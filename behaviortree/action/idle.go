package action

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type Idle struct {
	base.Action
	AnimationName string
	IsLoop        bool
}

func (p *Idle) OnAwake() {
	//rlog.Trace("=====Idle	OnAwake=====")
}

func (p *Idle) OnStart() {
	//rlog.Trace("=====Idle	OnStart=====")
}

func (p *Idle) OnUpdate() iface.TaskStatus {
	//rlog.Trace("=====Idle	OnUpdate=====")
	return iface.Running
}

func (p *Idle) OnEnd() {
	//rlog.Trace("=====Idle	OnEnd=====")
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.Idle", func() iface.ITask { return &Idle{} })
}
