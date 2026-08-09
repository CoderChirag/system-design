package fundamentals

import "testing"

var sink uint64

/*
A fixed lump of CPU work. See docs/lib/fundamentals/parallelism.md for explanation.
*/
func spin() uint64 {
	x := uint64(88172645463325252)
	for i := 0; i < 100; i++ {
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
	}
	return x
}

/*
Execution: go test -bench='PureCPUParallel' -benchtime=2s -cpu=1,2,4,8 ./...
*/

func BenchmarkPureCPUParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var acc uint64
		for pb.Next() {
			acc += spin()
		}
		sink = acc
	})
}

/*
Execution: go test -bench='PureCPUSerial' -benchtime=2s -cpu=1,2,4,8 ./...
*/
func BenchmarkPureCPUSerial(b *testing.B) {
	var acc uint64
	for b.Loop() {
		acc += spin()
	}
	sink = acc
}
