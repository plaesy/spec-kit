package semantic

import (
	"regexp"
	"strings"
)

// Ruby, Shell and PowerShell are the three hash-comment languages symbols.go
// extracts symbols for: their documentation lives in "#" lines written above
// the declaration (plus PowerShell's "<# ... #>" block and Ruby's
// "=begin"/"=end" block), not in a C-style "/* */" or '"""' block the way
// jsdocs.go and pydocs.go each had to handle. So the shape is shared and the
// language-specific part is only the symbol-name regexes and which extra block
// forms to honour — hashCommentLinesAbove/hashDocText below are that shared
// scan, and the three *DocComments functions are thin wrappers over it.
//
// Like jsdocs.go, these are regex-based rather than parsed: no Ruby/Shell/
// PowerShell parser in the Go standard library, and the value here is the doc
// text, not a syntax tree. Being permissive about *which* comment lines count
// matches go/parser, which attaches any contiguous comment block above a
// declaration as its doc comment.

// hashBlock is a multi-line block-comment form, expressed as its opener and
// closer delimiters — PowerShell's "<#" and "#>". The zero value means the
// language has single-line "#" comments only, which is the case for Ruby and
// Shell.
//
// Ruby's "=begin"/"=end" block is deliberately not one of these: its body is
// bare prose with no "#" on any line, so the per-line prefix test below would
// reject every line of it. It gets its own scan (hashRubyEmbed) instead.
type hashBlock struct {
	open  string
	close string
}

// hashNoBlock and hashPSBlock are the only two values hashBlock is called with,
// named so which comment form is in play is readable at each call site.
var (
	hashNoBlock = hashBlock{}
	hashPSBlock = hashBlock{open: "<#", close: "#>"}
)

// hasOpen reports whether an already-trimmed line opens a block comment (the
// single-line "<# text #>" form counts as opening).
func (b hashBlock) hasOpen(t string) bool {
	return b.open != "" && strings.HasPrefix(t, b.open)
}

// hasClose reports whether an already-trimmed line closes a block comment.
func (b hashBlock) hasClose(t string) bool {
	return b.close != "" && strings.HasPrefix(t, b.close)
}

// hashCommentLinesAbove returns the comment block whose last line is end (end
// inclusive), or an empty range when no comment sits directly above end. end is
// expected to be the line before a declaration.
//
// A block is a run of "#" lines, extended upwards past a "<#"-prefixed line
// when the language has that block form. A closing delimiter with no opener
// above it isn't a block that can be attributed to anything, so the scan gives
// up rather than pairing it with a distant opener. A blank line or any code
// line ends the scan, mirroring the rule go/parser applies: a doc comment must
// be contiguous with its declaration, which is what stops a comment belonging
// to some earlier statement from being attributed to the next one.
//
// The result is not yet known to be comment content — hashDocText checks that,
// since this scan is "contiguous with the declaration", not "all comments".
func hashCommentLinesAbove(lines []string, end int, block hashBlock) []string {
	if end < 0 {
		return nil
	}
	start := end + 1 // one past the block's first line
	for i := end; i >= 0; {
		t := strings.TrimSpace(lines[i])
		if block.hasClose(t) {
			// Block comment: jump back to its opening "<#".
			open := i
			for open >= 0 && !block.hasOpen(strings.TrimSpace(lines[open])) {
				open--
			}
			if open < 0 {
				return nil
			}
			start = open
			i = open - 1
			continue
		}
		if block.hasOpen(t) || hashCommentLine(t) {
			start = i
			i--
			continue
		}
		break
	}
	return lines[start : end+1]
}

// hashCommentLine reports whether an already-trimmed line is "#" comment
// content, shared by all three languages — that is the whole reason this scan
// can be shared.
//
// "#!" is excluded: it is an interpreter directive, not prose. A shell or Ruby
// script whose first line is "#!/usr/bin/env bash" would otherwise have that
// shebang attributed to the first function or def in the file, embedding an
// interpreter path as if it described it.
func hashCommentLine(t string) bool {
	return strings.HasPrefix(t, "#") && !strings.HasPrefix(t, "#!")
}

