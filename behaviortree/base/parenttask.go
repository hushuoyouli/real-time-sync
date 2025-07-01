package base

import (
	"math"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
)

type parentTask struct {
	task
	childTasks []iface.ITask
}

func (p *parentTask) MaxChildren() int             { return math.MaxInt }
func (p *parentTask) CanRunParallelChildren() bool { return false }

// CanRunParallelChildren	为false的时候调用
func (p *parentTask) OnChildExecuted1(childStatus iface.TaskStatus) {}
func (p *parentTask) OnChildStarted0()                              {}

// CanRunParallelChildren	为true的时候调用
func (p *parentTask) OnChildExecuted2(index int, childStatus iface.TaskStatus) {}
func (p *parentTask) OnChildStarted1(index int)                                {}

func (p *parentTask) CurrentChildIndex() int                                   { return 0 }
func (p *parentTask) CanExecute() bool                                         { return true }
func (p *parentTask) Decorate(status iface.TaskStatus) iface.TaskStatus        { return status }
func (p *parentTask) OverrideStatus0() iface.TaskStatus                        { return iface.Running }
func (p *parentTask) OverrideStatus1(status iface.TaskStatus) iface.TaskStatus { return status }
func (p *parentTask) OnConditionalAbort(index int)                             {}
func (p *parentTask) OnCancelConditionalAbort()                                {} //	当Abort取消的时候，会调用这个接口
func (p *parentTask) Children() []iface.ITask                                  { return p.childTasks }
func (p *parentTask) AddChild(task iface.ITask)                                { p.childTasks = append(p.childTasks, task) }
func (p *parentTask) IsImplementsIParentTask() bool                            { return true }
