---
applyTo: '**/domain/**,**/models/**,**/entities/**,**/aggregates/**'
description: 'Domain-Driven Design principles (ubiquitous language, bounded contexts, aggregates, rich vs anemic models, domain events, anti-corruption layers) to apply when shaping a domain model or service boundary around business concepts.'
---

# Domain-Driven Design Principles

The golden rule: **the domain model is the one place business rules live and are
enforced — not a data bag that services mutate from outside.** Every principle below
either protects that rule or defines the boundary within which one shared model and
one shared vocabulary are allowed to mean the same thing.

This file sits at a different altitude than the general OOP/code-quality principles
(SOLID, coupling/cohesion — see
`.plaesy/instructions/software-design-principles.md`) — DDD is about shaping the domain model and service/module boundaries
around business concepts, not general code hygiene; SOLID's Single Responsibility is
referenced below only where it's directly relevant to an aggregate's job, not
restated.

Sources (retrieved 2026-09-27): Eric Evans, *Domain-Driven Design: Tackling Complexity
in the Heart of Software* (2003) — the foundational source for all terms below —
[Martin Fowler — Bounded Context](https://martinfowler.com/bliki/BoundedContext.html),
[Domain Language — DDD Reference](https://www.domainlanguage.com/wp-content/uploads/2016/05/DDD_Reference_2015-03.pdf),
[Redis — Understanding DDD](https://redis.io/glossary/domain-driven-design-ddd/),
[Wikipedia — Anemic domain model](https://en.wikipedia.org/wiki/Anemic_domain_model)
(Martin Fowler named this anti-pattern in 2003),
[Milan Jovanović — Rich Domain Model vs Anemic Domain Model](https://milanjovanovic.tech/blog/rich-vs-anemic-domain-model).

## Speak one ubiquitous language, inside each bounded context

Use the exact same term for a concept in code, conversation, and documentation as the
domain expert uses — no translation layer between "what the business calls it" and
"what the code calls it." — A codebase that says `Client` while the business says
"Customer" forces every discussion to mentally translate both ways, and the drift
compounds: a new field gets named from whichever term the author happened to reach
for that day, and within a year the two vocabularies have quietly diverged.

## Draw bounded contexts around where a model's meaning changes, not around teams' desks

A **bounded context** is the boundary within which a term means exactly one thing. —
"Customer" in a *Billing* context (a paying account with an invoice history) and
"Customer" in a *Support* context (a person who opened a ticket, possibly never
having paid) are legitimately different models; forcing one shared `Customer` class
to satisfy both makes it either bloated with fields only one context needs, or
subtly wrong for whichever context lost the argument. Two contexts, each with their
own `Customer`, connected by an explicit mapping, is not duplication — it's the
correct boundary.

## Make the aggregate the one gate to a consistency boundary

- An **aggregate root** is the only entry point for external code into an aggregate;
  everything inside is only reachable through it. — Letting external code fetch and
  mutate an `OrderLine` directly (bypassing `Order`) means nothing enforces "an order's
  total must match the sum of its lines" — the invariant lives nowhere once there's a
  second way in.
- One transaction should change exactly one aggregate. — A single database
  transaction that updates both an `Order` aggregate and a `Customer` aggregate's
  loyalty-points balance couples their consistency together for no domain reason;
  if the business tolerates loyalty points updating a moment later, model that as an
  eventually-consistent domain event instead of a shared transaction.

## Keep behavior with the data it protects — avoid the anemic domain model

An **anemic domain model** (Martin Fowler, 2003) is an entity that's just a property
bag — public getters/setters, no self-validation — with all the actual business logic
living in an external service. It pays the structural cost of an object model
(classes, mapping, persistence) while getting none of the benefit (encapsulated,
enforced invariants), and it's the single most common way a domain model rots: once
one service reaches in and mutates a field directly, every other service now has to
duplicate the same validation to stay safe, and eventually one doesn't.

```typescript
// BAD: anemic model — Order is a data bag, invariant enforcement is
// scattered across every service that happens to touch it (or forgotten).
class Order {
  id: string;
  lines: OrderLine[];
  total: number;
  status: string;
}
class OrderService {
  applyDiscount(order: Order, percent: number) {
    order.total = order.total * (1 - percent / 100); // nothing stops this
    // going negative, or stops a second caller from doing the same thing
    // differently a week later
  }
}

// GOOD: rich model — Order enforces its own invariant, no path around it.
class Order {
  private lines: OrderLine[] = [];
  private status: "draft" | "placed" | "cancelled" = "draft";

  applyDiscount(percent: number) {
    if (this.status !== "draft") {
      throw new Error("cannot discount a placed order");
    }
    if (percent < 0 || percent > 100) {
      throw new Error("invalid discount percent");
    }
    this.lines.forEach((l) => l.applyDiscount(percent));
  }

  get total() {
    return this.lines.reduce((sum, l) => sum + l.total, 0);
  }
}
```

## Distinguish entities (identity persists) from value objects (only content matters)

An **entity** is compared by identity (`order.id === other.id`, even if every other
field differs); a **value object** is compared by its content and has no identity of
its own (`Money(10, "USD")` equals any other `Money(10, "USD")`). — Modeling an
address or a monetary amount as an entity (giving it its own ID and a mutable
lifecycle) invites accidental sharing bugs — two orders "sharing" the same address
row mutate each other's address when one is edited; a value object is immutable and
copied, so that can't happen by construction.

## Publish domain events for what already happened, don't ask other aggregates to reach in

A **domain event** (`OrderPlaced`, `PaymentCaptured`) describes something that
already occurred inside an aggregate, published so other parts of the system can
react without the aggregate knowing or caring who's listening. — Without domain
events, keeping inventory in sync with orders means `Order` either directly calls
into `Inventory` (coupling two aggregates' code together, and now `Order` can't be
tested or deployed independently of `Inventory`'s API) or a batch job periodically
reconciles them (introducing a lag nobody explicitly decided to accept); an
`OrderPlaced` event lets `Inventory` subscribe and react on its own terms.

## Protect the model from what's outside it with an anti-corruption layer

An **anti-corruption layer (ACL)** translates between an external system's model (a
third-party API, a legacy system, another team's bounded context) and your own,
sitting entirely on your side of the boundary. — Deserializing a third-party payment
provider's response directly into your own `Payment` domain object means every field
rename or semantic quirk on their side (their `status: "SUCCEEDED"` vs your domain's
`"captured"`) leaks straight into your domain model; an ACL absorbs that translation
in one place instead of every call site guessing at the mapping independently.

## Applying these under real constraints

- DDD's ceremony (bounded contexts, aggregates, ACLs, ubiquitous-language glossaries)
  is a cost paid for handling genuine domain complexity — a simple CRUD admin panel
  with no real business rules doesn't need an aggregate root; it needs a table and a
  form. Reach for these when the domain has real invariants and real complexity to
  protect, not by default.
- A bounded context needs a real organizational or team boundary behind it to be
  worth maintaining, not just a folder split — a context boundary that one team
  crosses freely in practice (editing "both" models in the same PR without a
  translation step) isn't actually a boundary; either enforce it or stop pretending
  it exists.
- When reviewing a domain model, name which principle is at risk and the concrete
  consequence (a service reaching past an aggregate root, two bounded contexts'
  models silently merging, business logic found in a service instead of the entity
  it's about) rather than "this isn't DDD."
