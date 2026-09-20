# Service Development Specification: [SERVICE NAME]

**Created**: [DATE]  
**Status**: Service Specification  
**Type**: API/Microservice Development  
**Raw Input**: "$ARGUMENTS"

## Auto-Generated Service Framework

```
1. Service Architecture Design
   → Define API contracts and service boundaries
   → Design data models and persistence layer
   → Plan authentication, authorization, and security
2. API Development Strategy
   → Design RESTful/GraphQL endpoints
   → Plan request/response schemas and validation
   → Implement rate limiting and caching strategies
3. Integration & Communication
   → Design service-to-service communication
   → Plan event streaming and message queues
   → Implement service discovery and load balancing
4. Observability & Monitoring
   → Implement logging, metrics, and tracing
   → Design health checks and alerting
   → Plan performance monitoring and SLA tracking
5. Deployment & Operations
   → Design containerization and orchestration
   → Plan CI/CD pipeline and deployment strategies
   → Implement backup, recovery, and scaling procedures
6. Return: SUCCESS (service ready for development)
```

---

## ⚡ Service Development Guidelines

- ✅ Design API-first with clear contracts
- ✅ Implement comprehensive error handling
- ✅ Plan for high availability and scalability
- ✅ Security and authentication from day one
- ✅ Observability and monitoring built-in
- ❌ Don't create monolithic services
- ❌ Don't ignore data consistency patterns
- ❌ Don't skip performance testing

---

## 🎯 Service Overview _(mandatory)_

### Core Purpose
[Describe what this service does and its business value]

**Example**: "A user authentication and authorization service that provides secure login, role management, and session handling for microservice architectures."

### Service Boundaries
- **Domain**: [What business domain does this service own?]
- **Responsibilities**: [What specific functions does it provide?]
- **Dependencies**: [What other services does it depend on?]
- **Consumers**: [What services or clients will use this?]

### Key Capabilities
1. [Primary capability this service provides]
2. [Secondary capability this service provides]
3. [Unique feature that differentiates this service]

---

## 🔌 API Design _(auto-generated)_

### API Contract Specification
```yaml
API Type: [REST/GraphQL/gRPC]
Base URL: https://api.domain.com/v1/service-name
Authentication: [Bearer Token/API Key/OAuth2]
Content Type: application/json
Rate Limiting: [Requests per minute/hour]
```

### Core Endpoints _(REST Example)_
```yaml
# Authentication & Authorization
POST   /auth/login              # User authentication
POST   /auth/refresh            # Token refresh
DELETE /auth/logout             # User logout
GET    /auth/me                 # Current user info

# Resource Management
GET    /resources               # List resources (paginated)
POST   /resources               # Create new resource
GET    /resources/{id}          # Get specific resource
PUT    /resources/{id}          # Update resource
DELETE /resources/{id}          # Delete resource

# Health & Monitoring
GET    /health                  # Health check endpoint
GET    /metrics                 # Prometheus metrics
GET    /status                  # Service status info
```

### Request/Response Schemas
```yaml
# Example User Login
POST /auth/login:
  request:
    email: string (required, email format)
    password: string (required, min 8 chars)
    remember_me: boolean (optional)
  response:
    access_token: string
    refresh_token: string
    expires_in: number
    user: object
  errors:
    400: Invalid credentials
    429: Too many requests
    500: Internal server error

# Example Resource Creation
POST /resources:
  request:
    name: string (required, max 100 chars)
    description: string (optional, max 500 chars)
    tags: array of strings (optional)
  response:
    id: string (UUID)
    name: string
    description: string
    created_at: string (ISO 8601)
    updated_at: string (ISO 8601)
```

### Error Handling Standards
```yaml
Error Response Format:
  error:
    code: string          # Machine-readable error code
    message: string       # Human-readable message
    details: object       # Additional error context
    trace_id: string      # Request tracing ID

HTTP Status Codes:
  200: Success
  201: Created
  400: Bad Request (client error)
  401: Unauthorized
  403: Forbidden
  404: Not Found
  429: Too Many Requests
  500: Internal Server Error
  503: Service Unavailable
```

---

## 🏗️ Technical Architecture _(auto-generated)_

### Technology Stack
```yaml
Language: [Go/Java/Python/Node.js/Rust]
Framework: [Gin/Spring Boot/FastAPI/Express/Actix]
Database: [PostgreSQL/MongoDB/Redis]
Cache: [Redis/Memcached]
Message Queue: [RabbitMQ/Apache Kafka/AWS SQS]
Monitoring: [Prometheus + Grafana/DataDog]
```