// hashDocText turns a raw comment range into doc text: strips the "#" markers
// (and a block form's "<#"/"#>" delimiters), drops blank lines, and joins what
// is left with newlines — the same shape go/parser's Doc.Text() produces, so
// nodeText's firstLine() truncation behaves the same across languages.
//
// Returns "" when any line in the range is not comment content. The scan above
// only establishes contiguity, and code sitting between comment lines must not
// be embedded as if it were prose; tag lines ("@param" in JSDoc, ".SYNOPSIS"
// in a PowerShell block) are kept verbatim, exactly as jsDocText keeps them, so
// that firstLine() decides what reaches the embedding rather than this function
// guessing which of a comment's lines are metadata.
func hashDocText(lines []string, kind hashBlock) string {
	kept := make([]string, 0, len(lines))
	inBlock := false
	for _, l := range lines {
		t := strings.TrimSpace(l)
		switch {
		case kind.hasClose(t):
			inBlock = false
		case inBlock:
			// Everything between "<#" and "#>" is comment content by
			// definition, so it is kept as written.
			if t != "" {
				kept = append(kept, t)
			}
		case kind.hasOpen(t):
			inBlock = true
			rest := strings.TrimSpace(strings.TrimPrefix(t, kind.open))
			if strings.HasSuffix(rest, kind.close) {
				rest = strings.TrimSpace(strings.TrimSuffix(rest, kind.close))
				inBlock = false
			}
			if rest != "" {
				kept = append(kept, rest)
			}
		case hashCommentLine(t):
			if text := strings.TrimSpace(strings.TrimPrefix(t, "#")); text != "" {
				kept = append(kept, text)
			}
		default:
			return ""
		}
	}
	return joinDocLines(kept)
}

// hashRubyEmbed returns the doc text of the "=begin"/"=end" block comment whose
// last line is end, or "" when end is not a closing line. Ruby accepts this
// second comment form alongside "#" lines, for documentation that needs blank
// lines or RDoc markup, which a "#" run can't carry.
//
// An "=end" with no "=begin" above it is treated as absent rather than paired
// with a distant opener, the same way pyDocstring treats an unterminated
// docstring — better to contribute no doc than to attribute a paragraph of
// unrelated prose.
func hashRubyEmbed(lines []string, end int) string {
	if end < 0 || strings.TrimSpace(lines[end]) != "=end" {
		return ""
	}
	open := end
	for open >= 0 && strings.TrimSpace(lines[open]) != "=begin" {
		open--
	}
	if open < 0 {
		return ""
	}
	body := make([]string, 0, end-open-1)
	for _, l := range lines[open+1 : end] {
		body = append(body, strings.TrimSpace(l))
	}
	return joinDocLines(body)
}

