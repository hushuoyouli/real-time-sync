package util

import "testing"

func TestStack(t *testing.T) {
	intStack := NewStackPtr[int](0)

	t.Logf("%p\n", intStack)
	intStack.Push(1)
	t.Logf("%p\n", intStack)
	intStack.Push(2)
	t.Logf("%p\n", intStack)
	intStack.Push(3)
	t.Logf("%p\n", intStack)
	intStack.Push(4)
	t.Logf("%p\n", intStack)
	intStack.Push(5)
	t.Logf("%p\n", intStack)

	t.Log(intStack)
	t.Log(intStack.Peak())
	t.Log(intStack.Peak())

	t.Log("================================================")

	t.Log(intStack.Pop())
	t.Log(intStack.Pop())
	t.Log(intStack.Pop())
	t.Log(intStack.Pop())
	t.Log(intStack.Pop())
}
