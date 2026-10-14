---
description: "Role-to-role handoff mapping for autonomous chaining"
applyTo: "**/*"
---

# Role-Handoff Mapping Reference

> **Purpose**: Route a role's finished work to the role built to take it next.
> This is the **canonical** role-routing table — `/implement`, `/fix`,
> `/optimize`, `/loop`, and `/continue` all point here instead of restating
> it. No other file in the framework may restate the per-role mapping.
>
> This is the role-level sibling of `.plaesy/instructions/dimension-mapping.md` (which routes
> *dimensions* to *commands*). This file routes *roles* to *roles*. The two
> are consulted together: a dimension finding picks the command, the command
> picks the first role, and this table picks every handoff after that.

Each role's own `## Boundaries & Escalation` section
(`.plaesy/roles/[role].md`) is the source of truth for its `Owns`/`Defers
to`/`Escalate to` relationships — this table is a transcription for lookup,
not a new design. If a role file changes its boundaries, update this table in
the same change.

---

## Routing Table

| Role | Owns | Defers to | Escalates to | Typical next role |
|---|---|---|---|---|
| **@ba** | Requirements elicitation/analysis, stakeholder requirements, user stories, business process/impact analysis | @pm (roadmap), @po (backlog/acceptance), @sa (technical shape), @bo (investment/ROI) | @bo — two validated requirements conflict and only a funding/pricing/priority call resolves it | @sa |
| **@sa** | System architecture, service boundaries, interface contracts, technology selection, NFR mapping, dev environment | @dev (implementation), @devops (CI/CD/IaC), @sre (operational readiness), @security (security architecture), @ai-architect (AI subsystems), @qa (test strategy) | @ai-architect — a service boundary is determined by model training/inference/governance, not domain logic | @dev |
| **@dev** | Application code, unit/integration test implementation, interface-contract implementation, refactoring, code review | @sa (structure/tech choice), @devops (pipelines/infra), @qa (test strategy/sign-off), @security (secure design); AI-facing prompt wording belongs to @pe/@ai-architect/@mlops | @sa — a change crosses a service boundary or requires a new technology choice | @qa |
| **@qa** | Test strategy, test suite implementation, coverage, defect reporting, quality metrics, quality sign-off | @security (pentest/threat models), @devsecops (pipeline gates), @sa (architecture review), @dev (fixing defects) | @sa — a defect's root cause is structural, not local | @devops |
| **@devops** | CI/CD pipelines, IaC, container/platform config, build/release automation, deployment strategy | @dev (application code), @sre (SLOs/alerting/incident response), @devsecops (security gates), @sa (what gets deployed) | @sre — the concern is production behaviour after deploy, not the pipeline | @sre |
| **@sre** | SLOs/SLIs, error budgets, observability, alerting, incident response/post-mortems, capacity planning, DR, internal dev platform | @devops (pipeline/IaC machinery), @security (incident forensics), @qa (performance validation), @sa (what is built), @pm (velocity vs. reliability) | @devops — the work is automating the pipeline/infra itself, not operating it | *(end of chain — reports status)* |
| **@designer** | UI/UX design, information/interaction architecture, design systems, usability research, visual design language | @accessibility (WCAG/conformance), @dev (implementation), @qa (usability test execution), @pm (flow prioritization) | @accessibility — a design decision breaks or cannot satisfy WCAG/AT conformance | @accessibility |
| **@accessibility** | WCAG/AT conformance, accessible component/interaction patterns, prioritized remediation, accessibility acceptance criteria | @designer (visual language), @dev (accessible markup/code), @qa (test execution), @privacy-legal (statutory obligations), @pm (prioritization) | @privacy-legal — a requirement is driven by statute (ADA, 508, EN 301 549, AODA) and its scope is disputed | @dev |
| **@pm** | Product vision/strategy, roadmap, prioritization frameworks, product metrics/KPIs, go-to-market | @po (backlog mechanics/sign-off), @bo (investment/ROI), @ba (requirements detail), @market-research-analyst (market evidence), @sm (sprint process) | @po — the work is ordering/refining existing backlog rather than setting direction | @po |
| **@po** | Backlog hygiene, story definition/sizing, acceptance criteria, story readiness, delivery-facing communication | @pm (strategy/roadmap), @ba (requirements analysis), @qa (acceptance verification), @dev (implementation feasibility) | @pm — backlog order or story scope requires changing product direction/roadmap | @dev |
| **@bo** | Business strategy, ROI/investment decisions, unit economics, organizational governance/policy, performance targets | @ba (requirements analysis), @pm (product strategy/roadmap), @sm (team process/capacity), @market-research-analyst (market evidence) | @pm — a strategic choice needs a roadmap decision to become real | @pm |
| **@market-research-analyst** | Market sizing (TAM/SAM/SOM), competitive landscape, trend/disruption analysis, segmentation, market-entry, source attribution | @pm (roadmap implication), @bo (investment/pricing), @ba (what to build), @tw (publishing findings) | @bo — findings imply an investment, pricing, or market-entry commitment | @pm |
| **@sm** | Ceremony facilitation, impediment surfacing/removal, team coaching, Agile practice health, scope-creep protection | @pm (priority/roadmap), @po (backlog order), @bo (organizational investment), @ba (requirements) | @bo — an impediment cannot be removed without budget, headcount, or a policy change | *(end of chain — reports status)* |
| **@security** | Threat modelling, vulnerability assessment, OWASP Top 10 validation, security-by-design, security code review, pentest scope/severity | @devsecops (pipeline/deploy enforcement), @compliance (regulatory mapping), @privacy-legal (data-protection law), @dev (remediation implementation) | @devsecops — the control must be enforced automatically in build/deploy | @dev |
| **@devsecops** | Security automation in CI/CD, hardened IaC, secret management, container/Kubernetes hardening, automated compliance gates, shift-left tooling | @security (threat models/pentest), @compliance (mandatory framework), @devops (pipeline plumbing), @privacy-legal (legal exposure) | @security — a finding requires exploitation analysis, pentesting, or threat modelling to confirm severity | @devops |
| **@compliance** | Regulatory framework selection, audit prep/evidence, audit-trail automation, automated policy enforcement, compliance risk scoring | @privacy-legal (privacy law/legal risk), @security (technical controls), @devsecops (pipeline automation), @ba (requirement analysis) | @privacy-legal — the obligation originates in privacy/data-protection law (GDPR, CCPA, HIPAA, PIPEDA) | @privacy-legal |
| **@privacy-legal** | Privacy law/legal basis, DPIAs, data processing agreements, consent management, data-subject rights, legal risk assessment, vendor agreement review | @compliance (audit mechanics), @security (technical controls), @data-engineer (governance implementation), @ba (legal requirements), @pm (risk/timeline trade-offs) | @security — a legal obligation needs a technical control implemented or verified | @security |
| **@data-engineer** | Data pipeline architecture, ETL/ELT, warehouse/lakehouse design, streaming/event-driven systems, data quality/monitoring, lineage/cataloging | @sa (consuming-service architecture), @dev (pipeline app code), @devops (infra provisioning), @compliance (retention/audit), @privacy-legal (lawful basis/DPIA) | @privacy-legal — a pipeline change moves, retains, combines, or re-identifies personal data | @sa |
| **@ai-architect** | AI/ML system architecture, model-in-the-loop service boundaries, AI governance/responsible-AI frameworks, hybrid human-AI workflow design | @mlops (ML platform/serving/retraining), @data-engineer (training/inference data pipelines), @sa (non-AI system parts), @security (model/data protection) | @mlops — the decision is about serving, scaling, monitoring, or retraining a model in production, not designing it | @mlops |
| **@mlops** | ML platform architecture, model lifecycle, training/validation/deployment pipelines, serving/auto-scaling, drift/performance monitoring, feature stores, experiment tracking | @ai-architect (ML system design/model selection), @data-engineer (upstream data pipelines), @sre (serving-infra reliability), @dev (ML integration code), @compliance (regulatory evidence) | @ai-architect — the decision is about model design, accuracy targets, or architecture the platform must support | @sre |
| **@tw** | Technical documentation content/accuracy, API references, user guides, runbooks, information architecture | @dev (implementation facts/code examples), @sa (architecture intent), @sre (runbook accuracy), @pe (system-prompt wording) | @dev — documenting a behaviour not present in the code or configuration | *(end of chain — reports status)* |
| **@pe** | Prompt structure, reasoning order, specificity, examples, the two-section output contract | @dev (the system the prompt drives), @pm/@po (what the prompt must achieve), @ai-architect (model capability constraints), @tw (documentation wording) | @dev — the improvement depends on code, config, or API behaviour that has not been read | *(invoked by name, never a handoff target — see Coverage table)* |
| **@nara** | Chief of Staff and final orchestrator: intakes every user request, routes it to specialist roles, runs council deliberation on contested choices, verifies their evidence, and is the only role that speaks to the user | every specialist role, for its domain analysis — Nara reads their returns and verifies evidence, never re-derives them | the user — only for explicit user instructions; Nara decides every other question, including rule 8 categories (secrets, destructive or irreversible actions, outward-facing actions, money/legal commitments), under the Constitution's NN articles | *(top of the chain — never a handoff target; see Coverage table)* |
| **@ideator** | Divergent option generation, each option with one stated assumption and one cheapest test | @critic (ranking), @market-research-analyst and @researcher (market facts), @verifier (citations), @nara (the choice) | @nara — the brief is too vague to produce distinct options | *(dispatched by Nara — never a handoff target; see Coverage table)* |
| **@critic** | Blind peer review of anonymized drafts against the stated criteria; names each draft's weakest claim | @verifier (whether a citation is real), @nara (synthesis and ties) | @nara — the criteria conflict so no draft scores consistently | *(dispatched by Nara — never a handoff target; see Coverage table)* |
| **@user-researcher** | Survey and interview instrument design, consent-text drafts, analysis of returned responses (drafts only, never contacts anyone) | @nara (whether and when to publish or contact), @privacy-legal and @compliance (consent and retention law) | @nara — the research needs personal data the user has not approved | *(dispatched by Nara — never a handoff target; see Coverage table)* |
| **@verifier** | Confirms each citation's claim text is in its source; marks each VERIFIED or UNVERIFIED | @nara (what to do with unverified claims), the citing role (the claim's meaning) | @nara — a verdict depends on unverified citations with no alternative source | *(dispatched by Nara — never a handoff target; see Coverage table)* |

