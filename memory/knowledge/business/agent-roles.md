---
version: "1.0.0"
updatedAt: "2025-09-16T07:38:00Z"
---

# Agent Roles and Responsibilities

The Spec-Kit framework defines clear role boundaries and responsibilities that scale from individual contributors to global enterprises. Each role has specific adaptation patterns as organizations grow.

## Role Structure Overview

```markdown
chatmodes/
├── ba.chatmode.md         # Business Analyst
├── bo.chatmode.md         # Business Owner
├── designer.chatmode.md   # UI/UX Designer
├── dev.chatmode.md        # Full Stack Developer
├── devops.chatmode.md     # DevOps Engineer
├── pe.chatmode.md         # Prompt Engineer
├── pm.chatmode.md         # Product Manager
├── po.chatmode.md         # Product Owner
├── qa.chatmode.md         # Quality Assurance
├── sa.chatmode.md         # Solution Architect
├── security.chatmode.md   # Security Specialist
├── sm.chatmode.md         # Scrum Master
├── tw.chatmode.md         # Technical Writer
└── ...                    # Other role-specific templates
```

## Business Layer

### Business Owner (BO)
**Responsibilities**: Strategic vision, business requirements, ROI analysis
**Scale Adaptation**: Individual contributor → Department head → Global business leader

**Key Activities**:
- Define business strategy and vision
- Analyze return on investment and business value
- Set strategic priorities and resource allocation
- Stakeholder relationship management
- Risk assessment and mitigation planning

### Product Owner (PO)
**Responsibilities**: Product strategy, user stories, prioritization, market analysis
**Scale Adaptation**: Single product → Product portfolio → Global product ecosystem

**Key Activities**:
- Product roadmap development and maintenance
- User story creation and acceptance criteria definition
- Backlog prioritization and management
- Market research and competitive analysis
- Customer feedback integration and product iteration

### Business Analyst (BA)
**Responsibilities**: Requirements gathering, process analysis, stakeholder management
**Scale Adaptation**: Local requirements → Cross-functional analysis → Global process optimization

**Key Activities**:
- Requirements elicitation and documentation
- Business process analysis and optimization
- Stakeholder communication and coordination
- Gap analysis and solution recommendation
- Change impact assessment and management

## Technical Layer

### Solution Architect (SA)
**Responsibilities**: Technical architecture, system design, technology strategy
**Scale Adaptation**: Application architecture → Enterprise architecture → Global technology governance

**Key Activities**:
- System architecture design and documentation
- Technology stack selection and evaluation
- Performance and scalability planning
- Integration patterns and API design
- Technical risk assessment and mitigation

### Full Stack Developer (DEV)
**Responsibilities**: Implementation, coding, testing, feature delivery
**Scale Adaptation**: Individual features → Team coordination → Platform contributions

**Key Activities**:
- Code implementation and maintenance
- Unit and integration testing
- Code review and quality assurance
- Technical documentation
- Feature delivery and bug resolution

### DevOps Engineer (DevOps)
**Responsibilities**: Infrastructure, deployment, automation, platform engineering
**Scale Adaptation**: CI/CD setup → Platform engineering → Global infrastructure automation

**Key Activities**:
- CI/CD pipeline design and implementation
- Infrastructure automation and management
- Monitoring and alerting configuration
- Security and compliance automation
- Deployment strategy and execution

### Security Engineer
**Responsibilities**: Security assessment, compliance, threat modeling, governance
**Scale Adaptation**: Application security → Enterprise security → Global security operations

**Key Activities**:
- Security architecture and threat modeling
- Vulnerability assessment and penetration testing
- Security policy development and enforcement
- Compliance monitoring and reporting
- Incident response and forensics

### Platform Engineer
**Responsibilities**: Developer experience, internal tooling, infrastructure automation
**Scale Adaptation**: Team tools → Department platform → Global developer platform

**Key Activities**:
- Developer tooling and platform development
- Self-service infrastructure provisioning
- Platform monitoring and optimization
- Developer experience improvement
- Platform adoption and training

