package action

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/base"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
)

type PlayAniForSyncData struct {
	Name string
	Age  int64
	Time float32
}

type PlayAniForSync struct {
	base.Action
	AnimationName string
	IsLoop        bool `behaviortree:"isLoop"`
	Data          PlayAniForSyncData
}

func (p *PlayAniForSync) OnStart() {
	//p.Unit().Log().Trace("=====PlayAniForSync	OnStart=====")
}

func (p *PlayAniForSync) OnUpdate() iface.TaskStatus {
	p.Unit().Log().Trace("=====PlayAniForSync	OnUpdate=====", p.AnimationName, p.Data.Name, p.Data.Age, p.Data.Time)
	return iface.Running
}

func (p *PlayAniForSync) OnEnd() {
	//p.Unit().Log().Trace("=====PlayAniForSync	OnEnd=====")
}

func (p *PlayAniForSync) IsSyncToClient() bool {
	return true
}

func (p *PlayAniForSync) RebuildSyncDatas() {
	p.SendSyncData([]byte("RebuildSyncDatas"))
}

func init() {
	register.RegisterCorrespondingType("BehaviorDesigner.Runtime.Tasks.PlayAniForSync", func() iface.ITask { return &PlayAniForSync{} })
}
