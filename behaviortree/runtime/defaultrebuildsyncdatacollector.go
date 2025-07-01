package runtime

import (
	"fmt"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

type DefaultRebuildSyncDataCollector struct {
	EmptyRebuildSyncDataCollector
}

func newDefaultRebuildSyncDataCollector() *DefaultRebuildSyncDataCollector {
	return &DefaultRebuildSyncDataCollector{}
}

func (p *DefaultRebuildSyncDataCollector) Stack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData) {
	fmt.Printf("恢复数据-执行栈	角色:%d 行为树:%d 堆栈:%d 生成时间:%d\n", behaviorTree.Unit().ID(), behaviorTree.ID(), data.StackID, data.StartTime)
}

func (p *DefaultRebuildSyncDataCollector) Action(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, datas [][]byte) {
	fmt.Printf("恢复数据-行为节点	角色:%d 行为树:%d 堆栈:%d 任务:%d-%d 生成时间:%d	恢复数据:%v\n", behaviorTree.Unit().ID(), behaviorTree.ID(), stackRuntimeData.StackID, taskRuntimeData.TaskID, taskRuntimeData.ExecuteID, taskRuntimeData.StartTime, datas)
}

func (p *DefaultRebuildSyncDataCollector) Parallel(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, childStackRuntimeDatas []*iface.StackRuntimeData) {
	childStackIds := make([]int, 0)
	for _, childStackRuntimeData := range childStackRuntimeDatas {
		childStackIds = append(childStackIds, childStackRuntimeData.StackID)
	}

	fmt.Printf("恢复数据-并发节点	角色:%d 行为树:%d 堆栈:%d 任务:%d-%d 生成时间:%d	子执行栈:%v\n", behaviorTree.Unit().ID(), behaviorTree.ID(), stackRuntimeData.StackID, taskRuntimeData.TaskID, taskRuntimeData.ExecuteID, taskRuntimeData.StartTime, childStackIds)
}
