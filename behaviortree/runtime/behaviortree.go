package runtime

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/hushuoyouli/real-time-sync/behaviortree/decorator"
	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser"
	"github.com/hushuoyouli/real-time-sync/util"
)

type BehaviorTree struct {
	id int64

	taskList           util.List[iface.ITask]
	parentIndex        util.List[int]
	childrenIndex      util.List[util.List[int]]
	relativeChildIndex util.List[int]

	activeStack              util.List[*util.Stack[int]]
	nonInstantTaskStatus     util.List[iface.TaskStatus]
	conditionalReevaluate    util.List[*ConditionalReevaluate]
	conditionalReevaluateMap util.Dictionary[int, *ConditionalReevaluate]

	parentCompositeIndex  util.List[int]
	childConditionalIndex util.List[util.List[int]]

	isRunning                        bool
	initializeFirstStackAndFirstTask bool //	是否需要初始化第一个执行栈和第一个任务
	executionStatus                  iface.TaskStatus
	config                           []byte
	unit                             iface.IUnit
	rootTask                         iface.ITask
	clock                            iface.IClock
	stackID                          int
	stackDatas                       map[*util.Stack[int]]*iface.StackRuntimeData
	stackID2StackData                map[int]*iface.StackRuntimeData

	taskDatas map[int]*iface.TaskRuntimeData

	stackID2ParallelTaskID  map[int]int
	parallelTaskID2StackIDs map[int][]int

	runtimeEventHandle    iface.IRuntimeEventHandle
	initializeForBaseFlag bool

	extraParam interface{}
}

func NewBehaviorTree(config []byte, unit iface.IUnit, clock iface.IClock, runtimeEventHandle iface.IRuntimeEventHandle) *BehaviorTree {
	return &BehaviorTree{
		taskList:           util.NewList[iface.ITask](20),
		parentIndex:        util.NewList[int](20),
		childrenIndex:      util.NewList[util.List[int]](20),
		relativeChildIndex: util.NewList[int](20),

		activeStack:              util.NewList[*util.Stack[int]](20),
		nonInstantTaskStatus:     util.NewList[iface.TaskStatus](20),
		conditionalReevaluate:    util.NewList[*ConditionalReevaluate](20),
		conditionalReevaluateMap: util.NewDictionary[int, *ConditionalReevaluate](),

		parentCompositeIndex:  util.NewList[int](20),
		childConditionalIndex: util.NewList[util.List[int]](20),

		isRunning:       false,
		executionStatus: iface.Inactive,
		config:          config,
		unit:            unit,
		rootTask:        nil,

		clock:             clock,
		stackID:           1,
		stackDatas:        make(map[*util.Stack[int]]*iface.StackRuntimeData),
		stackID2StackData: make(map[int]*iface.StackRuntimeData),
		taskDatas:         make(map[int]*iface.TaskRuntimeData),

		stackID2ParallelTaskID:  make(map[int]int),
		parallelTaskID2StackIDs: make(map[int][]int),

		runtimeEventHandle:    runtimeEventHandle,
		initializeForBaseFlag: false,

		extraParam: nil,
	}
}

func (p *BehaviorTree) ExtraParam() interface{} {
	return p.extraParam
}

func (p *BehaviorTree) SetExtraParam(param interface{}) {
	p.extraParam = param
}

func (p *BehaviorTree) Config() []byte {
	return p.config
}

func (p *BehaviorTree) Clock() iface.IClock {
	return p.clock
}

func (p *BehaviorTree) SetClock(clock iface.IClock) {
	p.clock = clock
}

func (p *BehaviorTree) SetRuntimeEventHandle(runtimeEventHandle iface.IRuntimeEventHandle) {
	p.runtimeEventHandle = runtimeEventHandle
}

func (p *BehaviorTree) PrintTask(task iface.ITask, tabCount int) {
	resultBytes, _ := json.Marshal(task.DebugInfo()["core"])
	infoBytes, _ := json.Marshal(task.DebugInfo()["info"])

	p.unit.Log().Trace(strings.Repeat("\t", tabCount), string(resultBytes), string(infoBytes))
	//if task.IsImplementsIParentTask() {
	if task.IsImplementsIParentTask() {
		parentTask := task.(iface.IParentTask)
		for _, childTask := range parentTask.Children() {
			p.PrintTask(childTask, tabCount+1)
		}
	}
}

