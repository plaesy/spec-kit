package graph

import (
	"regexp"
	"sort"
	"strings"
)

// symbolsOf is a Go port of plaesy-graph-symbols.awk: extracts a per-line,
// per-extension "what's actually inside this file" symbol list (function
// names, class names, markdown headings). Mirrors plaesy-graph.ps1's
// $SymbolPatterns table so both the bash and Go builders surface the same
// symbols for a given node.
func symbolsOf(rel, content string) []string {
	ext := extOf(rel)
	var pat *regexp.Regexp
	var extract func(line, matched string) string

	switch ext {
	case ".ps1":
		pat = symPS1FuncRE
		extract = func(line, matched string) string {
			return strings.TrimSpace(symPS1FuncStripRE.ReplaceAllString(matched, ""))
		}
	case ".sh":
		return shSymbols(content)
	case ".js", ".jsx", ".ts", ".tsx":
		return jsSymbols(content)
	case ".py":
		return pySymbols(content)
	case ".go":
		return goSymbols(content)
	case ".md":
		return mdSymbols(content)
	case ".rs":
		return rustSymbols(content)
	case ".dart":
		return dartSymbols(content)
	case ".java":
		return javaSymbols(content)
	case ".kt", ".kts":
		return kotlinSymbols(content)
	case ".cs":
		return csharpSymbols(content)
	case ".c", ".h", ".cpp", ".hpp", ".cc", ".hh", ".cxx":
		return cFamilySymbols(content)
	case ".swift":
		return swiftSymbols(content)
	case ".rb":
		return rubySymbols(content)
	case ".php":
		return phpSymbols(content)
	default:
		return nil
	}

	var out []string
	for _, line := range strings.Split(content, "\n") {
		m := pat.FindString(line)
		if m == "" {
			continue
		}
		out = append(out, extract(line, m))
	}
	return out
}

