---
description: "Generate comprehensive project documentation from source code and specifications"
subagent: true
---

# `/doc` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Objective

Generate professional, grounded documentation from project source code and specifications with verifiable accuracy and complete citations.

**Non-software deliverables** (business case, legal doc set, marketing campaign,
product roadmap): this prompt is source-code-centric by design below (file:line
citations, Diátaxis doc types, OpenAPI). For those dimensions, generate the
**Handoff Spec** instead, per `/assess` Design Spine table (business case +
assumption register, RACI matrix, runbook, etc.) — same grounding discipline
(§ Grounding Rules), but citations point at the deliverable document's own
sections/sources instead of `file:line`, and the multi-pass/Diátaxis/OpenAPI
machinery below does not apply.

## Why This Phase Matters

**Purpose**: Comprehensive documentation enables handoff (team velocity + future maintenance)

**Why Now**: Documentation anchors the completed feature into institutional memory—without it, knowledge stays locked in code and conversations. This phase captures rationale, usage patterns, and architecture at the moment when context is freshest.

**What Breaks If Skipped**: 
- Maintenance velocity plummets when future developers lack API contracts and design decisions
- Onboarding new team members requires archaeology through git history instead of clear handbooks
- Architecture decisions become implicit; refactoring introduces regressions

**Success Enables**: 
- Confident handoffs to other teams or contributors
- Self-service API discovery (reduce support burden)
- Foundation for code generation and testing frameworks that rely on accurate specs

## Protocol

**Execute documentation generation by analyzing source code → extracting specifications → producing deliverables (overview, architecture, components, reference, guides) → validating accuracy**

---

## Grounding Rules (MANDATORY — read first)

These rules prevent fabricated documentation. Violating any of them invalidates the output:

1. **Only cite what you read** — every claim about code MUST reference a file that was actually opened in this session. When unsure whether you have read something, re-open the file before citing it; if it cannot be found, record the gap instead of guessing.
2. **Verbatim over paraphrase** — when quoting signatures or snippets, copy them exactly from source; do not reconstruct from memory.
3. **Cite with location** — every documented entity includes `path/to/file.ts:L42` style references.
4. **Less is safer** — if context is insufficient for a section, produce fewer sections and list what is missing in `docs/metadata.json` under `gaps`. NEVER fill gaps with plausible-sounding inventions.
5. **Distinguish fact from inference** — inferred design rationale must be marked `(inferred)`; rationale sourced from prompts/ or instructions/ is cited to that file.
6. **Uncertainty is output too** — unresolved questions go into the report's "Open Questions" section.

---

## Validation Checklist (Before Running)
- ✅ Source files exist and readable
- ✅ Project type detectable (not unknown/ambiguous)
- ✅ Documentation structure planned
- ✅ User audience clear

## Input Analysis

### 1. Project Context Loading

- Scan source directory structure
- Read project metadata files (`package.json`, `pyproject.toml`, `go.mod`, `Cargo.toml`, `pubspec.yaml`, `pom.xml`, etc.)
- Load `prompts/` directory (workflow objectives)
- Load `instructions/` directory (design principles)
- Check for existing documentation
- Detect user language: generate documentation in the user's conversation language unless overridden via `--language`

### 2. Source Code Collection

- Auto-detect project languages and frameworks
- Identify all relevant source files (`.ts`, `.js`, `.py`, `.go`, `.rs`, `.dart`, `.java`, `.cs`, etc.)
- Read entry points first, then follow imports outward (dependency-driven reading order)
- Read source file contents (with JSDoc, docstrings, inline comments)
- Record every file read — this becomes the citation whitelist

### 3. Specification Files

- Read relevant prompts (project objectives, phases, protocols)
- Load design instructions (framework guidance, best practices)
- Extract checklist criteria (for quality validation)

---

## Execution Protocol (multi-pass — do not skip passes)

### Pass 1: Structure Planning

Produce an internal documentation plan BEFORE writing any docs:

