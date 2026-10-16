# Changelog

All notable changes to Plaesy Constitution Kit are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project uses [Semantic Versioning](https://semver.org/) once it reaches 1.0.0.

## [Unreleased]

### Added

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

- Accessibility standard is WCAG 2.2 AA (was WCAG 2.1 AA) across the instructions, agents, templates, prompts and docs. Text contrast is
  4.5:1 for normal text and 3:1 for large text and UI components (SC 1.4.3, 1.4.11). Updated: `assess-design`, `assess-technical`,
  `security-audit`, `quality-gates`, `angular`, `context7-protocol`, `tech-validation`, the `accessibility` agent, the `application` and
  `constitution` templates, `prompts/assess`, `fix`, `optimize`, `docs/instructions/README`, and the constitution test fixture.
- `.plaesy/memory/constitution.md` still names WCAG 2.1 AA. Its amendment to 2.2 AA is pending human approval and is not applied in this change.
- `redesign`: placeholder-looking figures and invented brand names are replaced by rules to use real or clearly marked placeholder content;
  `picsum.photos` is replaced by licence-checked imagery with credits; monospace is limited to code-like data; surface effects (glass, grain,
  tinted shadows) are opt-in.
- `quality-gates`: Gate 7 points to the measurable layout procedure in `frontend-design`.
- `redesign` and `brandkit` now carry an "Adapted from" line.

## [0.0.1] - Initial release

- Initial public commit of the Plaesy Spec-Kit framework: prompts, instructions,
  agents/roles, checklists, templates, and bash/PowerShell automation scripts.
