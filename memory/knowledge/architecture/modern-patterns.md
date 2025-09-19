---
version: "1.0.0"
updatedAt: "2025-09-16T07:38:00Z"
---

# Modern Architecture Patterns and Solutions

Comprehensive architecture patterns and solutions for modern application development, covering authentication, data architecture, microservices, and API design.

## Authentication and Authorization (2025 Edition)

### Modern Authentication Patterns

#### Passwordless Authentication
- **WebAuthn**: Hardware-based authentication using FIDO2 standards
  - Biometric authentication (fingerprint, face, voice)
  - Hardware security keys (YubiKey, Google Titan)
  - Platform authenticators (TouchID, FaceID, Windows Hello)
- **Magic Links**: Email-based passwordless authentication
- **Push Notifications**: Mobile app-based authentication approval
- **QR Code Authentication**: Scan-to-authenticate workflows

#### Zero Trust Identity
```typescript
interface ZeroTrustContext {
  user: {
    id: string;
    roles: string[];
    permissions: string[];
    lastAuthenticated: Date;
  };
  device: {
    id: string;
    trusted: boolean;
    lastSeen: Date;
    compliance: DeviceCompliance;
  };
  session: {
    riskScore: number;      // 0-100, higher = more risky
    location: GeoLocation;
    networkType: 'corporate' | 'public' | 'vpn';
    anomalyDetected: boolean;
  };
  context: {
    timeOfAccess: Date;
    resourceSensitivity: 'public' | 'internal' | 'confidential' | 'restricted';
    requiredMFA: boolean;
    additionalVerification: string[];
  };
}
```

#### Multi-Factor Authentication (MFA)
- **TOTP (Time-based OTP)**: Google Authenticator, Authy, Microsoft Authenticator
- **Hardware Tokens**: FIDO2/WebAuthn compliant devices
- **SMS/Voice**: Backup options (not primary due to security concerns)
- **Backup Codes**: Single-use recovery codes
- **Adaptive MFA**: Risk-based authentication requirements

#### Social Authentication & Federation
- **OAuth 2.1/OIDC**: Latest standards with PKCE and improved security
- **SAML 2.0**: Enterprise federation and single sign-on
- **Social Providers**: Google, Microsoft, Apple, GitHub, LinkedIn
- **Enterprise Identity**: Active Directory, Azure AD, Okta integration
- **Cross-domain SSO**: Single sign-on across multiple applications

### Authorization Frameworks

#### Fine-grained Access Control
```typescript
// Attribute-Based Access Control (ABAC)
interface ABACPolicy {
  subject: {
    userId: string;
    roles: string[];
    department: string;
    clearanceLevel: number;
  };
  resource: {
    id: string;
    type: 'document' | 'api' | 'database' | 'application';
    classification: 'public' | 'internal' | 'confidential' | 'restricted';
    owner: string;
    tags: string[];
  };
  action: 'read' | 'write' | 'delete' | 'execute' | 'admin';
  environment: {
    time: Date;
    location: GeoLocation;
    network: string;
    device: DeviceInfo;
  };
}

// Relationship-Based Access Control (ReBAC)
interface ResourceRelationship {
  user: string;
  relation: 'owner' | 'editor' | 'viewer' | 'admin' | 'member';
  resource: string;
  resourceType: 'workspace' | 'project' | 'document' | 'dataset';
  inherited?: boolean;
  conditions?: AccessCondition[];
}
```

#### Policy-as-Code Implementation
```rego
# Open Policy Agent (OPA) authorization policy
package authz

import future.keywords.if
import future.keywords.in

# Default deny
default allow = false

# Allow if user is owner of resource
allow if {
    input.user.id == data.resources[input.resource.id].owner
}

# Allow if user has explicit permission
allow if {
    permission := data.permissions[_]
    permission.user == input.user.id
    permission.resource == input.resource.id
    permission.action == input.action
    not permission.revoked
}

# Allow if user has role-based access
allow if {
    user_role := input.user.roles[_]
    role_permission := data.role_permissions[user_role][_]
    role_permission.resource_type == input.resource.type
    role_permission.action == input.action
}

# Deny if high risk context
allow = false if {
    input.context.risk_score > 80
    input.resource.classification in ["confidential", "restricted"]
}
```

