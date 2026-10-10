package config

// UnifiedConfig is the root configuration structure for all plaesy capabilities.
type UnifiedConfig struct {
	Version int            `json:"version"`
	Reach   *ReachConfig   `json:"reach,omitempty"`
	Search  *SearchConfig  `json:"search,omitempty"`
	Analyze *AnalyzeConfig `json:"analyze,omitempty"`
	Graph   *GraphConfig   `json:"graph,omitempty"`
}

// ReachConfig holds reach capability configuration.
type ReachConfig struct {
	CacheDir        string                         `json:"cache_dir,omitempty"`
	CacheSizeMB     int                            `json:"cache_size_mb,omitempty"`
	CacheMemorySize int                            `json:"cache_memory_size,omitempty"`
	DefaultCacheTTL string                         `json:"default_cache_ttl,omitempty"`
	Platforms       map[string]ReachPlatformConfig `json:"platforms,omitempty"`
	RateLimits      map[string]RateLimitConfig     `json:"rate_limits,omitempty"`
	Timeouts        map[string]string              `json:"timeouts,omitempty"`
	RetryConfig     *RetryConfig                   `json:"retry_config,omitempty"`
}

// ReachPlatformConfig holds per-platform reach configuration.
type ReachPlatformConfig struct {
	Enabled        bool              `json:"enabled"`
	Backends       []string          `json:"backends,omitempty"`
	DefaultBackend string            `json:"default_backend,omitempty"`
	Credentials    map[string]string `json:"credentials,omitempty"`
	Options        map[string]string `json:"options,omitempty"`
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	Requests int    `json:"requests"`
	Window   string `json:"window"`
	Burst    int    `json:"burst,omitempty"`
}

// RetryConfig holds retry configuration.
type RetryConfig struct {
	MaxRetries      int      `json:"max_retries"`
	BaseDelay       string   `json:"base_delay"`
	MaxDelay        string   `json:"max_delay"`
	RetryableErrors []string `json:"retryable_errors,omitempty"`
}

// SearchConfig holds search capability configuration.
type SearchConfig struct {
	Embedder    EmbedderConfig    `json:"embedder"`
	VectorStore VectorStoreConfig `json:"vector_store"`
	TextIndex   TextIndexConfig   `json:"text_index"`
	Chunker     ChunkerConfig     `json:"chunker"`
	Indexer     IndexerConfig     `json:"indexer"`
	Retriever   RetrieverConfig   `json:"retriever"`
	Reranker    RerankerConfig    `json:"reranker"`
}

// EmbedderConfig holds embedding model configuration.
type EmbedderConfig struct {
	Type         string `json:"type"`        // "onnx" | "http"
	Model        string `json:"model"`       // "intfloat/multilingual-e5-base"
	Dimension    int    `json:"dimension"`   // 768
	BatchSize    int    `json:"batch_size"`  // 32
	MaxSeqLen    int    `json:"max_seq_len"` // 512
	Device       string `json:"device"`      // "cpu"
	ModelPath    string `json:"model_path,omitempty"`
	HTTPEndpoint string `json:"http_endpoint,omitempty"`
}

// VectorStoreConfig holds vector store configuration.
type VectorStoreConfig struct {
	Type string     `json:"type"` // "chromem"
	HNSW HNSWConfig `json:"hnsw"`
}

// HNSWConfig holds HNSW index parameters.
type HNSWConfig struct {
	M              int `json:"m"`               // 16
	EFConstruction int `json:"ef_construction"` // 200
	EFSearch       int `json:"ef_search"`       // 100
}

// TextIndexConfig holds full-text search configuration.
type TextIndexConfig struct {
	Type string `json:"type"` // "bleve"
}

// ChunkerConfig holds document chunking configuration.
type ChunkerConfig struct {
	TargetTokens   int `json:"target_tokens"`    // 350
	OverlapTokens  int `json:"overlap_tokens"`   // 50
	MinChunkTokens int `json:"min_chunk_tokens"` // 100
}

// IndexerConfig holds indexing pipeline configuration.
type IndexerConfig struct {
	Workers int  `json:"workers"` // 4
	Resume  bool `json:"resume"`  // true
}

// RetrieverConfig holds hybrid retrieval configuration.
type RetrieverConfig struct {
	Hybrid      HybridConfig      `json:"hybrid"`
	TopK        int               `json:"top_k"` // 50
	ParentChild ParentChildConfig `json:"parent_child"`
}

// HybridConfig holds RRF hybrid search configuration.
type HybridConfig struct {
	VectorWeight float64 `json:"vector_weight"` // 0.6
	BM25Weight   float64 `json:"bm25_weight"`   // 0.4
	RRFK         int     `json:"rrf_k"`         // 60
}

// ParentChildConfig holds parent-child retrieval configuration.
type ParentChildConfig struct {
	Enabled      bool `json:"enabled"`       // true
	ParentTokens int  `json:"parent_tokens"` // 1000
}

// RerankerConfig holds cross-encoder reranker configuration.
type RerankerConfig struct {
	Enabled bool   `json:"enabled"` // false
	Model   string `json:"model"`   // "jina-reranker-v2-base-multilingual"
	TopN    int    `json:"top_n"`   // 10
}

// AnalyzeConfig holds analyze capability configuration.
type AnalyzeConfig struct {
	AutoIndex bool `json:"auto_index"` // true
	Workers   int  `json:"workers"`    // 4
}

