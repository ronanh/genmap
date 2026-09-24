package genmap_test

import (
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/ronanh/genmap"
)

func TestMap(t *testing.T) {
	m := genmap.NewMap[string, int](genmap.Equal[string], genmap.NewHasher[string]())
	if m.Len() != 0 {
		t.Errorf("expected empty map, got %d elements", m.Len())
	}

	m.Upsert("foo", func(elem *genmap.MapElement[string, int], exists bool) {
		elem.Value = 42
	})
	if m.Len() != 1 {
		t.Errorf("expected map with 1 element, got %d elements", m.Len())
	}
	if elem, ok := m.Get("foo"); !ok || elem != 42 {
		t.Errorf("expected element with key 'foo' and value 42, got %v", elem)
	}

	m.Upsert("foo", func(elem *genmap.MapElement[string, int], exists bool) {
		elem.Value = 43
	})
	if m.Len() != 1 {
		t.Errorf("expected map with 1 element, got %d elements", m.Len())
	}
	if elem, ok := m.Get("foo"); !ok || elem != 43 {
		t.Errorf("expected element with key 'foo' and value 43, got %v", elem)
	}

	m.Upsert("bar", func(elem *genmap.MapElement[string, int], exists bool) {
		elem.Value = 44
	})
	if m.Len() != 2 {
		t.Errorf("expected map with 2 elements, got %d elements", m.Len())
	}
	if elem, ok := m.Get("bar"); !ok || elem != 44 {
		t.Errorf("expected element with key 'bar' and value 44, got %v", elem)
	}

	elem, _ := m.Remove("foo")
	if m.Len() != 1 {
		t.Errorf("expected map with 1 element, got %d elements", m.Len())
	}
	if elem.Value != 43 {
		t.Errorf("expected removed element with value 43, got %v", elem)
	}
	if v, ok := m.Get("foo"); ok {
		t.Errorf("expected no element with key 'foo', got %v", v)
	}

	elem, _ = m.Remove("bar")
	if m.Len() != 0 {
		t.Errorf("expected empty map, got %d elements", m.Len())
	}
	if elem.Value != 44 {
		t.Errorf("expected removed element with value 44, got %v", elem)
	}
	if v, ok := m.Get("bar"); ok {
		t.Errorf("expected no element with key 'bar', got %v", v)
	}
}

func TestMapGet(t *testing.T) {
	m := genmap.NewMap[int, string](genmap.Equal[int], genmap.NewHasher[int]())
	m.Upsert(1, func(elem *genmap.MapElement[int, string], exists bool) {
		elem.Value = "one"
	})
	m.Upsert(2, func(elem *genmap.MapElement[int, string], exists bool) {
		elem.Value = "two"
	})
	m.Upsert(3, func(elem *genmap.MapElement[int, string], exists bool) {
		elem.Value = "three"
	})

	tests := []struct {
		name     string
		key      int
		expected string
	}{
		{
			name:     "existing key",
			key:      2,
			expected: "two",
		},
		{
			name:     "non-existent key",
			key:      4,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, ok := m.Get(tt.key)
			if !ok {
				if tt.expected != "" {
					t.Errorf("expected %q, but got nil", tt.expected)
				}
			} else if actual != tt.expected {
				t.Errorf("expected %q, but got %q", tt.expected, actual)
			}
		})
	}
}

func BenchmarkMapGet(b *testing.B) {
	m, keys := initMapAndKeys(100000, 64<<10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		me, ok := m.Get(keys[i%100000])
		if ok {
			_ = me
		}
	}
}

func BenchmarkStdMapGet(b *testing.B) {
	m, keys := initStdMapAndKeys(100000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if v, ok := m[keys[i%100000]]; ok {
			_ = v
		}
	}
}

