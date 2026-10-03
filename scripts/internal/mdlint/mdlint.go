// Package mdlint lints Markdown from Go, so the framework's documentation
// checks run with the same binary that installs the framework — no Node, no npm,
// and no second toolchain in CI.
//
// It implements the markdownlint rule set directly on top of goldmark, using the
// same rule IDs and config keys, plus the three GitHub accessibility rules that
// live in an npm-only package (see githubrules.go). The npm runner it replaces
// (markdownlint-cli2 + @github/markdownlint-github) is not an option here: that
// setup needed Node, and its shareable config could no longer be resolved.
//
// Two behaviors are deliberate and differ from the JavaScript tooling:
//
//   - Unknown rule names and unknown option keys are hard errors, never silent
//     no-ops. A config that half-loads is a check that reports success while
//     checking nothing.
//   - 'extends' is rejected. There is no registry to resolve it against, so the
//     rules a project needs must live in the config file.
//
// Rule coverage is documented in coverage.go and reported by
// `plaesy validate markdown --list-rules`; it is a subset of markdownlint's ~50
// rules, chosen to cover the violations this repository actually has.
package mdlint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// violation is one reported problem.
type violation struct {
	Rule    string
	Line    int
	Column  int
	Message string
}

// Config is a markdownlint-shaped configuration file: a default switch plus a map
// of rule ID (or alias) to either a boolean or a rule-options object.
type Config struct {
	Default bool
	Rules   map[string]json.RawMessage
}

// DefaultConfigName is the config file Lint and Fix read from the tree root when
// no config is passed in.
const DefaultConfigName = ".markdownlint.json"

// DefaultBaselineName is the ratchet file the CLI reads by default.
const DefaultBaselineName = ".markdownlint-baseline.json"

// DefaultIgnoreDirs are never walked: vendored dependencies, VCS metadata, the
// framework's own generated state, and release artifacts.
//
// `.claude` and `.kilo` are the per-platform command directories that
// `plaesy init` generates from `prompts/`. They are copies, not sources: this
// repository holds 69 prompt files and each of those platforms holds the same
// 69 again. Walking them meant every violation in a prompt was counted once per
// platform directory that happened to exist — three times here — and, worse,
// the total depended on which platforms the machine had initialised. A ratchet
// whose number moves with local state cannot hold a line: at HEAD the recorded
// ceiling of 1,860 had been measured against a tree without `.kilo/`, and the
// build was red on a clean checkout of the very commit that recorded it.
// Sources are linted; their generated copies are not, and any violation in
// `prompts/` still counts exactly once.
var DefaultIgnoreDirs = []string{
	".git", "node_modules", ".plaesy", "dist", "vendor", "bin", ".idea", ".vscode",
	".claude", ".kilo",
}

// Baseline is a ratchet ceiling: a run passes while the violation total stays at
// or below MaxViolations and fails when it rises. A zero ceiling is rejected —
// "no violations" is expressed by passing no baseline at all.
type Baseline struct {
	MaxViolations int            `json:"max_violations"`
	ByRule        map[string]int `json:"by_rule,omitempty"`
	MeasuredAt    string         `json:"measured_at"`
	Note          string         `json:"note"`
}

// Result summarizes a lint run.
type Result struct {
	Files      int            // files with at least one violation
	Scanned    int            // files scanned
	Total      int            // total violations
	ByRule     map[string]int // rule ID -> count
	ByFile     map[string]int // repo-relative path -> count
	Violations []fileResult   // per-file detail, worst first
}

type fileResult struct {
	Path   string
	Total  int
	Detail []violation
}

// goldmarkParser is configured once: GFM, so tables and task lists parse the way
// GitHub renders them.
//
// goldmark v1.7.8 has no built-in front-matter extension (it lives in the
// third-party abhinav/goldmark-frontmatter), so the YAML block is parsed as
// ordinary body markdown and every `#` comment inside it becomes a real
// ast.Heading. AST-based rules must therefore skip front matter by line -- see
// checkMD025. The text-based rules were always immune, because they go through
// markFrontMatter/inFrontMatter.
var goldmarkParser = goldmark.New(goldmark.WithExtensions(extension.GFM))

