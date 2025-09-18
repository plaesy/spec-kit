import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

interface ContextDetectionResult {
    detected_technologies: string[];
    detected_domains: string[];
    recommended_instructions: string[];
    recommended_chatmodes: string[];
    confidence_score: number;
}

export class ContextDetector {
    private readonly frameworkRoot: string;

    constructor() {
        // Get framework root - navigate up from tools/mcp-server
        this.frameworkRoot = path.resolve(__dirname, '../../../');
    }

    async detectAndLoadContext(args: any): Promise<any> {
        const { content, explicit_domains = [] } = args;

        const detection = this.analyzeContent(content, explicit_domains);
        const recommendations = await this.generateRecommendations(detection);

        return {
            content: [
                {
                    type: 'text',
                    text: JSON.stringify({
                        detection_results: detection,
                        recommendations,
                        instructions_to_load: recommendations.instructions,
                        chatmodes_to_load: recommendations.chatmodes,
                        constitutional_domains: recommendations.constitutional_domains,
                    }, null, 2),
                },
            ],
        };
    }

    private analyzeContent(content: string, explicitDomains: string[]): ContextDetectionResult {
        const technologies = this.detectTechnologies(content);
        const domains = this.detectDomains(content, explicitDomains);

        return {
            detected_technologies: technologies,
            detected_domains: domains,
            recommended_instructions: this.mapTechnologiesToInstructions(technologies),
            recommended_chatmodes: this.mapDomainsToAgents(domains),
            confidence_score: this.calculateConfidence(technologies, domains),
        };
    }

    private detectTechnologies(content: string): string[] {
        const techPatterns = {
            'react': /react|jsx|useState|useEffect|components?/i,
            'nextjs': /next\.js|next|getStaticProps|getServerSideProps|pages\/api/i,
            'angular': /angular|\@angular|ngOnInit|components?.*\.ts/i,
            'nestjs': /nestjs|\@nestjs|@Controller|@Service|@Injectable/i,
            'springboot': /spring.*boot|@SpringBootApplication|@RestController|@Service/i,
            'flutter': /flutter|dart|StatefulWidget|StatelessWidget/i,
            'go': /package main|func main|goroutine|go\s+func/i,
            'rust': /fn main|use std|cargo|impl.*for|trait/i,
            'csharp': /using System|class.*:|public.*static.*void|\.NET/i,
            'docker': /FROM.*:|COPY|RUN|EXPOSE|dockerfile/i,
            'kubernetes': /apiVersion|kind:|metadata:|spec:|kubectl/i,
            'typescript': /interface|type.*=|export.*type|\.ts$/i,
            'python': /def |import |from.*import|\.py$/i,
            'java': /public.*class|import.*java|\.java$/i,
        };

        const detected = [];
        for (const [tech, pattern] of Object.entries(techPatterns)) {
            if (pattern.test(content)) {
                detected.push(tech);
            }
        }

        return detected;
    }

    private detectDomains(content: string, explicitDomains: string[]): string[] {
        const domainPatterns = {
            'security': /security|auth|encrypt|vulnerability|owasp|cyber/i,
            'fintech': /payment|banking|financial|trading|blockchain|crypto|pci/i,
            'healthcare': /patient|medical|healthcare|hipaa|clinical/i,
            'testing': /test|testing|tdd|unit.*test|integration.*test/i,
            'performance': /performance|scalability|optimization|load.*test/i,
            'ai-ml': /ai|ml|machine.*learning|model|neural|tensorflow/i,
            'web3': /blockchain|smart.*contract|web3|solidity|ethereum/i,
        };

        const detected = [...explicitDomains];

        for (const [domain, pattern] of Object.entries(domainPatterns)) {
            if (pattern.test(content) && !detected.includes(domain)) {
                detected.push(domain);
            }
        }

        return detected;
    }

