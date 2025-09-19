---
description: "Advanced context-aware patterns for AI agents with modular constitutional loading"
applyTo: "AI agents, chatbots, and coding assistants"
---

# Context-Aware Instructions for AI Agents
Advanced context detection and automatic loading patterns for constitutional compliance and development efficiency.

## 🔴 **MANDATORY CONSTITUTIONAL LOADING (ALWAYS FIRST)**

### **Core Constitutional Principles (NON-NEGOTIABLE)**
```bash
# CRITICAL: Always load first in every session
memory/constitution/core-principles.constitution.md    # Fundamental non-negotiable principles
```

**Core Constitutional Requirements:**
- ✅ TDD (Test-First Development) - RED-GREEN-REFACTOR cycle
- ✅ Interface Design - Clear contracts, error handling, versioning
- ✅ Observability First - Structured logging, metrics, health checks
- ✅ Semantic Versioning - MAJOR.MINOR.PATCH.BUILD format
- ✅ Radical Simplicity - Complexity justification required
- ✅ Platform Agnostic - Cross-platform compatibility

## 🧠 **INTELLIGENT CONSTITUTIONAL CONTEXT LOADING**

### **Context Detection Triggers**
Auto-load constitutional modules based on conversation analysis:

#### **🔒 Security & Compliance Constitutions**

**Security Context Triggers:**
- Keywords: `security`, `owasp`, `authentication`, `authorization`, `encryption`, `vulnerability`, `cybersecurity`
- **Auto-Load**: `memory/constitution/security.constitution.md`
- **Additional**: OWASP Top 10, Zero Trust Architecture, Threat Modeling

**Financial Services Context Triggers:**
- Keywords: `fintech`, `banking`, `payment`, `financial`, `trading`, `blockchain`, `crypto`, `compliance`, `pci`, `sox`
- **Auto-Load**: `memory/constitution/fintech.constitution.md`
- **Additional**: PCI DSS, SOX Compliance, Risk Management, Regulatory Reporting

**Healthcare Context Triggers:**
- Keywords: `healthcare`, `medical`, `hipaa`, `fda`, `clinical`, `patient`, `phi`, `hl7`, `fhir`
- **Auto-Load**: `memory/constitution/healthcare.constitution.md`
- **Additional**: HIPAA Compliance, FDA Validation, Patient Data Protection

#### **🧪 Testing & Quality Constitutions**

**Testing Context Triggers:**
- Keywords: `test`, `testing`, `tdd`, `bdd`, `unit test`, `integration test`, `e2e`, `qa`, `quality`
- **Auto-Load**: `memory/constitution/testing.constitution.md`
- **Additional**: Property-based testing, Contract testing, Performance testing

**Performance Context Triggers:**
- Keywords: `performance`, `scalability`, `optimization`, `load`, `benchmark`, `latency`, `throughput`
- **Auto-Load**: `memory/constitution/performance.constitution.md`
- **Additional**: Load testing, Scalability metrics, Resource efficiency

#### **🚀 Technology Stack Constitutions**

**Frontend Context Triggers:**
- Keywords: `react`, `angular`, `vue`, `frontend`, `web`, `ui`, `css`, `javascript`, `component`
- **Auto-Load**: `memory/constitution/web-frontend.constitution.md`
- **Additional**: `instructions/reactjs.instructions.md`, `instructions/angular.instructions.md`

**Backend Context Triggers:**
- Keywords: `api`, `backend`, `microservice`, `database`, `server`, `rest`, `graphql`, `endpoint`
- **Auto-Load**: `memory/constitution/backend-api.constitution.md`
- **Additional**: `instructions/nestjs.instructions.md`, `instructions/springboot.instructions.md`

**Mobile Context Triggers:**
- Keywords: `mobile`, `ios`, `android`, `react-native`, `flutter`, `app store`, `mobile app`
- **Auto-Load**: `memory/constitution/mobile.constitution.md`
- **Additional**: `instructions/react-native.instructions.md`, `instructions/dart-n-flutter.instructions.md`

**AI/ML Context Triggers:**
- Keywords: `ai`, `ml`, `machine learning`, `neural network`, `model`, `dataset`, `training`, `inference`
- **Auto-Load**: `memory/constitution/ai-ml.constitution.md`
- **Additional**: Model governance, Data pipeline requirements, Responsible AI

