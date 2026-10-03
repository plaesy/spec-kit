---
applyTo: '**/routes/**,**/controllers/**,**/*.openapi.yaml,**/*.openapi.yml,**/*openapi*.json,**/*.api.ts,**/api/**/*.ts,**/api/**/*.py,**/api/**/*.go'
description: 'Core HTTP/REST API design principles (resource modeling, idempotency, versioning, pagination, error contract, backward compatibility) to apply when designing or reviewing an API surface.'
---

# API Design Principles

The golden rule: **an API is a contract with clients you cannot see and cannot force
to upgrade.** Every principle below exists to protect that contract, or to change it
in a way that doesn't break whoever is already depending on it.

Sources (retrieved 2026-09-27): [Moesif — 12 REST API Best Practices That Hold Up in 2026](https://www.moesif.com/blog/technical/api-development/essential-REST-API-best-practices/),
[OneUptime — REST API Design Best Practices for Production Services](https://oneuptime.com/blog/post/2026-02-20-api-design-rest-best-practices/view),
[DreamFactory — Best Practices for Naming REST API Endpoints](https://blog.dreamfactory.com/best-practices-for-naming-rest-api-endpoints),
[InfoWorld — How to make your REST APIs backward-compatible](https://www.infoworld.com/article/2261134/how-to-make-your-rest-apis-backward-compatible.html),
[AgileSeekers — API Backward Compatibility: Versioning Patterns and Examples](https://agileseekers.com/blog/handling-backward-compatibility-in-versioned-product-apis).

## Model resources as nouns, let HTTP methods carry the verb

- Endpoints name a resource (plural noun); the HTTP method says what happens to it. —
  `/getUser`, `/deleteUser` duplicate the verb the method already carries, and every
  new action demands a new endpoint name; `GET /users/{id}` and
  `DELETE /users/{id}` reuse one resource name across every operation on it.
- Keep collection and item endpoints grammatically consistent (always plural: `/users`,
  `/users/{id}`). — Mixing `/user` (singular, collection) with `/users/{id}` (plural,
  item) forces every client-code generator and every new contributor to guess which
  form a given resource uses.

```http
BAD:
POST /createOrder
POST /getOrderById
POST /cancelOrder/123

GOOD:
POST   /orders            # create
GET    /orders/123        # read
DELETE /orders/123        # cancel/remove
```

## Make unsafe operations idempotent

Give write operations that can be safely retried (especially `POST`, and any payment
or order-creation call) an idempotency key the client generates and the server
deduplicates on. — When a client's `POST /orders` times out, the client cannot tell
whether the order was created or not; without an idempotency key, retrying either
risks a duplicate order or leaves the client stuck not retrying at all. `PUT` and
`DELETE` are idempotent by HTTP definition (repeating them produces the same end
state) — don't build them in a way that violates that (e.g. a `PUT` that appends
instead of replaces).

## Version the contract, don't mutate it silently

- Pick one versioning mechanism (URL path `/v1/…`, a version header, or date-based
  pinning) and apply it everywhere — the specific mechanism matters less than
  applying it consistently. — A service where some endpoints are versioned and others
  aren't leaves clients unable to reason about what's safe to assume will keep
  working.
- Never change the *meaning* of an existing field, status code, or response shape in
  place — that's a breaking change regardless of whether the version number moved. —
  An endpoint that returns `200` on failure today and starts returning `500` tomorrow,
  under the same version, breaks every client that branches on status code without
  any signal that it should re-check its assumptions.
- Add new optional fields freely; treat removing a field, renaming a field, changing
  a field's type, or making an optional parameter required as breaking. — A client
  that only reads `email` from a response is unaffected by a new `phone` field
  appearing; the same client breaks the moment `email` is renamed to
  `email_address` or removed.

## Paginate collections, and don't leak internal ID structure

Use cursor-based pagination with an opaque token for any collection endpoint, rather
than raw offset or a bare sequential ID as the cursor. — A cursor value like
`?after=12345` tells every caller your primary keys are sequential integers and
breaks the moment the table is sharded or IDs switch to UUIDs; an opaque
(e.g. base64-encoded) cursor token can change its internal encoding later without
any client noticing.

## One consistent error contract, everywhere

Return errors in one fixed shape (e.g. `{ code, message, details }`, or the RFC 9457
`application/problem+json` structure) from every endpoint, never a shape that varies
by which handler happened to write it. — A client integrating against an API with
five different ad-hoc error shapes across five endpoints needs five different parsers
just for failure handling; one consistent shape needs exactly one.

```json
BAD (shape varies by endpoint):
{ "error": "not found" }
{ "message": "Invalid input", "field": "email" }

GOOD (one shape everywhere):
{ "code": "RESOURCE_NOT_FOUND", "message": "Order 123 not found", "details": {} }
{ "code": "VALIDATION_ERROR", "message": "email is invalid", "details": { "field": "email" } }
```

Never repurpose an HTTP status code's established meaning — if an endpoint returns
`404` for "not found" today, it must keep meaning that; encode finer-grained failure
reasons in the error body's `code`, not by inventing a new meaning for a status code.

## Applying these under real constraints

- A brand-new, unreleased API with zero consumers can break its own contract freely —
  these principles bind hardest once real clients exist; don't over-engineer
  versioning or idempotency for an endpoint nobody outside the team has integrated
  with yet.
- When a breaking change is genuinely necessary, prefer deprecating the old shape
  alongside the new one for a announced window over forcing a hard cutover — this
  trades short-term duplication for not breaking every client simultaneously.
- When reviewing an API change, name which contract guarantee is at risk (idempotency,
  backward compatibility, error-shape consistency) and the concrete client behavior
  that breaks, not just "this is a breaking change."

## Scaffolding a New Spec File

- OpenAPI: one YAML file with `paths`, `components.schemas`,
  `components.responses` (shared error shapes), and a `components.securitySchemes`
  block matching the project's real auth (`oauth2`, `apiKey`, `http bearer`) —
  never a placeholder scheme left unconfigured
- GraphQL SDL: types + a dedicated `Connection`/edges pattern for any list field
  that needs cursor pagination, mirroring the offset/cursor choice made above for
  REST; mutations return a payload type (`{ result, errors }`), not the bare
  mutated object, so field-level errors have somewhere to live
- Either format: examples on every schema/type used in a request or response —
  an unexampled schema is unreviewable and ungenerable (client SDKs, mocks)
- Generate the file at the path the project's own convention expects
  (`specs/*.yaml`, `schema.graphql`, etc.) rather than inventing a new location
