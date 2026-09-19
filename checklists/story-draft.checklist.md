# Story Draft Checklist

The Scrum Master should use this checklist to validate that each story contains sufficient context for a developer agent to implement it successfully, while assuming the dev agent has reasonable capabilities to figure things out.

[[LLM: Gather story doc, parent epic, referenced architecture/design docs, and prior related stories. Validate BEFORE implementation begins, against: clarity (what to build), context (why/fit), guidance (key decisions/patterns), testability, self-containment. Assume competent dev agents who research, decide reasonably, follow patterns, and ask only when truly stuck — check for sufficient guidance, not exhaustive detail.]]

## 1. GOAL & CONTEXT CLARITY

[[LLM: Verify the story states what to build, why it matters, how it fits the epic, explicit dependencies, and a concrete (not vague) definition of success.]]

- [ ] Story goal/purpose is clearly stated
- [ ] Relationship to epic goals is evident
- [ ] How the story fits into overall system flow is explained
- [ ] Dependencies on previous stories are identified (if applicable)
- [ ] Business context and value are clear

## 2. TECHNICAL IMPLEMENTATION GUIDANCE

[[LLM: Check that key files/components, non-obvious tech choices, integration points, data models/API contracts, and non-standard patterns are called out — important files only, not exhaustive.]]

- [ ] Key files to create/modify are identified (not necessarily exhaustive)
- [ ] Technologies specifically needed for this story are mentioned
- [ ] Critical APIs or interfaces are sufficiently described
- [ ] Necessary data models or structures are referenced
- [ ] Required environment variables are listed (if applicable)
- [ ] Any exceptions to standard coding patterns are noted

## 3. REFERENCE EFFECTIVENESS

[[LLM: References should help, not send readers on a treasure hunt — check they point to specific sections, explain relevance, summarize critical info inline, aren't broken, and carry forward needed prior-story context.]]

- [ ] References to external documents point to specific relevant sections
- [ ] Critical information from previous stories is summarized (not just referenced)
- [ ] Context is provided for why references are relevant
- [ ] References use consistent format (e.g., `docs/filename.md#section`)

## 4. SELF-CONTAINMENT ASSESSMENT

[[LLM: Stories should stand mostly alone — core requirements inline (not just referenced), domain terms explained, assumptions explicit, edge cases mentioned, understandable without reading many other documents.]]

- [ ] Core information needed is included (not overly reliant on external docs)
- [ ] Implicit assumptions are made explicit
- [ ] Domain-specific terms or concepts are explained
- [ ] Edge cases or error scenarios are addressed

## 5. TESTING GUIDANCE

[[LLM: Check that test approach (unit/integration/e2e), key scenarios, measurable success criteria, and special considerations are specified, and acceptance criteria are testable.]]

- [ ] Required testing approach is outlined
- [ ] Key test scenarios are identified
- [ ] Success criteria are defined
- [ ] Special testing considerations are noted (if applicable)

## VALIDATION RESULT

[[LLM: Produce a concise report: (1) readiness (READY/NEEDS REVISION/BLOCKED), clarity score 1-10, major gaps; (2) fill the table below with PASS/PARTIAL/FAIL per category; (3) concrete issues, fixes, and blocking dependencies; (4) developer-perspective check — could you implement this as written, what questions remain, what risks delay/rework. Be pragmatic: enough context to avoid a mess, not perfect documentation.]]

| Category                             | Status | Issues |
| ------------------------------------ | ------ | ------ |
| 1. Goal & Context Clarity            | _TBD_  |        |
| 2. Technical Implementation Guidance | _TBD_  |        |
| 3. Reference Effectiveness           | _TBD_  |        |
| 4. Self-Containment Assessment       | _TBD_  |        |
| 5. Testing Guidance                  | _TBD_  |        |

**Final Assessment:**

- READY: The story provides sufficient context for implementation
- NEEDS REVISION: The story requires updates (see issues)
- BLOCKED: External information required (specify what information)
