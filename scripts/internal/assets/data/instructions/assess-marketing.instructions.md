---
description: "Marketing assessment workflow - positioning, messaging, audience fit, go-to-market readiness"
applyTo: "**/*"
---

# Marketing Assessment Instructions

**Use with**: `/assess:marketing`

**Referenced from**: `/assess` (orchestrator), `.plaesy/instructions/dimension-mapping.md`

**Delegates to**: `.plaesy/roles/market-research-analyst.md`, `.plaesy/roles/pm.md`

---

## Assessment Workflow

### Step 1: Positioning Analysis

- Identify the stated (or inferable) market category and competitive frame
- Check positioning is specific to a target segment, not generically "for everyone"
- Validate the positioning is defensible (a competitor can't trivially claim the same statement)
- Compare against 2-3 direct competitors' public positioning

### Step 2: Messaging & Value Proposition

- Extract the core value proposition — does it lead with an outcome/benefit, or a feature list?
- Check message clarity: could a target user restate the value prop in their own words after one read?
- Identify proof points backing each claim (data, case study, benchmark) vs unsupported assertions
- Flag jargon or internal terminology that would not resonate with the target audience

### Step 3: Audience & Segmentation

- Confirm target audience is defined with enough specificity to guide channel/content choices (role, company size, pain point — not just "developers" or "businesses")
- Check messaging is segmented if multiple audiences exist (e.g., technical buyer vs economic buyer)
- Validate audience assumptions against any available usage/signup data

### Step 4: Competitive Messaging Analysis

- Map how 2-3 competitors position themselves on the same axes (price, ease of use, depth, ecosystem)
- Identify messaging gaps/whitespace not being claimed by competitors
- Flag if current messaging directly overlaps with a stronger-resourced competitor's claim

### Step 5: Go-to-Market Readiness

- Assess channel strategy fit (does the chosen channel match where the target audience actually is?)
- Review launch plan completeness (pre-launch, launch, post-launch milestones)
- Check content/collateral readiness (landing page, docs, demo, case studies) against launch stage
- Validate a feedback loop exists to measure message resonance post-launch (not just publish-and-hope)

### Step 6: Generate Report & Recommendations

**Report sections**: canonical names per `.plaesy/instructions/dimension-mapping.md`
→ *Assessment Report Section Registry*; the bullets below go *inside* those
sections and do not rename them.

- Overall marketing readiness score (0-100)
- Per-category scores (Positioning, Messaging, Audience, Competitive, GTM Readiness)
- Top 5 findings (ordered by impact on conversion/resonance)
- Ranked options with tradeoffs

---

## Quality Scoring

```text

positioning:      25%   # Category clarity, defensibility, competitive differentiation
messaging:         25%   # Value prop clarity, proof points, jargon-free
audience:           20%   # Segmentation specificity, validated assumptions
competitive:        15%   # Whitespace identified, no direct overlap with stronger competitor
gtm_readiness:      15%   # Channel fit, launch plan completeness, feedback loop
```

### Grade Mapping

| Grade | What it means here |
|---|---|
| A+ | Exceptional - Sharp positioning, validated messaging, launch-ready |
| A | Excellent - Clear differentiation, minor messaging polish needed |
| B+ | Very Good - Solid positioning, some audience/channel gaps |
| B | Good - Viable but generic in places, address before launch |
| C | Acceptable - Positioning unclear or unvalidated, needs research |
| D | Needs Work - No clear differentiation or audience definition, high launch risk |

The score edges behind these letters, and the general meaning of each band,
are defined once in `.plaesy/instructions/quality-gates.md` →
**Grade bands**. Do not restate them here.

### Success Criteria (Hard Stops)

- Must have: Target audience defined with enough specificity to pick a channel
- Must have: Value proposition stated in outcome/benefit terms, not just feature list
- Must have: At least one point of competitive differentiation identified
- Must have: Positioning does not directly collide with a dominant competitor's claim without a stated wedge

---

## Critical Rules

- ✅ **AUDIENCE-FIRST** — every messaging judgment is made from the target audience's vantage point, not the builder's
- ✅ **EVIDENCE OVER OPINION** — positioning claims backed by competitor research or user feedback, not taste
- ✅ **CITE SOURCES** — competitor claims cited with URL/date retrieved (see Web Research Requirement below)
- ✅ **ONE CORE MESSAGE** — flag messaging that tries to say too many things at once
- ✅ **DEFENSIBILITY CHECK** — explicitly test "could our strongest competitor say this exact sentence?"

---

## Finding Standards

- **Verify "unique" positioning against live competitor research**, not memory
- **Translate every feature into an outcome or benefit** before calling it a value
  proposition
- **Force a segmentation** — "everyone" is not an audience
- **Name the proof point** that supports each recommended message
- **Check channel-audience fit** — a correct message in the wrong channel fails
  just the same
- **Assess pre-launch and post-launch phases separately**

---

## Web Research Requirement (Mandatory)

Competitive and market claims in this assessment **must** be backed by a live source per
the Mode 1 anti-hallucination protocol in `/assess`:

1. Competitor positioning/pricing/feature claims → `WebSearch`/`WebFetch` against the competitor's own site or recent reviews, not model recall
2. Cite every competitive claim inline: `[claim] — Source: [name/URL], retrieved {{CURRENT_DATE}}`
3. If research is unavailable, label the finding SPECULATIVE rather than presenting it as researched

---

## Pre-Assessment Validation

```text

Is target audience defined?
  → yes: Validate specificity and segmentation
  → no: Flag as blocking finding, do not proceed to messaging validation

Is competitor information available?
  → yes: Run live web research, compare positioning
  → no: Run web research to establish baseline before assessing differentiation

Is usage/signup data available to validate audience assumptions?
  → yes: Cross-check assumed audience against actual data
  → no: Flag audience definition as ASSUMED, not validated
```

---

## Uncertainty Surfacing (Before Finalizing Assessment)

Load `.plaesy/instructions/uncertainty-surfacing.md`; do not restate it here.
Only this dimension's specifics follow.

**Confidence basis for this dimension**:

- HIGH: live competitor research with a citation, plus actual audience or usage data
- MEDIUM: live competitor research, but the audience itself is assumed
- LOW: no external validation; positioning inferred from the product only

**Worked example**:

```text
I could not validate the "fastest onboarding in the category" claim because
no competitor pricing pages or usage data were available.
To improve confidence, provide a signup-source breakdown, or three competitor
onboarding walkthroughs.
```

---

## Assessment Completion Checklist

Assessment is complete when:

- ✅ Positioning analyzed against target category and competitors (live research cited)
- ✅ Value proposition validated for clarity and proof points
- ✅ Audience segmentation specificity checked
- ✅ Competitive messaging whitespace identified
- ✅ GTM readiness (channel, launch plan, feedback loop) evaluated
- ✅ Scores calculated per category
- ✅ Top 5 findings with sources listed
- ✅ Next phase recommended (`/implement --marketing`, `/optimize --positioning`, or `/assess:marketing — Mode 1 if data is insufficient)
- ✅ Console output delivered to user

## Persona Protocol

Loaded by every `each family:marketing` persona prompt for this dimension — `/assess`,
`/implement`, `/fix`, `/optimize`, `/loop` and `/improve` all point their
`Loads:` line here. The rules below are dimension-neutral and are written here
once rather than restated in 54 persona files, where they had already drifted
apart from the parents they point at.

- **Follow the parent protocol.** The persona fixes the *dimension*; the family
  prompt fixes the *verb*. Follow the family prompt's full protocol. Where the
  parent and this file disagree, the parent governs the behaviour.
- **This file is the *assessment* criteria for `marketing`** — workflow, weights,
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
