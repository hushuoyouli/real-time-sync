package base

import (
	"fmt"
	"testing"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

func ConditionalTest(iface.IConditional) {

}

func TestConditional(t *testing.T) {
	ConditionalTest(&Conditional{})
	var task iface.ITask = &Conditional{}
	_, ok := task.(iface.IConditional)
	fmt.Println(ok)
}
