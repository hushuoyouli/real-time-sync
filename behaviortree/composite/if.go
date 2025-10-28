package composite

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type If struct {
	base.Composite
	currentChildIndex int
	executionStatus   iface.TaskStatus
}

func (p *If) OnStart() {

}

func (p *If) MaxChildren() int {
	return 3
}

func (p *If) CanRunParallelChildren() bool { return false }
func (p *If) OnChildExecuted1(childStatus iface.TaskStatus) {
	if p.currentChildIndex == 0 {
		if childStatus == iface.Success {
			p.currentChildIndex = 1
		} else {
			p.currentChildIndex = 2
		}

		p.executionStatus = childStatus
	} else {
		p.executionStatus = childStatus
		p.currentChildIndex = 3
	}
}
func (p *If) OnChildStarted0() { p.Composite.OnChildStarted0() }

func (p *If) CurrentChildIndex() int {
	return p.currentChildIndex
}

func (p *If) CanExecute() bool {
	return p.currentChildIndex < len(p.Children())
}

func (p *If) OnConditionalAbort(childIndex int) {
	p.Unit().Log().Trace("==========If OnConditionalAbort:", childIndex)
	p.currentChildIndex = childIndex
}

func (p *If) OnCancelConditionalAbort() {
	p.executionStatus = iface.Inactive
	p.currentChildIndex = 0
}

func (p *If) OnEnd() {
	p.executionStatus = iface.Inactive
	p.currentChildIndex = 0
	//p.Unit().Log().Trace("=====If	OnEnd=====")
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.If", func() iface.ITask { return &If{} })
}
