package iface

type IRebuildSyncDataCollector interface {
	//	执行栈
	Stack(behaviorTree IBehaviorTree, data *StackRuntimeData)

	//	需要同步的action的回调
	Action(behaviorTree IBehaviorTree, taskRuntimeData *TaskRuntimeData, stackRuntimeData *StackRuntimeData, task ITask, datas [][]byte)

	//	并发任务相关的执行栈恢复同步数据
	Parallel(behaviorTree IBehaviorTree, taskRuntimeData *TaskRuntimeData, stackRuntimeData *StackRuntimeData, task ITask, childStackRuntimeDatas []*StackRuntimeData)
}
