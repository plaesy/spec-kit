export declare class ContextDetector {
    private readonly frameworkRoot;
    constructor();
    detectAndLoadContext(args: any): Promise<any>;
    private analyzeContent;
    private detectTechnologies;
    private detectDomains;
    private mapTechnologiesToInstructions;
    private mapDomainsToAgents;
    private calculateConfidence;
    private generateRecommendations;
    private getConstitutionalDomains;
    private getWorkflowSuggestions;
    private getPriorityOrder;
}
//# sourceMappingURL=context-detector.d.ts.map