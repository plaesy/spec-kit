---
description: "Canonical 9-phase workflow ladder + Spec-Kit 15-stage business validation pipeline"
applyTo: "**/*"
---

# Workflow Phases

Executes work through **9 sequential phases** across **ALL scopes** (technical, design, business, product, marketing, operations, legal, financial, management). `/start` orchestrates all of them; individual commands stay available for resume/override.

**This list is the canonical phase model.** No other file may invent its own phase numbering — prompts reference phases by the numbers below, and any file that prints a phase total must print **9**.

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

`/start` runs two sub-steps before handing off: **Sub-step 1.0 (Constitution)** and **Sub-step 1.1 (Ambiguity Resolution)**. **Both are sub-steps of Phase 1** and are never counted as phases of their own — the ladder is 1-9 and every printed total is `/9`:

- **Sub-step 1.0 — Constitution** (`/start` only, one-time): detect the primary deliverable, generate `.plaesy/memory/constitution.md`. Everything downstream reads this file first.
- **Sub-step 1.1 — Research & Ambiguity Resolution** (`/assess` Mode 1): external research plus internal spec-ambiguity resolution, in one pass (bounded to ≤5 questions — see `/assess` Mode 1).

## Commands That Are Not Phases

These are always available and are **out of phase scope** — a reader never counts them when asking "which phase am I on?":

| Command | Status | Why it has no phase |
|---------|--------|---------------------|
| `/continue` | Out of phase | Driver: detects state, then executes the *next real phase*; sets no phase of its own |
| `/loop` | Out of phase | Driver: runs `assess → implement → verify` repeatedly until a stop condition — it crosses phases, it is not one |
| `/create` | Out of phase | Scaffolding a single artifact (`spec`, `tasks`, `plan`, …) for a phase, not a phase itself |
| `/improve` | Out of phase | Improvement gate invocable at *any* point, in any dimension or in Artifact Mode |
|| `/fix` (on demand) | Phase 7 when the ladder reaches it; otherwise on demand | A defect is fixed when found, not only at Phase 7 ||
|| `pipeline` | **Extended pipeline** | 15-stage business validation that may span multiple phases; not a single phase ||
|| `youtube-automation` | **Extended pipeline** | 16th stage for video content; optional, runs after marketing/launch ||

---

## Spec-Kit Business Validation Pipeline (Extended Workflow)

The **spec-kit pipeline** (`pipeline` skill) implements a **15-stage business validation workflow** that maps to the 9-phase ladder as follows:

| Pipeline Stage | Phase Mapping | Skill | Output |
|---|---|---|---|
| 1. Idea Discovery | Phase 1.1 (Research) | `ideator` | 8-12 candidates in `.plaesy/ideas/<topic>/` |
| 2. Idea Definition | Phase 1.1 | `founder: idea.md` | Core idea definition |
| 3. Board Review | Phase 1.1 | `founder: board.md` | 3-lens votes (offer, defensibility, product) |
| 4. Competitor Research | Phase 1.1 | `founder: competitors.md` | Pricing table + gap analysis |
| 5. Buyer Validation | Phase 1.1 + Phase 4 | `buyer-discovery` | **≥5 real conversations**, `buyers.md` |
| 6. Pricing | Phase 1.1 | `founder: pricing.md` | Segment-differentiated pricing |
| 7. Offer | Phase 1.1 | `founder: offer.md` | Outcome, guarantee, included/excluded |
| 8. Compliance (health) | Phase 1.1 + Legal | `compliance-health` | Risk-scored content + approved claims DB |
| 9. Economics → **VERDICT** | Phase 4 (Assessment) | `economics-engine` | **Incomplete \| Not yet \| Profitable** |
| **Pivot Decision Gate** | **Phase 7 (Error Recovery)** | `pipeline` | **If Not yet: choose pivot, document in plan.md** |
| 10. Marketing | Phase 2 (Implement) | `founder: marketing.md` | Channel, message, first test |
| 11. Brand | Phase 2 | `founder: brand.md` | Name, tone, anti-brand |
| 12. Operations | Phase 2 | `founder: ops.md` | Delivery requirements, capacity |
| 13. Launch | Phase 2 + Phase 8 | `founder: launch.md` | 30-day plan, smallest test, stop rules |
| 14. Plan Compilation | Phase 8 (Doc) | `plan-compiler` | `plan.md` with verdict + pivot docs |
| 15. Assessment (optional) | Phase 4 (Assessment) | `assess` | Grades per dimension (business, marketing, financial) |
| **16. YouTube Automation (optional)** | **Phase 2 + Phase 8** | `youtube-automation` | Video MP4 + thumbnail + metadata + findability report |

### YouTube Automation Integration
- **After Stage 10 (Marketing)**: Generate video assets for marketing collateral
- **After Stage 13 (Launch)**: Create launch videos, demo walkthroughs, tutorial content
- Output feeds into marketing collateral, launch execution, and ongoing content strategy
- Long-form (16:9) for deep-dive, Shorts (9:16) for teasers/social

1. **Stages 1-9 are Research/Validation** (Phase 1.1) — no code written, only market validation
2. **Stage 9 (Economics) is the critical gate** — produces verdict that determines path
3. **Pivot Decision Gate (between 9-10)** — if verdict = Not yet, MUST choose pivot before proceeding
4. **Stages 10-13 are Implementation** (Phase 2) — build marketing, brand, ops, launch artifacts
5. **Stage 14 is Documentation** (Phase 8) — compile final plan
6. **Stage 15 is Verification Assessment** (Phase 4/6) — audit the complete plan

### Resume Points

