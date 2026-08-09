# Parallelism Fundamentals

## 1. Testing Code Analysis
For understanding serial & parallel performances, let's first analyze the testing code `BenchmarkPureCPUParallel()` we have in `lib/fundamentals/bench_cpu_test.go`:

### `spin()` function used by the code: A fixed lump of CPU work
```go
x := uint64(88172645463325252)
x ^= x << 13;  x ^= x >> 7;  x ^= x << 17
```
That's **`xorshift64`** (Marsaglia's classic PRNG step), with his canonical seed and the standard `13/7/17` shift triple. It's used here not as a random generator but as a cheap, dependency-chained unit of work: each line feeds the next, so the CPU can't pipeline it away, and `100` rounds gives `~300 shift/xor ops` — **big enough to dominate loop overhead, small enough to stay in registers.**

**Key properties for the experiment:**
- No shared state, no locks, no clock, no allocation. Purely register-local. So it's ***embarrassingly parallel*** — `N` threads should get `N×` the throughput, with nothing to contend on.
- Deterministic. It returns the same constant every call. That's fine; the point is the cycles burned, not the value.
- It survives the optimizer because it's too big to inline and Go won't constant-fold a loop across a call boundary.

### `BenchmarkPureCPUParallel`
```go
b.RunParallel(func(pb *testing.PB) { ... })
```
`RunParallel` spawns `GOMAXPROCS` goroutines (controlled by `-cpu`) and hands each a `*testing.PB`. The total `b.N` iterations are split across them — `pb.Next()` pulls a batch off a shared atomic counter and returns false once the budget is exhausted. So all threads together perform `b.N` iterations, not `b.N` each.
```go
var acc uint64
for pb.Next() { acc += spin() }
sink = acc
```
`acc` is a goroutine-local accumulator, and `sink` is the package-level `var sink uint64` at line 9. Both exist purely to defeat dead-code elimination: without consuming the result, the compiler could delete the `spin()` call entirely and you'd benchmark an empty loop. Accumulating locally and storing once at the end keeps the per-iteration cost pure — a global write on every iteration would add a store to shared memory and pollute the measurement.

(All goroutines race on that final `sink = write`. Harmless here — the value is never read — but it's why you wouldn't run this under `-race`.)

## 2. Serial Code Performance Analysis
```bash
go test -bench='PureCPUSerial' -benchtime=2s -cpu=1,2,4,8 ./...

BenchmarkPureCPUSerial          17747835               135.5 ns/op
BenchmarkPureCPUSerial-2        17699413               135.7 ns/op
BenchmarkPureCPUSerial-4        17677620               137.5 ns/op
BenchmarkPureCPUSerial-8        17733219               137.5 ns/op
```
The control: `BenchmarkPureCPUSerial` stayed at `135.5` / `135.7` / `137.5` / `137.5` across `-cpu=1,2,4,8` — one goroutine can't use more cores, so `GOMAXPROCS` changes nothing.

## 3. Does `8 cores cpu` really means `~8x` performance improvement?
```bash
go test -bench='PureCPUParallel' -benchtime=2s -cpu=1,2,4,8 ./...

BenchmarkPureCPUParallel      136.5 ns/op   ← 1 thread
BenchmarkPureCPUParallel-2     70.92 ns/op  ← 1.92x faster
BenchmarkPureCPUParallel-4     36.50 ns/op  ← 3.74x faster
BenchmarkPureCPUParallel-8     26.88 ns/op  ← 5.08x faster
```
`8 threads` gave only `5.08×`, not `8×`, because the `Mac M1 CPU` has `4 performance + 4 efficiency cores`. Four threads land on `P-cores` and scale near-linearly (`3.74×`); the next four land on slower `E-cores` and add much less. **"8 cores" is rarely 8 equal cores.**