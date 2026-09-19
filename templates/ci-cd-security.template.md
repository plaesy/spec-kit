# CI/CD Security Pipeline Template

**Project**: [PROJECT_NAME]  
**Technology Stack**: [TECH_STACK]  
**Pipeline Platform**: [GITHUB_ACTIONS | GITLAB_CI | AZURE_DEVOPS | JENKINS]  
**Security Level**: [STANDARD | HIGH | CRITICAL]  

## Executive Summary

### Pipeline Overview
This CI/CD pipeline template implements security-first continuous integration and deployment with comprehensive security scanning, compliance validation, and quality gates aligned with constitutional framework principles.

### Security Integration Points
- **Static Analysis**: SAST scanning integrated into build process
- **Dependency Scanning**: Vulnerability detection in dependencies
- **Container Security**: Image scanning and hardening
- **Dynamic Testing**: DAST scanning in staging environments
- **Compliance Checks**: Automated compliance validation
- **Secrets Management**: Secure handling of sensitive data

## Pipeline Architecture

### Pipeline Stages Overview

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   SOURCE    │───▶│   BUILD     │───▶│    TEST     │───▶│   SECURITY  │
│             │    │             │    │             │    │             │
│ • Code      │    │ • Compile   │    │ • Unit      │    │ • SAST      │
│ • Secrets   │    │ • Dependencies │  │ • Integration │  │ • DAST      │
│ • Config    │    │ • Artifacts │    │ • E2E       │    │ • Containers │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │                   │
       ▼                   ▼                   ▼                   ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ COMPLIANCE  │───▶│   STAGING   │───▶│ PRODUCTION  │───▶│ MONITORING  │
│             │    │             │    │             │    │             │
│ • Policy    │    │ • Deploy    │    │ • Deploy    │    │ • Alerts    │
│ • Audit     │    │ • Validate  │    │ • Validate  │    │ • Metrics   │
│ • Reports   │    │ • Smoke Test │    │ • Rollback  │    │ • Logs      │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

### Security Gates and Quality Controls

| Stage | Security Controls | Quality Gates | Compliance Checks |
|-------|-------------------|---------------|-------------------|
| Source | Branch protection, signed commits | Pre-commit hooks | License scanning |
| Build | Dependency scanning, SBOM generation | Build success | Vulnerability thresholds |
| Test | Test coverage analysis | >80% coverage | Security test results |
| Security | SAST/DAST scanning | Zero critical vulnerabilities | Compliance validation |
| Compliance | Policy validation | Audit trail complete | Regulatory requirements |
| Staging | Runtime security monitoring | Performance benchmarks | Environment compliance |
| Production | Zero-downtime deployment | Health checks pass | Security monitoring active |
| Monitoring | Continuous security monitoring | SLA compliance | Ongoing audit |

## GitHub Actions Implementation

### Main Workflow (`.github/workflows/ci-cd-security.yml`)

