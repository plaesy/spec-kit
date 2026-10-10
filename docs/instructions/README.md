# 📖 Plaesy Instructions Documentation

**Comprehensive instruction library for technology-specific development guidance.**

⚠️ **CRITICAL: Before creating/editing instructions, understand:**

- How instructions are copied to `.plaesy/instructions/`
- Correct reference paths (use `.plaesy/instructions/` not `instructions/`)
- No self-references rule
- Cross-reference patterns

(See "📍 Architecture: Spec-Kit vs Per-Project" below.)

## 🎯 Overview

The Plaesy instructions system provides detailed, technology-specific guidance for development teams. Each instruction document contains best practices, coding standards, security guidelines, and
optimization techniques tailored to specific technologies and platforms.

## 📍 Architecture: Spec-Kit vs Per-Project

### Spec-Kit Repository (Shared Reference)

- **Location**: `instructions/` folder in spec-kit repo
- **Files**: `*.instructions.md` (comprehensive library)
- **Purpose**: Shared reference templates, mapped via `instructions/mapping.json`
- **Use**: Referenced by CLI and AI assistants for keyword-based detection

### Per-Project (Self-Contained)

- **Location**: `.plaesy/instructions/` in each project
- **Files**: Auto-copied during `plaesy init` — the 33 framework/scope files listed
  in `.plaesy/instructions.md` (see "What Actually Gets Installed" below)
- **Purpose**: Project-specific instructions (only the framework + per-dimension set
  is copied; technology files are added only when the stack is detected)
- **Benefits**: Self-contained (CLAUDE.md compliant), lightweight, no duplication

**How it works**: When `plaesy init` runs, it:

1. Reads `instructions/mapping.json` — the single source of truth for the install
2. Copies every `mappings.always_load` file to `.plaesy/instructions/` (renamed
   `name.instructions.md` → `name.md`)
3. Copies every `mappings.scope_load` file to the same directory (renamed the same
   way). These are **available on demand** for `/assess:{dimension}`, not always
   loaded into the agent's active context
4. Stops there. Technology/framework files (`go`, `reactjs`, `python`, `vue`, …) are
   **not** copied by `init`; they are only *reported* by `plaesy stack detect`, which
   matches the project's manifests, extensions and config filenames against the
   `frameworks` / `languages` / `cross_cutting` / `methodologies` categories

The implementation of steps 1–3 is `scripts/internal/scaffold/copy.go:loadMappingList`
— it reads exactly these two lists and nothing else.

---

## 📦 What Actually Gets Installed (33 of 73 — read this before Debugging "My Instruction Is Missing")

**`plaesy init` installs 33 of the 73 `*.instructions.md` files in
`instructions/`.** The installed set is exactly
`mapping.json → mappings.always_load ∪ mappings.scope_load`, and nothing else. The
generated index in `.plaesy/instructions.md` lists all 33 with a one-line
description each, and every path in it exists after `init`.

| Group | Count | What it is |
|-------|-------|------------|
| `always_load` | 8 | Loaded into context for every project: `plaesy`, `context-engineering`, `tasks`, `quality-gates`, `error-recovery`, `dimension-mapping`, `output-validation`, `universal-orchestrator` |
| `scope_load` | 25 | Copied but read on demand: the 9 `assess-*` per-dimension workflows, the 3 protocol references (`context7-protocol`, `uncertainty-surfacing`, `date-system`) + 13 cross-cutting `*-design-principles` files |

### The consequence, stated plainly

A scoped command in an installed project **cannot load the other 40 files**. If a
Go, React, Python, Terraform or security-audit instruction is not in
`.plaesy/instructions/`, no prompt in that project can read it — stack detection
must fire first. `plaesy stack detect` reports the matching filenames to stdout
and stops; **`plaesy stack detect --install` is the opt-in copy** — it copies
each detected instruction into `.plaesy/instructions/`, stripping the
`.instructions` suffix, and never overwrites an existing destination.

This is a deliberate, lightweight-by-design trade-off (a lean base context per
project), but it has three practical consequences:

1. **Technology guidance is opt-in.** A Python or Vue project is not automatically
   given `python.instructions.md` / `vue.instructions.md`; they appear in
   `stack detect` output and must be copied in (`stack detect --install`).
2. **Editing a non-installed file has no effect on a project.** Fixes to
   `go.instructions.md` reach a project only after that file is copied in.
3. **`instructions/mapping.json` categories are a detection registry, not an install
   list.** Adding a file to `frameworks`/`languages`/`cross_cutting`/`methodologies`
   makes it *detectable*; adding it to `always_load`/`scope_load` makes it
   *installed*. These are different outcomes and are easy to confuse.

