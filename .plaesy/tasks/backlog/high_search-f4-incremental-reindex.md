# Task: Search F4 - Incremental Indexing + Reindex + Resume

## Objective
Implement incremental indexing: hash-based change detection, re-embed only changed chunks, orphan cleanup, crash-resume capability.

## Scope
- Content-addressable storage: SHA256 of raw doc + chunk text
- Incremental pipeline: scan → hash → diff → embed changed → update indexes
- Orphan detection: chunks in index but not in current scan → delete
- Resume: checkpoint progress every N docs, resume from last checkpoint
- Full reindex command: `--force` bypasses incremental, rebuilds all

## Design Rationale (ADR)
**Hash Strategy:** Two-level hashing:
- Document hash: SHA256(raw file content) - detects any file change
- Chunk hash: SHA256(chunk text + chunk metadata) - detects chunk-level changes
- Stored in `ProcessedDocument.chunk_hashes: map[chunkID]hash`

**Incremental Pipeline:**
```
scan files → compute doc hashes → diff vs stored hashes
  → unchanged: skip entirely
  → changed: re-chunk → compute chunk hashes → diff vs stored chunk hashes
    → unchanged chunks: skip embedding
    → changed/new chunks: embed → upsert to vector+BM25
  → deleted files: remove all chunks from indexes (orphan cleanup)
```

**Checkpoint Format:** JSONL progress file (one line per document) for:
- Crash resilience: append-only, recoverable
- Resume: read last N lines, continue from last completed
- Schema: `{doc_path, doc_hash, status, chunks_processed, timestamp, error?}`

**Orphan Detection:** After scan, any chunk IDs in vector/BM25 index not in current scan → delete. Runs as final pipeline stage.

**Force Reindex:** `--force` clears all indexes and document store, runs full pipeline.

## Deliverables
1. `internal/search/indexer/incremental.go` - IncrementalIndexer with hash diff
2. `internal/search/store/processed.go` - Extend ProcessedDocument with `content_hash`, `chunk_hashes`
3. `internal/search/indexer/pipeline.go` - Checkpoint/resume logic (JSONL progress file)
4. `cmd/plaesy/search_index.go` - Add `--incremental` (default), `--force`, `--resume`
5. `cmd/plaesy/search_reindex.go` - `plaesy search reindex` (alias for `--force`)
6. `cmd/plaesy/search_stats.go` - `plaesy search stats` (index size, docs, chunks, last updated)
7. Tests: unchanged docs not re-embedded, changed docs updated, orphans removed, resume works

## Acceptance Criteria
- Second `index` run on unchanged repo: 0 embeddings, < 1s (verified by test)
- Modified file: only its chunks re-embedded (chunk hash diff verified)
- Deleted file: chunks removed from both vector + BM25 index (orphan cleanup test)
- Crash mid-index: `--resume` continues from checkpoint (integration test: kill -9, resume)
- `plaesy search stats` shows accurate counts (docs, chunks, index size, last updated)
- Checkpoint format versioned (`version: 1`) for forward compatibility
- Large repo performance: 100k docs incremental < 30s (no changes), full reindex < 10min

## Test Plan
```bash
# Unit tests
go test ./internal/search/indexer/... -v -run TestIncremental
go test ./internal/search/store/... -v -run TestProcessedDocument

# Integration tests
# 1. First index
plaesy search index .plaesy/memory --incremental
plaesy search stats

# 2. Second run (no changes)
plaesy search index .plaesy/memory --incremental
# Verify: 0 embeddings, < 1s

# 3. Modify file
echo "new content" >> .plaesy/memory/test.md
plaesy search index .plaesy/memory --incremental
# Verify: only test.md chunks re-embedded

# 4. Delete file
rm .plaesy/memory/test.md
plaesy search index .plaesy/memory --incremental
# Verify: chunks removed from indexes

# 5. Crash resume
plaesy search index .plaesy/memory --incremental &
sleep 2
kill -9 $!
plaesy search index .plaesy/memory --resume
# Verify: completes from checkpoint

# 6. Force reindex
plaesy search reindex .plaesy/memory
# Verify: full rebuild

# Regression
go test ./internal/search/... -count=1
```

## Dependencies
- F1 complete (indexing pipeline exists)
- Document store with hash tracking

## Config Schema Addition
```json
{
  "search": {
    "indexer": {
      "workers": 4,
      "resume": true,
      "batch_size": 100,
      "checkpoint_interval": 50,
      "checkpoint_format_version": 1
    }
  }
}
```

## Config Migration
- New fields added to `config.IndexerConfig` with defaults
- Checkpoint format version enables future migrations
- Existing progress files without version → treated as v0 (legacy), rebuilt

## Operations
- Progress file: `.plaesy/search/indexer_progress.jsonl` (JSONL, append-only)
- Disk usage: ~1KB per 100 docs indexed
- Resume cleanup: progress file truncated on successful completion
- Metrics: `plaesy search stats` includes `last_indexed_at`, `index_duration_ms`

## Estimated Effort
Medium (2-3 days)

## Definition of Done
- All deliverables implemented and tested
- Crash-resume verified with kill -9 test
- Performance benchmarks documented: `docs/perf/incremental-index-<date>.md`
- Documentation: `docs/search-incremental-indexing.md`