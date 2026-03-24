package genmap

type MapEntry[K any, V any] struct {
	elem *MapElement[K, V]
}

func (entry MapEntry[K, V]) MutateWith(f func(*MapElement[K, V])) {
	f(entry.elem)
}

type MaybeMapEntry[K any, V any] struct {
	m         *Map[K, V]
	elem      *MapElement[K, V]
	bucketPos uint64
	hash      uint64
	key       K
}

func makeOptionalEntry[K any, V any](m *Map[K, V], key K) MaybeMapEntry[K, V] {
	hash := m.hash(key)
	bucketPos := hash % uint64(len(m.buckets))
	bucket := m.buckets[bucketPos]
	if len(bucket) > 0 {
		if bucket[0].hash == hash && m.equal(bucket[0].Key, key) {
			return MaybeMapEntry[K, V]{m, &bucket[0], bucketPos, hash, key}
		}
		if len(bucket) > 1 {
			// slow path
			for pos := 1; pos < len(bucket); pos++ {
				if bucket[pos].hash == hash && m.equal(bucket[pos].Key, key) {
					return MaybeMapEntry[K, V]{m, &bucket[pos], bucketPos, hash, key}
				}
			}
		}
	}
	return MaybeMapEntry[K, V]{m, nil, bucketPos, hash, key}
}

func (entry *MaybeMapEntry[K, V]) Exists() bool {
	return entry.elem != nil
}

func (entry *MaybeMapEntry[K, V]) OrDefault() MapEntry[K, V] {
	if entry.elem != nil {
		return MapEntry[K, V]{entry.elem}
	}

	m := entry.m
	bucketPos := entry.bucketPos
	hash := entry.hash
	key := entry.key
	bucket := m.buckets[bucketPos]

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
	return MapEntry[K, V]{&bucket[pos]}
}
