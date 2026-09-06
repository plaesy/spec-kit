# Integration Examples and Implementation Patterns

## Overview
This document provides comprehensive real-world examples of how to implement the Spec-Kit constitutional framework across different project types, technologies, and team scales.

## Project Type Examples

### 1. Web Application (Full-Stack)
**Technology Stack**: React + Node.js + PostgreSQL

#### Constitutional Framework Implementation
```markdown
# Project: E-commerce Platform

## /idea Phase Output
- Problem: Online retailers need customizable storefront with inventory management
- Solution: Multi-tenant e-commerce platform with admin dashboard
- Users: Store owners, customers, platform administrators
- Success Metrics: 1000 stores onboarded in 6 months, 99.9% uptime

## /specify Phase Output
### API Contracts (OpenAPI 3.0)
```yaml
openapi: 3.0.0
info:
  title: E-commerce Platform API
  version: 1.0.0
paths:
  /api/v1/stores:
    post:
      summary: Create new store
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Store'
      responses:
        201:
          description: Store created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/StoreResponse'
        400:
          description: Validation error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

#### Database Schema (Constitutional)
```sql
-- stores table with constitutional requirements
CREATE TABLE stores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) UNIQUE NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    status store_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1,

    -- Constitutional requirements
    CONSTRAINT store_name_length CHECK (length(name) >= 2),
    CONSTRAINT store_domain_format CHECK (domain ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$')
);

-- Indexes for performance
CREATE INDEX idx_stores_owner_id ON stores(owner_id);
CREATE INDEX idx_stores_domain ON stores(domain);
CREATE INDEX idx_stores_status ON stores(status);
```

#### Test-First Implementation (TDD)
```javascript
// 1. Contract Test (written first)
describe('Store API Contract', () => {
  test('POST /api/v1/stores creates store with valid data', async () => {
    const storeData = {
      name: 'Test Store',
      domain: 'test-store',
      owner_id: '123e4567-e89b-12d3-a456-426614174000'
    };

    const response = await request(app)
      .post('/api/v1/stores')
      .send(storeData)
      .expect(201);

    expect(response.body).toMatchSchema(storeResponseSchema);
    expect(response.body.data.name).toBe(storeData.name);
  });
});

// 2. Unit Tests (written before implementation)
describe('Store Service', () => {
  test('createStore validates required fields', async () => {
    const storeService = new StoreService(mockDb);

    await expect(storeService.createStore({}))
      .rejects.toThrow('Name is required');
  });

  test('createStore generates UUID and timestamps', async () => {
    const storeService = new StoreService(mockDb);
    const storeData = { name: 'Test', domain: 'test', owner_id: 'uuid' };

    const store = await storeService.createStore(storeData);

    expect(store.id).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/);
    expect(store.created_at).toBeDefined();
  });
});

// 3. Implementation (written after tests)
class StoreService {
  async createStore(storeData) {
    // Validation
    if (!storeData.name) {
      throw new ValidationError('Name is required');
    }

    // Business logic
    const store = {
      id: uuid.v4(),
      ...storeData,
      status: 'active',
      created_at: new Date(),
      updated_at: new Date(),
      version: 1
    };

    // Database interaction
    return await this.db.stores.create(store);
  }
}

// 4. Integration Test (with real database)
describe('Store Integration', () => {
  beforeEach(async () => {
    await testDb.migrate.latest();
    await testDb.seed.run();
  });

  test('complete store creation workflow', async () => {
    // Test with real database, real API calls
    const user = await createTestUser();
    const storeData = {
      name: 'Integration Test Store',
      domain: 'integration-test',
      owner_id: user.id
    };

    const response = await request(app)
      .post('/api/v1/stores')
      .send(storeData)
      .expect(201);

    // Verify in database
    const dbStore = await testDb('stores')
      .where({ id: response.body.data.id })
      .first();

    expect(dbStore).toBeDefined();
    expect(dbStore.name).toBe(storeData.name);
  });
});
```

#### Monitoring & Observability
```javascript
// Constitutional logging
const logger = require('pino')({
  level: process.env.LOG_LEVEL || 'info',
  base: {
    service: 'ecommerce-api',
    version: process.env.APP_VERSION
  }
});

// Structured logging in service
class StoreService {
  async createStore(storeData) {
    const correlationId = uuid.v4();

    logger.info({
      correlationId,
      action: 'store_creation_started',
      owner_id: storeData.owner_id,
      domain: storeData.domain
    });

    try {
      const store = await this.db.stores.create(storeData);

      logger.info({
        correlationId,
        action: 'store_creation_completed',
        store_id: store.id,
        duration_ms: Date.now() - startTime
      });

      return store;
    } catch (error) {
      logger.error({
        correlationId,
        action: 'store_creation_failed',
        error: error.message,
        stack: error.stack
      });
      throw error;
    }
  }
}

// Health checks
app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    version: process.env.APP_VERSION,
    timestamp: new Date().toISOString(),
    dependencies: {
      database: db.isHealthy(),
      redis: redis.isHealthy()
    }
  });
});
```

