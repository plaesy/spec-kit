---
name: remove-onnx-and-http-embedders
description: "Only rembed remains as embedder. ONNX and HTTP embedders, the cache of vectors in chromem, HNSW settings, and the unused root search package were removed. onnxruntime_go is gone, so cgo is no longer required."
metadata:
  type: decision
  updatedAt: "2026-10-13T00:00:00.000Z"
---

# Decision: Remove ONNX and HTTP embedders; rembed is the only embedder

**Status**: Accepted — implemented
**Decided**: 2026-10-13
**Deciders**: user (asked for it: "saya tidak mau user install ollama dan semisalnya lagi") + session agent
**Supersedes**: [release-native-matrix-cgo](release-native-matrix-cgo.md) — the reason for the native matrix no longer holds
**Related**: [embedding-based-semantic-duplicate-search](embedding-based-semantic-duplicate-search.md) (rembed already used by `internal/semantic`)

## Context

`plaesy search` had two embedders that did not work on the target machine:

- `onnx`: disabled on Windows (`onnx_windows.go`); on Linux/macOS the model download was a placeholder, so the user had to fetch the model by hand.
- `http`: a stub. `Embed` returned "not yet implemented", so an Ollama or TEI server could not be used at all.

The error message told users to run Ollama, which did not work. `rembed` already ran in pure Go for `plaesy analyze`, downloaded its model on first use, and needed no server.

## Decision

1. `rembed` (`intfloat/multilingual-e5-base`, 768 dimensions) is the only embedder. It is the default in `config.DefaultSearchConfig`.
2. Removed: `embedder/onnx.go`, `embedder/onnx_windows.go`, the HTTP embedder, and the `device`, `model_path`, `http_endpoint`, `batch_size`, `max_seq_len` config keys. None of them was read by working code.
3. Removed `github.com/yalue/onnxruntime_go` from `go.mod`.
4. Removed the cache of vectors in `vectorstore.ChromemStore`. The store no longer needs an embedding function: callers pass the vectors they already computed. Before this, every chunk was embedded twice.
5. Removed the unused HNSW settings. chromem does brute-force search, so the values were never read.
6. Removed the root `internal/search` package (`search.go`, `config.go`). Nothing imported it.

## Consequences

- `CGO_ENABLED=0 go build ./cmd/plaesy` succeeds on Windows. The module has no cgo dependency left.
- Not done here: `.github/workflows/release.yml` still builds the native per-OS matrix. Returning it to a single-runner `CGO_ENABLED=0` cross-compile is a pipeline change and needs the user's go-ahead.
- Users with an old unified config that sets `"type": "onnx"` or `"http"` get an error on the next `plaesy search` run. The error names `rembed` as the supported type.
- E5 models expect `query: ` and `passage: ` prefixes. `rembed` embeds exactly the text it is given, so `RembedEmbedder` adds them: `EmbedQuery` uses `query: `, `EmbedPassages` uses `passage: `. Models outside the e5 family get no prefix. The interface splits the two methods so callers cannot mix them up.
- A top-1 check on six English/Indonesian pairs with the real model gave 6/6 with and without prefixes, and a smaller top-1/top-2 margin with them. The set is too small to decide, so the prefixes stay because the model documentation requires them. A larger labelled set is needed before claiming a quality gain.
- Indexes built before this change hold passage vectors without the prefix. Rebuild them with `plaesy search index --force`.

## Verification

- `go vet ./...` clean. Test suite passes in 28 packages. The only failure is `TestEveryShippedCommandHasAChangelogLine`, which fails on the untouched `HEAD` as well.
- Real model test (`PLAESY_TEST_REMBED=1`) passes: 768 dimensions, normalised vectors, cross-language similarity ordering.
- Reopen regression: a store written by one process returns results in another.
