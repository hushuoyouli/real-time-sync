package base

import (
	"fmt"
	"testing"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

func DecoratorTest(iface.IDecorator) {

}

func TestDecorator(t *testing.T) {
	DecoratorTest(&Decorator{})
	var task iface.ITask = &Decorator{}
	_, ok := task.(iface.IDecorator)
	fmt.Println(ok)
}
