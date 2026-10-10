package parser

import (
	"context"
	"os"
	"path/filepath"

	"github.com/plaesy/spec-kit/internal/config"
)

// NewDefaultRegistry creates a parser registry with all default parsers registered.
func NewDefaultRegistry(cfg *config.SearchConfig) *ParserRegistry {
	registry := NewParserRegistry()

	// Register default parsers
	registry.Register(NewMarkdownParser())
	registry.Register(NewHTMLParser())
	registry.Register(NewJSONParser())

	return registry
}

// ParseDirectory parses all supported files in a directory recursively.
func (r *ParserRegistry) ParseDirectory(ctx context.Context, rootPath string, extensions []string) ([]Document, error) {
	var allDocs []Document

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := getExtension(path)
		if len(extensions) > 0 {
			matched := false
			for _, e := range extensions {
				if e == ext {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}

		if _, ok := r.Get(ext); !ok {
			return nil
		}

		docs, err := r.ParseFile(ctx, path)
		if err != nil {
			// Log error but continue with other files
			return nil
		}
		allDocs = append(allDocs, docs...)
		return nil
	})

	return allDocs, err
}
