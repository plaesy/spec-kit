---
name: assess
description: "Assessment workflows per dimension: business, design, financial, legal, management, marketing, operations, product, technical. Use when the user asks to assess or audit one dimension; load the file for that dimension."
---

# Assessment

One file per dimension. Each file holds that dimension's workflow, quality weights,
critical rules, finding standards, and completion checklist.

| File | Dimension |
|------|-----------|
| `references/business.md` | Market fit, business model, financial viability |
| `references/design.md` | UI/UX, accessibility, design systems |
| `references/financial.md` | Cost structure, pricing, profitability, funding, unit economics |
| `references/legal.md` | Regulatory compliance, data privacy, risk management |
| `references/management.md` | Team capacity, project management, process efficiency |
| `references/marketing.md` | Positioning, messaging, audience fit, go-to-market readiness |
| `references/operations.md` | Deployment, release, reliability/SLOs, observability, infrastructure as code, runbooks |
| `references/product.md` | Feature set, roadmap, competitive positioning |
| `references/technical.md` | Code quality, architecture, tests, security, performance |

Run the dimension files for the scope requested. `/assess` runs them all.
