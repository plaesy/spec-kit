---
name: compliance-health
category: productivity
description: BPOM/FTC health claims compliance checker with DB.
---

# compliance-health

Health niche compliance skill: BPOM (Indonesia) + FTC (US) claims validation.

## When to Use
- Building health/wellness/supplement products that make claims
- Need to validate marketing copy before publish
- Want approved claims database for AI content generation

## Core Components

### 1. Approved Claims Database
- BPOM-approved health claims (Indonesia)
- FTC substantiation requirements (US)
- EFSA health claims (EU) - optional
- Structure: claim_text, category, evidence_level, source_url, last_verified

### 2. Rule Engine
- Keyword flagging: banned terms ("cure", "heal", "treat", "prevent disease")
- Required disclaimers: "This statement has not been evaluated by BPOM/FDA"
- Structure/function claim vs disease claim classification
- Auto-suggest compliant alternatives
- **Risk scoring (0-100): critical flags=25, high=15, medium=10, low=5; pass if <30**

### 3. Monthly Update Workflow
- Monitor BPOM/FTC warning letters
- Update database with new approved/revoked claims
- Version control for audit trail

### 4. Content Generation Integration
- Pre-generation filter: reject prompts likely to produce non-compliant content
- Post-generation check: all AI output passes through rule engine
- User-facing: real-time compliance score per content piece
- Export: compliant content + required disclaimers auto-appended

## Input/Output

```
Input: marketing_copy (string)
Output: {
  passed: boolean,
  flags: [{term, position, severity, suggestion, regulation}],
  required_disclaimers: [string],
  risk_score: 0-100
}
```

## Integration
- Called by founder skill during content generation
- Used in spec-compliance for formal requirements
- CI/CD gate for marketing content publish

## Evals

See skills/compliance-health/evals/evals.json and results.json (7 tests, 100% pass).