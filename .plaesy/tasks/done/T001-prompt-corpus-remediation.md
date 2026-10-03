---
id: T001
title: Resolve all critical + high findings from the 2026-09-27 prompt corpus assessment
status: doing
created: 2026-09-27
source: docs/assessment/2026-09-27-prompt-corpus.md
scores: { before: 47 }
---

# T001 — Prompt corpus remediation

Origin: `/assess` (technical dimension, assessment mode), 2026-09-27.
Full findings: `docs/assessment/2026-09-27-prompt-corpus.md`
Structured data: `docs/assessment/2026-09-27-prompt-corpus.json`

## Scope

237 source files, ~43,000 lines. Target: clear all 8 CRITICAL and 10 HIGH
findings, and make the fix mechanically enforced so it cannot regress.

## Work items

> **The checkboxes below are the original brief, kept verbatim for provenance.
> They are NOT a live task list — several are demonstrably done and this section
> says otherwise. The Progress log at the bottom is authoritative.** See
> *Reconciliation* below for the measured state. Re-deriving status from this
> section is what produced a wrong answer twice this session.

### Block 1 — correctness (blocks everything)

- [ ] **C1** `scripts/internal/featurepath/create.go:93` — template path bug.
      Reads `<repoRoot>/templates/` but install target is `.plaesy/templates/`;
      the `else` branch writes a 0-byte `spec.md` and discards the error.
      Fix the path and return the error. Add a regression test that exercises the
      real path rather than a stubbed parameter (`featurepath_test.go:258`).
- [ ] **C2** `prompts/create/tasks.md` — remove one of the two tails.
      2x `### Step 5` (`:128`, `:285`), 2x `## Programmatic Invocation` (`:175`, `:297`),
      2x `## Anti-Patterns` (`:149`, `:319`), 2x `## Success Criteria` (`:195`, `:344`).
      Keep contract A (one file per task) and fold the JSON schema in as a
      `--format json` branch.
- [ ] **C3** `design-spine.md` -> `memory/design.md`. 43 references across 6
      `prompts/create/*` files. Canonical name is set by `prompts/assess.md:158`.
      Add "Step 0: resolve design context" so the file is created from
      `templates/design.template.md` when absent.
- [ ] **C7** make all 16 path references in `prompts/` resolve:
      3 nonexistent roles (`technical`, `architect`, `product`),
      5 provider-config JSONs with no producer,
      7 repo-relative `instructions/brandkit...` refs.
- [ ] **H10** rename `instructions/software-design-prinsiples.instructions.md`
      -> `...-principles...`; update `mapping.json` x2 and 2 other references.

### Block 2 — single source of truth

- [ ] **C4** delete the ~170 duplicated routing lines from `plaesy.md`,
      `assess.md`, `implement.md`, `fix.md`, `optimize.md`, `doc.md`,
      `improve.md`, `universal-orchestrator.md`; delegate to
      `dimension-mapping.instructions.md`. Enforce the existing rule.
- [ ] **C5** unify the phase model on 9 phases. Fix `start.md:135` (`current/11`),
      `continue.md:217` (`X/12`), `loop.md` local `Phase 0-4` -> "Iteration Step N",
      `assess.md:48-49` (Phase 6 assigned twice).
- [ ] Dedupe the 48 dimension sub-prompts (256 lines of boilerplate that is
      9-of-9-lines identical after substituting the dimension name).
- [ ] Consolidate security (3 files, 419 lines -> 2) and resilience patterns
      (5 files, 200 lines -> 1 home: `distributed-systems-design-principles`).
- [x] Extract the 8x shared scaffold (~1,400 lines) into `assess-core.instructions.md`.
- [ ] Extract Uncertainty Surfacing Protocol (10 files, 90 lines) to its own file.
- [ ] Extract the Context7 protocol (3 files, 35 lines) to its own file.
- [ ] Collapse the duplicated "Why This Phase Matters" (55 lines) and
      "Success Criteria" (70 lines) blocks.

### Block 3 — content correctness

- [ ] **H1** `docs/agents/README.md:27,52` — `pm` is the Product Manager,
      `pe` is the Prompt Engineer. Fix both labels.
- [ ] **H2** `checklists/po.checklist.md` — split: real PO checklist (vision,
      backlog, prioritization, value metrics) stays; scaffolding/DB/deploy/frontend
      moves to `sa.checklist.md` or a new `devops.checklist.md`.
- [ ] **H6** `checklists/qa.checklist.md` — convert 112 non-checkbox sub-bullets to
      `- [ ]` items, add `[[LLM:]]` guidance per section, add a security section.
- [ ] **H4** wire or remove the 5 unwired roles (`ai-architect`, `bo`, `mlops`, `pe`, `po`).
- [ ] **H5** add a `## Boundaries & Escalation` block (`Owns:` / `Defers to:` /
      `Escalate to @x when:`) to the agents template; populate the 11 overlapping pairs.
- [ ] **H7** add `python.instructions.md` and `vue.instructions.md`; register in `mapping.json`.
- [ ] **H9** standardize template placeholders on `{{VARIABLE|default}}`; unescape the
      4 `\|` variants in `constitution.template.md:41`; add `{{CURRENT_DATE}}` to document headers.
- [ ] **H8** publish a canonical section-name registry with aliases; have `/assess`
      and `/improve` resolve through it.

