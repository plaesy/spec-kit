package analyze

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/plaesy/spec-kit/internal/graph"
)

// Options mirrors the CLI flags plaesy-analyze.sh accepts.
type Options struct {
	ProjectPath string
	NoGraph     bool
	Force       bool
	// IfChanged is accepted for backward compatibility with the legacy
	// --if-changed flag but is a no-op: the fingerprint fast path is now
	// always on by default (see should_skip_analysis in the bash version).
	IfChanged bool
}

// runContext holds everything computed once per run and reused across the
// three generator functions, mirroring the *_CACHE globals and mapfile in
// main() of plaesy-analyze.sh (detect_all_frameworks / detect_project_type /
// detect_all_languages / count_file_types are each called exactly once).
type runContext struct {
	walk             walkResult
	frameworks       []FrameworkHit
	projectType      FrameworkHit
	languages        []string
	devTools         []string
	buildSystems     []string
	fileCounts       fileCounts
	frameworkVersion string
}

// Logf and Successf mirror log_info/log_success. Kept as vars so callers
// (the cobra command) don't need any additional wiring; default behavior
// prints to stdout like the bash version's echo -e.
var (
	Logf = func(format string, a ...interface{}) {
		fmt.Printf("[INFO] "+format+"\n", a...)
	}
	Successf = func(format string, a ...interface{}) {
		fmt.Printf("[SUCCESS] "+format+"\n", a...)
	}
)

// Run mirrors main() in plaesy-analyze.sh end to end: resolve paths, ensure
// directories exist, check the fingerprint fast path, run every generator
// once, and (unless NoGraph) delegate to the graph builder.
func Run(opts Options) error {
	projectPath := opts.ProjectPath
	if projectPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		projectPath = wd
	}
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}
	projectPath = absProjectPath

	analysisDir := filepath.Join(projectPath, ".plaesy", "analysis")
	memoryDir := filepath.Join(projectPath, ".plaesy", "memory")
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		return fmt.Errorf("create analysis dir: %w", err)
	}
	if err := os.MkdirAll(memoryDir, 0o755); err != nil {
		return fmt.Errorf("create memory dir: %w", err)
	}

	frameworkVersion := frameworkVersionFor(projectPath)

	Logf("Starting comprehensive project analysis...")
	Logf("Project path: %s", projectPath)
	Logf("Analysis directory: %s", analysisDir)

	// Fingerprint fast path (default, not opt-in): skip all regeneration if
	// the project fingerprint matches the last run. --force bypasses this.
	if !opts.Force && shouldSkipAnalysis(projectPath, analysisDir, frameworkVersion) {
		Successf("Analysis unchanged since last run. Skipping regeneration (use --force to override).")
		Logf("Analysis files (in %s):", analysisDir)
		Logf("   - project.json - AI-optimized project summary (cached)")
		Logf("   - project.structure.json - Detailed project structure (cached)")
		Logf("   - overview.md - Analysis snapshot (cached)")
		if !opts.NoGraph {
			Logf("   - project.graph.json - Dependency graph (cached)")
			Logf("   - project.html - Interactive graph visualization (cached)")
			Logf("   - reports.md - Graph report (cached)")
		}
		return nil
	}

	walk, err := walkProject(projectPath)
	if err != nil {
		return fmt.Errorf("walk project: %w", err)
	}

	ctx := &runContext{
		walk:             walk,
		frameworkVersion: frameworkVersion,
	}
	// Detect once, reuse everywhere (mirrors FRAMEWORKS_CACHE /
	// PROJECT_TYPE_CACHE / ALL_LANGUAGES_CACHE in main()). Order matters:
	// projectType is derived from frameworks.
	ctx.frameworks = detectAllFrameworks(projectPath)
	ctx.projectType = detectProjectType(ctx.frameworks)
	ctx.languages = detectAllLanguages(fileNames(walk.files))
	ctx.devTools = detectDevelopmentTools(projectPath)
	ctx.buildSystems = detectBuildSystems(projectPath)
	ctx.fileCounts = countFileTypes(walk.files)

	if err := generateProjectJSON(analysisDir, projectPath, ctx); err != nil {
		return fmt.Errorf("generate project.json: %w", err)
	}
	if err := generateProjectStructureJSON(analysisDir, ctx); err != nil {
		return fmt.Errorf("generate project.structure.json: %w", err)
	}
	if err := generateOverviewMD(projectPath, analysisDir, ctx); err != nil {
		return fmt.Errorf("generate overview.md: %w", err)
	}
	Successf("Comprehensive project.json generated")
	Successf("Detailed project.structure.json generated")
	Successf("analysis/overview.md replaced with latest snapshot")

	if !opts.NoGraph {
		if opts.Force {
			Logf("Building dependency graph (forced)...")
		} else {
			Logf("Building dependency graph (incremental: only rebuilds if source changed)...")
		}
		if err := buildGraph(projectPath, opts.Force); err != nil {
			Logf("Graph build skipped/failed; summary files remain valid. (%v)", err)
		}
	}

	Successf("Comprehensive analysis completed!")
	Logf("Generated files:")
	Logf("Analysis files (in %s):", analysisDir)
	Logf("   - project.json - AI-optimized project summary")
	Logf("   - project.structure.json - Detailed project structure")
	if !opts.NoGraph {
		Logf("   - project.graph.json - Dependency graph (nodes + edges)")
		Logf("   - project.html - Interactive graph visualization")
		Logf("   - reports.md - Graph report (communities, god nodes, orphans)")
	}
	Logf("   - overview.md - Analysis snapshot (replaced every run)")
	Logf("No topic memory files are generated by analyze.")

	return nil
}

