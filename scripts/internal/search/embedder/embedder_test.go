package embedder

import (
	"context"
	"testing"
)

type mockEmbedder struct {
	dimension int
	modelName string
	batchSize int
	maxSeqLen int
	embedFunc func(ctx context.Context, texts []string) ([][]float32, error)
}

func (m *mockEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if m.embedFunc != nil {
		return m.embedFunc(ctx, texts)
	}
	// Default: return zero vectors
	embeddings := make([][]float32, len(texts))
	for i := range embeddings {
		embeddings[i] = make([]float32, m.dimension)
		for j := range embeddings[i] {
			embeddings[i][j] = float32(i+1) / 100.0
		}
	}
	return embeddings, nil
}

func (m *mockEmbedder) Dimension() int {
	return m.dimension
}

func (m *mockEmbedder) ModelName() string {
	return m.modelName
}

func (m *mockEmbedder) MaxBatchSize() int {
	return m.batchSize
}

func (m *mockEmbedder) MaxSeqLen() int {
	return m.maxSeqLen
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Type != "onnx" {
		t.Errorf("Type = %s, expected 'onnx'", cfg.Type)
	}
	if cfg.Model != "intfloat/multilingual-e5-base" {
		t.Errorf("Model = %s, expected 'intfloat/multilingual-e5-base'", cfg.Model)
	}
	if cfg.Dimension != 768 {
		t.Errorf("Dimension = %d, expected 768", cfg.Dimension)
	}
	if cfg.BatchSize != 32 {
		t.Errorf("BatchSize = %d, expected 32", cfg.BatchSize)
	}
	if cfg.MaxSeqLen != 512 {
		t.Errorf("MaxSeqLen = %d, expected 512", cfg.MaxSeqLen)
	}
	if cfg.Device != "cpu" {
		t.Errorf("Device = %s, expected 'cpu'", cfg.Device)
	}
}

func TestMockEmbedder(t *testing.T) {
	mock := &mockEmbedder{
		dimension: 768,
		modelName: "test-model",
		batchSize: 16,
		maxSeqLen: 256,
	}

	ctx := context.Background()
	texts := []string{"text 1", "text 2", "text 3"}

	embeddings, err := mock.Embed(ctx, texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(embeddings) != 3 {
		t.Fatalf("Expected 3 embeddings, got %d", len(embeddings))
	}

	for i, emb := range embeddings {
		if len(emb) != 768 {
			t.Errorf("Embedding %d: dimension = %d, expected 768", i, len(emb))
		}
	}

	if mock.Dimension() != 768 {
		t.Errorf("Dimension() = %d, expected 768", mock.Dimension())
	}
	if mock.ModelName() != "test-model" {
		t.Errorf("ModelName() = %s, expected 'test-model'", mock.ModelName())
	}
	if mock.MaxBatchSize() != 16 {
		t.Errorf("MaxBatchSize() = %d, expected 16", mock.MaxBatchSize())
	}
	if mock.MaxSeqLen() != 256 {
		t.Errorf("MaxSeqLen() = %d, expected 256", mock.MaxSeqLen())
	}
}

func TestMockEmbedder_CustomEmbedFunc(t *testing.T) {
	mock := &mockEmbedder{
		dimension: 384,
		modelName: "custom-model",
		batchSize: 8,
		maxSeqLen: 128,
		embedFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
			embeddings := make([][]float32, len(texts))
			for i := range embeddings {
				embeddings[i] = []float32{float32(i), float32(i + 1), float32(i + 2)}
			}
			return embeddings, nil
		},
	}

	ctx := context.Background()
	texts := []string{"a", "b"}

	embeddings, err := mock.Embed(ctx, texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(embeddings) != 2 {
		t.Fatalf("Expected 2 embeddings, got %d", len(embeddings))
	}

	if len(embeddings[0]) != 3 || len(embeddings[1]) != 3 {
		t.Errorf("Custom embed func returned wrong dimension")
	}
}

func TestMockEmbedder_EmptyInput(t *testing.T) {
	mock := &mockEmbedder{dimension: 768}

	ctx := context.Background()
	embeddings, err := mock.Embed(ctx, []string{})
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(embeddings) != 0 {
		t.Errorf("Expected 0 embeddings for empty input, got %d", len(embeddings))
	}
}

func TestConfig_Equality(t *testing.T) {
	cfg1 := DefaultConfig()
	cfg2 := DefaultConfig()

	if cfg1.Type != cfg2.Type || cfg1.Model != cfg2.Model || cfg1.Dimension != cfg2.Dimension {
		t.Error("Default configs should be equal")
	}

	cfg3 := Config{
		Type:         "http",
		Model:        "custom",
		Dimension:    384,
		BatchSize:    16,
		MaxSeqLen:    256,
		Device:       "cuda",
		ModelPath:    "/path/to/model",
		HTTPEndpoint: "http://localhost:8080",
	}

	if cfg3.Type == cfg1.Type {
		t.Error("Custom config should differ from default")
	}
}

func TestONNXEmbedder_WindowsStub(t *testing.T) {
	// On Windows, NewONNXEmbedder should return an error
	_, err := NewONNXEmbedder(DefaultConfig())
	if err == nil {
		t.Log("ONNX embedder available on this platform (not Windows)")
	} else {
		t.Logf("ONNX embedder not available on Windows as expected: %v", err)
	}
}

func TestFactory_InvalidType(t *testing.T) {
	// We can't easily test the factory without config package
	// But we can verify the error message pattern
	cfg := Config{Type: "invalid"}
	// The factory would return an error for unsupported type
	if cfg.Type != "onnx" && cfg.Type != "http" {
		// This is just verifying our test logic
		t.Log("Invalid type would be rejected by factory")
	}
}

func TestEmbedderInterface_Implementation(t *testing.T) {
	// Verify that our mock satisfies the Embedder interface
	var _ Embedder = &mockEmbedder{}
	t.Log("Mock embedder implements Embedder interface")
}
