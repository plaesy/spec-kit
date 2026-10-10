---
description: "Operations assessment workflow - deployment pipelines, release management, reliability/SLOs, observability, infrastructure as code, incident response and runbooks"
applyTo: "**/*"
---

# Operations Assessment Instructions

**Use with**: `/assess:operations` when dimension is **Operations**

**Referenced from**: `/assess` (orchestrator)

---

## Scope Definition

Operations owns the questions **"how does this get into production, and how do we
know it is working while it is there?"**:

- **Deployment**: how a build becomes a running system — pipeline stages, promotion
  between environments, environment parity, rollback and rollback verification
- **Release**: versioning, release cadence, change windows, release approval,
  communication of a release, deprecation policy
- **Reliability**: SLO/SLI definitions, error budgets, capacity headroom, dependency
  failure behaviour, blast radius, availability design as it is *operated*
- **Observability**: structured logs, metrics, traces, dashboards, alert quality
  (actionability, cardinality, alert fatigue), on-call signal
- **Infrastructure as Code**: whether the environment is reproducible from a repo,
  config/secrets handling, drift, state safety
- **Operational readiness**: runbooks, incident response, on-call rotation, MTTR,
  post-incident review, backup/restore and disaster recovery *exercises*

### Boundary — what does NOT belong here

The routing table in `dimension-mapping.md` is the authority; these are the
overlaps that caused findings to be scored twice, so state them once, here:

| Finding | Dimension | Not |
|---------|-----------|-----|
| CI/CD runs, the code is deployable, rollouts work | **Operations** | Technical |
| Code quality, test coverage, dependency CVEs in libraries, OWASP High/Critical in the source | **Technical** | Operations |
| Infrastructure/IaC, DR, HA, secrets management, monitoring — as deployed and operated | **Operations** | Technical |
| Cloud architecture, scalability of the *design*, code-level caching/concurrency | **Technical** | Operations |
| Incident response *process*, on-call *rotation*, change approval, runbooks, MTTR, post-incident review | **Operations** | Management |
| Team capacity, skills, backlog, org structure, hiring, morale, span of control | **Management** | Operations |
| Licence/compliance obligations attached to an artefact (legal review of a licence, GDPR handling) | **Legal** | Operations |

A deployment finding is scored **once**, here. If a finding is really two findings
(split it in the report), do not count the same evidence under two dimensions.

---

## Assessment Workflow

### Step 1: Operational Surface Inventory

- Detect deployment artefacts: CI/CD definitions, container files, IaC modules,
  chart/release directories, platform config in `scripts/configs/`
- Map the environments that exist (dev/staging/pre-prod/prod) and how each is
  created — manually, from IaC, or not at all
- Identify the runtime: services, jobs, schedulers, serverless functions, third-party
  dependencies the system cannot run without

**Rate**: record the inventory; no score yet — it is the input to the steps below

### Step 2: Deployment & Pipeline Review

- Trace one real change from commit to production: what triggers what, what is
  manual, what is gated
- Check pipeline stage ordering, failure handling (does a red stage block?), and
  whether a failed deploy leaves a half-applied environment
- Verify environment parity: config, schema/migrations, feature flags between
  staging and production; anything that only exists in one
- Verify rollback: is it documented, automated, and has it been executed before?
  "We can roll back" without a record of a successful rollback is **unverified**
- Check migration safety: reversible? backward-compatible with the running version?

**Rate**: 0-100 based on deployment automation and rollback maturity

### Step 3: Release Management Review

- Version scheme and whether it is applied consistently
- Release cadence, change-window policy, and what happens to a hotfix
- Change approval path: who approves, is it recorded, is it auditable
- Release notes / changelog discipline, and deprecation + customer communication
  policy for a breaking change

**Rate**: 0-100 based on release predictability and traceability

### Step 4: Reliability & SLO Review

- Are SLIs defined from user-visible behaviour (not CPU), and are SLOs written down
  with a target and a window?
- Is there an error budget, and does anything actually consume it (a freeze when
  exhausted)?
- Capacity: headroom, limits/throttling, and what the first limit to hit is
- Dependency failure behaviour in the *running* system: timeouts, retries with
  backoff, circuit breakers, bulkheads
- Single points of failure: one instance, one zone, one person, one credential

**Rate**: 0-100 based on measured/derivable reliability posture

### Step 5: Observability Review

- Logs: structured, with a correlation/trace id that survives a request across
  services; check for secrets/PII in log output
- Metrics: the golden signals for this system's user journey; cardinality bounded
- Traces: is there propagation end to end, and is anything sampled away that matters
- Dashboards: one a responder can open first — is it current or does it reference
  removed metrics?
- Alerts: are they actionable and owned? Count unactionable pages as a finding, and
  note the alerting channel/rotation
- Do the checks actually run, or does a check that cannot execute report as passing?

**Rate**: 0-100 based on whether an incident could be diagnosed from the telemetry

### Step 6: Infrastructure as Code Review

