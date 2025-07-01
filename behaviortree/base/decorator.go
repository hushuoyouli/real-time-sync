package base

type Decorator struct {
	parentTask
}

func (p *Decorator) MaxChildren() int {
	return 1
}

func (p *Decorator) IsDecorator() bool {
	return true
}

func (p *Decorator) DebugInfo() map[string]interface{} {
	info := p.parentTask.DebugInfo()
	coreSlice := info["core"].([]interface{})
	coreSlice = append(coreSlice, "Decorator")
	info["core"] = coreSlice
	return info
}

func (p *Decorator) IsImplementsIDecorator() bool {
	return true
}
