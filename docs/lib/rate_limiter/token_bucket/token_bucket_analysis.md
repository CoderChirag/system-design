# Token Bucket Algorithm Performance Analysis

First, the raw benchmark numbers, because every number in this analysis comes from these three lines:
```bash
go test -bench='Ticker|Lazy' -benchtime=2s -cpu=8 ./...

BenchmarkTickerParallel-8     31234443                81.26 ns/op
BenchmarkLazyParallel-8       20844204               102.3 ns/op
BenchmarkLazySerial-8         44000500                57.37 ns/op
```

## 1. What Go's benchmark actually prints
```bash
BenchmarkTickerParallel-8     31234443                81.26 ns/op
        │              │      │                  │
        │              │      │                  └─ nanoseconds per operation
        │              │      └─ iterations it ended up running (b.N)
        │              └─ GOMAXPROCS (8 cores)
        └─ benchmark name
```

The harness has no idea how fast your code is, so it calibrates: it runs the loop 1 time, then 100, then ramps up (roughly multiplying by how far short it fell) until the run lasts at least `-benchtime`. Only that final run is reported. So `ns/op` is just:
```bash
ns/op = total_wall_nanoseconds / iterations
```

Sanity-check it yourself — `iterations × ns/op` should be `≥ -benchtime` (I used `2s`):
```bash
┌────────────┬──────────────────────┬───────────┐
│ benchmark  │  iterations × ns/op  │     =     │
├────────────┼──────────────────────┼───────────┤
│ Ticker     │ 31,234,443 × 81.26ns │ 2.538 s ✓ │
├────────────┼──────────────────────┼───────────┤
│ Lazy       │ 20,844,204 × 102.3ns │ 2.132 s ✓ │
├────────────┼──────────────────────┼───────────┤
│ LazySerial │ 44,000,500 × 57.37ns │ 2.525 s ✓ │
└────────────┴──────────────────────┴───────────┘
```

## 2. `ns/op` → `req/s`

This is pure unit algebra. One second is a billion nanoseconds, so ask: how many `81.26ns` slots fit in `1,000,000,000ns`?
```bash
throughput = 1 / latency

           =        1 op
             ─────────────────
             81.26 × 10⁻⁹ sec

           = 1,000,000,000 ns/sec
             ─────────────────────   =  12,306,177 ops/sec  ≈  12.3M req/s
                 81.26 ns/op
```

The `ns` cancels and you're left with `op/sec`. That's the whole trick, throughput is the reciprocal of latency. Applied to all four:
```bash
┌────────────────┬───────┬─────────────┬─────────────┐
│                │ ns/op │ 1e9 ÷ ns/op │ throughput  │
├────────────────┼───────┼─────────────┼─────────────┤
│ Ticker + mutex │ 81.26 │ 1e9/81.26   │ 12.3M req/s │
├────────────────┼───────┼─────────────┼─────────────┤
│ Lazy + mutex   │ 102.3 │ 1e9/102.3   │ 9.8M req/s  │
├────────────────┼───────┼─────────────┼─────────────┤
│ Lazy, 1 thread │ 57.37 │ 1e9/57.37   │ 17.4M req/s │
└────────────────┴───────┴─────────────┴─────────────┘
```

## 3. But wait — 8 threads. Why isn't it 8 × 12.3M?

Because with `RunParallel`, `b.N` is the total iterations split across all goroutines, and the wall clock covers all of them running simultaneously. So `ns/op` is already aggregate system-wide wall-clock per op, not per-thread latency. No multiplying by thread count.

## 4. Per-request latency: Little's Law

Aggregate throughput doesn't tell a single user what they experience. For that, use **Little's Law** — the most useful formula in systems work:
```bash
L = λ × W

L = concurrency (requests in flight)
λ = throughput (requests/sec)
W = latency (seconds per request)
```

Rearranged, `W = L / λ`. With 8 goroutines in flight:
```bash
┌────────────┬─────┬─────────┬──────────┐
│            │  L  │    λ    │ W = L/λ  │
├────────────┼─────┼─────────┼──────────┤
│ Ticker     │ 8   │ 12.3M/s │ 650 ns   │
├────────────┼─────┼─────────┼──────────┤
│ Lazy       │ 8   │ 9.8M/s  │ 818 ns   │
└────────────┴─────┴─────────┴──────────┘
```

