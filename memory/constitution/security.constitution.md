---
version: "1.0.0"
updatedAt: "2025-09-16T07:38:00Z"
type: "security-constitution"
contextTriggers: ["security", "owasp", "authentication", "authorization", "encryption", "vulnerability", "cybersecurity"]
description: "Security-first development practices and OWASP compliance requirements"
---

# Security & OWASP Constitutional Standards
Comprehensive security requirements and OWASP compliance for secure software development.

## Security Requirements (OWASP Compliant)

### OWASP Top 10 Compliance
- **A01: Broken Access Control**: Implement principle of least privilege, deny by default
- **A02: Cryptographic Failures**: Use Argon2/bcrypt for passwords, AES-256 for data at rest, TLS 1.3+ for transit
- **A03: Injection**: Parameterized queries mandatory, input validation at every boundary
- **A04: Insecure Design**: Security-first design patterns, threat modeling required
- **A05: Security Misconfiguration**: Secure defaults, disable debug in production
- **A06: Vulnerable Components**: Automated dependency scanning, updates within 7 days
- **A07: Authentication Failures**: MFA support, secure session management, rate limiting
- **A08: Data Integrity Failures**: Secure deserialization, digital signatures for critical data
- **A09: Logging Failures**: Security event logging, audit trails, no sensitive data in logs
- **A10: SSRF**: Allowlist validation for external requests, network segmentation

### Mandatory Security Controls
Modern security-first approach required:

#### Zero Trust Architecture
- **Identity Verification**: Multi-factor authentication mandatory
- **Least Privilege**: Role-based access with just-in-time permissions
- **Network Segmentation**: Micro-segmentation with service-to-service encryption
- **Continuous Validation**: Real-time risk assessment and adaptive access

#### Privacy-by-Design
- **Data Minimization**: Collect only necessary data with explicit consent
- **Purpose Limitation**: Use data only for stated purposes
- **Retention Policies**: Automated data lifecycle management and deletion
- **Consent Management**: Granular consent with easy withdrawal mechanisms

#### Threat Modeling & Risk Management
- **STRIDE Analysis**: Systematic threat identification for all components
- **Attack Surface Mapping**: Comprehensive inventory of potential entry points
- **Security Champions**: Dedicated security advocates per team
- **Incident Response**: Automated detection with defined escalation procedures

#### Implementation Requirements
- HTTPS mandatory in all non-development environments
- Environment-based secret management (no hardcoded credentials)
- Input validation at every boundary (API, CLI, file upload)
- Authentication tokens with proper expiration and refresh
- CORS configuration restricted to known origins
- SQL injection prevention via parameterized queries only
- File upload restrictions (type, size, malware scanning)
- Security headers enforced: CSP, HSTS, X-Frame-Options, X-Content-Type-Options
- Regular security testing (SAST, DAST, dependency scanning)
- Supply chain security with dependency signing and verification

### Advanced Security Practices

#### Cryptographic Standards
- **Encryption at Rest**: AES-256 with proper key rotation
- **Encryption in Transit**: TLS 1.3 minimum, certificate pinning for mobile
- **Key Management**: Hardware Security Modules (HSM) for production
- **Password Security**: Argon2id or bcrypt with appropriate work factors
- **Digital Signatures**: ECDSA or EdDSA for critical operations
- **Random Number Generation**: Cryptographically secure random generators

#### Authentication & Authorization
- **Multi-Factor Authentication**: TOTP, WebAuthn, or SMS as fallback
- **OAuth 2.1**: Latest OAuth specification with PKCE
- **JWT Security**: Proper validation, short expiration, secure storage
- **Session Management**: Secure session tokens, proper invalidation
- **Role-Based Access Control**: Granular permissions with inheritance
- **API Security**: Rate limiting, API keys, request signing

#### Security Testing Integration
- **Static Analysis**: Automated vulnerability scanning in CI/CD
- **Dynamic Testing**: Runtime security testing with authentication flows
- **Dependency Scanning**: Third-party vulnerability assessment
- **Infrastructure Testing**: Configuration and compliance validation
- **Zero-Trust Validation**: Continuous identity and access verification
- **Supply Chain Security**: End-to-end artifact integrity validation

### Compliance & Regulatory Standards

#### Industry-Specific Security
- **Healthcare**: HIPAA compliance, PHI protection
- **Finance**: PCI DSS, SOX compliance, regulatory reporting
- **Government**: FedRAMP, FISMA, security clearance requirements
- **International**: GDPR, CCPA, regional privacy laws

#### Security Monitoring & Response
- **Security Information and Event Management (SIEM)**: Centralized logging and alerting
- **Intrusion Detection**: Network and host-based monitoring
- **Incident Response**: Automated response workflows, forensic capabilities
- **Threat Intelligence**: Integration with threat feeds and indicators
- **Security Metrics**: Continuous security posture measurement
- **Penetration Testing**: Regular third-party security assessments

---

**Type**: Security Constitution  
**Context Triggers**: Security-related keywords  
**Mandatory**: For security-sensitive applications  
**Version**: 1.0.0