---
applyTo: '**/*.ts,**/*.js,**/*.go,**/*.rs,**/*.java,**/*.py,**/*.cs,**/workers/**'
description: 'Core concurrency/parallelism design principles (shared mutable state, locks vs message-passing, deadlock prevention, async/await pitfalls, worker-pool sizing) for code running multiple threads/tasks within one process — distinct from cross-service failure handling.'
---

# Concurrency Design Principles

The golden rule: **the safest shared mutable state is the state you never shared.**
Every principle below either removes shared mutable state, or disciplines access to
it so two concurrent operations can never observe or leave it half-updated.

This file is about concurrency **within a single process/runtime** — threads, async
tasks, locks, shared memory. For failures *between* services (retries, circuit
breakers, CAP trade-offs), see [[distributed-systems-design-principles]] instead —
that file's concerns start where a network call begins; this one ends there.

Sources (retrieved 2026-09-27): [MIT 6.005 — Locks and Synchronization](https://web.mit.edu/6.005/www/sp16/classes/23-locks/),
[Medium — Race Conditions, Locks, Semaphores, and Deadlocks](https://medium.com/swlh/race-conditions-locks-semaphores-and-deadlocks-a4f783876529),
[Node.js Learn — Don't Block the Event Loop](https://nodejs.org/learn/asynchronous-work/dont-block-the-event-loop),
[OneUptime — Using asyncio Effectively for I/O-Bound Workloads](https://oneuptime.com/blog/post/2025-01-06-python-asyncio-io-bound/view),
[martinuke0 — Thread Pools In-Depth: Design, Tuning, and Real-World Pitfalls](https://martinuke0.github.io/posts/2025-12-07-thread-pools-in-depth-design-tuning-and-real-world-pitfalls/).

## Prefer no shared mutable state over disciplined access to it

Reach for immutable data and message-passing (channels, queues, actors) between
concurrent units before reaching for a lock. — Two goroutines that only communicate
by sending values over a channel structurally cannot race on that value, because
neither ever holds a reference the other can mutate at the same time; two goroutines
sharing a `map` behind no lock can, and the bug only shows up under real concurrent
load, not in a quick manual test.

## Make the check and the act one atomic unit, not two operations in sequence

- **Check-then-act (TOCTOU) race**: reading a value, deciding based on it, then
  acting, is unsafe unless the whole sequence is atomic — another thread can act
  between your check and your act. — `if (!fileExists(path)) createFile(path)`
  can still fail with "file exists" because another process created it between the
  check and the create; use an atomic operation (`open(path, O_CREAT | O_EXCL)`) that
  does both in one step instead.
- The same failure hits a plain counter increment: `count = count + 1` is a read, an
  add, and a write — three separate steps that two threads can interleave.

```typescript
// BAD: read-modify-write is three separate steps; two threads interleaving
// between them silently drop an increment (final count is short by however
// many increments collided).
let count = 0;
function increment() {
  count = count + 1;
}

// GOOD: one atomic operation — no window for another thread to interleave.
import { Worker } from "worker_threads";
const shared = new SharedArrayBuffer(4);
const count = new Int32Array(shared);
function increment() {
  Atomics.add(count, 0, 1);
}
```

## When you do need a lock, keep it small and its acquisition order fixed

- Hold a lock for the shortest critical section that makes the operation safe — never
  across a network call, a file read, or anything slow. — A lock held across an
  outbound HTTP call turns one slow request into a queue of every other thread
  waiting on that lock, serializing work that had no reason to be serialized.
- Always acquire multiple locks in the same global order, everywhere in the
  codebase. — Thread A locking `accountX` then `accountY`, while thread B locks
  `accountY` then `accountX`, deadlocks the instant both threads hold their first
  lock and wait forever for the other's second lock; a fixed global order (e.g.
  always lock in ID order) makes that interleaving impossible.

## Match the concurrency mechanism to the workload

- **I/O-bound work** (HTTP calls, DB queries, file I/O) — use async/await or an
  event loop; the task spends most of its time waiting, not computing, so many can
  overlap on one thread. — Spinning up a full OS thread per HTTP request, when the
  thread mostly just waits on the network, wastes memory and context-switch overhead
  that async I/O avoids entirely.
- **CPU-bound work** (parsing, image processing, compression) — use a worker/thread
  pool, not the async event loop. — Running a CPU-heavy loop inside an `async`
  function on a single-threaded event loop (Node.js, Python asyncio) blocks that
  loop entirely; every other "concurrent" request on that loop stalls until the CPU
  work finishes, because nothing was actually running in parallel.
- Size a CPU-bound pool near the core count; an I/O-bound pool can run several times
  larger, since its threads spend most of their time blocked/waiting, not competing
  for CPU. — An oversized CPU-bound pool doesn't add throughput, it adds context-switch
  and cache-thrashing overhead for threads that are all fighting for the same cores.

## Bound your fan-out and don't fire-and-forget errors

- Cap how many concurrent tasks/requests a single operation spawns — an unbounded
  `Promise.all` over ten thousand items opens ten thousand connections at once. —
  Fanning out one request per item with no concurrency limit can exhaust the
  process's file descriptors or a downstream service's connection pool in one call;
  a bounded queue/semaphore (process N at a time) gets the same work done without
  the spike.
- Never start an async task and discard its promise/future without handling its
  error. — A fire-and-forget task that throws with nothing awaiting it either
  disappears silently (the failure is never seen) or crashes the process outright
  (an unhandled rejection/`async void` in some runtimes) — neither is a deliberate
  choice, both are an accident of not awaiting or `.catch()`-ing it.

## Applying these under real constraints

- Message-passing/immutability isn't free — copying data between actors or channels
  costs more than in-place mutation for very hot, very small-scale loops; reach for
  it as the default for anything touching shared state across concurrent units, not
  for a tight single-threaded inner loop that never shares anything.
- Not every shared value needs a lock — a value only ever read after being written
  once (a config loaded at startup) needs no synchronization at all; reserve
  locks/atomics for values genuinely mutated by more than one concurrent unit.
- When reviewing concurrent code, name the specific race window (the exact two
  operations that can interleave, as in the examples above) rather than "this looks
  unsafe" — that's what tells the next person what actually has to be fixed.
