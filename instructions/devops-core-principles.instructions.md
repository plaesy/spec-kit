---
applyTo: 'deployments/**/*.yaml, deployments/**/*.yml, k8s/**/*.yaml, k8s/**/*.yml'
description: 'Foundational instructions covering core DevOps principles, culture (CALMS), and key metrics (DORA) to guide GitHub Copilot in understanding and promoting effective software delivery.'
---

# DevOps Core Principles

## Mission
Understand and advocate for core DevOps principles: guide developers toward a collaborative, automated, continuously-improving delivery culture. When generating/reviewing code, consider alignment with these principles.

## What is DevOps?
DevOps combines development (Dev) and operations (Ops) to shorten the delivery lifecycle while shipping features/fixes frequently, aligned with business goals - a cultural/technical shift toward higher-velocity delivery. Emphasizes communication, collaboration, integration, and automation between dev and ops, leading to faster time to market, reliability, security, and customer satisfaction. Not a methodology like Agile - a set of principles applicable across methodologies.

## The CALMS Framework (Pillars of DevOps)

### 1. Culture
**Principle**: foster a collaborative, blameless culture prioritizing shared responsibility, trust, continuous learning.
- Break down silos: communication/empathy between dev, ops, security, business; cross-functional understanding
- Shared responsibility: "we own this together," from commit to production and beyond
- Blameless post-mortems: focus on systemic issues and learning, not blame - builds transparency and psychological safety
- Continuous learning: growth mindset, experimentation, knowledge sharing
- Fast, clear feedback loops across the delivery lifecycle and between teams

**Copilot guidance**: in code reviews, suggest explaining the 'why' behind changes; in problem-solving, investigate across the whole pipeline, not just code; write docs accessible to both dev and ops; recommend communication practices (stand-ups, shared channels) where relevant; write runbooks executable by anyone on the team, not just experts.

**Key takeaway**: people and interactions are paramount - technology enables, doesn't replace, a strong culture.

### 2. Automation
**Principle**: automate everything possible across the delivery lifecycle to reduce manual effort, increase speed, minimize human error.
- CI/CD pipelines: automate build/test/integrate/deploy - backbone of fast, reliable releases
- Infrastructure as Code (Terraform, Ansible, Pulumi): consistency, repeatability, version control
- Configuration management (Ansible, Puppet, Chef)
- Automated testing: unit, integration, E2E, performance, security - run in pipeline
- Monitoring/alerting: automated metrics/logs/traces collection, automated anomaly alerts
- Security automation: SAST/DAST/SCA scans integrated into CI/CD

**Copilot guidance**: prioritize automating builds/tests/deploys/security scans in generated CI/CD pipelines (e.g. GitHub Actions); suggest IaC/config-management snippets; recommend automating repetitive ops tasks (log analysis, auto-scaling); help generate comprehensive automated test cases; suggest fully automated blue/green or canary deployments where feasible.

**Key takeaway**: if it's repeatable, automate it - frees engineers for complex problems, reduces error, accelerates feedback and delivery velocity.

### 3. Lean
**Principle**: apply lean manufacturing principles - eliminate waste, maximize flow, deliver value continuously.
- Eliminate waste: excessive docs, unnecessary approvals, waiting, manual handoffs, rework
- Maximize flow: smaller commits/PRs, frequent deployments, smooth idea-to-production flow
- Value stream mapping: understand the full delivery process to find bottlenecks
- Build quality in throughout development, not just end-of-cycle testing
- Just-in-time delivery: ship features/fixes as soon as ready

**Copilot guidance**: suggest breaking large features into small, frequent PRs/iterative deployments; advocate MVPs and iterative development; help identify/remove pipeline bottlenecks; promote continuous improvement from fast feedback and data; emphasize modularity/testability to reduce future waste.

**Key takeaway**: deliver value quickly and iteratively, minimize non-value-adding activities - enhances agility and responsiveness.

### 4. Measurement
**Principle**: measure everything relevant across the pipeline and app lifecycle to find bottlenecks and drive improvement.
- KPIs: delivery speed, quality, operational stability (DORA metrics)
- Centralized monitoring/logging: metrics, logs, traces
- Actionable dashboards visualizing system and pipeline health
- Effective alerting for critical issues
- Experimentation/A/B testing to validate hypotheses
- Capacity planning from resource-utilization metrics