```yaml
name: CI/CD Security Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 2 * * *'  # Daily security scan

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}
  NODE_VERSION: '18'
  SECURITY_SCAN_ENABLED: true
  COMPLIANCE_LEVEL: 'high'

jobs:
  security-scan:
    name: Security Scanning
    runs-on: ubuntu-latest
    permissions:
      contents: read
      security-events: write
      actions: read
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history for better analysis
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
      
      # Secrets Scanning
      - name: GitLeaks Secrets Scan
        uses: gitleaks/gitleaks-action@v2
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GITLEAKS_LICENSE: ${{ secrets.GITLEAKS_LICENSE }}
      
      # Static Application Security Testing (SAST)
      - name: CodeQL Analysis
        uses: github/codeql-action/init@v3
        with:
          languages: javascript, typescript
          config-file: ./.github/codeql/codeql-config.yml
      
      - name: Install dependencies
        run: npm ci
      
      - name: Build application
        run: npm run build
      
      - name: Perform CodeQL Analysis
        uses: github/codeql-action/analyze@v3
        with:
          category: "/language:${{matrix.language}}"
      
      # Dependency Vulnerability Scanning
      - name: npm audit
        run: npm audit --audit-level=high
      
      - name: Snyk Security Scan
        uses: snyk/actions/node@master
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          args: --severity-threshold=high
      
      # SBOM Generation
      - name: Generate SBOM
        uses: anchore/sbom-action@v0
        with:
          path: ./
          format: spdx-json
          output-file: "${{ github.event.repository.name }}-sbom.spdx.json"
      
      - name: Upload SBOM
        uses: actions/upload-artifact@v4
        with:
          name: sbom
          path: "${{ github.event.repository.name }}-sbom.spdx.json"
      
      # Container Security Scanning (if applicable)
      - name: Build Docker image
        if: env.CONTAINER_BUILD == 'true'
        run: |
          docker build -t ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }} .
      
      - name: Trivy Container Scan
        if: env.CONTAINER_BUILD == 'true'
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: '${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}'
          format: 'sarif'
          output: 'trivy-results.sarif'
      
      - name: Upload Trivy results
        if: env.CONTAINER_BUILD == 'true'
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: 'trivy-results.sarif'

  build-and-test:
    name: Build and Test
    runs-on: ubuntu-latest
    needs: security-scan
    
    strategy:
      matrix:
        node-version: [16, 18, 20]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Node.js ${{ matrix.node-version }}
        uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Lint code
        run: npm run lint
      
      - name: Type check
        run: npm run type-check
      
      - name: Run unit tests
        run: npm run test:unit -- --coverage
      
      - name: Run integration tests
        run: npm run test:integration
      
      - name: Run E2E tests
        run: npm run test:e2e
      
      - name: Upload coverage reports
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage/lcov.info
          flags: unittests
          name: codecov-umbrella
      
      - name: SonarCloud Scan
        uses: SonarSource/sonarcloud-github-action@master
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}

  compliance-check:
    name: Compliance Validation
    runs-on: ubuntu-latest
    needs: [security-scan, build-and-test]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: License Compliance Check
        uses: fossa-contrib/fossa-action@v2
        with:
          api-key: ${{ secrets.FOSSA_API_KEY }}
      
      - name: GDPR Compliance Check
        run: |
          # Custom script to validate GDPR compliance
          ./scripts/check-gdpr-compliance.sh
      
      - name: Security Policy Validation
        run: |
          # Validate security policies are implemented
          ./scripts/validate-security-policies.sh
      
      - name: Generate compliance report
        run: |
          # Generate comprehensive compliance report
          ./scripts/generate-compliance-report.sh
      
      - name: Upload compliance artifacts
        uses: actions/upload-artifact@v4
        with:
          name: compliance-reports
          path: ./reports/compliance/

  deploy-staging:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    needs: [security-scan, build-and-test, compliance-check]
    if: github.ref == 'refs/heads/develop'
    environment: staging
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Build for staging
        run: npm run build:staging
        env:
          NODE_ENV: staging
      
      - name: Deploy to staging
        run: |
          # Deploy application to staging environment
          ./scripts/deploy-staging.sh
        env:
          STAGING_API_KEY: ${{ secrets.STAGING_API_KEY }}
          STAGING_DATABASE_URL: ${{ secrets.STAGING_DATABASE_URL }}
      
      - name: Run smoke tests
        run: npm run test:smoke:staging
      
      - name: DAST Security Scan
        uses: zaproxy/action-full-scan@v0.7.0
        with:
          target: 'https://staging.example.com'
          rules_file_name: '.zap/rules.tsv'
          cmd_options: '-a'
      
      - name: Performance testing
        run: |
          # Run performance tests against staging
          npm run test:performance:staging

  deploy-production:
    name: Deploy to Production
    runs-on: ubuntu-latest
    needs: [deploy-staging]
    if: github.ref == 'refs/heads/main'
    environment: production
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Build for production
        run: npm run build:production
        env:
          NODE_ENV: production
      
      - name: Pre-deployment security check
        run: |
          # Final security validation before production
          ./scripts/pre-deploy-security-check.sh
      
      - name: Blue-Green Deployment
        run: |
          # Implement blue-green deployment strategy
          ./scripts/deploy-production-blue-green.sh
        env:
          PRODUCTION_API_KEY: ${{ secrets.PRODUCTION_API_KEY }}
          PRODUCTION_DATABASE_URL: ${{ secrets.PRODUCTION_DATABASE_URL }}
      
      - name: Health check validation
        run: |
          # Validate application health after deployment
          ./scripts/validate-production-health.sh
      
      - name: Setup monitoring alerts
        run: |
          # Configure production monitoring and alerting
          ./scripts/setup-monitoring.sh
        env:
          MONITORING_API_KEY: ${{ secrets.MONITORING_API_KEY }}
      
      - name: Deployment notification
        uses: 8398a7/action-slack@v3
        with:
          status: custom
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
          custom_payload: |
            {
              "text": "Production deployment successful",
              "attachments": [{
                "color": "good",
                "fields": [{
                  "title": "Version",
                  "value": "${{ github.sha }}",
                  "short": true
                }]
              }]
            }

  security-monitoring:
    name: Continuous Security Monitoring
    runs-on: ubuntu-latest
    needs: [deploy-production]
    if: github.ref == 'refs/heads/main'
    
    steps:
      - name: Setup security monitoring
        run: |
          # Configure continuous security monitoring
          ./scripts/setup-security-monitoring.sh
        env:
          SECURITY_MONITORING_API_KEY: ${{ secrets.SECURITY_MONITORING_API_KEY }}
      
      - name: Runtime security validation
        run: |
          # Validate runtime security configurations
          ./scripts/validate-runtime-security.sh
      
      - name: Generate security metrics
        run: |
          # Generate security metrics and KPIs
          ./scripts/generate-security-metrics.sh
```

