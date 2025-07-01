package base

import (
	"fmt"
	"testing"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

func TaskTest(parentTask iface.ITask) {

}

func TestTask(t *testing.T) {
	TaskTest(&task{})
	var task iface.ITask = &task{}
	_, ok := task.(iface.ITask)
	fmt.Println(ok)
}
