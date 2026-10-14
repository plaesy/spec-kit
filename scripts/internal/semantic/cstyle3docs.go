package semantic

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/plaesy/spec-kit/internal/graph"
)

// This file holds the doc-comment extractors for the three languages whose
// doc comments share the JSDoc "comment block above the declaration" shape:
// C-family (C/C++/Obj-C), Swift, and PHP. They are grouped in one file
// because the shared part of their logic is already in jsdocs.go —
// jsCommentBlock finds the contiguous comment block ending on the line above a
// declaration and jsDocText strips the markers — so no comment scanner is
// duplicated here. What is left per language is only:
//
//  1. which symbol name internal/graph's extractor would produce for a line,
//     copied from symbols.go so the key matches graph.Node.Symbols exactly,
//  2. the "///" / "//!" doc-marker fix-up, because jsDocText strips "//" and
//     would otherwise leave the third slash in the text, and
//  3. stepping over attribute lines that sit between the comment and the
//     declaration (Swift "@available(...)", PHP 8 "#[Route(...)]"). Those are
//     the lines that would otherwise end the comment block and silently drop
//     the documentation of an annotated declaration.
//
// C-family needs no equivalent of (3): the only C/C++ attributes there are,
// "[[nodiscard]]" and "__attribute__((...))", always begin the line they are on
// — and a line beginning with "[" is not a declaration line for graph either,
// so there is nothing to step over before one.
//
// Permissiveness is deliberate, most of all for C/C++. Unlike PHPDoc, which is
// literally JSDoc's shape, C and C++ have no doc-comment standard at all: a
// "/* */" or "//" block above a declaration is as close to a doc comment as
// that language offers. So any contiguous comment block directly above a
// declaration counts here, which is the same rule go/parser applies to Go's own
// "//" blocks and the rule jsDocComments already applies to JS/TS. Requiring a
// recognised doc marker ("/**", "///", "//!", "@") would silently drop docs
// written in the plain "// ..." style that plenty of C code actually uses, and
// a false negative inside an anti-duplication check is the worst failure mode
// for this feature (see the reasoning in drift_test.go).

// cFamilyDocComments maps each C/C++/Obj-C function, method, struct, or class
// name declared in a source file to its doc comment, if any — the comment
// block sitting immediately above the declaration. This is the C-family
// counterpart of goDocComments/jsDocComments.
//
// Line-based rather than parsed, for the same reason as jsDocComments: no C or
// C++ parser in the Go standard library, and the value here is doc text, not a
// syntax tree.
//
// The declaration shape that makes this language different from the others is
// that the return type shares the line with the name and the "{" may open on
// the *next* line:
//
//	/** Sums two integers. */
//	int add(int a, int b)
//	{
//	    return a + b;
//	}
//
// The comment therefore sits above the declaration line, not above the brace,
// and the name is the identifier immediately before "(" — which is what
// cfamilyFuncName reproduces from internal/graph.
//
// Returns nil when the file has no doc comments, matching goDocComments.
// Extraction cannot fail, so there is no error to return.
func cFamilyDocComments(content string) map[string]string {
	lines := strings.Split(content, "\n")

	var docs map[string]string
	for i := range lines {
		names := cfamilySymbolNames(lines, i)
		if len(names) == 0 {
			continue
		}
		text := cfamilyDocText(jsCommentBlock(lines, i-1))
		if text == "" {
			continue
		}
		if docs == nil {
			docs = make(map[string]string)
		}
		for _, name := range names {
			docs[name] = text
		}
	}
	return docs
}

// cFamilyDocsFor returns n's C/C++ doc comments if n is a C-family file under
// repoRoot, nil otherwise (see docsFor).
//
// The extension list mirrors symbolsOf's C-family case exactly — .c, .h, .cpp,
// .hpp, .cc, .hh, .cxx — so a doc comment is never computed for a file whose
// symbols graph.Build did not extract. Note that graph's own file scan
// (includeExtRE) does not list .hh or .cxx, so those two extensions cannot
// actually reach this function today; they are kept in the list because
// symbols.go already extracts symbols for them and the two lists must not
// drift apart silently.
func cFamilyDocsFor(repoRoot string, n graph.Node) map[string]string {
	switch strings.ToLower(filepath.Ext(n.ID)) {
	case ".c", ".h", ".cpp", ".hpp", ".cc", ".hh", ".cxx":
	default:
		return nil
	}
	content, ok := readSource(repoRoot, n)
	if !ok {
		return nil
	}
	return cFamilyDocComments(content)
}

