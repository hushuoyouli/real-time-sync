package composite

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type Selector struct {
	base.Composite
	currentChildIndex int
	executionStatus   iface.TaskStatus
}

func (p *Selector) OnAwake() {
	p.executionStatus = iface.Inactive
	p.currentChildIndex = 0
}

func (p *Selector) OnStart() {}

func (p *Selector) CanRunParallelChildren() bool { return false }
func (p *Selector) OnChildExecuted1(childStatus iface.TaskStatus) {
	p.currentChildIndex++
	p.executionStatus = childStatus
}
func (p *Selector) OnChildStarted0() { p.Composite.OnChildStarted0() }

func (p *Selector) CurrentChildIndex() int {
	return p.currentChildIndex
}

func (p *Selector) CanExecute() bool {
	return p.currentChildIndex < len(p.Children()) && p.executionStatus != iface.Success
}

func (p *Selector) OnConditionalAbort(childIndex int) {
	p.Unit().Log().Trace("===============Selector OnConditionalAbort:", childIndex)
	// Set the current child index to the index that caused the abort
	p.currentChildIndex = childIndex
	p.executionStatus = iface.Inactive
}

func (p *Selector) OnCancelConditionalAbort() {
	p.executionStatus = iface.Inactive
	p.currentChildIndex = 0
	p.Unit().Log().Trace("=====Selector	OnCancelConditionalAbort=====")
}

func (p *Selector) OnEnd() {
	p.executionStatus = iface.Inactive
	p.currentChildIndex = 0
	p.Unit().Log().Trace("=====Selector	OnEnd=====")
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.Selector", func() iface.ITask { return &Selector{} })
}
