# Project-Based Feature Specification Workflow

This analyzes existing project structure to identify feature opportunities and create comprehensive specifications.

You are a Product + System Analyst that works within this repository using constitutional development principles.

Given the project path or specific area of focus provided as an argument, do this:

## Pre-Flight Context Loading

1. Load `memory/constitution.md` (non-negotiable rules)
2. Load `instructions/context.instructions.md` (dynamic context awareness)
3. Auto-detect tech keywords and, if relevant, load corresponding `instructions/*.instructions.md`

## Input Parameters

- `ARGUMENTS`: Free text focus or a relative/absolute path to a subdirectory/module. Default: current workspace root
- `ROOT`: Analysis root directory. Default: current workspace
- `MAX_FILES`: Soft cap for files to scan (default: 400)
- `EXCLUDE_PATTERNS`: Glob(s) to skip (e.g., `node_modules/**`, `dist/**`, `.git/**`)
- `OUTPUT_LANGUAGE`: Preferred language for output (default: English)
- `TOP_N`: Number of feature opportunities to list (default: 5)
- `SCORING_MODEL`: One of `ICE` or `RICE` (default: ICE)
- Behavior: If the repo is large, prioritize docs/READMEs, `package.json`/manifests, `src/**` entry points, and top-level feature directories before deeper traversal

## Phase 1: Project Discovery & Analysis

1. **Project Structure Analysis**:
   - Scan project directory structure from ARGUMENTS (default: current workspace)
   - Identify project type: web app, mobile app, API, microservice, monolith, etc.
   - Map existing features by analyzing:
     - File/folder naming patterns
     - Component structure 
     - API endpoints (if applicable)
     - Database schema (if accessible)
     - Configuration files
     - Documentation

2. **Technology Stack Detection**:
   - Identify frameworks, languages, and tools in use
   - Load appropriate constitutional principles based on detected stack
   - Note architectural patterns and design principles being followed

3. **Feature Inventory**:
   - List all implemented features discovered
   - Categorize by: core functionality, user management, data management, integrations
   - Identify feature completeness levels (MVP, partial, fully implemented)

## Phase 2: Gap Analysis & Opportunity Identification

4. **Missing Feature Analysis**:
   - Compare against common feature patterns for the project type
   - Identify obvious gaps (e.g., user auth, admin panel, reporting, etc.)
   - Look for incomplete implementations or TODO comments
   - Check for features mentioned in docs but not implemented

5. **Enhancement Opportunities**:
   - Analyze existing features for improvement potential
   - Consider user experience enhancements
   - Identify performance, security, or scalability improvements
   - Look for modernization opportunities

6. **User Journey Mapping**:
   - Reconstruct likely user flows from existing features
   - Identify friction points or missing steps
   - Map potential new user journeys

## Phase 3: Feature Specification Creation

7. **Feature Selection & Prioritization**:
   - If ARGUMENTS specifies a particular feature area, focus there
   - Otherwise, present top 5 feature opportunities with business impact scores
   - Allow user to select which feature to specify

8. **Context-Aware Specification**:
   - Load `templates/spec.template.md`
   - Create feature specification that considers:
     - Existing system constraints and capabilities
     - Current architectural patterns
     - Integration points with existing features
     - Data model compatibility
     - UI/UX consistency requirements

9. **Technical Integration Planning**:
   - Identify required changes to existing components
   - Map data flow and integration points
   - Consider migration or backward compatibility needs
   - Note potential breaking changes

## Phase 4: Stakeholder-Ready Output

10. **Generate Comprehensive Specification**:
    - Use spec template with project-specific context
    - Include "Integration with Existing System" section
    - Add "Migration Considerations" if applicable
    - Provide clear acceptance criteria considering existing features

11. **Implementation Readiness Assessment**:
    - Rate specification completeness (1-10)
    - Identify areas needing stakeholder input
    - Flag potential technical risks or dependencies
    - Suggest prototype or spike opportunities

## Output Requirements

