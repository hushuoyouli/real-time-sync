package runtime

import (
	"testing"
	"time"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/rlog"
	"github.com/segmentio/kafka-go"
)

// MockBehaviorTree 用于测试
type MockBehaviorTree struct {
	id        int64
	unit      iface.IUnit
	clock     iface.IClock
	isRunning bool
}

func (m *MockBehaviorTree) Config() []byte {
	return nil
}

func (m *MockBehaviorTree) Name() string {
	return ""
}

func (m *MockBehaviorTree) ID() int64 {
	return m.id
}

func (m *MockBehaviorTree) Enable() error {
	m.isRunning = true
	return nil
}

func (m *MockBehaviorTree) Disable() error {
	m.isRunning = false
	return nil
}

func (m *MockBehaviorTree) Update() {
}

func (m *MockBehaviorTree) IsRunning() bool {
	return m.isRunning
}

func (m *MockBehaviorTree) Unit() iface.IUnit {
	return m.unit
}

func (m *MockBehaviorTree) RebuildSync(collector iface.IRebuildSyncDataCollector) {
}

func (m *MockBehaviorTree) Clock() iface.IClock {
	return m.clock
}

func (m *MockBehaviorTree) ExtraParam() interface{} {
	return nil
}

// MockTask 用于测试
type MockTask struct {
	correspondingType string
	owner             iface.IBehaviorTree
	parent            iface.IParentTask
	id                int
	name              string
	isInstant         bool
	disabled          bool
	unit              iface.IUnit
}

func (m *MockTask) CorrespondingType() string {
	return m.correspondingType
}

func (m *MockTask) SetCorrespondingType(correspondingType string) {
	m.correspondingType = correspondingType
}

func (m *MockTask) Owner() iface.IBehaviorTree {
	return m.owner
}

func (m *MockTask) SetOwner(owner iface.IBehaviorTree) {
	m.owner = owner
}

func (m *MockTask) Parent() iface.IParentTask {
	return m.parent
}

func (m *MockTask) SetParent(parent iface.IParentTask) {
	m.parent = parent
}

func (m *MockTask) ID() int {
	return m.id
}

func (m *MockTask) SetID(id int) {
	m.id = id
}

func (m *MockTask) Name() string {
	return m.name
}

func (m *MockTask) SetName(name string) {
	m.name = name
}

func (m *MockTask) IsInstant() bool {
	return m.isInstant
}

func (m *MockTask) SetIsInstant(isInstant bool) {
	m.isInstant = isInstant
}

func (m *MockTask) Disabled() bool {
	return m.disabled
}

func (m *MockTask) SetDisabled(disabled bool) {
	m.disabled = disabled
}

func (m *MockTask) Unit() iface.IUnit {
	return m.unit
}

func (m *MockTask) SetUnit(unit iface.IUnit) {
	m.unit = unit
}

func (m *MockTask) OnAwake() {
}

func (m *MockTask) OnStart() {
}

func (m *MockTask) OnUpdate() iface.TaskStatus {
	return iface.Success
}

func (m *MockTask) OnEnd() {
}

func (m *MockTask) OnComplete() {
}

func (m *MockTask) DebugInfo() map[string]interface{} {
	return make(map[string]interface{})
}

func (m *MockTask) IsImplementsIAction() bool {
	return false
}

func (m *MockTask) IsImplementsIComposite() bool {
	return false
}

func (m *MockTask) IsImplementsIDecorator() bool {
	return false
}

func (m *MockTask) IsImplementsIConditional() bool {
	return false
}

func (m *MockTask) IsImplementsIParentTask() bool {
	return false
}

func (m *MockTask) SetVariables(variableConfigs map[string]interface{}) error {
	return nil
}

// MockRuntimeEventHandle 用于测试委托调用
type MockRuntimeEventHandle struct {
	PostInitializeCalled           bool
	PostOnCompleteCalled           bool
	NewStackCalled                 bool
	RemoveStackCalled              bool
	PreOnStartCalled               bool
	PostOnUpdateCalled             bool
	PostOnEndCalled                bool
	ActionPostOnStartCalled        bool
	ActionPostOnUpdateCalled       bool
	ActionPostOnEndCalled          bool
	ParallelPreOnStartCalled       bool
	ParallelPostOnEndCalled        bool
	ParallelAddChildStackCalled    bool
	ParallelRemoveChildStackCalled bool
	CloseCalled                    bool
}