#### Just-in-Time (JIT) Access
```typescript
interface JITAccessRequest {
  user: string;
  resource: string;
  permissions: string[];
  duration: number;        // minutes
  justification: string;
  approver?: string;
  autoApprove: boolean;
  conditions: {
    maxDuration: number;
    requireMFA: boolean;
    allowedNetworks: string[];
    businessHours: boolean;
  };
}

class JITAccessManager {
  async requestAccess(request: JITAccessRequest): Promise<AccessGrant> {
    // Validate request
    await this.validateRequest(request);
    
    // Auto-approve or require approval
    if (request.autoApprove && this.canAutoApprove(request)) {
      return this.grantAccess(request);
    } else {
      return this.requestApproval(request);
    }
  }
  
  async grantAccess(request: JITAccessRequest): Promise<AccessGrant> {
    const grant = {
      id: generateId(),
      user: request.user,
      resource: request.resource,
      permissions: request.permissions,
      expiresAt: new Date(Date.now() + request.duration * 60000),
      conditions: request.conditions
    };
    
    // Schedule automatic revocation
    this.scheduleRevocation(grant);
    
    return grant;
  }
}
```

## Data Architecture and Persistence (2025 Edition)

### Modern Data Patterns

#### Event Sourcing
```typescript
interface DomainEvent {
  id: string;
  aggregateId: string;
  aggregateVersion: number;
  eventType: string;
  eventData: Record<string, any>;
  metadata: {
    timestamp: Date;
    userId: string;
    correlationId: string;
    causationId?: string;
  };
}

class EventStore {
  async appendEvents(
    aggregateId: string, 
    expectedVersion: number, 
    events: DomainEvent[]
  ): Promise<void> {
    // Optimistic concurrency control
    await this.checkExpectedVersion(aggregateId, expectedVersion);
    
    // Append events atomically
    await this.store.transaction(async (tx) => {
      for (const event of events) {
        await tx.events.insert(event);
      }
      await tx.aggregates.update(aggregateId, { 
        version: expectedVersion + events.length 
      });
    });
    
    // Publish events to event bus
    await this.publishEvents(events);
  }
  
  async getEvents(aggregateId: string, fromVersion?: number): Promise<DomainEvent[]> {
    return this.store.events
      .where('aggregateId', aggregateId)
      .where('aggregateVersion', '>=', fromVersion || 0)
      .orderBy('aggregateVersion');
  }
}
```

#### CQRS (Command Query Responsibility Segregation)
```typescript
// Command side - Write model
interface CreateOrderCommand {
  customerId: string;
  items: OrderItem[];
  shippingAddress: Address;
  paymentMethod: PaymentMethod;
}

class OrderCommandHandler {
  async handle(command: CreateOrderCommand): Promise<void> {
    // Load aggregate from event store
    const order = await this.orderRepository.create(command.customerId);
    
    // Execute business logic
    order.addItems(command.items);
    order.setShippingAddress(command.shippingAddress);
    order.setPaymentMethod(command.paymentMethod);
    
    // Save events
    await this.orderRepository.save(order);
  }
}

// Query side - Read model
interface OrderView {
  id: string;
  customerName: string;
  totalAmount: number;
  status: OrderStatus;
  items: OrderItemView[];
  createdAt: Date;
  updatedAt: Date;
}

class OrderProjection {
  async on(event: OrderCreatedEvent): Promise<void> {
    const customer = await this.customerRepository.findById(event.customerId);
    
    await this.orderViewRepository.create({
      id: event.orderId,
      customerName: customer.name,
      totalAmount: event.totalAmount,
      status: 'pending',
      items: event.items,
      createdAt: event.timestamp,
      updatedAt: event.timestamp
    });
  }
}
```

