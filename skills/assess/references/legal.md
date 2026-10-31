---
description: "Legal and compliance assessment workflow - regulatory, data privacy, risk management"
applyTo: "**/*"
---

# Legal/Compliance Assessment Instructions

**Use with**: `/assess:legal` when dimension is **Legal/Compliance**

**Referenced from**: `/assess` (orchestrator)

---

## Assessment Workflow

### Step 1: Regulatory Compliance Audit

- Identify applicable regulations and standards
- Map regulatory requirements to implementation
- Verify compliance with relevant laws (GDPR, CCPA, HIPAA, SOX, etc.)
- Assess compliance documentation completeness
- Identify compliance gaps and violations

**Rate**: 0-100 based on compliance coverage

### Step 2: Data Privacy Assessment

- Audit data collection practices and consent mechanisms
- Review data retention and deletion policies
- Assess data security and encryption standards
- Verify personal data handling procedures
- Review third-party data sharing agreements

**Rate**: 0-100 based on privacy standards

### Step 3: Risk Assessment

- Identify legal risks and liability exposures
- Assess intellectual property (IP) protection
- Review licensing and open-source compliance
- Evaluate contract terms and liabilities
- Identify regulatory change risks

**Rate**: 0-100 based on risk mitigation

### Step 4: Security Compliance

- Verify security standards compliance (ISO 27001, SOC 2, etc.)
- Assess data breach response procedures
- Review incident reporting requirements
- Evaluate access control and authentication
- Assess vulnerability management processes

**Rate**: 0-100 based on security standards

### Step 5: Documentation Review

- Check privacy policy completeness and accuracy
- Review terms of service and legal agreements
- Verify compliance policy documentation
- Assess regulatory filing completeness
- Review audit trail and compliance records

**Rate**: 0-100 based on documentation quality

### Step 6: Third-Party Compliance

- Audit vendor and supplier agreements
- Verify subcontractor compliance requirements
- Review data processing agreements (DPA)
- Assess supply chain compliance risks
- Evaluate third-party security standards

**Rate**: 0-100 based on third-party compliance

### Step 7: Generate Report

**Report sections**: canonical names per `.plaesy/instructions/dimension-mapping.md`
→ *Assessment Report Section Registry*; the bullets below go *inside* those
sections and do not rename them.

- Overall compliance score (0-100)
- Per-category scores (Regulatory, Privacy, Risk, Security, Docs, Third-Party)
- Top 5 findings (ordered by severity)
- Compliance gaps and remediation steps

**Blocking Issues**:

- GDPR/CCPA non-compliance = BLOCKER
- Data breach response missing = BLOCKER
- Major IP infringement = BLOCKER
- Missing privacy policy = BLOCKER

---

## Quality Scoring

### Compliance Assessment Components

```text

regulatory: 25%           # Regulatory requirements met, compliance verified
data_privacy: 25%         # GDPR/CCPA/HIPAA compliance, data handling, consent
risk_assessment: 15%      # Legal risks identified, IP protected, mitigated
security_compliance: 15%  # Security standards (ISO 27001, SOC 2), encryption
documentation: 15%        # Privacy policy, terms of service, compliance records
third_party: 5%          # Vendor compliance, DPA agreements, supply chain
```

### Grade Mapping

| Grade | What it means here |
|---|---|
| A+ | Exceptional - Full compliance, robust policies, strong controls |
| A | Excellent - Compliant, minor documentation gaps |
| B+ | Very Good - Compliant, some documentation needed |
| B | Good - Mostly compliant, some remediation needed |
| C | Acceptable - Significant compliance gaps, action plan needed |
| D | Needs Work - Major compliance issues, immediate action required |

The score edges behind these letters, and the general meaning of each band,
are defined once in `.plaesy/instructions/quality-gates.md` →
**Grade bands**. Do not restate them here.

### Success Criteria (Hard Stops)