The pipeline supports resume from any stage:
- `--resume-from ideator` — start from idea discovery
- `--resume-from buyers` — skip to buyer validation (if idea + board + competitors done)
- `--resume-from economics` — re-run economics with new inputs
- `--resume-from marketing` — re-do marketing after pivot
- `--resume-from assess` — re-run assessments after plan changes

### Error Recovery (Per Founder Skill Rules)

When any stage fails loudly (missing input):
1. Ask ≤3 clarifying questions in one message
2. Record missing item as `UNKNOWN` in the stage file
3. Continue stages that don't need the missing input
4. Final verdict reflects the gap (e.g., `Incomplete` if S has no basis)

---

## Per-Phase Detail (Core)

### Phase 4: Comprehensive Assessment — dimension checklist

Can assess any combination of: Technical (code quality, tests, security, performance, architecture, docs), Design (UI/UX, WCAG AA, design system), Business (market fit, viability, positioning, revenue), Product (feature completeness, competitive advantage), Marketing (positioning, GTM readiness), Legal/Compliance (regulatory, privacy, risk), Financial (cost, pricing, unit economics), Operations (deployment, release, observability, runbooks, IaC), Management (team capacity, process, org health).

Output: multi-dimensional quality scores (0-100 each) + findings + research-validated recommendations. Findings feed the loop: Assessment gaps → Task backlog → `/assess` (research mode) validation → `/implement` fixes → assess again.

**Spec-Kit Extension**: `assess` runs dimension assessments on `.plaesy/ideas/<topic>/` files. Outputs `assessment-<dimension>.md` with grade (A+-D), score, component scores, top 5 findings, recommendations.

### Phase 5: Comprehensive Optimization — scope

Optimizes performance (response time, throughput, memory, caching, algorithms), design quality (component reuse, tokens, dark mode, accessibility, animations — if UI), and code quality (refactoring, duplication, type safety, error handling, testing, docs).

**Spec-Kit Extension**: `/optimize:business --focus unit-economics` uses `economics-engine` sensitivity analysis. `/optimize:marketing --focus claim-sourcing` uses live competitor audit.

### Phase 6: Verification Assessment — must verify

Performance improvements (% gains measured), design improvements applied (if UI), code quality improvements complete, no regressions (tests still passing), all quality dimensions still within acceptable range. Output: verification report + confidence score (0-100).

**Spec-Kit Extension**: After pivot implementation, re-run `economics-engine` and `assess` to verify new model achieves Profitable.

### Phase 7: Error Recovery — Pivot Gate

When `economics-engine` returns **Not yet**:
1. Run segment analysis (per-tier break-even, cross-subsidy check)
2. Run segment validation matrix (kill segments with <30% would-pay after ≥3 conversations)
3. Present 4 pivot options with projected verdicts
4. **Require user decision** — document in `plan.md` → `## Pivot Analysis`
5. Only then proceed to Stage 10+

This is the **only mandatory human decision point** in the pipeline.

---

## Execution Patterns

- **New project** (`/start`): execute all phases sequentially (1 → 9).
- **Resume session** (`/continue`): auto-detect current phase → execute remaining (e.g., 4 → 9).
- **Assessment loop** (recommended): `/continue` runs implementation phases → quality-gates → `/assess` (assessment mode). On findings: design issues → `/implement` (refactor) → `/assess` (verify); performance issues → `/optimize` → `/assess` (verify); bugs → `/fix` → `/assess` (verify).
- **Agile continuous improvement**: `/continue` → `/assess` results seed new backlog tasks → `/continue` resumes with them → `/assess` (research mode) validates → `/implement` → `/assess` (assessment mode) → optimize/fix → `/assess` loop.
- **Autonomous** (`/loop`, optionally scoped e.g. `/loop:technical`): iterates assess → fix → verify on its own axis until you end the session, or until `consecutive_failures` reaches its limit. It does not run `/continue` for you, and does not stop when the task list empties — an empty checklist is a reason to widen the axis.
- **Selective phases**: `/assess` alone runs in research mode (validate tech choices upfront) or assessment mode (measure quality after implementation); `/optimize` runs performance/design/code optimization on demand.
- **Pipeline** (`/pipeline`): Runs 15-stage business validation. Can start from niche (`--niche health --idea "..."`) or resume existing topic (`--topic ai-health-influencer --resume-from economics`). Produces complete `.plaesy/ideas/<topic>/` with verdict.

---

## Pipeline Execution Patterns

| Pattern | Command | Use Case |
|---|---|---|
| **Full new idea** | `/pipeline --niche health --idea "AI health content"` | Start from scratch, run all 15 stages |
| **Validate existing** | `/pipeline --topic my-idea --resume-from buyers` | Have idea + board + competitors, need buyer validation |
| **Re-run economics** | `/pipeline --topic my-idea --resume-from economics` | Updated pricing/F/S inputs, need new verdict |
| **Pivot & rebuild** | `/pipeline --topic my-idea --resume-from marketing --pivot hybrid` | After Not yet verdict, chose pivot, rebuild marketing+ |
| **Assess only** | `/assess --dimension business --topic my-idea` | Audit existing plan without full pipeline |
| **Buyer sprint** | `/buyer-discovery --topic my-idea --target 5` | Focused buyer validation sprint |

---

## See Also

- `skills/using-plaesy/SKILL.md` — Bootstrap with spec-kit skill map
- `skills/pipeline/SKILL.md` — Master orchestrator with 15 stages
- `references/routing.md` — Dimension-to-command mapping with spec-kit extensions
- `.plaesy/instructions/plaesy.md` — Global mandatory instructions + 9-phase model