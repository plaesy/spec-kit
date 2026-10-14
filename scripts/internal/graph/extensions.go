package graph

import (
	"path/filepath"
	"regexp"
	"strings"
)

// SourceExtensions is the single list of extensions treated as source code.
// includeExtRE and nodeTypeOf are both derived from it, so a new language is
// added here once. Each entry must also have a case in symbolsOf (symbols.go),
// or its files are collected but get no symbols.
var SourceExtensions = []string{
	".js", ".jsx", ".ts", ".tsx",
	".py", ".go", ".dart", ".java", ".kt", ".kts", ".swift",
	".c", ".h", ".cc", ".cpp", ".hpp", ".cxx", ".hh",
	".cs", ".rs", ".rb", ".php",
}

// nonSourceExtensions are collected into the graph but not labelled "source":
// Markdown documents and shell scripts. Each has a symbol extractor.
var nonSourceExtensions = []string{".md", ".sh", ".ps1"}

// includeExtRE matches every extension the graph collects: the source list plus
// the non-source list.
var includeExtRE = buildIncludeExtRE()

func buildIncludeExtRE() *regexp.Regexp {
	names := make([]string, 0, len(SourceExtensions)+len(nonSourceExtensions))
	for _, ext := range append(append([]string{}, nonSourceExtensions...), SourceExtensions...) {
		names = append(names, regexp.QuoteMeta(strings.TrimPrefix(ext, ".")))
	}
	return regexp.MustCompile(`(?i)\.(` + strings.Join(names, "|") + `)$`)
}

// isSourceExt reports whether rel's extension is in SourceExtensions.
func isSourceExt(rel string) bool {
	ext := strings.ToLower(filepath.Ext(rel))
	for _, s := range SourceExtensions {
		if ext == s {
			return true
		}
	}
	return false
}
