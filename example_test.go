package genmap_test

import (
	"fmt"
	"testing"

	"github.com/ronanh/genmap"
)

type MyKey struct {
	k1 int
	k2 []string
}

type MyValue struct {
	v1 int
	v2 string
}

func NewMyKeyHasher() func(k MyKey) uint64 {
	var fieldHashes [2]uint64
	k1Hasher := genmap.NewHasher[int]()
	k2Hasher := genmap.NewHasher[string]()
	return func(k MyKey) uint64 {
		fieldHashes[0] = k1Hasher(k.k1)
		fieldHashes[1] = genmap.HashSeed
		for _, s := range k.k2 {
			fieldHashes[1] = genmap.CombineHash(fieldHashes[1], k2Hasher(s))
		}
		return genmap.CombineHashes(fieldHashes[:]...)
	}
}

func MyKeyEquals(a, b MyKey) bool {
	if a.k1 != b.k1 {
		return false
	}
	if len(a.k2) != len(b.k2) {
		return false
	}
	for i := range a.k2 {
		if a.k2[i] != b.k2[i] {
			return false
		}
	}
	return true
}

func TestMain(t *testing.T) {
	m := genmap.NewMap[MyKey, MyValue](MyKeyEquals, NewMyKeyHasher())
	m.Put(MyKey{1, []string{"a", "b"}}, MyValue{1, "a"})
	m.Put(MyKey{2, []string{"c", "d"}}, MyValue{2, "b"})
	m.Put(MyKey{1, []string{"a", "b"}}, MyValue{3, "c"})

	// increment the value for a key
	m.Upsert(MyKey{1, []string{"a", "b"}}, func(v *MyValue, exists bool) {
		v.v1++
	})

	// Get the value for a key
	v, ok := m.Get(MyKey{1, []string{"a", "b"}})
	if ok {
		println(v.v1)
		// prints 4
	}

	// Iterate over the map
	it := m.Iterator()
	for it.Next() {
		fmt.Printf("Key: %v, Value: %v\n", it.Cur().Key, it.Cur().Value)
	}
	// prints:
	// Key: {1 [a b]}, Value: {4 c}
	// Key: {2 [c d]}, Value: {2 b}
}
