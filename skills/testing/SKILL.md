---
name: testing
description: "Testing discipline: test-driven development (red-green-refactor, unit tests, a commit per step) and the testing strategy (what to test, golden master and characterization tests, coverage). Use when writing tests, adding a regression test, planning test coverage, or validating a change."
---

# Testing

| File | Load when |
|------|-----------|
| `references/tdd.md` | Writing code test-first: red-green-refactor, commit patterns, enforcement hooks |
| `references/strategy.md` | Choosing what to test and at which level, for any project type |

Decide the strategy with `references/strategy.md`, then work each change through `references/tdd.md`.

## Gotchas

- `FAIL <pkg> [build failed]` means zero tests ran, not that a test failed. `go test` runs vet first, so a vet error suppresses every test in the package. Re-run with `-vet=off` to see real results.
- `go test ... | tail` hides the exit code of `go test`. Check the exit status of the test command itself.
- Do not run `go test ./...` without explicit permission. Scope runs to the changed packages.
- A red suite is a finding. Read each failure before you call it pre-existing; `git status` alone does not prove that.
- `os.Getwd()` after `os.Chdir()` is not always the string you passed. Compare resolved paths in tests (macOS `/var` is a symlink to `/private/var`).
- Without a `.gitattributes`, line endings follow `core.autocrlf`. Regexes that assume `\n` break on CRLF files.
- `filepath.ToSlash` converts OS separators to `/`. It does not normalize backslashes in arbitrary strings.
- A coverage figure collected with `-vet=off` is complete, not partial. Do not discount it.
- Keep network steps behind a seam. A test that hits the real network depends on the machine it runs on.
