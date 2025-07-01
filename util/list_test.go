package util

import (
	"fmt"
	"testing"
)

func TestList(t *testing.T) {
	var b List[int]
	b.Clear()

	l := NewList[int](1)
	l.Add(100)
	l.Add(200)
	l.Add(300)
	l.Add(400)
	l.Add(500)

	fmt.Println(l.Count())
	fmt.Println(l)
	fmt.Println(l[0])

	fmt.Println(l.RemoveAt(0))
	fmt.Println(l)
}
