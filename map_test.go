package genmap_test

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	"github.com/ronanh/genmap"
)

func TestMap(t *testing.T) {
	m := genmap.NewMap[string, int](genmap.Equal[string], genmap.NewHasher[string]())
	if m.Len() != 0 {
		t.Errorf("expected empty map, got %d elements", m.Len())
	}

	m.Upsert("foo", func(val *int, exists bool) {
		*val = 42
	})
	if m.Len() != 1 {
		t.Errorf("expected map with 1 element, got %d elements", m.Len())
	}
	if elem, ok := m.Get("foo"); !ok || elem != 42 {
		t.Errorf("expected element with key 'foo' and value 42, got %v", elem)
	}

	m.Upsert("foo", func(val *int, exists bool) {
		*val = 43
	})
	if m.Len() != 1 {
		t.Errorf("expected map with 1 element, got %d elements", m.Len())
	}
	if elem, ok := m.Get("foo"); !ok || elem != 43 {
		t.Errorf("expected element with key 'foo' and value 43, got %v", elem)
	}

	m.Upsert("bar", func(val *int, exists bool) {
		*val = 44
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
	m.Upsert(1, func(val *string, exists bool) {
		*val = "one"
	})
	m.Upsert(2, func(val *string, exists bool) {
		*val = "two"
	})
	m.Upsert(3, func(val *string, exists bool) {
		*val = "three"
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
	m.Upsert("a", func(val *int, exists bool) {
		*val = 1
	})
	m.Upsert("b", func(val *int, exists bool) {
		*val = 2
	})
	m.Upsert("c", func(val *int, exists bool) {
		*val = 3
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
	m.Upsert("a", func(val *int, exists bool) {
		*val = 4
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
	update := func(val *MyValue, exists bool) {
		val.v1++
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
	update := func(val *MyValue, exists bool) {
		val.v1++
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
	update := func(val *MyValue, exists bool) {
		val.v1++
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
		m.Upsert(k, func(val *MyValue, exists bool) {
			val.v1 = v
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
	m.Upsert(1, func(val *string, exists bool) {
		*val = "one"
	})
	m.Upsert(2, func(val *string, exists bool) {
		*val = "two"
	})
	m.Upsert(3, func(val *string, exists bool) {
		*val = "three"
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
