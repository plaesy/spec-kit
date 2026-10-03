package mdlint

import (
	"fmt"
	"strings"
	"unicode"
)

// This file holds the line-level rules: everything decidable from the raw lines
// plus the front-matter and code-fence masks, with no AST needed.

// ---- MD009 no-trailing-spaces -------------------------------------------------

type MD009Options struct {
	// BrSpaces is how many trailing spaces are tolerated (markdownlint's
	// br_spaces: two trailing spaces are a hard line break).
	BrSpaces *int `json:"br_spaces"`
	// ListItemEmptyLines checks for trailing spaces on list item lines too.
	ListItemEmptyLines *bool `json:"list_item_empty_lines"`
}

func checkMD009(d *doc, rs *ruleSet) []violation {
	o := opt[MD009Options](rs, "MD009")
	allowed := 0
	if o.BrSpaces != nil {
		allowed = *o.BrSpaces
	}
	var out []violation
	for i := 0; i < d.n(); i++ {
		if d.inFrontMatter(i) {
			continue
		}
		line := d.lines[i]
		trimmedRight := strings.TrimRight(line, " \t")
		if trimmedRight == line {
			continue
		}
		trailing := len(line) - len(trimmedRight)
		if trailing <= allowed {
			continue
		}
		out = append(out, violation{"MD009", i + 1, len(trimmedRight) + 1,
			fmt.Sprintf("%d trailing space(s)", trailing)})
	}
	return out
}

// ---- MD010 no-hard-tabs -------------------------------------------------------

type MD010Options struct {
	// SpacesPerTab expands a tab to this width for column reporting.
	SpacesPerTab *int `json:"spaces_per_tab"`
	// IncludeCodeBlocks also checks lines inside fenced code. The key is
	// markdownlint's `code_blocks`, not a name of our own: the whole premise of
	// this linter is that a .markdownlint.json stays portable, so inventing a
	// key here would reject every config written for the tool it replaces.
	IncludeCodeBlocks *bool `json:"code_blocks"`
}

func checkMD010(d *doc, rs *ruleSet) []violation {
	o := opt[MD010Options](rs, "MD010")
	includeCode := o.IncludeCodeBlocks == nil || *o.IncludeCodeBlocks
	var out []violation
	for i := 0; i < d.n(); i++ {
		if d.inFrontMatter(i) || (d.inCode(i) && !includeCode) {
			continue
		}
		if strings.ContainsRune(d.lines[i], '\t') {
			out = append(out, violation{"MD010", i + 1, 1, "hard tab"})
		}
	}
	return out
}

// ---- MD012 no-multiple-blanks -------------------------------------------------

type MD012Options struct {
	Maximum *int `json:"maximum"`
}

func checkMD012(d *doc, rs *ruleSet) []violation {
	o := opt[MD012Options](rs, "MD012")
	max := 1
	if o.Maximum != nil {
		max = *o.Maximum
	}
	var out []violation
	run := 0
	for i := 0; i < d.n(); i++ {
		if d.inFrontMatter(i) {
			continue
		}
		if strings.TrimSpace(d.lines[i]) == "" {
			run++
			continue
		}
		if run > max {
			out = append(out, violation{"MD012", i + 1, 1,
				fmt.Sprintf("%d consecutive blank lines (maximum %d)", run, max)})
		}
		run = 0
	}
	return out
}

// ---- MD013 line-length --------------------------------------------------------

type MD013Options struct {
	LineLength          *int  `json:"line_length"`
	HeadingLineLength   *int  `json:"heading_line_length"`
	CodeBlockLineLength *int  `json:"code_block_line_length"`
	CodeBlocks          *bool `json:"code_blocks"`
	Headings            *bool `json:"headings"`
	Tables              *bool `json:"tables"`
	Strict              *bool `json:"strict"`
	StrictLineLength    *int  `json:"-"`
}

func checkMD013(d *doc, rs *ruleSet) []violation {
	o := opt[MD013Options](rs, "MD013")

	var out []violation
	for i := 0; i < d.n(); i++ {
		max, applies := md013Limit(d, i, o)
		if !applies {
			continue
		}
		if len([]rune(d.lines[i])) > max {
			out = append(out, md013Violation(i, d.lines[i], max))
		}
	}
	return out
}

