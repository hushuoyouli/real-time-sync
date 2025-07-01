package iface

type IClock interface {
	TimesampInMill() int64
	NextTaskExecuteID() int
}
