---
name: Git and Session State
description: What .plaesy tracks, which platform ids are canonical, how to publish
updatedAt: "2026-10-02T21:50:00.000Z"
---

# Git and Session State

- **`.plaesy/` holds two kinds of file and needs both tracked and ignored.**
  `instructions/`, `roles/`, `templates/`, `analysis/`, `scripts/` are generated;
  `context.md`, `memory.md`, `state.json` and `tasks/` are session state a
  collaborator needs. `~/.gitignore_global` excludes the whole directory, so
  `plaesy tasks move` silently dropped a task file out of version control on its
  first move — the tracked-binary bug from 2026-09-28 in a different costume.
  `.gitignore` re-includes with `!.plaesy/`, which **must precede** the
  per-directory re-ignores: git cannot re-include a file whose parent directory
  is excluded, and a reversed negation is silently inert. Guarded by
  `TestSessionStatePathsAreTrackable`.
- **`platform.json` keys are the canonical ids, and as of 2026-10-02 `claude`
  and `kilo` ARE the canonical ids** — not `claude_code`/`kilo_code`. That
  `*_suffix` convention is what this bullet said before the rename; it is
  stale. The long forms (`claude_code`, `kilo_code`) still resolve via the
  alias table in `scripts/internal/config/config.go` (kept for backward
  compatibility — a config, script, or muscle memory that types the old form
  still works), but the canonical spelling to write into `platform.json`, docs,
  or new code is the short one. See `[[canonical-platform-ids-claude-kilo]]`
  for the full decision record. Other platforms keep their existing
  `*_suffix`/`*_ai` ids (`cursor_ai`, `github_copilot`, `trae_ai`,
  `windsurf_ai`, …) — only claude/kilo changed.
- **Generating and documenting are different jobs.** A doc produced by the tool
  (`docs/`, `CHANGELOG.md`, `metadata.json` counts) drifts silently. Its
  `[Unreleased]` section was empty while shipping 19 documented commands — the
  exact failure its own guard exists to catch — and `goTestFiles`/`goTestFunctions`
  were stale. Treat a generated doc as a build product: regenerate, never
  hand-edit and assume.
- **Publishing is resolved: `origin/main` is a single-commit history, CI is
  green.** The 2026-09-30 unrelated-histories problem was never merged through —
  it was reset. 2026-10-02's session ran `make reset` (deletes every GitHub
  Release/tag, wipes local `.git`, recommits everything as one "Initial
  Commit", force-pushes, re-tags `v<VERSION>`) twice during active development,
  which is why `main`'s history is one commit, not a merge of the old and new
  trees. After that point, every further change was a normal commit + push —
  no more resets. Do not reach for `make reset` to resolve a future divergence
  without the owner's explicit sign-off: it is irreversible on the remote
  (deletes real GitHub Releases and their uploaded assets).

## Platform-agnostic paths (after `plaesy init`)

- Instructions → `.plaesy/instructions/[name].md` (flat)
- Agents → `.plaesy/roles/[name].md` (flat)
- Scripts → `.plaesy/scripts/{powershell,bash}/`
- Configs → `.plaesy/scripts/configs/`
- Analysis → `.plaesy/analysis/`
- Tasks → `.plaesy/tasks/{backlog,todo,doing,done,blocked}/`
- Knowledge → `.plaesy/memory/[topic].md` (flat)
