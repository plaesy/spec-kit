---
name: release-native-matrix-cgo
description: "Release workflow switched from single-runner CGO_ENABLED=0 cross-compile to a native per-OS/arch matrix with cgo enabled, because the search embedder's onnxruntime_go dependency is cgo-only"
metadata:
  type: decision
  updatedAt: "2026-10-10T22:10:00.000Z"
---

# Release: native matrix instead of cross-compiled CGO_ENABLED=0

## Decision

`.github/workflows/release.yml` now builds each of the 5 release targets
natively, on a runner whose OS/arch matches the target, with cgo enabled
(the default for a native build) — instead of cross-compiling all 5 from one
`ubuntu-latest` runner with `CGO_ENABLED=0`.

| Target | Runner |
|---|---|
| linux/amd64 | `ubuntu-latest` |
| linux/arm64 | `ubuntu-24.04-arm` (GitHub-hosted native arm64) |
| darwin/amd64 | `macos-13` (last Intel-hosted runner) |
| darwin/arm64 | `macos-latest` (Apple Silicon) |
| windows/amd64 | `windows-latest` |

A `build` job (matrix, 5 parallel runs) each uploads its binary as an
artifact; a `release` job downloads all 5 and creates the GitHub Release —
the same two-job split any multi-platform artifact collection needs, since a
cross-compile loop on one runner can no longer produce all 5 files itself.

## Why

Both the Release workflow and, independently, CI's `go build` started
failing. Three root causes, found in sequence — each looked sufficient until
the next `go test`/CI run disproved it:

1. **Wrong, then corrected**: first assumed `go.mod` pinning
   `onnxruntime_go` at v1.24.0 while the code targeted v1.36.0's API was the
   whole story. Bumped to v1.36.0. This did not fix anything — see (3).
2. **This decision**: the Release workflow kept failing with
   `onnxruntime_go: build constraints exclude all Go files`. Every non-test
   file in that package (`onnxruntime_go.go`, `legacy_code.go`,
   `setup_env{,_windows}.go`, `tensor_type_constraints.go`) does
   `import "C"` unconditionally — there is no non-cgo build-tag fallback.
   `CGO_ENABLED=0` therefore excludes every file in the package, for every
   OS/arch, independent of the version pinned. This can't be fixed by
   changing the Go code on our side; it's a property of the dependency.
3. **The real blocker, found after switching to the native matrix**: the
   matrix run (on `onnxruntime_go` v1.36.0, with cgo enabled) still failed
   with the exact same compile errors as before the version bump —
   `unknown field GraphOptimizationLevel`, `cannot use ... as []Value`, etc.
   `onnx.go`'s entire API usage was wrong, not version-mismatched:
   `SessionOptions` in both v1.24.0 and v1.36.0 is an **opaque struct**
   configured via `NewSessionOptions()` + setter methods
   (`SetGraphOptimizationLevel`, `SetExecutionMode`,
   `SetIntraOpNumThreads`, `SetInterOpNumThreads`) — never a struct literal
   with those names as public fields. `NewTensor` takes `(Shape, []T)` —
   shape first, flat data second — not `(data [][]int64, shape []int64)`.
   `DynamicAdvancedSession.Run(inputs, outputs []Value) error` takes two
   `[]Value` slices and returns only `error`; it does not take a
   `map[string]ArbitraryTensor` or return `([]Tensor, error)`. The file also
   imported `strings` and `time` without using either — confirming it had
   never once compiled on a non-Windows target. Rewrote `onnx.go`'s
   session setup, tensor creation, and inference call against the real
   v1.36.0 API (verified line-by-line against the module source, since no C
   toolchain was available locally to compile-check it end-to-end).

## Options considered

- **Native matrix, cgo on (chosen)** — matches what `ci.yml` already does
  successfully on these same three runner families (it never disables cgo).
  Lowest effort, no feature loss, no new embedder implementation.
- **Make ONNX optional, add a pure-Go fallback embedder** — would let
  Release keep `CGO_ENABLED=0` and the single-runner cross-compile loop.
  Rejected for now: no pure-Go embedder exists in `internal/search/embedder/`
  today (only `onnx.go`/`onnx_windows.go`), so this means designing and
  validating a new embedding implementation — a much larger, separate
  effort, not a release-pipeline fix.
- **Pure-Go ONNX runtime (`owulveryck/onnx-go` + Gorgonia)** — exists, but is
  far less actively maintained than Microsoft's `onnxruntime`, and swapping
  the runtime underneath an already-working embedder risks silent accuracy/
  performance regressions with no strong reason to take that risk here.
  Rejected.

## How to apply

- A future dependency that is cgo-only (no non-cgo build tag) and goes into
  a binary built by `release.yml` needs this same native-matrix treatment —
  don't add it back to the `CGO_ENABLED=0` cross-compile loop without
  checking this first.
- **A version bump alone does not prove an API matches** — compiling (or at
  minimum reading the dependency's actual source for every call site used)
  is the only check that catches a wrong call shape; "the versions now
  agree" and "the code now compiles" are different claims, and only the
  second one is the one that matters. This file is a case of claiming the
  first and only later checking the second.
- Not locally verified end-to-end (no C toolchain on the machine that made
  this change, for either the dependency-version fix or the `onnx.go`
  rewrite) — verify on the next real release run (next `v*` tag push, or a
  manual `workflow_dispatch`) that all 5 matrix jobs produce a binary before
  trusting this closes the issue.

## References

- `github.com/yalue/onnxruntime_go` v1.36.0 source, inspected 2026-10-10:
  every non-test `.go` file does `import "C"`, no non-cgo variant.
- GitHub Actions run `38061086140` (Release, failed) and `38061082055` (CI,
  failed) — both on commit `5d3c130`, retrieved 2026-10-10.
