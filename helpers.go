package genmap

import (
	"reflect"

	"github.com/dolthub/maphash"
)

const (
	prime    uint64 = 1099511628211        // Large prime number for mixing hashes
	HashSeed uint64 = 14695981039346656037 // Seed value for the initial hash
)

// CombineHash Combine two uint64 hashes into a single hash
func CombineHash(seed, hash uint64) uint64 {
	return seed ^ (hash + prime + (seed << 6) + (seed >> 2))
}

// CombineHashes Combine a list of uint64 hashes into a single hash
func CombineHashes(hashes ...uint64) uint64 {
	combinedHash := HashSeed
	for _, hash := range hashes {
		combinedHash = CombineHash(combinedHash, hash)
	}
	return combinedHash
}

// DeepEqual Deep equal comparison of two values
func DeepEqual[T any](a, b T) bool {
	return reflect.DeepEqual(a, b)
}

// Equal Compares two `comparable` values
func Equal[K comparable](k1, k2 K) bool {
	return k1 == k2
}

// Hasher Returns a hash function for `comparable`
func NewHasher[T comparable]() func(T) uint64 {
	mh := maphash.NewHasher[T]()
	return mh.Hash
}
