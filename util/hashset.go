package util

type HashSet[T comparable] map[T]bool

func NewHashSet[T comparable]() HashSet[T] {
	return make(HashSet[T])
}

func (p *HashSet[T]) Add(obj T) {
	(*p)[obj] = true
}

func (p *HashSet[T]) Clear() {
	for k := range *p {
		delete(*p, k)
	}
}

func (p *HashSet[T]) Contains(obj T) bool {
	_, ok := (*p)[obj]
	return ok
}

func (p *HashSet[T]) Remove(key T) {
	delete(*p, key)
}
