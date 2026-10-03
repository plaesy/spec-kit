---
applyTo: '**/*cache*,**/*.ts,**/*.js,**/*.go,**/*.py,**/*.java'
description: 'Core caching design principles (invalidation strategy, key design, stampede prevention, what not to cache) with the failure-scenario reasoning behind performance-optimization.instructions.md''s brief caching checklist items.'
---

# Caching Design Principles

The golden rule, from Phil Karlton's famous line — **"there are only two hard things
in computer science: cache invalidation and naming things"**: a cache is only safe if
you know exactly when a cached value stops being true, and every principle below
exists to make that moment explicit instead of accidental.

This file is about caching as a *design decision* — what to cache, how to invalidate
it, how to key it. For general performance tuning (profiling, algorithmic complexity,
asset bundling, etc.), see `.plaesy/instructions/performance-optimization.md` — caching is one
tool among many there; here it's the whole subject.

Sources (retrieved 2026-09-27): [DEV Community — Caching Strategies Explained](https://dev.to/young_gao/caching-strategies-when-where-and-how-to-cache-1f17),
[Levelop — Caching Strategies in System Design: 4 Patterns](https://levelop.dev/blog/caching-strategies-system-design-four-patterns-failure-modes),
[Levelop — Cache Invalidation Strategies in System Design](https://levelop.dev/blog/cache-aside-write-through-write-behind-read-through-four-strategies-four-failure),
attributed quote: Phil Karlton, via [dev.to — System Design: Caching](https://dev.to/karanpratapsingh/system-design-caching-18j4).

## Pick a caching pattern for the read/write shape, not by default

- **Cache-aside** (app checks cache, falls back to source on miss, populates cache) —
  fits read-heavy data that tolerates brief staleness. — Reaching for this on data
  that must be correct on the very next read after a write (e.g. an account balance
  right after a deposit) serves a stale value from before the write, because nothing
  repopulated the cache at write time.
- **Write-through** (write goes to cache and source together, synchronously) — fits
  low-write, consistency-critical data. — Using this on a high-write-volume path
  doubles write latency (every write now waits on two stores), for consistency
  guarantees a read-heavy path didn't need.
- **Write-behind** (write goes to cache immediately, source is updated asynchronously)
  — fits write-heavy data that can tolerate a small window of data-loss risk for
  speed. — Using this for data where losing the last few writes on a crash is
  unacceptable (financial ledgers) trades correctness for speed exactly where it
  can't be afforded.

## Bound staleness with a TTL even when you also invalidate explicitly

Set a TTL as a safety net on every cached value, in addition to (not instead of)
explicit invalidation on write. — A cache that relies purely on "delete this key when
the source changes" silently serves stale data forever the one time an invalidation
call is missed (a bug, a code path that forgot to call it, a message that never
arrived); a TTL guarantees the staleness window has a hard ceiling regardless of
whether invalidation succeeded.

```typescript
// BAD: no TTL — if the delete-on-write below is ever skipped, this key is
// stale forever with no self-correction.
await cache.set(`user:${id}`, user);

// GOOD: TTL as a backstop even though writes also explicitly invalidate.
await cache.set(`user:${id}`, user, { ttl: 300 }); // 5 min hard ceiling
// ...and on write:
await cache.delete(`user:${id}`);
```

## Prevent stampedes: jitter TTLs, and coalesce concurrent misses

- Add random jitter to TTLs instead of a fixed expiry. — A thousand keys all cached
  with the exact same `ttl: 300` at the same deploy moment all expire in the same
  second, so a thousand requests all miss simultaneously and all hammer the source at
  once (a "thundering herd"); jittering each TTL (e.g. `300 + random(-30, 30)`)
  spreads those expirations out.
- Coalesce concurrent misses on the same key into one source fetch (a lock, or an
  in-flight-request map), not one fetch per waiting request. — Ten requests arriving
  during the same millisecond-wide gap after one key expires, each independently
  missing and each independently querying the source, turns one cache miss into ten
  redundant source queries; the first request should fetch, and the other nine should
  wait for that result instead of duplicating the work.
- Consider **stale-while-revalidate** for read paths where serving a few extra
  seconds of staleness is cheaper than making the user wait: return the expired value
  immediately while refreshing it in the background for the next request.

## Design the cache key from everything that affects the result

Include every input that changes the output in the cache key — not just the most
obvious one. — A cache key of `products:${categoryId}` for an endpoint that also
accepts `sortOrder` and `page` serves page 1 of the "price ascending" result to a
request that asked for page 2 sorted by "newest," because the key can't distinguish
them; the key must encode `categoryId`, `sortOrder`, and `page` together, or the
cache will confidently return the wrong answer instead of failing loudly.

```typescript
// BAD: key omits inputs that change the result — wrong data served, not a miss
const key = `products:${categoryId}`;

// GOOD: key includes every input that affects the response
const key = `products:${categoryId}:${sortOrder}:${page}`;
```

## Know what must never be cached, or never be cached where it is now

- Never cache a per-user or permission-scoped response under a key that other users
  can hit. — Caching an authenticated `/api/me` response under a shared key like
  `route:/api/me` serves one user's private profile to the next unrelated user who
  hits that same route; the key must include the user/session identity whenever the
  response is scoped to it.
- Don't cache an error response as if it were a valid result, unless the TTL for
  errors is deliberately short. — Caching a `500` or a timeout for the same TTL as a
  successful response means a transient one-time failure gets served to every
  subsequent request for the full TTL window, turning a blip into an extended outage.

## Applying these under real constraints

- Not everything needs a cache. Adding one to a low-traffic or already-fast path adds
  invalidation complexity (a second thing that can be wrong) for a performance gain
  nobody will notice — reach for caching where a measured read is genuinely
  expensive/frequent, not by default.
- Watch cache hit ratio, not just presence of a cache — a cache sitting below roughly
  80% hit rate is burning memory and invalidation complexity without paying for
  itself; that's a signal to either fix the key design/TTL or remove the cache.
- When reviewing a caching change, name the specific staleness or leakage scenario at
  risk (stale-after-write, stampede on expiry, cross-user key collision) rather than
  "this could cause issues" — that's what makes the finding actionable.