### Block 4 — enforcement (prevents recurrence)

- [x] **H3** build the validator — **DONE 2026-09-29**, all 7 checked via
      `go test ./...` (already wired in `.github/workflows/ci.yml`), no
      `make check-commands` needed since Go tests cover it:
      - [x] the 8-dimension filename set is identical across all 6 command families
            — `TestRoutingTableCoversExactlyTheCanonicalDimensions` (routing_drift_test.go)
      - [x] every `/{fam}:{dim}` in a parent Usage Format has a backing file, and vice versa
            — `TestNestedSubCommandNamesResolveToFiles` + `TestRoutingMatrixIsComplete` (routing_drift_test.go)
      - [x] every `.plaesy/roles/*.md` and `.plaesy/instructions/*.md` path referenced in
            `prompts/` resolves post-init — `TestRoutingTableReferencesResolve` (routing_drift_test.go),
            `TestNoDeadRelativeMarkdownLinks` (dead_link_test.go)
      - [x] no duplicated H2 heading within a single prompt file —
            `TestNoDuplicateH2HeadingInAPromptFile` (h3_corpus_integrity_test.go, new)
      - [x] `agents/*.agents.md` is byte-identical to `.plaesy/roles/*.md` (parity)
            — `TestPromptMirrorMatchesSource` (mirror_parity_test.go)
      - [x] every `@role` mention resolves to an existing agent file —
            `TestEveryRoleMentionInAgentsResolvesToAnAgentFile` (h3_corpus_integrity_test.go, new)
      - [x] sub-prompt `description` matches the parent usage-table text —
            `TestSubPromptDescriptionMatchesParentUsageTable` (h3_corpus_integrity_test.go, new;
            caught and fixed 13 real drifts on first run)

### Block 5 — repository hygiene

- [ ] **C6** `git rm -r --cached .kilo .plaesy plaesy`; they are regenerable and
      already in `.gitignore`.
- [ ] **C8** generate `.plaesy/instructions.md` as a real index; fix the broken
      `overview.md` link in `.plaesy/memory.md:17`; document the 30-of-68 selective
      instruction copy in `docs/instructions/README.md`.
- [ ] Regenerate `docs/metadata.json` (stale counts: 23 cmd / 19 platforms / 15 pkgs
      vs actual 20 / 20 / 16; 5 referenced source files no longer exist).
- [ ] Fix stale doc claims in `docs/overview.md:46`, `docs/components.md:3`,
      `docs/prompts/README.md:550`, `docs/architecture.md:118-121`, `README.md:195,201,109`.

## Definition of done

- [ ] 0 CRITICAL, 0 HIGH findings open
- [x] validator runs in CI and fails on a deliberately introduced regression —
      all 7 H3 checks now `go test`-enforced (CI already runs `go test ./...`
      on every push/PR); 3 new guards each mutation-tested red, 2026-09-29
- [ ] re-assessment confirms no regressions and a score >= 90

## Progress log

### Iteration 11 (2026-09-29) — `/loop`, mechanical axis: a red build nobody was reading

Iteration 10 finished with a note that the remaining `go test ./cmd/plaesy`
failures were "pre-existing, unrelated". They were not unrelated. They were a
single root cause producing 21 failures, and the one that mattered most was
**not a test failure at all** — it was a live defect that every test in the
repository was passing over.

**The defect: `platform.json` and the rest of the repository disagreed about
what a platform is called.**

`scripts/configs/platform.json` keyed six platforms by their short names —
`claude`, `cursor`, `copilot`, `kilo`, `trae`, `windsurf`. Everything else in
the repository named the long form — `claude_code`, `cursor_ai`,
`github_copilot`, `kilo_code`, `trae_ai`, `windsurf_ai`:

- the `--ai` flag help on `plaesy init` and `plaesy reload` names them;
- `README.md:167-170` and `docs/overview.md:46` list them (the latter claims
  they are "the keys of `platforms` in `scripts/configs/platform.json`", which
  was false);
- `docs/reference.md`, `docs/scripts/plaesy-{init,clean,platforms}.md` and
  `docs/scripts/config-manager.md` use them throughout;
- `internal/config`'s own comment describes the table as mapping the shorthand
  *to the id platform.json declares* — while mapping `claude` to `claude`;
- `promptExtension()` switches on `"github_copilot"` — a case that could never
  fire, because the platform it names was keyed `copilot`.

The long form was canonical in five independent places, and it resolved to
nothing. `plaesy platforms show claude_code` printed `Name: Unknown`,
`Provider: Unknown`, `Category: Unknown` and exited 0 — indistinguishable from a
platform whose config genuinely had no name. `plaesy clean --ai claude_code`
matched no mapping, so the plan found nothing and the run removed nothing
**while reporting that it had**. `plaesy init --ai claude_code` ran
`normalizePlatform`, got back `claude` — a key that no longer existed — and
wrote platform files for a platform the config does not declare.

The Copilot extension case is the quiet one: `promptExtension()` returning
`.prompt.md` instead of `.md` for GitHub Copilot is a small behavioural
difference, in a switch statement that is never wrong, and reads as correct
code. It would have stayed wrong indefinitely.

