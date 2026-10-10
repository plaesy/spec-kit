package quality

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

// TemplateMarkerBaselineName is the file, next to scripts/go.mod, that records
// how many legacy fill-in markers each template may keep, per surface. It sits
// beside coverage-baseline.json for the same reason: the corpus states its
// rules in templates/README.md, and a rule that lives only in prose is a wish.
const TemplateMarkerBaselineName = "template-marker-baseline.json"

// A surface is a region of a template where a fill-in marker can hide. They are
// ratcheted separately because the honest count differs per region, and because
// the prose count is the one that was wrong to begin with: measuring only prose
// reported the corpus as clean while 255 markers sat inside fenced blocks and
// 28 underscore sign-off blanks were invisible to every mechanism in the repo.
const (
	// surfaceProse is everything outside fenced blocks and inline code spans.
	surfaceProse = "prose"
	// surfaceFenced is inside fenced code blocks. Markers here are usually a
	// block that illustrates a format — `[Module]`, `[Number]` — and usually
	// should stay, but not always: a fenced form is still a form, and its
	// slots were being filled by nobody.
	surfaceFenced = "fenced"
	// surfaceUnderscore is a run of 4+ underscores used as a sign-off blank,
	// e.g. `**Owner**: ______ Date: ______`. Three or fewer is a horizontal
	// rule or emphasis and is not counted, which is why the threshold is 4.
	surfaceUnderscore = "underscore"
)

var allSurfaces = []string{surfaceProse, surfaceFenced, surfaceUnderscore}

// Legacy markers are the two deprecated fill-in syntaxes: a [BRACKET] token and
// a {curly} token. Canonical {{PLACEHOLDER}} tokens are not counted — they are
// the target state, and counting them would reward a template for having more
// slots to fill.
//
// A bracket that is part of a markdown link or an image is not a fill-in. The
// regex alone cannot tell: it matches `[Overview]` inside `[Overview](#overview)`
// and stops at the closing bracket, because the `(` that would disqualify it is
// outside the pattern. isLinkOrImage is therefore checked separately, on the
// character that follows the match — `(` for a link or image destination — and
// on a `!` immediately before it.
var (
	bracketMarkerRE = regexp.MustCompile(`\[[A-Za-z0-9_][^\]\n]{0,60}\]`)
	curlyMarkerRE   = regexp.MustCompile(`\{[A-Za-z_][A-Za-z0-9_| ./():-]{0,60}\}`)
	underscoreRE    = regexp.MustCompile(`_{4,}`)
)

// startsCanonicalToken reports whether the match beginning at off is the tail
// of a canonical `{{NAME}}` / `{{NAME|default}}` fill-in rather than a legacy
// {curly} token.
//
// This distinction is the whole reason the ratchet is not a plain regex count.
// curlyMarkerRE cannot match the opening brace of a canonical token, because
// the character after `{` must be a letter, so the match always begins at the
// SECOND brace and off-1 is the opener. A canonical token carrying a prose
// format hint — `{{PERFORMANCE_TARGET|<200ms p95 response time}}` in the
// shipped constitution template — is otherwise matched in full, so counting
// without this check scored that template at its hint text rather than at its
// legacy debt.
//
// The check is on the single preceding character, not on the two before it:
// off is already the second brace, so prose[off-2:off] is a space and a brace,
// and comparing that to "{{" silently rejected every canonical token — which
// made the ratchet report 312 legacy markers in a template with none.
func startsCanonicalToken(prose string, off int) bool {
	if off < 1 || prose[off-1] != '{' {
		return false
	}
	end := off
	for end < len(prose) && prose[end] != '}' {
		end++
	}
	return end < len(prose) && prose[end] == '}'
}

// updateMarkerBaseline re-records every template's count. It is a flag rather
// than a second program for the reason the coverage ratchet uses one: the
// measurement already exists here, and a second implementation of it would be a
// second thing to keep honest.
var updateMarkerBaseline = flag.Bool("update-template-marker-baseline", false,
	"re-measure every template and rewrite "+TemplateMarkerBaselineName+" with the result")