### 2. CLI Tool
**Technology Stack**: Node.js CLI with Commander.js

#### Constitutional Implementation
```javascript
#!/usr/bin/env node

// Constitutional error handling and contracts
class CliError extends Error {
  constructor(message, code = 1) {
    super(message);
    this.name = 'CliError';
    this.code = code;
  }
}

// Interface contracts for CLI commands
class CommandContract {
  constructor() {
    this.requiredArgs = [];
    this.optionalArgs = [];
    this.flags = [];
  }

  validate(args) {
    const missing = this.requiredArgs.filter(arg => !args[arg]);
    if (missing.length > 0) {
      throw new CliError(`Missing required arguments: ${missing.join(', ')}`);
    }
  }
}

// Test-first CLI development
describe('CLI Store Command', () => {
  test('create command requires name and domain', async () => {
    const result = await runCli(['store', 'create']);

    expect(result.code).toBe(1);
    expect(result.stderr).toContain('Missing required arguments: name, domain');
  });

  test('create command generates valid output', async () => {
    const result = await runCli([
      'store', 'create',
      '--name', 'Test Store',
      '--domain', 'test-store'
    ]);

    expect(result.code).toBe(0);
    expect(result.stdout).toContain('Store created successfully');

    const output = JSON.parse(result.stdout);
    expect(output.data.id).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}/);
  });
});

// Implementation with constitutional compliance
class StoreCommand {
  constructor() {
    this.contract = new CommandContract();
    this.contract.requiredArgs = ['name', 'domain'];
    this.contract.flags = ['--json', '--verbose'];
  }

  async execute(args) {
    try {
      // Contract validation
      this.contract.validate(args);

      // Structured logging
      logger.info({
        command: 'store_create',
        args: { name: args.name, domain: args.domain }
      });

      // Business logic with real dependencies
      const store = await this.storeService.createStore({
        name: args.name,
        domain: args.domain,
        owner_id: args.userId
      });

      // Consistent output format
      const output = {
        success: true,
        data: store,
        timestamp: new Date().toISOString()
      };

      if (args.json) {
        console.log(JSON.stringify(output, null, 2));
      } else {
        console.log(`Store created successfully: ${store.id}`);
      }

      return 0;
    } catch (error) {
      logger.error({
        command: 'store_create',
        error: error.message
      });

      if (args.json) {
        console.error(JSON.stringify({
          success: false,
          error: error.message,
          timestamp: new Date().toISOString()
        }));
      } else {
        console.error(`Error: ${error.message}`);
      }

      return error.code || 1;
    }
  }
}
```

### 3. Microservice
**Technology Stack**: Go + gRPC + PostgreSQL

#### Constitutional gRPC Implementation
```protobuf
// store.proto - Interface contracts first
syntax = "proto3";

package store.v1;

service StoreService {
  // Constitutional error handling in contracts
  rpc CreateStore(CreateStoreRequest) returns (CreateStoreResponse);
  rpc GetStore(GetStoreRequest) returns (GetStoreResponse);
  rpc UpdateStore(UpdateStoreRequest) returns (UpdateStoreResponse);
  rpc DeleteStore(DeleteStoreRequest) returns (DeleteStoreResponse);
  rpc ListStores(ListStoresRequest) returns (ListStoresResponse);
}

message Store {
  string id = 1;
  string name = 2;
  string domain = 3;
  string owner_id = 4;
  StoreStatus status = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
  int32 version = 8;
}

message CreateStoreRequest {
  string name = 1;        // Required: min 2 chars
  string domain = 2;      // Required: unique, alphanumeric-dash
  string owner_id = 3;    // Required: valid UUID
}

message CreateStoreResponse {
  Store store = 1;
  Error error = 2;        // Constitutional error handling
}

message Error {
  string code = 1;        // ERROR_VALIDATION, ERROR_DUPLICATE, etc.
  string message = 2;
  repeated string details = 3;
}

enum StoreStatus {
  STORE_STATUS_UNSPECIFIED = 0;
  STORE_STATUS_ACTIVE = 1;
  STORE_STATUS_INACTIVE = 2;
  STORE_STATUS_DELETED = 3;
}
```

