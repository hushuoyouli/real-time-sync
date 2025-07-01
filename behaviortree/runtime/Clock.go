package runtime

import (
	"time"
)

type Clock struct {
	taskExecuteID int
}

func NewClock() *Clock {
	return &Clock{
		taskExecuteID: 1,
	}
}

func (p *Clock) TimesampInMill() int64 {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return timestamp
}

func (p *Clock) NextTaskExecuteID() int {
	taskExecuteID := p.taskExecuteID
	p.taskExecuteID++

	return taskExecuteID
}
