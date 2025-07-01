package iface

type StackRuntimeData struct {
	StackID   int
	StartTime int64
}

func NewStackRuntimeData(StackID int, StartTime int64) *StackRuntimeData {
	return &StackRuntimeData{
		StackID:   StackID,
		StartTime: StartTime,
	}
}