### Service Architecture
```
┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │────│   API Gateway   │
└─────────────────┘    └─────────────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
┌───────▼────────┐    ┌────────▼────────┐    ┌───────▼────────┐
│  Service A     │    │  Service B      │    │  Service C     │
│  (This Service)│    │  (Dependencies) │    │  (Dependencies)│
└────────────────┘    └─────────────────┘    └────────────────┘
        │                      │                      │
┌───────▼────────┐    ┌────────▼────────┐    ┌───────▼────────┐
│   Database     │    │   Message Queue │    │   Cache Layer  │
└────────────────┘    └─────────────────┘    └────────────────┘
```

### Data Architecture
```yaml
Primary Database:
  type: PostgreSQL
  schema: [Define core tables and relationships]
  indexes: [Performance optimization indexes]
  migrations: [Version-controlled schema changes]

Cache Layer:
  type: Redis
  patterns: [Cache-aside, write-through, write-behind]
  expiration: [TTL strategies for different data types]
  invalidation: [Cache invalidation strategies]

Message Queues:
  type: Apache Kafka
  topics: [Event types and message schemas]
  partitioning: [Data partitioning strategies]
  retention: [Message retention policies]
```

### Security Architecture
```yaml
Authentication:
  method: JWT Bearer Tokens
  issuer: [Authentication service/provider]
  validation: [Token validation strategy]
  refresh: [Token refresh mechanism]

Authorization:
  model: RBAC (Role-Based Access Control)
  roles: [Define service-specific roles]
  permissions: [Granular permission system]
  enforcement: [Authorization middleware/filters]

Network Security:
  tls: TLS 1.3 for all communications
  cors: [CORS policy configuration]
  rate_limiting: [Rate limiting strategies]
  firewall: [Network access controls]
```

---

## 📊 Data Management _(auto-generated)_

### Database Design
```sql
-- Example Core Tables
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_resources_user_id ON resources(user_id);
CREATE INDEX idx_resources_status ON resources(status);
```

### Data Migration Strategy
```yaml
Migration Framework: [Flyway/Liquibase/Alembic/Rails migrations]
Versioning: Sequential numbering with timestamps
Rollback: All migrations must be reversible
Testing: Migrations tested on staging data
Automation: Integrated into CI/CD pipeline
```

### Data Consistency
```yaml
Transactions: ACID compliance for critical operations
Distributed Transactions: Saga pattern for cross-service operations
Event Sourcing: [If applicable] Event store for audit trail
CQRS: [If applicable] Separate read/write models
```

---

## 🔐 Security Implementation _(auto-generated)_

### Authentication & Authorization
```yaml
JWT Configuration:
  algorithm: RS256
  expiration: 15 minutes (access), 7 days (refresh)
  issuer: auth-service
  audience: service-name
  
Role Definitions:
  admin: Full system access
  user: Standard user operations
  readonly: Read-only access
  
Permission Matrix:
  - resource:create → [admin, user]
  - resource:read → [admin, user, readonly]
  - resource:update → [admin, user (own resources)]
  - resource:delete → [admin, user (own resources)]
```

### Input Validation
```yaml
Validation Rules:
  - All inputs sanitized and validated
  - SQL injection prevention
  - XSS prevention in responses
  - JSON schema validation
  - File upload restrictions
  
Rate Limiting:
  - 100 requests/minute per user
  - 1000 requests/minute per IP
  - Stricter limits for sensitive endpoints
  
Security Headers:
  - Content-Security-Policy
  - X-Frame-Options: DENY
  - X-Content-Type-Options: nosniff
  - Strict-Transport-Security
```

### Audit & Monitoring
```yaml
Audit Logging:
  - All API requests logged
  - User actions tracked
  - Security events recorded
  - Data access logged
  
Security Monitoring:
  - Failed authentication attempts
  - Unusual access patterns
  - Rate limit violations
  - Privilege escalation attempts
```

---

## 📈 Performance & Scalability _(auto-generated)_

### Performance Requirements
```yaml
Response Time Targets:
  - API endpoints: < 200ms (95th percentile)
  - Database queries: < 50ms (95th percentile)
  - Health checks: < 10ms
  
Throughput Targets:
  - 1000 requests/second sustained
  - 5000 requests/second peak
  - 99.9% availability SLA
  
Resource Limits:
  - Memory: 512MB baseline, 2GB max
  - CPU: 0.5 cores baseline, 2 cores max
  - Storage: 10GB baseline, 100GB max
```

### Caching Strategy
```yaml
Cache Layers:
  - Application cache: In-memory for frequently accessed data
  - Distributed cache: Redis for session data
  - Database cache: Query result caching
  - CDN cache: Static asset caching
  
Cache Patterns:
  - Cache-aside for user data
  - Write-through for configuration
  - Write-behind for analytics
  - Refresh-ahead for predictable data
```

### Scaling Strategy
```yaml
Horizontal Scaling:
  - Stateless service design
  - Load balancer distribution
  - Auto-scaling based on CPU/memory
  - Database read replicas
  
Vertical Scaling:
  - Resource limit increases
  - Performance optimization
  - Database optimization
  - Connection pooling
```