func TestMapPut(t *testing.T) {
	m := genmap.NewMap[MyKey, MyValue](MyKeyEquals, NewMyKeyHasher())
	m.Put(MyKey{1, nil}, MyValue{1, "a"})
	m.Put(MyKey{2, nil}, MyValue{2, "b"})
	m.Put(MyKey{3, nil}, MyValue{3, "c"})
	if m.Len() != 3 {
		t.Errorf("expected length 3, got %d", m.Len())
	}
	if val, ok := m.Get(MyKey{1, nil}); !ok || !reflect.DeepEqual(val, (MyValue{1, "a"})) {
		t.Errorf("expected value 1 for key '1', got %v", val)
	}
	if val, ok := m.Get(MyKey{2, nil}); !ok || !reflect.DeepEqual(val, MyValue{2, "b"}) {
		t.Errorf("expected value 2 for key '2', got %v", val)
	}
	if val, ok := m.Get(MyKey{3, nil}); !ok || !reflect.DeepEqual(val, MyValue{3, "c"}) {
		t.Errorf("expected value 3 for key '3', got %v", val)
	}
}

func TestMapUpsert(t *testing.T) {
	m := genmap.NewMap[string, int](genmap.Equal[string], genmap.NewHasher[string]())
	m.Upsert("a", func(elem *genmap.MapElement[string, int], exists bool) {
		elem.Value = 1
	})
	m.Upsert("b", func(elem *genmap.MapElement[string, int], exists bool) {
		elem.Value = 2
	})
	m.Upsert("c", func(elem *genmap.MapElement[string, int], exists bool) {
		elem.Value = 3
	})
	if m.Len() != 3 {
		t.Errorf("expected length 3, got %d", m.Len())
	}
	if val, ok := m.Get("a"); !ok || val != 1 {
		t.Errorf("expected value 1 for key 'a', got %v", val)
	}
	if val, ok := m.Get("b"); !ok || val != 2 {
		t.Errorf("expected value 2 for key 'b', got %v", val)
	}
	if val, ok := m.Get("c"); !ok || val != 3 {
		t.Errorf("expected value 3 for key 'c', got %v", val)
	}
	m.Upsert("a", func(elem *genmap.MapElement[string, int], exists bool) {
		elem.Value = 4
	})
	if val, ok := m.Get("a"); !ok || val != 4 {
		t.Errorf("expected value 4 for key 'a', got %v", val)
	}
	if m.Len() != 3 {
		t.Errorf("expected length 3, got %d", m.Len())
	}
	if val, _ := m.Remove("b"); val.Value != 2 {
		t.Errorf("expected value 2 for removed key 'b', got %v", val)
	}
	if m.Len() != 2 {
		t.Errorf("expected length 2, got %d", m.Len())
	}
	if val, ok := m.Get("b"); ok {
		t.Errorf("expected nil value for removed key 'b', got %v", val)
	}
}

func TestMapEntry(t *testing.T) {
	m := genmap.NewMap[string, int](genmap.Equal[string], genmap.NewHasher[string]())
	m.Put("a", 1)
	increment := func(elem *genmap.MapElement[string, int]) { elem.Value++ }

	existing := m.Entry("a")
	if !existing.Exists() {
		t.Error("expected the entry for 'a' to exist")
	}
	existing.OrDefault().MutateWith(increment)

	missing := m.Entry("b")
	if missing.Exists() {
		t.Error("expected the entry for 'b' not to exist")
	}
	missing.OrDefault().MutateWith(increment)
	if !missing.Exists() {
		t.Error("expected the entry for 'b' to exist once OrDefault inserted it")
	}
	// the entry now refers to the inserted element: no second insertion
	missing.OrDefault().MutateWith(increment)

	assertStringMapContent(t, m, map[string]int{"a": 2, "b": 2})
}

func assertStringMapContent(t *testing.T, m *genmap.Map[string, int], want map[string]int) {
	t.Helper()
	if m.Len() != len(want) {
		t.Errorf("expected map with %d elements, got %d elements", len(want), m.Len())
	}
	got := make(map[string]int, len(want))
	it := m.Iterator()
	for it.Next() {
		if _, dup := got[it.Cur().Key]; dup {
			t.Errorf("iterator yielded key %q twice", it.Cur().Key)
		}
		got[it.Cur().Key] = it.Cur().Value
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected iterator to yield %v, got %v", want, got)
	}
}

func identityHash(k int) uint64 { return uint64(k) }

