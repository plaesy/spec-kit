---
description: "Universal orchestrator - guide spec-kit for ANY dimension (technical, design, business, product, marketing, operations, legal, financial)"
---

# Universal Orchestrator: Beyond Programming

⚡ **Framework Philosophy**: Spec-Kit is NOT just for code. It's for **ANY project need**.

---

## Spec-Kit Dimensions (8 Total)

### 1. 🔧 Technical
- Code quality, architecture, performance, security
- Testing, infrastructure, deployment
- Technology choices, dependencies
- Commands: `/implement`, `/optimize`, `/fix`, `/assess:technical`

### 2. 🎨 Design  
- UI/UX, accessibility (WCAG), design systems
- Component libraries, design tokens
- User experience optimization
- Commands: `/assess:design` (Mode 1 produces, Mode 2 audits — see Design Spine in `/assess`), `/optimize` (redesign/refactor)

### 3. 💼 Business
- Market fit, business model viability
- Revenue potential, unit economics
- Competitive positioning
- Commands: `/assess:business` (research + ambiguity resolution upfront, assessment mode post-build)

### 4. 📊 Product
- Feature set, roadmap, competitive analysis
- User needs fit, value proposition
- Product-market fit
- Commands: `/assess:product` (research + ambiguity resolution upfront, assessment mode post-build)

### 5. 📢 Marketing
- Positioning, messaging, brand
- Go-to-market strategy, audience fit
- Marketing campaigns, content strategy
- Commands: `/assess:marketing` (research mode upfront, assessment mode post-build), `/implement`

### 6. 🏢 Operations
- Team capacity, project management
- Process efficiency, scalability
- Deployment, release management
- Commands: `/assess:management`, `/optimize`, `/implement`

### 7. ⚖️ Legal/Compliance
- Regulatory compliance, risk assessment
- Data privacy (GDPR, CCPA, etc.)
- Legal contracts, IP management
- Commands: `/assess:legal` (research mode upfront, assessment mode post-build), `/fix`

### 8. 💰 Financial
- Pricing strategy, cost structure
- Profitability analysis, funding needs
- Budget forecasting, ROI calculation
- Commands: `/assess:financial` (research + ambiguity resolution upfront, assessment mode post-build)

---

## How Spec-Kit Works for ALL Dimensions

### Traditional (Limited to Technical)
```
/start "Build mobile app"
  → Only assesses technical feasibility
  → Only implements code
  → Ignores: business model, marketing, ops readiness
  → Result: Great code, wrong market
```

### Universal (Spec-Kit 9.2+)
```
/start "Build mobile app"
  ↓
/assess:technical → Code architecture ✅
/assess:design → UI/UX approach ✅
/assess:business → Market fit & viability ✅
/assess:product → Feature-market fit ✅
/assess:marketing → GTM strategy ✅
/assess:management → Team/process ready ✅
/assess:legal → Compliance gaps ✅
/assess:financial → Unit economics ✅
  ↓
Multi-dimensional validation complete
  ↓
/implement (technical + design + operations aligned)
  ↓
Result: Great code + right market + sustainable business ✅
```

---

## Autonomous Routing Across Dimensions

### Technical → Business Route
```
Technical Assessment finds: "10K concurrent users expected"
  ↓ Auto-routes to:
Business Assessment: "What's infrastructure cost per user?"
  ↓ Auto-routes to:
Financial Assessment: "Unit economics viable?"
  ↓ Routes back to:
Technical: "Yes → Scale to 10K, No → Redesign for 1K"
```

### Design → Marketing Route
```
Design creates: Beautiful SaaS dashboard
  ↓ Auto-routes to:
Marketing Assessment: "Does design match brand positioning?"
  ↓ If mismatch:
Design Refinement: "Update colors, tone, messaging"
  ↓ Routes to:
Marketing: "Ready for launch campaign"
```

### Legal → Operations Route
```
Legal Assessment finds: "GDPR compliance gaps"
  ↓ Auto-routes to:
Operations: "What process changes needed?"
  ↓ Implementation:
/implement --compliance → Update data handling
/assess:legal → Verify compliance
  ↓ Routes to:
Operations: "New process trained and deployed"
```

### Financial → Product Route
```
Financial Analysis: "CAC=$50, LTV=$200, need 3:1 ratio by month 6"
  ↓ Auto-routes to:
Product: "What features drive 3x LTV improvement?"
  ↓ Routes to:
Design + Marketing: "How to message new features?"
  ↓ Routes to:
Operations: "Can team execute in timeline?"
```

---

## Workflow Commands for ALL Dimensions

### `/start` — Universal Orchestrator
**Not just code**. Start any initiative (app, business, marketing campaign, operations overhaul, legal setup, financial model):

