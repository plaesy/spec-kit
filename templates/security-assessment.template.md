# Security Assessment Template

## Security Assessment for: [Feature/Project Name]

**Assessment Date**: [Date]  
**Assessed By**: [Security Engineer Name]  
**Project Phase**: [Design/Development/Pre-Production/Production]  
**Risk Level**: [Low/Medium/High/Critical]

---

## 1. Executive Summary

### Overview
Brief description of the feature/system being assessed and its security posture.

### Key Findings
- **Critical Issues**: [Number] - Issues requiring immediate attention
- **High Issues**: [Number] - Issues requiring prompt attention  
- **Medium Issues**: [Number] - Issues requiring attention within sprint
- **Low Issues**: [Number] - Issues for future consideration

### Recommendation Summary
High-level recommendations for addressing identified security concerns.

---

## 2. Scope and Methodology

### Assessment Scope
- **Components Assessed**: [List of systems, applications, APIs, databases]
- **Assessment Type**: [Code Review/Architecture Review/Penetration Test/Compliance Audit]
- **Environment**: [Development/Staging/Production]
- **Time Frame**: [Assessment duration]

### Methodology
- [ ] Threat modeling (STRIDE methodology)
- [ ] Static Application Security Testing (SAST)
- [ ] Dynamic Application Security Testing (DAST)
- [ ] Dependency vulnerability scanning
- [ ] Infrastructure security review
- [ ] Configuration security review
- [ ] Code review for security issues
- [ ] Compliance requirements validation

---

## 3. Threat Model

### Assets
List of valuable assets that need protection:
- **Data Assets**: [User data, financial data, intellectual property]
- **System Assets**: [Servers, databases, APIs, third-party services]
- **Business Assets**: [Reputation, compliance, revenue]

### Threat Agents
Potential attackers and their motivations:
- **External Attackers**: [Hackers, competitors, nation-states]
- **Internal Threats**: [Malicious insiders, compromised accounts]
- **Accidental Threats**: [Human error, system failures]

### Attack Vectors
Potential attack methods:
- **Network-based**: [SQL injection, XSS, CSRF, DDoS]
- **Application-based**: [Authentication bypass, privilege escalation]
- **Infrastructure-based**: [Server misconfiguration, unpatched systems]
- **Social Engineering**: [Phishing, pretexting, social manipulation]

### STRIDE Analysis
| Component | Spoofing | Tampering | Repudiation | Info Disclosure | DoS | Elevation |
|-----------|----------|-----------|-------------|-----------------|-----|-----------|
| [Component 1] | [Risk Level] | [Risk Level] | [Risk Level] | [Risk Level] | [Risk Level] | [Risk Level] |
| [Component 2] | [Risk Level] | [Risk Level] | [Risk Level] | [Risk Level] | [Risk Level] | [Risk Level] |

---

## 4. Security Requirements

### Authentication Requirements
- [ ] Multi-factor authentication (MFA) implemented
- [ ] Strong password policy enforced
- [ ] Account lockout mechanisms in place
- [ ] Session management secure
- [ ] OAuth/SAML integration where appropriate

### Authorization Requirements
- [ ] Role-based access control (RBAC) implemented
- [ ] Principle of least privilege applied
- [ ] Resource-level permissions defined
- [ ] Regular access reviews conducted

### Data Protection Requirements
- [ ] Data encryption at rest
- [ ] Data encryption in transit (TLS 1.3)
- [ ] Key management system implemented
- [ ] Data classification and handling procedures
- [ ] Data retention and disposal policies

### Input Validation Requirements
- [ ] Server-side input validation
- [ ] Output encoding/escaping
- [ ] SQL injection prevention
- [ ] XSS prevention measures
- [ ] File upload security controls

### Logging and Monitoring Requirements
- [ ] Security event logging implemented
- [ ] Log integrity protection
- [ ] Real-time security monitoring
- [ ] Incident response procedures
- [ ] Compliance logging requirements

---

## 5. Vulnerability Assessment

### Critical Vulnerabilities
| ID | Vulnerability | Component | CVSS Score | Impact | Recommendation |
|----|---------------|-----------|------------|--------|----------------|
| CRIT-001 | [Description] | [Component] | [Score] | [Impact] | [Recommendation] |

### High Vulnerabilities
| ID | Vulnerability | Component | CVSS Score | Impact | Recommendation |
|----|---------------|-----------|------------|--------|----------------|
| HIGH-001 | [Description] | [Component] | [Score] | [Impact] | [Recommendation] |

### Medium Vulnerabilities
| ID | Vulnerability | Component | CVSS Score | Impact | Recommendation |
|----|---------------|-----------|------------|--------|----------------|
| MED-001 | [Description] | [Component] | [Score] | [Impact] | [Recommendation] |

### Low Vulnerabilities
| ID | Vulnerability | Component | CVSS Score | Impact | Recommendation |
|----|---------------|-----------|------------|--------|----------------|
| LOW-001 | [Description] | [Component] | [Score] | [Impact] | [Recommendation] |

---

## 6. Compliance Assessment

### Regulatory Requirements
- [ ] **GDPR** - General Data Protection Regulation
  - Data protection impact assessment completed
  - Privacy by design principles implemented
  - Data subject rights mechanisms in place
  - Data processing lawful basis established

- [ ] **SOX** - Sarbanes-Oxley Act
  - Financial data controls implemented
  - Access controls for financial systems
  - Audit trail requirements met
  - Change management controls

- [ ] **HIPAA** - Health Insurance Portability and Accountability Act
  - Protected health information (PHI) safeguards
  - Business associate agreements in place
  - Minimum necessary access controls
  - Breach notification procedures

