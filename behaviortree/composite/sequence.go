package composite

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type Sequence struct {
	base.Composite
	currentChildIndex int
	executionStatus   iface.TaskStatus
}

func (p *Sequence) OnAwake() {
	p.executionStatus = iface.Inactive
	p.currentChildIndex = 0
}

func (p *Sequence) OnStart() {}

func (p *Sequence) CanRunParallelChildren() bool { return false }
func (p *Sequence) OnChildExecuted1(childStatus iface.TaskStatus) {
	p.currentChildIndex++
	p.executionStatus = childStatus
}
func (p *Sequence) OnChildStarted0() { p.Composite.OnChildStarted0() }

func (p *Sequence) CurrentChildIndex() int {
	return p.currentChildIndex
}

func (p *Sequence) CanExecute() bool {
	return p.currentChildIndex < len(p.Children()) && p.executionStatus != iface.Failure
}

func (p *Sequence) OnConditionalAbort(childIndex int) {
	// Set the current child index to the index that caused the abort
	p.currentChildIndex = childIndex
	p.executionStatus = iface.Inactive
}

func (p *Sequence) OnCancelConditionalAbort() {
	p.executionStatus = iface.Inactive
	p.currentChildIndex = 0
	p.Unit().Log().Trace("=====Sequence	OnCancelConditionalAbort=====")
}

func (p *Sequence) OnEnd() {
	p.executionStatus = iface.Inactive
	p.currentChildIndex = 0
	p.Unit().Log().Trace("=====Sequence	OnEnd=====")
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.Sequence", func() iface.ITask { return &Sequence{} })
}