`instructions/agents.instructions.md` is deliberately in neither group — it is the
per-platform core template (copied as `CLAUDE.md`/`AGENTS.md`/`.cursor/rules/…` by
the platform setup step, see the `excluded_files_note` in `mapping.json`).

## 📚 Instruction Categories

### 🌐 Frontend Framework Instructions

| Instruction | Technology | Focus Areas | Integration |
|-------------|------------|-------------|-------------|
| **[reactjs.instructions.md](../../instructions/reactjs.instructions.md)** | React.js | Component architecture, hooks, state management | Modern React patterns |
| **[vue.instructions.md](../../instructions/vue.instructions.md)** | Vue 3 | SFC structure, Composition API, reactivity, props/events | Progressive + full SPA |
| **[nextjs.instructions.md](../../instructions/nextjs.instructions.md)** | Next.js | SSR/SSG, routing, performance optimization | Full-stack React |
| **[angular.instructions.md](../../instructions/angular.instructions.md)** | Angular | Components, services, dependency injection | Enterprise Angular |
| **[react-native.instructions.md](../../instructions/react-native.instructions.md)** | React Native | Mobile development, native modules | Cross-platform mobile |
| **[dart-n-flutter.instructions.md](../../instructions/dart-n-flutter.instructions.md)** | Dart/Flutter | Mobile UI, state management | Cross-platform |

### ⚙️ Backend Framework Instructions

| Instruction | Technology | Focus Areas | Architecture |
|-------------|------------|-------------|--------------|
| **[springboot.instructions.md](../../instructions/springboot.instructions.md)** | Spring Boot | Microservices, Spring ecosystem | Enterprise Java |
| **[nestjs.instructions.md](../../instructions/nestjs.instructions.md)** | NestJS | Modular architecture, TypeScript | Node.js enterprise |
| **[ruby-on-rails.instructions.md](../../instructions/ruby-on-rails.instructions.md)** | Rails | Convention over configuration | Rapid development |

### 📊 Microsoft 365 Automation Instructions

| Instruction | Technology | Focus Areas | Integration |
|-------------|------------|-------------|--------------|
| **[powerpoint.instructions.md](../../instructions/powerpoint.instructions.md)** | python-pptx | Template-first slide generation, charts, tables | Programmatic PPTX |
| **[excel.instructions.md](../../instructions/excel.instructions.md)** | openpyxl / XlsxWriter | Template-fill reporting, large write-only datasets, memory | Programmatic XLSX |
| **[word.instructions.md](../../instructions/word.instructions.md)** | python-docx / docxtpl | Template-first documents, bulk generation | Programmatic DOCX |

### 🔧 Programming Language Instructions

| Instruction | Language | Focus Areas | Best Practices |
|-------------|----------|-------------|----------------|
| **[java.instructions.md](../../instructions/java.instructions.md)** | Java | OOP, JVM, performance | Enterprise Java |
| **[go.instructions.md](../../instructions/go.instructions.md)** | Go | Concurrency, performance | Cloud-native |
| **[python.instructions.md](../../instructions/python.instructions.md)** | Python | PEP 8, typing, packaging, async, testing | Scripting → services |
| **[rust.instructions.md](../../instructions/rust.instructions.md)** | Rust | Memory safety, performance | Systems programming |
| **[csharp.instructions.md](../../instructions/csharp.instructions.md)** | C# | .NET ecosystem, async programming | Microsoft stack |
| **[sql.instructions.md](../../instructions/sql.instructions.md)** | SQL | Database design, optimization | Data management |

### 🛡️ Security & DevOps Instructions

| Instruction | Domain | Focus Areas | Standards |
|-------------|--------|-------------|----------|
| **[security-audit.instructions.md](../../instructions/security-audit.instructions.md)** | Security | OWASP Top 10, WCAG accessibility, secure coding | Security & compliance |
| **[devops-core-principles.instructions.md](../../instructions/devops-core-principles.instructions.md)** | DevOps | CI/CD, infrastructure, monitoring | DevOps practices |
| **[terraform.instructions.md](../../instructions/terraform.instructions.md)** | Infrastructure | IaC, cloud provisioning | Cloud infrastructure |
| **[kubernetes-deployment-best-practices.instructions.md](../../instructions/kubernetes-deployment-best-practices.instructions.md)** | Kubernetes | Container orchestration | Cloud-native deployment |

### 🚀 Development Practice Instructions

