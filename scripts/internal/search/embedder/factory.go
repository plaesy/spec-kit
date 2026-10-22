package embedder

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/config"
)

// NewEmbedder builds the embedder named by cfg.Type. "rembed" is the only type:
// it runs in pure Go and needs no server or separate install.
func NewEmbedder(cfg *config.EmbedderConfig) (Embedder, error) {
	switch cfg.Type {
	case "rembed":
		// Explicit nil check: returning a nil *RembedEmbedder as the interface
		// would yield a non-nil interface holding a nil pointer.
		e, err := NewRembedEmbedder(Config{
			Type:      cfg.Type,
			Model:     cfg.Model,
			Dimension: cfg.Dimension,
		})
		if err != nil {
			return nil, err
		}
		return e, nil
	default:
		return nil, fmt.Errorf("unsupported embedder type %q (supported: rembed)", cfg.Type)
	}
}
