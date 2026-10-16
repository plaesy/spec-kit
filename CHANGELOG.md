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

### Changed

- `redesign`: placeholder-looking figures and invented brand names are replaced by rules to use real or clearly marked placeholder content;
  `picsum.photos` is replaced by licence-checked imagery with credits; monospace is limited to code-like data; surface effects (glass, grain,
  tinted shadows) are opt-in.
- `quality-gates`: Gate 7 points to the measurable layout procedure in `frontend-design`.
- `redesign` and `brandkit` now carry an "Adapted from" line.

## [0.0.1] - Initial release

- Initial public commit of the Plaesy Spec-Kit framework: prompts, instructions,
  agents/roles, checklists, templates, and bash/PowerShell automation scripts.