#### Event-Driven Architecture
```typescript
// Event definition
interface OrderCreatedEvent {
  type: 'OrderCreated';
  orderId: string;
  customerId: string;
  totalAmount: number;
  items: OrderItem[];
  timestamp: Date;
}

// Event handler
class InventoryService {
  @EventHandler('OrderCreated')
  async handleOrderCreated(event: OrderCreatedEvent): Promise<void> {
    // Reserve inventory for order items
    for (const item of event.items) {
      await this.reserveInventory(item.productId, item.quantity);
    }
    
    // Publish inventory reserved event
    await this.eventBus.publish({
      type: 'InventoryReserved',
      orderId: event.orderId,
      reservations: event.items.map(item => ({
        productId: item.productId,
        quantity: item.quantity
      })),
      timestamp: new Date()
    });
  }
}
```

### Database Selection Matrix

#### Scale-based Database Recommendations

```typescript
interface DatabaseRecommendation {
  useCase: string;
  smallScale: DatabaseSolution;
  mediumScale: DatabaseSolution;
  largeScale: DatabaseSolution;
}

const databaseMatrix: DatabaseRecommendation[] = [
  {
    useCase: 'OLTP (Transactional)',
    smallScale: { 
      technology: 'PostgreSQL', 
      setup: 'Single instance with local backup',
      capacity: '< 100GB, < 1000 QPS'
    },
    mediumScale: { 
      technology: 'PostgreSQL', 
      setup: 'Primary + Read Replicas + Connection Pooling',
      capacity: '< 1TB, < 10000 QPS'
    },
    largeScale: { 
      technology: 'CockroachDB', 
      setup: 'Distributed cluster with automatic sharding',
      capacity: '> 1TB, > 10000 QPS'
    }
  },
  {
    useCase: 'Analytics (OLAP)',
    smallScale: { 
      technology: 'PostgreSQL', 
      setup: 'With analytics extensions (pg_analytics)',
      capacity: '< 100GB data'
    },
    mediumScale: { 
      technology: 'ClickHouse', 
      setup: 'Dedicated analytics cluster',
      capacity: '< 10TB data'
    },
    largeScale: { 
      technology: 'Snowflake', 
      setup: 'Cloud data warehouse with auto-scaling',
      capacity: '> 10TB data'
    }
  },
  {
    useCase: 'Caching',
    smallScale: { 
      technology: 'Redis', 
      setup: 'Single instance with persistence',
      capacity: '< 10GB cache'
    },
    mediumScale: { 
      technology: 'Redis Cluster', 
      setup: 'Multi-node cluster with replication',
      capacity: '< 100GB cache'
    },
    largeScale: { 
      technology: 'Redis Enterprise', 
      setup: 'Global distributed cache with consistency',
      capacity: '> 100GB cache'
    }
  }
];
```

#### Polyglot Persistence Strategy
```typescript
class DataLayer {
  // Transactional data - PostgreSQL
  private readonly transactionalDB: PostgreSQLClient;
  
  // Analytics data - ClickHouse
  private readonly analyticsDB: ClickHouseClient;
  
  // Cache layer - Redis
  private readonly cache: RedisClient;
  
  // Full-text search - Elasticsearch
  private readonly searchEngine: ElasticsearchClient;
  
  // Time series data - InfluxDB
  private readonly timeSeriesDB: InfluxDBClient;
  
  // Graph data - Neo4j
  private readonly graphDB: Neo4jClient;
  
  async saveOrder(order: Order): Promise<void> {
    // Save to transactional database
    await this.transactionalDB.orders.save(order);
    
    // Update search index
    await this.searchEngine.index('orders', order.id, {
      customerName: order.customer.name,
      items: order.items.map(item => item.name),
      totalAmount: order.totalAmount
    });
    
    // Cache frequently accessed data
    await this.cache.set(`order:${order.id}`, order, { ttl: 3600 });
    
    // Store analytics events
    await this.analyticsDB.insert('order_events', {
      event_type: 'order_created',
      order_id: order.id,
      customer_id: order.customerId,
      amount: order.totalAmount,
      timestamp: new Date()
    });
  }
}
```

