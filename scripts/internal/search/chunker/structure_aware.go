package chunker

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/plaesy/spec-kit/internal/search/parser"
)

// StructureAwareChunker chunks documents respecting their structure (headings, paragraphs).
type StructureAwareChunker struct {
	config       Config
	tokenCounter TokenCounter

	// Heading regex for Markdown
	headingRegex *regexp.Regexp
}

// NewStructureAwareChunker creates a new structure-aware chunker.
func NewStructureAwareChunker(cfg Config, counter TokenCounter) *StructureAwareChunker {
	if cfg.TargetTokens == 0 {
		cfg = DefaultConfig()
	}
	if counter == nil {
		counter = &SimpleTokenCounter{}
	}

	return &StructureAwareChunker{
		config:       cfg,
		tokenCounter: counter,
		headingRegex: regexp.MustCompile(`^(#{1,6})\s+(.+)$`),
	}
}

// Name returns the chunker name.
func (c *StructureAwareChunker) Name() string {
	return "structure-aware"
}

// Config returns the chunker configuration.
func (c *StructureAwareChunker) Config() Config {
	return c.config
}

// Chunk splits a document into chunks based on structure.
func (c *StructureAwareChunker) Chunk(ctx context.Context, doc parser.Document) ([]Chunk, error) {
	switch {
	case parser.IsMarkdownFile(doc.Path):
		return c.chunkMarkdown(ctx, doc)
	case parser.IsHTMLFile(doc.Path):
		return c.chunkHTML(ctx, doc)
	default:
		return c.chunkPlainText(ctx, doc)
	}
}

// chunkMarkdown chunks a Markdown document by headings and paragraphs.
func (c *StructureAwareChunker) chunkMarkdown(ctx context.Context, doc parser.Document) ([]Chunk, error) {
	lines := strings.Split(doc.Content, "\n")

	// Parse document structure: headings and content sections
	type Section struct {
		Level     int    // Heading level (1-6)
		Title     string // Heading text
		StartLine int    // Start line index
		EndLine   int    // End line index (exclusive)
		Content   string // Section content
	}

	var sections []Section
	var currentSection *Section

	for i, line := range lines {
		matches := c.headingRegex.FindStringSubmatch(line)
		if matches != nil {
			// New heading found
			if currentSection != nil {
				currentSection.EndLine = i
				sections = append(sections, *currentSection)
			}

			level := len(matches[1])
			title := strings.TrimSpace(matches[2])
			currentSection = &Section{
				Level:     level,
				Title:     title,
				StartLine: i + 1, // Content starts after heading
			}
		}
	}

	// Close last section
	if currentSection != nil {
		currentSection.EndLine = len(lines)
		sections = append(sections, *currentSection)
	}

	// If no headings found, treat entire document as one section
	if len(sections) == 0 {
		sections = []Section{{
			Level:     0,
			Title:     "Document",
			StartLine: 0,
			EndLine:   len(lines),
			Content:   doc.Content,
		}}
	}

	// Build section content and chunk each section
	var chunks []Chunk
	for _, section := range sections {
		if section.StartLine >= section.EndLine {
			continue
		}

		sectionContent := strings.Join(lines[section.StartLine:section.EndLine], "\n")
		sectionContent = strings.TrimSpace(sectionContent)
		if sectionContent == "" {
			continue
		}

		sectionChunks := c.chunkText(ctx, doc, sectionContent, section.Title)
		chunks = append(chunks, sectionChunks...)
	}

	return chunks, nil
}

// chunkHTML chunks an HTML document (simplified - by paragraphs).
func (c *StructureAwareChunker) chunkHTML(ctx context.Context, doc parser.Document) ([]Chunk, error) {
	// For HTML, we already have cleaned text from parser
	// Split by double newlines (paragraphs)
	paragraphs := splitParagraphs(doc.Content)
	return c.chunkParagraphs(ctx, doc, paragraphs, "HTML")
}

// chunkPlainText chunks plain text by paragraphs.
func (c *StructureAwareChunker) chunkPlainText(ctx context.Context, doc parser.Document) ([]Chunk, error) {
	paragraphs := splitParagraphs(doc.Content)
	return c.chunkParagraphs(ctx, doc, paragraphs, "Text")
}