// Lint walks the given paths (or the whole root) and reports every violation.
// A nil config uses the built-in defaults (all implemented rules on, default
// options).
func Lint(root string, paths []string, cfg *Config, ignoreDirs []string) (*Result, error) {
	if cfg == nil {
		var err error
		if cfg, err = LoadConfig(filepath.Join(root, DefaultConfigName)); err != nil {
			return nil, err
		}
	}
	enabled, err := compile(cfg)
	if err != nil {
		return nil, err
	}
	files, err := collectFiles(root, paths, ignoreDirs)
	if err != nil {
		return nil, err
	}
	res := &Result{ByRule: map[string]int{}, ByFile: map[string]int{}}
	for _, path := range files {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, readErr
		}
		rel := relPath(root, path)
		found := checkFile(enabled, rel, data)
		res.Scanned++
		if len(found) == 0 {
			continue
		}
		res.Files++
		res.Total += len(found)
		res.ByFile[rel] = len(found)
		for _, v := range found {
			res.ByRule[v.Rule]++
		}
		res.Violations = append(res.Violations, fileResult{Path: rel, Total: len(found), Detail: found})
	}
	sort.Slice(res.Violations, func(i, j int) bool {
		if res.Violations[i].Total != res.Violations[j].Total {
			return res.Violations[i].Total > res.Violations[j].Total
		}
		return res.Violations[i].Path < res.Violations[j].Path
	})
	return res, nil
}

// Fix applies the machine-fixable rules (trailing spaces, hard tabs, repeated
// blanks, heading/list/fence/table spacing, final newline, fence style) in place
// and reports whether anything changed.
func Fix(root string, paths []string, cfg *Config, ignoreDirs []string) (bool, error) {
	if cfg == nil {
		var err error
		if cfg, err = LoadConfig(filepath.Join(root, DefaultConfigName)); err != nil {
			return false, err
		}
	}
	enabled, err := compile(cfg)
	if err != nil {
		return false, err
	}
	files, err := collectFiles(root, paths, ignoreDirs)
	if err != nil {
		return false, err
	}
	changedAny := false
	for _, path := range files {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			// readErr, not err: `err` is collectFiles' and is nil here, so
			// returning it reported a read failure as "nothing to fix" and
			// abandoned every remaining file in the batch with exit 0.
			return false, readErr
		}
		fixed := fixSource(enabled, data)
		if bytes.Equal(fixed, data) {
			continue
		}
		if writeErr := os.WriteFile(path, fixed, 0o644); writeErr != nil {
			return changedAny, writeErr
		}
		changedAny = true
	}
	return changedAny, nil
}

// checkFile runs every enabled rule over one document.
func checkFile(rs *ruleSet, path string, src []byte) []violation {
	d := newDoc(path, src)
	var out []violation
	for _, r := range rs.rules {
		out = append(out, r.check(d, rs)...)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Line != out[j].Line {
			return out[i].Line < out[j].Line
		}
		return out[i].Rule < out[j].Rule
	})
	return out
}

// doc is one parsed document plus the line masks the text rules need.
type doc struct {
	path  string
	src   []byte
	lines []string
	// fmStart/fmEnd bound YAML front matter (0-based, end exclusive); empty when absent.
	fmStart, fmEnd int
	// fence[i] is true when line i is inside a fenced code block.
	fence []bool
	// fenceLines[i] is true when line i is a fence delimiter itself.
	fenceLines []bool
	// fenceOpens[i] is true when line i *opens* a fenced code block. The
	// distinction matters for MD040: a closing fence carries no info string, so
	// treating it as an opening one reports every block in the document as
	// having no language.
	fenceOpens []bool
	// heading[i] is the ATX/setext heading level on line i, or 0.
	heading []int
	// headingText[i] is the heading text on line i, or "".
	headingText []string
	// headingEnd[i] is the last line of the heading anchored at line i. It equals
	// i for an ATX heading and reaches the underline for a setext one, whose
	// heading text spans the paragraph above it.
	headingEnd []int
	// list[i] is true when line i starts a list item at the outermost level.
	list []bool
	// table[i] is true when line i is part of a GFM table.
	table []bool
	// ast is the parsed document (lazily, only for AST rules).
	root ast.Node
	// source is the goldmark source view over the document.
	source []byte
}

