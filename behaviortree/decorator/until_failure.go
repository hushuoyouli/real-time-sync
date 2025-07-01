package decorator

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type UntilFailure struct {
	base.Decorator
	executionStatus iface.TaskStatus
}

func (p *UntilFailure) OnAwake() {
	p.executionStatus = iface.Inactive
}

func (p *UntilFailure) CanExecute() bool {
	return p.executionStatus == iface.Success || p.executionStatus == iface.Inactive
}

func (p *UntilFailure) OnEnd() {
	p.executionStatus = iface.Inactive
}

func (p *UntilFailure) OnChildExecuted1(childStatus iface.TaskStatus) {
	p.executionStatus = childStatus
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.UntilFailure", func() iface.ITask { return &UntilFailure{} })
}
