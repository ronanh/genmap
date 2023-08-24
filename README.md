# genmap

[![Go Reference](https://pkg.go.dev/badge/github.com/ronanh/genmap.svg)](https://pkg.go.dev/github.com/ronanh/genmap)

Go native map only accept [comparable](https://go.dev/ref/spec#Comparison_operators) types as key.
This prominently prevents the use of slices and maps as keys.

`genmap` provides a generic map implementation that does not have such limitation.

## Features

* `Get`, `Put`, `Delete`, `Upsert`
* `Len`, `Clear`
* `Iterator` allowing `Delete` while iterating

It's up to the user to provide a hash and an equality function for the key type (Helpers 
are provided for the common cases).

## Limitations

The rather simple implementation is designed for the case where the number of keys is 
roughly known in advance. Performance will suffer if the number of keys is much larger
than the number of buckets.

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
v := m.Get(MyKey{1, []string{"a", "b"}})
if v != nil {
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
```

## Benchmarks

Benchmarked against the standard map implementation.
Use of a 64k bucket size for 100k keys.

```
BenchmarkMapGet-10                   	26824383	        45.86 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdMapGet-10                	28797897	        40.86 ns/op	       0 B/op	       0 allocs/op
BenchmarkMapPut100k-10               	     294	   4057501 ns/op	 7716181 B/op	    1440 allocs/op
BenchmarkStdMapPut100k-10            	     390	   3043593 ns/op	 5208164 B/op	    1645 allocs/op
BenchmarkMapPutOverwrite-10          	22302621	        51.56 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdMapPutOverwrite-10       	29951390	        40.65 ns/op	       0 B/op	       0 allocs/op
BenchmarkMapUpsert100k-10            	     313	   3790824 ns/op	 7715365 B/op	    1445 allocs/op
BenchmarkStdMapUpsert100k-10         	     360	   3426847 ns/op	 5207400 B/op	    1642 allocs/op
BenchmarkMapUpsertIncrement-10       	24000559	        51.03 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdMapUpsertIncrement-10    	20924194	        57.68 ns/op	       0 B/op	       0 allocs/op
BenchmarkMapUpsertDelete-10          	25785609	        44.60 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdMapUpsertDelete-10       	20577380	        57.61 ns/op	       0 B/op	       0 allocs/op
BenchmarkMapIterator-10              	    1669	    712207 ns/op	       0 B/op	       0 allocs/op
BenchmarkStdMapIterator-10           	    1566	    766186 ns/op	       0 B/op	       0 allocs/op
```

## Credits

* [dolthub maphash](https://github.com/dolthub/maphash) generic hash function for comparables
