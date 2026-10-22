# Service Development Specification: {{SERVICE_NAME}}

**Created**: {{DATE}}
**Status**: {{DOC_STATUS|draft}}
**Type**: API or microservice
**Source request**: {{SOURCE_REQUEST|verbatim text that prompted this specification}}
**Spec owner**: {{SPEC_OWNER}}
**Constitution dimensions in force**: {{ACTIVE_DIMENSIONS|Technical}}
**Related specs**: {{RELATED_SPEC_IDS|spec directory numbers this specification depends on}}

## Working constraints

Each constraint below binds the delivered service, not this document. Every one is
either satisfied by a named section further down or it is not.

- The contract is written before the implementation, and it is the contract.
- Every failure mode has a defined response code, payload, and log line.
- Availability is a number with an error budget, not an adjective.
- Authentication and authorisation exist in the first increment, never after it.
- Logging, metrics, and tracing are part of the service, not an add-on.
- Consistency across services is a declared pattern (transactional outbox, saga,
  idempotent consumer), never an accident of call ordering.
- A service that cannot be deployed independently is not a service yet.

## Service Overview

### Core purpose

{{CORE_PURPOSE|one paragraph: what this service owns, and the business value of owning it}}

### Boundaries

| Aspect | Definition | Consequence if the boundary is wrong |
|--------|------------|--------------------------------------|
| Domain owned | {{DOMAIN_OWNED|bounded context this service is the source of truth for}} | {{DOMAIN_BOUNDARY_RISK}} |
| Responsibilities | {{RESPONSIBILITIES|functions this service performs}} | {{RESPONSIBILITY_RISK}} |
| Explicitly not owned | {{NOT_OWNED|adjacent data this service reads but never writes}} | {{NOT_OWNED_RISK}} |
| Upstream dependencies | {{UPSTREAM_DEPENDENCIES|services or systems this service calls}} | {{UPSTREAM_RISK}} |
| Downstream consumers | {{DOWNSTREAM_CONSUMERS|services or clients that call this one}} | {{DOWNSTREAM_RISK}} |

### Capabilities

| # | Capability | Input | Output | Consumer |
|---|------------|-------|--------|----------|
| C-01 | {{CAPABILITY_1}} | {{CAPABILITY_1_INPUT}} | {{CAPABILITY_1_OUTPUT}} | {{CAPABILITY_1_CONSUMER}} |
| C-02 | {{CAPABILITY_2}} | {{CAPABILITY_2_INPUT}} | {{CAPABILITY_2_OUTPUT}} | {{CAPABILITY_2_CONSUMER}} |
| C-03 | {{CAPABILITY_3}} | {{CAPABILITY_3_INPUT}} | {{CAPABILITY_3_OUTPUT}} | {{CAPABILITY_3_CONSUMER}} |

## API Design

### Protocol

| Property | Value | Rationale |
|----------|-------|-----------|
| Style | {{API_STYLE\|REST, GraphQL, or gRPC}} | {{API_STYLE_RATIONALE}} |
| Base path | {{API_BASE_PATH}} | {{API_BASE_PATH_RATIONALE}} |
| Authentication | {{API_AUTH}} | {{API_AUTH_RATIONALE}} |
| Media type | {{API_MEDIA_TYPE\|application/json}} | {{API_MEDIA_TYPE_RATIONALE}} |
| Versioning policy | {{API_VERSIONING}} | {{API_VERSIONING_RATIONALE}} |
| Deprecation policy | {{API_DEPRECATION}} | {{API_DEPRECATION_RATIONALE}} |
| Rate limit policy | {{API_RATE_LIMIT}} | {{API_RATE_LIMIT_RATIONALE}} |
| Pagination | {{API_PAGINATION}} | {{API_PAGINATION_RATIONALE}} |

### Endpoint catalogue

