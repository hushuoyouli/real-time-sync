package base

import (
	"fmt"
	"testing"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

func CompositeTest(iface.IComposite) {

}

func TestComposite(t *testing.T) {
	CompositeTest(&Composite{})
	var task iface.ITask = &Composite{}
	_, ok := task.(iface.IComposite)
	fmt.Println(ok)
}