func (p *BehaviorTree) Print() {
	p.unit.Log().Trace("id:", p.id)
	p.unit.Log().Trace("=====taskList start=====")
	p.PrintTask(p.rootTask, 0)
	p.unit.Log().Trace("=====taskList end=====")

	p.unit.Log().Trace("parentIndex:", p.parentIndex)
	p.unit.Log().Trace("childrenIndex:", p.childrenIndex)
	p.unit.Log().Trace("relativeChildIndex:", p.relativeChildIndex)

	activeStacks := make(map[int]util.Stack[int])
	for index, activeStack := range p.activeStack {
		activeStacks[p.getStackRuntimeData(index).StackID] = *activeStack
	}

	p.unit.Log().Trace("activeStack:", activeStacks)
	p.unit.Log().Trace("nonInstantTaskStatus:", p.nonInstantTaskStatus)

	p.unit.Log().Trace("=====conditionalReevaluate start=====")
	p.unit.Log().Trace("conditionalReevaluate:", len(p.conditionalReevaluateMap))
	for _, conditionalReevaluate := range p.conditionalReevaluate {
		p.unit.Log().Trace(conditionalReevaluate.DebugInfo()...)
	}
	p.unit.Log().Trace("=====conditionalReevaluate end=====")

	p.unit.Log().Trace("stackID2ParallelTaskID:", p.stackID2ParallelTaskID)
	p.unit.Log().Trace("parallelTaskID2StackIDs:", p.parallelTaskID2StackIDs)
}

// 注意这个接口，行为树每次enable后才会生成独一无二的id
func (p *BehaviorTree) ID() int64 {
	return p.id
}

func (p *BehaviorTree) IsRunning() bool {
	return p.isRunning
}

var behaviorTreeId int64 = 1
var behaviorTreeIdMutex sync.Mutex

func nextBehaviorTreeID() int64 {
	behaviorTreeIdMutex.Lock()
	defer behaviorTreeIdMutex.Unlock()
	val := behaviorTreeId
	behaviorTreeId++
	return val
}

func (p *BehaviorTree) Unit() iface.IUnit {
	return p.unit
}

func (p *BehaviorTree) SetUnit(unit iface.IUnit) {
	p.unit = unit
	if p.initializeForBaseFlag {
		for _, task := range p.taskList {
			task.SetUnit(unit)
		}
	}
}

func (p *BehaviorTree) RebuildSync(collector iface.IRebuildSyncDataCollector) {
	if p.isRunning {
		for stackIndex, _ := range p.activeStack {
			stackRuntimeData := p.getStackRuntimeData(stackIndex)
			collector.Stack(p, stackRuntimeData)
		}

		for stackIndex, stack := range p.activeStack {
			if stack.Len() > 0 {
				stackRuntimeData := p.getStackRuntimeData(stackIndex)
				taskId := stack.Peak()
				task := p.taskList[taskId]
				if task.IsImplementsIAction() {
					action := task.(iface.IAction)
					if action.IsSyncToClient() {
						taskRuntimeData := p.taskDatas[task.ID()]
						action.SyncDataCollector().GetAndClear()
						action.RebuildSyncDatas()
						syncDatas := action.SyncDataCollector().GetAndClear()
						collector.Action(p, taskRuntimeData, stackRuntimeData, task, syncDatas)
					}
				} else if task.IsImplementsIParentTask() {
					parentTask := task.(iface.IParentTask)
					if parentTask.CanRunParallelChildren() {
						taskRuntimeData := p.taskDatas[parentTask.ID()]
						childStackRuntimeDatas := make([]*iface.StackRuntimeData, 0)
						for _, childStackID := range p.parallelTaskID2StackIDs[parentTask.ID()] {
							childStackRuntimeDatas = append(childStackRuntimeDatas, p.stackID2StackData[childStackID])
						}

						collector.Parallel(p, taskRuntimeData, stackRuntimeData, task, childStackRuntimeDatas)
					}
				}
			}
		}
	}

}

func (p *BehaviorTree) ExecutionStatus() iface.TaskStatus {
	return p.executionStatus
}

func (p *BehaviorTree) Enable() error {
	if !p.isRunning {
		p.executionStatus = iface.Inactive
		p.id = nextBehaviorTreeID()

		if err := p.Initialize(); err != nil {
			return err
		}

		//	初始化数据收集器
		for _, task := range p.taskList {
			if task.IsImplementsIAction() {
				action := task.(iface.IAction)
				if action.IsSyncToClient() {
					action.SetSyncDataCollector(iface.NewSyncDataCollector())
				}
			}
		}

		for _, task := range p.taskList {
			if !task.Disabled() {
				task.OnAwake()
			}
		}

		p.executionStatus = iface.Running
		p.isRunning = true

		nowTimesampInMill := p.clock.TimesampInMill()
		p.runtimeEventHandle.PostInitialize(p, nowTimesampInMill)

		/*
			p.AddStack()
			p.PushTask(0, 0)
		*/
		p.initializeFirstStackAndFirstTask = true
		return nil
	} else {
		return errors.New("already running")
	}
}

