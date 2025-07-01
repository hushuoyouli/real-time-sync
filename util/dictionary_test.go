package util

import "testing"

func Test_Dictionary(t *testing.T) {
	var b Dictionary[int, int]
	b.Clear()

	a := NewDictionary[int, int]()
	a.Add(100, 200)
	c := a
	c.Add(200, 300)

	t.Log(c[100])
	t.Log(c[200])
}
