package graph

import (
	"regexp"
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
)

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

func stripMDHeading(line string) string {
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}
	s := strings.TrimLeft(line[i:], " \t")
	return strings.TrimRight(s, " \t")
}