// mustStay is a token that is permanently NOT a fill-in, in one file and one
// surface. It is a struct rather than a bare token list so that the reason
// travels with the entry: an allowance with no stated reason is one nobody
// re-checks, and the list is the part of this file most likely to rot.
type mustStay struct {
	File    string   `json:"file"`
	Surface string   `json:"surface"`
	Tokens  []string `json:"tokens"`
	Why     string   `json:"why"`
}

type markerBaseline struct {
	Comment    []string                  `json:"$comment"`
	MeasuredAt string                    `json:"measured_at"`
	Surfaces   map[string]map[string]int `json:"surfaces"`
	MustStay   []mustStay                `json:"must_stay"`
}

// fenceSpan reports whether each line of src is inside a fenced code block.
func fenceLines(src string) []bool {
	out := make([]bool, 0, strings.Count(src, "\n")+1)
	inFence := false
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			out = append(out, false)
			continue
		}
		out = append(out, inFence)
	}
	return out
}

// blankInlineCode replaces inline code spans with spaces, leaving the prose a
// filler actually reads. Fenced regions are handled separately by surface, so
// this is only applied to non-fenced lines.
func blankInlineCode(line string) string {
	blanked := line
	for _, fence := range []string{"```", "``", "`"} {
		for {
			start := strings.Index(blanked, fence)
			if start < 0 {
				break
			}
			rest := blanked[start+len(fence):]
			end := strings.Index(rest, fence)
			if end < 0 {
				break
			}
			blanked = blanked[:start] + strings.Repeat(" ", end+len(fence)) + rest[end+len(fence):]
		}
	}
	return blanked
}

// countLegacyIn counts deprecated fill-in syntax in one region, discounting
// markdown links, images, canonical tokens, and any token on the stays list.
func countLegacyIn(region, file, surface string, stays []string) int {
	n := 0
	for _, loc := range bracketMarkerRE.FindAllStringIndex(region, -1) {
		if isLinkOrImage(region, loc[0], loc[1]) {
			continue
		}
		n++
	}
	for _, loc := range curlyMarkerRE.FindAllStringIndex(region, -1) {
		if startsCanonicalToken(region, loc[0]) {
			continue
		}
		n++
	}
	// A token on the stays list is part of the file's own grammar, not a slot:
	// `[P]` is the parallel-execution flag in the tasks format, `[Ctrl+C]` is a
	// keypress in a facilitator guide. They are counted, then discounted, so the
	// exemption is visible in the raw number and a file cannot hide real debt
	// behind a broad allowance.
	for _, token := range stays {
		n -= strings.Count(region, token)
	}
	if n < 0 {
		n = 0
	}
	return n
}

func countSurfaces(src, file string, stays func(string) []string) map[string]int {
	fenced := fenceLines(src)
	lines := strings.Split(src, "\n")

	var proseBuf, fencedBuf strings.Builder
	for i, line := range lines {
		if fenced[i] {
			fencedBuf.WriteString(line)
			fencedBuf.WriteString("\n")
			continue
		}
		proseBuf.WriteString(blankInlineCode(line))
		proseBuf.WriteString("\n")
	}

	// Underscore blanks are only counted outside fenced blocks: a run of
	// underscores inside a code sample is a drawing, not a signature line. The
	// bracket/curly scan is deliberately NOT applied to this region — it is a
	// separate surface and counting both syntaxes in one total would make a
	// single number mean two different things.
	nonFenced := make([]string, 0, len(lines))
	for i, line := range lines {
		if !fenced[i] {
			nonFenced = append(nonFenced, blankInlineCode(line))
		}
	}
	underscoreRegion := strings.Join(nonFenced, "\n")
	underscores := len(underscoreRE.FindAllString(underscoreRegion, -1)) -
		discountUnderscores(underscoreRegion, stays(surfaceUnderscore))
	if underscores < 0 {
		underscores = 0
	}

	return map[string]int{
		surfaceProse:      countLegacyIn(proseBuf.String(), file, surfaceProse, stays(surfaceProse)),
		surfaceFenced:     countLegacyIn(fencedBuf.String(), file, surfaceFenced, stays(surfaceFenced)),
		surfaceUnderscore: underscores,
	}
}

