package register

import (
	"reflect"

	"github.com/hushuoyouli/real-time-sync/rlog"
)

// 自定义从jsonOb生成对象的解析函数
type valFromJsonObjTranform func(interface{}) (interface{}, error)

var ValFromJsonObjTranforms map[reflect.Type]valFromJsonObjTranform = make(map[reflect.Type]valFromJsonObjTranform)

func RegisterValFromJsonObjTranform(ty reflect.Type, handle valFromJsonObjTranform) {
	_, ok := ValFromJsonObjTranforms[ty]
	if ok {
		rlog.Panic("类型:" + ty.Name() + "已经注册过解析函数了")
	}

	ValFromJsonObjTranforms[ty] = handle
}
