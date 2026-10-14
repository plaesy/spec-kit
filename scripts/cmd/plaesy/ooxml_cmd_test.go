package main

// Command-level tests for `plaesy validate docx|pptx|xlsx` plus the two
// reporting commands that had no coverage at all (printRuleList, status).
//
// The internal/validate package is now at 93%, but these three commands were
// still near 5% because each one only needed a *file path* and nothing in the
// repo had a real .docx/.pptx/.xlsx to hand. The fixtures are built here with
// archive/zip for the same reason internal/validate does it: an OOXML file is a
// ZIP of XML parts.
//
// The behaviour worth pinning is the reporting contract, not the re-validation.
// The bug class is "the command returns success without having checked
// anything" — an error must propagate as a non-nil error so the shell sees a
// failure, and a passing run must print the no-render-check notice, because
// structural validation is not the same claim as "this renders correctly".

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ooxmlContentTypes is the minimal [Content_Types].xml every fixture carries.
const ooxmlContentTypes = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="xml" ContentType="application/xml"/>
</Types>`

// writeOoxmlZip builds an OOXML-shaped archive. An empty parts map is a valid
// but contentless archive — the case the validators must reject loudly rather
// than pass, because "the zip opened" is not "the file is usable".
func writeOoxmlZip(t *testing.T, path string, parts map[string]string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	for name, content := range parts {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return path
}

func docxCmdParts(paragraphs ...string) map[string]string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	for _, p := range paragraphs {
		b.WriteString(`<w:p><w:r><w:t>` + p + `</w:t></w:r></w:p>`)
	}
	b.WriteString(`</w:body></w:document>`)
	return map[string]string{
		"[Content_Types].xml": ooxmlContentTypes,
		"word/document.xml":   b.String(),
	}
}

func pptxCmdParts(titles ...bool) map[string]string {
	parts := map[string]string{"[Content_Types].xml": ooxmlContentTypes}
	var lst strings.Builder
	for i, hasTitle := range titles {
		n := string(rune('1' + i))
		lst.WriteString(`<p:sldId id="256` + n + `" r:id="rId` + n + `"/>`)
		ph := ""
		if hasTitle {
			ph = `<p:ph type="title"/>`
		}
		parts["ppt/slides/slide"+n+".xml"] =
			`<?xml version="1.0"?><p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">` +
				`<p:cSld><p:spTree><p:sp><p:nvSpPr><p:nvPr>` + ph + `</p:nvPr></p:nvSpPr></p:sp>` +
				`</p:spTree></p:cSld></p:sld>`
	}
	parts["ppt/presentation.xml"] =
		`<?xml version="1.0"?><p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">` +
			`<p:sldIdLst>` + lst.String() + `</p:sldIdLst></p:presentation>`
	return parts
}