// md013Limit is the per-line exemption switch MD013 applies, shared with the
// fixer so the two cannot disagree about which lines are in scope.
//
// It returns the character limit that applies to line i and whether the rule
// applies to it at all. The switch is order-sensitive and deliberately so: a
// heading inside a code block is still a heading as far as this rule is
// concerned, because the masks are exclusive and a heading cannot be inside a
// fence. Splitting the decision out here — rather than leaving the fixer to
// re-derive "is this prose?" — is the whole point: a fixer that wraps a line the
// checker never looked at produces a document the checker would have to be
// re-run to notice.
func md013Limit(d *doc, i int, o MD013Options) (limit int, applies bool) {
	if d.inFrontMatter(i) {
		return 0, false
	}
	lineLength := 80
	if o.LineLength != nil {
		lineLength = *o.LineLength
	}
	headingLineLength := lineLength
	if o.HeadingLineLength != nil {
		headingLineLength = *o.HeadingLineLength
	}
	codeBlockLineLength := lineLength
	if o.CodeBlockLineLength != nil {
		codeBlockLineLength = *o.CodeBlockLineLength
	}
	checkCode := o.CodeBlocks == nil || *o.CodeBlocks
	checkHeadings := o.Headings == nil || *o.Headings
	checkTables := o.Tables == nil || *o.Tables

	switch {
	case d.heading[i] > 0:
		if !checkHeadings {
			return 0, false
		}
		return headingLineLength, true
	case d.inCode(i):
		if !checkCode {
			return 0, false
		}
		return codeBlockLineLength, true
	case d.table[i]:
		if !checkTables {
			return 0, false
		}
	}
	return lineLength, true
}

func md013Violation(i int, line string, limit int) violation {
	runes := []rune(line)
	shown := limit
	if len(runes) > limit+20 {
		shown = limit + 20
	}
	return violation{"MD013", i + 1, 1,
		fmt.Sprintf("line %d characters (limit %d): %s…", len(runes), limit, string(runes[:shown]))}
}

// ---- MD022 blanks-around-headings ---------------------------------------------

type MD022Options struct {
	LinesAbove *int `json:"lines_above"`
	LinesBelow *int `json:"lines_below"`
}

func checkMD022(d *doc, rs *ruleSet) []violation {
	o := opt[MD022Options](rs, "MD022")
	above, below := 1, 1
	if o.LinesAbove != nil {
		above = *o.LinesAbove
	}
	if o.LinesBelow != nil {
		below = *o.LinesBelow
	}
	var out []violation
	for i := 0; i < d.n(); i++ {
		if d.heading[i] == 0 {
			continue
		}
		if above > 0 {
			// Nothing precedes the first heading in a document (or the first
			// heading after front matter), so there is nothing to separate.
			firstContent := i == 0 || (d.fmEnd > 0 && i == d.fmEnd)
			if !firstContent && strings.TrimSpace(d.lines[i-1]) != "" {
				out = append(out, violation{"MD022", i + 1, 1, "no blank line above the heading"})
			}
		}
		// A setext heading ends at its underline, so "below" is measured from
		// there rather than from the line holding the text.
		end := i
		if d.headingEnd[i] > end {
			end = d.headingEnd[i]
		}
		if below > 0 && end+1 < d.n() {
			if strings.TrimSpace(d.lines[end+1]) != "" && d.heading[end+1] == 0 {
				out = append(out, violation{"MD022", i + 1, 1, "no blank line below the heading"})
			}
		}
	}
	return out
}

// ---- MD025 single-title -------------------------------------------------------

type MD025Options struct {
	Level            *int   `json:"level"`
	FrontMatterTitle string `json:"front_matter_title"`
}

// ---- MD031 blanks-around-fences -----------------------------------------------

type MD031Options struct {
	ListItems *bool `json:"list_items"`
}

