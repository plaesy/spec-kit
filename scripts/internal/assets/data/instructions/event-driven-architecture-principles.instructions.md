---
applyTo: '**/events/**,**/*event*,**/*.consumer.ts,**/*.producer.ts,**/handlers/**,**/*.handler.ts'
description: 'Core event-driven architecture principles (event vs command, schema/versioning, choreography vs orchestration, ordering, coupling via fat events) for designing pub-sub/messaging systems — complements distributed-systems-design-principles.instructions.md rather than repeating its idempotency/retry/saga mechanics.'
---

# Event-Driven Architecture Principles

The golden rule: **an event is a public contract the moment a second consumer exists.**
Every principle below exists to keep that contract loosely coupled and survivable under
duplicate, out-of-order, or delayed delivery — the default behavior of any real message
broker.

For idempotency, retry/backoff, circuit breakers, CAP trade-offs, and the saga/outbox
pattern, see [[distributed-systems-design-principles]] — those mechanics apply to
event-driven systems too and are not repeated here. This file covers what's specific to
designing *with events* as the integration style: event shape, topology, and ordering.

Sources (retrieved 2026-09-27): [Conduktor — Event-Driven Architecture](https://www.conduktor.io/glossary/event-driven-architecture),
[Serverless Land — Choreography and orchestration with event-driven architectures](https://serverlessland.com/event-driven-architecture/choreography-and-orchestration),
[DEV Community — Event-Driven Architecture: Events, Commands, and the Tradeoffs Nobody Mentions][ref20],
[Baxchain — Resilient EDA: Idempotency, Retries, and Dead-Letter Queues](https://baxchain.com/blogs/resilient-event-driven-architecture-idempotency-retries-and-dead-letter-queues/),
[digitalapplied — Event-Driven Architecture & Message Queues: 2026 Reference](https://www.digitalapplied.com/blog/event-driven-architecture-message-queues-2026-engineering-reference).

## Distinguish an event from a command

An **event** ("OrderPlaced") states a fact that already happened and implies nothing
about who must act on it; a **command** ("PlaceOrder") tells one specific receiver what
to do. — Naming something `ChargeCard` and publishing it to a shared topic makes every
future consumer of that topic a potential (unintended) charger of cards; if it's really
an instruction for one service, send it directly to that service, not onto a broadcast
topic dressed up as an event.

## Keep events thin — publish facts, not internal state dumps

- Prefer a **thin/notification event** (what happened + an ID to look up detail) over a
  **fat event** (the full state snapshot embedded in the payload), unless you have a
  specific reason to accept the coupling that comes with the fat version. — A fat
  `OrderPlaced` event carrying every field of the order schema means every consumer
  that reads `event.payload.shippingAddress.postalCode` breaks the moment that internal
  field is renamed; a thin event carrying `{ orderId }` and requiring a lookup keeps
  consumers coupled to a stable read API instead of to internal storage shape.
- If you do choose event-carried state transfer for resilience (a consumer that must
  keep working when the source service is down), treat the event schema as a
  first-class public API with its own versioning discipline — not as "whatever fields
  happened to be on the internal object when we serialized it."

```json
// BAD: fat event — every consumer now depends on internal order-service fields
{
  "eventType": "OrderPlaced",
  "order": { "id": "o_123", "lineItems": [...], "shippingAddress": {...},
             "internalRiskScore": 0.42, "warehouseRoutingHint": "WH-04" }
}

// GOOD: thin event — a fact plus a reference, consumers fetch what they need
{
  "eventType": "OrderPlaced",
  "orderId": "o_123",
  "occurredAt": "2026-09-27T10:00:00Z"
}
```

## Version the event schema like an API contract

Add new optional fields freely; treat removing a field, renaming a field, or changing a
field's type as a breaking change that needs a new event version or a migration
window. — A consumer deployed last month that reads `event.total` breaks silently the
moment a producer renames that field to `event.totalAmount` without warning, because
nothing forces the two deploys to happen together in an async system. Put a schema
registry or a documented contract in front of every topic that has more than one
consumer team, so a breaking change is caught before it ships, not after a consumer
starts failing in production.

## Choose choreography vs. orchestration deliberately, per workflow

- **Choreography** (each service reacts to events with no central coordinator) fits
  simple, independent fan-outs — notifications, analytics, cache invalidation. —
  Forcing a multi-step business process (reserve → charge → ship, with compensations)
  into pure choreography scatters the "what happens next" logic across every
  participating service, so no single place shows the whole workflow or can answer
  "where did this order get stuck?"
- **Orchestration** (a central coordinator drives the sequence) fits workflows where
  ordering, visibility, or compensation logic matters. — Reintroducing a central
  point when only a one-hop notification was ever needed adds an unnecessary
  coordinator and a new single point of failure for a flow that didn't need one.
- Enterprise systems typically need both, applied per workflow, not one style forced
  onto everything.

## Design for at-least-once delivery — consumers must be idempotent

Assume every event can be delivered more than once (broker retries, consumer restarts
mid-processing) and design the consumer's handler to tolerate that, not just the
producer's publish. — A consumer that runs `inventory -= qty` directly on receipt
double-decrements stock the second time the same event is redelivered; track processed
event IDs (or make the operation naturally idempotent, e.g. `SET status = 'shipped'`)
so replaying the same event is a no-op the second time. (Mechanics: see
[[distributed-systems-design-principles]].)

## Don't assume ordering unless you've engineered for it

Most brokers guarantee order only within a partition/shard key, not globally across a
topic. — Two events for the *same* user processed out of order (`AddressUpdated` then
`AddressUpdated` again, but the older one arrives second) can silently overwrite the
newer state with stale data; if a consumer's correctness depends on order, partition
the topic by an entity ID (e.g. `userId`) so all events for that entity land on the
same partition and are processed in order — and treat "no ordering guarantee" as the
default anywhere you haven't done that.

## Route unprocessable events to a dead-letter queue, don't drop or block on them

A message that fails processing after a bounded number of retries goes to a monitored
dead-letter queue for inspection/replay — it doesn't retry forever, and it doesn't
silently vanish. — Without a DLQ, one malformed "poison" message either blocks the
partition behind it forever (if the consumer keeps retrying in place) or gets silently
dropped (if it doesn't) — a DLQ turns that single bad message into an alert someone can
act on, instead of an outage or silent data loss.

## Watch for the "distributed monolith via events" anti-pattern

Publishing and subscribing to events doesn't automatically decouple services — if
every service subscribes to (and depends on the exact shape of) every other service's
internal events, a schema change anywhere still requires coordinated redeploys
everywhere, exactly like a tightly-coupled synchronous call chain would. — A dozen
services all parsing the same fat `OrderUpdated` event for whichever few fields each
one needs recreates tight coupling through a message broker instead of through direct
calls; thin events plus deliberate per-consumer ownership of what each event contains
keeps the decoupling real, not just async-flavored.

## Applying these under real constraints

- Not every integration needs to be an event. A synchronous request/response call is
  simpler to reason about, trace, and debug than an async event chain — reach for
  events when you specifically need decoupling, fan-out to multiple consumers, or
  resilience to a downstream outage, not by default for every service-to-service call.
- Schema governance (a registry, a review step for new event fields) costs process
  overhead — scale it to how many independent consumer teams actually depend on a
  given topic; a topic with one producer and one consumer on the same team needs far
  less ceremony than a company-wide "OrderPlaced" topic.
- When reviewing an event-driven design, name the specific coupling or failure mode at
  stake (fat-event coupling, missing idempotency, unordered processing, no DLQ) rather
  than "this isn't really decoupled" — that's what makes the finding actionable.

[ref20]: https://dev.to/arnavsharma2711/event-driven-architecture-explained-events-commands-and-the-tradeoffs-nobody-mentions-2ki4
