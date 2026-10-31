# Product Owner (PO) Backlog & Value Delivery Checklist

This checklist validates that the product backlog is ready to deliver value — that the
vision is articulated, the backlog is clean, prioritization is defensible, stories are
ready, and acceptance is objective. It is a **product** checklist: it does not validate
infrastructure, deployment, or architecture. Those are owned by
`sa.checklist.md` (technical design), `qa.checklist.md` (test strategy and sign-off), and
`pm.checklist.md` (PRD and epic structure).

[[LLM: Required inputs: prd.md (docs/prd.md) or brownfield-prd.md, the current backlog/epic/story files, the product vision statement, and any value or usage metrics — derive whatever is
missing from existing docs/backlog history and label the gap `ASSUMED — <default>` rather than leaving it unresolved. Detect project type: GREENFIELD (new product, needs prd.md) or
BROWNFIELD (change to an existing product, needs brownfield-prd.md plus the existing backlog and release history). Skip [[BROWNFIELD ONLY]] sections that do not apply and note each skip
in the report. This is a backlog-readiness review, not a product-strategy review: if the vision itself is missing or contradictory, report it as a blocker under §1 rather than rewriting it
here. Validate each item with cited evidence from the backlog files, question assumptions, and judge whether a developer could pick up the top story without asking a question. Default to
comprehensive mode (full report at end) unless the user has already stated a preference for interactive (section-by-section).]]

## 1. VISION & PRODUCT DIRECTION

[[LLM: The vision is the input to every prioritization decision below. If it is missing, generic, or unfalsifiable, every downstream check fails — flag it first and do not paper over it.]]

### 1.1 Vision Statement

- [ ] Product vision is documented and states the problem and the intended outcome
- [ ] Target user is specific (a named segment, not "everyone")
- [ ] Vision is falsifiable — a reader can tell whether it was achieved
- [ ] Vision is consistent with the PRD's problem statement (no second, competing vision)
- [ ] Scope decisions can be traced back to a line in the vision

### 1.2 Value Proposition

- [ ] Each epic states the user or business value it delivers
- [ ] Value claims are supported by research or data, not assertion
- [ ] Known alternatives and why this product wins are documented
- [ ] The "do nothing" baseline is stated — what happens if this is not built

## 2. BACKLOG HYGIENE

[[LLM: A dirty backlog makes every priority decision arbitrary. Check for duplicates, orphans, stale items, and items nobody can act on.]]

### 2.1 Backlog Structure

- [ ] Backlog is a single ordered list with a stable ID per item
- [ ] Every item is an Epic or a Story — no free-floating notes, TODOs, or orphan tasks
- [ ] Every Story belongs to exactly one Epic
- [ ] Every Epic is reachable from the product vision (§1)
- [ ] Nothing in the backlog is blocked on a decision that was never made

### 2.2 Backlog Cleanliness

- [ ] No duplicate items (same user need described twice)
- [ ] No item has been sitting untouched past its review date without a reason
- [ ] Rejected or deferred items are removed or explicitly parked with a revisit trigger
- [ ] Technical-debt items are in the backlog as visible, prioritized items — not tracked elsewhere
- [ ] Items dependent on unfinished work are ordered after it

### 2.3 WIP & Flow

- [ ] Work-in-progress limits are set per item type
- [ ] Blocked items are flagged with a blocker and an owner, not left silently
- [ ] The current sprint's scope is drawn from the backlog order, not added ad hoc
- [ ] Anything added mid-sprint is an explicit, visible re-scope decision (§8)

## 3. PRIORITIZATION FRAMEWORK

[[LLM: The framework matters less than its consistent application. A single-item judgment call is fine if the method is stated.]]

### 3.1 Framework Definition

- [ ] A prioritization method is named (RICE, MoSCoW, Kano, WSJF, or an explicit local rule)
- [ ] The inputs the method needs are actually available for each item
- [ ] Scoring is not fabricated — unknown values are marked, not guessed
- [ ] Ties are broken by a stated rule, not by backlog age alone

### 3.2 Framework Application

- [ ] Every item in the top 3 delivery horizons has been scored with the chosen method
- [ ] Scores are recorded next to the item, not held in someone's head
- [ ] The order is the descending order of the scores
- [ ] Items deliberately ordered against the score carry a written reason
- [ ] Must-have, should-have, and won't-have are distinguished

### 3.3 Trade-off Decisions

- [ ] Build-vs-buy and fix-vs-build calls are recorded as backlog decisions
- [ ] The cost of the item being deferred is stated, not assumed to be zero
- [ ] Anything cut for capacity is listed so it can be revisited next horizon
- [ ] The MVP cut line is explicit — what is deliberately out of scope is named

## 4. USER STORIES

