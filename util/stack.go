package util

type Stack[T interface{}] []T

func NewStack[T interface{}](caps int) Stack[T] {
	a := make(Stack[T], 0, caps)

	return a
}

func NewStackPtr[T interface{}](caps int) *Stack[T] {
	a := make(Stack[T], 0, caps)

	return &a
}

func (p *Stack[T]) Push(value T) {
	AddToSlice((*[]T)(p), value)
}

func (p *Stack[T]) Pop() T {
	if len(*p) > 0 {
		lastIndex := len(*p) - 1
		e := (*p)[lastIndex]
		RemoveFromSlice((*[]T)(p), lastIndex)
		return e
	}

	return *new(T)
}

func (p *Stack[T]) Peak() T {
	if len(*p) > 0 {
		return (*p)[len(*p)-1]
	}

	return *new(T)
}

func (p *Stack[T]) Len() int {
	return len(*p)
}

func (p *Stack[T]) Empty() bool {
	return p.Len() == 0
}

func (p *Stack[T]) RemoveAt(index int) T {
	if index >= 0 && index < len(*p) {
		e := (*p)[index]
		RemoveFromSlice((*[]T)(p), index)
		return e
	}

	return *new(T)
}