// rubyDocComments maps each Ruby method/class/module name declared in a source
// file to its doc comment, if any — a run of "#" lines or a "=begin"/"=end"
// block written immediately above the declaration. This is the Ruby counterpart
// of goDocComments.
//
// The name regexes below are deliberately identical to the ones symbols.go
// feeds graph.Node.Symbols with (rubySymbols), and rubySymbolName reproduces
// rubySymbols' extraction logic — a doc comment keyed under a name that never
// reaches Node.Symbols would be silently useless, so the two must stay in
// step. symbols.go is the single source of truth; these are copies because
// they are unexported in another package.
//
// Returns nil when the file has no doc comments, matching goDocComments.
// Extraction cannot fail, so there is no error to return.
func rubyDocComments(content string) map[string]string {
	lines := strings.Split(content, "\n")

	var docs map[string]string
	for i, line := range lines {
		name, ok := rubySymbolName(line)
		if !ok {
			continue
		}
		// The "#" run is the common form and is tried first; an "=end" line
		// above the declaration is not comment content, so the run comes back
		// empty and the block form takes over. A "#" run above the block keeps
		// the run, which is the more specific attribution either way.
		text := hashDocText(hashCommentLinesAbove(lines, i-1, hashNoBlock), hashNoBlock)
		if text == "" {
			text = hashRubyEmbed(lines, i-1)
		}
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

// rubySymbolName returns the method/class/module name rubySymbols
// (internal/graph) would extract from line, so the returned key matches
// Node.Symbols exactly. Method definitions are checked first and win, exactly as
// rubySymbols does.
func rubySymbolName(line string) (string, bool) {
	if m := symRubyDefRE.FindString(line); m != "" {
		return strings.TrimSpace(symRubyDefStp.ReplaceAllString(m, "")), true
	}
	if m := symRubyTypeRE.FindString(line); m != "" {
		return strings.TrimSpace(symRubyTypeStp.ReplaceAllString(m, "")), true
	}
	return "", false
}

// shDocComments maps each Shell function name declared in a source file to its
// doc comment, if any — a run of "#" lines written immediately above the
// "function name" or "name() {" declaration. This is the Shell counterpart of
// goDocComments.
//
// The name regexes below are deliberately identical to the ones symbols.go
// feeds graph.Node.Symbols with (shSymbols), and shSymbolName reproduces
// shSymbols' extraction logic — a doc comment keyed under a name that never
// reaches Node.Symbols would be silently useless, so the two must stay in
// step. symbols.go is the single source of truth; these are copies because
// they are unexported in another package.
//
// Returns nil when the file has no doc comments, matching goDocComments.
// Extraction cannot fail, so there is no error to return.
func shDocComments(content string) map[string]string {
	lines := strings.Split(content, "\n")

	var docs map[string]string
	for i, line := range lines {
		name, ok := shSymbolName(line)
		if !ok {
			continue
		}
		text := hashDocText(hashCommentLinesAbove(lines, i-1, hashNoBlock), hashNoBlock)
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

// shSymbolName returns the function name shSymbols (internal/graph) would
// extract from line, so the returned key matches Node.Symbols exactly. The
// "function name" form is checked first and wins on a line that matches both,
// exactly as shSymbols does.
func shSymbolName(line string) (string, bool) {
	if m := symShFuncRE.FindString(line); m != "" {
		return strings.TrimSpace(symShFuncStrip.ReplaceAllString(m, "")), true
	}
	if m := symShBraceRE.FindString(line); m != "" {
		s := strings.TrimLeft(m, " \t")
		return symShBraceStrip.ReplaceAllString(s, ""), true
	}
	return "", false
}

// psDocComments maps each PowerShell function name declared in a source file to
// its doc comment, if any — a run of "#" lines or a "<# ... #>" block written
// immediately above the "function Name" declaration. This is the PowerShell
// counterpart of goDocComments.
//
// The name regexes below are deliberately identical to the ones symbols.go
// feeds graph.Node.Symbols with (the ".ps1" case of symbolsOf), and
// psSymbolName reproduces that extraction logic — a doc comment keyed under a
// name that never reaches Node.Symbols would be silently useless, so the two
// must stay in step. symbols.go is the single source of truth; these are copies
// because they are unexported in another package.
//
// Returns nil when the file has no doc comments, matching goDocComments.
// Extraction cannot fail, so there is no error to return.
func psDocComments(content string) map[string]string {
	lines := strings.Split(content, "\n")

	var docs map[string]string
	for i, line := range lines {
		name, ok := psSymbolName(line)
		if !ok {
			continue
		}
		text := hashDocText(hashCommentLinesAbove(lines, i-1, hashPSBlock), hashPSBlock)
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

// psSymbolName returns the function name symbolsOf (internal/graph) would
// extract from line, so the returned key matches Node.Symbols exactly.
func psSymbolName(line string) (string, bool) {
	if m := symPS1FuncRE.FindString(line); m != "" {
		return strings.TrimSpace(symPS1FuncStripRE.ReplaceAllString(m, "")), true
	}
	return "", false
}

// Copies of internal/graph/symbols.go's Ruby regexes (rubySymbols), Shell
// regexes (shSymbols) and PowerShell regexes (the ".ps1" case of symbolsOf).
// Kept byte-for-byte identical so the names they match are the names that end
// up in graph.Node.Symbols; see rubySymbolName, shSymbolName, psSymbolName.
var (
	symRubyDefRE   = regexp.MustCompile(`^[ \t]*def[ \t]+(self\.)?[A-Za-z0-9_?!=]+`)
	symRubyDefStp  = regexp.MustCompile(`.*def[ \t]+`)
	symRubyTypeRE  = regexp.MustCompile(`^[ \t]*(class|module)[ \t]+[A-Za-z0-9_:]+`)
	symRubyTypeStp = regexp.MustCompile(`^.*(class|module)[ \t]+`)

	symShFuncRE     = regexp.MustCompile(`^[ \t]*function[ \t]+[A-Za-z0-9_]+`)
	symShFuncStrip  = regexp.MustCompile(`^[ \t]*function[ \t]+`)
	symShBraceRE    = regexp.MustCompile(`^[ \t]*[A-Za-z0-9_]+[ \t]*\(\)[ \t]*\{`)
	symShBraceStrip = regexp.MustCompile(`[ \t]*\(\).*`)

	symPS1FuncRE      = regexp.MustCompile(`^[ \t]*function[ \t]+[A-Za-z0-9_-]+`)
	symPS1FuncStripRE = regexp.MustCompile(`^[ \t]*function[ \t]+`)
)