| Method | Path | Capability | Auth | Idempotent |
|--------|------|------------|------|------------|
| POST | /auth/login | C-01 | {{ENDPOINT_LOGIN_AUTH}} | no |
| POST | /auth/refresh | C-01 | {{ENDPOINT_REFRESH_AUTH}} | yes |
| GET | /auth/me | C-01 | {{ENDPOINT_ME_AUTH}} | yes |
| GET | /resources | C-02 | {{ENDPOINT_LIST_AUTH}} | yes |
| POST | /resources | C-02 | {{ENDPOINT_CREATE_AUTH}} | no |
| GET | /resources/{{ID}} | C-02 | {{ENDPOINT_GET_AUTH}} | yes |
| PUT | /resources/{{ID}} | C-02 | {{ENDPOINT_UPDATE_AUTH}} | yes |
| DELETE | /resources/{{ID}} | C-02 | {{ENDPOINT_DELETE_AUTH}} | yes |
| GET | /health | operations | none | yes |
| GET | /metrics | operations | {{METRICS_AUTH}} | yes |

### Request and response schemas

| Endpoint | Field | Type | Required | Validation | Example |
|----------|-------|------|----------|------------|---------|
| POST /auth/login | email | string | yes | {{LOGIN_EMAIL_RULE\|RFC 5322 format}} | {{LOGIN_EMAIL_EXAMPLE}} |
| POST /auth/login | password | string | yes | {{LOGIN_PASSWORD_RULE\|minimum length and entropy}} | {{LOGIN_PASSWORD_EXAMPLE\|redacted}} |
| POST /resources | name | string | yes | {{CREATE_NAME_RULE}} | {{CREATE_NAME_EXAMPLE}} |
| POST /resources | tags | array of string | no | {{CREATE_TAGS_RULE}} | {{CREATE_TAGS_EXAMPLE}} |
| GET /resources | limit | integer | no | {{LIST_LIMIT_RULE}} | {{LIST_LIMIT_EXAMPLE}} |
| GET /resources | cursor | string | no | {{LIST_CURSOR_RULE}} | {{LIST_CURSOR_EXAMPLE}} |

Response envelope: {{RESPONSE_ENVELOPE|shape every success response follows}}
Identifier format: {{ID_FORMAT|UUIDv4, ULID, or other; and which one}}

### Error catalogue

| Code | HTTP status | When it is returned | Client action | Retried |
|------|-------------|----------------------|---------------|---------|
| {{ERROR_CODE_1}} | {{ERROR_STATUS_1}} | {{ERROR_CONDITION_1}} | {{ERROR_CLIENT_ACTION_1}} | {{ERROR_RETRY_1}} |
| {{ERROR_CODE_2}} | {{ERROR_STATUS_2}} | {{ERROR_CONDITION_2}} | {{ERROR_CLIENT_ACTION_2}} | {{ERROR_RETRY_2}} |
| {{ERROR_CODE_3}} | {{ERROR_STATUS_3}} | {{ERROR_CONDITION_3}} | {{ERROR_CLIENT_ACTION_3}} | {{ERROR_RETRY_3}} |

Every error body carries a machine-readable code, a human-readable message, and
the request's trace id: {{ERROR_BODY_SHAPE}}.

## Data and Consistency

### Persistence

| Store | Purpose | Engine | Consistency | Backup and restore |
|-------|---------|--------|-------------|---------------------|
| {{STORE_1}} | {{STORE_1_PURPOSE}} | {{STORE_1_ENGINE}} | {{STORE_1_CONSISTENCY}} | {{STORE_1_BACKUP}} |
| {{STORE_2}} | {{STORE_2_PURPOSE}} | {{STORE_2_ENGINE}} | {{STORE_2_CONSISTENCY}} | {{STORE_2_BACKUP}} |

### Schema

| Table | Key | Columns | Indexes | Growth per day |
|-------|-----|---------|---------|-----------------|
| {{TABLE_1}} | {{TABLE_1_KEY}} | {{TABLE_1_COLUMNS}} | {{TABLE_1_INDEXES}} | {{TABLE_1_GROWTH}} |
| {{TABLE_2}} | {{TABLE_2_KEY}} | {{TABLE_2_COLUMNS}} | {{TABLE_2_INDEXES}} | {{TABLE_2_GROWTH}} |

### Migrations

| Rule | Decision |
|------|----------|
| Tooling | {{MIGRATION_TOOL\|Flyway, Liquibase, Alembic, or equivalent}} |
| Ordering | {{MIGRATION_ORDERING}} |
| Reversibility | {{MIGRATION_REVERSIBILITY\|every migration has a tested rollback, or the exceptions are named}} |
| Staging rehearsal | {{MIGRATION_STAGING_REHEARSAL}} |
| Deployment gate | {{MIGRATION_GATE}} |