// discountUnderscores subtracts the underscore runs an allowance covers. Kept
// separate from countLegacyIn because the token being counted there is a
// bracketed marker and here it is a bare run, which the regex would otherwise
// count once per underscore rather than once per blank.
func discountUnderscores(region string, stays []string) int {
	discounted := 0
	for _, token := range stays {
		for _, run := range underscoreRE.FindAllString(region, -1) {
			if strings.HasPrefix(run, token) {
				discounted++
			}
		}
	}
	return discounted
}

// isLinkOrImage reports whether a bracket match at start..end is really a
// markdown link or image rather than a fill-in.
func isLinkOrImage(prose string, start, end int) bool {
	if end < len(prose) && prose[end] == '(' {
		return true
	}
	return start > 0 && prose[start-1] == '!'
}

// TestTemplateLegacyMarkersDoNotGrow is the CI gate. For every template and
// every surface it counts deprecated fill-in syntax and fails on four
// conditions, each with its own fix:
//
//   - a count above the recorded number — a template is gaining legacy syntax
//     while the corpus is trying to shed it
//   - a file or surface with no entry — a new template is an unexamined default
//   - an entry for a template that no longer exists — a claim about gone code
//   - an entry the tree has outgrown — a bar the tree has passed is a claim
//     about the file as it was, and a too-low bar is the one error nothing else
//     catches
//
// Lowering a number is the point, so an improvement is logged and never fails.
// Pass -update-template-marker-baseline to re-record after a deliberate pass.
func TestTemplateLegacyMarkersDoNotGrow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping the template marker ratchet in -short mode")
	}
	repo := filepath.Dir(moduleDir(t))
	b := loadMarkerBaseline(t, repo)
	measured := measureTemplateMarkers(t, repo, b)

	if len(measured) == 0 {
		t.Fatalf("measured zero templates under %s: the runner is broken, not the corpus clean",
			filepath.Join(repo, "templates"))
	}

	if *updateMarkerBaseline {
		writeMarkerBaseline(t, filepath.Join(moduleDir(t), TemplateMarkerBaselineName), b, measured)
		t.Skipf("rewrote %s from a fresh measurement of %d template(s) across %d surface(s)",
			TemplateMarkerBaselineName, len(measured), len(allSurfaces))
	}

	var overlap int
	for name := range measured {
		if _, ok := b.Surfaces[surfaceProse][name]; ok {
			overlap++
		}
	}
	if overlap == 0 {
		t.Skipf("%s describes a different tree: %d template(s) measured, none listed; "+
			"delete it if this copy will never grow templates", TemplateMarkerBaselineName, len(measured))
	}

	for _, surface := range allSurfaces {
		recorded, ok := b.Surfaces[surface]
		if !ok {
			t.Errorf("%s has no %q section; re-record it so the gate governs that surface:\n\n"+
				"    go test ./internal/quality -run TestTemplateLegacyMarkersDoNotGrow -update-template-marker-baseline",
				TemplateMarkerBaselineName, surface)
			continue
		}
		for _, name := range sortedNames(measured) {
			got := measured[name][surface]
			want, listed := recorded[name]
			switch {
			case !listed:
				t.Errorf("%s surface %q: %s has no entry (%d legacy marker(s))\n\n"+
					"A new template is an unexamined default. Record it with -update-template-marker-baseline, "+
					"then migrate it, or record the count and accept the debt deliberately.",
					TemplateMarkerBaselineName, surface, name, got)
			case got > want:
				t.Errorf("legacy markers grew in %s (surface %q): %d recorded, %d now\n\n"+
					"Use {{PLACEHOLDER}} form, not a new [BRACKET] or {curly}. If this template was "+
					"rewritten deliberately, re-record with -update-template-marker-baseline and review "+
					"the diff: the new numbers are the record, not a target.", name, surface, want, got)
			case got < want:
				t.Logf("legacy markers down in %s (surface %q): %d -> %d, re-record with -update-template-marker-baseline",
					name, surface, want, got)
			}
		}
		for _, name := range sortedRecorded(recorded) {
			if _, ok := measured[name]; !ok {
				t.Errorf("stale entry in %s: %s is recorded under %q but does not exist\n\n"+
					"Remove the entry, or restore the template it describes.",
					TemplateMarkerBaselineName, name, surface)
			}
		}
	}
}