### Security Configuration Files

#### CodeQL Configuration (`.github/codeql/codeql-config.yml`)

```yaml
name: "Security Analysis"

queries:
  - uses: security-extended
  - uses: security-and-quality

paths-ignore:
  - node_modules
  - dist
  - coverage
  - '**/*.test.js'
  - '**/*.spec.js'

paths:
  - src
  - api
  - lib

disable-default-path-filters: false
```

#### ZAP Rules Configuration (`.zap/rules.tsv`)

```
10001	IGNORE	(Informational)
SQL Injection	FAIL	(High)
Cross Site Scripting (Reflected)	FAIL	(High)
Cross Site Scripting (Persistent)	FAIL	(High)
Path Traversal	FAIL	(High)
Remote File Inclusion	FAIL	(High)
Remote Code Execution	FAIL	(Critical)
External Redirect	WARN	(Medium)
Cookie No HttpOnly Flag	WARN	(Low)
Cookie Without Secure Flag	WARN	(Low)
```

## Deployment Scripts

### Pre-deployment Security Check (`scripts/pre-deploy-security-check.sh`)

```bash
#!/bin/bash
set -e

echo "🔒 Running pre-deployment security checks..."

# Check for secrets in environment variables
echo "Validating environment variables..."
if [ -z "$PRODUCTION_API_KEY" ]; then
    echo "❌ PRODUCTION_API_KEY not set"
    exit 1
fi

# Validate SSL certificates
echo "Validating SSL certificates..."
./scripts/validate-ssl-certs.sh

# Check security headers configuration
echo "Validating security headers..."
./scripts/validate-security-headers.sh

# Validate database security settings
echo "Validating database security..."
./scripts/validate-database-security.sh

# Check for required security middleware
echo "Validating security middleware..."
./scripts/validate-security-middleware.sh

echo "✅ Pre-deployment security checks passed"
```

### Blue-Green Deployment (`scripts/deploy-production-blue-green.sh`)

```bash
#!/bin/bash
set -e

echo "🚀 Starting blue-green deployment..."

# Determine current and new environments
CURRENT_ENV=$(./scripts/get-current-environment.sh)
if [ "$CURRENT_ENV" == "blue" ]; then
    NEW_ENV="green"
else
    NEW_ENV="blue"
fi

echo "Current environment: $CURRENT_ENV"
echo "Deploying to: $NEW_ENV"

# Deploy to new environment
echo "Deploying application to $NEW_ENV environment..."
./scripts/deploy-to-environment.sh $NEW_ENV

# Run health checks
echo "Running health checks on $NEW_ENV environment..."
./scripts/health-check.sh $NEW_ENV

if [ $? -eq 0 ]; then
    echo "Health checks passed. Switching traffic to $NEW_ENV..."
    ./scripts/switch-traffic.sh $NEW_ENV
    
    # Verify traffic switch
    sleep 30
    ./scripts/verify-traffic-switch.sh $NEW_ENV
    
    if [ $? -eq 0 ]; then
        echo "✅ Blue-green deployment successful"
        echo "Cleaning up old environment: $CURRENT_ENV"
        ./scripts/cleanup-environment.sh $CURRENT_ENV
    else
        echo "❌ Traffic switch verification failed. Rolling back..."
        ./scripts/switch-traffic.sh $CURRENT_ENV
        exit 1
    fi
else
    echo "❌ Health checks failed. Deployment aborted."
    ./scripts/cleanup-environment.sh $NEW_ENV
    exit 1
fi
```

