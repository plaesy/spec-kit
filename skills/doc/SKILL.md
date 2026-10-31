---
name: doc
description: "Generate grounded, cited project documentation (overview, architecture, components, reference, how-to) from source code and specifications. Use when the user asks to document a project, write docs or an API reference, or prepare a project for handoff"
disable-model-invocation: true
---

# `/doc` command instructions

## Purpose

Generate grounded, cited project documentation (overview, architecture, components, reference, how-to) from source code and specifications, with verifiable citations and structured outputs.

⚡ **Run with**: any parallel-capable mode. A single-agent host runs it serially, which is fine. See `/loop` for when fan-out is worth its cost.

## Success Contract

Write `contracts.json` in the task folder before work starts (one `run` and `expected` per item). Run `plaesy tasks validate <task>`. Report completion only when every item is `passed` (NN-08; `.plaesy/instructions/success-contract.md`).

## Objective

Generate grounded documentation from source code and specifications, with verifiable accuracy and full citations.

**Non-software deliverables** (business case, legal set, campaign, roadmap): this prompt is code-centric (`file:line` citations, Diátaxis, OpenAPI). For those dimensions, produce the **Handoff Spec** from the `/assess` Design Spine instead (business case with assumption register, RACI, runbook). Apply the same grounding rules, but cite the deliverable's own sections and sources. The multi-pass, Diátaxis, and OpenAPI steps do not apply.

**Why it matters:** documentation carries the rationale, usage, and architecture into institutional memory while context is fresh. Without it, the knowledge stays in the code and in conversations. Handoffs slow down, onboarding turns into git archaeology, and refactors start regressions.

## Grounding Rules (Mandatory, Read First)

Break any of these and the output is invalid.

1. **Only cite what you read.** Every claim about code references a file you opened this session. If you cannot find it, record the gap. Do not guess.
2. **Verbatim over paraphrase.** Copy signatures and snippets exactly from source. Never rebuild them from memory.
3. **Cite with location.** Every documented entity includes `path/to/file.ts:L42`.
4. **Less is safer.** If context is thin, write fewer sections and list what is missing in `docs/metadata.json.gaps`. Never fill a gap with a plausible invention.
5. **Separate fact from inference.** Mark inferred design rationale `(inferred)`. Rationale taken from `prompts/` or `skills/` cites that file.
6. **Uncertainty is output.** Unresolved questions go in the report's Open Questions.

## Before You Run

- [ ] `contracts.json` written before the first action.
- [ ] Source files exist and can be read.
- [ ] The project type can be detected. It is not unknown or ambiguous.
- [ ] The documentation structure is planned.
- [ ] The audience is clear.

## Input Analysis

1. **Project context.** Scan the directory structure. Read the metadata files (`package.json`, `pyproject.toml`, `go.mod`, `Cargo.toml`, `pubspec.yaml`, `pom.xml`). Load `prompts/` and `skills/`. Check for existing docs. Write in the user's conversation language unless `--language` says otherwise.
2. **Source collection.** Detect languages and frameworks. Find the relevant files (`.ts`, `.js`, `.py`, `.go`, `.rs`, `.dart`, `.java`, `.cs`, and so on). Read the entry points first, then follow the imports outward. Read docstrings, JSDoc, and comments. **Record every file read.** That list is the citation whitelist.
3. **Specifications.** Read the relevant prompts (objectives, phases, protocols), the design instructions, and the checklist criteria.

## Execution (Multi-Pass, Do Not Skip Passes)

### Pass 1: Structure Planning

Write the plan before any document:

1. **Project type:** `api | cli | library | datascience | unknown`, with the evidence and file:line citations.
2. **Document set** (Diátaxis-separated):
    - OVERVIEW: purpose, scope, core features, stack.
    - ARCHITECTURE: high-level design, plus a Mermaid diagram.
    - COMPONENTS: per component, responsibility and interactions.
    - REFERENCE: endpoints, functions, types. Lookup-oriented, dry, factual.
    - HOWTO/GUIDES: task steps. Only where the code shows usage patterns.
3. **Entity inventory:** every public function, class, and endpoint, with file:line.
4. **Gap list:** entities missing docstrings or comments.

Show the Pass 1 summary to the user before generating, unless `--no-confirm` is set.

### Pass 2: Per-Document Generation