```go
// Test-first implementation
func TestStoreService_CreateStore(t *testing.T) {
  tests := []struct {
    name     string
    request  *pb.CreateStoreRequest
    wantCode codes.Code
    wantErr  bool
  }{
    {
      name: "valid store creation",
      request: &pb.CreateStoreRequest{
        Name:    "Test Store",
        Domain:  "test-store",
        OwnerId: "123e4567-e89b-12d3-a456-426614174000",
      },
      wantCode: codes.OK,
    },
    {
      name: "missing name",
      request: &pb.CreateStoreRequest{
        Domain:  "test-store",
        OwnerId: "123e4567-e89b-12d3-a456-426614174000",
      },
      wantCode: codes.InvalidArgument,
      wantErr:  true,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      // Test with real database (constitutional requirement)
      db := setupTestDB(t)
      service := NewStoreService(db, logger)

      resp, err := service.CreateStore(context.Background(), tt.request)

      if tt.wantErr {
        assert.Error(t, err)
        assert.Equal(t, tt.wantCode, status.Code(err))
      } else {
        assert.NoError(t, err)
        assert.NotNil(t, resp.Store)
        assert.NotEmpty(t, resp.Store.Id)
      }
    })
  }
}

// Implementation with constitutional compliance
type StoreService struct {
  db     *sql.DB
  logger *zap.Logger
  pb.UnimplementedStoreServiceServer
}

func (s *StoreService) CreateStore(ctx context.Context, req *pb.CreateStoreRequest) (*pb.CreateStoreResponse, error) {
  // Request validation (constitutional requirement)
  if err := s.validateCreateStoreRequest(req); err != nil {
    s.logger.Warn("invalid create store request",
      zap.String("name", req.Name),
      zap.String("domain", req.Domain),
      zap.Error(err),
    )
    return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
  }

  // Structured logging (constitutional requirement)
  correlationID := uuid.New().String()
  s.logger.Info("creating store",
    zap.String("correlation_id", correlationID),
    zap.String("name", req.Name),
    zap.String("domain", req.Domain),
    zap.String("owner_id", req.OwnerId),
  )

  // Business logic with transaction
  tx, err := s.db.BeginTx(ctx, nil)
  if err != nil {
    return nil, status.Errorf(codes.Internal, "failed to begin transaction: %v", err)
  }
  defer tx.Rollback()

  store := &pb.Store{
    Id:        uuid.New().String(),
    Name:      req.Name,
    Domain:    req.Domain,
    OwnerId:   req.OwnerId,
    Status:    pb.StoreStatus_STORE_STATUS_ACTIVE,
    CreatedAt: timestamppb.Now(),
    UpdatedAt: timestamppb.Now(),
    Version:   1,
  }

  // Database interaction with constitutional error handling
  query := `
    INSERT INTO stores (id, name, domain, owner_id, status, created_at, updated_at, version)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
  `

  _, err = tx.ExecContext(ctx, query,
    store.Id, store.Name, store.Domain, store.OwnerId,
    store.Status, store.CreatedAt.AsTime(), store.UpdatedAt.AsTime(), store.Version,
  )

  if err != nil {
    if isDuplicateError(err) {
      return nil, status.Errorf(codes.AlreadyExists, "store domain already exists: %s", req.Domain)
    }

    s.logger.Error("failed to create store",
      zap.String("correlation_id", correlationID),
      zap.Error(err),
    )
    return nil, status.Errorf(codes.Internal, "failed to create store")
  }

  if err := tx.Commit(); err != nil {
    return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
  }

  s.logger.Info("store created successfully",
    zap.String("correlation_id", correlationID),
    zap.String("store_id", store.Id),
  )

  return &pb.CreateStoreResponse{Store: store}, nil
}

// Constitutional validation
func (s *StoreService) validateCreateStoreRequest(req *pb.CreateStoreRequest) error {
  if req.Name == "" {
    return errors.New("name is required")
  }

  if len(req.Name) < 2 {
    return errors.New("name must be at least 2 characters")
  }

  if req.Domain == "" {
    return errors.New("domain is required")
  }

  if !isValidDomain(req.Domain) {
    return errors.New("domain must contain only lowercase letters, numbers, and hyphens")
  }

  if req.OwnerId == "" {
    return errors.New("owner_id is required")
  }

  if !isValidUUID(req.OwnerId) {
    return errors.New("owner_id must be a valid UUID")
  }

  return nil
}
```

### 4. Mobile App Backend (Serverless)
**Technology Stack**: AWS Lambda + DynamoDB + API Gateway

