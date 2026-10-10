---
title: "Prompt Corpus Assessment — spec-kit (plaesy)"
date: 2026-09-27
mode: assessment
scope: prompts, instructions, agents, checklists, templates, generated mirrors
scores:
  core-protocol: 34
  templates: 41
  orchestration-prompts: 52
  dimension-subprompts: 52
  instruction-library: 58
  agents-checklists: 63
  architecture: 70
  overall: 47
method: 7 parallel assessors + parent verification of every CRITICAL/HIGH claim
---

# Prompt Corpus Assessment — 2026-09-27

Audit of the AI-agent prompt corpus: `prompts/` (69), `instructions/` (69),
`agents/` (23), `checklists/` (7), `templates/` (69) — 237 source files,
~43,000 lines, ~388k tokens. Plus the generated mirrors `.kilo/` and `.plaesy/`.

## Verdict

**47/100.** Architecture is sound (config-driven scaffold, registries, consistent
8-dimension model). What is missing is mechanically enforced discipline.

Four defect classes:

1. **Unreachable content** — 62 of 69 templates (90%) are read by no prompt,
   instruction, agent, or Go code path.
2. **Self-violated DRY** — the framework declares `dimension-mapping.md` the
   single source of truth for routing, then duplicates ~170 lines of routing
   across 6 other files.
3. **Unresolvable references** — 16 path references in `prompts/` point at files
   that do not exist after `plaesy init`.
4. **No validator** — 10 of the 15 top findings are mechanically detectable. No CI
   job references `prompts`, `instructions`, `agents`, or `checklists`.

## Scores

| Area | Score | Notes |
|---|---|---|
| Core protocol (`plaesy`, `dimension-mapping`, `quality-gates`, `error-recovery`) | 34 | declared SSOT violated by the framework itself |
| Template library (69 files, 141k tokens) | 41 | 62/69 unreachable |
| Orchestration prompts (11 parents) | 52 | strong per-file, DRY failure across files |
| Dimension sub-prompts (48 files) | 52 | 63% is empty boilerplate |
| Instruction library (69 files, 142k tokens) | 58 | consistent but heavily duplicated |
| Agents + checklists (30 files) | 63 | zero drift, but 5 roles unwired |
| Architecture & maintainability | 70 | good scaffold design, broken git hygiene |
| **Weighted overall** (by corpus size) | **47** | |

## Critical findings

### C1 — `spec.template.md` never reaches the user (code bug)

