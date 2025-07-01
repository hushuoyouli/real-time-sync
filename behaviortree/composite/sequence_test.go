package composite

import (
	"testing"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

func TestSequence(t *testing.T) {
	var _ iface.IComposite = &Sequence{}
	var task iface.ITask = &Sequence{}
	_, ok := task.(iface.IComposite)
	if !ok {
		t.Error("类型不满足接口")
	}
}
