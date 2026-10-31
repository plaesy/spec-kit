# Spec-Kit CLI Contract

## Overview

This document defines the observable behavior of every Spec-Kit CLI command. It serves as the **single source of truth** for:
- Command signatures (arguments, flags, env vars)
- Exit codes
- Output formats (stdout, stderr, files)
- Error behavior
- Side effects

**Rule**: The CLI behavior is the contract. Documentation that contradicts the CLI is a bug.

---

## Global Conventions

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Missing input (stage failed loudly) |
| 4 | Quality gate blocked |
| 5 | Pivot gate requires decision |
| 130 | Interrupted (SIGINT) |

### Environment Variables
| Variable | Description | Default |
|----------|-------------|---------|
| `SPEC_KIT_NICHE` | Default niche for pipeline | — |
| `SPEC_KIT_BUDGET` | Default budget (USD) | 5000 |
| `SPEC_KIT_TIMELINE` | Default timeline | "6 weeks" |
| `SPEC_KIT_WEB_BACKEND` | Web search backend | "exa" |
| `SPEC_KIT_LLM_PROVIDER` | LLM provider | "gemini" |
| `SPEC_KIT_YOUTUBE` | Enable YouTube automation | "false" |
| `SPEC_KIT_TOPIC` | Current topic | — |

### Output Conventions
- **stdout**: Machine-readable (JSON when `--json` flag)
- **stderr**: Human-readable progress, warnings, errors
- **Files**: All artifacts written to `.plaesy/ideas/<topic>/`

---

## Commands

### `ideator` — Divergent Idea Generation

```bash
ideator [OPTIONS]
```

**Options**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--niche` | string | required | Problem area (e.g., "health", "fintech") |
| `--interests` | string | "" | Your skills/interests |
| `--budget` | number | 5000 | Budget in USD |
| `--time` | string | "6 weeks" | Timeline for MVP |
| `--where` | string | "online" | Where to sell (city/online) |
| `--json` | bool | false | Output JSON to stdout |

**Behavior**
1. Asks ≤3 clarifying questions if inputs missing
2. Runs ≥3 web searches for problems people pay to solve
3. Generates 8-12 distinct options (different assumptions)
4. Each option: name, description, assumption, cheapest test, label (FACT/INFERRED/SPECULATIVE)
5. At least one contrarian option
5. Writes candidates to `.plaesy/ideas/<candidate-topic>/idea.md` with `status: candidate`
6. Updates `.plaesy/ideas.md` index

**Output (stdout, human)**
```
OPTIONS:
  1. AI Health Coach — Personalized wellness plans via AI
     assumption: Coaches want automated personalization
     cheapest test: Landing page + waitlist ($50 ads)
     label: INFERRED
  ...
STATUS: done
GAPS: []
```

**Output (stdout, `--json`)**
```json
{
  "options": [...],
  "status": "done",
  "gaps": []
}
```

**Files Created**
- `.plaesy/ideas/<candidate>/idea.md` (×8-12)
- `.plaesy/ideas.md` (index updated)

**Exit Codes**
- 0: Success
- 3: Needs escalation (brief too vague)

---

### `buyer-discovery` — Buyer Validation Sprint

```bash
buyer-discovery [OPTIONS]
```

**Options**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--topic` | string | required | Topic folder name |
| `--target` | number | 5 | Target real conversations |
| `--segment` | string | "" | Filter by segment (coach/brand/clinic) |
| `--json` | bool | false | Output JSON |

**Behavior**
1. Loads existing `.plaesy/ideas/<topic>/`
2. Runs outreach tracker with segment templates
3. Records interviews with would-pay signals
4. Runs Segment Validation Matrix
5. Outputs `buyers.md` (founder format) + `buyers-real.md`

**Quality Gates**
- **B1**: ≥5 real conversations OR verdict = Incomplete
- **B1**: ≥3 "would pay YES" OR verdict = Incomplete
- Segment kill: ≥3 convos AND would-pay rate < 30%

**Files Created/Updated**
- `.plaesy/ideas/<topic>/buyers.md` (founder format)
- `.plaesy/ideas/<topic>/buyers-real.md` (detailed records)

---

### `compliance-health` — Health Claims Validator

```bash
compliance-health [OPTIONS]
```

