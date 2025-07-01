package base

type Conditional struct {
	task
}

func (p *Conditional) IsConditional() bool {
	return true
}

func (p *Conditional) DebugInfo() map[string]interface{} {
	info := p.task.DebugInfo()
	coreSlice := info["core"].([]interface{})
	coreSlice = append(coreSlice, "Conditional")
	info["core"] = coreSlice
	return info
}

func (p *Conditional) IsImplementsIConditional() bool {
	return true
}