### Cross-service consistency

| Concern | Pattern chosen | Why | Failure behaviour |
|---------|----------------|-----|-------------------|
| Local transactions | {{LOCAL_TX_PATTERN}} | {{LOCAL_TX_RATIONALE}} | {{LOCAL_TX_FAILURE}} |
| Cross-service writes | {{DISTRIBUTED_TX_PATTERN\|transactional outbox, saga, or none}} | {{DISTRIBUTED_TX_RATIONALE}} | {{DISTRIBUTED_TX_FAILURE}} |
| Event publication | {{EVENT_PUBLICATION_PATTERN}} | {{EVENT_PUBLICATION_RATIONALE}} | {{EVENT_PUBLICATION_FAILURE}} |
| Idempotency | {{IDEMPOTENCY_MECHANISM}} | {{IDEMPOTENCY_RATIONALE}} | {{IDEMPOTENCY_FAILURE}} |

## Security

| Control | Mechanism | Verified by |
|---------|-----------|-------------|
| Authentication | {{AUTHN_METHOD\|JWT bearer, mTLS, API key}} | {{AUTHN_VERIFICATION}} |
| Token handling | {{TOKEN_HANDLING|issuer, audience, lifetime, refresh, revocation}} | {{TOKEN_VERIFICATION}} |
| Authorisation model | {{AUTHZ_MODEL\|RBAC, ABAC, or scoped claims}} | {{AUTHZ_VERIFICATION}} |
| Permission matrix | {{PERMISSION_MATRIX\|resource:action to role mapping}} | {{PERMISSION_VERIFICATION}} |
| Input validation | {{INPUT_VALIDATION}} | {{INPUT_VALIDATION_VERIFICATION}} |
| Transport security | {{TLS_POLICY}} | {{TLS_VERIFICATION}} |
| Secrets | {{SECRETS_HANDLING}} | {{SECRETS_VERIFICATION}} |
| Audit logging | {{AUDIT_EVENTS\|security-relevant actions recorded}} | {{AUDIT_VERIFICATION}} |
| Disclosure response | {{DISCLOSURE_RESPONSE}} | {{DISCLOSURE_VERIFICATION}} |

## Performance and Scaling

### Service level objectives

| SLI | Objective | Window | Measured by |
|-----|-----------|--------|-------------|
| Availability | {{AVAILABILITY_TARGET\|99.9%}} | {{AVAILABILITY_WINDOW\|rolling 30 days}} | {{AVAILABILITY_MEASUREMENT}} |
| Latency | {{LATENCY_TARGET\|p95 under 200ms}} | {{LATENCY_WINDOW}} | {{LATENCY_MEASUREMENT}} |
| Throughput | {{THROUGHPUT_TARGET}} | {{THROUGHPUT_WINDOW}} | {{THROUGHPUT_MEASUREMENT}} |
| Error budget | {{ERROR_BUDGET}} | {{ERROR_BUDGET_WINDOW}} | {{ERROR_BUDGET_MEASUREMENT}} |

### Resource envelope

| Resource | Baseline | Limit | Basis for the number |
|----------|----------|-------|----------------------|
| Memory | {{MEMORY_BASELINE}} | {{MEMORY_LIMIT}} | {{MEMORY_BASIS}} |
| CPU | {{CPU_BASELINE}} | {{CPU_LIMIT}} | {{CPU_BASIS}} |
| Storage | {{STORAGE_BASELINE}} | {{STORAGE_LIMIT}} | {{STORAGE_BASIS}} |
| Connections | {{CONNECTION_BASELINE}} | {{CONNECTION_LIMIT}} | {{CONNECTION_BASIS}} |

### Caching

| Layer | Cached data | Pattern | TTL | Invalidation trigger |
|-------|-------------|---------|-----|----------------------|
| {{CACHE_LAYER_1}} | {{CACHE_1_DATA}} | {{CACHE_1_PATTERN\|cache-aside, write-through, write-behind}} | {{CACHE_1_TTL}} | {{CACHE_1_INVALIDATION}} |
| {{CACHE_LAYER_2}} | {{CACHE_2_DATA}} | {{CACHE_2_PATTERN}} | {{CACHE_2_TTL}} | {{CACHE_2_INVALIDATION}} |