var (
	symPS1FuncRE      = regexp.MustCompile(`^[ \t]*function[ \t]+[A-Za-z0-9_-]+`)
	symPS1FuncStripRE = regexp.MustCompile(`^[ \t]*function[ \t]+`)

	symShFuncRE     = regexp.MustCompile(`^[ \t]*function[ \t]+[A-Za-z0-9_]+`)
	symShFuncStrip  = regexp.MustCompile(`^[ \t]*function[ \t]+`)
	symShBraceRE    = regexp.MustCompile(`^[ \t]*[A-Za-z0-9_]+[ \t]*\(\)[ \t]*\{`)
	symShBraceStrip = regexp.MustCompile(`[ \t]*\(\).*`)

	symJSFuncRE    = regexp.MustCompile(`^[ \t]*(export[ \t]+)?function[ \t]+[A-Za-z0-9_$]+[ \t]*\(`)
	symJSFuncStrip = regexp.MustCompile(`.*function[ \t]+`)
	symJSClassRE   = regexp.MustCompile(`^[ \t]*(export[ \t]+)?class[ \t]+[A-Za-z0-9_$]+`)
	symJSClassStrp = regexp.MustCompile(`.*class[ \t]+`)

	symPyDefRE     = regexp.MustCompile(`^[ \t]*def[ \t]+[A-Za-z0-9_]+[ \t]*\(`)
	symPyDefStrip  = regexp.MustCompile(`.*def[ \t]+`)
	symPyClassRE   = regexp.MustCompile(`^[ \t]*class[ \t]+[A-Za-z0-9_]+`)
	symPyClassStrp = regexp.MustCompile(`.*class[ \t]+`)

	symGoFuncRE       = regexp.MustCompile(`^func[ \t]+(\([^)]*\)[ \t]*)?[A-Za-z0-9_]+[ \t]*\(`)
	symGoFuncStripRcv = regexp.MustCompile(`^\([^)]*\)[ \t]*`)

	symMdHeadingRE = regexp.MustCompile(`^#{1,3}[ \t]+.+$`)

	symRustFuncRE  = regexp.MustCompile(`^[ \t]*(pub(\([^)]*\))?[ \t]+)?(async[ \t]+)?(unsafe[ \t]+)?fn[ \t]+[A-Za-z0-9_]+`)
	symRustFuncStp = regexp.MustCompile(`.*[ \t]fn[ \t]+|^fn[ \t]+`)
	symRustTypeRE  = regexp.MustCompile(`^[ \t]*(pub(\([^)]*\))?[ \t]+)?(struct|enum|trait)[ \t]+[A-Za-z0-9_]+`)
	symRustTypeStp = regexp.MustCompile(`^.*(struct|enum|trait)[ \t]+`)

	symDartClassRE = regexp.MustCompile(`^[ \t]*(abstract[ \t]+)?class[ \t]+[A-Za-z0-9_]+`)
	symDartClsStrp = regexp.MustCompile(`.*class[ \t]+`)

	symJavaClassRE  = regexp.MustCompile(`^[ \t]*(?:(?:public|private|protected|static|final|abstract)[ \t]+)*(class|interface|enum)[ \t]+[A-Za-z0-9_]+`)
	symJavaClassStp = regexp.MustCompile(`^.*(class|interface|enum)[ \t]+`)

	symKotlinFunRE   = regexp.MustCompile(`^[ \t]*(?:(?:public|private|internal|protected|suspend|inline|override)[ \t]+)*fun[ \t]+[A-Za-z0-9_]+`)
	symKotlinFunStp  = regexp.MustCompile(`.*fun[ \t]+`)
	symKotlinTypeRE  = regexp.MustCompile(`^[ \t]*(?:(?:public|private|internal|protected|abstract|open|sealed|data)[ \t]+)*(class|interface|object)[ \t]+[A-Za-z0-9_]+`)
	symKotlinTypeStp = regexp.MustCompile(`^.*(class|interface|object)[ \t]+`)

	symCSharpTypeRE  = regexp.MustCompile(`^[ \t]*(?:(?:public|private|protected|internal|static|abstract|sealed|partial)[ \t]+)*(class|interface|struct|enum)[ \t]+[A-Za-z0-9_]+`)
	symCSharpTypeStp = regexp.MustCompile(`^.*(class|interface|struct|enum)[ \t]+`)

	symCFamilyTypeRE  = regexp.MustCompile(`^[ \t]*(class|struct)[ \t]+[A-Za-z0-9_]+`)
	symCFamilyTypeStp = regexp.MustCompile(`^.*(class|struct)[ \t]+`)

	symSwiftFuncRE  = regexp.MustCompile(`^[ \t]*(?:(?:public|private|internal|fileprivate|open|static|override|mutating)[ \t]+)*func[ \t]+[A-Za-z0-9_]+`)
	symSwiftFuncStp = regexp.MustCompile(`.*func[ \t]+`)
	symSwiftTypeRE  = regexp.MustCompile(`^[ \t]*(?:(?:public|private|internal|fileprivate|open|final)[ \t]+)*(class|struct|enum|protocol)[ \t]+[A-Za-z0-9_]+`)
	symSwiftTypeStp = regexp.MustCompile(`^.*(class|struct|enum|protocol)[ \t]+`)

	symRubyDefRE   = regexp.MustCompile(`^[ \t]*def[ \t]+(self\.)?[A-Za-z0-9_?!=]+`)
	symRubyDefStp  = regexp.MustCompile(`.*def[ \t]+`)
	symRubyTypeRE  = regexp.MustCompile(`^[ \t]*(class|module)[ \t]+[A-Za-z0-9_:]+`)
	symRubyTypeStp = regexp.MustCompile(`^.*(class|module)[ \t]+`)

	symPHPFuncRE  = regexp.MustCompile(`^[ \t]*(?:(?:public|private|protected|static)[ \t]+)*function[ \t]+[A-Za-z0-9_]+`)
	symPHPFuncStp = regexp.MustCompile(`.*function[ \t]+`)
	symPHPTypeRE  = regexp.MustCompile(`^[ \t]*(?:(?:abstract|final)[ \t]+)?(class|interface|trait)[ \t]+[A-Za-z0-9_]+`)
	symPHPTypeStp = regexp.MustCompile(`^.*(class|interface|trait)[ \t]+`)

	// symCLikeFuncRE is a best-effort "typed-return method" detector for
	// languages with no function/method keyword (Dart, Java, C#, C/C++): two
	// or more identifier-like tokens followed by "(...)" and a same-line "{".
	// Requiring two tokens (return type + name) is what keeps it from
	// matching control-flow lines like "if (x) {" (only one token).
	// Requiring a same-line "{" is a known limitation: a brace on the next
	// line is missed, consistent with the rest of this file's heuristic,
	// not-AST approach.
	symCLikeFuncRE = regexp.MustCompile(`^[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*(?:<[^()]*>)?(?:\[\])?[ \t]+)+([A-Za-z_][A-Za-z0-9_]*)[ \t]*\([^;{}]*\)[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*[ \t]*)*\{[ \t]*$`)

	// symCLikeFuncNoBraceRE is symCLikeFuncRE's Allman-style counterpart: the
	// same declaration shape, but ending the line right after the ")" (plus
	// optional trailing modifiers) instead of requiring "{" on the same
	// line. Used only with a same-line-content next-line-"{" lookahead
	// (cLikeFuncHits), never on its own — on its own it would also match a
	// prototype/interface declaration that ends in ";" on the next line.
	symCLikeFuncNoBraceRE = regexp.MustCompile(`^[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*(?:<[^()]*>)?(?:\[\])?[ \t]+)+([A-Za-z_][A-Za-z0-9_]*)[ \t]*\([^;{}]*\)[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*[ \t]*)*$`)
)

