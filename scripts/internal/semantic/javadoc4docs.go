package semantic

import (
	"regexp"
	"strings"
)

// Doc-comment extraction for the four Javadoc/KDoc-shaped languages
// internal/graph/symbols.go supports but that had no extractor: Dart
// (.dart), Java (.java), Kotlin (.kt, .kts) and C# (.cs). All four document
// declarations with the same "/** ... */" block — Javadoc for Java, KDoc for
// Kotlin, dartdoc for Dart, XML doc for C# — or with a run of "///" lines
// immediately above the declaration. That is exactly the shape jsDocComments
// already handles, so every extractor here is a name matcher plus a call into
// jsCommentBlock/jsDocText rather than a second comment scanner; the language
// differences are all in how a declaration's name is recognised.
//
// Regex/line-based rather than parsed, for the same reason as jsDocComments:
// no Dart/Java/Kotlin/C# parser in the Go standard library, and the value here
// is the doc text, not a syntax tree.
//
// Name matching is deliberately per-language and mirrors symbols.go's
// dartSymbols, javaSymbols, kotlinSymbols and csharpSymbols line for line —
// same regexes, same order, same trimming. nodeText looks a doc comment up by
// the name in graph.Node.Symbols, so a doc keyed under a name graph never
// extracts would be silently useless, and a name graph extracts but this
// package fails to recognise would silently lose enrichment. symbols.go is the
// single source of truth; the regexes at the bottom of this file are copies
// because they are unexported in another package.
//
// Annotations sit between the doc comment and the declaration in all four
// languages (@Override / @JvmStatic / @visibleForTesting, or C#'s
// [Serializable] / [HttpGet]), and for Java and C# that is the common case
// rather than an edge case. jsCommentBlock treats a non-comment line as the end
// of the block — correct for the code it was written for — so handing it the
// line directly above an annotated declaration would drop the doc entirely.
// Each language therefore passes it xxxDocEndAbove, which walks up over the
// annotation lines first. The annotations are excluded from the block that way,
// so they never reach the embedding either.
//
// Known limitation, consistent with the rest of this package's heuristic
// approach: only single-line annotations are stepped over. A multi-line one
// (@SuppressWarnings({ ... }) in Java, an attribute list broken across lines in
// C#) is not recognised, and the member below it loses its doc comment exactly
// as it would without the fix.
//
// Known cosmetic caveat, shared with the JS/TS extractor: jsDocText strips
// exactly two leading slashes, so a "///" line keeps a stray "/" in front of its
// text ("/ Loads the user."). For Dart, Kotlin and C# that is the primary doc
// form rather than an edge case, so each of them removes the third slash on its
// own lines first (dartLineDocBlock and friends) and leaves every other line to
// jsDocText untouched. jsdocs.go is shared with the JS/TS extractor and is not
// this file's to change; if it ever grows the same fix-up, the three helpers
// here become no-ops rather than needing removal.
//
// Returns nil when the file has no doc comments, matching goDocComments and
// jsDocComments. Extraction cannot fail, so there is no error to return.

