package composite

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type ParallelSelector struct {
	base.Composite
	currentChildIndex int
	executionStatus   []iface.TaskStatus
}

func (p *ParallelSelector) OnAwake() {
	p.executionStatus = make([]iface.TaskStatus, len(p.Children()))
	p.currentChildIndex = 0
	for i := 0; i < len(p.executionStatus); i++ {
		p.executionStatus[i] = iface.Inactive
	}
}
func (p *ParallelSelector) OnStart() {
}

func (p *ParallelSelector) CanRunParallelChildren() bool {
	return true
}
func (p *ParallelSelector) OnChildExecuted2(childIndex int, childStatus iface.TaskStatus) {
	p.executionStatus[childIndex] = childStatus
}
func (p *ParallelSelector) OnChildStarted1(childIndex int) {
	p.currentChildIndex++
	p.executionStatus[childIndex] = iface.Running
}

func (p *ParallelSelector) CurrentChildIndex() int {
	return p.currentChildIndex
}
func (p *ParallelSelector) CanExecute() bool {
	return p.currentChildIndex < len(p.Children())
}
func (p *ParallelSelector) OverrideStatus1(status iface.TaskStatus) iface.TaskStatus {
	childrenComplete := true
	for i := 0; i < len(p.executionStatus); i++ {
		if p.executionStatus[i] == iface.Running {
			childrenComplete = false
		} else if p.executionStatus[i] == iface.Success {
			return iface.Success
		}
	}

	if childrenComplete {
		return iface.Failure
	} else {
		return iface.Running
	}
}

func (p *ParallelSelector) OnConditionalAbort(childIndex int) {
	p.Unit().Log().Trace("===============ParallelSelector OnConditionalAbort:", childIndex)

	p.currentChildIndex = 0
	for i := 0; i < len(p.executionStatus); i++ {
		p.executionStatus[i] = iface.Inactive
	}
}

func (p *ParallelSelector) OnCancelConditionalAbort() {
	p.Unit().Log().Trace("=====ParallelSelector	OnCancelConditionalAbort=====")
}

func (p *ParallelSelector) OnEnd() {
	p.Unit().Log().Trace("=====ParallelSelector	OnEnd=====")
	p.currentChildIndex = 0
	for i := 0; i < len(p.executionStatus); i++ {
		p.executionStatus[i] = iface.Inactive
	}
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.ParallelSelector", func() iface.ITask { return &ParallelSelector{} })
}