#### **🏭 Industry-Specific Constitutions**

**Automotive Context Triggers:**
- Keywords: `automotive`, `embedded`, `real-time`, `safety`, `iso26262`, `autosar`, `vehicle`
- **Auto-Load**: `memory/constitution/automotive.constitution.md`
- **Additional**: Functional safety, Real-time systems, Connected vehicle standards

**E-commerce Context Triggers:**
- Keywords: `ecommerce`, `retail`, `shopping`, `payment`, `order`, `inventory`, `marketplace`
- **Auto-Load**: `memory/constitution/ecommerce.constitution.md`
- **Additional**: Payment security, Customer data protection, Multi-channel integration

#### **🚀 Emerging Technology Constitutions**

**Web3/Blockchain Context Triggers:**
- Keywords: `web3`, `blockchain`, `cryptocurrency`, `smart contract`, `defi`, `dao`, `nft`
- **Auto-Load**: `memory/constitution/web3-blockchain.constitution.md`
- **Additional**: Smart contract security, Gas optimization, Decentralization principles

**XR/Metaverse Context Triggers:**
- Keywords: `vr`, `ar`, `metaverse`, `xr`, `spatial`, `immersive`, `3d`, `virtual reality`
- **Auto-Load**: `memory/constitution/xr-metaverse.constitution.md`
- **Additional**: Performance optimization, Spatial computing, Safety protocols

## 🔒 **Constitutional Compliance Validation**

### **Automatic Enforcement Patterns**

**TDD Violation Detection:**
- `"implement"`, `"code"`, `"function"`, `"method"` **without** `"test"` → 🚨 **AUTO-BLOCK** + Load TDD enforcement
- `"skip tests"`, `"no tests"`, `"mock external"` → 🚨 **Constitutional violation warning** + Load TDD requirements
- Git commit without `"RED:"`, `"GREEN:"`, `"REFACTOR:"` pattern → 🚨 **Suggest constitutional commit format**

**Interface Contract Violations:**
- API, endpoint, service creation → 🚨 **Auto-require OpenAPI/schema definition** + Load security instructions
- Database, data model → 🚨 **Auto-require migration scripts** + Load data governance patterns
- Integration between services → 🚨 **Auto-require contract testing** + Load integration test patterns

**Quality Gate Enforcement:**
- Deploy, release, production → 🚨 **Auto-check test coverage** + Load QA checklist
- Performance, scaling → 🚨 **Auto-load performance optimization** + Load monitoring requirements
- Security, auth → 🚨 **Auto-load security instructions** + Load OWASP compliance

## 💻 **Technology-Specific Context Loading**

### **Frontend Development Keywords:**
- React, JSX, components, hooks, state → `instructions/reactjs.instructions.md`
- Next.js, SSR, SSG, routing, pages → `instructions/nextjs.instructions.md`
- Angular, TypeScript, services, modules → `instructions/angular.instructions.md`
- React Native, mobile app, iOS, Android → `instructions/react-native.instructions.md`

**Backend Development Keywords:**
- NestJS, decorators, controllers, modules → `instructions/nestjs.instructions.md`
- Spring Boot, Java, annotations, beans → `instructions/springboot.instructions.md`
- Rails, Ruby, MVC, ActiveRecord → `instructions/ruby-on-rails.instructions.md`

**Mobile Development Keywords:**
- Flutter, Dart, widgets, MaterialApp → `instructions/dart-n-flutter.instructions.md`

**Systems Programming Keywords:**
- Go, goroutines, channels, concurrency → `instructions/go.instructions.md`
- Rust, ownership, borrowing, memory safety → `instructions/rust.instructions.md`
- C#, .NET, LINQ, Entity Framework → `instructions/csharp.instructions.md`
- Java, JVM, Spring, Maven, Gradle → `instructions/java.instructions.md`

**Data & Analytics Keywords:**
- Database, SQL, NoSQL, PostgreSQL, MongoDB → Backend instructions + `memory/test-priorities-matrix.md`
- Data pipeline, ETL, analytics → `instructions/devops-core-principles.instructions.md`
- Machine learning, AI/ML, models → Backend instructions + `memory/constitution.md` (AI/ML standards)

