---
description: "Performance optimization and code refactoring orchestrator"
subagent: true
---

# `/optimize` command instructions

⚡ **Run with**: any parallel-capable mode — a host that provides
subagents, or a mode that does. Not a requirement: a single-agent host
runs this identically, just serially. `/loop` documents when fan-out is
worth its cost.

## Usage Format

```bash

/optimize                        # Optimize all dimensions active in the constitution
/optimize:technical               # Performance, code quality, database, backend/frontend
/optimize:design                  # Design tokens, visual consistency, WCAG, dark mode
/optimize:business                # Unit-economics, viability
/optimize:legal                   # Clause traceability, compliance-owner coverage
/optimize:marketing               # Claim sourcing, brand consistency
/optimize:financial               # Cost, FinOps, pricing
/optimize:product                 # Roadmap prioritization, competitive baseline
/optimize:management              # Org sustainability, process completeness
/optimize:operations              # Reliability/SLO, release cadence, toil

# Narrow to one sub-area within a dimension
/optimize:technical --focus performance   # Just the Backend/Database performance priorities
/optimize:technical --focus database      # Just database optimization
/optimize:design --focus design-system    # Just component consolidation/tokens
```

Same 9-dimension scope set as `/assess:{scope}` (see `/assess`) — `technical`,
`design`, `business`, `marketing`, `legal`, `financial`, `product`, `management`,
and `operations` — `--focus` narrows to a specific sub-area within the chosen
dimension; it never selects the dimension itself.

## Objective

Comprehensive deliverable optimization across whatever dimension(s) are active
in `.plaesy/memory/constitution.md`: performance/code efficiency for
`software`, unit-economics/viability for `business`, clause traceability for
`legal`, claim sourcing/brand consistency for `marketing`, sustainability for
org structure, WCAG/design-token consistency for `design`/UX. Sections marked
"software" apply only to that dimension; the redesign step (Protocol §3) and
its Design Spine already cover every dimension generically.

**See also**: `.plaesy/instructions/performance-baseline.md` → baseline measurement, `.plaesy/instructions/performance-optimization.md` → detailed techniques. Both are in `mapping.json`'s
`scope_load`, so `plaesy init` installs them. If either is missing, run `plaesy stack detect --install` before skipping the step — the baseline step is step 1 and the command cannot start without it.

**No constitution / not a project — one standalone artifact instead** (a document,
deck, spreadsheet, design file with no known performance/quality target to hit)?
This command isn't the right entry point — use `/improve` (Artifact Mode) instead;
it doesn't require a project or constitution.

## Pre-Optimization Validation

✅ **MUST pass before starting**:

- Code passes quality gates (assess Mode 2 score ≥ the remediation floor in
  `.plaesy/instructions/quality-gates.md`; default 85 — optimizing below that
  is fixing, so `/fix` and `/loop` own the work first)
- Tests passing (all tests green)
- Baseline metrics recorded (before optimization)
- Production traffic representable in tests

## Protocol

**Optimize ALL aspects systematically**:

1. **Baseline Analysis** - Establish current metrics (see `.plaesy/instructions/performance-baseline.md`)
2. **Bottleneck Identification** - Profile to find hot paths
3. **Design Review/Redesign** - Any dimension whose Mode 2 mandatory audit failed
   (UI/UX, architecture, business model, org structure, process — see `/assess`'s
   Design Spine) gets refactored/redesigned here against that dimension's spine and
   audit criteria, then re-audited before this phase reports complete
4. **Code Quality** - Refactor for efficiency, readability, maintainability
5. **Prioritization** - Rank optimizations by impact vs effort
6. **Implementation** - Apply incrementally with validation
7. **Verification** - Measure improvements, validate no regressions

**MANDATORY NEXT STEP**: After `/optimize` completes, run `/assess` to verify all improvements + check for regressions

## Performance Targets (Software Dimension)

| Metric | Target | Context |
|--------|--------|---------|
| Response Time | <200ms | APIs, server responses |
| Throughput | >1000 req/s | Request handling |
| Memory Usage | <512MB | Baseline footprint |
| CPU Utilization | <70% avg | Under normal load |
| Database Query | <100ms | Avg query time |

## Optimization Targets by Other Dimension

