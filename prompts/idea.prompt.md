# Idea Exploration and Research Workflow

This is the first step in the Spec-Driven Development lifecycle. The AI will guide users through comprehensive idea exploration and research using advanced brainstorming techniques and systematic validation.

**Input**: {{$ARGUMENTS}} - Raw idea description or problem statement
**Output**: Validated, researched idea document with creative solutions and implementation readiness

Given the implementation details provided as {{$ARGUMENTS}}, do this:

## Phase 1: Initialize Idea Document

1. **Setup Project Structure**:
   - Run platform-appropriate script from repo root and parse JSON output for BRANCH_NAME and IDEA_FILE:
     - **Linux/macOS**: `.plaesy/scripts/bash/create-new-idea.sh "{{$ARGUMENTS}}" --json`
     - **Windows**: `.plaesy/scripts/powershell/create-new-idea.ps1 "{{$ARGUMENTS}}" -Json`
   - All file paths must be absolute
   - Verify script execution success before proceeding

2. **Load Foundation Materials**:
   - Load created IDEA_FILE to understand template structure and placeholders
   - Load `.plaesy/memory/brainstorming-techniques.md` for creative exploration methods
   - Load `.plaesy/memory/constitution.md` for constitutional framework alignment
   - Scan existing `.plaesy/memory/*.md` files for relevant domain knowledge

3. **Validate Input Quality**:
   - Assess idea clarity and completeness
   - Identify ambiguous areas requiring clarification
   - Flag potential scope issues (too broad/narrow)

## Phase 2: AI-Powered Idea Analysis

3. **Problem Definition Analysis**: 
   - Analyze the raw idea to identify the core problem being solved
   - Determine who has this problem and why it's important
   - Assess the problem urgency/pain level (Low/Medium/High/Critical)

4. **Solution Approach Exploration**:
   - **Creative Brainstorming**: Apply techniques from `brainstorming-techniques.md`:
     - **What If Scenarios**: "What if we approach this problem from [different angle]?"
     - **Analogical Thinking**: "What other industries solve similar problems?"
     - **Reversal/Inversion**: "What if we did the opposite of current solutions?"
     - **First Principles**: "What are the fundamental components of this problem?"
   - Generate multiple solution approaches (at least 3-5 different ways) using structured techniques
   - Evaluate technical complexity and feasibility for each approach
   - Apply **SCAMPER Method** (Substitute, Combine, Adapt, Modify, Put to other uses, Eliminate, Reverse) for solution enhancement
   - Identify preferred solution direction with reasoning

5. **Stakeholder & User Analysis**:
   - Identify primary and secondary users/stakeholders
   - Define user personas and their specific needs
   - Map user journey and pain points

## Phase 3: Market & Competitive Research

6. **Automated Market Research**:
   - Research existing solutions and competitors
   - Analyze feature gaps and opportunities
   - Estimate market size and opportunity
   - Identify key challenges and risks

7. **Technical Feasibility Assessment**:
   - Evaluate technical complexity and requirements
   - Identify necessary technologies and platforms
   - Assess resource and timeline implications
   - Flag potential technical risks

## Phase 4: Interactive Exploration & Creative Brainstorming

8. **Guided User Questions**: Ask the user targeted questions to fill knowledge gaps:
   - **Problem Validation**: "What specific problem does this solve that existing solutions don't?"
   - **User Focus**: "Who are your primary users? (be very specific)"
   - **Success Metrics**: "How will you measure success for this solution?"
   - **Constraints**: "What constraints (budget, timeline, technology) should we consider?"
   - **Risks**: "What could make this project fail?"

9. **Systematic Creative Exploration**: Apply structured techniques from `brainstorming-techniques.md`:

   **Phase 9a: Divergent Thinking (Generate Options)**
   - **SCAMPER Method**: Substitute, Combine, Adapt, Modify, Put to other uses, Eliminate, Reverse
   - **"Yes, And..." Building**: Build on ideas collaboratively with affirmative expansion
   - **Random Stimulation**: Use unexpected elements to spark innovative connections
   - **Morphological Analysis**: Break problem into parameters and explore all combinations

   **Phase 9b: Deep Problem Analysis**
   - **Five Whys**: Drill down to root causes and core motivations with systematic questioning
   - **Assumption Reversal**: Challenge and invert fundamental assumptions about the problem
   - **Role Playing**: Explore solutions from diverse stakeholder perspectives systematically
   - **Systems Thinking**: Map interconnections and feedback loops in the problem space

   **Phase 9c: Future-Oriented Innovation**
   - **Time Shifting**: "How would this be solved in 1995? 2030? 2050?" with technological context
   - **Resource Constraints**: Explore solutions with unlimited budget vs $10 constraint
   - **Forced Relationships**: Connect unrelated domains to find innovative solution bridges
   - **Metaphor Mapping**: Use extended metaphors to explore solution landscapes creatively

   **Phase 9d: Solution Convergence**
   - **Evaluation Matrix**: Score solutions on feasibility, impact, and innovation
   - **Risk-Benefit Analysis**: Assess each solution's potential outcomes
   - **Hybrid Solutions**: Combine best elements from multiple approaches

10. **Type-Specific Deep Dive**: Based on idea type (application, tool, service, etc.), ask specialized questions:
    - **Application**: Platforms, user journey, data handling, integrations
    - **Tool**: Input/output, distribution, configuration needs
    - **Service**: APIs, dependencies, performance, security
    - **Research**: Methodology, audience, prior research