// cfamilySymbolNames returns the symbol names internal/graph's cFamilySymbols
// would extract at line i, in the order cFamilySymbols would report them.
//
// It returns up to two names for one line, because cFamilySymbols itself runs
// two passes and merges them back into source order: a "class"/"struct" type
// scan, and the C-like function scan (cLikeFuncHits) that also has to look at
// the next line to decide whether a brace follows. A line such as
// "struct Point clone(void)" is a type hit ("Point") and a function hit
// ("clone") at once, and both are in Node.Symbols, so both get the comment.
//
// names is nil for a line that yields no symbol, which is also how the caller
// tells "no declaration here" from "a declaration with no doc comment".
func cfamilySymbolNames(lines []string, i int) []string {
	var out []string
	if m := cfamilyTypeRE.FindString(lines[i]); m != "" {
		out = append(out, strings.TrimSpace(cfamilyTypeStrip.ReplaceAllString(m, "")))
	}
	if name, ok := cfamilyFuncName(lines, i); ok {
		out = append(out, name)
	}
	return out
}

// cfamilyFuncName returns the function or method name cLikeFuncHits
// (internal/graph) would record at line i: the identifier immediately before
// "(" in a typed declaration, recognised only when a "{" follows on the same
// line or opens the next non-blank line.
//
// The brace requirement is load-bearing and is why this cannot be a plain
// single-line regex: it is what keeps a prototype ("int area();") and control
// flow ("if (ready) {", one identifier only) out of Node.Symbols, and this
// function must agree with it — keying a doc comment under a name graph never
// extracted would mean nodeText's lookup never fires.
func cfamilyFuncName(lines []string, i int) (string, bool) {
	if m := cfamilyFuncSameLineRE.FindStringSubmatch(lines[i]); m != nil {
		return m[1], true
	}
	m := cfamilyFuncAllmanRE.FindStringSubmatch(lines[i])
	if m == nil {
		return "", false
	}
	// Allman style: "{" alone on the next non-blank line. Blank lines between
	// the signature and the brace are skipped, as cLikeFuncHits does.
	j := i + 1
	for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
		j++
	}
	if j < len(lines) && strings.TrimSpace(lines[j]) == "{" {
		return m[1], true
	}
	return "", false
}

// cfamilyDocText turns a raw comment block above a C-family declaration into
// doc text: jsDocText's marker stripping plus the "///" / "//!" fix-up (see
// cfamilySlashDocLines).
func cfamilyDocText(block []string) string {
	return jsDocText(cfamilySlashDocLines(block))
}

// cfamilySlashDocLines normalises C-family triple-slash doc lines ("/// text"
// or "//! text") to the plain "// text" form jsDocText already knows how to
// strip, so the extracted text is "text" and not "/ text".
//
// Only the marker is rewritten; the comment body is left byte-for-byte alone.
// Rewriting line by line rather than trimming a "/" out of the joined result is
// what keeps a legitimate "// /etc/plaesy.conf is the config" line intact.
func cfamilySlashDocLines(block []string) []string {
	out := make([]string, len(block))
	for i, l := range block {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "///") || strings.HasPrefix(t, "//!") {
			out[i] = "//" + t[3:]
			continue
		}
		out[i] = l
	}
	return out
}

