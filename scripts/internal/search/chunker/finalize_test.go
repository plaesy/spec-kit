package chunker

import (
	"context"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/search/parser"
)

func chunkMarkdownDoc(t *testing.T, cfg Config, content string) []Chunk {
	t.Helper()
	chunker := NewStructureAwareChunker(cfg, &SimpleTokenCounter{})
	doc := parser.Document{ID: "docid-000001", Path: "notes.md", Content: content}
	chunks, err := chunker.Chunk(context.Background(), doc)
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	return chunks
}

// A document shorter than MinChunkTokens used to produce zero chunks, so it
// could never be found. It must produce one searchable chunk.
func TestChunk_ShortDocumentIsNotDropped(t *testing.T) {
	cfg := DefaultConfig() // MinChunkTokens = 100, far above this document
	chunks := chunkMarkdownDoc(t, cfg, "# Note\n\nOne short line about deployment.")

	if len(chunks) != 1 {
		t.Fatalf("got %d chunks, want 1", len(chunks))
	}
	if !strings.Contains(chunks[0].Content, "deployment") {
		t.Errorf("chunk content %q lost the document text", chunks[0].Content)
	}
}

// Each section used to number its chunks from 0, so two sections of the same
// document shared IDs and one overwrote the other in the stores.
func TestChunk_IDsAreUniqueAcrossSections(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MinChunkTokens = 5
	cfg.MaxChunkTokens = 40

	// Several paragraphs per section, so each section yields more than one chunk.
	para := "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu\n\n"
	body := strings.Repeat(para, 6)
	content := "# First\n\n" + body + "# Second\n\n" + body

	chunks := chunkMarkdownDoc(t, cfg, content)
	if len(chunks) < 4 {
		t.Fatalf("expected several chunks from two long sections, got %d", len(chunks))
	}
	seen := map[string]bool{}
	for _, c := range chunks {
		if seen[c.ID] {
			t.Fatalf("duplicate chunk ID %q: one chunk would overwrite another in the stores", c.ID)
		}
		seen[c.ID] = true
	}
}

// A tiny trailing section used to be dropped with its text. It must now be
// merged into the preceding chunk, so every word of the document is kept.
func TestChunk_SmallSectionIsMergedNotDropped(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MinChunkTokens = 20
	cfg.MaxChunkTokens = 200

	big := strings.Repeat("kubernetes deployment rolling update replica pods service ", 10)
	content := "# Main\n\n" + big + "\n\n# Tail\n\nfinal remark xylophone"

	chunks := chunkMarkdownDoc(t, cfg, content)
	var joined string
	for _, c := range chunks {
		joined += c.Content + "\n"
	}
	if !strings.Contains(joined, "xylophone") {
		t.Fatalf("tiny trailing section was dropped; chunks: %d", len(chunks))
	}
}
