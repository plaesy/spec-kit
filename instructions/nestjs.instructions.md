---
applyTo: '**/*.ts, **/*.js, **/*.json, **/*.spec.ts, **/*.e2e-spec.ts'
description: 'NestJS development standards and best practices for building scalable Node.js server-side applications'
---

# NestJS Development Best Practices

## Mission
Guide developers toward scalable, maintainable, well-architected server-side apps using NestJS principles and best practices.

## Core NestJS Principles

### 1. Dependency Injection (DI)
NestJS's DI container manages provider instantiation/lifetime.
- `@Injectable()` for services, repositories, other providers
- Inject via constructor parameters with proper typing
- Prefer interface-based DI for testability
- Custom providers for specific instantiation logic

### 2. Modular Architecture
Organize code into feature modules encapsulating related functionality.
- Feature modules via `@Module()`
- Import only what's needed, avoid circular deps
- `forRoot()`/`forFeature()` for configurable modules
- Shared modules for common functionality

### 3. Decorators and Metadata
Decorators define routes, middleware, guards, and other framework features.
- `@Controller()`, `@Get()`, `@Post()`, `@Injectable()` as appropriate
- `class-validator` validation decorators
- Custom decorators for cross-cutting concerns
- Metadata reflection for advanced scenarios

## Project Structure Best Practices

### Recommended Directory Structure
```
src/
├── app.module.ts
├── main.ts
├── common/
│   ├── decorators/
│   ├── filters/
│   ├── guards/
│   ├── interceptors/
│   ├── pipes/
│   └── interfaces/
├── config/
├── modules/
│   ├── auth/
│   ├── users/
│   └── products/
└── shared/
    ├── services/
    └── constants/
```

### File Naming Conventions
- Controllers: `*.controller.ts` (e.g. `users.controller.ts`)
- Services: `*.service.ts`
- Modules: `*.module.ts`
- DTOs: `*.dto.ts` (e.g. `create-user.dto.ts`)
- Entities: `*.entity.ts`
- Guards: `*.guard.ts`
- Interceptors: `*.interceptor.ts`
- Pipes: `*.pipe.ts`
- Filters: `*.filter.ts`

## API Development Patterns

### 1. Controllers
Keep thin - delegate business logic to services; proper HTTP methods/status codes; comprehensive DTO-based input validation; guards/interceptors at the appropriate level.

```typescript
@Controller('users')
@UseGuards(AuthGuard)
export class UsersController {
  constructor(private readonly usersService: UsersService) {}

  @Get()
  @UseInterceptors(TransformInterceptor)
  async findAll(@Query() query: GetUsersDto): Promise<User[]> {
    return this.usersService.findAll(query);
  }

  @Post()
  @UsePipes(ValidationPipe)
  async create(@Body() createUserDto: CreateUserDto): Promise<User> {
    return this.usersService.create(createUserDto);
  }
}
```

### 2. Services
Business logic lives here, not in controllers; constructor-based DI; focused, single-responsibility services; handle errors appropriately, let filters catch them.

```typescript
@Injectable()
export class UsersService {
  constructor(
    @InjectRepository(User)
    private readonly userRepository: Repository<User>,
    private readonly emailService: EmailService,
  ) {}

  async create(createUserDto: CreateUserDto): Promise<User> {
    const user = this.userRepository.create(createUserDto);
    const savedUser = await this.userRepository.save(user);
    await this.emailService.sendWelcomeEmail(savedUser.email);
    return savedUser;
  }
}
```

### 3. DTOs and Validation
`class-validator` decorators for input validation; separate DTOs per operation (create/update/query); proper transformation with `class-transformer`.

```typescript
export class CreateUserDto {
  @IsString()
  @IsNotEmpty()
  @Length(2, 50)
  name: string;

  @IsEmail()
  email: string;

  @IsString()
  @MinLength(8)
  @Matches(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/, {
    message: 'Password must contain uppercase, lowercase and number',
  })
  password: string;
}
```

## Database Integration

### TypeORM Integration
Primary ORM for DB operations; entities with proper decorators/relationships; repository pattern for data access; migrations for schema changes.

```typescript
@Entity('users')
export class User {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({ unique: true })
  email: string;

  @Column()
  name: string;

  @Column({ select: false })
  password: string;

  @OneToMany(() => Post, post => post.author)
  posts: Post[];

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;
}
```

### Custom Repositories
Extend base repository functionality when needed; implement complex queries in repository methods; query builders for dynamic queries.

## Authentication and Authorization

### JWT Authentication
JWT-based auth with Passport; guards to protect routes; custom decorators for user context.

```typescript
@Injectable()
export class JwtAuthGuard extends AuthGuard('jwt') {
  canActivate(context: ExecutionContext): boolean | Promise<boolean> {
    return super.canActivate(context);
  }

  handleRequest(err: any, user: any, info: any) {
    if (err || !user) {
      throw err || new UnauthorizedException();
    }
    return user;
  }
}
```

