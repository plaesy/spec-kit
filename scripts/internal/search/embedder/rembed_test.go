package embedder

import (
	"context"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/config"
)

// These tests need no network. They cover the rejection paths, the factory's
// nil-interface handling, and the prefix choice per model family.

func TestNewRembedEmbedder_RequiresModel(t *testing.T) {
	if _, err := NewRembedEmbedder(Config{Type: "rembed"}); err == nil {
		t.Fatal("expected an error when no model is configured")
	}
}

func TestFactory_RembedWithoutModelReturnsNilInterface(t *testing.T) {
	emb, err := NewEmbedder(&config.EmbedderConfig{Type: "rembed"})
	if err == nil {
		t.Fatal("expected an error for a rembed config without a model")
	}
	// A typed nil pointer would make this interface non-nil and cause a panic
	// for any caller that skips the error check.
	if emb != nil {
		t.Fatalf("embedder = %T, want nil interface on error", emb)
	}
}

// e5 needs its prefixes; other models must not get any, or their vectors
// shift for no reason.
func TestPrefixesFor(t *testing.T) {
	tests := []struct {
		model, query, passage string
	}{
		{"intfloat/multilingual-e5-base", "query: ", "passage: "},
		{"intfloat/multilingual-e5-small", "query: ", "passage: "},
		{"sentence-transformers/all-MiniLM-L6-v2", "", ""},
	}
	for _, tt := range tests {
		q, p := prefixesFor(tt.model)
		if q != tt.query || p != tt.passage {
			t.Errorf("prefixesFor(%q) = (%q, %q), want (%q, %q)", tt.model, q, p, tt.query, tt.passage)
		}
	}
}

// TestRembed_RealModel downloads the default model on first run, so it only
// runs when PLAESY_TEST_REMBED=1. It checks the default config end to end:
// dimension, normalisation, and a cross-language similarity sanity check.
func TestRembed_RealModel(t *testing.T) {
	if os.Getenv("PLAESY_TEST_REMBED") != "1" {
		t.Skip("set PLAESY_TEST_REMBED=1 to load the real model (downloads on first run)")
	}

	cfg := config.DefaultSearchConfig().Embedder
	emb, err := NewEmbedder(&cfg)
	if err != nil {
		t.Fatalf("NewEmbedder: %v", err)
	}
	if c, ok := emb.(interface{ Close() error }); ok {
		defer c.Close()
	}

	ctx := context.Background()
	query, err := emb.EmbedQuery(ctx, "binary search tree")
	if err != nil {
		t.Fatalf("EmbedQuery: %v", err)
	}
	passages, err := emb.EmbedPassages(ctx, []string{
		"pohon pencarian biner",
		"kubernetes deployment manifests",
	})
	if err != nil {
		t.Fatalf("EmbedPassages: %v", err)
	}
	if len(passages) != 2 {
		t.Fatalf("got %d passage vectors, want 2", len(passages))
	}

	for i, v := range append([][]float32{query}, passages...) {
		if len(v) != 768 {
			t.Fatalf("vector %d has length %d, want 768", i, len(v))
		}
		if n := norm(v); math.Abs(n-1) > 1e-3 {
			t.Errorf("vector %d norm = %.4f, want ~1 (L2-normalised)", i, n)
		}
	}

	// The chunker sizes text with this tokenizer, so check it behaves as the
	// code assumes: a short phrase counts in single digits, and text far past
	// the model's limit saturates at its maximum rather than growing.
	tok, ok := emb.(Tokenizer)
	if !ok {
		t.Fatal("rembed embedder must implement Tokenizer")
	}
	if n := tok.CountTokens("binary search tree"); n < 3 || n > 12 {
		t.Errorf("CountTokens(short phrase) = %d, want a small count", n)
	}
	long := strings.Repeat("binary search tree insertion ", 400)
	if n := tok.CountTokens(long); n != 512 {
		t.Errorf("CountTokens(long text) = %d, want 512 (the model's maximum, which marks truncation)", n)
	}

	related := cosine(query, passages[0])
	unrelated := cosine(query, passages[1])
	if related <= unrelated {
		t.Errorf("cross-language similarity: related=%.3f <= unrelated=%.3f", related, unrelated)
	}
}

func norm(v []float32) float64 {
	var s float64
	for _, x := range v {
		s += float64(x) * float64(x)
	}
	return math.Sqrt(s)
}

func cosine(a, b []float32) float64 {
	var dot float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
	}
	return dot / (norm(a) * norm(b))
}