---

## 📊 Observability _(auto-generated)_

### Logging Strategy
```yaml
Log Levels:
  - ERROR: Error conditions requiring attention
  - WARN: Warning conditions
  - INFO: General information
  - DEBUG: Detailed debugging information
  
Log Format (JSON):
  timestamp: ISO 8601 format
  level: Log level
  service: Service name
  trace_id: Request tracing ID
  message: Log message
  context: Additional context data
  
Log Destinations:
  - stdout: For container environments
  - File: For persistent storage
  - External: ELK stack, Splunk, etc.
```

### Metrics Collection
```yaml
Application Metrics:
  - Request count and duration
  - Error rate and types
  - Database query performance
  - Queue processing time
  
Business Metrics:
  - User registration rate
  - Feature usage statistics
  - Resource creation/deletion
  - API endpoint popularity
  
Infrastructure Metrics:
  - CPU and memory usage
  - Network I/O
  - Disk usage
  - Database connections
```

### Health Checks
```yaml
Health Check Endpoints:
  GET /health:
    purpose: Basic service availability
    response_time: < 10ms
    checks: Service startup, basic connectivity
  
  GET /health/ready:
    purpose: Readiness for traffic
    checks: Database connectivity, dependencies
  
  GET /health/live:
    purpose: Liveness check
    checks: Service responsiveness
```

---

## 🚀 Deployment Strategy _(auto-generated)_

### Containerization
```dockerfile
# Example Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o service ./cmd/server

FROM alpine:3.18
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/service .
EXPOSE 8080
CMD ["./service"]
```

### Kubernetes Deployment
```yaml
# Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: service-name
spec:
  replicas: 3
  selector:
    matchLabels:
      app: service-name
  template:
    metadata:
      labels:
        app: service-name
    spec:
      containers:
      - name: service-name
        image: service-name:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: service-secrets
              key: database-url
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### CI/CD Pipeline
```yaml
Pipeline Stages:
  1. Code Quality: Linting, formatting, security scan
  2. Testing: Unit, integration, and contract tests
  3. Build: Container image creation
  4. Security: Vulnerability scanning
  5. Deploy Staging: Automated staging deployment
  6. E2E Testing: End-to-end test suite
  7. Deploy Production: Blue-green deployment
  8. Monitoring: Health check verification
```

---

## 🧪 Testing Strategy _(auto-generated)_

### Test Pyramid
```yaml
Unit Tests (70%):
  focus: Business logic, utilities, data models
  tools: [Language-specific testing frameworks]
  coverage: 85% minimum
  
Integration Tests (20%):
  focus: Database, external services, API endpoints
  tools: [TestContainers, Docker Compose]
  coverage: Key integration points
  
E2E Tests (10%):
  focus: Complete user workflows, API contracts
  tools: [Postman, Newman, REST Assured]
  coverage: Critical business flows
```

### Test Scenarios
- [ ] **API Contract Tests**: Verify all endpoints and schemas
- [ ] **Authentication Tests**: Login, logout, token validation
- [ ] **Authorization Tests**: Role-based access control
- [ ] **Database Tests**: CRUD operations, transactions
- [ ] **Performance Tests**: Load testing, stress testing
- [ ] **Security Tests**: Input validation, injection attacks
- [ ] **Error Handling Tests**: Edge cases, failure scenarios

---

## ❓ Open Questions & Research Items

### Technical Architecture
- [NEEDS RESEARCH: Database choice optimization for use case]
- [NEEDS RESEARCH: Message queue vs direct service communication]
- [NEEDS RESEARCH: Caching strategy for specific data patterns]
- [NEEDS RESEARCH: Service mesh integration requirements]

### Security & Compliance
- [NEEDS RESEARCH: Compliance requirements (GDPR, HIPAA, etc.)]
- [NEEDS RESEARCH: Security audit and penetration testing needs]
- [NEEDS RESEARCH: Data encryption requirements]
- [NEEDS RESEARCH: Audit logging retention policies]

### Operations & Deployment
- [NEEDS RESEARCH: Infrastructure requirements and costs]
- [NEEDS RESEARCH: Disaster recovery and backup strategies]
- [NEEDS RESEARCH: Monitoring and alerting tool selection]
- [NEEDS RESEARCH: Multi-region deployment requirements]

---

## 🔄 Next Steps

1. **API Design Review**: Validate endpoints and data models
2. **Architecture Validation**: Review technical approach and stack
3. **Security Assessment**: Conduct threat modeling
4. **Performance Planning**: Define SLAs and monitoring strategy
5. **Infrastructure Setup**: Prepare development and staging environments

---

**Auto-Generated Checklist**: ✅ Service specification complete
**Next Template**: → Move to `implementation.template.md` for development
**Estimated Timeline**: [AUTO: Based on complexity analysis]
**Infrastructure Requirements**: [AUTO: Based on scale and features]