#### Constitutional Serverless Implementation
```typescript
// lambda/stores/create.ts - Constitutional serverless function

import { APIGatewayProxyEvent, APIGatewayProxyResult } from 'aws-lambda';
import { DynamoDBClient } from '@aws-sdk/client-dynamodb';
import { DynamoDBDocumentClient, PutCommand } from '@aws-sdk/lib-dynamodb';
import { v4 as uuidv4 } from 'uuid';
import { createLogger } from '../lib/logger';
import { validateCreateStoreRequest, CreateStoreRequest } from '../lib/validation';
import { StoreResponse, ErrorResponse } from '../lib/types';

// Constitutional error handling
class StoreError extends Error {
  constructor(
    message: string,
    public statusCode: number = 500,
    public code: string = 'INTERNAL_ERROR'
  ) {
    super(message);
    this.name = 'StoreError';
  }
}

// Initialize with constitutional observability
const dynamodb = DynamoDBDocumentClient.from(new DynamoDBClient({
  region: process.env.AWS_REGION,
}));

const logger = createLogger({
  service: 'store-api',
  version: process.env.VERSION || '1.0.0',
});

// Constitutional interface contract
export const handler = async (
  event: APIGatewayProxyEvent
): Promise<APIGatewayProxyResult> => {
  const correlationId = uuidv4();

  logger.info({
    correlationId,
    action: 'create_store_started',
    requestId: event.requestContext.requestId,
    userAgent: event.headers['User-Agent'],
  });

  try {
    // Input validation (constitutional requirement)
    if (!event.body) {
      throw new StoreError('Request body is required', 400, 'MISSING_BODY');
    }

    const requestData: CreateStoreRequest = JSON.parse(event.body);

    // Constitutional validation
    const validationResult = validateCreateStoreRequest(requestData);
    if (!validationResult.isValid) {
      throw new StoreError(
        `Validation failed: ${validationResult.errors.join(', ')}`,
        400,
        'VALIDATION_ERROR'
      );
    }

    // Extract user ID from JWT (constitutional auth requirement)
    const userId = extractUserIdFromJWT(event.headers.Authorization);
    if (!userId) {
      throw new StoreError('Authentication required', 401, 'UNAUTHORIZED');
    }

    // Business logic with constitutional compliance
    const store = {
      id: uuidv4(),
      name: requestData.name,
      domain: requestData.domain,
      ownerId: userId,
      status: 'active',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      version: 1,
      // DynamoDB specific
      pk: `STORE#${requestData.domain}`,
      sk: `STORE#${uuidv4()}`,
      gsi1pk: `USER#${userId}`,
      gsi1sk: `STORE#${new Date().toISOString()}`,
    };

    // Database interaction with constitutional error handling
    const putCommand = new PutCommand({
      TableName: process.env.STORES_TABLE,
      Item: store,
      ConditionExpression: 'attribute_not_exists(pk)', // Prevent duplicates
    });

    await dynamodb.send(putCommand);

    logger.info({
      correlationId,
      action: 'create_store_completed',
      storeId: store.id,
      domain: store.domain,
      ownerId: userId,
    });

    // Constitutional response format
    const response: StoreResponse = {
      success: true,
      data: {
        id: store.id,
        name: store.name,
        domain: store.domain,
        status: store.status,
        createdAt: store.createdAt,
        updatedAt: store.updatedAt,
      },
      meta: {
        correlationId,
        timestamp: new Date().toISOString(),
        version: '1.0.0',
      },
    };

    return {
      statusCode: 201,
      headers: {
        'Content-Type': 'application/json',
        'X-Correlation-ID': correlationId,
        'X-API-Version': '1.0.0',
      },
      body: JSON.stringify(response),
    };

  } catch (error) {
    logger.error({
      correlationId,
      action: 'create_store_failed',
      error: error.message,
      stack: error.stack,
    });

    // Constitutional error response
    if (error instanceof StoreError) {
      const errorResponse: ErrorResponse = {
        success: false,
        error: {
          code: error.code,
          message: error.message,
          correlationId,
          timestamp: new Date().toISOString(),
        },
      };

      return {
        statusCode: error.statusCode,
        headers: {
          'Content-Type': 'application/json',
          'X-Correlation-ID': correlationId,
        },
        body: JSON.stringify(errorResponse),
      };
    }

    // Unknown error handling
    const errorResponse: ErrorResponse = {
      success: false,
      error: {
        code: 'INTERNAL_ERROR',
        message: 'An unexpected error occurred',
        correlationId,
        timestamp: new Date().toISOString(),
      },
    };

    return {
      statusCode: 500,
      headers: {
        'Content-Type': 'application/json',
        'X-Correlation-ID': correlationId,
      },
      body: JSON.stringify(errorResponse),
    };
  }
};

// Constitutional testing with real dependencies
describe('Create Store Lambda', () => {
  const mockDynamoClient = mockClient(DynamoDBDocumentClient);

  beforeEach(() => {
    mockDynamoClient.reset();
  });

  test('creates store with valid data', async () => {
    // Mock successful DynamoDB response
    mockDynamoClient.on(PutCommand).resolves({});

    const event = createAPIGatewayEvent({
      body: JSON.stringify({
        name: 'Test Store',
        domain: 'test-store',
      }),
      headers: {
        Authorization: 'Bearer valid-jwt-token',
      },
    });

    const result = await handler(event);

    expect(result.statusCode).toBe(201);

    const response = JSON.parse(result.body);
    expect(response.success).toBe(true);
    expect(response.data.name).toBe('Test Store');
    expect(response.data.id).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}/);
  });

  test('handles validation errors', async () => {
    const event = createAPIGatewayEvent({
      body: JSON.stringify({
        name: '', // Invalid: empty name
        domain: 'test-store',
      }),
      headers: {
        Authorization: 'Bearer valid-jwt-token',
      },
    });

    const result = await handler(event);

    expect(result.statusCode).toBe(400);

    const response = JSON.parse(result.body);
    expect(response.success).toBe(false);
    expect(response.error.code).toBe('VALIDATION_ERROR');
  });

  test('handles duplicate domain errors', async () => {
    // Mock DynamoDB conditional check failure
    mockDynamoClient.on(PutCommand).rejects(
      new ConditionalCheckFailedException({
        message: 'The conditional request failed',
        $metadata: {},
      })
    );

    const event = createAPIGatewayEvent({
      body: JSON.stringify({
        name: 'Test Store',
        domain: 'existing-domain',
      }),
      headers: {
        Authorization: 'Bearer valid-jwt-token',
      },
    });

    const result = await handler(event);

    expect(result.statusCode).toBe(409);

    const response = JSON.parse(result.body);
    expect(response.success).toBe(false);
    expect(response.error.code).toBe('DOMAIN_EXISTS');
  });
});
```

## Team Scale Implementation Patterns

### Solo Developer (1 person)
```markdown
## Constitutional Implementation for Solo Developer

