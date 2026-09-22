---
title: "Spec-Kit Design Assessment 2026-09-22"
updatedAt: "2026-09-22T14:35:00Z"
---

# Design Assessment — Plaesy Spec-Kit Framework

**Scope**: Prompt quality, instruction quality, template quality, orphaned files, CLI ergonomics, docs UX
**Confidence**: HIGH for file inventory; MEDIUM for quality judgments

---

## Scores

| Dimension | Score | Notes |
|-----------|-------|-------|
| Prompt Quality | 78/100 | All 11 prompts have frontmatter and usage examples; some have inconsistent formatting |
| Instruction Quality | 72/100 | 40+ instructions exist; most have clear purposes; some lack usage examples |
| Template Quality | 65/100 | 10 templates exist; some are stubs; inconsistent example coverage |
| Orphaned Files | 85/100 | Most files referenced; a few docs/ files may be stale |
| CLI Ergonomics | 70/100 | Consistent bash/PS pairing; some CLI dispatch bugs (PowerShell `analyze` token bug) |
| Docs UX | 68/100 | README comprehensive but some docs contradict code (graph watch-mode) |

**Overall Design Score: 73/100** (below 85 threshold → minor improvements needed)

---

## Top 5 Findings

### 1. Inconsistent Usage Example Coverage (MEDIUM)
- **Location**: Various `instructions/*.instructions.md`, `prompts/*.md`
- **Issue**: Some instructions have usage examples, others don't. Constitution requires "every prompt/instruction/template ships a usage example."
- **Confidence**: MEDIUM — spot-check of 10 files showed 3 without examples
- **Impact**: New users get inconsistent onboarding experience

### 2. Stale Documentation Contradicting Code (MEDIUM)
- **Location**: `docs/scripts/plaesy-graph.md:16` vs `scripts/bash/plaesy-graph.sh`
- **Issue**: Graph documentation claims watch-mode support that the code doesn't implement
- **Confidence**: HIGH — direct doc vs code comparison
- **Impact**: Users follow docs that lead to dead ends

### 3. Template Stub Inconsistency (LOW)
- **Location**: `templates/` directory
- **Issue**: Some templates are full scaffolds, others are minimal stubs. No consistent template structure.
- **Confidence**: HIGH — file inventory
- **Impact**: Template output quality varies by template

### 4. PowerShell CLI Dispatch Bug (HIGH)
- **Location**: `scripts/powershell/install.ps1:387`
- **Issue**: Comment notes "plaesy-analyze.ps1 binds it to -ProjectPath" but the `analyze` token is still forwarded as a project path in some dispatch paths
- **Confidence**: MEDIUM — comment indicates known issue
- **Impact**: PowerShell users get confusing errors when running analyze

### 5. Docs Index May Have Orphans (LOW)
- **Location**: `docs/` directory
- **Issue**: Some doc files may not be referenced from the main README or any prompt/instruction
- **Confidence**: LOW — not exhaustively checked
- **Impact**: Users discover stale docs through search

---

## Missing Information

- Actual usage example coverage count (exact percentage)
- User feedback on CLI ergonomics
- Whether any docs/ files are truly orphaned

## Assumptions

- "Usage example" means a code block or command showing how to use the artifact
- Orphan check limited to docs/ and templates/ (not full repo sweep)

## Recommended Next Phase

**`/optimize:design`** — Add missing usage examples, fix stale docs, fix PowerShell CLI dispatch, standardize templates. Then re-assess.