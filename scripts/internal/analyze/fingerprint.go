package analyze

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// fingerprintExts mirrors the ext_names list in analyze_fingerprint() (bash).
var fingerprintExts = map[string]bool{
	".md": true, ".ps1": true, ".sh": true, ".js": true, ".jsx": true,
	".ts": true, ".tsx": true, ".py": true, ".go": true, ".dart": true,
	".java": true, ".kt": true, ".kts": true, ".swift": true, ".c": true,
	".h": true, ".cc": true, ".cpp": true, ".hpp": true, ".cs": true,
	".rs": true, ".rb": true, ".php": true, ".json": true, ".yaml": true,
	".yml": true, ".toml": true, ".xml": true, ".ini": true, ".cfg": true,
}

// isExcludedFingerprintDir mirrors `! -path '*/.*/*' ! -path
// '*/node_modules/*' ! -path '*/.plaesy/*'` — analyze_fingerprint()
// additionally excludes .plaesy to avoid a circular dependency on its own
// output, on top of the usual dotfile/node_modules pruning.
func isExcludedFingerprintDirName(name string) bool {
	return strings.HasPrefix(name, ".") || name == "node_modules"
}

// computeFingerprint mirrors analyze_fingerprint(): <count>|<maxmtime>|<version>.
// It walks the tree once, counting files whose extension is in
// fingerprintExts and tracking the newest mtime among them, excluding
// dotfile dirs, node_modules and .plaesy (matching the bash exclusions
// exactly, including that .plaesy is excluded here but NOT in the general
// walkProject() used by the generators).
func computeFingerprint(projectPath, frameworkVersion string) string {
	var count int
	var maxMtime int64

	_ = filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if path == projectPath {
			return nil
		}
		if d.IsDir() {
			if isExcludedFingerprintDirName(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if !fingerprintExts[ext] {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}
		count++
		if m := info.ModTime().Unix(); m > maxMtime {
			maxMtime = m
		}
		return nil
	})

	return strconv.Itoa(count) + "|" + strconv.FormatInt(maxMtime, 10) + "|" + frameworkVersion
}

// shouldSkipAnalysis mirrors should_skip_analysis(): returns true (skip) when
// the current fingerprint matches the one stored from the last run, and
// otherwise persists the new fingerprint and returns false (must regenerate).
// analysisDir must already exist.
func shouldSkipAnalysis(projectPath, analysisDir, frameworkVersion string) bool {
	fpFile := filepath.Join(analysisDir, ".analysis-fingerprint")
	current := computeFingerprint(projectPath, frameworkVersion)

	if data, err := os.ReadFile(fpFile); err == nil {
		if string(data) == current {
			return true
		}
	}
	_ = os.WriteFile(fpFile, []byte(current), 0o644)
	return false
}
