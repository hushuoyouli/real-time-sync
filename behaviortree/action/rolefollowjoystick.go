package action

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type RoleFollowJoystick struct {
	base.Action
}

func (p *RoleFollowJoystick) OnStart() {
	p.SendSyncData([]byte("OnStart"))
	p.Unit().Log().Trace("=====RoleFollowJoystick	OnStart=====")
}

func (p *RoleFollowJoystick) OnUpdate() iface.TaskStatus {
	p.SendSyncData([]byte("OnUpdate"))
	p.Unit().Log().Trace("=====RoleFollowJoystick	OnUpdate=====")
	return iface.Running
}

func (p *RoleFollowJoystick) OnEnd() {
	p.SendSyncData([]byte("OnEnd"))
	p.Unit().Log().Trace("=====RoleFollowJoystick	OnEnd=====")
}

func (p *RoleFollowJoystick) IsSyncToClient() bool {
	return true
}

func (p *RoleFollowJoystick) RebuildSyncDatas() {
	p.SendSyncData([]byte("RebuildSyncDatas"))
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.Role.MainRole.RoleFollowJoystick", func() iface.ITask { return &RoleFollowJoystick{} })
}
