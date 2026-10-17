// Package designlint checks produced frontend output (HTML, CSS, JSX and similar)
// against the rules in instructions/frontend-design.instructions.md and
// instructions/redesign.instructions.md.
//
// Every check is deterministic and offline: it reads text and never renders a
// page. The same input always produces the same findings in the same order.
// Checks that need a browser (headline line count, overlap, empty screens) are
// out of scope. Their evidence is the screenshots taken at each required width,
// which this package checks only for presence.
//
// Colour checks cover the declarations this package can read: CSS rules, inline
// style attributes, and JSX style objects. Utility classes (Tailwind and the
// like), CSS-in-JS helpers and Vue :style bindings are not read. Each text colour
// the lint cannot resolve is counted in ContrastSkipped, so the report says how
// much was not checked.
package designlint

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Rule identifiers. They are stable: fixtures and evals refer to them by name.
const (
	RuleImageHost = "DL-IMG-HOST"
	RuleFigure    = "DL-INVENTED-FIGURE"
	RuleFont      = "DL-DEFAULT-FONT"
	RuleContrast  = "DL-CONTRAST"
	RuleViewports = "DL-VIEWPORTS"
	RuleToken     = "DL-TOKEN"
)

// RequiredWidths are the viewport widths the render procedure must cover.
var RequiredWidths = []int{1440, 768, 390}

// rootPx is the browser default font size. em and % are resolved against it.
// That assumes the element inherits the root size, which the message says.
const rootPx = 16

// Severity separates findings that fail a check from findings to review.
type Severity string

const (
	SevFail Severity = "FAIL"
	SevWarn Severity = "WARN"
	SevInfo Severity = "INFO"
)

// Finding is one violation, or one note about something that could not be checked.
type Finding struct {
	File     string
	Line     int
	Rule     string
	Severity Severity
	Message  string
}

// Options configures a run.
type Options struct {
	// EvidenceDir holds screenshots taken at each required width. Empty means no
	// render evidence was supplied, and viewport coverage is reported as INFO.
	EvidenceDir string
}

// Result is the outcome of one run over a set of paths.
type Result struct {
	Files           int
	ContrastChecked int
	ContrastSkipped int
	Findings        []Finding
}

// Failed reports whether any finding has FAIL severity.
func (r Result) Failed() bool {
	for _, f := range r.Findings {
		if f.Severity == SevFail {
			return true
		}
	}
	return false
}

// Count returns how many findings carry the given rule and severity.
func (r Result) Count(rule string, sev Severity) int {
	n := 0
	for _, f := range r.Findings {
		if f.Rule == rule && f.Severity == sev {
			n++
		}
	}
	return n
}

var lintedExts = map[string]bool{
	".html": true, ".htm": true, ".css": true, ".scss": true,
	".jsx": true, ".tsx": true, ".vue": true, ".svelte": true,
}

// skipAnywhere are directories that never hold source worth linting.
var skipAnywhere = map[string]bool{".git": true, "node_modules": true}

// skipTopLevel are build outputs. They are skipped only when they sit directly
// under the target, so a source folder named build/ under src/ is still linted.
var skipTopLevel = map[string]bool{"dist": true, "build": true, ".next": true}

var imageExts = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true}

// bannedImageHosts are placeholder services. Their pictures are unlicensed or
// unknown, and the redesign rules forbid them. See redesign.instructions.md C-11.
var bannedImageHosts = []string{
	"picsum.photos", "placehold.co", "placeholder.com", "dummyimage.com",
	"placekitten.com", "placeimg.com", "loremflickr.com", "source.unsplash.com",
}

// bannedDefaultFonts are the browser and default-library families that the
// redesign rules name as defaults to replace (T-01).
var bannedDefaultFonts = map[string]bool{
	"inter": true, "arial": true, "roboto": true, "helvetica": true,
	"system-ui": true, "-apple-system": true, "segoe ui": true,
}

