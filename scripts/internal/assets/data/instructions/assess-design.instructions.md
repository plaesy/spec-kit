---
description: "Design assessment workflow - UI/UX, accessibility, design systems"
applyTo: "**/*"
---

# Design Assessment Instructions

**Use with**: `/assess` when dimension is **Design** or frontend/UI project detected

**Referenced from**: `/assess` (orchestrator)

**Note**: Load ONLY if project has frontend/UI (React, Vue, Angular, Flutter, web UI, mobile UI, or Figma files)

**Sibling**: `.plaesy/instructions/ui-ux-design-principles.md` applies the same
hierarchy/feedback/affordance/IA/navigation principles *while building* — this
file scores what's already shipped against them (plus WCAG) after the fact.

---

## Assessment Workflow

### Step 0: Design System Source of Truth (`.plaesy/memory/design.md`)

- Check whether `.plaesy/memory/design.md` exists (project-wide design tokens +
  rationale, following the Google Labs DESIGN.md spec — YAML front matter for
  machine-readable tokens, Markdown body for intent/rationale). See
  `.plaesy/instructions/how-to-create-designmd.md` for the field list and required
  body-section order.
- **Missing** → this is itself a finding (design tokens have no single source of
  truth across features) → Mode 1 (`/assess:design`, Design Spine) generates it
  from `.plaesy/templates/design.template.md` before Steps 3-4 below can be meaningfully
  scored
- **Exists** → use it as the reference for Steps 3 (Design Tokens) and 4 (Dark
  Mode) below: code tokens are audited *against this file*, not against a vague
  "are tokens used" heuristic
- **Stale** (code has tokens/colors not present in the file, or vice versa) →
  flag as a MEDIUM finding — the file has drifted from the actual implementation

### Step 1: Component Consistency Audit

- Check if UI components follow naming conventions
- Verify component variants are documented
- Assess component library organization
- Validate prop interfaces and documentation
- Review component naming vs design spec

**Rate**: 0-100 based on consistency level

### Step 2: Accessibility Audit (WCAG 2.2 AAA)

- **Color Contrast**: Verify WCAG AAA contrast ratios (4.5:1 for text, 3:1 for UI)
- **Keyboard Navigation**: Check all interactive elements keyboard accessible
- **ARIA Labels**: Verify ARIA labels and roles where appropriate
- **Screen Reader Support**: Test with screen reader
- **Focus Management**: Check focus indicators and tab order
- **Form Labels**: Verify all form inputs have proper labels
- **Alternative Text**: Check images have alt text

**Rate**: 0-100 based on WCAG AAA compliance

### Step 3: Design Tokens Assessment

- Compare code tokens against `.plaesy/memory/design.md` (Step 0) — every color/
  spacing/typography value in code should trace back to an entry there
- Verify colors defined as tokens (not hardcoded hex values)
- Check spacing/sizing tokens used consistently
- Validate typography tokens for fonts, sizes, weights
- Assess token naming and organization
- Review light/dark theme token variations
- Flag any code token with no corresponding entry in `design.md` (drift) and any
  `design.md` entry never referenced in code (dead token)

**Rate**: 0-100 based on token usage AND alignment with `design.md`

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

### Step 8: Uncertainty Surfacing (Before Finalizing Assessment)

Load `.plaesy/instructions/uncertainty-surfacing.md`; do not restate it here.
Only this dimension's specifics follow.

**Confidence basis for this dimension**:

- HIGH: the Figma file, design system, or shipped UI was read directly
- MEDIUM: a pattern matches comparable components, but this instance was not opened
- LOW: no design artefact was accessible — no Figma access, no rendered UI

**Worked example**:

```text
I could not assess keyboard navigation because Figma access was not available
and the prototype is not deployed.
To improve confidence, provide a Figma link covering the interactive component
set, or a deployed build to inspect.
```

### Step 9: Generate Report

**Report sections**: canonical names per `.plaesy/instructions/dimension-mapping.md`
→ *Assessment Report Section Registry*; the bullets below go *inside* those
sections and do not rename them.

- Overall design quality score (0-100)
- Per-category scores (Consistency, Accessibility, Tokens, Dark Mode, Library Adoption, Docs, Figma)
- Top 5 findings (ordered by impact)

**Blocking Issues**:

- WCAG AAA compliance failures = BLOCKER (must fix before shipping)
- Missing component documentation = medium priority
- Inconsistent naming = low priority

---

## Quality Scoring

### Design Quality Components

```text

component_consistency: 25%   # Naming conventions, variants, documentation
accessibility: 35%           # WCAG 2.2 AAA compliance, contrast, keyboard nav, ARIA
design_tokens: 20%          # Token usage, consistency, light/dark themes
design_system_adoption: 20% # Component reuse rate, library usage
```

### Grade Mapping

| Grade | What it means here |
|---|---|
| A+ | Exceptional - Design system mature, WCAG AAA compliant, well-documented |
| A | Excellent - Good design system adoption, minor accessibility issues |
| B+ | Very Good - Design tokens in place, WCAG AAA mostly compliant |
| B | Good - Design system started, some accessibility gaps, address in next iteration |
| C | Acceptable - Inconsistent design, significant accessibility issues |
| D | Needs Work - Design system gaps, critical accessibility issues |

