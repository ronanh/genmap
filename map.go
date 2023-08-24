package genmap

const (
	maxFreeSlices = 128
)

// MapElement is a generic key-value pair used in the Map[K, V] implementation.
type MapElement[K any, V any] struct {
	Key   K
	Value V
	hash  uint64
}

// Map is a generic hash map implementation that allows any type for keys.
// Map instance should be instantiated using the NewMap function.
type Map[K, V any] struct {
	equal       func(k1, k2 K) bool
	hash        func(k K) uint64
	buckets     [][]MapElement[K, V]
	len         int
	allocBuffer []MapElement[K, V]
	freeSlices  [][]MapElement[K, V]
}

// NewMap returns a new instance of Map[K, V] with the given equality and hash functions.
// The optional bucketSizeOpt parameter specifies the size of each bucket in the map.
// If not provided, a default bucket size (64k) is used.
// Special care should be taken when choosing a bucket size as it can have a significant impact on performance.
// For good performance, the bucket size should be close to the expected number of elements in the map.
func NewMap[K any, V any](equal func(k1, k2 K) bool, hash func(k K) uint64, bucketSizeOpt ...int) *Map[K, V] {
	if len(bucketSizeOpt) > 1 {
		panic("too many arguments")
	}
	bucketsSize := 64 << 10
	if len(bucketSizeOpt) == 1 {
		bucketsSize = bucketSizeOpt[0]
	}

	bucket := &Map[K, V]{
		equal:   equal,
		hash:    hash,
		buckets: make([][]MapElement[K, V], bucketsSize),
	}
	return bucket
}

// returns the number of elements in the map.
func (m *Map[K, V]) Len() int {
	return m.len
}

// Clear removes all elements from the map.
func (m *Map[K, V]) Clear() {
	for i := range m.buckets {
		m.buckets[i] = nil
	}
	m.len = 0
}

// returns the value associated with the given key.
func (m *Map[K, V]) Get(key K) (V, bool) {
	hash := m.hash(key)
	bucketID := hash % uint64(len(m.buckets))
	bucket := m.buckets[bucketID]
	if len(bucket) == 0 {
		return *new(V), false
	}

	if bucket[0].hash == hash && m.equal(bucket[0].Key, key) {
		return bucket[0].Value, true
	}

	if len(bucket) > 1 {
		// slow path
		for pos := 1; pos < len(bucket); pos++ {
			if bucket[pos].hash == hash && m.equal(bucket[pos].Key, key) {
				return bucket[pos].Value, true
			}
		}
	}
	return *new(V), false
}

// Put inserts the given key-value pair into the map.
func (m *Map[K, V]) Put(key K, val V) {
	hash := m.hash(key)
	bucket := m.buckets[hash%uint64(len(m.buckets))]
	if len(bucket) > 0 {
		if bucket[0].hash == hash && m.equal(bucket[0].Key, key) {
			bucket[0].Value = val
			return
		}
		if len(bucket) > 1 {
			// slow path
			for pos := 1; pos < len(bucket); pos++ {
				if bucket[pos].hash == hash && m.equal(bucket[pos].Key, key) {
					bucket[pos].Value = val
					return
				}
			}
		}
	}
	m.len++
	if bucket == nil {
		bucket = m.newElemSlice(0, 1)
	}
	if len(bucket)+1 > cap(bucket) {
		if len(bucket) < 3 {
			newBucket := m.newElemSlice(len(bucket)+1, 4)
			copy(newBucket, bucket)
			m.freeElemSlice(bucket)
			bucket = newBucket
		} else {
			bucket = append(bucket, MapElement[K, V]{
				Key:   key,
				Value: val,
				hash:  hash,
			})
		}
	} else {
		bucket = bucket[:len(bucket)+1]
		bucket[len(bucket)-1] = MapElement[K, V]{
			Key:   key,
			Value: val,
			hash:  hash,
		}
	}
	m.buckets[hash%uint64(len(m.buckets))] = bucket
}

// Upsert inserts or modifies the given entry into the map.
// The update function is called with the current value or the new one.
func (m *Map[K, V]) Upsert(key K, update func(elem *MapElement[K, V], exists bool)) {
	hash := m.hash(key)
	bucket := m.buckets[hash%uint64(len(m.buckets))]
	if len(bucket) > 0 {
		if bucket[0].hash == hash && m.equal(bucket[0].Key, key) {
			update(&bucket[0], true)
			return
		}
		if len(bucket) > 1 {
			// slow path
			for pos := 1; pos < len(bucket); pos++ {
				if bucket[pos].hash == hash && m.equal(bucket[pos].Key, key) {
					update(&bucket[pos], true)
					return
				}
			}
		}
	}
	m.len++
	if bucket == nil {
		bucket = m.newElemSlice(0, 1)
	}
	if len(bucket)+1 <= cap(bucket) {
		bucket = bucket[:len(bucket)+1]
	} else {
		if len(bucket) < 3 {
			newBucket := m.newElemSlice(len(bucket)+1, 4)
			copy(newBucket, bucket)
			m.freeElemSlice(bucket)
			bucket = newBucket
		} else {
			bucket = append(bucket, MapElement[K, V]{})
		}
	}
	pos := uint64(len(bucket)-1) % uint64(len(bucket)) // Eliminate bounds check
	bucket[pos].hash = hash
	bucket[pos].Key = key
	m.buckets[hash%uint64(len(m.buckets))] = bucket
	update(&bucket[pos], false)
}

