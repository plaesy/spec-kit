package semantic

import (
	"regexp"
	"strings"
)

// rustDocComments maps each Rust fn/struct/enum/trait name declared in a source
// file to its rustdoc comment, if any — the "///" outer line-doc run (or the
// "/** ... */" outer block-doc) written immediately above the item. This is the
// Rust counterpart of goDocComments/jsDocComments/pyDocComments.
//
// Regex/line-based rather than parsed, for the same reason as jsDocComments:
// Rust has no parser in the Go standard library, and pulling in a third-party
// AST parser just to recover doc-comment text would be disproportionate to the
// value (see .plaesy/decisions/embedding-based-semantic-duplicate-search.md).
// It mirrors the not-AST heuristic internal/graph/symbols.go already uses for
// every non-Go language.
//
// Which rustdoc form documents what (Rust Reference, "Comments") — the two line
// forms are NOT interchangeable, and getting this backwards is the whole
// hazard:
//
//   - "/// ..." is an OUTER line doc comment. It documents the item that
//     follows it, so it is the only form keyed to a graph.Node.Symbols name
//     here.
//   - "//! ..." is an INNER line doc comment. It documents the *parent* of the
//     comment — the enclosing module or crate — never the item below it, and
//     rustdoc requires it to sit at the top of the file. rustSymbols extracts
//     only fn/struct/enum/trait names (symbols.go's Rust regexes have no `mod`
//     or `impl` case), so a crate/module doc has no name in Node.Symbols to be
//     keyed against: emitting one would be an entry nodeText never looks up,
//     i.e. dead weight by construction.
//
// So "//!" is deliberately not emitted — and, more importantly, is recognised
// only to *terminate* a doc run. A file whose crate docs run straight into its
// first item (no blank line, which is perfectly legal and common) would
// otherwise have its whole crate header embedded as that item's doc, which is
// both wrong and poisonous signal. "/*! ... */" is excluded the same way.
//
// Also not treated as doc comments, per the same Reference: "//// ...",
// "/*** ... */" and "/**/" are ordinary comments (the outer forms are
// "exactly three slashes" and "exactly two asterisks"), so a section-divider
// or banner comment cannot be embedded as if it documented the item below it.
// "#[doc = \"...\"]" attributes are also out of scope: attributes precede the
// item like comments do, but they are expressions, not comments, and the
// regex-based extractor would have to parse them to tell `#[doc]` from every
// other attribute.
//
// The name regexes below are deliberately identical to the ones symbols.go
// feeds graph.Node.Symbols with (rustSymbols), and rustSymbolName reproduces
// rustSymbols' extraction logic — a doc comment keyed under a name that never
// reaches Node.Symbols would be silently useless, so the two must stay in
// step. symbols.go is the single source of truth; these are copies because
// they are unexported in another package.
//
// Doc comments are keyed by bare name, so two same-named items (a trait method
// and an impl of it, say) share one entry and the last one wins. That is the
// same granularity nodeText looks docs up at — Node.Symbols is a flat name list
// — so the two stay consistent rather than one being silently wrong.
//
// Returns nil when the file has no doc comments, matching goDocComments and
// jsDocComments. Extraction cannot fail, so there is no error to return.
func rustDocComments(content string) map[string]string {
	lines := strings.Split(content, "\n")

	var docs map[string]string
	for i, line := range lines {
		name, ok := rustSymbolName(line)
		if !ok {
			continue
		}
		text := rustDocText(rustDocBlock(lines, i-1))
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

// rustSymbolName returns the fn/struct/enum/trait name rustSymbols
// (internal/graph) would extract from line, so the returned key matches
// Node.Symbols exactly. The fn regex is checked first and wins — rustSymbols
// `continue`s after a fn hit instead of also testing the type regexes, and this
// must not turn one line into two candidate names.
func rustSymbolName(line string) (string, bool) {
	if m := symRustFuncRE.FindString(line); m != "" {
		return strings.TrimSpace(symRustFuncStp.ReplaceAllString(m, "")), true
	}
	if m := symRustTypeRE.FindString(line); m != "" {
		return strings.TrimSpace(symRustTypeStp.ReplaceAllString(m, "")), true
	}
	return "", false
}

// rustDocBlock returns the doc-comment block that ends on line end (end
// inclusive), or nil when no doc comment sits directly above end. end is the
// line before a declaration.
//
// The block is a run of "///" lines, optionally closed by a "/** ... */" block
// doc above them. A blank line, any code line, an inner doc comment ("//!" or
// "/*!") or an ordinary comment ends the scan — mirroring Go's rule that a doc
// comment must be contiguous with its declaration. That contiguity rule is what
// stops a comment belonging to some earlier statement from being attributed to
// the next item, and it is the only thing standing between a crate's "//!"
// header and the crate's first function.
func rustDocBlock(lines []string, end int) []string {
	if end < 0 {
		return nil
	}
	start := end + 1 // one past the block's first line
	for i := end; i >= 0; {
		t := strings.TrimSpace(lines[i])
		if rustDocLine(t) {
			start = i
			i--
			continue
		}
		if rustAttrLine(t) {
			// An attribute may sit between a doc comment and its item, and
			// #[derive(Debug)] above a struct is close to universal in real
			// Rust — treating it as code would end the doc run and silently
			// drop the documentation for nearly every derived type. Skip the
			// attribute and keep scanning; a multi-line attribute is skipped
			// as a unit by tracking bracket depth.
			start = i
			i--
			depth := strings.Count(t, "[") - strings.Count(t, "]")
			for ; i >= 0 && depth > 0; i-- {
				depth += strings.Count(lines[i], "[") - strings.Count(lines[i], "]")
			}
			continue
		}
		if strings.HasSuffix(t, "*/") {
			// Block doc: jump back to its opening "/**". Pairing a "*/" with
			// a distant opener would attribute unrelated text, so an opener
			// that is missing — or a "/*" (including the inner "/*!") closing
			// the block instead — is not a doc block we can attribute.
			open := i
			for open >= 0 {
				u := strings.TrimSpace(lines[open])
				if rustBlockDocOpen(u) {
					break
				}
				if strings.HasPrefix(u, "/*") {
					return nil
				}
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
	// Everything in the range has to be doc-comment content, not code that
	// merely sits between comment lines. For a pure "///" run that is already
	// guaranteed by the scan; this is the guard for the block-doc path, whose
	// interior lines are unrecognised until their opener is found.
	for _, l := range block {
		if !rustDocContent(strings.TrimSpace(l)) {
			return nil
		}
	}
	return block
}

// rustDocLine reports whether a trimmed line is an outer line doc comment:
// exactly three slashes, not four. "////" is a LINE_COMMENT per the Reference
// (the lexer takes the longer "//" "//" match), and treating a section-divider
// as a doc comment would embed the divider text as the item's documentation.
func rustDocLine(t string) bool {
	return strings.HasPrefix(t, "///") && !strings.HasPrefix(t, "////")
}

// rustBlockDocOpen reports whether a trimmed line opens an outer block doc
// comment: "/**" followed by neither "*" nor "/", since the Reference makes
// "/**/" and "/***" ordinary block comments.
func rustBlockDocOpen(t string) bool {
	rest, ok := strings.CutPrefix(t, "/**")
	if !ok || rest == "" {
		return false
	}
	return rest[0] != '*' && rest[0] != '/'
}

// rustAttrLine reports whether a trimmed line is an attribute. "#" at the start
// of a line in Rust is always an attribute ("#[...]" or the inner "#![...]"),
// never a comment — the language has no line comments at all. Attributes are
// allowed between a doc comment and the item it documents, so the scan steps
// over them.
//
// "#[doc = \"...\"]" is deliberately not harvested as doc text: it would mean
// parsing attribute expressions rather than comments, which is a different
// (and much larger) job than this extractor takes on. Such a declaration
// simply gets no doc comment from here.
func rustAttrLine(t string) bool {
	return strings.HasPrefix(t, "#")
}

// rustDocContent reports whether a trimmed line can be part of an outer doc
// comment: a "///" line, a "/**" opener, a line closing one, a leading "*"
// continuation line, or an attribute standing between the comment and its
// item. Blank lines deliberately fail this check — as they do in jsdocs.go,
// whose jsCommentLine is the same predicate with "//" — so a block doc whose
// interior lines are not the conventional "* " shape is dropped rather than
// risk embedding code as if it were documentation.
func rustDocContent(t string) bool {
	return rustDocLine(t) ||
		rustBlockDocOpen(t) ||
		strings.HasSuffix(t, "*/") ||
		strings.HasPrefix(t, "*") ||
		rustAttrLine(t)
}

// rustDocText turns a raw doc block into doc text: strips the "///" and
// "/** */" markers and the leading "*" continuation characters, and drops blank
// lines (a bare "///" separator included) — the same shape go/parser's Doc.Text()
// produces, so nodeText's firstLine() truncation behaves the same across
// languages. Attribute lines are dropped outright: they are accepted inside a
// block so the scan can step over them, but they are not documentation and must
// never reach the embedding. Reuses pydocs.go's joinDocLines for the same
// reason.
func rustDocText(block []string) string {
	kept := make([]string, 0, len(block))
	for _, l := range block {
		t := strings.TrimSpace(l)
		if rustAttrLine(t) {
			continue
		}
		t = strings.TrimPrefix(t, "///")
		t = strings.TrimPrefix(t, "/**")
		t = strings.TrimSuffix(t, "*/")
		t = strings.TrimSpace(t)
		t = strings.TrimSpace(strings.TrimPrefix(t, "*"))
		if t != "" {
			kept = append(kept, t)
		}
	}
	return joinDocLines(kept)
}

// Copies of internal/graph/symbols.go's Rust regexes (rustSymbols), kept
// byte-for-byte identical so the names they match are the names that end up in
// graph.Node.Symbols; see rustSymbolName. Watch the func-strip regex's "|fn "
// alternative: without it a column-0 "fn foo" would keep its keyword, since
// there is no whitespace before "fn" for the greedy ".* " branch to consume.
var (
	symRustFuncRE  = regexp.MustCompile(`^[ \t]*(pub(\([^)]*\))?[ \t]+)?(async[ \t]+)?(unsafe[ \t]+)?fn[ \t]+[A-Za-z0-9_]+`)
	symRustFuncStp = regexp.MustCompile(`.*[ \t]fn[ \t]+|^fn[ \t]+`)
	symRustTypeRE  = regexp.MustCompile(`^[ \t]*(pub(\([^)]*\))?[ \t]+)?(struct|enum|trait)[ \t]+[A-Za-z0-9_]+`)
	symRustTypeStp = regexp.MustCompile(`^.*(struct|enum|trait)[ \t]+`)
)
