---
description: "Canonical 9-phase workflow ladder: numbering, per-phase purpose/output/prerequisites, and execution patterns"
applyTo: "**/*"
---

# Workflow Phases

Executes work through **9 sequential phases** across **ALL scopes** (technical,
design, business, product, marketing, operations, legal, financial, management).
`/start` orchestrates all of them; individual commands stay available for
resume/override.

**This list is the canonical phase model.** No other file may invent its own
phase numbering — prompts reference phases by the numbers below, and any file that
prints a phase total must print **9**.

| # | Phase | Command | Prerequisites | Notes |
|---|-------|---------|----------------|-------|
| 1 | Universal Research & Validation | `/assess` (Mode 1) | None (first phase) | Sub-step 1.0 = Constitution, Sub-step 1.1 = Research + Ambiguity Resolution |
| 2 | Implementation | `/implement` | Tech stack clear, specs clear, TDD environment ready | Universal orchestrator; TDD for `software`, Draft-Review-Revise otherwise |
| 3 | Quality Assurance | *(none — automatic gate)* | Implementation complete | **Not a command.** Runs automatically at the end of Phase 2; blocking on failure — see `.plaesy/instructions/quality-gates.md` |
| 4 | Comprehensive Assessment | `/assess` (Mode 2) | Phase 2 complete | MANDATORY; blocks Phase 5 |
| 5 | Comprehensive Optimization | `/optimize` | Baseline metrics from `/assess` | Must be followed by Phase 6 |
| 6 | Verification Assessment | `/assess` (Mode 3) | Phase 5 complete | MANDATORY; blocks Phase 7/8 |
| 7 | Error Recovery | `/fix` | Issues found in assessment | Also runs on demand, outside the ladder |
| 8 | Documentation | `/doc` | Code complete + tests passing | OPTIONAL; skip for backend-only projects with minimal public API |
| 9 | Session Management | `/save` | None (can run anytime) | End of session or checkpoint |

## Sub-steps 1.0 and 1.1 Are Part of Phase 1

`/start` runs two sub-steps before handing off: **Sub-step 1.0 (Constitution)**
and **Sub-step 1.1 (Ambiguity Resolution)**. **Both are sub-steps of Phase 1** and
are never counted as phases of their own — the ladder is 1-9 and every printed
total is `/9`:

- **Sub-step 1.0 — Constitution** (`/start` only, one-time): detect the primary
  deliverable, generate `.plaesy/memory/constitution.md`. Everything downstream
  reads this file first.
- **Sub-step 1.1 — Research & Ambiguity Resolution** (`/assess` Mode 1): external
  research plus internal spec-ambiguity resolution, in one pass (bounded to ≤5
  questions — see `/assess` Mode 1).

## Commands That Are Not Phases

These are always available and are **out of phase scope** — a reader never counts
them when asking "which phase am I on?":

| Command | Status | Why it has no phase |
|---------|--------|--------------------|
| `/continue` | Out of phase | Driver: detects state, then executes the *next real phase*; sets no phase of its own |
| `/loop` | Out of phase | Driver: runs `assess → implement → verify` repeatedly until a stop condition — it crosses phases, it is not one |
| `/create` | Out of phase | Scaffolding a single artifact (`spec`, `tasks`, `plan`, …) for a phase, not a phase itself |
| `/improve` | Out of phase | Improvement gate invocable at *any* point, in any dimension or in Artifact Mode |
| `/fix` (on demand) | Phase 7 when the ladder reaches it; otherwise on demand | A defect is fixed when found, not only at Phase 7 |

## Per-Phase Detail

Only the phases with content beyond the table above are elaborated here.

### Phase 4: Comprehensive Assessment — dimension checklist

Can assess any combination of: Technical (code quality, tests, security,
performance, architecture, docs), Design (UI/UX, WCAG AA, design system),
Business (market fit, viability, positioning, revenue), Product (feature
completeness, competitive advantage), Marketing (positioning, GTM readiness),
Legal/Compliance (regulatory, privacy, risk), Financial (cost, pricing, unit
economics), Operations (deployment, release, observability, runbooks, IaC),
Management (team capacity, process, org health).

Output: multi-dimensional quality scores (0-100 each) + findings +
research-validated recommendations. Findings feed the loop: Assessment gaps →
Task backlog → `/assess` (research mode) validation → `/implement` fixes →
assess again.

### Phase 5: Comprehensive Optimization — scope

Optimizes performance (response time, throughput, memory, caching,
algorithms), design quality (component reuse, tokens, dark mode,
accessibility, animations — if UI), and code quality (refactoring,
duplication, type safety, error handling, testing, docs).

### Phase 6: Verification Assessment — must verify

Performance improvements (% gains measured), design improvements applied (if
UI), code quality improvements complete, no regressions (tests still
passing), all quality dimensions still within acceptable range. Output:
verification report + confidence score (0-100).

## Execution Patterns

- **New project** (`/start`): execute all phases sequentially (1 → 9).
- **Resume session** (`/continue`): auto-detect current phase → execute remaining (e.g., 4 → 9).
- **Assessment loop** (recommended): `/continue` runs implementation phases →
  quality-gates → `/assess` (assessment mode). On findings: design issues →
  `/implement` (refactor) → `/assess` (verify); performance issues →
  `/optimize` → `/assess` (verify); bugs → `/fix` → `/assess` (verify).
- **Agile continuous improvement**: `/continue` → `/assess` results seed new
  backlog tasks → `/continue` resumes with them → `/assess` (research mode)
  validates → `/implement` → `/assess` (assessment mode) → optimize/fix →
  `/assess` loop.
- **Autonomous** (`/loop`, optionally scoped e.g. `/loop:technical`): iterates
  assess → fix → verify on its own axis until you end the session, or until
  `consecutive_failures` reaches its limit. It does not run `/continue` for
  you, and does not stop when the task list empties — an empty checklist is a
  reason to widen the axis.
- **Selective phases**: `/assess` alone runs in research mode (validate tech
  choices upfront) or assessment mode (measure quality after implementation);
  `/optimize` runs performance/design/code optimization on demand.
