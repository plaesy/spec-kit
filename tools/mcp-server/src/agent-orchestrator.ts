import { promises as fs } from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

interface PersonaSwitchArgs {
    persona: string;
    context?: string;
}

interface AgentPersona {
    name: string;
    description: string;
    capabilities: string[];
    tools: string[];
    behavior: string;
    deliverables: string[];
}

export class AgentOrchestrator {
    private readonly frameworkRoot: string;
    private currentPersona: string = 'dev';

    constructor() {
        // Get framework root relative to this file location
        this.frameworkRoot = path.resolve(__dirname, '../../../');
    }

    async getChatModeContent(uri: string): Promise<string> {
        const persona = uri.replace('chatmode://', '');
        const chatModePath = path.join(this.frameworkRoot, 'chatmodes', `${persona}.chatmode.md`);

        try {
            return await fs.readFile(chatModePath, 'utf-8');
        } catch (error) {
            throw new Error(`Chat mode not found: ${persona}`);
        }
    }

    async switchPersona(args: PersonaSwitchArgs): Promise<any> {
        const { persona, context = '' } = args;

        // Validate persona exists
        const chatModePath = path.join(this.frameworkRoot, 'chatmodes', `${persona}.chatmode.md`);

        try {
            await fs.access(chatModePath);
        } catch (error) {
            throw new Error(`Invalid persona: ${persona}. Available personas: pm, sa, dev, qa, devops, security, ba, po`);
        }

        const previousPersona = this.currentPersona;
        this.currentPersona = persona;

        // Load the persona configuration
        const personaConfig = await this.loadPersonaConfig(persona);

        // Generate transition guidance
        const transitionGuidance = await this.generateTransitionGuidance(
            previousPersona,
            persona,
            context,
            personaConfig
        );

        return {
            content: [
                {
                    type: 'text',
                    text: JSON.stringify({
                        previous_persona: previousPersona,
                        current_persona: persona,
                        transition_context: context,
                        persona_configuration: personaConfig,
                        transition_guidance: transitionGuidance,
                        constitutional_requirements: await this.getPersonaConstitutionalRequirements(persona),
                    }, null, 2),
                },
            ],
        };
    }

    private async loadPersonaConfig(persona: string): Promise<AgentPersona> {
        const content = await this.getChatModeContent(`chatmode://${persona}`);

        // Parse the chatmode markdown file to extract configuration
        const config = this.parsePersonaConfig(content, persona);
        return config;
    }