| Document | Questions to answer |
|---|---|
| Overview | What problem does it solve? Core features? Target users? Main technologies? |
| Architecture | Main components and roles? Data flow? External dependencies? |
| Components | For each: purpose, interactions, design patterns, implementing file paths |
| Reference | Exact signatures, parameters with types and defaults, return values, errors thrown |
| How-to | Real task steps on actual code paths, with the expected outcome per step |

**Entry style.** Match the GOOD example. Never produce the BAD one.

GOOD: verbatim identifier, behavior, and cited locations.

```markdown
#### `common.NormalizeVersion()`

Sanitizes the plaesy version string.

- **Location**: `internal/common/version.go:17`
- **Build-time value**: `Version` is set via `-ldflags "-X .../common.Version=1.2.3"` (`internal/common/version.go:11`)
- **Fallback**: emits `0.0.0` when no valid semver is found (`internal/common/version.go:19-21`)
- **Caching**: none. `Version` is a package-level var set at build time (`internal/common/version.go:11`)
```

BAD: reconstructed from memory. Never emit this.

```markdown
#### `getVersion()`

Reads the version from package.json and caches it globally for performance.
```

Why it fails: the function name is invented (the real one is `NormalizeVersion`), the source claim is wrong (`package.json` is never read), there are no citations, and the performance claim cannot be checked. Every REFERENCE and COMPONENTS entry needs the GOOD shape: exact identifier, behavior summary, and `file:line` citations.

### Pass 3: Self-Validation

Run the checklists below against the output. Fix every failure before reporting completion.

## Diagrams

Every ARCHITECTURE document needs at least one valid Mermaid diagram:

- Architecture overview: `graph TD` of components and their relationships.
- Data flow: `flowchart LR` for the main request lifecycle (API projects).

