# Spec-Kit Architecture

## Overview

Spec-Kit is a **business validation pipeline** that takes a niche/idea through 15 stages to produce a make-money verdict (Profitable / Not yet / Incomplete). It includes an optional 16th stage for YouTube video automation.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        User Request                              │
└──────────────────────────┬──────────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    using-plaesy (Bootstrap)                     │
│  • Routes to correct skill based on task                        │
│  • Loads spec-kit pipeline stages                               │
└──────────────────────────┬──────────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                    spec-kit-pipeline (Orchestrator)             │
│  15 stages + optional 16th (YouTube)                            │
│  • Manages handoffs between stages                              │
│  • Pivot Decision Gate after Stage 9                            │
│  • Resume-from-stage support                                    │
└──────────────────────────┬──────────────────────────────────────┘
                           ▼
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│  Research &   │  │  Implement    │  │  Document &   │
│  Validation   │  │  (Stages      │  │  Verify       │
│  (Stages 1-9) │  │  10-13)       │  │  (Stages      │
└───────────────┘  └───────────────┘  └───────────────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│              .plaesy/ideas/<topic>/ (Artifact Store)            │
│  • All stage files (idea.md, board.md, competitors.md, ...)     │
│  • plan.md (compiled verdict + pivot docs)                      │
│  • assessment-*.md (dimension grades)                           │
│  • buyers-real.md (real conversations)                          │
│  • economics-real.md (real unit economics)                      │
└─────────────────────────────────────────────────────────────────┘
```

## Stage-to-Phase Mapping (9-Phase Ladder)

| Phase | Stages | Skills | Output |
|-------|--------|--------|--------|
| **1.1 Research** | 1-9 | ideator, founder, buyer-discovery, compliance-health, economics-engine | Market validation artifacts |
| **7 Error Recovery** | Pivot Gate | pipeline | Pivot decision documented |
| **2 Implement** | 10-13 | founder (marketing, brand, ops, launch) | Go-to-market artifacts |
| **8 Document** | 14 | plan-compiler | plan.md with verdict |
| **4/6 Verify** | 15 | assess | Dimension grades (A+-D) |
| **2+8 Optional** | 16 | youtube-automation | Video assets + findability report |

## Data Flow

```
User Input (niche/idea)
        │
        ▼
Stage 1: ideator → 8-12 candidates → .plaesy/ideas/<topic>/idea.md (candidate)
        │
        ▼ User chooses
Stage 2: founder/idea.md → Core definition
        │
        ▼
Stage 3: Board review → 3 lenses → board.md
        │
        ▼
Stage 4: Competitors → 3+ queries → competitors.md
        │
        ▼
Stage 5: buyer-discovery → ≥5 real conversations → buyers.md + buyers-real.md
        │
        ▼
Stage 6: Pricing → Segment-differentiated → pricing.md
        │
        ▼
Stage 7: Offer → Outcome/guarantee/included → offer.md
        │
        ▼
Stage 8: compliance-health (if health) → Risk scores → compliance-simulation.md
        │
        ▼
Stage 9: economics-engine → VERDICT (Incomplete/Not yet/Profitable) → economics.md
        │
        ├── Profitable ──► Continue to Stage 10
        └── Not yet ───► PIVOT GATE (must choose, document in plan.md)
                            │
                            ▼
Stage 10-13: Marketing, Brand, Ops, Launch → Artifacts
        │
        ▼
Stage 14: plan-compiler → plan.md (verdict + pivot docs + all stages)
        │
        ▼
Stage 15: assess → Dimension grades (business, marketing, financial)
        │
        ▼
Stage 16 (opt): youtube-automation → Video + thumbnail + findability report
```

## Key Components

### 1. Pipeline Orchestrator (`pipeline`)
- Manages 15+1 stages with handoffs
- Pivot Decision Gate (blocking)
- Resume-from-stage support
- Error recovery: ask ≤3 questions, record UNKNOWN, continue

### 2. Economics Engine (`economics-engine`)
- Core verdict calculation (Incomplete/Not yet/Profitable)
- Multi-tier segment analysis
- Sensitivity analysis (price, cost, volume, CAC, F, segment-mix)
- Oracle golden tests for verdict logic

### 3. Buyer Discovery (`buyer-discovery`)
- Outreach tracking → interviews → signal counting
- Segment Validation Matrix (kill <30% would-pay after ≥3 convos)
- Weighted would-pay scoring (decision-maker = 1.0, employee = 0.5)

### 4. Compliance Health (`compliance-health`)
- BPOM/FTC rule engine (banned terms, structure/function classification)
- Risk scoring (0-100, pass<30)
- Approved/banned claims DB with monthly update workflow

### 5. YouTube Automation (`youtube-automation`)
- 7-agent pipeline: Research → Fact-check → Script → Visuals → Voice → Render → Review/Publish
- Fail-soft API chains (Gemini→Groq→OpenRouter→Pollinations)
- Findability Score (11 checks, ≥80 to publish)
- Thumbnail Designer (4 styles × 2 ratios)

## Quality Gates (Blocking)

| Gate | Applies To | Threshold | Failure Action |
|------|------------|-----------|----------------|
| **B1 Real Buyers** | buyer-discovery, economics-engine | ≥5 real convos, ≥3 would-pay | Verdict = Incomplete |
| **B2 Economics** | economics-engine, plan-compiler | Rules-based verdict | Block or Pivot Gate |
| **B3 Pivot Gate** | pipeline | Not yet verdict | Must choose pivot |
| **B4 Sources** | competitors, marketing, assess | URL + date for all claims | Mark ASSUMED |
| **B5 Compliance** | compliance-health | Risk score <30 | Block auto-publish |
| **B6 Assessment** | assess | Mandatory audit items | Score capped at 49 |
| **B7 YouTube** | youtube-automation | Findability ≥80, thumb <2MB | Return to Review |

## External Dependencies (Fail-Soft Chains)

| Action | Primary | Backup | Keyless Fallback |
|--------|---------|--------|------------------|
| Web Search | Exa | Jina Reader | DuckDuckGo |
| Web Fetch | Jina Reader | Exa | Browser |
| LLM | Gemini | Groq | Pollinations |
| Visuals | Pexels | Pixabay → NASA → Openverse | AI-generated |
| Transcription | Groq Whisper | OpenAI Whisper | Local Whisper |
| Thumbnail | 4 Styles | Each style as fallback | N/A |

## File Structure

```
.plaesy/
├── ideas.md                              # Index: one line per idea
├── ideas/<topic>/
│   ├── idea.md                           # Core definition
│   ├── board.md                          # 3-lens review
│   ├── competitors.md                    # Price table + gap analysis
│   ├── buyers.md                         # Founder format (simulated)
│   ├── buyers-real.md                    # Real conversations
│   ├── pricing.md                        # Segment-differentiated
│   ├── offer.md                          # Outcome, guarantee, included/excluded
│   ├── economics.md                      # P,V,F,S → verdict
│   ├── economics-real.md                 # Real unit economics
│   ├── marketing.md                      # Channel, message, first test
│   ├── brand.md                          # Name, tone, anti-brand
│   ├── ops.md                            # Delivery requirements, capacity
│   ├── launch.md                         # 30-day plan, stop rules
│   ├── plan.md                           # Compiled: verdict + all stages
│   ├── assessment-business.md            # Grade + findings
│   ├── assessment-marketing.md           # Grade + findings
│   ├── assessment-financial.md           # Grade + findings
│   ├── compliance-simulation.md          # Rule engine test cases
│   └── oracle/                           # Golden test vectors
└── specs/<topic>/                        # Formal specs (spec-business, etc.)
```

## Skills Registry

| Skill | Category | Trigger | Output |
|-------|----------|---------|--------|
| `ideator` | Pipeline | "find business idea" | 8-12 candidates |
| `buyer-discovery` | Pipeline | "buyer interviews" | buyers.md ≥5 real |
| `compliance-health` | Pipeline | "BPOM/FTC compliance" | Risk score + alternatives |
| `economics-engine` | Pipeline | "unit economics" | Verdict + sensitivity |
| `plan-compiler` | Pipeline | "compile plan" | plan.md + pivot docs |
| `assess` | Assessment | "assess business" | Grade A+-D + findings |
| `pipeline` | Orchestration | "full pipeline" | Complete .plaesy/ideas/ |
| `youtube-automation` | Pipeline | "youtube video" | Video + thumbnail + metadata |

## Extension Points

1. **New Dimension Assessments** → Add to `assess` skill + `routing.md`
2. **New Compliance Regions** → Extend `compliance-health` with new rule engines
3. **New Video Platforms** → Add to `youtube-automation` visual chains
4. **New Pivot Options** → Extend Pivot Gate in `pipeline`
5. **New LLM Providers** → Add to LLM fallback chain in `fallback_chains.py`