// swiftDocComments maps each Swift func/type name declared in a source file to
// its doc comment, if any — the comment block sitting immediately above the
// declaration. This is the Swift counterpart of goDocComments/jsDocComments.
//
// Both of Swift's doc-comment styles are handled by jsCommentBlock: the "///"
// line form (Swift DocC, and the dominant convention) and the "/** */" block
// form that Doxygen uses, plus a plain "//" run, which is accepted for the
// same permissiveness reason as C-family.
//
// The name regexes below are deliberately identical to the ones symbols.go
// feeds graph.Node.Symbols with (swiftSymbols), and swiftSymbolName reproduces
// swiftSymbols' extraction logic — see the note at the top of jsdocs.go on why
// the copies must stay in step.
//
// Returns nil when the file has no doc comments, matching goDocComments.
func swiftDocComments(content string) map[string]string {
	lines := strings.Split(content, "\n")

	var docs map[string]string
	for i, line := range lines {
		name, ok := swiftSymbolName(line)
		if !ok {
			continue
		}
		text := swiftDocText(jsCommentBlock(lines, swiftDeclStart(lines, i)-1))
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

// swiftDocsFor returns n's Swift doc comments if n is a .swift file under
// repoRoot, nil otherwise (see docsFor).
func swiftDocsFor(repoRoot string, n graph.Node) map[string]string {
	if strings.ToLower(filepath.Ext(n.ID)) != ".swift" {
		return nil
	}
	content, ok := readSource(repoRoot, n)
	if !ok {
		return nil
	}
	return swiftDocComments(content)
}

// swiftSymbolName returns the func/type name swiftSymbols (internal/graph)
// would extract from line, so the returned key matches Node.Symbols exactly.
// A func declaration is checked first and wins, exactly as swiftSymbols does.
func swiftSymbolName(line string) (string, bool) {
	if m := swiftFuncRE.FindString(line); m != "" {
		return strings.TrimSpace(swiftFuncStrip.ReplaceAllString(m, "")), true
	}
	if m := swiftTypeRE.FindString(line); m != "" {
		return strings.TrimSpace(swiftTypeStrip.ReplaceAllString(m, "")), true
	}
	return "", false
}

// swiftDeclStart returns the first line of the declaration swiftSymbols matched
// at declLine — that is, declLine itself unless attribute lines sit above it.
//
// Swift attributes ("@available(iOS 15, *)", "@objc", "@discardableResult",
// "@MainActor") are written between the "///" comment and the declaration, and
// above public API they are close to universal. jsCommentBlock treats a
// non-comment line as the end of the block, so without this the documentation
// of every annotated declaration would be silently dropped. Attributes are
// stepped over rather than absorbed: they are not in the block handed to
// jsDocText, so "@available(...)" can never reach the embedding as doc text.
//
// A blank line still ends the walk, so an attribute run can never bridge a
// paragraph break and attach an unrelated comment.
func swiftDeclStart(lines []string, declLine int) int {
	i, depth := declLine, 0
	for i > 0 {
		t := strings.TrimSpace(lines[i-1])
		if t == "" {
			break
		}
		if depth == 0 && !swiftAttributeLine(t) {
			// Not an attribute: stop, unless this is the tail of a multi-line
			// attribute ("@available(\n    iOS 15,\n    *\n)"), whose opening
			// line is further up and still an attribute.
			if !strings.HasSuffix(t, ")") {
				break
			}
			j := swiftAttributeTailStart(lines, i-1)
			if j < 0 {
				break
			}
			i = j
			continue
		}
		depth += strings.Count(t, "(") - strings.Count(t, ")")
		i--
	}
	return i
}

// swiftAttributeTailStart returns the line index of the attribute whose closing
// paren sits on line end — a multi-line "@available(\n ...\n)" — or -1 when the
// line is not part of one. Bounded to a short run so a stray ")" in code cannot
// walk the scan far up the file.
func swiftAttributeTailStart(lines []string, end int) int {
	for j := end; j >= 0 && j > end-20; j-- {
		t := strings.TrimSpace(lines[j])
		if t == "" {
			return -1
		}
		if strings.Count(t, "(") > strings.Count(t, ")") && swiftAttributeLine(t) {
			return j
		}
	}
	return -1
}

// swiftAttributeLine reports whether an already-trimmed line is a Swift
// attribute (or the opening line of a multi-line one) rather than code: an "@"
// followed by a possibly-dotted identifier and optional arguments, with nothing
// else on the line. Requiring that identifier shape is what keeps a "@" that is
// not an attribute from being stepped over.
func swiftAttributeLine(t string) bool {
	if !strings.HasPrefix(t, "@") {
		return false
	}
	name := t[1:]
	if i := strings.IndexAny(name, "( \t"); i >= 0 {
		name = name[:i]
	}
	if name == "" {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '.':
		default:
			return false
		}
	}
	return true
}

// swiftDocText turns a raw comment block above a Swift declaration into doc
// text: jsDocText's marker stripping plus the "///" / "//!" fix-up (see
// swiftSlashDocLines).
func swiftDocText(block []string) string {
	return jsDocText(swiftSlashDocLines(block))
}

// swiftSlashDocLines normalises Swift triple-slash doc lines ("/// text") to
// the plain "// text" form jsDocText already knows how to strip.
//
// "///" is Swift's primary doc-comment marker (DocC), so this fix-up matters
// more here than in any other language covered so far: without it every
// extracted Swift doc would start with a stray "/".
//
// This is a per-language copy of cfamilySlashDocLines rather than a shared
// helper: the three languages being added to this package in parallel each
// carry their own prefixed copy, so no unprefixed helper is introduced into a
// package other agents are writing into at the same time.
func swiftSlashDocLines(block []string) []string {
	out := make([]string, len(block))
	for i, l := range block {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "///") {
			out[i] = "//" + t[3:]
			continue
		}
		out[i] = l
	}
	return out
}

