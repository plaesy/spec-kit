# /design — Generate UI/UX Designs and Specifications

## Overview

The `/design` prompt is a **new phase** in the Plaesy workflow for creating production-ready UI/UX designs. It bridges requirements (`/clarify`) and implementation (`/implement`), supporting three primary use cases:

1. **Prototype new features** — From specs to high-fidelity mockups and component library
2. **Refactor existing UI** — Audit → design improvements → implementation guide
3. **Build design systems** — Component libraries, design tokens, documentation

## When to Use `/design`

| Phase | Purpose | Input | Output |
|-------|---------|-------|--------|
| **Prototype** | Starting fresh UI | Feature spec from `/clarify` | Figma file + design spec |
| **Refactor** | Improve existing UI | Current app + audit findings | Before/after mockups + migration guide |
| **Design System** | Build reusable components | Component inventory | Design lib + tokens + guide |

## Key Features

### ✅ Constitutional Rules
- **ACCESSIBILITY-FIRST** — WCAG 2.1 AA minimum (contrast, keyboard navigation)
- **COMPONENT-DRIVEN** — All UI elements are reusable with variants/states
- **DESIGN-TOKENS** — Colors, spacing, typography defined as tokens
- **MOBILE-FIRST** — Design mobile first, scale to desktop
- **DARK MODE READY** — Colors work in both light and dark themes
- **MEASURABLE UX** — Success metrics for time-to-task, error rates, conversion
- **INTERACTION PATTERNS** — Document hover, focus, active, disabled, loading, error states
- **PERFORMANCE-AWARE** — Image optimization, lazy loading, animation frame rate targets

### 🎯 Workflow Steps

1. **Requirement Analysis** — Detect context (prototype/refactor/design system), identify personas, audit existing design
2. **Design Specification** — Wireframes → high-fidelity mockups → component inventory
3. **Component System** — Define components, props, states, variants, naming conventions
4. **Design Tokens** — Colors (light/dark), typography (hierarchy), spacing (8px grid), shadows/elevation
5. **Deliverables** — Figma file + design spec doc + component catalog + accessibility checklist
6. **Handoff** — Code Connect mappings, implementation guide, design dev guide

### 🔌 Integrations

**Context7 Protocol:**
- Design system library documentation (Material Design 3, Tailwind UI, Apple HIG)
- WCAG 2.1 accessibility standards
- Fallback: Industry-standard patterns if Context7 unavailable

**Figma Integration:**
- Full Figma support via `/figma-use` skill
- Component library management
- Code Connect for component-to-code mapping
- Asset export (icons, illustrations, responsive images)

### 📋 Design Specification Template

Every design includes:
- **Overview** — Feature goals, user personas, success metrics
- **Wireframes & Flows** — User journeys, screen hierarchy, navigation
- **High-Fidelity Designs** — All screens/states (default, hover, focus, active, disabled, loading, error)
- **Component Specs** — Props, states, variants, accessibility, keyboard navigation
- **Design Tokens** — Colors, typography, spacing, shadows (YAML format)
- **Interaction Patterns** — Animations, micro-interactions, loading/error handling
- **Responsive Design** — Mobile, tablet, desktop breakpoints
- **Accessibility Checklist** — WCAG compliance, contrast ratios, ARIA labels, keyboard navigation
- **Implementation Notes** — Engineer-specific considerations, performance targets

### 🚫 Anti-Patterns (Never Do These)

- ❌ **Skip accessibility review** — Every design must have WCAG audit
- ❌ **Hardcode colors/spacing** — Always use design tokens
- ❌ **Design only desktop** — Mobile-first or responsive always
- ❌ **Skip component documentation** — Every component needs props/states
- ❌ **Ignore dark mode** — Colors must work in both themes
- ❌ **Forget error/loading states** — Every flow needs error handling
- ❌ **Ship without handoff doc** — Engineers need clear specs
- ❌ **Speculate on UX** — Use data/research, never assumptions
- ❌ **Create infinite variants** — Max 5-7 variants per component

## Quality Criteria

### Content Quality
- ✅ All screens visually consistent with design system
- ✅ Component naming clear and follows conventions
- ✅ Token values specific and measurable
- ✅ Interaction patterns documented
- ✅ WCAG 2.1 AA compliance verified
- ✅ Dark mode variants match light mode contrast

### Handoff Quality
- ✅ Figma → Code Connect mappings defined
- ✅ Implementation guide clear for engineers
- ✅ All design decisions documented with rationale
- ✅ Performance considerations noted
- ✅ Accessibility implementation notes included

## Success Metrics

### Design Phase Completion
- ✅ All user flows documented
- ✅ All screens designed with required states
- ✅ Component library generated or updated
- ✅ Design tokens defined and exported
- ✅ WCAG 2.1 AA audit completed
- ✅ Figma file organized for handoff

### UX Success (Measurable)
- **Time-to-task** <2 min for common flows
- **Error rates** <5% of users encounter errors
- **Conversion improvement** vs baseline (if refactoring)
- **WCAG compliance** 100% WCAG 2.1 AA
- **Component reuse** 80%+ of UI from library

## Integration with Other Phases

```
/clarify → /design (NEW) → /flow → /implement → /assess → /optimize
```

### From `/clarify`:
- Receives clear feature specifications
- Uses user personas and use cases
- Inherits acceptance criteria

### To `/flow`:
- Hands off design decisions for architecture impact
- Provides component inventory for data modeling

### To `/implement`:
- Provides Figma mockups + design spec + handoff doc
- Enables TDD implementation with clear design contracts
- Code Connect mappings for component verification

### To `/assess`:
- Design quality is part of overall assessment
- Component library reuse metrics

### To `/optimize`:
- UX optimization based on metrics
- Design system performance improvements

## Decision Support

When facing design choices, call **`@nara`**:
- Which component pattern best fits this use case?
- Should we extend the design system or create new component?
- Mobile vs desktop trade-offs (which to optimize)?
- How to make this feature accessible?

## Related Commands & Skills

```bash
/design                    # Generate UI/UX designs and specifications
/clarify                   # Clarify requirements (predecessor phase)
/flow                      # Generate architecture diagrams (successor phase)
/implement                 # Build implementation with TDD (uses design specs)
/assess                    # Assess design quality (includes design audit)
/optimize                  # Optimize UX and performance

/figma-use                 # MANDATORY: Setup Figma integration
/figma-generate-library    # Build design system in Figma
/figma-code-connect        # Map Figma components to code
```

## Example Usage

### Prototype New Feature
```markdown
/design Implement authentication flow for SaaS app
Context: 
- User personas: new users, existing users
- Key flows: signup, login, password reset
- Success metric: <1 min signup time
```

### Refactor Existing UI
```markdown
/design Audit and improve payment checkout flow
Input: Current app checkout has 25% abandonment rate
Goal: Reduce abandonment, improve clarity
```

### Build Design System
```markdown
/design Create design system for React component library
Components: Button, Input, Card, Modal, etc.
Output: Figma library + design tokens + component guide
```

## See Also

- **[Clarify Phase](./clarify.md)** — Requirements and specifications
- **[Flow Phase](./flow.md)** — Architecture and workflow diagrams
- **[Implement Phase](./implement.md)** — TDD implementation from design specs
- **[Assess Phase](./assess.md)** — Design quality assessment
- **[Plaesy Core Guidelines](../instructions/plaesy.instructions.md)** — Constitutional rules for all phases
- **[Figma Integration Guide](../../chatmodes/nara-guide.md)** — Design decision support

---

**Created**: 2026-09-06  
**Status**: Active  
**Compliance**: Plaesy Spec-Kit v0.0.1
