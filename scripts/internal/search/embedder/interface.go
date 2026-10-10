package embedder

import (
	"context"
)

// Embedder defines the interface for generating embeddings.
type Embedder interface {
	// Embed generates embeddings for a batch of texts.
	Embed(ctx context.Context, texts []string) ([][]float32, error)

	// Dimension returns the embedding dimension.
	Dimension() int

	// ModelName returns the model name.
	ModelName() string

	// MaxBatchSize returns the maximum batch size.
	MaxBatchSize() int

	// MaxSeqLen returns the maximum sequence length.
	MaxSeqLen() int
}

// Config holds embedder configuration.
type Config struct {
	Type         string // "onnx" | "http"
	Model        string // Model name/path
	Dimension    int    // Embedding dimension
	BatchSize    int    // Batch size
	MaxSeqLen    int    // Maximum sequence length
	Device       string // "cpu" | "cuda"
	ModelPath    string // Local model path (for ONNX)
	HTTPEndpoint string // HTTP endpoint (for HTTP type)
}

// DefaultConfig returns default embedder configuration.
func DefaultConfig() Config {
	return Config{
		Type:      "onnx",
		Model:     "intfloat/multilingual-e5-base",
		Dimension: 768,
		BatchSize: 32,
		MaxSeqLen: 512,
		Device:    "cpu",
	}
}
