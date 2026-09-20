---
description: 'Guidelines for building C# applications'
applyTo: '**/*.cs'
---

# C# Development

## C# Instructions
- Always use the latest C# version (currently C# 13 features)
- Clear, concise comments per function

## General Instructions
- Only high-confidence suggestions when reviewing code changes
- Maintainable code, with comments on why (not just what) for design decisions
- Handle edge cases, clear exception handling
- Comment the usage/purpose of libraries and external dependencies

## Naming Conventions
- PascalCase for component names, method names, public members
- camelCase for private fields and local variables
- Prefix interfaces with "I" (e.g. `IUserService`)

## Formatting
- Follow `.editorconfig`
- File-scoped namespace declarations, single-line using directives
- Newline before the opening curly brace of any block (`if`, `for`, `while`, `foreach`, `using`, `try`, etc.)
- Final `return` on its own line
- Pattern matching and switch expressions wherever possible
- `nameof` instead of string literals for member names
- XML doc comments on public APIs, with `<example>`/`<code>` where applicable

## Project Setup and Structure
Guide new .NET project creation with appropriate templates; explain each generated file/folder; organize via feature folders or DDD; separate concerns (models, services, data access); explain Program.cs and ASP.NET Core 9's environment-specific configuration.

## Nullable Reference Types
- Declare variables non-nullable, check for `null` at entry points
- `is null`/`is not null`, never `== null`/`!= null`
- Trust null annotations - don't add redundant null checks the type system already guarantees

## Data Access Patterns
EF Core data access layer; SQL Server/SQLite/In-Memory options for dev vs prod; repository pattern (and when it's worth it); migrations and data seeding; efficient query patterns to avoid common perf issues.

## Authentication and Authorization
JWT Bearer token auth; OAuth 2.0/OpenID Connect in ASP.NET Core; role-based and policy-based authorization; Microsoft Entra ID (Azure AD) integration; consistent security across controller-based and Minimal APIs.

## Validation and Error Handling
Model validation via data annotations and FluentValidation; validation pipeline customization; global exception handling middleware; consistent error responses; RFC 7807 problem details.

## API Versioning and Documentation
Versioning strategies; Swagger/OpenAPI with proper docs (endpoints, parameters, responses, auth); versioning for both controller-based and Minimal APIs; documentation that actually helps API consumers.

## Logging and Monitoring
Structured logging (Serilog or similar); logging levels and when to use each; Application Insights telemetry; custom telemetry and correlation IDs for request tracking; monitoring performance/errors/usage.

## Testing
- Always include tests for critical paths
- Guide unit test creation
- No "Act"/"Arrange"/"Assert" comments
- Match existing style in nearby files for test naming/capitalization
- Integration testing for API endpoints, mocking dependencies, testing auth logic, TDD principles for API development

## Performance Optimization
Caching strategies (in-memory, distributed, response caching); async patterns and why they matter for API perf; pagination/filtering/sorting for large datasets; compression and other optimizations; measuring/benchmarking API performance.

## Deployment and DevOps
Containerize via .NET's built-in container support (`dotnet publish --os linux --arch x64 -p:PublishProfile=DefaultContainer`) vs manual Dockerfile; CI/CD pipelines; deployment to Azure App Service/Container Apps/other hosts; health checks and readiness probes; environment-specific configuration per deployment stage.
