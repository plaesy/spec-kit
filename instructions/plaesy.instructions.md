# Main Instructions

## 🏛️ **Constitutional Priority (LOAD FIRST)**
**MANDATORY**: Always load `.plaesy/memory/constitution.md` first for non-negotiable framework rules, then `.plaesy/instructions/context.instructions.md` for dynamic context awareness.

## Framework Overview
This is a **constitutional development framework** that enforces disciplined development practices through automation, AI orchestration, and constitutional compliance. It operates across any technology stack with platform-agnostic principles.

## 🔒 **Constitutional Enforcement (NON-NEGOTIABLE)**

### Interface Design Requirements (Universal)
Every library/service MUST expose:
- **Clear Contract**: Explicit input/output with validation (OpenAPI, GraphQL schema)
- **Error Handling**: Consistent error responses with meaningful context
- **Versioning**: Semantic versioning (MAJOR.MINOR.PATCH.BUILD) with backward compatibility
- **Documentation**: API docs with examples, error cases, and SDK generation
- **Observability**: Structured logging, metrics, health checks, distributed tracing

### Integration Testing Standards (Real Dependencies Only)
- **No Mocks Rule**: Use real dependencies, not mocks for external services (constitutional requirement)
- **Test Order Priority**: Contract → Integration → E2E → Unit
- **Coverage Requirements**: All API contracts, inter-service communication, auth flows
- **BDD Integration**: Given-When-Then structure for acceptance criteria

## 🧠 **AI-Powered Context Loading**

### Auto-Detection Patterns
Load instructions automatically based on conversation context:

**Technology Keywords** → Load corresponding instructions:
- React, JSX, hooks → `.plaesy/instructions/reactjs.instructions.md`
- Next.js, SSR, routing → `.plaesy/instructions/nextjs.instructions.md`
- Angular, TypeScript → `.plaesy/instructions/angular.instructions.md`
- NestJS, decorators → `.plaesy/instructions/nestjs.instructions.md`
- Spring Boot, Java → `.plaesy/instructions/springboot.instructions.md`
- Flutter, Dart → `.plaesy/instructions/dart-n-flutter.instructions.md`
- Go, goroutines → `.plaesy/instructions/go.instructions.md`
- Rust, ownership → `.plaesy/instructions/rust.instructions.md`
- C#, .NET → `.plaesy/instructions/csharp.instructions.md`
- Docker, K8s → `.plaesy/instructions/kubernetes-deployment-best-practices.instructions.md`
- Security, OWASP → `.plaesy/instructions/security-and-owasp.instructions.md`
- SQL → `.plaesy/instructions/sql.instructions.md`

**Role-Based Context** → Load corresponding chatmodes:
- Product strategy, roadmap → `chatmodes/pm.chatmode.md` + `chatmodes/bo.chatmode.md`
- Architecture, system design → `chatmodes/sa.chatmode.md`
- Implementation, coding → `chatmodes/dev.chatmode.md`
- Testing, QA → `chatmodes/qa.chatmode.md` + `checklists/qa.checklist.md`
- DevOps, deployment → `chatmodes/devops.chatmode.md`
- Security review → `chatmodes/security.chatmode.md`

**Intent Recognition** → Load appropriate workflows:
- "Start new", "Initialize" → `prompts/idea.prompt.md`
- "Plan feature", "Design" → `prompts/plan.prompt.md` + `chatmodes/sa.chatmode.md`
- "Requirements", "Specification" → `prompts/specify.prompt.md` + `chatmodes/ba.chatmode.md`
- "Break down", "Tasks" → `prompts/tasks.prompt.md` + `chatmodes/dev.chatmode.md`

## 🏗️ **Framework Structure & Workflow**

### Core Components
```
memory/constitution.md      # NON-NEGOTIABLE constitutional rules
instructions/context.instructions.md  # Dynamic context loading patterns
chatmodes/                  # Role-specific AI agent templates
prompts/                    # Structured workflow templates (idea→specify→plan→tasks)
templates/                  # Content structure blueprints
scripts/                    # Automation tools for feature creation
checklists/                 # Quality gates and validation steps
```

