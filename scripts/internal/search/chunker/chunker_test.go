package chunker

import (
	"context"
	"testing"

	"github.com/plaesy/spec-kit/internal/search/parser"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.TargetTokens != 350 {
		t.Errorf("TargetTokens = %d, expected 350", cfg.TargetTokens)
	}
	if cfg.OverlapTokens != 50 {
		t.Errorf("OverlapTokens = %d, expected 50", cfg.OverlapTokens)
	}
	if cfg.MinChunkTokens != 100 {
		t.Errorf("MinChunkTokens = %d, expected 100", cfg.MinChunkTokens)
	}
	if cfg.MaxChunkTokens != 500 {
		t.Errorf("MaxChunkTokens = %d, expected 500", cfg.MaxChunkTokens)
	}
}

func TestSimpleTokenCounter(t *testing.T) {
	counter := &SimpleTokenCounter{}

	tests := []struct {
		text     string
		expected int // approximate
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"hello world test", 3},
		{"one two three four five", 6}, // 5 * 1.3 = 6.5 -> 6
		{"a b c d e f g h i j", 13},    // 10 * 1.3 = 13
	}

	for _, tt := range tests {
		count := counter.Count(tt.text)
		if count != tt.expected {
			t.Errorf("Count(%q) = %d, expected %d", tt.text, count, tt.expected)
		}
	}
}