Stale-cache behaviour: {{STALE_CACHE_BEHAVIOUR|what a consumer sees when invalidation fails}}

### Scaling

| Direction | Mechanism | Trigger | Limit reached |
|-----------|-----------|---------|---------------|
| Horizontal | {{HORIZONTAL_SCALING}} | {{HORIZONTAL_TRIGGER}} | {{HORIZONTAL_LIMIT}} |
| Vertical | {{VERTICAL_SCALING}} | {{VERTICAL_TRIGGER}} | {{VERTICAL_LIMIT}} |
| Read path | {{READ_SCALING}} | {{READ_TRIGGER}} | {{READ_LIMIT}} |

## Observability

### Structured log fields

| Field | Format | Purpose |
|-------|--------|---------|
| timestamp | {{LOG_TIMESTAMP_FORMAT\|RFC 3339 UTC}} | {{LOG_TIMESTAMP_PURPOSE}} |
| level | {{LOG_LEVELS}} | {{LOG_LEVEL_PURPOSE}} |
| service | {{SERVICE_NAME}} | Correlates the line with the emitting instance |
| trace_id | {{LOG_TRACE_ID}} | Joins the line to the request's trace |
| span_id | {{LOG_SPAN_ID}} | Identifies the operation within the trace |
| message | {{LOG_MESSAGE}} | Human-readable event description |
| context | {{LOG_CONTEXT\|structured key-value fields}} | {{LOG_CONTEXT_PURPOSE}} |

Log destination and retention: {{LOG_DESTINATION}}

### Metrics

| Metric | Type | Labels | Used to answer |
|--------|------|--------|----------------|
| {{METRIC_1_NAME}} | {{METRIC_1_TYPE\|counter, gauge, histogram}} | {{METRIC_1_LABELS}} | {{METRIC_1_QUESTION}} |
| {{METRIC_2_NAME}} | {{METRIC_2_TYPE}} | {{METRIC_2_LABELS}} | {{METRIC_2_QUESTION}} |
| {{METRIC_3_NAME}} | {{METRIC_3_TYPE}} | {{METRIC_3_LABELS}} | {{METRIC_3_QUESTION}} |

### Tracing

Propagation format: {{TRACE_PROPAGATION\|W3C tracecontext}}
Spans sampled at entry: {{TRACE_SPANS}}

### Health and readiness

| Endpoint | Reports | Used by | Response budget |
|----------|---------|---------|------------------|
| /health/live | {{LIVE_CHECKS}} | {{LIVE_CONSUMER\|orchestrator liveness probe}} | {{LIVE_BUDGET}} |
| /health/ready | {{READY_CHECKS\|dependency reachability}} | {{READY_CONSUMER\|orchestrator readiness probe and load balancer}} | {{READY_BUDGET}} |
| /health/startup | {{STARTUP_CHECKS}} | {{STARTUP_CONSUMER}} | {{STARTUP_BUDGET}} |

### Alerts

| Alert | Condition | Severity | Runbook | Owner |
|-------|-----------|----------|---------|-------|
| {{ALERT_1_NAME}} | {{ALERT_1_CONDITION}} | {{ALERT_1_SEVERITY}} | {{ALERT_1_RUNBOOK}} | {{ALERT_1_OWNER}} |
| {{ALERT_2_NAME}} | {{ALERT_2_CONDITION}} | {{ALERT_2_SEVERITY}} | {{ALERT_2_RUNBOOK}} | {{ALERT_2_OWNER}} |
| {{ALERT_3_NAME}} | {{ALERT_3_CONDITION}} | {{ALERT_3_SEVERITY}} | {{ALERT_3_RUNBOOK}} | {{ALERT_3_OWNER}} |

An alert with no runbook is a notification, not a control: {{ALERT_RUNBOOK_POLICY}}

## Runtime and Deployment

| Property | Value |
|----------|-------|
| Base image | {{BASE_IMAGE}} |
| Build command | {{BUILD_COMMAND}} |
| Start command | {{START_COMMAND}} |
| Configuration source | {{CONFIG_SOURCE}} |
| Port | {{SERVICE_PORT}} |
| Orchestrator | {{ORCHESTRATOR}} |
| Replicas | {{REPLICA_COUNT\|with the scaling range}} |
| Resource requests and limits | {{RESOURCE_REQUESTS}} |
| Rollout strategy | {{ROLLOUT_STRATEGY\|rolling, canary, or blue-green}} |
| Rollback procedure | {{ROLLBACK_PROCEDURE}} |

