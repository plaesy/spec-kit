---
description: "Management assessment workflow - team capacity, project management, process efficiency"
applyTo: "**/*"
---

# Management Assessment Instructions

**Use with**: `/assess:management` when dimension is **Management/Organization**

**Referenced from**: `/assess` (orchestrator)

---

**Reference checklist**: `.plaesy/checklists/pm.checklist.md` (planning, resourcing, risk,
delivery) covers the same ground as this file's scoring dimensions in checklist
form — use it as a quick manual pass before or alongside a full assessment run.

## Assessment Workflow

### Step 1: Team Capacity Assessment

- Audit current team size and skill distribution
- Assess capability gaps and training needs
- Review workload distribution and burnout risk
- Identify critical person dependencies and key person risks
- Evaluate knowledge transfer and documentation practices

**Rate**: 0-100 based on team readiness

### Step 2: Project Management & Workflow

- Audit project management processes and governance
- Review sprint/iteration planning and estimation practices
- Assess prioritization and backlog management
- Evaluate communication and coordination mechanisms
- Review delivery cadence and predictability

**Rate**: 0-100 based on project management maturity

### Step 3: Process Efficiency Review

- Audit incident response procedures and MTTR expectations
- Review change management and approval processes
- Assess runbook completeness and accessibility
- Evaluate on-call rotation and alerting practices
- Review post-incident review and learning procedures

**Rate**: 0-100 based on process maturity

### Step 4: Organizational Structure & Dependencies

- Map organizational structure and reporting lines
- Identify single points of failure in team/knowledge
- Assess cross-team dependencies and collaboration
- Review knowledge distribution across team
- Evaluate organizational scalability

**Rate**: 0-100 based on organizational health

### Step 5: Training & Development

- Assess training programs and skill development initiatives
- Review onboarding procedures and effectiveness
- Evaluate knowledge documentation and accessibility
- Assess mentoring and career development opportunities
- Review continuous learning culture

**Rate**: 0-100 based on development culture

### Step 6: Generate Report

**Report sections**: canonical names per `.plaesy/instructions/dimension-mapping.md`
→ *Assessment Report Section Registry*; the bullets below go *inside* those
sections and do not rename them.

- Overall management score (0-100)
- Per-category scores (Team Capacity, Project Management, Process Efficiency, Organization, Training)
- Top 5 findings (ordered by impact)
- Recommended improvements and roadmap

**Blocking Issues**:

- Critical skill gaps with no mitigation = BLOCKER
- No documented procedures for critical processes = BLOCKER
- Severe burnout risk or team morale issues = BLOCKER
- Single points of failure in critical knowledge = BLOCKER

---

## Quality Scoring

### Management Assessment Components

```text

team_capacity: 25%        # Team size, skills, workload, gaps
project_management: 20%   # Planning, prioritization, delivery cadence
process_efficiency: 20%   # Procedures, change mgmt, incident response
organization: 18%        # Structure, dependencies, scalability
training_development: 17% # Learning, onboarding, knowledge sharing
```

### Grade Mapping

| Grade | What it means here |
|---|---|
| A+ | Exceptional - Strong team, clear processes, high engagement |
| A | Excellent - Capable team, documented processes, minor gaps |
| B+ | Very Good - Good team dynamics, some process gaps, effective overall |
| B | Good - Functional team, process improvements identified |
| C | Acceptable - Team capacity concerns, significant process gaps |
| D | Needs Work - Critical team/process issues, high risk |

The score edges behind these letters, and the general meaning of each band,
are defined once in `.plaesy/instructions/quality-gates.md` →
**Grade bands**. Do not restate them here.

### Success Criteria (Hard Stops)

- MUST HAVE: Clear team structure and roles
- MUST HAVE: Documented project management process
- MUST HAVE: Incident response procedures (runbooks, on-call rotation)
- MUST HAVE: Knowledge documentation and sharing practices
- MUST HAVE: No critical single points of failure in team

---

## Critical Rules

