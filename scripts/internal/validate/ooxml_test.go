package validate

// The OOXML validators were the last completely untested surface in the module:
// Docx, Pptx, Xlsx, and every helper they share measured 0%, because each one
// needs a real .docx/.pptx/.xlsx and the repo carries none. Building them here
// rather than checking binaries in is deliberate — an OOXML file is a ZIP of XML
// parts, which archive/zip writes in a few lines, so the fixtures stay readable
// in the test and cannot rot out of sync with the format.
//
// What these tests are really for is the failure direction. The bash originals
// shelled out to python-docx/openpyxl and reported "could not open it"; a Go port
// that only checks "the zip opens" would pass a file no tool can use, and one
// that reports a missing part as success would pass an empty archive.

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeZip builds a zip archive at path from name -> content. An empty or nil
// map is a valid, empty archive: that is the "opens fine, contains nothing" case
// several validators have to reject with a specific message.
func writeZip(t *testing.T, path string, parts map[string]string) string {
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

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="xml" ContentType="application/xml"/>
</Types>`

// docxDocument builds a word/document.xml with the given paragraphs. Each entry
// becomes one <w:p>; a paragraph containing <w:tbl> is not expressible here, so
// table counting is exercised separately.
func docxDocument(paragraphs ...string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	for _, p := range paragraphs {
		b.WriteString(`<w:p><w:r><w:t>` + p + `</w:t></w:r></w:p>`)
	}
	b.WriteString(`</w:body></w:document>`)
	return b.String()
}

func docxParts(paragraphs ...string) map[string]string {
	return map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"word/document.xml":   docxDocument(paragraphs...),
	}
}

func TestDocxCountsParagraphsAndTables(t *testing.T) {
	parts := docxParts("First paragraph.", "Second paragraph.")
	// A table is a <w:tbl> sibling of the paragraphs; it must be counted
	// separately, since the "empty document" check accepts either.
	parts["word/document.xml"] = strings.Replace(
		docxParts("Only paragraph.")["word/document.xml"],
		"</w:body>",
		`<w:tbl><w:tr><w:tc><w:p><w:r><w:t>cell</w:t></w:r></w:p></w:tc></w:tr></w:tbl></w:body>`, 1)
	path := writeZip(t, filepath.Join(t.TempDir(), "doc.docx"), parts)

	res, err := Docx(path)
	if err != nil {
		t.Fatalf("a well-formed docx must validate: %v", err)
	}
	// Two paragraphs: the explicit one and the one inside the table cell. A
	// validator that ignored nesting would report one.
	if res.Paragraphs != 2 {
		t.Errorf("Paragraphs = %d, want 2", res.Paragraphs)
	}
	if res.Tables != 1 {
		t.Errorf("Tables = %d, want 1", res.Tables)
	}
	if len(res.Leftover) != 0 {
		t.Errorf("Leftover = %v, want none", res.Leftover)
	}
}

// A document whose only content is a table is not empty. The bash check was
// "no paragraphs or tables", so the error must require both to be zero.
func TestDocxWithOnlyATableIsNotEmpty(t *testing.T) {
	path := writeZip(t, filepath.Join(t.TempDir(), "table-only.docx"), map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"word/document.xml": `<?xml version="1.0"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>
<w:tbl><w:tr><w:tc><w:p><w:r><w:t>cell</w:t></w:r></w:p></w:tc></w:tr></w:tbl>
</w:body></w:document>`,
	})
	res, err := Docx(path)
	if err != nil {
		t.Fatalf("a table-only document must not be reported as empty: %v", err)
	}
	if res.Tables != 1 {
		t.Errorf("Tables = %d, want 1", res.Tables)
	}
}

func TestDocxRejectsAnEmptyDocument(t *testing.T) {
	path := writeZip(t, filepath.Join(t.TempDir(), "empty.docx"), map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"word/document.xml": `<?xml version="1.0"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body></w:body></w:document>`,
	})
	_, err := Docx(path)
	if err == nil || !strings.Contains(err.Error(), "no paragraphs or tables") {
		t.Fatalf("err = %v, want an empty-document error", err)
	}
}

// The check that exists to catch a generated report that still has its
// template in it. {{ }} split across two runs must still be caught, because that
// is how a Jinja placeholder actually lands in a document.
func TestDocxRejectsLeftoverTemplateTags(t *testing.T) {
	// A Jinja placeholder lands in a generated document split across runs, which
	// is why the paragraph text is joined before the tag is matched. Both shapes
	// have to be caught, or the split one slips through every time.
	cases := map[string]string{
		"within one run": `<w:p><w:r><w:t>Report for {{ customer_name }}.</w:t></w:r></w:p>`,
		"split across runs": `<w:p><w:r><w:t>Report for {{ customer</w:t></w:r>` +
			`<w:r><w:t>_name }}.</w:t></w:r></w:p>`,
	}
	for name, paragraph := range cases {
		t.Run(name, func(t *testing.T) {
			parts := docxParts("Clean paragraph.")
			parts["word/document.xml"] = strings.Replace(
				docxParts("Clean paragraph.")["word/document.xml"],
				"</w:body>", paragraph+"</w:body>", 1)
			path := writeZip(t, filepath.Join(t.TempDir(), "tagged.docx"), parts)
			res, err := Docx(path)
			if err == nil || !strings.Contains(err.Error(), "leftover unfilled template tags") {
				t.Fatalf("err = %v, want a leftover-tag error", err)
			}
			if len(res.Leftover) == 0 {
				t.Error("the offending paragraph text was not reported")
			}
		})
	}
}

func TestDocxRejectsBrokenContainers(t *testing.T) {
	dir := t.TempDir()
	notAZip := filepath.Join(dir, "not-a-zip.docx")
	if err := os.WriteFile(notAZip, []byte("this is plain text"), 0o644); err != nil {
		t.Fatal(err)
	}
	emptyArchive := writeZip(t, filepath.Join(dir, "empty.docx"), nil)
	noContentTypes := writeZip(t, filepath.Join(dir, "no-ct.docx"), map[string]string{
		"word/document.xml": docxDocument("Text."),
	})
	noDocument := writeZip(t, filepath.Join(dir, "no-doc.docx"), map[string]string{
		"[Content_Types].xml": contentTypesXML,
	})
	brokenXML := writeZip(t, filepath.Join(dir, "broken.docx"), map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"word/document.xml":   `<?xml version="1.0"?><w:document><w:body><w:p>`,
	})

	cases := []struct {
		name string
		path string
		want string
	}{
		{"not a zip", notAZip, "not a well-formed OOXML"},
		{"missing file", filepath.Join(dir, "absent.docx"), "file not found"},
		{"directory", dir, "file not found"},
		{"empty archive", emptyArchive, "missing [Content_Types].xml"},
		{"no content types", noContentTypes, "missing [Content_Types].xml"},
		{"no document part", noDocument, "missing required part"},
		{"malformed document xml", brokenXML, "not well-formed XML"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Docx(tc.path)
			if err == nil {
				t.Fatal("a broken docx must be an error")
			}
			// Each case has its own message: a caller that only sees "invalid"
			// cannot tell a truncated file from a missing part.
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %q, want it to mention %q", err, tc.want)
			}
		})
	}
}

// pptxParts builds a presentation whose sldIdLst names count slides, with the
// given per-slide title-placeholder state (true = has a title).
func pptxParts(titles ...bool) map[string]string {
	parts := map[string]string{"[Content_Types].xml": contentTypesXML}
	var lst strings.Builder
	for i := range titles {
		lst.WriteString(`<p:sldId id="256` + string(rune('0'+i)) + `" r:id="rId` + string(rune('1'+i)) + `"/>`)
		ph := ""
		if titles[i] {
			ph = `<p:ph type="title"/>`
		}
		parts["ppt/slides/slide"+string(rune('1'+i))+".xml"] =
			`<?xml version="1.0"?><p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">` +
				`<p:cSld><p:spTree><p:sp><p:nvSpPr><p:nvPr>` + ph + `</p:nvPr></p:nvSpPr></p:sp>` +
				`</p:spTree></p:cSld></p:sld>`
	}
	parts["ppt/presentation.xml"] =
		`<?xml version="1.0"?><p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">` +
			`<p:sldIdLst>` + lst.String() + `</p:sldIdLst></p:presentation>`
	return parts
}