### Simplified Quality Gates
- **TDD**: Red-Green-Refactor for all features
- **Testing**: 70%+ coverage minimum
- **Documentation**: README + API docs
- **CI/CD**: Basic GitHub Actions workflow
- **Monitoring**: Simple health checks + logs

### Workflow
1. `/idea` - Validate concept, define MVP
2. `/specify` - Basic API contracts, simple data model
3. `/plan` - Minimal architecture, focus on core features
4. `/tasks` - Small, manageable tasks (1-2 hours each)

### Technology Choices
- **Stack**: Keep it simple (React + Node.js + SQLite)
- **Deployment**: Serverless or simple VPS
- **Monitoring**: Built-in cloud provider tools
- **Testing**: Jest + Supertest for integration
```

### Small Team (2-5 people)
```markdown
## Constitutional Implementation for Small Team

### Enhanced Quality Gates
- **TDD**: Mandatory for all features
- **Testing**: 85%+ coverage, integration tests required
- **Code Review**: All changes require peer review
- **Documentation**: API docs + deployment guides
- **CI/CD**: Automated testing + staging deployment

### Workflow
1. **Planning**: Weekly planning meetings using spec-kit phases
2. **Development**: Parallel task execution with [P] markers
3. **Review**: Constitutional checklist for all PRs
4. **Integration**: Continuous integration with real dependencies

### Collaboration Patterns
- **Pair Programming**: For complex features
- **Code Review**: Constitutional compliance checks
- **Daily Standups**: Progress on spec-kit phases
- **Retrospectives**: Framework process improvements
```

### Medium Team (6-15 people)
```markdown
## Constitutional Implementation for Medium Team

### Advanced Quality Gates
- **Architecture Reviews**: Senior developer approval required
- **Contract Testing**: All service interfaces tested
- **Performance Testing**: Automated load testing
- **Security Scanning**: OWASP compliance automation
- **Documentation**: Comprehensive API + user docs

### Organization Structure
- **Feature Teams**: 3-4 developers per feature
- **Architecture Team**: Cross-cutting concerns
- **QA Team**: Constitutional compliance validation
- **DevOps Team**: Infrastructure and deployment

### Process Enhancement
- **Spec Reviews**: Formal specification approval process
- **Implementation Plans**: Detailed technical planning phase
- **Integration Testing**: Real dependency testing mandated
- **Performance Baselines**: Established SLAs and monitoring
```

## Technology Integration Examples

### Frontend Frameworks

#### React with Constitutional Principles
```typescript
// Constitutional React component structure
interface StoreFormProps {
  onSubmit: (store: CreateStoreRequest) => Promise<StoreResponse>;
  onError: (error: FormError) => void;
  loading?: boolean;
}

// Test-first component development
describe('StoreForm', () => {
  test('validates required fields before submission', async () => {
    const onSubmit = jest.fn();
    const onError = jest.fn();

    render(<StoreForm onSubmit={onSubmit} onError={onError} />);

    const submitButton = screen.getByRole('button', { name: /create store/i });
    fireEvent.click(submitButton);

    expect(onError).toHaveBeenCalledWith({
      code: 'VALIDATION_ERROR',
      message: 'Name and domain are required',
      fields: ['name', 'domain'],
    });
    expect(onSubmit).not.toHaveBeenCalled();
  });

  test('submits valid form data', async () => {
    const onSubmit = jest.fn().mockResolvedValue({
      success: true,
      data: { id: '123', name: 'Test Store', domain: 'test' },
    });

    render(<StoreForm onSubmit={onSubmit} onError={jest.fn()} />);

    fireEvent.change(screen.getByLabelText(/store name/i), {
      target: { value: 'Test Store' },
    });
    fireEvent.change(screen.getByLabelText(/domain/i), {
      target: { value: 'test-store' },
    });

    fireEvent.click(screen.getByRole('button', { name: /create store/i }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({
        name: 'Test Store',
        domain: 'test-store',
      });
    });
  });
});

