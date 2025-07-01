package util

import "github.com/hushuoyouli/real-time-sync/behaviortree/iface"

type TaskAddData struct {
	Parent               iface.IParentTask
	ParentIndex          int
	Depth                int
	CompositeParentIndex int
	Owner                iface.IBehaviorTree
	Unit                 iface.IUnit
	ErrorTask            int
	ErrrorTaskName       string
}

func NewTaskAddData(Owner iface.IBehaviorTree, Unit iface.IUnit) *TaskAddData {
	return &TaskAddData{
		Parent:               nil,
		ParentIndex:          -1,
		Depth:                0,
		CompositeParentIndex: 0,
		Owner:                Owner,
		Unit:                 Unit,
		ErrorTask:            -1,
		ErrrorTaskName:       "",
	}
}
