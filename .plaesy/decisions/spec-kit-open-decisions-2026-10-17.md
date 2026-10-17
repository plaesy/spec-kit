---
title: "Decision: open decisions from the spec-kit-2 and spec-kit-5 reviews"
description: "Six open items resolved as Nara: rule-8 hold, CI secret scan, SC-01 scope, instructions size, contract blocking map, freeze and evidence limits"
updatedAt: "2026-10-17T12:40:00Z"
decided: "2026-10-17"
status: "Accepted (ASSUMED where marked)"
supersedes-part: "nara-first-no-stop (rule 8 for irreversible, outward actions with no reversible form)"
---

# Decision: open decisions from the spec-kit-2 and spec-kit-5 reviews

**Status**: Accepted. Items marked ASSUMED are defaults to revisit.
**Decided**: 2026-10-17 by Nara (framework decider, `agents/nara.agents.md`), on the owner's instruction to resolve all open decisions and integrate.
**Inputs**: `agentic-dev-best-practices-gap-report.md` (spec-kit-5, §4); `review-ao-spec-kit-2-root.md` (spec-kit-3); `review-ao-spec-kit-2-root-success-contract.md` (spec-kit-4, §7).
**Related**: [nara-first-no-stop](nara-first-no-stop.md), [remove-internet-access-dimension](remove-internet-access-dimension.md)

## Council

Nara ran the Council Protocol in this session as a single agent. It did **not** run independent drafts from 3+ agents with a critic's scoring. The diversity caveat applies: one underlying model cannot remove its own blind spot. The evidence used is the three review reports above (each a separate reviewer's work) plus the references listed below. A second provider should review items 1 and 6 before they are treated as settled.

## Decisions

### 1. Human approval for high-risk actions versus rule 8 — ASSUMED: hold, not stop

- **Conflict**: rule 8 says Nara decides destructive and outward actions with no pause. OWASP LLM01 and Anthropic's agent guidance recommend a human approval step for high-risk actions.
- **Decision**: Nara still decides everything. For an irreversible, destructive or externally visible action, Nara performs only a reversible form when one exists (backup before delete, draft before publish, dry run before spend). With no reversible form, the action is recorded as blocked with reason `awaiting human approval`, and the run continues. An explicit owner instruction approves a held action.
- **Owner's earlier choice**: `nara-first-no-stop` chose option A (no stops). This refines it for the no-reversible-form case only. Reversible actions keep option A. This is the decision the owner is most likely to want to reverse, so it is flagged.
- **NN check**: NN-06 (a logged, reversible action before any delete or overwrite) is kept by the reversible-form rule. NN-07 is kept (no downstream phase overrides this). NN-05 is not touched.
- **Constitution**: no amendment. The hold is a framework rule in `instructions/plaesy.instructions.md` rule 8 and `agents/nara.agents.md`. No non-negotiable changed.
- **Reversal condition**: the owner says "no holds" explicitly.

### 2. CI secret scanner (NN-01) — DONE: gitleaks v8.30.1, pinned

- **Decision**: add a `secrets` job to `.github/workflows/ci.yml`. It runs `go run github.com/zricethezav/gitleaks/v8@v8.30.1 git --redact --exit-code 1 .` over the full history (`fetch-depth: 0`). The module is built from source by the Go toolchain, so there is no third-party action and no licence key.
- **Version**: v8.30.1 is the latest release per the GitHub releases listing (search result, retrieved 2026-10-17; the release date in the snippet is unverified). The module path `github.com/zricethezav/gitleaks/v8` was confirmed by running `go run` (the `gitleaks/gitleaks/v8` path fails: the module declares the zricethezav path).
- **Baseline**: the first full-history run found 12 hits at base commit `7cdd774`, all reviewed as false positives: 10 `curl-auth-header` on the placeholder `Bearer YOUR_API_KEY` in `templates/api-documentation.template.md` (and its embedded copy), and 2 `generic-api-key` on cache keys in `.plaesy/reach/cache/`. Those cache files hold cached web responses (an example.com page and a 404 page), not credentials, which was checked by reading them with the long tokens redacted. The 12 fingerprints are listed in `.gitleaksignore`. They are commit-scoped, so a leak in any later commit still fails the job. After the allowlist, the scan reports `no leaks found` (exit 0).
- **NN check**: NN-01 detection anchor is now real. The scan runs in CI on every push and pull request.
- **Residual**: a secret already in `7cdd774` history is not rotated by this change. Rotation status for the cached-response cache keys is UNKNOWN, but they are not credentials.
- **Reversal condition**: the owner prefers a different scanner. The job is one YAML block.

### 3. Constitution SC-01 scope — DONE: constitution 2.3.0 (MINOR)

