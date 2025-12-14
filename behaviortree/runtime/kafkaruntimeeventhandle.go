package runtime

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
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
	//writer    *kafka.Writer
	log rlog.ILogger

	kafkaMessageChannel   chan kafka.Message
	messageWriteWaitGroup sync.WaitGroup
	messageWriteContext   context.Context
	messageWriteCancel    context.CancelFunc
	frameInterval         int //帧间隔，单位是毫秒，默认是100毫秒
}

// 触发自动创建 topic的消息

// topicName,可以是场景的名字，或者是一场战斗的名字，系统自动加上时间
// broker,可以是kafka的地址等等
func NewKafkaRuntimeEventHandle(handle iface.IRuntimeEventHandle, topicName string, broker string, log rlog.ILogger, frameInterval int) (*KafkaRuntimeEventHandle, error) {
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
		RequiredAcks:           kafka.RequireNone, //在局域网中，应该不会丢失
		BatchSize:              1,                 // 批量大小为1，立即发送
		BatchTimeout:           0,                 // 不等待批量超时，立即发送（这是关键！）
		BatchBytes:             0,                 // 不限制批量字节数
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Tracef("正在向 Topic %s 发送消息...", topicName)
	log.Tracef("Writer 配置: AllowAutoTopicCreation=%v, Topic=%s", writer.AllowAutoTopicCreation, writer.Topic)
	messageWriteContext, messageWriteCancel := context.WithCancel(context.Background())

	p := &KafkaRuntimeEventHandle{
		handle:    handle,
		topicName: topicName,
		broker:    broker,
		conn:      conn,
		//writer:                writer,
		log:                   log,
		messageWriteWaitGroup: sync.WaitGroup{},
		messageWriteContext:   messageWriteContext,
		messageWriteCancel:    messageWriteCancel,
		kafkaMessageChannel:   make(chan kafka.Message, 3000),
		frameInterval:         frameInterval,
	}

	// 策略：先尝试发送消息（这会触发 topic 自动创建）
	startCreate := time.Now()
	err = writer.WriteMessages(ctx, p.newInitializeTopicMsg())
	if err != nil {
		if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr == kafka.UnknownTopicOrPartition {
			log.Tracef("检测到 Topic 不存在错误，等待元数据同步...")
			if p.waitForTopicReady(15 * time.Second) {
				//log.Tracef("Topic 元数据已同步，重试发送消息...")
			} else {
				log.Errorf("等待topic就绪超时...")
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
	p.messageWriteWaitGroup.Add(1)
	createDuration := time.Since(startCreate)
	log.Tracef("KafkaRuntimeEventHandle 创建 topic 的耗时: %v", createDuration)

	go func() {
		defer p.messageWriteWaitGroup.Done()
		defer writer.Close()
		defer close(p.kafkaMessageChannel)
		stop := false
		msgCount := 0
		msgBatch := 500
		msgCache := make([]kafka.Message, 0, msgBatch)
		for {
			if stop {
				break
			}

			select {
			case msg := <-p.kafkaMessageChannel:
				msgCache = append(msgCache, msg)
				msgCount++
				if msgCount >= msgBatch {
					startTime := time.Now()
					err = writer.WriteMessages(context.Background(), msgCache...)
					if err != nil {
						p.log.Errorf("发送消息失败: %v", err)
					} else {
						p.log.Tracef("发送%d条消息成功，耗时%v,吞吐量%.2f条/秒,平均每条消息耗时%v", msgCount, time.Since(startTime), float64(msgCount)/time.Since(startTime).Seconds(), time.Since(startTime)/time.Duration(msgCount))
					}
					msgCache = msgCache[:0]
					msgCount = 0
				}
				// err = writer.WriteMessages(p.messageWriteContext, msg)
				// if err != nil {
				// 	p.log.Errorf("发送消息失败: %v", err)
				// }
			case <-p.messageWriteContext.Done():
				stop = true
			}
		}

		stop = false
		for {
			if stop {
				break
			}

			select {
			case msg := <-p.kafkaMessageChannel:
				msgCache = append(msgCache, msg)
				msgCount++
				if msgCount >= msgBatch {
					err = writer.WriteMessages(context.Background(), msgCache...)
					if err != nil {
						p.log.Errorf("发送消息失败: %v", err)
					}
					msgCache = msgCache[:0]
					msgCount = 0
				}
			default:
				stop = true
			}
		}

		if msgCount > 0 {
			startTime := time.Now()
			err = writer.WriteMessages(context.Background(), msgCache...)
			if err != nil {
				p.log.Errorf("发送消息失败: %v", err)
			} else {
				p.log.Tracef("发送%d条消息成功，耗时%v,吞吐量%.2f条/秒,平均每条消息耗时%v", msgCount, time.Since(startTime), float64(msgCount)/time.Since(startTime).Seconds(), time.Since(startTime)/time.Duration(msgCount))
			}
		}
	}()

	//defer conn.Close()
	return p, nil
}

func (p *KafkaRuntimeEventHandle) newInitializeTopicMsg() kafka.Message {
	return kafka.Message{
		Key:   []byte(p.topicName),
		Value: []byte("initialize_topic"),
	}
}

// 这个函数因为在最后会等待写入最后一批的消息，会有延迟和阻塞，所以需要异步关闭
func (p *KafkaRuntimeEventHandle) Close() {
	//p.writer.Close()
	p.messageWriteCancel()
	p.messageWriteWaitGroup.Wait()
	p.conn.Close()
}

func (p *KafkaRuntimeEventHandle) sendMessage(content []byte, nowtimestampInMilli int64, eventType string, behaviorTreeID int64, behaviorTreeName string, objectName string, objectID int64) {
	p.kafkaMessageChannel <- kafka.Message{
		Key:   []byte(p.topicName),
		Value: content,
		Headers: []kafka.Header{
			{Key: "nowtimestamp_milli", Value: []byte(strconv.FormatInt(nowtimestampInMilli, 10))}, //事件发生的时间
			{Key: "type", Value: []byte(eventType)},                                                //事件类型
			{Key: "behavior_tree_id", Value: []byte(strconv.FormatInt(behaviorTreeID, 10))},        //行为树的ID
			{Key: "behavior_tree_name", Value: []byte(behaviorTreeName)},                           //行为树的名称
			{Key: "object_name", Value: []byte(objectName)},                                        //对象的名称
			{Key: "object_id", Value: []byte(strconv.FormatInt(objectID, 10))},                     //对象的ID
		},
	}
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
	p.sendMessage([]byte(strconv.FormatInt(int64(p.frameInterval), 10)), nowtimestampInMilli, "behavior_tree_frame_interval", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //树的帧率
	p.sendMessage(behaviorTree.Config(), nowtimestampInMilli, "post_initialize", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID())                                              //树初始化
}

func (p *KafkaRuntimeEventHandle) PostOnComplete(behaviorTree iface.IBehaviorTree, nowtimestampInMilli int64) {
	p.handle.PostOnComplete(behaviorTree, nowtimestampInMilli)
	p.sendMessage(nil, nowtimestampInMilli, "post_complete", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //树完成
}

func (p *KafkaRuntimeEventHandle) NewStack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData) {
	p.handle.NewStack(behaviorTree, data)
	p.sendMessage([]byte(strconv.FormatInt(int64(data.StackID), 10)), behaviorTree.Clock().TimesampInMill(), "new_stack", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //新栈
}

func (p *KafkaRuntimeEventHandle) RemoveStack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData, nowtimestampInMilli int64) {
	p.handle.RemoveStack(behaviorTree, data, nowtimestampInMilli)
	p.sendMessage([]byte(strconv.FormatInt(int64(data.StackID), 10)), nowtimestampInMilli, "remove_stack", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //删除栈
}

func (p *KafkaRuntimeEventHandle) PreOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask) {
	p.handle.PreOnStart(behaviorTree, taskRuntimeData, stackRuntimeData, task)
	data := map[string]interface{}{
		"task_id":    taskRuntimeData.TaskID,
		"execute_id": taskRuntimeData.ExecuteID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		p.log.Errorf("序列化数据失败: %v", err)
		return
	}
	p.sendMessage(jsonData, taskRuntimeData.StartTime, "pre_on_start", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //预开始
}

func (p *KafkaRuntimeEventHandle) PostOnUpdate(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, status iface.TaskStatus) {
	p.handle.PostOnUpdate(behaviorTree, taskRuntimeData, stackRuntimeData, task, nowtimestampInMilli, status)
}

func (p *KafkaRuntimeEventHandle) PostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, status iface.TaskStatus) {
	p.handle.PostOnEnd(behaviorTree, taskRuntimeData, stackRuntimeData, task, nowtimestampInMilli, status)
	data := map[string]interface{}{
		"task_id":    taskRuntimeData.TaskID,
		"execute_id": taskRuntimeData.ExecuteID,
		"status":     status.ToString(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		p.log.Errorf("序列化数据失败: %v", err)
		return
	}
	p.sendMessage(jsonData, nowtimestampInMilli, "post_on_end", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //结束
}

func (p *KafkaRuntimeEventHandle) ActionPostOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, datas [][]byte) {
	p.handle.ActionPostOnStart(behaviorTree, taskRuntimeData, stackRuntimeData, task, datas)
	data := map[string]interface{}{
		"task_id":    taskRuntimeData.TaskID,
		"execute_id": taskRuntimeData.ExecuteID,
		"datas":      datas,
		"stack_id":   stackRuntimeData.StackID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		p.log.Errorf("序列化数据失败: %v", err)
		return
	}
	p.sendMessage(jsonData, taskRuntimeData.StartTime, "action_post_on_start", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //动作开始
}

func (p *KafkaRuntimeEventHandle) ActionPostOnUpdate(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, status iface.TaskStatus, datas [][]byte) {
	p.handle.ActionPostOnUpdate(behaviorTree, taskRuntimeData, stackRuntimeData, task, nowtimestampInMilli, status, datas)
	if len(datas) > 0 {
		data := map[string]interface{}{
			"task_id":    taskRuntimeData.TaskID,
			"execute_id": taskRuntimeData.ExecuteID,
			"datas":      datas,
			"stack_id":   stackRuntimeData.StackID,
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			p.log.Errorf("序列化数据失败: %v", err)
			return
		}
		p.sendMessage(jsonData, nowtimestampInMilli, "action_post_on_update", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //动作更新
	}
}

func (p *KafkaRuntimeEventHandle) ActionPostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, datas [][]byte) {
	p.handle.ActionPostOnEnd(behaviorTree, taskRuntimeData, stackRuntimeData, task, nowtimestampInMilli, datas)
	p.sendMessage(nil, nowtimestampInMilli, "action_post_on_end", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //动作结束
	//if len(datas) > 0 {
	data := map[string]interface{}{
		"task_id":    taskRuntimeData.TaskID,
		"execute_id": taskRuntimeData.ExecuteID,
		"datas":      datas,
		"stack_id":   stackRuntimeData.StackID,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		p.log.Errorf("序列化数据失败: %v", err)
		return
	}
	p.sendMessage(jsonData, nowtimestampInMilli, "action_post_on_end", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //动作结束
}

func (p *KafkaRuntimeEventHandle) ParallelPreOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask) {
	p.handle.ParallelPreOnStart(behaviorTree, taskRuntimeData, stackRuntimeData, task)
	data := map[string]interface{}{
		"task_id":    taskRuntimeData.TaskID,
		"execute_id": taskRuntimeData.ExecuteID,
		"stack_id":   stackRuntimeData.StackID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		p.log.Errorf("序列化数据失败: %v", err)
		return
	}
	p.sendMessage(jsonData, taskRuntimeData.StartTime, "parallel_pre_on_start", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //并发任务开始
}

func (p *KafkaRuntimeEventHandle) ParallelPostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64) {
	p.handle.ParallelPostOnEnd(behaviorTree, taskRuntimeData, stackRuntimeData, task, nowtimestampInMilli)
	data := map[string]interface{}{
		"task_id":    taskRuntimeData.TaskID,
		"execute_id": taskRuntimeData.ExecuteID,
		"stack_id":   stackRuntimeData.StackID,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		p.log.Errorf("序列化数据失败: %v", err)
		return
	}

	p.sendMessage(jsonData, nowtimestampInMilli, "parallel_post_on_end", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //并发任务结束
}

func (p *KafkaRuntimeEventHandle) ParallelAddChildStack(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, childStackRuntimeData *iface.StackRuntimeData) {
	p.handle.ParallelAddChildStack(behaviorTree, taskRuntimeData, stackRuntimeData, task, childStackRuntimeData)
	data := map[string]interface{}{
		"task_id":        taskRuntimeData.TaskID,
		"execute_id":     taskRuntimeData.ExecuteID,
		"stack_id":       stackRuntimeData.StackID,
		"child_stack_id": childStackRuntimeData.StackID,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		p.log.Errorf("序列化数据失败: %v", err)
		return
	}
	p.sendMessage(jsonData, behaviorTree.Clock().TimesampInMill(), "parallel_add_child_stack", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //并发任务增加子栈
}

func (p *KafkaRuntimeEventHandle) ParallelRemoveChildStack(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, childStackRuntimeData *iface.StackRuntimeData, nowtimestampInMilli int64) {
	p.handle.ParallelRemoveChildStack(behaviorTree, taskRuntimeData, stackRuntimeData, task, childStackRuntimeData, nowtimestampInMilli)
	data := map[string]interface{}{
		"task_id":        taskRuntimeData.TaskID,
		"execute_id":     taskRuntimeData.ExecuteID,
		"stack_id":       stackRuntimeData.StackID,
		"child_stack_id": childStackRuntimeData.StackID,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		p.log.Errorf("序列化数据失败: %v", err)
		return
	}
	p.sendMessage(jsonData, nowtimestampInMilli, "parallel_remove_child_stack", behaviorTree.ID(), behaviorTree.Name(), behaviorTree.Unit().Name(), behaviorTree.Unit().ID()) //并发任务删除子栈
}
