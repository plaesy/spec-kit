# CLAUDE.md - Constitutional Development Framework Configuration
**Optimized for Claude Code 2024-2025 best practices**
**Auto-loaded project context for AI-assisted development**

## Project Overview
This project uses the Plaesy Spec-Kit constitutional development framework - a systematic approach to software development ensuring quality through discipline and excellence through automation.

## Constitutional Framework (Non-Negotiable Principles)

### 1. Test-Driven Development (TDD)
- **Red-Green-Refactor cycle**: Write failing tests first, implement minimal code, refactor
- **Coverage requirement**: Minimum 90% test coverage with meaningful test scenarios
- **Real dependencies**: Integration tests use actual databases, APIs, external services (no mocks)
- **Test hierarchy**: Unit → Integration → Contract → End-to-end testing

### 2. Interface-First Design
- **API contracts**: OpenAPI 3.0 specifications before implementation
- **Error handling**: Comprehensive error responses with meaningful HTTP status codes
- **Input validation**: Strict validation with clear, actionable error messages
- **Versioning strategy**: Semantic versioning with backward compatibility

### 3. Built-in Observability
- **Structured logging**: JSON format with correlation IDs for request tracing
- **Health monitoring**: Health check endpoints for all services
- **Metrics collection**: Performance and business metrics with dashboards
- **Alerting rules**: Proactive monitoring with intelligent alerting

### 4. Security by Design
- **OWASP compliance**: Apply OWASP Top 10 principles from specification phase
- **Authentication**: Robust identity verification mechanisms
- **Authorization**: Proper access controls and permission systems
- **Data protection**: Encryption at rest and in transit with key management

### 5. Platform Agnostic Architecture
- **Containerization**: Docker for consistent deployment environments
- **Infrastructure as Code**: Automated infrastructure provisioning
- **Environment parity**: Development-production consistency
- **Scalability**: Horizontal and vertical scaling considerations

## Development Workflow Commands
- `/idea` - Creative exploration and problem validation
- `/specify` - Technical requirements and specification
- `/plan` - Implementation strategy and architecture
- `/tasks` - Development task breakdown with dependencies

## File Structure Context
```
.plaesy/                    # Constitutional framework
├── memory/constitution.md  # Core framework principles
├── templates/             # Specification templates
├── chatmodes/            # Role-based development modes
├── instructions/         # Technology-specific guidance
└── checklists/           # Quality validation gates
```

## Build Commands & Scripts
- `plaesy init --ai claude` - Initialize project with Claude configuration
- Refer to `.plaesy/scripts/` for framework automation scripts
- Use `.plaesy/templates/` for structured specification development

## Code Generation Standards
- **Test-first approach**: Generate comprehensive test cases before implementation
- **Complete solutions**: Provide fully working, production-ready code
- **Error resilience**: Include comprehensive error handling and recovery
- **Performance aware**: Consider optimization and resource efficiency
- **Security minded**: Embed security practices throughout implementation
- **Documentation**: Include API documentation and usage examples

## Quality Gates & Validation
- All code must pass constitutional framework validation
- Database schemas require proper constraints, indexes, and migrations
- API endpoints need complete OpenAPI specifications
- Integration tests must use real external dependencies
- Performance baselines established with monitoring
- Security threats identified and mitigated

## Memory Integration
- Apply established patterns from `.plaesy/memory/` directory
- Maintain architectural consistency across project components
- Follow technology stack decisions and team conventions
- Reference constitutional principles for all development decisions

---

# Constitutional Framework Context

You are operating within the Plaesy Spec-Kit constitutional framework. This is a systematic approach to software development that enforces quality, security, and maintainability through non-negotiable principles.

## Core Constitutional Principles

### 1. Test-Driven Development (TDD)
- **Mandatory Red-Green-Refactor cycle** for all feature development
- **90%+ test coverage** with meaningful test scenarios
- **Real dependency testing** - no mocks in integration tests
- **Contract testing** between services and external APIs

### 2. Interface-First Design
- **API contracts before implementation** using OpenAPI specifications
- **Comprehensive error handling** with meaningful error responses
- **Input validation** with clear validation rules
- **Versioning strategy** with backward compatibility

### 3. Built-in Observability
- **Structured logging** with correlation IDs for request tracing
- **Health check endpoints** for service monitoring
- **Metrics collection** for performance and business KPIs
- **Alerting configuration** for proactive issue detection

### 4. Security by Design
- **OWASP compliance** from specification through deployment
- **Authentication and authorization** for all protected resources
- **Data protection** including encryption at rest and in transit
- **Security scanning** integrated into CI/CD pipeline

### 5. Platform Agnostic Architecture
- **Containerized deployment** with Docker and orchestration
- **Infrastructure as Code** for reproducible environments
- **Environment parity** between development and production
- **Scalability considerations** built into design

## Implementation Standards

### Code Quality
- Clean, readable, and maintainable code
- Comprehensive error handling and logging
- Performance optimization and resource efficiency
- Security best practices and threat mitigation

### Documentation
- API documentation with examples and error scenarios
- Architecture decision records (ADRs)
- Deployment and operational runbooks
- User guides and developer documentation

### Testing Strategy
- Unit tests for individual components
- Integration tests with real dependencies
- Contract tests for service boundaries
- End-to-end tests for critical user journeys

---