- [ ] **PCI DSS** - Payment Card Industry Data Security Standard
  - Cardholder data protection
  - Network security controls
  - Regular security testing
  - Information security policy

### Industry Standards
- [ ] **ISO 27001** - Information Security Management
- [ ] **NIST Cybersecurity Framework**
- [ ] **OWASP Top 10** - Web Application Security
- [ ] **CIS Controls** - Critical Security Controls

---

## 7. Security Controls Assessment

### Technical Controls
- [ ] **Encryption**: [Status and recommendations]
- [ ] **Access Controls**: [Status and recommendations]
- [ ] **Network Security**: [Status and recommendations]
- [ ] **Application Security**: [Status and recommendations]
- [ ] **Database Security**: [Status and recommendations]

### Administrative Controls
- [ ] **Security Policies**: [Status and recommendations]
- [ ] **Security Training**: [Status and recommendations]
- [ ] **Incident Response**: [Status and recommendations]
- [ ] **Business Continuity**: [Status and recommendations]
- [ ] **Vendor Management**: [Status and recommendations]

### Physical Controls
- [ ] **Data Center Security**: [Status and recommendations]
- [ ] **Workstation Security**: [Status and recommendations]
- [ ] **Media Protection**: [Status and recommendations]
- [ ] **Environmental Controls**: [Status and recommendations]

---

## 8. Risk Assessment

### Risk Matrix
| Risk ID | Description | Likelihood | Impact | Risk Level | Mitigation Status |
|---------|-------------|------------|--------|------------|-------------------|
| RISK-001 | [Description] | [H/M/L] | [H/M/L] | [Critical/High/Medium/Low] | [Status] |

### Risk Treatment Plan
| Risk ID | Treatment Strategy | Owner | Target Date | Status |
|---------|-------------------|-------|-------------|--------|
| RISK-001 | [Accept/Mitigate/Transfer/Avoid] | [Owner] | [Date] | [Status] |

---

## 9. Security Testing Results

### Static Analysis (SAST)
- **Tool Used**: [Tool name and version]
- **Scan Date**: [Date]
- **Issues Found**: [Number by severity]
- **False Positives**: [Number]
- **Remediation Status**: [Status]

### Dynamic Analysis (DAST)
- **Tool Used**: [Tool name and version]
- **Scan Date**: [Date]
- **Issues Found**: [Number by severity]
- **Coverage**: [Percentage of application tested]
- **Remediation Status**: [Status]

### Penetration Testing
- **Testing Scope**: [Scope description]
- **Testing Date**: [Date range]
- **Methodology**: [OWASP Testing Guide, PTES, etc.]
- **Key Findings**: [Summary of findings]
- **Proof of Concept**: [If applicable]

### Dependency Scanning
- **Tool Used**: [Tool name and version]
- **Scan Date**: [Date]
- **Vulnerable Dependencies**: [Number]
- **Critical Vulnerabilities**: [Number]
- **Update Status**: [Status]

---

## 10. Recommendations and Remediation

### Immediate Actions (Critical/High Priority)
1. **[Recommendation 1]**
   - **Priority**: Critical/High
   - **Effort**: [Time estimate]
   - **Owner**: [Responsible party]
   - **Target Date**: [Date]

2. **[Recommendation 2]**
   - **Priority**: Critical/High
   - **Effort**: [Time estimate]
   - **Owner**: [Responsible party]
   - **Target Date**: [Date]

### Medium-Term Actions
1. **[Recommendation 1]**
   - **Priority**: Medium
   - **Effort**: [Time estimate]
   - **Owner**: [Responsible party]
   - **Target Date**: [Date]

### Long-Term Actions
1. **[Recommendation 1]**
   - **Priority**: Low
   - **Effort**: [Time estimate]
   - **Owner**: [Responsible party]
   - **Target Date**: [Date]

---

## 11. Security Monitoring and Incident Response

### Monitoring Requirements
- [ ] Security Information and Event Management (SIEM) integration
- [ ] Intrusion Detection System (IDS) configuration
- [ ] Security metrics and KPIs defined
- [ ] Automated alerting for security events
- [ ] Regular security reviews scheduled

### Incident Response Plan
- [ ] Incident response team identified
- [ ] Escalation procedures defined
- [ ] Communication protocols established
- [ ] Recovery procedures documented
- [ ] Post-incident review process

---

## 12. Security Training and Awareness

### Required Training
- [ ] Secure coding practices
- [ ] Security awareness training
- [ ] Incident response procedures
- [ ] Compliance requirements
- [ ] Tool-specific security training

### Training Schedule
| Training Topic | Target Audience | Frequency | Next Due Date |
|----------------|-----------------|-----------|---------------|
| [Topic] | [Audience] | [Frequency] | [Date] |

---

## 13. Conclusion and Next Steps

### Security Posture Summary
[Overall assessment of the security posture and readiness for next phase]

### Key Metrics
- **Security Score**: [Score out of 100]
- **Critical Issues**: [Number remaining]
- **Compliance Status**: [Percentage compliant]
- **Risk Level**: [Overall risk level]

### Next Steps
1. [Action item 1]
2. [Action item 2]
3. [Action item 3]

### Follow-up Schedule
- **Next Assessment**: [Date]
- **Progress Review**: [Date]
- **Compliance Audit**: [Date]

---

## Appendices

### Appendix A: Security Tools and Configurations
[Details about security tools used and their configurations]

### Appendix B: Compliance Documentation
[Links to relevant compliance documentation and evidence]

### Appendix C: Security Policies and Procedures
[References to applicable security policies and procedures]

### Appendix D: Technical Details
[Detailed technical findings and evidence]

---

**Assessment Approval**

**Security Engineer**: _________________________ Date: _________

**Technical Lead**: __________________________ Date: _________

**Product Owner**: __________________________ Date: _________