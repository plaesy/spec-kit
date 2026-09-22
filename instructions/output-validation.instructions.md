---
description: "Structured output validation framework for all Spec-Kit phases"
---

# Output Validation Framework

Every phase output MUST validate against these criteria:

## 1. Completeness Check
- All required fields present? ✓/✗
- All arrays populated (not empty)? ✓/✗
- All references resolved? ✓/✗

## 2. Format Consistency
- Dates in ISO 8601? ✓/✗
- Numbers properly formatted? ✓/✗
- Markdown headings consistent? ✓/✗

## 3. Coherence Check
- No contradictions between sections? ✓/✗
- Cross-references consistent? ✓/✗
- Tone consistent throughout? ✓/✗

## 4. Quality Check
- Spelling/grammar correct? ✓/✗
- Technical terms accurate? ✓/✗
- Examples valid/runnable? ✓/✗

## 5. Self-Audit Before Delivery
- Have I verified every assumption?
- Could output be misinterpreted?
- Are all references/links valid?
- Would I confidently deliver this?

## Usage Example

```
Phase output: assess-technical-2026-09-22.md

1. Completeness: all 6 dimension scores present ✓, no empty findings array ✓
2. Format: dates ISO 8601 (2026-09-22) ✓, headings consistent ✓
3. Coherence: score table matches the findings listed below it ✓
4. Quality: every code reference resolves to a real file:line ✓
5. Self-audit: findings re-verified against source before delivery ✓
→ Passes validation; safe to hand off to /save
```
