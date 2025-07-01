package util

type Dictionary[KeyType comparable, ValueType interface{}] map[KeyType]ValueType

func NewDictionary[KeyType comparable, ValueType interface{}]() Dictionary[KeyType, ValueType] {
	return make(Dictionary[KeyType, ValueType])
}

func (p *Dictionary[KeyType, ValueType]) Clear() {
	for k := range *p {
		delete(*p, k)
	}
}

func (p *Dictionary[KeyType, ValueType]) Remove(key KeyType) {
	delete(*p, key)
}

func (p *Dictionary[KeyType, ValueType]) TryGetValue(key KeyType, valPtr *ValueType) bool {
	val, ok := (*p)[key]
	if ok {
		*valPtr = val
	}

	return ok
}

func (p *Dictionary[KeyType, ValueType]) Add(key KeyType, value ValueType) {
	(*p)[key] = value
}

func (p *Dictionary[KeyType, ValueType]) Contains(obj KeyType) bool {
	_, ok := (*p)[obj]
	return ok
}