| Dimension | Target | Context |
|-----------|--------|---------|
| Business | CAC payback period within constitution threshold; every unit-economics assumption sourced | Model canvas / business case |
| Legal | 100% of clauses/claims traceable to a source; zero unassigned compliance owners | Contract / compliance doc set |
| Marketing | 100% of claims sourced; brand voice/positioning consistent across assets | Campaign / content set |
| Design | WCAG 2.1 AA pass; zero hardcoded values outside the design-token set | UI/UX, brand system |
| Management/Org | Span of control ≤6; no single point of failure in org chart | Org structure, process docs |
| Product | Every roadmap item traceable to a validated user need; no unsourced priority claim | Roadmap / prioritization set |
| Financial | Every cost/price figure sourced; no vendor or FinOps commitment without a review date | Cost model / budget set |
| Operations | Every release has a rollback path; every service has an alert and a runbook | Deploy pipeline / on-call docs |

## Optimization Standards

- **Measure the baseline first** — an optimization without one is a guess
- **Test every path after optimizing**, not just the one you changed
- **Keep correctness ahead of speed** — a behaviour change is a separate decision
  with its own justification
- **State the metric and the before/after for every optimization**, so the gain is
  a measurement rather than a claim

## Optimization Priorities (Ranked by Impact)

*Sections below are software-dimension specifics. For other active dimensions,
rank by the same impact-vs-effort logic against the table above — e.g. an
unsourced pricing claim used in three places (high impact, low effort)
outranks a minor copy-edit (low impact, low effort).*

### Backend Performance (Top Priority)

1. **Database Optimization** - Indexes, query optimization, N+1 fixes, connection pooling
2. **Caching Strategy** - Query caching, result caching, multi-level caching (memory → Redis → CDN)
3. **Algorithm Efficiency** - Complexity reduction, better data structures
4. **Async Processing** - Non-blocking I/O, batch operations, queue-based processing
5. **Memory Management** - Object pooling, stream processing, GC tuning

### Frontend Performance (If UI present)

1. **Component Performance** - Memoization, lazy loading, code-splitting, reduce re-renders
2. **Asset Optimization** - WebP/AVIF images, responsive sizing, lazy loading
3. **Animation Performance** - Target 60fps, GPU acceleration, respect prefers-reduced-motion
4. **Bundle Optimization** - Tree-shaking, minification, compression (Gzip/Brotli)
5. **Design System** - Component reuse, consolidate duplicate components, design tokens

### Design Optimization (If UI/Frontend)

1. **Design Tokens** - Standardize colors, spacing, typography
2. **Visual Consistency** - Ensure consistent patterns across UI
3. **Accessibility Performance** - WCAG AA compliance, keyboard navigation, screen reader
4. **Dark Mode** - Optimize contrast, ensure theme parity

### Code Quality (All Projects)

1. **Refactoring** - Simplify complex functions, improve readability
2. **Duplication Removal** - Consolidate similar code patterns
3. **Type Safety** - Add type annotations, reduce any types
4. **Error Handling** - Improve error messages, edge case handling
5. **Testing** - Increase coverage, improve test quality

## Impact vs Effort Matrix

| Impact | Effort | Priority | Examples |
|--------|--------|----------|----------|
| High | Low | **Immediate** | Add indexes, cache queries, consolidate components |
| High | High | Next Sprint | Architecture refactor, redesign system |
| Low | Low | Quick Wins | Code cleanup, minor UI tweaks |
| Low | High | Defer/Skip | Premature optimization, edge case polish |

## Techniques by Technology

For detailed optimization techniques per technology stack, see `.plaesy/instructions/performance-optimization.md`, which covers:

- **Frontend**: React/Vue/Angular, CSS, JavaScript bundle optimization
- **Backend**: Node.js, Python, Java, Go, C#, Ruby performance tuning
- **Database**: SQL optimization, query planning, indexing strategies
- **Network**: HTTP/2, compression, CDN strategies, serialization

## Implementation Strategy

### Phase Flow

```text
1. Baseline (`.plaesy/instructions/performance-baseline.md`)
   ↓
2. Analyze & Prioritize (impact vs effort matrix)
   ↓
3. Implement Incrementally (test each optimization)
   ↓
4. Validate (measure against baseline, check regressions)
   ↓
5. Deploy Gradually (no big bang releases)
   ↓
6. Assess & Verify (/assess to confirm no regressions)
```

### Risk Mitigation

- **Measure before/after** - Establish baseline metrics first
- **Test thoroughly** - Unit, integration, performance tests required
- **Deploy gradually** - Roll out incrementally, monitor closely
- **Have a rollback plan** - Ready to revert if regressions appear
- **Document decisions** - Why each optimization was chosen

## Success Criteria (VERIFY Before Completing)

✅ **Baseline Recorded**:

- [ ] Before measurements taken WITH TIMESTAMPS
- [ ] Tool outputs captured (not estimates)
- [ ] Same conditions as post-optimization measurements

✅ **Performance Improvements Measurable**:

