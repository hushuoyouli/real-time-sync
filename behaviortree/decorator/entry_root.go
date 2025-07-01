package decorator

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

type EntryRoot struct {
	base.Decorator
	executionStatus iface.TaskStatus
}

func (p *EntryRoot) OnAwake() {
	p.executionStatus = iface.Inactive

}

func (p *EntryRoot) CanExecute() bool {
	return p.executionStatus == iface.Inactive
}

func (p *EntryRoot) OnStart() {
}

func (p *EntryRoot) OnEnd() {
	p.executionStatus = iface.Inactive
}

func (p *EntryRoot) OnChildExecuted1(childStatus iface.TaskStatus) {
	p.executionStatus = childStatus
}