func TestMapPutCollidingKeys(t *testing.T) {
	// One bucket makes every key collide, so every Put after the first one
	// has to grow the bucket.
	m := genmap.NewMap[int, string](genmap.Equal[int], identityHash, 1)
	m.Put(1, "one")
	m.Put(2, "two")
	m.Put(3, "three")
	m.Put(4, "four")
	m.Put(5, "five")

	want := map[int]string{1: "one", 2: "two", 3: "three", 4: "four", 5: "five"}
	if m.Len() != len(want) {
		t.Errorf("expected map with %d elements, got %d elements", len(want), m.Len())
	}
	for k, v := range want {
		if got, ok := m.Get(k); !ok || got != v {
			t.Errorf("expected value %q for key %d, got %q (found: %v)", v, k, got, ok)
		}
	}
	got := make(map[int]string, len(want))
	it := m.Iterator()
	for it.Next() {
		got[it.Cur().Key] = it.Cur().Value
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected iterator to yield %v, got %v", want, got)
	}
}

func TestNilMap(t *testing.T) {
	// Like a nil built-in map, a nil Map reads as empty and ignores removals.
	var m *genmap.Map[int, int]
	if m.Len() != 0 {
		t.Errorf("expected empty map, got %d elements", m.Len())
	}
	if v, ok := m.Get(1); ok {
		t.Errorf("expected no element with key 1, got %v", v)
	}
	if elem, ok := m.Remove(1); ok {
		t.Errorf("expected nothing to remove for key 1, got %+v", elem)
	}
	m.Clear()
	it := m.Iterator()
	if it.Next() {
		t.Errorf("expected the iterator to yield nothing, got %+v", it.Cur())
	}
}

// TestMapReleasesRemovedValues checks that the map does not keep removed or
// cleared values reachable, which would stop the garbage collector from
// freeing them for as long as the map lives.
func TestMapReleasesRemovedValues(t *testing.T) {
	type value struct{ _ [64]byte }
	tests := []struct {
		name string
		drop func(m *genmap.Map[int, *value], v *value)
	}{
		{
			name: "removed",
			drop: func(m *genmap.Map[int, *value], v *value) {
				m.Put(0, v)
				m.Put(1, new(value))
				m.Remove(0)
			},
		},
		{
			name: "removed after its bucket outgrew 4 elements",
			drop: func(m *genmap.Map[int, *value], v *value) {
				m.Put(0, v)
				for k := 1; k <= 4; k++ {
					m.Put(k, new(value))
				}
				m.Remove(0)
			},
		},
		{
			name: "cleared",
			drop: func(m *genmap.Map[int, *value], v *value) {
				m.Put(0, v)
				m.Clear()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// One bucket, so all the keys share it.
			m := genmap.NewMap[int, *value](genmap.Equal[int], identityHash, 1)
			released := make(chan struct{})
			v := new(value)
			runtime.SetFinalizer(v, func(*value) { close(released) })
			tt.drop(m, v)
			v = nil

			if !isReleased(released) {
				t.Error("expected the value to be garbage collected, but the map still references it")
			}
			// the map must outlive the check: the leak goes through the map
			runtime.KeepAlive(m)
		})
	}
}

// isReleased runs the garbage collector until released is closed by a
// finalizer, and gives up after about a second.
func isReleased(released chan struct{}) bool {
	for i := 0; i < 100; i++ {
		runtime.GC()
		select {
		case <-released:
			return true
		case <-time.After(10 * time.Millisecond):
		}
	}
	return false
}

func TestMapWithZeroBucketSize(t *testing.T) {
	// Callers size maps from the expected number of keys, which can be 0.
	m := genmap.NewMap[int, int](genmap.Equal[int], identityHash, 0)
	m.Put(1, 10)
	m.Put(2, 20)
	assertMapContent(t, m, map[int]int{1: 10, 2: 20}, 3)
}

