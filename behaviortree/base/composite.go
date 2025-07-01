package base

import "github.com/hushuoyouli/real-time-sync/behaviortree/iface"

type Composite struct {
	parentTask
	abortType iface.AbortType
}

func (p *Composite) AbortType() iface.AbortType {
	return p.abortType
}
func (p *Composite) SetAbortType(abortType iface.AbortType) {
	p.abortType = abortType
}

func (p *Composite) IsComposite() bool {
	return true
}

func (p *Composite) IsImplementsIComposite() bool {
	return true
}

func (p *Composite) DebugInfo() map[string]interface{} {
	info := p.parentTask.DebugInfo()
	coreSlice := info["core"].([]interface{})
	coreSlice = append(coreSlice, "Composite")
	coreSlice = append(coreSlice, p.abortType.ToString())
	info["core"] = coreSlice
	return info
}
