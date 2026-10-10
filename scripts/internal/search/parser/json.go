package parser

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// JSONParser parses JSON files.
type JSONParser struct{}

// NewJSONParser creates a new JSON parser.
func NewJSONParser() *JSONParser {
	return &JSONParser{}
}

// Name returns the parser name.
func (p *JSONParser) Name() string {
	return "json"
}

// SupportedExtensions returns the supported file extensions.
func (p *JSONParser) SupportedExtensions() []string {
	return []string{".json", ".jsonl", ".ndjson"}
}

// Parse parses a JSON file and returns documents.
// For JSON arrays, each element becomes a separate document.
// For JSON objects, the entire object becomes one document.
func (p *JSONParser) Parse(ctx context.Context, path string) ([]Document, error) {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, &ParseError{
			Code:    "READ_ERROR",
			Message: "failed to read file",
			Err:     err,
		}
	}

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

	// Parse JSON
	var data interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, &ParseError{
			Code:    "PARSE_ERROR",
			Message: "failed to parse JSON",
			Err:     err,
		}
	}

	// Extract text content based on JSON structure
	var docs []Document
	switch v := data.(type) {
	case []interface{}:
		// JSON array - each element becomes a document
		for i, item := range v {
			text := p.extractText(item)
			if text != "" {
				docs = append(docs, Document{
					ID:       fmt.Sprintf("%s-%d", hashStr[:12], i),
					Path:     path,
					Content:  text,
					Metadata: map[string]string{"source": "json", "index": fmt.Sprintf("%d", i)},
					ModTime:  info.ModTime(),
					Hash:     hashStr,
					Chunked:  false,
				})
			}
		}
	case map[string]interface{}:
		// JSON object - single document
		text := p.extractText(v)
		docs = append(docs, Document{
			ID:       hashStr[:16],
			Path:     path,
			Content:  text,
			Metadata: map[string]string{"source": "json"},
			ModTime:  info.ModTime(),
			Hash:     hashStr,
			Chunked:  false,
		})
	default:
		// Primitive value
		text := p.extractText(v)
		docs = append(docs, Document{
			ID:       hashStr[:16],
			Path:     path,
			Content:  text,
			Metadata: map[string]string{"source": "json"},
			ModTime:  info.ModTime(),
			Hash:     hashStr,
			Chunked:  false,
		})
	}

	return docs, nil
}

// extractText recursively extracts text from a JSON value.
func (p *JSONParser) extractText(v interface{}) string {
	var parts []string
	p.extractTextRecursive(v, &parts)
	return strings.Join(parts, " ")
}

// extractTextRecursive recursively extracts text from a JSON value.
func (p *JSONParser) extractTextRecursive(v interface{}, parts *[]string) {
	switch val := v.(type) {
	case string:
		if s := strings.TrimSpace(val); s != "" {
			*parts = append(*parts, s)
		}
	case float64, int, int64, bool:
		*parts = append(*parts, fmt.Sprintf("%v", val))
	case []interface{}:
		for _, item := range val {
			p.extractTextRecursive(item, parts)
		}
	case map[string]interface{}:
		for k, v := range val {
			if s := strings.TrimSpace(k); s != "" {
				*parts = append(*parts, s)
			}
			p.extractTextRecursive(v, parts)
		}
	case nil:
		// Skip null values
	default:
		*parts = append(*parts, fmt.Sprintf("%v", val))
	}
}

// IsJSONFile checks if a file is a JSON file.
func IsJSONFile(path string) bool {
	ext := strings.ToLower(path)
	return strings.HasSuffix(ext, ".json") || strings.HasSuffix(ext, ".jsonl") || strings.HasSuffix(ext, ".ndjson")
}