// TestMapMatchesBuiltinMap applies the same random operations to a Map and to
// a built-in map and checks after each step that both hold the same entries.
// Few buckets and a small key space force collisions, overwrites and removals,
// so bucket growth, shrinking and slice reuse all get exercised.
func TestMapMatchesBuiltinMap(t *testing.T) {
	const nbKeys, nbSteps = 64, 3000
	for _, nbBuckets := range []int{1, 3, 16} {
		t.Run(strconv.Itoa(nbBuckets)+"-buckets", func(t *testing.T) {
			rnd := rand.New(rand.NewSource(int64(nbBuckets)))
			m := genmap.NewMap[int, int](genmap.Equal[int], identityHash, nbBuckets)
			want := make(map[int]int)
			for step := 0; step < nbSteps; step++ {
				k := rnd.Intn(nbKeys)
				switch op := rnd.Intn(10); {
				case op < 4:
					m.Put(k, step)
					want[k] = step
				case op < 6:
					m.Upsert(k, func(elem *genmap.MapElement[int, int], exists bool) {
						elem.Value++
					})
					want[k]++
				case op < 9:
					elem, ok := m.Remove(k)
					wantV, wantOK := want[k]
					if ok != wantOK || ok && (elem.Key != k || elem.Value != wantV) {
						t.Fatalf("step %d: Remove(%d) = %+v, %v; want value %d, %v", step, k, elem, ok, wantV, wantOK)
					}
					delete(want, k)
				default:
					// remove about half of the entries through the iterator
					it := m.Iterator()
					for it.Next() {
						if rnd.Intn(2) == 0 {
							elem := it.Remove()
							if wantV, ok := want[elem.Key]; !ok || elem.Value != wantV {
								t.Fatalf("step %d: iterator removed %+v, want value %d (present: %v)", step, elem, wantV, ok)
							}
							delete(want, elem.Key)
						}
					}
				}
				assertMapContent(t, m, want, nbKeys)
				if t.Failed() {
					t.Fatalf("map diverged at step %d", step)
				}
			}
		})
	}
}

// assertMapContent checks that m holds exactly the entries of want, through
// Len, Get on every key in [0, nbKeys) and a full iteration.
func assertMapContent(t *testing.T, m *genmap.Map[int, int], want map[int]int, nbKeys int) {
	t.Helper()
	if m.Len() != len(want) {
		t.Errorf("expected map with %d elements, got %d elements", len(want), m.Len())
	}
	for k := 0; k < nbKeys; k++ {
		got, ok := m.Get(k)
		if wantV, wantOK := want[k]; ok != wantOK || got != wantV {
			t.Errorf("Get(%d) = %d, %v; want %d, %v", k, got, ok, wantV, wantOK)
		}
	}
	got := make(map[int]int, len(want))
	it := m.Iterator()
	for it.Next() {
		if _, dup := got[it.Cur().Key]; dup {
			t.Errorf("iterator yielded key %d twice", it.Cur().Key)
		}
		got[it.Cur().Key] = it.Cur().Value
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected iterator to yield %v, got %v", want, got)
	}
}

func BenchmarkMapPut100k(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := genmap.NewMap[int, MyValue](genmap.Equal[int], genmap.NewHasher[int](), 64<<10)
		for j := 0; j < 100000; j++ {
			m.Put(j, MyValue{j, "a"})
		}
	}
}

func BenchmarkStdMapPut100k(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := make(map[int]MyValue, 64<<10)
		for j := 0; j < 100000; j++ {
			m[j] = MyValue{j, "a"}
		}
	}
}

func BenchmarkMapPutOverwrite(b *testing.B) {
	m, keys := initMapAndKeys(100000, 64<<10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Put(keys[i%100000], MyValue{i, "a"})
	}
}

func BenchmarkStdMapPutOverwrite(b *testing.B) {
	m, keys := initStdMapAndKeys(100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m[keys[i%100000]] = MyValue{i, "a"}
	}
}

func BenchmarkMapUpsert100k(b *testing.B) {
	update := func(elem *genmap.MapElement[int, MyValue], exists bool) {
		elem.Value.v1++
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := genmap.NewMap[int, MyValue](genmap.Equal[int], genmap.NewHasher[int](), 64<<10)
		for j := 0; j < 100000; j++ {
			m.Upsert(j, update)
		}
	}
}

func BenchmarkStdMapUpsert100k(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := make(map[int]MyValue, 64<<10)
		for j := 0; j < 100000; j++ {
			v := m[j]
			v.v1++
			m[j] = v
		}
	}
}