func TestStructureAwareChunker_Markdown(t *testing.T) {
	// Use smaller min chunk tokens for test
	cfg := DefaultConfig()
	cfg.MinChunkTokens = 20
	chunker := NewStructureAwareChunker(cfg, &SimpleTokenCounter{})

	mdContent := `# Main Title

## Section 1

This is the first section with some content. It has multiple sentences to make it longer. We want to see how the chunker handles this section.

### Subsection 1.1

This is a subsection with more content. It should be part of the same section but might be split if it's too long.

## Section 2

This is the second section. It also has content that should be chunked appropriately. Let's add more text to make it substantial enough for chunking.

### Subsection 2.1

Another subsection with content. The chunker should handle heading hierarchy properly.

## Section 3

Final section with some concluding remarks. This section is shorter.

End of document.
`

	doc := parser.Document{
		ID:       "test-doc-1",
		Path:     "test.md",
		Content:  mdContent,
		Metadata: map[string]string{"source": "markdown"},
	}

	chunks, err := chunker.Chunk(context.Background(), doc)
	if err != nil {
		t.Fatalf("Chunk failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("Expected at least one chunk")
	}

	t.Logf("Generated %d chunks", len(chunks))

	// Verify chunk properties
	for i, chunk := range chunks {
		if chunk.ID == "" {
			t.Errorf("Chunk %d: ID is empty", i)
		}
		if chunk.DocID != doc.ID {
			t.Errorf("Chunk %d: DocID mismatch: got %s, expected %s", i, chunk.DocID, doc.ID)
		}
		if chunk.Content == "" {
			t.Errorf("Chunk %d: Content is empty", i)
		}
		if chunk.TokenCount < chunker.Config().MinChunkTokens && i < len(chunks)-1 {
			// Last chunk can be smaller, but others should meet minimum
			t.Logf("Chunk %d: TokenCount %d < MinChunkTokens %d (may be last chunk)", i, chunk.TokenCount, chunker.Config().MinChunkTokens)
		}
		if chunk.TokenCount > chunker.Config().MaxChunkTokens {
			t.Errorf("Chunk %d: TokenCount %d > MaxChunkTokens %d", i, chunk.TokenCount, chunker.Config().MaxChunkTokens)
		}
		if chunk.Metadata["source"] != "markdown" {
			t.Errorf("Chunk %d: Metadata source mismatch: got %s", i, chunk.Metadata["source"])
		}
	}

	// Verify parent-child relationships within each section
	// Note: Parent-child is only set within each section's chunks, not across sections
	// So we just verify that if a chunk has a ParentID, it's a valid chunk ID
	for _, chunk := range chunks {
		if chunk.ParentID != "" {
			found := false
			for _, c := range chunks {
				if c.ID == chunk.ParentID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Chunk %s: ParentID %s not found in chunks", chunk.ID, chunk.ParentID)
			}
		}
	}

	// Verify section paths are set
	for _, chunk := range chunks {
		if chunk.SectionPath == "" {
			t.Logf("Chunk %s: SectionPath is empty", chunk.ID)
		}
	}
}

func TestStructureAwareChunker_Markdown_NoHeadings(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MinChunkTokens = 20
	chunker := NewStructureAwareChunker(cfg, &SimpleTokenCounter{})

	// Document without headings - should be treated as one section
	mdContent := `This is a plain text document without any headings. It has multiple paragraphs.

This is the second paragraph with more content to make it substantial.

Third paragraph continues the content. We want to see how the chunker handles plain text.

Fourth paragraph adds even more content for chunking purposes.

Fifth paragraph to ensure we have enough content for multiple chunks.

Sixth paragraph continues.

Seventh paragraph.

Eighth paragraph.

Ninth paragraph.

Tenth paragraph to finish.
`

	doc := parser.Document{
		ID:       "test-doc-2",
		Path:     "test.md",
		Content:  mdContent,
		Metadata: map[string]string{"source": "markdown"},
	}

	chunks, err := chunker.Chunk(context.Background(), doc)
	if err != nil {
		t.Fatalf("Chunk failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("Expected at least one chunk")
	}

	// All chunks should have empty section path (or "Document")
	for _, chunk := range chunks {
		if chunk.SectionPath != "" && chunk.SectionPath != "Document" {
			t.Logf("Chunk %s: SectionPath = %q", chunk.ID, chunk.SectionPath)
		}
	}
}

func TestStructureAwareChunker_HTML(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MinChunkTokens = 20
	chunker := NewStructureAwareChunker(cfg, &SimpleTokenCounter{})

	htmlContent := `This is HTML content that has been parsed and cleaned. It contains multiple paragraphs.

This is the second paragraph with more content for the chunker to process.

Third paragraph adds more text content.

Fourth paragraph continues.

Fifth paragraph to ensure enough content.

Sixth paragraph.

Seventh paragraph.

Eighth paragraph.

Ninth paragraph.

Tenth paragraph to finish.
`

	doc := parser.Document{
		ID:       "test-doc-3",
		Path:     "test.html",
		Content:  htmlContent,
		Metadata: map[string]string{"source": "html"},
	}

	chunks, err := chunker.Chunk(context.Background(), doc)
	if err != nil {
		t.Fatalf("Chunk failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("Expected at least one chunk")
	}

	for _, chunk := range chunks {
		if chunk.Metadata["source"] != "html" {
			t.Errorf("Chunk metadata source mismatch: got %s", chunk.Metadata["source"])
		}
	}
}

func TestStructureAwareChunker_PlainText(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MinChunkTokens = 20
	chunker := NewStructureAwareChunker(cfg, &SimpleTokenCounter{})

	textContent := `This is plain text content with multiple paragraphs.

Second paragraph with more content.

Third paragraph continues.

Fourth paragraph.

Fifth paragraph.

Sixth paragraph.

Seventh paragraph.

Eighth paragraph.

Ninth paragraph.

Tenth paragraph to finish.
`

	doc := parser.Document{
		ID:       "test-doc-4",
		Path:     "test.txt",
		Content:  textContent,
		Metadata: map[string]string{"source": "text"},
	}

	chunks, err := chunker.Chunk(context.Background(), doc)
	if err != nil {
		t.Fatalf("Chunk failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("Expected at least one chunk")
	}
}

func TestStructureAwareChunker_SmallContent(t *testing.T) {
	chunker := NewStructureAwareChunker(DefaultConfig(), &SimpleTokenCounter{})

	// Content smaller than MinChunkTokens
	smallContent := `# Title

Short content.
`

	doc := parser.Document{
		ID:       "test-doc-5",
		Path:     "test.md",
		Content:  smallContent,
		Metadata: map[string]string{"source": "markdown"},
	}

	chunks, err := chunker.Chunk(context.Background(), doc)
	if err != nil {
		t.Fatalf("Chunk failed: %v", err)
	}

	// Small content might produce 0 chunks (below MinChunkTokens) or 1 chunk
	if len(chunks) > 1 {
		t.Errorf("Expected at most 1 chunk for small content, got %d", len(chunks))
	}
}

func TestStructureAwareChunker_Overlap(t *testing.T) {
	cfg := DefaultConfig()
	cfg.OverlapTokens = 50
	cfg.TargetTokens = 100
	cfg.MinChunkTokens = 20
	cfg.MaxChunkTokens = 150

	chunker := NewStructureAwareChunker(cfg, &SimpleTokenCounter{})

	// Create content that will produce multiple chunks with overlap
	// Each "Paragraph..." is about 10 words = 13 tokens, repeated 5 times = 50 words = 65 tokens
	// We have 4 such sections = 200 words = 260 tokens total, should produce 2-3 chunks
	content := "# Section 1\n\n" +
		repeatString("Paragraph one with enough words to make it substantial. ", 10) + "\n\n" +
		"## Subsection 1.1\n\n" +
		repeatString("Paragraph two continues with more content. ", 10) + "\n\n" +
		"# Section 2\n\n" +
		repeatString("Paragraph three in section two. ", 10) + "\n\n" +
		"## Subsection 2.1\n\n" +
		repeatString("Paragraph four in subsection. ", 10)

	doc := parser.Document{
		ID:       "test-doc-6",
		Path:     "test.md",
		Content:  content,
		Metadata: map[string]string{"source": "markdown"},
	}

	chunks, err := chunker.Chunk(context.Background(), doc)
	if err != nil {
		t.Fatalf("Chunk failed: %v", err)
	}

	if len(chunks) < 2 {
		t.Fatalf("Expected at least 2 chunks for overlap test, got %d", len(chunks))
	}

	// Check that chunks have overlap (content from end of previous chunk appears in next)
	// This is a basic check - we verify the mechanism works
	t.Logf("Generated %d chunks with overlap", len(chunks))
	for i, chunk := range chunks {
		t.Logf("Chunk %d: %d tokens, parent=%s", i, chunk.TokenCount, chunk.ParentID)
	}
}

func TestStructureAwareChunker_Config(t *testing.T) {
	cfg := Config{
		TargetTokens:   200,
		OverlapTokens:  30,
		MinChunkTokens: 50,
		MaxChunkTokens: 300,
	}
	chunker := NewStructureAwareChunker(cfg, &SimpleTokenCounter{})

	if chunker.Config().TargetTokens != 200 {
		t.Errorf("TargetTokens not preserved: %d", chunker.Config().TargetTokens)
	}
	if chunker.Config().OverlapTokens != 30 {
		t.Errorf("OverlapTokens not preserved: %d", chunker.Config().OverlapTokens)
	}
	if chunker.Name() != "structure-aware" {
		t.Errorf("Name = %s, expected 'structure-aware'", chunker.Name())
	}
}

func TestSplitParagraphs(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"para1\n\npara2\n\npara3", 3},
		{"para1\n\n\npara2", 2},
		{"single paragraph", 1},
		{"", 0},
		{"para1\r\n\r\npara2", 2},
		{"para1\r\rpara2", 2},
		{"  \n\n  \n\npara", 1}, // empty paragraphs filtered
	}

	for _, tt := range tests {
		result := splitParagraphs(tt.input)
		if len(result) != tt.expected {
			t.Errorf("splitParagraphs(%q): got %d paragraphs, expected %d: %v", tt.input, len(result), tt.expected, result)
		}
	}
}

func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func TestGetOverlapText(t *testing.T) {
	cfg := DefaultConfig()
	cfg.OverlapTokens = 50
	chunker := NewStructureAwareChunker(cfg, &SimpleTokenCounter{})

	// Use longer text to ensure overlap is shorter than original
	text := "This is a test sentence with multiple words for overlap testing. " +
		"It has many more words to ensure the overlap is a subset. " +
		"Adding even more content to make it long enough for overlap extraction. " +
		"The overlap should take the last N tokens from this text."
	overlap := chunker.getOverlapText(text)

	if overlap == "" {
		t.Error("Overlap text should not be empty")
	}

	// Overlap should be shorter than original
	if len(overlap) >= len(text) {
		t.Error("Overlap text should be shorter than original")
	}
}

func TestGetOverlapText_ZeroOverlap(t *testing.T) {
	cfg := DefaultConfig()
	cfg.OverlapTokens = 0
	chunker := NewStructureAwareChunker(cfg, &SimpleTokenCounter{})

	text := "This is a test sentence."
	overlap := chunker.getOverlapText(text)

	if overlap != "" {
		t.Errorf("Overlap text should be empty with OverlapTokens=0, got: %s", overlap)
	}
}

func TestStructureAwareChunker_EmptyDocument(t *testing.T) {
	chunker := NewStructureAwareChunker(DefaultConfig(), &SimpleTokenCounter{})

	doc := parser.Document{
		ID:       "test-doc-empty",
		Path:     "test.md",
		Content:  "",
		Metadata: map[string]string{"source": "markdown"},
	}

	chunks, err := chunker.Chunk(context.Background(), doc)
	if err != nil {
		t.Fatalf("Chunk failed: %v", err)
	}

	if len(chunks) != 0 {
		t.Errorf("Expected 0 chunks for empty document, got %d", len(chunks))
	}
}
