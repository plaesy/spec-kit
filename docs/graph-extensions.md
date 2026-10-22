# Graph file extensions

`plaesy analyze` collects a file into the project graph when its extension is in
`scripts/internal/graph/extensions.go`.

| List | Extensions | Node type |
|------|-----------|-----------|
| `SourceExtensions` | `.js .jsx .ts .tsx .py .go .dart .java .kt .kts .swift .c .h .cc .cpp .hpp .cxx .hh .cs .rs .rb .php` | `source` |
| Non-source (collected) | `.md .sh .ps1` | `other` (unless the directory decides, see below) |

A file's node type follows its top-level directory first (`docs` → `doc`,
`scripts` → `script`, `templates` → `template`, and so on). Only then does the
extension decide: a file in `SourceExtensions` is `source`, and everything else is
`other`.

Every entry in `SourceExtensions` needs a case in `symbolsOf`
(`scripts/internal/graph/symbols.go`), or its files are collected with no symbols.

Decision record: `.plaesy/decisions/graph-extension-consolidation.md`.

## What analyze leaves out

Every scan (the graph, `plaesy analyze`'s structure and fingerprint walks, the
document index, and the assumptions validator) uses the same rules, defined in
`scripts/internal/exclude/exclude.go`:

- **Directories by name:** any dot-directory (`.git`, `.plaesy`, `.dart_tool`,
  `.venv`, ...) and `node_modules`, `vendor`, `dist`, `build`, `target`,
  `__pycache__`, `venv`, `Pods`, `DerivedData`, `coverage`.
- **Flutter tooling by location:** `flutter/ephemeral` under a platform folder
  (`ios/Flutter/ephemeral`, `linux/flutter/ephemeral`, ...).
- **The project's root `.gitignore`:** read on every run. Supported syntax is the
  common subset: `#` comments, `!` negation, a trailing `/` for directories,
  a leading or inner `/` to anchor a pattern, `*`, `?`, `[...]`, and `**`. Nested
  `.gitignore` files are not read.

`bin/`, `obj/`, `out/` and `env/` are not excluded by name: each is a common
source folder in some projects. Add them to `.gitignore` if they hold output.