func BenchmarkMapUpsertIncrement(b *testing.B) {
	m, keys := initMapAndKeys(100000, 64<<10)
	update := func(elem *genmap.MapElement[string, MyValue], exists bool) {
		elem.Value.v1++
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Upsert(keys[i%100000], update)
	}
}

func BenchmarkStdMapUpsertIncrement(b *testing.B) {
	m, keys := initStdMapAndKeys(100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := m[keys[i%100000]]
		v.v1++
		m[keys[i%100000]] = v
	}
}

func BenchmarkMapUpsertDelete(b *testing.B) {
	m, keys := initMapAndKeys(100000, 128<<10)
	update := func(elem *genmap.MapElement[string, MyValue], exists bool) {
		elem.Value.v1++
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%50 == 0 {
			m.Remove(keys[i%100000])
		} else {
			m.Upsert(keys[i%100000], update)
		}
	}
}

func BenchmarkStdMapUpsertDelete(b *testing.B) {
	m, keys := initStdMapAndKeys(100000)
	var j int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j++
		if i%50 == 0 {
			delete(m, keys[i%10000])
		} else {
			v := m[keys[i%100000]]
			v.v1++
			m[keys[i%100000]] = v
		}
	}
}

func initMapAndKeys(size, bucketsSize int) (*genmap.Map[string, MyValue], []string) {
	m := genmap.NewMap[string, MyValue](genmap.Equal[string], genmap.NewHasher[string](), bucketsSize)
	keys := make([]string, size)
	for i := 0; i < size; i++ {
		v := rand.Int()
		k := strconv.Itoa(v)
		m.Upsert(k, func(elem *genmap.MapElement[string, MyValue], exists bool) {
			elem.Value.v1 = v
		})
		keys[i] = k
	}
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	return m, keys
}

func initStdMapAndKeys(size int) (map[string]MyValue, []string) {
	m := make(map[string]MyValue)
	keys := make([]string, size)
	for i := 0; i < size; i++ {
		v := rand.Int()
		k := strconv.Itoa(v)
		m[k] = MyValue{v1: v}
		keys[i] = k
	}
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	return m, keys
}

func TestMapIterator(t *testing.T) {
	m := genmap.NewMap[int, string](genmap.Equal[int], genmap.NewHasher[int]())
	m.Upsert(1, func(elem *genmap.MapElement[int, string], exists bool) {
		elem.Value = "one"
	})
	m.Upsert(2, func(elem *genmap.MapElement[int, string], exists bool) {
		elem.Value = "two"
	})
	m.Upsert(3, func(elem *genmap.MapElement[int, string], exists bool) {
		elem.Value = "three"
	})

	if m.Len() != 3 {
		t.Errorf("expected map with 3 elements, got %d elements", m.Len())
	}

	var nbIt int
	chkValues := func(it *genmap.MapIterator[int, string]) {
		nbIt++
		switch it.Cur().Key {
		case 1:
			if it.Cur().Value != "one" {
				t.Errorf("Expected Cur() to return {1, \"one\"}")
			}
		case 2:
			if it.Cur().Value != "two" {
				t.Errorf("Expected Cur() to return {2, \"two\"}")
			}
		case 3:
			if it.Cur().Value != "three" {
				t.Errorf("Expected Cur() to return {3, \"three\"}")
			}
		default:
			t.Errorf("Unexpected key %d", it.Cur().Key)
		}
	}

	it := m.Iterator()
	if it.Next() != true {
		t.Errorf("Expected Next() to return true")
	}
	chkValues(it)

	if it.Next() != true {
		t.Errorf("Expected Next() to return true")
	}
	chkValues(it)

	if it.Next() != true {
		t.Errorf("Expected Next() to return true")
	}
	chkValues(it)

	if it.Next() != false {
		t.Errorf("Expected Next() to return false")
	}

	if nbIt != 3 {
		t.Errorf("Expected 3 iterations, got %d", nbIt)
	}

	nbIt = 0
	it.Reset()
	if it.Next() != true {
		t.Errorf("Expected Next() to return true")
	}
	chkValues(it)

	v := *it.Cur()
	if it.Remove().Key != v.Key {
		t.Errorf("Expected Remove() to return %v", v)
	}

	if it.Next() != true {
		t.Errorf("Expected Next() to return true")
	}
	chkValues(it)

	v = *it.Cur()
	if it.Remove().Key != v.Key {
		t.Errorf("Expected Remove() to return %v", v)
	}

	if it.Next() != true {
		t.Errorf("Expected Next() to return true")
	}
	chkValues(it)

	v = *it.Cur()
	if it.Remove().Key != v.Key {
		t.Errorf("Expected Remove() to return %v", v)
	}

	if it.Next() != false {
		t.Errorf("Expected Next() to return false")
	}
	if nbIt != 3 {
		t.Errorf("Expected 3 iterations, got %d", nbIt)
	}
	if m.Len() != 0 {
		t.Errorf("Expected empty map, got %d elements", m.Len())
	}
}

