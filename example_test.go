package genmap_test

import (
	"fmt"

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
	k1Hasher := genmap.NewHasher[int]()
	k2Hasher := genmap.NewHasher[string]()
	return func(k MyKey) uint64 {
		// local to each call, so that concurrent Gets can share the hasher
		var fieldHashes [2]uint64
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

func Example() {
	m := genmap.NewMap[MyKey, MyValue](MyKeyEquals, NewMyKeyHasher())
	m.Put(MyKey{1, []string{"a", "b"}}, MyValue{1, "a"})
	m.Put(MyKey{2, []string{"c", "d"}}, MyValue{2, "b"})
	m.Put(MyKey{1, []string{"a", "b"}}, MyValue{3, "c"})

	// increment the value for a key
	m.Upsert(MyKey{1, []string{"a", "b"}}, func(elem *genmap.MapElement[MyKey, MyValue], exists bool) {
		elem.Value.v1++
	})

	// Get the value for a key
	if v, ok := m.Get(MyKey{1, []string{"a", "b"}}); ok {
		fmt.Println(v.v1)
	}

	// Iterate over the map
	it := m.Iterator()
	for it.Next() {
		fmt.Printf("Key: %v, Value: %v\n", it.Cur().Key, it.Cur().Value)
	}
	// Unordered output:
	// 4
	// Key: {1 [a b]}, Value: {4 c}
	// Key: {2 [c d]}, Value: {2 b}
}

func ExampleMap_Entry() {
	m := genmap.NewMap[string, int](genmap.Equal[string], genmap.NewHasher[string]())
	m.Put("foo", 42)

	// Look the key up once, then update its element in place.
	foo := m.Entry("foo")
	if foo.Exists() {
		foo.OrDefault().MutateWith(func(elem *genmap.MapElement[string, int]) {
			elem.Value++
		})
	}

	// OrDefault inserts a missing key with a zero value.
	bar := m.Entry("bar")
	bar.OrDefault().MutateWith(func(elem *genmap.MapElement[string, int]) {
		elem.Value += 10
	})

	fooValue, _ := m.Get("foo")
	barValue, _ := m.Get("bar")
	fmt.Println(fooValue, barValue, m.Len())
	// Output: 43 10 2
}
