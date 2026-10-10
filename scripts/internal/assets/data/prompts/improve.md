---
description: "Continuous improvement gate: bring a project OR a single artifact to current best practice, routing defects/known gaps to /assess, /optimize, /fix"
subagent: true
---

# `/improve` command instructions

⚡ **Run with**: any parallel-capable mode — a host that provides
subagents, or a mode that does. Not a requirement: a single-agent host
runs this identically, just serially. `/loop` documents when fan-out is
worth its cost.

🔄 **Invocation mode**: dispatched as a background task per
`.plaesy/instructions/plaesy.md` → Long-Running Commands. Control returns to
the user immediately; completion is reported, not polled. Run synchronously
only if the user explicitly asks to wait or this conversation's next step
needs the result right away.

## Usage Format

```bash

# Project Mode (constitution + dimensions exist)
/improve                          # All dimensions active in the constitution
/improve:technical                # Tech stack, architecture pattern, tooling currency
/improve:design                   # Design system, component, accessibility-pattern currency
/improve:business                 # Business-model, pricing-model pattern currency
/improve:marketing                # Positioning/channel/GTM practice currency
/improve:legal                    # Regulatory/compliance-practice currency
/improve:financial                # Cost/FinOps practice currency
/improve:product                  # Roadmap/prioritization practice currency
/improve:management               # Process/org-practice currency
/improve:operations               # Deployment, release, observability, IaC practice currency
/improve:adoption                 # Only unused Plaesy Spec-Kit capabilities + version drift
/improve:spec                     # Only spec/plan/tasks quality + traceability

# Artifact Mode (one deliverable, no project scaffolding required)
/improve ./MOM-2026-09-16.docx           # A single document
/improve https://figma.com/file/...      # A Figma file (via Figma MCP)
/improve ./pitch-deck.pptx               # A presentation
/improve "improve my Q3 budget sheet"     # Any file/link named in the request
```

Project Mode takes the same 9-dimension scope set as `/assess:{scope}` — `technical`,
`design`, `business`, `marketing`, `legal`, `financial`, `product`, `management`,
and `operations`. `/improve:adoption` and `/improve:spec` are **not** dimensions —
they are the two dimension-agnostic checks (framework adoption, spec
traceability) that Step 2 runs on every project regardless of scope. Artifact Mode
takes no dimension scope — the artifact's own type determines what applies.

## Mode Detection (Run This First)

`/improve` needs neither a spec/tasks system nor a constitution — a user is as
likely to hand it one document as a whole project.

1. **Artifact Mode** — request names a specific file/path/link, OR no
  `.plaesy/memory/constitution.md` exists → improve that deliverable directly.
  Skip to the Artifact Mode protocol; no baseline `/assess` required (no
  project-level score exists for a single document).
2. **Project Mode** — constitution exists and no single artifact is named →
  run the dimension-scoped protocol.
3. **Both at once** is valid — e.g. "improve the pitch deck in this project"
  runs Artifact Mode on that file while still checking Framework Adoption for
  the project it lives in.

---

## Objective

Bring a project (all active dimensions) or a single artifact to current best
practice by applying cited modernization findings directly, routing defects and
known gaps to `/assess`, `/optimize`, or `/fix`.

---

## Artifact Mode Protocol

For one deliverable — a document, spreadsheet, presentation, design file, or
anything else a user might ask to "make better":

### Step A1: Identify the Artifact Type

| Artifact | Format instructions | Evidence source |
|---|---|---|
| Word document (MOM, brief, contract, report, proposal) | `word.instructions.md` | `WebSearch`/`WebFetch` — current convention for *that document type* (e.g. what a current MOM template includes: attendees, decisions log, action items with owner + due date) |
| Presentation (pitch deck, status deck, all-hands) | `powerpoint.instructions.md` | `WebSearch`/`WebFetch` — current deck-design convention (one idea/slide, data-viz practice, accessible contrast) |
| Spreadsheet (budget, tracker, model) | `excel.instructions.md` | `WebSearch`/`WebFetch` — current spreadsheet-structure convention (structured tables, named ranges, data validation vs brittle formulas) |
| Figma file / online design | Figma MCP (`get_design_context`, `get_screenshot`, `get_metadata`) + `.plaesy/roles/designer.md` + `.plaesy/roles/accessibility.md` | Figma MCP read (ground truth) + `WebSearch` for current design-system convention |
| Static website (plain HTML/CSS/JS or a generator — Astro, Eleventy, Hugo, Jekyll) | `static-site.instructions.md` | Live page read (view source / dev server) + Lighthouse-style checks (performance, SEO, accessibility) + `WebSearch` for current static-hosting/SEO convention |
| Anything matching a `templates/*.template.md` shape (PRD, runbook, RACI, threat model, SLA, ADR, etc.) | that template file | The template itself — deliverable's structure compared against it |
| Codebase / repo | *(this is Project Mode territory — not Artifact Mode)* | route to Project Mode instead |

