package composite

import (
	"testing"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

func ParallelTest(iface.IComposite) {

}

func TestParallel(t *testing.T) {
	ParallelTest(&Parallel{})
	var task iface.ITask = &Parallel{}
	_, ok := task.(iface.IComposite)
	if !ok {
		t.Error("类型不满足接口")
	}
}