func (p *BehaviorTree) PushTask(taskIndex, stackIndex int) {
	if !p.isRunning || stackIndex >= p.activeStack.Count() {
		return
	}

	if p.activeStack[stackIndex].Len() == 0 || p.activeStack[stackIndex].Peak() != taskIndex {
		p.activeStack[stackIndex].Push(taskIndex)
		p.nonInstantTaskStatus[stackIndex] = iface.Running

		task := p.taskList[taskIndex]

		stack := p.activeStack[stackIndex]
		stackData := p.stackDatas[stack]
		nowTimestamp := p.clock.TimesampInMill()
		taskExecuteID := p.nextTaskExecuteID()

		taskRuntimeData := iface.NewTaskRuntimeData(task.ID(), nowTimestamp, taskExecuteID, stackData.StackID)
		p.taskDatas[task.ID()] = taskRuntimeData

		//	TODO:这里需要截获初始化的数据？
		p.runtimeEventHandle.PreOnStart(p, taskRuntimeData, stackData, task)
		if task.IsImplementsIParentTask() {
			parentTask := task.(iface.IParentTask)
			if parentTask.CanRunParallelChildren() {
				p.runtimeEventHandle.ParallelPreOnStart(p, taskRuntimeData, stackData, task)
			}
		}

		//	先清理数据
		if task.IsImplementsIAction() {
			action := task.(iface.IAction)
			if action.IsSyncToClient() {
				action.SyncDataCollector().GetAndClear()
			}
		}
		task.OnStart()
		if task.IsImplementsIAction() {
			action := task.(iface.IAction)
			if action.IsSyncToClient() {
				datas := action.SyncDataCollector().GetAndClear()
				p.runtimeEventHandle.ActionPostOnStart(p, taskRuntimeData, stackData, task, datas)
			}
		}

		if task.IsImplementsIConditional() {
			var conditionalReevaluate *ConditionalReevaluate
			if p.conditionalReevaluateMap.TryGetValue(taskIndex, &conditionalReevaluate) {
				conditionalReevaluate.compositeIndex = -1
			}
		}

		if task.IsImplementsIParentTask() {
			//	可以并发的父节点有特殊处理
			parentTask := task.(iface.IParentTask)
			if parentTask.CanRunParallelChildren() {
				p.parallelTaskID2StackIDs[task.ID()] = make([]int, 0)
			}

			if task.IsImplementsIComposite() {
				compositeTask := task.(iface.IComposite)

				if compositeTask.AbortType() != iface.None {
					for _, conditionalReevaluate := range p.conditionalReevaluate {
						if p.IsParentTask(taskIndex, conditionalReevaluate.index) {
							conditionalReevaluate.compositeIndex = taskIndex
						}
					}

					if compositeTask.AbortType() == iface.LowerPriority {
						childConditionalIndexes := p.childConditionalIndex[compositeTask.ID()]
						for _, childConditionalIndex := range childConditionalIndexes {
							var conditionalReevaluate *ConditionalReevaluate
							if p.conditionalReevaluateMap.TryGetValue(childConditionalIndex, &conditionalReevaluate) {
								conditionalReevaluate.compositeIndex = -1
							}
						}
					}
				}
			}
		}
	}
}

func (p *BehaviorTree) initializeForBase() error {
	p.taskList.Clear()
	p.parentIndex.Clear()
	p.childrenIndex.Clear()
	p.relativeChildIndex.Clear()

	p.parentCompositeIndex.Clear()
	p.childConditionalIndex.Clear()
	p.rootTask = nil

	taskAddData := util.NewTaskAddData(p, p.unit)
	/* 	bytes, err := os.ReadFile(p.configFile)
	   	if err != nil {
	   		return err
	   	} */

	rootTask, err := parser.Deserialize(p.config, taskAddData)
	if err != nil {
		return err
	}

	/*增加一个空的EntryRoot*/
	entryRoot := &decorator.EntryRoot{}
	entryRoot.SetOwner(rootTask.Owner())
	entryRoot.SetUnit(rootTask.Unit())
	entryRoot.SetCorrespondingType("EntryRoot")
	entryRoot.SetName("EntryRoot")
	entryRoot.AddChild(rootTask)
	rootTask = entryRoot

	p.rootTask = rootTask
	p.taskList.Add(p.rootTask)
	p.parentIndex.Add(-1)
	p.parentCompositeIndex.Add(-1)
	p.childConditionalIndex.Add(util.NewList[int](10))
	p.childrenIndex.Add(util.NewList[int](10))
	p.relativeChildIndex.Add(-1)

	parentCompositeIndex := -1
	p.rootTask.SetID(0)
	/*
		if util.IsImplementsIAction(reflect.TypeOf(p.rootTask).Elem()) {
			action := p.rootTask.(iface.IAction)
			if action.IsSyncToClient() {
				action.SetSyncDataCollector(iface.NewSyncDataCollector())
			}
		}
	*/

	if p.rootTask.IsImplementsIParentTask() {
		if p.rootTask.IsImplementsIComposite() {
			parentCompositeIndex = p.rootTask.ID()
		}

		parentTask := p.rootTask.(iface.IParentTask)
		for _, childTask := range parentTask.Children() {
			if err := p.ParseChildTask(childTask, parentTask, parentCompositeIndex); err != nil {
				return err
			}
		}
	}

	return err
}