### Data Governance Framework

#### Data Lineage Tracking
```typescript
interface DataLineage {
  datasetId: string;
  upstreamSources: DataSource[];
  transformations: DataTransformation[];
  downstreamConsumers: DataConsumer[];
  dataQuality: DataQualityMetrics;
  accessPatterns: AccessPattern[];
}

class DataLineageTracker {
  async trackTransformation(
    sourceDataset: string, 
    targetDataset: string, 
    transformation: DataTransformation
  ): Promise<void> {
    await this.lineageStore.recordTransformation({
      id: generateId(),
      sourceDataset,
      targetDataset,
      transformation,
      timestamp: new Date(),
      userId: this.getCurrentUser(),
      jobId: this.getCurrentJobId()
    });
  }
  
  async getLineage(datasetId: string, direction: 'upstream' | 'downstream' | 'both'): Promise<DataLineage> {
    return this.lineageStore.getLineage(datasetId, direction);
  }
}
```

## Microservices Architecture (2025 Edition)

### Service Design Patterns

#### Domain-Driven Design (DDD)
```typescript
// Bounded Context
class OrderManagement {
  // Aggregate Root
  class Order {
    private constructor(
      private readonly id: OrderId,
      private readonly customerId: CustomerId,
      private items: OrderItem[],
      private status: OrderStatus
    ) {}
    
    // Business logic encapsulated in aggregate
    addItem(productId: ProductId, quantity: number, price: Money): void {
      if (this.status !== OrderStatus.Draft) {
        throw new Error('Cannot modify confirmed order');
      }
      
      const existingItem = this.items.find(item => item.productId.equals(productId));
      if (existingItem) {
        existingItem.updateQuantity(quantity);
      } else {
        this.items.push(new OrderItem(productId, quantity, price));
      }
    }
    
    confirm(): DomainEvent[] {
      if (this.items.length === 0) {
        throw new Error('Cannot confirm empty order');
      }
      
      this.status = OrderStatus.Confirmed;
      
      return [
        new OrderConfirmedEvent(this.id, this.customerId, this.getTotalAmount())
      ];
    }
  }
}
```

#### Hexagonal Architecture
```typescript
// Domain layer - pure business logic
interface OrderRepository {
  save(order: Order): Promise<void>;
  findById(id: OrderId): Promise<Order | null>;
}

interface PaymentService {
  processPayment(amount: Money, paymentMethod: PaymentMethod): Promise<PaymentResult>;
}

// Application layer - orchestration
class OrderService {
  constructor(
    private orderRepository: OrderRepository,
    private paymentService: PaymentService,
    private eventBus: EventBus
  ) {}
  
  async confirmOrder(orderId: OrderId): Promise<void> {
    const order = await this.orderRepository.findById(orderId);
    if (!order) {
      throw new Error('Order not found');
    }
    
    // Process payment
    const paymentResult = await this.paymentService.processPayment(
      order.getTotalAmount(),
      order.getPaymentMethod()
    );
    
    if (paymentResult.isSuccessful()) {
      const events = order.confirm();
      await this.orderRepository.save(order);
      await this.eventBus.publishAll(events);
    }
  }
}

// Infrastructure layer - adapters
class PostgreSQLOrderRepository implements OrderRepository {
  async save(order: Order): Promise<void> {
    // Map domain object to database schema
    const orderData = OrderMapper.toPersistence(order);
    await this.db.orders.upsert(orderData);
  }
}
```

### Communication Patterns

