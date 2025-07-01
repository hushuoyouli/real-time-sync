package decorator

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type ReturnSuccess struct {
	base.Decorator
	executionStatus iface.TaskStatus
}

func (p *ReturnSuccess) OnAwake() {
	p.executionStatus = iface.Inactive
}

func (p *ReturnSuccess) CanExecute() bool {
	return p.executionStatus == iface.Running || p.executionStatus == iface.Inactive
}

func (p *ReturnSuccess) OnEnd() {
	p.executionStatus = iface.Inactive
}

func (p *ReturnSuccess) OnChildExecuted1(childStatus iface.TaskStatus) {
	p.executionStatus = childStatus
}

func (p *ReturnSuccess) Decorate(status iface.TaskStatus) iface.TaskStatus {
	if status == iface.Failure {
		return iface.Success
	}
	return status
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.ReturnSuccess", func() iface.ITask { return &ReturnSuccess{} })
}
