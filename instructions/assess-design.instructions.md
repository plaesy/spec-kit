---
description: "Design assessment workflow - UI/UX, accessibility, design systems"
---

# Design Assessment Instructions

**Use with**: `/assess` when dimension is **Design** or frontend/UI project detected

**Referenced from**: `.plaesy/memory/assess.md` (orchestrator)

**Note**: Load ONLY if project has frontend/UI (React, Vue, Angular, Flutter, web UI, mobile UI, or Figma files)

---

## Assessment Workflow

### Step 1: Component Consistency Audit

- Check if UI components follow naming conventions
- Verify component variants are documented
- Assess component library organization
- Validate prop interfaces and documentation
- Review component naming vs design spec

**Rate**: 0-100 based on consistency level

### Step 2: Accessibility Audit (WCAG 2.1 AA)

- **Color Contrast**: Verify WCAG AA contrast ratios (4.5:1 for text, 3:1 for UI)
- **Keyboard Navigation**: Check all interactive elements keyboard accessible
- **ARIA Labels**: Verify ARIA labels and roles where appropriate
- **Screen Reader Support**: Test with screen reader
- **Focus Management**: Check focus indicators and tab order
- **Form Labels**: Verify all form inputs have proper labels
- **Alternative Text**: Check images have alt text

**Rate**: 0-100 based on WCAG AA compliance

### Step 3: Design Tokens Assessment

- Verify colors defined as tokens (not hardcoded hex values)
- Check spacing/sizing tokens used consistently
- Validate typography tokens for fonts, sizes, weights
- Assess token naming and organization
- Review light/dark theme token variations

**Rate**: 0-100 based on token usage

### Step 4: Dark Mode Assessment

- Verify light/dark color pairs defined
- Check contrast ratios in both themes
- Validate theme consistency across components
- Assess user preference detection (prefers-color-scheme)
- Review theme switching mechanism

**Rate**: 0-100 based on dark mode implementation

### Step 5: Design System Adoption

- Measure component reuse rate from design library
- Check design library usage consistency
- Assess one-off component creation vs library reuse
- Review design spec vs implementation alignment
- Identify design system gaps

**Rate**: 0-100 based on design system adoption

### Step 6: Design Documentation

- Check if design spec exists and is current
- Verify component documentation completeness
- Review design rationale documentation
- Assess design guidelines and usage
- Validate Figma organization (if applicable)

**Rate**: 0-100 based on documentation

### Step 7: Figma Organization (if applicable)

- Check file structure and organization
- Verify component library setup
- Validate Code Connect mappings (if present)
- Assess Figma plugin usage
- Review design-to-code workflow

**Rate**: 0-100 based on Figma setup

### Step 8: Generate Report

- Overall design quality score (0-100)
- Per-category scores (Consistency, Accessibility, Tokens, Dark Mode, Library Adoption, Docs, Figma)
- Top 5 findings (ordered by impact)

**Blocking Issues**:
- WCAG AA compliance failures = BLOCKER (must fix before shipping)
- Missing component documentation = medium priority
- Inconsistent naming = low priority

---

## Quality Scoring

### Design Quality Components

```
component_consistency: 25%   # Naming conventions, variants, documentation
accessibility: 35%           # WCAG 2.1 AA compliance, contrast, keyboard nav, ARIA
design_tokens: 20%          # Token usage, consistency, light/dark themes
design_system_adoption: 20% # Component reuse rate, library usage
```

### Grade Mapping

```
95-100: A+ (Exceptional - Design system mature, WCAG AA compliant, well-documented)
90-94:  A  (Excellent - Good design system adoption, minor accessibility issues)
85-89:  B+ (Very Good - Design tokens in place, WCAG AA mostly compliant)
80-84:  B  (Good - Design system started, some accessibility gaps, address in next iteration)
70-79:  C  (Acceptable - Inconsistent design, significant accessibility issues)
<70:    D  (Needs Work - Design system gaps, critical accessibility issues)
```

### Success Criteria (Hard Stops)

- MUST have: WCAG 2.1 AA compliance (or documented exceptions)
- MUST have: No critical accessibility failures (color contrast, keyboard nav, screen reader)
- SHOULD have: Design tokens in place (or color vars defined)
- SHOULD have: Component documentation

---

## Critical Rules

- ✅ **WCAG AA MANDATORY** - Accessibility is non-negotiable for UI projects
- ✅ **MEASURE CONTRAST** - Use automated tools to verify contrast ratios
- ✅ **TEST KEYBOARD NAV** - Tab through entire interface
- ✅ **TEST SCREEN READER** - Verify with actual screen reader
- ✅ **CITE WITH LOCATION** - Every finding includes component name and location
- ✅ **VERIFY TOKENS** - Check actual code, don't assume token usage

---

## Anti-Patterns (NEVER Do These)

- ❌ **Never assess design on backend-only projects** - Design quality only for frontend/UI
- ❌ **Never skip WCAG audit** - Accessibility is mandatory
- ❌ **Never speculate on contrast** - Use tools to verify actual ratios
- ❌ **Never assume keyboard nav works** - Test tab order and focus
- ❌ **Never ignore dark mode** - Test both light and dark themes
- ❌ **Never report unverified issues** - Test with actual tools/browsers
- ❌ **Never mix design and code issues** - Separate component design from implementation quality
- ❌ **Never recommend design changes** - Only assess what exists

---

## Testing Tools

**Accessibility**:
- Axe DevTools (WCAG scanning)
- WAVE (contrast and accessibility audit)
- Lighthouse (WCAG audit)
- Screen reader testing (NVDA, JAWS)
- Keyboard navigation testing (manual)

**Design**:
- Chrome DevTools (color inspection)
- Contrast analyzers (WebAIM, Contrast Ratio)
- Figma plugins for token validation
- Component library audits

---

## Accessibility Checklist

- [ ] Color contrast ≥4.5:1 for text (WCAG AA)
- [ ] Color contrast ≥3:1 for UI components
- [ ] All interactive elements keyboard accessible
- [ ] Tab order logical and predictable
- [ ] Focus indicators visible
- [ ] Form inputs have labels
- [ ] Images have alt text
- [ ] ARIA labels where appropriate
- [ ] Screen reader compatible
- [ ] No keyboard traps

---

## Pre-Assessment Validation

Run these checks BEFORE attempting assessment:

```
Is this a frontend/UI project?
  → yes (React/Vue/Angular/Flutter/web-UI): Continue with design assessment
  → no (backend-only): Skip design review entirely

Can you access the UI/design files?
  → yes (codebase or Figma): Continue
  → no: Stop, request access

Do design tools exist?
  → yes (Figma/design spec): Assess design quality
  → no: Assess component documentation only
```

---

## Success Criteria

Assessment is complete when:

- ✅ Component consistency audited with examples
- ✅ WCAG 2.1 AA compliance verified with tools
- ✅ Design tokens assessed and documented
- ✅ Dark mode implementation reviewed
- ✅ Design system adoption measured
- ✅ Design documentation reviewed
- ✅ Scores calculated per category
- ✅ Top 5 findings listed with component locations
- ✅ WCAG violations clearly flagged as blockers
- ✅ Next phase recommended (fix WCAG → refactor design → optimize)
- ✅ Console output delivered to user
