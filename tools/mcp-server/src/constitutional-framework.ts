import * as fs from 'fs-extra';
import path from 'path';
import { glob } from 'glob';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export class ConstitutionalFramework {
    private readonly frameworkRoot: string;

    constructor() {
        // Navigate up from tools/mcp-server to the framework root
        this.frameworkRoot = path.resolve(__dirname, '../../../');
    }

    async getConstitutionalContent(uri: string): Promise<string> {
        if (uri === 'constitutional://core-principles') {
            return await this.loadCoreConstitution();
        }

        if (uri.startsWith('constitutional://context/')) {
            const domain = uri.replace('constitutional://context/', '');
            return await this.loadDomainConstitution(domain);
        }

        throw new Error(`Unknown constitutional URI: ${uri}`);
    }

    private async loadCoreConstitution(): Promise<string> {
        const constitutionPath = path.join(this.frameworkRoot, 'memory/constitution.md');

        if (await fs.pathExists(constitutionPath)) {
            return await fs.readFile(constitutionPath, 'utf-8');
        }

        // Fallback to core principles if main constitution doesn't exist
        const corePrinciplesPath = path.join(this.frameworkRoot, 'memory/constitution/core-principles.constitution.md');

        if (await fs.pathExists(corePrinciplesPath)) {
            return await fs.readFile(corePrinciplesPath, 'utf-8');
        }

        throw new Error('Core constitutional principles not found');
    }

    private async loadDomainConstitution(domain: string): Promise<string> {
        const domainPath = path.join(this.frameworkRoot, `memory/constitution/${domain}.constitution.md`);

        if (await fs.pathExists(domainPath)) {
            return await fs.readFile(domainPath, 'utf-8');
        }

        // If specific domain doesn't exist, return guidance
        return `# ${domain.charAt(0).toUpperCase() + domain.slice(1)} Constitutional Principles

Domain-specific constitutional principles for **${domain}** are not yet defined.

## Default Constitutional Requirements

Based on core principles, all ${domain} implementations must follow:

1. **Test-First Development (TDD)**
   - RED: Write failing test first
   - GREEN: Write minimal code to pass
   - REFACTOR: Improve code while keeping tests green

2. **Interface Design**
   - Clear contracts with input/output validation
   - Consistent error handling and responses
   - Semantic versioning (MAJOR.MINOR.PATCH.BUILD)
   - Comprehensive API documentation

3. **Observability First**
   - Structured logging with correlation IDs
   - Metrics collection and monitoring
   - Health checks and readiness probes
   - Distributed tracing support

4. **Security by Design** (Critical for ${domain})
   - Authentication and authorization
   - Input validation and sanitization
   - Secure communication (HTTPS/TLS)
   - Audit logging and compliance

5. **Platform Agnostic**
   - Cross-platform compatibility
   - Container-ready deployment
   - Configuration externalization
   - Environment-specific settings

To create domain-specific constitutional principles, add them to:
\`memory/constitution/${domain}.constitution.md\`
    `;
    }

    async validateCompliance(args: any): Promise<any> {
        const { code, type, domains = [] } = args;

        const validationResults = {
            constitutional_compliance: true,
            violations: [] as string[],
            recommendations: [] as string[],
            domains_checked: domains,
        };

        // Load applicable constitutional rules
        const constitutionalRules = await this.loadApplicableRules(domains);

        // Perform validation based on type
        switch (type) {
            case 'code':
                this.validateCodeCompliance(code, constitutionalRules, validationResults);
                break;
            case 'api':
                this.validateApiCompliance(code, constitutionalRules, validationResults);
                break;
            case 'architecture':
                this.validateArchitectureCompliance(code, constitutionalRules, validationResults);
                break;
            case 'test':
                this.validateTestCompliance(code, constitutionalRules, validationResults);
                break;
        }

        return {
            content: [
                {
                    type: 'text',
                    text: JSON.stringify(validationResults, null, 2),
                },
            ],
        };
    }

    private async loadApplicableRules(domains: string[]): Promise<any> {
        const rules = {
            core: await this.loadCoreConstitution(),
            domains: {} as Record<string, string>,
        };

        for (const domain of domains) {
            try {
                rules.domains[domain] = await this.loadDomainConstitution(domain);
            } catch (error) {
                // Domain-specific rules not found, continue with core rules
            }
        }

        return rules;
    }

    private validateCodeCompliance(code: string, rules: any, results: any): void {
        // TDD Validation
        if (!this.hasTests(code)) {
            results.violations.push('Missing tests - TDD requires test-first development');
            results.constitutional_compliance = false;
        }

        // Interface Design Validation
        if (this.hasPublicMethods(code) && !this.hasErrorHandling(code)) {
            results.violations.push('Missing error handling - All public interfaces must handle errors');
            results.constitutional_compliance = false;
        }

        // Observability Validation
        if (!this.hasLogging(code)) {
            results.recommendations.push('Consider adding structured logging for observability');
        }

        // Security Validation
        if (this.hasUserInput(code) && !this.hasInputValidation(code)) {
            results.violations.push('Missing input validation - All user inputs must be validated');
            results.constitutional_compliance = false;
        }
    }

    private validateApiCompliance(code: string, rules: any, results: any): void {
        // API Contract Validation
        if (!this.hasApiDocumentation(code)) {
            results.violations.push('Missing API documentation - All APIs must be documented');
            results.constitutional_compliance = false;
        }

        // Versioning Validation
        if (!this.hasVersioning(code)) {
            results.violations.push('Missing versioning - All APIs must support semantic versioning');
            results.constitutional_compliance = false;
        }

        // Health Check Validation
        if (!this.hasHealthCheck(code)) {
            results.recommendations.push('Consider adding health check endpoints');
        }
    }

    private validateArchitectureCompliance(code: string, rules: any, results: any): void {
        // Platform Agnostic Validation
        if (this.hasPlatformSpecificCode(code)) {
            results.violations.push('Platform-specific code detected - Maintain platform agnostic design');
            results.constitutional_compliance = false;
        }

        // Service Independence Validation
        if (this.hasTightCoupling(code)) {
            results.recommendations.push('Consider reducing coupling between services');
        }
    }

    private validateTestCompliance(code: string, rules: any, results: any): void {
        // No Mocks Rule for Integration Tests
        if (this.isIntegrationTest(code) && this.hasMocks(code)) {
            results.violations.push('Mocks detected in integration tests - Use real dependencies only');
            results.constitutional_compliance = false;
        }

        // Test Structure Validation
        if (!this.hasGivenWhenThen(code)) {
            results.recommendations.push('Consider using Given-When-Then structure for better test clarity');
        }
    }

    // Helper methods for validation
    private hasTests(code: string): boolean {
        return /test|spec|describe|it\(/i.test(code);
    }

    private hasPublicMethods(code: string): boolean {
        return /public|export|api/i.test(code);
    }

    private hasErrorHandling(code: string): boolean {
        return /try|catch|throw|error|exception/i.test(code);
    }

    private hasLogging(code: string): boolean {
        return /log|logger|console\./i.test(code);
    }

    private hasUserInput(code: string): boolean {
        return /input|request|params|body|query/i.test(code);
    }

    private hasInputValidation(code: string): boolean {
        return /validate|schema|joi|zod|yup/i.test(code);
    }

    private hasApiDocumentation(code: string): boolean {
        return /swagger|openapi|@api|\/\*\*.*@param/i.test(code);
    }

    private hasVersioning(code: string): boolean {
        return /version|v\d+|\/api\/v/i.test(code);
    }

    private hasHealthCheck(code: string): boolean {
        return /health|status|ping|ready/i.test(code);
    }

    private hasPlatformSpecificCode(code: string): boolean {
        return /\.NET|\.net|windows|linux|macos|platform\.is/i.test(code);
    }

    private hasTightCoupling(code: string): boolean {
        return /import.*\.\.\/.*\.\.\/|require.*\.\.\/.*\.\.\//i.test(code);
    }

    private isIntegrationTest(code: string): boolean {
        return /integration.*test|test.*integration/i.test(code);
    }

    private hasMocks(code: string): boolean {
        return /mock|stub|fake|spy|jest\.mock/i.test(code);
    }

    private hasGivenWhenThen(code: string): boolean {
        return /given|when|then|arrange|act|assert/i.test(code);
    }
}