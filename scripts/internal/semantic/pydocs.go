package semantic

import (
	"regexp"
	"strings"
)

// pyDocComments maps each Python function/class name declared in a source
// file to its docstring, if any — the first triple-quoted string (double or
// single quotes) in its body. This is the Python counterpart of
// goDocComments.
//
// Regex/line-based rather than parsed, for the same reason as jsDocComments:
// no Python AST parser in the Go standard library, and the value here is the
// doc text, not a syntax tree. Python's docstring convention is *inside* the
// block (unlike Go's and JS's *above* the declaration), so this reads the
// line(s) after the matched declaration rather than before it.
//
// The name regexes below are deliberately identical to the ones symbols.go
// feeds graph.Node.Symbols with (pySymbols), and pySymbolName reproduces
// pySymbols' extraction logic — a docstring keyed under a name that never
// reaches Node.Symbols would be silently useless, so the two must stay in
// step. symbols.go is the single source of truth; these are copies because
// they are unexported in another package.
//
// Returns nil when the file has no docstrings, matching goDocComments.
// Extraction cannot fail, so there is no error to return.
func pyDocComments(content string) map[string]string {
	lines := strings.Split(content, "\n")

	var docs map[string]string
	for i, line := range lines {
		name, ok := pySymbolName(line)
		if !ok {
			continue
		}
		text := pyDocstring(lines, i)
		if text == "" {
			continue
		}
		if docs == nil {
			docs = make(map[string]string)
		}
		docs[name] = text
	}
	return docs
}

// pySymbolName returns the function/class name pySymbols (internal/graph)
// would extract from line, so the returned key matches Node.Symbols exactly.
// Function declarations are checked first and win, exactly as pySymbols does.
func pySymbolName(line string) (string, bool) {
	if m := symPyDefRE.FindString(line); m != "" {
		s := strings.TrimSuffix(m, "(")
		return symPyDefStrip.ReplaceAllString(s, ""), true
	}
	if m := symPyClassRE.FindString(line); m != "" {
		return symPyClassStrp.ReplaceAllString(m, ""), true
	}
	return "", false
}

// pyDocstring returns the docstring of the def/class declared on line i, or
// "" if it has none. The header has to end with ":" (ignoring a trailing
// comment); after it, the first line that isn't blank or a "#" comment must
// open a triple-quoted string, otherwise the body simply has no docstring.
func pyDocstring(lines []string, i int) string {
	if i >= len(lines)-1 || !pyHeaderEnd(lines[i]) {
		return ""
	}
	for j := i + 1; j < len(lines); j++ {
		t := strings.TrimSpace(lines[j])
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		delim, first, ok := pyTripleQuote(t)
		if !ok {
			return ""
		}
		kept := []string{strings.TrimSpace(first)}
		if end := strings.Index(first, delim); end >= 0 {
			kept[0] = strings.TrimSpace(first[:end])
			return joinDocLines(kept)
		}
		for k := j + 1; k < len(lines); k++ {
			if end := strings.Index(lines[k], delim); end >= 0 {
				if pre := strings.TrimSpace(lines[k][:end]); pre != "" {
					kept = append(kept, pre)
				}
				return joinDocLines(kept)
			}
			if body := strings.TrimSpace(lines[k]); body != "" {
				kept = append(kept, body)
			}
		}
		// Unterminated docstring: not usable text, so treat it as absent
		// rather than embedding the rest of the file.
		return ""
	}
	return ""
}

// pyHeaderEnd reports whether a def/class header line ends with the ":"
// that opens its body. A "#" only counts as a comment start when a colon
// already appeared before it — a docstring body may legitimately contain
// "#" (an issue reference, say), and truncating there would corrupt it.
func pyHeaderEnd(line string) bool {
	if h := strings.IndexByte(line, '#'); h >= 0 {
		if !strings.Contains(line[:h], ":") {
			return false
		}
		line = line[:h]
	}
	return strings.HasSuffix(strings.TrimSpace(line), ":")
}

// pyStringPrefixes are the legal Python string prefixes that may precede a
// triple-quoted docstring (r"""...""", f"""...""", rb"""...""", ...).
const pyStringPrefixes = "rRbBuUfF"

// pyTripleQuote splits a triple-quoted string opener off the start of an
// already-trimmed line, returning the delimiter and whatever follows it on
// that same line.
func pyTripleQuote(t string) (delim, rest string, ok bool) {
	for rest = t; len(rest) > 0 && strings.IndexByte(pyStringPrefixes, rest[0]) >= 0; {
		rest = rest[1:]
	}
	if strings.HasPrefix(rest, `"""`) {
		return `"""`, rest[3:], true
	}
	if strings.HasPrefix(rest, `'''`) {
		return `'''`, rest[3:], true
	}
	return "", "", false
}

// joinDocLines joins already-cleaned doc lines with newlines, dropping any
// that came out empty — the same shape go/parser's Doc.Text() produces, so
// nodeText's firstLine() truncation behaves the same across languages.
func joinDocLines(lines []string) string {
	kept := make([]string, 0, len(lines))
	for _, l := range lines {
		if l != "" {
			kept = append(kept, l)
		}
	}
	return strings.Join(kept, "\n")
}

// Copies of internal/graph/symbols.go's Python regexes (pySymbols). Kept
// byte-for-byte identical so the names they match are the names that end up
// in graph.Node.Symbols; see pySymbolName.
var (
	symPyDefRE     = regexp.MustCompile(`^[ \t]*def[ \t]+[A-Za-z0-9_]+[ \t]*\(`)
	symPyDefStrip  = regexp.MustCompile(`.*def[ \t]+`)
	symPyClassRE   = regexp.MustCompile(`^[ \t]*class[ \t]+[A-Za-z0-9_]+`)
	symPyClassStrp = regexp.MustCompile(`.*class[ \t]+`)
)
