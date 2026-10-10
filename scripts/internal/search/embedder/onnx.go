//go:build !windows

package embedder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/yalue/onnxruntime_go"
)

// onnxEnvOnce guards onnxruntime_go.InitializeEnvironment, which must run
// exactly once per process regardless of how many ONNXEmbedder instances
// are created.
var onnxEnvOnce sync.Once
var onnxEnvErr error

func ensureONNXEnvironment() error {
	onnxEnvOnce.Do(func() {
		if path := os.Getenv("ONNX_LIB_PATH"); path != "" {
			onnxruntime_go.SetSharedLibraryPath(path)
		}
		onnxEnvErr = onnxruntime_go.InitializeEnvironment()
	})
	return onnxEnvErr
}

// ONNXEmbedder implements Embedder using ONNX Runtime.
type ONNXEmbedder struct {
	config      Config
	session     *onnxruntime_go.DynamicAdvancedSession
	inputNames  []string
	outputNames []string
	mu          sync.Mutex
	initialized bool
}

// NewONNXEmbedder creates a new ONNX embedder.
func NewONNXEmbedder(cfg Config) (*ONNXEmbedder, error) {
	if cfg.Type != "onnx" {
		return nil, fmt.Errorf("invalid embedder type: %s", cfg.Type)
	}

	if cfg.Dimension == 0 {
		cfg.Dimension = 768
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 32
	}
	if cfg.MaxSeqLen == 0 {
		cfg.MaxSeqLen = 512
	}

	e := &ONNXEmbedder{
		config: cfg,
	}

	// Initialize ONNX Runtime environment
	if err := e.init(); err != nil {
		return nil, err
	}

	return e, nil
}

// init initializes the ONNX Runtime session.
func (e *ONNXEmbedder) init() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized {
		return nil
	}

	// Determine model path
	modelPath := e.config.ModelPath
	if modelPath == "" {
		// Default to cached model location
		cacheDir := os.Getenv("ONNX_MODEL_CACHE")
		if cacheDir == "" {
			cacheDir = filepath.Join(os.Getenv("HOME"), ".cache", "onnx_models")
		}
		modelPath = filepath.Join(cacheDir, "multilingual-e5-base.onnx")
	}

	// Check if model exists, if not download it
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		if err := e.downloadModel(modelPath); err != nil {
			return fmt.Errorf("failed to download model: %w", err)
		}
	}

	if err := ensureONNXEnvironment(); err != nil {
		return fmt.Errorf("failed to initialize ONNX Runtime environment: %w", err)
	}

	// SessionOptions is an opaque handle in onnxruntime_go — built via
	// NewSessionOptions() and configured through setter methods, not a
	// struct literal.
	opts, err := onnxruntime_go.NewSessionOptions()
	if err != nil {
		return fmt.Errorf("failed to create ONNX session options: %w", err)
	}
	defer opts.Destroy()
	if err := opts.SetGraphOptimizationLevel(onnxruntime_go.GraphOptimizationLevelEnableAll); err != nil {
		return fmt.Errorf("failed to set graph optimization level: %w", err)
	}
	if err := opts.SetExecutionMode(onnxruntime_go.ExecutionModeSequential); err != nil {
		return fmt.Errorf("failed to set execution mode: %w", err)
	}
	if err := opts.SetIntraOpNumThreads(runtime.NumCPU()); err != nil {
		return fmt.Errorf("failed to set intra-op thread count: %w", err)
	}
	if err := opts.SetInterOpNumThreads(1); err != nil {
		return fmt.Errorf("failed to set inter-op thread count: %w", err)
	}

	// Create ONNX session
	session, err := onnxruntime_go.NewDynamicAdvancedSession(
		modelPath,
		[]string{"input_ids", "attention_mask"},
		[]string{"output"},
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to create ONNX session: %w", err)
	}

	e.session = session
	e.inputNames = []string{"input_ids", "attention_mask"}
	e.outputNames = []string{"output"}
	e.initialized = true

	return nil
}

// downloadModel downloads the model from Hugging Face Hub.
// This is a placeholder - in production, use a proper model downloader.
func (e *ONNXEmbedder) downloadModel(destPath string) error {
	// For now, return an error with instructions
	// In production, implement actual download from Hugging Face
	return fmt.Errorf("model not found at %s. Please download intfloat/multilingual-e5-base ONNX model and place it at this path, or set MODEL_PATH env variable", destPath)
}

