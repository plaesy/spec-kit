---
description: "Usage guide for plaesy-trim — command-output and memory-file token compression. Load when compressing instruction/memory files, wrapping noisy shell commands, or investigating high token spend."
---

# Plaesy Trim — Token & Context Compression

## Purpose

Plaesy Trim (available at `.plaesy/scripts/powershell/plaesy-trim.ps1` for Windows/PowerShell and `.plaesy/scripts/bash/plaesy-trim.sh` for Linux/macOS/Git Bash/WSL) cuts token spend at the three points it actually accrues:

1. **Layer 1 — command output**: `run <command...>` executes a shell command and compresses its output
   before it reaches context, per-tool aware where it matters (git, npm, cargo, pytest, dotnet, docker
   and others via `.plaesy/scripts/configs/plaesy-trim-rules.json`). Commands with `"mode": "test-results"`
   (pytest, cargo test, npm test, dotnet test) get filtered to failures + summary only, not just
   truncated — a passing 500-line test run collapses to one line instead of a generic head/tail cut.
2. **Layer 2 — memory/instruction files**, two modes:
   - `compress -Path <file|dir>` — fast, local, heuristic (filler-phrase stripping + whitespace
     collapsing). Saved once, re-read every session after. Compression level is chosen per file from
      its `plaesy-graph` in-degree (`degree` field in `.plaesy/memory/analysis/project.graph.json`) when
     `-Level`/`--level` is not given: frequently-referenced files ("god nodes") get the lightest touch
     (`lite`), rarely-referenced files get compressed harder (`ultra`).
   - `llm-queue` / `apply-llm` — LLM-quality rewrite done by the calling assistant itself, no API key,
     same no-network pattern as `plaesy-graph`'s `-SemanticQueue`/`-ApplySemantic`. Denser and more
     semantically faithful than the heuristic mode at the cost of a manual round-trip; use it when the
     heuristic mode's savings are too small to bother with.
3. **Layer 3 — generated output discipline**: not a script, a rule. See "Token & Output Efficiency
   Protocol" in `.plaesy/memory/plaesy.md` (decision ladder: skip → reuse → stdlib → native →
   one-liner → minimal solution) — applies to every code-writing task, always loaded.

## What is never touched

Code blocks (fenced ``` regions), shell commands, error messages/stack traces, and structured data
(JSON/YAML/JS objects) are always preserved byte-exact in both layers. Only narrative prose is compressed.

## Which script to run

Same rule as `plaesy-graph`: PowerShell shell → `.ps1`, bash/zsh/sh shell → `.sh`. Don't ask the user to
choose; infer from the current tool-execution shell.

```powershell
# PowerShell
.plaesy/scripts/powershell/plaesy-trim.ps1 run git status                          # Layer 1: compress one command's output
.plaesy/scripts/powershell/plaesy-trim.ps1 compress -Path .plaesy/memory.md          # Layer 2: compress one file
.plaesy/scripts/powershell/plaesy-trim.ps1 compress -Path .plaesy/memory -Recurse            # Layer 2: compress memory folder
.plaesy/scripts/powershell/plaesy-trim.ps1 compress -Path .plaesy/memory -DryRun             # preview only, no write
.plaesy/scripts/powershell/plaesy-trim.ps1 compress -Path README.md -Level ultra  # force a level
.plaesy/scripts/powershell/plaesy-trim.ps1 llm-queue -Path .plaesy/memory/plaesy.md   # export prose to rewrite
.plaesy/scripts/powershell/plaesy-trim.ps1 apply-llm -Path .plaesy/memory/plaesy.md -Annotations ann.json  # merge back
.plaesy/scripts/powershell/plaesy-trim.ps1 report                                  # cumulative savings across layers
```

```bash
# Bash (identical behavior/output shape)
.plaesy/scripts/bash/plaesy-trim.sh run git status
.plaesy/scripts/bash/plaesy-trim.sh compress --path .plaesy/memory/plaesy.md
.plaesy/scripts/bash/plaesy-trim.sh compress --path .plaesy/memory --recurse
.plaesy/scripts/bash/plaesy-trim.sh compress --path .plaesy/memory --dry-run
.plaesy/scripts/bash/plaesy-trim.sh compress --path README.md --level ultra
.plaesy/scripts/bash/plaesy-trim.sh llm-queue --path .plaesy/memory/plaesy.md
.plaesy/scripts/bash/plaesy-trim.sh apply-llm --path .plaesy/memory/plaesy.md --annotations ann.json
.plaesy/scripts/bash/plaesy-trim.sh report
```

## Compression levels (Layer 2)

- `lite` — strip filler phrases (`.plaesy/scripts/configs/plaesy-trim-rules.json` → `layer2_filler_patterns`), collapse repeated blank lines
- `full` (default when degree is mid-range) — `lite` + merge short redundant sentences, trim hedging clauses
- `ultra` — `full` + telegram-style fragments; use only for low-traffic files, readability drops

Every `compress` run backs up the original as `<file>.bak` before writing, and appends a record to
`.plaesy/memory/token-stats.json`. `report` reads that log — no re-scanning needed.

## Layer 1 rule format

`.plaesy/scripts/configs/plaesy-trim-rules.json` → `layer1_commands` maps a command prefix (e.g. `"git status"`)
to `{ max_lines, dedupe_consecutive, mode }`. Unmatched commands fall back to `_default`.
- `mode` omitted or `"generic"` — consecutive identical lines collapse to `line (xN)`; output beyond
  `max_lines` keeps the first/last few lines with a `... N lines omitted ...` marker.
- `mode: "test-results"` — keeps only lines matching fail/error markers plus the trailing summary line
  (e.g. `X failed, Y passed`); an all-passing run collapses to a single confirmation line.

Still a per-tool-*aware* line-shape compressor, not a full AST/output parser — extend `layer1_commands`
with a new `mode` (and matching branch in `compress_command_output`/`Compress-CommandOutput`) if a
command's output needs finer handling than dedupe/truncate/filter.

## LLM-mode round-trip (Layer 2)

1. `llm-queue -Path <file>` splits the file on fenced code blocks (never touching code) and writes every
   prose segment ≥40 chars to `.plaesy/memory/analysis/trim-queue.json` as `{file, segments:[{index, text}]}`.
2. Hand that file to the calling assistant (this session) with: "rewrite each segment.text denser, same
   meaning, don't touch anything that looks like code/commands/paths — return the same `{index, text}`
   shape as JSON."
3. Save the reply as `annotations.json`, then `apply-llm -Path <file> -Annotations annotations.json` —
   segment indices must match the original file's fence-toggle split exactly (don't re-run `llm-queue`
   against a since-modified file and apply against the old queue, or indices will misalign).

## Notes

- Layer 2's degree-based level selection requires `.plaesy/memory/analysis/project.graph.json` to exist —
  run `graph` first (see `.plaesy/memory/plaesy-graph.md`); if missing, `compress`
  falls back to `full` for every file.
- `run` does not install a permanent shell hook — it's an explicit per-invocation wrapper, so nothing is
  silently rewritten. Wrap it in an alias/function in your own shell profile if you want it default.
  Per `.plaesy/memory/plaesy.md`'s Token & Output Efficiency Protocol, the assistant itself
  is expected to call `run` by default for noisy commands, without the user asking each time.
- All processing is local — no network calls, no telemetry, matches the rest of this repo's tooling.
