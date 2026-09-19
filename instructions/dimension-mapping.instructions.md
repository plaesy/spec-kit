---
description: "Dimension-to-command mapping for intelligent workflow routing"
applyTo: "routing-decisions, phase-orchestration, autonomous-routing"
---

# Dimension-Command Mapping Reference

> **Purpose**: Route findings from any dimension to the command built to handle
> them. This is the **canonical** routing table — `/assess`, `/implement`,
> `/optimize`, `/fix`, and `/improve` all point here instead of restating it.

What gets measured/built/fixed per dimension lives in that dimension's own file —
`assess-{dimension}.instructions.md` for assessment criteria, `/optimize`'s
Optimization Targets table for tuning targets. This file only maps dimension →
command; it does not restate their content.

---

## Routing Table

| Dimension | Scope | Assess | Implement | Optimize (top priority) | Error Recovery (critical path) |
|---|---|---|---|---|---|
| **Technical** | Code quality, architecture, tests, security, infrastructure, deployment | `/assess:technical` | `/implement:technical` | `/optimize:technical --focus performance` (DB/caching first) | `/fix:technical --focus security` (immediate) |
| **Design** | UI/UX, WCAG accessibility, design systems, visual consistency | `/assess:design` | `/implement:technical --focus wcag` (code); `/assess:design` Mode 1 (spec + tokens) | `/optimize:design --focus design-system` | `/fix:design` (WCAG violation — immediate) |
| **Business** | Market fit, business model viability, revenue, ROI | `/assess:business` | `/implement:business` | `/optimize:business` (unit-economics); `/loop:business` (continuous) | `/assess:business --research` (re-validate assumption) |
| **Product** | Feature set, roadmap, competitive advantage | `/assess:product` | `/implement:product` | `/optimize:product` (roadmap prioritization) | `/fix:product` |
| **Marketing** | Positioning, messaging, go-to-market | `/assess:marketing` | `/implement:marketing` | `/optimize:marketing` (claim sourcing); `/loop:marketing` | `/assess:marketing --research` (reposition) |
| **Legal** | Regulatory compliance, data privacy, legal risk | `/assess:legal` | `/fix:legal --focus compliance` (drafting is compliance-first) | N/A — compliance is binary pass/fail, not gradual | `/fix:legal` (CRITICAL — immediate) |
| **Financial** | Cost structure, pricing, profitability, funding | `/assess:financial` | `/implement:financial` | `/optimize:financial --focus cost` or `--focus pricing` | `/assess:financial --research` (revalidate model) |
| **Management/Operations** | Team capacity, process efficiency, org health, deployment readiness | `/assess:management` | `/implement:management` | `/loop:management --focus process` | `/fix:management --focus deployment` |

Every scope selector above is `/{command}:{dimension}`; `--focus {sub-area}`
narrows within it — it never selects the dimension itself (see any prompt's Usage
Format for the full pattern).

**Not scored/fixed here**: framework adoption gaps, version drift, tech-stack/
design modernization, spec traceability — these are `/improve`'s lane, not a
dimension-routing decision (see `/improve`).

---

## Priority Mapping (When Multiple Findings Compete)

| Priority | Trigger | Response Time | Command |
|----------|---------|----------------|---------|
| **CRITICAL** | Security / legal / compliance breach, data loss, service down | Immediate (≤1h) | `/fix` (blocking) |
| **HIGH** | Performance below constitution target | Next phase | `/optimize` or `/fix` |
| **HIGH** | Test coverage below constitution minimum | Before merge | `/implement --focus tdd` |
| **MEDIUM** | Design inconsistency, unsourced business/market claim | Next sprint / research phase | `/optimize:design` or `/assess:{dimension} --research` |
| **LOW** | Documentation gap, minor code-quality cleanup | Backlog / batch | `/doc` or `/loop` |

---

## See Also

- `plaesy.instructions.md` — Global mandatory instructions + routing logic
- `/assess`, `/implement`, `/optimize`, `/fix`, `/improve` — the commands this
  table routes to
- `assess-technical.instructions.md`, `assess-design.instructions.md`,
  `assess-business.instructions.md`, `assess-financial.instructions.md`,
  `assess-marketing.instructions.md`, `assess-legal.instructions.md`,
  `assess-management.instructions.md`, `assess-product.instructions.md` —
  per-dimension assessment criteria (what gets measured, not routed)
- `error-recovery-predictive.instructions.md` — Predictive error prevention