**Copilot guidance**: suggest relevant metrics when designing systems/pipelines (latency, error rate, deployment frequency, lead time, MTTR, change failure rate); recommend robust logging/monitoring (structured logging, tracing); encourage dashboards/alerts (Prometheus, Grafana); use data to validate changes and justify architecture; when debugging, check metrics/logs first.

**Key takeaway**: you can't improve what you don't measure - data-driven decisions are essential.

### 5. Sharing
**Principle**: promote knowledge sharing, collaboration, transparency across teams.
- Shared tools/platforms/practices across teams
- Clear, up-to-date documentation (runbooks, ADRs)
- Open, accessible communication channels (Slack, Teams, wikis)
- Cross-functional teams: dev/ops working closely together
- Pair/mob programming to spread knowledge
- Internal meetups/workshops for best practices

**Copilot guidance**: suggest documenting processes/architecture decisions/runbooks (e.g. ADR/runbook markdown templates); advocate shared tools/platforms; recommend cross-training/pairing in project setup; recommend clear communication channels; include explanatory comments/internal docs in generated code.

**Key takeaway**: breaking down silos and open communication is crucial - sharing accelerates learning and builds resilience.

## The Four Key DORA Metrics
Empirically linked to higher organizational performance (profitability, productivity, market share). Help teams improve these.

### 1. Deployment Frequency (DF)
**Definition**: how often the org successfully releases to production - measures delivery speed.
- Elite performers deploy multiple times/day; high frequency reduces per-deployment risk, eases debugging, speeds feedback
- Enabled by small, incremental changes over large infrequent ones

**Copilot guidance**: design CI/CD for frequent, small, safe deployments (automated testing, blue/green); break large features into independently deployable units; use feature flags to decouple deployment from release.

**Goal**: high (elite: multiple/day). **Impact**: faster time to market, quicker feedback, reduced risk per change.

### 2. Lead Time for Changes (LTFC)
**Definition**: time from commit to production - measures dev-to-delivery speed.
- Spans the full value stream, commit to deployment
- High lead time signals bottlenecks in dev, test, or deployment

**Copilot guidance**: reduce bottlenecks (smaller PRs, automated testing, faster builds, efficient review); streamline approvals, eliminate manual handoffs; recommend frequent CI merges/testing; suggest CI/CD caching strategies.

**Goal**: low (elite: <1 hour). **Impact**: rapid response to market changes, faster defect resolution, higher productivity.

### 3. Change Failure Rate (CFR)
**Definition**: % of deployments causing service degradation (rollback/hotfix/outage) - measures delivery quality.
- Lower = higher quality/stability; high CFR often from insufficient testing, missing automated checks, poor rollback, overly complex deployments

**Copilot guidance**: emphasize robust testing (unit/integration/E2E), automated rollbacks, comprehensive monitoring, secure coding; integrate static/dynamic/security analysis into CI/CD; add pre/post-deployment health checks; design resilient architectures (circuit breakers, retries, graceful degradation).

**Goal**: low (elite: 0-15%). **Impact**: system stability, less downtime, more customer trust.

### 4. Mean Time to Recovery (MTTR)
**Definition**: time to restore service after degradation/outage - measures resilience/recovery.
- Low MTTR = fast detect/diagnose/resolve, minimized failure impact; relies on strong observability (monitoring, alerting, centralized logging, tracing)

**Copilot guidance**: implement clear monitoring/alerting (dashboards, automated anomaly notifications); recommend automated incident response and documented runbooks; advise efficient (one-click) rollback strategies; build observability in (structured logging, metrics, tracing); guide debugging via logs/metrics/traces first.

**Goal**: low (elite: <1 hour). **Impact**: minimized disruption, better customer satisfaction, operational confidence.

## Conclusion
DevOps is culture and continuous improvement driven by feedback and metrics, not just tools. Guide developers via CALMS and DORA improvement toward reliable, scalable, efficient delivery pipelines - every code change, infrastructure change, and pipeline modification should align with delivering high-quality software rapidly and reliably.

---

<!-- End of DevOps Core Principles Instructions -->