### Role-Based Access Control
RBAC via custom guards/decorators; metadata to define required roles; flexible permission systems.

```typescript
@SetMetadata('roles', ['admin'])
@UseGuards(JwtAuthGuard, RolesGuard)
@Delete(':id')
async remove(@Param('id') id: string): Promise<void> {
  return this.usersService.remove(id);
}
```

## Error Handling and Logging

### Exception Filters
Global exception filters for consistent error responses; handle exception types appropriately; log errors with proper context.

```typescript
@Catch()
export class AllExceptionsFilter implements ExceptionFilter {
  private readonly logger = new Logger(AllExceptionsFilter.name);

  catch(exception: unknown, host: ArgumentsHost): void {
    const ctx = host.switchToHttp();
    const response = ctx.getResponse<Response>();
    const request = ctx.getRequest<Request>();

    const status = exception instanceof HttpException 
      ? exception.getStatus() 
      : HttpStatus.INTERNAL_SERVER_ERROR;

    this.logger.error(`${request.method} ${request.url}`, exception);

    response.status(status).json({
      statusCode: status,
      timestamp: new Date().toISOString(),
      path: request.url,
      message: exception instanceof HttpException 
        ? exception.message 
        : 'Internal server error',
    });
  }
}
```

### Logging
Built-in `Logger` class for consistent logging; proper log levels (error/warn/log/debug/verbose); contextual info in logs.

## Testing Strategies

### Unit Testing
Test services independently with mocks; Jest as the framework; comprehensive suites for business logic.

```typescript
describe('UsersService', () => {
  let service: UsersService;
  let repository: Repository<User>;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        UsersService,
        {
          provide: getRepositoryToken(User),
          useValue: {
            create: jest.fn(),
            save: jest.fn(),
            find: jest.fn(),
          },
        },
      ],
    }).compile();

    service = module.get<UsersService>(UsersService);
    repository = module.get<Repository<User>>(getRepositoryToken(User));
  });

  it('should create a user', async () => {
    const createUserDto = { name: 'John', email: 'john@example.com' };
    const user = { id: '1', ...createUserDto };

    jest.spyOn(repository, 'create').mockReturnValue(user as User);
    jest.spyOn(repository, 'save').mockResolvedValue(user as User);

    expect(await service.create(createUserDto)).toEqual(user);
  });
});
```

### Integration Testing
`TestingModule` for integration tests; test full request/response cycles; mock external deps appropriately.

### E2E Testing
Test complete application flows; `supertest` for HTTP testing; test auth/authz flows.

## Performance and Security

### Performance Optimization
Redis caching; interceptors for response transformation; proper DB indexing; pagination for large datasets.

### Security Best Practices
Validate all inputs (`class-validator`); rate limiting against abuse; appropriate CORS; sanitize outputs against XSS; env vars for sensitive config.

```typescript
// Rate limiting example
@Controller('auth')
@UseGuards(ThrottlerGuard)
export class AuthController {
  @Post('login')
  @Throttle(5, 60) // 5 requests per minute
  async login(@Body() loginDto: LoginDto) {
    return this.authService.login(loginDto);
  }
}
```

## Configuration Management

### Environment Configuration
`@nestjs/config` for configuration; validate config at startup; different configs per environment.

```typescript
@Injectable()
export class ConfigService {
  constructor(
    @Inject(CONFIGURATION_TOKEN)
    private readonly config: Configuration,
  ) {}

  get databaseUrl(): string {
    return this.config.database.url;
  }

  get jwtSecret(): string {
    return this.config.jwt.secret;
  }
}
```

## Common Pitfalls to Avoid
- **Circular Dependencies**: avoid modules importing each other in a cycle
- **Heavy Controllers**: no business logic in controllers
- **Missing Error Handling**: always handle errors appropriately
- **Improper DI Usage**: don't manually instantiate when DI can handle it
- **Missing Validation**: always validate input data
- **Synchronous Operations**: async/await for DB and external API calls
- **Memory Leaks**: dispose subscriptions/event listeners properly

## Development Workflow

### Development Setup
1. NestJS CLI for scaffolding: `nest generate module users`
2. Consistent file organization
3. TypeScript strict mode
4. Comprehensive ESLint
5. Prettier for formatting

### Code Review Checklist
- [ ] Proper use of decorators and dependency injection
- [ ] Input validation with DTOs and class-validator
- [ ] Appropriate error handling and exception filters
- [ ] Consistent naming conventions
- [ ] Proper module organization and imports
- [ ] Security considerations (authentication, authorization, input sanitization)
- [ ] Performance considerations (caching, database optimization)
- [ ] Comprehensive testing coverage

## Conclusion
Follow these practices to build maintainable, testable, efficient NestJS applications leveraging TypeScript and modern development patterns.

---

<!-- End of NestJS Instructions -->
