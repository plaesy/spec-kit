---
description: "Enforces Test-Driven Development (TDD) practices and commit patterns"
applyTo: '**/*.js, **/*.ts, **/*.java, **/*.py, **/*.rb, **/*.go, **/*.rs, **/*.cs, **/*.cpp, **/*.c, **/*.swift, **/*.kt, **/*.kts'
---

# Test-Driven Development (TDD) Instructions

## Overview
Constitutional requirement: Test-First Development, non-negotiable, applies to all implementation work.

## TDD Cycle: Red-Green-Refactor

### Phase 1: RED - Write Failing Test
Write a test that fails because the functionality doesn't exist yet.
1. Review spec/plan for the specific functionality
2. Write test describing expected behavior
3. Run it, verify it fails (RED)
4. Commit: "RED: [test description]"

**Example Commit Messages**:
```
RED: Add user authentication validation test
RED: Implement API endpoint response validation
RED: Add database connection error handling test
```

### Phase 2: GREEN - Make Test Pass
Write minimal code to make the test pass.
1. Implement just enough to pass
2. Run it, verify it passes (GREEN)
3. Run all tests, check no regression
4. Commit: "GREEN: [implementation description]"

**Example Commit Messages**:
```
GREEN: Implement user authentication validation
GREEN: Add API endpoint with proper response format
GREEN: Handle database connection errors gracefully
```

### Phase 3: REFACTOR - Improve Code Quality
Improve structure, keep behavior unchanged.
1. Identify code smells/duplication/complexity
2. Refactor without changing behavior
3. Run all tests, confirm still passing
4. Commit: "REFACTOR: [improvement description]"

**Example Commit Messages**:
```
REFACTOR: Extract authentication logic to separate module
REFACTOR: Simplify API response handling
REFACTOR: Improve database error handling structure
```

## TDD Best Practices

### Test Design Principles
1. Arrange-Act-Assert structure
2. One assertion per test - single behavior
3. Descriptive names explaining expected behavior
4. Cover boundary conditions and error scenarios

### BDD Integration with TDD
Combine TDD with BDD for requirement clarity: Given-When-Then acceptance criteria, outside-in (acceptance -> unit tests), tests as living documentation.

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
1. **YAGNI**: implement only what tests require
2. **DRY**: refactor duplication
3. **SOLID**: follow during refactoring
4. **Clean Code**: readable, maintainable

### Test Organization Patterns

#### Test Builders Pattern
Create complex test objects with a readable, chainable API.

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
Pre-configured objects for common test scenarios.
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
Generate test data with proper relationships/constraints.
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
Scope: individual functions/methods/classes. Tools: Jest/Mocha (JS/TS), JUnit (Java), PyTest (Python), Go Test, Cargo Test (Rust), NUnit (C#). Focus: business logic, data transformation, validation.

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
Scope: component interactions, API endpoints, database ops. Tools: Supertest/Cypress (JS/TS), TestContainers (Java), Requests (Python), Testify (Go), Reqwest (Rust). Focus: integration, data flow, external deps.

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
Scope: complete user workflows, system behavior. Tools: Playwright, Cypress, Selenium. Focus: user journeys, system integration, business workflows.

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
Scope: non-functional requirements, response times, throughput. Tools: Artillery/K6 (JS), JMeter (Java), Locust (Python), Vegeta (Go), Criterion (Rust). Focus: benchmarks, load testing, optimization validation.

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
All libraries must have CLI interfaces with comprehensive tests.

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
Test library integration without mocks.

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
Pre-commit hook verifies TDD cycle compliance.
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
Pipeline validates TDD compliance in CI/CD for all project types.
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
- Test coverage: min 90% line coverage
- Test quality: mutation testing for effectiveness
- TDD cycle time: RED-GREEN-REFACTOR duration
- Test maintenance: test code quality/maintainability

### Monitoring Dashboard
Commit message compliance, pass/fail rates over time, coverage trends, TDD cycle adherence.

## TDD Troubleshooting

### Test Doubles Strategy
Prefer real dependencies; use test doubles only when necessary.

**Use real dependencies for**: database integration tests (real DB + test data), HTTP client tests (real calls to test services), file system ops (real files in temp dirs), auth services (real providers in test env).

**Use test doubles for**: external APIs (mock third-party services), slow operations (stub time-consuming processes), error scenarios hard to reproduce (fake), isolated unit tests (mock dependencies).

#### Types of Test Doubles

**Dummy Objects** - passed but never used
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
1. Verify interactions via spies/mocks
2. Realistic fakes - behave like real implementations
3. Avoid over-mocking - focus on boundaries
4. Mock roles (interfaces), not objects
5. Reset mock state between tests

### Common Anti-Patterns -> Solutions
- Writing tests after code (violates TDD) -> enforce RED phase, verify tests fail first
- Testing implementation details -> behavior-driven tests, focus on user-facing behavior
- Over-mocking -> integration testing with real dependencies
- Large test cases -> test decomposition, break into smaller ones
- Mocking value objects -> mock behavior, not data structures

## Language-Specific TDD Considerations

| Language | Frameworks | Key considerations |
|---|---|---|
| JS/TS | Jest, Mocha, Vitest | `async/await` for async tests; `jest.mock()`; test types + runtime behavior; Testing Library for DOM |
| Java | JUnit 5, TestNG, Mockito | `@DisplayName`; `@ParameterizedTest`; TestContainers for integration; consistent AAA |
| Python | pytest, unittest, mock | fixtures for setup/teardown; `pytest.mark.parametrize`; `unittest.mock`; PEP 8 naming |
| Go | Go testing, Testify, GoMock | table-driven tests; interfaces for DI/mocking; `t.Helper()`; Go naming conventions |
| Rust | built-in testing, proptest | `#[cfg(test)]`; `Result<T, E>` in tests; property-based testing; test success + error paths |
| C# | xUnit, NUnit, MSTest, Moq | `[Theory]`/`[InlineData]`; async/await; DI for testability; descriptive method names |
| Ruby | RSpec, Minitest | RSpec describe/context/it; FactoryBot; VCR for HTTP; Ruby naming conventions |

## TDD Training and Onboarding
**Required skills**: testing frameworks, test design patterns, TDD cycle/principles, CI experience.
**Resources**: TDD workshops, code review sessions, pair programming with TDD experts, regular retrospectives.