package util

import "testing"

func Test_HashSet(t *testing.T) {
	var b HashSet[int]
	b.Clear()

	hashSet := NewHashSet[int]()
	for i := 0; i < 10; i++ {
		hashSet.Add(i)
	}

	t.Log(hashSet.Contains(1))
	t.Log(hashSet.Contains(11))

	hashSet.Remove(1)
	hashSet.Add(11)

	t.Log(hashSet.Contains(1))
	t.Log(hashSet.Contains(11))
}