func (p *BehaviorTree) Initialize() error {
	if !p.initializeForBaseFlag {
		if err := p.initializeForBase(); err != nil {
			return err
		}

		p.initializeForBaseFlag = true
	}

	p.stackID = 1

	p.activeStack.Clear()
	p.nonInstantTaskStatus.Clear()
	p.conditionalReevaluate.Clear()
	p.conditionalReevaluateMap.Clear()

	return nil
}

func (p *BehaviorTree) ParseChildTask(task iface.ITask, parent iface.IParentTask, parentCompositeIndex int) error {
	index := p.taskList.Count()
	parentIndex := parent.ID()

	p.childrenIndex[parentIndex].Add(index)
	p.relativeChildIndex.Add(p.childrenIndex[parentIndex].Count() - 1)
	p.taskList.Add(task)
	p.parentIndex.Add(parent.ID())
	p.parentCompositeIndex.Add(parentCompositeIndex)
	p.childConditionalIndex.Add(util.NewList[int](10))
	p.childrenIndex.Add(util.NewList[int](10))

	task.SetID(index)
	task.SetParent(parent)
	task.SetOwner(p)

	/*
		if task.IsImplementsIAction() {
			action := task.(iface.IAction)
			if action.IsSyncToClient() {
				action.SetSyncDataCollector(iface.NewSyncDataCollector())
			}
		}
	*/

	if task.IsImplementsIParentTask() {
		if task.IsImplementsIComposite() {
			parentCompositeIndex = task.ID()
		}

		parentTask := task.(iface.IParentTask)
		for _, childTask := range parentTask.Children() {
			if err := p.ParseChildTask(childTask, parentTask, parentCompositeIndex); err != nil {
				return err
			}
		}
	} else {
		if task.IsImplementsIConditional() {
			if parentCompositeIndex != -1 {
				p.childConditionalIndex[parentCompositeIndex].Add(task.ID())
			}
		}
	}

	return nil
}

func (p *BehaviorTree) Disable() error {
	if p.isRunning {
		status := iface.Success
		for i := p.activeStack.Count() - 1; i >= 0; i-- {
			for p.activeStack[i].Len() > 0 {
				stackCount := p.activeStack[i].Len()
				status = p.PopTask(p.activeStack[i].Peak(), i, status, false)
				if stackCount == 1 {
					break
				}
			}
		}

		for _, task := range p.taskList {
			if !task.Disabled() {
				task.OnComplete()
			}
		}

		p.RemoveChildConditionalReevaluate(-1)

		//数据收集器解析掉
		for _, task := range p.taskList {
			if task.IsImplementsIAction() {
				action := task.(iface.IAction)
				if action.IsSyncToClient() {
					collector := action.SyncDataCollector()
					if collector != nil {
						collector.GetAndClear()
						action.SetSyncDataCollector(nil)
					}
				}
			}
		}

		p.executionStatus = status
		p.isRunning = false
		p.runtimeEventHandle.PostOnComplete(p, p.clock.TimesampInMill())
		return nil
	} else {
		return errors.New("not running")
	}
}