func (m *MockRuntimeEventHandle) Close() {
	m.CloseCalled = true
}

func (m *MockRuntimeEventHandle) PostInitialize(behaviorTree iface.IBehaviorTree, nowtimestampInMilli int64) {
	m.PostInitializeCalled = true
}

func (m *MockRuntimeEventHandle) PostOnComplete(behaviorTree iface.IBehaviorTree, nowtimestampInMilli int64) {
	m.PostOnCompleteCalled = true
}

func (m *MockRuntimeEventHandle) NewStack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData) {
	m.NewStackCalled = true
}

func (m *MockRuntimeEventHandle) RemoveStack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData, nowtimestampInMilli int64) {
	m.RemoveStackCalled = true
}

func (m *MockRuntimeEventHandle) PreOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask) {
	m.PreOnStartCalled = true
}

func (m *MockRuntimeEventHandle) PostOnUpdate(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, status iface.TaskStatus) {
	m.PostOnUpdateCalled = true
}

func (m *MockRuntimeEventHandle) PostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, status iface.TaskStatus) {
	m.PostOnEndCalled = true
}

func (m *MockRuntimeEventHandle) ActionPostOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, datas [][]byte) {
	m.ActionPostOnStartCalled = true
}

func (m *MockRuntimeEventHandle) ActionPostOnUpdate(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, status iface.TaskStatus, datas [][]byte) {
	m.ActionPostOnUpdateCalled = true
}

func (m *MockRuntimeEventHandle) ActionPostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, datas [][]byte) {
	m.ActionPostOnEndCalled = true
}

func (m *MockRuntimeEventHandle) ParallelPreOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask) {
	m.ParallelPreOnStartCalled = true
}

func (m *MockRuntimeEventHandle) ParallelPostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64) {
	m.ParallelPostOnEndCalled = true
}

func (m *MockRuntimeEventHandle) ParallelAddChildStack(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, childStackRuntimeData *iface.StackRuntimeData) {
	m.ParallelAddChildStackCalled = true
}

func (m *MockRuntimeEventHandle) ParallelRemoveChildStack(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, childStackRuntimeData *iface.StackRuntimeData, nowtimestampInMilli int64) {
	m.ParallelRemoveChildStackCalled = true
}

// TestNewKafkaRuntimeEventHandle_Success 测试成功创建 KafkaRuntimeEventHandle
func TestNewKafkaRuntimeEventHandle_Success(t *testing.T) {
	broker := "localhost:9092"

	// 检查 Kafka 是否可用
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		t.Skipf("跳过测试：无法连接到 Kafka: %v", err)
	}
	conn.Close()

	mockHandle := &MockRuntimeEventHandle{}
	logger := &rlog.SLogger{}

	handle, err := NewKafkaRuntimeEventHandle(mockHandle, "test-topic", broker, logger)
	if err != nil {
		t.Fatalf("创建 KafkaRuntimeEventHandle 失败: %v", err)
	}

	defer handle.Close()

	if handle == nil {
		t.Fatal("返回的 handle 不应该为 nil")
	}

	if handle.handle != mockHandle {
		t.Error("内部 handle 应该指向传入的 mockHandle")
	}

	if handle.broker != broker {
		t.Errorf("broker 不匹配: 期望 %s, 实际 %s", broker, handle.broker)
	}

	if handle.topicName == "" {
		t.Error("topicName 不应该为空")
	}

	// 验证 topicName 包含时间戳格式
	if len(handle.topicName) <= len("test-topic") {
		t.Error("topicName 应该包含时间戳后缀")
	}
}

// TestNewKafkaRuntimeEventHandle_InvalidBroker 测试无效的 broker 地址
func TestNewKafkaRuntimeEventHandle_InvalidBroker(t *testing.T) {
	mockHandle := &MockRuntimeEventHandle{}
	logger := &rlog.SLogger{}

	handle, err := NewKafkaRuntimeEventHandle(mockHandle, "test-topic", "invalid-broker:9999", logger)
	if err == nil {
		if handle != nil {
			handle.Close()
		}
		t.Fatal("应该返回错误，但 err 为 nil")
	}
}