```
1. Project type classification: api | cli | library | datascience | unknown
   Evidence: [list concrete indicators + file:line citations]
2. Documentation set selection (Diátaxis-separated):
   - OVERVIEW      → purpose, scope, core features, tech stack
   - ARCHITECTURE  → high-level design + Mermaid diagram
   - COMPONENTS    → per-component responsibility & interactions
   - REFERENCE     → API endpoints / functions / types (lookup-oriented, dry, factual)
   - HOWTO/GUIDES  → task-oriented usage (only if usage patterns are evident in code)
3. Entity inventory: all public functions/classes/endpoints found, with file:line
4. Gap list: entities lacking docstrings/comments
```

Present Pass 1 summary to the user before generating (unless `--no-confirm`).

### Pass 2: Per-Document Generation

Generate each document separately using its type-specific focus:

| Document     | Focus questions                                                                                                      |
| ------------ | -------------------------------------------------------------------------------------------------------------------- |
| Overview     | What problem does this solve? Core features? Target users? Main technologies?                                        |
| Architecture | Main components and roles? Data flow between them? External dependencies?                                            |
| Components   | For each: purpose/responsibility? Interactions with other components? Design patterns used? Implementing file paths? |
| Reference    | Exact signatures? Parameters + types + defaults? Return values? Errors thrown?                                       |
| How-to       | Real task steps grounded in actual code paths? Expected outcomes per step?                                           |

#### Few-shot: entry style (grounded vs fabricated)

GOOD — verbatim identifier + cited location (match this style):

```markdown
#### `get_plaesy_version()`

Resolves the plaesy version string at script load time.

- **Location**: `scripts/bash/common.sh:9`
- **Search order**: checks `./VERSION`, repo root, then `${HOME}/.plaesy/VERSION`; first hit wins (`scripts/bash/common.sh:23-35`)
- **Fallback**: emits `0.0.0` when no valid semver is found (`scripts/bash/common.sh:37-48`)
- **Caching**: result stored once in readonly `PLAESY_VERSION` (`scripts/bash/common.sh:51`)
```

BAD — reconstructed from memory (never emit this):

```markdown
#### `getVersion()`

Reads the version from package.json and caches it globally for performance.
```

Why BAD fails: function name invented (`get_plaesy_version` is the real one), wrong source claim (`package.json` is never read), zero citations, and an unverifiable performance rationale. Every REFERENCE/COMPONENTS entry must match the GOOD pattern: exact identifier + behavior summary + `file:line` citations.

### Pass 3: Self-Validation

Run quality checklists (below) against generated output. Fix failures before reporting completion.

---

## Diagrams (Mermaid)

Every ARCHITECTURE document MUST include at least one valid Mermaid diagram:

- **Architecture overview**: `graph TD` of components and their relationships
- **Data flow**: `flowchart LR` for primary request/data lifecycle (if API project)

Rules:

- Write each diagram in standard Mermaid syntax inside a single ```` ```mermaid ```` fenced block — one fence per diagram, nothing else inside it
- Every node in a diagram must correspond to a real component/file (grounding rule #1 applies)
- Keep diagrams ≤ 15 nodes; split larger systems into multiple diagrams

---

## Output Contracts

All files written under `--output-dir` (default `./docs/`). Length budgets (hard caps — trim content rather than exceed):

| Document        | Length budget                                          |
| --------------- | ------------------------------------------------------ |
| overview.md     | ≤ 400 words                                            |
| architecture.md | ≤ 600 words narrative + diagrams                       |
| components.md   | ≤ 150 words per component; index-only + split into `components/{name}.md` above 15 components (same threshold as the diagram node cap above — don't let one file grow unbounded) |
| reference.md    | No prose cap — one compact cited entry per entity      |
| howto/guides    | ≤ 10 numbered steps per guide                          |

### docs/overview.md

Purpose, scope, core features, tech stack, quick links to other docs.

### docs/architecture.md

High-level design narrative + Mermaid diagrams + data flow + external dependencies.

### docs/components.md

Per-component analysis (responsibility, interactions, patterns, implementing paths).

**At 15 or fewer components**: one flat file, ≤150 words per component, as above.

**Above 15 components**: `docs/components.md` becomes an index only — same
index-links-to-topic-files pattern as `.plaesy/memory.md` — one line per
component (name + one-sentence responsibility + link), full ≤150-word analysis
moved to `docs/components/{component-name}.md`. Never let the flat file grow
past 15 entries; that threshold is the point real design-system docs report
single-file component docs becoming unmanageable to navigate and review.

### docs/reference.md

Dry, factual, complete lookup: endpoints/functions/types with exact signatures, parameters, returns, errors. Route all advice and tutorials to HOWTO/GUIDES instead, linked from here (Diátaxis separation).

### docs/openapi.json *(API projects only)*

Valid OpenAPI 3.0 spec; import-ready for Postman/Insomnia/Hoppscotch.

### docs/metadata.json

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
  "inferences": ["claims marked (inferred) with basis"]
}
```

