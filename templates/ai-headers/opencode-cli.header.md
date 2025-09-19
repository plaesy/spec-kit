# Constitutional Development Framework - OpenCode CLI Configuration

## Framework Configuration
```json
{
  "framework": "plaesy-spec-kit",
  "version": "3.0.0",
  "enforcement": "constitutional",
  "compliance": "mandatory",
  "quality_gates": "automated"
}
```

## Constitutional Development Standards

### Mandatory Principles
The constitutional framework enforces these non-negotiable development standards:

#### 1. Test-Driven Development (TDD)
```json
{
  "methodology": "red-green-refactor",
  "coverage_minimum": "90%",
  "integration_testing": "real_dependencies_only",
  "mock_usage": "forbidden_in_integration",
  "test_categories": ["unit", "integration", "contract", "e2e"]
}
```

#### 2. Interface-First Design
```json
{
  "api_contracts": "openapi_3_0_required",
  "error_handling": "comprehensive",
  "input_validation": "strict",
  "versioning": "semantic",
  "backward_compatibility": "maintained"
}
```

#### 3. Built-in Observability
```json
{
  "logging": {
    "format": "structured_json",
    "correlation_ids": "required",
    "context": "comprehensive"
  },
  "monitoring": {
    "health_checks": "mandatory",
    "metrics": "business_and_technical",
    "alerting": "proactive"
  }
}
```

#### 4. Security by Design
```json
{
  "compliance": "owasp_top_10",
  "authentication": "robust",
  "authorization": "rbac_or_abac",
  "encryption": {
    "at_rest": "required",
    "in_transit": "required"
  },
  "security_testing": "automated"
}
```

#### 5. Platform Agnostic
```json
{
  "containerization": "docker_required",
  "infrastructure": "as_code",
  "environment_parity": "dev_prod_consistent",
  "scalability": "horizontal_and_vertical"
}
```

### Development Workflow

#### Implementation Process
1. **Requirements Validation**: Ensure alignment with constitutional principles
2. **Test Strategy Design**: Create comprehensive testing approach with real dependencies
3. **Interface Definition**: Design API contracts and error handling
4. **TDD Implementation**: Follow Red-Green-Refactor cycle strictly
5. **Integration**: Test with actual external dependencies
6. **Security Review**: Validate OWASP compliance
7. **Documentation**: Complete API and operational documentation

#### Code Quality Standards
```json
{
  "readability": "high",
  "maintainability": "prioritized",
  "error_handling": "comprehensive",
  "performance": "optimized",
  "security": "embedded",
  "logging": "structured_contextual"
}
```

### Response Requirements

#### Solution Structure
Every response must include:

1. **Constitutional Validation**
   - Verify request aligns with framework principles
   - Identify applicable constitutional requirements

2. **Test-First Implementation**
   - Comprehensive test cases before implementation
   - Multiple test levels (unit, integration, contract)
   - Real dependencies in integration tests

3. **Implementation with Standards**
   - Clean, maintainable code following constitutional principles
   - Comprehensive error handling and validation
   - Structured logging and monitoring integration

4. **Security Integration**
   - Authentication and authorization implementation
   - Data protection measures
   - Security threat consideration and mitigation

5. **Documentation and Examples**
   - Complete API documentation (OpenAPI format)
   - Usage examples with error scenarios
   - Operational guides and troubleshooting information

#### Quality Gates Checklist
```json
{
  "tdd_followed": true,
  "real_dependencies_used": true,
  "error_handling_comprehensive": true,
  "structured_logging_added": true,
  "security_addressed": true,
  "performance_considered": true,
  "documentation_complete": true,
  "constitutional_compliance": true
}
```

### Context Integration

#### Project Memory Integration
- Apply patterns from `.plaesy/memory/` directory
- Maintain consistency with existing architecture
- Follow established technology decisions
- Adhere to team coding standards

#### Technology Stack Alignment
- Respect project technology choices
- Maintain architectural coherence
- Support team development workflows
- Enable collaborative development practices

### Output Format

#### Code Examples
```typescript
// Constitutional framework compliant implementation
describe('Store Service', () => {
  test('creates store with valid data', async () => {
    // Real dependency testing (constitutional requirement)
    const store = await storeService.create(validStoreData);
    expect(store.id).toBeDefined();
  });
});

class StoreService {
  async create(data: CreateStoreRequest): Promise<Store> {
    // Structured logging (constitutional requirement)
    logger.info('Store creation started', {
      correlationId: uuid(),
      action: 'store_create',
      data: sanitizeForLogging(data)
    });

    // Input validation (constitutional requirement)
    const validation = validateStoreData(data);
    if (!validation.isValid) {
      throw new ValidationError(validation.errors);
    }

    // Implementation with error handling
    try {
      const store = await this.repository.save(data);

      logger.info('Store creation completed', {
        correlationId: uuid(),
        action: 'store_create_success',
        storeId: store.id
      });

      return store;
    } catch (error) {
      logger.error('Store creation failed', {
        correlationId: uuid(),
        action: 'store_create_error',
        error: error.message
      });
      throw error;
    }
  }
}
```

### Constitutional Compliance Validation

Before finalizing any solution, ensure:
- Constitutional principles are followed
- Quality gates are met
- Security requirements are addressed
- Documentation is complete
- Testing strategy includes real dependencies
- Error handling is comprehensive
- Observability is integrated

---