// Implementation with constitutional error handling
export const StoreForm: React.FC<StoreFormProps> = ({
  onSubmit,
  onError,
  loading = false,
}) => {
  const [formData, setFormData] = useState<CreateStoreRequest>({
    name: '',
    domain: '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  // Constitutional validation
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Store name is required';
    } else if (formData.name.trim().length < 2) {
      newErrors.name = 'Store name must be at least 2 characters';
    }

    if (!formData.domain.trim()) {
      newErrors.domain = 'Domain is required';
    } else if (!/^[a-z0-9][a-z0-9-]*[a-z0-9]$/.test(formData.domain)) {
      newErrors.domain = 'Domain must contain only lowercase letters, numbers, and hyphens';
    }

    setErrors(newErrors);

    if (Object.keys(newErrors).length > 0) {
      onError({
        code: 'VALIDATION_ERROR',
        message: 'Please fix the validation errors',
        fields: Object.keys(newErrors),
      });
      return false;
    }

    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    try {
      await onSubmit(formData);
    } catch (error) {
      onError({
        code: 'SUBMISSION_ERROR',
        message: error.message || 'Failed to create store',
        fields: [],
      });
    }
  };

  return (
    <form onSubmit={handleSubmit} className="store-form">
      <div className="form-group">
        <label htmlFor="name">Store Name</label>
        <input
          id="name"
          type="text"
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          className={errors.name ? 'error' : ''}
          disabled={loading}
          aria-describedby={errors.name ? 'name-error' : undefined}
        />
        {errors.name && (
          <div id="name-error" className="error-message" role="alert">
            {errors.name}
          </div>
        )}
      </div>

      <div className="form-group">
        <label htmlFor="domain">Domain</label>
        <input
          id="domain"
          type="text"
          value={formData.domain}
          onChange={(e) => setFormData({ ...formData, domain: e.target.value })}
          className={errors.domain ? 'error' : ''}
          disabled={loading}
          aria-describedby={errors.domain ? 'domain-error' : undefined}
        />
        {errors.domain && (
          <div id="domain-error" className="error-message" role="alert">
            {errors.domain}
          </div>
        )}
      </div>

      <button
        type="submit"
        disabled={loading}
        className="submit-button"
        aria-describedby="submit-status"
      >
        {loading ? 'Creating Store...' : 'Create Store'}
      </button>

      <div id="submit-status" aria-live="polite" className="sr-only">
        {loading ? 'Form is being submitted' : ''}
      </div>
    </form>
  );
};
```

### Database Integration Patterns

#### PostgreSQL with Constitutional Schema
```sql
-- Constitutional database schema with complete constraints
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Enums for type safety
CREATE TYPE store_status AS ENUM ('active', 'inactive', 'deleted');
CREATE TYPE user_role AS ENUM ('owner', 'admin', 'user');

-- Users table with constitutional requirements
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role user_role NOT NULL DEFAULT 'user',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1,

    -- Constitutional constraints
    CONSTRAINT users_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$'),
    CONSTRAINT users_name_length CHECK (
        length(trim(first_name)) >= 1 AND
        length(trim(last_name)) >= 1
    ),
    CONSTRAINT users_password_length CHECK (length(password_hash) >= 60) -- bcrypt minimum
);

-- Stores table with full constitutional compliance
CREATE TABLE stores (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status store_status NOT NULL DEFAULT 'active',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1,

    -- Constitutional business rules
    CONSTRAINT stores_name_length CHECK (length(trim(name)) >= 2),
    CONSTRAINT stores_domain_format CHECK (domain ~* '^[a-z0-9][a-z0-9-]*[a-z0-9]$'),
    CONSTRAINT stores_domain_length CHECK (length(domain) BETWEEN 3 AND 63),
    CONSTRAINT stores_valid_settings CHECK (jsonb_typeof(settings) = 'object')
);

-- Audit table for constitutional compliance
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_name VARCHAR(100) NOT NULL,
    record_id UUID NOT NULL,
    action VARCHAR(20) NOT NULL, -- INSERT, UPDATE, DELETE
    old_values JSONB,
    new_values JSONB,
    user_id UUID REFERENCES users(id),
    correlation_id UUID,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT audit_valid_action CHECK (action IN ('INSERT', 'UPDATE', 'DELETE'))
);

-- Performance indexes (constitutional requirement)
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_created_at ON users(created_at);

CREATE INDEX idx_stores_owner_id ON stores(owner_id);
CREATE INDEX idx_stores_domain ON stores(domain);
CREATE INDEX idx_stores_status ON stores(status);
CREATE INDEX idx_stores_created_at ON stores(created_at);

-- Full-text search support
CREATE INDEX idx_stores_name_search ON stores USING GIN (to_tsvector('english', name));
CREATE INDEX idx_stores_description_search ON stores USING GIN (to_tsvector('english', description));

-- Audit indexes
CREATE INDEX idx_audit_table_record ON audit_log(table_name, record_id);
CREATE INDEX idx_audit_timestamp ON audit_log(timestamp);
CREATE INDEX idx_audit_user_id ON audit_log(user_id);

-- Constitutional triggers for audit logging
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO audit_log (
        table_name,
        record_id,
        action,
        old_values,
        new_values,
        user_id,
        correlation_id
    ) VALUES (
        TG_TABLE_NAME,
        COALESCE(NEW.id, OLD.id),
        TG_OP,
        CASE WHEN TG_OP = 'DELETE' THEN row_to_json(OLD) ELSE NULL END,
        CASE WHEN TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN row_to_json(NEW) ELSE NULL END,
        current_setting('app.current_user_id', true)::UUID,
        current_setting('app.correlation_id', true)::UUID
    );

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Apply audit triggers
CREATE TRIGGER users_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON users
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER stores_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON stores
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Constitutional update timestamp trigger
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_update_timestamp
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER stores_update_timestamp
    BEFORE UPDATE ON stores
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- Constitutional RLS (Row Level Security) policies
ALTER TABLE stores ENABLE ROW LEVEL SECURITY;

