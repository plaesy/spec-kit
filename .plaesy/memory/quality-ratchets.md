---
name: Quality Ratchets
description: Markdown and prompt-craft ratchets, their floors, and the instrument bugs
updatedAt: "2026-09-30T19:40:00.000Z"
---

# Quality Ratchets

## Current floors

- **Markdown is gated by a ratchet, and the ratchet is now zero.** The corpus
  went 1417 violations → 0 on 2026-09-30 across 275 files.
  `.markdownlint-baseline.json` is **deleted**, not emptied: `LoadBaseline`
  treats absence as "no ceiling" and rejects a literal `max_violations: 0` as
  unable to express a ratchet. Per-file limits that matter for prompts: MD013
  line_length 200, MD040 (fence a language or not at all), MD036 (do not fake a
  heading with bold text).
- **`repo_ratchet_test.go` had to change with the deletion.** It treated an
  absent baseline as "markdown would be unchecked", so simply deleting the file
  would have disabled linting entirely while reporting success. Absence now
  means zero tolerance — strictly stronger than the 1417 ceiling and not
  gameable by removing the file.
- **Prompt craft has a measured floor** (`internal/quality/prompt_craft_ratchet_test.go`,
  baseline in `internal/quality/prompt-craft-baseline.json`). 2026-09-30, 10 prompts:
  `with_objective` 9→10, `with_output_contract` 6→10,
  `with_worked_example` 1→10, `## Protocol` restating its own objective 1→0.
  Guard against regressions *and* has been mutation-tested per metric.

## These metrics are structural proxies

- **Structural sweeps are not prompt quality.** Dead links, stale counts and
  unresolved paths are a different axis from whether a prompt works.
  Reporting a sweep as a prompting score is a known blind spot in this project,
  and it recurred: for most of the 100/100 push a markdownlint score was allowed
  to stand in for a prompt-quality claim.
- **The craft metrics cannot judge content.** They check heading presence, body
  length and duplication. They cannot tell you whether a worked example is any
  good or whether an objective is the right one. 10/10 means "no measurable
  regression", not "well-crafted".
- **The 8 new worked examples were written from scratch**, not reused from
  existing fences, grounded in each prompt's own output format. Invented examples
  teach the model a pattern; this is the main thing to review before shipping.
- **A metric that looks one level down is not measuring the thing it claims.**
  The craft ratchet's `headingRe` matched only `^##\s+`, so `optimize.md`'s real
  worked example (under a deeper heading) was scored absent — 9/10 where the
  truth was 10/10. The same blindness applied to the objective and output-contract
  checks. Caught only by measuring independently and getting a different answer
  from the harness. Fixed to `^#{1,6}\s+`.
- **Orphan/reachability rates must not be applied to human-called entry points.**
  See `prompt-authoring.md`.

## Linter bugs found while driving the count to zero

Three defects in the tool, not the corpus. Each is verified to fail without its guard.
- **MD025 counted YAML front matter as headings.** goldmark v1.7.8 (pinned) has
  no front-matter extension; the doc-level worka