    private parsePersonaConfig(content: string, persona: string): AgentPersona {
        // Extract YAML frontmatter if present
        const frontmatterMatch = content.match(/^---\n(.*?)\n---/s);
        let description = '';
        let tools: string[] = [];

        if (frontmatterMatch) {
            try {
                // Simple YAML parsing for description and tools
                const frontmatter = frontmatterMatch[1];
                const descMatch = frontmatter.match(/description:\s*"([^"]+)"/);
                const toolsMatch = frontmatter.match(/tools:\s*\[(.*?)\]/s);

                if (descMatch) description = descMatch[1];
                if (toolsMatch) {
                    tools = toolsMatch[1]
                        .split(',')
                        .map(t => t.trim().replace(/['"]/g, ''))
                        .filter(t => t.length > 0);
                }
            } catch (error) {
                // Fallback if YAML parsing fails
            }
        }

        // Extract capabilities and deliverables from content
        const capabilities = this.extractSectionContent(content, 'Key Capabilities');
        const deliverables = this.extractSectionContent(content, 'Deliverables');
        const behavior = this.extractSectionContent(content, 'Persona Behavior')[0] || '';

        return {
            name: this.getPersonaName(persona),
            description,
            capabilities,
            tools,
            behavior,
            deliverables,
        };
    }

    private extractSectionContent(content: string, sectionName: string): string[] {
        const sectionRegex = new RegExp(`## ${sectionName}[\\s\\S]*?(?=##|$)`, 'i');
        const sectionMatch = content.match(sectionRegex);

        if (!sectionMatch) return [];

        const sectionContent = sectionMatch[0];
        const listItems = sectionContent.match(/^- \*\*(.*?)\*\*: (.*?)$/gm) || [];

        return listItems.map(item => {
            const match = item.match(/^- \*\*(.*?)\*\*: (.*?)$/);
            return match ? `${match[1]}: ${match[2]}` : item.replace(/^- \*\*/, '').replace(/\*\*:/, ':');
        });
    }

    private getPersonaName(persona: string): string {
        const names: Record<string, string> = {
            'pm': 'Product Manager',
            'sa': 'Solution Architect',
            'dev': 'Full Stack Developer',
            'qa': 'Quality Assurance Engineer',
            'devops': 'DevOps Engineer',
            'security': 'Security Engineer',
            'ba': 'Business Analyst',
            'po': 'Product Owner',
        };

        return names[persona] || persona.toUpperCase();
    }

    private async generateTransitionGuidance(
        fromPersona: string,
        toPersona: string,
        context: string,
        personaConfig: AgentPersona
    ): Promise<any> {
        const transitions: Record<string, Record<string, string>> = {
            'pm': {
                'sa': 'Transitioning from product strategy to technical architecture. Focus on translating business requirements into technical solutions.',
                'dev': 'Moving from product planning to implementation. Ensure development aligns with product goals and user stories.',
                'qa': 'Shifting from product definition to quality validation. Define acceptance criteria and test scenarios.',
            },
            'sa': {
                'dev': 'Transitioning from architecture design to implementation. Follow architectural patterns and design principles.',
                'devops': 'Moving from solution design to deployment strategy. Focus on scalability and operational requirements.',
                'security': 'Shifting from general architecture to security architecture. Apply security by design principles.',
            },
            'dev': {
                'qa': 'Transitioning from implementation to quality assurance. Provide context on code changes and potential test scenarios.',
                'devops': 'Moving from development to deployment. Ensure code is deployment-ready with proper configurations.',
                'sa': 'Shifting from implementation details to architectural review. Discuss technical decisions and patterns used.',
            },
            'qa': {
                'dev': 'Transitioning from testing to development. Provide feedback on issues found and suggested improvements.',
                'devops': 'Moving from quality assurance to deployment validation. Focus on production readiness and monitoring.',
            },
        };

        const guidance = transitions[fromPersona]?.[toPersona] ||
            `Transitioning from ${this.getPersonaName(fromPersona)} to ${this.getPersonaName(toPersona)} context.`;

        return {
            transition_message: guidance,
            context_carryover: context,
            recommended_actions: this.getPersonaRecommendedActions(toPersona),
            focus_areas: personaConfig.capabilities.slice(0, 3), // Top 3 capabilities
            deliverable_expectations: personaConfig.deliverables,
        };
    }

    private getPersonaRecommendedActions(persona: string): string[] {
        const actions: Record<string, string[]> = {
            'pm': [
                'Review and prioritize requirements',
                'Define acceptance criteria',
                'Validate business value alignment',
            ],
            'sa': [
                'Design technical architecture',
                'Define interface contracts',
                'Validate scalability requirements',
            ],
            'dev': [
                'Implement features following TDD',
                'Write comprehensive tests',
                'Follow coding standards and patterns',
            ],
            'qa': [
                'Design test scenarios and cases',
                'Validate functional requirements',
                'Perform integration testing with real dependencies',
            ],
            'devops': [
                'Setup deployment pipelines',
                'Configure monitoring and observability',
                'Ensure security and compliance',
            ],
            'security': [
                'Perform security assessment',
                'Validate authentication and authorization',
                'Review for security vulnerabilities',
            ],
        };

        return actions[persona] || ['Perform role-specific activities'];
    }

    private async getPersonaConstitutionalRequirements(persona: string): Promise<string[]> {
        const requirements: Record<string, string[]> = {
            'pm': [
                'Define clear acceptance criteria',
                'Ensure business value alignment',
                'Validate user story completeness',
            ],
            'sa': [
                'Design with interface contracts',
                'Ensure platform agnostic architecture',
                'Plan for observability and monitoring',
            ],
            'dev': [
                'Follow TDD (RED-GREEN-REFACTOR)',
                'Implement proper error handling',
                'Write comprehensive tests before code',
            ],
            'qa': [
                'Use real dependencies in integration tests',
                'Follow Given-When-Then test structure',
                'Validate all acceptance criteria',
            ],
            'devops': [
                'Implement health checks',
                'Configure structured logging',
                'Ensure deployment automation',
            ],
            'security': [
                'Apply security by design principles',
                'Validate input sanitization',
                'Implement proper authentication',
            ],
        };

        return requirements[persona] || ['Follow core constitutional principles'];
    }
}