// chunkParagraphs chunks a list of paragraphs into chunks with overlap.
func (c *StructureAwareChunker) chunkParagraphs(ctx context.Context, doc parser.Document, paragraphs []string, sectionTitle string) ([]Chunk, error) {
	var chunks []Chunk
	var currentChunk strings.Builder
	var currentTokens int
	chunkIndex := 0

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		paraTokens := c.tokenCounter.Count(para)

		// Check if adding this paragraph would exceed target
		if currentTokens > 0 && currentTokens+paraTokens > c.config.MaxChunkTokens {
			// Finalize current chunk
			chunk := c.createChunk(doc, currentChunk.String(), chunkIndex, sectionTitle)
			if chunk.TokenCount >= c.config.MinChunkTokens {
				chunks = append(chunks, chunk)
				chunkIndex++
			}

			// Start new chunk with overlap
			overlapText := c.getOverlapText(currentChunk.String())
			currentChunk.Reset()
			currentChunk.WriteString(overlapText)
			currentChunk.WriteString("\n\n")
			currentChunk.WriteString(para)
			currentTokens = c.tokenCounter.Count(overlapText) + paraTokens
		} else {
			if currentChunk.Len() > 0 {
				currentChunk.WriteString("\n\n")
			}
			currentChunk.WriteString(para)
			currentTokens += paraTokens
		}
	}

	// Add final chunk
	if currentChunk.Len() > 0 {
		chunk := c.createChunk(doc, currentChunk.String(), chunkIndex, sectionTitle)
		if chunk.TokenCount >= c.config.MinChunkTokens {
			chunks = append(chunks, chunk)
		}
	}

	// Set parent-child relationships
	for i := range chunks {
		if i > 0 {
			chunks[i].ParentID = chunks[i-1].ID
		}
	}

	return chunks, nil
}

// chunkText chunks arbitrary text.
func (c *StructureAwareChunker) chunkText(ctx context.Context, doc parser.Document, text, sectionTitle string) []Chunk {
	paragraphs := splitParagraphs(text)
	chunks, _ := c.chunkParagraphs(ctx, doc, paragraphs, sectionTitle)
	return chunks
}

// createChunk creates a chunk with metadata.
func (c *StructureAwareChunker) createChunk(doc parser.Document, content string, index int, sectionTitle string) Chunk {
	tokenCount := c.tokenCounter.Count(content)

	chunkID := fmt.Sprintf("%s-%d", doc.ID[:8], index)

	// Build section path
	sectionPath := sectionTitle
	if sectionTitle == "Document" || sectionTitle == "HTML" || sectionTitle == "Text" {
		sectionPath = ""
	}

	return Chunk{
		ID:          chunkID,
		DocID:       doc.ID,
		ParentID:    "",
		Content:     content,
		TokenCount:  tokenCount,
		SectionPath: sectionPath,
		StartOffset: 0, // Would need more precise tracking
		EndOffset:   len(content),
		Metadata: map[string]string{
			"source": doc.Metadata["source"],
		},
	}
}

// getOverlapText returns the last N tokens as overlap text.
func (c *StructureAwareChunker) getOverlapText(text string) string {
	if c.config.OverlapTokens <= 0 {
		return ""
	}

	words := splitWords(text)
	if len(words) == 0 {
		return ""
	}

	// Estimate words for overlap tokens
	overlapWords := int(float64(c.config.OverlapTokens) / 1.3)
	if overlapWords > len(words) {
		overlapWords = len(words)
	}

	return strings.Join(words[len(words)-overlapWords:], " ")
}

// splitParagraphs splits text by double newlines (paragraphs).
func splitParagraphs(text string) []string {
	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// Split by double newline
	parts := strings.Split(text, "\n\n")

	var paragraphs []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			paragraphs = append(paragraphs, part)
		}
	}
	return paragraphs
}

// Config holds chunker configuration for the StructureAwareChunker.
type StructureAwareConfig struct {
	Config
}