func newDoc(path string, src []byte) *doc {
	d := &doc{path: path, src: src, source: src}
	d.lines = splitLines(src)
	d.markFrontMatter()
	d.markFences()
	d.markHeadings()
	d.markLists()
	d.markTables()
	return d
}

func splitLines(src []byte) []string {
	s := string(src)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	// A trailing newline yields an empty final element that is not a real line.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func (d *doc) n() int { return len(d.lines) }

// inFrontMatter reports whether the 0-based line index sits in YAML front matter.
func (d *doc) inFrontMatter(i int) bool { return i >= d.fmStart && i < d.fmEnd }

// isFenceLine reports whether the line is a fence delimiter itself. It
// bounds-checks because the fix passes read masks that may lag the working
// slice; a stale mask must mean "no opinion", not a panic.
func (d *doc) isFenceLine(i int) bool { return i >= 0 && i < len(d.fenceLines) && d.fenceLines[i] }

// inCode reports whether the line is inside a fenced code block (not the fence
// delimiters themselves).
func (d *doc) inCode(i int) bool { return i < len(d.fence) && d.fence[i] }

// frontMatterTitleField returns the field name whose value counts as the
// document's top-level heading, or "" when front matter titles are ignored.
func (d *doc) frontMatterTitleField(pattern string) string {
	if pattern == "" {
		return "title"
	}
	if pattern == "^$" {
		return ""
	}
	return pattern
}

func (d *doc) frontMatterTitle() (string, bool) {
	if d.fmEnd == 0 {
		return "", false
	}
	field := d.frontMatterTitleField(frontMatterTitlePattern)
	for i := d.fmStart; i < d.fmEnd; i++ {
		key, value, found := strings.Cut(d.lines[i], ":")
		if !found {
			continue
		}
		if strings.TrimSpace(key) == field {
			return strings.Trim(strings.TrimSpace(value), `"'`), true
		}
	}
	return "", false
}

// frontMatterTitlePattern is the MD025/MD041 setting, resolved from the config
// at rule-compile time.
var frontMatterTitlePattern string

// markFrontMatter records the YAML front matter bounds.
func (d *doc) markFrontMatter() {
	if len(d.lines) == 0 || strings.TrimSpace(d.lines[0]) != "---" {
		return
	}
	for i := 1; i < len(d.lines); i++ {
		trimmed := strings.TrimSpace(d.lines[i])
		if trimmed == "---" || trimmed == "..." {
			d.fmStart, d.fmEnd = 1, i
			return
		}
	}
}

// markFences records which lines sit inside fenced code blocks.
func (d *doc) markFences() {
	d.fence = make([]bool, d.n())
	d.fenceLines = make([]bool, d.n())
	d.fenceOpens = make([]bool, d.n())
	var marker string
	for i := 0; i < d.n(); i++ {
		if d.inFrontMatter(i) {
			continue
		}
		line := d.lines[i]
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent > 3 {
			// Four or more leading spaces is an indented code block, so such
			// a line cannot open or close a fence. It is not a reason to
			// forget that we are inside one, though: when a block is already
			// open this line is that block's content, and marking it as such
			// is what inCode() reports. Skipping the marking here left
			// indented content looking like prose, and MD031 then reported
			// the closing fence of such a block as "should be preceded by a
			// blank line" — a finding about a line the author cannot fix
			// without corrupting the code it is reporting on.
			if marker != "" {
				d.fence[i] = true
			}
			continue
		}
		trimmed := strings.TrimLeft(line, " ")
		if marker == "" {
			if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
				char := trimmed[:1]
				count := len(trimmed) - len(strings.TrimLeft(trimmed, char))
				if count >= 3 {
					marker = strings.Repeat(char, count)
					d.fenceLines[i] = true
					d.fenceOpens[i] = true
				}
			}
			continue
		}
		if strings.HasPrefix(trimmed, marker) {
			d.fenceLines[i] = true
			marker = ""
			continue
		}
		d.fence[i] = true
	}
}