### Security Monitoring Setup (`scripts/setup-security-monitoring.sh`)

```bash
#!/bin/bash
set -e

echo "🔍 Setting up security monitoring..."

# Configure SIEM integration
echo "Configuring SIEM integration..."
curl -X POST "https://siem.example.com/api/v1/sources" \
  -H "Authorization: Bearer $SECURITY_MONITORING_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production-app",
    "type": "application",
    "endpoints": ["https://api.example.com"]
  }'

# Setup intrusion detection
echo "Configuring intrusion detection..."
./scripts/configure-ids.sh

# Configure WAF rules
echo "Updating WAF rules..."
./scripts/update-waf-rules.sh

# Setup vulnerability scanning schedule
echo "Scheduling vulnerability scans..."
./scripts/schedule-vulnerability-scans.sh

# Configure security alerts
echo "Setting up security alerts..."
./scripts/configure-security-alerts.sh

echo "✅ Security monitoring setup complete"
```

## Security Policies and Compliance

### Security Policy Validation (`scripts/validate-security-policies.sh`)

```bash
#!/bin/bash
set -e

echo "🛡️ Validating security policies..."

# Check password policy implementation
echo "Validating password policies..."
grep -r "password.*complexity" src/ || {
    echo "❌ Password complexity requirements not found"
    exit 1
}

# Validate encryption at rest
echo "Validating encryption at rest..."
grep -r "encrypt.*at.*rest" config/ || {
    echo "❌ Encryption at rest not configured"
    exit 1
}

# Check session management
echo "Validating session management..."
grep -r "session.*timeout" src/ || {
    echo "❌ Session timeout not configured"
    exit 1
}

# Validate audit logging
echo "Validating audit logging..."
grep -r "audit.*log" src/ || {
    echo "❌ Audit logging not implemented"
    exit 1
}

# Check access control implementation
echo "Validating access control..."
grep -r "rbac\|role.*based" src/ || {
    echo "❌ Role-based access control not found"
    exit 1
}

echo "✅ Security policy validation passed"
```

### GDPR Compliance Check (`scripts/check-gdpr-compliance.sh`)

```bash
#!/bin/bash
set -e

echo "🇪🇺 Checking GDPR compliance..."

# Check for data processing consent
echo "Validating consent management..."
grep -r "consent" src/ || {
    echo "❌ Consent management not implemented"
    exit 1
}

# Validate data subject rights implementation
echo "Validating data subject rights..."
grep -r "data.*deletion\|right.*erasure" src/ || {
    echo "❌ Data deletion/erasure not implemented"
    exit 1
}

# Check for data portability
echo "Validating data portability..."
grep -r "data.*export\|portability" src/ || {
    echo "❌ Data portability not implemented"
    exit 1
}

# Validate privacy by design
echo "Validating privacy by design..."
grep -r "privacy.*design\|data.*minimization" src/ || {
    echo "❌ Privacy by design principles not found"
    exit 1
}

# Check for breach notification procedures
echo "Validating breach notification..."
test -f "docs/incident-response-plan.md" || {
    echo "❌ Incident response plan not found"
    exit 1
}

echo "✅ GDPR compliance check passed"
```

## Monitoring and Alerting

### Security Metrics Collection (`scripts/generate-security-metrics.sh`)

```bash
#!/bin/bash
set -e

echo "📊 Generating security metrics..."

# Create metrics output directory
mkdir -p reports/security-metrics

# Collect vulnerability metrics
echo "Collecting vulnerability metrics..."
curl -s "https://security-api.example.com/vulnerabilities" \
  -H "Authorization: Bearer $SECURITY_API_KEY" \
  > reports/security-metrics/vulnerabilities.json

# Collect compliance metrics
echo "Collecting compliance metrics..."
curl -s "https://compliance-api.example.com/status" \
  -H "Authorization: Bearer $COMPLIANCE_API_KEY" \
  > reports/security-metrics/compliance.json

# Generate security dashboard data
echo "Generating security dashboard data..."
node scripts/generate-security-dashboard.js

# Calculate security score
echo "Calculating security score..."
python scripts/calculate-security-score.py

echo "✅ Security metrics generated"
```

### Alert Configuration (`scripts/configure-security-alerts.sh`)

