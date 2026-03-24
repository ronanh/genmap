// Package genmap provides a simple generic hash‑map implementation that
// operates directly on slices of buckets.  This file defines the entry
// abstractions used to access and optionally create map elements.
//
// The map stores its data in a slice of buckets (`[][]MapElement[K,V]`).  Each
// bucket holds one or more `MapElement`s that share the same hash value
// (collision handling via chaining).  `MapEntry` is a thin wrapper around a
// pointer to an existing element, while `MaybeMapEntry` represents a lookup
// that may or may not have found an element.  The latter can be turned into a
// concrete entry (creating a new element if necessary) via `OrDefault`.
//
// These helpers are used by the public `Map` API (e.g. `Get`, `Set`, `Delete`)
// to provide a convenient, zero‑allocation way to mutate entries in place.
package genmap

// MapEntry is a lightweight handle to an existing map element.  It does not
// own the element; it merely holds a pointer that can be used to mutate the
// element via `MutateWith`.
type MapEntry[K any, V any] struct {
	elem *MapElement[K, V] // pointer to the underlying element
}

// MutateWith runs the supplied function on the underlying map element.
// The caller can modify the key, value, or other fields directly.
func (entry MapEntry[K, V]) MutateWith(f func(*MapElement[K, V])) {
	f(entry.elem)
}

// MaybeMapEntry represents the result of a lookup that may be absent.
// It stores enough context (map reference, bucket position, hash, key) to
// either return the found element or create a new one on demand.
type MaybeMapEntry[K any, V any] struct {
	m         *Map[K, V]        // reference to the parent map
	elem      *MapElement[K, V] // nil if the key was not found
	bucketPos uint64            // index of the bucket in the map's bucket slice
	hash      uint64            // pre‑computed hash of the key
	key       K                 // the key being looked up
}

// makeOptionalEntry performs a lookup for `key` in map `m`.  If an element
// with a matching hash and key is found, it returns a `MaybeMapEntry` that
// points to that element.  Otherwise it returns a `MaybeMapEntry` with a nil
// `elem`, allowing the caller to create a new entry via `OrDefault`.
func makeOptionalEntry[K any, V any](m *Map[K, V], key K) MaybeMapEntry[K, V] {
	hash := m.hash(key)
	bucketPos := hash % uint64(len(m.buckets))
	bucket := m.buckets[bucketPos]
	if len(bucket) > 0 {
		if bucket[0].hash == hash && m.equal(bucket[0].Key, key) {
			return MaybeMapEntry[K, V]{m, &bucket[0], bucketPos, hash, key}
		}
		if len(bucket) > 1 {
			// slow path – iterate over the rest of the bucket
			for pos := 1; pos < len(bucket); pos++ {
				if bucket[pos].hash == hash && m.equal(bucket[pos].Key, key) {
					return MaybeMapEntry[K, V]{m, &bucket[pos], bucketPos, hash, key}
				}
			}
		}
	}
	// Not found – return a placeholder with nil element
	return MaybeMapEntry[K, V]{m, nil, bucketPos, hash, key}
}

// Exists reports whether the lookup succeeded (i.e. an element was found).
func (entry *MaybeMapEntry[K, V]) Exists() bool {
	return entry.elem != nil
}

// OrDefault returns a concrete `MapEntry`.  If the element already exists it
// is returned unchanged; otherwise a new element is allocated, inserted into
// the appropriate bucket, and a handle to that new element is returned.
func (entry *MaybeMapEntry[K, V]) OrDefault() MapEntry[K, V] {
	if entry.elem != nil {
		return MapEntry[K, V]{entry.elem}
	}

	m := entry.m
	bucketPos := entry.bucketPos
	hash := entry.hash
	key := entry.key
	bucket := m.buckets[bucketPos]

	// Grow the map length to account for the new element
	m.len++

	// Ensure the bucket slice exists
	if bucket == nil {
		bucket = m.newElemSlice(0, 1)
	}
	// Make room for the new element, reusing capacity when possible
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
	// Insert the new element at the end of the bucket (modulo length to
	// avoid bounds checks)
	pos := uint64(len(bucket)-1) % uint64(len(bucket))
	bucket[pos].hash = hash
	bucket[pos].Key = key

	// Write the bucket back to the map's bucket array
	m.buckets[hash%uint64(len(m.buckets))] = bucket
	return MapEntry[K, V]{&bucket[pos]}
}