## Phase 5: Documentation & Synthesis

11. **Update Idea Document**: Write comprehensive findings to IDEA_FILE, including:
    - **Enhanced Problem Statement**: Clear, validated problem description with evidence
    - **Market Research**: Competitive analysis, opportunities, challenges with sources
    - **Solution Approaches**: Multiple options with pros/cons and recommendation matrix
      - Include brainstorming technique used for each approach
      - Document creative insights and breakthrough moments
      - Map solution evolution through different brainstorming methods
    - **Stakeholder Analysis**: User personas, requirements, and decision makers
    - **Technical Assessment**: Complexity, risks, requirements, and technology recommendations
    - **Success Criteria**: Measurable goals, KPIs, and acceptance metrics
    - **Research Questions**: Mark unclear aspects with [NEEDS RESEARCH: specific question]
    - **Creative Process Documentation**: Record which brainstorming techniques were most effective
    - **Metadata Section**: Add standardized metadata for handoff:
      ```yaml
      ---
      idea_id: generated_id
      branch_name: current_branch
      research_status: COMPLETE|PARTIAL|PENDING
      complexity_assessment: Low|Medium|High|Critical
      recommended_solution: solution_name
      brainstorming_techniques_used: [technique1, technique2, technique3]
      creative_insights: [insight1, insight2]
      next_phase: SPECIFICATION
      handoff_ready: true|false
      ---
      ```
    - **Constitutional Compliance**: Validate against framework principles
    - **Specification Readiness**: Assessment of readiness for next phase

12. **Quality Gate Validation**: Validate readiness for specification phase:
    - **Completeness Check**: All required sections documented
    - **Evidence Validation**: Research findings backed by credible sources  
    - **User Validation**: Stakeholder input incorporated and validated
    - **Technical Feasibility**: Architecture approach confirmed as viable
    - **Constitutional Alignment**: Framework compliance verified
    - **Handoff Readiness**: All necessary context available for specify phase
    - **Research Gap Assessment**: Outstanding questions documented with mitigation plans

## Phase 6: Completion & Handoff

13. **Generate Standardized Output**: Report completion with JSON metadata for seamless handoff:
    ```json
    {
      "BRANCH_NAME": "branch_name",
      "IDEA_FILE": "/absolute/path/to/idea.md",
      "STATUS": "RESEARCH_COMPLETE",
      "HANDOFF_DATA": {
        "validated_problem": "Clear problem statement",
        "recommended_solution": "Preferred solution approach",
        "priority_features": ["feature1", "feature2"],
        "technical_complexity": "Low|Medium|High|Critical",
        "user_personas": ["persona1", "persona2"],
        "success_metrics": ["metric1", "metric2"],
        "research_gaps": ["gap1", "gap2"],
        "brainstorming_insights": ["insight1", "insight2"],
        "creative_breakthroughs": ["breakthrough1", "breakthrough2"]
      }
    }
    ```

14. **Prepare Specify Phase Input**: Create ready-to-use input for `specify.prompt.md`:
    - Validated problem statement with evidence
    - Prioritized solution approaches with rationale
    - Complete stakeholder and user analysis
    - Technical architecture recommendations
    - Identified constraints and dependencies
    - Creative insights and breakthrough solutions from brainstorming sessions

## Quality Assurance Guidelines

### Research Excellence
- **Deep Understanding**: Prioritize comprehensive problem comprehension over premature solutions
- **Evidence-Based**: Back all findings with credible research, data, and validated assumptions
- **Multiple Perspectives**: Consider diverse stakeholder viewpoints and cultural contexts
- **Competitive Intelligence**: Thoroughly analyze existing solutions and market gaps

### Creative Methodology
- **Systematic Innovation**: Apply structured brainstorming techniques from `brainstorming-techniques.md`
- **Divergent-Convergent Flow**: Generate many options before selecting optimal approaches
- **Cross-Pollination**: Draw inspiration from unrelated domains and industries
- **Future-Proofing**: Consider technological evolution and changing user behaviors

### User-Centricity
- **Empathy-Driven**: Keep user needs, pains, and goals at the center of all exploration
- **Persona Validation**: Test ideas against detailed user persona scenarios
- **Journey Mapping**: Understand complete user experience from awareness to advocacy
- **Accessibility**: Consider diverse user abilities and circumstances

### Constitutional Framework Alignment
- **TDD Readiness**: Structure ideas for test-driven development implementation
- **Interface Design**: Prepare for clear contracts, error handling, and versioning
- **Real Dependencies**: Plan for integration testing with actual external systems
- **Observability**: Design for monitoring, logging, and performance measurement
- **Platform Agnostic**: Ensure solutions work across different platforms and environments
- **Security First**: Embed security considerations from the earliest ideation phase

### Collaboration & Validation
- **Interactive Exploration**: Engage users in co-creation and idea refinement
- **Transparent Uncertainty**: Clearly mark areas requiring additional research with [NEEDS RESEARCH: specific question]
- **Iterative Refinement**: Build on user feedback to improve and validate ideas
- **Stakeholder Alignment**: Ensure ideas align with business objectives and constraints

Note: This phase focuses on understanding and researching the problem/opportunity thoroughly using structured creative techniques. The goal is a well-researched idea document that provides a solid foundation for creating detailed specifications that comply with the constitutional framework. The integration of `brainstorming-techniques.md` ensures systematic creativity and comprehensive solution exploration.