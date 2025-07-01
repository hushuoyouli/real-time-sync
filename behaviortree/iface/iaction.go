package iface

type IAction interface {
	ITask
	IsAction() bool

	//	是否需要同步到客户端
	IsSyncToClient() bool
	SendSyncData(data []byte)
	RebuildSyncDatas()
	SetSyncDataCollector(collector *SyncDataCollector)
	SyncDataCollector() *SyncDataCollector
}
