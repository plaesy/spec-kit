// Package graph is a Go port of scripts/bash/plaesy-graph.sh (plus its two
// awk helpers, plaesy-graph-mentions.awk and plaesy-graph-symbols.awk): a
// lightweight, dependency-free knowledge-graph builder for the plaesy
// spec-kit repo (or any project). It scans source/doc/config files, extracts
// structural references (markdown links, dot-sourcing, imports, ...) plus
// plain-text "mentions" and per-file symbols, runs label-propagation
// community detection, and emits:
//
//	<outdir>/project.graph.json  - nodes + edges
//	<outdir>/project.html        - self-contained force-directed viz
//	<outdir>/reports.md          - plain-language summary
//
// Only the Go standard library is used.
package graph

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Node mirrors one row of the bash script's NODES_TSV / one entry of
// project.graph.json's "nodes" array.
type Node struct {
	ID        string   `json:"id"`
	Label     string   `json:"label"`
	Type      string   `json:"type"`
	Group     string   `json:"group"`
	Degree    int      `json:"degree"`
	Community int      `json:"community"`
	Layer     string   `json:"layer,omitempty"`
	Symbols   []string `json:"symbols"`
}

// Edge mirrors one row of EDGES_TSV / one entry of project.graph.json's
// "edges" array.
type Edge struct {
	Source     string `json:"source"`
	Target     string `json:"target"`
	Type       string `json:"type"`
	Confidence string `json:"confidence"` // "EXTRACTED" or "INFERRED"
}

// Graph is the full in-memory result of a build, and also the shape
// (de)serialized to/from project.graph.json for the standalone query
// commands (--query, --explain, --path-query, --impact-check, ...).
type Graph struct {
	GeneratedAt string `json:"generated_at"`
	Root        string `json:"root"`
	Description string `json:"description"`
	Nodes       []Node `json:"nodes"`
	Edges       []Edge `json:"edges"`
}

// includeExtRE / excludeDirsRE mirror INCLUDE_EXT_RE / EXCLUDE_DIRS_RE.
var includeExtRE = regexp.MustCompile(`(?i)\.(md|ps1|sh|js|jsx|ts|tsx|py|go|dart|java|kt|kts|swift|c|h|cc|cpp|hpp|cs|rs|rb|php)$`)

// excludeDirSegment reports whether a path segment marks an excluded
// directory: a dot-directory, or one of the well-known build/vendor dirs.
func excludeDirSegment(seg string) bool {
	if seg == "" {
		return false
	}
	if strings.HasPrefix(seg, ".") {
		return true
	}
	switch seg {
	case "node_modules", "dist", "build", "__pycache__", "vendor":
		return true
	}
	return false
}

// pathExcluded mirrors EXCLUDE_DIRS_RE: true if any path segment is excluded.
func pathExcluded(relSlash string) bool {
	for _, seg := range strings.Split(relSlash, "/") {
		if excludeDirSegment(seg) {
			return true
		}
	}
	return false
}

// toSlash normalizes a filepath.WalkDir-relative path to forward slashes,
// matching the bash script's `find . -type f | sed 's|^\./||'` output shape.
func toSlash(p string) string {
	return filepath.ToSlash(p)
}

// nodeTypeOf mirrors node_type_of().
func nodeTypeOf(rel string) string {
	top := rel
	if i := strings.IndexByte(rel, '/'); i >= 0 {
		top = rel[:i]
	}
	switch top {
	case "chatmodes":
		return "chatmode"
	case "instructions":
		return "instruction"
	case "checklists":
		return "checklist"
	case "templates":
		return "template"
	case "scripts":
		return "script"
	case "docs":
		return "doc"
	case "prompts":
		return "prompt"
	case "testing":
		return "testing"
	}
	switch filepath.Ext(rel) {
	case ".js", ".jsx", ".ts", ".tsx", ".py", ".go", ".dart", ".java", ".kt", ".kts", ".swift",
		".c", ".h", ".cc", ".cpp", ".hpp", ".cs", ".rs", ".rb", ".php":
		return "source"
	}
	return "other"
}

var (
	layerAPIFolderRE  = regexp.MustCompile(`^(api|controllers|handlers|routes|endpoints|gateway)$`)
	layerAPIFileRE    = regexp.MustCompile(`(controller|handler|route|gateway|middleware|app|server)`)
	layerAPIPathRE    = regexp.MustCompile(`/(api|controllers|routes|handlers)/`)
	layerSvcFolderRE  = regexp.MustCompile(`^(services|business|use-cases|use_cases|orchestration|workflows)$`)
	layerSvcFileRE    = regexp.MustCompile(`(service|business|usecase|workflow|orchestration)`)
	layerSvcPathRE    = regexp.MustCompile(`/(services|business|use[-_]cases)/`)
	layerDataFolderRE = regexp.MustCompile(`^(repositories|repo|models|database|db|schemas|data|persistence)$`)
	layerDataFileRE   = regexp.MustCompile(`(repo|repository|model|entity|schema|mapper|query|dao)`)
	layerDataPathRE   = regexp.MustCompile(`/(repositories|models|database|db|persistence|entities)/`)
	layerUIFolderRE   = regexp.MustCompile(`^(components|views|pages|screens|ui|presentation|widgets)$`)
	layerUIFileRE     = regexp.MustCompile(`(component|view|page|screen|widget)`)
	layerUIPathRE     = regexp.MustCompile(`/(components|views|pages|screens|ui)/`)
	layerUtilFolderRE = regexp.MustCompile(`^(utils|helpers|common|lib|tools|config|constants)$`)
	layerUtilFileRE   = regexp.MustCompile(`(util|helper|common|constant|config|logger|validator|formatter)`)
	layerUtilPathRE   = regexp.MustCompile(`/(utils|helpers|common|config|constants)/`)
)

// getArchitectureLayer mirrors get_architecture_layer().
func getArchitectureLayer(rel string) string {
	relLower := strings.ToLower(rel)
	folder := rel
	if i := strings.IndexByte(rel, '/'); i >= 0 {
		folder = rel[:i]
	}
	filename := rel
	if i := strings.LastIndexByte(rel, '/'); i >= 0 {
		filename = rel[i+1:]
	}
	if i := strings.LastIndexByte(filename, '.'); i >= 0 {
		filename = filename[:i]
	}
	filename = strings.ToLower(filename)

	switch {
	case layerAPIFolderRE.MatchString(folder), layerAPIFileRE.MatchString(filename), layerAPIPathRE.MatchString(relLower):
		return "API"
	case layerSvcFolderRE.MatchString(folder), layerSvcFileRE.MatchString(filename), layerSvcPathRE.MatchString(relLower):
		return "Service"
	case layerDataFolderRE.MatchString(folder), layerDataFileRE.MatchString(filename), layerDataPathRE.MatchString(relLower):
		return "Data"
	case layerUIFolderRE.MatchString(folder), layerUIFileRE.MatchString(filename), layerUIPathRE.MatchString(relLower):
		return "UI"
	case layerUtilFolderRE.MatchString(folder), layerUtilFileRE.MatchString(filename), layerUtilPathRE.MatchString(relLower):
		return "Utility"
	}
	return ""
}