func fileNames(files []fileInfo) []string {
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.relPath
	}
	return names
}

// frameworkVersionFor mirrors FRAMEWORK_VERSION resolution in the bash
// script (`cat "$SCRIPT_DIR/../../VERSION" 2>/dev/null || echo "0.0.1"`):
// it reads the VERSION file two levels up from this binary's scripts
// directory, but since the Go build already threads a version through
// common.Version at link time, prefer that when set and fall back to
// reading VERSION relative to the project/executable, then "0.0.1".
func frameworkVersionFor(projectPath string) string {
	if v := versionFromCommon(); v != "" && v != "0.0.0" {
		return v
	}
	// Fall back to a VERSION file two directories above the running
	// executable (mirrors $SCRIPT_DIR/../../VERSION for scripts/bash/*.sh).
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "..", "..", "VERSION")
		if data, err := os.ReadFile(candidate); err == nil {
			return trimVersion(string(data))
		}
	}
	return "0.0.1"
}

func trimVersion(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r' || s[len(s)-1] == ' ') {
		s = s[:len(s)-1]
	}
	return s
}

// buildGraph calls the internal/graph package directly (in-process, no
// shell-out), matching the --if-changed / force flag the bash analyzer used:
// force triggers a full rebuild, non-force only rebuilds if source files
// changed since the last graph build.
func buildGraph(projectPath string, force bool) error {
	opts := graph.Options{RepoPath: projectPath}
	paths, err := graph.ResolvePaths(opts)
	if err != nil {
		return fmt.Errorf("resolving graph paths: %w", err)
	}

	if !force {
		fingerprint, err := graph.SourceFingerprint(opts, paths)
		if err == nil {
			if prev, readErr := os.ReadFile(paths.FingerprintFile); readErr == nil && string(prev) == fingerprint {
				return nil // unchanged since last build, skip
			}
		}
	}

	g, err := graph.Build(opts, paths)
	if err != nil {
		return fmt.Errorf("building graph: %w", err)
	}
	if err := graph.SaveJSON(g, paths.ProjectJSON); err != nil {
		return fmt.Errorf("saving project.graph.json: %w", err)
	}
	if err := graph.WriteHTML(g, paths.OutFull); err != nil {
		return fmt.Errorf("writing project.html: %w", err)
	}
	if err := graph.WriteReport(g, paths.OutFull, ""); err != nil {
		return fmt.Errorf("writing reports.md: %w", err)
	}
	if fingerprint, err := graph.SourceFingerprint(opts, paths); err == nil {
		_ = os.WriteFile(paths.FingerprintFile, []byte(fingerprint), 0o644)
	}
	return nil
}