- Is the environment reproducible from a repo? Any manually-created resource is drift
- Config and secret handling: are secrets outside source control, rotated, and scoped
  least-privilege
- State safety: destroy/apply ordering, drift detection, and change review on IaC
- Unpinned versions/tags in IaC and images (a "latest" tag is a finding)
- Cost visibility: is there any tag/label/report that makes spend attributable?

**Rate**: 0-100 based on reproducibility and change safety

### Step 7: Operational Readiness Review

- Runbooks: for the top alerts and the top failure modes, does a runbook exist, is it
  current, and is it reachable at 03:00?
- Incident response: severity definition, escalation path, comms owner
- On-call: rotation, load, handoff, and whether on-call is sustainable
- MTTD/MTTR: measured, or asserted? Report asserted-only numbers as unverified
- Backup and restore: does a restore have been *exercised*, with a date
- Post-incident review: does it happen, and are actions tracked to completion?

**Rate**: 0-100 based on demonstrated readiness

### Step 8: Generate Report & Routing Decision

**Report sections**: canonical names per `.plaesy/instructions/dimension-mapping.md`
→ *Assessment Report Section Registry*; the bullets below go *inside* those
sections and do not rename them.

- Overall operations score (0-100) and per-category scores
- Top 5 findings, each with the evidence that supports it
- For each finding: which dimension owns it, per the boundary table above

**Routing**:

- Deployment/release blocked, no runbook, no restore drill → `/implement:operations`
- Detected but badly tuned (noisy alerts, no SLO, cache/deadline problems in the
  running system) → `/optimize:operations`
- Incident in flight, or a hard availability/consistency break → `/fix:operations`
- If the finding turns out to be a team/process issue, route to
  `/implement:management`; if it is a code defect, `/fix:technical`
- All categories at or above `{{QUALITY_REMEDIATION_FLOOR}}` (85) → proceed to
  `/optimize`; below it, the work is remediation and belongs to `/fix` or
  `/loop` first. Defined in `.plaesy/instructions/quality-gates.md`; do not
  restate the number here.

### Uncertainty Surfacing (Before Finalizing Assessment)

Load `.plaesy/instructions/uncertainty-surfacing.md`; do not restate it here.
Only this dimension's specifics follow.

**Confidence basis for this dimension**:

- HIGH: the running system, pipeline config, or incident record was read directly
- MEDIUM: configuration or dashboard inspected but the live behaviour was not observed
- LOW: the environment is not reachable and no configuration was available — nothing was verified

Operations has an unusually wide gap between HIGH and LOW: reading a pipeline
definition is not evidence that the pipeline works, and a dimension that cannot
reach a running environment has verified nothing about that environment. Say so
rather than scoring it MEDIUM on the strength of a YAML file.

**Worked example**:

```text
I could not assess rollback reliability because the staging cluster was not
reachable and no restore-drill record exists in the repository.
To improve confidence, provide the most recent rollback drill record, or access
to the staging cluster.
```

---

## Quality Scoring

### Operations Assessment Components

```text
deployment_pipeline: 20%   # commit→prod path, gating, env parity, rollback, migrations
release_management: 15%   # versioning, cadence, approvals, changelog, deprecation
reliability_slo: 20%       # SLIs/SLOs, error budget, capacity, dependency failure, SPFs
observability: 20%         # logs/metrics/traces, dashboards, alert quality, ownership
infrastructure_as_code: 15% # reproducibility, config/secrets, state safety, pinning, cost
operational_readiness: 10% # runbooks, incident response, on-call, MTTR, backup/restore, PIR
```

Weights sum to 100. Apply the formula strictly — do not adjust a weight to move a
score. If a category cannot be assessed (no pipeline exists at all, no IaC), score
it 0 and state why; do not silently drop its weight and rescale the rest, because
that hides absence of evidence as a neutral score.

### Grade Mapping

| Grade | What it means here |
|---|---|
| A+ | Exceptional - deploy is boring, rollbacks are routine, incidents are diagnosable |
| A | Excellent - automated, measurable, minor gaps |
| B+ | Very Good - solid pipeline, some observability or runbook gaps |
| B | Good - works, evidence is thin in places |
| C | Acceptable - manual steps and unverified rollbacks |
| D | Needs Work - releases are risky and outages are not diagnosable |

The score edges behind these letters, and the general meaning of each band,
are defined once in `.plaesy/instructions/quality-gates.md` →
**Grade bands**. Do not restate them here.

### Success Criteria (Hard Stops)

- MUST HAVE: a documented, repeatable path from commit to production
- MUST HAVE: a rollback that has been executed at least once, with a date
- MUST HAVE: alerting wired to a named rotation, with no unactionable pages
- MUST HAVE: runbooks for the top failure modes, or an explicit acknowledged gap
- MUST HAVE: no secret committed to an IaC module, manifest, or environment file

---

## Critical Rules

- ✅ **EVIDENCE OVER ASSERTION** - a claimed rollback, restore, or drill counts only
  with a record (a log, a run, a dated ticket); otherwise mark it UNVERIFIED
