# Spec-Kit Design System

## Design Principles

1. **Clarity over cleverness** — Every stage output is human-readable Markdown
2. **Fail-soft by default** — External calls never block; always have fallback
3. **Evidence over opinion** — Every claim needs source URL + date or marked ASSUMED
4. **Explicit over implicit** — Missing inputs = UNKNOWN, not guessed
5. **Human-in-the-loop** — Pivot Gate requires human decision

## Visual Language

### Document Structure
All stage outputs follow consistent Markdown structure:

```markdown
# Stage Title

**Key Metadata**: [values]

## Section 1
Content with evidence...

## Section 2
| Table | With | Sources |
|-------|------|---------|

**Verdict/Decision**: [clear statement]

---
*Generated: YYYY-MM-DD | Stage X of 15*
```

### Color Coding (for rendered HTML)
| Status | Color | Usage |
|--------|-------|-------|
| ✅ PASS | `#10B981` | Green - gate passed |
| ❌ FAIL | `#EF4444` | Red - gate failed |
| ⚠️ FLAG | `#F59E0B` | Amber - non-blocking issue |
| 🔵 INFO | `#3B82F6` | Blue - informational |
| 🟣 PIVOT | `#8B5CF6` | Purple - pivot decision needed |

### Typography
- **Headers**: System UI font, semibold
- **Body**: System UI font, regular
- **Code/Tables**: Monospace (JetBrains Mono / Fira Code)
- **Evidence links**: Underlined, `#3B82F6`

## Stage Output Templates

### 1. Idea Definition (`idea.md`)
```markdown
# <Working Name>

- What it is: <one sentence>
- Who it is for: <specific customer>
- What it sells, at what price: <product + price>
- Where and how: <channel + delivery>
- Budget and constraints: <money, time, skills, decisions>
```

### 2. Board Review (`board.md`)
```markdown
# Board Review: <topic>

## Brief
<2-3 sentences>

---

## Reviewer 1: Offer Lens
**Focus**: Value prop clarity
**Kill Risks**: [bullet list]
**Vote**: fund / fund with conditions / pass

## Reviewer 2: Defensibility Lens
**Focus**: Moat sustainability
**Kill Risks**: [bullet list]
**Vote**: fund / fund with conditions / pass

## Reviewer 3: Product Lens
**Focus**: Build feasibility
**Kill Risks**: [bullet list]
**Vote**: fund / fund with conditions / pass

---

## Summary
**Consensus**: [Fund with conditions / Pass]
**Key Conditions**: [numbered list]
```

### 3. Competitors (`competitors.md`)
```markdown
# Competitors: <topic>

| Alternative | Price (as published) | Promise | Source | Accessed |
|---|---|---|---|---|
| Do nothing | 0 | <status quo> | - | - |
| <Name> | <$X/mo or "not published"> | <promise> | <URL> | YYYY-MM-DD |
```

### 4. Buyers (`buyers.md`)
```markdown
# Buyers: <topic>

**Real conversations so far: <N>**

| Date | Person (role) | Question asked | Answer | Would pay? |
|---|---|---|---|---|
| YYYY-MM-DD | <role> | <question> | <answer> | YES/NO/MAYBE |
```

### 5. Economics (`economics.md`)
```markdown
# Economics: <topic>

## Inputs

| Input | Symbol | Value | Source |
|---|---|---|---|
| Price per sale | P | $X | pricing.md |
| Direct cost per sale | V | $X | API costs + hosting |
| Fixed cost per month | F | $X | tools + marketing + compliance |
| Expected sales/month | S | X | <stated basis> |

## Formulas
- Margin: m = P - V = $X (Y%)
- Break-even: B = ceil(F/m) = X sales/mo
- Margin of Safety: (S-B)/S = Z%

## Verdict: <Profitable | Not yet | Incomplete>
**Numbers behind it**: Margin $X (Y%), Break-even X, Expected X, MoS Z%
**Basis**: <S basis>
```

### 6. Plan (`plan.md`)
```markdown
# <topic>: business plan

**Verdict:** <Profitable | Not yet | Incomplete>
**Numbers:** margin $X (Y%), break-even Z, expected S, MoS W%
**Board votes:** <summary>

## Pivot Analysis (if Not yet)
**Original Model:** <segments, P, F, S, B, verdict>
**Failing Number:** <S < B | m <= 0 | missing input>
**Pivot Options:**
| Option | Segments | P | F | Target S | B | MoS | Verdict |

**Chosen Pivot:** <name>
**Rationale:** <why>
**Validation Required:** <what to test>

## Summary
<one page: what, who, offer, price, verdict in plain words>

## Stages
### idea.md
<content>
### board.md
<content>
...
```

## Interactive Elements (HTML Rendering)

### Findability Score Card (YouTube)
```
┌─────────────────────────────────────┐
│  Findability Score: 87/100 ✅       │
├─────────────────────────────────────┤
│ ✅ Title length 40-60 chars         │
│ ✅ Keyword in first 3 words         │
│ ✅ Description > 150 chars          │
│ ❌ End screen elements              │
│ ✅ 10/11 checks passed              │
└─────────────────────────────────────┘
```

### Thumbnail Style Selector
```
┌─────────┬─────────────┬─────────────┬─────────────┐
│ Impact  │ Cinematic   │ Split       │ Studio      │
│ ████████│ ████████    │ ████████    │ ████████    │
│ 1.2 MB  │ 1.5 MB      │ 1.3 MB      │ 1.4 MB      │
│ ✅      │ ✅          │ ✅          │ ✅          │
└─────────┴─────────────┴─────────────┴─────────────┘
```

### Risk Score Display (Compliance)
```
Risk Score: 25/100 ✅ PASS (<30)

Flags:
  ⚠️ "boost metabolism" - MEDIUM - Structure/function OK with disclaimer
  ✅ "supports immune function" - PASS - Approved claim
  ❌ "cures fatigue" - CRITICAL - Auto-reject
```

## Error States

### Missing Input (Stage Fails Loudly)
```markdown
⚠️ **MISSING INPUT: Real buyer conversations**

Stage 5 (buyers) requires ≥5 real conversations.
Current: 0 real, 5 simulated.

Options:
1. Provide real conversation data
2. Record as UNKNOWN, continue with simulated (verdict will be Incomplete)
3. Ask ≤3 clarifying questions

Questions:
1. Can you share 3 real conversation records?
2. What's your outreach method?
3. What's your current would-pay rate?
```

### Pivot Gate (Blocking)
```markdown
🟣 **PIVOT DECISION REQUIRED**

Economics verdict: **Not yet**
Failing number: **S=18 < B=32** (MoS -77.8%)

Pivot Options:
| Option | Segments | P | F | Target S | B | MoS | Verdict |
| B2B Only | Brands, Clinics, Apps | $250 | $2,000 | 10 | 10 | 0% | Profitable |
| Hybrid | All, diff pricing | $140 | $2,200 | 20 | 19 | 5% | Profitable |

**Required**: Choose option → Document in plan.md → Continue to Stage 10
```

## Accessibility

- All tables have proper headers
- Color coding supplemented with icons (✅❌⚠️🔵🟣)
- Alt text for all generated images
- Semantic HTML structure for screen readers
- Keyboard navigation for interactive elements

## Responsive Breakpoints
- **Desktop** (≥1024px): Full side-by-side layouts
- **Tablet** (768-1023px): Stacked tables with horizontal scroll
- **Mobile** (<768px): Card-based layouts, collapsible sections