    private mapTechnologiesToInstructions(technologies: string[]): string[] {
        const techToInstructions: Record<string, string> = {
            'react': 'reactjs.instructions.md',
            'nextjs': 'nextjs.instructions.md',
            'angular': 'angular.instructions.md',
            'nestjs': 'nestjs.instructions.md',
            'springboot': 'springboot.instructions.md',
            'flutter': 'dart-n-flutter.instructions.md',
            'go': 'go.instructions.md',
            'rust': 'rust.instructions.md',
            'csharp': 'csharp.instructions.md',
            'kubernetes': 'kubernetes-deployment-best-practices.instructions.md',
            'java': 'java.instructions.md',
        };

        const instructions = ['context.instructions.md']; // Always load context awareness

        for (const tech of technologies) {
            if (techToInstructions[tech]) {
                instructions.push(techToInstructions[tech]);
            }
        }

        return [...new Set(instructions)];
    }

    private mapDomainsToAgents(domains: string[]): string[] {
        const domainToAgents: Record<string, string[]> = {
            'security': ['security.chatmode.md'],
            'testing': ['qa.chatmode.md'],
            'performance': ['devops.chatmode.md', 'sa.chatmode.md'],
            'fintech': ['security.chatmode.md', 'sa.chatmode.md'],
            'healthcare': ['security.chatmode.md', 'sa.chatmode.md'],
        };

        const agents = ['dev.chatmode.md']; // Default to dev mode

        for (const domain of domains) {
            if (domainToAgents[domain]) {
                agents.push(...domainToAgents[domain]);
            }
        }

        return [...new Set(agents)];
    }

    private calculateConfidence(technologies: string[], domains: string[]): number {
        const techWeight = Math.min(technologies.length * 0.3, 0.6);
        const domainWeight = Math.min(domains.length * 0.2, 0.4);

        return Math.min(techWeight + domainWeight, 1.0);
    }

    private async generateRecommendations(detection: ContextDetectionResult): Promise<any> {
        return {
            instructions: detection.recommended_instructions,
            chatmodes: detection.recommended_chatmodes,
            constitutional_domains: this.getConstitutionalDomains(detection.detected_domains),
            workflow_suggestions: this.getWorkflowSuggestions(detection),
            priority_order: this.getPriorityOrder(detection),
        };
    }

    private getConstitutionalDomains(domains: string[]): string[] {
        const constitutionalMapping: Record<string, string> = {
            'security': 'security.constitution.md',
            'fintech': 'fintech.constitution.md',
            'healthcare': 'healthcare.constitution.md',
            'testing': 'testing.constitution.md',
            'performance': 'performance.constitution.md',
            'ai-ml': 'ai-ml.constitution.md',
            'web3': 'web3-blockchain.constitution.md',
        };

        return domains
            .map(domain => constitutionalMapping[domain])
            .filter(Boolean);
    }

    private getWorkflowSuggestions(detection: ContextDetectionResult): string[] {
        const suggestions = [];

        if (detection.detected_technologies.length > 0) {
            suggestions.push('Load technology-specific instructions first');
        }

        if (detection.detected_domains.includes('security')) {
            suggestions.push('Apply security constitutional principles');
            suggestions.push('Use security chatmode for sensitive operations');
        }

        if (detection.detected_domains.includes('testing')) {
            suggestions.push('Enforce TDD constitutional requirements');
            suggestions.push('Switch to QA chatmode for test strategy');
        }

        if (detection.detected_technologies.includes('kubernetes')) {
            suggestions.push('Load DevOps best practices');
            suggestions.push('Apply deployment constitutional principles');
        }

        return suggestions;
    }

    private getPriorityOrder(detection: ContextDetectionResult): string[] {
        const order = [
            'core-principles.constitution.md', // Always first
            'context.instructions.md', // Always second
        ];

        // Add domain-specific constitutions based on criticality
        const criticalDomains = ['security', 'fintech', 'healthcare'];
        for (const domain of criticalDomains) {
            if (detection.detected_domains.includes(domain)) {
                order.push(`${domain}.constitution.md`);
            }
        }

        // Add technology instructions
        order.push(...detection.recommended_instructions.filter(i => i !== 'context.instructions.md'));

        // Add chatmodes
        order.push(...detection.recommended_chatmodes);

        return order;
    }
}