### Pipeline stages

| # | Stage | Gate that must pass before the next stage |
|---|-------|---------------------------------------------|
| 1 | {{PIPELINE_STAGE_1}} | {{PIPELINE_GATE_1}} |
| 2 | {{PIPELINE_STAGE_2}} | {{PIPELINE_GATE_2}} |
| 3 | {{PIPELINE_STAGE_3}} | {{PIPELINE_GATE_3}} |
| 4 | {{PIPELINE_STAGE_4}} | {{PIPELINE_GATE_4}} |
| 5 | {{PIPELINE_STAGE_5}} | {{PIPELINE_GATE_5}} |

## Testing Strategy

| Level | Scope | Target | Tools |
|-------|-------|--------|-------|
| Unit | Business logic and data models | {{TEST_COVERAGE_MIN\|90}}% | {{UNIT_TEST_TOOLS}} |
| Contract | Producer and consumer agree on every schema | {{CONTRACT_TEST_SCOPE\|every published endpoint}} | {{CONTRACT_TEST_TOOLS}} |
| Integration | Database, queues, and external services | {{INTEGRATION_COVERAGE_TARGET}} | {{INTEGRATION_TEST_TOOLS}} |
| End to end | Critical business flows across services | {{E2E_FLOW_COUNT}} flows | {{E2E_TEST_TOOLS}} |
| Load and soak | Behaviour under target and peak load | {{LOAD_PROFILE}} | {{LOAD_TEST_TOOLS}} |
| Resilience | Dependency failure and retry behaviour | {{RESILIENCE_SCENARIOS}} | {{RESILIENCE_TEST_TOOLS}} |

- [ ] {{TEST_SCENARIO_1|e.g. invalid credentials return the documented code}}
- [ ] {{TEST_SCENARIO_2}}
- [ ] {{TEST_SCENARIO_3}}
- [ ] {{TEST_SCENARIO_4}}

## Open Questions

| # | Question | Why it matters | Owner | Blocking what | Resolution date |
|---|----------|----------------|-------|----------------|-----------------|
| Q1 | {{QUESTION_1}} | {{QUESTION_1_IMPACT}} | {{QUESTION_1_OWNER}} | {{QUESTION_1_BLOCKS}} | {{QUESTION_1_DUE}} |
| Q2 | {{QUESTION_2}} | {{QUESTION_2_IMPACT}} | {{QUESTION_2_OWNER}} | {{QUESTION_2_BLOCKS}} | {{QUESTION_2_DUE}} |
| Q3 | {{QUESTION_3}} | {{QUESTION_3_IMPACT}} | {{QUESTION_3_OWNER}} | {{QUESTION_3_BLOCKS}} | {{QUESTION_3_DUE}} |
| Q4 | {{QUESTION_4}} | {{QUESTION_4_IMPACT}} | {{QUESTION_4_OWNER}} | {{QUESTION_4_BLOCKS}} | {{QUESTION_4_DUE}} |

## Completion Checklist

Nothing below is true when this file is created. Each box is ticked only when the
thing it names has actually happened; an unticked box is the honest state.

- [ ] Every section above is filled in, not merely headed.
- [ ] Every `{{PLACEHOLDER}}` token in this document has been replaced.
- [ ] Every endpoint has an owner, an auth rule, and a defined error catalogue entry.
- [ ] Every SLO carries a measurement and a window.
- [ ] Every consistency decision names the failure behaviour, not only the happy path.
- [ ] Every alert names a runbook.
- [ ] Every assumption is labelled as fact, inferred, speculative, or unknown.
- [ ] Every threshold used here matches the active-dimension values in the constitution.

## Handover

| Item | Value |
|------|-------|
| Next artifact | design.template.md |
| Reviewers | {{REVIEWERS}} |
| Approval required from | {{APPROVERS}} |
| Estimated timeline | {{ESTIMATED_TIMELINE}} |
| Infrastructure requirements | {{INFRASTRUCTURE_REQUIREMENTS}} |