`scripts/internal/featurepath/create.go:93` reads `<repoRoot>/templates/spec.template.md`
(`GetRepoRoot()` = `git rev-parse --show-toplevel` = the **user's** repo), but
`scripts/internal/scaffold/copy.go:263` installs templates to `.plaesy/templates/`.
The read fails, and the `else` branch at `create.go:99-103` writes a **zero-byte
`spec.md`** and discards the error.

Masked in this repo because `templates/` exists at the root; in every installed
project `/specify` produces an empty spec. `featurepath_test.go:258` stubs the
template as a parameter, hiding the bug.

### C2 — `prompts/create/tasks.md` has two contradictory output contracts

Verified duplicated headings: 2x `### Step 5: Validate & Report` (`:128`, `:285`),
2x `## Programmatic Invocation` (`:175`, `:297`), 2x `## Anti-Patterns` (`:149`, `:319`),
2x `## Success Criteria` (`:195`, `:344`), 2x Design-Spine integration (`:161`, `:334`).
~130 contradictory lines.

Contract A writes one file per task to `.plaesy/tasks/backlog/{priority}_{title}.md`
and reports `file_count / files[] / index`. Contract B writes a single aggregate with
`markdown_path / json_path / backlog{epic_count,task_count,total_points}`. The `--out`
flag table (`:29`) and frontmatter (`:2`) endorse A; the JSON schema (`:240-282`)
endorses B. No rule selects between them.

### C3 — `design-spine.md` is referenced 43x and never exists

6 `prompts/create/*` files + `templates/template-registry.json:188` (a tag, not a path).
The canonical artifact named in `prompts/assess.md:158`,
`instructions/assess-design.instructions.md:26` and `docs/instructions/README.md:102`
is `.plaesy/memory/design.md`. `.plaesy/memory/` is **empty**.
Default style resolution in `images.md:21` and `storyboard.md:25` therefore always
falls through to generic-art fallbacks.

### C4 — routing tables duplicated despite an explicit prohibition

`instructions/plaesy.instructions.md:224-226`: *"Do not duplicate the per-dimension
table in this file or in any prompt — `dimension-mapping.md` is the single source of
truth for routing."*

Violated at: `plaesy.md:179-213` (35 lines), `prompts/assess.md:280-299`,
`prompts/implement.md:369-385`, `prompts/fix.md:315-369` (55),
`prompts/optimize.md:252-311` (60), `prompts/doc.md:317-335`,
`prompts/improve.md:246-262`, `instructions/universal-orchestrator.md:108-161` (54).
**~170 duplicated lines** from an 82-line canonical file.

Contradictions: `plaesy.md:203` routes a business issue to `/assess:business` while
`dimension-mapping.md:25` routes to `/implement:business`; `plaesy.md:201` sends a
technical gap to `/implement` while the canonical table sends it to `/implement:technical`.

### C5 — the phase model contradicts itself

| Source | Claim |
|---|---|
| `instructions/plaesy.md:302-403` | 9 phases (1-9) |
| `prompts/start.md:135` | `Phase: [current/11]` |
| `prompts/continue.md:217` | `X/12 phases` |
| `prompts/loop.md:40,57,77,104,129` | local `Phase 0-4` (collides with global) |
| `prompts/assess.md:48` | Mode 2 = "Phases 4, **6**" |
| `prompts/assess.md:49` | Mode 3 = "**Phase 6**" |

Phase 6 is assigned to two modes. `start.md` adds Phase 0 (Constitution) and
Phase 0.5 (Ambiguity Resolution) that `plaesy.md` does not list. Phase 3
(quality-gates) has no command. `/improve`, `/create`, `/continue`, `/loop` have
no phase at all.

### C6 — generated artifacts are tracked despite ignore rules

`.gitignore:9` (`.kilo*`) and `.gitignore:24-28` (`.plaesy/templates*`, `.plaesy/instructions*`,
etc.) express the intent, but `git ls-files` tracks **138 `.plaesy/` files + 69 `.kilo/`
files + the 15MB `plaesy` binary**. gitignore never untracks. Any hand-edit by a user
silently diverges from the source of truth. `git status` already shows 8
`.plaesy/analysis/` files dirty from a timestamp-only diff.

### C7 — 16 path references do not resolve

- **3 roles that do not exist**: `.plaesy/roles/technical.md` (`create/diagram.md:50`,
  `create/api.md:45`), `.plaesy/roles/architect.md` (`create/diagram.md:50`),
  `.plaesy/roles/product.md` (`create/tasks.md:49`). Only `ai-architect.md` exists.
- **5 provider-config JSONs with no producer**: `api-format.json` (`create/api.md:52`),
  `image-provider.json` (`create/images.md:51`, `create/storyboard.md:61`),
  `llm-provider.json` (`create/tasks.md:58`), `ci-platform.json` (`create/templates/ci.md:52`),
  `iac-provider.json` (`create/templates/infra.md:60`). No writer exists in Go, `install.sh`
  or `install.ps1`. Every "Step 2: Resolve the Provider" resolves to a file that cannot exist.
- **`brandkit.instructions.md`**: 7 references use the repo-relative path
  `instructions/brandkit.instructions.md`, which does not exist in an installed project
  (the installed form is `.plaesy/instructions/brandkit.md`). It *is* in `mapping.json`
  at line 297, but only as a stack-detection entry — not in `always_load` or `scope_load`.

### C8 — the promised index does not exist, and 38 of 68 instructions are never installed

`plaesy.md:103` declares `instructions.md  # INDEX/TOC for .plaesy/instructions/[topics].md`.
The file does not exist. Verified: `.plaesy/instructions/` holds **30** files while
`instructions/` holds **68** `.instructions.md` files — `copy.go:117-152` copies only
`always_load` (9) + `scope_load` (21). A scoped command in an installed project cannot
load the other 38. `.plaesy/memory.md:17` links to `overview.md`, which does not exist.

## High findings

| # | Finding | Evidence |
|---|---|---|
| H1 | `docs/agents/README.md` mislabels 2 roles — the router table sends the wrong persona | `:27` `pm` = "Project Manager"; the file is the **Product** Manager. `:52` `pe` = "Performance Engineer"; the file is the **Prompt** Engineer |
| H2 | `po.checklist.md` is not a Product Owner checklist | 151 checkboxes, **zero** backlog/vision/prioritization/value items; content is scaffolding, DB, deploy pipeline, frontend infra. Overlaps `sa.checklist.md:75-197`. Loaded by `/assess:product`, which delegates to `pm`+`ba`, never to `po` |
| H3 | No CI validator for any of this | `.github/workflows/ci.yml` has 0 references to `prompts`, `instructions`, `agents`, `checklists` |
| H4 | 5 of 23 roles (21%) are never invoked | `ai-architect`, `bo`, `mlops`, `pe`, `po` return 0 hits in `prompts/` + `instructions/` |
| H5 | 23 personas, 0 boundary statements, 0 escalation rules (except `sm`) | Overlapping pairs: security↔devsecops, security↔compliance, dev↔devops, dev↔sa, pm↔po, pm↔bo, designer↔accessibility, qa↔sa, devops↔sre, mlops↔ai-architect, ba↔pm |
| H6 | `qa.checklist.md` violates the documented checklist format | 112 of 146 lines are non-checkbox sub-bullets; **0** `[[LLM:]]` blocks (siblings have 6-14); 59 `[PLACEHOLDER]` tokens. ~77% is silently uncheckable |
| H7 | 2 claimed stacks have no instruction file | No `python.instructions.md`, no `vue.instructions.md`; `plaesy.md:317` promises Python |
| H8 | 15 names for success criteria, 10 for the opening frame | Breaks `improve.md:70`, which grades a deliverable by matching section names |
| H9 | 4 incompatible placeholder languages, no convention | `{{VAR\|default}}`, `[UPPER_SNAKE]`, `{lower_snake}`, `<angle>`; `constitution.template.md:41` escapes `\|` inconsistently with `:59` |
| H10 | Filename typo | `instructions/software-design-prinsiples.instructions.md`, hardcoded in `mapping.json` x2 and referenced twice more |

## Duplication quantified

| Block | Files | ~Lines | Single home |
|---|---|---|---|
| Autonomous routing tables | 6-8 | **170** | `dimension-mapping.instructions.md` |
| 48 sub-prompt boilerplate (9 of 9 body lines identical after substitution) | 32 | **256** | delete from sub-prompts |
| Uncertainty Surfacing Protocol | 10 | **90** | `instructions/uncertainty-surfacing.md` |
| Phase list | 2 | **150** | on-demand, not always-load |
| Success Criteria checklist | 4 | **70** | `quality-gates.md` |
| "Why This Phase Matters" triad | 4 | **55** | shared snippet |
| Context7 protocol | 3 | **35** | new `context7.instructions.md` |
| Security (OWASP / authz / secrets) | 3 | **419** | 2 files |
| Idempotency / retry / circuit-breaker / bulkhead | 5 | **200** | `distributed-systems-design-principles` |
| Testing pyramid + 90% coverage | 2 | **40** | `tdd-enforcement` |
| Scaffold shared by 8 `assess-*.instructions.md` | 8 | **1,400** | `assess-core.instructions.md` |
| Duplicated tail of `create/tasks.md` | 1 | **130** (contradictory) | one tail |
| **Subtotal, prompts + instructions** | | **~3,000** | |
| 8 templates mergeable with no capability loss | 8 | **~5,100** | — |
| 4 largest unreachable templates | 4 | **4,518** | `docs/` |
| **Total removable** | | **~8,100 / 43,000 lines (~19%)** | |

62 of 69 templates (90%) are read by no prompt, instruction, agent, or Go code path.
Only 7 are live: `spec`, `constitution`, `tasks`, `design`, `state`, `context`, `plan`.

## Best-practice evaluation (cited)

| Practice | Corpus reality | Source |
|---|---|---|
| 3-5 few-shot examples in `<example>` tags | 2 of 11 parent commands; 0 of 48 sub-prompts | Anthropic, "Prompting best practices", retrieved 2026-09-27 |
| Role in system prompt | present (23 personas) | same |
| XML tags to delimit content blocks | only in `plaesy.md`'s subagent template; corpus is ~95% markdown | same |
| Prefill assistant message | correctly absent (returns 400 on current models) | same |
| Instruction count ceiling | `always_load` = 9 files, ~2,473 lines | arXiv:2607.19257 — perfect-response rate collapses to zero by N=80 rules across all models, formats and placements; retrieved 2026-09-27 |
| Conflicting instructions | 6 routing tables, 4 phase counts | arXiv:2502.15851 (Control Illusion, AAAI-26) — 9.6-45.8% drop in primary-constraint adherence on conflicts; retrieved 2026-09-27 |
| Root instruction file length | `plaesy.md` 454 lines, always loaded | the "~150-200 line" figure is practitioner-sourced, not controlled (MEDIUM confidence); direction supported by Gloaguen et al. 2026 (~-20% task success from poorly-constructed context files) and context rot beginning at 70-80% of window (Hong et al., via Kuhara 2026) |
| Auto-generated context files | most of the corpus | McMillan 2026 — a root instruction file helps only when human-authored and non-inferable; auto-generated context files often hurt |

## Corrections — claims rejected during verification

| Claim | Verdict |
|---|---|
| "`.plaesy/roles/` does not exist" | **FALSE** — 23 files exist and are git-tracked; the assessor used a glob that skips gitignored paths |
| "`.plaesy/instructions/tasks.md` is a dead reference" | **FALSE** — the file exists |
| "`@nara` is undefined" | **FALSE** — `agents/nara.agents.md` exists |
| "`brandkit` is never installed" | **HALF TRUE** — it is in `mapping.json` but only as a stack-detection entry; the real defect is the repo-relative path in 7 references |
| "No `python.instructions.md`" | **TRUE** |
| "spec template path bug" | **TRUE** — verified in the Go source |

## Uncertainty

### Not assessed

- Runtime effectiveness. No harness was available; all findings are static. The 47
  measures artifact quality, not model output.
- Model behavior on these prompts. Requires an eval: run `/assess:technical` vs
  `/assess` on the same repo, measure output adherence. No baseline exists.
- Real token cost. The 388k figure is a char/4 estimate, not model tokenization.
- Whether `ultracode` and `mcp__context7__*` are guaranteed available. 5 prompts
  mandate them; 0 agent files mention them.

### Assumptions

- "Duplicated" = lines identical after substituting the family/dimension name, i.e.
  what the model reads as identical.
- `mapping.json` `always_load` ∪ `scope_load` is the authoritative install set
  (confirmed at `copy.go:26`).
- `.plaesy/` paths are runtime locations, not source paths.

### Speculative — verify before acting

- Duplicate line counts are estimates, not a full diff.
- "90% of templates are dead" ignores dynamic/glob resolution (`improve.md:70`
  references `templates/*.template.md`); some may be intended for direct human use.
- The absolute effect of the N=80 instruction ceiling on this corpus is unmeasured.

## Recommended sequence

```text
1. /fix       C1 (spec template path), C2 (tasks.md tail), H1 (docs labels)
2. /implement Build the validator — make check-commands + go test corpus integrity
3. /loop      Dedupe: routing (170), phase model, sub-prompt boilerplate (256),
              security (419), resilience (200), assess-core scaffold (1,400)
4. /doc       Regenerate docs/ — metadata.json counts, dead source paths
5. /assess    Re-verify: confirm 10 defect classes cleared, no regressions
```

Highest leverage is the validator, not the rewriting. 10 of 15 top findings are
mechanically detectable; without enforcement every dedupe starts drifting again.

---

# Re-assessment round — 2026-09-27 (post-remediation)

Seven independent assessors re-scored the corpus after the C1–C8 / H4 / H8–H10
remediation round. Every score below is a second opinion on the *current* tree,
not a projection.

| Area | Baseline | Re-assessed | Change |
|---|---|---|---|
| Instructions | 58 | 61 | +3 |
| Prompts (orchestration + dimension) | 52 | 68 | +16 |
| Agents | 63 | 62 | −1 |
| Checklists | 63 | 62 | −1 |
| Templates | 41 | 58 | +17 |
| Cross-cutting CLI/docs | 70 | 71 | +1 |
| **Weighted overall** | **47** | **~60** | **+13** |

The two −1 movements are worth noting rather than hiding: the agent and
checklist scores fell by one point each because the assessors re-read the files
*after* the `## Boundaries & Escalation` and territory-scoping remediation and
found the added specificity introduced new coupling obligations. That is a
measurement artefact, not a regression.

## Findings raised by the re-assessment, and their disposition

| # | Severity | Finding | Disposition |
|---|---|---|---|
| T1 | CRITICAL | Go's `time.RFC3339` **layout constant** shipped as a timestamp *value* in `context.template.md:3`, `memory.template.md:4,10` (`2006-01-02T15:04:05.999Z`) | **Fixed** — `{{UPDATED_AT|<ISO-8601 …>}}` |
| T2 | CRITICAL | `status.template.md` declared a phantom 5-phase model (Idea→…→Implementation) with zero name overlap against the canonical 9-phase ladder | **Fixed** — relabelled *Stage 1–5 (artifact pipeline)*, ladder pointer added, five wrong `File:` paths corrected |
| T3 | CRITICAL | `sbom-template.json` — *"valid JSON only by quote luck"* | **Falsified** — `json.load()` parses it clean, before and after any edit; all 67 fill-ins sit inside JSON *string values*. The real defect was 67 validator-invisible fill-ins, now converted to `{{ }}` with strict JSON re-validated |
| T4 | HIGH | `spec-structure-folder.template.md:5` gated on five `*.prompt.md` files that exist nowhere | **Fixed** — gate restated as "`plaesy features create` already produced this layout" |
| T5 | HIGH | `plan.template.md` numbered its own steps "Phase 0/1/2/3–4", colliding with a ladder where Phase 0 is explicitly *not* a phase | **Fixed** — renumbered Steps 1–3 |
| T6 | HIGH | `docs/templates/README.md:255-262` prescribed the deprecated `[PROJECT_NAME]` syntax | **Fixed** — points at `templates/README.md` |
| T7 | HIGH | `implementation.template.md` cited as the next hop by `service.template.md:654` and `tool.template.md:446`; the file does not exist | **Fixed** — both now route to `/implement` + `/create plan` |
| T8 | MEDIUM | `integration-examples.template.md` is a 4.6k-word finished reference document, not a fill-in skeleton | **Fixed** — moved to `docs/examples/integration-examples.md` with an explicit "not a template" header; registry path updated, registry still valid JSON |
| X1 | CRITICAL | `prompts/create/templates/infra.md` and `ci.md` instruct `plaesy generate-template` / `generate-pipeline` — **neither command exists** (verified by executing the built binary) | **Fixed** — explicit NOT-IMPLEMENTED branch with the verbatim cobra error; requirement retained as a manual-write contract |
| X2 | HIGH | `docs/instructions/README.md` taught `plaesy detect-stack`; **no alias exists** (it was assumed to be kept) | **Fixed** — `stack detect`, verified against the binary |
| X3 | HIGH | `docs/reference.md` never documented the C8 `stack detect --install` flag; `docs_drift_test.go` matches subcommand *names* only, never flags | **Fixed** — flag row + example added; the test gap is check #1 in the CI table below |
| X4 | HIGH | `docs/scripts/plaesy-init.md` documented three config commands that do not exist (`get-mapping-value`, `get-mapping-excludes`, `get-plaesy-structure`) | **Fixed** — real names, each verified against the binary |
| X5 | MEDIUM | `Makefile:19` ran `plaesy config detect-platform`, which does not exist | **Fixed** — `platforms detect` |
| X6 | MEDIUM | `Makefile` `clean` ran `git add -Af && git commit -m "Initial Commit"` — staging *every* file in the tree. The repo already carries three commits by that name | **Fixed** — `clean` only removes generated dirs; the old behaviour is now an opt-in `snapshot` target scoped to those paths |
| X7 | MEDIUM | `docs/scripts/plaesy-init.md` named the wrong source for the platform core file, placed agents at `.claude/roles/`, and omitted 6 of 8 `core_directories` | **Fixed** — source is `instructions/agents.instructions.md` per `platform.json plaesy.mapping.core.value`; full tree corrected |
| X8 | MEDIUM | `PLAESY_IAC_TOOL` and `PLAESY_LLM_PROVIDER` are documented as configuration inputs; **no Go code reads either** | **Fixed** — both now carry an inline "convention the agent honours, not plaesy config" warning |

## The placeholder debt — re-measured, and mostly falsified

The previous version of this section carried this table, described as
"measured":

| Syntax | Occurrences | Status |
|---|---|---|
| `{{NAME\|default}}` | 119 | Canonical |
| `{lower_snake}` | 262 | Legacy, deprecated, invisible to validators |
| `<angle_bracket>` | 93 | Legacy, deprecated, invisible to validators |
| `[UPPER_SNAKE]` | 54 | Legacy, deprecated, invisible to validators |

Those three legacy counts were still eyeballed, and they were wrong by more
than an order of magnitude. They counted *tokens*, not *fill-ins*. Almost every
hit is a documentation metavariable inside a code span or fence:

```text
docs/reference.md:122            <section> | <mapping_type>    documenting mapping.json's shape
prompts/create/diagram.md:32     <output-dir> <slugified-title> <format>    a usage line
```

`<output-dir>` in a usage line is the conventional way to write a parameter
slot. Rewriting it to `{{OUTPUT_DIR}}` would make the line unrunnable and the
document unreadable. Calling it "debt" and proposing a migration would have
made the corpus worse.

`scripts/cmd/plaesy/placeholder_drift_test.go` now draws the line where it
belongs — not by spelling, but by **role** — and counts what is left:

| Role | Skeletons | Legacy fill-ins found |
|---|---|---|
| `templates/*.template.md` + the `.plaesy/` copies | 114 | **0** |
| Everything else (`prompts/ instructions/ agents/ checklists/ docs/`) | — | 18 tokens in 8 files, all metavariables |

The skeletons are clean. The `templates/README.md` taxonomy that prompted this
section described a debt that had already been paid, by the template work
earlier in this same round. The finding is **falsified**, not fixed.

What the validator did find, and what is fixed here:

- **`.plaesy/templates/integration-examples.template.md` was an orphan.** The
  T8 fix moved the file to `docs/examples/integration-examples.md` and updated
  the registry, but left the generated copy behind — at a different length
  (46,415 vs 46,841 bytes), so it had already diverged from its source. It is
  git-tracked, which is why the deletion is part of this change. Its fences were
  also unbalanced (29 fence lines), which desynchronised the code-span stripping
  and produced the three false `{loading}` reports below.
- **`templates/user-documentation.template.md:57`** — `[FAQ]` in a table of
  contents. A markdown link label, not a fill-in. The validator now skips
  `[X](…)`; it is not a defect in the template.

Both new tests are mutation-checked: injecting `{project_name}` and `[OWNER]`
into a skeleton makes the gate fail with the right rule and line number, and a
stale count or a missing index entry in `.plaesy/instructions.md` fails the
index guard. A green test that cannot go red is not a guard.

The original instinct — enforcement over rewriting — was right. It just needed
the enforcement to distinguish a fill-in from a metavariable before it could
count anything honestly.

## The highest-leverage remaining work is a validator, not more editing

A cross-cutting assessor enumerated the checks that would have caught every
HIGH finding above. All of them extend tests that already exist and pass.

| # | Check | Catches | Where it belongs |
|---|---|---|---|
| 1 | Walk every `.md` in `prompts/ instructions/ agents/ checklists/ docs/ AGENTS.md README.md`, extract `plaesy <cmd…>`, and require each to resolve against the live cobra tree — **and each `--flag` to appear in that command's `cmd.Flags()`** | X1, X2, X4, X5, and all future flag drift | new `cmd/plaesy/corpus_cli_drift_test.go`; reuses `rootForTest()` from `docs_drift_test.go` |
| 2 | For every installed `instructions/<n>` and `templates/<f>`, assert the `.plaesy/` copy is **byte-identical** to source | the 12-instruction + 4-template drift, permanently | `internal/scaffold/generated_freshness_test.go` |
| 3 | Every `[[wikilink]]` in `.plaesy/roles/`, `prompts/`, `instructions/` resolves to a real agent or instruction file | the H10 class (filename typo). Currently **passes** (23↔23, zero orphans) — the point is to lock it in | `cmd/plaesy/corpus_*_test.go` |
| 4 | Every `mapping.json` value in `always_load`/`scope_load`/category `file` exists on disk | dangling mapping targets | `internal/detectstack/detectstack_test.go` — it already loads the file |
| 5 | `docs/metadata.json.counts` equals the measured value of all 14 keys | two counts were already wrong *within hours* of the file being "regenerated from the tree" | `cmd/plaesy/metadata_drift_test.go` |
| 6 | `plaesy validate markdown --no-baseline` as a hard gate in CI | the ratchet exists only as a *ceiling* | no new code — one step in `ci.yml` |
| 7 | `.plaesy/instructions.md` lists **exactly** the files `mapping.json` installs, and its three counts (installed / in-repo / not-installed) and its Always Loaded section match the tree | the index is hand-maintained and had drifted to 30/71/41 against a real 33/73/40, with `uncertainty-surfacing.md` and `context7-protocol.md` missing from the list entirely — an agent reading it concluded those protocols did not exist | `cmd/plaesy/instructions_index_drift_test.go` |
| 8 | Every `templates/*.template.md` and its `.plaesy/` copy uses only `{{UPPER_SNAKE}}` fill-ins, with code spans and fences excluded | the 4-to-1 convention loss described above, now gated by role rather than by spelling | `cmd/plaesy/placeholder_drift_test.go` |

Check 7 exists because `internal/scaffold/generated_freshness_test.go` cannot
cover `.plaesy/instructions.md`: that check requires generated files to be
byte-identical copies of a source, and a hand-maintained index is legitimately
not a copy. So its claims are verified semantically instead — recomputed from
`mapping.json` — rather than left to rot.

On `docs/metadata.json`: **test-guard it.** Deleting it breaks the `/doc` prompt
contract, which requires the agent to write into its `gaps`/`openQuestions`/
`inferences` sections. Generating it just relocates the same drift behind one
more generator that must be kept in sync.

## Deliberately not done

- **H3, the CI validator/glob, stays deferred** at the owner's instruction. The
  check table above is therefore a plan, not a schedule.
- ~~**`.plaesy/` was not regenerated.**~~ **RESOLVED — see `plaesy reload`
  below.** The original reason for not doing it (`rm -rf .plaesy` in a worktree
  with peers editing) was avoidable, because the right operation was never a
  wipe. `plaesy reload` refreshes the generated files in place and leaves
  memory, context and everything the user owns untouched: 28 files brought
  forward, 3 preserved, 0 drifted instructions and 0 drifted templates
  remaining, and a second run reports 0 changes, so it is idempotent.

## Verification state

`go build ./...` and `go vet ./...` clean. Targeted tests pass:
`docs_drift_test.go`, `templates_drift_test.go`, `changelog_drift_test.go`,
`internal/featurepath` (C1 regression), `internal/mdlint`.

Markdown ratchet: **1350 violations against a 1417 ceiling — 67 of headroom,
down 63 from the 1413 recorded at the start of this round.** The remediation was
net-negative in violations, so the ceiling needs no re-baselining. No global
`go test ./...` was run, per the owner's per-file testing constraint.

---

# `plaesy reload` — the missing refresh path

The cross-cutting assessment's finding #13 was the important one and I had
deferred it: `copy.go` returns early when a destination exists, `init` says so
explicitly, and `repair`/`upgrade` are stubs — so 12 of 33 installed
instructions and 4 of 58 templates had drifted with **no command that would fix
them**. The proposed remedy in the assessment was a freshness *test*, which
would catch the next drift. That is necessary but not sufficient: a test tells
you your `.plaesy/` is stale, it does not make it current, and it fails the
build instead of the file.

So the tree now has the command it was missing.

```text
plaesy reload [directory] [--ai <platform>] [--plaesy-root] [--dry-run] [--prune] [--prune-apply]
```

| It does | It will not |
|---|---|
| Overwrite drifted instructions, templates, checklists and scripts | Overwrite `memory.md`, `context.md` or `state.json` |
| Create a deleted `memory.md`/`context.md`/`state.json` (repair) | Write anything under `specs/`, `tasks/`, `analysis/`, `memory/` |
| Refresh roles and platform files when `--ai` is given | Guess the AI platform — it is not recorded on disk |
| Report created / updated / unchanged / preserved / failed per file | Delete anything from a tree you own, so a hand-written template survives |
| List orphans in tool-owned trees with `--prune` | Delete anything without `--prune-apply` as well |
| Delete orphans in the agent roles and prompt mirror with `--prune-apply` | Prune a prompt directory shared with hand-written files |
| Exit non-zero if any file failed to refresh | Touch platform files without `--ai` |

Measured on this repo, which is itself the drift case: **28 files updated,
77 unchanged, 3 preserved, 0 failed.** Instruction drift went 13 → 0, template
drift 4 → 0, and a second run reported `0 created, 0 updated` — idempotent.

## The one thing reload could not do: delete

The table above said "delete anything" for years of this project's life, and
the reason was sound. A file in `.plaesy/` with no counterpart in the sources
is usually the user's work, and deleting it on sight would destroy the thing
reload exists to protect.

The same file in a generated mirror is not user work. It is a copy that
outlived its source, and it is worse than dead weight: **it is a complete,
loadable command.** When `/create:doc:design` moved to `/spec:design`,
`prompts/create/doc/design.md` was renamed and
`.kilo/commands/create/doc/design.md` stayed behind, still titled
"`/create:doc:design` command instructions" and still invokable. Both commands
answered; one of them was a file the project no longer had. An agent that
picked it would report having followed the project's own convention.

So the discriminator is **ownership, not path shape**:

- **Tool-owned trees** — the agent roles and the platform prompt mirror. Every
  file is a copy reload itself wrote, so rsync's `--delete` semantics are
  correct. `scripts/configs/platform.json` marks a platform's prompt
  destination with `"prune_prompts": "true"`.
- **Mixed trees** — everything else under `.plaesy/`. These are left alone
  entirely. The right mechanism here is manifest-ownership, not path pruning: a
  manifest recording the path *and content hash* of every file reload wrote, so
  an orphan the user has since edited is recognised as work rather than garbage.
  That is not implemented, and `--prune` is the honest substitute for it.

Two opt-ins, not one, because `.kilo*` and `.claude*` are gitignored and
version control cannot undo a deletion. `--prune` prints; `--prune-apply`
removes. The three safety properties that make the second one defensible:

1. **The source is checked first.** A missing or empty source directory makes
   the expected set empty, which would make every file look like an orphan. A
   misconfigured `--plaesy-home` would then delete the tree. Prune skips
   instead.
2. **Shared destinations are denied.** `.cursor/rules` and `.qoder/rules` hold
   hand-written rules next to copied prompts, so a file there is
   indistinguishable from a stale mirror. The config marks them prunable and
   the deny list still wins — a mutation check confirms that removing the deny
   list entry deletes a hand-written file.
3. **The platform must be named.** Reload refuses to guess the platform when
   copying; guessing it when deleting would prune a directory belonging to
   some *other* tool.

## How the implementation avoids the obvious trap

`reload` reuses the existing copy helpers rather than reimplementing them. The
mechanism is a `copyPolicy` threaded into `copyFile`:

```go
type copyPolicy struct {
    overwrite bool              // init's zero value: false, so init is unchanged
    dryRun    bool
    protected map[string]bool   // never overwrite (but still create if missing)
    report    *fileReport
}
```

The zero value reproduces `init`'s behaviour byte for byte, which is why all 15
pre-existing `TestInit*` tests pass untouched. Duplicating the copy loops instead
would have been shorter to write and would have created the second source of
truth this framework forbids.

Two genuine bugs surfaced while writing the tests, both of which would have
destroyed data on the first real run:

1. **The `memory.md` "create empty" fallback fired on an existing file.** The
   old chain's empty-scaffold branch was reachable for a file that already
   existed, so a reload against a home without `memory.template.md` blanked out
   the user's memory. The branch is now gated on there being nothing to lose.
2. **`context.md` was clobbered through a subtler path.** `copyFile` correctly
   reported "protected, copied=false", and the caller read that as *failure* and
   took the empty-file branch. Protection that is honoured in one function and
   unwound in the next is worse than no protection, because it looks like it
   worked.

Both were caught by `TestReloadPreservesUserData`, which asserts the content of
all eight user-owned paths rather than asserting that a function was called.
A test that checked the call would have passed while the data was destroyed.

## What still stands

The freshness test (cross-cutting check #2) is **not** implemented and is still
worth doing. `reload` fixes drift when someone runs it; only a test stops it
from arriving. The two are complementary, and the reason to build the command
first is that it gives the test something to compare against.

---

# Enforcement round — the checks, built

The previous section argued that a validator outranks more editing. This is
that work: six new test files, 23 test functions, all of which fail the build
via the `go test ./...` that `ci.yml` already ran. **No CI file was edited** —
the enforcement was always one step away; nobody had written the step.

| File | Guards | Would have caught |
|---|---|---|
| `cmd/plaesy/corpus_cli_drift_test.go` | every `plaesy <cmd>` in `prompts/ instructions/ agents/ checklists/ docs/ README.md AGENTS.md` resolves against the live cobra tree, and every `--flag` exists on that command | `generate-template`, `generate-pipeline`, `detect-stack`, `config detect-platform`, the three `get-mapping-*` names, the undocumented `--install` |
| `internal/scaffold/generated_freshness_test.go` | every installed `.plaesy/instructions` and `.plaesy/templates` file is byte-identical to source; `plaesy reload` is a no-op on this repo | the 13-file instruction drift and 4-file template drift |
| `cmd/plaesy/metadata_drift_test.go` | all 14 `docs/metadata.json` counts recomputed, plus every `filesRead` path exists | the 8 counts that were wrong |
| `cmd/plaesy/owasp_taxonomy_drift_test.go` | the OWASP category codes and names are identical in the two security files | the A01/A10 and A05/A06 merges |
| `cmd/plaesy/reload_cmd_test.go` | `reload`'s output contract, exit code, `--help`, and `--prune` / `--prune-apply` end to end | — (guards the new command) |
| `internal/scaffold/reload_test.go` | `reload` preserves all eight user-owned paths; prune reports without deleting; a missing or empty source deletes nothing | — (guards the new command) |
| `internal/scaffold/prune_test.go` | every way a platform fails to qualify as prunable; `withinRoot` refuses escaping paths; a candidate outside every owned tree is refused | — (guards the new command) |

## Every guard was proven non-vacuous

A drift test that has never failed is indistinguishable from a test that cannot
fail. Each was checked by breaking the tree on purpose:

| Guard | Injected fault | Result |
|---|---|---|
| freshness | appended a line to an installed instruction and a template | failed, named both, `reload` fixed both |
| CLI commands | `plaesy generate-pipelins` in a prompt | failed — and correctly ignored the allowlisted `generate-pipeline` |
| CLI flags | `plaesy graph --watcht` in `docs/overview.md` | failed, and distinguished it from the real `--watch` |
| OWASP | re-merged `A01 & A10` | failed |
| metadata | 3 new test functions | failed with the delta |

Three of these tests were **wrong on first run** and had to be fixed, which is
the more useful half of the story:

1. The command index registered only leaf names, so `features create` reported
   as a typo. 116 false findings.
2. `pflag.Lookup` was called with `--ai` instead of `ai`, so every flag in the
   corpus reported as missing. 12 false findings.
3. The OWASP comparison accepted a prefix relationship — which is precisely
   the defect ("broken access control & SSRF" matching "broken access control").
   The test passed **on the exact bug it was written for**, and only failed once
   the comparison was tightened. A lenient comparator is worse than no test,
   because it reports safety it did not verify.

Two other test-design notes worth keeping. The CLI walk needs longest-prefix
matching, or `plaesy trim run go build` and `plaesy context update claude` are
reported as typos — a check that cries wolf twenty times gets deleted, not
fixed. And the command test has to redirect the real file descriptor: the
logger writes to `os.Stdout` directly, like every other command here, so
capturing cobra's buffer would have asserted nothing. It also deadlocks without
a concurrent reader, because a reload of this repo writes past the 64 KiB pipe
buffer.

## The placeholder migration, measured

327 fill-ins across 54 files converted to `{{ }}`. Not a regex over the tree:
`[X](url)` is a markdown link, `<div>` is HTML, `{` and `}` are JSON and Go
template syntax, and `<AuthResult>` in a TypeScript example is shaped exactly
like a fill-in. Each form was converted against an explicit allowlist, and the
ten tokens that remain are documented in `templates/README.md` with the reason
each is *not* a placeholder.

A fifth syntax turned up that nobody had catalogued: `{web_app|mobile_app|...}`
value enums, which have no canonical form because a pipe inside `{{ }}` splits
a markdown table cell. Written as `{{PROJECT_TYPE}}` plus a plain "one of:" list.

## Two numbers moved, and both are honest

`cmd/plaesy` coverage 74.6 → 74.8 and `internal/scaffold` 86.9 → **84.6**. The
second went down. `reload` added ~290 statements across the two packages, about
half of them `if err != nil` returns on stat, `MkdirAll` and copy calls that
need a filesystem which refuses reads to reach — the same unreachable class the
baseline already documents. The reachable behaviour is tested; the remainder is
error handling. It is recorded as measured, with the reasoning, rather than
raised to make a red build green.

## Still open, and deliberately so

- **H3 proper** — a `ci.yml` step for `validate markdown --no-baseline`. The
  Markdown ratchet is still only a *ceiling* (1353 against 1417), so a new
  violation is permitted until the total rises 64. Every other guard in this
  section is a hard gate; that one is not.
- **`docs/assessment/` is excluded from the CLI walk**, because those dated
  reports quote commands that no longer exist precisely as a record of what was
  broken. Rewriting them would make the report lie about what it found.

---

# Autonomous loop — iterations 1 and 2

Run under `/loop`, `technical` dimension, per-file test targeting only. Loop
state advanced in `.plaesy/state.json` to iteration 2/10, `consecutive_failures: 0`,
then paused: iteration 2 surfaced no new assessable defect, and the two items
that remain are decisions for the maintainer, not fixes an agent should make.

## Iteration 1 — six defect groups, fixed

**Dead relative links (13 files).** The quietest class in the corpus: they render
as ordinary blue text, never fail a build, and an agent following one concludes
the file was deleted. 596 relative links swept, in four shapes:

| Where | What |
|---|---|
| `docs/instructions/README.md` (37) | wrote `./<n>.instructions.md` as if the files sat beside the doc; they are two levels up in `instructions/` |
| `prompts/{doc,fix,implement,optimize}.md` (8) | wrapped the href in backticks, so the backticks became part of the URL and the link could never resolve |
| `docs/reference.md:202` | pointed at `docs/scripts/create.md`, which does not exist — the `/create` prompt is the only spec of that command |
| `README.md`, `docs/README.md` (3) | claimed a top-level `testing/` directory that has never existed; the real folder is `docs/testing/` |

**Four more legacy fill-ins, found only after widening the placeholder rule.**
`templates/context.template.md` carried `{tasks-topics}` immediately beside a
canonical `{{STATUS}}` in the same path; `status.template.md` had
`{{NUMBER}}-{branch-name}`, the identical mixed shape. Widening the rule to
admit hyphens then surfaced two more on its own — `spec-structure-folder` at
`{tasks-id}` and `status` again. The original rule was
`\{[a-z][a-z0-9_]*\}`, and every one of those placeholders contains a hyphen.
A check that cannot see a whole class of the thing it checks is worse than no
check, because it reports green.

**Three guards added, each mutation-checked** so that a green test is known to
be able to go red: `dead_link_test.go`, plus the index and placeholder guards
from the preceding round. An injected `{project_name}`, a stale index count, and
a link to a file that does not exist each turn the right guard red with the
right line number.

## Iteration 2 — no new defect; one scan falsified twice

**Dangling file references: zero.** A prose sweep for `*.instructions.md` and
`templates/*` names that resolve to nothing reported **191** hits. Every one was
a bug in the sweep — it compared a filename *stem* against full filenames, so it
flagged `go`, `sql`, `terraform` and 60 others that all exist. Corrected, it
reports 7, and all 7 are legitimate prose: four are merger provenance notes
("merged from the former `plaesy-trim.instructions.md`"), one is a glob
(`*-design-principles.instructions.md`), one is a dated historical note, one is a
`<name>` pattern example. **The axis is clean.**

This is the second finding in two rounds that a hand-rolled scan produced a
number an order of magnitude too high. Both times the corpus was fine and the
measurement was wrong. That is the argument for the general one: encode the
check, let it run in CI, and stop re-deriving it with a regex.

**OWASP taxonomy conflict: already resolved.** A peer board message flagged
`security-and-owasp` and `security-audit` as using two irreconcilable A01–A10
schemes and marked it "not mine to fix". It is already reconciled — both files
carry the 2021 labels, both document the decision in place, and
`owasp_taxonomy_drift_test.go` locks it. The flag predated the fix. Recorded here
so nobody re-opens it.

## Left for the maintainer

- **A 14.9 MB Windows PE binary is committed at `plaesy`.** `/plaesy` in
  `.gitignore` does nothing, because ignore rules do not apply to a tracked
  file. It entered through the old `clean` target's `git add -Af` (X6). The fix
  is `git rm --cached plaesy`; it is not applied here because it changes what a
  clone contains.
- **Six source files are untracked.** `.gitignore` had a bare `plaesy`, which
  matched the path *component* in `scripts/cmd/plaesy/` and silently ignored
  every new file there while leaving already-tracked ones alone. The rule is
  fixed and the files are now visible to `git add`, but they are not committed:
  `reload.go` and five drift tests, including `corpus_cli_drift_test.go` and
  `metadata_drift_test.go` — two of the six checks this report credits as
  highest-leverage. They build and pass, and would not survive a clone.

## Verification

`gofmt` clean, `go vet ./...` clean, `go build ./...` clean. `cmd/plaesy` (33s),
`internal/scaffold`, `internal/mdlint`, `internal/detectstack` all pass. Markdown
ratchet **1355 against the 1417 ceiling** — 62 headroom, ceiling never rewritten.
No global `go test ./...`, per the owner constraint. `docs/metadata.json` counts
moved 72→75 / 705→711, and `metadata_drift_test.go` caught that on each of the
three new test files, which is the check working.

---

# Corpus round — 2026-09-27 (scope: everything except `scripts/`)

Goal: 100/100 overall on the prompt corpus. 266 files, ~51,000 lines. Five axes
swept mechanically, because every finding this session has come from a sweep
rather than from reading, and three of the last four were my own measurement
bugs rather than corpus defects.

| Axis | Scope swept | Result |
|---|---|---|
| Dangling anchor fragments | 112 fragments | **17 dangling, fixed** |
| Dead relative links | 265 files | 0 |
| Self-reported counts | 73/23/7/64/28/75 nouns | 0 stale |
| Cross-file verbatim duplication | 51k lines, ≥6-line blocks | 2, both defensible |
| Placeholder syntax | 114 skeletons | 0 |

## The one real finding, and it was severe

**`/create:template:api` vs `/create:templates:api`.** `prompts/create/templates/ci.md`
contradicted *itself*: its heading and its `CALL` line said singular, its usage
examples said plural. Same in `infra.md` and `project.md`, and `prompts/create.md`
was singular throughout. Four files, 14 references, two names for one command —
and a prompt that cannot name itself consistently is a prompt an agent invokes
wrong.

Three independent signals settle which is canonical: the on-disk directory is
`prompts/create/templates/`, the generated `.kilo/commands/create/templates/`
mirror is plural, and `docs/prompts/README.md` is plural. Normalised to plural
everywhere.

## A second copy of a procedure that must not drift

`prompts/create/images.md` and `prompts/create/storyboard.md` each carried the
same 11-line image-provider resolution procedure — same env var name, same
override-file semantics, verbatim. Two prompts that must agree exactly on a
config contract, each holding its own copy. The definition now lives once, in
`prompts/create.md` → Provider Resolution, and both point at it.

Worth recording how that went: the first attempt left the numbered list in place
and appended a pointer. That is the worst outcome — it reads as thorough and is
still a second copy. The second attempt made both pointers identical, which is
the same defect one level down. It took a one-line, scope-specific pointer to
actually remove the duplication. **A pointer that is long enough is still a
duplicate.**

## Two of my own scans were wrong, and that is the finding

The anchor sweep first reported 30 dangling fragments. 13 were bugs in the scan:
a slug function that collapsed runs of whitespace, so the *correct* anchor
`#382-authentication--authorization` (GitHub gives each space its own dash)
looked broken; hex colour literals; links inside code spans. The dedupe script
then reported a target missing that was sitting on disk, because it resolved
`../../prompts/…` against the repo root instead of the document's own directory.

Two hand-rolled scans in a row produced numbers an order of magnitude off, in
both directions, while the corpus was fine. That is the strongest argument this
session for the pattern it kept converging on: encode the check, let CI run it,
and never re-derive a count with a regex.

## Left deliberately, with reasons

- **`prompts/create/api.md` sits outside `templates/`** while `ci`, `infra` and
  `project` sit inside it, and the generated mirror agrees with the disk. No
  routing table in the repository defines these sub-commands, so which spelling
  is authoritative cannot be determined from here. Moving the file on a guess
  would break a router that cannot be inspected. **Needs review.**
- **`docs/reference.md:377` and `docs/scripts/platforms.md:11` share a 7-line
  subcommand table verbatim.** A reference index and a per-command detail page
  legitimately repeat a small table, and `docs_drift_test.go` checks that both
  mention every subcommand but not that the tables match. De-duplicating means
  either gutting the reference's self-containedness or breaking a passing guard
  for a 7-line win. Recorded as accepted duplication with a known unguarded
  drift risk.

## What "100/100" can and cannot mean here

Every axis above is *provable* and is now guarded — ten tests in total, each
mutation-checked, so a green result is known to be able to go red. That part is
at 100% clean: no known defect remains that a mechanical check can find.

What it is not is a 100/100 overall score. The remaining dimensions — whether a
file is actionable without guessing, whether it sits at the right abstraction
level, whether its prose is concise, whether two files agree on *meaning* rather
than on links — need all 51,000 lines read, not swept. The assessor subagents
that produced the earlier rounds timed out on every attempt this session, so
that read did not happen. Claiming the number without it would be the one
unacceptable outcome here.
