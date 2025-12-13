package runtime

import (
	"time"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/rlog"
	"github.com/segmentio/kafka-go"
)

type KafkaRuntimeEventHandle struct {
	EmptyRuntimeEventHandle
	//kafkaClient *kafka.Client

	handle    iface.IRuntimeEventHandle
	topicName string
	broker    string
	conn      *kafka.Conn
	writer    *kafka.Writer
	log       rlog.ILogger
}

// topicName,可以是场景的名字，或者是一场战斗的名字，系统自动加上时间
// broker,可以是kafka的地址等等
func NewKafkaRuntimeEventHandle(handle iface.IRuntimeEventHandle, topicName string, broker string, log rlog.ILogger) (*KafkaRuntimeEventHandle, error) {
	// 格式化当前时间，返回形如"20060102-150405"的字符串
	nowTimeStr := time.Now().Format("20060102-150405")
	topicName = topicName + "-" + nowTimeStr

	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return nil, err
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(broker),
		Topic:                  topicName, // 在 Writer 中指定 topic
		Balancer:               &kafka.CRC32Balancer{},
		AllowAutoTopicCreation: true, // 允许自动创建 topic
		WriteTimeout:           10 * time.Second,
		RequiredAcks:           kafka.RequireAll,
	}

	//defer conn.Close()
	return &KafkaRuntimeEventHandle{
		handle:    handle,
		topicName: topicName,
		broker:    broker,
		conn:      conn,
		writer:    writer,
		log:       log,
	}, nil
}

func (p *KafkaRuntimeEventHandle) Close() {
	p.writer.Close()
	p.conn.Close()
}

// 等待topic就绪
func (pt *KafkaRuntimeEventHandle) waitForTopicReady(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	checkInterval := 200 * time.Millisecond // 每 200ms 检查一次

	for time.Now().Before(deadline) {
		partitions, err := pt.conn.ReadPartitions()
		if err == nil {
			for _, p := range partitions {
				if p.Topic == pt.topicName {
					// Topic 存在，再等待一小段时间确保元数据完全同步
					time.Sleep(300 * time.Millisecond)
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

	return false
}

func (p *KafkaRuntimeEventHandle) PostInitialize(behaviorTree iface.IBehaviorTree, nowtimestampInMilli int64) {
}

func (p *KafkaRuntimeEventHandle) PostOnComplete(behaviorTree iface.IBehaviorTree, nowtimestampInMilli int64) {
}

func (p *KafkaRuntimeEventHandle) NewStack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData) {
}

func (p *KafkaRuntimeEventHandle) RemoveStack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData, nowtimestampInMilli int64) {
}
