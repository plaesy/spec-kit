---
name: stack-detection-fixes-2026-09-17
description: plaesy-init selective-install root-cause fixes (bash target_dir bug, jq gating) + mapping.json detection gaps for Go/Java/Rust/Dart/Next.js + new "filenames" field
metadata:
  type: project
---

# Stack detection fixes (2026-09-17)

User reported: `plaesy init` on a project containing Word/Excel/PowerPoint files didn't
install `word.md`/`excel.md`/`powerpoint.md`. Investigation expanded into a full audit
of `instructions/mapping.json` detection for every language/framework, across both
`scripts/bash/plaesy-init.sh` and `scripts/powershell/plaesy-init.ps1`.

## Root causes found (bash — the user's actual runtime)

1. **`copy_instructions()` scanned the wrong directory.** It called
   `detect-stack.sh "$target_dir"`, but inside that function `$target_dir` is the
   `.plaesy` destination folder (empty at that point), not the project root. Fixed
   to scan `"."` (the script has already `cd`'d into the project root by then).
   This was the actual bug the user hit — nothing was ever detected via
   `detect-stack.sh` in practice, silently falling through to the hardcoded
   `always_load` list.
2. **`detect-stack.sh` was gated behind `command -v jq`**, even though
   `detect-stack.sh` does its own JSON parsing and never uses jq. On any machine
   without `jq` installed, detection was skipped entirely (jq is only needed for
   the `always_load` fallback read). Fixed: only that fallback read is now gated
   on jq.
3. **`--all-instructions` flag was silently a no-op.** It set `$ALL_INSTRUCTIONS`,
   but the only code that ever checked it lived inside a `case "instructions")`
   branch in `setup_platform_config()` that can never execute (`$platform_mappings`
   never contains `"instructions"`, only `"core"`/`"prompts"`). Removed that dead
   branch (~55 lines) and added a real `$ALL_INSTRUCTIONS` check inside
   `copy_instructions()` (the function that actually runs).
4. **Dead code removed**: `create_task_system()` duplicated `create_task_structure()`
   (same 5 `mkdir -p` calls, called back-to-back); the tail of `create_structure()`
   computed `expanded_root`/`plaesy_root` and then never used it (leftover from a
   pre-refactor design where framework files were copied there).

## PowerShell (`plaesy-init.ps1`) — separate, narrower bug

`Copy-Instructions` never called `detect-stack.ps1` at all — it only ever read
`always_load` from `mapping.json`. (`Set-AISpecificConfig` *did* call
`detect-stack.ps1` and store the result in `$selectedInstructions`, but only to
print a log line — the variable was never used to copy anything; that dead code
was left as documentation of intent, not removed.) Fixed by making
`Copy-Instructions` call `detect-stack.ps1 -TargetDir "."` directly (correctly,
unlike the bash bug above) with `-AllInstructions` support mirroring bash.

## `mapping.json` detection gaps (affects both scripts — shared registry)

Systematically tested every category entry with a **minimal, realistic** manifest
(not one that happens to name the technology). Found real gaps: an entry with no
`"extensions"` field is 100% dependent on its `"keywords"` appearing literally in
scanned manifest text, which often isn't true for ordinary manifests.

| Entry | Problem | Fix |
|---|---|---|
| `dart` | `pubspec.yaml` without literal "dart"/"flutter" text wasn't detected | added `"extensions": [".dart"]` |
| `go` | Standard `go.mod`/`main.go` has no literal "golang"/"go module" | added `"extensions": [".go"]` |
| `java` | Standard `pom.xml` has no literal "java"/"jvm"/"maven"/"gradle" | added `"extensions": [".java"]` |
| `rust` | Standard `Cargo.toml`/`main.rs` has no literal "rust"/"cargo"/"rustc" | added `"extensions": [".rs"]` |
| `spring-boot` | Real `pom.xml` artifactIds are `spring-boot-starter-*` / `org.springframework.boot`, not "spring boot"/"springboot" | added keywords `"spring-boot"`, `"springframework"` |
| `nextjs` | `package.json` dependency key is bare `"next"` — too generic/prose-colliding to add as a keyword (would match "next steps" in spec/context.md) | new `"filenames"` field: `next.config.js/.mjs/.ts/.cjs` — presence-based, like `"extensions"` but exact filename instead of suffix |

**Verified NOT broken** (no fix needed): React, NestJS, Angular, React Native,
Rails — all match because their package/gem names appear as distinctive literal
substrings in real manifests (`@nestjs/core`, `@angular/core`, `react-native`,
`gem 'rails'`).

**Deliberately not "fixed"**: the `markdown` entry has no extension-based
detection on purpose — `.md`/`README.md` exist in nearly every project, so
extension-matching it would defeat the whole point of selective install. It's
meant to trigger only when a task explicitly discusses documentation/readme, via
keywords in spec/context files.

## New mechanism: `"filenames"` field

Added to both `scripts/bash/detect-stack.sh` and `scripts/powershell/detect-stack.ps1`,
implemented exactly parallel to the existing `"extensions"` mechanism (one
tree-walk total, generic — no hardcoded per-framework filename list in either
script). Use `"filenames"` instead of a keyword whenever the only reliable signal
is a specific config file's *name* (not its content), and the dependency name
itself is too generic/collision-prone for keyword matching.

**Bash gotcha discovered while implementing this**: `detect-stack.sh`'s keyword
extraction does `kw="${kw//\"/}"` — a naive strip of every `"` character, with no
JSON-unescape step. A keyword containing a literal embedded quote (e.g. the
tempting `"\"next\":"` to match `"next":` in package.json) gets silently mangled
into garbage (`\next\:`) and will never match, no error. Documented in
`mapping.json`'s `field_semantics_note`. Don't put `"` inside a `keywords` entry —
use `"filenames"` for that class of problem instead.

**PowerShell gotcha discovered while implementing this**: `Get-ChildItem -Include`
silently returns nothing when passed a `System.Collections.Generic.HashSet[string]`
object directly (no error, no exception — `$ErrorActionPreference =
"SilentlyContinue"` at the top of the script masks it further). Must pipe through
`| ForEach-Object { $_ }` first to materialize a plain array before passing to
`-Include`, exactly like the pre-existing `$allExtensions` → `$includePatterns`
conversion already did. If you add a third such "collect distinct values across
all entries" pattern in `detect-stack.ps1`, remember this conversion step or it
will silently detect nothing.

## Verification performed
Full regression matrix (office docx/xlsx/pptx, dart, go, java, spring-boot, rust,
nextjs via config file, react non-regression) re-run in **both** bash and
PowerShell after every change, plus one full `plaesy-init.sh` end-to-end run on a
synthetic Next.js project confirming `nextjs.md` + `reactjs.md` land in
`.plaesy/instructions/`.
