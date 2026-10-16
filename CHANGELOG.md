# Changelog

All notable changes to Plaesy Constitution Kit are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project uses [Semantic Versioning](https://semver.org/) once it reaches 1.0.0.

## [Unreleased]

### Added

- `THIRD_PARTY_NOTICES.md` credits the taste-skill (MIT) and anthropics/skills `frontend-design` (Apache-2.0) text adapted into the instructions
  below.
- `frontend-design`: an idea test (grounding sentence with a competitor check), a fullness rule against thin pages, and a "Verify with renders"
  procedure (1440, 768 and 390 px, fail conditions, three-round cap, explicit "not run" when no browser is available).
- `plaesy validate design <path...>`: offline design lint for produced frontend output. Checks placeholder image hosts (FAIL), text contrast at
  WCAG 2.2 AA with tokens resolved (FAIL), render evidence at 1440, 768 and 390 px with `--evidence` (FAIL), and placeholder-looking figures and
  default fonts (WARN, failing only with `--strict`).
- `plaesy validate evals` and `make evals`: behavioural cases for `frontend-design` and `redesign` under `evals/`. Static phrase checks and
  fixture lint run offline in `go test`. Brief cases need an LLM through `PLAESY_EVAL_RUNNER`, and are reported as SKIP without it.
- `redesign`: a contrast rule (WCAG 2.2 AA on every changed colour pair).

### Changed

- Accessibility standard is WCAG 2.2 AA (was WCAG 2.1 AA) across the instructions, agents, templates, prompts and docs. Text contrast is
  4.5:1 for normal text and 3:1 for large text and UI components (SC 1.4.3, 1.4.11). Updated: `assess-design`, `assess-technical`,
  `security-audit`, `quality-gates`, the `accessibility` agent, the `application` and `constitution` templates, `prompts/assess`, `fix`,
  `optimize`, `docs/instructions/README`, and the constitution test fixture.
- `.plaesy/memory/constitution.md` still names WCAG 2.1 AA. Its amendment to 2.2 AA is pending human approval and is not applied in this change.
- `redesign`: placeholder-looking figures and invented brand names are replaced by rules to use real or clearly marked placeholder content;
  `picsum.photos` is replaced by licence-checked imagery with credits; monospace is limited to code-like data; surface effects (glass, grain,
  tinted shadows) are opt-in.
- `quality-gates`: Gate 7 points to the measurable layout procedure in `frontend-design`.
- `redesign` and `brandkit` now carry an "Adapted from" line.

## [0.0.1] - Initial release

- Initial public commit of the Plaesy Spec-Kit framework: prompts, instructions,
  agents/roles, checklists, templates, and bash/PowerShell automation scripts.
