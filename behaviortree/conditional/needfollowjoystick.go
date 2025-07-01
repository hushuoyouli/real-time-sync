package conditional

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type NeedFollowJoystick struct {
	base.Conditional
}

var NeedFollowJoystickFlag = false

func (p *NeedFollowJoystick) OnStart() {
	//p.Unit().Log().Trace("=====NeedFollowJoystick	OnStart=====")
}

func (p *NeedFollowJoystick) OnUpdate() iface.TaskStatus {
	p.Unit().Log().Trace("=====NeedFollowJoystick	OnUpdate=====")
	if NeedFollowJoystickFlag {
		return iface.Success
	} else {
		return iface.Failure
	}
}

func (p *NeedFollowJoystick) OnEnd() {
	//p.Unit().Log().Trace("=====NeedFollowJoystick	OnEnd=====")
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.Role.MainRole.NeedFollowJoystick", func() iface.ITask { return &NeedFollowJoystick{} })
}
