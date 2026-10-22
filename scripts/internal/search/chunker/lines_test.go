package chunker

import (
	"context"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/search/parser"
)

// linesFixture has known line numbers (1-based):
//
//	1 # Title
//	2
//	3 Alpha paragraph one.
//	4 continues here.
//	5
//	6 Beta paragraph.
//	7
//	8 ## Second
//	9
//	10 Gamma text.
const linesFixture = "# Title\n\nAlpha paragraph one.\ncontinues here.\n\nBeta paragraph.\n\n## Second\n\nGamma text.\n"

// noMergeConfig turns off small-chunk merging, so each section's chunks keep
// exactly the lines they came from.
func noMergeConfig() Config {
	cfg := DefaultConfig()
	cfg.MinChunkTokens = 0
	return cfg
}

func chunkDoc(t *testing.T, cfg Config, path, content string) []Chunk {
	t.Helper()
	c := NewStructureAwareChunker(cfg, &SimpleTokenCounter{})
	doc := parser.Document{
		ID:       "abcdef1234567890",
		Path:     path,
		Content:  content,
		Metadata: map[string]string{"source": "markdown"},
	}
	chunks, err := c.Chunk(context.Background(), doc)
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	return chunks
}

func TestChunk_MarkdownLineRanges(t *testing.T) {
	chunks := chunkDoc(t, noMergeConfig(), "a.md", linesFixture)
	want := []struct{ start, end int }{
		{3, 6},   // the Title section: both paragraphs
		{10, 10}, // the Second section
	}
	if len(chunks) != len(want) {
		t.Fatalf("got %d chunks, want %d: %+v", len(chunks), len(want), chunks)
	}
	for i, w := range want {
		if chunks[i].StartLine != w.start || chunks[i].EndLine != w.end {
			t.Errorf("chunk %d lines = %d-%d, want %d-%d", i, chunks[i].StartLine, chunks[i].EndLine, w.start, w.end)
		}
	}
}

func TestChunk_MarkdownCRLFKeepsLineNumbers(t *testing.T) {
	crlf := strings.ReplaceAll(linesFixture, "\n", "\r\n")
	chunks := chunkDoc(t, noMergeConfig(), "a.md", crlf)
	if len(chunks) != 2 {
		t.Fatalf("got %d chunks, want 2", len(chunks))
	}
	if chunks[0].StartLine != 3 || chunks[0].EndLine != 6 {
		t.Errorf("CRLF chunk 0 lines = %d-%d, want 3-6", chunks[0].StartLine, chunks[0].EndLine)
	}
	if chunks[1].StartLine != 10 || chunks[1].EndLine != 10 {
		t.Errorf("CRLF chunk 1 lines = %d-%d, want 10-10", chunks[1].StartLine, chunks[1].EndLine)
	}
}

// TestChunk_LineRangesStayInsideTheDocument splits a long section into many
// chunks, and checks that every chunk's range is valid and inside the file.
func TestChunk_LineRangesStayInsideTheDocument(t *testing.T) {
	var b strings.Builder
	b.WriteString("# Long\n\n")
	for i := 0; i < 40; i++ {
		b.WriteString("Paragraph number ")
		b.WriteString(strings.Repeat("word ", 30))
		b.WriteString("\n\n")
	}
	content := b.String()
	lines := strings.Count(strings.TrimRight(content, "\n"), "\n") + 1

	cfg := noMergeConfig()
	cfg.MaxChunkTokens = 60
	chunks := chunkDoc(t, cfg, "long.md", content)
	if len(chunks) < 2 {
		t.Fatalf("expected the section to split into several chunks, got %d", len(chunks))
	}
	for i, ch := range chunks {
		if ch.StartLine < 1 || ch.EndLine < ch.StartLine || ch.EndLine > lines {
			t.Errorf("chunk %d lines %d-%d not valid for a %d-line file", i, ch.StartLine, ch.EndLine, lines)
		}
	}
}

func TestChunk_PlainTextHasNoLineRange(t *testing.T) {
	chunks := chunkDoc(t, noMergeConfig(), "notes.txt", "one\n\ntwo\n")
	if len(chunks) == 0 {
		t.Fatal("got no chunks")
	}
	for i, ch := range chunks {
		if ch.StartLine != 0 || ch.EndLine != 0 {
			t.Errorf("plain-text chunk %d lines = %d-%d, want 0-0 (extracted text has no source lines)", i, ch.StartLine, ch.EndLine)
		}
	}
}

func TestSplitUnits_LinesAndBlankGroups(t *testing.T) {
	units := splitUnits("\n  first\n  line\n\n\nsecond", 5)
	if len(units) != 2 {
		t.Fatalf("got %d units, want 2: %+v", len(units), units)
	}
	if units[0].text != "first\n  line" || units[0].first != 6 || units[0].last != 7 {
		t.Errorf("unit 0 = %+v, want text %q lines 6-7", units[0], "first\n  line")
	}
	if units[1].text != "second" || units[1].first != 10 || units[1].last != 10 {
		t.Errorf("unit 1 = %+v, want text %q line 10", units[1], "second")
	}
}