// TestKafkaRuntimeEventHandle_Close 测试 Close 方法
func TestKafkaRuntimeEventHandle_Close(t *testing.T) {
	broker := "localhost:9092"

	// 检查 Kafka 是否可用
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		t.Skipf("跳过测试：无法连接到 Kafka: %v", err)
	}
	conn.Close()

	mockHandle := &MockRuntimeEventHandle{}
	logger := &rlog.SLogger{}

	handle, err := NewKafkaRuntimeEventHandle(mockHandle, "test-topic", broker, logger)
	if err != nil {
		t.Fatalf("创建 KafkaRuntimeEventHandle 失败: %v", err)
	}

	// 关闭应该不报错
	handle.Close()

	// 再次关闭应该也不报错（幂等性）
	handle.Close()

	if !mockHandle.CloseCalled {
		t.Error("应该调用内部 handle 的 Close 方法")
	}
}

// TestKafkaRuntimeEventHandle_DelegationMethods 测试所有委托方法
func TestKafkaRuntimeEventHandle_DelegationMethods(t *testing.T) {
	broker := "localhost:9092"

	// 检查 Kafka 是否可用
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		t.Skipf("跳过测试：无法连接到 Kafka: %v", err)
	}
	conn.Close()

	mockHandle := &MockRuntimeEventHandle{}
	logger := &rlog.SLogger{}

	handle, err := NewKafkaRuntimeEventHandle(mockHandle, "test-topic", broker, logger)
	if err != nil {
		t.Fatalf("创建 KafkaRuntimeEventHandle 失败: %v", err)
	}
	defer func() {
		if handle != nil {
			handle.Close()
		}
	}()

	// 创建测试用的 mock 对象
	mockUnit := NewTestUnit("test-unit")
	mockClock := NewClock()
	mockBT := &MockBehaviorTree{
		id:    1,
		unit:  mockUnit,
		clock: mockClock,
	}
	mockTaskRuntimeData := &iface.TaskRuntimeData{}
	mockStackRuntimeData := &iface.StackRuntimeData{}
	mockTask := &MockTask{
		correspondingType: "test-task",
		name:              "test-task",
		unit:              mockUnit,
	}

	// 测试 PostInitialize
	handle.PostInitialize(mockBT, time.Now().UnixMilli())
	if !mockHandle.PostInitializeCalled {
		t.Error("PostInitialize 应该委托给内部 handle")
	}

	// 测试 PostOnComplete
	handle.PostOnComplete(mockBT, time.Now().UnixMilli())
	if !mockHandle.PostOnCompleteCalled {
		t.Error("PostOnComplete 应该委托给内部 handle")
	}

	// 测试 NewStack
	handle.NewStack(mockBT, mockStackRuntimeData)
	if !mockHandle.NewStackCalled {
		t.Error("NewStack 应该委托给内部 handle")
	}

	// 测试 RemoveStack
	handle.RemoveStack(mockBT, mockStackRuntimeData, time.Now().UnixMilli())
	if !mockHandle.RemoveStackCalled {
		t.Error("RemoveStack 应该委托给内部 handle")
	}

	// 测试 PreOnStart
	handle.PreOnStart(mockBT, mockTaskRuntimeData, mockStackRuntimeData, mockTask)
	if !mockHandle.PreOnStartCalled {
		t.Error("PreOnStart 应该委托给内部 handle")
	}

	// 测试 PostOnUpdate
	handle.PostOnUpdate(mockBT, mockTaskRuntimeData, mockStackRuntimeData, mockTask, time.Now().UnixMilli(), iface.Success)
	if !mockHandle.PostOnUpdateCalled {
		t.Error("PostOnUpdate 应该委托给内部 handle")
	}

	// 测试 PostOnEnd
	handle.PostOnEnd(mockBT, mockTaskRuntimeData, mockStackRuntimeData, mockTask, time.Now().UnixMilli(), iface.Success)
	if !mockHandle.PostOnEndCalled {
		t.Error("PostOnEnd 应该委托给内部 handle")
	}

	// 测试 ActionPostOnStart
	handle.ActionPostOnStart(mockBT, mockTaskRuntimeData, mockStackRuntimeData, mockTask, [][]byte{[]byte("test")})
	if !mockHandle.ActionPostOnStartCalled {
		t.Error("ActionPostOnStart 应该委托给内部 handle")
	}

	// 测试 ActionPostOnUpdate
	handle.ActionPostOnUpdate(mockBT, mockTaskRuntimeData, mockStackRuntimeData, mockTask, time.Now().UnixMilli(), iface.Success, [][]byte{[]byte("test")})
	if !mockHandle.ActionPostOnUpdateCalled {
		t.Error("ActionPostOnUpdate 应该委托给内部 handle")
	}

	// 测试 ActionPostOnEnd
	handle.ActionPostOnEnd(mockBT, mockTaskRuntimeData, mockStackRuntimeData, mockTask, time.Now().UnixMilli(), [][]byte{[]byte("test")})
	if !mockHandle.ActionPostOnEndCalled {
		t.Error("ActionPostOnEnd 应该委托给内部 handle")
	}

	// 测试 ParallelPreOnStart
	handle.ParallelPreOnStart(mockBT, mockTaskRuntimeData, mockStackRuntimeData, mockTask)
	if !mockHandle.ParallelPreOnStartCalled {
		t.Error("ParallelPreOnStart 应该委托给内部 handle")
	}

	// 测试 ParallelPostOnEnd
	handle.ParallelPostOnEnd(mockBT, mockTaskRuntimeData, mockStackRuntimeData, mockTask, time.Now().UnixMilli())
	if !mockHandle.ParallelPostOnEndCalled {
		t.Error("ParallelPostOnEnd 应该委托给内部 handle")
	}

	// 测试 ParallelAddChildStack
	handle.ParallelAddChildStack(mockBT, mockTaskRuntimeData, mockStackRuntimeData, mockTask, mockStackRuntimeData)
	if !mockHandle.ParallelAddChildStackCalled {
		t.Error("ParallelAddChildStack 应该委托给内部 handle")
	}

	// 测试 ParallelRemoveChildStack
	handle.ParallelRemoveChildStack(mockBT, mockTaskRuntimeData, mockStackRuntimeData, mockTask, mockStackRuntimeData, time.Now().UnixMilli())
	if !mockHandle.ParallelRemoveChildStackCalled {
		t.Error("ParallelRemoveChildStack 应该委托给内部 handle")
	}
}

