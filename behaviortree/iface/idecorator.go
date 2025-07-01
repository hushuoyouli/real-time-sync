package iface

type IDecorator interface {
	IParentTask
	IsDecorator() bool
}