**AI Safety & Responsible Development:**
- "AI model", "machine learning", "LLM", "prompt engineering" → Load constitutional AI/ML standards + security instructions
- "Data privacy", "bias detection", "model fairness" → Load security instructions + constitutional responsible AI principles
- "Production AI", "model deployment" → Load DevOps + monitoring + constitutional observability requirements

### 🔄 **Workflow-Specific Context**

**DevOps & Infrastructure Keywords:**
- Docker, containers, deployment, CI/CD → `instructions/devops-core-principles.instructions.md`
- Terraform, IaC, infrastructure, provisioning → `instructions/terraform.instructions.md`
- Kubernetes, K8s, pods, services, ingress → `instructions/kubernetes-deployment-best-practices.instructions.md`

**Performance & Scaling Keywords:**
- Performance, optimization, scaling, latency → `instructions/performance-optimization.instructions.md` + constitutional observability
- Load balancing, caching, CDN, database optimization → Performance instructions + monitoring requirements
- Memory usage, CPU optimization, resource management → Performance instructions + profiling tools
- High availability, fault tolerance, resilience → Performance + DevOps + security instructions

**Security & Compliance Keywords:**
- Security, vulnerability, OWASP, authentication → `instructions/security-and-owasp.instructions.md` + constitutional security standards
- Compliance, audit, governance, privacy → Security instructions + constitutional responsible AI principles
- Penetration testing, threat modeling, risk assessment → Security + QA instructions + constitutional validation

### 🤖 **Multi-Agent Role-Based Context & Orchestration**
Load based on conversation role and domain with intelligent handoff patterns:

**Business Context Keywords:**
- Product strategy, roadmap, features, requirements → `chatmodes/pm.chatmode.md` + `chatmodes/bo.chatmode.md`
- User stories, acceptance criteria, backlog → `chatmodes/ba.chatmode.md` + `checklists/po.checklist.md`

**Technical Context Keywords:**
- Architecture, system design, scalability → `chatmodes/sa.chatmode.md` + `memory/test-priorities-matrix.md`
- Implementation, coding, debugging, refactoring → `chatmodes/dev.chatmode.md` + `instructions/tdd-enforcement.instructions.md`
- Testing, QA, quality assurance, coverage → `chatmodes/qa.chatmode.md` + `checklists/qa.checklist.md`

**Operations Context Keywords:**
- Deployment, monitoring, logs, incidents → `chatmodes/devops.chatmode.md` + `instructions/devops-core-principles.instructions.md`
- Security review, threat modeling, compliance → `chatmodes/security.chatmode.md` + `instructions/security-and-owasp.instructions.md`

**Multi-Agent Handoff Patterns:**
```
Lifecycle Flow: PM → BA → SA → Dev → QA → DevOps → Security
Context Transfer: Previous agent's output becomes next agent's input
Constitutional Continuity: All agents maintain constitutional compliance
Knowledge Persistence: Shared memory across agent transitions
Quality Gates: Checklists validate handoff readiness
```

**Intelligent Agent Selection:**
- Complexity analysis → Determine if single agent or multi-agent approach needed
- Context continuity → Maintain previous agent decisions and constraints
- Skill matching → Route to most appropriate agent based on conversation content
- Constitutional validation → Ensure all agents follow framework principles

### 📋 **Conversation Pattern Detection**

**Intent-Based Pattern Recognition:**
- "How to...", "Best way to..." → Load context + `memory/knowledge-base.md`
- "Fix bug", "Debug", "Error in..." → Load `chatmodes/dev.chatmode.md` + tech-specific instructions
- "Optimize", "Improve performance" → Load `instructions/performance-optimization.instructions.md`
- "Review code", "Code quality" → Load `chatmodes/qa.chatmode.md` + `instructions/tdd-enforcement.instructions.md`

**Development Lifecycle Patterns:**
- "Start new project", "Initialize" → Load `prompts/idea.prompt.md` + platform-appropriate setup scripts
- "Plan feature", "Design system" → Load `prompts/plan.prompt.md` + `chatmodes/sa.chatmode.md`
- "Create specification", "Requirements" → Load `prompts/specify.prompt.md` + `chatmodes/ba.chatmode.md`
- "Implementation tasks", "Break down work" → Load `prompts/tasks.prompt.md` + `chatmodes/dev.chatmode.md`