func xlsxCmdParts(sheets ...string) map[string]string {
	parts := map[string]string{"[Content_Types].xml": ooxmlContentTypes}
	var list, rels strings.Builder
	for i, name := range sheets {
		id := string(rune('1' + i))
		list.WriteString(`<sheet name="` + name + `" sheetId="` + id + `" r:id="rId` + id + `"/>`)
		rels.WriteString(`<Relationship Id="rId` + id + `" Target="worksheets/sheet` + id + `.xml"/>`)
		parts["xl/worksheets/sheet"+id+".xml"] =
			`<?xml version="1.0"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
				`<sheetData><row r="1"><c r="A1"><v>1</v></c></row></sheetData></worksheet>`
	}
	parts["xl/workbook.xml"] =
		`<?xml version="1.0"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" ` +
			`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
			`<sheets>` + list.String() + `</sheets>` +
			`<Relationships>` + rels.String() + `</Relationships></workbook>`
	return parts
}

func runLeafCmd(t *testing.T, cmd *cobra.Command, args ...string) error {
	t.Helper()
	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	return cmd.Execute()
}

// ---- validate docx -------------------------------------------------------

// A well-formed document must pass AND print the paragraph/table counts: the
// counts are the only evidence the command did any work at all, so a silent pass
// is indistinguishable from a command that returned nil unconditionally.
func TestDocxCmdPassesAndReportsCounts(t *testing.T) {
	path := writeOoxmlZip(t, filepath.Join(t.TempDir(), "plaesy-test", "ok.docx"),
		docxCmdParts("First.", "Second."))

	out, err := captureStdout(t, func() error { return runLeafCmd(t, newDocxCmd(), path) })
	if err != nil {
		t.Fatalf("a well-formed docx must validate, got %v\n%s", err, out)
	}
	for _, want := range []string{"paragraphs: 2", "tables: 0", "structural validation passed"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	// Structural validation is a weaker claim than "renders correctly"; saying
	// so is what stops a user treating a pass as a rendering guarantee.
	if !strings.Contains(out, "render") {
		t.Errorf("a docx pass must disclose that no render check happened:\n%s", out)
	}
}

// The failure direction is the whole point. An empty document must return an
// error, not a zero count and nil.
func TestDocxCmdFailsOnEmptyDocument(t *testing.T) {
	path := writeOoxmlZip(t, filepath.Join(t.TempDir(), "plaesy-test", "empty.docx"), docxCmdParts())

	out, err := captureStdout(t, func() error { return runLeafCmd(t, newDocxCmd(), path) })
	if err == nil {
		t.Fatalf("an empty document must fail, not pass:\n%s", out)
	}
	if strings.Contains(out, "passed validation") {
		t.Errorf("failure output must not claim success:\n%s", out)
	}
}

// A file that is not a zip at all is the most common real-world mistake, and it
// must be an error rather than a panic from the archive reader.
func TestDocxCmdFailsOnNonZip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plaesy-test", "notes.docx")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("this is plain text, not a docx"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := captureStdout(t, func() error { return runLeafCmd(t, newDocxCmd(), path) }); err == nil {
		t.Fatal("a non-zip file must not be reported as a valid document")
	}
}

// Missing file: the command takes the path from the user, so a typo has to fail
// loudly rather than validate nothing.
func TestDocxCmdFailsOnMissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "plaesy-test", "nope.docx")
	if _, err := captureStdout(t, func() error { return runLeafCmd(t, newDocxCmd(), missing) }); err == nil {
		t.Fatal("a missing file must be an error")
	}
}

// ExactArgs(1): a bare `validate docx` validates nothing, so it must be rejected
// at the argument layer rather than defaulting to some path.
func TestDocxCmdRequiresExactlyOneArgument(t *testing.T) {
	for _, args := range [][]string{{}, {"a.docx", "b.docx"}} {
		if err := runLeafCmd(t, newDocxCmd(), args...); err == nil {
			t.Errorf("args %v must be rejected, want an argument-count error", args)
		}
	}
}

// ---- validate pptx -------------------------------------------------------

func TestPptxCmdPassesAndReportsSlideCount(t *testing.T) {
	path := writeOoxmlZip(t, filepath.Join(t.TempDir(), "plaesy-test", "deck.pptx"),
		pptxCmdParts(true, false))

	out, err := captureStdout(t, func() error { return runLeafCmd(t, newPptxCmd(), path) })
	if err != nil {
		t.Fatalf("a well-formed deck must validate, got %v\n%s", err, out)
	}
	// The slide count is read from presentation.xml's sldIdLst. This is the
	// regression guard for the duplicated sldIdLst XML tag that silently
	// reported every deck as empty.
	if !strings.Contains(out, "slide count: 2") {
		t.Errorf("expected the deck to report 2 slides:\n%s", out)
	}
}

// The per-slide title warning is the check's whole value: a deck that validates
// but has untitled slides is the thing a human must act on, so the warning must
// be printed even though the run passes.
func TestPptxCmdWarnsAboutSlidesWithNoTitle(t *testing.T) {
	path := writeOoxmlZip(t, filepath.Join(t.TempDir(), "plaesy-test", "deck.pptx"),
		pptxCmdParts(true, false))

	out, err := captureStdout(t, func() error { return runLeafCmd(t, newPptxCmd(), path) })
	if err != nil {
		t.Fatalf("a missing title is a warning, not a failure: %v\n%s", err, out)
	}
	if !strings.Contains(out, "slide 2 has no title placeholder") {
		t.Errorf("the untitled slide must be reported by number:\n%s", out)
	}
	if strings.Contains(out, "slide 1 has no title") {
		t.Errorf("the titled slide must NOT be reported:\n%s", out)
	}
}

// The optional second argument is a caller-supplied expectation. Matching it
// must pass; ignoring it would make the argument decorative.
func TestPptxCmdEnforcesTheExpectedSlideCount(t *testing.T) {
	path := writeOoxmlZip(t, filepath.Join(t.TempDir(), "plaesy-test", "deck.pptx"),
		pptxCmdParts(true, false, true))

	if out, err := captureStdout(t, func() error { return runLeafCmd(t, newPptxCmd(), path, "3") }); err != nil {
		t.Fatalf("a deck with the expected 3 slides must pass, got %v\n%s", err, out)
	}
	out, err := captureStdout(t, func() error { return runLeafCmd(t, newPptxCmd(), path, "7") })
	if err == nil {
		t.Fatalf("expecting 7 slides of a 3-slide deck must fail:\n%s", out)
	}
}

// A non-numeric expectation is a typo. It must be reported as such rather than
// silently treated as "no expectation", which would turn a wrong number into a
// passing run.
func TestPptxCmdRejectsANonNumericExpectedCount(t *testing.T) {
	path := writeOoxmlZip(t, filepath.Join(t.TempDir(), "plaesy-test", "deck.pptx"), pptxCmdParts(true))
	out, err := captureStdout(t, func() error { return runLeafCmd(t, newPptxCmd(), path, "three") })
	if err == nil {
		t.Fatalf("a non-numeric slide count must be an error:\n%s", out)
	}
	if !strings.Contains(err.Error(), "three") {
		t.Errorf("the error should quote what could not be parsed: %v", err)
	}
}

func TestPptxCmdRejectsMoreThanTwoArguments(t *testing.T) {
	if err := runLeafCmd(t, newPptxCmd(), "a.pptx", "1", "extra"); err == nil {
		t.Fatal("pptx takes at most a path and an expected count")
	}
}

func TestPptxCmdFailsOnArchiveWithNoSlides(t *testing.T) {
	parts := map[string]string{"[Content_Types].xml": ooxmlContentTypes}
	path := writeOoxmlZip(t, filepath.Join(t.TempDir(), "plaesy-test", "empty.pptx"), parts)

	out, err := captureStdout(t, func() error { return runLeafCmd(t, newPptxCmd(), path) })
	if err == nil {
		t.Fatalf("a deck with no slides must fail:\n%s", out)
	}
	if strings.Contains(out, "passed validation") {
		t.Errorf("failure output must not claim success:\n%s", out)
	}
}

// A slide with no title placeholder is a legitimate draft, not a broken file —
// the command must not reject it, or the check would be unusable day to day.
func TestPptxCmdAcceptsADeckWithoutATitleSlide(t *testing.T) {
	path := writeOoxmlZip(t, filepath.Join(t.TempDir(), "plaesy-test", "notitle.pptx"),
		pptxCmdParts(false, false))
	if out, err := captureStdout(t, func() error { return runLeafCmd(t, newPptxCmd(), path) }); err != nil {
		t.Fatalf("a deck with no title slide must still validate, got %v\n%s", err, out)
	}
}

func TestPptxCmdFailsOnNonZip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plaesy-test", "deck.pptx")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := captureStdout(t, func() error { return runLeafCmd(t, newPptxCmd(), path) }); err == nil {
		t.Fatal("a non-zip deck must not be reported as valid")
	}
}

// ---- validate xlsx -------------------------------------------------------

func TestXlsxCmdPassesAndReportsSheetNames(t *testing.T) {
	path := writeOoxmlZip(t, filepath.Join(t.TempDir(), "plaesy-test", "book.xlsx"),
		xlsxCmdParts("Data", "Notes"))

	out, err := captureStdout(t, func() error { return runLeafCmd(t, newXlsxCmd(), path) })
	if err != nil {
		t.Fatalf("a well-formed workbook must validate, got %v\n%s", err, out)
	}
	for _, want := range []string{"Data", "Notes", "structural validation passed"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

// The optional trailing arguments are an expectation the caller is making about
// the workbook. A workbook that lacks the named sheet is a real failure, and it
// is the reason the variadic exists — so it must be enforced, not ignored.
func TestXlsxCmdFailsWhenAnExpectedSheetIsAbsent(t *testing.T) {
	path := writeOoxmlZip(t, filepath.Join(t.TempDir(), "plaesy-test", "book.xlsx"),
		xlsxCmdParts("Data"))

	out, err := captureStdout(t, func() error { return runLeafCmd(t, newXlsxCmd(), path, "Missing") })
	if err == nil {
		t.Fatalf("naming a sheet the workbook lacks must fail:\n%s", out)
	}
	if !strings.Contains(out, "Missing") {
		t.Errorf("the error should name the sheet the caller asked for:\n%s", out)
	}
}

func TestXlsxCmdPassesWhenExpectedSheetsArePresent(t *testing.T) {
	path := writeOoxmlZip(t, filepath.Join(t.TempDir(), "plaesy-test", "book.xlsx"),
		xlsxCmdParts("Data", "Notes"))
	if out, err := captureStdout(t, func() error { return runLeafCmd(t, newXlsxCmd(), path, "Data", "Notes") }); err != nil {
		t.Fatalf("expectations that match must pass, got %v\n%s", err, out)
	}
}

func TestXlsxCmdFailsOnNonZip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plaesy-test", "book.xlsx")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := captureStdout(t, func() error { return runLeafCmd(t, newXlsxCmd(), path) }); err == nil {
		t.Fatal("a non-zip workbook must not be reported as valid")
	}
}

// MinimumNArgs(1): the path is mandatory, unlike the sheet expectations.
func TestXlsxCmdRequiresAPath(t *testing.T) {
	if err := runLeafCmd(t, newXlsxCmd()); err == nil {
		t.Fatal("xlsx with no path must be rejected")
	}
}

// ---- printRuleList -------------------------------------------------------

// `validate markdown --list-rules` is how a user decides whether a clean run
// means anything. Two things must hold: the implemented count must match the
// number of lines printed (a count that drifts from the list is a lie), and the
// disclaimer must be present, because a green run covers ~half the rule set.
func TestPrintRuleListCountMatchesTheListedRules(t *testing.T) {
	out, err := captureStdout(t, func() error { printRuleList(); return nil })
	if err != nil {
		t.Fatal(err)
	}
	var count int
	for _, line := range strings.Split(out, "\n") {
		if idx := strings.Index(line, "Implemented markdownlint rules ("); idx >= 0 {
			rest := line[idx+len("Implemented markdownlint rules ("):]
			rest = rest[:strings.Index(rest, ")")]
			if count, err = strconv.Atoi(rest); err != nil {
				t.Fatalf("the header count must be a number, got %q: %v", rest, err)
			}
		}
	}
	if count == 0 {
		t.Fatalf("no rule count found in output:\n%s", out)
	}
	// Count the indented "  <RULEID>  description" lines that follow the header.
	// The registry is not markdownlint-only: the GitHub ruleset rules carry GHA
	// ids, so a count that only looked for MD would silently under-report.
	listed := 0
	for _, line := range strings.Split(out, "\n") {
		if isRuleListLine(line) {
			listed++
		}
	}
	if listed != count {
		t.Errorf("header claims %d rules but %d are listed:\n%s", count, listed, out)
	}
}

// The honest framing is the feature: without it a user reads "no violations" as
// "markdownlint would be happy", which is false for the unimplemented rules.
func TestPrintRuleListDisclosesWhatIsNotChecked(t *testing.T) {
	out, err := captureStdout(t, func() error { printRuleList(); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Not implemented") {
		t.Errorf("the list must say which rules are not checked:\n%s", out)
	}
	if !strings.Contains(out, "markdownlint would be happy") {
		t.Errorf("the list must state that a clean run is not markdownlint-clean:\n%s", out)
	}
}

// Every rule the linter reports must be printable, and every printed rule must
// have a description. A registry entry with an empty description would render as
// a bare ID and tell the user nothing about what just failed.
func TestPrintRuleListGivesEveryRuleADescription(t *testing.T) {
	out, err := captureStdout(t, func() error { printRuleList(); return nil })
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(out, "\n") {
		if !isRuleListLine(line) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[1] == "" {
			t.Errorf("rule line has no description: %q", line)
		}
	}
}

// isRuleListLine recognises one "  <ID>  <description>" entry as printRuleList
// renders it. The ids are MDxxx for markdownlint rules and GHAxxx for the
// GitHub ruleset rules, so both prefixes are real.
func isRuleListLine(line string) bool {
	return strings.HasPrefix(line, "  MD") || strings.HasPrefix(line, "  GHA")
}

// ---- status --------------------------------------------------------------

// `status` is the first command a new user runs, so both facts it reports must
// be present: the version, and which of the two environment facts hold. It has
// no error path by design — a missing git is a fact to report, not a failure.
func TestStatusReportsVersionAndEnvironmentFacts(t *testing.T) {
	out, err := captureStdout(t, func() error { return runLeafCmd(t, newStatusCmd()) })
	if err != nil {
		t.Fatalf("status must not fail on a healthy environment: %v", err)
	}
	if !strings.Contains(out, "Plaesy Constitution Kit") {
		t.Errorf("status must print the banner:\n%s", out)
	}
	if !strings.Contains(out, "version") {
		t.Errorf("status must report the version:\n%s", out)
	}
	// Exactly one of the two git facts must be resolved, never neither: a
	// status that silently omits git is indistinguishable from a broken build.
	if !strings.Contains(out, "git") {
		t.Errorf("status must say something about git:\n%s", out)
	}
	if !strings.Contains(out, "git repository") {
		t.Errorf("status must state whether this is a git repository:\n%s", out)
	}
}

// status takes no arguments; extra ones are a typo and must be surfaced.
// Without Args: cobra.NoArgs, cobra's ArbitraryArgs default swallowed
// `plaesy status stray` and reported a clean bill of health.
func TestStatusRejectsArguments(t *testing.T) {
	if err := runLeafCmd(t, newStatusCmd(), "extra"); err == nil {
		t.Fatal("status must reject an unexpected argument")
	}
}

// Same class of defect in the other two argument-free leaf commands: a report
// command that quietly accepts a typo still prints a plausible report, so the
// user believes they asked for something they did not.
func TestArgumentFreeLeafCommandsRejectArguments(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{"features paths", newFeaturesPathsCmd()},
		{"features validate", newFeaturesValidateCmd()},
	} {
		if err := runLeafCmd(t, tc.cmd, "stray"); err == nil {
			t.Errorf("%s must reject an unexpected argument", tc.name)
		}
	}
}