#### Saga Pattern for Distributed Transactions
```typescript
// Orchestration-based Saga
class OrderProcessingSaga {
  private steps: SagaStep[] = [
    { name: 'ReserveInventory', service: 'inventory', compensate: 'ReleaseInventory' },
    { name: 'ProcessPayment', service: 'payment', compensate: 'RefundPayment' },
    { name: 'CreateShipment', service: 'shipping', compensate: 'CancelShipment' },
    { name: 'ConfirmOrder', service: 'order', compensate: 'CancelOrder' }
  ];
  
  async execute(sagaData: OrderSagaData): Promise<SagaResult> {
    const executedSteps: ExecutedStep[] = [];
    
    try {
      for (const step of this.steps) {
        const result = await this.executeStep(step, sagaData);
        executedSteps.push({ step, result });
        
        if (!result.isSuccess()) {
          await this.compensate(executedSteps.reverse());
          return SagaResult.failed(result.error);
        }
      }
      
      return SagaResult.success();
    } catch (error) {
      await this.compensate(executedSteps.reverse());
      return SagaResult.failed(error);
    }
  }
  
  private async compensate(executedSteps: ExecutedStep[]): Promise<void> {
    for (const { step } of executedSteps) {
      if (step.compensate) {
        await this.executeCompensation(step.compensate, sagaData);
      }
    }
  }
}
```

#### Circuit Breaker Pattern
```typescript
class CircuitBreaker {
  private state: 'CLOSED' | 'OPEN' | 'HALF_OPEN' = 'CLOSED';
  private failures = 0;
  private lastFailureTime?: Date;
  
  constructor(
    private failureThreshold: number = 5,
    private recoveryTimeout: number = 60000,
    private successThreshold: number = 3
  ) {}
  
  async execute<T>(operation: () => Promise<T>): Promise<T> {
    if (this.state === 'OPEN') {
      if (this.shouldAttemptReset()) {
        this.state = 'HALF_OPEN';
      } else {
        throw new Error('Circuit breaker is OPEN');
      }
    }
    
    try {
      const result = await operation();
      this.onSuccess();
      return result;
    } catch (error) {
      this.onFailure();
      throw error;
    }
  }
  
  private onSuccess(): void {
    this.failures = 0;
    this.state = 'CLOSED';
  }
  
  private onFailure(): void {
    this.failures++;
    this.lastFailureTime = new Date();
    
    if (this.failures >= this.failureThreshold) {
      this.state = 'OPEN';
    }
  }
  
  private shouldAttemptReset(): boolean {
    if (!this.lastFailureTime) return false;
    
    return Date.now() - this.lastFailureTime.getTime() > this.recoveryTimeout;
  }
}
```

### Service Mesh Implementation

#### Istio Configuration
```yaml
# Virtual Service for traffic routing
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: order-service
spec:
  hosts:
  - order-service
  http:
  - match:
    - headers:
        x-user-type:
          exact: premium
    route:
    - destination:
        host: order-service
        subset: v2
      weight: 100
  - route:
    - destination:
        host: order-service
        subset: v1
      weight: 90
    - destination:
        host: order-service
        subset: v2
      weight: 10
    fault:
      delay:
        percentage:
          value: 0.1
        fixedDelay: 5s
---
# Destination Rule for load balancing
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: order-service
spec:
  host: order-service
  trafficPolicy:
    circuitBreaker:
      consecutiveGatewayErrors: 3
      interval: 30s
      baseEjectionTime: 30s
      maxEjectionPercent: 50
    loadBalancer:
      simple: LEAST_CONN
  subsets:
  - name: v1
    labels:
      version: v1
  - name: v2
    labels:
      version: v2
```

## Error Handling and Resilience