```bash
#!/bin/bash
set -e

echo "🚨 Configuring security alerts..."

# Configure critical vulnerability alerts
echo "Setting up critical vulnerability alerts..."
curl -X POST "https://alerts.example.com/api/v1/rules" \
  -H "Authorization: Bearer $ALERTS_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Critical Vulnerabilities",
    "condition": "vulnerability.severity == critical",
    "action": "immediate_notification",
    "channels": ["slack", "email", "pagerduty"]
  }'

# Configure intrusion detection alerts
echo "Setting up intrusion detection alerts..."
curl -X POST "https://alerts.example.com/api/v1/rules" \
  -H "Authorization: Bearer $ALERTS_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Intrusion Detected",
    "condition": "intrusion.detected == true",
    "action": "immediate_notification",
    "channels": ["slack", "email", "pagerduty"]
  }'

# Configure compliance violation alerts
echo "Setting up compliance violation alerts..."
curl -X POST "https://alerts.example.com/api/v1/rules" \
  -H "Authorization: Bearer $ALERTS_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Compliance Violation",
    "condition": "compliance.status == violated",
    "action": "notification",
    "channels": ["slack", "email"]
  }'

echo "✅ Security alerts configured"
```

## Constitutional Framework Integration

### TDD Security Testing

```javascript
// Example security test that should be written before implementation
describe('Security Controls', () => {
  describe('Authentication', () => {
    it('should enforce strong password requirements', () => {
      // Test password complexity requirements
      const weakPassword = 'password123';
      const strongPassword = 'P@ssw0rd!2024#';
      
      expect(validatePassword(weakPassword)).toBe(false);
      expect(validatePassword(strongPassword)).toBe(true);
    });
    
    it('should implement session timeout', () => {
      // Test session timeout functionality
      const session = createSession();
      jest.advanceTimersByTime(SESSION_TIMEOUT);
      
      expect(session.isValid()).toBe(false);
    });
  });
  
  describe('Data Protection', () => {
    it('should encrypt sensitive data at rest', () => {
      // Test data encryption
      const sensitiveData = 'user-personal-info';
      const encrypted = encryptData(sensitiveData);
      
      expect(encrypted).not.toBe(sensitiveData);
      expect(decryptData(encrypted)).toBe(sensitiveData);
    });
    
    it('should implement data deletion for GDPR compliance', () => {
      // Test data deletion functionality
      const userId = 'user-123';
      const result = deleteUserData(userId);
      
      expect(result.success).toBe(true);
      expect(result.deletedRecords).toBeGreaterThan(0);
    });
  });
});
```

### Interface Contracts for Security

```typescript
// Security service interface contract
interface SecurityService {
  authenticate(credentials: Credentials): Promise<AuthResult>;
  authorize(user: User, resource: Resource): Promise<boolean>;
  encryptData(data: string): Promise<string>;
  auditLog(event: AuditEvent): Promise<void>;
  validateCompliance(): Promise<ComplianceStatus>;
}

// Security configuration contract
interface SecurityConfig {
  passwordPolicy: PasswordPolicy;
  sessionTimeout: number;
  encryptionSettings: EncryptionConfig;
  auditSettings: AuditConfig;
  complianceSettings: ComplianceConfig;
}
```

## Best Practices and Guidelines

### Security Development Lifecycle (SDL)

1. **Threat Modeling**: Conduct threat modeling for all new features
2. **Secure Coding**: Follow secure coding guidelines and standards
3. **Security Testing**: Implement comprehensive security testing
4. **Security Review**: Mandatory security review for all changes
5. **Incident Response**: Maintain incident response procedures
6. **Continuous Monitoring**: Implement continuous security monitoring

### Compliance Framework

1. **GDPR**: Implement privacy by design and data protection
2. **SOC 2**: Maintain security, availability, and confidentiality
3. **ISO 27001**: Implement information security management
4. **OWASP**: Follow OWASP security guidelines and practices
5. **PCI DSS**: Implement payment card security standards

### Quality Gates

- [ ] All security tests pass
- [ ] Zero critical vulnerabilities
- [ ] Compliance requirements validated
- [ ] Security policies implemented
- [ ] Monitoring and alerting configured
- [ ] Incident response procedures tested
- [ ] Security documentation complete
- [ ] Security training completed

---

**Document Control**
- **Version**: 1.0
- **Created**: [Date]
- **Last Modified**: [Date]
- **Security Level**: [Classification]
- **Owner**: [Security Team]