- [ ] Post-optimization metrics collected WITH SAME TOOLS
- [ ] % improvement calculated: (before - after) / before × 100
- [ ] Target improvement reached (>10% OR no regression <5%)
- [ ] Specific metrics: [Response time %, Throughput %, Memory %, CPU %]

✅ **No Regressions**:

- [ ] All tests passing (full test suite, not partial)
- [ ] No new errors in logs
- [ ] Functionality unchanged (smoke test performed)
- [ ] Optimization only improved target area

✅ **Code Quality**:

- [ ] Type safety maintained (no new 'any' types)
- [ ] Complexity not increased
- [ ] Test coverage ≥{{TEST_COVERAGE_MIN|90}} target, or ≥ the pre-optimization baseline if the project was already below it (no reduction either way)

✅ **Self-Audit Checklist**:

- [ ] Have I verified every performance assumption?
- [ ] Could measurements be misinterpreted? (clarified if yes)
- [ ] Are regressions truly eliminated?
- [ ] Would I confidently deploy to production?

## Completion Report Format

✅ **Optimization Complete**

- **Baseline**: [metrics, timestamp, tool]
- **Result**: [metrics, timestamp, tool]
- **Improvement**: [% change, calculation shown]
- **Performance improvements**: [before → after, % change]
- **Code improvements**: [refactoring, duplication, type safety]
- **Design improvements**: [if UI: tokens, consistency, accessibility]
- **Regressions**: None detected ✓ (test results attached)
- **Confidence**: HIGH (verified with [specific evidence])

### Example Report

```text

✅ Optimization Complete (45 minutes)

📊 Metrics:
- API Response: 850ms → 240ms (↓72%)
- Throughput: 800 → 2400 req/s (↑200%)
- Memory: 580MB → 320MB (↓45%)

🔧 Changes:
✓ Database: 3 indexes added, 2 N+1 queries fixed
✓ Caching: Redis layer + query result cache
✓ Algorithm: Sort O(n²) → O(n log n)
✓ Async: 4 blocking calls → async/await

📈 Impact:
- 15 slow endpoints now < 200ms
- Memory footprint down 45%
- Peak CPU usage down 33%
```

## Autonomous Routing (After Optimization)

**See also**: [Universal Autonomous Routing](.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory)

After `/optimize` completes, `/assess` (verification mode) is MANDATORY and
cannot be skipped — it gates deployment readiness. Regressions → revert, try
an alternative approach, re-verify; no regressions → loop back to `/optimize`
if opportunities remain, else route to `/fix`/`/implement`, else proceed to
`/doc`.

### Dimension-Based Routing After Verification

**Routing**: per-dimension routing after verification and the severity→priority
order are in `.plaesy/instructions/dimension-mapping.md`. Look the dimension up
there; do not restate the table here.

What applies to **every** dimension, regardless of which row you land on:

- **Target met, no regression** → ready to proceed to the next phase.
- **Target missed** → the dimension's own optimize command, narrowed to the
  sub-area that missed (`--focus performance`, `--focus cost`, `--focus pricing`,
  …). One more pass, not a new dimension.
- **Regression introduced by the optimization** → the dimension's error-recovery
  command, narrowed to the regression. For the technical dimension that is
  `/fix:technical --focus regression` (broken test/behavior) or
  `/fix:technical --focus performance` (revert + alternative approach); for any
  other dimension it is that dimension's `/fix` cell. A failed verification blocks
  progression until cleared.
- **Accessibility gap that turns out to be a code defect** rather than a design
  one → `/implement:technical --focus wcag` (the Design row's implement cell), not
  another optimize pass.

**Role handoff**: once the target is met with no regression, look up
`.plaesy/instructions/role-mapping.md`'s Routing Table for the optimizing
role's "Typical next role," write the handoff to `state.json.role_handoff_log`,
and invoke `/continue` rather than stopping. No hop limit; a hard-stop category
(`.plaesy/instructions/plaesy.md` rule 8) is the only thing that ends the
chain early.

---

## Error Recovery

**If optimization causes regression:**

1. Identify which optimization caused it (binary search)
2. Revert the problematic change
3. Document why it failed (may need a different approach)
4. Try an alternative optimization for the same bottleneck

**If metrics are unavailable:**

- Use profiling tools: Python cProfile, Node.js clinic, Go pprof, Chrome DevTools
- Report baseline from profiler output, not live metrics

## Detailed Technique Reference

For comprehensive optimization techniques, examples, and language-specific approaches:
→ See `.plaesy/instructions/performance-optimization.md` (platform-specific best practices)
→ See `.plaesy/instructions/performance-baseline.md` (measurement framework & tools)

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md` →
[Global Routing](.plaesy/instructions/plaesy.md#-universal-autonomous-routing-global-mandatory)
