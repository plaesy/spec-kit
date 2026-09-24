---
description: "Generate API boilerplate and specification files (OpenAPI YAML or GraphQL SDL) with modular components, authentication, and pagination support"
---

# `/create:api` command instructions

This is the **api**-scoped entry point into `/create`. It produces real specification files (OpenAPI YAML or GraphQL SDL), not descriptions—complete with schemas, components, examples, and security definitions ready for code generation.

## Usage Format

```bash
/create:api "user-management-api" --format openapi --auth oauth2 --out specs/api.yaml
/create:api "product-catalog" --format graphql --version 2.0 --auth jwt --paginate cursor
/create:api --design-spine .plaesy/memory/design-spine.md --resource users posts comments --methods GET POST PUT DELETE
/create:api "payment-service" --format openapi --version 3.2.0 --auth api-key --paginate offset --out specs/payment-api.yaml
```

| Flag | Default | Meaning |
|---|---|---|
| `description` or first arg | — | API name, title, or one-line purpose |
| `--format` | `openapi` | Output specification format: `openapi` (YAML), `graphql` (SDL), or `asyncapi` |
| `--version` | `1.0.0` (API version) | Semantic version for the generated API |
| `--spec-version` | `3.2.0` (OpenAPI) or `SDL` (GraphQL) | Specification standard version (OpenAPI 3.0.0–3.2.0, GraphQL SDL) |
| `--auth` | `api-key` | Authentication scheme: `api-key`, `oauth2`, `jwt`, `bearer`, `basic` |
| `--paginate` | `offset` | Pagination strategy: `offset`, `cursor`, `relay` (GraphQL connection style) |
| `--methods` | `GET POST` (OpenAPI) | HTTP methods to include (ignored for GraphQL): `GET POST PUT DELETE PATCH` |
| `--resources` | `users` (generic) | Resource models to scaffold (e.g., `users posts comments`) |
| `--out` | `specs/api.{yaml\|graphql}` | Output file path relative to project root |
| `--design-spine` | `.plaesy/memory/design-spine.md` | Path to design-spine file for API naming conventions, auth patterns |
| `--validate` | `true` | Run Spectral linting on output (OpenAPI only) |
| `--n` | `1` | Number of variants to generate (each with `-1`, `-2` suffix) |

## Protocol

### Step 1: Resolve the API Brief

- If `--design-spine` given: read that file, extract API naming conventions (camelCase fields, PascalCase types), authentication patterns, resource models, and pagination preferences.
- If `--resources` given: build the resource list from that; otherwise use sensible defaults (e.g., `users`, `items`, `events`).
- If `--auth` and `--paginate` not specified, check `.plaesy/memory/api-design-spine.md` for project defaults.
- Compose the API brief: `{description} — {resources list} — {auth scheme} — {pagination} — {HTTP methods or GraphQL operations}`.
- Layer in project API design standards when available:
  - `.plaesy/memory/api-design-spine.md` — naming conventions, auth patterns, error response formats
  - `.plaesy/memory/design-spine.md` — if it contains API sections (base paths, server definitions)
  - `.plaesy/roles/technical.md` — API design principles (idempotency, rate limiting, versioning)

### Step 2: Resolve the Provider & Format

Read format preference, in this order, stop at first match:
1. `--format` CLI flag (`openapi` | `graphql` | `asyncapi`)
2. `.plaesy/scripts/configs/api-format.json` → `{"format": "openapi", "specVersion": "3.2.0"}`
3. Default: `openapi` with spec version `3.2.0`

For OpenAPI: ensure `--spec-version` is one of `3.0.0`, `3.0.1`, `3.0.2`, `3.0.3`, `3.1.0`, `3.2.0` (current best practice per OpenAPI Specification v3.2.0).
For GraphQL: SDL output (no version flag needed; if provided, ignore).

### Step 3: Generate Specification

Construct the API specification file with these sections (OpenAPI YAML structure):

