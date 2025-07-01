package iface

import (
	"github.com/hushuoyouli/real-time-sync/rlog"
)

type IUnit interface {
	ID() int64
	Log() rlog.ILogger
}
