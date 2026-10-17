# Changelog

All notable changes to Plaesy Constitution Kit are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project uses [Semantic Versioning](https://semver.org/) once it reaches 1.0.0.

## [Unreleased]

### Added

- **Agent and tool trust boundaries** in `security-design-principles`: repository files, web pages, issue and PR text, tool output and MCP
  descriptions are untrusted data (constitution SC-01 extended to these channels). An action requested only by untrusted content is never
  taken. Repository-defined MCP and hook configuration is reviewed before first use. Pointer added to `plaesy.instructions.md`.
- **Dependency existence check** in `security-and-owasp` §6.2: every new package name is checked against its registry before it enters a
  manifest, with a committed lockfile. Based on Spracklen et al., arXiv:2406.10279 (USENIX Security 2025).
- **Agent secret hygiene** in `git` §7: no printing of secret files into the transcript, named staging instead of `git add -A`, and a secret
  scan of the staged diff before every commit. **Checkpoint before risky agent edits** in `git` §1.
- **Failure modes and controls** table and the simplest-pattern rule in `multi-agent-patterns`, plus the Orchestrator-Workers and
  Evaluator-Optimizer patterns that were missing from its table.
- **Trials and consistency** section in `evals/README.md`: pass@k vs pass^k, grade outcomes not paths, read transcripts, start from real
  failures. Records that brief cases currently run once, with no k-trial runner.
- **Static eval cases** for every rule above under `evals/cases/` (`security-design-principles`, `security-and-owasp`, `git`,
  `multi-agent-patterns`, `plaesy`, and a new `quality-gates` entry).
- **DoD checklist items** in `story-done`: dependency check and lockfile, secret scan, small batches with checkpoints, no duplicated blocks.

- **Success Contract (mandatory for every phase).** Before a phase starts it writes a measurable KPI set and checklist, freezes it, and at
  the end compares every item with its result. Full rule: `instructions/success-contract.instructions.md` (installed as
  `.plaesy/instructions/success-contract.md`, scope-loaded). Template: `templates/success-contract.template.md`. A short pointer is in
  `plaesy.instructions.md`. Gate 0 in `quality-gates` is the precondition and the contract comparison table is part of its output. Every
  phase prompt carries the same pointer block, and the verifier, critic and qa agents and the story checklists point to the rule.
  Contracts are stored under `.plaesy/contracts/`.
- `plaesy validate contract <file> [--gate]`: validates the contract's schema, measurable targets (no vague words; a number with a comparison or a
  binary condition), freeze and amendments (a changed target needs a recorded reason, and blocking cannot be loosened), and with `--gate` the
  recorded results. A `met` numeric claim is checked against its target. Unit tests and fixtures under `evals/fixtures/contracts/`.
- `THIRD_PARTY_NOTICES.md` and `licenses/Apache-2.0-frontend-design.txt` credit the taste-skill (MIT) and anthropics/skills `frontend-design`
  (Apache-2.0) text adapted into the instructions. The Apache text is a byte-for-byte copy of the upstream licence file.
- `frontend-design`: an idea test (grounding sentence with a competitor check), a fullness rule against thin pages, and a "Verify with renders"
  procedure (1440, 768 and 390 px, fail conditions, three-round cap, explicit "not run" when no browser is available).
- `plaesy validate design <path...>`: offline design lint for produced frontend output. FAIL: placeholder image hosts, text contrast below WCAG 2.2
  AA (CSS rules, inline `style=`, and JSX style objects; translucent text composited over its background; px, rem, em, % and clamp() sizes),
  and, with `--evidence`, a render at 1440, 768 and 390 px. WARN, failing only with `--strict`: placeholder-looking figures and default fonts.
  Utility classes and CSS-in-JS are not read and are counted as skipped. The verdict is the same on every run.
- `plaesy validate evals` and `make evals`: case files under `evals/` with three kinds of check. Phrase checks confirm that required wording is
  present in an instruction; they do not check meaning. Fixture checks lint hand-written frontend folders with known expected results; they test
  the lint, not a model. Brief checks are the only test of model output, and they need `PLAESY_EVAL_RUNNER`. Without it the briefs are reported
  SKIP and the output says no model behaviour was checked.
- `redesign`: a contrast rule (WCAG 2.2 AA on every changed colour pair).

### Changed

- Quality gate verifier: findings block only when they break an acceptance criterion, a constitution article or correctness. Other findings
  go in a non-blocking list (`quality-gates`).
- Eval guidance: brief cases are described as single-trial results until a k-trial runner exists (`evals/README.md`).

- Accessibility standard is WCAG 2.2 AA (was WCAG 2.1 AA) across the instructions, agents, templates, prompts and docs. Text contrast is
  4.5:1 for normal text and 3:1 for large text and UI components (SC 1.4.3, 1.4.11). Updated: `assess-design`, `assess-technical`,
  `security-audit`, `quality-gates`, `angular`, `context7-protocol`, `tech-validation`, the `accessibility` agent, the `application` and
  `constitution` templates, `prompts/assess`, `fix`, `optimize`, `docs/instructions/README`, and the constitution test fixture.
- `.plaesy/memory/constitution.md` 2.0.0 → 2.1.0 (MINOR, approved): the Design threshold and accessibility bar name WCAG 2.2 AA, with a
  Sync Impact Report in the amendment log.
- `redesign`: placeholder-looking figures and invented brand names are replaced by rules to use real or clearly marked placeholder content;
  `picsum.photos` is replaced by licence-checked imagery with credits; monospace is limited to code-like data; surface effects (glass, grain,
  tinted shadows) are opt-in.
- `quality-gates`: Gate 7 points to the measurable layout procedure in `frontend-design`.
- `redesign` and `brandkit` now carry an "Adapted from" line.

## [0.0.1] - Initial release

- Initial public commit of the Plaesy Spec-Kit framework: prompts, instructions,
  agents/roles, checklists, templates, and bash/PowerShell automation scripts.