// Remove removes the given key from the map and returns it.
func (m *Map[K, V]) Remove(key K) (MapElement[K, V], bool) {
	hash := m.hash(key)
	bucketID := hash % uint64(len(m.buckets))
	bucket := m.buckets[bucketID]
	if len(bucket) == 0 {
		return MapElement[K, V]{}, false
	}
	if bucket[0].hash == hash && m.equal(bucket[0].Key, key) {
		return m.remove(bucketID, uint64(0)), true
	}
	if len(bucket) > 1 {
		// slow path
		for pos := 1; pos < len(bucket); pos++ {
			if bucket[pos].hash == hash && m.equal(bucket[pos].Key, key) {
				return m.remove(bucketID, uint64(pos)), true
			}
		}
	}
	return MapElement[K, V]{}, false
}

func (m *Map[K, V]) remove(bucketID uint64, pos uint64) (elem MapElement[K, V]) {
	m.len--
	bucket := m.buckets[bucketID%uint64(len(m.buckets))] // Eliminate bounds check
	pos = pos % uint64(len(bucket))                      // Eliminate bounds check
	elem = bucket[pos]
	copy(bucket[pos:], bucket[pos+1:])
	// force clear the last element to avoid memory leak
	bucket[len(bucket)-1] = MapElement[K, V]{}
	bucket = bucket[:len(bucket)-1]
	if len(bucket) == 0 {
		// free the bucket
		m.freeElemSlice(bucket)
		m.buckets[bucketID%uint64(len(m.buckets))] = nil
		return
	} else if len(bucket)+1 < cap(bucket)/3 {
		// shrink the bucket
		newBucket := make([]MapElement[K, V], cap(bucket)/2)
		copy(newBucket, bucket)
		bucket = newBucket
	}
	m.buckets[bucketID%uint64(len(m.buckets))] = bucket // Eliminate bounds check
	return
}

// Iterator returns a new iterator over the map.
func (m *Map[K, V]) Iterator() *MapIterator[K, V] {
	return &MapIterator[K, V]{m: m}
}

func (m *Map[K, V]) newElemSlice(size, capacity int) []MapElement[K, V] {
	if len(m.freeSlices) > 0 && len(m.freeSlices[len(m.freeSlices)-1]) >= size {
		last := len(m.freeSlices) - 1
		slice := m.freeSlices[last]
		m.freeSlices = m.freeSlices[:last]
		return slice
	}
	if len(m.allocBuffer) < capacity {
		m.allocBuffer = make([]MapElement[K, V], 1024)
	}
	last := len(m.allocBuffer) - capacity
	slice := m.allocBuffer[last : last+size : last+capacity]
	m.allocBuffer = m.allocBuffer[:last]
	return slice
}

func (m *Map[K, V]) freeElemSlice(slice []MapElement[K, V]) {
	if len(slice) > 0 {
		for i := range slice {
			slice[i] = MapElement[K, V]{}
		}
		slice = slice[:0]
	}
	if len(m.freeSlices) < maxFreeSlices {
		m.freeSlices = append(m.freeSlices, slice)
	}
}

// MapIterator is an iterator over a map.
type MapIterator[K any, V any] struct {
	m      *Map[K, V]
	mapPos uint64
	pos    uint64
	ready  bool
}

// Next advances the iterator and returns true if there is another element
func (it *MapIterator[K, V]) Next() bool {
	if it.ready {
		// ensure the cursor is moved
		it.pos++
	}
	// ensure the cursor is at a valid position
	// otherwise move to the next valid position
	for it.mapPos < uint64(len(it.m.buckets)) {
		if it.pos < uint64(len(it.m.buckets[it.mapPos])) {
			it.ready = true
			return true
		}
		it.mapPos++
		it.pos = 0
	}
	it.ready = false
	return false
}

// Cur returns the current element
func (it *MapIterator[K, V]) Cur() *MapElement[K, V] {
	if !it.ready || it.mapPos >= uint64(len(it.m.buckets)) || it.pos >= uint64(len(it.m.buckets[it.mapPos])) {
		panic("iterator position not set")
	}
	return &it.m.buckets[it.mapPos][it.pos]
}

// Remove removes the current element from the map and returns it.
// After calling Remove, Next must be called before calling Cur again.
func (it *MapIterator[K, V]) Remove() MapElement[K, V] {
	if !it.ready {
		panic("iterator position not set")
	}
	it.ready = false
	return it.m.remove(uint64(it.mapPos), it.pos)
}

// Reset resets the iterator to the beginning of the map.
func (it *MapIterator[K, V]) Reset() {
	it.mapPos = 0
	it.pos = 0
	it.ready = false
}
