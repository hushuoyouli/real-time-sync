package iface

type IParentTask interface {
	ITask
	MaxChildren() int
	CanRunParallelChildren() bool
	/*
		跟是否可以并发有关的
		OnChildExecuted
		OnChildStarted
		OverrideStatus
	*/
	//	CanRunParallelChildren	为false的时候调用
	OnChildExecuted1(childStatus TaskStatus)
	OnChildStarted0()
	//	CanRunParallelChildren	为true的时候调用
	OnChildExecuted2(index int, childStatus TaskStatus)
	OnChildStarted1(index int)

	CurrentChildIndex() int
	CanExecute() bool
	Decorate(status TaskStatus) TaskStatus

	/*
		TODO：这个部分还需要继续了解
		OverrideStatus
	*/
	OverrideStatus0() TaskStatus
	OverrideStatus1(status TaskStatus) TaskStatus

	OnConditionalAbort(index int)
	OnCancelConditionalAbort() //当Abort取消的时候，会调用这个接口

	Children() []ITask
	AddChild(task ITask)
}