// markHeadings records heading level and text per line (ATX and setext).
func (d *doc) markHeadings() {
	d.heading = make([]int, d.n())
	d.headingText = make([]string, d.n())
	d.headingEnd = make([]int, d.n())
	for i := 0; i < d.n(); i++ {
		if d.inFrontMatter(i) || d.inCode(i) {
			continue
		}
		line := d.lines[i]
		trimmed := strings.TrimLeft(line, " ")
		if indent := len(line) - len(trimmed); indent > 3 {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			level := 0
			for level < len(trimmed) && trimmed[level] == '#' {
				level++
			}
			if level <= 6 {
				d.heading[i] = level
				d.headingEnd[i] = i
				d.headingText[i] = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(trimmed[level:]), "#"))
			}
			continue
		}
		// Setext: the paragraph above a = or - underline becomes a heading.
		//
		// The heading spans that whole paragraph, so it is anchored at the
		// paragraph's FIRST line and carries the paragraph's full text, which is
		// where goldmark anchors it too. Anchoring at the line directly above the
		// underline is wrong in a way that is invisible until a document uses the
		// common "paragraph, then --- separator" shape: every continuation line
		// then looks like a heading with content jammed against it, and the rules
		// report phantom MD022 (no blank line above) and MD026 (trailing
		// punctuation) on prose that is not a heading at all.
		if i > 0 && (strings.HasPrefix(trimmed, "=") || strings.HasPrefix(trimmed, "-")) && isSetextUnderline(trimmed) {
			start := setextStart(d, i)
			if start >= 0 && d.heading[start] == 0 {
				level := 1
				if strings.HasPrefix(trimmed, "-") {
					level = 2
				}
				d.heading[start] = level
				d.headingEnd[start] = i
				d.headingText[start] = strings.Join(trimBlank(d.lines[start:i]), " ")
			}
		}
	}
}

// setextStart walks back from a setext underline to the first line of the
// paragraph it underlines, or returns -1 when there is no paragraph to underline.
// The walk stops at a blank line, a code block, front matter, an existing
// heading, or a line that opens another block (list, quote, fence): a `---` after
// those is a thematic break, not a heading underline.
func setextStart(d *doc, underline int) int {
	start := -1
	for j := underline - 1; j >= 0; j-- {
		line := d.lines[j]
		trimmed := strings.TrimLeft(line, " ")
		if strings.TrimSpace(line) == "" || d.inCode(j) || d.inFrontMatter(j) {
			break
		}
		if strings.HasPrefix(trimmed, "#") || isFenceDelimiter(trimmed) ||
			strings.HasPrefix(trimmed, ">") || isListStart(trimmed) {
			break
		}
		start = j
	}
	return start
}

// trimBlank joins the lines of a span, dropping a hard line break's trailing
// spaces so the text reads the way the heading renders.
func trimBlank(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, strings.TrimRight(l, " \t"))
	}
	return out
}

// isFenceDelimiter reports whether a line opens or closes a fenced code block.
func isFenceDelimiter(trimmed string) bool {
	if len(trimmed) < 3 {
		return false
	}
	c := trimmed[0]
	if c != '`' && c != '~' {
		return false
	}
	n := 0
	for n < len(trimmed) && trimmed[n] == c {
		n++
	}
	return n >= 3
}

func isSetextUnderline(s string) bool {
	if s == "" {
		return false
	}
	char := s[:1]
	if char != "=" && char != "-" {
		return false
	}
	return strings.Trim(s, char) == ""
}

// markLists records lines that start a top-level list item.
func (d *doc) markLists() {
	d.list = make([]bool, d.n())
	for i := 0; i < d.n(); i++ {
		if d.inFrontMatter(i) || d.inCode(i) {
			continue
		}
		line := d.lines[i]
		trimmed := strings.TrimLeft(line, " ")
		indent := len(line) - len(trimmed)
		if indent > 3 {
			continue
		}
		if isListStart(trimmed) {
			d.list[i] = true
		}
	}
}

