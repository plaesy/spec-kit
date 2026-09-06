# Threat Model: {System/Component Name}

**System**: {system name} · **Version**: {version} · **Date**: {YYYY-MM-DD}
**Security Team**: {team members} · **Review Status**: {Draft | In Review | Approved | Needs Update}

Approach: STRIDE (Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege), plus LINDDUN for privacy where personal data is involved.

## 1. Executive Summary

{Brief overview of the system and key security considerations}

## 2. Scope and Architecture

- **Components**: {component1}, {component2}
- **Data flows**: {flow1}, {flow2}
- **Trust boundaries**: {boundary1}, {boundary2}
- **External dependencies**: {dependency1}, {dependency2}

## 3. Data Classification

| Type | Sensitivity | Examples |
|------|-------------|----------|
| {data type} | Public | {examples} |
| {data type} | Internal | {examples} |
| {data type} | Confidential | {examples} |
| {data type} | Restricted | {examples} |

## 4. Threat Enumeration (STRIDE)

| ID | Category | Description | Component | Attack Vector | Likelihood | Impact | Current Controls | Residual Risk |
|----|----------|--------------|-----------|----------------|------------|--------|-------------------|----------------|
| S001 | Spoofing | | | | Low/Med/High | Low/Med/High/Critical | | |
| T001 | Tampering | | | | | | | |
| R001 | Repudiation | | | | | | | |
| I001 | Info Disclosure | | | | | | | |
| D001 | Denial of Service | | | | | | | |
| E001 | Elevation of Privilege | | | | | | | |

## 5. Privacy Threats (LINDDUN)

| ID | Category | Description | Data | Likelihood | Impact | Mitigation |
|----|----------|--------------|------|------------|--------|------------|

## 6. Abuse/Misuse Cases

- {Scenario}: {defense/mitigation}

## 7. Security Controls

**Current**: {control} — type: preventive/detective/corrective, effectiveness: Low/Med/High, coverage: {what it protects}

**Required**: {control} — addresses {threat IDs}, priority {Low/Med/High/Critical}, effort {Small/Medium/Large}, owner {team}

## 8. Risk Ranking and Treatment

- Critical: {count} · High: {count} · Medium: {count} · Low: {count} (unmitigated critical/high called out separately)
- **Overall posture**: {Poor/Fair/Good/Excellent} — **Recommendation**: {Block/Review/Approve with conditions/Approve}
- Top risk scenarios ranked by score, each with threat IDs, impact, likelihood, and required mitigation

## 9. Compliance Mapping

Applicable regimes (GDPR, ISO 27001, SOC 2, PCI-DSS, ...) with relevant requirements/controls and compliance status per regime.

## 10. Implementation Plan

- **Critical (0-30d)**: {mitigations}
- **High (30-90d)**: {mitigations}
- **Medium (90-180d)**: {mitigations}
- **Low (180d+)**: {mitigations}

## 11. Monitoring & Incident Response

- **Monitoring**: SIEM integration, log sources, correlation rules, alerting/escalation, threat intel feeds
- **Incident response**: playbooks, response team, communication plan, recovery procedures

## 12. Validation

- Security testing, code review, penetration testing performed against the threats above

## 13. Review & Approval

- **Review cadence**: security review / threat model update / risk assessment / control effectiveness
- **Approvals**: security architect, CISO, business owner — name, date, status

## Appendix

- Attack trees, data flow diagrams, network diagrams, reference materials (link or embed as needed)

---
*Template version: 2025.1 · Created {timestamp} · Last updated {timestamp}*
