package util

import (
	"reflect"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

func IsImplementsIAction(taskType reflect.Type) bool {
	iTaskType := reflect.TypeOf((*iface.IAction)(nil)).Elem()
	taskPtrType := reflect.PointerTo(taskType)
	return taskPtrType.Implements(iTaskType)
}

func IsImplementsIComposite(taskType reflect.Type) bool {
	iTaskType := reflect.TypeOf((*iface.IComposite)(nil)).Elem()
	taskPtrType := reflect.PointerTo(taskType)
	return taskPtrType.Implements(iTaskType)
}

func IsImplementsIDecorator(taskType reflect.Type) bool {
	iTaskType := reflect.TypeOf((*iface.IDecorator)(nil)).Elem()
	taskPtrType := reflect.PointerTo(taskType)
	return taskPtrType.Implements(iTaskType)
}

func IsImplementsIConditional(taskType reflect.Type) bool {
	iTaskType := reflect.TypeOf((*iface.IConditional)(nil)).Elem()
	taskPtrType := reflect.PointerTo(taskType)
	return taskPtrType.Implements(iTaskType)
}

func IsImplementsIParentTask(taskType reflect.Type) bool {
	iTaskType := reflect.TypeOf((*iface.IParentTask)(nil)).Elem()
	taskPtrType := reflect.PointerTo(taskType)
	return taskPtrType.Implements(iTaskType)
}
