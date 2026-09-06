---
description: 'Guidelines for building Java base applications'
applyTo: '**/*.java'
---

# Java Development

## General Instructions

- Ask if the user wants static analysis tools (SonarQube, PMD, Checkstyle) integrated; if yes, guide selection/config
- If declined, still apply the best-practices/bug-pattern/code-smell guidelines below
- Address code smells proactively, don't accumulate technical debt
- Prioritize readability, maintainability, performance when refactoring
- Use IDE-reported warnings to catch patterns early

## Best practices

- **Records**: use Java Records instead of traditional classes for data-holding classes (DTOs, immutable structures)
- **Pattern Matching**: use pattern matching for `instanceof`/`switch` to simplify conditional logic and type casting
- **Type Inference**: use `var` for local variables when the type is explicitly clear from the right-hand side
- **Immutability**: favor immutable objects; `final` classes/fields where possible; `List.of()`/`Map.of()` for fixed data; `Stream.toList()` for immutable lists
- **Streams and Lambdas**: Streams API + lambdas for collection processing; method references (e.g. `stream.map(Foo::toBar)`)
- **Null Handling**: avoid returning/accepting `null` - `Optional<T>` for possibly-absent values, `Objects.equals()`/`requireNonNull()`

### Naming Conventions

Follow Google's Java style guide: `UpperCamelCase` for classes/interfaces, `lowerCamelCase` for methods/variables, `UPPER_SNAKE_CASE` for constants, `lowercase` for packages. Nouns for classes (`UserService`), verbs for methods (`getUserById`). Avoid abbreviations and Hungarian notation.

### Bug Patterns

| Rule ID | Description                                                 | Example / Notes                                                                                  |
| ------- | ----------------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| `S2095` | Resources should be closed                                  | Use try-with-resources when working with streams, files, sockets, etc.                           |
| `S1698` | Objects should be compared with `.equals()` instead of `==` | Especially important for Strings and boxed primitives.                                           |
| `S1905` | Redundant casts should be removed                           | Clean up unnecessary or unsafe casts.                                                            |
| `S3518` | Conditions should not always evaluate to true or false      | Watch for infinite loops or if-conditions that never change.                                     |
| `S108`  | Unreachable code should be removed                          | Code after `return`, `throw`, etc., must be cleaned up.                                          |

## Code Smells

| Rule ID | Description                                            | Example / Notes                                                               |
| ------- | ------------------------------------------------------ | ------------------------------------------------------------------------------- |
| `S107`  | Methods should not have too many parameters            | Refactor into helper classes or use builder pattern.                          |
| `S121`  | Duplicated blocks of code should be removed            | Consolidate logic into shared methods.                                        |
| `S138`  | Methods should not be too long                         | Break complex logic into smaller, testable units.                             |
| `S3776` | Cognitive complexity should be reduced                 | Simplify nested logic, extract methods, avoid deep `if` trees.                |
| `S1192` | String literals should not be duplicated               | Replace with constants or enums.                                              |
| `S1854` | Unused assignments should be removed                   | Avoid dead variables—remove or refactor.                                      |
| `S109`  | Magic numbers should be replaced with constants        | Improves readability and maintainability.                                     |
| `S1188` | Catch blocks should not be empty                       | Always log or handle exceptions meaningfully.                                 |

## Build and Verification

After changes, verify the build: `mvn clean install` (Maven) or `./gradlew build` / `gradlew.bat build` (Gradle, Windows). Ensure all tests pass.