```yaml
openapi: 3.2.0
info:
  title: {API name from brief}
  version: {--version flag}
  description: {full brief}
servers:
  - url: https://api.example.com/v1
components:
  securitySchemes:
    {auth scheme}: {OAuth2/JWT/API-Key definition}
  schemas:
    {Resource}: {JSON Schema per resource, with properties, required, examples}
paths:
  /resources:
    get: {list operation with pagination, sorting}
    post: {create operation with input schema}
  /resources/{id}:
    get: {retrieve by ID}
    put: {replace}
    delete: {remove}
security:
  - {auth scheme}: []
```

For GraphQL (SDL):
```graphql
type Query {
  {resources}(first: Int!, after: String): {Resource}Connection!
}

type Mutation {
  create{Resource}(input: Create{Resource}Input!): {Resource}!
}

type {Resource} {
  id: ID!
  # fields per resource schema
}

input Create{Resource}Input {
  # input fields
}
```

Apply these standards (from OpenAPI Specification v3.2.0 & GraphQL best practices):
- Use camelCase for field names, PascalCase for type/schema names, UPPERCASE for enums
- Include detailed `description` fields on all schemas, operations, and parameters
- Add `example` values for every schema (realistic, minimal)
- Implement modular components via `$ref` (e.g., `$ref: '#/components/schemas/Error'`) to avoid duplication
- Define error responses (400, 401, 403, 404, 500) with schema references
- Include pagination parameters (`limit`/`offset` or `first`/`after` for cursor)
- Set `required` arrays for mandatory fields
- Define `responses` object with `application/json` content type per operation

For pagination (OpenAPI):
- **Offset**: `limit` (default 20) + `offset` (default 0) query parameters
- **Cursor**: `first` (default 20) + `after` (null for first page) parameters
- Cursor variant returns `pageInfo: {hasNextPage, endCursor}`

For pagination (GraphQL):
- **Relay Connection Pattern**: `{Resource}Connection`, `{Resource}Edge`, `PageInfo`
- Always include `first`, `after`, `last`, `before` arguments on list queries
- Return `edges: [{node, cursor}]` + `pageInfo`

### Step 4: Index & Update Design Spine

Write or update `.plaesy/memory/api-design-spine.md` with:
- API name and version
- Authentication scheme + token/key location (header, query param)
- Base URL/server definition
- Pagination pattern used
- Resource models defined
- Naming conventions applied (camelCase, PascalCase, UPPERCASE)
- Link to generated specification file

If `.plaesy/memory/api-design-spine.md` does not exist, create it with the template from `.plaesy/instructions/how-to-create-designmd.md`.

### Step 5: Validate & Report

- If `--validate true` (default) and format is OpenAPI:
  - Run Spectral CLI: `spectral lint --ruleset=recommended {output-file}` (requires Spectral installed or skip with graceful fallback).
  - Surface validation errors/warnings, or report "validation passed" if all clear.
- Confirm the output file exists and is valid YAML (OpenAPI) or valid GraphQL SDL (GraphQL).
- Report:
  - Output file path (e.g., `specs/payment-api.yaml`)
  - Format and specification version (e.g., `OpenAPI 3.2.0`)
  - Resources scaffolded (e.g., `users, products, orders`)
  - Authentication scheme (e.g., `OAuth 2.0`)
  - Pagination pattern (e.g., `cursor-based`)
  - Validation status (passed, failed with errors, or skipped if Spectral missing)
  - API Design Spine update path (`.plaesy/memory/api-design-spine.md`)

If validation fails, surface the Spectral errors verbatim; do not attempt retry loops.

## Programmatic Invocation (Called By Other Prompts)

Prompts that need an API spec mid-run (`/implement:api`, `/implement:graphql`, `/doc`, `/assess:technical`) call this protocol directly:

