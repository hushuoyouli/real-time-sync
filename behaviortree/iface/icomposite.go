package iface

type IComposite interface {
	IParentTask
	AbortType() AbortType
	SetAbortType(abortType AbortType)
	IsComposite() bool
}