func (p *BehaviorTree) PopTask(taskIndex, stackIndex int, status iface.TaskStatus, popChildren bool) iface.TaskStatus {
	if !p.isRunning {
		return status
	}

	if stackIndex >= p.activeStack.Count() {
		return status
	}

	if p.activeStack[stackIndex].Len() == 0 || taskIndex != p.activeStack[stackIndex].Peak() {
		return status
	}

	p.activeStack[stackIndex].Pop()
	p.nonInstantTaskStatus[stackIndex] = iface.Inactive

	task := p.taskList[taskIndex]

	//	清理数据
	if task.IsImplementsIAction() {
		action := task.(iface.IAction)
		if action.IsSyncToClient() {
			action.SyncDataCollector().GetAndClear()
		}
	}

	task.OnEnd()

	parentIndex := p.parentIndex[taskIndex]
	if parentIndex != -1 {
		if task.IsImplementsIConditional() {
			compositeParentIndex := p.parentCompositeIndex[taskIndex]
			if compositeParentIndex != -1 {
				compositeTask := p.taskList[compositeParentIndex].(iface.IComposite)
				if compositeTask.AbortType() != iface.None {
					conditionalReevaluate := p.conditionalReevaluateMap[taskIndex]
					composite := -1
					if compositeTask.AbortType() != iface.LowerPriority {
						composite = compositeParentIndex
					}

					if conditionalReevaluate == nil {
						//index int, taskStatus iface.TaskStatus, compositeIndex int,	stackIndex int
						conditionalReevaluate = newConditionalReevaluate(taskIndex, status, composite)
						p.conditionalReevaluate.Add(conditionalReevaluate)
						p.conditionalReevaluateMap.Add(taskIndex, conditionalReevaluate)
					} else {
						conditionalReevaluate.Initialize(taskIndex, status, composite)
					}
				}
			}
		}

		parentTask := p.taskList[parentIndex].(iface.IParentTask)
		if !parentTask.CanRunParallelChildren() {
			parentTask.OnChildExecuted1(status)
			status = parentTask.Decorate(status)
		} else {
			parentTask.OnChildExecuted2(p.relativeChildIndex[taskIndex], status)
		}
	}

	if task.IsImplementsIParentTask() {
		if task.IsImplementsIComposite() {
			compositeTask := task.(iface.IComposite)
			if compositeTask.AbortType() == iface.Self || compositeTask.AbortType() == iface.None {
				p.RemoveChildConditionalReevaluate(taskIndex)
			} else if compositeTask.AbortType() == iface.Both || compositeTask.AbortType() == iface.LowerPriority {
				for _, conditionalReevaluate := range p.conditionalReevaluate {
					if p.IsParentTask(taskIndex, conditionalReevaluate.index) {
						conditionalReevaluate.compositeIndex = p.parentCompositeIndex[taskIndex]
					}
				}
			}
		}
	}

	if popChildren {
		for i := p.activeStack.Count() - 1; i > stackIndex; i-- {
			stack := p.activeStack[i]
			for i < p.activeStack.Count() && stack == p.activeStack[i] && p.activeStack[i].Len() != 0 {
				if p.IsParentTask(taskIndex, p.activeStack[i].Peak()) {
					childStatus := iface.Failure
					p.PopTask(p.activeStack[i].Peak(), i, childStatus, false)
				} else {
					break
				}
			}
		}
	}

	taskRuntimeData := p.taskDatas[task.ID()]
	stackData := p.getStackRuntimeData(stackIndex)
	nowTimestamp := p.clock.TimesampInMill()
	p.runtimeEventHandle.PostOnEnd(p, taskRuntimeData, stackData, task, nowTimestamp)

	if task.IsImplementsIAction() {
		action := task.(iface.IAction)
		if action.IsSyncToClient() {
			datas := action.SyncDataCollector().GetAndClear()
			p.runtimeEventHandle.ActionPostOnEnd(p, taskRuntimeData, stackData, task, nowTimestamp, datas)
		}
	}

	//	删除任务运行时数据
	if task.IsImplementsIParentTask() {
		parentTask := task.(iface.IParentTask)
		if parentTask.CanRunParallelChildren() {
			p.runtimeEventHandle.ParallelPostOnEnd(p, taskRuntimeData, stackData, task, nowTimestamp)
			delete(p.parallelTaskID2StackIDs, task.ID())
		}
	}
	delete(p.taskDatas, task.ID())

	if p.activeStack[stackIndex].Len() == 0 {
		if stackIndex == 0 {
			p.RemoveStack(stackIndex)
			p.Disable()
			p.executionStatus = status
			status = iface.Inactive
		} else {
			p.RemoveStack(stackIndex)
			status = iface.Running
		}
	}

	return status
}

func (p *BehaviorTree) nextStackID() int {
	stackID := p.stackID
	p.stackID++
	return stackID
}

/* var globaTaskExecuteID = int64(1)
var globaTaskExecuteIDMutex sync.Mutex */

func (p *BehaviorTree) nextTaskExecuteID() int {
	return p.clock.NextTaskExecuteID()

	/* 	globaTaskExecuteIDMutex.Lock()
	   	defer globaTaskExecuteIDMutex.Unlock()
	   	taskExecuteID := globaTaskExecuteID
	   	globaTaskExecuteID++
	   	return taskExecuteID */
	/*
		 	taskExecuteID := p.taskExecuteID
			p.taskExecuteID++

			return taskExecuteID
	*/

}

