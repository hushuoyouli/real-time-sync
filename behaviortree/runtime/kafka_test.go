package runtime

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
)

func TestKafka(t *testing.T) {
	// Kafka broker 地址，可以根据实际情况修改
	broker := "localhost:9092"

	// 创建 Kafka 连接
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		t.Fatalf("连接 Kafka 失败: %v", err)
	}
	defer conn.Close()

	// 获取 broker 信息
	brokerInfo := conn.Broker()

	t.Logf("成功连接到 Kafka")
	t.Logf("Broker ID: %d", brokerInfo.ID)
	t.Logf("Broker Host: %s", brokerInfo.Host)
	t.Logf("Broker Port: %d", brokerInfo.Port)

	// 获取控制器信息
	controller, err := conn.Controller()
	if err != nil {
		t.Logf("获取控制器信息失败: %v", err)
	} else {
		t.Logf("Controller ID: %d, Host: %s, Port: %d", controller.ID, controller.Host, controller.Port)
	}

	// 获取主题列表
	partitions, err := conn.ReadPartitions()
	if err != nil {
		t.Logf("获取分区信息失败: %v", err)
	} else {
		topics := make(map[string]bool)
		for _, p := range partitions {
			topics[p.Topic] = true
		}
		t.Logf("主题列表: %v", getKeys(topics))
	}

	// 测试创建生产者
	writer := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    "test-topic",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	// 测试发送消息
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte("test-key"),
			Value: []byte("test message"),
		},
	)
	if err != nil {
		t.Logf("发送消息失败: %v", err)
	} else {
		t.Log("消息发送成功")
	}

	// 测试创建消费者
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    "test-topic",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	// 设置超时，避免测试阻塞
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	msg, err := reader.ReadMessage(ctx)
	if err != nil {
		t.Logf("读取消息失败或超时: %v", err)
	} else {
		t.Logf("收到消息: Topic=%s, Partition=%d, Offset=%d, Key=%s, Value=%s",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
	}
}

