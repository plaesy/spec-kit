---
description: Create or update the feature specification from researched idea document or natural language feature description.
---

# Feature Specification Creation Workflow

This creates detailed technical specifications from researched ideas or natural language descriptions.

Given the feature description provided as an argument, do this:

## Phase 1: Input Analysis & Setup

1. **Determine Input Type**:
   - If ARGUMENTS references an existing idea file (contains `.md` or file path), treat as researched idea
   - Otherwise, treat as natural language feature description requiring new setup

2. **Setup Based on Input Type**:
   
   **For Researched Ideas (from idea.prompt.md)**:
   - Parse ARGUMENTS to extract idea file path 
   - Load the idea file to extract research findings, problem analysis, and solution approaches
   - Use existing branch and directory structure from idea phase
   - Extract BRANCH_NAME and SPECS_DIR from idea file metadata

   **For New Feature Descriptions**:
   - Run `.plaesy/scripts/bash/create-new-feature.sh --json "$ARGUMENTS"` from repo root
   - Parse JSON output for BRANCH_NAME and SPEC_FILE
   - All file paths must be absolute

## Phase 2: Research Integration & Analysis

3. **Load Foundation Materials**:
   - Load `.plaesy/templates/spec.template.md` to understand required specification sections
   - If working from idea file, extract all research findings including:
     - Problem statement and user analysis
     - Market research and competitive analysis  
     - Technical feasibility assessment
     - Stakeholder requirements
     - Success criteria and KPIs

4. **Gap Analysis**:
   - Identify missing information needed for complete specification
   - Flag areas where additional research or stakeholder input is required
   - Prioritize missing elements based on specification criticality

## Phase 3: Specification Development

5. **Create Comprehensive Specification**:
   - Write specification to SPEC_FILE using template structure
   - Replace placeholders with concrete details from research/analysis
   - Preserve section order and headings from template
   
   **Key Sections to Complete**:
   - **Problem Statement**: Use researched and validated problem definition
   - **User Stories & Scenarios**: Convert research into specific user stories with acceptance criteria
   - **Functional Requirements**: List specific capabilities based on solution approach
   - **Non-Functional Requirements**: Add performance, security, scalability needs from technical assessment
   - **System Architecture**: High-level design based on technical complexity analysis
   - **API Specifications**: Define interfaces and data contracts
   - **Security Requirements**: Authentication, authorization, data protection based on risk assessment
   - **Testing Strategy**: Test plans aligned with constitutional framework requirements
   - **Implementation Plan**: Phased approach based on complexity and dependencies

6. **Quality Assurance**:
   - Ensure all research findings are properly integrated
   - Validate that specification addresses all identified user needs
   - Check that technical requirements align with feasibility assessment
   - Verify security and compliance requirements are addressed

## Phase 4: Research Gaps & Next Steps

7. **Document Remaining Unknowns**:
   - Clearly mark items requiring additional research with [NEEDS RESEARCH: specific question]
   - Identify dependencies on external stakeholders or technical validation
   - Suggest specific research activities to resolve gaps

8. **Integration Recommendations**:
   - Define how this feature integrates with existing systems
   - Specify data migration requirements if applicable
   - Identify potential breaking changes and mitigation strategies

## Phase 5: Completion & Handoff

9. **Validation & Review**:
   - Ensure specification is complete and actionable for development teams
   - Validate alignment with original problem statement and success criteria
   - Check constitutional framework compliance (TDD, real dependency testing, etc.)

10. **Report Completion**:
    - Provide branch name, spec file path, and readiness assessment
    - Summarize key findings and decisions from research integration
    - Recommend next phase actions (planning, architecture, prototyping)

## Guidelines

- **Research-Driven**: Leverage all available research to create evidence-based specifications
- **User-Centric**: Ensure specifications directly address identified user needs and pain points
- **Technically Grounded**: Base technical decisions on feasibility assessments and constraints
- **Implementation-Ready**: Create specifications detailed enough for development teams to begin work
- **Constitutional Compliance**: Ensure all specifications align with constitutional framework principles

Note: This phase transforms researched ideas into actionable technical specifications. When working from idea.prompt.md output, leverage all research findings to create comprehensive, well-founded specifications.
