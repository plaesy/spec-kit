package analyze

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/plaesy/spec-kit/internal/exclude"
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

// computeFingerprint mirrors analyze_fingerprint(): <count>|<maxmtime>|<version>.
// It walks the tree once, counting files whose extension is in
// fingerprintExts and tracking the newest mtime among them, excluding
// dotfile dirs, node_modules and .plaesy (matching the bash exclusions
// exactly, including that .plaesy is excluded here but NOT in the general
// walkProject() used by the generators).
func computeFingerprint(projectPath, frameworkVersion string) string {
	var count int
	var maxMtime int64
	ig := exclude.LoadGitignore(projectPath)

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
		rel, relErr := filepath.Rel(projectPath, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if ig.SkipDir(rel, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(d.Name())), ".")
		if !fingerprintExts[ext] || ig.SkipFile(rel) {
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

// builtComponents lists the artifact components a run writes. "structure" is
// always written; "graph" and "symbols" only when the graph is not disabled.
func builtComponents(noGraph bool) []string {
	if noGraph {
		return []string{"structure"}
	}
	return []string{"structure", "graph", "symbols"}
}

// componentsTag separates the walk fingerprint from the list of artifact
// components a completed run built, e.g. "<count>|<mtime>|<version>|components=
// structure,graph,symbols". A fingerprint without the tag (written before
// components were tracked) records no components, so it never satisfies a skip.
const componentsTag = "|components="

// fingerprintWithComponents appends the built components to a walk fingerprint.
// With no components it returns fp unchanged, so the plain form stays valid.
func fingerprintWithComponents(fp string, components []string) string {
	if len(components) == 0 {
		return fp
	}
	return fp + componentsTag + strings.Join(components, ",")
}

// splitComponents returns the walk fingerprint and the set of components a
// stored fingerprint recorded.
func splitComponents(stored string) (string, map[string]bool) {
	built := map[string]bool{}
	i := strings.Index(stored, componentsTag)
	if i < 0 {
		return stored, built
	}
	for _, c := range strings.Split(stored[i+len(componentsTag):], ",") {
		built[c] = true
	}
	return stored[:i], built
}

// shouldSkipAnalysis mirrors should_skip_analysis(): returns true (skip) when
// the current walk fingerprint matches the one recorded by the last *completed*
// run, and that run built every component in need. It records nothing — see
// recordFingerprint. analysisDir must already exist.
func shouldSkipAnalysis(projectPath, analysisDir, frameworkVersion string, need ...string) bool {
	fpFile := filepath.Join(analysisDir, ".analysis-fingerprint")
	data, err := os.ReadFile(fpFile)
	if err != nil {
		return false
	}
	base, built := splitComponents(string(data))
	if base != computeFingerprint(projectPath, frameworkVersion) {
		return false
	}
	for _, c := range need {
		if !built[c] {
			return false
		}
	}
	return true
}

// recordFingerprint stores the fingerprint of a finished analysis.
//
// It is deliberately a separate call made after the generators succeed. Writing
// it from shouldSkipAnalysis — which the bash original did — means a run that
// fails halfway leaves a valid fingerprint behind, and the next run skips,
// prints "Analysis unchanged since last run" and lists project.json as
// "(cached)" for a file that was never written. The fingerprint is a claim that
// the artifacts exist, so it is written when that becomes true.
func recordFingerprint(projectPath, analysisDir, frameworkVersion string, components ...string) error {
	fpFile := filepath.Join(analysisDir, ".analysis-fingerprint")
	fp := fingerprintWithComponents(computeFingerprint(projectPath, frameworkVersion), components)
	return os.WriteFile(fpFile, []byte(fp), 0o644)
}
