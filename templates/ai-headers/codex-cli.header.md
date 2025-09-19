# Constitutional Development Framework - Codex CLI Configuration

## Project Memory Integration
This project follows the Plaesy Spec-Kit constitutional framework. All development must adhere to systematic quality gates and non-negotiable principles stored in the project memory system.

## Constitutional Framework Context

### Core Development Principles (MANDATORY)
The constitutional framework establishes these immutable development standards:

**Test-Driven Development (TDD)**
- Red-Green-Refactor cycle is mandatory for all feature development
- Minimum 90% test coverage with meaningful test scenarios
- Integration tests must use real external dependencies (no mocks allowed)
- Contract testing required between all service boundaries

**Interface-First Design**
- OpenAPI specifications must be created before implementation
- Comprehensive error handling with meaningful error messages
- Strict input validation with clear error feedback
- Semantic versioning with backward compatibility planning

**Built-in Observability**
- Structured JSON logging with correlation IDs for request tracing
- Health check endpoints for all services
- Performance and business metrics collection
- Proactive monitoring and intelligent alerting

**Security by Design**
- OWASP Top 10 compliance from specification phase
- Authentication and authorization for all protected resources
- Data encryption at rest and in transit
- Automated security scanning in CI/CD pipeline

**Platform Agnostic Architecture**
- Docker containerization for consistent deployment
- Infrastructure as Code for reproducible environments
- Development-production environment parity
- Horizontal and vertical scalability considerations

### Implementation Guidelines

**Development Workflow**
1. Validate requirements against constitutional principles
2. Design comprehensive test strategy with real dependencies
3. Create API contracts and error handling specifications
4. Implement using strict TDD methodology
5. Integrate with actual external dependencies
6. Validate OWASP compliance and security measures
7. Document APIs and create operational guides

**Code Quality Requirements**
- Clean, readable, and maintainable code structure
- Comprehensive error scenarios and response handling
- Performance optimization and resource efficiency
- Secure coding practices and threat mitigation
- Structured logging with contextual information
- Multi-level testing with real dependency validation

**Response Structure**
When providing solutions, follow this structure:

1. **Constitutional Validation**: Verify alignment with framework principles
2. **Test Implementation**: Provide comprehensive test cases first
3. **Core Implementation**: Implement following constitutional standards
4. **Error Handling**: Add robust error scenarios and responses
5. **Observability**: Include logging, metrics, and health checks
6. **Security**: Address authentication, authorization, and data protection
7. **Documentation**: Provide API docs and usage examples

**Quality Gates**
All solutions must pass these validation points:
- [ ] TDD process followed (tests before implementation)
- [ ] Real dependencies used in integration tests
- [ ] Comprehensive error handling implemented
- [ ] Structured logging with correlation IDs
- [ ] Security considerations addressed (OWASP)
- [ ] Performance implications considered
- [ ] Complete documentation provided

### Project Memory Reference
- Constitutional rules: `.plaesy/memory/constitution.md`
- Domain patterns: `.plaesy/memory/*.md`
- Technology decisions: Project architecture documentation
- Team standards: Coding and documentation conventions

### Context Awareness
Apply these integration patterns:
- Use established patterns from project memory
- Maintain consistency with existing system architecture
- Follow project technology stack decisions
- Support team collaboration and development workflows

Remember: The constitutional framework is designed to ensure systematic quality, security, and maintainability. All development decisions should prioritize these constitutional principles over convenience or speed.

---