package runtime

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

type DefaultRuntimeEventHandle struct {
	EmptyRuntimeEventHandle
}

func NewDefaultRuntimeEventHandle() *DefaultRuntimeEventHandle {
	return &DefaultRuntimeEventHandle{}
}

func (p *DefaultRuntimeEventHandle) PostInitialize(behaviorTree iface.IBehaviorTree, nowtimestampInMilli int64) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d =>初始化完成\n", nowtimestampInMilli, behaviorTree.Unit().ID(), behaviorTree.ID())
}

func (p *DefaultRuntimeEventHandle) PostOnComplete(behaviorTree iface.IBehaviorTree, nowtimestampInMilli int64) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d =>结束执行\n", nowtimestampInMilli, behaviorTree.Unit().ID(), behaviorTree.ID())
}

func (p *DefaultRuntimeEventHandle) NewStack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d =>增加执行栈%d,时间戳:%d\n", behaviorTree.Clock().TimesampInMill(), behaviorTree.Unit().ID(), behaviorTree.ID(), data.StackID, data.StartTime)
}

func (p *DefaultRuntimeEventHandle) RemoveStack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData, nowtimestampInMilli int64) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d =>删除执行栈%d,时间戳:%d\n", behaviorTree.Clock().TimesampInMill(), behaviorTree.Unit().ID(), behaviorTree.ID(), data.StackID, nowtimestampInMilli)
}

func (p *DefaultRuntimeEventHandle) PreOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d =>在时间:%d任务%s:%d-%d进入堆栈:%d\n", behaviorTree.Clock().TimesampInMill(), behaviorTree.Unit().ID(), behaviorTree.ID(), taskRuntimeData.StartTime, task.CorrespondingType(), task.ID(), taskRuntimeData.ExecuteID, stackRuntimeData.StackID)
}

func (p *DefaultRuntimeEventHandle) PostOnUpdate(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, status iface.TaskStatus) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d => 在时间:%d任务:%s:%d-%d在堆栈:%d,执行结果:%s\n", behaviorTree.Clock().TimesampInMill(), behaviorTree.Unit().ID(), behaviorTree.ID(), nowtimestampInMilli, task.CorrespondingType(), task.ID(), taskRuntimeData.ExecuteID, stackRuntimeData.StackID, status.ToString())
}

func (p *DefaultRuntimeEventHandle) PostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d => 在时间:%d任务:%s:%d-%d离开堆栈:%d\n", behaviorTree.Clock().TimesampInMill(), behaviorTree.Unit().ID(), behaviorTree.ID(), nowtimestampInMilli, task.CorrespondingType(), task.ID(), taskRuntimeData.ExecuteID, stackRuntimeData.StackID)
}

func (p *DefaultRuntimeEventHandle) ActionPostOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, datas [][]byte) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d => 在时间:%d	同步任务%s:%d-%d	进入堆栈:%d	同步数据:%v\n", behaviorTree.Clock().TimesampInMill(), behaviorTree.Unit().ID(), behaviorTree.ID(), taskRuntimeData.StartTime, task.CorrespondingType(), task.ID(), taskRuntimeData.ExecuteID, stackRuntimeData.StackID, datas)
}

func (p *DefaultRuntimeEventHandle) ActionPostOnUpdate(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, status iface.TaskStatus, datas [][]byte) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d => 在时间:%d	同步任务:%s:%d-%d在堆栈:%d,执行结果:%s	同步数据:%v\n", behaviorTree.Clock().TimesampInMill(), behaviorTree.Unit().ID(), behaviorTree.ID(), nowtimestampInMilli, task.CorrespondingType(), task.ID(), taskRuntimeData.ExecuteID, stackRuntimeData.StackID, status.ToString(), datas)
}

func (p *DefaultRuntimeEventHandle) ActionPostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64, datas [][]byte) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d => 在时间:%d	同步任务:%s:%d-%d离开堆栈:%d	同步数据:%v\n", behaviorTree.Clock().TimesampInMill(), behaviorTree.Unit().ID(), behaviorTree.ID(), nowtimestampInMilli, task.CorrespondingType(), task.ID(), taskRuntimeData.ExecuteID, stackRuntimeData.StackID, datas)
}

func (p *DefaultRuntimeEventHandle) ParallelPreOnStart(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d =>在时间:%d	并发任务%s:%d-%d 	进入堆栈:%d\n", behaviorTree.Clock().TimesampInMill(), behaviorTree.Unit().ID(), behaviorTree.ID(), taskRuntimeData.StartTime, task.CorrespondingType(), task.ID(), taskRuntimeData.ExecuteID, stackRuntimeData.StackID)
}

func (p *DefaultRuntimeEventHandle) ParallelPostOnEnd(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, nowtimestampInMilli int64) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d => 在时间:%d	并发任务:%s:%d-%d	离开堆栈:%d\n", behaviorTree.Clock().TimesampInMill(), behaviorTree.Unit().ID(), behaviorTree.ID(), nowtimestampInMilli, task.CorrespondingType(), task.ID(), taskRuntimeData.ExecuteID, stackRuntimeData.StackID)
}

func (p *DefaultRuntimeEventHandle) ParallelAddChildStack(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, childStackRuntimeData *iface.StackRuntimeData) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d => 在时间:%d任务:%s:%d-%d在堆栈:%d,生成一个子执行栈:%d\n", behaviorTree.Clock().TimesampInMill(), behaviorTree.Unit().ID(), behaviorTree.ID(), childStackRuntimeData.StartTime, task.CorrespondingType(), task.ID(), taskRuntimeData.ExecuteID, stackRuntimeData.StackID, childStackRuntimeData.StackID)
}

func (p *DefaultRuntimeEventHandle) ParallelRemoveChildStack(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, childStackRuntimeData *iface.StackRuntimeData, nowtimestampInMilli int64) {
	behaviorTree.Unit().Log().Tracef("在时间:%d 角色:%d 行为树:%d => 在时间:%d任务:%s:%d-%d在堆栈:%d,弹出子执行栈:%d\n", behaviorTree.Clock().TimesampInMill(), behaviorTree.Unit().ID(), behaviorTree.ID(), nowtimestampInMilli, task.CorrespondingType(), task.ID(), taskRuntimeData.ExecuteID, stackRuntimeData.StackID, childStackRuntimeData.StackID)
}