func checkMD031(d *doc, rs *ruleSet) []violation {
	_ = opt[MD031Options](rs, "MD031")
	var out []violation
	for i := 0; i < d.n(); i++ {
		if !d.fenceLines[i] {
			continue
		}
		// No blank line required directly after an opening fence, or directly
		// before a closing fence, when the block is empty.
		if i > 0 && !d.fenceLines[i-1] && strings.TrimSpace(d.lines[i-1]) != "" && !d.inCode(i-1) {
			out = append(out, violation{"MD031", i + 1, 1, "fenced code block should be preceded by a blank line"})
			continue
		}
		// "Followed by a blank line" is a statement about the line after the
		// *closing* fence. Asking it of an opening fence means asking for a
		// blank line inside the code block, and it fired on every non-empty
		// block in the document — 171 of the 1,092-count era's MD031 reports
		// were this. The fixer had been inserting the blank line it asked for,
		// which is how a check that reported the wrong thing stayed invisible:
		// its bogus finding was being satisfied by a real edit.
		if d.fenceOpens[i] {
			continue
		}
		if i+1 < d.n() && !d.fenceLines[i+1] && strings.TrimSpace(d.lines[i+1]) != "" {
			out = append(out, violation{"MD031", i + 1, 1, "fenced code block should be followed by a blank line"})
		}
	}
	return out
}

// ---- MD032 blanks-around-lists ------------------------------------------------

type MD032Options struct {
	Lists *bool `json:"lists"`
}

// isIndented reports whether a line's own indentation makes it a continuation
// of the block above rather than a new block.
func isIndented(line string) bool {
	return strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "	")
}

func checkMD032(d *doc, rs *ruleSet) []violation {
	_ = opt[MD032Options](rs, "MD032")
	var out []violation
	for i := 0; i < d.n(); i++ {
		if !d.list[i] {
			continue
		}
		// A list that continues from the previous line (nested item, or a
		// wrapped item) does not need a blank line above. The indent test reads
		// the raw line, not the trimmed one: a wrapped item's continuation is
		// distinguished by its indentation, which trimming would erase.
		if i > 0 {
			prev := d.lines[i-1]
			continues := d.list[i-1] || isIndented(prev) ||
				strings.HasSuffix(strings.TrimRight(prev, " "), ">") || d.fence[i-1] || d.fenceLines[i-1]
			if !continues && strings.TrimSpace(prev) != "" && !d.inCode(i-1) {
				out = append(out, violation{"MD032", i + 1, 1, "list should be preceded by a blank line"})
			}
		}
		// A list that ends here needs a blank line after its last line.
		if i+1 < d.n() {
			next := strings.TrimSpace(d.lines[i+1])
			follows := d.list[i+1] || isIndented(d.lines[i+1]) ||
				strings.HasPrefix(next, "- ") || strings.HasPrefix(next, "* ") ||
				strings.HasPrefix(next, "> ") || d.fenceLines[i+1] || d.fence[i+1]
			if follows {
				continue
			}
			if next != "" {
				out = append(out, violation{"MD032", i + 1, 1, "list should be followed by a blank line"})
			}
		}
	}
	return out
}

// ---- MD040 fenced-code-language -----------------------------------------------

type MD040Options struct {
	AllowedLanguages *[]string `json:"allowed_languages"`
	LanguageOnly     *bool     `json:"language_only"`
}

func checkMD040(d *doc, rs *ruleSet) []violation {
	o := opt[MD040Options](rs, "MD040")
	languageOnly := o.LanguageOnly == nil || *o.LanguageOnly
	var allowed map[string]bool
	if o.AllowedLanguages != nil {
		allowed = map[string]bool{}
		for _, l := range *o.AllowedLanguages {
			allowed[strings.ToLower(l)] = true
		}
	}
	var out []violation
	for i := 0; i < d.n(); i++ {
		// Only opening fences carry an info string; a closing one is empty by
		// definition and would be reported as a block with no language.
		if !d.fenceOpens[i] {
			continue
		}
		info := fenceInfo(d.lines[i])
		if info.lang == "" {
			out = append(out, violation{"MD040", i + 1, 1, "fenced code block has no language"})
			continue
		}
		if allowed != nil && !allowed[strings.ToLower(info.lang)] {
			out = append(out, violation{"MD040", i + 1, 1,
				"language " + info.lang + " is not in allowed_languages"})
			continue
		}
		if languageOnly && strings.TrimSpace(info.rest) != "" {
			out = append(out, violation{"MD040", i + 1, 1,
				"language_only: info string must be just the language, got " + strings.TrimSpace(info.rest)})
		}
	}
	return out
}