| Instruction | Practice | Focus Areas | Implementation |
|-------------|-----------|-------------|----------------|
| **[tdd-enforcement.instructions.md](../../instructions/tdd-enforcement.instructions.md)** | TDD | Test-driven development, testing patterns | Quality assurance |
| **[testing-strategy.instructions.md](../../instructions/testing-strategy.instructions.md)** | Testing | Test pyramid, coverage, CI/CD gates | Comprehensive testing |
| **[performance-baseline.instructions.md](../../instructions/performance-baseline.instructions.md)** | Performance | Baseline establishment, profiling, optimization | Performance engineering |
| **[brainstorming-techniques.instructions.md](../../instructions/brainstorming-techniques.instructions.md)** | Brainstorming | Creative problem-solving, ideation | Innovation techniques |

### 📝 Documentation & Communication Instructions

| Instruction | Skill | Focus Areas | Application |
|-------------|-------|-------------|------------|
| **[how-to-create-markdown-document.instructions.md](../../instructions/how-to-create-markdown-document.instructions.md)** | Documentation | Markdown, technical writing | Documentation standards |
| **[how-to-create-designmd.instructions.md](../../instructions/how-to-create-designmd.instructions.md)** | Design System | `.plaesy/memory/design.md`, design tokens | DESIGN.md spec (Google Labs) authoring |
| **[changelog.instructions.md](../../instructions/changelog.instructions.md)** | Release Management | Changelog, semantic versioning, release notes | Version management |

### 📋 Workflow & Collaboration Instructions

| Instruction | Purpose | Status | Priority |
|-------------|---------|--------|----------|
| **[git.instructions.md](../../instructions/git.instructions.md)** | Git workflows, branching strategies, commit discipline | ✅ **ACTIVE** | HIGH |
| **[tech-validation.instructions.md](../../instructions/tech-validation.instructions.md)** | Technology evaluation, framework selection, fallbacks | ✅ **ACTIVE** | HIGH |
| **[skills-packaging.instructions.md](../../instructions/skills-packaging.instructions.md)** | When to use an instruction file vs. a Claude Code Skill | ✅ **ACTIVE** | MEDIUM |

### ⚙️ Shared Protocols & System Instructions (Installed)

Installed by `init` either way; the **Always Load** column says whether the file is
in the agent's active context for every project (`always_load`) or is copied and
read on demand (`scope_load`). Only the first is genuinely automatic.

| Instruction | Purpose | Scope | Always Load |
|-------------|---------|-------|------------|
| **[tasks.instructions.md](../../instructions/tasks.instructions.md)** | Task lifecycle across all workflows | Workflow states, task management | ✅ YES |
| **[quality-gates.instructions.md](../../instructions/quality-gates.instructions.md)** | Quality validation framework | Coverage, performance, security gates | ✅ YES |
| **[error-recovery.instructions.md](../../instructions/error-recovery.instructions.md)** | Error classification & recovery | Per-phase recovery strategies | ✅ YES |
| **[date-system.instructions.md](../../instructions/date-system.instructions.md)** | Date variable system | {{CURRENT_DATE}}, {{CURRENT_YEAR}} | ❌ NO — `scope_load`, read on demand |
| **[context-engineering.instructions.md](../../instructions/context-engineering.instructions.md)** | Token & context management | `plaesy trim` (compression) + `plaesy graph` (knowledge graph, impact analysis) | ✅ YES |

---

## 🔧 Instruction Integration

### Instruction Selection Framework

```mermaid

graph TD
    A[Project Analysis] --> B{Technology Stack}
    B --> C[Frontend Framework]
    B --> D[Backend Framework]
    B --> E[Database]
    B --> F[Infrastructure]

    C --> C1[React.js]
    C --> C2[Next.js]
    C --> C3[Angular]
    C --> C4[React Native]

    D --> D1[Spring Boot]
    D --> D2[NestJS]
    D --> D3[Ruby on Rails]
    D --> D4[Dart/Flutter]

    E --> E1[SQL]
    F --> F1[Terraform]
    F --> F2[Kubernetes]

    G[Development Practices] --> H[TDD]
    G --> I[Performance Optimization]
    G --> J[Security]
    G --> K[DevOps]

    L[Load Relevant Instructions] --> M[Apply Best Practices]
    M --> N[Quality Development]

```

### Multi-Technology Projects

#### **Instruction Combination Strategy**

For projects using multiple technologies, combine relevant instructions:

1. **Primary Technology**: Load main framework instruction
2. **Supporting Technologies**: Add supplementary instructions
3. **Cross-Cutting Concerns**: Include security, performance, and DevOps instructions
4. **Development Practices**: Apply TDD and other practice instructions

#### **Example: Full-Stack Web Application**

