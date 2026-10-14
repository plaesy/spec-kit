package semantic

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	chromem "github.com/philippgille/chromem-go"

	"github.com/plaesy/spec-kit/internal/graph"
)

// Stats describes the semantic index on disk.
type Stats struct {
	Documents         int       // embedded documents: declaration units and file-level nodes
	LastBuilt         time.Time // when the fingerprint was last recorded, zero if never built
	Stale             bool      // true when the sources changed since LastBuilt
	CheckpointPresent bool      // an interrupted build left a resumable checkpoint
}

// HasCheckpoint reports whether an interrupted build left a checkpoint in outDir.
func HasCheckpoint(outDir string) bool {
	_, err := os.Stat(checkpointPath(outDir))
	return err == nil
}

// IndexStats reads the semantic index under paths.OutFull. A missing index is not
// an error: it reports zero documents and Stale true.
func IndexStats(opts graph.Options, paths graph.Paths) (Stats, error) {
	var s Stats
	s.CheckpointPresent = HasCheckpoint(paths.OutFull)

	stale, err := IsIndexStale(opts, paths, paths.OutFull)
	if err != nil {
		return s, fmt.Errorf("checking index staleness: %w", err)
	}
	s.Stale = stale

	if info, err := os.Stat(FingerprintPath(paths.OutFull)); err == nil {
		s.LastBuilt = info.ModTime()
	}

	dbPath := filepath.Join(paths.OutFull, "embeddings")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return s, nil
	}
	db, err := chromem.NewPersistentDB(dbPath, true)
	if err != nil {
		return s, fmt.Errorf("semantic: opening embedding store at %s: %w", dbPath, err)
	}
	if coll := db.GetCollection(collectionName, nil); coll != nil {
		s.Documents = coll.Count()
	}
	return s, nil
}
