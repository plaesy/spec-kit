---
applyTo: '**'
description: 'Core software design principles (SOLID, DRY, KISS, YAGNI, separation of concerns, composition over inheritance, coupling/cohesion) to apply when writing or reviewing any code, regardless of language.'
---

# Software Design Principles

Every principle below exists to push code toward two properties: **high cohesion**
(each unit does one clearly-related thing) and **low coupling** (units depend on as
little of each other's internals as possible). When a rule below and a rule elsewhere
conflict, prefer whichever gets closer to that pair — these are forces to balance, not
laws to apply blindly regardless of context.

Sources (retrieved 2026-09-27): [BMC — SOLID Design Principles](https://www.bmc.com/blogs/solid-design-principles/),
[freeCodeCamp — SOLID Design Principles](https://www.freecodecamp.org/news/solid-design-principles-in-software-development/),
[DigitalOcean — SOLID: the first five principles of OOD](https://www.digitalocean.com/community/conceptual-articles/s-o-l-i-d-the-first-five-principles-of-object-oriented-design),
[Scalastic — SOLID, DRY, KISS...](https://scalastic.io/en/solid-dry-kiss/).
SOLID was formalized by Robert C. Martin in his 2000 essay "Design Principles and
Design Patterns."

## SOLID (object-oriented / component design)

- **Single Responsibility** — a unit (class, module, function) should have one reason
  to change. — When a `UserService` both validates input, talks to the database, and
  formats API responses, a UI-formatting tweak risks breaking database logic that has
  nothing to do with it; split by responsibility instead.
- **Open/Closed** — extend behavior by adding new code, not by editing code that
  already works. — A `switch` on a `type` field that every new case forces you to
  revisit is a sign to replace it with polymorphism (strategy/plugin) so adding a case
  means adding a file, not editing a shared one.
- **Liskov Substitution** — a subtype must be usable anywhere its base type is
  expected, without surprising the caller. — A `Bird` base class with a `fly()` method
  breaks the moment `Penguin extends Bird`; the fix is a different abstraction
  (`FlightlessBird`), not a `fly()` override that throws.
- **Interface Segregation** — don't force a caller to depend on methods it never
  calls. — One fat `Worker` interface with `cook()`, `clean()`, and `code()` forces a
  `Cleaner` implementation to fake-implement `code()`; split into `Cook`, `Cleaner`,
  `Coder`.
- **Dependency Inversion** — depend on an abstraction (interface/protocol), not a
  concrete implementation. — A class that directly `new`s a `PostgresClient` can't be
  unit-tested without a real database; depend on a `Repository` interface and inject
  the concrete client at the boundary.

```typescript
// BAD: violates SRP + DIP — mixes formatting, persistence, and a hard-coded dependency
class UserService {
  save(user: User) {
    const db = new PostgresClient(); // concrete dependency baked in
    db.query("INSERT INTO users ..."); // persistence
    return `Welcome, ${user.name}!`; // presentation, unrelated to saving
  }
}

// GOOD: one reason to change per class, dependency injected as an abstraction
interface UserRepository {
  save(user: User): void;
}
class UserService {
  constructor(private repo: UserRepository) {}
  save(user: User) {
    this.repo.save(user);
  }
}
```

## DRY — Don't Repeat Yourself

Eliminate duplicated **knowledge**, not just duplicated lines. — Two functions that
happen to look similar but encode unrelated business rules are not a DRY violation;
copy-pasted validation logic that must change in lockstep in three places *is* — the
risk is one copy gets updated and the others silently drift out of sync. Applies to
code, config, and tests alike, not only source files.

## KISS — Keep It Simple

Prefer the simplest design that satisfies the actual requirement. — A generic plugin
system for a feature with exactly two fixed variants adds indirection with no present
payoff; write the two branches directly, and only generalize once a third variant
actually shows up.

## YAGNI — You Aren't Gonna Need It

Don't build for a requirement that doesn't exist yet. — An extra config flag,
abstraction layer, or parameter added "in case we need it later" is a cost paid today
(more to read, more to test, more to keep correct) against a benefit that may never
arrive; add it when the real need appears, not before.

## Separation of Concerns

Keep distinct responsibilities in distinct layers — data access, business logic, and
presentation should not live in the same function. — A React component that both
fetches from an API and renders JSX makes the fetch logic untestable without a
renderer and the render logic untestable without a mocked network; split fetch into a
hook/service, keep the component about rendering.

## Composition over Inheritance

Prefer combining small, focused objects/functions over building deep inheritance
hierarchies for code reuse. — A `FlyingCar extends Car, Plane` (or an equivalent
multi-level inheritance chain) tightly couples a subclass to every implementation
detail of its ancestors; composing a `Car` that *has a* `FlightModule` keeps each
piece independently replaceable and testable.

## Law of Demeter (principle of least knowledge)

A unit should only talk to its immediate collaborators, not reach through them into
their internals. — `order.getCustomer().getAddress().getCity()` couples the caller to
three levels of another object's structure; any of those return types changing shape
breaks this call. Prefer asking the immediate collaborator to do the work:
`order.getBillingCity()`.

## Applying these under real constraints

- These principles trade off against each other and against delivery speed — e.g.
  strict SRP can fragment a trivial CRUD module into more files than it needs. Apply
  proportionally to the actual complexity and expected churn of the code, not
  uniformly everywhere.
- Prefer refactoring toward a principle when a **second** violation of the same kind
  appears (duplicated logic, a class picking up a second responsibility), rather than
  pre-abstracting on the first instance — this keeps YAGNI and DRY in balance instead
  of fighting each other.
- When reviewing code, flag violations by naming which principle is at stake and the
  concrete failure mode it enables (as in the examples above) — not by citing the
  acronym alone.