```text

Primary: nextjs.instructions.md + nestjs.instructions.md
Supporting: sql.instructions.md + terraform.instructions.md
Cross-cutting: security-and-owasp.instructions.md + performance-optimization.instructions.md
Practices: tdd-enforcement.instructions.md + devops-core-principles.instructions.md
```

---

## 🎯 Instruction Structure

### Standard Instruction Format

Each instruction follows this standardized structure:

```markdown

# [Technology/Practice Name] Instructions

## 🎯 Purpose
[Brief description of instruction purpose and scope]

## 📋 Prerequisites
[Required knowledge, tools, or setup]

## 🔧 Core Concepts
[Fundamental concepts and principles]

## 📝 Best Practices
[List of best practices with explanations]

## 🛡️ Security Considerations
[Security-specific guidance if applicable]

## 🚀 Performance Optimization
[Performance tips and techniques]

## 🔍 Code Examples
[Practical code examples]

## 📚 Additional Resources
[Links to additional documentation and resources]

## 🔧 Integration with Plaesy
[How to use with Plaesy framework]
```

### Quality Standards

#### **Instruction Quality Criteria**

- **Accuracy**: Up-to-date and technically accurate
- **Completeness**: Covers all essential aspects
- **Clarity**: Clear and easy to understand
- **Practicality**: Practical and applicable examples
- **Integration**: Compatible with Plaesy framework

#### **Content Standards**

- **Current Best Practices**: Reflects current industry standards
- **Version Specific**: Clearly indicates version compatibility
- **Code Quality**: Includes high-quality, tested code examples
- **Security Focus**: Emphasizes secure development practices
- **Performance Awareness**: Includes performance considerations

---

## 🔧 Plaesy Framework Integration

### Automated Instruction Loading

#### **Context-Based Instruction Selection**

The Plaesy framework automatically selects relevant instructions based on:

1. **Project Analysis**: Results from `plaesy analyze`
2. **Technology Detection**: Identified technologies in project
3. **Framework Detection**: Detected frameworks and platforms
4. **Configuration**: User-specified preferences

#### **Integration with AI Assistants**

```bash

# When AI assistant encounters project:
1. plaesy analyze           # Detect technologies
2. plaesy features paths    # Get context
3. Load relevant instructions based on analysis
4. Apply instruction-specific best practices
5. Validate compliance with guidelines
```

### Instruction Application Process

#### **Development Workflow Integration**

1. **Project Initialization**: Load base instructions
2. **Feature Development**: Apply feature-specific instructions
3. **Code Review**: Validate against instruction guidelines
4. **Quality Assurance**: Ensure instruction compliance
5. **Documentation**: Document instruction usage

#### **Quality Gate Integration**

- **Instruction Compliance**: Check adherence to loaded instructions
- **Best Practice Validation**: Validate against instruction guidelines
- **Security Review**: Ensure security instructions are followed
- **Performance Validation**: Check performance optimization guidelines

---

## 🚀 Usage Guidelines

### For AI Assistants

#### **Instruction Loading Protocol**

```bash

# Critical sequence for AI assistants:
1. Run project analysis: plaesy analyze
2. Get feature context: plaesy features paths
3. Load relevant instructions based on technology stack
4. Apply instruction-specific guidance
5. Validate compliance with loaded instructions
```

#### **Instruction Application Best Practices**

- **Context Awareness**: Understand project context before applying instructions
- **Technology Specificity**: Use technology-specific instructions when available
- **Cross-Technology Integration**: Combine instructions for multi-technology projects
- **Quality Validation**: Ensure compliance with instruction guidelines

### For Development Teams

#### **Instruction Selection Guidelines**

1. **Project Analysis**: Analyze project requirements and technology stack
2. **Instruction Mapping**: Map technologies to relevant instructions
3. **Customization**: Customize instructions based on project needs
4. **Integration**: Integrate instructions into development workflow
5. **Validation**: Validate instruction compliance regularly

#### **Team Training**

- **Instruction Workshops**: Train team on relevant instructions
- **Best Practice Sessions**: Review instruction best practices
- **Code Reviews**: Include instruction compliance in code reviews
- **Knowledge Sharing**: Share insights from instruction application

---

## 📊 Instruction Categories Deep Dive

### 🌐 Frontend Framework Instructions

#### **React.js Instructions Focus**

- **Component Architecture**: Modern React patterns and best practices
- **State Management**: Redux, Context API, and state management patterns
- **Performance**: Optimization techniques and performance monitoring
- **Testing**: React Testing Library and component testing strategies
- **Security**: Secure React development practices

#### **Next.js Instructions Focus**