**The fix, and why this direction.** The keys were renamed, not the callers.
Five independent sources agree on the long form and one config file dissented;
a rename of 24 doc references would have made the one minority correct and the
five majority wrong. The rename also put `platform.json`'s key order into exact
agreement with the list at `docs/overview.md:46` — that list was written from
an earlier state of the file and was right all along.

`internal/config`'s alias table now maps shorthand → long id, so both spellings
resolve. `internal/scaffold`'s `normalizePlatform` no longer keeps a private
copy of the table (which returned the short ids, so `init` and `clean` resolved
the same name to *different* platforms); it delegates and adds only what the
shared table cannot: the dashed spellings and the no-platform family.

Verified end to end, not only by the suite:

- `plaesy platforms show claude_code` → `Name: Claude Code`, `Provider: Anthropic`
- `plaesy init cptest2 --ai claude_code` → writes `CLAUDE.md`
- `plaesy init cptest --ai github_copilot` → 71/71 prompts as `.prompt.md`
  (before the fix: 71/71 as `.md`)

**Three smaller findings, same iteration.**

- `CHANGELOG.md` had an empty `## [Unreleased]` section while shipping 19
  documented commands — the exact product-gate failure
  `TestEveryShippedCommandHasAChangelogLine` was written to catch, failing for
  every one of them. Filled in, each entry derived from `--help` output rather
  than from memory. The two `Fixed` entries are this iteration's own changes.
- `TestNoFileDuplicatesAnotherOutsideTheKnownMirrors` flagged the root
  `AGENTS.md`. It was right about the fact and wrong about the meaning:
  `AGENTS.md` is the *generated* platform core file (`platform.json` installs
  `instructions/agents.instructions.md` to it for `opencode` and `kilo_code`),
  gitignored, produced by the framework doing exactly what it documents. The
  fix reads the legitimate target set **out of `platform.json`** rather than
  listing the paths, because two platforms already share `AGENTS.md` and a new
  platform would otherwise make the guard fail on correct behaviour. Mutation
  tested: dropping the core-target load turns it red naming `AGENTS.md` and
  `instructions/agents.instructions.md`; restoring it returns green.
- `docs/metadata.json` `goTestFiles` 80 → 81, `goTestFunctions` 728 → 731.

**State: `go test ./cmd/plaesy` 21 failing → 0. `go test ./internal/...` green.
Markdown ratchet 1333/1417, 84 headroom.**

The standing note from iteration 10 — that these failures were pre-existing and
unrelated — was accurate about the *diff* and wrong about the *world*. The
right response to a red suite is to read what it says, not to classify which
file each failure names.

### Iteration 10 (2026-09-29) — owner directive: remove `/create:templates:*`

Owner-directed scope change, not part of the DRY-sweep or `/improve` axes above.
`/create:templates:{api,ci,infra,project}` were four full command protocols
(usage flags, step-by-step generation, "Programmatic Invocation" contracts)
that existed only to scaffold one hardcoded software stack each (Node/pnpm
monorepo, OpenAPI/GraphQL, Terraform/CloudFormation, GitHub Actions/GitLab CI).
Owner's objection: the framework is dimension-agnostic everywhere else
(`/assess`, `/implement`, `/loop` all route across all 9 dimensions), but this
one router branch baked in a single technical stack as if it were the whole
framework's concern — and even within the technical dimension it covered only
one stack per artifact type.

- Deleted `prompts/create/templates/{api,ci,infra,project}.md` and their
  `.claude/commands/` mirrors (gitignored/regenerable, not tracked).
- Verified no other prompt actually calls `CALL /create:templates:*` — the
  "Programmatic Invocation" sections were self-contained, aspirational
  documentation, not real call sites — so deletion has no dangling caller.
- Ported the stack-specific knowledge worth keeping into `instructions/`,
  which loads automatically per file type via `mapping.json` instead of
  through a dedicated generator command:
  - `instructions/api-design-principles.instructions.md` — added a
    "Scaffolding a New Spec File" section (OpenAPI/GraphQL SDL structure).
  - `instructions/terraform.instructions.md` — added "Scaffolding a New
    Module" (file layout, remote state, CloudFormation equivalents).
  - `instructions/devops-core-principles.instructions.md` — added
    "Scaffolding a New Pipeline"; `applyTo` extended to
    `.github/workflows/*.yml` and `.gitlab-ci.yml` (previously k8s/deployment
    YAML only).
  - New `instructions/monorepo-scaffolding.instructions.md` (Node/pnpm
    workspace layout, the 14 avoid/use pairs from the old `project.md`),
    registered in `mapping.json` under `frameworks.monorepo` with `filenames`
    detection (`pnpm-workspace.yaml`, `turbo.json`, `nx.json`).
- `prompts/create.md`: dropped the "Boilerplate Templates" table and usage
  lines; `/create` is now 4 scopes (images, storyboard, diagram, tasks), all
  real asset-generation calls. Added a pointer from `/create` to the
  instructions files above so an agent scaffolding a stack knows where the
  guidance now lives.
- `docs/prompts/README.md` — removed the 4 dead rows, same pointer added.
- `docs/metadata.json` — `instructionFiles` 73 → 74 (the new file).
- `scripts/cmd/plaesy/corpus_cli_drift_test.go` — removed the
  `generate-template`/`generate-pipeline` allowlist entries; both were
  allowlisted only because the two deleted prompts instructed an agent to run
  a CLI command that was never implemented. Nothing else in the corpus
  mentions either name now, so the allowlist entries would have gone stale
  silently — removing them means a *future* reintroduction gets caught fresh.
