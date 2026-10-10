package analyze

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/plaesy/spec-kit/internal/graph"
	"github.com/plaesy/spec-kit/internal/semantic"
)

// Options mirrors the CLI flags plaesy-analyze.sh accepts.
type Options struct {
	ProjectPath string
	NoGraph     bool
	// NoIndex skips the semantic-search index build. The index is on by
	// default, like the graph and the symbols index, so that one analyze run
	// leaves `plaesy search` queryable instead of stale. --no-index is the
	// opt-out for the same reason --no-graph is: the step costs an embedding
	// model load and a full re-embed of every symbol-bearing node.
	NoIndex bool
	Force   bool
	// IfChanged is accepted for backward compatibility with the legacy
	// --if-changed flag but is a no-op: the fingerprint fast path is now
	// always on by default (see should_skip_analysis in the bash version).
	IfChanged bool
	// Ctx bounds the semantic index build. nil means context.Background(), so
	// Run stays callable without one — which is how its tests call it.
	Ctx context.Context
}

// ctx returns the context the semantic index build runs under, defaulting to
// context.Background() so a zero Options is still usable.
func (o Options) ctx() context.Context {
	if o.Ctx != nil {
		return o.Ctx
	}
	return context.Background()
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

// ensureSemantic is the seam tests replace with a stub. The real
// ensureSemanticIndex loads a local embedding model — a download on a machine's
// first run — so no test may reach it through Run. Same pattern, and the same
// reason, as the Logf/Successf vars above.
var ensureSemantic = ensureSemanticIndex

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
			Logf("   - project.symbols.md - Function/class index (cached)")
		}
		// "Skipping regeneration" is a claim about the analysis artifacts, and
		// the semantic index is not one of them: it carries its own
		// fingerprint, and an earlier run that could not load the embedding
		// model (an offline first run) leaves it stale or missing — a state
		// `plaesy search` refuses to query. Re-checking it here is what makes
		// "analyze, then search" work on the path where nothing was
		// regenerated. Cheap when it is fresh: one fingerprint walk, no model
		// load, no embedding.
		if !opts.NoGraph && !opts.NoIndex {
			rebuilt, err := ensureSemantic(opts.ctx(), projectPath, false)
			switch {
			case err != nil:
				Logf("Semantic search index not rebuilt: %v. Run 'plaesy search --index' to retry.", err)
			case rebuilt:
				Successf("analysis/embeddings/ generated (semantic search index)")
			default:
				Logf("   - embeddings/ - Semantic search index (fresh)")
			}
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

	graphFailed := false
	indexFailed := false
	if !opts.NoGraph {
		if opts.Force {
			Logf("Building dependency graph (forced)...")
		} else {
			Logf("Building dependency graph (incremental: only rebuilds if source changed)...")
		}
		if err := buildGraph(projectPath, opts.Force); err != nil {
			// A failed graph build is not fatal: the summary artifacts were
			// written above and stay valid, and --no-graph is the opt-out.
			// What must not happen is what used to: the run then printed
			// "Comprehensive analysis completed!", listed the three graph
			// files under "Generated files" (built from the flag, not from
			// what was written), and recorded the fingerprint — so the next
			// run took the skip path and reported the missing graph as
			// "(cached)", and it was never built again without --force.
			Logf("Graph build skipped/failed; summary files remain valid. (%v)", err)
			graphFailed = true
		} else if err := generateSymbolsIndex(projectPath, analysisDir); err != nil {
			// Treated like a graph build failure for fingerprinting purposes:
			// if we recorded the fingerprint anyway, the next run would take
			// the skip path and report project.symbols.md as "(cached)" even
			// though it was never written.
			Logf("Symbols index skipped/failed; other analysis files remain valid. (%v)", err)
			graphFailed = true
		} else {
			Successf("analysis/project.symbols.md generated (function/class index, grouped by file)")

			// Built here, after the graph, because the index embeds the graph's
			// symbols: an index built from a stale graph is exactly the stale
			// index `plaesy search` refuses to query.
			if !opts.NoIndex {
				rebuilt, err := ensureSemantic(opts.ctx(), projectPath, opts.Force)
				switch {
				case err != nil:
					// Non-fatal, and deliberately not folded into graphFailed.
					// The analysis fingerprint is a claim that project.json, the
					// graph artifacts and project.symbols.md exist — all three
					// were written. The index has its own fingerprint and its own
					// actionable error in `plaesy search`, so withholding the
					// analysis fingerprint here would only mean a machine that
					// cannot load the embedding model re-runs the entire analysis
					// on every invocation, forever, over a step that is not
					// analysis.
					Logf("Semantic search index not rebuilt: %v. Run 'plaesy search --index' to retry.", err)
					indexFailed = true
				case rebuilt:
					Successf("analysis/embeddings/ generated (semantic search index)")
				default:
					Logf("   - embeddings/ - Semantic search index unchanged (source fingerprint matches)")
				}
			}
		}
	}

	Successf("Comprehensive analysis completed!")

	// Recorded only when every artifact this run was supposed to write
	// actually was: the fingerprint is the claim that they exist. A run
	// whose graph build failed must not leave one behind, or the fast path
	// above would report a graph that was never built as cached forever.
	if graphFailed {
		Logf("Not recording the analysis fingerprint: the graph artifacts are missing, so the next run must regenerate.")
	} else if err := recordFingerprint(projectPath, analysisDir, frameworkVersion); err != nil {
		Logf("Could not record the analysis fingerprint (a later run will regenerate): %v", err)
	}

	Logf("Generated files:")
	Logf("Analysis files (in %s):", analysisDir)
	Logf("   - project.json - AI-optimized project summary")
	Logf("   - project.structure.json - Detailed project structure")
	if !opts.NoGraph && !graphFailed {
		Logf("   - project.graph.json - Dependency graph (nodes + edges)")
		Logf("   - project.html - Interactive graph visualization")
		Logf("   - reports.md - Graph report (communities, god nodes, orphans)")
		Logf("   - project.symbols.md - Function/class index, grouped by file (check before adding new functions)")
		if !opts.NoIndex {
			if indexFailed {
				Logf("   - embeddings/ - NOT rebuilt this run (`plaesy search --index` retries it on demand)")
			} else {
				Logf("   - embeddings/ - Semantic search index, what `plaesy search` queries")
			}
		}
	}
	Logf("   - overview.md - Analysis snapshot (replaced every run)")
	Logf("No topic memory files are generated by analyze.")

	return nil
}

