package runtime

import (
	"time"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
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
}

// topicName,可以是场景的名字，或者是一场战斗的名字，系统自动加上时间
// broker,可以是kafka的地址等等
func NewKafkaRuntimeEventHandle(handle iface.IRuntimeEventHandle, topicName string, broker string) (*KafkaRuntimeEventHandle, error) {
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
	}, nil
}
