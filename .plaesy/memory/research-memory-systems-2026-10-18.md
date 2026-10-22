---
title: "Research: Memory Systems Landscape"
description: "Nine external agent-memory repos checked 2026-10-18, what to adopt, what to skip, framework gaps"
updatedAt: "2026-10-18T00:00:00.000Z"
---

# Research: Memory Systems Landscape (2026-10-18)

Source summaries were produced by WebFetch (model summaries, not code audits). Treat star counts and
benchmark claims as self-reported. Accessed 2026-10-18 for every entry.

## Repos checked

| Repo | What it is | Verdict for plaesy |
|---|---|---|
| [tobi/qmd](https://github.com/tobi/qmd) | Local BM25 + vector + LLM rerank search over markdown; MCP server | Adopt the idea: search over memory and decisions. Retrieval gap. |
| [ottosulin/hwiki](https://github.com/ottosulin/hwiki) | Markdown memory for Hermes; `raw/` never indexed; secret guard; append-only staged writes | Adopt: raw vs curated split, secret guard on writes. Small repo (2 stars). |
| [nashsu/llm_wiki](https://github.com/nashsu/llm_wiki) | Desktop LLM wiki; `purpose.md` (why) split from `schema.md` (how); per-page `sources:` | Adopt the pattern: intent vs structure, provenance per page. |
| [nashsu/llm_wiki_skill](https://github.com/nashsu/llm_wiki_skill) | Read-only skill over llm_wiki HTTP API | Skip. Wrapper only. |
| [writerslogic/holographic-memory](https://github.com/writerslogic/holographic-memory) (HMS) | Local hypervector (VSA) memory, Rust core | **Skip.** See [hms-not-adopted](../decisions/hms-not-adopted.md). |
| [wysie/mnemosyne-dashboard](https://github.com/wysie/mnemosyne-dashboard) | Read-only dashboard over Hermes memory DB | Adopt the pattern: supersede or expire, never hard delete; audit log. |
| [tuancookiez-hub/HyAtlas-Memory](https://github.com/tuancookiez-hub/HyAtlas-Memory) | Go memory server; supersession with recorded reason; per-turn extraction vs cross-memory consolidation | Adopt: supersession fields, consolidation split. |
| [eugeniughelbur/obsidian-second-brain](https://github.com/eugeniughelbur/obsidian-second-brain) | 48-command markdown vault skill; abstain on low confidence; human gate for learned rules | Adopt: abstain rule and human gate. Skip the command sprawl. |
| [Hermes curator](https://hermes-agent.nousresearch.com/docs/user-guide/features/curator) | Lifecycle for agent-created skills: active, then stale at 14 days, then archived at 30 days; never auto-delete; ledger and snapshots | **Most relevant.** Framework has no memory lifecycle. |

## Framework gaps (checked locally 2026-10-18)

- Memory only grows. `session-2026-10-10` through `session-2026-10-17` are never archived. `memory.md` index is 98 lines.
- No supersession. Replaced decisions stay active in `decisions.md`.
- Frontmatter is inconsistent: some files use `name`, others `title`, `constitution.md` uses `version` and `ratified`.
- `plaesy search` indexes code symbols only. Memory and decisions have no search.
- Provenance is required for decisions but not for memory files.
- No raw vs curated split. Web research can land directly in `memory/`.
- No secret guard on memory writes.
- No ledger for memory curation. Git exists but batches are not recorded.

## Proposed order (not yet decided)

1. Memory lifecycle with dry-run curation (curator pattern). Archive, never delete.
2. `superseded_by` and `supersedes` on decisions and memory (HyAtlas pattern).
3. Standard frontmatter plus `plaesy validate memory`.
4. `.plaesy/raw/` for unreviewed research, excluded from automatic context.
5. Search over memory and decisions (qmd pattern; BM25 first).
6. Secret guard on memory writes (hwiki pattern).