var (
	reHTMLNoise     = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>`)
	reTag           = regexp.MustCompile(`<[^>]*>`)
	reFigure        = regexp.MustCompile(`99\.99%|\$100\.00|\b50%`)
	reFontFamily    = regexp.MustCompile(`(?i)font-family\s*:\s*([^;}\n]+)`)
	reFontFamilyJ   = regexp.MustCompile(`fontFamily\s*:\s*['"]([^'"]+)['"]`)
	reTokenDecl     = regexp.MustCompile(`--([A-Za-z0-9_-]+)\s*:\s*([^;}{]+)`)
	reBlock         = regexp.MustCompile(`([^{}]+)\{([^{}]*)\}`)
	reCSSComment    = regexp.MustCompile(`(?s)/\*.*?\*/`)
	reInlineStyle   = regexp.MustCompile(`(?i)\bstyle\s*=\s*(?:"([^"]*)"|'([^']*)')`)
	reJSXStyle      = regexp.MustCompile(`\bstyle\s*=\s*\{\{([^{}]*)\}\}`)
	reStyleBrace    = regexp.MustCompile(`\bstyle\s*=\s*\{`)
	reVueStyle      = regexp.MustCompile(`:style\s*=`)
	reSizeToken     = regexp.MustCompile(`^([0-9]*\.?[0-9]+)(px|rem|em|%)?$`)
	reSizeInShort   = regexp.MustCompile(`(?:^|\s)([0-9]*\.?[0-9]+(?:px|rem|em))(?:/|\s|$)`)
	reWeightInShort = regexp.MustCompile(`(?:^|\s)([1-9]00|bold|normal)(?:\s|$)`)
)

// Lint runs every check over paths. A path may be a file or a directory; a
// directory is walked for the extensions this package reads. The result does not
// depend on the order of paths or on map iteration.
func Lint(paths []string, opts Options) (Result, error) {
	var res Result
	files, err := collectFiles(paths)
	if err != nil {
		return res, err
	}
	contents := make(map[string]string, len(files))
	for _, f := range files {
		b, readErr := os.ReadFile(f)
		if readErr != nil {
			return res, readErr
		}
		contents[f] = string(b)
	}
	res.Files = len(files)

	tokens := collectTokens(files, contents)
	for _, name := range sortedKeys(tokens.conflicts) {
		res.Findings = append(res.Findings, Finding{
			File: "(tokens)", Rule: RuleToken, Severity: SevWarn,
			Message: fmt.Sprintf("--%s is declared with %d different values across the files; any contrast check that uses it is skipped", name, tokens.conflicts[name]),
		})
	}
	for _, f := range files {
		text := contents[f]
		lintImageHosts(&res, f, text)
		if isHTML(f) {
			lintFigures(&res, f, text)
		}
		lintFonts(&res, f, text)
		lintContrast(&res, f, text, tokens)
	}
	lintViewports(&res, opts.EvidenceDir)
	sort.SliceStable(res.Findings, func(i, j int) bool {
		a, b := res.Findings[i], res.Findings[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Rule < b.Rule
	})
	return res, nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// collectFiles returns the lintable files under paths, sorted, so every later
// step sees the same order whatever order the paths were given in.
func collectFiles(paths []string) ([]string, error) {
	var out []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			out = append(out, p)
			continue
		}
		walkErr := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path == p {
					return nil
				}
				if skipAnywhere[d.Name()] || (skipTopLevel[d.Name()] && filepath.Dir(path) == p) {
					return filepath.SkipDir
				}
				return nil
			}
			if lintedExts[strings.ToLower(filepath.Ext(path))] {
				out = append(out, path)
			}
			return nil
		})
		if walkErr != nil {
			return nil, walkErr
		}
	}
	sort.Strings(out)
	return sortedUnique(out), nil
}

func sortedUnique(in []string) []string {
	out := in[:0]
	for i, s := range in {
		if i == 0 || s != in[i-1] {
			out = append(out, s)
		}
	}
	return out
}

func isHTML(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".html" || ext == ".htm"
}

// lineOf returns the 1-based line that contains byte offset idx.
func lineOf(text string, idx int) int {
	return strings.Count(text[:idx], "\n") + 1
}

func lintImageHosts(res *Result, file, text string) {
	lower := strings.ToLower(text)
	for _, host := range bannedImageHosts {
		for from := 0; ; {
			i := strings.Index(lower[from:], host)
			if i < 0 {
				break
			}
			at := from + i
			res.Findings = append(res.Findings, Finding{
				File: file, Line: lineOf(text, at), Rule: RuleImageHost, Severity: SevFail,
				Message: fmt.Sprintf("picture from placeholder host %q: use licence-checked imagery with a recorded credit, or a flat colour field", host),
			})
			from = at + len(host)
		}
	}
}

