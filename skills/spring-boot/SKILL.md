---
name: spring-boot
description: "Spring Boot guidelines: dependency injection, configuration, code organization, service layer, logging, and input handling. Use when building or reviewing Spring Boot services."
metadata:
  applies-to: "**/*.java, **/*.kt"
---

# Spring Boot Development

Apply these when the project uses Spring Boot. Base Java rules are in `../java/SKILL.md`.

## General

- Only high-confidence suggestions when reviewing code changes
- Maintainable code with comments on why (not just what)
- Handle edge cases, clear exception handling
- Comment usage/purpose of libraries and external dependencies

## Dependency Injection

- Constructor injection for all required dependencies
- Dependency fields declared `private final`

## Configuration

- YAML (`application.yml`) for externalized config
- Spring profiles for environments (dev, test, prod)
- `@ConfigurationProperties` for type-safe config binding
- Externalize secrets via env vars or a secrets manager

## Code Organization

- Package by feature/domain, not by layer
- Thin controllers, focused services, simple repositories
- Utility classes: `final` with private constructors

## Service Layer

- Business logic in `@Service`-annotated classes
- Stateless, testable services
- Repositories injected via constructor
- Method signatures use domain IDs or DTOs, not repository entities directly unless necessary

## Logging

- SLF4J for all logging (`private static final Logger logger = LoggerFactory.getLogger(MyClass.class);`)
- No direct Logback/Log4j2 or `System.out.println()`
- Parameterized logging: `logger.info("User {} logged in", userId);`

## Security & Input Handling

- Parameterized queries: Spring Data JPA or `NamedParameterJdbcTemplate` to prevent SQL injection
- Validate request bodies/params via JSR-380 (`@NotNull`, `@Size`, etc.) and `BindingResult`

## Usage Example

```java

@Service
public class UserService {
    private final UserRepository repository;

    public UserService(UserRepository repository) {
        this.repository = repository;
    }

    public User findById(Long id) {
        return repository.findById(id)
            .orElseThrow(() -> new UserNotFoundException(id));
    }
}
```