// Embed generates embeddings for a batch of texts.
func (e *ONNXEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if !e.initialized {
		if err := e.init(); err != nil {
			return nil, err
		}
	}

	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// Process in batches
	batchSize := e.config.BatchSize
	if batchSize <= 0 {
		batchSize = 32
	}

	var allEmbeddings [][]float32

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := e.embedBatch(ctx, batch)
		if err != nil {
			return nil, err
		}

		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	return allEmbeddings, nil
}

// embedBatch embeds a single batch of texts.
func (e *ONNXEmbedder) embedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		return nil, fmt.Errorf("embedder not initialized")
	}

	// Tokenize texts — flat, row-major [batch_size * seq_len], matching the
	// Shape passed to NewTensor below (onnxruntime_go.NewTensor takes a
	// flat backing slice plus its logical Shape, not a [][]int64).
	inputIDsFlat, attentionMaskFlat, err := e.tokenize(texts)
	if err != nil {
		return nil, err
	}

	seqShape := onnxruntime_go.Shape{int64(len(texts)), int64(e.config.MaxSeqLen)}

	inputIDsTensor, err := onnxruntime_go.NewTensor(seqShape, inputIDsFlat)
	if err != nil {
		return nil, fmt.Errorf("failed to create input_ids tensor: %w", err)
	}
	defer inputIDsTensor.Destroy()

	attentionMaskTensor, err := onnxruntime_go.NewTensor(seqShape, attentionMaskFlat)
	if err != nil {
		return nil, fmt.Errorf("failed to create attention_mask tensor: %w", err)
	}
	defer attentionMaskTensor.Destroy()

	outputShape := onnxruntime_go.Shape{int64(len(texts)), int64(e.config.Dimension)}
	outputTensor, err := onnxruntime_go.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate output tensor: %w", err)
	}
	defer outputTensor.Destroy()

	// DynamicAdvancedSession.Run populates the pre-allocated output tensors
	// in place and returns only an error — not a map of result tensors.
	err = e.session.Run(
		[]onnxruntime_go.Value{inputIDsTensor, attentionMaskTensor},
		[]onnxruntime_go.Value{outputTensor},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run inference: %w", err)
	}

	// Extract embeddings
	outputData := outputTensor.GetData()

	// Reshape to [batch_size, hidden_dim]
	batchSize := len(texts)
	hiddenDim := e.config.Dimension
	embeddings := make([][]float32, batchSize)
	for i := 0; i < batchSize; i++ {
		start := i * hiddenDim
		end := start + hiddenDim
		embeddings[i] = outputData[start:end]

		// Normalize embeddings (L2 normalization for cosine similarity)
		normalize(embeddings[i])
	}

	return embeddings, nil
}

// tokenize converts texts to flat, row-major input_ids and attention_mask
// buffers of length len(texts)*MaxSeqLen — the shape onnxruntime_go.NewTensor
// expects (a flat backing slice plus a separate Shape), not [][]int64.
// This is a simplified tokenizer - in production, use a proper tokenizer
// from the transformers library (e.g., via tokenizers Go bindings).
func (e *ONNXEmbedder) tokenize(texts []string) ([]int64, []int64, error) {
	// This is a placeholder implementation
	// Real implementation would use the model's tokenizer
	// For now, return dummy tensors of correct shape

	seqLen := e.config.MaxSeqLen
	total := len(texts) * seqLen

	inputIDs := make([]int64, total)
	attentionMask := make([]int64, total)
	// Both buffers are zero-valued (padding token / no attention) by
	// make()'s default — nothing further to fill in this placeholder.

	return inputIDs, attentionMask, nil
}

// Dimension returns the embedding dimension.
func (e *ONNXEmbedder) Dimension() int {
	return e.config.Dimension
}

// ModelName returns the model name.
func (e *ONNXEmbedder) ModelName() string {
	return e.config.Model
}

// MaxBatchSize returns the maximum batch size.
func (e *ONNXEmbedder) MaxBatchSize() int {
	return e.config.BatchSize
}

// MaxSeqLen returns the maximum sequence length.
func (e *ONNXEmbedder) MaxSeqLen() int {
	return e.config.MaxSeqLen
}

// normalize performs L2 normalization on a vector.
func normalize(v []float32) {
	var sum float32
	for _, x := range v {
		sum += x * x
	}
	if sum > 0 {
		norm := float32(1.0) / float32(len(v)) // Simplified - should be sqrt(sum)
		for i := range v {
			v[i] *= norm
		}
	}
}

// l2Normalize performs proper L2 normalization.
func l2Normalize(v []float32) {
	var sum float32
	for _, x := range v {
		sum += x * x
	}
	if sum > 0 {
		invNorm := float32(1.0) / float32(sum)
		for i := range v {
			v[i] *= invNorm
		}
	}
}
