---
title: "Go CLI Migration"
description: "bash/PowerShell scripts replaced with a single Go binary; where things live now"
updatedAt: "2026-09-23"
---

# Go CLI Migration (2026-09-23)

**Decision**: User's target audience includes non-technical users (business,
legal, marketing, design dimensions), not just developers — a terminal-shy
audience can't be expected to run bash/PowerShell scripts comfortably, and
the dual-maintenance cost (every fix written twice, ~18k lines total) was
already a known pain point. User chose a full "force migration" (no
backward-compat shims needed — framework has no external users yet) to a
single cross-platform Go binary.

**What changed**:
- `scripts/bash/*.sh`, `scripts/powershell/*.ps1`, `scripts/bash/*.awk`, and
  their `testing/bash/`, `testing/powershell/`, `testing/smoke/*.sh|*.ps1`
  test harnesses are **deleted** (commit `41a5704`).
- Replaced by `scripts/cmd/plaesy/*.go` (cobra command definitions, one file
  per command) + `scripts/internal/*/*.go` (one package per former script:
  `common`, `featurepath`, `agentcontext`, `detectstack`, `analyze`, `graph`,
  `scaffold` (was plaesy-init), `installer` (was install.sh), `config` (was
  config-manager), `cleaner`, `trimmer`, `taskmanage`, `aiheaders`,
  `validate`, `imagegen`) — commit `ac358c2`.
- Go module root is `scripts/go.mod` (NOT repo root) — `cmd/` and `internal/`
  live directly under `scripts/`, sibling to where `scripts/bash/` and
  `scripts/powershell/` used to be. Build from repo root: `cd scripts && go
  build ./cmd/plaesy`.
- Stdlib only — no new go.mod dependencies beyond `github.com/spf13/cobra`
  (jq→encoding/json, awk/sed/grep→regexp, find→filepath.WalkDir,
  curl→net/http).
- `plaesy analyze` calls the `internal/graph` package **in-process**, not via
  subprocess — avoids the shell-out fragility the bash version had (the
  original migration attempt shelled out via `exec.LookPath("plaesy")` and
  hit "Unknown command: graph"; fixed by wiring `graph.Build`/`SaveJSON`/
  `WriteHTML`/`WriteReport` directly into `analyze.go`'s `buildGraph()`).
- Command registration uses an `init()`-based registry
  (`scripts/cmd/plaesy/registry.go`'s `register()` function) specifically so
  parallel-agent ports never conflict editing a shared `main.go` — each
  command file does `func init() { register(newXCmd()) }` and `main.go` only
  does `root.AddCommand(registry...)`. Keep this pattern for any new command.
- E2E smoke coverage: `scripts/cmd/plaesy/e2e_smoke_test.go` is a Go port of
  the deleted `testing/smoke/smoke-e2e.sh` (same fixture, same analyze+graph
  contract assertions), run via `go test ./...` from `scripts/`.
- CI (`.github/workflows/ci.yml`) is now one `go build`/`vet`/`test` matrix
  job across ubuntu/windows/macos — replaced 7 bash-lint/shellcheck/
  powershell-parity jobs.
- `Makefile`'s `reload`/`detect` targets now `go build` the binary first,
  then invoke it (`./plaesy init`, `./plaesy config detect-platform`).
- `README.md` install instructions now say "build from source" (`go build`)
  — there is no published release binary yet, so no curl-a-prebuilt-binary
  instructions were added (would be fabricated).

**Still open / not yet done**:
- `scripts/internal/installer`'s `repair`/`upgrade` subcommands are
  deliberately stubbed ("not yet implemented") — a real self-update needs a
  release/distribution pipeline that doesn't exist yet.
- No GitHub Releases / prebuilt binaries published — anyone installing today
  must build from source.
- Doc cleanup pass (this session, still in progress as of this checkpoint):
  updating `docs/scripts/*.md`, `instructions/*.instructions.md`,
  `prompts/*.md`, `templates/*.md`, `CONTRIBUTING.md`,
  `.github/ISSUE_TEMPLATE/bug_report.md`, `docs/checklists/README.md` to
  reference `plaesy <command>` instead of the deleted script paths. This
  cleanup is now COMPLETE.

**Post-migration bug fixes (2026-09-23, commit `2962edb`)**:
- **Windows path resolution**: `GetRepoRoot()` used hardcoded `/` in
  `filepath.Join`, breaking Windows. Fixed by routing all path joins through
  `filepath.Join` (platform-native) + `filepath.ToSlash` for display only.
- **PowerShell install script**: `scripts/install.ps1` created for Windows
  users, mirroring the existing `scripts/install.sh`.
- **README.md**: stale `.sh`/`.ps1` references updated; both bash and
  PowerShell install instructions documented.

**Superseded guidance** (previously in `.plaesy/memory.md`, now WRONG — see
that file's Established Rules section, corrected in this same checkpoint):
- "Technology detection... must go through `scripts/bash/detect-stack.sh`/
  `.ps1`" → now `scripts/internal/detectstack` (Go), invoked as `plaesy
  detect-stack`.
- "Scripts → `.plaesy/scripts/{powershell,bash}/`" (post-init path) → that
  per-project copy mechanism is gone; a project no longer gets a
  `.plaesy/scripts/` copy of anything, since there's one binary, not
  per-project script copies.
- Any prior rule assuming bash+PowerShell must be kept in parity no longer
  applies — there is exactly one implementation now.
