package chunker

import (
	"context"

	"github.com/plaesy/spec-kit/internal/search/parser"
)

// Chunk represents a document chunk with metadata.
type Chunk struct {
	ID          string            // Unique chunk ID
	DocID       string            // Parent document ID
	ParentID    string            // Parent chunk ID (for parent-child retrieval)
	Content     string            // Chunk text content
	TokenCount  int               // Estimated token count
	SectionPath string            // Hierarchical section path (e.g., "Introduction > Overview")
	StartOffset int               // Start character offset in original document
	EndOffset   int               // End character offset in original document
	Metadata    map[string]string // Additional metadata
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
	TargetTokens   int // Target tokens per chunk (default: 350)
	OverlapTokens  int // Overlap tokens between chunks (default: 50)
	MinChunkTokens int // Minimum tokens for a valid chunk (default: 100)
	MaxChunkTokens int // Maximum tokens per chunk (default: 500)
}

// DefaultConfig returns the default chunker configuration.
func DefaultConfig() Config {
	return Config{
		TargetTokens:   350,
		OverlapTokens:  50,
		MinChunkTokens: 100,
		MaxChunkTokens: 500,
	}
}

// TokenCounter estimates token count for text.
type TokenCounter interface {
	Count(text string) int
}

// SimpleTokenCounter provides a simple token estimation (words * 1.3).
type SimpleTokenCounter struct{}

// Count estimates token count from text.
func (c *SimpleTokenCounter) Count(text string) int {
	words := len(splitWords(text))
	return int(float64(words) * 1.3)
}

// splitWords splits text into words.
func splitWords(text string) []string {
	// Simple word splitting - can be replaced with proper tokenizer
	return splitByWhitespace(text)
}

func splitByWhitespace(text string) []string {
	var words []string
	current := ""
	for _, r := range text {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}