// GraphConfig holds graph capability configuration.
type GraphConfig struct {
	IncludeExtensions []string `json:"include_extensions,omitempty"`
	ExcludePatterns   []string `json:"exclude_patterns,omitempty"`
	MaxFileSize       int64    `json:"max_file_size,omitempty"`
}

// DefaultUnifiedConfig returns a default unified configuration.
func DefaultUnifiedConfig() *UnifiedConfig {
	return &UnifiedConfig{
		Version: 1,
		Reach:   DefaultReachConfig(),
		Search:  DefaultSearchConfig(),
		Analyze: DefaultAnalyzeConfig(),
		Graph:   DefaultGraphConfig(),
	}
}

// DefaultReachConfig returns default reach configuration.
func DefaultReachConfig() *ReachConfig {
	return &ReachConfig{
		CacheDir:        ".plaesy/reach/cache",
		CacheSizeMB:     100,
		CacheMemorySize: 1000,
		DefaultCacheTTL: "24h",
		Platforms:       make(map[string]ReachPlatformConfig),
		RateLimits:      DefaultRateLimits(),
		Timeouts:        DefaultTimeouts(),
		RetryConfig:     DefaultRetryConfig(),
	}
}

// DefaultRateLimits returns default per-platform rate limits.
func DefaultRateLimits() map[string]RateLimitConfig {
	return map[string]RateLimitConfig{
		"github":   {Requests: 100, Window: "1m", Burst: 10},
		"web":      {Requests: 20, Window: "1m", Burst: 5},
		"youtube":  {Requests: 10, Window: "1m", Burst: 2},
		"rss":      {Requests: 5, Window: "1m", Burst: 1},
		"twitter":  {Requests: 3, Window: "1m", Burst: 1},
		"reddit":   {Requests: 5, Window: "1m", Burst: 1},
		"bilibili": {Requests: 10, Window: "1m", Burst: 2},
		"exa":      {Requests: 10, Window: "1m", Burst: 2},
		"v2ex":     {Requests: 20, Window: "1m", Burst: 5},
		"xueqiu":   {Requests: 20, Window: "1m", Burst: 5},
	}
}

// DefaultTimeouts returns default per-platform timeouts.
func DefaultTimeouts() map[string]string {
	return map[string]string{
		"github":   "30s",
		"web":      "15s",
		"youtube":  "60s",
		"rss":      "10s",
		"twitter":  "15s",
		"reddit":   "15s",
		"bilibili": "15s",
		"exa":      "30s",
		"v2ex":     "10s",
		"xueqiu":   "10s",
	}
}

// DefaultRetryConfig returns default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  "500ms",
		MaxDelay:   "10s",
		RetryableErrors: []string{
			"timeout",
			"connection refused",
			"temporary failure",
			"rate limit",
			"503",
			"504",
		},
	}
}

// DefaultSearchConfig returns default search configuration.
func DefaultSearchConfig() *SearchConfig {
	return &SearchConfig{
		Embedder: EmbedderConfig{
			Type:      "onnx",
			Model:     "intfloat/multilingual-e5-base",
			Dimension: 768,
			BatchSize: 32,
			MaxSeqLen: 512,
			Device:    "cpu",
		},
		VectorStore: VectorStoreConfig{
			Type: "chromem",
			HNSW: HNSWConfig{
				M:              16,
				EFConstruction: 200,
				EFSearch:       100,
			},
		},
		TextIndex: TextIndexConfig{
			Type: "bleve",
		},
		Chunker: ChunkerConfig{
			TargetTokens:   350,
			OverlapTokens:  50,
			MinChunkTokens: 100,
		},
		Indexer: IndexerConfig{
			Workers: 4,
			Resume:  true,
		},
		Retriever: RetrieverConfig{
			Hybrid: HybridConfig{
				VectorWeight: 0.6,
				BM25Weight:   0.4,
				RRFK:         60,
			},
			TopK: 50,
			ParentChild: ParentChildConfig{
				Enabled:      true,
				ParentTokens: 1000,
			},
		},
		Reranker: RerankerConfig{
			Enabled: false,
			Model:   "jina-reranker-v2-base-multilingual",
			TopN:    10,
		},
	}
}

// DefaultAnalyzeConfig returns default analyze configuration.
func DefaultAnalyzeConfig() *AnalyzeConfig {
	return &AnalyzeConfig{
		AutoIndex: true,
		Workers:   4,
	}
}

// DefaultGraphConfig returns default graph configuration.
func DefaultGraphConfig() *GraphConfig {
	return &GraphConfig{
		IncludeExtensions: []string{
			".go", ".js", ".jsx", ".ts", ".tsx", ".py",
			".rs", ".rb", ".java", ".kt", ".kts", ".cs",
			".c", ".h", ".cpp", ".hpp", ".cc", ".hh", ".cxx",
			".swift", ".php", ".dart", ".sh", ".ps1",
			".md", ".mdx", ".txt",
		},
		ExcludePatterns: []string{
			"vendor/**", "node_modules/**", ".git/**", "dist/**", "build/**",
			"**/*.min.js", "**/*.min.css", "**/testdata/**", "**/vendor/**",
		},
		MaxFileSize: 10 * 1024 * 1024, // 10MB
	}
}
