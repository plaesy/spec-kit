---
title: "Session 2026-10-20: installer PATH, uninstall cleanup, Go toolchain pitfalls"
updatedAt: "2026-10-20T00:00:00.000Z"
---

# Session 2026-10-20: installers and PATH

## What changed (uncommitted)

- `install.ps1` adds `%LOCALAPPDATA%\Plaesy\bin` to the user PATH after `plaesy install`.
- `install.sh` appends a marked `export PATH` block to `~/.bashrc`, `~/.bash_profile` (macOS) or `~/.zshrc`.
- `plaesy uninstall` removes that PATH entry again (HKCU registry on Windows; marked block on Unix).
- Decision: [installers-add-binary-dir-to-path](../decisions/installers-add-binary-dir-to-path.md). It reverses the earlier "never edit PATH" rule.

## Learnings

- **Stale binaries shadow new ones.** Three `plaesy.exe` copies existed (`C:\Users\User\bin`, `AppData\Local\Plaesy\bin`, repo-local). The PATH order picked the oldest. Check `where plaesy` and `plaesy validate --list` before trusting a validator message from another project.
- **Go toolchain mismatch.** `go` on PATH was 1.27.0 while GOROOT pointed at 1.27.2. `go test` failed with `version "go1.27.2" does not match go tool version`. Working fix: call the GOROOT binary directly (`...\mise\installs\go\1.27.2\bin\go.exe`).
- **`go fmt` rewrites whole packages.** It reformatted 68 unrelated files in `scripts/cmd/plaesy/`. Use `gofmt -l` (read only) and format only the files you changed.
- **`git checkout --` is not a safe undo here.** It reverted hand edits along with the formatter damage. Restore only paths you have listed and checked.
- **Stderr inside captured lists.** Capturing `git diff --name-only 2>&1` put git warnings into a pathspec list. Keep stderr out of path lists.
- **Test isolation for PATH.** Uninstall tests stub the PATH remover (`removeFromPath`) so they never edit the real user PATH.

## Open

- Nothing committed. The owner has not approved commits (see `.plaesy/context.md`).
- Windows registry write in `removeFromUserPath` is not exercised end to end. Only the pure logic is unit-tested.
- `plaesy uninstall` does not clean a leftover PATH entry when both the binary and the Plaesy home were already deleted by hand.
- GitHub Actions cost question answered from third-party sources; not recorded as a decision.
