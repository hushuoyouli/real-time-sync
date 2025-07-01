package runtime

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

type EmptyRebuildSyncDataCollector struct {
}

func newEmptyRebuildSyncDataCollector() *EmptyRebuildSyncDataCollector {
	return &EmptyRebuildSyncDataCollector{}
}

func (p *EmptyRebuildSyncDataCollector) Stack(behaviorTree iface.IBehaviorTree, data *iface.StackRuntimeData) {
}

func (p *EmptyRebuildSyncDataCollector) Action(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, datas [][]byte) {
}

func (p *EmptyRebuildSyncDataCollector) Parallel(behaviorTree iface.IBehaviorTree, taskRuntimeData *iface.TaskRuntimeData, stackRuntimeData *iface.StackRuntimeData, task iface.ITask, childStackRuntimeDatas []*iface.StackRuntimeData) {
}
