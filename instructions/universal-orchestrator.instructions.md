---
description: "Universal orchestrator - guide spec-kit for ANY dimension (technical, design, business, product, marketing, operations, legal, financial)"
applyTo: "**/*"
---

# Universal Orchestrator: Beyond Programming

⚡ **Framework Philosophy**: Spec-Kit is NOT just for code. It's for **ANY project need**.

---

## Spec-Kit Dimensions (9 Total)

Technical, Design, Business, Product, Marketing, Management, Operations,
Legal/Compliance, Financial. What each one covers, and which command handles
it, is the **Scope** column of the Routing Table in
`.plaesy/instructions/dimension-mapping.md` — not restated here to avoid two
sources of truth for the same list.

The point of this file is the *orchestration* layer above that table: a
request rarely belongs to one dimension cleanly, so the sections below cover
what happens at the seams — multiple dimensions assessed together, and
findings handed off between them.

---

## Autonomous Routing Across Dimensions

**Canonical routing table**: `.plaesy/instructions/dimension-mapping.md` — which
command each dimension routes to (assess / implement / optimize / error recovery /
loop), the severity→priority order, and the inventory of every shipped
sub-command. This section adds only what that table cannot express: the *order* in
which dimensions hand off to each other when a finding has no single home.

Every cross-dimension flow below follows the same three rules:

1. **A finding is owned by exactly one dimension.** Route it with the canonical
   table, not by re-deriving a dimension from a flow diagram.
2. **A hand-off is not a re-assessment.** The receiving dimension consumes the
   source dimension's output as input; only the owning dimension re-runs its own
   assess/implement cycle.
3. **Every hand-off closes with a return to the source dimension** — the loop
   never ends with an unverified finding. Verification is Phase 6 of the canonical
   9-phase model (`.plaesy/instructions/workflow-phases.md`).

Example chains (illustrative, not a routing table):

- **Technical ↔ Business ↔ Financial** — a stated concurrency target (technical)
  becomes an infrastructure-cost question (business) and a unit-economics question
  (financial), then routes back to technical, which either scales to that target or
  redesigns for a smaller one.
- **Design ↔ Marketing** — a finished design is checked against brand positioning
  (marketing); on a mismatch, design refinement and marketing re-check each other
  until aligned.
- **Legal → Operations → Technical** — a compliance gap (legal) becomes a process
  and tooling change (operations, technical), then re-enters legal for
  verification. Compliance is the CRITICAL, blocking leg of this chain.
- **Financial → Product ↔ Design/Marketing → Operations** — a CAC/LTV target
  (financial) becomes a feature-priority question (product), a messaging question
  (design, marketing) and a capacity question (operations).

---

## Workflow Commands Across Dimensions

`/start` auto-assesses whichever dimensions a request touches (not just
`technical`), then coordinates implementation across them — a marketing
campaign request pulls in Marketing/Business/Financial/Product, not just
code:

```bash

/start Launch AI-powered e-commerce platform
  → Auto-assess all 9 dimensions → integrated plan → coordinated implementation

/start Launch marketing campaign for Q4
  → Auto-assess: Marketing, Business, Financial, Product → campaign strategy + budget
```

`/assess`, `/implement`, `/optimize`, `/fix`, `/loop` all take the same
`{command}:{dimension}` selector over the same 9 dimensions, narrowed with
`--focus {sub-area}` — never a bare `--backend`/`--design`/`--bug` flag. The
canonical mapping (which dimension routes to which command, in which mode)
is `.plaesy/instructions/dimension-mapping.md`; this file does not restate
it. One illustrative selector per command is enough to show the form:

```bash

/assess:business,marketing          # multi-dimensional in one call
/implement:technical --focus wcag   # dimension + sub-area, never combined into one flag
/optimize:financial --focus cost
/fix:legal
/loop:management --focus process
```

`/assess` also runs in Research Mode (pre-spec, pulls external evidence) and
Mode 1 (resolves a named ambiguity) — same `{scope}` selector, different
trigger; see `/assess` for the mode definitions.

---

## Conflict Resolution Across Dimensions

Two dimensions' findings compete (e.g. Technical wants a database refactor,
Business wants to enter a market now). Resolve with the same two facts every
time: which side is time-boxed or legally non-negotiable, and whether
deferring the other side blocks something irreversible. A legal/compliance
constraint that is mandatory in the operating jurisdiction always wins over a
budget preference; everything else is a genuine trade-off to size, not an
autonomous auto-pick — surface the trade-off and its cost on each side,
default to the lower-risk option per rule 2 (`plaesy.instructions.md`
§Assistant Behavior), and record which one was chosen and why.

---

## Integration Example (All Dimensions)

```text

/start Launch AI analytics SaaS

PHASE 1 (/assess Mode 1): assess all touched dimensions
├─ /assess:technical  → "Scalability: ✅ ready"
├─ /assess:design     → "UX: ⚠️ dashboard needs accessibility work"
├─ /assess:business   → "Model: ⚠️ pricing needs validation"
└─ /assess:financial  → "Economics: ⚠️ CAC/LTV needs improvement"

PHASE 4 (/assess assessment mode): route each finding to its owning dimension
├─ Design finding   → /implement:technical --focus wcag
├─ Business finding  → /assess:business Mode 1 (revalidate pricing)
└─ Financial finding → /optimize:financial --focus cost

PHASE 2/5 (/implement, /optimize): fix in the owning dimension
PHASE 6 (/assess verification mode): each dimension re-assesses its own fix

Result: every finding owned, fixed, and verified by the dimension that owns it.
```

This is the same hand-off loop as the Autonomous Routing rules above, just
walked through one request end to end.
