package analyze

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// fingerprintExts is the set of files whose change makes a stored analysis
// stale: everything the analysis itself classifies (source, documentation,
// configuration) plus the build/lock/style files that change what a project is.
// The first three come from the taxonomy in output.go, so an extension the
// analysis counts cannot be missing here — the drift that let a .txt edit skip
// regeneration happened because the two lists were written separately.
var fingerprintExts = buildFingerprintExts()

func buildFingerprintExts() map[string]bool {
	out := make(map[string]bool, len(sourceCodeExts)+len(docExts)+len(cfgExts)+len(extraFingerprintExts))
	for _, group := range [][]string{sourceCodeExts, docExts, cfgExts, extraFingerprintExts} {
		for _, ext := range group {
			out[ext] = true
		}
	}
	return out
}

// extraFingerprintExts are tracked by the fingerprint but counted in no bucket:
// stylesheets, SQL, lock files and editor/project files. They are not
// documentation, source or configuration as the analysis defines those words, and
// leaving them out meant the same class of stale analysis as the .txt case.
var extraFingerprintExts = []string{
	"sql", "scss", "sass", "less", "lock", "env", "gradle", "properties", "txt",
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
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(d.Name())), ".")
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
// the current fingerprint matches the one recorded by the last *completed* run.
// It records nothing — see recordFingerprint. analysisDir must already exist.
func shouldSkipAnalysis(projectPath, analysisDir, frameworkVersion string) bool {
	fpFile := filepath.Join(analysisDir, ".analysis-fingerprint")
	current := computeFingerprint(projectPath, frameworkVersion)

	data, err := os.ReadFile(fpFile)
	return err == nil && string(data) == current
}

// recordFingerprint stores the fingerprint of a finished analysis.
//
// It is deliberately a separate call made after the generators succeed. Writing
// it from shouldSkipAnalysis — which the bash original did — means a run that
// fails halfway leaves a valid fingerprint behind, and the next run skips,
// prints "Analysis unchanged since last run" and lists project.json as
// "(cached)" for a file that was never written. The fingerprint is a claim that
// the artifacts exist, so it is written when that becomes true.
func recordFingerprint(projectPath, analysisDir, frameworkVersion string) error {
	fpFile := filepath.Join(analysisDir, ".analysis-fingerprint")
	return os.WriteFile(fpFile, []byte(computeFingerprint(projectPath, frameworkVersion)), 0o644)
}
