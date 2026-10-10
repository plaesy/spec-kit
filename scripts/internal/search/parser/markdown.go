package parser

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// MarkdownParser parses Markdown files.
type MarkdownParser struct {
	markdown goldmark.Markdown
}

// NewMarkdownParser creates a new Markdown parser.
func NewMarkdownParser() *MarkdownParser {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Table, extension.Strikethrough),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	return &MarkdownParser{markdown: md}
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

	// Convert markdown to text (extract plain text)
	text := p.extractText(content)

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

// extractText converts markdown to plain text by rendering to HTML then stripping tags.
func (p *MarkdownParser) extractText(content []byte) string {
	// Use goldmark to render to HTML
	var buf strings.Builder
	if err := p.markdown.Convert(content, &buf); err != nil {
		// Fallback: return raw content
		return string(content)
	}

	// Strip HTML tags to get plain text
	htmlContent := buf.String()
	return stripHTMLTags(htmlContent)
}

// stripHTMLTags removes HTML tags from a string.
func stripHTMLTags(html string) string {
	var result strings.Builder
	inTag := false
	for _, r := range html {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	// Clean up whitespace
	text := result.String()
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&", "&")
	text = strings.ReplaceAll(text, "<", "<")
	text = strings.ReplaceAll(text, ">", ">")
	text = strings.ReplaceAll(text, "\"", "\"")
	text = strings.ReplaceAll(text, "'", "'")

	// Normalize whitespace
	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, "\n")
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
