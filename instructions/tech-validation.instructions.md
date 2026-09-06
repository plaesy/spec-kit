---
description: "Technology validation and fallback sources when Context7 unavailable"
applyTo: "package.json, go.mod, pom.xml, Cargo.toml, pubspec.yaml, pyproject.toml, *.csproj"
---

# Technology Validation & Fallback Sources

## Overview
When Context7 MCP is unavailable, this protocol provides verified fallback sources for all major technologies.

---

## JavaScript/TypeScript Ecosystem

### Frontend Frameworks
| Tech | Fallback Source |
|------|-----------------|
| **React** | [react.dev](https://react.dev) |
| **Next.js** | [nextjs.org](https://nextjs.org) |
| **Vue.js** | [vuejs.org](https://vuejs.org) |
| **Angular** | [angular.io](https://angular.io) |

### Backend Frameworks
| Tech | Fallback Source |
|------|-----------------|
| **Express.js** | [expressjs.com](https://expressjs.com) |
| **NestJS** | [nestjs.com](https://nestjs.com) |
| **Fastify** | [fastify.io](https://fastify.io) |

### Testing
| Tech | Fallback Source |
|------|-----------------|
| **Jest** | [jestjs.io](https://jestjs.io) |
| **Vitest** | [vitest.dev](https://vitest.dev) |
| **Cypress** | [cypress.io](https://cypress.io) |
| **Playwright** | [playwright.dev](https://playwright.dev) |

---

## Python Ecosystem

### Web Frameworks
| Tech | Fallback Source |
|------|-----------------|
| **Django** | [djangoproject.com](https://www.djangoproject.com) |
| **FastAPI** | [fastapi.tiangolo.com](https://fastapi.tiangolo.com) |
| **Flask** | [flask.palletsprojects.com](https://flask.palletsprojects.com) |

### Data Science
| Tech | Fallback Source |
|------|-----------------|
| **NumPy** | [numpy.org](https://numpy.org) |
| **Pandas** | [pandas.pydata.org](https://pandas.pydata.org) |
| **Scikit-learn** | [scikit-learn.org](https://scikit-learn.org) |

---

## Java Ecosystem

| Tech | Fallback Source |
|------|-----------------|
| **Spring Boot** | [spring.io/projects/spring-boot](https://spring.io/projects/spring-boot) |
| **Quarkus** | [quarkus.io](https://quarkus.io) |
| **JUnit 5** | [junit.org/junit5](https://junit.org/junit5) |
| **Maven** | [maven.apache.org](https://maven.apache.org) |

---

## Go Ecosystem

| Tech | Fallback Source |
|------|-----------------|
| **Gin** | [gin-gonic.com](https://gin-gonic.com) |
| **Echo** | [echo.labstack.com](https://echo.labstack.com) |
| **Testing** | [pkg.go.dev/testing](https://pkg.go.dev/testing) |

---

## Rust Ecosystem

| Tech | Fallback Source |
|------|-----------------|
| **Actix-web** | [actix-rs.github.io](https://actix-rs.github.io) |
| **Axum** | [tokio.rs](https://tokio.rs) |
| **Rocket** | [rocket.rs](https://rocket.rs) |

---

## Databases

### Relational
| Tech | Fallback Source |
|------|-----------------|
| **PostgreSQL** | [postgresql.org](https://www.postgresql.org) |
| **MySQL** | [mysql.com](https://www.mysql.com) |
| **SQLite** | [sqlite.org](https://www.sqlite.org) |

### NoSQL
| Tech | Fallback Source |
|------|-----------------|
| **MongoDB** | [mongodb.com](https://www.mongodb.com) |
| **Redis** | [redis.io](https://redis.io) |

---

## DevOps & Infrastructure

| Tech | Fallback Source |
|------|-----------------|
| **Docker** | [docker.com](https://docs.docker.com) |
| **Kubernetes** | [kubernetes.io](https://kubernetes.io) |
| **Terraform** | [terraform.io](https://www.terraform.io) |
| **GitHub Actions** | [github.com/features/actions](https://github.com/features/actions) |

---

## Always Available (No Context7 Needed)

| Standard | Source |
|----------|--------|
| **OWASP Top 10** | [owasp.org](https://owasp.org/www-project-top-ten/) |
| **WCAG 2.1** | [w3.org/WAI/WCAG21](https://www.w3.org/WAI/WCAG21/quickref/) |
| **REST API Best Practices** | [restfulapi.net](https://restfulapi.net) |
| **GraphQL Spec** | [spec.graphql.org](https://spec.graphql.org) |
| **Semantic Versioning** | [semver.org](https://semver.org) |
| **OpenAPI 3.0** | [spec.openapis.org/oas/v3.0.3](https://spec.openapis.org/oas/v3.0.3) |

---

## Fallback Decision Tree

```
When Context7 unavailable:

1. Technology identified?
   ├─ YES → Check table above
   └─ NO → Ask user for technology name

2. Official documentation available?
   ├─ YES → Use official docs
   └─ NO → Use GitHub repo or community

3. Multiple sources?
   ├─ YES → Prefer: Official > GitHub > Community
   └─ NO → Use only available source
```

---

## Validation Rules

When using fallback sources:

```
✅ DO:
- Cite source with URL
- Note date retrieved
- Cross-reference if unsure
- Flag if documentation >2 years old

❌ DON'T:
- Use Stack Overflow as primary source
- Assume community answers are authoritative
- Use deprecated versions without noting deprecation
```

---

*Technology Validation & Fallback Sources v1.0.0*  
*Framework: Spec Kit Constitutional Development Framework*
