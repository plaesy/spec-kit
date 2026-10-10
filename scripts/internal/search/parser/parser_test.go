package parser

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMarkdownParser(t *testing.T) {
	// Create temp markdown file
	tmpDir := t.TempDir()
	mdPath := filepath.Join(tmpDir, "test.md")
	mdContent := "# Test Document\n\n## Introduction\n\nThis is a test markdown document with **bold** and *italic* text.\n\n## Code Example\n\n```go\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n```\n\n## List\n\n- Item 1\n- Item 2\n- Item 3\n\n## Table\n\n| Column 1 | Column 2 |\n|----------|----------|\n| Value 1  | Value 2  |\n| Value 3  | Value 4  |\n\nEnd of document.\n"
	err := os.WriteFile(mdPath, []byte(mdContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser := NewMarkdownParser()
	docs, err := parser.Parse(context.Background(), mdPath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("Expected 1 document, got %d", len(docs))
	}

	doc := docs[0]
	if doc.Path != mdPath {
		t.Errorf("Path mismatch: got %s, expected %s", doc.Path, mdPath)
	}

	if doc.Content == "" {
		t.Error("Content should not be empty")
	}

	// Check for expected content
	expectedTerms := []string{"Test Document", "Introduction", "bold", "italic", "Code Example", "Hello, World", "List", "Item 1", "Table", "Column 1", "End of document"}
	for _, term := range expectedTerms {
		if !contains(doc.Content, term) {
			t.Errorf("Content missing expected term: %s", term)
		}
	}

	if doc.Metadata["source"] != "markdown" {
		t.Errorf("Metadata source mismatch: got %s", doc.Metadata["source"])
	}

	if doc.Hash == "" {
		t.Error("Hash should not be empty")
	}

	if doc.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestMarkdownParser_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	mdPath := filepath.Join(tmpDir, "empty.md")
	err := os.WriteFile(mdPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser := NewMarkdownParser()
	docs, err := parser.Parse(context.Background(), mdPath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("Expected 1 document, got %d", len(docs))
	}

	if docs[0].Content != "" {
		t.Errorf("Empty file should have empty content, got: %s", docs[0].Content)
	}
}

func TestMarkdownParser_NonExistentFile(t *testing.T) {
	parser := NewMarkdownParser()
	_, err := parser.Parse(context.Background(), "/non/existent/file.md")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestMarkdownParser_SupportedExtensions(t *testing.T) {
	parser := NewMarkdownParser()
	exts := parser.SupportedExtensions()
	expected := []string{".md", ".markdown", ".mdx"}
	if len(exts) != len(expected) {
		t.Errorf("Expected %d extensions, got %d", len(expected), len(exts))
	}
	for i, ext := range expected {
		if exts[i] != ext {
			t.Errorf("Extension %d: got %s, expected %s", i, exts[i], ext)
		}
	}
}

func TestHTMLParser(t *testing.T) {
	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "test.html")
	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<title>Test Page</title>
	<style>body { color: red; }</style>
	<script>console.log("test");</script>
</head>
<body>
	<h1>Main Title</h1>
	<p>This is a paragraph with <strong>bold</strong> text.</p>
	<div>
		<h2>Section Title</h2>
		<ul>
			<li>Item 1</li>
			<li>Item 2</li>
		</ul>
	</div>
	<footer>Footer content</footer>
</body>
</html>
`
	err := os.WriteFile(htmlPath, []byte(htmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser := NewHTMLParser()
	docs, err := parser.Parse(context.Background(), htmlPath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("Expected 1 document, got %d", len(docs))
	}

	doc := docs[0]
	if doc.Path != htmlPath {
		t.Errorf("Path mismatch: got %s, expected %s", doc.Path, htmlPath)
	}

	if doc.Content == "" {
		t.Error("Content should not be empty")
	}

	// Check that script and style content is removed
	if contains(doc.Content, "console.log") {
		t.Error("Script content should be removed")
	}
	if contains(doc.Content, "color: red") {
		t.Error("Style content should be removed")
	}

	// Check for expected content
	expectedTerms := []string{"Main Title", "paragraph", "bold", "Section Title", "Item 1", "Item 2", "Footer content"}
	for _, term := range expectedTerms {
		if !contains(doc.Content, term) {
			t.Errorf("Content missing expected term: %s", term)
		}
	}

	if doc.Metadata["source"] != "html" {
		t.Errorf("Metadata source mismatch: got %s", doc.Metadata["source"])
	}
}

func TestHTMLParser_SupportedExtensions(t *testing.T) {
	parser := NewHTMLParser()
	exts := parser.SupportedExtensions()
	expected := []string{".html", ".htm", ".xhtml"}
	if len(exts) != len(expected) {
		t.Errorf("Expected %d extensions, got %d", len(expected), len(exts))
	}
	for i, ext := range expected {
		if exts[i] != ext {
			t.Errorf("Extension %d: got %s, expected %s", i, exts[i], ext)
		}
	}
}

func TestJSONParser(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "test.json")
	jsonContent := `{
	"title": "Test Document",
	"author": "Test Author",
	"tags": ["tag1", "tag2", "tag3"],
	"content": {
		"summary": "This is a summary",
		"details": "Detailed content here"
	},
	"count": 42,
	"active": true
}
`
	err := os.WriteFile(jsonPath, []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser := NewJSONParser()
	docs, err := parser.Parse(context.Background(), jsonPath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("Expected 1 document for JSON object, got %d", len(docs))
	}

	doc := docs[0]
	if doc.Content == "" {
		t.Error("Content should not be empty")
	}

	expectedTerms := []string{"Test Document", "Test Author", "tag1", "tag2", "tag3", "summary", "This is a summary", "details", "Detailed content here", "42", "true"}
	for _, term := range expectedTerms {
		if !contains(doc.Content, term) {
			t.Errorf("Content missing expected term: %s", term)
		}
	}

	if doc.Metadata["source"] != "json" {
		t.Errorf("Metadata source mismatch: got %s", doc.Metadata["source"])
	}
}

func TestJSONParser_Array(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "test_array.json")
	jsonContent := `[
	{"id": 1, "name": "Item One", "value": "first"},
	{"id": 2, "name": "Item Two", "value": "second"},
	{"id": 3, "name": "Item Three", "value": "third"}
]
`
	err := os.WriteFile(jsonPath, []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser := NewJSONParser()
	docs, err := parser.Parse(context.Background(), jsonPath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(docs) != 3 {
		t.Fatalf("Expected 3 documents for JSON array, got %d", len(docs))
	}

	for i, doc := range docs {
		if doc.Metadata["index"] != string(rune('0'+i)) {
			t.Errorf("Doc %d: index metadata mismatch: got %s", i, doc.Metadata["index"])
		}
		if doc.Metadata["source"] != "json" {
			t.Errorf("Doc %d: source metadata mismatch: got %s", i, doc.Metadata["source"])
		}
	}
}

func TestJSONParser_JSONL(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "test.jsonl")
	jsonContent := `{"id": 1, "message": "First line"}
{"id": 2, "message": "Second line"}
{"id": 3, "message": "Third line"}
`
	err := os.WriteFile(jsonPath, []byte(jsonContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	parser := NewJSONParser()
	_, err = parser.Parse(context.Background(), jsonPath)
	if err == nil {
		t.Error("Expected error for JSONL (not valid JSON)")
	}
}

func TestJSONParser_SupportedExtensions(t *testing.T) {
	parser := NewJSONParser()
	exts := parser.SupportedExtensions()
	expected := []string{".json", ".jsonl", ".ndjson"}
	if len(exts) != len(expected) {
		t.Errorf("Expected %d extensions, got %d", len(expected), len(exts))
	}
	for i, ext := range expected {
		if exts[i] != ext {
			t.Errorf("Extension %d: got %s, expected %s", i, exts[i], ext)
		}
	}
}

func TestParserRegistry(t *testing.T) {
	registry := NewParserRegistry()
	registry.Register(NewMarkdownParser())
	registry.Register(NewHTMLParser())
	registry.Register(NewJSONParser())

	// Test getting parsers
	mdParser, ok := registry.Get(".md")
	if !ok {
		t.Error("Markdown parser not found")
	}
	if mdParser.Name() != "markdown" {
		t.Errorf("Wrong parser name: %s", mdParser.Name())
	}

	htmlParser, ok := registry.Get(".html")
	if !ok {
		t.Error("HTML parser not found")
	}
	if htmlParser.Name() != "html" {
		t.Errorf("Wrong parser name: %s", htmlParser.Name())
	}

	jsonParser, ok := registry.Get(".json")
	if !ok {
		t.Error("JSON parser not found")
	}
	if jsonParser.Name() != "json" {
		t.Errorf("Wrong parser name: %s", jsonParser.Name())
	}

	// Test unsupported extension
	_, ok = registry.Get(".xyz")
	if ok {
		t.Error("Should not find parser for .xyz")
	}

	// Test supported extensions
	exts := registry.GetSupportedExtensions()
	expectedCount := 9 // 3 md + 3 html + 3 json
	if len(exts) != expectedCount {
		t.Errorf("Expected %d extensions, got %d: %v", expectedCount, len(exts), exts)
	}
}

func TestParserRegistry_ParseFile(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewDefaultRegistry(nil)

	// Create test files
	mdPath := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(mdPath, []byte("# Test\n\nContent"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	htmlPath := filepath.Join(tmpDir, "test.html")
	err = os.WriteFile(htmlPath, []byte("<html><body><h1>Test</h1></body></html>"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	jsonPath := filepath.Join(tmpDir, "test.json")
	err = os.WriteFile(jsonPath, []byte(`{"key": "value"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Parse each
	mdDocs, err := registry.ParseFile(context.Background(), mdPath)
	if err != nil {
		t.Fatalf("Markdown parse failed: %v", err)
	}
	if len(mdDocs) != 1 {
		t.Errorf("Expected 1 markdown doc, got %d", len(mdDocs))
	}

	htmlDocs, err := registry.ParseFile(context.Background(), htmlPath)
	if err != nil {
		t.Fatalf("HTML parse failed: %v", err)
	}
	if len(htmlDocs) != 1 {
		t.Errorf("Expected 1 html doc, got %d", len(htmlDocs))
	}

	jsonDocs, err := registry.ParseFile(context.Background(), jsonPath)
	if err != nil {
		t.Fatalf("JSON parse failed: %v", err)
	}
	if len(jsonDocs) != 1 {
		t.Errorf("Expected 1 json doc, got %d", len(jsonDocs))
	}

	// Test unsupported extension
	txtPath := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(txtPath, []byte("plain text"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	_, err = registry.ParseFile(context.Background(), txtPath)
	if err == nil {
		t.Error("Expected error for unsupported extension")
	}
	if err != ErrUnsupportedExtension {
		t.Errorf("Expected ErrUnsupportedExtension, got: %v", err)
	}
}

func TestParserRegistry_ParseDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := []struct {
		name    string
		content string
	}{
		{"doc1.md", "# Doc 1\n\nContent one"},
		{"doc2.html", "<html><body>Content two</body></html>"},
		{"doc3.json", `{"data": "content three"}`},
		{"doc4.txt", "ignored"},
		{"subdir/doc5.md", "# Sub Doc\n\nContent five"},
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f.name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(f.content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	registry := NewDefaultRegistry(nil)
	docs, err := registry.ParseDirectory(context.Background(), tmpDir, nil)
	if err != nil {
		t.Fatalf("ParseDirectory failed: %v", err)
	}

	// Should parse 4 files (3 types + 1 subdir) but not .txt
	if len(docs) != 4 {
		t.Errorf("Expected 4 documents, got %d", len(docs))
	}

	// Test with extension filter
	docs, err = registry.ParseDirectory(context.Background(), tmpDir, []string{".md"})
	if err != nil {
		t.Fatalf("ParseDirectory with filter failed: %v", err)
	}
	if len(docs) != 2 {
		t.Errorf("Expected 2 markdown documents with filter, got %d", len(docs))
	}
}

func TestGetFileHash(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.md")
	content := "# Test\n\nContent"
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	hash, err := GetFileHash(filePath)
	if err != nil {
		t.Fatalf("GetFileHash failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	// Hash should be deterministic
	hash2, err := GetFileHash(filePath)
	if err != nil {
		t.Fatalf("GetFileHash failed: %v", err)
	}
	if hash != hash2 {
		t.Error("Hash should be deterministic")
	}
}

func TestIsMarkdownFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"test.md", true},
		{"test.markdown", true},
		{"test.mdx", true},
		{"test.MD", true},
		{"test.txt", false},
		{"test.html", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsMarkdownFile(tt.path)
		if result != tt.expected {
			t.Errorf("IsMarkdownFile(%q) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}

func TestIsHTMLFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"test.html", true},
		{"test.htm", true},
		{"test.xhtml", true},
		{"test.HTML", true},
		{"test.txt", false},
		{"test.md", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsHTMLFile(tt.path)
		if result != tt.expected {
			t.Errorf("IsHTMLFile(%q) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}

func TestIsJSONFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"test.json", true},
		{"test.jsonl", true},
		{"test.ndjson", true},
		{"test.JSON", true},
		{"test.txt", false},
		{"test.md", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsJSONFile(tt.path)
		if result != tt.expected {
			t.Errorf("IsJSONFile(%q) = %v, expected %v", tt.path, result, tt.expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
