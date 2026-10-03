---
description: "Dimension-to-command mapping for intelligent workflow routing"
applyTo: "**/*"
---

# Dimension-Command Mapping Reference

> **Purpose**: Route findings from any dimension to the command built to handle
> them. This is the **canonical** routing table **and** the canonical command/phase
> inventory — `/assess`, `/implement`, `/optimize`, `/fix`, `/improve`, and
> `universal-orchestrator` all point here instead of restating it. No other file
> in the framework may restate the per-dimension mapping.

What gets measured/built/fixed per dimension lives in that dimension's own file —
`assess-{dimension}.instructions.md` for assessment criteria, `/optimize`'s
Optimization Targets table for tuning targets. This file only maps dimension →
command; it does not restate their content.

---

## Routing Table

| Dimension | Scope | Assess | Implement | Optimize (top priority) | Error Recovery (critical path) | Continuous (`/loop`) |
|---|---|---|---|---|---|---|
| **Technical** | Code quality, architecture, tests, security, performance, dependencies, technology choices | `/assess:technical` | `/implement:technical` | `/optimize:technical --focus performance` (DB/caching first) | `/fix:technical --focus security` (immediate) | `/loop:technical` (the default scope when the constitution declares only `software`) |
| **Design** | UI/UX, WCAG accessibility, design systems, visual consistency | `/assess:design` | `/implement:design` (component/token implementation); `/implement:technical --focus wcag` (code-level accessibility); `/assess:design` Mode 1 (spec + tokens, no implementation yet) | `/optimize:design --focus design-system` | `/fix:design` (WCAG violation — immediate) | `/loop:design` |
| **Business** | Market fit, business model viability, revenue, ROI | `/assess:business` | `/implement:business` | `/optimize:business` (unit-economics) | `/assess:business` (Mode 1: re-validate assumption) | `/loop:business` (continuous) |
| **Product** | Feature set, roadmap, competitive advantage | `/assess:product` | `/implement:product` | `/optimize:product` (roadmap prioritization) | `/fix:product` | `/loop:product` |
| **Marketing** | Positioning, messaging, go-to-market | `/assess:marketing` | `/implement:marketing` | `/optimize:marketing` (claim sourcing) | `/assess:marketing` (Mode 1: reposition) | `/loop:marketing` |
| **Legal** | Regulatory compliance, data privacy, legal risk | `/assess:legal` | `/fix:legal --focus compliance` (drafting is compliance-first); `/implement:legal` (clause-set authoring) | `/optimize:legal` (clause traceability + zero unassigned compliance owners) | `/fix:legal` (CRITICAL — immediate) | `/loop:legal` |
| **Financial** | Cost structure, pricing, profitability, funding | `/assess:financial` | `/implement:financial` | `/optimize:financial --focus cost` or `--focus pricing` | `/assess:financial` (Mode 1: revalidate model) | `/loop:financial` |
| **Management** | Team capacity, process efficiency, org health | `/assess:management` | `/implement:management` | `/optimize:management` (org sustainability — span of control, single points of failure) | `/fix:management` | `/loop:management --focus process` |
| **Operations** | Deployment, release, observability, runbooks, IaC | `/assess:operations` | `/implement:operations` | `/optimize:operations` (reliability/SLO, release cadence) | `/fix:operations` | `/loop:operations` |

Every family ships a file per dimension, but the **Error-Recovery** column is not
uniformly `/fix:{dimension}`: for Business, Marketing and Financial the canonical
destination is `/assess:{dimension}` Mode 1 (re-validate the assumption rather
than patch the figure), and the corresponding `/fix` files are invokable but are
not the routing destination. The Command Coverage table below states this
explicitly. The boundary between `technical` and `operations` (which side owns
CI/CD, IaC, monitoring) is stated once, in
`.plaesy/instructions/assess-operations.md` → "Boundary" — not here.

Every scope selector above is `/{command}:{dimension}`; `--focus {sub-area}`
narrows within it — it never selects the dimension itself (see any prompt's Usage
Format for the full pattern).