**Options**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--copy` | string | required | Marketing text to validate |
| `--format` | string | "json" | Output format (json/text) |
| `--update-db` | bool | false | Trigger monthly DB update |

**Behavior**
1. Runs rule engine on input text
2. Returns risk score (0-100), flags, required disclaimers
3. Auto-suggests compliant alternatives for flagged terms
4. If risk_score ≥ 30: content blocked for auto-publish

**Output (JSON)**
```json
{
  "passed": false,
  "risk_score": 75,
  "flags": [
    {"term": "cures", "position": 12, "severity": "critical", "suggestion": "supports", "regulation": "FTC/BPOM: drug claim"}
  ],
  "required_disclaimers": [
    "This statement has not been evaluated by BPOM.",
    "This product is not intended to diagnose, treat, cure, or prevent any disease."
  ]
}
```

**Quality Gate B5**
- Risk score < 30 for auto-publish
- Critical flags (cure/heal/treat/prevent) = 0 allowed

---

### `economics-engine` — Unit Economics Calculator

```bash
economics-engine [OPTIONS]
```

**Options**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--topic` | string | required | Topic folder |
| `--P` | number | — | Price per sale |
| `--V` | number | — | Direct cost per sale |
| `--F` | number | — | Fixed cost/month |
| `--S` | number | — | Expected sales/month |
| `--basis` | string | — | S basis (required) |
| `--CAC` | number | — | Customer acquisition cost |
| `--churn` | number | — | Monthly churn rate |
| `--json` | bool | false | Output JSON |

**Behavior**
1. Reads existing `pricing.md`, `offer.md`, `ops.md` if inputs not provided
2. Calculates: margin, break-even, MoS, LTV, payback, months-to-BE
3. Runs segment analysis (multi-tier)
4. Runs sensitivity analysis (price, cost, volume, CAC, F, segment-mix)
5. Applies verdict rules (Incomplete/Not yet/Profitable)
5. Writes `economics.md` + `economics-real.md`

**Verdict Rules (Blocking)**
1. **Incomplete**: Any input missing, S no basis, required stage missing
2. **Not yet**: m ≤ 0 OR S < B (name failing number)
3. **Profitable**: m > 0 AND S ≥ B AND S has basis (label estimate if weak)

**Quality Gate B2**
- All 4 inputs present with sources
- S basis explicitly stated
- Margin calc shown, break-even shown, MoS shown (if S ≥ B)

---

### `plan-compiler` — Business Plan Compiler

```bash
plan-compiler [OPTIONS]
```

**Options**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--topic` | string | required | Topic folder |
| `--json` | bool | false | Output JSON |

**Behavior**
1. Verifies all 11 required stage files exist
2. Reads verdict + numbers from `economics.md`
3. Reads board votes from `board.md`
4. Concatenates all stages in order with headers
5. Auto-generates one-page summary
6. Updates `.plaesy/ideas.md` index with verdict + date
7. If verdict = Not yet: appends Pivot Analysis section

**Files Created/Updated**
- `.plaesy/ideas/<topic>/plan.md` (complete)
- `.plaesy/ideas.md` (index updated)

---

### `assess` — Dimension Assessment

```bash
assess [OPTIONS]
```

**Options**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dimension` | string | required | business/marketing/financial/product/ops/tech/design/legal/management |
| `--topic` | string | required | Topic folder |
| `--json` | bool | false | Output JSON |

**Behavior**
1. Loads dimension reference from `spec-kit/skills/assess/references/<dimension>.md`
2. Loads all stage files from `.plaesy/ideas/<topic>/`
3. Runs assessment workflow (7 steps)
4. Scores per component weights
5. Outputs grade (A+-D), score, top 5 findings, recommendations

**Output Format (stdout)**
```markdown
# Assessment: <dimension> for <topic>

**Grade**: <A+|A|B+|B|C|D>
**Score**: <0-100>

## Component Scores
- <component>: <score>/100 (<weight>%)

## Top 5 Findings
1. <finding with source>

## Recommendations
- <action>

## Confidence Basis: HIGH|MEDIUM|LOW
## Next Phase: <command>
```

**Quality Gate B6**
- Mandatory audit items per dimension present
- Score capped at 49 if mandatory audit fails
- Output must include all canonical sections

---

### `pipeline` — Master Orchestrator

```bash
pipeline [OPTIONS]
```

**Options**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--niche` | string | — | Niche for new idea |
| `--idea` | string | — | Idea description |
| `--topic` | string | — | Existing topic to resume |
| `--resume-from` | string | "ideator" | Stage to resume from |
| `--pivot` | string | — | Chosen pivot (b2b-only/hybrid/coach-volume/partnership) |
| `--json` | bool | false | Output JSON |

**Behavior**
1. If `--niche` + `--idea`: runs full 15 stages from Stage 1
2. If `--topic` + `--resume-from`: resumes from specified stage
3. If `--pivot`: applies pivot decision, rebuilds from Stage 10
4. Runs all stages with proper handoffs
5. Stage 9 (economics) produces verdict
6. If verdict = Not yet: **Pivot Decision Gate** (blocking)
   - Runs segment analysis + validation matrix
   - Presents 4 options with projected verdicts
   - **Requires user decision** before Stage 10
   - Documents in `plan.md` → `## Pivot Analysis`
7. Stage 16 (optional): `youtube-automation` after marketing/launch

**Stages**
```
1. ideator (if no idea)
2. founder: idea.md
3. founder: board.md
4. founder: competitors.md
5. buyer-discovery (≥5 real convos)
6. founder: pricing.md
7. founder: offer.md
8. compliance-health (if health niche)
9. economics-engine → VERDICT
   ├── Profitable → continue
   └── Not yet → PIVOT GATE (blocking)
10. founder: marketing.md
11. founder: brand.md
12. founder: ops.md
13. founder: launch.md
14. plan-compiler → plan.md
15. assess (optional) → grades
16. youtube-automation (optional) → video assets
```

