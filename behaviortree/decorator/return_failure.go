package decorator

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type ReturnFailure struct {
	base.Decorator
	executionStatus iface.TaskStatus
}

func (p *ReturnFailure) OnAwake() {
	p.executionStatus = iface.Inactive
}

func (p *ReturnFailure) CanExecute() bool {
	return p.executionStatus == iface.Running || p.executionStatus == iface.Inactive
}

func (p *ReturnFailure) OnEnd() {
	p.executionStatus = iface.Inactive
}

func (p *ReturnFailure) OnChildExecuted1(childStatus iface.TaskStatus) {
	p.executionStatus = childStatus
}

func (p *ReturnFailure) Decorate(status iface.TaskStatus) iface.TaskStatus {
	if status == iface.Success {
		return iface.Failure
	}
	return status
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.ReturnFailure", func() iface.ITask { return &ReturnFailure{} })
}
