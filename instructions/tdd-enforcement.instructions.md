---
description: "Enforces Test-Driven Development (TDD) practices and commit patterns"
applyTo: '**/*.js, **/*.ts, **/*.java, **/*.py, **/*.rb, **/*.go, **/*.rs, **/*.cs, **/*.cpp, **/*.c, **/*.swift, **/*.kt, **/*.kts'
---

# Test-Driven Development (TDD) Instructions

## Overview

These instructions enforce the constitutional requirement for Test-First Development across all Spec-Kit projects. TDD is non-negotiable and must be followed for all implementation work.

## TDD Cycle: Red-Green-Refactor

### Phase 1: RED - Write Failing Test
**Objective**: Write a test that fails because the functionality doesn't exist yet

**Process**:
1. **Understand Requirements**: Review spec and plan for specific functionality
2. **Write Test First**: Create test that describes expected behavior
3. **Run Test**: Verify test fails (RED state)
4. **Commit Failing Test**: Use git commit with message "RED: [test description]"

**Example Commit Messages**:
```
RED: Add user authentication validation test
RED: Implement API endpoint response validation
RED: Add database connection error handling test
```

### Phase 2: GREEN - Make Test Pass
**Objective**: Write minimal code to make the test pass

**Process**:
1. **Write Minimal Code**: Implement just enough to make test pass
2. **Run Test**: Verify test passes (GREEN state)
3. **Run All Tests**: Ensure no regression in existing functionality
4. **Commit Passing Code**: Use git commit with message "GREEN: [implementation description]"

**Example Commit Messages**:
```
GREEN: Implement user authentication validation
GREEN: Add API endpoint with proper response format
GREEN: Handle database connection errors gracefully
```

### Phase 3: REFACTOR - Improve Code Quality
**Objective**: Improve code structure while maintaining functionality

**Process**:
1. **Identify Improvements**: Look for code smells, duplication, complexity
2. **Refactor Code**: Improve structure without changing behavior
3. **Run All Tests**: Ensure all tests still pass
4. **Commit Refactored Code**: Use git commit with message "REFACTOR: [improvement description]"

**Example Commit Messages**:
```
REFACTOR: Extract authentication logic to separate module
REFACTOR: Simplify API response handling
REFACTOR: Improve database error handling structure
```

## TDD Best Practices

### Test Design Principles
1. **Arrange-Act-Assert**: Structure tests clearly
2. **One Assertion Per Test**: Focus on single behavior
3. **Descriptive Test Names**: Test name explains expected behavior
4. **Test Edge Cases**: Include boundary conditions and error scenarios

### BDD Integration with TDD
**Objective**: Combine TDD with Behavior-Driven Development for better requirement clarity

**BDD-TDD Process**:
1. **Given-When-Then**: Structure acceptance criteria using BDD format
2. **Outside-In TDD**: Start with acceptance tests, work down to unit tests
3. **Living Documentation**: Tests serve as executable specifications

**Example BDD-TDD Structure (Multi-Language)**:

**JavaScript/TypeScript**:
```javascript
// Feature: User Authentication
// Scenario: User logs in with valid credentials
describe('Feature: User Authentication', () => {
  describe('Scenario: User logs in with valid credentials', () => {
    it('Given a registered user exists, When they provide valid credentials, Then they should be logged in', async () => {
      // Given - Arrange
      const user = await UserFactory.create({ email: 'user@example.com' });
      
      // When - Act
      const result = await authService.login('user@example.com', 'validPassword');
      
      // Then - Assert
      expect(result.success).toBe(true);
      expect(result.token).toBeDefined();
    });
  });
});
```

**Java**:
```java
// Feature: User Authentication
// Scenario: User logs in with valid credentials
@Test
class UserAuthenticationTest {
    @Test
    @DisplayName("Given a registered user exists, When they provide valid credentials, Then they should be logged in")
    void userLoginWithValidCredentials() {
        // Given - Arrange
        User user = UserFactory.create("user@example.com");
        
        // When - Act
        AuthResult result = authService.login("user@example.com", "validPassword");
        
        // Then - Assert
        assertTrue(result.isSuccess());
        assertNotNull(result.getToken());
    }
}
```

