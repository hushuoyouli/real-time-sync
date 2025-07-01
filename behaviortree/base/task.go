package base

import (
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

type task struct {
	correspondingType string
	owner             iface.IBehaviorTree
	parent            iface.IParentTask
	id                int
	name              string
	isInstant         bool
	disabled          bool
	unit              iface.IUnit

	/*
		isExecute         bool
		startTime         int64
		executeID         int
		activeStackID     int
	*/
}

// 对应客户端的类型
func (p *task) CorrespondingType() string { return p.correspondingType }
func (p *task) SetCorrespondingType(correspondingType string) {
	p.correspondingType = correspondingType
}

func (p *task) Owner() iface.IBehaviorTree         { return p.owner }
func (p *task) SetOwner(owner iface.IBehaviorTree) { p.owner = owner }

func (p *task) Parent() iface.IParentTask          { return p.parent }
func (p *task) SetParent(parent iface.IParentTask) { p.parent = parent }

// ID
func (p *task) ID() int      { return p.id }
func (p *task) SetID(id int) { p.id = id }

func (p *task) Name() string        { return p.name }
func (p *task) SetName(name string) { p.name = name }

func (p *task) IsInstant() bool             { return p.isInstant }
func (p *task) SetIsInstant(isInstant bool) { p.isInstant = isInstant }

func (p *task) Disabled() bool            { return p.disabled }
func (p *task) SetDisabled(disabled bool) { p.disabled = disabled }

func (p *task) Unit() iface.IUnit        { return p.unit }
func (p *task) SetUnit(unit iface.IUnit) { p.unit = unit }

/*
	func (p *task) IsExecute() bool             { return p.isExecute }
	func (p *task) SetIsExecute(isExecute bool) { p.isExecute = isExecute }

	func (p *task) StartTime() int64             { return p.startTime }
	func (p *task) SetStartTime(startTime int64) { p.startTime = startTime }

	func (p *task) ExecuteID() int             { return p.executeID }
	func (p *task) SetExecuteID(executeID int) { p.executeID = executeID }

	func (p *task) ActiveStackID() int                 { return p.activeStackID }
	func (p *task) SetActiveStackID(activeStackID int) { p.activeStackID = activeStackID }
*/

func (p *task) OnAwake()                   {}
func (p *task) OnStart()                   {}
func (p *task) OnUpdate() iface.TaskStatus { return iface.Success }
func (p *task) OnEnd()                     {}
func (p *task) OnComplete()                {}

/*
correspondingType string

owner             iface.IBehaviorTree
parent            iface.IParentTask
id                int
name              string
isInstant         bool
disabled          bool
unit              iface.IUnit
isExecute         bool
startTime         int64
executeID         int
activeStackID     int
*/
func (p *task) DebugInfo() map[string]interface{} {
	return map[string]interface{}{
		"core": []interface{}{p.ID(), p.Name()},
		"info": []interface{}{
			[]interface{}{"Owner", p.owner.ID()},
			[]interface{}{"IsInstant", p.IsInstant()},
			[]interface{}{"Disabled", p.Disabled()},
			[]interface{}{"Unit", p.Unit().ID()},
		},
	}
}

func (p *task) IsImplementsIAction() bool {
	return false
}

func (p *task) IsImplementsIComposite() bool {
	return false
}

func (p *task) IsImplementsIDecorator() bool {
	return false
}

func (p *task) IsImplementsIConditional() bool {
	return false
}

func (p *task) IsImplementsIParentTask() bool {
	return false
}

func (p *task) SetVariables(variableConfigs map[string]interface{}) error {
	return nil
}
