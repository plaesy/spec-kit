//go:build windows

package embedder

import (
	"context"
	"fmt"
)

// NewONNXEmbedder creates a new ONNX embedder.
// On Windows, ONNX Runtime is not available via onnxruntime_go.
// Use the HTTP embedder type instead with a local TEI/Ollama server.
func NewONNXEmbedder(cfg Config) (*ONNXEmbedder, error) {
	return nil, fmt.Errorf("ONNX embedder not available on Windows. Use embedder type 'http' with a local TEI/Ollama server, or run on Linux/macOS")
}

// ONNXEmbedder is a stub on Windows.
type ONNXEmbedder struct{}

func (e *ONNXEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, fmt.Errorf("ONNX embedder not available on Windows")
}

func (e *ONNXEmbedder) Dimension() int {
	return 768
}

func (e *ONNXEmbedder) ModelName() string {
	return "intfloat/multilingual-e5-base (unavailable on Windows)"
}

func (e *ONNXEmbedder) MaxBatchSize() int {
	return 32
}

func (e *ONNXEmbedder) MaxSeqLen() int {
	return 512
}