- ✅ **MEASURE ACTUAL PRACTICES** - Assess what's documented AND what's actually done
- ✅ **CITE PROCESSES** - Reference documented procedures and policies
- ✅ **NO SPECULATION** - Only assess observable team practices
- ✅ **RISK PRIORITIZATION** - Identify key person dependencies and mitigation
- ✅ **REALISTIC ASSESSMENT** - Account for team size and resource constraints
- ✅ **HUMAN-CENTERED** - Focus on team health and sustainability, not just productivity

---

## Finding Standards

- **Assess burnout risk as a first-class finding** — sustainability outranks
  short-term velocity
- **Map single points of failure**, key-person dependencies especially
- **Compare documented procedure against observed practice** before scoring compliance
- **Record tacit knowledge as a finding** — undocumented know-how leaves with
  the person who holds it
- **Assess the system, not the person**, when something fails
- **Trace communication gaps** — they cause cascading failures
- **Report training and skills gaps**; they compound over time

---

## Key Metrics to Assess

**Team Capacity**:

- Current team size vs. project needs
- Skills distribution and gaps
- Workload balance and overtime rates
- Turnover rate and retention
- Average experience level

**Project Management**:

- Sprint velocity and predictability
- On-time delivery rate
- Estimation accuracy
- Cycle time from requirement to deploy
- Backlog health (prioritization, clarity)

**Process Efficiency**:

- Change approval time
- Incident resolution time
- Post-incident review cadence
- Documentation accuracy and currency
- Process compliance rate

**Organization**:

- Key person risk index
- Cross-team communication effectiveness
- Knowledge sharing frequency
- Organizational clarity/alignment
- Scalability capacity

**Training & Development**:

- Onboarding time to productivity
- Training hours per person per year
- Certification and skill progression rate
- Mentoring participation rate
- Internal promotion rate

---

## Pre-Assessment Validation

Run these checks BEFORE attempting assessment:

```text

Can you access team information?
  → yes: Review team structure and composition
  → no: Request organizational documentation

Are management processes documented?
  → yes: Review and assess procedures
  → partial: Flag gaps, recommend documentation

Is team feedback available?
  → yes: Incorporate feedback into assessment
  → no: Recommend team survey or interviews
```

---

## Uncertainty Surfacing (Before Finalizing Assessment)

Load `.plaesy/instructions/uncertainty-surfacing.md`; do not restate it here.
Only this dimension's specifics follow.

**Confidence basis for this dimension**:

- HIGH: actual delivery data, or a process document that was read as it exists
- MEDIUM: team or process pattern inferred from partial observation
- LOW: organisational data not available — no tracker, no process document, no team access

**Worked example**:

```text
I could not assess delivery predictability because sprint history was not shared
and no tracker is connected.
To improve confidence, provide the last six sprints of completed work with
their cycle times.
```

---

## Assessment Completion Checklist

Assessment is complete when:

- ✅ Team capacity and skill gaps identified with specificity
- ✅ Project management processes documented and assessed
- ✅ Process efficiency reviewed (incident response, change management)
- ✅ Organizational structure and dependencies mapped
- ✅ Training and development practices evaluated
- ✅ Key person risks identified and mitigation assessed
- ✅ Scores calculated per category
- ✅ Top 5 findings listed with impact assessment
- ✅ Blocking issues (burnout, key person dependency, critical gaps) flagged prominently
- ✅ Improvement roadmap and recommendations provided
- ✅ Console output delivered to user

## Persona Protocol

Loaded by every `each family:management` persona prompt for this dimension — `/assess`,
`/implement`, `/fix`, `/optimize`, `/loop` and `/improve` all point their
`Loads:` line here. The rules below are dimension-neutral and are written here
once rather than restated in 54 persona files, where they had already drifted
apart from the parents they point at.

- **Follow the parent protocol.** The persona fixes the *dimension*; the family
  prompt fixes the *verb*. Follow the family prompt's full protocol. Where the
  parent and this file disagree, the parent governs the behaviour.
- **This file is the *assessment* criteria for `management`** — workflow, weights,
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