func TestPptxCountsSlidesAndReportsMissingTitles(t *testing.T) {
	path := writeZip(t, filepath.Join(t.TempDir(), "deck.pptx"), pptxParts(true, false, true))

	res, err := Pptx(path, 0)
	if err != nil {
		t.Fatalf("a well-formed pptx must validate: %v", err)
	}
	if res.SlideCount != 3 {
		t.Errorf("SlideCount = %d, want 3", res.SlideCount)
	}
	// A missing title is a warning in the bash original, not a failure: the
	// validator's exit code has to stay usable for a deck that is merely imperfect.
	if len(res.SlidesNoTitle) != 1 || res.SlidesNoTitle[0] != 2 {
		t.Errorf("SlidesNoTitle = %v, want [2]", res.SlidesNoTitle)
	}
}

func TestPptxSlideCountExpectation(t *testing.T) {
	path := writeZip(t, filepath.Join(t.TempDir(), "deck.pptx"), pptxParts(true, true))

	// 0 means "no expectation", which is what the bare dispatch passes.
	res, err := Pptx(path, 0)
	if err != nil || res.SlideCount != 2 {
		t.Fatalf("no expectation: res = %+v, err = %v", res, err)
	}
	// A wrong expectation is an error, and it names both numbers so the reader
	// does not have to count the slides again.
	res, err = Pptx(path, 5)
	if err == nil || !strings.Contains(err.Error(), "expected 5 slides, got 2") {
		t.Fatalf("err = %v, want a slide-count mismatch naming both numbers", err)
	}
	// The result still comes back, so a caller can report what it found.
	if res == nil || res.SlideCount != 2 {
		t.Errorf("the result was discarded on failure: %+v", res)
	}
}

