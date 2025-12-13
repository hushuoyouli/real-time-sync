package runtime

import (
	"context"
	"strconv"
	"strings"
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
	// 优化配置以减少延迟，同时保持 RequireOne 确认
	// 关键优化：设置 BatchTimeout=0 和 BatchSize=1，立即发送，不等待批量
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(broker),
		Topic:                  topicName,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,             // 允许自动创建 topic
		WriteTimeout:           5 * time.Second,  // 减少超时时间，快速失败
		RequiredAcks:           kafka.RequireOne, // 至少需要一个 broker 确认
		Async:                  false,            // 同步发送，确保每条消息都确认
		BatchSize:              1,                // 批量大小为1，当累积1条消息时立即发送
		BatchTimeout:           0,                // 不等待批量超时，立即发送（这是关键！）
		BatchBytes:             0,                // 不限制批量字节数
		// 不启用压缩，减少 CPU 开销和延迟
		// 注意：
		// 1. kafka-go 默认 BatchTimeout 可能是 1 秒，这就是为什么每条消息都正好 1 秒的原因
		// 2. BatchSize=1 表示当 Writer 内部累积的消息达到 1 条时触发发送
		// 3. 但是，如果一次 WriteMessages() 调用传入多个消息，这些消息会作为一个批次一起发送（更高效）
		//    例如：WriteMessages(ctx, msg1, msg2, msg3) 会一次性发送 3 条消息，而不是分 3 次发送
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
	const messageCount = 10000
	const batchSize = 5000 // 每次批量发送的消息数量
	t.Logf("开始发送 %d 条消息（批量发送，每次 %d 条）...", messageCount, batchSize)
	t.Logf("当前配置: RequiredAcks=%v, WriteTimeout=%v, Async=%v, BatchSize=%d, BatchTimeout=%v",
		writer.RequiredAcks, writer.WriteTimeout, writer.Async, writer.BatchSize, writer.BatchTimeout)

	// 测试网络延迟
	networkTestStart := time.Now()
	testConn, testErr := kafka.Dial("tcp", broker)
	if testErr == nil {
		testConn.Close()
		networkLatency := time.Since(networkTestStart)
		t.Logf("网络连接测试耗时: %v", networkLatency)
		if networkLatency > 100*time.Millisecond {
			t.Logf("⚠️  网络延迟较高，可能影响发送性能")
		}
	}

	startTime := time.Now()

	// 批量发送消息，并记录每个批次的耗时
	var batchTimes []time.Duration
	totalSent := 0

	for i := 0; i < messageCount; i += batchSize {
		batchStartTime := time.Now()

		// 构建当前批次的消息
		end := i + batchSize
		if end > messageCount {
			end = messageCount
		}

		messages := make([]kafka.Message, 0, end-i)
		for j := i; j < end; j++ {
			messages = append(messages, kafka.Message{
				Key:   []byte("key-" + strconv.Itoa(j)),
				Value: []byte("消息内容-" + strconv.Itoa(j) + "-" + time.Now().Format("20060102-150405.000")),
			})
		}

		// 使用带超时的 context，避免长时间等待
		msgCtx, msgCancel := context.WithTimeout(context.Background(), 30*time.Second)

		// 批量发送当前批次的所有消息
		err = writer.WriteMessages(msgCtx, messages...)

		msgCancel() // 立即取消 context
		batchElapsed := time.Since(batchStartTime)
		batchTimes = append(batchTimes, batchElapsed)

		if err != nil {
			t.Fatalf("发送消息批次失败 (消息 %d-%d): %v", i, end-1, err)
		}

		totalSent += len(messages)

		// 输出每个批次的耗时和性能
		avgTimePerMsg := batchElapsed / time.Duration(len(messages))
		throughput := float64(len(messages)) / batchElapsed.Seconds()
		t.Logf("批次 %d: 发送 %d 条消息，耗时 %v (平均每条: %v, 吞吐量: %.2f 消息/秒)",
			len(batchTimes), len(messages), batchElapsed, avgTimePerMsg, throughput)

		// 如果批次耗时超过1秒，输出警告
		if batchElapsed > 1*time.Second {
			t.Logf("⚠️  批次 %d 耗时异常: %v，可能存在问题", len(batchTimes), batchElapsed)
		}
	}

	elapsed := time.Since(startTime)

	// 步骤4: 输出统计信息
	t.Logf("\n=== 性能统计 ===")
	t.Logf("✓ 成功发送 %d 条消息（共 %d 个批次，每批次 %d 条）", totalSent, len(batchTimes), batchSize)
	t.Logf("✓ 总耗时: %v", elapsed)
	t.Logf("✓ 平均每条消息耗时: %v", elapsed/time.Duration(totalSent))
	t.Logf("✓ 平均每批次耗时: %v", elapsed/time.Duration(len(batchTimes)))
	t.Logf("✓ 总吞吐量: %.2f 消息/秒", float64(totalSent)/elapsed.Seconds())
	t.Logf("✓ 总吞吐量: %.2f 消息/毫秒", float64(totalSent)/float64(elapsed.Milliseconds()))

	// 详细分析每个批次的耗时
	if len(batchTimes) > 0 {
		var minTime, maxTime, totalTime time.Duration
		minTime = batchTimes[0]
		maxTime = batchTimes[0]
		for _, bt := range batchTimes {
			if bt < minTime {
				minTime = bt
			}
			if bt > maxTime {
				maxTime = bt
			}
			totalTime += bt
		}
		t.Logf("\n=== 详细性能分析 ===")
		t.Logf("最短批次耗时: %v (平均每条: %v)", minTime, minTime/time.Duration(batchSize))
		t.Logf("最长批次耗时: %v (平均每条: %v)", maxTime, maxTime/time.Duration(batchSize))
		t.Logf("平均批次耗时: %v (平均每条: %v)", totalTime/time.Duration(len(batchTimes)), (totalTime/time.Duration(len(batchTimes)))/time.Duration(batchSize))

		// 分析可能的原因
		t.Logf("\n=== 性能分析建议 ===")

		// 计算平均每条消息的耗时
		avgPerMsg := maxTime / time.Duration(batchSize)

		// 检查是否是 BatchTimeout 导致的延迟（正好1秒左右）
		if avgPerMsg > 900*time.Millisecond && avgPerMsg < 1100*time.Millisecond {
			t.Logf("🔍 检测到平均每条消息延迟正好在 1 秒左右（%v），这很可能是 BatchTimeout 导致的！", avgPerMsg)
			t.Logf("   kafka-go Writer 默认 BatchTimeout 可能是 1 秒，即使只发送一条消息也会等待")
			t.Logf("   ✓ 已设置 BatchTimeout=0 和 BatchSize=1，应该能解决此问题")
			t.Logf("   如果仍然很慢，请检查其他配置")
		}

		if avgPerMsg > 500*time.Millisecond {
			t.Logf("⚠️⚠️  平均每条消息耗时异常（>500ms），严重性能问题！")
			t.Logf("   可能原因：")
			t.Logf("   1. BatchTimeout 配置问题（已设置 BatchTimeout=0，BatchSize=1）")
			t.Logf("   2. Kafka broker 配置问题（检查 server.properties）")
			t.Logf("   3. 网络延迟或丢包（检查: ping localhost, telnet localhost 9092）")
			t.Logf("   4. Kafka broker 负载过高或资源不足")
			t.Logf("   5. 防火墙或网络配置问题")
			t.Logf("   6. Docker 容器网络配置问题（如果使用 Docker）")
			t.Logf("\n   诊断步骤：")
			t.Logf("   1. 检查 Kafka broker 日志: docker logs kafka")
			t.Logf("   2. 测试网络延迟: ping localhost")
			t.Logf("   3. 测试端口连接: telnet localhost 9092")
			t.Logf("   4. 检查 Kafka broker 配置:")
			t.Logf("      - socket.request.max.bytes")
			t.Logf("      - socket.send.buffer.bytes")
			t.Logf("      - socket.receive.buffer.bytes")
			t.Logf("   5. 检查系统资源: CPU、内存、磁盘 I/O")
		} else if avgPerMsg > 100*time.Millisecond {
			t.Logf("⚠️  平均每条消息耗时较长（>100ms），可能原因：")
			t.Logf("   1. RequiredAcks=%v 需要等待 broker 确认（网络往返时间）", writer.RequiredAcks)
			t.Logf("   2. 网络延迟较高（本地应该 <10ms）")
			t.Logf("   3. Kafka broker 处理较慢")
			t.Logf("\n   优化建议：")
			t.Logf("   - 检查网络延迟: ping localhost（应该 <1ms）")
			t.Logf("   - 检查 Kafka broker 状态和性能")
			t.Logf("   - 如果使用 Docker，检查容器网络配置")
			t.Logf("   - 检查 Kafka broker 日志是否有错误或警告")
		} else if avgPerMsg > 10*time.Millisecond {
			t.Logf("ℹ️  平均每条消息耗时中等（10-100ms）")
			t.Logf("   当前 RequiredAcks=%v 需要等待 broker 确认", writer.RequiredAcks)
			t.Logf("   批量发送 %d 条消息，平均每条 %v，性能良好", batchSize, avgPerMsg)
			t.Logf("   对于本地 Kafka，这个时间可能偏长，建议检查网络配置")
		} else {
			t.Logf("✓ 平均每条消息耗时较短（<10ms），性能良好")
			t.Logf("   批量发送 %d 条消息，平均每条 %v，吞吐量优秀", batchSize, avgPerMsg)
		}
	}
}

