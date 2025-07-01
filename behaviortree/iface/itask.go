package iface

type ITask interface {
	//	对应客户端的名字
	CorrespondingType() string
	SetCorrespondingType(correspondingType string)
	//	所属的树
	Owner() IBehaviorTree
	SetOwner(owner IBehaviorTree)
	//	父节点
	Parent() IParentTask
	SetParent(parent IParentTask)
	//	ID
	ID() int
	SetID(id int)
	//	名字
	Name() string
	SetName(name string)
	//是否是Instant
	IsInstant() bool
	SetIsInstant(isInstant bool)
	//是否无效
	Disabled() bool
	SetDisabled(disabled bool)
	//树的宿主
	Unit() IUnit
	SetUnit(unit IUnit)

	OnAwake()
	OnStart()
	OnUpdate() TaskStatus
	OnEnd()
	OnComplete()

	DebugInfo() map[string]interface{}

	IsImplementsIAction() bool
	IsImplementsIComposite() bool
	IsImplementsIDecorator() bool
	IsImplementsIConditional() bool
	IsImplementsIParentTask() bool
	SetVariables(variableConfigs map[string]interface{}) error
}
