package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Placeholder syntax is the one convention in this corpus that is both declared
// binding and unenforced. `templates/README.md` states the rule — fill-ins are
// `{{UPPER_SNAKE}}`, optionally `{{UPPER_SNAKE|default}}` — and then 48 of 69
// templates carried no `{{ }}` at all. A convention that loses 4-to-1 without a
// check does not survive a refactor, so the rule is checked here.
//
// What is deliberately NOT checked: legacy syntax anywhere outside a skeleton.
// A prior pass counted `{lower_snake}` at 230, `<angle_bracket>` at 93 and
// `[UPPER_SNAKE]` at 23 and reported all three as "invisible to validators"
// debt. Those numbers were eyeballed, and two of the three syntaxes are not
// debt at all:
//
//	docs/reference.md:122   <section> | <mapping_type>   documenting mapping.json's shape
//	prompts/create/diagram.md:32  <output-dir> <slugified-title> <format>   a usage line
//
// Those are documentation metavariables — the conventional way to write a
// parameter slot. Rewriting them to `{{SECTION}}` would make the docs
// unreadable and the CLI examples unrunnable. So the distinction this test
// enforces is not spelling, it is *role*: a skeleton is a file meant to be
// filled in, and only a skeleton may use a fill-in. Everywhere else, angle
// brackets and braces are prose.
//
// Skeletons are templates/*.template.md, plus the .plaesy/ copies that ship
// them. A legacy fill-in in a skeleton is a defect. A legacy fill-in in
// documentation is not, and is counted and logged so the debt stays measured
// rather than guessed.

var (
	// The three legacy forms. Kept narrow on purpose: `[A-Z][A-Z0-9_]*`
	// would match every Go array index and every bracketed aside, and
	// `<[a-z][a-z-]*>` would match every HTML tag.
	// `{[a-z][a-z0-9_]*}` missed `{tasks-topics}`, which is why
	// templates/context.template.md shipped a legacy fill-in next to a
	// canonical `{{STATUS}}` in the same path and the gate stayed green.
	// Real placeholders carry hyphens and dots, so the class does.
	legacyLowerSnake = regexp.MustCompile(`\{[a-z][a-z0-9_.-]*\}`)
	legacyAngle      = regexp.MustCompile(`<[a-z][a-z0-9_-]*>`)
	legacyUpperSnake = regexp.MustCompile(`\[[A-Z][A-Z0-9_]+\]`)
	// The default form {{NAME|some default}} has to match too. This pattern
	// was only ever used to *skip* canonical fill-ins when hunting legacy ones,
	// so failing to match the default form cost nothing. The moment it was
	// used as a positive assertion -- does this skeleton have any fill-in at
	// all -- the gap surfaced as memory.template.md reported as having none,
	// while its own line 10 reads {{UPDATED_AT|2026-01-31}}.
	canonicalFill = regexp.MustCompile(`\{\{[A-Z0-9_]+(\|[^}]*)?\}\}`)
	// codeSpan is the one already declared in docs_drift_test.go, reused
	// rather than redeclared: it matches a single-line span, which is the
	// granularity that matters here.
	phFenceLine = regexp.MustCompile("^" + `\s*` + "```")
)

// stripNonProse blanks out regions where a placeholder-shaped token is not a
// fill-in: fenced code blocks and inline code spans. `<dir>` in a usage line
// that happens to be set in backticks is a metavariable; `<name>` in a
// template's fill-in table is a defect. Blanking by byte offset keeps the
// remaining text aligned so reported line numbers stay true.
func stripNonProse(body string) string {
	out := []byte(body)
	blank := func(a, b int) {
		for i := a; i < b && i < len(out); i++ {
			if out[i] != '\n' {
				out[i] = ' '
			}
		}
	}
	inFence := false
	for i, line := 0, 0; i < len(body); i, line = endOfLine(body, i), line+1 {
		if phFenceLine.MatchString(body[i:endOfLine(body, i)]) {
			inFence = !inFence
			blank(i, endOfLine(body, i))
			continue
		}
		if inFence {
			blank(i, endOfLine(body, i))
			continue
		}
		for _, m := range codeSpan.FindAllStringIndex(body[i:endOfLine(body, i)], -1) {
			blank(i+m[0], i+m[1])
		}
	}
	return string(out)
}

