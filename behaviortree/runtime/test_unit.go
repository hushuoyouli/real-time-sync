package runtime

import "github.com/hushuoyouli/real-time-sync/rlog"

type TestUnit struct {
	log  rlog.ILogger
	name string
}

func NewTestUnit(name string) *TestUnit {
	return &TestUnit{
		log:  &rlog.SLogger{},
		name: name,
	}
}

func (p *TestUnit) Name() string {
	return p.name
}

func (p *TestUnit) ID() int64 {
	return 0
}

func (p *TestUnit) Log() rlog.ILogger {
	if p.log == nil {
		p.log = &rlog.SLogger{}
	}
	return p.log
}
