interface PersonaSwitchArgs {
    persona: string;
    context?: string;
}
export declare class AgentOrchestrator {
    private readonly frameworkRoot;
    private currentPersona;
    constructor();
    getChatModeContent(uri: string): Promise<string>;
    switchPersona(args: PersonaSwitchArgs): Promise<any>;
    private loadPersonaConfig;
    private parsePersonaConfig;
    private extractSectionContent;
    private getPersonaName;
    private generateTransitionGuidance;
    private getPersonaRecommendedActions;
    private getPersonaConstitutionalRequirements;
}
export {};
//# sourceMappingURL=agent-orchestrator.d.ts.map