// startsCanonical reports whether the token at off opens a canonical
// `{{UPPER_SNAKE}}` / `{{UPPER_SNAKE|default}}` fill-in.
func startsCanonical(prose string, off int) bool {
	if off < 2 || prose[off-2:off] != "{{" {
		return false
	}
	end := off
	for end < len(prose) && prose[end] != '}' {
		end++
	}
	return end < len(prose) && canonicalFill.MatchString(prose[off-1:end+1])
}

func endOfLine(s string, i int) int {
	if j := strings.IndexByte(s[i:], '\n'); j >= 0 {
		return i + j + 1
	}
	return len(s)
}

func lineOf(body string, offset int) int {
	return strings.Count(body[:offset], "\n") + 1
}

type legacyHit struct {
	rule, file string
	line       int
	text       string
}

func findLegacy(body, file string) []legacyHit {
	var hits []legacyHit
	prose := stripNonProse(body)
	rules := []struct {
		name string
		re   *regexp.Regexp
	}{
		{"lower_snake", legacyLowerSnake},
		{"angle", legacyAngle},
		{"UPPER_SNAKE", legacyUpperSnake},
	}
	for _, r := range rules {
		for _, m := range r.re.FindAllStringIndex(prose, -1) {
			// `{{NAME}}` and `{{NAME|default}}` are canonical. The
			// lower_snake rule would match the braces of a canonical
			// fill-in whose default happens to be lowercase, so the
			// `{{` opener is what has to be excluded.
			if startsCanonical(prose, m[0]) {
				continue
			}
			// `[FAQ](#frequently-asked-questions)` is a markdown link
			// label in a table of contents, not a fill-in. Only the
			// bracket form needs this; a `<` or `{` opener is not
			// how markdown spells a link.
			if r.name == "UPPER_SNAKE" && m[1] < len(prose) && prose[m[1]] == '(' {
				continue
			}
			hits = append(hits, legacyHit{r.name, file, lineOf(body, m[0]), prose[m[0]:m[1]]})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].line != hits[j].line {
			return hits[i].line < hits[j].line
		}
		return hits[i].rule < hits[j].rule
	})
	return hits
}

// isSkeleton reports whether a path is a fill-in template rather than prose.
func isSkeleton(rel string) bool {
	rel = filepath.ToSlash(rel)
	if !strings.HasSuffix(rel, ".template.md") {
		return false
	}
	// templates/README.md is the taxonomy that *documents* the four syntaxes;
	// it quotes them on purpose, so it is the one skeleton excluded.
	return rel != "templates/README.md"
}