If the artifact's type matches none of these (e.g. a format the framework has no
instructions file for yet): still evaluate it against general current best
practice for that document type via `WebSearch`, AND separately note — as a
**framework gap, not a project finding** — that Plaesy Spec-Kit has no dedicated
instructions/template for this artifact type. Report that note to the user; it's
feedback for the framework itself, not something to silently work around.

### Step A2: Check Against Current Best Practice (Cited)

Same discipline as `/assess` Mode 1: every "this is behind current practice"
claim cites its source inline (`[claim] — Source: [name/URL], retrieved
{{CURRENT_DATE}}`). Read the artifact fully before judging it (open the file;
for Figma, call the MCP read tools — never assess from the filename or a guess).

### Step A3: Apply & Report

Same apply/report rule as Project Mode's Step 3 below — every cited finding
(e.g. "add a decisions-log table with owner + due-date columns, per current
MOM convention — Source: ...") is applied directly to the artifact as part of
this run, then reported. No per-item confirmation pause — see Step 3 for the
exceptions that still require it.

---

## Project Mode Protocol

`/improve` brings **every active dimension** of a project to current best
practice — not "is it broken" (`/fix`), not "is it slow/low-quality against a
known finding" (`/optimize`), not "what's the score" (`/assess`), but *"is this
still how it should be done, or has practice moved on?"* across technical,
design, business, marketing, legal, financial, management, product, and
operations — the full constitution, not just code.

It is a **gate**, not another scorer or fixer:

| Finding is a... | `/improve` action |
|---|---|
| Unknown/unmeasured quality state | Route to `/assess` (get a score first) |
| A known performance/code-quality/audit gap already identified | Route to `/optimize` |
| A bug, error, regression, or compliance breach | Route to `/fix` |
| Nothing wrong per se, just behind current best practice for its dimension | Handle directly in this command |

The first three rows are *classes* of finding, not dimensions — which command a
given dimension-scoped finding routes to is decided by
`.plaesy/instructions/dimension-mapping.md` (the canonical routing table); use
`/optimize:{dimension}` and `/fix:{dimension}` from its Optimize and
Error-Recovery columns, and do not restate that table here.

Never re-implement `/assess`'s scoring, `/optimize`'s tuning, or `/fix`'s
root-cause debugging — call those commands and stop. Only the fourth row is
`/improve`'s own job, applied across **every** dimension, not just software.

### Step 1: Triage

1. If no current quality baseline exists for the project (no recent `/assess`
  output in `.plaesy/memory/`) → run `/assess` first, then continue with its
  findings.
2. Scan the most recent findings per active dimension (from `/assess` or surfaced
  during this run):
  - HIGH/CRITICAL finding with a known fix → route to `/optimize:{dimension}`
  - Bug/error/regression/compliance breach → route to `/fix:{dimension}`
3. Findings that are **not** defects and **not** already-scored gaps — "this
  passes, but isn't current best practice anymore" — go to Step 2, per dimension.

### Step 2: Direct Handling — Best-Practice Currency, Per Dimension

For each active dimension (from `.plaesy/memory/constitution.md`), check current
practice against what's actually current — evidence-backed the same way `/assess`
Mode 1 requires (Context7 for technical, `WebSearch`/`WebFetch` for everything
else, cited inline):

| Dimension | "Behind best practice" looks like | Evidence source |
|---|---|---|
| **Technical** | Library/framework version behind current stable; architecture pattern superseded by a better-supported one; tooling (CI, linting, build) missing a now-standard step | Context7 (`resolve-library-id` → `get-library-docs`) |
| **Design** | Design tokens/components not matching current design-system conventions; accessibility pattern superseded (e.g. newer WCAG technique) | `WebSearch`/`WebFetch` (design-system docs), `assess-design.instructions.md` |
| **Business** | Business/pricing model pattern outdated for the current market (e.g. one-time fee where subscription now dominates the segment) | `WebSearch`/`WebFetch` (market/competitor sources) |
| **Legal** | Compliance approach meets the letter of a regulation but not current best practice (e.g. consent flow technically compliant but below current UX/legal norms) | `WebSearch`/`WebFetch` (regulatory guidance, legal commentary) |
| **Marketing** | Positioning/channel mix not reflecting current GTM practice for the segment | `WebSearch`/`WebFetch` (market reports) |
| **Financial** | Cost/FinOps practice missing a now-standard optimization (e.g. no reserved-capacity/committed-use review) | `WebSearch`/`WebFetch` (FinOps sources) |
| **Management** | Process/org practice superseded (e.g. release process missing a now-standard safeguard) | `WebSearch`/`WebFetch` (process/org sources) |
| **Product** | Roadmap prioritization framework outdated, or missing a now-standard discovery practice | `WebSearch`/`WebFetch` (product-practice sources) |

Every finding here cites its source inline (`[claim] — Source: [name/URL],
retrieved {{CURRENT_DATE}}`), same rule as `/assess` Mode 1 — an uncited
"modernize this" recommendation is not usable.

**Plus two dimension-agnostic checks, every run**:

1. **Framework Adoption Audit** — compare what this project actually uses
  (`.plaesy/memory/`, `.plaesy/roles/`) against the full Plaesy catalog
  (`instructions/*.instructions.md`, `agents/*.agents.md`,
  `templates/*.template.md`). Flag high-value capabilities never adopted (judged
  against the detected stack/active dimensions — never recommend an irrelevant
  one). Flag version drift: `.plaesy/memory/*` files older than the source
  repo's current versions.
2. **Spec Quality & Traceability Review** — **only if a spec/tasks system is
  actually in use** (`.plaesy/specs/*` or `.plaesy/tasks/*` has real content,
  not just the empty scaffold) — review `spec.md`/`plan.md`/`tasks.md` for:
  testability of acceptance criteria, every requirement↔task traceable in
  both directions (no orphans), unresolved ambiguity (reuse `/assess` Mode
  1's ambiguity resolution, ≤5 items), and staleness (spec says X, deliverable
  does Y). No spec/tasks system in use → report **N/A**, not a finding — some
  deliverables (a single document, a small campaign) never need one. When the
  artifact under review is itself a prompt or AI-facing instruction — this
  repo's own corpus, an agent definition, a system prompt — delegate that
  portion to `.plaesy/roles/pe.md`, which owns reasoning order, specificity,
  few-shot examples, and output contracts. A prompt reviewed only as a text
  file will pass every readability check and still fail as a prompt.

### Step 3: Prioritize & Apply

Rank all findings (routed + directly-handled) by impact vs effort (same matrix
as `/optimize`). Apply every directly-handled finding immediately, in that
order — adopt the instruction file, edit the outdated pattern, revise the doc
section — then report what changed. No per-item confirmation pause.

Two exceptions still require explicit user confirmation before writing, per
the general risk-of-action rule (destructive, hard-to-reverse, or externally
visible):

- A change that deletes/overwrites content the user might still want (e.g.
  replacing a section wholesale rather than extending it) — prefer additive edits;
  ask only when an overwrite is unavoidable.
- A change with real-world cost or external effect (e.g. upgrading a dependency
  across a major version, revising a pricing model, publishing/pushing anything).

Everything else — adopting an instruction file, adding a CI lint step, fixing a
parity bug, adding `set -euo pipefail`, extending a document with a missing
section — is applied directly, then listed in the report so the user can review
the diff.

Routed findings (`/assess`, `/optimize`, `/fix`) are NOT applied here — call
those commands and let their own protocols decide whether to apply or confirm.

---

## Routing and Evidence Standards

- **Run Artifact Mode against the artifact alone** — it deliberately does not
  require a constitution, spec or tasks system
- **Route quality scoring to `/assess`** (Project Mode) rather than scoring inline
- **Route performance, code and audit fixes to `/optimize`**
- **Route bug and compliance-breach debugging to `/fix`**
- **Cite evidence that a modernization is current practice** before recommending it
- **Recommend only what fits the detected stack, the active dimensions, and the
  artifact's actual type** — an irrelevant adoption is churn, not improvement
- **Confirm the specific change with the user first** for anything destructive or
  externally visible: an overwrite, a major-version upgrade, a pricing-model
  change, a publish or a push
- **Apply a cited, non-destructive, in-scope finding, then report it** — Step 3's
  default is apply-then-report, not report-then-wait
- **Check whether version or practice drift was intentional** before flagging it —
  a deliberate pin and deliberate simplicity are decisions, not drift
- **Treat a missing spec/tasks system as N/A** — there is no defect to report

## Output Format

```text

✅ Improvement Gate Complete — [Artifact Mode: <file/link> | Project Mode]

# Artifact Mode
├─ Artifact Type: [detected type + matched instructions/template]
├─ Applied (cited): [gap vs current convention + source + what was changed, per item]
├─ Pending Confirmation: [destructive/costly change description, if any — awaiting user go-ahead]
└─ Framework Gap (if any): [artifact type with no matching instructions/template]

# Project Mode
├─ Routed to /assess: [reason, if baseline was missing]
├─ Routed to /optimize: [findings + dimension]
├─ Routed to /fix: [findings + dimension]
├─ Applied Directly (per dimension, cited)
│  ├─ Technical: [outdated dependency/pattern + current alternative + source + file(s) changed]
│  ├─ Design: [pattern behind current practice + source + what changed]
│  ├─ Business/Legal/Marketing/Financial/Management/Product: [same, per active dimension]
│  ├─ Framework Adoption: [unused capability adopted + why relevant] / [version drift fixed]
│  └─ Spec Quality: [orphaned requirements/tasks resolved] / [ambiguity resolved, ≤5 items] / [N/A]
└─ Pending Confirmation: [destructive/costly change description, per item — awaiting user go-ahead]

Summary (ranked by impact/effort, applied vs pending):
1. [highest impact, lowest effort item — routed, applied, or pending]
...
```

## Autonomous Routing

- Routed finding (assess/optimize/fix) → hand off to that command; `/improve`
  does not track its outcome, the called command's own routing does
- Modernization/adoption finding, non-destructive and no external effect →
  applied directly by `/improve` (instruction file copied, CI step added,
  pattern updated), then reported
- Destructive, major-version, pricing, or publish/push-type change → held for
  explicit user confirmation before writing (see Step 3 exceptions)
- Artifact finding, non-destructive → applied directly to the file/Figma design,
  then reported; a wholesale-overwrite edit is held for confirmation
- Spec ambiguity resolved → written back into the spec directly (same rule as
  `/assess` Mode 1)
- Orphaned task/requirement found → route to `/doc` or manual triage, never
  auto-deleted

`/improve` is an **out-of-phase gate** — it is not one of the 9 workflow phases
and may be invoked at any point, in any dimension, or in Artifact Mode. Its
dimension-scoped routes are the canonical ones in
`.plaesy/instructions/dimension-mapping.md`; it defines no routing of its own.

---

## Worked Example — Modernize Go Dependencies (Project Mode)

**Input**: Project with constitution (technical active). Last `/assess` was 3 weeks ago.
User runs `/improve:technical`.

```text
✅ Improvement Gate Complete — Project Mode

# Project Mode
├─ Routed to /assess: [baseline missing — ran assess:technical first, score 82/100]
├─ Applied Directly (per dimension, cited)
│  ├─ Technical: github.com/go-chi/chi v4 → v5 (Context7: resolved, current stable,
│  │  migration guide followed — internal/handler/*.go updated) + added golangci-lint
│  │  v1.61 (was v1.59) with new govet checks
│  ├─ Framework Adoption: adopted `.plaesy/instructions/tdd-enforcement.instructions.md`
│  │  (was unused) + fixed version drift in `.plaesy/memory/constitution.md`
│  └─ Spec Quality: resolved 2 ambiguities in spec.md (ASSUMED defaults labeled)
└─ Pending Confirmation: major-version upgrade chi/v4→v5 — awaiting user go-ahead

Summary (ranked by impact/effort):
1. chi v5 upgrade + migration — high impact, medium effort (pending confirmation)
2. golangci-lint v1.61 — medium impact, low effort (applied)
3. TDD enforcement adoption — high impact, low effort (applied)
4. Spec ambiguity resolution — medium impact, low effort (applied)
```

**Artifact Mode not run** (no single artifact named).

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
