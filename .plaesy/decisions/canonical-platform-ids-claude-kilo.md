---
title: "Canonical platform ids: claude and kilo, not claude_code/kilo_code"
updatedAt: "2026-10-02T21:50:00.000Z"
---

# Canonical platform ids: `claude` and `kilo`

**Decision**: `platform.json`'s keys for these two platforms are now `claude`
and `kilo` (previously `claude_code` and `kilo_code`). The long forms stay
registered as aliases in `scripts/internal/config/config.go`'s
`platformAliases` map (`"claude_code": "claude"`, `"kilo_code": "kilo"`) for
backward compatibility — an existing config, script, or person's muscle memory
that types the old form keeps resolving, it just no longer appears as the
canonical spelling anywhere new.

**Why**: The user wanted `plaesy init --ai claude` to work without having to
remember or type `claude_code`. It already did, via the alias table — but the
docs, `--ai` flag help text, example commands, and `platform.json` itself all
said `claude_code`, which is confusing when the shorthand is the one anyone
actually types. Rather than just fix the docs to recommend the shorthand while
the config disagreed, the canonical id itself was renamed to match.

**Scope of the rename**: `scripts/configs/platform.json` (and its generated
mirrors: `.plaesy/scripts/configs/platform.json`,
`scripts/internal/assets/data/scripts/configs/platform.json`), the alias
table and its comments in `scripts/internal/config/config.go`, help text in
`cmd/plaesy/{init,reload,platforms,clean}.go`, ~30 test assertions across
`scripts/internal/config/config_test.go`,
`scripts/internal/scaffold/scaffold_test.go`,
`scripts/cmd/plaesy/{clean,config,platforms,reload}_cmd_test.go`, and docs
(`README.md`, `docs/overview.md`, `docs/reference.md`,
`docs/scripts/{plaesy-init,plaesy-clean,platforms,README,config-manager}.md`).
CHANGELOG.md and `.plaesy/tasks/done/` entries were deliberately left alone —
they document what a *past* version of the tool did, and rewriting them to
match current spelling would be falsifying the historical record.

**A subtle bug the rename surfaced**: `scaffold.normalizePlatform()` used to
detect "the alias table actually did something" by checking whether its
output differed from the lowercased input. Once `claude`'s own alias entry
became `"claude": "claude"` (an identity mapping, because the canonical id now
equals its own lowercased shorthand), that heuristic broke for case variants
like `"CLAUDE"` — lowering it produces `"claude"`, which now equals the
alias-resolved output, so the function concluded "nothing happened" and fell
through to a case-sensitive display-name match that `"CLAUDE"` cannot win.
Fixed by adding `config.IsKnownAlias()`, which asks the alias table directly
rather than inferring a hit from whether the spelling changed. See
`scripts/internal/scaffold/init.go` and `scripts/internal/config/config.go`.

**Not done**: `cmd/plaesy/platforms.go`'s and `cmd/plaesy/clean.go`'s inline
comments referencing the old convention were updated; `mirror_parity_test.go`
was separately generalized in the same session to read every platform's
prompt-mirror destination from `platform.json` instead of hardcoding
`.kilo/commands/` — a related but independent fix (the hardcoding predates
this rename and would have been a bug regardless of which id was canonical).