**Not scored/fixed here**: framework adoption gaps, version drift, tech-stack/
design modernization, spec traceability — these are `/improve`'s lane, not a
dimension-routing decision (see `/improve`).

---

## Priority Mapping (When Multiple Findings Compete)

| Priority | Trigger | Response Time | Command |
|----------|---------|----------------|---------|
| **CRITICAL** | Security / legal / compliance breach, data loss, service down | Immediate (≤1h) | `/fix` (blocking) |
| **HIGH** | Performance below constitution target | Next phase | `/optimize` or `/fix` |
| **HIGH** | Test coverage below constitution minimum | Before merge | `/implement --focus tdd` |
| **HIGH** | Deployment/release blocked, missing observability or runbook | Before the next release | `/implement:operations` (build it) or `/optimize:operations` (tune it) |
| **MEDIUM** | Design inconsistency, unsourced business/market claim | Next sprint / research phase | `/optimize:design` or `/assess:{dimension}` (Mode 1) |
| **LOW** | Documentation gap, minor code-quality cleanup | Backlog / batch | `/doc` or `/loop` |

Priority is a property of the *severity* of a finding, not of its dimension — a
`CRITICAL` legal finding outranks a `LOW` technical one. Use this table to order
competing findings; use the Routing Table above to pick the command for the
dimension the winner belongs to.

---

## Assessment Report Section Registry (Canonical Names)

An assessment report is read by a later command, a reviewer, and a `/loop`
iteration. That only works if the section names are predictable. These four are
**canonical** — emit them with exactly these names, in this order, in every
dimension's report. A dimension may append its own sections after them, never
rename them.

| # | Canonical section | Content | Optional |
|---|-------------------|---------|----------|
| 1 | `## Overall Score` | One 0-100 number for the dimension, plus the letter grade from that dimension's Grade Mapping | no |
| 2 | `## Category Scores` | One line per scored category with its weight, so the number above is auditable | no |
| 3 | `## Findings` | Ordered by the Priority Mapping above, most severe first, each with severity, evidence (`file:line`, URL, or Figma node), and the command that resolves it | no |
| 4 | `## Next Phase` | The single recommended next command, or an explicit "no action required" | no |

Allowed appended sections, and only these:

| Section | Emitted by | When |
|---------|-----------|------|
| `## Handoff: <dimension>` | any dimension | A finding belongs to another dimension. It is recorded here and **not scored** — see the boundary rule in the Routing Table. |
| `## Uncertainty` | any dimension | The Uncertainty Surfacing protocol raised items. Never folded into the score. |
| `## Research` | any dimension | Sources cited for a Mode 1 research pass. |
| `## N/A` | any dimension | A conditional step that did not apply. Stated, not silently omitted. |

Two rules make this a registry rather than a suggestion:

1. **No synonyms.** "Top 5 Findings" is not a variant of `## Findings` — it is
   a second, competing name for the same section, and a consumer that greps for
   one will miss the other. Emit `## Findings` and put any "top 5" framing
   inside it.
2. **A missing section is a finding, not a style choice.** If a report has
   scores but no `## Findings`, the assessment did not run to completion. Say so
   rather than emitting a report that parses as complete.

Human-facing prompt files (`prompts/*.md`) may still open with `## Objective` or
`## Usage Format` — those are command-invocation scaffolding, not report
sections, and they never appear in an assessment report.

---

## Command Coverage (Every Shipped Sub-Command)

Each of the six dimension-scoped families (`assess`, `implement`, `fix`,
`optimize`, `loop`, `improve`) ships one file per dimension. Every shipped
selector appears in the Routing Table above except the deliberate exclusions
below — so a reader can tell from this one file what is routed and what is not.

