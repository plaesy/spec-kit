---
applyTo: '**/services/**,**/*.ts,**/*.go,**/*.py,**/*.java,**/*.cs'
description: 'Core observability design principles (structured logging, correlation/trace IDs, cardinality-aware metrics, SLO-based alerting) for designing systems that can be understood in production, beyond performance-baseline.instructions.md''s measurement mechanics.'
---

# Observability Design Principles

The golden rule: **design to answer the question you haven't thought to ask yet, not
just the ones you predicted.** Monitoring watches for failures you anticipated;
observability is what lets you explain the ones you didn't — the dividing line is
whether your telemetry preserves enough per-request detail to be sliced by any
attribute after the fact, or only pre-aggregated numbers for the cases you thought of
in advance.

This file is about designing systems to be debuggable in production generally; for
performance-specific measurement/baselining mechanics, see
`.plaesy/instructions/performance-baseline.md`.

Sources (retrieved 2026-09-27): [IBM — Three Pillars of Observability](https://www.ibm.com/think/insights/observability-pillars),
[Grepr — Structured Logging Best Practices for Production in 2026](https://www.grepr.ai/blog/structured-logging-best-practices),
[Last9 — Correlation ID vs Trace ID](https://last9.io/blog/correlation-id-vs-trace-id/),
[CloudTweaks — Monitoring vs Observability: Which Catches Unknown Failures First](https://cloudtweaks.com/2026/07/monitoring-vs-observability-which-catches-unknown-failures-first/),
[O'Reilly — Observability Engineering, ch. 13: Acting on and Debugging SLO-Based Alerts](https://www.oreilly.com/library/view/observability-engineering/9781492076438/ch13.html).

## Emit structured logs, not free text

Log in a structured format (JSON or equivalent) with consistent field names, never an
interpolated sentence. — A log line like `"User 123 failed to checkout: timeout"`
can only be found later by grepping for a substring you have to guess in advance;
`{ event: "checkout_failed", user_id: 123, reason: "timeout" }` can be filtered,
aggregated, and correlated automatically by every log platform, with no regex
required.

```text
BAD:  log.info(`User ${userId} failed to checkout: ${reason}`)
GOOD: log.info("checkout_failed", { user_id: userId, reason, order_id: orderId })
```

## Propagate one correlation ID through every hop

Assign a correlation/trace ID at the request's entry point, and pass it through every
downstream call and every log line for that request. — When an error report says
"a checkout failed around 14:32," but the request touched five services with no
shared ID, there is no way to reassemble which of the thousands of log lines across
those five services belong to that one request; a single propagated ID makes it one
query instead of an investigation.

```text
BAD:  service B receives a call from service A with no request/trace header —
      B's logs for that request are indistinguishable from every other request B handled.

GOOD: A sends X-Correlation-Id (and/or a W3C traceparent header) to B; B logs it on
      every line and forwards it to C; one ID ties every log/span across A, B, C together.
```

## Pick the right pillar for the question, don't default to one

- **Metrics** answer "how is the system doing right now, in aggregate" (rate, error %,
  latency percentile) — cheap to store, but they've already lost the detail of any
  single request. — A metric showing "1% error rate" can't tell you which customer, or
  which input, triggered the failing 1%.
- **Logs** answer "what exactly happened for this one event." — Useful once you know
  roughly where to look; expensive to keep at full detail and full retention forever.
- **Traces** answer "where did the time/failure go across service boundaries" for one
  request. — A trace shows a checkout took 4s because a downstream inventory call
  took 3.8s of it; a metric alone only shows "checkout is slow," not where.

Design telemetry with all three available, and route each question to the pillar that
actually answers it — reaching for logs to answer an aggregate-rate question, or for a
dashboard to debug one specific user's one specific failure, wastes time either way.

## Manage cardinality deliberately, don't let it explode by accident

Never attach an unbounded value (raw user ID, full URL with query string, free-text
error message) as a metric *label* — put unbounded values in logs/traces, not in
metric dimensions. — A `request_duration{url="/orders/48291?ref=email-campaign-7"}`
label creates a new time series per unique URL; at scale this silently multiplies
storage cost and can take down the metrics backend itself. Use a low-cardinality label
instead (`route="/orders/{id}"`) and keep the specific ID in the log line or trace
span, where it belongs.

## Alert on user-facing symptoms, not on every possible cause

Page a human on SLO burn rate / user-visible impact (error rate, latency breaching a
target), not on every internal signal that *might* precede a problem (CPU%, queue
depth in isolation). — Alerting on CPU at 80% pages someone for a server that is
working exactly as intended under load and causing zero user impact; that page trains
the on-call engineer to start ignoring alerts, so the *next* page — the one that
actually matters — gets dismissed too. Alert on the thing users actually experience,
and let internal signals feed dashboards/investigation instead of pages.

## Applying these under real constraints

- Full distributed tracing on every internal call has real cost (both infra spend and
  instrumentation effort) — prioritize tracing the request paths that actually cross
  multiple services and matter for user-facing latency, not every function call.
- Retention is a cost decision, not a technical default: keep high-cardinality/
  high-detail data (raw logs, full traces) for a shorter window and let cheaper
  aggregated metrics cover the long-term trend.
- When reviewing telemetry code, name the specific question it will fail to answer
  during an incident (as in the examples above) — "we should add more logging" isn't
  actionable; "this handler has no correlation ID, so a failure here can't be tied back
  to the request that caused it" is.