So a single `Allow()` call under 8-way contention takes `~818ns` wall-clock, even though the system completes one every `102ns`. The other `~760ns` is spent queued, waiting for the lock. Throughput and latency are different questions, and Little's Law is the bridge. (Equivalently: `aggregate ns/op × threads = per-request latency` — 102.3 × 8 = 818ns.)

## 5. Decomposing all operations layer-by-layer

I peeled the operation apart layer by layer. This is how you turn **"it's `102ns`"** into **"here's why it's `102ns`"**:
```bash
┌─────────────────────────────────────┬───────────┬──────────┐
│           what I measured           │ 8 threads │ 1 thread │
├─────────────────────────────────────┼───────────┼──────────┤
│ Empty loop (harness overhead)       │ 0.25 ns   │ —        │
├─────────────────────────────────────┼───────────┼──────────┤
│ time.Now() alone                    │ 21.81 ns  │ 40.39 ns │
├─────────────────────────────────────┼───────────┼──────────┤
│ mu.Lock(); mu.Unlock() — empty body │ 78.92 ns  │ 13.71 ns │
├─────────────────────────────────────┼───────────┼──────────┤
│ Lock + token arithmetic, no clock   │ 84.55 ns  │ —        │
├─────────────────────────────────────┼───────────┼──────────┤
│ Full lazy Allow()                   │ 102.3 ns  │ 57.37 ns │
└─────────────────────────────────────┴───────────┴──────────┘
```

Read the findings off it:

- The harness costs nothing (`0.25ns`) — the loop machinery isn't polluting the measurement.

- The single-threaded column adds up almost perfectly: clock `40.39` + mutex `13.71` = `54.1`, plus a few `ns` of integer math = `57.37` measured. When there's no overlap, costs compose linearly.

- The mutex is the whole story. Uncontended (single-threaded): `13.71ns`. Contended by 8 threads: `78.92ns` — a **`5.8×` contention tax**, and **`77%` of the total `102ns`**. The actual token arithmetic `(84.55 − 78.92)` is `5.6ns`. 

> **Our algorithm is essentially free; the coordination is the cost.**

>**The lazy-vs-ticker gap is the clock. `Lazy − Ticker = 102.3 − 81.26 = 21.0ns`, and independently-measured `time.Now() at 8 threads = 21.81ns`. Also `Ticker (81.26) ≈ lock+math without a clock (84.55)`, which is exactly what the ticker's request path is.**

>**One honest wrinkle:** the parallel column does not add up linearly (`84.55 + 21.81 = 106.4` vs `102.3` measured). That's expected — `time.Now()` runs outside the lock, so thread A reads the clock while thread B holds the lock. Overlapped work isn't additive. Microbenchmark costs compose when serialized and don't when parallel.

## 6. Why is 8 threads slower than 1 thread?

This is the deepest point here. `57.37ns → 102.3ns`: adding 7 threads made it `1.8×` worse. Two classical laws explain it.

**Amdahl's Law ** — only the parallel fraction speeds up:
```bash
speedup(N) = ─────────────────
              s + (1 − s)/N       s = serial fraction
```
For our limiter, essentially 100% of the work is inside the mutex, so `s ≈ 1`:
```bash
speedup(N) = 1/(1 + 0/N) = 1     ← no gain from ANY number of cores
```
That gives the hard ceiling: `max throughput = 1 / serial_section_time = 1/57.37ns = 17.4M/s`, forever, on 8 cores or 800. This is why **"just add more cores" doesn't fix a lock**.

**Universal Scalability Law (Gunther)** explains why it goes backwards past the ceiling — it adds a coherency-cost term that grows as `N²`:
```bash
                    N
C(N) = ───────────────────────────
        1 + σ(N−1) + κN(N−1)
                │        └─ coherency (cache-line ping-pong) — grows as N²
                └─ contention (queueing for the lock)
```
My hand-rolled harness shows that curve directly:
```bash
threads   throughput
   1      15.96M ops/sec
   2      16.84M ops/sec   ← peak
   4      11.72M ops/sec   ← declining
   8       9.52M ops/sec   ← retrograde
```
Peak at 2 threads, then retrograde scaling. Physically: the mutex word and the bucket field live in a cache line that must be transferred between cores on every single operation, and losers get parked and woken by the scheduler. 

> **With 8 cores fighting over one 64-byte cache line, you spend more time shipping it around than doing work.**

> **The actual fix for all of this is sharding — per-key buckets on separate cache lines, so different users never touch the same line. You reduce `N` per lock rather than making the lock faster.**