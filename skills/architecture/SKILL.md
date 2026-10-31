---
name: architecture
description: "Architecture design principles: API, caching, DevOps, CLI, concurrency, database, distributed systems, domain-driven design, observability, resilience, event-driven, monorepo, and general software design. Use when designing or reviewing a system; load only the file that matches the decision."
---

# Architecture Principles

Principles for designing and reviewing systems. Each file states the principle, the
failure it prevents, and when it applies. Load only the file for the decision at hand.

| File | Load when |
|------|-----------|
| `references/software-design.md` | Writing or reviewing any code: SOLID, DRY, KISS, YAGNI, coupling and cohesion |
| `references/api-design.md` | Designing or reviewing an HTTP/REST surface: resources, idempotency, versioning, pagination, errors |
| `references/database-design.md` | Designing a schema, migration, or data model: normalization, keys, indexing |
| `references/caching.md` | Adding or changing a cache: invalidation, keys, stampede prevention, what not to cache |
| `references/concurrency.md` | Code running multiple threads or tasks: shared state, locks, deadlock, worker pools |
| `references/distributed-systems.md` | Cross-service calls: idempotency, retries, circuit breakers, CAP, sagas |
| `references/domain-driven-design.md` | Shaping a domain model or service boundary: bounded contexts, aggregates |
| `references/observability.md` | Making a system understandable in production: logs, trace IDs, metrics, SLO alerts |
| `references/cli-design.md` | Designing or reviewing a command-line tool: streams, exit codes, machine-readable output |
| `references/resilience.md` | Designing for failure: graceful degradation, timeouts, graceful shutdown, load shedding, chaos experiments |
| `references/event-driven.md` | Designing messaging or pub-sub systems: event vs command, schema versioning, choreography vs orchestration |
| `references/devops.md` | Designing CI/CD pipelines, deployment workflows, or delivery metrics: CALMS culture, DORA metrics |
| `references/monorepo.md` | Setting up a JavaScript/TypeScript monorepo: workspace layout, lockfile discipline, pnpm pitfalls |

Security architecture is in `../security/references/principles.md`.
