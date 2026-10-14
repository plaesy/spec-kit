package semantic

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/plaesy/spec-kit/internal/graph"
)

const fingerprintFileName = ".fingerprint"

// FingerprintPath returns the path to the fingerprint file stored alongside
// the embeddings directory.
func FingerprintPath(outDir string) string {
	return filepath.Join(outDir, "embeddings", fingerprintFileName)
}

// StoredFingerprint reads the fingerprint recorded at the last BuildIndex run.
// Returns empty string if no fingerprint file exists.
func StoredFingerprint(outDir string) (string, error) {
	path := FingerprintPath(outDir)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("reading fingerprint: %w", err)
	}
	return string(b), nil
}

// RecordFingerprint computes the current source fingerprint via
// graph.SourceFingerprint and writes it to the fingerprint file.
func RecordFingerprint(opts graph.Options, paths graph.Paths) error {
	fp, err := graph.SourceFingerprint(opts, paths)
	if err != nil {
		return fmt.Errorf("computing source fingerprint: %w", err)
	}
	path := FingerprintPath(paths.OutFull)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating fingerprint dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(fp), 0o644); err != nil {
		return fmt.Errorf("writing fingerprint: %w", err)
	}
	return nil
}

// IsIndexStale compares the current source fingerprint against the one
// stored at the last BuildIndex. Returns true if the index is stale
// (no stored fingerprint, or fingerprint mismatch).
func IsIndexStale(opts graph.Options, paths graph.Paths, outDir string) (bool, error) {
	stored, err := StoredFingerprint(outDir)
	if err != nil {
		return true, err
	}
	if stored == "" {
		return true, nil
	}
	current, err := graph.SourceFingerprint(opts, paths)
	if err != nil {
		return true, err
	}
	return stored != current, nil
}
