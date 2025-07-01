package conditional

import (
	"testing"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

func TestNeedFollowJoystick(t *testing.T) {
	var _ iface.IConditional = &NeedFollowJoystick{}
	var task iface.ITask = &NeedFollowJoystick{}
	_, ok := task.(iface.IConditional)
	if !ok {
		t.Error("类型不满足接口")
	}
}