func TestTemplateSkeletonsUseOnlyCanonicalPlaceholders(t *testing.T) {
	docs := repoDocs(t)
	if docs == "" {
		t.Skip("not a repository checkout")
	}
	root := filepath.Dir(filepath.Dir(docs))

	roots := []string{
		filepath.Join(root, "templates"),
		filepath.Join(root, ".plaesy", "templates"),
	}
	type found struct {
		rel  string
		hits []legacyHit
	}
	var all []found
	skeletons := 0
	for _, dir := range roots {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".template.md") {
				return nil
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			rel = filepath.ToSlash(rel)
			if !isSkeleton(rel) {
				return nil
			}
			skeletons++
			body, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			if h := findLegacy(string(body), rel); len(h) > 0 {
				all = append(all, found{rel, h})
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}
	if skeletons == 0 {
		t.Fatal("no skeletons found; the walk is wrong, which would make this test vacuous")
	}

	for _, f := range all {
		for _, h := range f.hits {
			t.Errorf("%s:%d: %s fill-in %q in a skeleton. Skeletons are filled in by a human or an "+
				"agent reading the file literally, so the form must be the canonical one: "+
				"{{UPPER_SNAKE}} or {{UPPER_SNAKE|default}}.", h.file, h.line, h.rule, h.text)
		}
	}
	t.Logf("%d skeleton(s) checked, %d with legacy fill-ins", skeletons, len(all))
}

// TestEverySkeletonHasAtLeastOneCanonicalFillIn is the positive half of the same
// rule, and it is the half a regex cannot defeat.
//
// The three patterns above all require a space-free body, so a legacy fill-in
// containing a space, a digit or a slash walks straight through them:
// `[APP NAME]`, `[Current Version]`, `{tasks/{id}}`. The gate was green on the
// corpus's five largest skeletons while they contained no canonical fill-in at
// all — 1009 lines, zero `{{ }}` — because there was nothing for it to flag.
//
// Asserting the absence of a fill-in is only as good as the pattern that looks
// for one. Asserting that a file named `.template.md` has at least one
// canonical fill-in needs no pattern: a template with nothing to fill in is not
// a template. This is the check that catches a finished reference document
// masquerading as a skeleton, which is the defect the previous version of this
// test was blind to.
// pendingSkeletonMigration names the skeletons that have no canonical fill-in.
// It is a judgement call left open, not an oversight, and it is frozen: the
// list cannot grow, so this test degrades to a ratchet rather than a gate.
//
// Each entry was independently flagged by two assessors that had read the whole
// file. They are finished reference documents shipped under a .template.md
// name, which is why there is nothing to fill in. The fix is to move each to
// docs/examples/ and repoint template-registry.json and its README row, the way
// integration-examples.md already was. That is deliberately NOT done here,
// because all five are registered templates that the /doc command resolves: the
// restructure changes the command surface, which is the maintainer's call.
//
// user-documentation.template.md was removed from this list on 2026-09-30. It
// used to be "the one with teeth": it shipped `support@yourcompany.com`,
// `sales@`, `partners@`, `press@` and `[@yourcompany]` as unmarked example
// prose that a filler emitted verbatim. Those are now canonical fill-ins with
// format-hint defaults, so the specific gap this list records — no canonical
// fill-in at all — is closed. Per the ratchet rule below, a resolved gap is
// removed from the list rather than left to hide the next real one. Removing it
// makes the file *stricter* from here: it is now permanently held to the
// at-least-one-canonical-fill-in standard and fails if it ever regresses to
// zero. The file's remaining legacy `[BRACKET]` markers are tracked by the
// census in templates/README.md, not by this list.
// test-plan.template.md was removed from this list on 2026-09-30. Its only
// pre-existing `{{ }}` was `test-cases/{{AREA}}.md` inside a code span, which
// stripNonProse blanks, so it genuinely had no prose-level fill-in. Unifying its
// hardcoded coverage thresholds onto the constitution's {{TEST_COVERAGE_MIN|90}}
// definition site put a real one in prose, which closes the gap this list
// records. Same rule as user-documentation above: a resolved gap is removed, and
// the file is thereafter held to the at-least-one-canonical-fill-in standard.
// api-documentation.template.md was removed from this list on 2026-09-30. It was
// blocked on exactly this gap: the file is a published API reference, so every
// slot read as example prose, and none of them used the canonical form. Its
// header, auth, and support links are now `{{API_NAME}}`, `{{BASE_URL|url}}` and
// the rest, which puts a real fill-in in prose and closes the gap. The gap was
// closed in the file rather than the list deleted to make a test pass, because
// the alternative — leaving ten legacy [BRACKET] slots in a template that is
// copied verbatim by consumers — is the defect the list exists to prevent.
// user-training-guide.template.md was removed from this list on 2026-09-30. Its
// entry said "zero fill-ins, facilitator content rather than a form", which
// misread the document. A facilitator guide is the densest fill-in surface in
// the corpus, because its reader is handed a blank copy of a course: every
// learning objective, the audience profile, the three workflows, the navigation
// tree, and every row of the settings, report, glossary and shortcut tables is
// a slot someone must complete before teaching. All 227 legacy [BRACKET] slots
// are now {{PLACEHOLDER|default}} fill-ins — 242 including the ones inside the
// fenced procedure blocks and the menu tree. The relocation this entry proposed
// (move it to docs/examples/, because it "reads like a finished reference
// document") is the fix for a reference document, and this is not one: there was
// never a finished document here to preserve, only unfilled slots.
// security-assessment.template.md was removed from this list on 2026-09-30, the
// last entry. Its entry said "worked assessment rather than a form", which
// described how it READ — a filled-in report — rather than what it IS: a blank
// assessment whose every cell, rating and owner is a slot. Its [BRACKET] tokens
// are now {{PLACEHOLDER}} fill-ins, the enums that remain are the legal values
// for their fields, and the bracket lists that look like examples are examples.
// The list is now empty, which is the first time that has been true.
//
// It is kept, empty, rather than deleted. The mechanism is the point: an entry
// is a dated claim that a file is KNOWN to be short of canonical fill-ins, and
// the test's rule is that a resolved gap leaves the list so the next real one is
// visible. Deleting the list would delete the ability to record the next gap,
// which is how a corpus quietly stops being checked.
var pendingSkeletonMigration = map[string]string{}

func TestEverySkeletonHasAtLeastOneCanonicalFillIn(t *testing.T) {
	docs := repoDocs(t)
	if docs == "" {
		t.Skip("not a repository checkout")
	}
	root := filepath.Dir(filepath.Dir(docs))

	var empty, stillPending []string
	checked := 0
	for _, dir := range []string{
		filepath.Join(root, "templates"),
		filepath.Join(root, ".plaesy", "templates"),
	} {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".template.md") {
				return nil
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			rel = filepath.ToSlash(rel)
			if !isSkeleton(rel) {
				return nil
			}
			// The .plaesy tree is a generated mirror of templates/; report the
			// source file once rather than every defect twice.
			if strings.HasPrefix(rel, ".plaesy/") {
				return nil
			}
			checked++
			raw, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			body := stripNonProse(string(raw))
			if !canonicalFill.MatchString(body) {
				lines := len(strings.Split(string(raw), "\n"))
				if why, ok := pendingSkeletonMigration[rel]; ok {
					stillPending = append(stillPending, rel)
					t.Logf("PENDING migration: %s (%d lines) — %s", rel, lines, why)
				} else {
					empty = append(empty, fmt.Sprintf("%s (%d lines, no {{ }} fill-in)", rel, lines))
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}
	if checked == 0 {
		t.Fatal("no skeletons found; the walk is wrong, which would make this test vacuous")
	}
	sort.Strings(empty)
	for _, e := range empty {
		t.Errorf("%s has no canonical fill-in at all. A file named .template.md with nothing to "+
			"fill in is either a finished reference document in the wrong place — move it to "+
			"docs/examples/ and repoint template-registry.json, the way integration-examples.md "+
			"was — or a skeleton that was never migrated.", e)
	}

	// The ratchet half: a file may be *removed* from the pending list by
	// migrating it, but nothing may be added. Without this the exception list
	// would quietly become a dumping ground and the assertion would stop
	// meaning anything.
	var live []string
	for rel := range pendingSkeletonMigration {
		found := false
		for _, s := range stillPending {
			if s == rel {
				found = true
				break
			}
		}
		if !found {
			live = append(live, rel)
		}
	}
	sort.Strings(live)
	for _, rel := range live {
		t.Errorf("%s is still listed in pendingSkeletonMigration but now has a canonical "+
			"fill-in. Remove it from the list — the exception exists to record a known gap, "+
			"and a resolved gap left in the list hides the next real one.", rel)
	}
	t.Logf("%d skeleton(s) checked, %d failing, %d pending migration", checked, len(empty), len(stillPending))
}

// TestDocumentedPlaceholdersAreMeasuredNotMigrated records the legacy syntax
// that lives in prose. It asserts nothing, on purpose: this is the measurement
// the earlier eyeballed counts (230 / 93 / 23) should have been, and it logs
// the real numbers so a future reader does not have to guess again. It also
// pins the reasoning, so nobody "fixes" `<section>` in a mapping.json table
// into `{{SECTION}}` and breaks the document.
func TestDocumentedPlaceholdersAreMeasuredNotMigrated(t *testing.T) {
	docs := repoDocs(t)
	if docs == "" {
		t.Skip("not a repository checkout")
	}
	root := filepath.Dir(filepath.Dir(docs))

	counts := map[string]int{}
	files := map[string]map[string]bool{}
	for _, dir := range []string{"prompts", "instructions", "agents", "checklists", "docs"} {
		full := filepath.Join(root, dir)
		if _, err := os.Stat(full); err != nil {
			continue
		}
		err := filepath.Walk(full, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				base := filepath.Base(path)
				// Generated and historical: .plaesy mirrors the
				// skeleton, docs/assessment is a dated record of
				// what was broken and must keep quoting it.
				if base == "assessment" || base == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".md") {
				return nil
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			body, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			for _, h := range findLegacy(string(body), filepath.ToSlash(rel)) {
				counts[h.rule]++
				if files[h.rule] == nil {
					files[h.rule] = map[string]bool{}
				}
				files[h.rule][h.file] = true
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}

	rules := []string{"lower_snake", "angle", "UPPER_SNAKE"}
	var parts []string
	for _, r := range rules {
		parts = append(parts, fmt.Sprintf("%s=%d/%dfiles", r, counts[r], len(files[r])))
	}
	t.Logf("documented (non-skeleton) legacy tokens: %s", strings.Join(parts, "  "))
	t.Log("these are documentation metavariables, not debt: rewriting them would make usage " +
		"lines and JSON-shape tables unreadable. Only skeletons are gated, by " +
		"TestTemplateSkeletonsUseOnlyCanonicalPlaceholders.")
}
