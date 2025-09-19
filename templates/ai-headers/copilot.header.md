## GitHub Copilot Instructions for Constitutional Development Framework
### Auto-loaded from .github/copilot-instructions.md

## Project Overview
This project uses the Plaesy Spec-Kit constitutional development framework - a systematic approach ensuring quality through discipline and excellence through automation.

## Tech Stack & Architecture
- **Framework**: Constitutional development with non-negotiable quality gates
- **Testing**: TDD Red-Green-Refactor cycle (mandatory)
- **Dependencies**: Real external systems (no mocks in integration tests)
- **Contracts**: API-first design with OpenAPI specifications
- **Observability**: Structured logging, metrics, health checks
- **Security**: OWASP compliance embedded from specification phase

## Coding Standards (Non-Negotiable)
- **TDD Enforcement**: Write tests before implementation code
- **Coverage Requirement**: Minimum 90% test coverage
- **Real Dependencies**: Integration tests use actual databases, APIs, services
- **Interface Contracts**: All services expose OpenAPI specifications
- **Error Handling**: Comprehensive error responses with meaningful messages
- **Structured Logging**: JSON logs with correlation IDs for tracing
- **Security First**: Authentication, authorization, data protection built-in

## File Structure
```
.plaesy/                    # Constitutional framework
├── memory/constitution.md  # Framework principles
├── templates/             # Specification templates
├── chatmodes/            # Role-based development modes
└── instructions/         # Technology-specific guidance
```

## Key Commands & Tools
- **Initialize**: `plaesy init --ai copilot`
- **Develop**: Follow /idea → /specify → /plan → /tasks workflow
- **Test**: Always write tests first (Red-Green-Refactor)
- **Validate**: Use `.plaesy/checklists/` for quality gates

## Constitutional Principles for Copilot
1. **Test-Driven Development**: Red-Green-Refactor cycle is mandatory
2. **Real Dependencies**: No mocks in integration tests
3. **Interface Contracts**: API-first with comprehensive error handling
4. **Observability**: Built-in logging, metrics, monitoring
5. **Security by Design**: OWASP principles from day one
6. **Platform Agnostic**: Solutions work across environments

## Response Guidelines
- Generate test-first code following TDD principles
- Include error handling and input validation
- Add structured logging with correlation tracking
- Provide OpenAPI specifications for endpoints
- Consider security implications in all suggestions
- Maintain consistency with constitutional patterns

---