// symHit pairs an extracted symbol with the line it was found on, so callers
// that merge multiple passes (class/type scan + lookahead function scan) can
// restore source order instead of reporting types-then-funcs regardless of
// where they actually appear.
type symHit struct {
	line int
	name string
}

// cLikeFuncHits finds C-like method declarations in languages with no
// function keyword (Dart, Java, C#, C/C++), matching both "same-line brace"
// (`void run() {`) and "Allman-style" (`void run()` with `{` alone on the
// next non-blank line).
func cLikeFuncHits(lines []string) []symHit {
	var hits []symHit
	for i, line := range lines {
		if m := symCLikeFuncRE.FindStringSubmatch(line); m != nil {
			hits = append(hits, symHit{i, m[1]})
			continue
		}
		m := symCLikeFuncNoBraceRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		j := i + 1
		for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
			j++
		}
		if j < len(lines) && strings.TrimSpace(lines[j]) == "{" {
			hits = append(hits, symHit{i, m[1]})
		}
	}
	return hits
}

// mergeHits combines pre-collected (line, name) hits with the results of
// cLikeFuncHits and returns the names in source order.
func mergeHits(typeHits []symHit, lines []string) []string {
	hits := append(append([]symHit(nil), typeHits...), cLikeFuncHits(lines)...)
	sort.SliceStable(hits, func(a, b int) bool { return hits[a].line < hits[b].line })
	out := make([]string, len(hits))
	for i, h := range hits {
		out[i] = h.name
	}
	return out
}

func shSymbols(content string) []string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		if m := symShFuncRE.FindString(line); m != "" {
			out = append(out, strings.TrimSpace(symShFuncStrip.ReplaceAllString(m, "")))
			continue
		}
		if m := symShBraceRE.FindString(line); m != "" {
			s := strings.TrimLeft(m, " \t")
			s = symShBraceStrip.ReplaceAllString(s, "")
			out = append(out, s)
		}
	}
	return out
}

func jsSymbols(content string) []string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		if m := symJSFuncRE.FindString(line); m != "" {
			s := strings.TrimSuffix(m, "(")
			s = symJSFuncStrip.ReplaceAllString(s, "")
			out = append(out, s)
			continue
		}
		if m := symJSClassRE.FindString(line); m != "" {
			out = append(out, symJSClassStrp.ReplaceAllString(m, ""))
		}
	}
	return out
}

func pySymbols(content string) []string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		if m := symPyDefRE.FindString(line); m != "" {
			s := strings.TrimSuffix(m, "(")
			s = symPyDefStrip.ReplaceAllString(s, "")
			out = append(out, s)
			continue
		}
		if m := symPyClassRE.FindString(line); m != "" {
			out = append(out, symPyClassStrp.ReplaceAllString(m, ""))
		}
	}
	return out
}

func goSymbols(content string) []string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		m := symGoFuncRE.FindString(line)
		if m == "" {
			continue
		}
		s := strings.TrimSuffix(m, "(")
		s = strings.TrimPrefix(s, "func")
		s = strings.TrimLeft(s, " \t")
		s = symGoFuncStripRcv.ReplaceAllString(s, "")
		out = append(out, s)
	}
	return out
}