// ensureSemanticIndex leaves the semantic-search index matching the graph in
// the project's analysis directory, and reports whether it had to rebuild it.
//
// It gates on semantic.IsIndexStale — the same graph.SourceFingerprint check
// `plaesy search` runs before every query — so analyze and search cannot
// disagree about what "fresh" means. A run that finds the fingerprint unchanged
// pays neither a model load nor an embedding, and the query that follows is not
// refused as stale. force skips that gate, the way it does for the graph.
func ensureSemanticIndex(ctx context.Context, projectPath string, force bool) (bool, error) {
	opts := graph.Options{RepoPath: projectPath}
	paths, err := graph.ResolvePaths(opts)
	if err != nil {
		return false, fmt.Errorf("resolving graph paths: %w", err)
	}

	if !force {
		stale, err := semantic.IsIndexStale(opts, paths, paths.OutFull)
		if err != nil {
			return false, fmt.Errorf("checking index staleness: %w", err)
		}
		if !stale {
			return false, nil
		}
	}

	g, err := graph.LoadJSON(paths.ProjectJSON)
	if err != nil {
		return false, fmt.Errorf("loading project.graph.json: %w", err)
	}
	emb, err := semantic.LoadDefaultEmbedder()
	if err != nil {
		return false, err
	}
	if err := semantic.BuildIndex(ctx, g, paths.OutFull, paths.RepoRoot, emb); err != nil {
		return false, err
	}
	// Recorded last, and only once the collection is written: this fingerprint
	// is the claim that the index matches the source, so a failed build has to
	// leave the previous (stale) claim standing for `plaesy search` to catch.
	if err := semantic.RecordFingerprint(opts, paths); err != nil {
		return false, fmt.Errorf("recording fingerprint: %w", err)
	}
	return true, nil
}