```bash
/start Build AI-powered e-commerce platform
  → Auto-assess all 8 dimensions
  → Create integrated plan across all areas
  → Coordinate implementation

/start Launch marketing campaign for Q4
  → Auto-assess: Marketing, Business, Financial, Product
  → Create campaign strategy + budget + metrics
  → Implement across channels

/start Restructure company operations
  → Auto-assess: Operations, Legal, Financial, Management
  → Create org design + process changes
  → Implement rollout plan

/start Ensure GDPR compliance
  → Auto-assess: Legal, Technical, Operations
  → Identify all gaps
  → Implement fixes + verify
```

### `/assess` — Universal Quality Gate
Run assessment on ANY dimension:

```bash
/assess:technical           # Code quality, performance, security
/assess:design              # UI/UX, accessibility, systems
/assess:business            # Market fit, model viability
/assess:product             # Features, roadmap, fit
/assess:marketing           # Positioning, GTM readiness
/assess:management          # Team, process, scalability
/assess:legal               # Compliance, risk, privacy
/assess:financial           # Pricing, cost structure, ROI

/assess:business,marketing  # Multi-dimensional: market + go-to-market
/assess                     # All dimensions (comprehensive audit)
```

### `/implement` — Universal Executor
Execute work across dimensions:

```bash
# Technical implementation
/implement --backend        # Build API, database, services
/implement --frontend       # Build UI, components, state management
/implement --devops         # Build infrastructure, deployment

# Design implementation
/implement --design         # Create design system, components
/implement --ui             # Build UI mockups, prototypes
/implement --accessibility  # Implement WCAG compliance

# Business/Operations implementation
/implement --operations     # Build processes, workflows, templates
/implement --compliance     # Implement regulatory requirements
/implement --team-structure # Restructure org, update roles

# Marketing/Product implementation
/implement --marketing      # Create campaigns, content, messaging
/implement --product        # Build features, update roadmap
```

### `/optimize` — Improve Any Dimension
```bash
/optimize --performance     # Speed up technical systems
/optimize:design-quality  # Improve UX, accessibility, systems
/optimize --user-experience # Streamline workflows
/optimize --cost           # Reduce operational costs
/optimize --marketing-roi  # Improve campaign efficiency
/optimize --team-process   # Speed up delivery cycles
```

### `/fix` — Resolve Issues Across Dimensions
```bash
/fix --security            # Fix security vulnerabilities
/fix --compliance          # Fix legal/regulatory gaps
/fix --performance         # Fix slow queries, endpoints
/fix --bug                 # Fix technical bugs
/fix --design              # Fix UI/UX issues
/fix --process             # Fix operational bottlenecks
```

### `/assess` (Research Mode) — Deep Dive Any Topic

Run `/assess:{scope}` before a spec/decision exists (Mode 1 — see `/assess`)
and it researches the web/Context7 for evidence instead of assessing existing output:
```bash
/assess:business Market opportunity for feature X
/assess:business Competitive landscape in segment Y
/assess:legal Best practices for GDPR compliance
/assess:management Optimal team structure for scale
/assess:financial Pricing models for SaaS products
/assess:technical Technology stack for use case
```

### `/assess` (Mode 1) — Resolve Ambiguity
```bash
/assess:business Business model (revenue streams unclear)
/assess:product Product roadmap (feature priorities unclear)
/assess:technical Technical architecture (design uncertain)
/assess:management Team structure (role responsibilities unclear)
/assess:marketing Marketing strategy (positioning uncertain)
```

### `/assess` (Mode 1, Design Spine) — Create Design Across Dimensions
```bash
/assess:design              # UI/UX mockups, system, components, flows
/assess:technical            # Technical architecture diagrams (Design Spine)
/assess:management           # Organization structure (Design Spine)
/assess:business             # Business processes (Design Spine)
```

---

## Autonomous Decision Making (All Dimensions)

### Technical vs Business Priority Conflict
```
Technical Assessment: "Database refactor needed"
Business Assessment: "New market entry urgent"

Framework Auto-Decides:
IF market opportunity is time-sensitive
  → Delay refactor (defer technical debt)
  → Hire contractor for technical work if needed
  → Proceed with market entry
ELSE IF refactor blocks scaling
  → Do refactor first
  → Then market entry
ELSE
  → Parallel work (technical team + business team)
```

### Design vs Speed Conflict
```
Design Assessment: "Complete overhaul needed"
Marketing Assessment: "Launch campaign in 2 weeks"

Framework Auto-Decides:
IF timeline non-negotiable
  → Use existing design + quick polish
  → Plan full redesign for post-launch (phase 2)
ELSE IF design critical for positioning
  → Delay launch, nail design
  → Launch with strong positioning
```

### Compliance vs Cost Conflict
```
Legal Assessment: "GDPR compliance: $50K implementation"
Financial Assessment: "Budget: $30K max"

Framework Auto-Decides:
IF GDPR mandatory (operates in EU)
  → Allocate $50K (non-negotiable)
  → Cut other expenses
ELSE IF optional (not EU users)
  → Build for future ($15K/phase)
  → Defer full compliance
```

