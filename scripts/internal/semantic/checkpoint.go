package semantic

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// An interrupted build resumes from a checkpoint instead of embedding every
// document again. The defaults below are ASSUMED (see the F4 task file):
//   - a batch is 64 documents; each batch is written and synced before the next
//     one starts, so a kill loses at most the batch in progress;
//   - a checkpoint is reused only for the exact document set it was written for
//     (its digest). Any change to that set discards it with a message;
//   - a build that finishes deletes its checkpoint.
const (
	checkpointVersion   = 1
	checkpointBatchSize = 64
	checkpointFile      = "embeddings.checkpoint.jsonl"
)

type checkpointHeader struct {
	Version int    `json:"version"`
	Digest  string `json:"digest"`
}

type checkpointEntry struct {
	ID     string    `json:"id"`
	Hash   string    `json:"hash"`
	Vector []float32 `json:"vector"`
}

func checkpointPath(outDir string) string {
	return filepath.Join(outDir, checkpointFile)
}

// inputDigest identifies the exact set of documents to embed: each document's id
// and text hash, sorted, so the order of the graph does not matter.
func inputDigest(docs []indexDoc) string {
	lines := make([]string, len(docs))
	for i, d := range docs {
		lines[i] = d.id + "\x00" + d.meta[metaHash]
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])
}

// loadCheckpoint returns the vectors an earlier interrupted build embedded for
// this same document set. valid reports whether the file's header matched, which
// means new batches can be appended to it. A checkpoint with a different digest
// or version is removed, and the user is told why.
func loadCheckpoint(path, digest string, docs []indexDoc) (vectors map[string][]float32, valid bool) {
	vectors = map[string][]float32{}
	f, err := os.Open(path)
	if err != nil {
		return vectors, false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	var hdr checkpointHeader
	if !sc.Scan() || json.Unmarshal(sc.Bytes(), &hdr) != nil ||
		hdr.Version != checkpointVersion || hdr.Digest != digest {
		fmt.Fprintf(os.Stderr, "plaesy: discarding embedding checkpoint %s: the documents changed since it was written\n", path)
		_ = f.Close()
		_ = os.Remove(path)
		return vectors, false
	}

	hashes := make(map[string]string, len(docs))
	for _, d := range docs {
		hashes[d.id] = d.meta[metaHash]
	}
	for sc.Scan() {
		var e checkpointEntry
		// The last line can be cut short by an interrupted write. Stop there and
		// keep everything before it.
		if json.Unmarshal(sc.Bytes(), &e) != nil {
			break
		}
		if hashes[e.ID] == e.Hash {
			vectors[e.ID] = e.Vector
		}
	}
	return vectors, true
}

// writeCheckpointBatch appends one batch of embedded documents and syncs it to
// disk. When header is false the file is started afresh with the header first.
func writeCheckpointBatch(path, digest string, entries []checkpointEntry, header bool) error {
	flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if !header {
		flags |= os.O_TRUNC
	}
	f, err := os.OpenFile(path, flags, 0o600)
	if err != nil {
		return fmt.Errorf("semantic: opening checkpoint %s: %w", path, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	enc := json.NewEncoder(w)
	if !header {
		if err := enc.Encode(checkpointHeader{Version: checkpointVersion, Digest: digest}); err != nil {
			return fmt.Errorf("semantic: writing checkpoint header: %w", err)
		}
	}
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			return fmt.Errorf("semantic: writing checkpoint entry: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("semantic: flushing checkpoint: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("semantic: syncing checkpoint: %w", err)
	}
	return nil
}

// removeCheckpoint deletes a finished or obsolete checkpoint. A missing file is
// not an error.
func removeCheckpoint(path string) {
	_ = os.Remove(path)
}