// phpDocComments maps each PHP function/method/class/interface/trait name
// declared in a source file to its doc comment, if any — the comment block
// sitting immediately above the declaration. This is the PHP counterpart of
// goDocComments/jsDocComments.
//
// PHPDoc is literally JSDoc's shape ("/** ... */" above the declaration), so
// there is nothing language-specific in the comment handling at all: the block
// scan and the marker stripping are jsCommentBlock's and jsDocText's, called
// directly. PHP's "///" style is not a thing and no fix-up is applied.
//
// The name regexes below are deliberately identical to the ones symbols.go
// feeds graph.Node.Symbols with (phpSymbols), and phpSymbolName reproduces
// phpSymbols' extraction logic — see the note at the top of jsdocs.go on why
// the copies must stay in step.
//
// Returns nil when the file has no doc comments, matching goDocComments.
func phpDocComments(content string) map[string]string {
	lines := strings.Split(content, "\n")

	var docs map[string]string
	for i, line := range lines {
		name, ok := phpSymbolName(line)
		if !ok {
			continue
		}
		text := jsDocText(jsCommentBlock(lines, phpDeclStart(lines, i)-1))
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

// phpDocsFor returns n's PHPDoc comments if n is a .php file under repoRoot,
// nil otherwise (see docsFor).
func phpDocsFor(repoRoot string, n graph.Node) map[string]string {
	if strings.ToLower(filepath.Ext(n.ID)) != ".php" {
		return nil
	}
	content, ok := readSource(repoRoot, n)
	if !ok {
		return nil
	}
	return phpDocComments(content)
}

// phpDeclStart returns the first line of the declaration phpSymbols matched at
// declLine — that is, declLine itself unless PHP 8 attribute lines sit above
// it.
//
// A controller method commonly looks like this, docblock first and attribute
// below it:
//
//	/** Removes an account permanently. */
//	#[Route('/accounts/{id}', methods: ['DELETE'])]
//	public function purge(string $id) { ... }
//
// jsCommentBlock treats the "#[...]" line as the end of the block, so without
// this step the docblock of every attributed method would be dropped. The
// attribute is stepped over, not absorbed: it never reaches the embedded text.
//
// A blank line still ends the walk, so an attribute run cannot bridge a
// paragraph break.
func phpDeclStart(lines []string, declLine int) int {
	i, depth := declLine, 0
	for i > 0 {
		t := strings.TrimSpace(lines[i-1])
		if t == "" {
			break
		}
		if depth == 0 && !phpAttributeLine(t) {
			if !strings.HasSuffix(t, "]") {
				break
			}
			j := phpAttributeTailStart(lines, i-1)
			if j < 0 {
				break
			}
			i = j
			continue
		}
		depth += strings.Count(t, "[") - strings.Count(t, "]")
		i--
	}
	return i
}

// phpAttributeTailStart returns the line index of the "#[...]" whose closing
// bracket sits on line end — a multi-line attribute — or -1 when the line is
// not part of one. Bounded to a short run so a stray "]" in code cannot walk
// the scan far up the file.
func phpAttributeTailStart(lines []string, end int) int {
	for j := end; j >= 0 && j > end-20; j-- {
		t := strings.TrimSpace(lines[j])
		if t == "" {
			return -1
		}
		if strings.Count(t, "[") > strings.Count(t, "]") && phpAttributeLine(t) {
			return j
		}
	}
	return -1
}

// phpAttributeLine reports whether an already-trimmed line is a PHP 8 attribute
// or the opening line of a multi-line one. PHPDoc has no "@annotation" form, so
// "#[Attr]" is the whole shape to recognise.
func phpAttributeLine(t string) bool {
	return strings.HasPrefix(t, "#[")
}

// phpSymbolName returns the function/type name phpSymbols (internal/graph)
// would extract from line, so the returned key matches Node.Symbols exactly.
// A function declaration is checked first and wins, exactly as phpSymbols does.
func phpSymbolName(line string) (string, bool) {
	if m := phpFuncRE.FindString(line); m != "" {
		return strings.TrimSpace(phpFuncStrip.ReplaceAllString(m, "")), true
	}
	if m := phpTypeRE.FindString(line); m != "" {
		return strings.TrimSpace(phpTypeStrip.ReplaceAllString(m, "")), true
	}
	return "", false
}

// Copies of internal/graph/symbols.go's C-family regexes (cFamilySymbols plus
// the shared cLikeFuncHits pair) and Swift/PHP regexes (swiftSymbols,
// phpSymbols). Kept byte-for-byte identical so the names they match are the
// names that end up in graph.Node.Symbols; see cfamilySymbolNames,
// swiftSymbolName, phpSymbolName.
var (
	cfamilyTypeRE    = regexp.MustCompile(`^[ \t]*(class|struct)[ \t]+[A-Za-z0-9_]+`)
	cfamilyTypeStrip = regexp.MustCompile(`^.*(class|struct)[ \t]+`)

	// cfamilyFuncSameLineRE / cfamilyFuncAllmanRE are copies of symbols.go's
	// symCLikeFuncRE / symCLikeFuncNoBraceRE, restricted to C-family by where
	// they are used. Both are best-effort "typed-return declaration" detectors
	// for languages with no function keyword: two or more identifier-like
	// tokens followed by "(...)". Requiring two tokens (return type + name) is
	// what keeps "if (x) {" (one token) out; requiring a following brace is
	// what keeps prototypes ("int area();") out. See cfamilyFuncName.
	cfamilyFuncSameLineRE = regexp.MustCompile(`^[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*(?:<[^()]*>)?(?:\[\])?[ \t]+)+([A-Za-z_][A-Za-z0-9_]*)[ \t]*\([^;{}]*\)[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*[ \t]*)*\{[ \t]*$`)
	cfamilyFuncAllmanRE   = regexp.MustCompile(`^[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*(?:<[^()]*>)?(?:\[\])?[ \t]+)+([A-Za-z_][A-Za-z0-9_]*)[ \t]*\([^;{}]*\)[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*[ \t]*)*$`)

	swiftFuncRE    = regexp.MustCompile(`^[ \t]*(?:(?:public|private|internal|fileprivate|open|static|override|mutating)[ \t]+)*func[ \t]+[A-Za-z0-9_]+`)
	swiftFuncStrip = regexp.MustCompile(`.*func[ \t]+`)
	swiftTypeRE    = regexp.MustCompile(`^[ \t]*(?:(?:public|private|internal|fileprivate|open|final)[ \t]+)*(class|struct|enum|protocol)[ \t]+[A-Za-z0-9_]+`)
	swiftTypeStrip = regexp.MustCompile(`^.*(class|struct|enum|protocol)[ \t]+`)

	phpFuncRE    = regexp.MustCompile(`^[ \t]*(?:(?:public|private|protected|static)[ \t]+)*function[ \t]+[A-Za-z0-9_]+`)
	phpFuncStrip = regexp.MustCompile(`.*function[ \t]+`)
	phpTypeRE    = regexp.MustCompile(`^[ \t]*(?:(?:abstract|final)[ \t]+)?(class|interface|trait)[ \t]+[A-Za-z0-9_]+`)
	phpTypeStrip = regexp.MustCompile(`^.*(class|interface|trait)[ \t]+`)
)