func isListStart(trimmed string) bool {
	if trimmed == "" {
		return false
	}
	switch trimmed[0] {
	case '-', '*', '+':
		// "- " starts a list; "---" is a thematic break.
		return len(trimmed) > 1 && trimmed[1] == ' '
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		rest := trimmed
		for len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
			rest = rest[1:]
		}
		return strings.HasPrefix(rest, ". ") || strings.HasPrefix(rest, ") ")
	}
	return false
}

// markTables records lines belonging to a GFM table (header, delimiter, body).
func (d *doc) markTables() {
	d.table = make([]bool, d.n())
	for i := 0; i+1 < d.n(); i++ {
		if d.inFrontMatter(i) || d.inCode(i) || !strings.Contains(d.lines[i], "|") {
			continue
		}
		if !isTableDelimiter(d.lines[i+1]) {
			continue
		}
		// Walk up to the first non-table line and down through the body.
		start := i
		for start > 0 && !d.inFrontMatter(start) && !d.inCode(start) && strings.Contains(d.lines[start-1], "|") && strings.TrimSpace(d.lines[start-1]) != "" {
			start--
		}
		for j := start; j < d.n(); j++ {
			if d.inCode(j) || strings.TrimSpace(d.lines[j]) == "" || !strings.Contains(d.lines[j], "|") {
				break
			}
			d.table[j] = true
		}
	}
}

func isTableDelimiter(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.Contains(trimmed, "-") {
		return false
	}
	cells := strings.Split(trimmed, "|")
	if len(cells) < 2 {
		return false
	}
	for _, c := range cells {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if strings.Trim(c, "-:") != "" {
			return false
		}
	}
	return true
}

// parse builds the goldmark AST on first use.
func (d *doc) parse() ast.Node {
	if d.root == nil {
		src := d.src
		d.root = goldmarkParser.Parser().Parse(text.NewReader(src))
	}
	return d.root
}

// lineOf reports the 1-based line a node starts on (0 when unknown). Inline
// nodes carry no line information in goldmark 1.7 and panic if Lines() is
// called on them, so the lookup climbs to the nearest enclosing block.
func (d *doc) lineOf(node ast.Node) int {
	for n := node; n != nil; n = n.Parent() {
		if n.Type() == ast.TypeInline {
			continue
		}
		seg := n.Lines()
		if seg == nil || seg.Len() == 0 {
			continue
		}
		return d.lineAt(seg.At(0).Start)
	}
	return 0
}

// lineAt maps a byte offset to a 1-based line number.
func (d *doc) lineAt(offset int) int {
	if offset < 0 || offset > len(d.src) {
		return 0
	}
	return 1 + bytes.Count(d.src[:offset], []byte("\n"))
}

// walk visits every node in the document, depth-first. Returning false from fn
// skips that node's children. (goldmark 1.7 has no ast.Node.Walk, so the
// traversal is spelled out here.)
func (d *doc) walk(fn func(node ast.Node) bool) {
	var visit func(n ast.Node)
	visit = func(n ast.Node) {
		if n == nil {
			return
		}
		if !fn(n) {
			return
		}
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			visit(child)
		}
	}
	visit(d.parse())
}

// DefaultConfig returns the built-in configuration: every implemented rule on,
// with its default options.
func DefaultConfig() *Config {
	return &Config{Default: true, Rules: map[string]json.RawMessage{}}
}

// LoadConfig reads a config file. A missing file is not an error: the defaults
// apply.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	cfg := DefaultConfig()
	if v, ok := raw["default"]; ok {
		if err := json.Unmarshal(v, &cfg.Default); err != nil {
			return nil, fmt.Errorf("parse %s: 'default' must be a boolean: %w", path, err)
		}
	}
	if _, ok := raw["extends"]; ok {
		return nil, fmt.Errorf("%s uses \"extends\": npm shareable configs cannot be resolved by the Go linter — copy the rules you need into the file", path)
	}
	for k, v := range raw {
		if k == "default" {
			continue
		}
		// $-prefixed keys are documentation for humans ($schema, $comment), not
		// rules. Ignoring them is what markdownlint does, and it keeps
		// explaining a config inside the config.
		if strings.HasPrefix(k, "$") {
			continue
		}
		cfg.Rules[k] = v
	}
	return cfg, nil
}

