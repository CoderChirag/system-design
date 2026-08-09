# Fundamentals

## Aggregate Throughput via Aggregate Latency (`ns/op` -> `req/s`)
Calculating aggregate throughput for aggregate latency of `81.26 ns/op`
```bash
throughput = 1 / latency

           =        1 op
             ─────────────────
             81.26 × 10⁻⁹ sec

           = 1,000,000,000 ns/sec
             ─────────────────────   =  12,306,177 ops/sec  ≈  12.3M req/s
                 81.26 ns/op
```

## Latency from Throughput or Aggregate Latency: Little's Law

Aggregate throughput doesn't tell a single user what they experience. For that, use **Little's Law** — the most useful formula in systems work:
```bash
L = λ × W

L = concurrency (requests in flight)
λ = aggregate throughput (requests/sec)
W = latency (seconds per request)
```

Rearranged, `W = L / λ`, for Aggregate Throughput of `12.3M req/s` with a concurrency of `8` (`GOMAXPROCS=8`):
```bash
Per Request Latency, W = Concurrency L / Aggregate Throughput λ
           =        8
             ───────────────── = 650 ns
                12.3 M req/s
```

Equivalently, 
```bash
Per Request Latency, W = Concurrency L / Aggregate Throughput λ (req/s)
                       = Concurrency L / (1 / Aggregate Latency (ns/op))
                       = Concurrency L * Aggregate Latency (ns/op)
                       = 8 * 81.26 ns/op
                       = 650.08 ns
```
## Common Latency Nubmers

| Operation Name | Time |
| --- | --- |
| L1 Cache Hit | `~1ns` |
| L2 Cache Reference | `~7ns` |
| Mutex Lock / Unlock | `~100ns` |
| Main Memory Reference | `~100ns` |
| SSD Read | `~100μs` |
| Round Trip within same datacenter | `500μs` |
| Disk Seek | `10ms` |
| Round Trip across Continent | `100ms` |