package iface

type IRuntimeEventHandle interface {
	PostInitialize(behaviorTree IBehaviorTree, nowtimestampInMilli int64)
	//	树结束
	PostOnComplete(behaviorTree IBehaviorTree, nowtimestampInMilli int64)

	//	同步需要
	NewStack(behaviorTree IBehaviorTree, data *StackRuntimeData)
	RemoveStack(behaviorTree IBehaviorTree, data *StackRuntimeData, nowtimestampInMilli int64)

	//	以下3个回调可以用于追踪树的执行
	PreOnStart(behaviorTree IBehaviorTree, taskRuntimeData *TaskRuntimeData, stackRuntimeData *StackRuntimeData, task ITask)
	PostOnUpdate(behaviorTree IBehaviorTree, taskRuntimeData *TaskRuntimeData, stackRuntimeData *StackRuntimeData, task ITask, nowtimestampInMilli int64, status TaskStatus) //	任何的任务每帧调用的结果
	PostOnEnd(behaviorTree IBehaviorTree, taskRuntimeData *TaskRuntimeData, stackRuntimeData *StackRuntimeData, task ITask, nowtimestampInMilli int64)

	//	需要同步的action的回调，同步需要
	ActionPostOnStart(behaviorTree IBehaviorTree, taskRuntimeData *TaskRuntimeData, stackRuntimeData *StackRuntimeData, task ITask, datas [][]byte)
	ActionPostOnUpdate(behaviorTree IBehaviorTree, taskRuntimeData *TaskRuntimeData, stackRuntimeData *StackRuntimeData, task ITask, nowtimestampInMilli int64, status TaskStatus, datas [][]byte) //	任何的任务每帧调用的结果
	ActionPostOnEnd(behaviorTree IBehaviorTree, taskRuntimeData *TaskRuntimeData, stackRuntimeData *StackRuntimeData, task ITask, nowtimestampInMilli int64, datas [][]byte)

	//	需要同步的并发任务进入调用，同步需要
	ParallelPreOnStart(behaviorTree IBehaviorTree, taskRuntimeData *TaskRuntimeData, stackRuntimeData *StackRuntimeData, task ITask)
	ParallelPostOnEnd(behaviorTree IBehaviorTree, taskRuntimeData *TaskRuntimeData, stackRuntimeData *StackRuntimeData, task ITask, nowtimestampInMilli int64)

	//	并发任务相关的执行栈的增加/减少，调用顺序是NewStack/ParallelAddChildStack/ParallelRemoveChildStack/RemoveStack
	ParallelAddChildStack(behaviorTree IBehaviorTree, taskRuntimeData *TaskRuntimeData, stackRuntimeData *StackRuntimeData, task ITask, childStackRuntimeData *StackRuntimeData)
	ParallelRemoveChildStack(behaviorTree IBehaviorTree, taskRuntimeData *TaskRuntimeData, stackRuntimeData *StackRuntimeData, task ITask, childStackRuntimeData *StackRuntimeData, nowtimestampInMilli int64)
}
