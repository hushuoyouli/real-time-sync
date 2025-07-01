package iface

type TaskStatus int

const (
	Inactive TaskStatus = iota
	Running
	Success
	Failure
)

func (status TaskStatus) ToString() string {
	switch status {
	case Inactive:
		return "Inactive"
	case Running:
		return "Running"
	case Success:
		return "Success"
	case Failure:
		return "Failure"
	}

	return "ErrorTaskStatus"
}

type AbortType int

const (
	None AbortType = iota
	Self
	LowerPriority
	Both
)

func (val AbortType) ToString() string {
	switch val {
	case None:
		return "None"
	case Self:
		return "Self"
	case LowerPriority:
		return "LowerPriority"
	case Both:
		return "Both"
	default:
		return "Error"
	}
}
