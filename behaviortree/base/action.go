package base

import "github.com/hushuoyouli/real-time-sync/behaviortree/iface"

type Action struct {
	task
	collector *iface.SyncDataCollector
}

func (p *Action) IsAction() bool {
	return true
}

func (p *Action) DebugInfo() map[string]interface{} {
	info := p.task.DebugInfo()

	coreSlice := info["core"].([]interface{})
	coreSlice = append(coreSlice, "Action")
	info["core"] = coreSlice

	return info
}

func (p *Action) IsSyncToClient() bool {
	return false
}

func (p *Action) SendSyncData(data []byte) {
	if p.collector != nil {
		p.collector.AddData(data)
	}
}

func (p *Action) RebuildSyncDatas() {

}

func (p *Action) SetSyncDataCollector(collector *iface.SyncDataCollector) {
	p.collector = collector
}

func (p *Action) SyncDataCollector() *iface.SyncDataCollector {
	return p.collector
}
func (p *Action) IsImplementsIAction() bool {
	return true
}
