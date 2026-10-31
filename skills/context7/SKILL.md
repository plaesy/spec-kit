---
name: context7
description: "Context7 documentation lookup: when it is mandatory, the exact tool sequence, the citation form, and the fallback. Use when the user asks about a library, framework, SDK, or API, or wants current docs or examples."
metadata:
  applies-to: "**/*"
---

# Context7 Protocol

| File | Load when |
|------|-----------|
| `SKILL.md` (this file) | Looking up library documentation through Context7, and the fallback when it is unavailable |
| `references/date.md` | Citing a retrieval date or validating a time-based claim |

Context7 resolves a library ID to current, version-matched documentation. It is
the authority for "what does this API actually look like in the version we are
installing" — the question general model knowledge answers confidently and
sometimes wrongly.

This file is the only place the protocol is stated. Commands that require it
(`/implement`, `/fix`) point here rather than restating the sequence, because a
restated sequence drifts: one command ends up citing, another does not, and a
reader cannot tell which rule is current.

## When It Is Mandatory

| Situation | Required |
|-----------|----------|
| Adding or upgrading any third-party dependency | yes |
| Using an API whose signature may have changed since training | yes |
| Answering "is this library deprecated / what is the current version" | yes |
| Editing only the project's own code, no new dependency | no |
| Reading a framework config file whose format is not in doubt | no |
| The library is in the "Always Available" table below | no — cite the primary source directly |

"Not mandatory" means *not required*, not *forbidden*. When a lookup is cheap and
the answer will be acted on, do it.

## The Sequence

1. `mcp__context7__resolve-library-id` for **every** technology involved — one
   call per library, not per project. A stack of three technologies is three
   resolve calls.
2. `mcp__context7__get-library-docs` (tokens=500) against each resolved ID.
3. Cite the result inline, in the form below.
4. On failure, follow the fallback chain. Never silently proceed as if the
   lookup had succeeded.

## Required Citation Form

```text
Implementation based on Context7 (/library/id) — Retrieved {{CURRENT_DATE}}
```

The citation carries three obligations: the resolved library ID (so a reader can
re-resolve it), the retrieval date (so a stale answer is visible), and the word
Context7 (so it is distinguishable from a general-knowledge claim). A finding,
decision, or implementation line that used Context7 and carries no citation is
unverifiable and should be treated as uncited.

## Fallback Chain

Context7 being unavailable degrades the *sourcing*, never the requirement to
source. Work down this list and record which rung you landed on:

1. **Local cache** — `~/.context7/cache`, if populated. Say that you used a
   cached copy and give its date; a cache entry from six months ago is not a
   current answer and must not be presented as one.
2. **Primary documentation** — the library's own docs, or its repository. Cite
   the URL. This is the preferred fallback: it is the same authority Context7
   indexes. The tables in "Fallback Source Tables" below give the URL for each
   common technology.
3. **General knowledge** — acceptable only for stable, unversioned facts, and
   it must be labelled as general knowledge rather than presented as verified.
   Anything version-dependent stops here: if the answer depends on which version
   is current, you do not know it, and the honest output is the question, not a
   guess.
4. **Apply the most conservative documented default and label it** — when the
   answer changes what gets built but no source above resolved it, pick the
   behavior most consistent with the library's last known-stable release,
   label it `ASSUMED — <default>, unverifiable at Context7 fallback rung 4`,
   and record it in `.plaesy/context.md` so the next session (or an
   `ASSUMED`-tag audit) can revisit it. Nara decides how to proceed (`.plaesy/instructions/plaesy.md`
   rule 8), e.g. by choosing the reversible option for a destructive/irreversible operation, and
   records the choice.

When using a fallback source, cite the URL, note the date retrieved, and
cross-reference when unsure. Flag documentation older than two years, and note
deprecation when using a deprecated version. Do not use Stack Overflow as the
primary source, and do not treat community answers as authoritative. Source
priority when several are available: official docs > GitHub repo > community.

## Fallback Source Tables

Used at rung 2 above. Each row maps a technology to its maintained source.

### JavaScript/TypeScript

| Tech | Fallback Source |
|------|-----------------|
| **React** | `https://react.dev` |
| **Next.js** | `https://nextjs.org` |
| **Vue.js** | `https://vuejs.org` |
| **Angular** | `https://angular.io` |
| **Express.js** | `https://expressjs.com` |
| **NestJS** | `https://nestjs.com` |
| **Fastify** | `https://fastify.io` |
| **Jest** | `https://jestjs.io` |
| **Vitest** | `https://vitest.dev` |
| **Cypress** | `https://cypress.io` |
| **Playwright** | `https://playwright.dev` |

### Python

| Tech | Fallback Source |
|------|-----------------|
| **Django** | `https://www.djangoproject.com` |
| **FastAPI** | `https://fastapi.tiangolo.com` |
| **Flask** | `https://flask.palletsprojects.com` |
| **NumPy** | `https://numpy.org` |
| **Pandas** | `https://pandas.pydata.org` |
| **Scikit-learn** | `https://scikit-learn.org` |

### Java

| Tech | Fallback Source |
|------|-----------------|
| **Spring Boot** | `https://spring.io/projects/spring-boot` |
| **Quarkus** | `https://quarkus.io` |
| **JUnit 5** | `https://junit.org/junit5` |
| **Maven** | `https://maven.apache.org` |

### Go

| Tech | Fallback Source |
|------|-----------------|
| **Gin** | `https://gin-gonic.com` |
| **Echo** | `https://echo.labstack.com` |
| **Testing** | `https://pkg.go.dev/testing` |

### Rust

| Tech | Fallback Source |
|------|-----------------|
| **Actix-web** | `https://actix-rs.github.io` |
| **Axum** | `https://tokio.rs` |
| **Rocket** | `https://rocket.rs` |

### Databases

| Tech | Fallback Source |
|------|-----------------|
| **PostgreSQL** | `https://www.postgresql.org` |
| **MySQL** | `https://www.mysql.com` |
| **SQLite** | `https://www.sqlite.org` |
| **MongoDB** | `https://www.mongodb.com` |
| **Redis** | `https://redis.io` |

### DevOps & Infrastructure

| Tech | Fallback Source |
|------|-----------------|
| **Docker** | `https://docs.docker.com` |
| **Kubernetes** | `https://kubernetes.io` |
| **Terraform** | `https://www.terraform.io` |
| **GitHub Actions** | `https://github.com/features/actions` |

## What Context7 Is Not For

- **Not a licence to skip reading the code.** Context7 documents the library as
  published; your installed version may be patched, vendored, or wrapped. Read
  what you actually depend on.
- **Not a substitute for the project's own instructions.** Stack conventions,
  testing strategy, and the constitution's thresholds live in
  `.plaesy/memory/constitution.md` and the `*-design-principles` files, and
  Context7 says nothing about them.
- **Not applicable to private or internal packages.** There is no public ID to
  resolve; read the package's own source and types instead.

## Always Available (No Lookup Needed)

| Standard | Source |
|----------|--------|
| **OWASP Top 10** | `https://owasp.org/www-project-top-ten/` |
| **WCAG 2.2** | `https://www.w3.org/WAI/WCAG22/quickref/` |
| **REST API Best Practices** | `https://restfulapi.net` |
| **GraphQL Spec** | `https://spec.graphql.org` |
| **Semantic Versioning** | `https://semver.org` |
| **OpenAPI 3.0** | `https://spec.openapis.org/oas/v3.0.3` |

These are specifications, not libraries: they have no version to resolve and no
release cycle, so a Context7 call would return nothing. Cite them directly.