func (p *BehaviorTree) AddStack() int {
	stackIndex := p.activeStack.Count()
	stack := util.NewStackPtr[int](10)
	p.activeStack.Add(stack)
	p.nonInstantTaskStatus.Add(iface.Inactive)

	stackID := p.nextStackID()
	timestampInMill := p.clock.TimesampInMill()
	stackData := iface.NewStackRuntimeData(stackID, timestampInMill)
	p.runtimeEventHandle.NewStack(p, stackData)
	p.stackDatas[stack] = stackData
	p.stackID2StackData[stackID] = stackData

	return stackIndex
}

func (p *BehaviorTree) RemoveStack(stackIndex int) {
	if stackIndex < p.activeStack.Count() {
		stack := p.activeStack[stackIndex]
		stackData := p.stackDatas[stack]
		nowTimesampInMill := p.clock.TimesampInMill()
		if _, ok := p.stackID2ParallelTaskID[stackData.StackID]; ok {
			parallelTaskID := p.stackID2ParallelTaskID[stackData.StackID]
			taskRunTimeData := p.taskDatas[parallelTaskID]
			parentStackData := p.stackID2StackData[taskRunTimeData.ActiveStackID]
			task := p.taskList[taskRunTimeData.TaskID]
			p.runtimeEventHandle.ParallelRemoveChildStack(p, taskRunTimeData, parentStackData, task, stackData, nowTimesampInMill)

			delete(p.stackID2ParallelTaskID, stackData.StackID)
			oldParallelTaskID2StackIDs := p.parallelTaskID2StackIDs[task.ID()]
			p.parallelTaskID2StackIDs[task.ID()] = make([]int, 0)
			for _, childStackID := range oldParallelTaskID2StackIDs {
				if childStackID != stackData.StackID {
					p.parallelTaskID2StackIDs[task.ID()] = append(p.parallelTaskID2StackIDs[task.ID()], childStackID)
				}
			}
		}

		p.runtimeEventHandle.RemoveStack(p, stackData, nowTimesampInMill)
		delete(p.stackID2StackData, stackData.StackID)
		delete(p.stackDatas, stack)
		p.activeStack.RemoveAt(stackIndex)
		p.nonInstantTaskStatus.RemoveAt(stackIndex)
	}
}

func (p *BehaviorTree) RemoveChildConditionalReevaluate(compositeIndex int) {
	/*
		for i := p.conditionalReevaluate.Count() - 1; i > -1; i-- {
			if p.conditionalReevaluate[i].compositeIndex == compositeIndex {
				conditionalIndex := p.conditionalReevaluate[i].index
				p.conditionalReevaluateMap.Remove(conditionalIndex)
				p.conditionalReevaluate.RemoveAt(i)
			}
		}
	*/

	for i := p.conditionalReevaluate.Count() - 1; i > -1; i-- {
		if p.IsParentTask(compositeIndex, p.conditionalReevaluate[i].index) {
			conditionalIndex := p.conditionalReevaluate[i].index
			p.conditionalReevaluateMap.Remove(conditionalIndex)
			p.conditionalReevaluate.RemoveAt(i)
		}
	}
}

func (p *BehaviorTree) IsParentTask(possibleParent, possibleChild int) bool {
	parentIndex := 0
	childIndex := possibleChild

	for childIndex != -1 {
		parentIndex = p.parentIndex[childIndex]
		if parentIndex == possibleParent {
			return true
		}

		childIndex = parentIndex
	}

	return false
}