Rules: one diagram per ```` ```mermaid ```` fence, with nothing else inside it. Every node must match a real component or file. Keep each diagram to 15 nodes or fewer. Split larger systems.

## Output Contracts

Write under `--output-dir` (default `./docs/`). These are hard length caps. Trim content rather than exceed them.

| Document | Budget |
|---|---|
| `overview.md` | ≤400 words |
| `architecture.md` | ≤600 words of narrative, plus diagrams |
| `components.md` | ≤150 words per component. Above 15 components, split into `components/{name}.md` |
| `reference.md` | No prose cap. One compact cited entry per entity |
| how-to and guides | ≤10 numbered steps each |

**`components.md` above 15 components** becomes an index: one line per component (name, one-sentence responsibility, link). Each full analysis moves to `docs/components/{name}.md`. The flat file never grows past 15 entries. That is the point where single-file component docs become hard to review.

**`reference.md`** is a dry, complete lookup: signatures, parameters, returns, errors. Advice and tutorials go to the how-to guides, linked from here.

**`openapi.json`** (API projects only): valid OpenAPI 3.0, importable into Postman, Insomnia, or Hoppscotch.

**`metadata.json`:**

```json
{
  "generator": "plaesy",
  "generatedAt": "<ISO-8601>",
  "projectType": "api|cli|library|datascience|unknown",
  "languages": [],
  "filesRead": [],
  "counts": { "functions": 0, "types": 0, "endpoints": 0 },
  "coveragePercent": 0,
  "gaps": ["entities missing documentation"],
  "openQuestions": ["uncertainties requiring human input"],
  "inferences": ["claims marked (inferred), with basis"]
}
```

## Quality Validation (Pass 3)

**Grounding (blocking):**

- [ ] Every documented entity exists at its cited path:line.
- [ ] Every quoted signature matches the source verbatim.
- [ ] No invented function names, parameters, or endpoints.
- [ ] `filesRead` covers every cited file.

**Diátaxis:**

- [ ] REFERENCE holds no tutorials or advice.
- [ ] HOWTO holds no concept explanations. Link to ARCHITECTURE instead.
- [ ] No page mixes learning, goal, lookup, and understanding modes.

**Format:**

- [ ] Mermaid blocks are valid.
- [ ] Headings nest without skipped levels.
- [ ] Code fences have language tags.

**OpenAPI (API projects):**

- [ ] Valid OpenAPI 3.0.
- [ ] Every detected endpoint is included.
- [ ] Request and response schemas are present.

## Success Criteria

- [ ] Every Success Contract item is compared with its result at the gate. Blocking items are met.

**Clarity:** specs are clear and documented. No ambiguity remains. Every public API and function is documented.

**Quality:** tests and metrics are validated. No regressions. All grounding checks pass. No fabricated or unverified claims.

**Self-audit:**

- [ ] The docs are handoff-ready.
- [ ] Closing check applied (`.plaesy/instructions/quality-gates.md` → **Closing check**): ready for the next phase, and would this ship?
- [ ] A new team member can onboard from these docs alone.
- [ ] Every gap and open question is recorded in `metadata.json`.

## Troubleshooting

- **No source files found:** confirm they exist, and name the directory explicitly.
- **Context too large:** document the public API and entry points first. Summarize internal modules. Record skipped areas in `metadata.json.gaps`. Never fabricate unread code to fill space.
- **No design rationale:** populate `prompts/` and `skills/`. Otherwise mark the rationale `(inferred)` and state its basis.
- **Code contradicts existing docs or prompts:** trust the code. Record the conflict in `metadata.json.openQuestions`.

## References

- Grounding and multi-pass pattern: DeepWiki (github.com/AsyncFuncAI/deepwiki-open, api/prompts.py).
- Document types: Diátaxis (diataxis.fr).
- Spec-kit specific: rationale extracted from `prompts/` and `skills/`, which generic tools cannot do.
- Hallucination in repository-level code work, arXiv 2409.20550 (taxonomy and retrieval-based mitigation): abstract only, secondary. Supports Grounding Rules 1 and 2.
- Grounding with API documentation, arXiv 2407.09726: not read, secondary. Cited only as a direction to check.
- CodeWiki, arXiv 2510.24428 (repository-level documentation benchmark): abstract only, secondary.

## Autonomous Routing

**See also:** [Universal Autonomous Routing](.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory)

After `/doc`, review `metadata.json.gaps` and `openQuestions`:

- **Gaps** (missing or stubbed implementation, undocumented public API): `/implement` to close them, then re-run `/doc`.
- **Open questions** (architecture, business, or UX uncertainty): `/assess` in Mode 1, scoped to the dimension the question belongs to, then re-run `/doc`. The dimension's research route is in the **Assess** column of `.plaesy/instructions/dimension-mapping.md`. Look it up there. `/doc` takes no scope: it documents whatever the project produced.
- **Both empty**, `coveragePercent` ≥{{TEST_COVERAGE_MIN|90}}, every public API documented with examples, and the deployment, security, and error-handling sections present: the docs are **handoff-ready**. Run `/save` (Phase 9).

Loop `/doc` → `/implement` or `/assess` → `/doc` until both lists are empty. Never run `/save` with open gaps or questions.

---

## Worked Example — Document a Go REST API

**Input:** a Go project with `main.go`, `internal/handler/`, `internal/service/`, and `internal/repository/`. 21 public functions across 3 packages. Run: `/doc --output-dir ./docs`.

```text
Pass 1: Structure Planning
1. Project type: api (main.go:15 chi.Router, internal/handler/*.go)
2. Doc set: OVERVIEW, ARCHITECTURE, COMPONENTS, REFERENCE, HOWTO (3 guides)
3. Entities: 21 public functions, 8 types, 12 endpoints, all with file:line
4. Gaps: 3 handlers lack docstrings (health.go, metrics.go, version.go)

Pass 2: Generation
- overview.md: 320 words (purpose, stack, quick links)
- architecture.md: 480 words, 2 Mermaid diagrams (component graph, request flow)
- components.md: 5 components × ~140 words
- reference.md: 21 entries, verbatim signatures with file:line (e.g. handler.CreateUser, internal/handler/user.go:28)
- howto/deploy.md: 8 steps. howto/test.md: 6 steps.

Pass 3: Self-Validation
✅ All 21 entities exist at the cited path:line
✅ Signatures match source verbatim
✅ filesRead covers all 14 cited files
✅ REFERENCE is dry facts only. HOWTO has no concept explanations.
✅ Mermaid renders. coveragePercent = 86 (3 gaps recorded)

Output: ./docs/{overview,architecture,components,reference}.md, ./docs/howto/{deploy,test}.md,
./docs/openapi.json, ./docs/metadata.json
```

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md` →
[Global Routing](.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory)