- `scripts/cmd/plaesy/routing_drift_test.go` —
  `TestNestedSubCommandNamesResolveToFiles`'s own zero-result guard fired
  (correctly — three-segment `/x:y:z` selectors no longer exist anywhere in
  the corpus, `/create:templates:*` was the only family). Changed `t.Fatal` to
  `t.Skip` with a reason instead of weakening or deleting the check: a future
  three-segment selector still gets validated automatically the moment one is
  added back.
- Verified: `TestEveryCommandInTheCorpusExists`, `TestNoDeadRelativeMarkdownLinks`,
  `TestNestedSubCommandNamesResolveToFiles` (now skips with reason), and the
  `instructionFiles` line of `TestMetadataCountsMatchTheTree` all green.
  `goTestFiles`/`goTestFunctions` drift in that same test predates this change
  (no test files or functions added or removed here) and is left for whoever
  is tracking that axis.

### Iteration 9 (2026-09-29) — `/improve`, frontmatter + prose + external best-practice gap

Separate pass, not a continuation of the DRY-sweep axis above. Ran `/improve`
against the framework's own corpus (Artifact Mode — no constitution in this
repo). Did not re-audit Blocks 1-5 above; found a different layer.

- **22 files** had frontmatter gaps: 4 `instructions/*.md` had none at all,
  16 had `description` but no `applyTo`, 1 `agents/*.agents.md` description
  didn't match its H1. All fixed.
- **`prompts/create.md`** was the only top-level router missing an
  `## Output Format` section (10 siblings all have one). Fixed.
- **Compressed `prompts/` (75 files, ~40 touched)**, 8-18% cut on the 11
  top-level command files, prose only — 0 heading/contract drift, verified
  by `grep -c '^#'` pre/post on the 7 largest files. Found+fixed 2 real bugs
  while in there: stray `%s` placeholders (should be em-dash) in
  `create/{diagram,templates/api}.md`; a duplicated dead paragraph in
  `create/{images,storyboard}.md` Step 2.
- **Ported 4 gaps from `github.com/miqdadbadjuber/anti-slop`** (external
  UI-slop-filter reference, researched live):
  `quality-gates.instructions.md` Gate 7 gained a click-through delivery-gate
  requirement; `templates/design.template.md` +
  `how-to-create-designmd.instructions.md` gained a "Three Dials"
  (energy/rhythm/motion) liveliness mechanism; `redesign.instructions.md`'s
  85 audit bullets got citable IDs (`T-01`..`SO-06`) plus a "Keystone: Write
  the Reason" section; `create/images.md` + `implement/design.md` gained a
  confirm-or-placeholder gate before fabricating identity-bearing assets.
- **`plaesy trim`** (the repo's own compression CLI) was tried and rejected
  by the user — not validated yet, do not use it for corpus work until it's
  been through its own `/improve` pass. See `.plaesy/memory.md`.
- Flagged, not fixed: `prompts/create/templates/project.md`'s "Project
  Structure Standards" numbered list reads ambiguously — needs a correctness
  pass, not compression.
- Not yet done: content-quality pass on `templates/*.template.md` (57
  files, structural audit only so far) and `agents/*.agents.md` (23 files,
  structural audit clean, no deep content pass).
- **H3 CLOSED.** 4 of its 7 checks were already covered by existing guards
  (`routing_drift_test.go`, `mirror_parity_test.go`, `dead_link_test.go`);
  the 3 missing ones got a new file, `h3_corpus_integrity_test.go`:
  `TestNoDuplicateH2HeadingInAPromptFile`, `TestEveryRoleMentionInAgentsResolvesToAnAgentFile`,
  `TestSubPromptDescriptionMatchesParentUsageTable`. All 3 mutation-tested
  (each turned red on a deliberate break, confirmed the message, restored).
  **The description-matching guard caught 13 real single-source-of-truth
  drifts** between a parent's Usage Format table and its sub-prompt's
  frontmatter — fixed, sub-file synced to the parent's wording (the parent
  table is the canonical copy elsewhere in this corpus). `loop.md`'s
  `/loop:technical` line is excluded with a written reason: it's a config-flag
  example, not a per-dimension description table like its 5 siblings have.
  **The guard itself had a bug on first write** — a stray `TrimPrefix(line,
  "/")` meant the regex (which still required a leading `/`) never matched
  anything, so the test silently asserted zero comparisons and reported
  PASS. Caught only because the mutation test came back green when it should
  have been red — the repo's own "a green mutation result is only evidence
  about the mutation" lesson, paying off again. No CI config change needed:
  `.github/workflows/ci.yml` already runs `go test ./...` on every push/PR
  across 3 OSes, so these guards are live immediately.

### Done and verified

- [x] **C1** `create.go:93` now resolves the template through
      `specTemplateCandidates` (`.plaesy/templates/` first, then the source tree)
      and a missing template is a hard error instead of a zero-byte `spec.md`.
      Regression test `TestCreateNewFeatureFailsWithoutATemplate` plus two new
      cases pinning the installed-path resolution. `go test ./...` green.