- ✅ **SCOPE EACH FINDING ONCE** - use the boundary table; a deployment finding
  belongs to Operations, not to Technical and Management as well
- ✅ **CITE WITH LOCATION** - every finding names the file (and line, or resource
  name) it came from
- ✅ **MEASURE DON'T ESTIMATE** - MTTR, coverage of alerts, and error-budget burn
  come from the system or are reported as unmeasured
- ✅ **A CHECK THAT CANNOT RUN IS NOT A PASS** - a pipeline that is configured but
  never ran, or a monitor that cannot execute, is a finding
- ✅ **NO RECOMMENDED PRODUCT CHOOSES** - assess what exists; selecting a tool is
  `/implement`'s job

---

## Finding Standards

- **Score each finding under exactly one dimension** — double-scoring is the
  defect this dimension exists to remove
- **Credit only checks that were actually run and are reachable** — "we have a
  dashboard" is not evidence that it works
- **Require the record for a rollback**, or report it as unverified
- **Keep vendor selection with Technical** — this dimension covers reproducibility
  and state safety
- **Read team health as Management's finding**, not Operations'
- **Score secret handling in deployed environments here** — source-code controls are
  Technical's
- **Score an absent category as 0 with a stated reason** — never as 100

---

## Pre-Assessment Validation

Run these checks BEFORE attempting assessment:

```text
Is there anything deployed at all?
  → yes: Continue
  → no (library, local tool, docs): Stop. Report "not an operations assessment"
    and route to /assess:technical instead. Do not score 0.

Can you read the delivery configuration?
  → readable (CI config, IaC, deploy scripts): Continue
  → denied/absent: Report the permission issue, and score only the categories
    that have evidence, marking the rest UNASSESSABLE (not 0)

Is there a running system whose telemetry you can read?
  → yes: Steps 4-5 use real measurements
  → no: State it. Rate 4-5 from configuration only, labelled MEDIUM confidence
    at best, and say which number is unverified

Is the constitution's active_dimensions list 'operations'?
  → yes: Run this file
  → no: Report the gap. A dimension not active in the constitution cannot be
    reported as a passing audit even if the score is 80+
```

---

## Assessment Completion Checklist

Assessment is complete when:

- ✅ Step 1 inventory produced: delivery artefacts, environments, runtime, dependencies
- ✅ Deployment and rollback reviewed, with rollback evidence cited or marked unverified
- ✅ Release management reviewed (versioning, cadence, approvals, changelog)
- ✅ Reliability reviewed: SLI/SLO presence, error budget, capacity, dependency failure
- ✅ Observability reviewed across logs, metrics, traces, dashboards, alerts
- ✅ IaC reviewed: reproducibility, secrets, state safety, pinning, cost visibility
- ✅ Operational readiness reviewed: runbooks, incident response, on-call, DR
- ✅ Score calculated with the formula above, summing to 100 across six categories
- ✅ Every finding routed to exactly one dimension per the boundary table
- ✅ Uncertainty surfaced: confidence level, assumptions, missing information, and
  confirmed vs. inferred for each finding
- ✅ Top 5 findings listed with evidence and a next command
- ✅ Console output delivered to the user

## Persona Protocol

Loaded by every `each family:operations` persona prompt for this dimension — `/assess`,
`/implement`, `/fix`, `/optimize`, `/loop` and `/improve` all point their
`Loads:` line here. The rules below are dimension-neutral and are written here
once rather than restated in 54 persona files, where they had already drifted
apart from the parents they point at.

- **Follow the parent protocol.** The persona fixes the *dimension*; the family
  prompt fixes the *verb*. Follow the family prompt's full protocol. Where the
  parent and this file disagree, the parent governs the behaviour.
- **This file is the *assessment* criteria for `operations`** — workflow, weights,
  grade bands, mandatory audit. Under `/assess` that is what you are scored
  against. Under `/implement`, `/fix`, `/optimize`, `/loop` and `/improve`
  it is the only per-dimension reference that exists: the parent supplies the
  verb, this file supplies the subject matter. If you need criteria the parent
  does not define, say so in your report rather than borrowing the assessment
  workflow — an `/optimize` run is not an `/assess` run with a different verb.
- **`$ARGUMENTS` is scoping detail, never a selector.** It narrows work *within*
  the dimension. It cannot select a different dimension, and it cannot be used to
  broaden scope to another dimension — that requires the user to invoke the
  other family explicitly (`/fix:technical,design`). A literal-minded reading of
  `$ARGUMENTS` as a command name is the failure this rule exists to prevent.
- **Do not manufacture work to justify being loaded.** A persona that finds
  nothing in its dimension reports "no findings in this dimension". Loading every
  persona for every run is what produces filler findings, not smaller honest ones.
- **Report the dimension you actually assessed.** Say which dimension ran. A
  finding filed under a neighbouring dimension's name is routed by that
  dimension's rules and will be closed by the wrong reviewer.
