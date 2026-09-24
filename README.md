# genmap

[![Go Reference](https://pkg.go.dev/badge/github.com/ronanh/genmap.svg)](https://pkg.go.dev/github.com/ronanh/genmap)

Go native map only accept [comparable](https://go.dev/ref/spec#Comparison_operators) types as key.
This prominently prevents the use of slices and maps as keys.

`genmap` provides a generic map implementation that does not have such limitation.

## Features

* `Get`, `Put`, `Remove`, `Upsert`
* `Entry` to check for a key, then update or insert its element with a single lookup
* `Len`, `Clear`
* `Iterator` allowing `Remove` while iterating
* A nil map reads as empty, like a nil Go map

It's up to the user to provide a hash and an equality function for the key type (Helpers 
are provided for the common cases).

## Limitations

The rather simple implementation is designed for the case where the number of keys is 
roughly known in advance. Performance will suffer if the number of keys is much larger
than the number of buckets.

The number of buckets never changes. Each empty bucket still costs 24 bytes, and
iterating over or clearing the map visits every bucket: with the default of 64k buckets,
a map takes 1.5 MB however few keys it holds. Pass a bucket count close to the expected
number of keys to `NewMap`.

A map is not safe for concurrent use while it is being modified.

## Example usage

```go
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
```

## Benchmarks

Benchmarked against the standard map implementation.
Use of a 64k bucket size for 100k keys (Apple M1 Max, Go 1.27, median of 6 runs).

```
BenchmarkMapGet-10                     25298931        48.96 ns/op          0 B/op        0 allocs/op
BenchmarkStdMapGet-10                  72832658        16.99 ns/op          0 B/op        0 allocs/op
BenchmarkMapPut100k-10                      241      4769061 ns/op    7720841 B/op     1440 allocs/op
BenchmarkStdMapPut100k-10                   823      1487336 ns/op    5248145 B/op      257 allocs/op
BenchmarkMapPutOverwrite-10            21949846        55.16 ns/op          0 B/op        0 allocs/op
BenchmarkStdMapPutOverwrite-10         52113963        23.29 ns/op          0 B/op        0 allocs/op
BenchmarkMapUpsert100k-10                   248      5037831 ns/op    7712815 B/op     1445 allocs/op
BenchmarkStdMapUpsert100k-10                614      1979311 ns/op    5248143 B/op      257 allocs/op
BenchmarkMapUpsertIncrement-10         21710925        58.95 ns/op          0 B/op        0 allocs/op
BenchmarkStdMapUpsertIncrement-10      33433012        37.88 ns/op          0 B/op        0 allocs/op
BenchmarkMapUpsertDelete-10            20493829        55.69 ns/op          0 B/op        0 allocs/op
BenchmarkStdMapUpsertDelete-10         31600706        39.64 ns/op          0 B/op        0 allocs/op
BenchmarkMapIterator-10                    1690       704702 ns/op          0 B/op        0 allocs/op
BenchmarkStdMapIterator-10                 2162       555166 ns/op          0 B/op        0 allocs/op
```

## Credits

* [dolthub maphash](https://github.com/dolthub/maphash) generic hash function for comparables
