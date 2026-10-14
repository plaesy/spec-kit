package semantic

import (
	"regexp"
	"strings"
)

// jsDocComments maps each JavaScript/TypeScript function/class name declared
// in a source file to its doc comment, if any — the "/** ... */" JSDoc
// block, or a run of "//" lines, written immediately above the declaration.
// This is the JS/TS counterpart of goDocComments.
//
// Regex-based rather than parsed, unlike the Go extractor: Go has go/parser
// in the standard library but JS/TS does not, and pulling in a third-party
// AST parser just to recover doc-comment text would be disproportionate to
// the value (see .plaesy/decisions/embedding-based-semantic-duplicate-search.md).
// It mirrors the not-AST heuristic internal/graph/symbols.go already uses for
// every non-Go language.
//
// The name regexes below are deliberately identical to the ones symbols.go
// feeds graph.Node.Symbols with (jsSymbols), and jsSymbolName reproduces
// jsSymbols' extraction logic — a doc comment keyed under a name that never
// reaches Node.Symbols would be silently useless, so the two must stay in
// step. symbols.go is the single source of truth; these are copies because
// they are unexported in another package.
//
// Returns nil when the file has no doc comments, matching goDocComments.
// Extraction cannot fail, so there is no error to return.
func jsDocComments(content string) map[string]string {
	lines := strings.Split(content, "\n")

	var docs map[string]string
	for i, line := range lines {
		name, ok := jsSymbolName(line)
		if !ok {
			continue
		}
		text := jsDocText(jsCommentBlock(lines, i-1))
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

// jsSymbolName returns the function/class name jsSymbols (internal/graph)
// would extract from line, so the returned key matches Node.Symbols exactly.
// Function declarations are checked first and win, exactly as jsSymbols does.
func jsSymbolName(line string) (string, bool) {
	if m := symJSFuncRE.FindString(line); m != "" {
		s := strings.TrimSuffix(m, "(")
		return symJSFuncStrip.ReplaceAllString(s, ""), true
	}
	if m := symJSClassRE.FindString(line); m != "" {
		return symJSClassStrp.ReplaceAllString(m, ""), true
	}
	return "", false
}

// jsCommentBlock returns the comment block that ends on line end (end
// inclusive), or nil when no comment sits directly above end. end is
// expected to be the line before a declaration.
//
// A comment block is either a run of "//" lines or a /* ... */ block
// comment, which may span lines. A blank line or any code line ends the
// scan, mirroring Go's rule that a doc comment must be contiguous with its
// declaration — that's what stops a comment belonging to some earlier
// statement from being attributed to the next one.
func jsCommentBlock(lines []string, end int) []string {
	if end < 0 {
		return nil
	}
	start := end + 1 // one past the block's first line
	for i := end; i >= 0; {
		t := strings.TrimSpace(lines[i])
		if jsCommentLine(t) {
			start = i
			i--
			continue
		}
		if strings.HasSuffix(t, "*/") {
			// Block comment: jump back to its opening "/*". An unterminated
			// "*/" with no "/*" above it isn't a comment block we can
			// attribute, so give up rather than pair it with a distant one.
			open := i
			for open >= 0 && !strings.Contains(lines[open], "/*") {
				open--
			}
			if open < 0 {
				return nil
			}
			start = open
			i = open - 1
			continue
		}
		break
	}
	block := lines[start : end+1]
	// Everything in the range has to be comment content, not code that
	// merely sits between comment lines.
	for _, l := range block {
		if !jsCommentLine(strings.TrimSpace(l)) {
			return nil
		}
	}
	return block
}

// jsCommentLine reports whether a trimmed line is comment content: a "//"
// line, a line opening a /* */ block, a line closing one, or JSDoc's "*"
// continuation line. None of these can be code, which is what makes
// jsCommentBlock's range safe to attribute to the declaration below it.
//
// The block-comment cases test prefix/suffix, not containment: a code line
// like `let x = 1; /* why` contains "/*" but is not a comment, and treating it
// as one would embed the code itself into the embedding as if it were a doc
// comment.
func jsCommentLine(t string) bool {
	return strings.HasPrefix(t, "//") ||
		strings.HasPrefix(t, "/*") ||
		strings.HasSuffix(t, "*/") ||
		strings.HasPrefix(t, "*")
}

// jsDocText turns a raw comment block into doc text: strips the "//" and
// "/* */" markers and JSDoc's leading "*" continuation characters, and
// drops blank lines — the same shape go/parser's Doc.Text() produces, so
// nodeText's firstLine() truncation behaves the same for both.
func jsDocText(block []string) string {
	if len(block) == 0 {
		return ""
	}
	kept := make([]string, 0, len(block))
	for _, l := range block {
		t := strings.TrimSpace(l)
		// "///" first, before "//". It is the primary doc marker for Rust, C#,
		// Swift and Dart, and a valid triple-slash directive in TypeScript, so
		// it reaches here for every language that uses it. Stripping only "//"
		// left a stray "/" at the front of the embedded text ("/ Target docs."),
		// which quietly polluted the vector in all of them. Languages with
		// their own pre-normalisation (cfamilySlashDocLines,
		// swiftSlashDocLines) already rewrote "///" to "//" by this point, so
		// this is simply a no-op for them.
		t = strings.TrimPrefix(t, "///")
		t = strings.TrimPrefix(t, "//")
		t = strings.TrimPrefix(t, "/*")
		t = strings.TrimSuffix(t, "*/")
		t = strings.TrimSpace(t)
		t = strings.TrimSpace(strings.TrimPrefix(t, "*"))
		if t != "" {
			kept = append(kept, t)
		}
	}
	return strings.Join(kept, "\n")
}

// Copies of internal/graph/symbols.go's JS/TS regexes (jsSymbols). Kept
// byte-for-byte identical so the names they match are the names that end up
// in graph.Node.Symbols; see jsSymbolName.
var (
	symJSFuncRE    = regexp.MustCompile(`^[ \t]*(export[ \t]+)?function[ \t]+[A-Za-z0-9_$]+[ \t]*\(`)
	symJSFuncStrip = regexp.MustCompile(`.*function[ \t]+`)
	symJSClassRE   = regexp.MustCompile(`^[ \t]*(export[ \t]+)?class[ \t]+[A-Za-z0-9_$]+`)
	symJSClassStrp = regexp.MustCompile(`.*class[ \t]+`)
)