### Constitutional Development Pipeline
1. **Ideation** → `prompts/idea.prompt.md`: Capture and structure initial concept
2. **Specification** → `prompts/specify.prompt.md`: Define business requirements with acceptance criteria
3. **Planning** → `prompts/plan.prompt.md`: Technical architecture with constitutional compliance
4. **Task Breakdown** → `prompts/tasks.prompt.md`: Implementation tasks with requirements
5. **Implementation** → Follow cycle with real-time compliance checking

## 🚀 **Advanced Capabilities**

### Platform Agnostic Design
- **Cross-Platform**: Works across any language (Python, JavaScript, Go, Rust, Java, C#, etc.)
- **Multi-Environment**: Web, mobile, desktop, embedded, serverless
- **Integration Patterns**: APIs, databases, message queues, real-time streams

### AI Orchestration
- **Dynamic Model Selection**: Match AI capabilities to task complexity
- **Constitutional Learning**: AI agents learn framework compliance patterns
- **Real-Time Collaboration**: Live pair programming with constitutional validation
- **Predictive Development**: Anticipate issues and suggest proactive solutions

### Quality Enforcement
- **Real-Time Compliance**: Live constitutional adherence monitoring
- **Automatic Validation**: Code review automation with constitutional checks
- **Multi-Tier Testing**: Property-based, integration, contract testing
- **Breaking Change Management**: Semantic versioning with migration plans

## 🎯 **Operational Guidelines**

### Session Initialization (Execute in Order)
1. **Constitutional Loading**: Load `memory/constitution.md` for non-negotiable framework rules
2. **Context Awareness**: Load `.plaesy/instructions/context.instructions.md` for dynamic patterns
3. **Technology Detection**: Analyze conversation for tech keywords (React, Go, K8s, etc.)
4. **Role Identification**: Detect role context (PM, SA, Dev, QA, DevOps, Security)
5. **Intent Recognition**: Identify workflow intent (ideation, planning, implementation)
6. **Auto-Load Context**: Load relevant instructions and chatmodes based on detection
7. **TDD Activation**: Always load `.plaesy/instructions/tdd-enforcement.instructions.md` for coding tasks
8. **Constitutional Validation**: Apply constitutional principles to all interactions

### Development Workflow (Constitutional Compliance)
- **Always enforce TDD cycle** with `.plaesy/instructions/tdd-enforcement.instructions.md`
- **Use automation scripts** from `.plaesy/scripts/` for feature creation and consistency
- **Follow structured templates** from `.plaesy/templates/` for consistent deliverables
- **Validate against quality checklists** from `.plaesy/checklists/` before completion
- **Document constitutional reasoning** for architectural and design decisions
- **Monitor real-time compliance** through automated validation hooks

### Multi-Agent Collaboration Patterns
- **Context Switching**: Dynamically switch between chatmodes based on task requirements
- **Knowledge Transfer**: Maintain context across agent transitions (PM → SA → Dev → QA)
- **Constitutional Consistency**: Ensure all agents follow the same constitutional principles
- **Quality Gates**: Use checklists for handoffs between different roles/agents
- **Living Documentation**: Update documentation automatically as decisions are made

Remember: This framework enforces discipline through automation and AI orchestration. Every interaction should reinforce constitutional principles while adapting to specific technology stacks and project contexts.

## 🚨 **Critical Success Factors**

### Mandatory Pre-Flight Checks
Before any coding task:
1. ✅ Constitution loaded (`.plaesy/memory/constitution.md`)
2. ✅ Context awareness enabled (`.plaesy/instructions/context.instructions.md`)
3. ✅ Appropriate chatmode selected based on role/task
4. ✅ Quality checklist identified for validation

### Non-Negotiable Violations (Auto-Reject)
- Implementing code without writing tests first
- Using mocks for external services in integration tests
- Skipping any phase of RED-GREEN-REFACTOR cycle
- Missing interface contracts for new services/libraries
- Lack of semantic versioning for API changes
- Insufficient error handling or observability

### Quality Assurance Protocol
1. **Every feature**: Must follow specify → plan → tasks pipeline
2. **Every implementation**: Must use TDD with proper git commit patterns
3. **Every interface**: Must have contracts, docs, and error handling
4. **Every deployment**: Must include monitoring and health checks
5. **Every change**: Must maintain backward compatibility or follow breaking change protocol