type fence struct {
	lang string
	rest string
}

func fenceInfo(line string) fence {
	trimmed := strings.TrimLeft(line, " ")
	idx := 0
	for idx < len(trimmed) && (trimmed[idx] == '`' || trimmed[idx] == '~') {
		idx++
	}
	rest := strings.TrimSpace(trimmed[idx:])
	if rest == "" {
		return fence{}
	}
	fields := strings.Fields(rest)
	return fence{lang: fields[0], rest: strings.TrimSpace(strings.TrimPrefix(rest, fields[0]))}
}

// ---- MD047 single-trailing-newline --------------------------------------------

type MD047Options struct {
	// Strict requires exactly one trailing newline; otherwise a missing one is
	// still reported, just as "no newline at end of file".
	Strict *bool `json:"-"`
}

func checkMD047(d *doc, _ *ruleSet) []violation {
	if len(d.src) == 0 {
		return nil
	}
	src := string(d.src)
	if strings.HasSuffix(src, "\n\n") {
		return []violation{{"MD047", d.n(), 1, "multiple trailing newlines"}}
	}
	if !strings.HasSuffix(src, "\n") {
		return []violation{{"MD047", d.n() + 1, 1, "no newline at end of file"}}
	}
	return nil
}

// ---- MD058 blanks-around-tables -----------------------------------------------

type MD058Options struct {
	// Tables is accepted for config compatibility; the rule is about tables.
	Tables *bool `json:"tables"`
}

func checkMD058(d *doc, rs *ruleSet) []violation {
	_ = opt[MD058Options](rs, "MD058")
	var out []violation
	seen := map[int]bool{}
	for i := 0; i < d.n(); i++ {
		if !d.table[i] || seen[i] {
			continue
		}
		// Report the block's first line for a missing blank above...
		if i > 0 && !d.table[i-1] && strings.TrimSpace(d.lines[i-1]) != "" && !d.inCode(i-1) {
			out = append(out, violation{"MD058", i + 1, 1, "table should be preceded by a blank line"})
		}
		// ...and its last line for a missing blank below.
		j := i
		for j+1 < d.n() && d.table[j+1] {
			j++
		}
		seen[j] = true
		if j+1 < d.n() && strings.TrimSpace(d.lines[j+1]) != "" && !d.fenceLines[j+1] {
			out = append(out, violation{"MD058", j + 1, 1, "table should be followed by a blank line"})
		}
	}
	return out
}

// ---- MD004 unordered-list-style -----------------------------------------------

type MD004Options struct {
	Style string `json:"style"`
}

func checkMD004(d *doc, rs *ruleSet) []violation {
	o := opt[MD004Options](rs, "MD004")
	style := o.Style
	if style == "" {
		style = "consistent"
	}
	var first string
	var out []violation
	for i := 0; i < d.n(); i++ {
		if d.inFrontMatter(i) || d.inCode(i) {
			continue
		}
		trimmed := strings.TrimLeft(d.lines[i], " ")
		if len(trimmed) < 2 || (trimmed[0] != '-' && trimmed[0] != '*' && trimmed[0] != '+') || trimmed[1] != ' ' {
			continue
		}
		marker := string(trimmed[0])
		if style == "consistent" {
			if first == "" {
				first = marker
			}
			if marker != first {
				out = append(out, violation{"MD004", i + 1, 1,
					"unordered list marker " + marker + " is inconsistent with " + first + " (style: consistent)"})
			}
			continue
		}
		if marker != styleMarker(style) {
			out = append(out, violation{"MD004", i + 1, 1,
				"unordered list marker " + marker + " (style: " + style + ")"})
		}
	}
	return out
}

// styleMarker maps a markdownlint style name to the marker character it requires.
func styleMarker(style string) string {
	switch style {
	case "dash":
		return "-"
	case "asterisk":
		return "*"
	case "plus":
		return "+"
	}
	return style
}

// ---- MD007 unordered-list-indent ---------------------------------------------

