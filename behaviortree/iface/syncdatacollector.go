package iface

type SyncDataCollector struct {
	datas [][]byte
}

func NewSyncDataCollector() *SyncDataCollector {
	return &SyncDataCollector{
		datas: make([][]byte, 0),
	}
}

func (p *SyncDataCollector) AddData(data []byte) {
	p.datas = append(p.datas, data)
}

func (p *SyncDataCollector) GetAndClear() [][]byte {
	datas := p.datas
	if len(p.datas) > 0 {
		p.datas = make([][]byte, 0)
	}

	return datas
}