// TestMapIteratorAfterBucketShrink pins the invariant that the iterator yields
// exactly the elements the map holds, no more. Removing enough elements from a
// bucket makes remove() reallocate it smaller; sizing that replacement by
// capacity rather than by length leaves the tail filled with zero-valued
// MapElements, which the iterator then reports as if they were real entries.
func TestMapIteratorAfterBucketShrink(t *testing.T) {
	// One bucket makes every key collide, so the removals below are certain to
	// drive a bucket's length far enough under its capacity to trigger a shrink.
	m := genmap.NewMap[int, int](genmap.Equal[int], genmap.NewHasher[int](), 1)

	const nbInserted, nbRemoved = 16, 13
	for k := 0; k < nbInserted; k++ {
		m.Upsert(k, func(elem *genmap.MapElement[int, int], exists bool) {
			elem.Value = k + 1 // never zero, so a zero value marks a phantom
		})
	}
	for k := 0; k < nbRemoved; k++ {
		m.Remove(k)
	}

	want := make(map[int]int, nbInserted-nbRemoved)
	for k := nbRemoved; k < nbInserted; k++ {
		want[k] = k + 1
	}

	if m.Len() != len(want) {
		t.Errorf("expected map with %d elements, got %d elements", len(want), m.Len())
	}

	got := make(map[int]int, len(want))
	var nbIt int
	it := m.Iterator()
	for it.Next() {
		nbIt++
		got[it.Cur().Key] = it.Cur().Value
	}

	if nbIt != len(want) {
		t.Errorf("Expected %d iterations, got %d", len(want), nbIt)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected iterator to yield %v, got %v", want, got)
	}
}

func TestMapIteratorRemoveAfterMapChange(t *testing.T) {
	// Removing a key through the map can leave the iterator positioned past
	// the end of its bucket. it.Remove() must then panic, like it.Cur(), and
	// leave the map untouched instead of removing some other key.
	tests := []struct {
		name       string
		keys       []int
		nbNext     int
		removedKey int
		want       map[int]int
	}{
		{
			name:       "position past the end of the bucket",
			keys:       []int{0, 1, 2},
			nbNext:     3, // on key 2
			removedKey: 0,
			want:       map[int]int{1: 10, 2: 20},
		},
		{
			name:       "bucket emptied",
			keys:       []int{1},
			nbNext:     1, // on key 1
			removedKey: 1,
			want:       map[int]int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// One bucket, so the keys are stored in insertion order.
			m := genmap.NewMap[int, int](genmap.Equal[int], identityHash, 1)
			for _, k := range tt.keys {
				m.Put(k, k*10)
			}
			it := m.Iterator()
			for i := 0; i < tt.nbNext; i++ {
				it.Next()
			}
			m.Remove(tt.removedKey)

			func() {
				defer func() {
					if recover() == nil {
						t.Error("expected it.Remove() to panic")
					}
				}()
				it.Remove()
			}()
			assertMapContent(t, m, tt.want, 3)
		})
	}
}

func BenchmarkMapIterator(b *testing.B) {
	m, _ := initMapAndKeys(100000, 64<<10)
	it := m.Iterator()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		it.Reset()
		for it.Next() {
			_ = it.Cur().Key
			_ = it.Cur().Value
		}
	}
}

func BenchmarkStdMapIterator(b *testing.B) {
	m, _ := initStdMapAndKeys(100000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for k, v := range m {
			_ = k
			_ = v
		}
	}
}