---

## Integration Examples

### Example 1: Launch SaaS Product (All Dimensions)

```
/start Launch AI analytics SaaS

PHASE 1: Universal Assessment
├─ /assess:technical → "Scalability: ✅ PostgreSQL + Redis + K8s ready"
├─ /assess:design → "UX: ⚠️ Dashboard needs accessibility improvements"
├─ /assess:business → "Model: ⚠️ Pricing needs validation"
├─ /assess:product → "Features: ✅ Core features ready, roadmap clear"
├─ /assess:marketing → "GTM: ⚠️ Positioning unclear vs competitors"
├─ /assess:management → "Team: ✅ Org ready, processes defined"
├─ /assess:legal → "Compliance: ✅ SOC 2 ready, GDPR compliant"
└─ /assess:financial → "Economics: ⚠️ CAC/LTV ratio needs 20% improvement"

PHASE 2: Autonomous Routing
├─ Design issue → /implement --accessibility
├─ Business issue → /assess:business --research (pricing validation)
├─ Marketing issue → /assess:marketing --research (competitive positioning)
└─ Financial issue → /optimize --economics (CAC reduction strategy)

PHASE 3: Implementation
├─ /implement --accessibility → Add keyboard nav, ARIA labels, color contrast
├─ /implement --marketing → Create positioning docs, messaging framework
├─ /implement --operations → Sales playbook, onboarding process

PHASE 4: Verification
├─ /assess:design → Verify accessibility ≥ WCAG AA
├─ /assess:business → Validate pricing with customers
├─ /assess:financial → Verify CAC/LTV ratio improved
└─ /assess:marketing → Validate positioning resonates

Result: Ready for launch ✅ (all dimensions validated)
```

### Example 2: Restructure Operations (All Dimensions)

```
/start Restructure company for rapid growth

PHASE 1: Assessment
├─ /assess:management → "Org bottlenecks identified, processes unclear"
├─ /assess:technical → "Technical onboarding slow (4 weeks)"
├─ /assess:legal → "Role documentation missing, compliance gaps"
├─ /assess:financial → "Headcount costs scaling faster than revenue"
└─ /assess:business → "Team structure misaligned with strategy"

PHASE 2: Auto-Routing
├─ Operations issues → /assess:management (Design Spine: new structure)
├─ Technical issues → /implement --onboarding (faster setup)
├─ Legal issues → /fix --compliance (role docs, policies)
├─ Financial issues → /optimize --headcount (efficiency)
└─ Business issues → /assess:business --research (org alignment options)

PHASE 3: Implementation
├─ /assess:management (Design Spine) → New reporting structure, clear roles
├─ /implement --processes → Updated workflows, handoffs
├─ /implement --onboarding → Automated setup, training
├─ /implement --documentation → Policy manual, role guides

PHASE 4: Verification
├─ /assess:management → Team satisfaction, process efficiency
├─ /assess:financial → Cost per hire, time-to-productivity
├─ /assess:technical → Onboarding time reduced 50%

Result: Org optimized ✅ (all dimensions improved)
```

---

## 10/10 Autonomy: The Last Gap

**9.2/10 → 10/10 requires one final piece**:

### Gap: Cross-Dimensional Learning

**Problem**: Each dimension optimizes independently. No system learning.

**Solution**: Framework remembers patterns:
- "When we prioritize market over tech debt, outcome is X"
- "If financial constraint <$50K, timeline extends 3 weeks"
- "Design-marketing misalignment causes 20% lower adoption"

**Implementation** (future):
```
/start --learn-from [past-projects]
  → Loads historical patterns
  → Makes better decisions across dimensions
  → Improves prediction accuracy

/assess --predict
  → Based on learned patterns
  → "This pricing likely fails because..."
  → "This team structure will cause..."
  → "This timeline is optimistic because..."
```

**This achieves 10/10**: Framework not just autonomous, but **intelligent** (learns + adapts).

---

## Summary: Spec-Kit is Universal

Spec-Kit is **NOT** a programming tool.

It's a **universal orchestrator** for:
- ✅ Building products (technical + design + product)
- ✅ Launching companies (business + financial + operations)
- ✅ Running operations (management + legal + compliance)
- ✅ Growing marketing (marketing + product positioning)
- ✅ Any initiative requiring coordinated multi-dimensional work

**Framework handles**:
- ✅ Assessment across all dimensions
- ✅ Autonomous routing between dimensions
- ✅ Conflict resolution (priority, budget, timeline)
- ✅ Integration testing (all dimensions together)
- ✅ Continuous improvement (learn + adapt)

**Result**: One framework for ALL work types. No more switching tools.
