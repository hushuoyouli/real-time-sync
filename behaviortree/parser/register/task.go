package register

import (
	"errors"
	"reflect"
	"strings"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/rlog"
	"github.com/hushuoyouli/real-time-sync/util"
)

var isImplementsIAction = util.IsImplementsIAction
var isImplementsIComposite = util.IsImplementsIComposite
var isImplementsIConditional = util.IsImplementsIConditional
var isImplementsIDecorator = util.IsImplementsIDecorator

var correspondingType2Type map[string]func() iface.ITask = make(map[string]func() iface.ITask)

func RegisterCorrespondingType(correspondingType string, newFunc func() iface.ITask) {
	rlog.Debug("注册任务类型:" + correspondingType)
	taskObj := newFunc()
	taskType := reflect.TypeOf(taskObj).Elem()

	_, ok := correspondingType2Type[correspondingType]
	if ok {
		rlog.Panic("类型:" + correspondingType + "已经注册过了")
	}

	if !isImplementsITask(taskType) {
		rlog.Panic("类型:" + correspondingType + "必须是IAction, IComposite, IDecorator, IConditional中的一种")
	}

	prePath := []string{taskType.Name()}
	if err := checkTaskValid(taskType, prePath); err != nil {
		rlog.Panic(err)
	}

	correspondingType2Type[correspondingType] = newFunc
}

// 检测任务的内嵌任务不能是指针
func checkTaskValid(taskType reflect.Type, prePath []string) error {
	for i := 0; i < taskType.NumField(); i++ {
		field := taskType.Field(i)
		if field.Anonymous {
			if field.Type.Kind() == reflect.Pointer {
				if isImplementsITask(field.Type.Elem()) {
					return errors.New("内嵌路径:" + strings.Join(prePath, "/") + " 内嵌字段:" + field.Name + "内嵌了指针")
				}
			} else {
				if isImplementsITask(field.Type) {
					newPrePath := []string{}
					newPrePath = append(newPrePath, prePath...)
					newPrePath = append(newPrePath, field.Name)

					err := checkTaskValid(field.Type, newPrePath)
					if err != nil {
						return err
					}
				}
			}

		}
	}

	return nil
}

func isImplementsITask(taskType reflect.Type) bool {
	checks := [](func(reflect.Type) bool){
		isImplementsIAction,
		isImplementsIComposite,
		isImplementsIDecorator,
		isImplementsIConditional,
	}

	for _, checkFun := range checks {
		if checkFun(taskType) {
			return true
		}
	}

	return false
}

func NewTask(correspondingType string) (iface.ITask, error) {
	newFunc, ok := correspondingType2Type[correspondingType]
	if !ok {
		return nil, errors.New("客户端类型:" + correspondingType + "对应的golang类型不存在")
	}

	/* 	val := reflect.New(taskType)
	   	task, ok := val.Interface().(iface.ITask)
	   	if !ok {
	   		return nil, errors.New("类型:" + correspondingType + "生成的对象不是ITask接口的类型")
	   	}
	*/

	return newFunc(), nil
	/* 	return task, nil */
}
