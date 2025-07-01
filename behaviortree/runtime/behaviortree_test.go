package runtime

import (
	"os"
	"strings"
	"testing"

	_ "github.com/hushuoyouli/real-time-sync/behaviortree/action"    //	通过导入来让相应的模块注册进去
	_ "github.com/hushuoyouli/real-time-sync/behaviortree/composite" //	通过导入来让相应的模块注册进去

	_ "github.com/hushuoyouli/real-time-sync/behaviortree/conditional" //	通过导入来让相应的模块注册进去

	"github.com/hushuoyouli/real-time-sync/behaviortree/conditional"
)

func Test_BehaviorTree_Parser(t *testing.T) {
	bytes, err := os.ReadFile("../parser/test_behaviortree.json")
	if err != nil {
		t.Error(err)
		return
	}

	behaviorTree := NewBehaviorTree(bytes, &TestUnit{}, NewClock(), NewDefaultRuntimeEventHandle())
	behaviorTree.Enable()
}

func Test_BehaviorTree_Enable(t *testing.T) {
	bytes, err := os.ReadFile("../parser/test_behaviortree.json")
	if err != nil {
		t.Error(err)
		return
	}
	behaviorTree := NewBehaviorTree(bytes, &TestUnit{}, NewClock(), NewDefaultRuntimeEventHandle())
	if err := behaviorTree.Enable(); err != nil {
		t.Log(err)
	}

	behaviorTree.Print()
	rebuildSyncDataCollector := newDefaultRebuildSyncDataCollector()
	t.Log(strings.Repeat("1", 100))
	behaviorTree.Update()
	behaviorTree.Print()
	behaviorTree.RebuildSync(rebuildSyncDataCollector)

	t.Log(strings.Repeat("2", 100))
	conditional.NeedFollowJoystickFlag = true
	behaviorTree.Update()
	behaviorTree.Print()
	behaviorTree.RebuildSync(rebuildSyncDataCollector)

	t.Log(strings.Repeat("3", 100))
	conditional.NeedFollowJoystickFlag = false
	behaviorTree.Update()
	behaviorTree.Print()
	behaviorTree.RebuildSync(rebuildSyncDataCollector)

	t.Log(strings.Repeat("4", 100))
	conditional.NeedFollowJoystickFlag = true
	behaviorTree.Update()
	behaviorTree.Print()
	behaviorTree.RebuildSync(rebuildSyncDataCollector)

	t.Log(strings.Repeat("5", 100))
	conditional.NeedFollowJoystickFlag = false
	behaviorTree.Update()
	behaviorTree.Print()
	behaviorTree.RebuildSync(rebuildSyncDataCollector)

	t.Log(strings.Repeat("6", 100))
	behaviorTree.Disable()
	behaviorTree.Print()
	behaviorTree.RebuildSync(rebuildSyncDataCollector)
	/*
		t.Log(strings.Repeat("3", 100))
		conditional.NeedFollowJoystickFlag = false
		behaviorTree.Update()
		behaviorTree.Update()
		behaviorTree.Print()

		t.Log(strings.Repeat("4", 100))
		conditional.NeedFollowJoystickFlag = true
		behaviorTree.Update()
		behaviorTree.Update()
		behaviorTree.Print()

		t.Log(strings.Repeat("5", 100))
		conditional.NeedFollowJoystickFlag = false
		behaviorTree.Update()
		behaviorTree.Update()
		behaviorTree.Print()

		t.Log(strings.Repeat("6", 100))
		conditional.NeedFollowJoystickFlag = true
		behaviorTree.Update()
		behaviorTree.Update()
		behaviorTree.Print()

		t.Log(strings.Repeat("9", 100))
		behaviorTree.Disable()
		behaviorTree.Print()

		t.Log(strings.Repeat("=", 100))
		behaviorTree.Enable()
		behaviorTree.Print()
	*/
	//t.Log(strings.Repeat("=", 15))
}
