package decorator

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type UntilForever struct {
	base.Decorator
}

func (p *UntilForever) OnAwake() {
}

func (p *UntilForever) CanExecute() bool {
	return true
}

func (p *UntilForever) OnEnd() {
}

func (p *UntilForever) OnChildExecuted1(childStatus iface.TaskStatus) {
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.UntilForever", func() iface.ITask { return &UntilForever{} })
}