type MD007Options struct {
	Indent        *int  `json:"indent"`
	StartIndented *bool `json:"start_indented"`
	StartIndent   *int  `json:"start_indent"`
}

func checkMD007(d *doc, rs *ruleSet) []violation {
	o := opt[MD007Options](rs, "MD007")
	indent := 2
	if o.Indent != nil {
		indent = *o.Indent
	}
	startIndented := o.StartIndented != nil && *o.StartIndented
	startIndent := indent
	if o.StartIndent != nil {
		startIndent = *o.StartIndent
	}
	var out []violation
	for i := 0; i < d.n(); i++ {
		if d.inFrontMatter(i) || d.inCode(i) {
			continue
		}
		lead, offending := md007Offender(d.lines[i], indent, startIndented, startIndent)
		if !offending {
			continue
		}
		out = append(out, violation{"MD007", i + 1, lead + 1,
			fmt.Sprintf("unordered list indentation %d spaces (expected a multiple of %d)", lead, indent)})
	}
	return out
}

// md007Offender is MD007's per-line decision, shared with fixMD007.
//
// It returns the line's leading-space count and whether the line is an unordered
// list item whose indentation the rule objects to. Sharing this is what stops
// the fixer from re-indenting lines the checker never reported — the failure
// this file's MD032 and MD031 comments describe, where a fixer with its own
// idea of the predicate edits lines that are not violations and leaves the real
// violations in place.
func md007Offender(line string, indent int, startIndented bool, startIndent int) (lead int, offending bool) {
	if indent < 1 {
		// A zero width has no multiple to test against. The old code divided by
		// it; refusing the rule is better than a panic in a fixer.
		return 0, false
	}
	trimmed := strings.TrimLeft(line, " ")
	lead = len(line) - len(trimmed)
	if lead == 0 || len(trimmed) < 2 {
		return lead, false
	}
	if trimmed[0] != '-' && trimmed[0] != '*' && trimmed[0] != '+' {
		return lead, false
	}
	if trimmed[1] != ' ' {
		return lead, false
	}
	want := indent
	if lead <= startIndent && startIndented {
		want = startIndent
	}
	if lead%indent == 0 || lead%want == 0 {
		return lead, false
	}
	return lead, true
}

// ---- MD029 ordered-list-prefix ------------------------------------------------

type MD029Options struct {
	Style string `json:"style"`
}

// listRun is one ordered list: the lines its markers sit on and the numbers
// they carry, in document order.
type listRun struct {
	indent int
	lines  []int // 1-based
	nums   []int
}

func (r *listRun) push(line, num int) {
	r.lines = append(r.lines, line)
	r.nums = append(r.nums, num)
}

// allEqual reports whether every marker carries the same number.
func (r *listRun) allEqual(want int) bool {
	for _, n := range r.nums {
		if n != want {
			return false
		}
	}
	return true
}

// increments reports whether the numbers run 1, 2, 3, ... in order.
func (r *listRun) increments() bool {
	for i, n := range r.nums {
		if n != i+1 {
			return false
		}
	}
	return true
}

func checkMD029(d *doc, rs *ruleSet) []violation {
	style := opt[MD029Options](rs, "MD029").Style
	if style == "" {
		style = "one_or_ordered"
	}
	var out []violation
	for _, r := range orderedRuns(d) {
		out = append(out, checkListRun(style, r)...)
	}
	return out
}