// LoadBaseline reads a baseline file, rejecting a zero ceiling. A missing file
// returns (nil, nil), meaning "no ceiling".
func LoadBaseline(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if b.MaxViolations <= 0 {
		return nil, fmt.Errorf("%s has max_violations = %d: a zero ceiling cannot express a ratchet — omit the baseline file to require zero violations", path, b.MaxViolations)
	}
	return &b, nil
}

// WriteBaseline records the current total as the new ceiling. It is the ratchet's
// only writer: the ceiling may only move by deliberately re-measuring, and the
// timestamp plus the histogram in the file make that move reviewable.
func WriteBaseline(path string, r *Result) error {
	b := Baseline{
		MaxViolations: r.Total,
		ByRule:        r.ByRule,
		MeasuredAt:    time.Now().UTC().Format(time.RFC3339),
		Note: "Ratchet ceiling re-measured with 'plaesy validate markdown --write-baseline'. " +
			"This is the debt already in the tree, not a target: it may only go down. " +
			"Fix violations (--fix, then by hand) and re-measure to lower it. Never raise it to make a red build green.",
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// Exceeds reports whether the run broke the baseline's ceiling.
//
// A nil baseline is `--no-baseline`, which promises zero violations, so it is
// treated as a ceiling of 0 rather than as the absence of a ceiling. Treating
// it as "nothing to compare" let the flag print its own violations and exit 0.
func (r *Result) Exceeds(b *Baseline) bool {
	if b == nil {
		return r.Total > 0
	}
	return r.Total > b.MaxViolations
}

// Histogram renders a per-rule breakdown, most violations first.
func (r *Result) Histogram() string {
	if len(r.ByRule) == 0 {
		return "no violations"
	}
	type kv struct {
		rule string
		n    int
	}
	pairs := make([]kv, 0, len(r.ByRule))
	for rule, n := range r.ByRule {
		pairs = append(pairs, kv{rule, n})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].n != pairs[j].n {
			return pairs[i].n > pairs[j].n
		}
		return pairs[i].rule < pairs[j].rule
	})
	var b strings.Builder
	for _, p := range pairs {
		fmt.Fprintf(&b, "\n  %-8s %d", p.rule, p.n)
	}
	return strings.TrimPrefix(b.String(), "\n")
}

// WorstFiles returns up to n file paths, most violations first.
func (r *Result) WorstFiles(n int) []string {
	out := make([]string, 0, len(r.Violations))
	for _, f := range r.Violations {
		out = append(out, f.Path)
		if n > 0 && len(out) >= n {
			break
		}
	}
	return out
}

// collectFiles returns the Markdown files to lint, sorted for deterministic output.
func collectFiles(root string, paths []string, ignoreDirs []string) ([]string, error) {
	if len(ignoreDirs) == 0 {
		ignoreDirs = DefaultIgnoreDirs
	}
	skip := map[string]bool{}
	for _, d := range ignoreDirs {
		skip[d] = true
	}
	var out []string
	seen := map[string]bool{}
	add := func(path string) {
		abs, err := filepath.Abs(path)
		if err != nil || seen[abs] {
			return
		}
		seen[abs] = true
		out = append(out, abs)
	}
	walkDir := func(start string, skipRoot bool) error {
		return filepath.WalkDir(start, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if skip[d.Name()] && !(skipRoot && filepath.Clean(path) == filepath.Clean(start)) {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.EqualFold(filepath.Ext(d.Name()), ".md") {
				add(path)
			}
			return nil
		})
	}
	if len(paths) == 0 {
		if err := walkDir(root, true); err != nil {
			return nil, err
		}
	} else {
		for _, p := range paths {
			info, statErr := os.Stat(p)
			if statErr != nil {
				return nil, statErr
			}
			if !info.IsDir() {
				if strings.EqualFold(filepath.Ext(p), ".md") {
					add(p)
				}
				continue
			}
			if err := walkDir(p, false); err != nil {
				return nil, err
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

func relPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}
