package composite

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type Parallel struct {
	base.Composite
	currentChildIndex int
	executionStatus   []iface.TaskStatus
}

func (p *Parallel) OnAwake() {
	p.executionStatus = make([]iface.TaskStatus, len(p.Children()))
	p.currentChildIndex = 0
	for i := 0; i < len(p.executionStatus); i++ {
		p.executionStatus[i] = iface.Inactive
	}
}
func (p *Parallel) OnStart() {
}

func (p *Parallel) CanRunParallelChildren() bool {
	return true
}
func (p *Parallel) OnChildExecuted2(childIndex int, childStatus iface.TaskStatus) {
	p.executionStatus[childIndex] = childStatus
}
func (p *Parallel) OnChildStarted1(childIndex int) {
	p.currentChildIndex++
	p.executionStatus[childIndex] = iface.Running
}

func (p *Parallel) CurrentChildIndex() int {
	return p.currentChildIndex
}
func (p *Parallel) CanExecute() bool {
	return p.currentChildIndex < len(p.Children())
}
func (p *Parallel) OverrideStatus1(status iface.TaskStatus) iface.TaskStatus {
	childrenComplete := true
	for i := 0; i < len(p.executionStatus); i++ {
		if p.executionStatus[i] == iface.Running {
			childrenComplete = false
		} else if p.executionStatus[i] == iface.Failure {
			return iface.Failure
		}
	}

	if childrenComplete {
		return iface.Success
	} else {
		return iface.Running
	}
}

func (p *Parallel) OnConditionalAbort(childIndex int) {
	p.Unit().Log().Trace("===============Parallel OnConditionalAbort:", childIndex)

	p.currentChildIndex = 0
	for i := 0; i < len(p.executionStatus); i++ {
		p.executionStatus[i] = iface.Inactive
	}
}

func (p *Parallel) OnCancelConditionalAbort() {
	p.Unit().Log().Trace("=====Parallel	OnCancelConditionalAbort=====")
}

func (p *Parallel) OnEnd() {
	p.Unit().Log().Trace("=====Parallel	OnEnd=====")
	p.currentChildIndex = 0
	for i := 0; i < len(p.executionStatus); i++ {
		p.executionStatus[i] = iface.Inactive
	}
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.Parallel", func() iface.ITask { return &Parallel{} })
}