The score edges behind these letters, and the general meaning of each band,
are defined once in `.plaesy/instructions/quality-gates.md` →
**Grade bands**. Do not restate them here.

### Success Criteria (Hard Stops)

- MUST have: WCAG 2.2 AAA compliance (or documented exceptions)
- MUST have: No critical accessibility failures (color contrast, keyboard nav, screen reader)
- SHOULD have: `.plaesy/memory/design.md` present and not drifted from code tokens
- SHOULD have: Component documentation

---

## Critical Rules

- ✅ **WCAG AAA MANDATORY** - Accessibility is non-negotiable for UI projects
- ✅ **MEASURE CONTRAST** - Use automated tools to verify contrast ratios
- ✅ **TEST KEYBOARD NAV** - Tab through entire interface
- ✅ **TEST SCREEN READER** - Verify with actual screen reader
- ✅ **CITE WITH LOCATION** - Every finding includes component name and location
- ✅ **VERIFY TOKENS** - Check actual code, don't assume token usage

---

## Finding Standards

- **Assess design only where a frontend/UI surface exists** — a backend-only
  project scores this category N/A with a stated reason
- **Run the WCAG audit** wherever a UI exists; accessibility is mandatory there
- **Verify contrast with a tool** and cite the measured ratio
- **Test tab order and focus**, and report what actually happens
- **Test both light and dark themes**
- **Reproduce each issue in a real browser or tool** before reporting it
- **Separate component design from implementation quality** — a design finding and
  a code finding are two findings, not one
- **Assess the interface as it exists**; recommendations belong to `/optimize:design`

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

- [ ] Color contrast ≥4.5:1 for text (WCAG AAA)
- [ ] Color contrast ≥3:1 for UI components
- [ ] All interactive elements keyboard accessible
- [ ] Tab order logical and predictable
- [ ] Focus indicators visible
- [ ] Form inputs have labels
- [ ] Images have alt text
- [ ] ARIA labels where appropriate
- [ ] Screen reader compatible
- [ ] No keyboard traps
- [ ] Target size ≥24×24px for interactive elements (WCAG 2.2 SC 2.5.8, unless an exception applies)
- [ ] Focus indicator not fully obscured by other content (WCAG 2.2 SC 2.4.11)
- [ ] Any drag-based interaction has a single-pointer (non-dragging) alternative achieving the same outcome (WCAG 2.2 SC 2.5.7)

---

## Pre-Assessment Validation

Run these checks BEFORE attempting assessment:

```text

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

## Assessment Completion Checklist

Assessment is complete when:

- ✅ Component consistency audited with examples
- ✅ WCAG 2.2 AAA compliance verified with tools
- ✅ Design tokens assessed and documented
- ✅ Dark mode implementation reviewed
- ✅ Design system adoption measured
- ✅ Design documentation reviewed
- ✅ Scores calculated per category
- ✅ Top 5 findings listed with component locations
- ✅ WCAG violations clearly flagged as blockers
- ✅ Next phase recommended (fix WCAG → refactor design → optimize)
- ✅ Console output delivered to user

## Persona Protocol

Loaded by every `each family:design` persona prompt for this dimension — `/assess`,
`/implement`, `/fix`, `/optimize`, `/loop` and `/improve` all point their
`Loads:` line here. The rules below are dimension-neutral and are written here
once rather than restated in 54 persona files, where they had already drifted
apart from the parents they point at.

- **Follow the parent protocol.** The persona fixes the *dimension*; the family
  prompt fixes the *verb*. Follow the family prompt's full protocol. Where the
  parent and this file disagree, the parent governs the behaviour.
- **This file is the *assessment* criteria for `design`** — workflow, weights,
  grade bands, mandatory audit. Under `/assess` that is what you are scored
  against. Under `/implement`, `/fix`, `/optimize`, `/loop` and `/improve`
  it is the only per-dimension reference that exists: the parent supplies the
  verb, this file supplies the subject matter. If you need criteria the parent
  does not define, say so in your report rather than borrowing the assessment
  workflow — an `/optimize` run is not an `/assess` run with a different verb.
- **`$ARGUMENTS` is scoping detail, never a selector.** It narrows work *within*
  the dimension. It cannot select a different dimension, and it cannot be used to
  broaden scope to another dimension — that requires the user to invoke the
  other family explicitly (`/fix:technical,design`). A literal-minded reading of
  `$ARGUMENTS` as a command name is the failure this rule exists to prevent.
- **Do not manufacture work to justify being loaded.** A persona that finds
  nothing in its dimension reports "no findings in this dimension". Loading every
  persona for every run is what produces filler findings, not smaller honest ones.
- **Report the dimension you actually assessed.** Say which dimension ran. A
  finding filed under a neighbouring dimension's name is routed by that
  dimension's rules and will be closed by the wrong reviewer.
