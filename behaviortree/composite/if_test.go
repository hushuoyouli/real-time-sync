package composite

import (
	"fmt"
	"testing"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

func IfTest(iface.IComposite) {

}

func TestComposite(t *testing.T) {
	IfTest(&If{})
	var task iface.ITask = &If{}
	_, ok := task.(iface.IComposite)
	fmt.Println(ok)
}