```
CALL /create:api
  description: "<API name or purpose>"
  format: "openapi" | "graphql"
  auth: "<auth scheme>"
  paginate: "offset" | "cursor" | "relay"
  resources: [<resource-1>, <resource-2>, …]
  out: "<path the calling prompt will reference>"
RETURNS
  path: <written file path, or null if generation was skipped>
  format: "openapi" | "graphql"
  resources: [<generated resource list>]
  auth_scheme: "<auth scheme>"
  validation_status: "passed" | "failed" | "skipped"
  design_spine_updated: true | false
```

If no write permissions or Spectral validation fails, return `path: null` and `validation_status: "failed"` with error message. The calling prompt must treat this as "spec generation pending" and surface the issue, never silently drop the spec or fabricate a path.

## Anti-Patterns (NEVER Do These)

- ❌ Generate a specification without example values — examples are essential for testing and documentation
- ❌ Skip the modular components (`$ref`) pattern — copy-pasting schemas leads to inconsistency and maintenance debt
- ❌ Use generic naming (e.g., `data`, `info`, `result`) instead of explicit field names — be specific per resource
- ❌ Omit error response definitions — every endpoint must define 4xx/5xx error schemas
- ❌ Forget pagination on list endpoints — unprotected lists are a DoS vector; always paginate
- ❌ Mix authentication schemes in one API — pick one (API key, OAuth2, JWT, etc.) and apply consistently
- ❌ Omit `required` field lists — let consumers know which fields are mandatory vs. optional
- ❌ Generate a GraphQL schema without a `Query` root type or mutation inputs — incomplete schema breaks resolvers
- ❌ Use UPPERCASE for field names or camelCase for type names — violate the naming convention standard
- ❌ Validate with Spectral and ignore errors — surface all linting issues to the caller
- ❌ Commit the generated spec without updating `.plaesy/memory/api-design-spine.md` — future API work depends on this reference
- ❌ Retry provider calls on validation failure — report the error once and stop; let the user fix the issue

## Integration Points

- **With `/implement:api`**: `/create:api` scaffolds the spec; `/implement:api` reads it and generates server stubs, business logic, and tests
- **With `/implement:graphql`**: `/create:api --format graphql` produces the SDL schema; `/implement:graphql` implements resolvers
- **With `/doc`**: Generated spec is used to produce interactive API documentation via Scalar UI or ReDoc
- **With `/assess:technical`**: Specification is validated against OpenAPI/GraphQL standards via quality gates
- **With `.plaesy/memory/api-design-spine.md`**: Stores project API naming conventions, auth patterns, and pagination strategies for reuse
- **With Spectral CLI** (optional): Lints OpenAPI specs for conformance to rules defined in `.spectral.yaml`

## Design-Spine & Memory Integration

If `.plaesy/memory/api-design-spine.md` exists:
1. Extract the project's API naming rules (camelCase fields, PascalCase types, UPPERCASE enums)
2. Extract defined authentication schemes and apply consistently
3. Extract pagination preference (offset, cursor, relay) and use as default
4. Extract list of approved resources/models and include in generation

If the file does not exist, create it during Step 4 (Index & Update Design Spine) with sensible defaults.

## Success Criteria

An API specification is complete when:
- ✅ Specification file exists (OpenAPI YAML or GraphQL SDL)
- ✅ `info`/`title`/`version` defined (OpenAPI) or `type Query` + mutations present (GraphQL)
- ✅ All resources have schemas with properties, types, and examples
- ✅ Authentication scheme defined (`securitySchemes` in OpenAPI, auth directives in GraphQL)
- ✅ List endpoints include pagination parameters (`limit`/`offset` or `first`/`after`)
- ✅ Error responses defined (400, 401, 403, 404, 500 with schemas)
- ✅ Naming conventions applied (camelCase fields, PascalCase types, UPPERCASE enums)
- ✅ Modular components used (`$ref` references) to avoid duplication
- ✅ Spectral validation passed (if enabled and Spectral installed)
- ✅ `.plaesy/memory/api-design-spine.md` created or updated with API metadata

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