**Quality Gates**
- **S1-S8**: Stage-specific gates (see quality-gates.md)
- **B1-B7**: Business validation gates
- **B3 Pivot Gate**: Blocks if Not yet without decision

**Exit Codes**
- 0: Complete (Profitable or documented pivot)
- 3: Pivot gate requires decision
- 4: Quality gate blocked

---

### `youtube-automation` — Video Creation

```bash
youtube-automation [OPTIONS]
```

**Options**
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--topic` | string | required | Topic for video |
| `--format` | string | "long" | "long" (16:9) or "short" (9:16) |
| `--niche` | string | — | Niche for research context |
| `--json` | bool | false | Output JSON |

**Behavior**
1. Research Agent: Trends + Wikipedia + arXiv + HN → brief
2. Fact-Check Agent: Wikipedia + arXiv grounding → sources
3. Script Agent: Hook + Body + CTA + chapters (long) / punchy (Shorts)
4. Visuals Agent: Pexels→Pixabay→NASA→Openverse→AI (fail-soft)
5. Voice Agent: Edge TTS + SRT captions (sync <500ms)
4. Render Agent: MoviePy → MP4 + thumbnail + SRT
5. Review Agent: Title Lab (6 options), Findability Score (11 checks), Thumbnail Designer (4 styles)
5. Output: `output/library/<video-id>/` with all assets

**Quality Gate B7 (Blocking)**
- Research sources ≥3 with URLs
- All claims grounded or marked assumption
- Script: hook + body + CTA
- Visuals ≥5 with credits
- Voice SRT synced (<500ms drift)
- Render: MP4 valid, duration ±10%
- Findability Score ≥80
- Thumbnail <2MB, correct ratio (16:9/9:16)
- At least 1 thumbnail style generated

**Output Structure**
```
output/library/<video-id>/
├── video.mp4
├── thumbnail.jpg
├── captions.srt
├── script.md
├── sources.json
├── metadata.json
└── findability_report.md
```

---

## Error Behavior

### Missing Input (Stage Fails Loudly)
```
Error: Stage 5 (buyers) requires ≥5 real conversations
Current: 0 real, 5 simulated

Options:
1. Provide real conversation data
2. Record as UNKNOWN, continue (verdict = Incomplete)
3. Ask ≤3 clarifying questions

Questions:
1. Can you share 3 real conversation records?
2. What's your outreach method?
3. What's your current would-pay rate?

Exit code: 3
```

### Pivot Gate Blocking
```
🟣 PIVOT DECISION REQUIRED
Economics verdict: Not yet
Failing number: S=18 < B=32 (MoS -77.8%)

Pivot Options:
| Option | Segments | P | F | Target S | B | MoS | Verdict |
| B2B Only | Brands, Clinics, Apps | $250 | $2,000 | 10 | 10 | 0% | Profitable |
| Hybrid | All, diff pricing | $140 | $2,200 | 20 | 19 | 5% | Profitable |

Required: Choose option → Document in plan.md → Continue
Run: pipeline --topic ai-health-influencer --pivot hybrid

Exit code: 5
```

### Quality Gate Blocked
```
❌ QUALITY GATE BLOCKED: B7 YouTube Automation
Findability Score: 72/100 (< 80 threshold)

Failed checks:
- End screen elements missing
- Cards/pinned comment missing

Fix and re-run: youtube-automation --topic "AI tools" --format long

Exit code: 4
```

---

## File Output Contracts

All commands write to `.plaesy/ideas/<topic>/` with these guarantees:

| File | Created By | Required For |
|------|------------|--------------|
| `idea.md` | ideator / founder | Stage 2+ |
| `board.md` | founder | Stage 3 |
| `competitors.md` | founder | Stage 4 |
| `buyers.md` | buyer-discovery | Stage 5, 6, 7, 9 |
| `buyers-real.md` | buyer-discovery | Audit |
| `pricing.md` | founder | Stage 6, 7, 9 |
| `offer.md` | founder | Stage 7, 9 |
| `economics.md` | economics-engine | Stage 9, 14 |
| `economics-real.md` | economics-engine | Audit |
| `marketing.md` | founder | Stage 10, 13, 16 |
| `brand.md` | founder | Stage 11 |
| `ops.md` | founder | Stage 12, 9 |
| `launch.md` | founder | Stage 13, 16 |
| `plan.md` | plan-compiler | Final |
| `assessment-*.md` | assess | Audit |
| `compliance-simulation.md` | compliance-health | Audit |

---

## Versioning

This contract is versioned with the spec-kit release. Breaking changes require major version bump.

**Current Version**: 1.0.0
**Schema**: Commands follow semantic versioning for flags/outputs.