func (p *BehaviorTree) ReevaluateConditionalTasks() {
	updateConditionIndexes := util.NewList[*ConditionalReevaluate](10)

	for i := p.conditionalReevaluate.Count() - 1; i > -1; i-- {
		conditionalReevaluate := p.conditionalReevaluate[i]
		if conditionalReevaluate.compositeIndex != -1 {
			conditionalIndex := conditionalReevaluate.index
			conditionalStatus := p.taskList[conditionalIndex].OnUpdate()
			if conditionalStatus != conditionalReevaluate.taskStatus {
				compositeIndex := conditionalReevaluate.compositeIndex
				for j := p.activeStack.Count() - 1; j > -1; j-- {
					if p.activeStack[j].Len() > 0 {
						taskIndex := p.activeStack[j].Peak()
						if !p.IsParentTask(compositeIndex, taskIndex) {
							continue
						}

						stackCount := p.activeStack.Count()
						for taskIndex != -1 && taskIndex != compositeIndex && p.activeStack.Count() == stackCount {
							status := iface.Failure
							p.PopTask(taskIndex, j, status, false)
							taskIndex = p.parentIndex[taskIndex]
						}
					}
				}

				for j := p.conditionalReevaluate.Count() - 1; j > i; j-- {
					jConditionalReval := p.conditionalReevaluate[j]
					if p.IsParentTask(compositeIndex, jConditionalReval.index) {
						p.conditionalReevaluateMap.Remove(jConditionalReval.index)
						p.conditionalReevaluate.RemoveAt(j)
					}
				}

				//	原先abort过的要设置为原位
				for i := updateConditionIndexes.Count() - 1; i > -1; i-- {
					jConditionalReval := updateConditionIndexes[i]
					if p.IsParentTask(compositeIndex, jConditionalReval.index) {
						taskIndex := p.parentIndex[jConditionalReval.index]
						for taskIndex != -1 && taskIndex != jConditionalReval.compositeIndex {
							task := p.taskList[taskIndex].(iface.IParentTask)
							task.OnCancelConditionalAbort()
							taskIndex = p.parentIndex[taskIndex]
						}
						updateConditionIndexes.RemoveAt(i)
					}
				}

				updateConditionIndexes.Add(conditionalReevaluate)
				//是否需要把当前的conditionalReevaluate也删除掉？需要
				p.conditionalReevaluateMap.Remove(conditionalIndex)
				p.conditionalReevaluate.RemoveAt(i)
				/*
					for j := i - 1; j > -1; j-- {
						jConditionalReval := p.conditionalReevaluate[j]
						if jConditionalReval.compositeIndex == compositeIndex {
							commonCompositeIndex := p.FindLCA(jConditionalReval.index, conditionalIndex)
							if commonCompositeIndex != compositeIndex {
								jConditionalReval.compositeIndex = commonCompositeIndex
							}
						}
					}
				*/
				conditionalParentIndexes := util.NewList[int](10)
				parentIndex := conditionalIndex
				for {
					parentIndex = p.parentIndex[parentIndex]
					conditionalParentIndexes.Add(parentIndex)
					if parentIndex == compositeIndex {
						break
					}
				}

				for j := conditionalParentIndexes.Count() - 1; j > -1; j-- {
					parentTask := p.taskList[conditionalParentIndexes[j]].(iface.IParentTask)
					if j == 0 {
						parentTask.OnConditionalAbort(p.relativeChildIndex[conditionalIndex])
					} else {
						parentTask.OnConditionalAbort(p.relativeChildIndex[conditionalParentIndexes[j-1]])
					}
				}
			}
		}
	}
}

func (p *BehaviorTree) FindLCA(taskindex1, taskIndex2 int) int {
	set := util.NewHashSet[int]()
	parentIndex := p.parentCompositeIndex[taskindex1]

	for {
		set.Add(parentIndex)
		if parentIndex == -1 {
			break
		}
		parentIndex = p.parentCompositeIndex[parentIndex]
	}

	parentIndex = p.parentCompositeIndex[taskIndex2]
	for {
		if set.Contains(parentIndex) {
			return parentIndex
		}

		if parentIndex == -1 {
			break
		}

		parentIndex = p.parentCompositeIndex[parentIndex]
	}

	return -1
}

func (p *BehaviorTree) Update() {
	if p.isRunning {
		if p.initializeFirstStackAndFirstTask {
			p.AddStack()
			p.PushTask(0, 0)
			p.initializeFirstStackAndFirstTask = false
		}

		p.ReevaluateConditionalTasks()

		for j := p.activeStack.Count() - 1; j > -1; j-- {
			status := iface.Inactive
			startIndex := -1
			taskIndex := 0

			//	通过判断当前位置上的队列是否是同一个
			currentStack := p.activeStack[j]
			for status != iface.Running && j < p.activeStack.Count() && p.activeStack[j].Len() > 0 && currentStack == p.activeStack[j] {
				taskIndex = p.activeStack[j].Peak()
				if !p.isRunning {
					break
				}

				if p.activeStack[j].Len() > 0 && startIndex == p.activeStack[j].Peak() {
					break
				}

				startIndex = taskIndex
				status = p.RunTask(taskIndex, j, status)
			}
		}
	}
}