| Selector | Status |
|---|---|
| `{assess,implement,fix,optimize,loop,improve}:{technical,design,business,product,marketing,legal,financial,management,operations}` | Routed — see Routing Table |
| `/fix:business`, `/fix:marketing`, `/fix:financial` | **Ship and invokable, but not the routing destination.** Their Error-Recovery cell is `/assess:{dimension}` Mode 1 — re-validate the assumption rather than patch the figure. Use `/fix` here only when a specific figure is known wrong and re-validation would not change it. |
| `/improve:adoption` | **Not a dimension.** Framework-adoption audit only (unused framework capabilities + version drift) — `/improve` Step 2, not a routing destination |
| `/improve:spec` | **Not a dimension.** Spec/plan/tasks quality + traceability only — `/improve` Step 2, not a routing destination |
| `/improve:artifact` (a file path/URL) | **Not a dimension.** Artifact Mode — no scope selector applies |
| `.plaesy/roles/nara.md` | **Not a dimension persona.** Nara is the Decision Oracle — multi-perspective deliberation for a genuinely ambiguous decision during `/fix` or `/implement`. Invoked by name when a decision is a hard fork with no safe default; never loaded as a scope. |
| `.plaesy/roles/pe.md` | **Not a dimension persona.** Prompt Engineer — strict prompt improvement and output contract. Invoked by name when the artifact under review is a prompt; never loaded as a scope. |
| `.plaesy/roles/sa.md`, `.plaesy/roles/security.md` | **Dimension-conditional.** Loaded for `technical` only when system design, integration boundaries, or threat modeling is itself a scored finding (see the Conditional Roles table in `/assess`). |
| `/doc` | **Dimension-agnostic.** Documented deliverables in any dimension; dimension-agnostic by design (produces the Handoff Spec for non-code dimensions) |
| `/create:images`, `/create:storyboard`, `/create:diagram` | **Not dimensions.** Asset-generation scopes (pixels/vectors) — `/create` router, not a routing destination |
| `/create:tasks` | **Not a dimension.** Structured-text asset generation (task files under `.plaesy/tasks/`) — `/create` router |
| `/spec:design` | **Not a dimension, despite the shape.** The selector looks like `{family}:{dimension}` and `design` is one of the nine canonical dimensions, but it names a *document type* — the design-system specification at `.plaesy/memory/design.md` — not a lens. Routing an assessment finding to `/spec:design` is a category error; a design-dimension finding goes to `/assess:design` or `/implement:design` |
| `/spec` (bare) | **No scope selector.** Specification-document router, out of phase scope — see `prompts/spec.md` |
| `/save`, `/start`, `/continue`, `/create`, `/spec`, `/loop` (bare) | **No scope selector.** Session/project orchestration commands, out of phase scope — see `.plaesy/instructions/workflow-phases.md` |
| `--focus {sub-area}` | Not a selector — narrows within an already-chosen `{command}:{dimension}` cell |

---

## See Also

- `.plaesy/instructions/plaesy.md` — Global mandatory instructions + the 9-phase model
- `/assess`, `/implement`, `/optimize`, `/fix`, `/improve`, `/loop` — the
  commands this table routes to
- `.plaesy/instructions/assess-technical.md`, `.plaesy/instructions/assess-design.md`,
  `.plaesy/instructions/assess-business.md`, `.plaesy/instructions/assess-financial.md`,
  `.plaesy/instructions/assess-marketing.md`, `.plaesy/instructions/assess-legal.md`,
  `.plaesy/instructions/assess-management.md`, `.plaesy/instructions/assess-product.md`,
  `.plaesy/instructions/assess-operations.md` — per-dimension assessment
  criteria (what gets measured, not routed)
- `.plaesy/instructions/error-recovery.md` — Error recovery protocol (both reactive heal-after + predictive prevent-before modes)

## Adversarial Verifier Scope

Every dimension's fix, not only `/implement` code diffs, gets a fresh
adversarial re-check against the original finding before it is marked
resolved. Procedure: `.plaesy/instructions/quality-gates.md` →
**Adversarial Verifier Pattern**.

## Usage Example

```text

Finding: "Pricing page claims '50% cheaper than competitors' with no source."
Dimension: Marketing (unsourced claim)
→ Lookup row: Marketing | Optimize (top priority) = claim sourcing
→ Route to: /optimize:marketing --focus claim-sourcing
```