---

## Handoff Section Registry (Role-Level)

Sibling to `.plaesy/instructions/dimension-mapping.md`'s `## Handoff: <dimension>` convention. Any
role's output may declare:

| Section | Emitted by | When |
|---------|-----------|------|
| `## Handoff: @<role>` | any role | This role's work is done and the Routing Table's "Typical next role" (or an `Escalates to` fork) names where it goes next. Recorded here and written to `state.json.role_handoff_log` — not silently implied. |
| `## Handoff: none (end of chain)` | any role | The role has no "Typical next role" (see table — e.g. `@sre`, `@sm`, `@tw`). |

---

## Priority: Role Chain vs. Nara

Every handoff goes through Nara. A role's output that touches a rule 8 category
(secrets, destructive actions, externally-visible actions, money/legal commitments)
is not a stop. Nara decides it under the Constitution's Non-Negotiables
(`.plaesy/memory/constitution.md`, NN-01..07) and the chain continues. The chain
has **no maximum depth or hop count** — see `.plaesy/instructions/plaesy.instructions.md` rule 8.

---

## Coverage Table

| Role | Status |
|---|---|
| All roles in the Routing Table above except `@nara`, `@pe`, `@ideator`, `@critic`, `@user-researcher`, `@verifier` | Loaded as a scope persona and participates in auto-chaining |
| `.plaesy/roles/nara.md` | **Top of the chain, not a chain participant.** Chief of Staff — the entry point for user requests; it dispatches the roles above and is never a `Typical next role` target. Its Decision Memo Mode still answers a genuinely ambiguous hard fork escalated from `/fix` or `/implement`. |
| `.plaesy/roles/ideator.md`, `critic.md`, `user-researcher.md`, `verifier.md` | **Not chain participants.** Dispatched only by Nara (see its Agent Roster). Never a `Typical next role` target. |
| `.plaesy/roles/pe.md` | **Not a chain participant.** Prompt Engineer — invoked by name when the artifact under review is a prompt. Never a `Typical next role` target. |

---

## See Also

- `.plaesy/instructions/plaesy.md` → §Assistant Behavior rules 2 and 8 — the
  default-and-record policy and the decision rights Nara holds under the Constitution's NN articles
- `.plaesy/instructions/dimension-mapping.md` — the dimension-level sibling of this table
  (dimension → command, where this file is role → role)
- `.plaesy/roles/[role].md` — each role's own `## Boundaries & Escalation`
  section, the source of truth this table transcribes

## Usage Example

```text

@ba finishes a spec (requirements + user stories).
→ Lookup row: @ba | Typical next role = @sa
→ Handoff: @sa (architecture/technical shape)
→ @sa finishes the architecture.
→ Lookup row: @sa | Typical next role = @dev
→ Handoff: @dev (implementation)
→ @dev finishes the code.
→ Lookup row: @dev | Typical next role = @qa
→ Handoff: @qa (test strategy/sign-off)
→ @qa finishes test sign-off.
→ Lookup row: @qa | Typical next role = @devops
→ Handoff: @devops (deployment)
→ @devops deploys.
→ Lookup row: @devops | Typical next role = @sre
→ Handoff: @sre (production operation)
→ @sre's row has no "Typical next role" → `## Handoff: none (end of chain)`,
  chain ends, status reported.
```
