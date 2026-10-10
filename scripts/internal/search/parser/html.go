package parser

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// HTMLParser parses HTML files.
type HTMLParser struct{}

// NewHTMLParser creates a new HTML parser.
func NewHTMLParser() *HTMLParser {
	return &HTMLParser{}
}

// Name returns the parser name.
func (p *HTMLParser) Name() string {
	return "html"
}

// SupportedExtensions returns the supported file extensions.
func (p *HTMLParser) SupportedExtensions() []string {
	return []string{".html", ".htm", ".xhtml"}
}

// Parse parses an HTML file and returns a document.
func (p *HTMLParser) Parse(ctx context.Context, path string) ([]Document, error) {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, &ParseError{
			Code:    "READ_ERROR",
			Message: "failed to read file",
			Err:     err,
		}
	}

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(content)))
	if err != nil {
		return nil, &ParseError{
			Code:    "PARSE_ERROR",
			Message: "failed to parse HTML",
			Err:     err,
		}
	}

	// Extract text content
	text := p.extractText(doc)

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

	d := Document{
		ID:       hashStr[:16],
		Path:     path,
		Content:  text,
		Metadata: map[string]string{"source": "html"},
		ModTime:  info.ModTime(),
		Hash:     hashStr,
		Chunked:  false,
	}

	return []Document{d}, nil
}

// extractText extracts plain text from an HTML document.
func (p *HTMLParser) extractText(doc *goquery.Document) string {
	// Remove script and style elements
	doc.Find("script, style, noscript, iframe, svg").Remove()

	// Get text from body or entire document
	body := doc.Find("body")
	if body.Length() > 0 {
		return cleanText(body.Text())
	}
	return cleanText(doc.Text())
}

// cleanText normalizes whitespace in text.
func cleanText(text string) string {
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

// IsHTMLFile checks if a file is an HTML file.
func IsHTMLFile(path string) bool {
	ext := strings.ToLower(path)
	return strings.HasSuffix(ext, ".html") || strings.HasSuffix(ext, ".htm") || strings.HasSuffix(ext, ".xhtml")
}
