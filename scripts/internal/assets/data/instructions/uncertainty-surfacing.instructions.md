---
description: "Single source for the Uncertainty Surfacing protocol - the four disclosures an assessment must make before it is called complete, the confidence ladder, and the sentence forms for stating an assumption or a gap"
applyTo: "**/*"
---

# Uncertainty Surfacing (Before Finalizing Assessment)

An assessment that does not say what it does not know is indistinguishable from
one that is confident. This file is the single source for the four disclosures
every assessment makes before it is finalized. The per-dimension
`assess-{dimension}.instructions.md` files carry only what is *specific to that
dimension* — which evidence earns which confidence level, and one worked
example — and point here for everything else.

**The score is not the disclosure.** A dimension scores what it could measure.
An assessment with a HIGH score and three unresolved assumptions has still not
told the reader whether to trust it. Both are required, and neither substitutes
for the other.

## The Four Disclosures

State all four, in order, before the score is presented as final.

### 1. Confidence Level

One level per category assessed, with the evidence that earned it. Confidence is
a property of the *evidence*, not of the finding's severity or the assessor's
tone.

| Level | Earned by |
|-------|-----------|
| **HIGH** | Direct evidence: source read, test run, metric measured — the thing itself was inspected |
| **MEDIUM** | Indirect evidence: pattern match, inference from a comparable, partial data where some was read |
| **LOW** | No direct evidence: specification gap, missing input, or the category could not be examined at all |

Each dimension's own file states what counts as direct evidence *for it* — a
code review is direct evidence for a technical finding, a cited competitor page
is direct evidence for a marketing claim, and neither is evidence for the other.

A LOW confidence is a legitimate result and must be reported as one. Padding a
LOW to MEDIUM to make a report look decisive is the failure this section exists
to prevent.

### 2. Assumptions Made

Every assumption that a finding depends on, in this exact sentence form:

```text
I assumed [X] because [Y was not provided].
If [assumption] is false, [finding] changes to [alternative].
```

The second sentence is the part that gets skipped, and it is the part that
matters: it tells the reader what the finding is worth if the assumption breaks.
An assumption with no stated consequence is a caveat, not an assumption — either
give the alternative reading or drop it.

### 3. Missing Information

Every area that could not be assessed, in this form:

```text
I could not assess [area] because [specific data unavailable].
To improve confidence, provide [specific evidence].
```

Name the artefact that would close the gap — a file, a metric, an access grant,
an answer. "Insufficient information" is not a reason; it is a restatement of the
gap. A dimension with a conditional step that did not apply reports it as `N/A`
(see the section registry in `.plaesy/instructions/dimension-mapping.md`), which is
different from "could not assess" and must not be blurred into it.

### 4. Speculative vs Confirmed

Tag every finding, so a reader can filter on evidence quality without reading
the reasoning:

| Tag | Meaning |
|-----|---------|
| **CONFIRMED** | Direct evidence of the specific problem, at the specific location |
| **LIKELY** | A pattern matches, but the specific instance was not observed |
| **POSSIBLE** | Could occur **if** a stated condition holds — the condition is named |
| **SPECULATIVE** | Plausible, unverified, and not yet actionable |

A POSSIBLE finding without its condition stated is a SPECULATIVE one wearing a
labelling that promises specificity it does not have.

## Where It Goes in the Report

Uncertainty goes in the `## Uncertainty` section — a permitted appended section
under the Assessment Report Section Registry in
`.plaesy/instructions/dimension-mapping.md`. It is **never** folded into the
score and never netted against a finding's severity. An assessment that reports
`82/100` and hides three LOW-confidence categories behind that number has
misrepresented its result, and this section is what stops it.

## What This Is Not

- **Not a confidence-damped verdict.** Surfacing uncertainty adds disclosure; it
  does not mean hedging every finding. A CONFIRMED critical defect is stated as
  a critical defect, with its evidence, and separately reports that some other
  category went unexamined.
- **Not a licence to skip work.** "I could not assess X" is a reason to say what
  would let you assess X, not a reason to not have looked. If the artefact was
  available and unread, the finding is CONFIRMED that the check was not done.
- **Not dimension-specific.** The four disclosures, the ladder, and the two
  sentence forms are identical across all nine dimensions. Only the evidence
  that earns a level, and the worked example, differ — and those stay with the
  dimension.