### Comprehensive Error Classification
```typescript
enum ErrorCategory {
  // Client errors (4xx)
  VALIDATION = 'VALIDATION',         // 400 - Invalid input
  AUTHENTICATION = 'AUTHENTICATION', // 401 - Invalid credentials
  AUTHORIZATION = 'AUTHORIZATION',   // 403 - Insufficient permissions
  NOT_FOUND = 'NOT_FOUND',          // 404 - Resource not found
  CONFLICT = 'CONFLICT',            // 409 - Business rule violation
  RATE_LIMITED = 'RATE_LIMITED',    // 429 - Too many requests
  
  // Server errors (5xx)
  INTERNAL = 'INTERNAL',            // 500 - Internal system error
  BAD_GATEWAY = 'BAD_GATEWAY',      // 502 - Upstream service error
  SERVICE_UNAVAILABLE = 'SERVICE_UNAVAILABLE', // 503 - Service down
  GATEWAY_TIMEOUT = 'GATEWAY_TIMEOUT', // 504 - Upstream timeout
}

interface ErrorResponse {
  category: ErrorCategory;
  code: string;
  message: string;
  details?: Record<string, any>;
  traceId: string;
  timestamp: string;
  retryable: boolean;
  retryAfter?: number;
}

class ErrorHandler {
  handle(error: Error, context: RequestContext): ErrorResponse {
    const traceId = context.traceId || generateTraceId();
    
    // Log error with context
    this.logger.error('Request failed', {
      error: error.message,
      stack: error.stack,
      traceId,
      userId: context.userId,
      requestId: context.requestId,
      path: context.path,
      method: context.method
    });
    
    // Convert to standard error response
    if (error instanceof ValidationError) {
      return {
        category: ErrorCategory.VALIDATION,
        code: 'VALIDATION_FAILED',
        message: 'Request validation failed',
        details: error.validationErrors,
        traceId,
        timestamp: new Date().toISOString(),
        retryable: false
      };
    }
    
    if (error instanceof AuthenticationError) {
      return {
        category: ErrorCategory.AUTHENTICATION,
        code: 'AUTHENTICATION_FAILED',
        message: 'Authentication required',
        traceId,
        timestamp: new Date().toISOString(),
        retryable: false
      };
    }
    
    // Default to internal error
    return {
      category: ErrorCategory.INTERNAL,
      code: 'INTERNAL_ERROR',
      message: 'An internal error occurred',
      traceId,
      timestamp: new Date().toISOString(),
      retryable: true,
      retryAfter: 30
    };
  }
}
```

### Resilience Patterns Implementation
```typescript
class ResilientHttpClient {
  private circuitBreaker: CircuitBreaker;
  private retryPolicy: RetryPolicy;
  
  constructor(
    private baseUrl: string,
    options: ResilientClientOptions = {}
  ) {
    this.circuitBreaker = new CircuitBreaker(options.circuitBreaker);
    this.retryPolicy = new RetryPolicy(options.retry);
  }
  
  async request<T>(path: string, options: RequestOptions = {}): Promise<T> {
    return this.circuitBreaker.execute(async () => {
      return this.retryPolicy.execute(async () => {
        const response = await fetch(`${this.baseUrl}${path}`, {
          ...options,
          timeout: options.timeout || 5000,
          signal: AbortSignal.timeout(options.timeout || 5000)
        });
        
        if (!response.ok) {
          throw new HttpError(response.status, response.statusText);
        }
        
        return response.json();
      });
    });
  }
}

class RetryPolicy {
  constructor(
    private maxAttempts: number = 3,
    private baseDelay: number = 1000,
    private maxDelay: number = 10000,
    private backoffMultiplier: number = 2
  ) {}
  
  async execute<T>(operation: () => Promise<T>): Promise<T> {
    let lastError: Error;
    
    for (let attempt = 1; attempt <= this.maxAttempts; attempt++) {
      try {
        return await operation();
      } catch (error) {
        lastError = error;
        
        if (attempt === this.maxAttempts || !this.isRetryable(error)) {
          throw error;
        }
        
        const delay = Math.min(
          this.baseDelay * Math.pow(this.backoffMultiplier, attempt - 1),
          this.maxDelay
        );
        
        await this.sleep(delay);
      }
    }
    
    throw lastError!;
  }
  
  private isRetryable(error: Error): boolean {
    if (error instanceof HttpError) {
      // Retry on 5xx errors and specific 4xx errors
      return error.status >= 500 || error.status === 429;
    }
    
    // Retry on network errors
    return error.name === 'NetworkError' || error.name === 'TimeoutError';
  }
}
```

---

> **Related Knowledge**:
> - [Security Best Practices](../security/best-practices.md)
> - [Performance Optimization](../operations/performance.md)
> - [Platform Engineering & DevOps](../operations/platform-engineering.md)