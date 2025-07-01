package runtime

import "github.com/hushuoyouli/real-time-sync/behaviortree/iface"

type ConditionalReevaluate struct {
	index          int
	taskStatus     iface.TaskStatus
	compositeIndex int
	//stackIndex     int
}

func newConditionalReevaluate(index int, taskStatus iface.TaskStatus, compositeIndex int) *ConditionalReevaluate {
	val := &ConditionalReevaluate{}
	val.Initialize(index, taskStatus, compositeIndex)
	return val
}

func (p *ConditionalReevaluate) Initialize(index int, taskStatus iface.TaskStatus, compositeIndex int) {
	p.index = index
	p.taskStatus = taskStatus
	p.compositeIndex = compositeIndex
}

func (p *ConditionalReevaluate) DebugInfo() []interface{} {
	return []interface{}{p.index, p.taskStatus.ToString(), p.compositeIndex}
}
