package iface

type TaskRuntimeData struct {
	TaskID        int
	StartTime     int64
	ExecuteID     int
	ActiveStackID int
}

func NewTaskRuntimeData(TaskID int, StartTime int64, ExecuteID int, ActiveStackID int) *TaskRuntimeData {
	return &TaskRuntimeData{
		TaskID:        TaskID,
		StartTime:     StartTime,
		ExecuteID:     ExecuteID,
		ActiveStackID: ActiveStackID,
	}
}