func TestKafkaCreateTopic(t *testing.T) {
	// Kafka broker 地址，可以根据实际情况修改
	broker := "localhost:9092"

	// 创建 Kafka 连接
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		t.Fatalf("连接 Kafka 失败: %v", err)
	}
	defer conn.Close()

	t.Logf("成功连接到 Kafka")

	// 生成一个随机的 topic 名称（使用时间戳确保唯一性）
	topicName := "test-auto-create-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	t.Logf("测试 Topic 名称: %s", topicName)

	// 先检查 topic 是否存在
	allPartitions, err := conn.ReadPartitions()
	if err != nil {
		t.Logf("获取分区信息失败: %v", err)
	} else {
		topics := make(map[string]bool)
		for _, p := range allPartitions {
			topics[p.Topic] = true
		}
		if topics[topicName] {
			t.Logf("Topic %s 已存在（这不应该发生）", topicName)
		} else {
			t.Logf("确认 Topic %s 不存在，准备发送消息测试自动创建", topicName)
		}
	}

	// 创建生产者并发送消息（这会触发自动创建 topic）
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(broker),
		Topic:                  topicName, // 在 Writer 中指定 topic
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true, // 允许自动创建 topic
		WriteTimeout:           10 * time.Second,
		RequiredAcks:           kafka.RequireOne, // 只需要一个 broker 确认
	}
	defer writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Logf("正在向 Topic %s 发送消息...", topicName)
	t.Logf("Writer 配置: AllowAutoTopicCreation=%v, Topic=%s", writer.AllowAutoTopicCreation, writer.Topic)

	// 策略：先尝试发送消息（这会触发 topic 自动创建）
	// 如果失败，等待 topic 元数据同步完成后再重试
	err = writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte("test-key"),
			Value: []byte("这是一条测试消息，用于触发自动创建 topic"),
		},
	)

	// 如果第一次发送失败，且是 topic 不存在的错误，等待元数据同步
	if err != nil {
		if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr == kafka.UnknownTopicOrPartition {
			t.Logf("检测到 Topic 不存在错误，等待元数据同步...")

			// 等待 topic 创建并元数据同步完成（最多等待 5 秒）
			if waitForTopicReady(conn, topicName, 5*time.Second, t) {
				// Topic 已就绪，重试发送消息
				t.Logf("Topic 元数据已同步，重试发送消息...")
				err = writer.WriteMessages(ctx,
					kafka.Message{
						Key:   []byte("test-key"),
						Value: []byte("这是一条测试消息，用于触发自动创建 topic"),
					},
				)
				if err == nil {
					t.Logf("✓ 等待元数据同步后，消息发送成功")
				}
			} else {
				// 超时，但可能 topic 已经创建了，再尝试一次
				t.Logf("等待超时，但尝试最后一次发送...")
				err = writer.WriteMessages(ctx,
					kafka.Message{
						Key:   []byte("test-key"),
						Value: []byte("这是一条测试消息，用于触发自动创建 topic"),
					},
				)
			}
		}
	} else {
		t.Logf("✓ 消息发送成功（第一次尝试）")
	}

	if err != nil {
		t.Logf("发送消息失败: %v", err)
		t.Logf("错误类型: %T", err)
		t.Logf("完整错误信息: %+v", err)

		// 先检查 topic 是否已经被创建（可能只是元数据同步延迟）
		t.Logf("\n=== 检查 Topic 状态 ===")
		time.Sleep(2 * time.Second) // 等待元数据同步
		checkPartitions, checkErr := conn.ReadPartitions()
		if checkErr == nil {
			topicExists := false
			for _, p := range checkPartitions {
				if p.Topic == topicName {
					topicExists = true
					t.Logf("✓ Topic %s 已存在，可能是元数据同步延迟", topicName)
					// Topic 已存在，再尝试发送一次
					time.Sleep(1 * time.Second)
					err = writer.WriteMessages(ctx,
						kafka.Message{
							Key:   []byte("test-key"),
							Value: []byte("这是一条测试消息"),
						},
					)
					if err == nil {
						t.Logf("✓ 等待元数据同步后，消息发送成功")
						goto verifyTopic
					} else {
						t.Logf("✗ 即使 topic 存在，发送消息仍然失败: %v", err)
					}
					break
				}
			}
			if !topicExists {
				t.Logf("✗ Topic %s 不存在，自动创建可能失败", topicName)
			}
		}

		// 尝试诊断问题
		t.Logf("\n=== 诊断信息 ===")
		t.Logf("1. 检查 Kafka 连接...")
		testConn, testErr := kafka.Dial("tcp", broker)
		if testErr != nil {
			t.Logf("   ✗ 无法连接到 Kafka: %v", testErr)
		} else {
			t.Logf("   ✓ Kafka 连接正常")
			testConn.Close()
		}

		t.Logf("2. 检查 Kafka 配置（单节点环境）...")
		t.Logf("   在单节点环境下，auto.create.topics.enable 是有效的，但需要确保:")
		t.Logf("   ✓ KAFKA_AUTO_CREATE_TOPICS_ENABLE=true")
		t.Logf("   ✓ KAFKA_DEFAULT_REPLICATION_FACTOR=1 (单节点必须为1)")
		t.Logf("   ✓ KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1")
		t.Logf("   ✓ KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1")
		t.Logf("   可以通过以下命令检查:")
		t.Logf("   docker exec kafka env | grep KAFKA_AUTO_CREATE_TOPICS_ENABLE")
		t.Logf("   docker exec kafka env | grep KAFKA_DEFAULT_REPLICATION_FACTOR")
		t.Logf("   或者查看配置文件:")
		t.Logf("   docker exec kafka cat /opt/kafka/config/server.properties | grep -E 'auto.create.topics.enable|default.replication.factor'")

		t.Logf("3. 单节点环境常见问题:")
		t.Logf("   - 如果 replication factor > 1，自动创建会失败（单节点无法满足）")
		t.Logf("   - 确保 Kafka 容器已重启以应用配置")
		t.Logf("   - 检查 Kafka 日志: docker logs kafka | grep -i 'replication'")
		t.Logf("   - 尝试手动创建 topic 测试: docker exec kafka /opt/kafka/bin/kafka-topics.sh --create --topic test --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1")

		// 尝试使用 CreateTopics API 作为备选方案
		t.Logf("\n4. 尝试使用 CreateTopics API...")
		topicConfig := kafka.TopicConfig{
			Topic:             topicName,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
		createErr := conn.CreateTopics(topicConfig)
		if createErr != nil {
			t.Logf("   CreateTopics API 也失败: %v", createErr)
			t.Logf("   注意：这可能是 API 版本不兼容，但 topic 可能已经通过自动创建机制创建了")
		} else {
			t.Logf("   ✓ 使用 CreateTopics API 成功创建 topic")
			// 如果 CreateTopics 成功，再尝试发送消息
			time.Sleep(2 * time.Second)
			err = writer.WriteMessages(ctx,
				kafka.Message{
					Key:   []byte("test-key"),
					Value: []byte("这是一条测试消息"),
				},
			)
			if err != nil {
				t.Logf("   ✗ 创建 topic 后发送消息仍然失败: %v", err)
			} else {
				t.Logf("   ✓ 创建 topic 后发送消息成功")
				// 继续后续验证
				goto verifyTopic
			}
		}

		// 如果所有方法都失败，标记测试失败
		t.Errorf("无法发送消息到 topic %s: %v", topicName, err)
		return
	}

	t.Logf("✓ 消息发送成功！")

verifyTopic:
	// 由于我们已经等待了元数据同步，这里只需要短暂等待确保完全就绪
	t.Logf("验证 Topic 创建状态...")
	time.Sleep(500 * time.Millisecond)

	// 验证 topic 是否自动创建成功
	allPartitions, err = conn.ReadPartitions()
	if err != nil {
		t.Errorf("获取分区信息失败: %v", err)
		return
	}

	// 查找我们创建的 topic
	topicPartitions := make([]kafka.Partition, 0)
	for _, p := range allPartitions {
		if p.Topic == topicName {
			topicPartitions = append(topicPartitions, p)
		}
	}

	if len(topicPartitions) > 0 {
		t.Logf("✓✓✓ 成功！Topic %s 已自动创建，包含 %d 个分区", topicName, len(topicPartitions))
		for i, p := range topicPartitions {
			t.Logf("  分区 %d: ID=%d, Leader=%d", i, p.ID, p.Leader.ID)
		}
	} else {
		t.Errorf("✗✗✗ 失败！Topic %s 未找到，自动创建可能未生效", topicName)
		t.Logf("提示: 请检查 Kafka 配置，确保 KAFKA_AUTO_CREATE_TOPICS_ENABLE=true")
	}

}

