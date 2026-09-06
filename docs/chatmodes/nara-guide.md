# `@nara` - Decision Oracle Chat Mode Guide

**nara** = "narasumber" (expert source in Indonesian) — oracle that gathers expert perspectives and **makes decisions**

## What is `@nara`?

A decision-making oracle chat mode that:
1. **Gathers perspectives** from relevant expert viewpoints (developer, architect, operations, product, security, etc)
2. **Deliberates** across viewpoints to identify tensions and trade-offs
3. **Makes the decision** (not "it depends" or "both work")
4. **Documents rationale** with clear explanation of why this decision serves the team

**Key difference from advisory**: Nara doesn't just present options—it decides.

**How to call it**: `@nara [question]` from anywhere in your workflow

## When to Use `@nara`

### ✅ Good Decisions for Nara
```
"Should we refactor this service now or later?"
"PostgreSQL or MongoDB for this feature?"
"When should we migrate to microservices?"
"Should we add feature X or fix tech debt?"
"REST API or GraphQL?"
```

### ❌ Not the Right Tool For
```
❌ "Review my code" → use /code-review
❌ "Is my project good?" → use /assess
❌ "What are the options?" → use /research
❌ "How do I build X?" → use /implement
```

## How to Use `@nara`

### Basic Syntax
```
@nara Should we do X?
Context: [Situation, constraints, options being considered]
```

Nara is a **chat mode** (not a prompt), so you call it with `@` to consult it:
- From any phase of `/start`, `/clarify`, `/implement`, etc
- During discussions and planning
- Anytime you need to make a decision

### What Nara Will Do

1. **Analyze the decision** → Understand the core question
2. **Select relevant perspectives** → Which chatmodes to consult
   - Code decision → @dev, @sa, @qa
   - Architecture decision → @sa, @dev, @devops, @security
   - Feature decision → @pm, @ba, @dev, @qa
   - Deployment decision → @devops, @sre, @security, @qa
3. **Consult each perspective** → Ask for their input
4. **Deliberate** → Identify where perspectives align/conflict
5. **Decide** → Make the call with clear rationale
6. **Document** → Save decision record to memory

### Example Input

```
/nara Should we refactor UserService?

Context:
- Currently 400 lines with 3 mixed responsibilities
- Hard to test, slowing down feature development
- No immediate feature pressure
- Team has 8 hours available this week
```

**Nara will consult**:
- @dev: "Is this slowing us down?"
- @sa: "Does this violate architecture principles?"
- @qa: "Does this affect testability?"
- @pm: "Does this impact roadmap?"

**Nara decides**: "Yes, refactor now. Reason: Testability is the blocker. 8 hours now saves time on next 3 features."

## Available Perspectives (Chatmodes)

Nara can consult any of these perspectives:

| Chatmode | Perspective | Care About |
|----------|-------------|-----------|
| `@ba` | Business Analyst | Requirements, stakeholder alignment |
| `@bo` | Business Owner | ROI, strategic alignment |
| `@dev` | Developer | Code quality, velocity, tech debt |
| `@sa` | Solution Architect | Scalability, integration, design |
| `@pm` | Product Manager | User value, roadmap |
| `@po` | Product Owner | Feature scope, priorities |
| `@qa` | QA Lead | Testing, reliability, coverage |
| `@security` | Security Expert | Vulnerabilities, compliance, threats |
| `@devops` | DevOps Engineer | Deployment, monitoring, ops |
| `@designer` | Designer | UX, accessibility, user experience |
| `@sre` | SRE | Reliability, automation, incident response |
| `@compliance` | Compliance Officer | Regulations, standards, audit |
| `@pe` | Performance Engineer | Optimization, benchmarks |

Nara automatically selects the most relevant ones for each decision.

## What You Get from Nara

### Decision Record Format