func (p *BehaviorTree) RunTask(taskIndex, stackIndex int, previousStatus iface.TaskStatus) iface.TaskStatus {
	if taskIndex >= p.taskList.Count() {
		return previousStatus
	}

	task := p.taskList[taskIndex]
	if task.Disabled() {
		parentIndex := p.parentIndex[taskIndex]
		if parentIndex != -1 {
			parentTask := p.taskList[parentIndex].(iface.IParentTask)
			if !parentTask.CanRunParallelChildren() {
				parentTask.OnChildExecuted1(iface.Inactive)
			} else {
				parentTask.OnChildExecuted2(p.relativeChildIndex[taskIndex], iface.Inactive)
			}
		}

		status := iface.Success
		if p.activeStack[stackIndex].Len() == 0 {
			if stackIndex == 0 {
				p.RemoveStack(stackIndex)
				p.Disable()
				p.executionStatus = status
				status = iface.Inactive
			} else {
				p.RemoveStack(stackIndex)
				status = iface.Running
			}
		}

		return status
	}

	status := previousStatus
	if !task.IsInstant() && (p.nonInstantTaskStatus[stackIndex] == iface.Failure || p.nonInstantTaskStatus[stackIndex] == iface.Success) {
		status = p.nonInstantTaskStatus[stackIndex]
		status = p.PopTask(taskIndex, stackIndex, status, true)
		return status
	}

	p.PushTask(taskIndex, stackIndex)
	if task.IsImplementsIParentTask() {
		status, stackIndex = p.RunParentTask(taskIndex, stackIndex, status)
		parentTask := task.(iface.IParentTask)
		status = parentTask.OverrideStatus1(status)
	} else {
		//	清理同步数据
		if task.IsImplementsIAction() {
			action := task.(iface.IAction)
			if action.IsSyncToClient() {
				action.SyncDataCollector().GetAndClear()
			}
		}
		status = task.OnUpdate()
	}

	taskRunTimeData := p.taskDatas[taskIndex]
	stack := p.activeStack[stackIndex]
	stackRuntimeData := p.stackDatas[stack]
	nowTimesampInMill := p.clock.TimesampInMill()
	p.runtimeEventHandle.PostOnUpdate(p, taskRunTimeData, stackRuntimeData, task, nowTimesampInMill, status)

	if task.IsImplementsIAction() {
		action := task.(iface.IAction)
		if action.IsSyncToClient() {
			datas := action.SyncDataCollector().GetAndClear()
			p.runtimeEventHandle.ActionPostOnUpdate(p, taskRunTimeData, stackRuntimeData, task, nowTimesampInMill, status, datas)
		}
	}

	if status != iface.Running {
		if task.IsInstant() {
			status = p.PopTask(taskIndex, stackIndex, status, true)
		} else {
			p.nonInstantTaskStatus[stackIndex] = status
			status = iface.Running
		}
	}

	return status
}

func (p *BehaviorTree) getStackRuntimeData(stackIndex int) *iface.StackRuntimeData {
	stack := p.activeStack[stackIndex]
	return p.stackDatas[stack]
}

func (p *BehaviorTree) RunParentTask(taskIndex, stackIndex int, status iface.TaskStatus) (iface.TaskStatus, int) {
	parentTask := p.taskList[taskIndex].(iface.IParentTask)
	if !parentTask.CanRunParallelChildren() || parentTask.OverrideStatus1(iface.Running) != iface.Running {
		childStatus := iface.Inactive
		parentStack := stackIndex
		parentStackRunTimeData := p.getStackRuntimeData(stackIndex)
		taskRuntimeData := p.taskDatas[taskIndex]
		childrenIndexes := p.childrenIndex[taskIndex]

		for parentTask.CanExecute() && (childStatus != iface.Running || parentTask.CanRunParallelChildren()) && p.isRunning {
			childIndex := parentTask.CurrentChildIndex()

			if parentTask.CanRunParallelChildren() {
				/*
					p.activeStack.Add(util.NewStackPtr[int](10))
					p.nonInstantTaskStatus.Add(iface.Inactive)
					stackIndex = p.activeStack.Count() - 1
				*/
				stackIndex = p.AddStack()
				childStackRunTimeData := p.getStackRuntimeData(stackIndex)
				p.stackID2ParallelTaskID[childStackRunTimeData.StackID] = parentTask.ID()
				p.parallelTaskID2StackIDs[parentTask.ID()] = append(p.parallelTaskID2StackIDs[parentTask.ID()], childStackRunTimeData.StackID)
				p.runtimeEventHandle.ParallelAddChildStack(p, taskRuntimeData, parentStackRunTimeData, parentTask, childStackRunTimeData)

				parentTask.OnChildStarted1(childIndex)
			} else {
				parentTask.OnChildStarted0()
			}

			childStatus = p.RunTask(childrenIndexes[childIndex], stackIndex, status)
			status = childStatus
		}
		stackIndex = parentStack
	}

	return status, stackIndex
}