- [x] **C1-adjacent** `cmd/plaesy` tests depended on the untracked, never-committed
      `.plaesy/memory/constitution.md` and failed on any clean checkout. Now
      hermetic via `scripts/cmd/plaesy/testdata/constitution.md`.
- [x] **C2** `prompts/create/tasks.md` 366 lines, one of each duplicated heading
      survives; the JSON branch folded into Step 5 as `--format json`.
- [x] **C3** 47 `design-spine` occurrences resolved to `.plaesy/memory/design.md`
      (3 kept as the `--design` flag name, since renamed for self-consistency).
      Step 0 added to the 5 scopes that read design context.
- [x] **C3-adjacent** `infra-spine.md` -> `infrastructure.md` (same bug class,
      4 occurrences); `ci.md` mislabelled heading corrected; the hardcoded
      `~/.claude/templates/{preset}/` machine path in `project.md` replaced with a
      project-local resolution that can actually fail.
- [x] **C7** all 7 role references resolve (`technical`->`dev`, `architect`->`sa`,
      `product`->`po`); 5 provider-config reads are now honest optional overrides
      at a path the user can create; 7 `brandkit` refs use the installed path.
- [x] **C4** `dimension-mapping.instructions.md` is the only routing table.
      ~225 duplicated/stale routing lines removed across 12 files, including
      three command blocks in `universal-orchestrator` that referenced
      non-existent selectors (`--backend`, `--design-quality`, `--bug`).
      A Command Coverage table now accounts for every shipped sub-command.
- [x] **C5** one phase model. Canonical is 9 phases; `/start`'s Constitution and
      Ambiguity Resolution are sub-steps 1.0/1.1 of Phase 1; Phase 3 is an
      automatic gate; `/continue`, `/loop`, `/create`, `/improve` are out of
      phase. `loop.md`'s local ladder renamed to Iteration Step 0-4. All printed
      totals now read `/9`.
- [x] **`operations` dimension** (owner decision): `assess-operations.instructions.md`
      plus all six `prompts/*/operations.md` scope files. The `⚠️ PARTIAL` marker
      and its interim fallback table in `dimension-mapping.instructions.md` are
      removed — all nine dimensions route to their own file in every family.
- [ ] Validator (H3) — **deferred at owner request**; to be done separately.

### Round 2 — extraction, consolidation, docs

- [x] **Uncertainty Surfacing extracted** — 8 copies (21-22 lines each) ->
      `instructions/uncertainty-surfacing.instructions.md`, registered in
      `scope_load`. Each dimension keeps only its own confidence basis + one
      worked example. `assess-operations` had no copy and gained one.
      **Two files were carrying the wrong dimension's criteria**: `assess-design`
      and `assess-legal` both held the *technical* confidence basis verbatim
      ("actual code review + test results + metrics"). Both rewritten.
- [x] **Context7 protocol extracted** — `instructions/context7-protocol.instructions.md`
      (when mandatory, tool sequence, citation form, 5-rung fallback chain, what
      it is not for). `implement.md` and `fix.md` restated the sequence; both now
      point at it. `assess.md` Mode 1 too.