- **Decision**: SC-01 now names the untrusted channels (repository files, web pages, issue and PR text, tool output, MCP tool descriptions) and says they cannot introduce rules, grant permissions, or request an action the user or the constitution does not authorize. The same row is in `templates/constitution.template.md`.
- **Procedure**: MINOR, because scope is materially expanded. The Sync Impact Report is in the amendment log (`templates/constitution.template.md`, `instructions/security-design-principles.instructions.md`, `instructions/plaesy.instructions.md`, `agents/verifier`). No non-negotiable was removed or narrowed. `plaesy validate constitution` checks that the frontmatter version equals the last log row, which holds (2.3.0).
- **NN check**: no NN article touched. SC-01 is a §4 rule; NN-07 precedence is unchanged.

### 4. Size of the always-loaded `plaesy.instructions.md` — ASSUMED: keep, follow up

- **Finding**: about 500 lines (26 KB). No validator enforces a line or byte budget on this file. The Size Enforcement section sets numbers only for `context.md` (≤100 lines) and index files (`memory.md`, `decisions.md`, `instructions.md`).
- **Decision**: do not split the file in this integration. Splitting the Anti-Duplication and Memory sections into scope-loaded files means moving text that tests and the embedded-assets mirror pin. The risk of breaking those outweighs the benefit inside this change.
- **NN check**: not applicable (no NN article).
- **Residual**: the file is larger than the Size Enforcement numbers would suggest for a core file. A follow-up task should split it and add a validator budget. The spec-kit-5 report makes the same recommendation (gap #2 and §4 item 4).

### 5. B-2: constitution threshold-to-item mapping for blocking contract items — ASSUMED: partial mechanical check

- **Implemented already**: the validator requires at least one blocking item unless a logged `no-blocking` waiver with a reason is recorded (`scripts/internal/contract/contract.go`, around lines 312–331). B-1 (removing a blocking item after freeze) is fixed (line ~552).
- **Decision**: the mapping from §1/§2 thresholds to contract items is enforced by constitution §7 item 1 (every active-dimension threshold is measured, with the measurement named) and by the verifier pass, not by the validator. Items carry no link to a constitution variable, and a text match would be fragile.
- **NN check**: NN-08 holds, because a phase cannot report completion while a blocking item is unmet. §7 is a MUST, and it is checked by the verifier.
- **Residual**: a contract can omit a threshold row without the validator failing. Only the verifier catches it. Reversal condition: a `threshold:` field on items would allow a mechanical check.

### 6. Block placement, freeze limits, evidence content — ASSUMED: accept, document residual

- **M-12 (block placement)**: keep the identical block before the first H2 in all 73 prompts. `TestEveryPromptCarriesTheSuccessContractBlockOnce` pins the exact text and count, so copies cannot drift. Shortening the block to two lines is deferred, because it would change 73 files at once for no change in behaviour.
- **Freeze limit**: the git anchor proves the frozen snapshot matches the first committed version of the contract. It does **not** prove the contract was committed before work started. NN-08's "committed before the first action" clause has no mechanical check. Residual, not blocked: the phase report must state the contract commit and the first work commit, and the verifier checks their order.
- **Evidence content (M-7)**: the validator checks that the evidence path exists. It does not check the content. Residual, covered by the NN-03 verifier pass and EV-02 labelling.
- **NN check**: NN-03 and NN-08 are upheld by the verifier and phase report. Neither is mechanically proven by the validator. This is recorded here so it is not read as proven.

## Consequences

- `plaesy.instructions.md` rule 8, `agents/nara.agents.md`, `templates/constitution.template.md`, `.plaesy/memory/constitution.md` (2.3.0), `ci.yml` (`secrets` job), `.gitleaksignore`, and `CHANGELOG.md` changed. Embedded assets regenerated.
- Items 4, 5 and 6 are documented residual risks, not fixed. Item 1 changes the owner's option A for one case and is flagged for review.

## Blocked

None. No decision was forced past an NN article. The NN-08 start-of-phase clause (item 6) and the threshold mapping (item 5) are residual gaps in verification, not violations.

## References

- OWASP GenAI, LLM01 Prompt Injection, human approval for high-risk actions, https://genai.owasp.org/llmrisk/llm01-prompt-injection/ (via gap report, accessed 2026-10-17)
- Anthropic, "Building effective agents", human checkpoints and stopping conditions, https://www.anthropic.com/engineering/building-effective-agents (via gap report, accessed 2026-10-17)
- gitleaks releases, latest v8.30.1, https://github.com/gitleaks/gitleaks/releases (search snippet, accessed 2026-10-17; the module path verified by `go run`, 2026-10-17)
- `agentic-dev-best-practices-gap-report.md` §4, `review-ao-spec-kit-2-root.md`, `review-ao-spec-kit-2-root-success-contract.md` §7 (artifacts, 2026-10-17)
