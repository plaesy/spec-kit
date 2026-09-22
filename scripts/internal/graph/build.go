package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Options mirrors the bash script's parsed CLI flags that affect a build.
type Options struct {
	RepoPath string // --path
	OutDir   string // --outdir (relative to RepoPath, default ".plaesy/analysis")
}

// Paths holds the resolved absolute locations a build reads/writes, mirroring
// REPO_ROOT / OUT_FULL / PROJECT_JSON / etc. in the bash script.
type Paths struct {
	RepoRoot        string // OS-native absolute path
	RepoRootSlash   string // forward-slash form, used for all internal joins
	OutFull         string // OS-native absolute path
	ProjectJSON     string
	FingerprintFile string
}

func ResolvePaths(opts Options) (Paths, error) {
	repoRoot, err := filepath.Abs(opts.RepoPath)
	if err != nil {
		return Paths{}, fmt.Errorf("resolving --path: %w", err)
	}
	outDir := opts.OutDir
	if outDir == "" {
		outDir = ".plaesy/analysis"
	}
	outFull := filepath.Join(repoRoot, outDir)
	return Paths{
		RepoRoot:        repoRoot,
		RepoRootSlash:   toSlash(repoRoot),
		OutFull:         outFull,
		ProjectJSON:     filepath.Join(outFull, "project.graph.json"),
		FingerprintFile: filepath.Join(outFull, ".fingerprint"),
	}, nil
}

// outDirNorm mirrors `${OUT_DIR#./}; ${%/}` - the relative-slash form of the
// output dir used to exclude it from the file scan.
func outDirNorm(opts Options, p Paths) string {
	rel, err := filepath.Rel(p.RepoRoot, p.OutFull)
	if err != nil {
		return ""
	}
	rel = toSlash(rel)
	rel = strings.TrimPrefix(rel, "./")
	return strings.TrimSuffix(rel, "/")
}

// Build mirrors build_graph(): scans the repo, extracts structural + mention
// edges + symbols, computes degree and communities, and returns the full
// Graph. It does not write any files - callers use render.go's writers.
func Build(opts Options, p Paths) (*Graph, error) {
	fmt.Fprintf(os.Stderr, "[plaesy-graph] Scanning %s ...\n", p.RepoRoot)
	if err := os.MkdirAll(p.OutFull, 0o755); err != nil {
		return nil, err
	}

	files, err := collectFiles(p.RepoRoot, outDirNorm(opts, p))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files found under %s", p.RepoRoot)
	}

	known := make(map[string]bool, len(files))
	ntype := make(map[string]string, len(files))
	ngroup := make(map[string]string, len(files))
	nlayer := make(map[string]string, len(files))
	for _, rel := range files {
		known[rel] = true
		ntype[rel] = nodeTypeOf(rel)
		if i := strings.IndexByte(rel, '/'); i >= 0 {
			ngroup[rel] = rel[:i]
		} else {
			ngroup[rel] = rel
		}
		nlayer[rel] = getArchitectureLayer(rel)
	}

	// 2) structural extraction (2a-2g)
	edges := extractEdges(p.RepoRootSlash, files, known)

	// 2g) mentions (one pass mirroring plaesy-graph-mentions.awk)
	edges = append(edges, mentionEdges(p.RepoRootSlash, files, files)...)

	// 2h) symbols (one pass mirroring plaesy-graph-symbols.awk)
	symbols := make(map[string][]string, len(files))
	for _, rel := range files {
		content := readAll(joinRoot(p.RepoRootSlash, rel))
		if content == "" {
			continue
		}
		if syms := symbolsOf(rel, content); len(syms) > 0 {
			symbols[rel] = syms
		}
	}

	// 3) mirrors: same basename across different folders
	byBase := make(map[string][]string)
	for _, rel := range files {
		base := rel
		if i := strings.LastIndexByte(rel, '/'); i >= 0 {
			base = rel[i+1:]
		}
		byBase[base] = append(byBase[base], rel)
	}
	for _, members := range byBase {
		if len(members) < 2 {
			continue
		}
		sort.Strings(members)
		for i := 0; i < len(members); i++ {
			for j := i + 1; j < len(members); j++ {
				edges = append(edges, Edge{Source: members[i], Target: members[j], Type: "mirrors", Confidence: "EXTRACTED"})
			}
		}
	}

	edges = dropWeakMentions(edges)
	sort.Slice(edges, func(i, j int) bool {
		a, b := edges[i], edges[j]
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		if a.Target != b.Target {
			return a.Target < b.Target
		}
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		return a.Confidence < b.Confidence
	})

	// 4) degree, 5) communities
	degree := computeDegree(files, edges)
	adj := buildAdjacency(edges)
	label := labelPropagation(files, adj)
	community := assignCommunityIDs(files, label)

	// 6) assemble nodes (sorted by id, matching `sort "$NODES_TSV"`)
	nodes := make([]Node, 0, len(files))
	for _, rel := range files {
		nodes = append(nodes, Node{
			ID:        rel,
			Label:     baseName(rel),
			Type:      ntype[rel],
			Group:     ngroup[rel],
			Degree:    degree[rel],
			Community: community[rel],
			Layer:     nlayer[rel],
			Symbols:   symbols[rel],
		})
	}

	g := &Graph{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		Root:        p.RepoRoot,
		Description: fmt.Sprintf("Knowledge graph of %s (%d files): source, docs, config and their references, calls, imports, and text mentions.", filepath.Base(p.RepoRoot), len(nodes)),
		Nodes:       nodes,
		Edges:       edges,
	}
	return g, nil
}

func baseName(rel string) string {
	if i := strings.LastIndexByte(rel, '/'); i >= 0 {
		return rel[i+1:]
	}
	return rel
}

// SourceFingerprint mirrors source_fingerprint(): a cheap "did anything
// change" signature (file count + newest mtime) over the same file set Build
// scans, used by --if-changed / --watch.
func SourceFingerprint(opts Options, p Paths) (string, error) {
	files, err := collectFiles(p.RepoRoot, outDirNorm(opts, p))
	if err != nil {
		return "", err
	}
	var maxMtime time.Time
	for _, rel := range files {
		fi, err := os.Stat(joinRoot(p.RepoRootSlash, rel))
		if err != nil {
			continue
		}
		if fi.ModTime().After(maxMtime) {
			maxMtime = fi.ModTime()
		}
	}
	return fmt.Sprintf("%d|%d", len(files), maxMtime.Unix()), nil
}
