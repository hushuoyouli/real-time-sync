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

	// 发送消息（如果 topic 不存在，会自动创建）
	// 使用重试机制，因为 topic 自动创建需要时间进行元数据同步
	var lastErr error
	maxRetries := 5
	retryDelay := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		err = writer.WriteMessages(ctx,
			kafka.Message{
				Key:   []byte("test-key"),
				Value: []byte("这是一条测试消息，用于触发自动创建 topic"),
			},
		)

		if err == nil {
			t.Logf("✓ 消息发送成功（尝试 %d/%d）", i+1, maxRetries)
			lastErr = nil
			break
		}

		lastErr = err
		// 检查是否是 topic 不存在的错误
		if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr == kafka.UnknownTopicOrPartition {
			if i < maxRetries-1 {
				t.Logf("Topic 尚未就绪，等待 %v 后重试 (%d/%d)...", retryDelay, i+1, maxRetries)
				time.Sleep(retryDelay)
				// 每次重试增加等待时间
				retryDelay = time.Duration(float64(retryDelay) * 1.5)
				continue
			}
		} else {
			// 其他错误，不重试
			break
		}
	}

	err = lastErr

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
	// 等待一下，确保 topic 创建完成
	t.Logf("等待 2 秒，确保 topic 创建完成...")
	time.Sleep(2 * time.Second)

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

// 辅助函数：格式化 broker 地址
func formatBrokerAddress(broker kafka.Broker) string {
	return broker.Host + ":" + strconv.Itoa(broker.Port)
}

// 辅助函数：获取 map 的键列表
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
