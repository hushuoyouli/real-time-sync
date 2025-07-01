package composite

import (
	"testing"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

func SelectorTest(iface.IComposite) {

}

func TestSelector(t *testing.T) {
	SelectorTest(&Selector{})
	var task iface.ITask = &Selector{}
	_, ok := task.(iface.IComposite)
	if !ok {
		t.Error("类型不满足接口")
	}
}