-- Store owners can see and modify their own stores
CREATE POLICY stores_owner_policy ON stores
    FOR ALL
    TO authenticated_users
    USING (owner_id = current_setting('app.current_user_id')::UUID);

-- Admins can see all stores
CREATE POLICY stores_admin_policy ON stores
    FOR ALL
    TO admin_users
    USING (true);

-- Constitutional data validation functions
CREATE OR REPLACE FUNCTION validate_store_domain(domain TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    -- Check format
    IF domain !~* '^[a-z0-9][a-z0-9-]*[a-z0-9]$' THEN
        RETURN FALSE;
    END IF;

    -- Check length
    IF length(domain) < 3 OR length(domain) > 63 THEN
        RETURN FALSE;
    END IF;

    -- Check for reserved words
    IF domain IN ('admin', 'api', 'www', 'mail', 'ftp', 'localhost') THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Sample data for testing (constitutional requirement)
INSERT INTO users (id, email, password_hash, first_name, last_name, role, email_verified) VALUES
('123e4567-e89b-12d3-a456-426614174000', 'admin@example.com', '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj0NQqLYe.Pk', 'Admin', 'User', 'admin', true),
('123e4567-e89b-12d3-a456-426614174001', 'user@example.com', '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj0NQqLYe.Pk', 'Regular', 'User', 'user', true);

INSERT INTO stores (id, name, domain, description, owner_id) VALUES
('223e4567-e89b-12d3-a456-426614174000', 'Demo Store', 'demo-store', 'A demonstration store for testing', '123e4567-e89b-12d3-a456-426614174001');
```

## Continuous Integration Examples

### GitHub Actions Workflow (Constitutional)
```yaml
# .github/workflows/constitutional-ci.yml
name: Constitutional CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  NODE_VERSION: '18'
  CONSTITUTIONAL_FRAMEWORK_VERSION: '1.0.0'

jobs:
  constitutional-validation:
    name: Constitutional Framework Validation
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'

    - name: Install dependencies
      run: npm ci

    - name: Validate constitutional compliance
      run: |
        echo "Checking constitutional framework compliance..."
        # Verify TDD structure
        npm run test:structure
        # Verify interface contracts exist
        npm run validate:contracts
        # Verify real dependency testing
        npm run validate:integration-tests
        # Verify observability setup
        npm run validate:observability

    - name: Run linting (constitutional code quality)
      run: npm run lint

    - name: Type checking (constitutional type safety)
      run: npm run typecheck

  unit-tests:
    name: Unit Tests (TDD Constitutional Requirement)
    runs-on: ubuntu-latest
    needs: constitutional-validation

    steps:
    - uses: actions/checkout@v4

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'

    - name: Install dependencies
      run: npm ci

    - name: Run unit tests with coverage
      run: npm run test:unit -- --coverage

    - name: Verify constitutional coverage threshold (90%+)
      run: |
        COVERAGE=$(npm run test:coverage -- --silent | grep "All files" | awk '{print $10}' | sed 's/%//')
        if [ "$COVERAGE" -lt 90 ]; then
          echo "Coverage $COVERAGE% is below constitutional requirement of 90%"
          exit 1
        fi
        echo "Coverage $COVERAGE% meets constitutional requirements"

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage/lcov.info

  contract-tests:
    name: Contract Tests (Constitutional Interface Validation)
    runs-on: ubuntu-latest
    needs: constitutional-validation

    steps:
    - uses: actions/checkout@v4

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'

    - name: Install dependencies
      run: npm ci

    - name: Start mock services for contract testing
      run: |
        docker-compose -f docker-compose.test.yml up -d
        sleep 10

    - name: Run API contract tests
      run: npm run test:contracts

    - name: Validate OpenAPI specifications
      run: |
        npx swagger-codegen validate -i contracts/openapi.yaml
        echo "All API contracts are valid"

    - name: Stop test services
      run: docker-compose -f docker-compose.test.yml down

  integration-tests:
    name: Integration Tests (Constitutional Real Dependencies)
    runs-on: ubuntu-latest
    needs: [unit-tests, contract-tests]

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: testpassword
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
    - uses: actions/checkout@v4

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'

    - name: Install dependencies
      run: npm ci

    - name: Run database migrations
      run: npm run db:migrate
      env:
        DATABASE_URL: postgresql://postgres:testpassword@localhost:5432/testdb

    - name: Seed test data
      run: npm run db:seed
      env:
        DATABASE_URL: postgresql://postgres:testpassword@localhost:5432/testdb

    - name: Run integration tests with real dependencies
      run: npm run test:integration
      env:
        DATABASE_URL: postgresql://postgres:testpassword@localhost:5432/testdb
        REDIS_URL: redis://localhost:6379
        LOG_LEVEL: error

    - name: Validate no mocks in integration tests (constitutional requirement)
      run: |
        if grep -r "jest.mock\|sinon.stub\|td.replace" tests/integration/; then
          echo "ERROR: Found mocks in integration tests - violates constitutional framework"
          exit 1
        fi
        echo "Integration tests properly use real dependencies"

  security-scan:
    name: Security Scanning (Constitutional Security Requirement)
    runs-on: ubuntu-latest
    needs: constitutional-validation

    steps:
    - uses: actions/checkout@v4

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'

    - name: Install dependencies
      run: npm ci

    - name: Run npm audit (dependency vulnerabilities)
      run: npm audit --audit-level high

    - name: Run OWASP dependency check
      uses: dependency-check/Dependency-Check_Action@main
      with:
        project: 'constitutional-project'
        path: '.'
        format: 'HTML,JSON'

    - name: Run static security analysis
      run: npx semgrep --config=auto

    - name: Upload security scan results
      uses: actions/upload-artifact@v3
      with:
        name: security-scan-results
        path: reports/

  performance-tests:
    name: Performance Testing (Constitutional Performance Validation)
    runs-on: ubuntu-latest
    needs: integration-tests
    if: github.ref == 'refs/heads/main'

    steps:
    - uses: actions/checkout@v4

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'

    - name: Install dependencies
      run: npm ci

    - name: Start application for performance testing
      run: |
        npm run build
        npm start &
        sleep 30
      env:
        NODE_ENV: production

    - name: Run load tests
      run: |
        npx k6 run tests/performance/load-test.js
        npx k6 run tests/performance/stress-test.js

    - name: Validate performance metrics
      run: |
        # Check if 95th percentile response time is under 200ms
        RESPONSE_TIME=$(cat performance-results.json | jq '.metrics.http_req_duration.values.p95')
        if (( $(echo "$RESPONSE_TIME > 200" | bc -l) )); then
          echo "Performance degraded: 95th percentile response time is ${RESPONSE_TIME}ms"
          exit 1
        fi
        echo "Performance meets constitutional requirements: ${RESPONSE_TIME}ms"

  build-and-deploy:
    name: Build and Deploy
    runs-on: ubuntu-latest
    needs: [integration-tests, security-scan, performance-tests]
    if: github.ref == 'refs/heads/main'

    steps:
    - uses: actions/checkout@v4

    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: ${{ env.NODE_VERSION }}
        cache: 'npm'

    - name: Install dependencies
      run: npm ci

    - name: Build application
      run: npm run build

    - name: Generate deployment artifacts
      run: |
        # Create deployment package with constitutional metadata
        tar -czf deployment.tar.gz dist/ package.json package-lock.json

        # Generate deployment manifest
        cat > deployment-manifest.json << EOF
        {
          "version": "${{ github.sha }}",
          "branch": "${{ github.ref_name }}",
          "constitutional_framework_version": "${{ env.CONSTITUTIONAL_FRAMEWORK_VERSION }}",
          "build_timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
          "tests_passed": {
            "unit_tests": true,
            "contract_tests": true,
            "integration_tests": true,
            "security_scan": true,
            "performance_tests": true
          },
          "compliance": {
            "tdd_enforced": true,
            "real_dependencies_tested": true,
            "interface_contracts_validated": true,
            "observability_configured": true,
            "security_scanned": true
          }
        }
        EOF

    - name: Deploy to staging
      run: |
        echo "Deploying to staging environment..."
        # Deploy with constitutional compliance validation
        ./scripts/deploy.sh staging deployment.tar.gz deployment-manifest.json

    - name: Run smoke tests on staging
      run: |
        sleep 60  # Wait for deployment
        npm run test:smoke -- --base-url https://staging.example.com

    - name: Deploy to production (if staging tests pass)
      run: |
        echo "Deploying to production environment..."
        ./scripts/deploy.sh production deployment.tar.gz deployment-manifest.json

    - name: Verify production deployment
      run: |
        sleep 60  # Wait for deployment
        npm run test:smoke -- --base-url https://api.example.com

        # Verify health endpoints
        curl -f https://api.example.com/health
        curl -f https://api.example.com/metrics

  notify-success:
    name: Notify Deployment Success
    runs-on: ubuntu-latest
    needs: build-and-deploy

    steps:
    - name: Notify team of successful deployment
      run: |
        curl -X POST -H 'Content-type: application/json' \
          --data '{"text":"✅ Constitutional deployment successful for `${{ github.sha }}` to production. All constitutional requirements validated."}' \
          ${{ secrets.SLACK_WEBHOOK_URL }}
```

This comprehensive implementation guide demonstrates how the Spec-Kit constitutional framework can be applied across different project types, team sizes, and technology stacks while maintaining consistency, quality, and constitutional compliance throughout the development lifecycle.

---
**Document Version**: 1.0
**Last Updated**: [Date]
**Framework Version**: Constitutional 3.0.0
**Compatibility**: All supported AI assistants and development environments