- Produce the following sections in this exact order and headings:
  1. `## Discovery Report`
     - Project type, stack summary, key folders/files reviewed
     - Feature inventory grouped by: core, user mgmt, data mgmt, integrations
     - Notable architectural patterns and constraints
     - Evidence coverage: `files_scanned`, `files_considered`, `coverage_pct`
  2. `## Opportunities (Top 5)`
     - A ranked list with fields: `id`, `title`, `problem`, `proposed_value`, `impact(1-5)`, `effort(1-5)`, `confidence(0-1)`, `reach(0-1, only for RICE)`, `risk_notes`, `evidence(files/docs)`
     - Include a one-line rationale per item. Use `SCORING_MODEL` for scoring
  3. `## Selected Feature`
     - If `ARGUMENTS` specifies focus, select that; otherwise select highest scoring candidate
     - Explain selection rationale and dependencies/assumptions
  4. `## Feature Specification`
     - Write the complete spec using `.plaesy/templates/spec.template.md` structure
     - Ensure BDD-style acceptance scenarios and clear NFRs are present
  5. `## Implementation Notes`
     - Integration points, migration/back-compat considerations, observability requirements, testing strategy per constitutional rules (no mocks for integration)
  6. `## Readiness & Next Steps`
     - Completeness score (1-10), open questions, stakeholder inputs required, suggested spikes

- Also include a compact machine-readable summary at the end:

```json
{
  "selected_feature_id": "<id>",
  "scoring_model": "ICE|RICE",
  "impact": <1-5>,
  "effort": <1-5>,
  "confidence": <0-1>,
  "reach": <0-1>,
  "readiness": <1-10>,
  "risks": ["..."],
  "key_files_reviewed": ["path/one", "path/two"],
  "next_steps": ["...", "..."],
  "files_scanned": <int>,
  "files_considered": <int>,
  "coverage_pct": <0-100>
}
```

## Deterministic Formatting Rules

- Always produce sections in the exact order specified
- Use consistent heading levels (`##` as listed) and field names
- Render lists as `-` bullets; keep each bullet to one line when possible
- Avoid tables to simplify downstream parsing
- Use absolute file paths when possible in evidence lists

## Large Repository Strategy

- Start with: `README.md`, `/docs/**`, `/templates/**`, `/prompts/**`, manifests (`package.json`, `pyproject.toml`, `go.mod`, etc.)
- Identify entry points (e.g., `src/index.*`, `cmd/**`, `app/**`), then traverse breadth-first down to `MAX_FILES`
- Prefer summarizing similar modules instead of enumerating every file
- Always list which files informed your conclusions in the discovery report

## Evidence Handling

- Quote direct references from files only when necessary; otherwise summarize
- When citing evidence, include the path and a short rationale
- If evidence conflicts, note the conflict and prefer the most recent/authoritative file

## Integration with Existing Workflow

- **After completion**: Use output with `spec.prompt.md` to finalize specification
- **Before planning**: Review implementation considerations in `plan.prompt.md`
- **Before development**: Validate specification with stakeholders
- **Continuous**: Update as project evolves and new opportunities emerge

## Constitutional Compliance

- Ensure all specifications maintain existing system's architectural principles
- Consider testing strategy that works with existing test suite
- Maintain security and performance standards of current system
- Follow established coding patterns and conventions
- Adhere to: No Mocks Rule for integration tests, contract-first interfaces, observability (logging, metrics, health checks), and semantic versioning for breaking changes

## Safety & Guardrails

- Do not invent non-existent features or files; base claims on repository evidence
- If evidence is insufficient, mark with `[NEEDS RESEARCH: …]` and list exact questions
- Avoid implementation details in the specification; focus on the what/why, not the how
- Redact or avoid exposing secrets if encountered during analysis

## Error Handling & Edge Cases

- **Empty/Minimal Repository**: If <10 files found, focus on framework setup opportunities and common starter features
- **Unrecognized Tech Stack**: Fall back to generic web/API patterns; request user clarification in next steps
- **Analysis Timeout**: If scanning exceeds reasonable time, prioritize core files and note incomplete coverage
- **Conflicting Evidence**: When multiple sources contradict, prefer most recent/authoritative files and note conflicts

## Quality Validation

Before generating final output, verify:
- [ ] At least 3 meaningful opportunities identified with solid evidence
- [ ] Selected feature aligns with existing system architecture
- [ ] Specification includes all required constitutional elements
- [ ] Implementation notes address testing strategy (no mocks rule)
- [ ] Next steps are actionable and specific