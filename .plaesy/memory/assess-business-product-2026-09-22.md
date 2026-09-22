---
title: "Spec-Kit Business/Product Assessment 2026-09-22"
updatedAt: "2026-09-22T14:35:00Z"
---

# Business/Product Assessment — Plaesy Spec-Kit Framework

**Scope**: Business viability, product completeness, marketing readiness, roadmap, adoption signals
**Confidence**: HIGH for product facts; MEDIUM for business judgments; LOW for competitive claims

---

## Scores

| Dimension | Score | Notes |
|-----------|-------|-------|
| Business Viability | 72/100 | Clear value proposition for spec-driven development; no revenue model yet (open source) |
| Product Completeness | 68/100 | 11 prompt phases exist but some are stubs; analyzer has known gaps |
| Marketing Readiness | 75/100 | README comprehensive; CHANGELOG updated; no aspirational claims found |
| Roadmap/CHANGELOG | 80/100 | CHANGELOG follows Keep a Changelog; recent entries accurate |
| Adoption Signals | 60/100 | Dogfooding works but analyzer gaps undermine credibility; no usage metrics |

**Overall Business/Product Score: 71/100** (below 85 threshold → improvements needed before next release)

---

## Top 5 Findings

### 1. Product Completeness Gap: Analyzer (HIGH)
- **Location**: `scripts/bash/plaesy-analyze.sh`, `scripts/powershell/plaesy-analyze.ps1`
- **Issue**: The analyzer is the framework's core differentiator but has known gaps: no `--if-changed`, GNU-only Bash, narrower PowerShell insights. Users running `plaesy analyze` get inconsistent results across platforms.
- **Confidence**: HIGH — source inspection and prior assessment
- **Impact**: Core feature is unreliable; undermines framework credibility

### 2. No Usage Metrics or Feedback Loop (MEDIUM)
- **Location**: Entire repo
- **Issue**: No telemetry, no feedback collection mechanism, no issue templates for user reports. Maintainers can't tell what's breaking in the wild.
- **Confidence**: HIGH — repo inspection
- **Impact**: Issues go undetected; product decisions lack user data

### 3. CHANGELOG Accuracy (LOW)
- **Location**: `CHANGELOG.md`
- **Issue**: CHANGELOG is accurate but only tracks framework-internal changes. User-visible feature changes in prompts/instructions aren't always reflected.
- **Confidence**: MEDIUM — CHANGELOG inspection
- **Impact**: Users can't track what changed in their workflow

### 4. README Positioning vs Reality (MEDIUM)
- **Location**: `README.md`
- **Issue**: README claims comprehensive capabilities but the analyzer gaps (no `--if-changed`, parity bugs) mean the actual experience falls short of the pitch.
- **Confidence**: MEDIUM — README vs code comparison
- **Impact**: First-time users hit friction; may not return

### 5. No Competitive Analysis or Differentiation Articulation (LOW)
- **Location**: README, docs/
- **Issue**: The framework's unique value vs other spec-driven tools isn't clearly articulated. No comparison table, no "why Spec-Kit" section.
- **Confidence**: LOW — absence of evidence
- **Impact**: Prospective adopters can't justify choosing Spec-Kit over alternatives

---

## Missing Information

- Actual user adoption numbers or download counts
- Competitive landscape analysis
- User satisfaction surveys or feedback
- Time-to-value metrics for new users

## Assumptions

- "Product completeness" means the framework can be used end-to-end by a typical developer
- The repo IS the product (no separate marketing site)

## Recommended Next Phase

**`/assess:marketing`** — Clarify positioning, add differentiation section to README, add issue templates. Then **`/implement`** to fix analyzer gaps (the product completeness blocker).