// TestKafkaRuntimeEventHandle_TopicNameFormat 测试 topic 名称格式
func TestKafkaRuntimeEventHandle_TopicNameFormat(t *testing.T) {
	broker := "localhost:9092"

	// 检查 Kafka 是否可用
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		t.Skipf("跳过测试：无法连接到 Kafka: %v", err)
	}
	conn.Close()

	mockHandle := &MockRuntimeEventHandle{}
	logger := &rlog.SLogger{}

	baseTopicName := "test-scenario"
	handle1, err := NewKafkaRuntimeEventHandle(mockHandle, baseTopicName, broker, logger)
	if err != nil {
		t.Fatalf("创建 KafkaRuntimeEventHandle 失败: %v", err)
	}
	defer handle1.Close()

	// 等待一小段时间确保时间戳不同
	time.Sleep(100 * time.Millisecond)

	handle2, err := NewKafkaRuntimeEventHandle(mockHandle, baseTopicName, broker, logger)
	if err != nil {
		t.Fatalf("创建第二个 KafkaRuntimeEventHandle 失败: %v", err)
	}
	defer handle2.Close()

	// 验证 topic 名称包含基础名称和时间戳
	if handle1.topicName == handle2.topicName {
		t.Error("不同时间创建的 handle 应该有不同的 topic 名称")
	}

	// 验证 topic 名称格式：应该包含基础名称和时间戳
	if len(handle1.topicName) <= len(baseTopicName) {
		t.Errorf("topic 名称应该包含时间戳后缀，但得到: %s", handle1.topicName)
	}

	// 验证 topic 名称以基础名称开头
	if handle1.topicName[:len(baseTopicName)] != baseTopicName {
		t.Errorf("topic 名称应该以基础名称开头，但得到: %s", handle1.topicName)
	}
}

