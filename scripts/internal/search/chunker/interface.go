package chunker

import (
	"context"

	"github.com/plaesy/spec-kit/internal/search/parser"
)

// Chunk represents a document chunk with metadata.
type Chunk struct {
	ID          string // Unique chunk ID
	DocID       string // Parent document ID
	ParentID    string // Parent chunk ID (for parent-child retrieval)
	Content     string // Chunk text content
	TokenCount  int    // Token count, in the embedding model's tokenizer
	SectionPath string // Hierarchical section path (e.g., "Introduction > Overview")
	StartOffset int    // Start character offset in original document
	EndOffset   int    // End character offset in original document
	// StartLine and EndLine are the 1-based, end-inclusive source lines this
	// chunk covers. 0 means the source has no line numbers the chunk can be
	// mapped to (HTML, JSON, and plain text, whose content is extracted text).
	StartLine int
	EndLine   int
	Metadata  map[string]string // Additional metadata
}

// Chunker defines the interface for chunking documents.
type Chunker interface {
	// Chunk splits a document into chunks.
	Chunk(ctx context.Context, doc parser.Document) ([]Chunk, error)

	// Name returns the chunker name.
	Name() string

	// Config returns the chunker configuration.
	Config() Config
}

// Config holds chunker configuration.
type Config struct {
	OverlapTokens  int // Overlap tokens between chunks (default: 50)
	MinChunkTokens int // Minimum tokens for a valid chunk; smaller ones are merged (default: 100)
	MaxChunkTokens int // Hard cap per chunk, in model tokens (default: 256)
}

// DefaultConfig returns the default chunker configuration.
//
// MaxChunkTokens is 256. The e5 model takes at most 512 input tokens, including
// its [CLS] and [SEP] tokens and the "passage: " prefix, and a chunk over that
// limit is cut off silently when embedded. 256 measured better than 480 for
// hybrid and vector search on 101 labels.
func DefaultConfig() Config {
	return Config{
		OverlapTokens:  50,
		MinChunkTokens: 100,
		MaxChunkTokens: 256,
	}
}

// TokenCounter counts tokens for text. It must use the same tokenizer as the
// embedding model, or the cap does not mean what it says.
type TokenCounter interface {
	Count(text string) int
}
