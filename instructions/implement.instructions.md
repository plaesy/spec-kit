---
description: "Generate complete implementation from specifications with TDD enforcement"
---

# `/implement` command instructions

⚡ **Run with**: `ultracode` (multi-agent parallel execution)

## Objective
Universal implementation orchestrator: Execute specialized implementation based on project type
- Code generation (Go, Java, TypeScript, Python, Rust, etc.)
- Design implementation (Figma design specs, component generation)
- Full-stack implementation (code + design + tools)

## Protocol
**Smart orchestration**: Detect project type → Load appropriate technology/tool instructions → Execute per those specs

## Implementation Orchestration Process
1. **Project Analysis** - Detect project type, technology stack, requirements
2. **Instruction Loading** - Load appropriate instruction files:
   - **Technology-specific**: go.instructions.md, java.instructions.md, reactjs.instructions.md, etc.
   - **Tool-specific**: figma.instructions.md, terraform.instructions.md, etc.
   - **Specialty**: tdd-enforcement.instructions.md, security-and-owasp.instructions.md, etc.
3. **Context Validation** - Verify all specs, architecture, design requirements
4. **Execution** - Execute per loaded instructions (TDD, security, tooling, etc.)
5. **Quality Verification** - All quality gates + automated validation
6. **Documentation** - Per technology best practices

## Constitutional Rules (Mandatory)
- ✅ **TDD REQUIRED** - Red-Green-Refactor for all code
- ✅ **90% COVERAGE** - Minimum test coverage enforced
- ✅ **MODERN TECH ONLY** - Context7 validated technologies
- ✅ **SECURITY-FIRST** - OWASP compliance built-in
- ✅ **QUALITY GATES** - Automated validation mandatory

## Context7 Protocol (Required)
**Before ANY implementation**:
1. `mcp__context7__resolve-library-id` for ALL technologies
2. `mcp__context7__get-library-docs` (tokens=500)
3. **Citation**: "Implementation based on Context7 (/library/id) - Retrieved {{CURRENT_DATE}}"
4. **Error Recovery**: Alternative sources if Context7 fails

## TDD Enforcement (Non-Negotiable)
### Red-Green-Refactor Cycle
1. **RED** - Write failing test first
2. **GREEN** - Make test pass with minimal code
3. **REFACTOR** - Improve code quality
4. **COMMIT** - Each phase with specific commit messages

### Commit Pattern
```
RED: Add user authentication validation test
GREEN: Implement authentication service with JWT
REFACTOR: Extract validation logic to separate module
```

## Validation Checklist (Before Running)
- ✅ Specs are clear (research done, gaps found via `/assess`)
- ✅ Project type detected (language, framework, tools identified)
- ✅ Appropriate instruction files available (tech + tools)
- ✅ Design specifications available (if frontend/UI needed)
- ✅ TDD environment ready (test framework installed)
- ✅ No blocking dependencies missing

## Technology-Specific & Tool-Specific Implementation

### Step 1: Detect Project Needs & Load Instructions

Based on project analysis, load appropriate instructions:

**By Technology (Languages & Frameworks)**:
- Go: `go.instructions.md`
- Java: `java.instructions.md`
- C#: `csharp.instructions.md`
- Rust: `rust.instructions.md`
- Python: (covered by backend frameworks)
- TypeScript/JavaScript: 
  - React: `reactjs.instructions.md`
  - Vue: (covered in framework mapping)
  - Angular: `angular.instructions.md`
  - Next.js: `nextjs.instructions.md`
  - NestJS: `nestjs.instructions.md`
- Mobile: React Native: `react-native.instructions.md`, Flutter: `dart-n-flutter.instructions.md`
- Backend: Spring Boot: `springboot.instructions.md`, Rails: `ruby-on-rails.instructions.md`

**By Tools & Specialties**:
- **Figma Design**: `figma.instructions.md` (if UI/UX needed)
- **Database**: `sql.instructions.md`
- **DevOps**: `devops-core-principles.instructions.md`, `kubernetes-deployment-best-practices.instructions.md`
- **Infrastructure**: `terraform.instructions.md`
- **Testing**: `testing-strategy.instructions.md`, `tdd-enforcement.instructions.md`
- **Security**: `security-and-owasp.instructions.md`

### Step 2: Execute Per Loaded Instructions

Follow the loaded instruction file(s) exactly. Multi-technology projects load multiple instructions:
```
Example: React + Node.js + PostgreSQL + Docker
Load: reactjs.instructions.md + nestjs.instructions.md + sql.instructions.md + devops-core-principles.instructions.md
Execute in sequence per each instruction spec
```

### Step 3: Universal Constraints (All Technologies)

Regardless of technology chosen:
- ✅ TDD enforcement (per tdd-enforcement.instructions.md)
- ✅ 90%+ test coverage mandatory
- ✅ OWASP security compliance
- ✅ Modern tech stack (Context7 validated)
- ✅ Complete documentation

## Anti-Patterns (NEVER Do These)
- ❌ **Never skip TDD** - Tests written FIRST (Red-Green-Refactor per loaded instructions)
- ❌ **Never commit <90% coverage** - All new code must have tests
- ❌ **Never use deprecated libraries** - All tech stack verified current via Context7
- ❌ **Never mix technologies arbitrarily** - Use spec'd tech stack only
- ❌ **Never skip security implementation** - OWASP compliance mandatory
- ❌ **Never deviate from specs** - Follow research/requirements exactly
- ❌ **Never skip accessibility on UI** - WCAG AA compliance mandatory for frontend
- ❌ **Never ignore loaded instructions** - Execute per technology-specific specs
- ❌ **Never implement both design & code simultaneously without coordination** - Load both instruction sets, coordinate execution

## Error Recovery
**If tests fail:**
- Report which test failed + error message
- Do NOT skip failing tests
- Fix root cause + re-run

**If build fails:**
- Report compilation error + line number
- Fix + rebuild before proceeding

## Deliverables
- Complete source code
- Comprehensive tests (90%+ coverage)
- Documentation with examples
- Configuration files
- Deployment scripts

## Completion Format
✅ Implementation Complete
- Coverage: [XX%] (must be ≥90%)
- Tests: [all passing/[N] failing]
- Deliverables: [code, tests, docs, config]
- Next phase: /assess

## Progress Format
```
Status: Implementing...
Component: [current component]
TDD Phase: [RED/GREEN/REFACTOR]
Coverage: [XX%]
```

## Need Help Deciding?

When facing difficult decisions during implementation, call **`@nara`**:
- Architecture patterns (which pattern best?)
- Refactoring vs shipping (now or later?)
- Technology choices (library X vs Y?)

Nara will gather perspectives from dev, architect, product and make the decision.

## Next Steps
- `/assess` - Quality assessment
- `/optimize` - Performance improvements
- `/fix` - Resolve issues
- `@nara` - Get help on tough implementation decisions

/save

---

**Follow shared protocols**: `.plaesy/memory/quality-gates.md` → `.plaesy/memory/error-recovery.md`