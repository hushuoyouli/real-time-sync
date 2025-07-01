package register

import (
	"reflect"
	"testing"
)

// 注册测试
func TestRegisterValFromJsonObjTranform(t *testing.T) {
	transformFun := func(jsonObj interface{}) (interface{}, error) {
		return "TestvalFromJsonvalStruct", nil
	}

	RegisterValFromJsonObjTranform(reflect.TypeOf(int(0)), transformFun)

}
