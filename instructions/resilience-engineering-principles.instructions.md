---
applyTo: '**'
description: 'Principles for deliberately designing and verifying resilience (graceful degradation, timeouts, graceful shutdown, load shedding, chaos experiments) — one layer above the individual patterns in distributed-systems-design-principles.instructions.md.'
---

# Resilience Engineering Principles

The golden rule: **resilience is a property you design and verify, not one you hope
for.** A system that has never had a dependency deliberately killed under controlled
conditions has an untested assumption, not a resilience guarantee.

This file is about designing for failure *proactively* and *proving* the design holds
(chaos experiments, graceful degradation, load shedding, timeouts, graceful shutdown) —
a layer above individual reliability patterns. For idempotency, retries/backoff,
circuit breakers, CAP trade-offs, and saga/outbox, see
[[distributed-systems-design-principles]]; this file doesn't re-explain those.

Sources (retrieved 2026-09-27): [Principles of Chaos Engineering](https://principlesofchaos.org/),
[Uptime Labs — What Is Chaos Engineering? Core Principles](https://www.uptimelabs.io/learn/what-is-chaos-engineering),
[Kunal Ganglani — Chaos engineering principles & steady-state hypotheses](https://www.kunalganglani.com/learning-paths/sre/sre-chaos-principles),
[OneUptime — Graceful Shutdown in Go for Kubernetes](https://oneuptime.com/blog/post/2026-01-07-go-graceful-shutdown-kubernetes/view),
[Webalert — Graceful Shutdown and SIGTERM: Deploy Without Dropping Requests](https://web-alert.io/blog/graceful-shutdown-sigterm-zero-downtime-deploys-guide).

## Define steady state before you break anything

State the system's normal, healthy behavior as a measurable metric (e.g. "checkout
success rate ≥ 99.5%, p99 latency ≤ 400ms") *before* running a chaos experiment, and
form a hypothesis that this steady state holds even when a specific failure is
injected. — Killing a dependency "to see what happens" with no baseline defined means
you can't tell whether anything actually broke, or the system was already degraded
before you touched it; a stated steady state turns "seemed fine" into a pass/fail
result you can act on.

## Start experiments with the smallest possible blast radius

Run a new chaos experiment against a single instance/pod, in staging or against a
small slice of production traffic with an immediate abort switch — never full
production on the first run. — Injecting a network partition against every replica
of a payment service at once, with no rollback path, turns a learning exercise into
the outage it was supposed to help prevent; expand scope only after the small
experiment has already built confidence that the abort mechanism and the hypothesis
both hold.

## Give every remote call a timeout and a deadline, not an infinite wait

A call with no timeout is a resource leak waiting to happen: a hung downstream
dependency slowly consumes every thread/connection in the caller's pool until the
caller itself stops responding to anything, unrelated request included. Prefer
propagating one deadline through a whole request's call chain over setting an
independent timeout at each hop — otherwise a chain of 300ms timeouts across 5
sequential hops adds up to a 1.5s wait the original caller never agreed to.

```typescript
// BAD: no timeout — a hung dependency can block this call (and this thread/
// connection) forever, starving every other caller sharing the pool.
async function getInventory(sku: string) {
  return fetch(`https://inventory.internal/sku/${sku}`);
}

// GOOD: explicit timeout, deadline propagated to the downstream call.
async function getInventory(sku: string, deadline: AbortSignal) {
  return fetch(`https://inventory.internal/sku/${sku}`, {
    signal: AbortSignal.any([deadline, AbortSignal.timeout(500)]),
  });
}
```

## Shut down by draining, never by dropping

On `SIGTERM` (deploy, scale-down, autoscaler eviction): stop accepting new work
first, let in-flight requests finish within a bounded grace period, then exit — don't
hard-kill the process the moment the signal arrives. — A process that exits
immediately on `SIGTERM` drops every request that was mid-flight at that exact
instant, on every single deploy; a orchestrator (Kubernetes, an ALB) typically already
gives a grace period specifically so the process can drain into it — not using that
window turns routine deploys into a steady trickle of dropped requests.

```typescript
// BAD: exits immediately, dropping every in-flight request.
process.on("SIGTERM", () => process.exit(0));

// GOOD: stop accepting new work, drain what's in flight, then exit — with a
// hard ceiling so a stuck request can't block shutdown forever.
process.on("SIGTERM", async () => {
  server.close(); // stop accepting new connections
  await Promise.race([drainInFlightRequests(), sleep(10_000)]);
  process.exit(0);
});
```

## Degrade gracefully — partial function beats total outage

When a non-critical dependency is unavailable, serve the request with reduced
functionality instead of failing the whole request. — A product page that 500s
entirely because the "recommended items" service is down takes an unrelated,
fully-working checkout flow down with it; the same page returning the product details
with recommendations simply omitted keeps the primary function working and confines
the failure to the part that actually failed.

## Shed load on purpose, before the system falls over on its own

When incoming load exceeds capacity, deliberately reject the excess (fast, explicit
`503`/`429` responses) rather than letting every request degrade together as queues
grow unbounded. — A service that accepts every request and lets its queue grow
without limit doesn't serve everyone slowly — past a point it serves *no one*, because
memory/connections exhaust and the whole process falls over; explicitly rejecting the
requests above a known-safe threshold keeps the requests under that threshold fast
and correct, instead of every request (accepted or not) sharing in the collapse.

## Applying these under real constraints

- Chaos experiments carry real risk if blast-radius control is skipped — never run
  one without an abort mechanism, and never run a first-time experiment directly
  against unconstrained production traffic, regardless of how confident the hypothesis
  feels.
- Not every dependency needs a graceful-degradation path — reserve it for
  dependencies whose failure shouldn't fail the whole request; a dependency that truly
  is required for correctness (e.g. the payment step of checkout) should fail loudly,
  not silently degrade into an inconsistent state.
- When reviewing resilience work, name the specific untested assumption or dropped
  work a gap allows (no timeout → thread-pool exhaustion, no drain → dropped requests
  on every deploy, no load shedding → full collapse under a traffic spike) rather than
  "this isn't resilient."