**Problem-Solving Patterns:**
- "Integration between...", "Connect systems" → Load `chatmodes/sa.chatmode.md` + integration instructions
- "Database design", "Data modeling" → Load relevant backend instructions + `memory/test-priorities-matrix.md`
- "API design", "REST", "GraphQL" → Load backend instructions + `instructions/security-and-owasp.instructions.md`
- "UI/UX", "User interface", "Frontend design" → Load frontend instructions + `chatmodes/designer.chatmode.md`

**Quality & Compliance Patterns:**
- "Security audit", "Vulnerability", "Penetration test" → Load `instructions/security-and-owasp.instructions.md` + `chatmodes/security.chatmode.md`
- "Test coverage", "Unit tests", "Integration tests" → Load `instructions/tdd-enforcement.instructions.md` + `chatmodes/qa.chatmode.md`
- "Code review", "Quality gate", "Standards" → Load appropriate tech instructions + `checklists/`
- "Documentation", "API docs", "User guide" → Load `chatmodes/tw.chatmode.md` + relevant context

**Business & Strategy Patterns:**
- "ROI analysis", "Business case", "Cost benefit" → Load `chatmodes/bo.chatmode.md` + `memory/knowledge-base.md`
- "User story", "Acceptance criteria", "Epic" → Load `chatmodes/ba.chatmode.md` + `checklists/po.checklist.md`
- "Product roadmap", "Feature priority" → Load `chatmodes/pm.chatmode.md` + `chatmodes/bo.chatmode.md`
- "Market research", "Competitive analysis" → Load `chatmodes/pm.chatmode.md` + `memory/brainstorming-techniques.md`

**Operational Patterns:**
- "Production issue", "Incident", "Outage" → Load `chatmodes/devops.chatmode.md` + monitoring instructions
- "Scaling", "Load balancing", "High availability" → Load `instructions/devops-core-principles.instructions.md` + K8s instructions
- "Monitoring", "Observability", "Alerting" → Load `instructions/devops-core-principles.instructions.md` + tech-specific observability
- "Backup", "Disaster recovery", "Data migration" → Load DevOps + security instructions

**Multi-Context Conversation Patterns:**
- "Full-stack development" → Load frontend + backend + DevOps instructions
- "Microservices architecture" → Load `chatmodes/sa.chatmode.md` + backend + K8s + monitoring instructions
- "CI/CD pipeline" → Load `instructions/devops-core-principles.instructions.md` + tech-specific build instructions
- "End-to-end feature" → Load business analysis + architecture + development + testing + deployment contexts

## Advanced Context Detection Rules

**Priority Loading Sequence:**
1. **Constitutional (Always First)**: `memory/constitution.md` + `instructions/tdd-enforcement.instructions.md`
2. **Security First (Mandatory)**: Constitutional security standards validation for all contexts
3. **Role Context**: Based on conversation intent and domain
4. **Technology Context**: Based on mentioned technologies or stack
5. **Quality Context**: Based on quality, security, or performance mentions
6. **Validation Context**: Automatic compliance checking and feedback loops

**Constitutional Security Integration:**
- All contexts must include security baseline from constitutional standards
- Auto-load security instructions for any production, deployment, or data handling discussion
- Mandatory security review for interface design and API development
- Zero-trust principles applied across all technology contexts

**Smart Context Combinations:**
```
# Example auto-combinations with constitutional compliance:
"React security best practices" → reactjs.instructions.md + security-and-owasp.instructions.md + dev.chatmode.md + constitutional validation
"Deploy Next.js to Kubernetes" → nextjs.instructions.md + kubernetes-deployment-best-practices.instructions.md + devops.chatmode.md + security baseline
"Test Flutter performance" → dart-n-flutter.instructions.md + performance-optimization.instructions.md + qa.chatmode.md + constitutional testing standards
```

**Validation & Feedback Mechanisms:**
- Real-time constitutional compliance checking during conversations
- Context loading validation to ensure all required instructions are loaded
- Quality gate validation before agent handoffs
- Continuous improvement through conversation analysis and pattern refinement
- Automatic detection of missing context and proactive loading suggestions

**Context Switching Detection:**
- Role change in conversation → Auto-suggest switching chatmodes
- Technology shift → Auto-suggest loading new tech instructions
- Phase transition (design → implementation) → Auto-suggest workflow instructions