### Site Reliability Engineer (SRE)
**Responsibilities**: System reliability, performance, incident response
**Scale Adaptation**: Service monitoring → Multi-service reliability → Global SRE practices

**Key Activities**:
- Service level objective (SLO) definition and monitoring
- Incident response and post-mortem analysis
- System performance optimization
- Reliability engineering and chaos testing
- On-call rotation and escalation procedures

## Quality Layer

### QA Engineer (QA)
**Responsibilities**: Testing strategy, quality assurance, automation
**Scale Adaptation**: Manual testing → Automated testing → Quality engineering platform

**Key Activities**:
- Test strategy and plan development
- Test case design and execution
- Test automation framework development
- Quality metrics and reporting
- Defect tracking and resolution

### Product Manager (PM)
**Responsibilities**: Product delivery, coordination, stakeholder management
**Scale Adaptation**: Single product → Product portfolio → Global product strategy

**Key Activities**:
- Product delivery coordination
- Cross-functional team collaboration
- Stakeholder communication and reporting
- Product metrics and analytics
- Go-to-market strategy and execution

### Designer
**Responsibilities**: User experience, visual design, design systems
**Scale Adaptation**: Interface design → Design system → Global brand consistency

**Key Activities**:
- User experience research and design
- Visual design and prototyping
- Design system development and maintenance
- Usability testing and optimization
- Design review and quality assurance

### Technical Writer (TW)
**Responsibilities**: Documentation, knowledge management, content strategy
**Scale Adaptation**: Team documentation → Knowledge systems → Global content platform

**Key Activities**:
- Technical documentation creation and maintenance
- Knowledge management system development
- Content strategy and information architecture
- Documentation review and quality assurance
- Training material development

## Support Layer

### Project/Scrum Master (SM)
**Responsibilities**: Process management, facilitation, agile coaching
**Scale Adaptation**: Team facilitation → Multi-team coordination → Agile transformation

**Key Activities**:
- Agile process facilitation and coaching
- Team coordination and communication
- Sprint planning and retrospectives
- Impediment removal and escalation
- Process improvement and optimization

### Prompt Engineer (PE)
**Responsibilities**: AI prompt optimization, AI tool integration, automation
**Scale Adaptation**: Individual productivity → Team AI tools → Enterprise AI platform

**Key Activities**:
- AI prompt design and optimization
- AI tool integration and automation
- AI workflow development and training
- AI governance and best practices
- AI performance monitoring and improvement

### Data Engineer
**Responsibilities**: Data pipeline, analytics, machine learning infrastructure
**Scale Adaptation**: Simple analytics → Data platform → Global data governance

**Key Activities**:
- Data pipeline design and implementation
- Data warehouse and lake architecture
- ETL/ELT process development
- Data quality and governance
- Analytics and reporting infrastructure

### AI/ML Engineer
**Responsibilities**: Machine learning models, AI system integration, MLOps
**Scale Adaptation**: Model development → ML platform → Global AI governance

**Key Activities**:
- Machine learning model development and training
- ML pipeline and infrastructure development
- Model deployment and monitoring
- AI system integration and optimization
- MLOps and model lifecycle management

## Role Collaboration Patterns

### Cross-Functional Teams
- **Product Trio**: PO + Designer + Tech Lead
- **Delivery Team**: SM + Dev + QA + DevOps
- **Architecture Team**: SA + Security + Platform Engineer
- **Business Team**: BO + BA + PM

### Scale-Specific Adaptations

#### Small Scale (1-10 people)
- Individual contributors wear multiple hats
- Direct communication and informal processes
- Minimal overhead and maximum agility

#### Medium Scale (10-50 people)
- Specialized roles with clear boundaries
- Formal processes and communication channels
- Team coordination and knowledge sharing

#### Large Scale (50+ people)
- Dedicated role specialists and sub-specializations
- Formal governance and process management
- Global coordination and standardization

---

> **Related Knowledge**:
> - [Core Philosophy](./core-philosophy.md)
> - [Quality Standards & Metrics](../operations/quality-standards.md)
> - [Platform Engineering & DevOps](../operations/platform-engineering.md)