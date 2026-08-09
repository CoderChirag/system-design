# Token Bucket Rate Limiting (Lazy)

## Why do we do `tb.lastRefill = tb.lastRefill.Add(time.Duration(n) * tb.refillInterval)` instead of `tb.lastRefill = time.Now()`?
Because integer division throws away the remainder, and `= now` would throw away the tokens that remainder represents.

The mechanics. With `refillInterval = 100ms`, suppose `lastRefill = T0` and a call arrives at `T0 + 250ms`:

- `n = int(250ms / 100ms) = 2` → you credit 2 intervals' worth of tokens.
- The extra `50ms` is unearned, it hasn't matured into a token yet.

Adding `n * refillInterval` sets `lastRefill = T0 + 200ms`, so that leftover `50ms` stays on the clock and counts toward the next refill. Setting `lastRefill = now` would move it to `T0 + 250ms` and silently discard those 50ms.

Why it matters: drift. The loss compounds on every call. If requests arrive every `150ms` with a `100ms` interval:

| call at | `now - lastRefill` |  `n` | `lastRefill` after `Add` | `lastRefill` after `=now` |
| --- | --- | --- | --- | --- |
| `T+150` | `150ms` | `1` | `T+100` | `T+150` |
| `T+300` | `200ms` / `150ms` | `2` / `1` | `T+300` | `T+300` |
| `T+450` | `150ms` / `150ms` | `1` / `1` | `T+400` | `T+450` |
| `T+600` | `200ms` / `150ms` | `2` / `1` | `T+600` | `T+600` |

Over `600ms` the `Add` version grants `6` refills, exactly the true rate. The `=now` version grants `4`. Under a request pattern with period just under `2 * refillInterval`, `=now` leaks up to `~50%` of the configured rate. In the worst case (arrivals just under one interval apart, `n` always 0 — wait, `n = 0` skips the branch entirely, so no loss there), the leak shows up whenever `n >= 1` and there's a nonzero remainder, which is the common case under real traffic.

The invariant. `lastRefill` means **"the timestamp of the last interval boundary I have already paid out,"** **NOT** **"the last time I looked."** Advancing it by whole intervals keeps it aligned to the original boundary grid anchored at construction time; assigning `now` re-anchors the grid on every call and makes the effective rate depend on arrival timing rather than on config.