// lintFigures flags round, placeholder-looking figures in visible HTML text. The
// redesign rules forbid inventing statistics, so each hit is a WARN to review,
// not a FAIL: a real brief can state 50%.
func lintFigures(res *Result, file, text string) {
	visible := reHTMLNoise.ReplaceAllStringFunc(text, func(m string) string {
		return strings.Repeat("\n", strings.Count(m, "\n"))
	})
	visible = reTag.ReplaceAllStringFunc(visible, func(m string) string {
		return strings.Repeat(" ", len(m))
	})
	for _, loc := range reFigure.FindAllStringIndex(visible, -1) {
		res.Findings = append(res.Findings, Finding{
			File: file, Line: lineOf(visible, loc[0]), Rule: RuleFigure, Severity: SevWarn,
			Message: fmt.Sprintf("figure %q looks placeholder-like: confirm it comes from the brief, or mark it as a sample for the client to confirm", visible[loc[0]:loc[1]]),
		})
	}
}

func lintFonts(res *Result, file, text string) {
	check := func(re *regexp.Regexp) {
		for _, loc := range re.FindAllStringSubmatchIndex(text, -1) {
			first := strings.TrimSpace(strings.Split(text[loc[2]:loc[3]], ",")[0])
			first = strings.Trim(first, `"' `)
			if bannedDefaultFonts[strings.ToLower(first)] {
				res.Findings = append(res.Findings, Finding{
					File: file, Line: lineOf(text, loc[0]), Rule: RuleFont, Severity: SevWarn,
					Message: fmt.Sprintf("default font %q is first in the stack: choose a face with character (for example Geist or Outfit) or justify it in the Keystone reason", first),
				})
			}
		}
	}
	check(reFontFamily)
	check(reFontFamilyJ)
}

// tokenTable holds the colour tokens declared across the run. A token declared with
// more than one value (for example a light and a dark theme both defining --ink) is
// a conflict. It is not resolved, so no contrast result depends on which value wins.
type tokenTable struct {
	colours   map[string]colour
	conflicts map[string]int
}

// collectTokens reads every `--name: value` declaration that resolves to a colour.
// Files are visited in sorted order, so the table is the same on every run.
func collectTokens(files []string, contents map[string]string) tokenTable {
	seen := map[string]map[string]bool{}
	colours := map[string]colour{}
	for _, f := range files {
		for _, m := range reTokenDecl.FindAllStringSubmatch(contents[f], -1) {
			c, ok := parseColor(m[2])
			if !ok {
				continue
			}
			if seen[m[1]] == nil {
				seen[m[1]] = map[string]bool{}
				colours[m[1]] = c
			}
			seen[m[1]][c.key()] = true
		}
	}
	t := tokenTable{colours: colours, conflicts: map[string]int{}}
	for name, values := range seen {
		if len(values) > 1 {
			t.conflicts[name] = len(values)
			delete(t.colours, name)
		}
	}
	return t
}

// resolve returns the opaque or translucent colour a value stands for: a literal,
// or a token from var(--name). Gradients, images and keywords do not resolve.
func (t tokenTable) resolve(val string) (colour, bool) {
	lower := strings.ToLower(val)
	if strings.Contains(lower, "gradient") || strings.Contains(lower, "url(") {
		return colour{}, false
	}
	if strings.HasPrefix(lower, "var(") {
		name := strings.TrimPrefix(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(val[4:]), ")")), "--")
		if i := strings.Index(name, ","); i >= 0 {
			name = strings.TrimSpace(name[:i])
		}
		c, ok := t.colours[name]
		return c, ok
	}
	return parseColor(val)
}

// splitTop splits s at sep wherever sep is outside parentheses and quotes.
func splitTop(s string, sep byte) []string {
	var parts []string
	depth, start := 0, 0
	var quote byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '(':
			depth++
		case c == ')':
			depth--
		case c == sep && depth == 0:
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	return append(parts, s[start:])
}

// cssDecls reads `prop: value; prop: value` into a map with lower-case property names.
func cssDecls(body string) map[string]string {
	props := map[string]string{}
	for _, decl := range splitTop(body, ';') {
		i := strings.Index(decl, ":")
		if i < 0 {
			continue
		}
		prop := strings.ToLower(strings.TrimSpace(decl[:i]))
		if prop != "" {
			props[prop] = strings.TrimSpace(decl[i+1:])
		}
	}
	return props
}

