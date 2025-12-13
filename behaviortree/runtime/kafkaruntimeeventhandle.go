package runtime

import (
	"context"
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

// 触发自动创建 topic的消息
func newInitializeTopicMsg() kafka.Message {
	return kafka.Message{
		Key:   []byte("initialize_topic"),
		Value: []byte("触发自动创建 topic"),
	}
}

// topicName,可以是场景的名字，或者是一场战斗的名字，系统自动加上时间
// broker,可以是kafka的地址等等
func NewKafkaRuntimeEventHandle(handle iface.IRuntimeEventHandle, topicName string, broker string, log rlog.ILogger) (*KafkaRuntimeEventHandle, error) {
	// 格式化当前时间，返回形如"20060102-150405"的字符串
	nowTimeStr := time.Now().Format("20060102-150405.000")
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

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Tracef("正在向 Topic %s 发送消息...", topicName)
	log.Tracef("Writer 配置: AllowAutoTopicCreation=%v, Topic=%s", writer.AllowAutoTopicCreation, writer.Topic)

	p := &KafkaRuntimeEventHandle{
		handle:    handle,
		topicName: topicName,
		broker:    broker,
		conn:      conn,
		writer:    writer,
		log:       log,
	}

	// 策略：先尝试发送消息（这会触发 topic 自动创建）
	startCreate := time.Now()
	err = writer.WriteMessages(ctx, newInitializeTopicMsg())
	if err != nil {
		if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr == kafka.UnknownTopicOrPartition {
			log.Tracef("检测到 Topic 不存在错误，等待元数据同步...")
			if p.waitForTopicReady(15 * time.Second) {
				log.Tracef("Topic 元数据已同步，重试发送消息...")
			} else {
				log.Errorf("等待topic就绪超时:%s", err.Error())
				writer.Close()
				conn.Close()
				return nil, err
			}
		} else {
			writer.Close()
			conn.Close()
			log.Errorf("发送消息失败: %v", err)
			return nil, err
		}
	}
	createDuration := time.Since(startCreate)
	log.Tracef("KafkaRuntimeEventHandle 创建 topic 的耗时: %v", createDuration)

	//defer conn.Close()
	return p, nil
}

func (p *KafkaRuntimeEventHandle) Close() {
	p.writer.Close()
	p.conn.Close()
}

// 等待topic就绪
func (pt *KafkaRuntimeEventHandle) waitForTopicReady(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	checkInterval := 200 * time.Millisecond // 每 200ms 检查一次

	pt.log.Tracef("等待topic%s就绪，最多等待%f秒", pt.topicName, timeout.Seconds())
	for time.Now().Before(deadline) {
		partitions, err := pt.conn.ReadPartitions()
		if err == nil {
			for _, p := range partitions {
				if p.Topic == pt.topicName {
					// Topic 存在，再等待一小段时间确保元数据完全同步
					time.Sleep(300 * time.Millisecond)
					pt.log.Tracef("topic%s就绪", pt.topicName)
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
	p.handle.PostInitialize(behaviorTree, nowtimestampInMilli)
}

func (p *KafkaRuntimeEventHandle) PostOnComplete(behaviorTree iface.IBehaviorTree, nowtimestampInMilli int64) {
	p.handle.PostOnComplete(behaviorTree, nowtimestampInMilli)
}

func (p *KafkaRuntimeEventHandle) NewStack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData) {
	p.handle.NewStack(behaviorTree, data)
}

func (p *KafkaRuntimeEventHandle) RemoveStack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData, nowtimestampInMilli int64) {
	p.handle.RemoveStack(behaviorTree, data, nowtimestampInMilli)
}

func (p *KafkaRuntimeEventHandle) PreOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask) {
	p.handle.PreOnStart(behaviorTree, taskRuntimeData, stackRuntimeData, task)
}

func (p *KafkaRuntimeEventHandle) PostOnUpdate(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, status iface.TaskStatus) {
	p.handle.PostOnUpdate(behaviorTree, taskRuntimeData, stackRuntimeData, task, nowtimestampInMilli, status)
}

func (p *KafkaRuntimeEventHandle) PostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64) {
	p.handle.PostOnEnd(behaviorTree, taskRuntimeData, stackRuntimeData, task, nowtimestampInMilli)
}

func (p *KafkaRuntimeEventHandle) ActionPostOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, datas [][]byte) {
	p.handle.ActionPostOnStart(behaviorTree, taskRuntimeData, stackRuntimeData, task, datas)
}

func (p *KafkaRuntimeEventHandle) ActionPostOnUpdate(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, status iface.TaskStatus, datas [][]byte) {
	p.handle.ActionPostOnUpdate(behaviorTree, taskRuntimeData, stackRuntimeData, task, nowtimestampInMilli, status, datas)
}

func (p *KafkaRuntimeEventHandle) ActionPostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, datas [][]byte) {
	p.handle.ActionPostOnEnd(behaviorTree, taskRuntimeData, stackRuntimeData, task, nowtimestampInMilli, datas)
}

func (p *KafkaRuntimeEventHandle) ParallelPreOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask) {
	p.handle.ParallelPreOnStart(behaviorTree, taskRuntimeData, stackRuntimeData, task)
}

func (p *KafkaRuntimeEventHandle) ParallelPostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64) {
	p.handle.ParallelPostOnEnd(behaviorTree, taskRuntimeData, stackRuntimeData, task, nowtimestampInMilli)
}

func (p *KafkaRuntimeEventHandle) ParallelAddChildStack(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, childStackRuntimeData *iface.StackRuntimeData) {
	p.handle.ParallelAddChildStack(behaviorTree, taskRuntimeData, stackRuntimeData, task, childStackRuntimeData)
}

func (p *KafkaRuntimeEventHandle) ParallelRemoveChildStack(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, childStackRuntimeData *iface.StackRuntimeData, nowtimestampInMilli int64) {
	p.handle.ParallelRemoveChildStack(behaviorTree, taskRuntimeData, stackRuntimeData, task, childStackRuntimeData, nowtimestampInMilli)
}
