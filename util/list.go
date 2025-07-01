package util

/*
NOTICE:这个类型不能拷贝使用,
比如a=b
a.Add(1)
b.Add(2)
因为实现机制的缘故，得到的不一定是你需要的结果
如果想要更改后还是一致的，可以使用指针
a = &b
a.Add(1)
b.Add(2)
*/
type List[T interface{}] []T

func NewList[T interface{}](caps int) List[T] {
	a := make([]T, 0, caps)
	return (List[T])(a)
}

func NewListPointer[T interface{}](caps int) *List[T] {
	a := make([]T, 0, caps)
	return (*List[T])(&a)
}

func (p *List[T]) Add(value T) {
	AddToSlice((*[]T)(p), value)
}

func (p *List[T]) Count() int {
	return len(*p)
}

func (p *List[T]) RemoveAt(index int) T {
	if index >= 0 && index < len(*p) {
		e := (*p)[index]
		RemoveFromSlice((*[]T)(p), index)
		return e
	}

	return *new(T)
}

func (p *List[T]) Clear() {
	(*p) = (*p)[0:0]
}
