package search

import (
	"github.com/plaesy/spec-kit/internal/config"
)

// SearchConfigManager manages search configuration using the unified config manager.
type SearchConfigManager struct {
	*config.ConfigManager[config.SearchConfig]
}

// NewSearchConfigManager creates a new search config manager.
func NewSearchConfigManager() *SearchConfigManager {
	return &SearchConfigManager{
		ConfigManager: config.NewConfigManager[config.SearchConfig](
			"", // Use default unified config path
			config.DefaultSearchConfig(),
		),
	}
}

// Load loads the search configuration from the unified config file.
func (scm *SearchConfigManager) Load() (*config.SearchConfig, error) {
	// First load the unified config
	unifiedMgr := config.NewConfigManager[config.UnifiedConfig](
		"", // Use default unified config path
		config.DefaultUnifiedConfig(),
	)
	unified, err := unifiedMgr.Load()
	if err != nil {
		return nil, err
	}

	// Return search section, or default if not present
	if unified.Search != nil {
		return unified.Search, nil
	}
	return config.DefaultSearchConfig(), nil
}

// Save saves the search configuration to the unified config file.
func (scm *SearchConfigManager) Save(cfg *config.SearchConfig) error {
	unifiedMgr := config.NewConfigManager[config.UnifiedConfig](
		"", // Use default unified config path
		config.DefaultUnifiedConfig(),
	)
	unified, err := unifiedMgr.Load()
	if err != nil {
		return err
	}

	unified.Search = cfg
	return unifiedMgr.Save(unified)
}

// SearchConfig aliases config.SearchConfig for backward compatibility.
type SearchConfig = config.SearchConfig

// EmbedderConfig aliases config.EmbedderConfig.
type EmbedderConfig = config.EmbedderConfig

// VectorStoreConfig aliases config.VectorStoreConfig.
type VectorStoreConfig = config.VectorStoreConfig

// HNSWConfig aliases config.HNSWConfig.
type HNSWConfig = config.HNSWConfig

// TextIndexConfig aliases config.TextIndexConfig.
type TextIndexConfig = config.TextIndexConfig

// ChunkerConfig aliases config.ChunkerConfig.
type ChunkerConfig = config.ChunkerConfig

// IndexerConfig aliases config.IndexerConfig.
type IndexerConfig = config.IndexerConfig

// RetrieverConfig aliases config.RetrieverConfig.
type RetrieverConfig = config.RetrieverConfig

// HybridConfig aliases config.HybridConfig.
type HybridConfig = config.HybridConfig

// ParentChildConfig aliases config.ParentChildConfig.
type ParentChildConfig = config.ParentChildConfig

// RerankerConfig aliases config.RerankerConfig.
type RerankerConfig = config.RerankerConfig

// DefaultSearchConfig returns the default search configuration.
var DefaultSearchConfig = config.DefaultSearchConfig
