package embedder

import (
	"context"
)

// Embedder turns text into vectors. Search queries and indexed passages are
// separate methods on purpose: retrieval models such as e5 expect a different
// prefix for each, and a caller that mixes them up degrades results silently.
type Embedder interface {
	// EmbedQuery embeds a search query.
	EmbedQuery(ctx context.Context, query string) ([]float32, error)

	// EmbedPassages embeds documents to be indexed, one vector per passage.
	EmbedPassages(ctx context.Context, passages []string) ([][]float32, error)
}

// Tokenizer counts tokens with the embedding model's own tokenizer. Chunk sizes
// must be measured this way, because the model truncates input at its limit.
type Tokenizer interface {
	// CountTokens returns the number of model tokens in text, including the
	// special tokens the model adds.
	CountTokens(text string) int
}

// Config holds embedder configuration.
type Config struct {
	Type      string // "rembed"
	Model     string // Hugging Face model id
	Dimension int    // Expected vector size; checked against the loaded model
}

// DefaultConfig returns default embedder configuration.
func DefaultConfig() Config {
	return Config{
		Type:      "rembed",
		Model:     "intfloat/multilingual-e5-base",
		Dimension: 768,
	}
}
