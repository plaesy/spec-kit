package parser

import (
	"context"
	"time"
)

// Document represents a parsed document with content and metadata.
type Document struct {
	ID       string            // Unique identifier (content hash)
	Path     string            // Source file path
	Content  string            // Extracted text content
	Metadata map[string]string // Additional metadata (source, author, etc.)
	ModTime  time.Time         // File modification time
	Hash     string            // SHA256 hash of raw content
	Chunked  bool              // Whether this document has been chunked
}

// Parser defines the interface for parsing documents.
type Parser interface {
	// Parse parses a file and returns a slice of documents.
	// A single file may produce multiple documents (e.g., JSON array).
	Parse(ctx context.Context, path string) ([]Document, error)

	// SupportedExtensions returns the file extensions this parser handles.
	SupportedExtensions() []string

	// Name returns the parser name.
	Name() string
}

// ParserRegistry manages multiple parsers and routes files to appropriate parser.
type ParserRegistry struct {
	parsers map[string]Parser
}

// NewParserRegistry creates a new parser registry.
func NewParserRegistry() *ParserRegistry {
	return &ParserRegistry{
		parsers: make(map[string]Parser),
	}
}

// Register adds a parser to the registry.
func (r *ParserRegistry) Register(p Parser) {
	for _, ext := range p.SupportedExtensions() {
		r.parsers[ext] = p
	}
}

// Get returns the parser for a given file extension.
func (r *ParserRegistry) Get(ext string) (Parser, bool) {
	p, ok := r.parsers[ext]
	return p, ok
}

// GetSupportedExtensions returns all supported file extensions from registered parsers.
func (r *ParserRegistry) GetSupportedExtensions() []string {
	extSet := make(map[string]bool)
	for _, p := range r.parsers {
		for _, ext := range p.SupportedExtensions() {
			extSet[ext] = true
		}
	}

	extensions := make([]string, 0, len(extSet))
	for ext := range extSet {
		extensions = append(extensions, ext)
	}
	return extensions
}

// ParseFile parses a file using the appropriate parser.
func (r *ParserRegistry) ParseFile(ctx context.Context, path string) ([]Document, error) {
	ext := getExtension(path)
	p, ok := r.Get(ext)
	if !ok {
		return nil, ErrUnsupportedExtension
	}
	return p.Parse(ctx, path)
}

// getExtension returns the file extension including the dot.
func getExtension(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i:]
		}
		if path[i] == '/' || path[i] == '\\' {
			break
		}
	}
	return ""
}

// ErrUnsupportedExtension is returned when no parser supports the file extension.
var ErrUnsupportedExtension = &ParseError{
	Code:    "UNSUPPORTED_EXTENSION",
	Message: "no parser registered for this file extension",
}

// ParseError represents a parsing error.
type ParseError struct {
	Code    string
	Message string
	Err     error
}

func (e *ParseError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *ParseError) Unwrap() error {
	return e.Err
}
