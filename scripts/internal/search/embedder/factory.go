package embedder

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
)

// NewEmbedder creates an embedder based on configuration.
func NewEmbedder(cfg *config.EmbedderConfig) (Embedder, error) {
	ec := Config{
		Type:         cfg.Type,
		Model:        cfg.Model,
		Dimension:    cfg.Dimension,
		BatchSize:    cfg.BatchSize,
		MaxSeqLen:    cfg.MaxSeqLen,
		Device:       cfg.Device,
		ModelPath:    cfg.ModelPath,
		HTTPEndpoint: cfg.HTTPEndpoint,
	}

	switch ec.Type {
	case "onnx":
		return NewONNXEmbedder(ec)
	case "http":
		return NewHTTPEmebedder(ec)
	default:
		return nil, fmt.Errorf("unsupported embedder type: %s", ec.Type)
	}
}

// HTTPEmebedder implements Embedder using an HTTP endpoint (e.g., TEI, Ollama).
type HTTPEmebedder struct {
	config Config
	client *http.Client
}

func NewHTTPEmebedder(cfg Config) (*HTTPEmebedder, error) {
	if cfg.HTTPEndpoint == "" {
		return nil, fmt.Errorf("http_endpoint required for HTTP embedder")
	}
	return &HTTPEmebedder{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (e *HTTPEmebedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	// Implementation for HTTP-based embedding (TEI, Ollama, etc.)
	// This would call the HTTP endpoint with the texts
	return nil, fmt.Errorf("HTTP embedder not yet implemented")
}

func (e *HTTPEmebedder) Dimension() int {
	return e.config.Dimension
}

func (e *HTTPEmebedder) ModelName() string {
	return e.config.Model
}

func (e *HTTPEmebedder) MaxBatchSize() int {
	return e.config.BatchSize
}

func (e *HTTPEmebedder) MaxSeqLen() int {
	return e.config.MaxSeqLen
}
