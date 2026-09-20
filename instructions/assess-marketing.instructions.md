---
description: "Marketing assessment workflow - positioning, messaging, audience fit, go-to-market readiness"
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

- Overall marketing readiness score (0-100)
- Per-category scores (Positioning, Messaging, Audience, Competitive, GTM Readiness)
- Top 5 findings (ordered by impact on conversion/resonance)
- Ranked options with tradeoffs

---

## Quality Scoring

```
positioning:      25%   # Category clarity, defensibility, competitive differentiation
messaging:         25%   # Value prop clarity, proof points, jargon-free
audience:           20%   # Segmentation specificity, validated assumptions
competitive:        15%   # Whitespace identified, no direct overlap with stronger competitor
gtm_readiness:      15%   # Channel fit, launch plan completeness, feedback loop
```

### Grade Mapping

```
95-100: A+ (Exceptional - Sharp positioning, validated messaging, launch-ready)
90-94:  A  (Excellent - Clear differentiation, minor messaging polish needed)
85-89:  B+ (Very Good - Solid positioning, some audience/channel gaps)
80-84:  B  (Good - Viable but generic in places, address before launch)
70-79:  C  (Acceptable - Positioning unclear or unvalidated, needs research)
<70:    D  (Needs Work - No clear differentiation or audience definition, high launch risk)
```

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

## Anti-Patterns (NEVER Do These)

- ❌ **Never approve positioning as "unique" without checking competitors** — verify via live research, not memory
- ❌ **Never accept a feature list as a value proposition** — always translate to outcome/benefit
- ❌ **Never assume audience is "everyone"** — force segmentation
- ❌ **Never recommend a message without stating what proof point supports it**
- ❌ **Never ignore channel-audience mismatch** — a technically correct message in the wrong channel still fails
- ❌ **Never treat launch as a single event** — assess pre/post-launch phases separately

---

## Web Research Requirement (Mandatory)

Competitive and market claims in this assessment **must** be backed by a live source per
the Mode 1 anti-hallucination protocol in `/assess`:

1. Competitor positioning/pricing/feature claims → `WebSearch`/`WebFetch` against the competitor's own site or recent reviews, not model recall
2. Cite every competitive claim inline: `[claim] — Source: [name/URL], retrieved {{CURRENT_DATE}}`
3. If research is unavailable, label the finding SPECULATIVE rather than presenting it as researched

---

## Pre-Assessment Validation

```
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

**MUST state explicitly**:

1. **Confidence Levels**:
   - HIGH: Based on live competitor research + actual audience/usage data
   - MEDIUM: Based on live competitor research but assumed audience
   - LOW: Based on no external validation, positioning inferred from product only

2. **Assumptions Made**:
   - "I assumed the target audience is [X] because [no segmentation data was provided]"
   - "If [assumption] is false, [positioning recommendation] changes to [alternative]"

3. **Missing Information**:
   - "I could not validate [claim] because [no competitor data / no usage data available]"
   - "To improve confidence, provide [specific data: signup source breakdown, customer interview notes]"

4. **Speculative vs Confirmed**:
   - CONFIRMED: [Based on live competitor research with citation]
   - LIKELY: [Pattern matches across similar products, not directly verified]
   - SPECULATIVE: [Needs validation before acting]

---

## Success Criteria

Assessment is complete when:

- ✅ Positioning analyzed against target category and competitors (live research cited)
- ✅ Value proposition validated for clarity and proof points
- ✅ Audience segmentation specificity checked
- ✅ Competitive messaging whitespace identified
- ✅ GTM readiness (channel, launch plan, feedback loop) evaluated
- ✅ Scores calculated per category
- ✅ Top 5 findings with sources listed
- ✅ Next phase recommended (`/implement --marketing`, `/optimize --positioning`, or `/assess:marketing --research` if data is insufficient)
- ✅ Console output delivered to user