func mdSymbols(content string) []string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		m := symMdHeadingRE.FindString(line)
		if m == "" {
			continue
		}
		out = append(out, stripMDHeading(m))
	}
	return out
}

func rustSymbols(content string) []string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		if m := symRustFuncRE.FindString(line); m != "" {
			out = append(out, strings.TrimSpace(symRustFuncStp.ReplaceAllString(m, "")))
			continue
		}
		if m := symRustTypeRE.FindString(line); m != "" {
			out = append(out, strings.TrimSpace(symRustTypeStp.ReplaceAllString(m, "")))
		}
	}
	return out
}

func dartSymbols(content string) []string {
	lines := strings.Split(content, "\n")
	var typeHits []symHit
	for i, line := range lines {
		if m := symDartClassRE.FindString(line); m != "" {
			typeHits = append(typeHits, symHit{i, strings.TrimSpace(symDartClsStrp.ReplaceAllString(m, ""))})
		}
	}
	return mergeHits(typeHits, lines)
}

func javaSymbols(content string) []string {
	lines := strings.Split(content, "\n")
	var typeHits []symHit
	for i, line := range lines {
		if m := symJavaClassRE.FindString(line); m != "" {
			typeHits = append(typeHits, symHit{i, strings.TrimSpace(symJavaClassStp.ReplaceAllString(m, ""))})
		}
	}
	return mergeHits(typeHits, lines)
}

func kotlinSymbols(content string) []string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		if m := symKotlinFunRE.FindString(line); m != "" {
			out = append(out, strings.TrimSpace(symKotlinFunStp.ReplaceAllString(m, "")))
			continue
		}
		if m := symKotlinTypeRE.FindString(line); m != "" {
			out = append(out, strings.TrimSpace(symKotlinTypeStp.ReplaceAllString(m, "")))
		}
	}
	return out
}

func csharpSymbols(content string) []string {
	lines := strings.Split(content, "\n")
	var typeHits []symHit
	for i, line := range lines {
		if m := symCSharpTypeRE.FindString(line); m != "" {
			typeHits = append(typeHits, symHit{i, strings.TrimSpace(symCSharpTypeStp.ReplaceAllString(m, ""))})
		}
	}
	return mergeHits(typeHits, lines)
}

func cFamilySymbols(content string) []string {
	lines := strings.Split(content, "\n")
	var typeHits []symHit
	for i, line := range lines {
		if m := symCFamilyTypeRE.FindString(line); m != "" {
			typeHits = append(typeHits, symHit{i, strings.TrimSpace(symCFamilyTypeStp.ReplaceAllString(m, ""))})
		}
	}
	return mergeHits(typeHits, lines)
}

func swiftSymbols(content string) []string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		if m := symSwiftFuncRE.FindString(line); m != "" {
			out = append(out, strings.TrimSpace(symSwiftFuncStp.ReplaceAllString(m, "")))
			continue
		}
		if m := symSwiftTypeRE.FindString(line); m != "" {
			out = append(out, strings.TrimSpace(symSwiftTypeStp.ReplaceAllString(m, "")))
		}
	}
	return out
}

func rubySymbols(content string) []string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		if m := symRubyDefRE.FindString(line); m != "" {
			out = append(out, strings.TrimSpace(symRubyDefStp.ReplaceAllString(m, "")))
			continue
		}
		if m := symRubyTypeRE.FindString(line); m != "" {
			out = append(out, strings.TrimSpace(symRubyTypeStp.ReplaceAllString(m, "")))
		}
	}
	return out
}

func phpSymbols(content string) []string {
	var out []string
	for _, line := range strings.Split(content, "\n") {
		if m := symPHPFuncRE.FindString(line); m != "" {
			out = append(out, strings.TrimSpace(symPHPFuncStp.ReplaceAllString(m, "")))
			continue
		}
		if m := symPHPTypeRE.FindString(line); m != "" {
			out = append(out, strings.TrimSpace(symPHPTypeStp.ReplaceAllString(m, "")))
		}
	}
	return out
}

func stripMDHeading(line string) string {
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}
	s := strings.TrimLeft(line[i:], " \t")
	return strings.TrimRight(s, " \t")
}