func loadMarkerBaseline(t *testing.T, repo string) markerBaseline {
	t.Helper()
	path := filepath.Join(repo, "scripts", TemplateMarkerBaselineName)
	data, err := os.ReadFile(path)
	if err != nil {
		// Bootstrap only. A ratchet that is being introduced has no file yet,
		// and the flag exists precisely to write the first one.
		if *updateMarkerBaseline && os.IsNotExist(err) {
			return markerBaseline{Surfaces: map[string]map[string]int{}}
		}
		t.Fatalf("read %s: %v\n\nWithout a recorded baseline there is no ratchet, and a bar "+
			"that only exists in prose is a wish. Record one with:\n\n"+
			"    go test ./internal/quality -run TestTemplateLegacyMarkersDoNotGrow -update-template-marker-baseline",
			path, err)
	}
	var b markerBaseline
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if len(b.Surfaces) == 0 && !*updateMarkerBaseline {
		// Skipped while recording: widening the ratchet to a new surface means
		// the file has no section for it yet, and that is precisely the run that
		// writes one. Enforcing it here would make the section uncreatable.
		t.Fatalf("%s records no surfaces: the ratchet would pass on an empty map", path)
	}
	if b.MeasuredAt == "" {
		t.Errorf("%s has no measured_at date: a number with no date cannot be re-measured on purpose", path)
	}
	// A must_stay entry that names no measured template/surface discounts
	// nothing, and a mistyped key is indistinguishable from an allowance that
	// has already been applied — the failure looks exactly like success.
	for _, m := range b.MustStay {
		if strings.Contains(m.File, "/") {
			t.Errorf("%s must_stay file %q is a path; keys are bare filenames", path, m.File)
		}
		if !contains(allSurfaces, m.Surface) {
			t.Errorf("%s must_stay entry for %s names surface %q; known surfaces are %v",
				path, m.File, m.Surface, allSurfaces)
		}
		if m.Why == "" {
			t.Errorf("%s must_stay entry for %s (%s) has no why: an allowance nobody can re-check is "+
				"an allowance nobody will", path, m.File, m.Surface)
		}
	}
	return b
}

func measureTemplateMarkers(t *testing.T, repo string, b markerBaseline) map[string]map[string]int {
	t.Helper()
	dir := filepath.Join(repo, "templates")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	measured := make(map[string]map[string]int)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".template.md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		name := e.Name()
		measured[name] = countSurfaces(string(data), name, func(surface string) []string {
			var out []string
			for _, m := range b.MustStay {
				if m.File == name && m.Surface == surface {
					out = append(out, m.Tokens...)
				}
			}
			return out
		})
	}
	return measured
}

func writeMarkerBaseline(t *testing.T, path string, old markerBaseline, measured map[string]map[string]int) {
	t.Helper()
	next := markerBaseline{
		Comment:    old.Comment,
		MeasuredAt: time.Now().Format("2006-01-02"),
		Surfaces:   make(map[string]map[string]int, len(allSurfaces)),
		MustStay:   old.MustStay,
	}
	for _, surface := range allSurfaces {
		next.Surfaces[surface] = make(map[string]int, len(measured))
		for name, counts := range measured {
			next.Surfaces[surface][name] = counts[surface]
		}
	}
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		t.Fatalf("encode %s: %v", TemplateMarkerBaselineName, err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	t.Logf("recorded %d template(s) across %d surface(s) in %s",
		len(measured), len(allSurfaces), TemplateMarkerBaselineName)
}

func sortedNames(m map[string]map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedRecorded(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
