#!/usr/bin/env node
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { CallToolRequestSchema, ListResourcesRequestSchema, ListToolsRequestSchema, ReadResourceRequestSchema, } from '@modelcontextprotocol/sdk/types.js';
import { ConstitutionalFramework } from './constitutional-framework.js';
import { ContextDetector } from './context-detector.js';
import { AgentOrchestrator } from './agent-orchestrator.js';
class SpecKitMCPServer {
    server;
    framework;
    contextDetector;
    agentOrchestrator;
    constructor() {
        this.server = new Server({
            name: 'spec-kit-mcp-server',
            version: '1.0.0',
        }, {
            capabilities: {
                resources: {},
                tools: {},
            },
        });
        this.framework = new ConstitutionalFramework();
        this.contextDetector = new ContextDetector();
        this.agentOrchestrator = new AgentOrchestrator();
        this.setupHandlers();
    }
    setupHandlers() {
        // List available resources
        this.server.setRequestHandler(ListResourcesRequestSchema, async () => {
            return {
                resources: [
                    {
                        uri: 'constitutional://core-principles',
                        name: 'Core Constitutional Principles',
                        description: 'Non-negotiable framework rules (TDD, Interface Design, etc.)',
                        mimeType: 'text/markdown',
                    },
                    {
                        uri: 'constitutional://context/security',
                        name: 'Security Constitutional Rules',
                        description: 'Security-specific constitutional principles',
                        mimeType: 'text/markdown',
                    },
                    {
                        uri: 'constitutional://context/fintech',
                        name: 'Fintech Constitutional Rules',
                        description: 'Financial technology constitutional principles',
                        mimeType: 'text/markdown',
                    },
                    {
                        uri: 'constitutional://context/healthcare',
                        name: 'Healthcare Constitutional Rules',
                        description: 'Healthcare-specific constitutional principles',
                        mimeType: 'text/markdown',
                    },
                    {
                        uri: 'chatmode://dev',
                        name: 'Developer Agent Mode',
                        description: 'Full stack developer persona configuration',
                        mimeType: 'text/markdown',
                    },
                    {
                        uri: 'chatmode://qa',
                        name: 'QA Agent Mode',
                        description: 'Quality assurance engineer persona configuration',
                        mimeType: 'text/markdown',
                    },
                    {
                        uri: 'chatmode://sa',
                        name: 'Solution Architect Agent Mode',
                        description: 'Solution architect persona configuration',
                        mimeType: 'text/markdown',
                    },
                ],
            };
        });
        // Read specific resource content
        this.server.setRequestHandler(ReadResourceRequestSchema, async (request) => {
            const { uri } = request.params;
            if (uri.startsWith('constitutional://')) {
                const content = await this.framework.getConstitutionalContent(uri);
                return {
                    contents: [
                        {
                            uri,
                            mimeType: 'text/markdown',
                            text: content,
                        },
                    ],
                };
            }
            if (uri.startsWith('chatmode://')) {
                const content = await this.agentOrchestrator.getChatModeContent(uri);
                return {
                    contents: [
                        {
                            uri,
                            mimeType: 'text/markdown',
                            text: content,
                        },
                    ],
                };
            }
            throw new Error(`Unknown resource: ${uri}`);
        });
        // List available tools
        this.server.setRequestHandler(ListToolsRequestSchema, async () => {
            return {
                tools: [
                    {
                        name: 'detect_and_load_context',
                        description: 'Auto-detect technology/domain and load appropriate constitutional context',
                        inputSchema: {
                            type: 'object',
                            properties: {
                                content: {
                                    type: 'string',
                                    description: 'Code or conversation content to analyze for context detection',
                                },
                                explicit_domains: {
                                    type: 'array',
                                    items: { type: 'string' },
                                    description: 'Explicitly specified domains to load (security, fintech, healthcare, etc.)',
                                },
                            },
                            required: ['content'],
                        },
                    },
                    {
                        name: 'switch_agent_persona',
                        description: 'Switch between different agent personas (PM, SA, Dev, QA, etc.)',
                        inputSchema: {
                            type: 'object',
                            properties: {
                                persona: {
                                    type: 'string',
                                    enum: ['pm', 'sa', 'dev', 'qa', 'devops', 'security', 'ba', 'po'],
                                    description: 'The agent persona to switch to',
                                },
                                context: {
                                    type: 'string',
                                    description: 'Additional context for the persona switch',
                                },
                            },
                            required: ['persona'],
                        },
                    },
                    {
                        name: 'validate_constitutional_compliance',
                        description: 'Validate code or design against constitutional principles',
                        inputSchema: {
                            type: 'object',
                            properties: {
                                code: {
                                    type: 'string',
                                    description: 'Code to validate against constitutional principles',
                                },
                                type: {
                                    type: 'string',
                                    enum: ['code', 'api', 'architecture', 'test'],
                                    description: 'Type of validation to perform',
                                },
                                domains: {
                                    type: 'array',
                                    items: { type: 'string' },
                                    description: 'Specific domains to validate against',
                                },
                            },
                            required: ['code', 'type'],
                        },
                    },
                ],
            };
        });
        // Handle tool calls
        this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
            const { name, arguments: args } = request.params;
            switch (name) {
                case 'detect_and_load_context':
                    return await this.contextDetector.detectAndLoadContext(args);
                case 'switch_agent_persona':
                    return await this.agentOrchestrator.switchPersona(args);
                case 'validate_constitutional_compliance':
                    return await this.framework.validateCompliance(args);
                default:
                    throw new Error(`Unknown tool: ${name}`);
            }
        });
    }
    async run() {
        const transport = new StdioServerTransport();
        await this.server.connect(transport);
        console.error('Spec-Kit MCP Server running on stdio');
    }
}
// Start the server
const server = new SpecKitMCPServer();
server.run().catch(console.error);
//# sourceMappingURL=index.js.map