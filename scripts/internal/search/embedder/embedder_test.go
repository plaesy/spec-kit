package embedder

import (
	"testing"

	"github.com/plaesy/spec-kit/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Type != "rembed" {
		t.Errorf("Type = %s, expected 'rembed'", cfg.Type)
	}
	if cfg.Model != "intfloat/multilingual-e5-base" {
		t.Errorf("Model = %s, expected 'intfloat/multilingual-e5-base'", cfg.Model)
	}
	if cfg.Dimension != 768 {
		t.Errorf("Dimension = %d, expected 768", cfg.Dimension)
	}
}

// Only rembed exists. Any other type must fail loudly rather than fall back.
func TestNewEmbedder_RejectsUnsupportedTypes(t *testing.T) {
	for _, typ := range []string{"onnx", "http", "openai", ""} {
		emb, err := NewEmbedder(&config.EmbedderConfig{Type: typ, Model: "x", Dimension: 768})
		if err == nil {
			t.Errorf("type %q: expected an error", typ)
		}
		if emb != nil {
			t.Errorf("type %q: expected a nil embedder on error", typ)
		}
	}
}

// Compile-time check that the only embedder satisfies the interface.
var _ Embedder = (*RembedEmbedder)(nil)
