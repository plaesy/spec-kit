---
description: "Legal and compliance assessment workflow - regulatory, data privacy, risk management"
---

# Legal/Compliance Assessment Instructions

**Use with**: `/assess:legal` when dimension is **Legal/Compliance**

**Referenced from**: `.plaesy/memory/assess.md` (orchestrator)

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

```
regulatory: 25%           # Regulatory requirements met, compliance verified
data_privacy: 25%         # GDPR/CCPA/HIPAA compliance, data handling, consent
risk_assessment: 15%      # Legal risks identified, IP protected, mitigated
security_compliance: 15%  # Security standards (ISO 27001, SOC 2), encryption
documentation: 15%        # Privacy policy, terms of service, compliance records
third_party: 5%          # Vendor compliance, DPA agreements, supply chain
```

### Grade Mapping

```
95-100: A+ (Exceptional - Full compliance, robust policies, strong controls)
90-94:  A  (Excellent - Compliant, minor documentation gaps)
85-89:  B+ (Very Good - Compliant, some documentation needed)
80-84:  B  (Good - Mostly compliant, some remediation needed)
70-79:  C  (Acceptable - Significant compliance gaps, action plan needed)
<70:    D  (Needs Work - Major compliance issues, immediate action required)
```

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

## Anti-Patterns (NEVER Do These)

- ❌ **Never speculate on legal interpretation** - Cite actual law/regulation
- ❌ **Never hide compliance violations** - Flag all identified issues
- ❌ **Never ignore data protection requirements** - GDPR/CCPA are non-negotiable
- ❌ **Never assume exemptions apply** - Verify explicitly
- ❌ **Never mix legal advice with assessment** - Recommend legal counsel review
- ❌ **Never ignore third-party compliance** - Vendor risk is real
- ❌ **Never skip incident response assessment** - Critical for compliance

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

```
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

## Success Criteria

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
