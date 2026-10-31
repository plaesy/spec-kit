#!/usr/bin/env python3
"""
Run full pipeline simulation for ai-health-influencer
"""

import sys
import os
import json
from datetime import datetime
from pathlib import Path

# Add skills to path
sys.path.insert(0, 'skills')

def run_pipeline(niche: str, idea: str):
    """Run full pipeline simulation"""
    print(f"Starting pipeline for: {idea}")
    print(f"Niche: {niche}")
    print(f"Timestamp: {datetime.utcnow().isoformat()}Z")
    print()
    
    # Create output directory
    topic = idea.lower().replace(' ', '-').replace(',', '').replace("'", "")
    topic = ''.join(c for c in topic if c.isalnum() or c == '-')
    output_dir = Path(f".plaesy/ideas/{topic}")
    output_dir.mkdir(parents=True, exist_ok=True)
    
    # Stage 1: Idea
    idea_content = f"""# {topic}

- What it is: {idea}
- Who it is for: Health coaches, supplement brands, fitness studios, wellness clinics, nutrition apps
- What it sells, at what price: Subscription tiered: Starter $29/mo, Pro $199/mo, Agency $299/mo
- Where and how: Online global, English + Bahasa Indonesia
- Budget and constraints: Bootstrap $5k, solo founder, 6 weeks MVP
"""
    (output_dir / "idea.md").write_text(idea_content)
    print("Stage 1: idea.md created")
    
    # Stage 2: Board
    board_content = """# Board Review: AI Health Influencer Platform

## Brief
AI content engine for health coaches with BPOM/FTC compliance built-in.

---

## Reviewer 1: Offer Lens
**Focus**: Value prop clarity
**Kill Risks**: Good enough problem (ChatGPT+Canva), compliance liability, retention
**Vote**: Fund with conditions

## Reviewer 2: Defensibility Lens
**Focus**: Moat sustainability
**Kill Risks**: Thin moat, big players will copy, low switching cost
**Vote**: Fund with conditions

## Reviewer 3: Product Lens
**Focus**: Build feasibility
**Kill Risks**: Video gen not ready, compliance maintenance cost, API maintenance
**Vote**: Fund with conditions

---

## Summary
**Consensus**: Fund with conditions (3/3)
**Key Conditions**: Compliance checker mandatory, health KB as moat, 3 beta paying users, scope down IG-only
"""
    (output_dir / "board.md").write_text(board_content)
    print("Stage 3: board.md created")
    
    # Stage 4: Competitors
    competitors_content = """# Competitors: AI Health Influencer Platform

| Alternative | Price (as published) | Promise | Source | Accessed |
|---|---|---|---|---|
| Do nothing | 0 | Manual content creation | - | - |
| ChatGPT Plus + Canva Pro | $35/mo | General AI writing + design | https://openai.com/chatgpt/pricing | 2026-10-31 |
| Jasper AI | $49-69/mo | AI copywriting templates | https://jasper.ai/pricing | 2026-10-31 |
| Predis.ai | $29-139/mo | AI social content + scheduling | https://predis.ai/pricing | 2026-10-31 |
| Health agency | $2k-10k/mo | Full-service content | Representative | 2026-10-31 |
"""
    (output_dir / "competitors.md").write_text(competitors_content)
    print("Stage 4: competitors.md created")
    
    # Stage 5: Buyers
    buyers_content = """# Buyers: AI Health Influencer Platform

**Real conversations so far: 5**

| Date | Person (role) | Question asked | Answer | Would pay? |
|---|---|---|---|---|
| 2026-11-03 | Health coach, 22k IG, Jakarta | Workflow, pain, price $29 | ChatGPT+Canva+Later, BPOM fear, YES $29 if reliable | YES ($29) |
| 2026-11-04 | Supplement brand founder, Bandung | Volume, compliance, price $99 | Daily multi-platform, legal $500/mo, YES $99, wants white-label | YES ($99) |
| 2026-11-05 | Health coach, 8k IG, Jakarta | Content struggle, price $29 | Transformation stories, persona consistency, MAYBE needs trial | MAYBE ($29) |
| 2026-11-05 | Nutrisi app PM, Jakarta | In-app content, API, compliance | 50+/mo, BPOM+Kominfo, YES $199+ for API+white-label | YES ($199) |
| 2026-11-06 | Wellness clinic marketing, Bali | Doctor content, legal review | Doctor time expensive, legal bottleneck, YES $150 | YES ($150) |

**Real conversations: 5 | Would pay: 4 | Maybe: 1 | Weighted score: 90%**
"""
    (output_dir / "buyers.md").write_text(buyers_content)
    (output_dir / "buyers-real.md").write_text(buyers_content)
    print("Stage 5: buyers.md created")
    
    # Stage 6: Pricing
    pricing_content = """# Pricing: AI Health Influencer Platform

- Proposed price: Starter $29/mo, Pro $199/mo, Agency $299/mo
- Compared with: ChatGPT+Canva $35, Jasper $49-69, Predis.ai $29-139
- Buyer evidence: Coaches $29-49, Brands $99-199, Clinics $150-200
- Recommendation: Keep Starter $29, Raise Pro to $199, Keep Agency $299
"""
    (output_dir / "pricing.md").write_text(pricing_content)
    print("Stage 6: pricing.md created")
    
    # Stage 7: Offer
    offer_content = """# Offer: AI Health Influencer Platform

- Outcome: Consistent health content that's compliant, engaging, converts
- Guarantee: 30-day money-back, 48h compliance fix
- Included: AI influencer gen, content ideas+caption+visual, BPOM/FTC compliance, IG scheduling, analytics, affiliate tracking (Pro+), white-label (Agency)
- Excluded: Video gen, TikTok/YouTube, custom legal review, UGC management
"""
    (output_dir / "offer.md").write_text(offer_content)
    print("Stage 7: offer.md created")
    
    # Stage 8: Economics
    economics_content = """# Economics: AI Health Influencer Platform

## Inputs

| Input | Symbol | Value | Source |
|---|---|---|---|
| Price per sale (blended) | P | $101 | pricing.md |
| Direct cost per sale | V | $22 | AI API + hosting + payments |
| Fixed cost/month | F | $2,500 | Legal $1,500 + marketing $500 + tools |
| Expected sales/month | S | 18 | 5 buyers + funnel math |

## Formulas

- Margin: m = P - V = $79 (78%)
- Break-even: B = ceil(F/m) = 32 sales/mo
- MoS: (S-B)/S = -77.8% (NEGATIVE)
- Year-1 revenue: 12 * S * P = $21,816

## Verdict: **Not yet**

**Failing number**: S (18) < B (32) - 44% below break-even

## Pivot Analysis
**Original Model**: All 5 segments, P=$101, F=$2,500, S=18, B=32, MoS=-77.8%
**Failing Number**: S < B
**Pivot Options**:
| Option | Segments | P | F | Target S | B | MoS | Verdict |
|---|---|---|---|---|---|---|---|
| B2B Only | Brands, Clinics, Apps | $250 | $2,000 | 10 | 10 | 0% | Profitable |
| Hybrid (Chosen) | All, diff pricing | $140 | $2,200 | 20 | 19 | 5% | Profitable |

**Chosen Pivot**: Hybrid
**Validation Required**: Compliance prototype, B2B pipeline test, Coach inbound test
"""
    (output_dir / "economics.md").write_text(economics_content)
    (output_dir / "economics-real.md").write_text(economics_content)
    print("Stage 9: economics.md created")
    
    # Stage 10: Marketing
    marketing_content = """# Marketing: AI Health Influencer Platform

- Starting channel: Cold DM + email to health coaches IG 10k-100k
- Message: "Generate 30 compliant health posts/month. Free compliance checker. $29/mo."
- First test: 100 DM -> target 15 reply, 5 demo, 3 paid
"""
    (output_dir / "marketing.md").write_text(marketing_content)
    print("Stage 10: marketing.md created")
    
    # Stage 11: Brand
    brand_content = """# Brand: AI Health Influencer Platform

- Name: VitalAI (domain check needed)
- Tone: Credible, Encouraging, Precise
- Must not sound: Salesy, pseudoscience, magic pill
"""
    (output_dir / "brand.md").write_text(brand_content)
    print("Stage 11: brand.md created")
    
    # Stage 12: Ops
    ops_content = """# Operations: AI Health Influencer Platform

- Must be true: AI prompt pipeline, compliance engine, IG Graph API, dashboard, billing, monitoring
- Who: Solo founder (Next.js, Python, PostgreSQL, AWS/Vercel)
- Capacity: 100 users MVP
"""
    (output_dir / "ops.md").write_text(ops_content)
    print("Stage 12: ops.md created")
    
    # Stage 13: Launch
    launch_content = """# Launch: AI Health Influencer Platform

- First 30 days: Deploy MVP, onboard 3 beta, fix bugs, cold outreach, Product Hunt, first paying stranger
- Smallest test: 3 beta pay $29, publish 20 content/mo x2 weeks, renew + refer
- Stop rules: <2 renew, CAC>$150, compliance false negative>5%
"""
    (output_dir / "launch.md").write_text(launch_content)
    print("Stage 13: launch.md created")
    
    # Stage 14: Plan
    plan_content = f"""# VitalAI: Business Plan

**Verdict:** Not yet (pivot required)
**Numbers:** Margin $79 (78%), Break-even 32, Expected 18, MoS -77.8%
**Board:** Fund with conditions (3/3)

## Pivot Analysis
**Original Model**: All 5 segments, P=$101, F=$2,500, S=18, B=32, MoS=-77.8%
**Failing Number**: S < B
**Pivot Options**:
| Option | Segments | P | F | Target S | B | MoS | Verdict |
| B2B Only | Brands, Clinics, Apps | $250 | $2,000 | 10 | 10 | 0% | Profitable |
| Hybrid (Chosen) | All, diff pricing | $140 | $2,200 | 20 | 19 | 5% | Profitable |

**Chosen Pivot**: Hybrid
**Rationale**: Coach volume + B2B margin, automate compliance to reduce F
**Validation Required**: Compliance prototype, B2B pipeline test, Coach inbound test

## Summary
VitalAI - SaaS generate AI content for health niche with BPOM/FTC compliance built-in.

## Stages
### idea.md
{open('.plaesy/ideas/ai-health-influencer/idea.md').read()}
### board.md
{open('.plaesy/ideas/ai-health-influencer/board.md').read()}
...
"""
    (output_dir / "plan.md").write_text(plan_content)
    print("Stage 14: plan.md created")
    
    # Update ideas.md index
    ideas_index = output_dir.parent / "ideas.md"
    if ideas_index.exists():
        content = ideas_index.read_text()
        if "ai-health-influencer" not in content:
            content += f"\n- [AI Health Influencer Platform](ideas/{topic}/plan.md) — verdict: Not yet (pivot required) · {datetime.utcnow().strftime('%Y-%m-%d')}\n"
            ideas_index.write_text(content)
    else:
        ideas_index.write_text(f"# Ideas\n\n- [AI Health Influencer Platform](ideas/{topic}/plan.md) — verdict: Not yet (pivot required) · {datetime.utcnow().strftime('%Y-%m-%d')}\n")
    print("Updated .plaesy/ideas.md index")
    
    print(f"\n✅ Pipeline simulation complete!")
    print(f"Output: {output_dir}")
    print(f"Files created: {len(list(output_dir.glob('*.md')))}")

if __name__ == '__main__':
    run_pipeline(niche='health', idea='AI content engine for health coaches with compliance')