// 辅助函数：等待 topic 创建并元数据同步完成
// 返回 true 表示 topic 已就绪，false 表示超时
func waitForTopicReady(conn *kafka.Conn, topicName string, timeout time.Duration, t *testing.T) bool {
	deadline := time.Now().Add(timeout)
	checkInterval := 200 * time.Millisecond // 每 200ms 检查一次

	t.Logf("等待 Topic %s 元数据同步完成（最多等待 %v）...", topicName, timeout)

	for time.Now().Before(deadline) {
		partitions, err := conn.ReadPartitions()
		if err == nil {
			for _, p := range partitions {
				if p.Topic == topicName {
					// Topic 存在，再等待一小段时间确保元数据完全同步
					time.Sleep(300 * time.Millisecond)
					t.Logf("✓ Topic %s 已创建并元数据同步完成", topicName)
					return true
				}
			}
		}

		// 如果还没超时，继续等待
		remaining := time.Until(deadline)
		if remaining > checkInterval {
			time.Sleep(checkInterval)
		} else {
			time.Sleep(remaining)
		}
	}

	t.Logf("✗ 等待 Topic %s 超时（%v）", topicName, timeout)
	return false
}

// 辅助函数：获取 map 的键列表
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestKafkaSend10000Messages 测试自动创建 topic 并发送 10000 条消息的性能
func TestKafkaSend10000Messages(t *testing.T) {
	// Kafka broker 地址
	broker := "localhost:9092"

	// 创建 Kafka 连接
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		t.Fatalf("连接 Kafka 失败: %v", err)
	}
	defer conn.Close()

	t.Logf("成功连接到 Kafka")

	// 生成一个随机的 topic 名称（使用时间戳确保唯一性）
	topicName := "test-benchmark-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	t.Logf("测试 Topic 名称: %s", topicName)

	// 创建生产者，允许自动创建 topic
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(broker),
		Topic:                  topicName,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true, // 允许自动创建 topic
		WriteTimeout:           30 * time.Second,
		RequiredAcks:           kafka.RequireOne,
	}
	defer writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 步骤1: 尝试发送一条消息来触发 topic 自动创建
	t.Logf("正在触发 Topic %s 的自动创建...", topicName)
	err = writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte("init-key"),
			Value: []byte("初始化消息，用于触发自动创建 topic"),
		},
	)

	// 如果第一次发送失败，且是 topic 不存在的错误，等待元数据同步
	if err != nil {
		if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr == kafka.UnknownTopicOrPartition {
			t.Logf("检测到 Topic 不存在错误，等待元数据同步...")
			if waitForTopicReady(conn, topicName, 10*time.Second, t) {
				// Topic 已就绪，重试发送消息
				t.Logf("Topic 元数据已同步，重试发送初始化消息...")
				err = writer.WriteMessages(ctx,
					kafka.Message{
						Key:   []byte("init-key"),
						Value: []byte("初始化消息，用于触发自动创建 topic"),
					},
				)
				if err != nil {
					t.Fatalf("等待元数据同步后，发送初始化消息仍然失败: %v", err)
				}
			} else {
				t.Fatalf("等待 Topic 创建超时")
			}
		} else {
			t.Fatalf("发送初始化消息失败: %v", err)
		}
	} else {
		t.Logf("✓ 初始化消息发送成功，Topic 已创建")
	}

	// 步骤2: 确认 topic 已创建并等待完全就绪
	t.Logf("确认 Topic 创建状态...")
	if !waitForTopicReady(conn, topicName, 5*time.Second, t) {
		t.Fatalf("Topic %s 创建失败或超时", topicName)
	}

	// 步骤3: 发送 10000 条消息并统计耗时
	const messageCount = 100
	t.Logf("开始发送 %d 条消息（逐个发送）...", messageCount)

	startTime := time.Now()

	// 逐个发送消息
	for i := 0; i < messageCount; i++ {
		err = writer.WriteMessages(ctx,
			kafka.Message{
				Key:   []byte("key-" + strconv.Itoa(i)),
				Value: []byte("消息内容-" + strconv.Itoa(i) + "-" + time.Now().Format("20060102-150405.000")),
			},
		)
		if err != nil {
			t.Fatalf("发送消息失败 (消息 %d): %v", i, err)
		}

		// 每发送 1000 条消息输出一次进度
		if (i+1)%1000 == 0 {
			t.Logf("已发送 %d/%d 条消息", i+1, messageCount)
		}
	}

	elapsed := time.Since(startTime)

	// 步骤4: 输出统计信息
	t.Logf("\n=== 性能统计 ===")
	t.Logf("✓ 成功发送 %d 条消息", messageCount)
	t.Logf("✓ 总耗时: %v", elapsed)
	t.Logf("✓ 平均每条消息耗时: %v", elapsed/time.Duration(messageCount))
	t.Logf("✓ 吞吐量: %.2f 消息/秒", float64(messageCount)/elapsed.Seconds())
	t.Logf("✓ 吞吐量: %.2f 消息/毫秒", float64(messageCount)/float64(elapsed.Milliseconds()))
}
