package parser

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MarkdownParser parses Markdown files.
type MarkdownParser struct{}

// NewMarkdownParser creates a new Markdown parser.
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{}
}

// Name returns the parser name.
func (p *MarkdownParser) Name() string {
	return "markdown"
}

// SupportedExtensions returns the supported file extensions.
func (p *MarkdownParser) SupportedExtensions() []string {
	return []string{".md", ".markdown", ".mdx"}
}

// Parse parses a Markdown file and returns a document.
func (p *MarkdownParser) Parse(ctx context.Context, path string) ([]Document, error) {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, &ParseError{
			Code:    "READ_ERROR",
			Message: "failed to read file",
			Err:     err,
		}
	}

	// Keep the source text. Content must keep its line structure, because the
	// chunker reports each chunk's line range against it. Rendering to HTML and
	// stripping tags would lose both the headings and the line numbers.
	text := string(content)

	// Compute hash
	hash := sha256.Sum256(content)
	hashStr := fmt.Sprintf("%x", hash)

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, &ParseError{
			Code:    "STAT_ERROR",
			Message: "failed to stat file",
			Err:     err,
		}
	}

	doc := Document{
		ID:       hashStr[:16],
		Path:     path,
		Content:  text,
		Metadata: map[string]string{"source": "markdown"},
		ModTime:  info.ModTime(),
		Hash:     hashStr,
		Chunked:  false,
	}

	return []Document{doc}, nil
}

// GetFileHash computes the SHA256 hash of a file.
func GetFileHash(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash), nil
}

// IsMarkdownFile checks if a file is a Markdown file.
func IsMarkdownFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown" || ext == ".mdx"
}
