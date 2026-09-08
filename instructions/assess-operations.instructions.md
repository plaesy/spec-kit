---
description: "Operations and deployment assessment workflow - infrastructure, team capacity, process efficiency"
---

# Operations Assessment Instructions

**Use with**: `/assess:operations` when dimension is **Operations**

**Referenced from**: `.plaesy/memory/assess.md` (orchestrator)

---

## Assessment Workflow

### Step 1: Infrastructure Assessment

- Audit current infrastructure (on-prem, cloud, hybrid)
- Verify scalability and capacity planning
- Assess redundancy and high-availability setup
- Review disaster recovery and backup procedures
- Evaluate monitoring and alerting coverage

**Rate**: 0-100 based on infrastructure readiness

### Step 2: Deployment Readiness

- Verify CI/CD pipeline setup and automation
- Assess deployment procedures and rollback capability
- Review release process and change management
- Evaluate deployment frequency and reliability
- Assess environment parity (dev, staging, production)

**Rate**: 0-100 based on deployment maturity

### Step 3: Team Capacity Assessment

- Audit current team size and skill distribution
- Assess capability gaps and training needs
- Review workload distribution and burnout risk
- Evaluate knowledge transfer and documentation
- Identify critical person dependencies

**Rate**: 0-100 based on team readiness

### Step 4: Process Efficiency Review

- Audit incident response procedures and MTTR
- Review change management and approval processes
- Assess runbook completeness and accuracy
- Evaluate on-call rotation and alerting fatigue
- Review post-incident review procedures

**Rate**: 0-100 based on process maturity

### Step 5: Monitoring & Observability

- Verify monitoring coverage (infrastructure, application)
- Assess alerting rules and alert fatigue
- Review logging completeness and retention
- Evaluate metrics and SLO definition
- Assess observability tool stack

**Rate**: 0-100 based on observability maturity

### Step 6: Security Operations

- Verify access control and secrets management
- Assess security patching procedures
- Review vulnerability scanning and remediation
- Evaluate compliance monitoring
- Assess audit logging and forensics capability

**Rate**: 0-100 based on security operations maturity

### Step 7: Generate Report

- Overall operations score (0-100)
- Per-category scores (Infrastructure, Deployment, Team, Processes, Monitoring, Security)
- Top 5 findings (ordered by impact)
- Recommended improvements and roadmap

**Blocking Issues**:
- No backup/disaster recovery = BLOCKER
- No monitoring/alerting = BLOCKER
- Deployment process failures = BLOCKER
- Critical skill gaps = BLOCKER

---

## Quality Scoring

### Operations Assessment Components

```
infrastructure: 20%       # Scalability, redundancy, capacity, DR
deployment: 20%           # CI/CD, automation, release process
team_capacity: 20%        # Team size, skills, workload, knowledge transfer
process_efficiency: 15%   # Incident response, change management, runbooks
monitoring: 15%           # Infrastructure/app monitoring, logging, metrics
security_ops: 10%         # Access control, patching, vulnerability management
```

### Grade Mapping

```
95-100: A+ (Exceptional - Mature operations, high reliability, well-documented)
90-94:  A  (Excellent - Reliable operations, minor gaps)
85-89:  B+ (Very Good - Good operational coverage, some optimization needed)
80-84:  B  (Good - Functional operations, improvement areas identified)
70-79:  C  (Acceptable - Basic operational capability, significant gaps)
<70:    D  (Needs Work - Immature operations, high risk)
```

### Success Criteria (Hard Stops)

- MUST HAVE: Backup and disaster recovery procedures
- MUST HAVE: Monitoring and alerting (infrastructure and application)
- MUST HAVE: Incident response procedures (runbooks, on-call rotation)
- MUST HAVE: CI/CD pipeline and deployment automation
- MUST HAVE: Access control and secrets management

---

## Critical Rules

- ✅ **MEASURE RELIABILITY** - Use actual MTTR, deployment frequency, error rates
- ✅ **CITE PROCEDURES** - Reference documented runbooks and processes
- ✅ **NO SPECULATION** - Only assess what's observable or documented
- ✅ **RISK PRIORITIZATION** - Identify single points of failure
- ✅ **REALISTIC ASSESSMENT** - Account for team size and resource constraints
- ✅ **OPERATIONAL REALITY** - Distinguish between documented and actual practices

---

## Anti-Patterns (NEVER Do These)

- ❌ **Never assume redundancy without verification** - Test failover
- ❌ **Never skip monitoring assessment** - Observability is critical
- ❌ **Never ignore team capacity issues** - Burnout impacts reliability
- ❌ **Never assume runbooks are current** - Verify and test them
- ❌ **Never overlook single points of failure** - Identify critical dependencies
- ❌ **Never skip disaster recovery testing** - It must actually work
- ❌ **Never ignore security operations** - Patching and access control matter

---

## Key Metrics to Assess

**Reliability**:
- Mean Time To Recovery (MTTR)
- Mean Time Between Failures (MTBF)
- Availability/uptime percentage
- Error rate and incident frequency

**Deployment**:
- Deployment frequency
- Lead time for changes
- Change failure rate
- Time to rollback

**Team**:
- On-call rotation coverage
- Alert fatigue (false positive rate)
- Knowledge bus factor
- Training and skill development

**Process**:
- Change approval time
- Incident resolution time
- Post-incident review cadence
- Documentation accuracy

---

## Pre-Assessment Validation

Run these checks BEFORE attempting assessment:

```
Can you access infrastructure information?
  → yes: Audit configuration and setup
  → no: Request infrastructure documentation

Are operations procedures documented?
  → yes: Review and test procedures
  → partial: Flag gaps, recommend documentation

Is monitoring in place?
  → yes: Assess coverage and effectiveness
  → no: Critical gap, recommend immediate implementation
```

---

## Success Criteria

Assessment is complete when:

- ✅ Infrastructure components assessed for scalability and redundancy
- ✅ Deployment process evaluated for automation and reliability
- ✅ Team capacity and skill gaps identified
- ✅ Operational procedures reviewed (incident response, change management)
- ✅ Monitoring and observability coverage assessed
- ✅ Security operations practices evaluated
- ✅ Scores calculated per category
- ✅ Key metrics and KPIs identified
- ✅ Top 5 findings listed with impact assessment
- ✅ Blocking issues (backup, monitoring, DR) flagged prominently
- ✅ Improvement roadmap recommended
- ✅ Console output delivered to user