// jsxDecls reads a React style object body (`color: '#777', backgroundColor: '#fff'`)
// into CSS property names. A bare number for a length property means pixels.
func jsxDecls(body string) map[string]string {
	props := map[string]string{}
	for _, pair := range splitTop(body, ',') {
		i := strings.Index(pair, ":")
		if i < 0 {
			continue
		}
		key := camelToKebab(strings.TrimSpace(pair[:i]))
		val := strings.Trim(strings.TrimSpace(pair[i+1:]), `'"`)
		if key == "font-size" {
			if _, err := strconv.ParseFloat(val, 64); err == nil {
				val += "px"
			}
		}
		if key != "" {
			props[key] = val
		}
	}
	return props
}

func camelToKebab(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			b.WriteByte('-')
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// lintContrast checks every rule and inline style it can read. A rule is checked
// when it sets a text colour and a background colour, both resolvable, and the
// size and weight are known. Anything else is counted as skipped, with no verdict.
func lintContrast(res *Result, file, text string, tokens tokenTable) {
	clean := reCSSComment.ReplaceAllStringFunc(text, func(m string) string {
		return strings.Repeat(" ", len(m))
	})
	for _, loc := range reBlock.FindAllStringSubmatchIndex(clean, -1) {
		checkRule(res, file, lineOf(clean, loc[0]), cssDecls(clean[loc[4]:loc[5]]), tokens)
	}
	for _, loc := range reInlineStyle.FindAllStringSubmatchIndex(text, -1) {
		body := ""
		if loc[2] >= 0 {
			body = text[loc[2]:loc[3]]
		} else {
			body = text[loc[4]:loc[5]]
		}
		checkRule(res, file, lineOf(text, loc[0]), cssDecls(body), tokens)
	}
	literal := map[int]bool{}
	for _, loc := range reJSXStyle.FindAllStringSubmatchIndex(text, -1) {
		literal[loc[0]] = true
		checkRule(res, file, lineOf(text, loc[0]), jsxDecls(text[loc[2]:loc[3]]), tokens)
	}
	// style={...} that is not a literal object (a variable, a function call) cannot be read.
	for _, loc := range reStyleBrace.FindAllStringIndex(text, -1) {
		if !literal[loc[0]] {
			res.ContrastSkipped++
		}
	}
	res.ContrastSkipped += len(reVueStyle.FindAllStringIndex(text, -1))
}

// checkRule applies WCAG 2.2 AA to one set of declarations, or counts it as skipped.
func checkRule(res *Result, file string, line int, props map[string]string, tokens tokenTable) {
	fgVal, hasFG := props["color"]
	if !hasFG {
		return
	}
	bgVal, hasBG := props["background-color"]
	if !hasBG {
		bgVal, hasBG = props["background"]
	}
	if !hasBG {
		res.ContrastSkipped++ // the background comes from elsewhere in the cascade
		return
	}
	fg, okF := tokens.resolve(fgVal)
	bg, okB := tokens.resolve(bgVal)
	if !okF || !okB || bg.a < 1 {
		res.ContrastSkipped++
		return
	}
	size, weight, ok, basis := fontMetrics(props)
	if !ok {
		res.ContrastSkipped++
		return
	}
	if fg.a < 1 {
		// Translucent text: the reader sees it composited over the background.
		basis += fmt.Sprintf(" (text is %.0f%% opaque, measured as composited over %s)", fg.a*100, bg.hex())
		fg = colour{composite(fg, bg), 1}
	}
	res.ContrastChecked++
	ratio := contrastRatio(fg.c, bg.c)
	need, kind := 4.5, "normal text"
	if size >= 24 || (size >= 18.66 && weight >= 700) {
		need, kind = 3.0, "large text"
	}
	if ratio < need {
		res.Findings = append(res.Findings, Finding{
			File: file, Line: line, Rule: RuleContrast, Severity: SevFail,
			Message: fmt.Sprintf("%s on %s is %.2f:1; WCAG 2.2 AA needs %.1f:1 for %s%s", fg.hex(), bg.hex(), ratio, need, kind, basis),
		})
	}
}

// fontMetrics returns the font size in px and the weight the text is set in. ok is
// false when either cannot be read, so the caller skips the rule rather than guess.
// basis explains an assumption (em and % taken against the root size), for messages.
func fontMetrics(props map[string]string) (size float64, weight int, ok bool, basis string) {
	size, weight = rootPx, 400
	shorthand := props["font"]
	if v, has := props["font-size"]; has {
		s, sok, rel := parseSize(v)
		if !sok {
			return 0, 0, false, ""
		}
		size = s
		if rel {
			basis = fmt.Sprintf(" (size taken as %.0fpx from %q, assuming a %dpx root)", s, v, rootPx)
		}
	} else if shorthand != "" {
		m := reSizeInShort.FindStringSubmatch(strings.ToLower(shorthand))
		if m == nil {
			return 0, 0, false, ""
		}
		s, sok, _ := parseSize(m[1])
		if !sok {
			return 0, 0, false, ""
		}
		size = s
	}
	if v, has := props["font-weight"]; has {
		w, wok := parseWeight(v)
		if !wok {
			return 0, 0, false, ""
		}
		weight = w
	} else if shorthand != "" {
		low := strings.ToLower(shorthand)
		if strings.Contains(low, "bolder") || strings.Contains(low, "lighter") {
			return 0, 0, false, "" // relative to the parent, not readable here
		}
		if m := reWeightInShort.FindStringSubmatch(low); m != nil {
			w, _ := parseWeight(m[1])
			weight = w
		}
	}
	return size, weight, true, basis
}

// parseSize reads a CSS length as px. rel is true when the value was relative to the
// root size. clamp() resolves to the smallest of its arguments, so the strictest
// threshold applies, never the looser one.
func parseSize(val string) (px float64, ok, rel bool) {
	val = strings.ToLower(strings.TrimSpace(val))
	if strings.HasPrefix(val, "clamp(") && strings.HasSuffix(val, ")") {
		// The rendered size lies between the minimum and the maximum, so its smallest
		// possible value is the minimum. The preferred value may be any expression
		// (2vw, calc()), so it is not read, and the bounds decide the result.
		args := splitTop(val[len("clamp("):len(val)-1], ',')
		if len(args) != 3 {
			return 0, false, false
		}
		lo, lok, lrel := parseSize(args[0])
		if !lok {
			return 0, false, false
		}
		best, rel := lo, lrel
		if hi, hok, hrel := parseSize(args[2]); hok {
			if hi < best {
				best = hi
			}
			rel = rel || hrel
		}
		return best, true, rel
	}
	m := reSizeToken.FindStringSubmatch(val)
	if m == nil {
		return 0, false, false
	}
	n, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false, false
	}
	switch m[2] {
	case "px":
		return n, true, false
	case "rem", "em":
		return n * rootPx, true, true
	case "%":
		return n * rootPx / 100, true, true
	case "":
		if n == 0 {
			return 0, true, false
		}
	}
	return 0, false, false
}