- MUST HAVE: GDPR/CCPA compliance (or documented exemption)
- MUST HAVE: Privacy policy and terms of service
- MUST HAVE: Data protection agreement with processors
- MUST HAVE: Incident response procedures documented
- MUST HAVE: No unaddressed critical compliance violations

---

## Critical Rules

- ✅ **CITE REGULATIONS** - Reference specific laws and standards
- ✅ **DOCUMENT EVIDENCE** - Every finding includes evidence/reference
- ✅ **NO SPECULATION** - Only report documented compliance gaps
- ✅ **RISK PRIORITIZATION** - Rate by business and legal impact
- ✅ **ACCURACY > COMPLETENESS** - Fewer verified findings > many guesses
- ✅ **LEGAL EXPERTISE** - Defer to legal counsel for interpretation

---

## Finding Standards

- **Cite the actual statute or regulation** behind every legal conclusion
- **Report every compliance violation found**, at the severity it carries
- **Assess GDPR/CCPA applicability explicitly** — they are not optional
- **Verify an exemption before relying on it**, in writing
- **State the finding and recommend counsel review** — this assessment is not
  legal advice
- **Assess vendor third-party compliance** as part of scope
- **Assess incident response readiness** — it is a compliance control, not an
  operational extra

---

## Regulatory Frameworks to Assess

**Global**:

- GDPR (EU) — Data protection, privacy
- CCPA (California) — Consumer privacy
- LGPD (Brazil) — Personal data protection

**Industry-Specific**:

- HIPAA (Healthcare) — Medical data protection
- PCI DSS (Payment) — Credit card data security
- SOX (Finance) — Financial controls
- FERPA (Education) — Student records

**Security Standards**:

- ISO 27001 — Information security
- SOC 2 — Service organization controls
- ISO 9001 — Quality management

---

## Pre-Assessment Validation

Run these checks BEFORE attempting assessment:

```text

Does organization have legal/compliance function?
  → yes: Interview or review documentation
  → no: Recommend legal counsel engagement

Are compliance requirements documented?
  → yes: Review documentation completeness
  → partial: Flag gaps, recommend legal review

Is data handling documented?
  → yes: Audit data practices
  → no: Critical gap, recommend immediate action
```

---

## Uncertainty Surfacing (Before Finalizing Assessment)

Load `.plaesy/instructions/uncertainty-surfacing.md`; do not restate it here.
Only this dimension's specifics follow.

**Confidence basis for this dimension**:

- HIGH: the specific clause, statute, or regulation was read and cited
- MEDIUM: the applicable framework is identified, but jurisdiction or interpretation is unconfirmed
- LOW: the answer depends on jurisdiction or contract text that was not supplied

**Worked example**:

```text
I assumed GDPR applies because the product serves EU residents; jurisdiction
was not stated.
If the product is EU-only in practice, the data-transfer assessment changes to
inapplicable and the DPIA requirement does not trigger.
```

---

## Assessment Completion Checklist

Assessment is complete when:

- ✅ Applicable regulations identified and mapped
- ✅ Data privacy practices assessed
- ✅ Legal risks identified and rated
- ✅ Security compliance verified
- ✅ Documentation reviewed for completeness
- ✅ Third-party compliance assessed
- ✅ Scores calculated per category
- ✅ Compliance gaps clearly identified with remediation steps
- ✅ Blocking issues flagged prominently
- ✅ Legal counsel review recommended where needed
- ✅ Console output delivered to user

## Persona Protocol

Loaded by every `each family:legal` persona prompt for this dimension — `/assess`,
`/implement`, `/fix`, `/optimize`, `/loop` and `/improve` all point their
`Loads:` line here. The rules below are dimension-neutral and are written here
once rather than restated in 54 persona files, where they had already drifted
apart from the parents they point at.

- **Follow the parent protocol.** The persona fixes the *dimension*; the family
  prompt fixes the *verb*. Follow the family prompt's full protocol. Where the
  parent and this file disagree, the parent governs the behaviour.
- **This file is the *assessment* criteria for `legal`** — workflow, weights,
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
