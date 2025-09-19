# Constitutional Development Framework - ChatGPT Configuration

## System Context
You are working within the Plaesy Spec-Kit constitutional framework. This framework enforces systematic, high-quality software development through non-negotiable principles and quality gates.

## Constitutional Principles (MANDATORY)

### 1. Test-Driven Development (TDD)
- **Red-Green-Refactor cycle**: Write failing tests, implement minimal code, refactor
- **90%+ test coverage**: Comprehensive testing with meaningful scenarios
- **Real dependencies**: Integration tests use actual external systems (no mocks)
- **Test categories**: Unit, integration, contract, and end-to-end testing

### 2. Interface-First Design
- **API contracts**: Define OpenAPI specifications before implementation
- **Error handling**: Comprehensive error responses with meaningful messages
- **Input validation**: Strict validation with clear error feedback
- **Versioning**: Semantic versioning with backward compatibility planning

### 3. Observability by Design
- **Structured logging**: JSON logs with correlation IDs and contextual information
- **Health monitoring**: Health check endpoints for service status
- **Metrics collection**: Performance and business metrics with dashboards
- **Alerting**: Proactive monitoring with intelligent alerting rules

### 4. Security-First Development
- **OWASP compliance**: Apply OWASP Top 10 principles from design phase
- **Authentication/Authorization**: Robust identity and access management
- **Data protection**: Encryption at rest and in transit with proper key management
- **Security testing**: Automated security scanning in CI/CD pipeline

### 5. Platform-Agnostic Architecture
- **Containerization**: Docker-based deployment with orchestration
- **Infrastructure as Code**: Reproducible infrastructure provisioning
- **Environment consistency**: Development-production parity
- **Scalability**: Horizontal and vertical scaling considerations

## Response Guidelines

### Code Implementation
- **Complete solutions**: Provide full, working implementations
- **Error handling**: Include comprehensive error scenarios and handling
- **Logging integration**: Add structured logging throughout code
- **Performance considerations**: Include optimization and scalability notes
- **Security integration**: Embed security practices in all code

### Testing Strategy
- **Test-first approach**: Show test cases before implementation
- **Real dependencies**: Use actual databases, APIs, and external services
- **Coverage validation**: Ensure tests cover edge cases and error scenarios
- **Integration testing**: Demonstrate service-to-service communication testing

### Documentation Standards
- **API documentation**: Complete OpenAPI specifications with examples
- **Code comments**: Explain complex logic and business rules
- **Architecture decisions**: Document technical choices and trade-offs
- **Operational guides**: Include deployment and troubleshooting information

## Quality Validation

Before providing solutions, validate:
- [ ] Tests written before implementation (TDD compliance)
- [ ] Real dependencies used in integration tests
- [ ] Comprehensive error handling implemented
- [ ] Structured logging with correlation IDs added
- [ ] Security considerations addressed
- [ ] Performance implications considered
- [ ] Documentation provided for APIs and complex logic

## Context Integration
- **Memory patterns**: Apply relevant patterns from `.plaesy/memory/`
- **Existing architecture**: Maintain consistency with current system design
- **Technology stack**: Follow established technology choices
- **Team standards**: Adhere to project coding and documentation standards

## Example Response Structure
1. **Constitutional validation**: Confirm requirements align with framework
2. **Test implementation**: Provide comprehensive test cases first
3. **Core implementation**: Implement solution following constitutional principles
4. **Error handling**: Add robust error scenarios and responses
5. **Observability**: Include logging, metrics, and health checks
6. **Security considerations**: Address authentication, authorization, and data protection
7. **Documentation**: Provide API docs and usage examples

---