// TestNewInitializeTopicMsg 测试初始化消息的创建
/* func TestNewInitializeTopicMsg(t *testing.T) {
	msg := newInitializeTopicMsg()

	if string(msg.Key) != "initialize_topic" {
		t.Errorf("消息 Key 应该是 'initialize_topic'，但得到: %s", string(msg.Key))
	}

	if string(msg.Value) != "触发自动创建 topic" {
		t.Errorf("消息 Value 应该是 '触发自动创建 topic'，但得到: %s", string(msg.Value))
	}
} */

// TestKafkaRuntimeEventHandle_SendMessage 测试 sendMessage 方法
func TestKafkaRuntimeEventHandle_SendMessage(t *testing.T) {
	broker := "localhost:9092"

	// 检查 Kafka 是否可用
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		t.Skipf("跳过测试：无法连接到 Kafka: %v", err)
	}
	conn.Close()

	mockHandle := &MockRuntimeEventHandle{}
	logger := &rlog.SLogger{}

	handle, err := NewKafkaRuntimeEventHandle(mockHandle, "test-send-message", broker, logger)
	if err != nil {
		t.Fatalf("创建 KafkaRuntimeEventHandle 失败: %v", err)
	}
	defer handle.Close()

	// 创建测试消息
	testMsg := kafka.Message{
		Key:   []byte("test-key"),
		Value: []byte("test message value"),
		Headers: []kafka.Header{
			{Key: "header1", Value: []byte("value1")},
		},
	}

	// 调用 sendMessage（由于是私有方法，在同一个包中可以访问）
	handle.sendMessage(testMsg.Value, time.Now().UnixMilli(), "test", 0, "", "", 0)

	// 从通道中读取消息并验证
	select {
	case receivedMsg := <-handle.kafkaMessageChannel:
		// 验证消息的 Key
		if string(receivedMsg.Key) != string(testMsg.Key) {
			t.Errorf("消息 Key 不匹配: 期望 %s, 实际 %s", string(testMsg.Key), string(receivedMsg.Key))
		}

		// 验证消息的 Value
		if string(receivedMsg.Value) != string(testMsg.Value) {
			t.Errorf("消息 Value 不匹配: 期望 %s, 实际 %s", string(testMsg.Value), string(receivedMsg.Value))
		}

		// 验证消息的 Headers
		if len(receivedMsg.Headers) != len(testMsg.Headers) {
			t.Errorf("消息 Headers 数量不匹配: 期望 %d, 实际 %d", len(testMsg.Headers), len(receivedMsg.Headers))
		} else if len(receivedMsg.Headers) > 0 {
			if receivedMsg.Headers[0].Key != testMsg.Headers[0].Key {
				t.Errorf("消息 Header Key 不匹配: 期望 %s, 实际 %s", testMsg.Headers[0].Key, receivedMsg.Headers[0].Key)
			}
			if string(receivedMsg.Headers[0].Value) != string(testMsg.Headers[0].Value) {
				t.Errorf("消息 Header Value 不匹配: 期望 %s, 实际 %s", string(testMsg.Headers[0].Value), string(receivedMsg.Headers[0].Value))
			}
		}
	case <-time.After(2 * time.Second):
		t.Error("超时：未能从通道中接收到消息")
	}
}

// TestKafkaRuntimeEventHandle_SendMessage_Multiple 测试 sendMessage 发送多条消息
func TestKafkaRuntimeEventHandle_SendMessage_Multiple(t *testing.T) {
	broker := "localhost:9092"

	// 检查 Kafka 是否可用
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		t.Skipf("跳过测试：无法连接到 Kafka: %v", err)
	}
	conn.Close()

	mockHandle := &MockRuntimeEventHandle{}
	logger := &rlog.SLogger{}

	handle, err := NewKafkaRuntimeEventHandle(mockHandle, "test-send-multiple", broker, logger)
	if err != nil {
		t.Fatalf("创建 KafkaRuntimeEventHandle 失败: %v", err)
	}
	defer handle.Close()

	// 发送多条消息
	messageCount := 9999
	for i := 0; i < messageCount; i++ {
		/* 		testMsg := kafka.Message{
			Key:   []byte("test-key"),
			Value: []byte("test message " + string(rune(i+'0'))),
		} */
		handle.sendMessage([]byte("test message "+string(rune(i+'0'))), time.Now().UnixMilli(), "test", 0, "", "", 0)
	}

}