- [x] **OWASP taxonomy conflict resolved** — `security-and-owasp` used OWASP
      2021 labels; `security-audit` used a private A1-A10 scheme where numbers
      meant different things (`A1` = Broken Authentication = 2021's `A07`;
      `A2` = Broken Authorization = 2021's `A01`). Reading a finding against the
      wrong file checked the wrong category. Audit file renumbered to 2021, and
      **A02 Cryptographic Failures — which the private scheme had no slot for,
      so nothing audited it — now has a checklist**. Both files state the taxonomy
      and their coverage split.
- [x] Security/resilience consolidation — agent measured first and reported
      **no duplication** in the resilience/ops cluster (max pairwise Jaccard
      0.16; bulkhead absent corpus-wide). Real overlap was OWASP restatement
      between the two security files: audit checklist 45 -> 25 items, each
      removal verified present in its owner first. **Total lines went UP 2.6%** —
      this was a provenance win, not a size win.
- [x] **H8 registry** — `dimension-mapping.instructions.md` now carries the
      canonical Assessment Report Section Registry (4 required sections, 3
      permitted appends, no-synonyms rule). All 9 assess files point at it.
- [x] **H9** — `constitution.template.md` had two variables with contradictory
      defaults in two places (`PERFORMANCE_TARGET`; and `DOC_BAR` meaning two
      different things). Fixed, `CLAIM_BAR` split out, §2 declared the
      definition site. `templates/README.md` had been *endorsing* `[UPPER_SNAKE]`
      / `{lower_snake}` — the exact forms the validators cannot see. Now mandates
      `{{ }}`, legacy marked deprecated.
- [x] **Docs** — `docs/metadata.json` regenerated from the tree (it had 5
      `filesRead` paths that no longer exist and counts off by 1-4; it has **no
      generator**, now says so and ships a `countsHowToVerify` block).
      `docs/overview.md` said 18 platforms while listing 20; `docs/components.md`
      said 17 packages while listing 16.
- [x] **`stack detect --install`** — detection only ever printed, so detected
      instructions never reached `.plaesy/`, closing C8's real gap. 3 new
      per-file tests: installs matches, never overwrites, plain mode writes
      nothing.

### Corrections to this task file's own estimates

- [x] "Extract the 8x shared scaffold (~1,400 lines) into assess-core" —
      **wrong, not done.** Measured: 8 lines are byte-identical across the 9
      assess-* files; 103 of 111 sections are unique to one file. The scaffold
      was already extracted. Replaced with the two extractions that were real.
- [x] "Collapse the duplicated 'Why This Phase Matters' (55 lines)" — **wrong.**
      The 4 copies are command-specific, not boilerplate. Only real defect was an
      unsourced statistic ("30% faster delivery") in `start.md`, now removed.

### In flight

- [x] C8 instruction index + H7 python/vue — **H7 was a false finding**: both
      `python.instructions.md` and `vue.instructions.md` already exist with real
      content and both are registered in `mapping.json`. C8's index half was
      also already done (`.plaesy/instructions.md` exists); its remaining half
      was the dangling `overview.md` link in `memory.md`, now fixed in the
      template, the generated file, and `plaesy.instructions.md`.
- [x] H1/H2/H4/H5/H6 agents, checklists and boundaries — 33 files rewritten.
      H4 needed more than the brief's 5: **`data-engineer`, `qa`, and `tw` were
      also unwired**. All 7 now reachable; `bo`/`po` added to the unconditional
      technical/financial/product delegation cells, the 5 conditional technical
      roles gated on a stated trigger, `pe` wired to `/improve` spec/artifact
      mode.
- [x] 54 dimension sub-prompts dedupe — 54 files, 782 -> 791 lines. The win is
      functional (54 `subagent: true` + 54 `**Loads**:` lines added), not size.
- [x] H10 — `git mv software-design-prinsiples -> software-design-principles`
      and BOTH mapping entries restored (`scope_load` row + `cross_cutting`
      block). The 13 weeks of "unmapped SOLID guidance" note in
      `assess-technical.instructions.md` is gone; `go build` clean.

### Reconciliation 2026-09-27 — measured, not remembered

A DRY sweep was run to get a real number for the remaining work instead of
trusting the estimates above. Getting that number took three attempts, each of
which was wrong in a way that would have been reported as fact:

| Version | Reported | Why it was wrong |
|---|---|---|
| v1 | **101.9%** of the corpus | Ran `SequenceMatcher(a=lines, b=lines)`, comparing every file to itself, so each file matched itself 100%. |
| v2 | 11.1% (5,321 lines) | Real cross-file figure, but attributed a region's redundancy to *every* file containing it, so a 7-line block in 9 files counted 9x. |
| v3 | 8.0% (1,752 lines) | prompts/ + instructions/ only; same per-file attribution bug. |
| **v4** | **3.8% (948 lines)** | Each region counted once via union-find over shared windows. **This is the figure.** |

The known-answer test from this session's own history
(`docs/reference.md:377` ↔ `docs/scripts/platforms.md:11`) was carried through
every version and passed in all of them — including v1, whose cross-file code
was fine and whose self-match code was not. A passing calibration check
validates the code it exercises and nothing else.

**Measured remaining duplication: 948 redundant lines across 65 regions in 34
of 178 files (prompts, instructions, agents, checklists).**

| Family | Redundant lines | Share |
|---|---|---|
| `instructions/` | 661 | 70% |
| `prompts/` | 227 | 24% |
| `checklists/` | 60 | 6% |

Root cause is one thing: the nine `assess-*.instructions.md` files each carry
the same 6-8 line block verbatim — 9-10 copies of it. The largest is the
"**Report sections**: use the canonical names from
`.plaesy/instructions/dimension-mapping.md`" pointer, present 9-10 times. That
path is correct; `.plaesy/instructions/dimension-mapping.md` exists among 39
installed files. This is the defect the corpus report named a round earlier and
never fixed: *a pointer that is long enough is still a duplicate.*

`templates/` is **excluded** from this figure on purpose. v2 measured 25 copies
of a 6-line document header (`Version 0.1` / `Prepared by {{AUTHOR}}` /
`{{ORGANIZATION}}` / `{{DATE_CREATED}}`) across 25 templates and scored it as
144 redundant lines each. That repetition is correct — a template is copied
into a project as a standalone deliverable, so a header held in a shared
include would leave the generated document without it. Literal repetition and
DRY violation are different things.

This also **confirms the correction above** that killed the "~1,400 line
scaffold" claim: the assess-* files are 185-307 lines each and share only
6-8 line blocks. The scaffold was already extracted. `assess-core` is not
needed; the residual is pointer replication, and it is ~661 lines across the
whole `instructions/` family, not 1,400.

### Iteration 3 (2026-09-27) — `/loop`, paused

Collapsed the two multi-line pointers that all nine `assess-*.instructions.md`
files carried verbatim into one-line, scope-specific pointers. 9/9 uniform,
36 lines removed, 26 corpus guards + the markdown ratchet green under
`-count=1`.

**Line-count duplication is not the metric for this corpus.** The sweep moved
948 → 936 after a fix that demonstrably removed 36 lines. The reason is that
the remaining "duplication" is overwhelmingly *structural*, not prose:

- `## Success Criteria` / `## Pre-Assessment Validation` /
  `## Uncertainty Surfacing` — 9 copies each, and each assess file is **required**
  to carry its own. Collapsing them would delete something required.
- The shortened pointers are still 9 copies, just 1-2 lines instead of 5.

So the residual ~936 is mostly headings that must stay. Continuing to optimise
against this number means either churning on required structure or stopping —
which is the protocol's own infinite-loop signal ("quality score doesn't
improve", "same fix applied twice").

**Deliberately not done: the sweep was not promoted to a CI guard.** It has now
produced four wrong numbers this session (101.9% self-match, 11.1% per-file
attribution, 8.0% same bug narrowed, 948/936 dominated by headings). Encoding a
check is right; encoding *this* check is not, and its known-answer test passing
in v1 did not protect it. A DRY guard is worth building only after a metric is
agreed that counts prose and excludes structure — that is a design decision,
not an auto-fix.

Left for a human decision, unchanged:

- **H3** validator — deferred at owner request.
- **`prompts/create/api.md`** placement — no routing table in the repo defines
  the authoritative spelling, so moving it would break a router that cannot be
  inspected.
- **`docs/reference.md` ↔ `docs/scripts/platforms.md`** 7-line table — accepted
  duplication with a known *unguarded* drift risk.

### Iteration 4 (2026-09-27) — `/loop` set to run until stopped

Owner directive: the loop runs continuously until a manual stop. Exhaustion and
completion are no longer stop conditions.

`prompts/loop.md` Phase 4 rewritten. `max_iterations` is now a progress readout
that resets rather than a budget. "All assessable issues fixed", "quality target
reached" and "remaining items need a user decision" all became **continue-and-widen**
instead of stop. The only agent-side stop left is
`consecutive_failures >= max_consecutive_failures` — a real failure, not a
stall; the human stop is ending the session, since there is no `/stop` command
for the agent to observe.
Infinite-loop detection was re-scoped from "pause" to "change axis and continue",
because a stall is a signal that the *axis* is exhausted, not that the work is.

Two new sections: *When the Checklist Is Empty* (widen in a defined order —
re-run mechanical axes, mutation-test the guards, read what no sweep can reach,
compare against external best practice, audit the tooling that judges the work)
and *Plateaus and Stale Metrics* (re-measure from the tree, change axis, discard
a metric that cannot discriminate). Both encode the lesson of this session: a
check that has never been seen red is an assumption, and a metric frozen from an
uncalibrated measurement just freezes the error.

Also fixed: `checkpoint_interval: 0` is no longer a modulo of zero.

### Iteration 4 — needs-review items closed with data

- **`prompts/create/api.md` -> `prompts/create/templates/api.md`.** All four
  `/create:templates:*` sub-commands are documented with the segment, and the
  segment is a directory in the path. `api.md` was the only one outside it, in
  the source tree *and* the `.kilo` mirror, which is what made the wrong layout
  look deliberate. Moved, plus a dead link in `docs/prompts/README.md:41` fixed.
  `TestNestedSubCommandNamesResolveToFiles` now binds every three-segment
  selector to a real file (4 selectors), mutation-tested.
- **`docs/reference.md` <-> `docs/scripts/platforms.md`.** Duplication kept — a
  reference index has to be self-contained — and the drift risk closed with
  `TestDuplicatedDocTablesStayInSync`, mutation-tested. Writing it surfaced a bug
  in the guard itself: `reference.md` carries an earlier `| Subcommand |` table
  for the `plaesy config` verbs, so matching on the header compared the wrong
  pair and reported a length mismatch that did not exist. It now locates the
  table by a marker row.

### Iteration 8 (2026-09-28) — `/loop`, owner decision applied: the tracked binary

Owner: "delete the plaesy binary, and do not track it." Done, plus the second
thing that decision exposed.

**`plaesy` untracked and deleted** (blob `10a8d880`, 14,911,488 bytes). Tracked
files: **659 -> 657**.

The defect was not the missing ignore rule — `/plaesy` has been in `.gitignore`
the whole time — nor the `git add -Af` in `clean`, which was fixed and
documented earlier. It was that **an ignore rule does not apply to a file that is
already tracked**. Two correct measures, neither of which could ever have
worked, and a binary that sat there through both. "We wrote the ignore rule" and
"the ignore rule is doing something" are different claims.

**A second instance of the same shape, found by the guard written to prevent the
first.** `.plaesy/templates/integration-examples.template.md` was tracked but
absent from the working tree: a template correctly relocated to
`docs/examples/integration-examples.md` (finding T8) whose original was never
removed from the index. Every clone was materialising a file that the templates
set had deliberately dropped. Untracked.

**New guards** (`scripts/cmd/plaesy/repo_hygiene_test.go`):

- `TestNoBuildArtifactIsTracked` — by construction, not by a list of forbidden
  names: any tracked file containing a NUL byte in its first 8000 bytes fails,
  and so does any tracked file missing from the working tree. 657 files checked.
- `TestBuildOutputPathsAreIgnored` — untracked is not sufficient, because
  `make reset` runs `git add -A`; the next staging re-adds it. Covers both build
  spellings, since `make build` writes `plaesy` and a bare `go build` in the CLI
  source writes `plaesy.exe` on Windows. `.gitignore` gained the `.exe` rules.

Both mutation-tested: force-adding a freshly built binary turns the first red
with the byte count; deleting one ignore line turns the second red. Fixing the
second red also corrected the first guard's message, which had credited
`make build` with writing `plaesy.exe` — it does not.

Both guards then immediately caught `docs/metadata.json`
(`goTestFiles` 79 -> 80, `goTestFunctions` 726 -> 728).

### Iteration 7 (2026-09-27) — `/loop`, read-and-judge axis

The guards are all proven. The next axis is "what no sweep can reach": files a
script cannot judge because nothing in the build reads them. The candidate was
`docs/instructions/README.md` — a **second hand-maintained description of the
same install set** the index already describes.

It had drifted on every claim, and no guard saw it:

| Claim in the README | README said | Reality |
|---|---|---|
| installed of total | 30 of 71 | **33 of 73** |
| `always_load` count and membership | 9, including `date-system` | **8**, `date-system` is `scope_load` |
| `scope_load` count | 21 | **25** |
| files a scoped command cannot load | 41 | **40** |
| "index-size" claim | 30 framework/scope files | **33** |

The `date-system` one is not a number, it is a behavioural claim: the table
marked it "Always Load: YES" long after a slimming pass moved it to
`scope_load`, so an agent reading that table would rely on a file that is copied
into the project but never loaded into its context.

**The structural finding is more useful than the correction.** One
hand-maintained file describing a fact was fully guarded; a second, adjacent
file describing the same fact had no guard at all. The enforced one being
*correct* is exactly what kept the gap invisible — nothing fails, the numbers
just disagree with each other.

Corrected, and `TestInstructionsReadmeAgreesWithMapping` added so the second
description cannot drift unobserved again. Mutation-tested on all three of the
classes that were actually wrong:

- `always_load` count 8 -> 9 -> red, names the number and the source of truth;
- `date-system` back to "Always Load: YES" -> red, with the consequence in the
  message;
- deleting the sentence entirely -> red twice, *"makes no 'installed count'
  claim, so this test cannot verify it"*. A claim that disappears must fail, or
  the guard silently stops guarding.

Also renamed the section it corrects: "Shared Protocols & System Instructions
(**Auto-Loaded**)" held a file that is not auto-loaded, which is the heading
restating the same wrong claim in a place the numbers did not reach.

### Iteration 6 (2026-09-27) — `/loop`, guard-bite axis closed

Finished the axis opened in iteration 5: prove every guard has been seen red.

**Two guards were still unproven, and both had been given invalid mutations —
not both were blind guards.**

- `TestInstructionsIndexAlwaysLoadedSectionIsExact` — the mutation was
  `sed 's/...Always Loaded (8)$/.../'`, and `$` never matched because the file
  is CRLF, so the file was never changed and the test correctly stayed green.
  A byte-level replace of `(8)` -> `(9)` turns it red: *"the Always Loaded
  heading claims 9 but the section lists 8 file(s)"*.
- `TestEveryShippedCommandHasAChangelogLine` — the mutation renamed a command
  that appears **twice** in the Unreleased section, so the second occurrence
  still satisfied the assertion. `plaesy upgrade` appears exactly once;
  renaming that one turns it red.

The lesson is the one this project has now paid for three times: a green
mutation result is only evidence about the mutation, never about the guard.

**A real red build, found by the finished axis.** `go test ./...` failed on
`cmd/plaesy` and (as a cascade — coverage cannot be measured when a package's
tests fail) on `internal/quality`. Cause: **`docs/CHANGELOG.md`**, an untracked,
unreferenced file byte-identical to `.plaesy/instructions.md` — an instructions
index filed under a changelog name, carrying 60+ relative links that resolve to
nothing. Deleted.

**New guard — `TestNoFileDuplicatesAnotherOutsideTheKnownMirrors`.**
`TestPromptMirrorMatchesSource` compares a known pair in one direction. Neither
it nor the dead-link guard asks whether a file is *allowed* to exist twice, so
the stray copy was found only by accident, as sixty link failures with a common
cause nobody looked for; a stray copy containing no relative links would have
been invisible to the whole package. The new test hashes the 677 text files and
requires every byte-identical group to be one of the repo's 207 known
source/mirror pairs across six families. Mutation-tested twice:

- the exact `docs/CHANGELOG.md` copy -> red, naming both files;
- a copy planted in `docs/instructions/` -> red, and the report names the
  legitimate family *and* the intruder in one line, which is the diagnostic the
  dead-link test could not give.

It also caught its own author: adding one `Test` function moved
`docs/metadata.json` `counts.goTestFunctions` to 725, which
`TestMetadataCountsMatchTheTree` reported immediately.

### Iteration 5 (2026-09-27) — `/loop`, guard-bite axis opened

### Notes

- Markdown ratchet is the guardrail for this work: ceiling 1417, currently 1415.
  Deleting duplicated prose is what buys the headroom; new prose must be clean.
- The user chose to leave `.kilo/` and `.plaesy/` tracked in git (C6 closed as
  "not a defect — re-init regenerates them"), so the SOURCE/GENERATED contract
  stays unenforced by tooling rather than by a CI check. Superseded in one
  respect by iteration 8: `.plaesy/` may stay tracked, but a *stale* file
  under it no longer can.
- **`plaesy`, a 14.9 MB Windows PE binary, was tracked at the repo root.**
  **RESOLVED in iteration 8 (2026-09-28)** — untracked and deleted, guarded by
  `TestNoBuildArtifactIsTracked`.