**Python**:
```python
# Feature: User Authentication
# Scenario: User logs in with valid credentials
class TestUserAuthentication:
    def test_given_registered_user_when_valid_credentials_then_logged_in(self):
        # Given - Arrange
        user = UserFactory.create(email="user@example.com")
        
        # When - Act
        result = auth_service.login("user@example.com", "valid_password")
        
        # Then - Assert
        assert result.success == True
        assert result.token is not None
```

**Go**:
```go
// Feature: User Authentication
// Scenario: User logs in with valid credentials
func TestUserAuthentication_ValidCredentials(t *testing.T) {
    // Given - Arrange
    user := UserFactory.Create("user@example.com")
    
    // When - Act
    result := authService.Login("user@example.com", "validPassword")
    
    // Then - Assert
    if !result.Success {
        t.Error("Expected login to succeed")
    }
    if result.Token == "" {
        t.Error("Expected token to be present")
    }
}
```

### Code Design Principles
1. **YAGNI**: You Aren't Gonna Need It - implement only what tests require
2. **DRY**: Don't Repeat Yourself - refactor duplication
3. **SOLID**: Follow SOLID principles during refactoring
4. **Clean Code**: Write readable, maintainable code

### Test Organization Patterns

#### Test Builders Pattern
**Purpose**: Create complex test objects with readable, chainable API

**JavaScript/TypeScript**:
```javascript
class UserBuilder {
  constructor() {
    this.user = { id: 'default-id', email: 'test@example.com', role: 'user' };
  }
  
  withEmail(email) { this.user.email = email; return this; }
  withRole(role) { this.user.role = role; return this; }
  build() { return { ...this.user }; }
}
```

**Java**:
```java
public class UserBuilder {
    private String id = "default-id";
    private String email = "test@example.com";
    private String role = "user";
    
    public UserBuilder withEmail(String email) { this.email = email; return this; }
    public UserBuilder withRole(String role) { this.role = role; return this; }
    public User build() { return new User(id, email, role); }
}
```

**Python**:
```python
class UserBuilder:
    def __init__(self):
        self.user_data = {"id": "default-id", "email": "test@example.com", "role": "user"}
    
    def with_email(self, email):
        self.user_data["email"] = email
        return self
    
    def with_role(self, role):
        self.user_data["role"] = role
        return self
    
    def build(self):
        return User(**self.user_data)
```

**Go**:
```go
type UserBuilder struct {
    user User
}

func NewUserBuilder() *UserBuilder {
    return &UserBuilder{
        user: User{ID: "default-id", Email: "test@example.com", Role: "user"},
    }
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
    b.user.Email = email
    return b
}

func (b *UserBuilder) WithRole(role string) *UserBuilder {
    b.user.Role = role
    return b
}

func (b *UserBuilder) Build() User {
    return b.user
}
```

#### Object Mother Pattern
**Purpose**: Provide pre-configured objects for common test scenarios
```javascript
class UserMother {
  static validUser() {
    return {
      id: '123',
      email: 'valid@example.com',
      password: 'SecurePass123',
      role: 'user'
    };
  }
  
  static adminUser() {
    return {
      ...this.validUser(),
      role: 'admin'
    };
  }
  
  static userWithInvalidEmail() {
    return {
      ...this.validUser(),
      email: 'invalid-email'
    };
  }
}
```

#### Test Data Factory Pattern
**Purpose**: Generate test data with proper relationships and constraints
```javascript
class TestDataFactory {
  static async createUser(overrides = {}) {
    const userData = {
      email: `user${Date.now()}@example.com`,
      password: 'SecurePass123',
      ...overrides
    };
    return await User.create(userData);
  }
  
  static async createUserWithPosts(postCount = 3) {
    const user = await this.createUser();
    const posts = await Promise.all(
      Array(postCount).fill().map((_, i) => 
        Post.create({
          title: `Post ${i + 1}`,
          userId: user.id
        })
      )
    );
    return { user, posts };
  }
}
```

## TDD for Different Layers

