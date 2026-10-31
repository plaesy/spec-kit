# Changelog

All notable changes to Spec-Kit are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project uses [Semantic Versioning](https://semver.org/) once it reaches 1.0.0.

## [Unreleased]

## [1.0.0] - 2026-10-31

### Added
- **Core Pipeline Skills** (8 skills with evals):
  - `ideator` — Divergent idea generation (8-12 options, assumption + cheapest test + FACT/INFERRED/SPECULATIVE labels)
  - `buyer-discovery` — Outreach tracking → interviews → signal counting + Segment Validation Matrix
  - `compliance-health` — BPOM/FTC rule engine, risk scoring (0-100), approved/banned claims DB
  - `economics-engine` — Unit economics calculator with verdict engine, segment analysis, sensitivity analysis
  - `plan-compiler` — Auto-compiles 11 stages → plan.md with verdict + pivot docs
  - `assess` — Dimension assessments (business, marketing, financial, product, ops) with grades A+-D
  - `pipeline` — Master orchestrator: 15 stages + Pivot Decision Gate + resume-from-stage
  - `youtube-automation` — 7-agent video pipeline: Research → Fact-check → Script → Visuals → Voice → Render → Review/Publish

- **Oracle Harness** (`scripts/oracle_runner.py`):
  - Golden test recording/replay for critical skills
  - Function-level call vectors
  - Replay against alternate implementations

- **Fail-Soft Tool Chains** (`skills/fallback_chains.py`):
  - Web Search: Exa → Jina → DuckDuckGo
  - Web Fetch: Jina Reader → Exa → Browser
  - LLM: Gemini → Groq → OpenRouter → Pollinations
  - Visuals: Pexels → Pixabay → NASA → Openverse → AI-generated
  - Transcription: Groq Whisper → OpenAI Whisper → Local Whisper
  - Thumbnail: 4 styles as fallback chain

- **Quality Gates** (`.plaesy/instructions/quality-gates.md`):
  - 7 blocking business validation gates (B1-B7)
  - 9 pipeline stage gates (S1-S9)
  - Adversarial verifier pattern

- **Success Contracts** (`.plaesy/instructions/success-contract.md`):
  - 8 contract templates (Idea Discovery, Buyer Discovery, Economics, Pivot Gate, Plan Compilation, Assessment, Full Pipeline, YouTube Automation)

- **Plugin Manifests** (multi-platform):
  - `.claude-plugin/plugin.json` — Claude Code
  - `.codex-plugin/plugin.json` — OpenAI Codex
  - `.agents/plugins/marketplace.json` — Agents SDK

- **CI/CD** (`.github/workflows/ci.yml`):
  - Evals gate (all 8 skills)
  - Oracle harness replay (4 critical skills)
  - Plugin manifest validation
  - Skill structure validation
  - Fallback chain import test
  - Documentation check

- **Documentation**:
  - `docs/ARCHITECTURE.md` — System architecture, data flow, stage mapping
  - `docs/DESIGN.md` — Design system, stage templates, error states
  - `docs/CLI-CONTRACT.md` — Complete CLI specifications for all 8 commands

- **Using Plaesy Updates**:
  - Skill registry with 8 spec-kit skills
  - Pipeline stage map with 15+1 stages
  - Routing table with pivot gate
  - Phase mapping (9-phase ladder + 15-stage pipeline)
  - Quality gates with B1-B7 + S1-S9
  - Success contracts with 8 templates

### Changed
- Renamed skills from `spec-kit-*` prefix to clean names (`ideator`, `buyer-discovery`, etc.)
- Updated `using-plaesy` routing to use clean skill names
- Extended pipeline to 16 stages (added YouTube Automation as optional Stage 16)
- Added Pivot Decision Gate as blocking gate between Stage 9 and 10

### Fixed
- Quality gate B3 now properly blocks pipeline until pivot decision documented
- Oracle harness records function-level call vectors for replay
- Fail-soft chains properly skip unavailable tools (no API key = skip)

---

## [0.0.1] - Initial release

- Initial release of the Spec-Kit framework: prompts, instructions, agents/roles, checklists, templates, and automation scripts.