[[LLM: Read the stories as a developer who has never seen the product. If a sentence needs the product owner's memory to interpret, the story is not ready.]]

### 4.1 Story Form

- [ ] Story states user, capability wanted, and the reason it matters
- [ ] Story describes need and outcome, not the implementation
- [ ] Story is written from one user's perspective — a story spanning three roles is three stories
- [ ] Story is small enough to deliver and verify inside one iteration
- [ ] Story states the value that makes it worth doing

### 4.2 Story Validity (INVEST)

- [ ] **Independent** — deliverable without depending on an unfinished sibling
- [ ] **Negotiable** — leaves the solution open, does not over-specify the implementation
- [ ] **Valuable** — delivers user or business value, not just a technical step
- [ ] **Estimable** — the team can size it; unknowns are listed
- [ ] **Small** — fits an iteration, with an explicit split if it does not
- [ ] **Testable** — a test for pass/fail can be written from the acceptance criteria

### 4.3 Story Detail

- [ ] Preconditions and postconditions are stated
- [ ] Edge and error scenarios are named, even if deferred — a deferral is a decision, not an omission
- [ ] Non-functional expectations that affect the story (limits, permissions, data volume) are stated
- [ ] Anything the story depends on is linked, not described from memory

## 5. ACCEPTANCE CRITERIA

[[LLM: Acceptance criteria are the contract between the backlog and QA. They must be objective: a tester who did not attend planning must be able to decide pass or fail from the text alone.]]

### 5.1 Criteria Quality

- [ ] Every story in the delivery horizon has acceptance criteria
- [ ] Criteria are written as observable outcomes, not implementation instructions
- [ ] Each criterion is pass/fail decidable with no interpretation
- [ ] Criteria cover the happy path, the error path, and the boundary values
- [ ] Criteria state the expected result, not only the action
- [ ] No criterion is a restatement of the story title

### 5.2 Criteria Testability

- [ ] Local testability is defined for backend and data stories (e.g. verifiable through the CLI) — `story-done.checklist.md` checks this at completion
- [ ] Any manual/visual verification step is labelled as such with its pass condition
- [ ] Data-setup requirements for the test are specified
- [ ] Criteria do not depend on data that does not exist yet

### 5.3 Criteria Traceability

- [ ] Each criterion traces to the requirement or user need it validates
- [ ] Every PRD requirement is covered by at least one criterion
- [ ] Changed criteria show what changed and why (version-controlled with the story)
- [ ] Criteria amended after implementation are reviewed, not silently accepted

## 6. VALUE METRICS & RELEASE DECISION

[[LLM: Value delivery is only demonstrable against a metric set fixed before the build. Metrics chosen after launch are always favourable.]]

### 6.1 Metric Definition

- [ ] Each epic has at least one named value metric
- [ ] Metrics are measurable with data this project can actually access
- [ ] Baselines are recorded before the work starts, or the baseline gap is stated
- [ ] Targets are numeric, with a date
- [ ] Counter-metrics are named where a metric can be gamed (e.g. speed vs. correctness)

### 6.2 Instrumentation

- [ ] The event or measurement that produces each metric is specified
- [ ] Instrumentation is itself a backlog item, not an assumption
- [ ] Metrics are reviewable by someone who did not build the feature
- [ ] Qualitative signal (support tickets, interviews) has a collection route

### 6.3 Release Decision

- [ ] The condition for calling the epic done-and-valuable is written down
- [ ] The condition for iterating, cutting, or rolling back is written down
- [ ] Who makes the call is named, and it is not the person who built it
- [ ] Post-release review is scheduled at first release

## 7. DEFINITION OF READY

[[LLM: A story enters an iteration only when every item below is true. Treat this as the entry gate, and report the count of stories that fail it — that count is the backlog's real state.]]

### 7.1 Story Readiness

- [ ] Story exists, is in the backlog, and is ordered
- [ ] Acceptance criteria are written and pass/fail decidable (§5)
- [ ] Dependencies are resolved or explicitly sequenced
- [ ] Design is available where the story needs one
- [ ] Open questions are answered, not parked in the story

### 7.2 Team Readiness

- [ ] Team has the skills or a plan to acquire them
- [ ] Environment and test data needed to work the story are available — validated by `sa.checklist.md` §7.4
- [ ] Anything the story needs from outside the team has a named owner and a date

### 7.3 Gate Result

- [ ] Every story entering the current iteration satisfies §7.1 and §7.2
- [ ] Stories that fail the gate are returned to the backlog, not carried as debt
- [ ] The gate result is visible to stakeholders

## 8. SCOPE CHANGE & RISK [[BROWNFIELD ONLY]]

[[LLM: For a change to an existing product, the risk is not "will it build" but "what will it break". Be pessimistic: assume the change breaks something nobody documented.]]

### 8.1 Change Impact

- [ ] Existing behaviour the change touches is identified, not assumed to be safe
- [ ] Data and API compatibility with the existing system is assessed
- [ ] Existing user workflows affected by the change are listed
- [ ] Migration path for existing data is specified, or explicitly declared unnecessary

### 8.2 Re-scope & Rollback Options

- [ ] Rollback procedure exists for the change, and its trigger is defined
- [ ] A feature-flag or staged-rollout option is available for irreversible parts
- [ ] Backup and restore procedure is confirmed with the owning team
- [ ] The change can be split so a risky part ships separately

### 8.3 User Impact

- [ ] Affected users are identified and communication is planned
- [ ] Support documentation and training material updates are backlog items
- [ ] Support is briefed before release
- [ ] A deprecation or migration path is communicated if behaviour changes

### 8.4 Risk Register

- [ ] Top risks are recorded with an owner and a mitigation
- [ ] Each risk has a trigger that would escalate it
- [ ] The risk owner is empowered to cut scope, not just to report

## 9. STAKEHOLDER ALIGNMENT

[[LLM: Disagreement found here is cheaper than disagreement in the iteration. Ask who has not been consulted — that list is usually the finding.]]

### 9.1 Stakeholder Coverage

- [ ] Every stakeholder group affected by the backlog is identified
- [ ] Each group has a feedback route into the backlog
- [ ] Groups not yet consulted are named with a date to consult them
- [ ] Decision authority for backlog changes is explicit

### 9.2 Communication

- [ ] Backlog state and near-term intent are communicated on a known cadence
- [ ] Scope changes are communicated with the reason, not just the new state
- [ ] Deferrals are communicated to whoever asked for the item
- [ ] Feedback received from stakeholders is reflected in the backlog or its rejection is explained

## 10. HANDOFF TO DELIVERY

[[LLM: The handoff is the last Product-owned gate. Whatever is unresolved here becomes the team's surprise.]]

### 10.1 Handoff Completeness

- [ ] Stories are ordered and the first increment is unambiguous
- [ ] Acceptance criteria have been reviewed with whoever will test them
- [ ] Constraints and non-negotiables are stated explicitly
- [ ] High-complexity or high-risk items are flagged for deeper technical work — `sa.checklist.md`
- [ ] Open product questions are listed with an owner and a due date

### 10.2 Handoff Quality

- [ ] A developer could start the first story without asking a product question
- [ ] Anything requiring a product decision during implementation is pre-empted
- [ ] Feedback loop back to the backlog is defined (how learnings re-enter)

## VALIDATION SUMMARY

[[LLM: Generate a validation report covering: (1) Executive Summary — vision state, backlog readiness %, Go/No-Go for the delivery horizon, blocking issues count, skipped sections; (2)
Category Analysis — per-section status PASS (90%+ items met) / PARTIAL (60-89%) / FAIL (<60%) with critical issues; (3) Backlog Health — items failing the Definition of Ready, orphans,
duplicates, and blocked items, each with a count; (4) Prioritization Assessment — framework in use, whether the order matches the scores, and items ordered against their score with the
reason; (5) Acceptance Quality — stories missing criteria, criteria that are not pass/fail decidable, and uncovered requirements; (6) Value Delivery — epic-to-metric coverage,
instrumentation gaps, release decision status; (7) [BROWNFIELD ONLY] Change Risk — top risks, rollback readiness, user impact; (8) Recommendations — must-fix / should-fix / consider.
Then offer: detailed analysis of a failed section, a Definition-of-Ready pass over the top stories, or a re-prioritization pass.]]

### Category Statuses

| Category | Status | Critical Issues |
|----------|--------|-----------------|
| 1. Vision & Product Direction | _TBD_ | |
| 2. Backlog Hygiene | _TBD_ | |
| 3. Prioritization Framework | _TBD_ | |
| 4. User Stories | _TBD_ | |
| 5. Acceptance Criteria | _TBD_ | |
| 6. Value Metrics & Release Decision | _TBD_ | |
| 7. Definition of Ready | _TBD_ | |
| 8. Scope Change & Risk (Brownfield) | _TBD_ | |
| 9. Stakeholder Alignment | _TBD_ | |
| 10. Handoff to Delivery | _TBD_ | |

### Critical Deficiencies

(To be populated during validation)

### Recommendations

(To be populated during validation)

### Final Decision

- **READY FOR DELIVERY**: The backlog is prioritized, stories are ready, and value delivery is measurable — the team can start.
- **CONDITIONAL**: The backlog is usable once the listed blockers are closed; name the horizon it applies to.
- **NOT READY**: Vision, ordering, or story readiness fails badly enough that work started now would be rework.
