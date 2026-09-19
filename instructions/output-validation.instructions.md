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
