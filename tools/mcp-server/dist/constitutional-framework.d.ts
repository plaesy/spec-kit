export declare class ConstitutionalFramework {
    private readonly frameworkRoot;
    constructor();
    getConstitutionalContent(uri: string): Promise<string>;
    private loadCoreConstitution;
    private loadDomainConstitution;
    validateCompliance(args: any): Promise<any>;
    private loadApplicableRules;
    private validateCodeCompliance;
    private validateApiCompliance;
    private validateArchitectureCompliance;
    private validateTestCompliance;
    private hasTests;
    private hasPublicMethods;
    private hasErrorHandling;
    private hasLogging;
    private hasUserInput;
    private hasInputValidation;
    private hasApiDocumentation;
    private hasVersioning;
    private hasHealthCheck;
    private hasPlatformSpecificCode;
    private hasTightCoupling;
    private isIntegrationTest;
    private hasMocks;
    private hasGivenWhenThen;
}
//# sourceMappingURL=constitutional-framework.d.ts.map