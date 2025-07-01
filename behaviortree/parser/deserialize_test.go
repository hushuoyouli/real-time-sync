package parser

import (
	"os"
	"reflect"
	"testing"

	_ "github.com/hushuoyouli/real-time-sync/behaviortree/action"      //	通过导入来让相应的模块注册进去
	_ "github.com/hushuoyouli/real-time-sync/behaviortree/composite"   //	通过导入来让相应的模块注册进去
	_ "github.com/hushuoyouli/real-time-sync/behaviortree/conditional" //	通过导入来让相应的模块注册进去
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
	"github.com/hushuoyouli/real-time-sync/util"
)

func TestDeserialize(t *testing.T) {
	bytes, err := os.ReadFile("./test_behaviortree.json")
	if err != nil {
		t.Error(err)
	}

	if _, err := Deserialize(bytes, util.NewTaskAddData(nil, nil)); err != nil {
		t.Error(err)
	} else {
		//t.Log(rootTask)
		//fmt.Println(reflect.TypeOf(rootTask))
	}
}

func TestSetTask(t *testing.T) {
	correspondingType := "BehaviorDesigner.Runtime.Tasks.PlayAniForSync"
	task, err := register.NewTask(correspondingType)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(task)
	}
}

func Test_valFromJsonval_Pointer(t *testing.T) {
	var p **int
	val, err := valFromJsonval(reflect.TypeOf(p), 30)
	if err != nil {
		t.Error(err)
		return
	}

	pp := reflect.ValueOf(&p)
	pp.Elem().Set(reflect.ValueOf(val))

	t.Log(p)
	t.Log(*p)
	t.Log(**p)
}

func Test_valFromJsonval_Slice(t *testing.T) {
	var p [][]int
	val, err := valFromJsonval(reflect.TypeOf(p), [][]interface{}{{1, 2, 3, 4, 5}, {6, 7, 8}, {9, 10}, {11}})
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(val)
}

func Test_valFromJsonval_array(t *testing.T) {
	var p [3][3]int
	val, err := valFromJsonval(reflect.TypeOf(p), [][]interface{}{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}})
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(val)
}

type TestvalFromJsonvalStruct struct {
	Name string
	Age  int
}

func Test_valFromJsonval_struct(t *testing.T) {
	var val TestvalFromJsonvalStruct
	valMap := map[string]interface{}{
		"Name": "TestvalFromJsonvalStruct",
		"Age":  30,
	}

	result, err := valFromJsonval(reflect.TypeOf(val), valMap)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(result)
	}
}

func Test_RegisterValFromJsonObjTranform(t *testing.T) {
	transformFun := func(jsonObj interface{}) (interface{}, error) {
		return "TestvalFromJsonvalStruct", nil
	}

	register.RegisterValFromJsonObjTranform(reflect.TypeOf(int(0)), transformFun)

	result, err := valFromJsonval(reflect.TypeOf(int(0)), 1000)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(result)
	}
}