// orderedRuns groups the document's ordered-list markers into the same lists
// checkMD029 used to build inline, and is shared with fixMD029.
//
// "What counts as one list" is the whole difficulty in MD029: a nested list
// ends when a marker at the parent's indent reappears, a blank line keeps a
// loose list together, and an indented line continues the innermost open list.
// A fixer that decided that question on its own would renumber items in a
// different grouping from the one the rule reported against, which is how a
// document can come out of a fix pass with the violation count unchanged. The
// grouping lives here, once, so that cannot happen.
func orderedRuns(d *doc) []*listRun {
	var stack []*listRun
	var out []*listRun

	// A line belongs to the innermost open list when it is blank (a loose list
	// keeps its items together) or indented past that list's markers.
	continues := func(indent int) bool {
		return len(stack) > 0 && indent > stack[len(stack)-1].indent
	}
	flushAll := func() {
		for j := len(stack) - 1; j >= 0; j-- {
			out = append(out, stack[j])
		}
		stack = nil
	}

	for i := 0; i < d.n(); i++ {
		if d.inFrontMatter(i) || d.inCode(i) {
			flushAll()
			continue
		}
		line := d.lines[i]
		indent := leadingSpaces(line)
		if num, _, ok := orderedMarker(line); ok {
			// A marker at or below the innermost open list's indent closes any
			// deeper list first: that is how a nested list ends and its parent
			// resumes.
			for len(stack) > 0 && stack[len(stack)-1].indent > indent {
				out = append(out, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) > 0 && stack[len(stack)-1].indent == indent {
				stack[len(stack)-1].push(i+1, num)
				continue
			}
			stack = append(stack, &listRun{indent: indent, lines: []int{i + 1}, nums: []int{num}})
			continue
		}
		if strings.TrimSpace(line) == "" || continues(indent) {
			continue
		}
		flushAll()
	}
	flushAll()
	return out
}

// checkListRun applies the configured style to one list.
func checkListRun(style string, r *listRun) []violation {
	report := func(i int, msg string) violation {
		return violation{"MD029", r.lines[i], 1, msg}
	}
	var out []violation
	switch style {
	case "one":
		for i, n := range r.nums {
			if n != 1 {
				out = append(out, report(i, "ordered list item must start at 1 (style: one)"))
			}
		}
	case "zero":
		for i, n := range r.nums {
			if n != 0 {
				out = append(out, report(i, "ordered list item must start at 0 (style: zero)"))
			}
		}
	case "ordered":
		for i, n := range r.nums {
			if want := i + 1; n != want {
				out = append(out, report(i, fmt.Sprintf("ordered list item numbered %d, expected %d (style: ordered)", n, want)))
			}
		}
	case "one_or_ordered":
		if r.allEqual(1) || r.increments() {
			return nil
		}
		for i, n := range r.nums {
			if n != 1 && n != i+1 {
				out = append(out, report(i, fmt.Sprintf("ordered list item numbered %d (style: one_or_ordered)", n)))
			}
		}
	default:
		for i, n := range r.nums {
			if n > 1 {
				out = append(out, report(i, fmt.Sprintf("ordered list item numbered %d (style: %s)", n, style)))
			}
		}
	}
	return out
}

func leadingSpaces(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func orderedMarker(line string) (num int, delim byte, ok bool) {
	trimmed := strings.TrimLeft(line, " ")
	idx := 0
	for idx < len(trimmed) && unicode.IsDigit(rune(trimmed[idx])) {
		idx++
	}
	if idx == 0 || idx >= len(trimmed) {
		return 0, 0, false
	}
	if trimmed[idx] != '.' && trimmed[idx] != ')' {
		return 0, 0, false
	}
	if idx+1 >= len(trimmed) || trimmed[idx+1] != ' ' {
		return 0, 0, false
	}
	for _, r := range trimmed[:idx] {
		num = num*10 + int(r-'0')
	}
	return num, trimmed[idx], true
}

// ---- MD026 no-trailing-punctuation-in-heading ---------------------------------

type MD026Options struct {
	Punctuation *string `json:"punctuation"`
}

func checkMD026(d *doc, rs *ruleSet) []violation {
	o := opt[MD026Options](rs, "MD026")
	punctuation := ".,;:!。，；：！"
	if o.Punctuation != nil {
		punctuation = *o.Punctuation
	}
	if punctuation == "" {
		return nil
	}
	var out []violation
	for i := 0; i < d.n(); i++ {
		if d.heading[i] == 0 {
			continue
		}
		text := strings.TrimSpace(strings.Trim(d.headingText[i], "#"))
		if text == "" {
			continue
		}
		last := []rune(text)
		if !strings.ContainsRune(punctuation, last[len(last)-1]) {
			continue
		}
		// A heading that is only punctuation is exempt.
		if strings.Trim(string(last), punctuation) == "" {
			continue
		}
		out = append(out, violation{"MD026", i + 1, 1, "heading ends with punctuation"})
	}
	return out
}