func TestPptxRejectsBrokenContainers(t *testing.T) {
	dir := t.TempDir()
	noPresentation := writeZip(t, filepath.Join(dir, "no-pres.pptx"), map[string]string{
		"[Content_Types].xml": contentTypesXML,
	})
	brokenPresentation := writeZip(t, filepath.Join(dir, "broken.pptx"), map[string]string{
		"[Content_Types].xml":  contentTypesXML,
		"ppt/presentation.xml": `<?xml version="1.0"?><p:presentation><p:sldIdLst>`,
	})
	brokenSlide := writeZip(t, filepath.Join(dir, "broken-slide.pptx"), func() map[string]string {
		parts := pptxParts(true)
		parts["ppt/slides/slide1.xml"] = `<?xml version="1.0"?><p:sld><p:cSld>`
		return parts
	}())

	cases := []struct {
		name string
		path string
		want string
	}{
		{"no presentation part", noPresentation, "missing required part"},
		{"malformed presentation xml", brokenPresentation, "not well-formed XML"},
		{"malformed slide xml", brokenSlide, "not well-formed XML"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Pptx(tc.path, 0); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}

// A slide part that is not named slideN.xml must not be counted, and a
// relationship or layout part living under ppt/slides/ must not be mistaken for
// one. Real decks have ppt/slides/_rels/slide1.xml.rels.
func TestPptxIgnoresNonSlidePartsUnderSlides(t *testing.T) {
	parts := pptxParts(true)
	parts["ppt/slides/_rels/slide1.xml.rels"] = `<?xml version="1.0"?><Relationships/>`
	parts["ppt/slideLayouts/slideLayout1.xml"] = `<?xml version="1.0"?><p:sldLayout/>`
	path := writeZip(t, filepath.Join(t.TempDir(), "deck.pptx"), parts)

	res, err := Pptx(path, 0)
	if err != nil {
		t.Fatalf("a deck with rels parts must validate: %v", err)
	}
	if res.SlideCount != 1 {
		t.Errorf("SlideCount = %d, want 1", res.SlideCount)
	}
}

func xlsxParts(sheets ...string) map[string]string {
	parts := map[string]string{"[Content_Types].xml": contentTypesXML}
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
			`<sheets>` + list.String() + `</sheets></workbook>`
	parts["xl/_rels/workbook.xml.rels"] =
		`<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
			rels.String() + `</Relationships>`
	return parts
}

func TestXlsxListsSheetsAndChecksExpectations(t *testing.T) {
	path := writeZip(t, filepath.Join(t.TempDir(), "book.xlsx"), xlsxParts("Summary", "Detail"))

	res, err := Xlsx(path, nil)
	if err != nil {
		t.Fatalf("a well-formed xlsx must validate: %v", err)
	}
	if strings.Join(res.Sheets, ",") != "Summary,Detail" {
		t.Errorf("Sheets = %v, want [Summary Detail]", res.Sheets)
	}
	// A named sheet that exists passes, in any order and alongside others.
	if _, err := Xlsx(path, []string{"Detail"}); err != nil {
		t.Errorf("an existing sheet was reported missing: %v", err)
	}
	// One that does not is an error naming what was available.
	_, err = Xlsx(path, []string{"Missing"})
	if err == nil || !strings.Contains(err.Error(), "Missing") {
		t.Fatalf("err = %v, want a missing-sheet error", err)
	}
	// The available list is in the message, so the reader can see the typo.
	if !strings.Contains(err.Error(), "Summary") {
		t.Errorf("error %q does not list the sheets that do exist", err)
	}
}

func TestXlsxRejectsAWorkbookWithNoSheets(t *testing.T) {
	parts := xlsxParts()
	parts["xl/workbook.xml"] = `<?xml version="1.0"?><workbook><sheets/></workbook>`
	path := writeZip(t, filepath.Join(t.TempDir(), "empty.xlsx"), parts)
	_, err := Xlsx(path, nil)
	if err == nil || !strings.Contains(err.Error(), "no sheets") {
		t.Fatalf("err = %v, want a no-sheets error", err)
	}
}

// The worksheet part is what openpyxl actually parsed, so its being malformed
// has to fail the run rather than be swallowed as "best effort". The r:id is
// resolved through workbook.xml.rels, so the relationship has to be honoured for
// this to be reachable.
func TestXlsxRejectsAMalformedWorksheet(t *testing.T) {
	parts := xlsxParts("Summary")
	parts["xl/worksheets/sheet1.xml"] = `<?xml version="1.0"?><worksheet><sheetData>`
	path := writeZip(t, filepath.Join(t.TempDir(), "broken.xlsx"), parts)
	_, err := Xlsx(path, nil)
	if err == nil || !strings.Contains(err.Error(), "not well-formed XML") {
		t.Fatalf("err = %v, want a malformed-worksheet error", err)
	}
}

// Relationship targets are written three different ways by three producers.
// All three name the same part, so all three must resolve — otherwise the
// worksheet check silently stops running for whoever wrote it that way.
func TestNormalizeRelTarget(t *testing.T) {
	cases := map[string]string{
		"worksheets/sheet1.xml":     "worksheets/sheet1.xml",
		"/xl/worksheets/sheet1.xml": "worksheets/sheet1.xml",
		"xl/worksheets/sheet1.xml":  "worksheets/sheet1.xml",
		"./worksheets/sheet1.xml":   "worksheets/sheet1.xml",
		// A target that is exactly the prefix is left alone rather than reduced
		// to the empty string, which would join to "xl/" and match everything.
		"xl/":   "xl/",
		"/xl/":  "/xl/",
		"./":    "./",
		"other": "other",
	}
	for in, want := range cases {
		if got := normalizeRelTarget(in); got != want {
			t.Errorf("normalizeRelTarget(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestXlsxToleratesAMissingRelsPart(t *testing.T) {
	// No rels at all: the conventional sheetN.xml naming is the fallback, and a
	// workbook missing its relationship part is still openable by Excel.
	parts := xlsxParts("Summary")
	delete(parts, "xl/_rels/workbook.xml.rels")
	path := writeZip(t, filepath.Join(t.TempDir(), "norels.xlsx"), parts)
	if _, err := Xlsx(path, nil); err != nil {
		t.Fatalf("a workbook with no rels part must still validate: %v", err)
	}
}

func TestXlsxRejectsBrokenContainers(t *testing.T) {
	dir := t.TempDir()
	noWorkbook := writeZip(t, filepath.Join(dir, "no-wb.xlsx"), map[string]string{
		"[Content_Types].xml": contentTypesXML,
	})
	brokenWorkbook := writeZip(t, filepath.Join(dir, "broken.xlsx"), map[string]string{
		"[Content_Types].xml": contentTypesXML,
		"xl/workbook.xml":     `<?xml version="1.0"?><workbook><sheets>`,
	})
	notAZip := filepath.Join(dir, "plain.xlsx")
	if err := os.WriteFile(notAZip, []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		path string
		want string
	}{
		{"no workbook part", noWorkbook, "missing required part"},
		{"malformed workbook xml", brokenWorkbook, "not well-formed XML"},
		{"not a zip", notAZip, "not a well-formed OOXML"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Xlsx(tc.path, nil); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}

func TestWellFormedXML(t *testing.T) {
	if err := wellFormedXML([]byte(`<?xml version="1.0"?><a><b/></a>`)); err != nil {
		t.Errorf("well-formed XML rejected: %v", err)
	}
	for _, bad := range []string{`<a><b></a>`, `<a attr=unquoted/>`} {
		if err := wellFormedXML([]byte(bad)); err == nil {
			t.Errorf("malformed XML accepted: %q", bad)
		}
	}
	// Two cases the decoder alone accepted: it returns bare text as character
	// data with no error, and it returns EOF immediately for an empty part. Both
	// are corrupt OOXML parts, and a slide part like this would be counted.
	for _, notXML := range []string{`not xml at all`, ``, `   `} {
		if err := wellFormedXML([]byte(notXML)); err == nil {
			t.Errorf("content with no XML element accepted: %q", notXML)
		}
	}
}

// The bash originals optionally rendered the file with LibreOffice to catch
// corruption the object model misses. This port cannot, and the contract is that
// it says so rather than doing less silently.
func TestNoRenderCheckNoticeIsExplicit(t *testing.T) {
	notice := NoRenderCheckNotice()
	for _, want := range []string{"LibreOffice", "skipping", "OOXML"} {
		if !strings.Contains(notice, want) {
			t.Errorf("the notice does not mention %q: %s", want, notice)
		}
	}
}

// The memory validator is the check `plaesy validate` runs by default, and it
// was at 0% alongside the OOXML ones. Its whole job is catching a path that
// points outside the project, so the patterns have to be pinned individually —
// one dropped pattern is a class of leak that goes unnoticed.
func TestMemoryFindsExternalReferences(t *testing.T) {
	dir := t.TempDir()
	mem := filepath.Join(dir, ".plaesy", "memory")
	if err := os.MkdirAll(mem, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mem, "leaky.md"),
		[]byte("# Notes\n\nSee ~/.claude/plans for the draft.\nAlso /tmp/scratch.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mem, "clean.md"),
		[]byte("# Clean\n\nThis file only refers to files in the project.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Only .md files are scanned: a stray .json must not be reported.
	if err := os.WriteFile(filepath.Join(mem, "config.json"), []byte(`{"path":"/tmp/x"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Memory(dir)
	if err != nil {
		t.Fatalf("Memory: %v", err)
	}
	if res.FilesChecked != 2 {
		t.Errorf("FilesChecked = %d, want 2 (.md files only)", res.FilesChecked)
	}
	if len(res.Issues) != 1 {
		t.Fatalf("Issues = %d, want exactly the one leaky file: %+v", len(res.Issues), res.Issues)
	}
	issue := res.Issues[0]
	if !strings.HasSuffix(filepath.ToSlash(issue.File), "leaky.md") {
		t.Errorf("the wrong file was reported: %q", issue.File)
	}
	// Both patterns on the line, reported with the line number the bash `grep -n`
	// printed, so a reader can jump straight to it.
	total := 0
	for _, lines := range issue.Lines {
		total += len(lines)
	}
	if total < 2 {
		t.Errorf("expected both external paths on the line, got %+v", issue.Lines)
	}
	// The fixture's two external paths are on lines 3 and 4, and the references
	// have to say which: "the file matched" is much less useful than the line.
	gotLines := map[string]bool{}
	for _, lines := range issue.Lines {
		for _, line := range lines {
			num, _, ok := strings.Cut(line, ":")
			if !ok {
				t.Errorf("line reference %q carries no line number", line)
				continue
			}
			gotLines[num] = true
		}
	}
	if len(gotLines) != 2 || !gotLines["3"] || !gotLines["4"] {
		t.Errorf("line numbers reported = %v, want 3 and 4", gotLines)
	}
}

func TestMemoryOnAProjectWithNoMemoryDirectory(t *testing.T) {
	_, err := Memory(t.TempDir())
	if err == nil {
		t.Fatal("a project with no .plaesy/memory must be an error here")
	}
	// The path is in the message, because "memory directory not found" without
	// saying where it looked leaves the user guessing.
	if !strings.Contains(err.Error(), "memory") {
		t.Errorf("err = %q, does not name what was missing", err)
	}
}

// A sheet named in workbook.xml whose worksheet part is absent from the
// archive. The name check above only proves the name is listed, and the
// per-sheet loop then skipped the missing part under a "best-effort" comment
// claiming presence had already been validated — so the workbook was reported
// clean while a sheet in it could not be opened. openpyxl, the bash
// original's engine, cannot read it either. A check that cannot run must not
// report as passing.
func TestXlsxRejectsASheetWhoseWorksheetPartIsAbsent(t *testing.T) {
	parts := xlsxParts("Summary", "Detail")
	delete(parts, "xl/worksheets/sheet2.xml")
	path := writeZip(t, filepath.Join(t.TempDir(), "nosheet2.xlsx"), parts)

	res, err := Xlsx(path, nil)
	if err == nil {
		t.Fatalf("a workbook whose Detail worksheet part is missing validated clean: %+v", res)
	}
	if !strings.Contains(err.Error(), "sheet2.xml") || !strings.Contains(err.Error(), "Detail") {
		t.Errorf("error %q should name both the missing part and the sheet it belongs to", err)
	}
}

// The same defect with no rels part to resolve through: the conventional
// sheetN.xml fallback is what the loop is really searching for, so the
// conventional name must be held to the same standard.
func TestXlsxRejectsASheetMissingUnderConventionalNaming(t *testing.T) {
	parts := xlsxParts("Summary", "Detail")
	delete(parts, "xl/_rels/workbook.xml.rels")
	delete(parts, "xl/worksheets/sheet2.xml")
	path := writeZip(t, filepath.Join(t.TempDir(), "norels-missing.xlsx"), parts)

	if _, err := Xlsx(path, nil); err == nil {
		t.Error("a workbook whose second sheet has no part validated clean under the conventional-naming fallback")
	}
}