### Unit Tests
**Scope**: Individual functions, methods, classes
**Tools**: Jest/Mocha (JS/TS), JUnit (Java), PyTest (Python), Go Test (Go), Cargo Test (Rust), NUnit (C#)
**Focus**: Business logic, data transformation, validation

**Example Structure (Multi-Language)**:

**JavaScript/TypeScript**:
```javascript
describe('UserValidator', () => {
  it('should validate email format correctly', () => {
    // Arrange
    const validator = new UserValidator();
    const validEmail = 'user@example.com';
    
    // Act
    const result = validator.validateEmail(validEmail);
    
    // Assert
    expect(result.isValid).toBe(true);
  });
});
```

**Java**:
```java
@Test
class UserValidatorTest {
    @Test
    void shouldValidateEmailFormatCorrectly() {
        // Arrange
        UserValidator validator = new UserValidator();
        String validEmail = "user@example.com";
        
        // Act
        ValidationResult result = validator.validateEmail(validEmail);
        
        // Assert
        assertTrue(result.isValid());
    }
}
```

**Python**:
```python
class TestUserValidator:
    def test_should_validate_email_format_correctly(self):
        # Arrange
        validator = UserValidator()
        valid_email = "user@example.com"
        
        # Act
        result = validator.validate_email(valid_email)
        
        # Assert
        assert result.is_valid == True
```

**Go**:
```go
func TestUserValidator_ValidateEmail(t *testing.T) {
    // Arrange
    validator := NewUserValidator()
    validEmail := "user@example.com"
    
    // Act
    result := validator.ValidateEmail(validEmail)
    
    // Assert
    if !result.IsValid {
        t.Errorf("Expected email to be valid, got invalid")
    }
}
```

**Rust**:
```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn should_validate_email_format_correctly() {
        // Arrange
        let validator = UserValidator::new();
        let valid_email = "user@example.com";
        
        // Act
        let result = validator.validate_email(valid_email);
        
        // Assert
        assert!(result.is_valid);
    }
}
```

**C#**:
```csharp
[TestClass]
public class UserValidatorTests
{
    [TestMethod]
    public void ShouldValidateEmailFormatCorrectly()
    {
        // Arrange
        var validator = new UserValidator();
        var validEmail = "user@example.com";
        
        // Act
        var result = validator.ValidateEmail(validEmail);
        
        // Assert
        Assert.IsTrue(result.IsValid);
    }
}
```

### Integration Tests
**Scope**: Component interactions, API endpoints, database operations
**Tools**: Supertest/Cypress (JS/TS), TestContainers (Java), Requests (Python), Testify (Go), Reqwest (Rust)
**Focus**: Component integration, data flow, external dependencies

**Example Structure (Multi-Language)**:

**JavaScript/TypeScript**:
```javascript
describe('User API', () => {
  it('should create user and return user ID', async () => {
    const userData = { email: 'user@example.com', name: 'Test User' };
    const response = await request(app).post('/api/users').send(userData);
    
    expect(response.status).toBe(201);
    expect(response.body.userId).toBeDefined();
  });
});
```

**Java**:
```java
@SpringBootTest
@AutoConfigureTestDatabase
class UserApiIntegrationTest {
    @Test
    void shouldCreateUserAndReturnUserId() {
        // Arrange
        UserDto userData = new UserDto("user@example.com", "Test User");
        
        // Act
        ResponseEntity<UserResponse> response = testRestTemplate
            .postForEntity("/api/users", userData, UserResponse.class);
        
        // Assert
        assertEquals(HttpStatus.CREATED, response.getStatusCode());
        assertNotNull(response.getBody().getUserId());
    }
}
```

**Python**:
```python
class TestUserAPI:
    def test_should_create_user_and_return_user_id(self, client):
        # Arrange
        user_data = {"email": "user@example.com", "name": "Test User"}
        
        # Act
        response = client.post("/api/users", json=user_data)
        
        # Assert
        assert response.status_code == 201
        assert "user_id" in response.json()
```

**Go**:
```go
func TestUserAPI_CreateUser(t *testing.T) {
    // Arrange
    userData := UserRequest{Email: "user@example.com", Name: "Test User"}
    body, _ := json.Marshal(userData)
    
    // Act
    resp, err := http.Post(testServer.URL+"/api/users", "application/json", bytes.NewBuffer(body))
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
```

### End-to-End Tests
**Scope**: Complete user workflows, system behavior
**Tools**: Playwright, Cypress, Selenium
**Focus**: User journeys, system integration, business workflows

**Example Structure**:
```javascript
test('User can register and login successfully', async ({ page }) => {
  // Arrange
  await page.goto('/register');
  
  // Act
  await page.fill('#email', 'user@example.com');
  await page.fill('#password', 'SecurePass123');
  await page.click('#register-button');
  
  // Assert
  await expect(page.locator('#welcome-message')).toBeVisible();
});
```

### Performance Tests
**Scope**: Non-functional requirements, response times, throughput
**Tools**: Artillery/K6 (JS), JMeter (Java), Locust (Python), Vegeta (Go), Criterion (Rust)
**Focus**: Performance benchmarks, load testing, optimization validation

**TDD for Performance (Multi-Language)**:

**JavaScript/TypeScript**:
```javascript
describe('API Performance', () => {
  it('should handle user creation within 200ms', async () => {
    const userData = UserMother.validUser();
    const startTime = performance.now();
    
    const response = await request(app).post('/api/users').send(userData);
    const responseTime = performance.now() - startTime;
    
    expect(response.status).toBe(201);
    expect(responseTime).toBeLessThan(200);
  });
});
```

**Java**:
```java
@Test
void shouldHandleUserCreationWithin200ms() {
    // Arrange
    UserDto userData = UserMother.validUser();
    long startTime = System.currentTimeMillis();
    
    // Act
    ResponseEntity<UserResponse> response = restTemplate
        .postForEntity("/api/users", userData, UserResponse.class);
    long responseTime = System.currentTimeMillis() - startTime;
    
    // Assert
    assertEquals(HttpStatus.CREATED, response.getStatusCode());
    assertTrue(responseTime < 200);
}
```

**Python**:
```python
def test_should_handle_user_creation_within_200ms(client):
    # Arrange
    user_data = UserMother.valid_user()
    start_time = time.time()
    
    # Act
    response = client.post("/api/users", json=user_data)
    response_time = (time.time() - start_time) * 1000
    
    # Assert
    assert response.status_code == 201
    assert response_time < 200
```

**Go**:
```go
func TestUserAPI_Performance(t *testing.T) {
    // Arrange
    userData := UserMother.ValidUser()
    startTime := time.Now()
    
    // Act
    response := createUser(userData)
    responseTime := time.Since(startTime)
    
    // Assert
    assert.Equal(t, http.StatusCreated, response.StatusCode)
    assert.Less(t, responseTime.Milliseconds(), int64(200))
}
```

## Library-First TDD

### CLI Interface Testing
**Requirement**: All libraries must have CLI interfaces with comprehensive tests

**Test Structure**:
```javascript
describe('CLI Interface', () => {
  it('should return JSON output when --format=json flag is used', () => {
    // Test CLI commands and output formats
  });
  
  it('should handle invalid input with proper error codes', () => {
    // Test error handling and exit codes
  });
});
```

### Library Integration Testing
**Requirement**: Test library integration without mocks

**Test Structure**:
```javascript
describe('Library Integration', () => {
  it('should integrate with external services correctly', () => {
    // Test real integrations, no mocks
  });
});
```

## TDD Automation and Enforcement

### Git Hooks
**Pre-commit Hook**: Verify TDD cycle compliance
```bash
#!/bin/bash
# Enhanced TDD compliance validation

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get commit message
COMMIT_MSG=$(cat "$1")

# Function to validate commit message format
validate_commit_message() {
    if [[ ! "$COMMIT_MSG" =~ ^(RED|GREEN|REFACTOR):[[:space:]].+ ]]; then
        echo -e "${RED}❌ Error: Invalid commit message format${NC}"
        echo -e "${YELLOW}Required format: RED:|GREEN:|REFACTOR: <description>${NC}"
        echo -e "${YELLOW}Examples:${NC}"
        echo -e "  ${GREEN}RED: Add user validation test${NC}"
        echo -e "  ${GREEN}GREEN: Implement user validation logic${NC}"
        echo -e "  ${GREEN}REFACTOR: Extract validation to separate class${NC}"
        exit 1
    fi
}

# Function to validate TDD cycle progression
validate_tdd_cycle() {
    local commit_type=$(echo "$COMMIT_MSG" | grep -o '^[^:]*')
    local last_commit_type=$(git log --oneline -n 1 --pretty=format:'%s' | grep -o '^[^:]*' || echo "")
    
    case "$commit_type" in
        "RED")
            # RED can follow any state
            echo -e "${GREEN}✅ Starting TDD cycle with failing test${NC}"
            ;;
        "GREEN")
            if [[ "$last_commit_type" != "RED" ]]; then
                echo -e "${YELLOW}⚠️  Warning: GREEN commit should typically follow RED commit${NC}"
                echo -e "${YELLOW}Current: $commit_type, Previous: $last_commit_type${NC}"
                read -p "Continue anyway? (y/N): " -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                    exit 1
                fi
            fi
            ;;
        "REFACTOR")
            if [[ "$last_commit_type" != "GREEN" ]]; then
                echo -e "${YELLOW}⚠️  Warning: REFACTOR should typically follow GREEN commit${NC}"
                echo -e "${YELLOW}Current: $commit_type, Previous: $last_commit_type${NC}"
            fi
            ;;
    esac
}

# Function to run tests based on commit type
run_appropriate_tests() {
    local commit_type=$(echo "$COMMIT_MSG" | grep -o '^[^:]*')
    
    case "$commit_type" in
        "RED")
            echo -e "${YELLOW}🔴 Running tests - expecting failures...${NC}"
            run_tests_for_project 2>/dev/null
            if [ $? -eq 0 ]; then
                echo -e "${RED}❌ Error: RED commit should have failing tests${NC}"
                echo -e "${YELLOW}Hint: Make sure your new test fails before committing${NC}"
                exit 1
            fi
            echo -e "${GREEN}✅ Tests failing as expected for RED commit${NC}"
            ;;
        "GREEN"|"REFACTOR")
            echo -e "${YELLOW}🟢 Running tests - expecting all to pass...${NC}"
            run_tests_for_project
            if [ $? -ne 0 ]; then
                echo -e "${RED}❌ Error: All tests must pass for $commit_type commit${NC}"
                exit 1
            fi
            echo -e "${GREEN}✅ All tests passing${NC}"
            ;;
    esac
}

# Function to detect project type and run appropriate tests
run_tests_for_project() {
    if [ -f "package.json" ]; then
        # Node.js/JavaScript/TypeScript project
        npm test
    elif [ -f "pom.xml" ]; then
        # Java Maven project
        mvn test
    elif [ -f "build.gradle" ] || [ -f "build.gradle.kts" ]; then
        # Java Gradle project
        ./gradlew test
    elif [ -f "requirements.txt" ] || [ -f "pyproject.toml" ]; then
        # Python project
        python -m pytest
    elif [ -f "go.mod" ]; then
        # Go project
        go test ./...
    elif [ -f "Cargo.toml" ]; then
        # Rust project
        cargo test
    elif [ -f "*.csproj" ] || [ -f "*.sln" ]; then
        # C# .NET project
        dotnet test
    elif [ -f "Gemfile" ]; then
        # Ruby project
        bundle exec rspec
    else
        echo -e "${YELLOW}⚠️  Warning: Could not detect project type for testing${NC}"
        echo -e "${YELLOW}Please configure test command manually${NC}"
        exit 1
    fi
}

# Function to check test coverage for GREEN/REFACTOR commits
check_coverage() {
    local commit_type=$(echo "$COMMIT_MSG" | grep -o '^[^:]*')
    
    if [[ "$commit_type" == "GREEN" || "$commit_type" == "REFACTOR" ]]; then
        echo -e "${YELLOW}📊 Checking test coverage...${NC}"
        
        if command -v nyc &> /dev/null; then
            COVERAGE=$(nyc --reporter=text-summary npm test 2>/dev/null | grep "All files" | grep -o '[0-9.]*%' | head -1 | tr -d '%')
            if (( $(echo "$COVERAGE < 90" | bc -l) )); then
                echo -e "${YELLOW}⚠️  Warning: Test coverage is ${COVERAGE}% (target: 90%)${NC}"
            else
                echo -e "${GREEN}✅ Test coverage: ${COVERAGE}%${NC}"
            fi
        fi
    fi
}

# Main validation flow
echo -e "${YELLOW}🔍 Validating TDD compliance...${NC}"

validate_commit_message
validate_tdd_cycle
run_appropriate_tests
check_coverage

echo -e "${GREEN}✅ TDD validation passed!${NC}"
```

### CI/CD Integration
**Pipeline Validation**: Ensure TDD compliance in CI/CD for all project types
```yaml
name: TDD Validation
on: [push, pull_request]

jobs:
  validate-tdd:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      
      - name: Check commit message format
        run: |
          git log --oneline -n 1 | grep -E "^[a-f0-9]+ (RED|GREEN|REFACTOR):"
      
      - name: Setup Node.js (if Node project)
        if: hashFiles('package.json') != ''
        uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm'
      
      - name: Setup Java (if Java project)
        if: hashFiles('pom.xml', '**/*.gradle*') != ''
        uses: actions/setup-java@v3
        with:
          java-version: '17'
          distribution: 'temurin'
      
      - name: Setup Python (if Python project)
        if: hashFiles('requirements.txt', 'pyproject.toml') != ''
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      
      - name: Setup Go (if Go project)
        if: hashFiles('go.mod') != ''
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Setup Rust (if Rust project)
        if: hashFiles('Cargo.toml') != ''
        uses: actions-rs/toolchain@v1
        with:
          toolchain: stable
      
      - name: Setup .NET (if C# project)
        if: hashFiles('**/*.csproj', '**/*.sln') != ''
        uses: actions/setup-dotnet@v3
        with:
          dotnet-version: '7.0'
      
      - name: Install dependencies and run tests
        run: |
          if [ -f "package.json" ]; then
            npm ci && npm test
          elif [ -f "pom.xml" ]; then
            mvn clean test
          elif [ -f "build.gradle" ]; then
            ./gradlew test
          elif [ -f "requirements.txt" ]; then
            pip install -r requirements.txt && pytest
          elif [ -f "go.mod" ]; then
            go test ./...
          elif [ -f "Cargo.toml" ]; then
            cargo test
          elif [ -f "*.csproj" ]; then
            dotnet test
          fi
      
      - name: Check test coverage
        run: |
          if [ -f "package.json" ]; then
            npm run test:coverage
          elif [ -f "pom.xml" ]; then
            mvn jacoco:report
          elif [ -f "build.gradle" ]; then
            ./gradlew jacocoTestReport
          elif [ -f "requirements.txt" ]; then
            pytest --cov
          elif [ -f "go.mod" ]; then
            go test -coverprofile=coverage.out ./...
          elif [ -f "Cargo.toml" ]; then
            cargo tarpaulin
          elif [ -f "*.csproj" ]; then
            dotnet test --collect:"XPlat Code Coverage"
          fi
```

## TDD Metrics and Monitoring

### Quality Metrics
- **Test Coverage**: Minimum 90% line coverage required
- **Test Quality**: Mutation testing for test effectiveness
- **TDD Cycle Time**: Time between RED-GREEN-REFACTOR cycles
- **Test Maintenance**: Test code quality and maintainability

### Monitoring Dashboard
- Commit message compliance rates
- Test pass/fail rates over time
- Code coverage trends
- TDD cycle adherence metrics

## TDD Troubleshooting

### Test Doubles Strategy
**Principle**: Use real dependencies when possible, test doubles when necessary

#### When to Use Real Dependencies
- **Database Integration Tests**: Use real database with test data
- **HTTP Client Tests**: Use real HTTP calls to test services
- **File System Operations**: Use real files in temporary directories
- **Authentication Services**: Use real auth providers in test environment

#### When to Use Test Doubles
- **External APIs**: Use mocks for third-party services
- **Slow Operations**: Stub time-consuming processes
- **Error Scenarios**: Fake error conditions that are hard to reproduce
- **Isolated Unit Tests**: Mock dependencies to test single units

#### Types of Test Doubles

**Dummy Objects** - Objects passed but never used
```javascript
// Dummy object for parameter fulfillment
const dummyLogger = {};
const service = new UserService(database, dummyLogger);
```

**Fake Objects** - Working implementations with shortcuts
```javascript
// Fake in-memory database
class FakeUserRepository {
  constructor() {
    this.users = new Map();
  }
  
  async save(user) {
    this.users.set(user.id, user);
    return user;
  }
  
  async findById(id) {
    return this.users.get(id);
  }
}
```

**Stubs** - Provide canned answers to calls
```javascript
// Stub for external API
const emailServiceStub = {
  sendEmail: jest.fn().mockResolvedValue({ success: true, messageId: '123' })
};
```

**Spies** - Record information about calls
```javascript
// Spy to verify method calls
const loggerSpy = jest.spyOn(console, 'log');
service.processUser(user);
expect(loggerSpy).toHaveBeenCalledWith('Processing user:', user.id);
```

**Mocks** - Pre-programmed with expectations
```javascript
// Mock with expectations
const mockPaymentService = {
  processPayment: jest.fn()
    .mockResolvedValueOnce({ success: true, transactionId: 'tx123' })
    .mockRejectedValueOnce(new Error('Payment failed'))
};
```

#### Test Doubles Best Practices
1. **Verify Interactions**: Use spies/mocks to verify important interactions
2. **Realistic Fakes**: Make fakes behave like real implementations
3. **Avoid Over-Mocking**: Don't mock everything, focus on boundaries
4. **Mock Roles, Not Objects**: Mock at the interface level
5. **Reset Between Tests**: Clear mock state between test cases

### Common Anti-Patterns
1. **Writing Tests After Code**: Violates TDD principles
2. **Testing Implementation Details**: Tests should focus on behavior
3. **Over-Mocking**: Use real dependencies where possible
4. **Large Test Cases**: Keep tests small and focused
5. **Mocking Value Objects**: Mock behavior, not data structures

### Solutions
1. **Enforce RED Phase**: Verify tests fail before implementation
2. **Behavior-Driven Tests**: Focus on user-facing behavior
3. **Integration Testing**: Test with real dependencies
4. **Test Decomposition**: Break large tests into smaller ones

## Language-Specific TDD Considerations

### JavaScript/TypeScript
**Frameworks**: Jest, Mocha, Vitest
**Considerations**:
- Use `async/await` for asynchronous tests
- Mock modules with `jest.mock()` or similar
- Test both TypeScript types and runtime behavior
- Use test utilities like Testing Library for DOM testing

### Java
**Frameworks**: JUnit 5, TestNG, Mockito
**Considerations**:
- Use `@DisplayName` for readable test descriptions
- Leverage `@ParameterizedTest` for data-driven tests
- Use TestContainers for integration tests with real dependencies
- Follow AAA pattern consistently

### Python
**Frameworks**: pytest, unittest, mock
**Considerations**:
- Use fixtures for test setup and teardown
- Leverage `pytest.mark.parametrize` for multiple test cases
- Use `unittest.mock` for mocking dependencies
- Follow PEP 8 naming conventions for test methods

### Go
**Frameworks**: Go testing, Testify, GoMock
**Considerations**:
- Use table-driven tests for multiple scenarios
- Leverage interfaces for dependency injection and mocking
- Use `t.Helper()` for test helper functions
- Follow Go naming conventions for test functions

### Rust
**Frameworks**: Built-in testing, proptest
**Considerations**:
- Use `#[cfg(test)]` module for unit tests
- Leverage `Result<T, E>` for error handling in tests
- Use property-based testing with proptest
- Test both success and error paths explicitly

### C#
**Frameworks**: xUnit, NUnit, MSTest, Moq
**Considerations**:
- Use `[Theory]` and `[InlineData]` for parameterized tests
- Leverage async/await patterns in tests
- Use dependency injection for testable design
- Follow naming conventions with descriptive test method names

### Ruby
**Frameworks**: RSpec, Minitest
**Considerations**:
- Use RSpec's descriptive syntax (describe, context, it)
- Leverage FactoryBot for test data generation
- Use VCR for HTTP interaction testing
- Follow Ruby naming conventions

## TDD Training and Onboarding

### Required Skills
- Understanding of testing frameworks
- Knowledge of test design patterns
- Familiarity with TDD cycle and principles
- Experience with continuous integration

### Training Resources
- TDD workshops and hands-on exercises
- Code review sessions focused on TDD practices
- Pair programming with TDD experts
- Regular TDD retrospectives and improvements