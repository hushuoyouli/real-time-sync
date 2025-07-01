package runtime

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

type EmptyRuntimeEventHandle struct {
}

func NewEmptyRuntimeEventHandle() *EmptyRuntimeEventHandle {
	return &EmptyRuntimeEventHandle{}
}

func (p *EmptyRuntimeEventHandle) PostInitialize(behaviorTree iface.IBehaviorTree, nowtimestampInMilli int64) {
}

func (eventHandle *EmptyRuntimeEventHandle) PostOnComplete(behaviorTree iface.IBehaviorTree, nowtimestampInMilli int64) {

}

func (p *EmptyRuntimeEventHandle) NewStack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData) {
}

func (p *EmptyRuntimeEventHandle) RemoveStack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData, nowtimestampInMilli int64) {
}

func (p *EmptyRuntimeEventHandle) PreOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask) {

}

func (p *EmptyRuntimeEventHandle) PostOnUpdate(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, status iface.TaskStatus) {

}

func (p *EmptyRuntimeEventHandle) PostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64) {

}

func (p *EmptyRuntimeEventHandle) ActionPostOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, datas [][]byte) {

}

func (p *EmptyRuntimeEventHandle) ActionPostOnUpdate(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, status iface.TaskStatus, datas [][]byte) {

}

func (p *EmptyRuntimeEventHandle) ActionPostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, datas [][]byte) {

}

func (p *EmptyRuntimeEventHandle) ParallelPreOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask) {

}

func (p *EmptyRuntimeEventHandle) ParallelPostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64) {

}

func (p *EmptyRuntimeEventHandle) ParallelAddChildStack(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, childStackRuntimeData *iface.StackRuntimeData) {

}

func (p *EmptyRuntimeEventHandle) ParallelRemoveChildStack(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, childStackRuntimeData *iface.StackRuntimeData, nowtimestampInMilli int64) {

}
