package decorator

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type UntilSuccess struct {
	base.Decorator
	executionStatus iface.TaskStatus
}

func (p *UntilSuccess) OnAwake() {
	p.executionStatus = iface.Inactive
}

func (p *UntilSuccess) CanExecute() bool {
	return p.executionStatus == iface.Failure || p.executionStatus == iface.Inactive
}

func (p *UntilSuccess) OnEnd() {
	p.executionStatus = iface.Inactive
}

func (p *UntilSuccess) OnChildExecuted1(childStatus iface.TaskStatus) {
	p.executionStatus = childStatus
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.UntilSuccess", func() iface.ITask { return &UntilSuccess{} })
}
