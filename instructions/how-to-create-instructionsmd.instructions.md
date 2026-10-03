---
applyTo: '**/*.instructions.md'
description: 'Instructions for writing a *.instructions.md scoped instructions file for AI coding agents, platform-agnostic.'
---

# Scoped Instructions File Creation Instructions

Create or update a `NAME.instructions.md` file — a **scoped** instructions file: a
piece of guidance for an AI agent that applies only to a matching subset of files
(by glob pattern), rather than the whole repository. Several such files can live
side by side, each narrow and non-conflicting, instead of one giant file trying to
cover every concern.

Unlike a single repo-wide instructions file (e.g. `AGENTS.md` — see
[[how-to-create-agentsmd]]), a `.instructions.md` file targets a subset of files via
glob matching. The concepts below (frontmatter with a scope glob + description,
precedence rules, content principles) are common across the AI coding tools that
support this pattern — write the file against these general concepts rather than
against any one tool's exact field names, since tools vary in naming (some use
`applyTo`, others `paths`) but converge on the same underlying idea.

Sources consulted (fetched directly, not taken from this repo's own instruction set),
retrieved 2026-09-26 — cited for the general concepts distilled below, not for
platform-specific field names:
[GitHub Docs — Add repository instructions in your IDE](https://docs.github.com/en/copilot/how-tos/configure-custom-instructions-in-your-ide/add-repository-instructions-in-your-ide),
[GitHub Docs — Response customization concepts](https://docs.github.com/en/copilot/concepts/response-customization),
[GitHub Blog — 5 tips for writing better custom instructions](https://github.blog/ai-and-ml/github-copilot/5-tips-for-writing-better-custom-instructions-for-copilot/),
[VS Code Docs — Custom instructions (raw source)](https://github.com/microsoft/vscode-docs/blob/main/docs/agent-customization/custom-instructions.md),
[awesome-copilot — Defining custom instructions](https://awesome-copilot.github.com/learning-hub/defining-custom-instructions/),
[awesome-copilot — acreadiness-generate-instructions skill][ref30],
[Agent Skills authoring best practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices) (conciseness/progressive-disclosure/degrees-of-freedom
principles transfer directly, independent of the skills-specific mechanics).

## File location and naming

- Filename must end in `.instructions.md`. The part before it names the instruction's
  purpose (e.g. `angular.instructions.md`, `security-and-owasp.instructions.md`), and
  should use kebab-case.
- Store scoped instruction files together in one predictable place in the repo (this
  repository keeps them under `instructions/`, indexed in `instructions/mapping.json`
  via `keywords`/`triggers`/`extensions`) rather than scattering them. Whatever tool
  reads them, it discovers them from a known folder — keep that folder the single
  source, and treat any tool-specific mirroring as generated from it, not hand-edited
  separately.
- When more than one scoped file could apply to the same target file (overlapping
  globs), resolve it by **specificity**: the file whose glob is the narrowest match
  for that path should be treated as taking precedence, and a repo-wide instructions
  file is combined with (not overridden by) any matching scoped file.
- You can have many `.instructions.md` files side by side. Keep each one scoped to one
  concern (one language, one framework, one workflow) rather than one giant file.

## Choosing a strategy: flat vs. nested

Before writing content, decide how many files this concern needs:

- **Flat** — a single instructions file. Fine for a small/medium repo with one
  dominant stack.
- **Nested** — one repo-wide "hub" file plus several topic- or area-scoped
  `.instructions.md` files, each with its own scope glob. Better for large repos,
  monorepos with multiple stacks, or when a single file would exceed a comfortable
  length (see Length below).

For a monorepo, scope one `.instructions.md` per detected area/package rather than one
repo-wide file trying to cover every stack's conventions at once.

## Required and optional frontmatter

```markdown
---
applyTo: 'src/**/*.ts,src/**/*.tsx'
description: 'One sentence: what this file governs and why it exists.'
---
```

Field names vary slightly by tool — use whatever field name the target tool expects;
the concepts are the same everywhere:

- **Scope glob** (the field that makes a file *scoped* rather than global — commonly
  `applyTo`, sometimes an array field like `paths`) — a glob, or comma/array list of
  globs, matching the files this instruction applies to.
  - `*` — all files in the current directory only.
  - `**` or `**/*` — all files in all directories (effectively "always apply").
  - `*.py` — `.py` files in the current directory only.
  - `src/*.py` — matches `src/foo.py`, **not** `src/foo/bar.py`.
  - `src/**/*.py` — matches recursively: `src/foo.py`, `src/foo/bar.py`, `src/foo/bar/baz.py`.
  - `**/subdir/**/*.py` — matches `*.py` under any `subdir` at any depth.
- **Description** — short summary used for discovery/relevance matching (some tools
  can attach a file based on this even without a scope-glob match). Write it as one
  sentence stating what it covers and when it's relevant, not a title restatement.
- **Display name** (optional, where supported) — a human-readable label distinct from
  the filename.
- **Exclusion controls** (optional, where supported) — some tools let a file opt out
  of specific agent surfaces (e.g. automated code review vs. an interactive session)
  even when the scope glob would otherwise match; use this when a rule is only safe
  or relevant in one mode.
- Omitting both the scope glob and description typically means the file is never
  attached automatically — only by explicit manual reference. Prefer narrowing the
  scope over leaving it broad or empty — an overly broad scope makes the agent apply
  the wrong rules to unrelated files, which is the exact problem scoped instructions
  exist to solve.
- Known limitation across most tools: scoped instructions apply to chat/agent
  interactions, not to raw inline "as you type" autocomplete suggestions.

## Body content rules

Write the body as direct imperative instructions to the agent, not a description of
the topic. A well-structured instruction file has four parts: title/overview,
specific guidelines, code examples, and the reasoning ("why") behind non-obvious rules
— reasoning matters because it lets the agent generalize correctly to edge cases you
didn't enumerate.

- State conventions and constraints as rules ("Use X", "Never do Y"), not narrated
  background ("This project uses X").
- **Default assumption: the model is already competent.** Only include what it would
  otherwise get wrong or have to guess — exact commands, non-default conventions,
  gotchas, required ordering. For each sentence, ask "does this justify its token
  cost?" A concise rule with one short code example beats a paragraph explaining
  concepts the model already knows.
- Prefer negative framing for hard prohibitions ("Never mock the DB in integration
  tests") over vague positive guidance ("write good tests") — explicit prohibitions
  are followed more reliably than open-ended advice.
- Match specificity to how fragile the task is (its "degrees of freedom"):
  - **High freedom** — heuristics/text guidance, when multiple valid approaches exist
    and context should decide (e.g. code review checklists).
  - **Medium freedom** — a preferred pattern or parameterized template, when some
    variation is acceptable.
  - **Low freedom** — an exact command/script with no room to deviate, when the
    operation is fragile or must run in a specific sequence (e.g. "run exactly
    `python scripts/migrate.py --verify --backup`, do not add flags").
- Use consistent terminology throughout — pick one term per concept ("API endpoint",
  not a mix of "endpoint"/"route"/"URL") so the agent isn't left resolving synonyms.
- Avoid time-sensitive phrasing ("before/after <date>, use X"). Prefer a "Current
  method" section plus a collapsed/clearly-labeled "Old patterns" section for
  deprecated approaches, so the file doesn't silently go stale.
- Add short code/config examples for anything with easy-to-miss syntax; input/output
  example pairs work better than descriptions when output *style* matters (e.g.
  commit message format).
- Exclude generic language/framework advice the model already knows; exclude padding,
  marketing language, or restating the filename in prose.
- If the same rule belongs in more than one instructions file, put it in the more
  general one and cross-reference (`[[other-file]]`, a plain link, or an import
  mechanism if the target tool supports one) rather than duplicating it.

### Length and progressive disclosure

- Keep a single `.instructions.md` body tight — a comfortable ceiling is roughly
  300–500 lines; several tools also enforce a hard ceiling around 1000 lines. Past
  that, split by concern (see Flat vs. nested above) rather than letting one file
  grow indefinitely.
- If a concern genuinely needs more reference material than fits comfortably, link out
  to a separate reference file **one level deep** from this file (don't nest a
  reference file's own links three deep — an agent skimming a long chain may not read
  the last file in full). Put a table of contents at the top of any linked reference
  file longer than ~100 lines.

Prefer short sections and bullets over long prose. If the concern is simple, keep the
file short — a five-line `.instructions.md` that is followed correctly beats a
fifty-line one that gets skimmed.

## Common failure modes to avoid

- **Scope too broad**: a glob like `**/*.md` on a file meant only for API docs will
  also fire on changelogs and READMEs. Scope it to the actual target directory/extension.
- **Scope too narrow or wrong syntax**: forgetting `**/` drops nested files silently;
  test the glob mentally against a few real paths in the repo before committing it.
- **Duplicating the repo-wide instructions file**: if a rule is truly universal and
  has no narrower audience, it belongs in the repo-wide file, not a new scoped one.
- **Mixing multiple unrelated concerns in one file**: split by concern so each file's
  scope can stay tight.
- **Contradicting another scoped instructions file**: when a repo has several
  `.instructions.md` files, check that a new rule doesn't conflict with an existing
  one whose scope overlaps — overlaps are resolved by specificity/proximity (see
  above), not by which file happens to be "right" in the abstract.
- **Writing instructions that fight the tool's own defaults**: requests requiring the
  agent to fetch external references, tone/style/persona directives, or fixed
  response-length constraints tend to be unreliable across tools and are not a good
  fit for this file type — put persona/tone guidance elsewhere if a tool supports it
  separately.

## Writing process (don't skip straight to prose)

1. Start from what actually goes wrong today — if you're generating this file for an
   existing repo, base content on the repo's actual detected stack/conventions rather
   than a generic template. Where possible, notice 2-3 concrete instances of an agent
   guessing wrong before writing the rule that would have prevented it, rather than
   pre-writing rules for hypothetical problems.
2. Start minimal: 3-5 core instructions (naming, architecture boundary, one security
   rule) beat an exhaustive first draft. Add more iteratively as real gaps surface.
3. Preview before committing: if the repo has an automated instructions-generation
   flow, run it in dry-run mode first, and land the result via a normal PR (not an
   unreviewed direct commit, and never a non-interactive CI-only step) so a human
   reviews the generated rules once before they start steering the agent.
4. Treat the file as living documentation — revisit it when conventions change, not
   only when it's first created.

## Usage Example

```markdown
---
applyTo: '**/*.test.ts,**/*.spec.ts'
description: 'Conventions for writing TypeScript test files in this repo.'
---

Write tests using `describe`/`it`, not `test()`. One assertion focus per `it` block.

- Mock network calls with `msw`, never `jest.mock('axios')` directly — the repo's
  request layer expects handlers, not mocked module internals.
- Test files live next to the file they cover (`foo.ts` -> `foo.test.ts`), not in a
  parallel `__tests__/` tree.
```

## Verification before finishing

- Re-read the scope glob against 2–3 real file paths in the repo and confirm it
  matches only the intended set.
- Check the new file's scope against existing `.instructions.md` files for
  overlapping scope and contradictory rules.
- If this repo has an instructions index (e.g. `instructions/mapping.json`), add an
  entry for the new file so it is discoverable the same way existing entries are.

[ref30]: https://raw.githubusercontent.com/github/awesome-copilot/6c4d33b9cfca967a28bb2962ef4d55e4a384c88c/skills/acreadiness-generate-instructions/SKILL.md
