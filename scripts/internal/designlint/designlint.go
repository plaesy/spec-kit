// Package designlint checks produced frontend output (HTML, CSS, JSX and similar)
// against the rules in instructions/frontend-design.instructions.md and
// instructions/redesign.instructions.md.
//
// Every check is deterministic and offline: it reads text and never renders a
// page. Checks that need a browser (headline line count, overlap, empty screens)
// are out of scope. Their evidence is the screenshots taken at each required
// width, which this package checks only for presence.
package designlint

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
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
)

// RequiredWidths are the viewport widths the render procedure must cover.
var RequiredWidths = []int{1440, 768, 390}

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

var skipDirs = map[string]bool{
	"node_modules": true, ".git": true, "dist": true, "build": true, ".next": true,
}

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
	reHTMLNoise   = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>`)
	reTag         = regexp.MustCompile(`<[^>]*>`)
	reFigure      = regexp.MustCompile(`99\.99%|\$100\.00|\b50%`)
	reFontFamily  = regexp.MustCompile(`(?i)font-family\s*:\s*([^;}\n]+)`)
	reFontFamilyJ = regexp.MustCompile(`fontFamily\s*:\s*['"]([^'"]+)['"]`)
	reTokenDecl   = regexp.MustCompile(`--([A-Za-z0-9_-]+)\s*:\s*([^;}{]+)`)
	reBlock       = regexp.MustCompile(`([^{}]+)\{([^{}]*)\}`)
	reCSSComment  = regexp.MustCompile(`(?s)/\*.*?\*/`)
)

// Lint runs every check over paths. A path may be a file or a directory; a
// directory is walked for the extensions this package reads.
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

	tokens := collectTokens(contents)
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
	return res, nil
}

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
				if path != p && skipDirs[d.Name()] {
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
	return out, nil
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

// collectTokens reads every `--name: value` declaration that resolves to a colour.
// A token is recorded under its bare name, so `var(--ink)` finds `--ink: #111`.
func collectTokens(contents map[string]string) map[string]rgb {
	tokens := map[string]rgb{}
	for _, text := range contents {
		for _, m := range reTokenDecl.FindAllStringSubmatch(text, -1) {
			if c, ok := parseColor(m[2]); ok {
				tokens[m[1]] = c
			}
		}
	}
	return tokens
}

// lintContrast checks each CSS rule that sets both a text colour and a background
// colour, both resolvable to a concrete value. A rule that depends on a gradient,
// an image, an inherited value or a parent's background is skipped and counted,
// so the report says how much was not checked.
func lintContrast(res *Result, file, text string, tokens map[string]rgb) {
	clean := reCSSComment.ReplaceAllStringFunc(text, func(m string) string {
		return strings.Repeat(" ", len(m))
	})
	for _, loc := range reBlock.FindAllStringSubmatchIndex(clean, -1) {
		body := clean[loc[4]:loc[5]]
		var fgVal, bgVal string
		var size float64
		var weight int
		for _, decl := range strings.Split(body, ";") {
			i := strings.Index(decl, ":")
			if i < 0 {
				continue
			}
			prop := strings.ToLower(strings.TrimSpace(decl[:i]))
			val := strings.TrimSpace(decl[i+1:])
			switch prop {
			case "color":
				fgVal = val
			case "background", "background-color":
				bgVal = val
			case "font-size":
				size = parseSize(val)
			case "font-weight":
				weight = parseWeight(val)
			}
		}
		if fgVal == "" || bgVal == "" {
			continue
		}
		fg, okF := resolveColor(fgVal, tokens)
		bg, okB := resolveColor(bgVal, tokens)
		if !okF || !okB {
			res.ContrastSkipped++
			continue
		}
		res.ContrastChecked++
		ratio := contrastRatio(fg, bg)
		need, kind := 4.5, "normal text"
		if size >= 24 || (size >= 18.66 && weight >= 700) {
			need, kind = 3.0, "large text"
		}
		if ratio < need {
			res.Findings = append(res.Findings, Finding{
				File: file, Line: lineOf(clean, loc[0]), Rule: RuleContrast, Severity: SevFail,
				Message: fmt.Sprintf("%s on %s is %.2f:1; WCAG 2.2 AA needs %.1f:1 for %s", fg.hex(), bg.hex(), ratio, need, kind),
			})
		}
	}
}

// resolveColor returns the colour a value stands for: a literal, or a token from
// var(--name). Gradients, images and keywords do not resolve.
func resolveColor(val string, tokens map[string]rgb) (rgb, bool) {
	lower := strings.ToLower(val)
	if strings.Contains(lower, "gradient") || strings.Contains(lower, "url(") {
		return rgb{}, false
	}
	if strings.HasPrefix(lower, "var(") {
		name := strings.TrimPrefix(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(val[4:]), ")")), "--")
		if i := strings.Index(name, ","); i >= 0 {
			name = strings.TrimSpace(name[:i])
		}
		c, ok := tokens[name]
		return c, ok
	}
	return parseColor(val)
}

func parseSize(val string) float64 {
	val = strings.ToLower(strings.TrimSpace(val))
	mult := 1.0
	switch {
	case strings.HasSuffix(val, "px"):
		val = strings.TrimSuffix(val, "px")
	case strings.HasSuffix(val, "rem"):
		val = strings.TrimSuffix(val, "rem")
		mult = 16
	default:
		return 0
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
	if err != nil {
		return 0
	}
	return f * mult
}

func parseWeight(val string) int {
	val = strings.ToLower(strings.TrimSpace(val))
	if val == "bold" {
		return 700
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return n
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
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && imageExts[strings.ToLower(filepath.Ext(path))] {
			names = append(names, filepath.Base(path))
		}
		return nil
	})
	for _, w := range RequiredWidths {
		key := strconv.Itoa(w)
		found := false
		for _, n := range names {
			if strings.Contains(n, key) {
				found = true
				break
			}
		}
		if !found {
			res.Findings = append(res.Findings, Finding{
				File: dir, Rule: RuleViewports, Severity: SevFail,
				Message: fmt.Sprintf("no screenshot at %d px in the evidence directory (file name must contain %q)", w, key),
			})
		}
	}
}
