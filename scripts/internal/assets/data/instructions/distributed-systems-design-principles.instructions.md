---
applyTo: '**/services/**,**/*.client.ts,**/*client*,**/*gateway*,**/*queue*,**/*event*'
description: 'Core distributed-systems design principles (idempotency, retries/backoff, circuit breakers, CAP trade-offs, saga/outbox over distributed transactions) to apply when designing or reviewing cross-service calls, messaging, or multi-step workflows.'
---

# Distributed Systems Design Principles

The golden rule: **assume partial failure is normal, not exceptional.** A remote call
can be slow, drop the response, or succeed on the server while the client sees a
timeout — every principle below exists to make that survivable instead of silently
corrupting state or cascading into a full outage.

Sources (retrieved 2026-09-27): [GeeksforGeeks — Retries Strategies in Distributed Systems](https://www.geeksforgeeks.org/system-design/retries-strategies-in-distributed-systems/),
[System Design Handbook — Distributed Systems Principles Explained](https://www.systemdesignhandbook.com/blog/distributed-systems-principles/),
[iam.slys.dev — How systems handle failure: retries, circuit breakers, idempotency](https://iam.slys.dev/p/how-systems-handle-failure-retries),
[microservices.io — Pattern: Saga](https://microservices.io/patterns/data/saga.html),
[Conduktor — Saga Pattern for Microservices Explained](https://www.conduktor.io/glossary/saga-pattern-for-distributed-transactions).

This file is about **application-level design** for calls, messages, and multi-step
workflows across services — not deployment or infrastructure ops (see
[[devops-core-principles]] and [[kubernetes-deployment-best-practices]] for those).

## Make retried operations idempotent

- **Idempotency** — repeating the same operation must produce the same end state as
  running it once. — A "charge card" call retried after a timeout, with no
  idempotency key, can charge the customer twice if the first attempt actually
  succeeded server-side and only the response was lost; an idempotency key lets the
  server recognize the retry and return the original result instead of repeating the
  charge.
- Generate the idempotency key on the **client**, once per logical operation (not
  per HTTP attempt). — A key generated fresh on every retry defeats the purpose;
  the whole point is that every attempt of the *same* logical operation carries the
  *same* key.

```typescript
// BAD: naive retry loop — no backoff, no jitter, no idempotency key.
// A timeout on a successful charge, retried blindly, can double-charge the
// customer, and a fleet of clients retrying in lockstep can hammer a
// recovering service back down (a "retry storm").
async function chargeCard(order: Order) {
  for (let i = 0; i < 5; i++) {
    try {
      return await paymentApi.charge(order.amount);
    } catch {
      /* retry immediately */
    }
  }
}

// GOOD: idempotency key + exponential backoff with jitter + bounded attempts.
async function chargeCard(order: Order) {
  const idempotencyKey = order.id; // same key on every retry of this order
  for (let attempt = 0; attempt < 5; attempt++) {
    try {
      return await paymentApi.charge(order.amount, { idempotencyKey });
    } catch (err) {
      if (!isRetryable(err) || attempt === 4) throw err;
      const backoff = 2 ** attempt * 100;
      const jitter = Math.random() * backoff;
      await sleep(backoff + jitter);
    }
  }
}
```

## Retry with backoff and jitter, inside a budget

- **Exponential backoff** — space retries out progressively (100ms, 200ms, 400ms...)
  instead of hammering at a constant rate. — Constant-interval retries from many
  clients at once keep load on a struggling service exactly where it started; growing
  the gap gives it room to actually recover.
- **Jitter** — randomize the backoff instead of using the exact computed value. —
  Without jitter, every client that failed at the same moment retries at the exact
  same moment again, turning a transient blip into a synchronized retry storm that
  re-triggers the outage it's trying to recover from.
- **Retry budget** — cap total retries/time, and don't retry non-retryable errors
  (e.g. `400 Bad Request`). — Retrying a request that will deterministically fail
  again (bad input, not-found) wastes latency and load for zero chance of success;
  only retry errors that are plausibly transient (timeouts, `5xx`, connection resets).

## Fail fast with circuit breakers, don't queue hope

- **Circuit breaker** — after enough consecutive failures to a downstream dependency,
  stop calling it for a cooldown period and fail immediately instead. — Without a
  breaker, every caller keeps retrying a downstream service that's already
  overloaded, adding load that delays its recovery; a breaker gives it room to
  recover and gives callers a fast, predictable failure instead of a slow timeout.
- Not every call needs one — reserve breakers for calls to dependencies that can
  actually degrade under load (see "Applying these" below).

## Isolate failure blast radius (bulkheads)

- **Bulkhead** — give different downstream dependencies separate resource pools
  (connection pools, thread pools, queues) instead of sharing one. — If a slow
  third-party API and a fast internal service share one connection pool, the slow
  API alone can exhaust the pool and take the fast internal calls down with it, even
  though the fast service is perfectly healthy.

## Know which consistency trade-off you're making (CAP/PACELC)

- Under a network partition, a system must choose between staying fully
  **consistent** (reject/block requests it can't guarantee are correct) or staying
  **available** (serve requests, possibly with stale/conflicting data) — it cannot
  have both. — Treating "the database will just handle it" as a given, without
  deciding on purpose which side the system falls on for a given operation, means
  the trade-off gets made implicitly and inconsistently across the codebase.
- Most operations don't need strong consistency — decide per-operation. — A "like
  count" can tolerate eventual consistency (briefly stale is fine); a "deduct
  inventory before selling the last unit" usually can't. Applying the same
  consistency model everywhere over- or under-protects different operations.

## Avoid distributed transactions — use saga + outbox instead

- Don't reach for a two-phase-commit/distributed transaction across services. —
  It holds locks across multiple systems for the duration of the whole operation,
  so one slow or down participant blocks (or fails) an operation that has nothing
  architecturally to do with it; it doesn't scale with the number of services
  involved.
- **Saga** — model a multi-step cross-service operation as a sequence of local
  transactions, each with a compensating action to undo it if a later step fails. —
  "Reserve inventory → charge payment → ship order" as a saga means a failed charge
  triggers a compensating "release inventory" step, instead of requiring all three
  services to commit or roll back atomically together, which they structurally
  can't.
- **Transactional outbox** — when a step needs to both update its own data *and*
  publish an event, write the event to an `outbox` table in the *same* local
  transaction as the data change, and have a separate relay process publish it from
  there. — Updating the database and publishing to a message broker as two separate
  operations means a crash between them either loses the event or double-updates
  the data; writing both in one local transaction makes "event published" and "data
  changed" atomic with each other, even though the actual broker publish happens
  asynchronously after.
- A saga trades ACID isolation for availability — its intermediate states are
  visible to the rest of the system before the saga completes. Design each step to
  tolerate that (e.g. mark reserved-but-not-yet-paid inventory as unavailable to
  other orders) rather than assuming the whole saga is invisible until it finishes.

## Applying these under real constraints

- Not every call needs a circuit breaker or bulkhead — reserve them for calls to
  dependencies that can meaningfully degrade or fail under load (external APIs,
  shared databases, anything with its own capacity limit); wrapping a trivial
  in-process function call in a breaker adds complexity with no failure mode it
  actually protects against.
- A system that can't be observed can't be debugged when one of these patterns
  trips — log/emit metrics on every retry, breaker state change, and saga
  compensation, or the pattern silently hides failures instead of surfacing them.
- When reviewing a cross-service call or workflow, name the specific failure mode a
  missing pattern would allow (double-charge, retry storm, cascading outage, torn
  write) rather than citing the pattern name alone.