// generateSymbolsIndex writes project.symbols.md: every function/class/heading
// symbol the graph builder already extracted (graph.Node.Symbols), grouped by
// file. It deliberately omits line numbers — lines drift on every edit, so a
// cached line number is actively misleading, whereas a file path + symbol
// name stays correct until the symbol itself is renamed or removed. Intent
// is to let an AI session scan this before adding a new function, so it can
// spot an existing one with the same name (or go Read the file to check
// intent) instead of creating a duplicate.
func generateSymbolsIndex(projectPath, analysisDir string) error {
	paths, err := graph.ResolvePaths(graph.Options{RepoPath: projectPath})
	if err != nil {
		return fmt.Errorf("resolving graph paths: %w", err)
	}
	g, err := graph.LoadJSON(paths.ProjectJSON)
	if err != nil {
		return fmt.Errorf("loading project.graph.json: %w", err)
	}

	type fileSymbols struct {
		path    string
		symbols []string
	}
	var files []fileSymbols
	for _, n := range g.Nodes {
		if len(n.Symbols) == 0 {
			continue
		}
		files = append(files, fileSymbols{path: n.ID, symbols: n.Symbols})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].path < files[j].path })

	// Cross-file name collisions: the single highest-signal entry for
	// duplicate-avoidance, since a repeated name is a duplicate (or a
	// near-duplicate worth checking) regardless of which file it's in.
	// Markdown headings are excluded: the same heading text (e.g.
	// "Overview") recurring across docs is normal and not a code
	// duplication signal.
	byName := map[string][]string{}
	for _, f := range files {
		if filepath.Ext(f.path) == ".md" {
			continue
		}
		for _, s := range f.symbols {
			byName[s] = append(byName[s], f.path)
		}
	}
	var dupNames []string
	for name, locs := range byName {
		if len(locs) > 1 {
			dupNames = append(dupNames, name)
		}
	}
	sort.Strings(dupNames)

	// Near-duplicates: same concept, different spelling convention
	// (getUserById / get_user_by_id / GetUserById all normalize to
	// "getuserbyid"). This is the case exact-name matching above cannot
	// catch, and it's exactly the case a mixed-convention codebase (or a
	// change in house style over time) produces. It still can't catch a
	// genuine synonym (fetchUser vs getUser) — that's a semantic judgment,
	// not a spelling one, and stays the AI's job to catch by reading code.
	byNormalized := map[string][]string{} // normalized -> distinct original spellings
	seenPerNorm := map[string]map[string]bool{}
	for name := range byName {
		norm := normalizeSymbolName(name)
		if seenPerNorm[norm] == nil {
			seenPerNorm[norm] = map[string]bool{}
		}
		if !seenPerNorm[norm][name] {
			seenPerNorm[norm][name] = true
			byNormalized[norm] = append(byNormalized[norm], name)
		}
	}
	var similarKeys []string
	for norm, spellings := range byNormalized {
		if len(spellings) > 1 {
			similarKeys = append(similarKeys, norm)
			sort.Strings(spellings)
		}
	}
	sort.Strings(similarKeys)

	var b strings.Builder
	b.WriteString("# Symbols Index\n\n")
	b.WriteString("Function/class/heading names extracted per file, for duplicate-avoidance checks before adding new code. ")
	b.WriteString("No line numbers: they go stale on every edit. Regenerated every `plaesy analyze` run alongside the dependency graph.\n\n")
	if len(files) == 0 {
		b.WriteString("_No symbols extracted for this project's current file set._\n")
	}

	b.WriteString("## Duplicate Names (appear more than once)\n\n")
	if len(dupNames) == 0 {
		b.WriteString("_None found._\n\n")
	} else {
		b.WriteString("Same name more than once: could be intentional (interface implementations, per-package `New`, overloads in languages that support them) or an accidental duplicate — check before adding another one.\n\n")
		for _, name := range dupNames {
			fmt.Fprintf(&b, "- `%s` — %s\n", name, formatLocations(byName[name]))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Similar Names (different spelling, same normalized form)\n\n")
	if len(similarKeys) == 0 {
		b.WriteString("_None found._\n\n")
	} else {
		b.WriteString("Same letters/digits once case and separators (`_`) are stripped — likely the same concept named inconsistently (or copy-pasted and renamed). Not caught by exact-name matching above.\n\n")
		for _, norm := range similarKeys {
			spellings := byNormalized[norm]
			parts := make([]string, len(spellings))
			for i, name := range spellings {
				parts[i] = fmt.Sprintf("`%s` (%s)", name, formatLocations(byName[name]))
			}
			fmt.Fprintf(&b, "- %s\n", strings.Join(parts, ", "))
		}
		b.WriteString("\n")
	}

	for _, f := range files {
		fmt.Fprintf(&b, "## %s\n\n", f.path)
		for _, s := range f.symbols {
			fmt.Fprintf(&b, "- `%s`\n", s)
		}
		b.WriteString("\n")
	}

	return os.WriteFile(filepath.Join(analysisDir, "project.symbols.md"), []byte(b.String()), 0o644)
}

// normalizeSymbolName reduces a symbol name to a convention-independent key:
// lowercase letters and digits only, separators (_, -, etc.) dropped. Used
// to group getUserById / get_user_by_id / GetUserById as the same concept.
func normalizeSymbolName(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r - 'A' + 'a')
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		}
	}
	return b.String()
}

// formatLocations collapses repeated occurrences of the same file (e.g. two
// methods named RoundTrip in one test file) into "path (xN)" instead of
// listing the path N times, which reads as a cross-file duplicate when it
// isn't one.
func formatLocations(paths []string) string {
	counts := map[string]int{}
	var order []string
	for _, p := range paths {
		if counts[p] == 0 {
			order = append(order, p)
		}
		counts[p]++
	}
	sort.Strings(order)
	out := make([]string, len(order))
	for i, p := range order {
		if n := counts[p]; n > 1 {
			out[i] = fmt.Sprintf("%s (x%d)", p, n)
		} else {
			out[i] = p
		}
	}
	return strings.Join(out, ", ")
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
