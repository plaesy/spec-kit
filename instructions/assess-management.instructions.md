---
description: "Management assessment workflow - team capacity, project management, process efficiency"
---

# Management Assessment Instructions

**Use with**: `/assess:management` when dimension is **Management/Organization**

**Referenced from**: `.plaesy/memory/assess.md` (orchestrator)

---

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

```
team_capacity: 25%        # Team size, skills, workload, gaps
project_management: 20%   # Planning, prioritization, delivery cadence
process_efficiency: 20%   # Procedures, change mgmt, incident response
organization: 18%        # Structure, dependencies, scalability
training_development: 17% # Learning, onboarding, knowledge sharing
```

### Grade Mapping

```
95-100: A+ (Exceptional - Strong team, clear processes, high engagement)
90-94:  A  (Excellent - Capable team, documented processes, minor gaps)
85-89:  B+ (Very Good - Good team dynamics, some process gaps, effective overall)
80-84:  B  (Good - Functional team, process improvements identified)
70-79:  C  (Acceptable - Team capacity concerns, significant process gaps)
<70:    D  (Needs Work - Critical team/process issues, high risk)
```

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

## Anti-Patterns (NEVER Do These)

- ❌ **Never ignore team burnout risks** - Sustainability matters more than short-term velocity
- ❌ **Never overlook single points of failure** - Key person dependencies are critical risks
- ❌ **Never assume procedures are actually followed** - Verify documented vs. actual practice
- ❌ **Never skip knowledge documentation** - Tacit knowledge is a liability
- ❌ **Never blame individuals for systemic issues** - Assess the system, not the person
- ❌ **Never ignore communication gaps** - Miscommunication causes cascading failures
- ❌ **Never overlook training needs** - Skills gaps compound over time

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

```
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

## Success Criteria

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