```
DECISION: [Clear statement of what we're doing]

PERSPECTIVES HEARD:
- Developer: "Hard to test, takes 8 hours"
- Architect: "Violates SRP, will scale poorly"
- Product: "No user impact, can wait"
- QA: "Can't hit 80% coverage with current structure"

RATIONALE: Testability is slowing velocity. 8 hours now saves time on next 3 features.

TRADE-OFFS:
- Gain: 50% faster feature development, clearer code
- Lose: 8 hours now, 3 files instead of 1
- Accept: Small complexity increase for much better maintainability

DISSENTING VIEWS: None strong - all perspectives aligned

RISKS & MITIGATION:
1. Refactoring introduces bugs → All existing tests run before/after
2. Blocks other work → Can be done in 1 sprint, no critical features

SUCCESS CRITERIA:
✓ All tests pass
✓ Code coverage >85%
✓ Method length <50 lines
✓ New features 30% faster

NEXT STEPS:
1. Create feature branch - Dev Team, Day 1
2. Refactor - Dev Team, Days 2-3
3. Code review - Lead, Day 4
4. Merge & measure - Dev Team, Day 5

REVISIT PLAN: None needed unless deeper issues discovered
```

## Decision Types

### Strategic Decisions
"Long-term architectural/organizational choices"
- Should we go microservices?
- Should we switch to PostgreSQL?
- Should we rebuild this system?
- When should we hire more engineers?

### Tactical Decisions
"Important choices affecting immediate work"
- Should we refactor this class?
- Should we add caching here?
- Should we implement feature A or B first?
- Should we upgrade this dependency?

### Technical Disputes
"Team disagrees on approach"
- REST vs GraphQL?
- SQL vs NoSQL?
- Build vs buy?
- Monolith vs microservices?

## Tips for Better Decisions

1. **Be clear about the question** - "Should we X?" not "What should we do?"
2. **Provide full context** - Current state, options, constraints, pressure
3. **Identify trade-offs you see** - "We want X but Y would require..."
4. **Be ready to accept the decision** - Nara decides, doesn't present options
5. **Document the decision** - Save record for future reference
6. **Revisit when needed** - Some decisions should be reassessed later

## Example Scenarios

### Example 1: Code Decision
```
/nara Should we refactor this 400-line service?

Context: 3 mixed responsibilities, hard to test, slowing development

Nara: YES. Testability is the blocker. Pays off in 3 features. Do it this week.
```

### Example 2: Architecture Decision
```
@nara When should we migrate monolith to microservices?

Context: 50K LOC, 3 independent services, 2-hour deployments

Nara: START NOW with auth service. Phase 1 (4 weeks), then payment, then notification.
Clean boundaries, improves deployment speed, aligns with team skills.
```

### Example 3: Technology Decision
```
@nara PostgreSQL or MongoDB for new feature?

Context: Need ACID transactions, scale to 100K records, team knows SQL, speed to market matters

Nara: PostgreSQL. Transaction requirement is critical. Team expertise matters. 
Can scale vertically for now. Revisit in 2 years if hitting limits.
```

### Example 4: Priority Decision
```
/nara Should we fix tech debt or ship new features?

Context: 30% velocity loss from tech debt, but roadmap pressure for features

Nara: Both, but sequenced: Fix top 2 tech debts (2 weeks), then resume features (4 weeks).
Unblocks velocity without delaying roadmap. Revisit every quarter.
```

## Related Prompts

| Prompt | When | Output |
|--------|------|--------|
| `/nara` | **Need to DECIDE** | Final decision with rationale |
| `/assess` | Want quality evaluation | Assessment scores, feedback |
| `/code-review` | Want code feedback | Code improvements, issues |
| `/research` | Want to explore options | Research findings, comparisons |
| `/implement` | Know what to do | Implementation guidance |
| `/clarify` | Resolve ambiguities | Specifications, decisions |

## Decision Ownership

After Nara decides:
- **Owner**: Who is responsible for implementation?
- **Timeline**: When does this start?
- **Success criteria**: How do we know it worked?
- **Revisit plan**: When should we reassess?

Decisions without ownership are just conversations. Nara makes sure every decision has clear next steps.

---

**Use `/nara` or `@nara` when your team needs to make an important decision!**

