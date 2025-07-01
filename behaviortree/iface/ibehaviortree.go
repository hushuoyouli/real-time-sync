package iface

type IBehaviorTree interface {
	ID() int64

	Enable() error
	Disable() error
	Update()
	IsRunning() bool

	Unit() IUnit
	RebuildSync(collector IRebuildSyncDataCollector)

	Clock() IClock

	ExtraParam() interface{}
}