- **SSR/SSG**: Server-side rendering and static site generation
- **Routing**: Advanced routing patterns and optimization
- **API Routes**: Backend integration and API development
- **Performance**: Optimization for production environments
- **Deployment**: Production deployment best practices

### ⚙️ Backend Framework Instructions

#### **Spring Boot Instructions Focus**

- **Microservices**: Microservice architecture with Spring Boot
- **Spring Ecosystem**: Integration with Spring projects
- **Security**: Spring Security and secure API development
- **Performance**: JVM optimization and performance tuning
- **Testing**: Comprehensive testing strategies

#### **NestJS Instructions Focus**

- **Modular Architecture**: Module organization and dependency injection
- **TypeScript**: TypeScript best practices and patterns
- **API Development**: RESTful API design and GraphQL
- **Authentication**: JWT and authentication strategies
- **Microservices**: Microservice patterns with NestJS

### 🛡️ Security & Compliance Instructions

#### **Security Audit (OWASP & WCAG)**

- **OWASP Top 10**: Protection against common vulnerabilities
- **Secure Coding**: Security-focused development practices
- **Authentication**: Secure authentication and authorization
- **Data Protection**: Encryption and data security
- **Accessibility (WCAG 2.1 AA)**: Accessibility compliance guidelines
- **Compliance Frameworks**: GDPR, HIPAA, and other standards

---

## 🔧 Customization and Extension

### Creating Custom Instructions

#### **Instruction Development Process**

1. **Identify Need**: Determine technology or practice not covered
2. **Research Best Practices**: Research current best practices
3. **Create Structure**: Follow standard instruction format
4. **Content Development**: Write comprehensive instruction content
5. **Validation**: Validate instruction accuracy and completeness
6. **Integration**: Integrate with Plaesy framework

> **Golden rule self-check** (Anthropic prompting best practices, retrieved
> 2026-09-25): before finalizing any instruction file, show it to a colleague
> with minimal context on the task and ask them to follow it. If they'd be
> confused, Claude will be too — revise until the instruction is unambiguous to
> a first-time reader.

#### **Quality Assurance**

- **Technical Review**: Review by technical experts
- **Practical Testing**: Test instructions in real projects
- **Peer Review**: Review by other developers
- **Regular Updates**: Keep instructions current with technology changes

### Instruction Maintenance

#### **Regular Updates**

- **Technology Changes**: Update for new versions and features
- **Best Practice Evolution**: Incorporate new best practices
- **Security Updates**: Include new security considerations
- **Performance Improvements**: Add new optimization techniques

#### **Community Contributions**

- **Feedback Collection**: Collect user feedback on instructions
- **Community Contributions**: Accept community contributions
- **Quality Standards**: Maintain high quality standards
- **Version Management**: Track instruction versions and changes

---

## 🆘 Troubleshooting

### Common Instruction Issues

#### **Instruction Selection Problems**

```bash

# Symptoms: Wrong instructions selected or missing instructions
# Solutions:
1. Verify technology detection in project analysis
2. Check instruction naming conventions
3. Validate instruction content for technology match
4. Consider creating custom instructions if needed
```

#### **Integration Issues**

```bash

# Symptoms: Instructions don't integrate with development workflow
# Solutions:
1. Check instruction compatibility with Plaesy framework
2. Validate instruction structure and format
3. Test instruction application in sample projects
4. Update integration scripts if needed
```

#### **Quality Issues**

```bash

# Symptoms: Instructions are outdated or inaccurate
# Solutions:
1. Review and update instruction content
2. Validate against current best practices
3. Test with current technology versions
4. Get community feedback and validation
```

---

## 📞 Support and Resources

### Documentation Resources

- **Main Documentation**: [../../README.md](../../README.md)
- **Templates Integration**: [../templates/README.md](../templates/README.md)
- **Prompts Integration**: [../prompts/README.md](../prompts/README.md)

### External Resources

- **Official Documentation**: Links to official technology documentation
- **Community Resources**: Community forums and discussion boards
- **Training Materials**: Training and certification resources
- **Best Practice Guides**: Industry best practice guides

### Support Channels

- **GitHub Issues**: [github.com/plaesy/spec-kit/issues](https://github.com/plaesy/spec-kit/issues)
- **Discussions**: [github.com/plaesy/spec-kit/discussions](https://github.com/plaesy/spec-kit/discussions)
- **Instruction Requests**: Request new instructions or updates

---

**📖 This instructions documentation provides comprehensive guidance for using, customizing, and extending technology-specific instructions within the Plaesy framework. All instructions are designed to
integrate seamlessly with automated development workflows and ensure best practice compliance.**