func parseWeight(val string) (int, bool) {
	val = strings.ToLower(strings.TrimSpace(val))
	switch val {
	case "normal":
		return 400, true
	case "bold":
		return 700, true
	}
	n, err := strconv.Atoi(val)
	if err != nil || n < 100 || n > 900 || n%100 != 0 {
		return 0, false
	}
	return n, true
}

// viewportWidth reports whether a file name carries the width as its own number, so
// that "1768.png" does not count as 768 px and "3900.png" does not count as 390 px.
func viewportWidth(name string, width int) bool {
	re := regexp.MustCompile(`(?:^|[^0-9])` + strconv.Itoa(width) + `(?:[^0-9]|$)`)
	return re.MatchString(name)
}

func lintViewports(res *Result, dir string) {
	if dir == "" {
		res.Findings = append(res.Findings, Finding{
			File: "(no evidence)", Rule: RuleViewports, Severity: SevInfo,
			Message: "no render evidence supplied: viewport coverage (1440, 768, 390) was not checked",
		})
		return
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		res.Findings = append(res.Findings, Finding{
			File: dir, Rule: RuleViewports, Severity: SevFail,
			Message: "evidence directory does not exist: renders at 1440, 768 and 390 px are required",
		})
		return
	}
	var names []string
	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && imageExts[strings.ToLower(filepath.Ext(path))] {
			names = append(names, filepath.Base(path))
		}
		return nil
	})
	if walkErr != nil {
		res.Findings = append(res.Findings, Finding{
			File: dir, Rule: RuleViewports, Severity: SevFail,
			Message: fmt.Sprintf("evidence directory cannot be read: %v", walkErr),
		})
		return
	}
	for _, w := range RequiredWidths {
		found := false
		for _, n := range names {
			if viewportWidth(n, w) {
				found = true
				break
			}
		}
		if !found {
			res.Findings = append(res.Findings, Finding{
				File: dir, Rule: RuleViewports, Severity: SevFail,
				Message: fmt.Sprintf("no screenshot at %d px in the evidence directory (file name must contain %d as its own number)", w, w),
			})
		}
	}
}