// TestKafkaFindOptimalBatchSize 测试不同批次大小，找出最佳批次大小
func TestKafkaFindOptimalBatchSize(t *testing.T) {
	// Kafka broker 地址
	broker := "localhost:9092"

	// 创建 Kafka 连接
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		t.Fatalf("连接 Kafka 失败: %v", err)
	}
	defer conn.Close()

	t.Logf("成功连接到 Kafka")

	// 测试不同的批次大小
	batchSizes := []int{100, 500, 1000, 2000, 5000}
	const messageCount = 10000
	const testRuns = 2 // 每个批次大小测试2次，取平均值

	type BatchResult struct {
		BatchSize     int
		TotalTime     time.Duration
		AvgTimePerMsg time.Duration
		Throughput    float64 // 消息/秒
		BatchCount    int
		AvgBatchTime  time.Duration
	}

	results := make([]BatchResult, 0, len(batchSizes))

	for _, batchSize := range batchSizes {
		separator := strings.Repeat("=", 80)
		t.Logf("\n%s", separator)
		t.Logf("测试批次大小: %d 条消息", batchSize)
		t.Logf("%s", separator)

		// 生成一个随机的 topic 名称
		topicName := "test-batch-" + strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + strconv.Itoa(batchSize)

		// 创建 Writer
		writer := &kafka.Writer{
			Addr:                   kafka.TCP(broker),
			Topic:                  topicName,
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
			WriteTimeout:           30 * time.Second,
			RequiredAcks:           kafka.RequireOne,
			Async:                  false,
			BatchSize:              1,
			BatchTimeout:           0,
			BatchBytes:             0,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		// 触发 topic 创建
		err = writer.WriteMessages(ctx,
			kafka.Message{
				Key:   []byte("init"),
				Value: []byte("初始化"),
			},
		)
		if err != nil {
			if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr == kafka.UnknownTopicOrPartition {
				waitForTopicReady(conn, topicName, 10*time.Second, t)
				err = writer.WriteMessages(ctx,
					kafka.Message{
						Key:   []byte("init"),
						Value: []byte("初始化"),
					},
				)
			}
			if err != nil {
				t.Logf("⚠️  批次大小 %d: 初始化失败，跳过: %v", batchSize, err)
				writer.Close()
				continue
			}
		}

		// 等待 topic 就绪
		waitForTopicReady(conn, topicName, 5*time.Second, t)

		var totalTime time.Duration
		var totalBatches int

		// 运行多次测试取平均值
		for run := 0; run < testRuns; run++ {
			startTime := time.Now()
			batches := 0

			for i := 0; i < messageCount; i += batchSize {
				end := i + batchSize
				if end > messageCount {
					end = messageCount
				}

				messages := make([]kafka.Message, 0, end-i)
				for j := i; j < end; j++ {
					messages = append(messages, kafka.Message{
						Key:   []byte("key-" + strconv.Itoa(j)),
						Value: []byte("消息-" + strconv.Itoa(j)),
					})
				}

				msgCtx, msgCancel := context.WithTimeout(context.Background(), 60*time.Second)
				err = writer.WriteMessages(msgCtx, messages...)
				msgCancel()

				if err != nil {
					t.Logf("⚠️  批次大小 %d, 运行 %d: 发送失败: %v", batchSize, run+1, err)
					break
				}

				batches++
			}

			elapsed := time.Since(startTime)
			totalTime += elapsed
			totalBatches += batches

			if run == 0 {
				t.Logf("  运行 %d: 耗时 %v, 批次 %d", run+1, elapsed, batches)
			}
		}

		writer.Close()

		// 计算平均值
		avgTime := totalTime / time.Duration(testRuns)
		avgBatches := totalBatches / testRuns
		avgTimePerMsg := avgTime / time.Duration(messageCount)
		throughput := float64(messageCount) / avgTime.Seconds()
		avgBatchTime := avgTime / time.Duration(avgBatches)

		result := BatchResult{
			BatchSize:     batchSize,
			TotalTime:     avgTime,
			AvgTimePerMsg: avgTimePerMsg,
			Throughput:    throughput,
			BatchCount:    avgBatches,
			AvgBatchTime:  avgBatchTime,
		}
		results = append(results, result)

		t.Logf("  平均总耗时: %v", avgTime)
		t.Logf("  平均每条消息耗时: %v", avgTimePerMsg)
		t.Logf("  吞吐量: %.2f 消息/秒", throughput)
		t.Logf("  平均每批次耗时: %v", avgBatchTime)
	}

	// 输出对比结果
	separator := strings.Repeat("=", 80)
	lineSeparator := strings.Repeat("-", 80)
	t.Logf("\n%s", separator)
	t.Logf("批次大小性能对比总结")
	t.Logf("%s", separator)
	t.Logf("%-10s %-15s %-20s %-15s %-15s", "批次大小", "总耗时", "平均每条耗时", "吞吐量(消息/秒)", "平均批次耗时")
	t.Logf("%s", lineSeparator)

	bestThroughput := 0.0
	bestBatchSize := 0
	bestResult := BatchResult{}

	for _, r := range results {
		t.Logf("%-10d %-15v %-20v %-15.2f %-15v",
			r.BatchSize, r.TotalTime, r.AvgTimePerMsg, r.Throughput, r.AvgBatchTime)
		if r.Throughput > bestThroughput {
			bestThroughput = r.Throughput
			bestBatchSize = r.BatchSize
			bestResult = r
		}
	}

	t.Logf("%s", separator)
	t.Logf("\n🏆 最佳批次大小建议:")
	t.Logf("   批次大小: %d 条消息", bestBatchSize)
	t.Logf("   总耗时: %v", bestResult.TotalTime)
	t.Logf("   平均每条消息耗时: %v", bestResult.AvgTimePerMsg)
	t.Logf("   吞吐量: %.2f 消息/秒 (最高)", bestResult.Throughput)
	t.Logf("   平均每批次耗时: %v", bestResult.AvgBatchTime)
	t.Logf("\n💡 建议:")
	if bestBatchSize <= 1000 {
		t.Logf("   - 对于实时性要求高的场景，建议使用较小的批次大小 (%d)", bestBatchSize)
	} else if bestBatchSize <= 2000 {
		t.Logf("   - 对于平衡性能和延迟的场景，建议使用中等批次大小 (%d)", bestBatchSize)
	} else {
		t.Logf("   - 对于高吞吐量场景，建议使用较大的批次大小 (%d)", bestBatchSize)
	}
	t.Logf("   - 批次大小 %d 提供了最佳的吞吐量性能", bestBatchSize)
}