---

## Quality Validation (Pass 3)

### Grounding Checks (blocking — fix before completion)

- [ ] Every documented entity exists at its cited path:line
- [ ] All quoted signatures match source verbatim
- [ ] No invented function names, parameters, or endpoints
- [ ] `filesRead` in metadata.json covers all cited files

### Diátaxis Separation Checks

- [ ] REFERENCE contains no tutorials/advice (dry facts only)
- [ ] HOWTO contains no concept explanations (link to ARCHITECTURE instead)
- [ ] No single page mixes learning/goal/lookup/understanding modes

### Format Checks

- [ ] Mermaid blocks render-valid syntax
- [ ] Markdown headings hierarchical (no skipped levels)
- [ ] Code examples include language tags on fences

### OpenAPI Checks (if API)

- [ ] Valid OpenAPI 3.0 schema
- [ ] All detected endpoints included
- [ ] Request/response schemas present

---

## Success Criteria (VERIFY Before Completing)

✅ **Clarity**:
- [ ] All specifications clear and documented
- [ ] No ambiguities remain
- [ ] Every public API/function has accompanying documentation

✅ **Quality**:
- [ ] Tests passing/metrics validated
- [ ] No regressions detected
- [ ] All grounding checks (Pass 3) passing
- [ ] No fabricated or unverified claims

✅ **Self-Audit**:
- [ ] Ready for next phase? (Documentation is handoff-ready)
- [ ] Would ship this to production? (Complete, accurate, usable)
- [ ] Can a new team member onboard using only this documentation?
- [ ] Are all gaps and open questions recorded in metadata.json?

---

## Troubleshooting

### Issue: "No source files found"

**Solution**: Ensure source files exist; specify scope via explicit directory in conversation.

### Issue: "Context too large"

**Solution**: Prioritize public API surface and entry points; summarize internal modules; record skipped areas in `metadata.json.gaps`. Never fabricate unread code to fill space.

### Issue: "Missing design rationale"

**Solution**: Populate `prompts/` and `instructions/`; otherwise mark rationale as `(inferred)` with basis stated.

### Issue: "Conflicting information"

**Solution**: When code contradicts existing docs/prompts, trust the code, flag the conflict in `metadata.json.openQuestions`.

---

## References

- Grounding & multi-pass pattern: DeepWiki/deepwiki-open prompt architecture (github.com/AsyncFuncAI/deepwiki-open, api/prompts.py)
- Document type separation: Diátaxis framework (diataxis.fr)
- Spec-kit integration: unique capability — extracts design rationale from prompts/ + instructions/, which generic tools cannot do

---

## Autonomous Routing (After Documentation)

**See also**: [Universal Autonomous Routing](`.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory`)

After `/doc` completes, review `metadata.json.gaps` and `.openQuestions`:

- **Gaps** (missing/stubbed implementation, undocumented public API) → `/implement`
  to close them, then re-run `/doc`
- **Open questions** (architecture/business/UX uncertainty) → `/assess` (Mode 1,
  scoped to the relevant dimension — `technical`/`product`/`design`) to resolve,
  then re-run `/doc`
- **Both empty**, `coveragePercent` ≥90%, all public APIs documented with examples,
  and deployment/security/error-handling sections present → documentation is
  **handoff-ready**: run `/save` to persist the session

Loop `/doc` → `/implement`/`/assess` → `/doc` until both lists are empty; don't
proceed to `/save` with unresolved gaps or open questions.

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md` → [Global Routing](`.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory`)

**Execute this phase with `/doc` command to generate comprehensive project documentation!**
