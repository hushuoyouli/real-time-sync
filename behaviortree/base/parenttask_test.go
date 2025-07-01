package base

import (
	"fmt"
	"testing"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

func ParentTaskTest(parentTask iface.IParentTask) {

}

func TestParentTask(t *testing.T) {
	ParentTaskTest(&parentTask{})
	var task iface.ITask = &parentTask{}
	_, ok := task.(iface.IParentTask)
	fmt.Println(ok)
}