// dartDocComments maps each Dart class/method name declared in a source file to
// its dartdoc, if any — the "/** ... */" block or a run of "///" lines written
// immediately above the declaration.
func dartDocComments(content string) map[string]string {
	lines := strings.Split(content, "\n")

	var docs map[string]string
	for i := range lines {
		names := dartSymbolNames(lines, i)
		if len(names) == 0 {
			continue
		}
		text := jsDocText(dartLineDocBlock(jsCommentBlock(lines, dartDocEndAbove(lines, i))))
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

// dartSymbolNames returns the symbol names dartSymbols (internal/graph) would
// extract from lines[i], in mergeHits' order: the class name from the type pass
// first, then the method name from the C-like pass. A slice rather than
// jsSymbolName's single value because those two passes are independent, so one
// declaration line can yield both. Every name it returns is spelled exactly as
// graph.Node.Symbols spells it.
func dartSymbolNames(lines []string, i int) []string {
	var names []string
	if m := symDartClassRE.FindString(lines[i]); m != "" {
		names = append(names, strings.TrimSpace(symDartClsStrp.ReplaceAllString(m, "")))
	}
	if name, ok := dartCLikeFuncName(lines, i); ok {
		names = append(names, name)
	}
	return names
}

// dartCLikeFuncName reproduces cLikeFuncHits' per-line logic for .dart: the
// same-line-brace form wins outright, otherwise the Allman form (signature on
// one line, "{" alone on the next non-blank line) is accepted. The captured
// name is returned verbatim, exactly as cLikeFuncHits does — it does not
// TrimSpace it either, and trimming here would key a doc under a name
// Node.Symbols does not contain.
func dartCLikeFuncName(lines []string, i int) (string, bool) {
	if m := dartCLikeFuncRE.FindStringSubmatch(lines[i]); m != nil {
		return m[1], true
	}
	m := dartCLikeFuncNoBraceRE.FindStringSubmatch(lines[i])
	if m == nil {
		return "", false
	}
	j := i + 1
	for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
		j++
	}
	if j < len(lines) && strings.TrimSpace(lines[j]) == "{" {
		return m[1], true
	}
	return "", false
}

// dartDocEndAbove returns the line index jsCommentBlock should treat as the end
// of the doc comment for the declaration on line i: normally i-1, but stepped
// up over any dartdoc annotations (@override, @visibleForTesting, ...) sitting
// between the comment and the declaration. Stepping over them rather than
// letting jsCommentBlock see them is what keeps an annotated member's doc
// comment attached to it — the annotation is excluded from the returned block,
// so it never becomes doc text.
func dartDocEndAbove(lines []string, i int) int {
	j := i - 1
	for j >= 0 && dartAnnotationLine(strings.TrimSpace(lines[j])) {
		j--
	}
	return j
}

// dartAnnotationLine reports whether a trimmed line is a Dart annotation. Only
// lines starting with "@" qualify; in the position this is called from (directly
// above a declaration) nothing else can start with one.
func dartAnnotationLine(t string) bool {
	return strings.HasPrefix(t, "@")
}

// dartLineDocBlock returns block with one leading slash removed from every "///"
// line, so that jsDocText's exactly-two-slash strip lands on the marker rather
// than leaving a stray "/" in front of the words. "///" is Dart's primary doc
// form, so without this nearly every Dart doc would be embedded as "/ Loads the
// user." Only "///" lines are touched; every other line is passed through
// unchanged, so a "/** ... */" block comes out exactly as jsDocText produces
// it. This is a one-shape fix-up in front of the shared cleaner, not a second
// comment scanner: jsCommentBlock still decides where the block starts and ends.
func dartLineDocBlock(block []string) []string {
	has := false
	for _, l := range block {
		if strings.HasPrefix(strings.TrimSpace(l), "///") {
			has = true
			break
		}
	}
	if !has {
		return block
	}
	out := make([]string, len(block))
	for i, l := range block {
		out[i] = l
		// "///x" -> "//x": dropping the *third* slash (index 2) leaves the two
		// jsDocText knows how to strip, so the text comes out clean.
		if t := strings.TrimSpace(l); strings.HasPrefix(t, "///") {
			out[i] = "//" + t[3:]
		}
	}
	return out
}

// javaDocComments maps each Java class/interface/enum name and method name
// declared in a source file to its Javadoc, if any.
func javaDocComments(content string) map[string]string {
	lines := strings.Split(content, "\n")

	var docs map[string]string
	for i := range lines {
		names := javaSymbolNames(lines, i)
		if len(names) == 0 {
			continue
		}
		text := jsDocText(jsCommentBlock(lines, javaDocEndAbove(lines, i)))
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

// javaSymbolNames returns the symbol names javaSymbols (internal/graph) would
// extract from lines[i] — type name first, then method name, matching
// mergeHits' order.
func javaSymbolNames(lines []string, i int) []string {
	var names []string
	if m := symJavaClassRE.FindString(lines[i]); m != "" {
		names = append(names, strings.TrimSpace(symJavaClassStp.ReplaceAllString(m, "")))
	}
	if name, ok := javaCLikeFuncName(lines, i); ok {
		names = append(names, name)
	}
	return names
}

// javaCLikeFuncName is dartCLikeFuncName's Java counterpart, mirroring
// cLikeFuncHits per line for .java.
func javaCLikeFuncName(lines []string, i int) (string, bool) {
	if m := javaCLikeFuncRE.FindStringSubmatch(lines[i]); m != nil {
		return m[1], true
	}
	m := javaCLikeFuncNoBraceRE.FindStringSubmatch(lines[i])
	if m == nil {
		return "", false
	}
	j := i + 1
	for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
		j++
	}
	if j < len(lines) && strings.TrimSpace(lines[j]) == "{" {
		return m[1], true
	}
	return "", false
}

// javaDocEndAbove returns the line index jsCommentBlock should treat as the end
// of the Javadoc for the declaration on line i, stepped up over any annotations
// (@Override, @SuppressWarnings, @Deprecated, ...) between the comment and the
// declaration. This matters far more for Java than for any other language here:
// an @Override above an overriding method is close to universal, so refusing to
// step over it would drop most methods' documentation.
func javaDocEndAbove(lines []string, i int) int {
	j := i - 1
	for j >= 0 && javaAnnotationLine(strings.TrimSpace(lines[j])) {
		j--
	}
	return j
}

// javaAnnotationLine reports whether a trimmed line is a Java annotation
// (@Override, @SuppressWarnings("unchecked"), ...).
func javaAnnotationLine(t string) bool {
	return strings.HasPrefix(t, "@")
}

// kotlinDocComments maps each Kotlin function and type name declared in a source
// file to its KDoc, if any — "/** ... */", identical in shape to Javadoc, or a
// run of "///" lines.
func kotlinDocComments(content string) map[string]string {
	lines := strings.Split(content, "\n")

	var docs map[string]string
	for i := range lines {
		names := kotlinSymbolNames(lines, i)
		if len(names) == 0 {
			continue
		}
		text := jsDocText(kotlinLineDocBlock(jsCommentBlock(lines, kotlinDocEndAbove(lines, i))))
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

// kotlinSymbolNames returns the symbol names kotlinSymbols (internal/graph)
// would extract from lines[i]. Unlike the three mergeHits languages, Kotlin
// keys off an explicit `fun` keyword rather than the C-like "return type plus
// name plus brace" shape, so kotlinSymbols checks the function regex first and
// `continue`s — a line that matched is never also tested as a type. That is
// reproduced here by returning early, which is also why this one returns a
// slice of at most one name.
func kotlinSymbolNames(lines []string, i int) []string {
	if m := symKotlinFunRE.FindString(lines[i]); m != "" {
		return []string{strings.TrimSpace(symKotlinFunStp.ReplaceAllString(m, ""))}
	}
	if m := symKotlinTypeRE.FindString(lines[i]); m != "" {
		return []string{strings.TrimSpace(symKotlinTypeStp.ReplaceAllString(m, ""))}
	}
	return nil
}

// kotlinDocEndAbove returns the line index jsCommentBlock should treat as the
// end of the KDoc for the declaration on line i, stepped up over any
// annotations (@JvmStatic, @Throws, @Deprecated, ...) between the comment and
// the declaration.
func kotlinDocEndAbove(lines []string, i int) int {
	j := i - 1
	for j >= 0 && kotlinAnnotationLine(strings.TrimSpace(lines[j])) {
		j--
	}
	return j
}

// kotlinAnnotationLine reports whether a trimmed line is a Kotlin annotation
// (@JvmStatic, @Throws(Exception::class), ...).
func kotlinAnnotationLine(t string) bool {
	return strings.HasPrefix(t, "@")
}

// kotlinLineDocBlock is dartLineDocBlock's Kotlin counterpart: KDoc is
// normally "/** ... */", but "///" line docs are also idiomatic Kotlin, and
// jsDocText's two-slash strip would leave the third one in front of the words.
func kotlinLineDocBlock(block []string) []string {
	has := false
	for _, l := range block {
		if strings.HasPrefix(strings.TrimSpace(l), "///") {
			has = true
			break
		}
	}
	if !has {
		return block
	}
	out := make([]string, len(block))
	for i, l := range block {
		out[i] = l
		if t := strings.TrimSpace(l); strings.HasPrefix(t, "///") {
			out[i] = "//" + t[3:]
		}
	}
	return out
}

// csharpDocComments maps each C# class/interface/struct/enum name and method
// name declared in a source file to its XML doc comment, if any.
func csharpDocComments(content string) map[string]string {
	lines := strings.Split(content, "\n")

	var docs map[string]string
	for i := range lines {
		names := csharpSymbolNames(lines, i)
		if len(names) == 0 {
			continue
		}
		text := csharpDocText(jsCommentBlock(lines, csharpDocEndAbove(lines, i)))
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

// csharpSymbolNames returns the symbol names csharpSymbols (internal/graph)
// would extract from lines[i] — type name first, then method name, matching
// mergeHits' order.
func csharpSymbolNames(lines []string, i int) []string {
	var names []string
	if m := symCSharpTypeRE.FindString(lines[i]); m != "" {
		names = append(names, strings.TrimSpace(symCSharpTypeStp.ReplaceAllString(m, "")))
	}
	if name, ok := csharpCLikeFuncName(lines, i); ok {
		names = append(names, name)
	}
	return names
}

// csharpCLikeFuncName is dartCLikeFuncName's C# counterpart, mirroring
// cLikeFuncHits per line for .cs.
func csharpCLikeFuncName(lines []string, i int) (string, bool) {
	if m := csharpCLikeFuncRE.FindStringSubmatch(lines[i]); m != nil {
		return m[1], true
	}
	m := csharpCLikeFuncNoBraceRE.FindStringSubmatch(lines[i])
	if m == nil {
		return "", false
	}
	j := i + 1
	for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
		j++
	}
	if j < len(lines) && strings.TrimSpace(lines[j]) == "{" {
		return m[1], true
	}
	return "", false
}

// csharpDocEndAbove returns the line index jsCommentBlock should treat as the
// end of the XML doc for the declaration on line i, stepped up over any
// attribute lines ([Serializable], [HttpGet("api/x")], ...) between the doc
// comment and the declaration. For C# this is the norm rather than the
// exception: every MVC controller action and most DTO members carry one, and
// since C#'s documentation convention is "///" lines, the attribute would
// otherwise separate almost every member from its own documentation.
func csharpDocEndAbove(lines []string, i int) int {
	j := i - 1
	for j >= 0 && csharpAttributeLine(strings.TrimSpace(lines[j])) {
		j--
	}
	return j
}

// csharpAttributeLine reports whether a trimmed line is a C# attribute
// ([Serializable], [HttpGet("api/accounts")], ...). Only "[" starts count; in
// the position this is called from nothing else can. Multi-line attribute lists
// (an argument broken onto the next line) are not handled — see the
// single-line-annotation limitation noted at the top of this file.
func csharpAttributeLine(t string) bool {
	return strings.HasPrefix(t, "[")
}

// csharpLineDocBlock is dartLineDocBlock's C# counterpart, and the one that
// matters most here: "///" is C#'s documentation convention, so jsDocText's
// two-slash strip would leave the third slash in front of every C# doc line.
func csharpLineDocBlock(block []string) []string {
	has := false
	for _, l := range block {
		if strings.HasPrefix(strings.TrimSpace(l), "///") {
			has = true
			break
		}
	}
	if !has {
		return block
	}
	out := make([]string, len(block))
	for i, l := range block {
		out[i] = l
		if t := strings.TrimSpace(l); strings.HasPrefix(t, "///") {
			out[i] = "//" + t[3:]
		}
	}
	return out
}

// csharpDocText is jsDocText over a raw C# comment block, with two C# specifics
// applied first: the third slash of a "///" line doc is removed (see
// csharpLineDocBlock), then the XML doc tags are stripped from the text.
//
// The tags matter more than they look. The dominant C# documentation form is
//
//	/// <summary>
//	/// Reloads accounts from the remote store.
//	/// </summary>
//
// whose first line is the bare tag "<summary>", and nodeText embeds only
// firstLine — so leaving tags in would embed the tag name instead of the
// documentation for every member documented that way. Only the tags are
// removed; whatever they wrap is kept, so a single-line
// "/// <summary>Finds every account.</summary>" becomes "Finds every account."
func csharpDocText(block []string) string {
	return csharpStripXMLTags(jsDocText(csharpLineDocBlock(block)))
}

// csharpStripXMLTags removes "<...>" doc tags from already-cleaned doc text and
// drops any line that held nothing else. Splitting on "\n" and rejoining keeps
// the multi-line shape jsDocText produces, so firstLine still sees the first
// line that carries words.
func csharpStripXMLTags(text string) string {
	lines := strings.Split(text, "\n")
	kept := make([]string, 0, len(lines))
	for _, l := range lines {
		if t := strings.TrimSpace(csharpXMLTagRE.ReplaceAllString(l, "")); t != "" {
			kept = append(kept, t)
		}
	}
	return strings.Join(kept, "\n")
}

// Copies of internal/graph/symbols.go's Dart/Java/Kotlin/C# regexes, kept
// byte-for-byte identical so the names they match are the names that end up in
// graph.Node.Symbols; see dartSymbolNames and friends.
//
// The C-like method regexes are per-language copies of symCLikeFuncRE /
// symCLikeFuncNoBraceRE for the same reason. They are duplicated rather than
// shared on purpose: this package cannot reach symbols.go's unexported helpers,
// and each language keeps its own matcher so a change to one language's
// declarations cannot be silently absorbed by another.
var (
	symDartClassRE = regexp.MustCompile(`^[ \t]*(abstract[ \t]+)?class[ \t]+[A-Za-z0-9_]+`)
	symDartClsStrp = regexp.MustCompile(`.*class[ \t]+`)

	dartCLikeFuncRE = regexp.MustCompile(`^[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*(?:<[^()]*>)?(?:\[\])?[ \t]+)+([A-Za-z_][A-Za-z0-9_]*)[ \t]*\([^;{}]*\)[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*[ \t]*)*\{[ \t]*$`)
	// dartCLikeFuncNoBraceRE is dartCLikeFuncRE's Allman-style counterpart:
	// the same declaration shape, but ending the line right after the ")"
	// (plus optional trailing modifiers) instead of requiring "{" on the same
	// line. Only ever used with the same-line-content next-line-"{" lookahead
	// in dartCLikeFuncName, never on its own — on its own it would also match
	// a prototype/interface declaration that ends in ";" on the next line.
	dartCLikeFuncNoBraceRE = regexp.MustCompile(`^[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*(?:<[^()]*>)?(?:\[\])?[ \t]+)+([A-Za-z_][A-Za-z0-9_]*)[ \t]*\([^;{}]*\)[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*[ \t]*)*$`)

	symJavaClassRE  = regexp.MustCompile(`^[ \t]*(?:(?:public|private|protected|static|final|abstract)[ \t]+)*(class|interface|enum)[ \t]+[A-Za-z0-9_]+`)
	symJavaClassStp = regexp.MustCompile(`^.*(class|interface|enum)[ \t]+`)

	javaCLikeFuncRE        = regexp.MustCompile(`^[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*(?:<[^()]*>)?(?:\[\])?[ \t]+)+([A-Za-z_][A-Za-z0-9_]*)[ \t]*\([^;{}]*\)[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*[ \t]*)*\{[ \t]*$`)
	javaCLikeFuncNoBraceRE = regexp.MustCompile(`^[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*(?:<[^()]*>)?(?:\[\])?[ \t]+)+([A-Za-z_][A-Za-z0-9_]*)[ \t]*\([^;{}]*\)[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*[ \t]*)*$`)

	symKotlinFunRE   = regexp.MustCompile(`^[ \t]*(?:(?:public|private|internal|protected|suspend|inline|override)[ \t]+)*fun[ \t]+[A-Za-z0-9_]+`)
	symKotlinFunStp  = regexp.MustCompile(`.*fun[ \t]+`)
	symKotlinTypeRE  = regexp.MustCompile(`^[ \t]*(?:(?:public|private|internal|protected|abstract|open|sealed|data)[ \t]+)*(class|interface|object)[ \t]+[A-Za-z0-9_]+`)
	symKotlinTypeStp = regexp.MustCompile(`^.*(class|interface|object)[ \t]+`)

	symCSharpTypeRE  = regexp.MustCompile(`^[ \t]*(?:(?:public|private|protected|internal|static|abstract|sealed|partial)[ \t]+)*(class|interface|struct|enum)[ \t]+[A-Za-z0-9_]+`)
	symCSharpTypeStp = regexp.MustCompile(`^.*(class|interface|struct|enum)[ \t]+`)

	csharpCLikeFuncRE        = regexp.MustCompile(`^[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*(?:<[^()]*>)?(?:\[\])?[ \t]+)+([A-Za-z_][A-Za-z0-9_]*)[ \t]*\([^;{}]*\)[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*[ \t]*)*\{[ \t]*$`)
	csharpCLikeFuncNoBraceRE = regexp.MustCompile(`^[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*(?:<[^()]*>)?(?:\[\])?[ \t]+)+([A-Za-z_][A-Za-z0-9_]*)[ \t]*\([^;{}]*\)[ \t]*(?:[A-Za-z_][A-Za-z0-9_]*[ \t]*)*$`)

	// csharpXMLTagRE matches the XML doc tags of a C# "///" comment — <summary>,
	// </summary>, <param name="id">, <returns>, ... — so csharpStripXMLTags can
	// remove the markup and keep the words it wraps.
	csharpXMLTagRE = regexp.MustCompile(`</?[A-Za-z][^>]*>`)
)
