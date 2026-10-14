package chunker

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/plaesy/spec-kit/internal/search/parser"
)

// StructureAwareChunker chunks documents respecting their structure (headings, paragraphs).
//
// Every chunk stays at or below MaxChunkTokens as counted by the token counter.
// The counter must be the embedding model's tokenizer: a chunk the model cannot
// hold in full is silently truncated when embedded, so the cap has to be in the
// model's own units.
type StructureAwareChunker struct {
	config       Config
	tokenCounter TokenCounter

	// Heading regex for Markdown
	headingRegex *regexp.Regexp

	// Sentence boundary: punctuation followed by whitespace.
	sentenceEnd *regexp.Regexp
}

// NewStructureAwareChunker creates a new structure-aware chunker. The counter
// is required: sizes are only meaningful in the units of the embedding model.
func NewStructureAwareChunker(cfg Config, counter TokenCounter) *StructureAwareChunker {
	if counter == nil {
		panic("chunker: a token counter is required")
	}
	if cfg.MaxChunkTokens == 0 {
		cfg = DefaultConfig()
	}

	return &StructureAwareChunker{
		config:       cfg,
		tokenCounter: counter,
		headingRegex: regexp.MustCompile(`^(#{1,6})\s+(.+)$`),
		sentenceEnd:  regexp.MustCompile(`[.!?]\s+`),
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
	var chunks []Chunk
	var err error
	switch {
	case parser.IsMarkdownFile(doc.Path):
		chunks, err = c.chunkMarkdown(ctx, doc)
	case parser.IsHTMLFile(doc.Path):
		chunks, err = c.chunkHTML(ctx, doc)
	default:
		chunks, err = c.chunkPlainText(ctx, doc)
	}
	if err != nil {
		return nil, err
	}
	return c.finalize(doc, chunks), nil
}

// finalize makes the chunk list for one document consistent. It runs once per
// document, after every section has been chunked.
//
//   - A chunk below MinChunkTokens is merged into the chunk before it, so no
//     text is dropped. The merge is skipped when the result would exceed
//     MaxChunkTokens. A leading small chunk stays on its own, because it has
//     nothing before it to merge into. A document shorter than the minimum
//     still gets one searchable chunk.
//   - IDs are renumbered from 0 across the whole document. Each section used to
//     number its chunks from 0, so two sections produced the same ID and one
//     overwrote the other in the stores.
//   - ParentID links each chunk to the previous one, as before.
func (c *StructureAwareChunker) finalize(doc parser.Document, chunks []Chunk) []Chunk {
	var out []Chunk
	for _, ch := range chunks {
		if ch.Content == "" {
			continue
		}
		if n := len(out); n > 0 && ch.TokenCount < c.config.MinChunkTokens {
			prev := &out[n-1]
			merged := prev.Content + "\n\n" + ch.Content
			if mergedTokens := c.tokenCounter.Count(merged); mergedTokens <= c.config.MaxChunkTokens {
				prev.Content = merged
				prev.TokenCount = mergedTokens
				prev.EndOffset = len(merged)
				if prev.StartLine == 0 {
					prev.StartLine = ch.StartLine
				}
				if ch.EndLine > prev.EndLine {
					prev.EndLine = ch.EndLine
				}
				continue
			}
		}
		out = append(out, ch)
	}

	for i := range out {
		out[i].ID = fmt.Sprintf("%s-%d", doc.ID[:8], i)
		out[i].ParentID = ""
		if i > 0 {
			out[i].ParentID = out[i-1].ID
		}
	}
	return out
}

// chunkMarkdown chunks a Markdown document by headings and paragraphs.
func (c *StructureAwareChunker) chunkMarkdown(ctx context.Context, doc parser.Document) ([]Chunk, error) {
	lines := strings.Split(normalizeNewlines(doc.Content), "\n")

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

		// Section line index i is source line i+1, so the units start at StartLine+1.
		sectionText := strings.Join(lines[section.StartLine:section.EndLine], "\n")
		units := splitUnits(sectionText, section.StartLine+1)
		if len(units) == 0 {
			continue
		}

		sectionChunks, err := c.chunkParagraphs(ctx, doc, units, section.Title)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, sectionChunks...)
	}

	return chunks, nil
}

// chunkHTML chunks an HTML document (simplified - by paragraphs).
func (c *StructureAwareChunker) chunkHTML(ctx context.Context, doc parser.Document) ([]Chunk, error) {
	// For HTML, we already have cleaned text from parser
	// Split by double newlines (paragraphs)
	return c.chunkParagraphs(ctx, doc, splitUnits(doc.Content, 0), "HTML")
}

// chunkPlainText chunks plain text by paragraphs.
func (c *StructureAwareChunker) chunkPlainText(ctx context.Context, doc parser.Document) ([]Chunk, error) {
	return c.chunkParagraphs(ctx, doc, splitUnits(doc.Content, 0), "Text")
}

// chunkParagraphs chunks a list of paragraphs into chunks with overlap. No
// chunk it returns exceeds MaxChunkTokens. A chunk's line range covers the
// paragraphs it adds, not the overlap text it carries from the chunk before.
func (c *StructureAwareChunker) chunkParagraphs(ctx context.Context, doc parser.Document, units []unit, sectionTitle string) ([]Chunk, error) {
	var chunks []Chunk
	var currentChunk strings.Builder
	var currentTokens int
	var first, last int // source lines of the paragraphs this chunk adds

	// addLines widens the range to cover u. A unit with no source lines, as in
	// HTML, leaves the range at 0.
	addLines := func(u unit) {
		if u.first == 0 {
			return
		}
		if first == 0 {
			first = u.first
		}
		if u.last > last {
			last = u.last
		}
	}
	flush := func() {
		chunks = append(chunks, c.createChunk(doc, currentChunk.String(), sectionTitle, first, last))
		first, last = 0, 0
	}

	for _, u := range c.fitParagraphs(units) {
		para := u.text
		paraTokens := c.tokenCounter.Count(para)

		// Check if adding this paragraph would exceed target
		if currentTokens > 0 && currentTokens+paraTokens > c.config.MaxChunkTokens {
			// Finalize current chunk
			flush()

			// Start the next chunk with overlap, but only as much overlap as
			// still leaves room for the paragraph under the cap.
			overlapText := c.getOverlapText(currentChunk.String())
			overlapTokens := c.tokenCounter.Count(overlapText)
			if overlapTokens+paraTokens > c.config.MaxChunkTokens {
				overlapText, overlapTokens = "", 0
			}
			currentChunk.Reset()
			if overlapText != "" {
				currentChunk.WriteString(overlapText)
				currentChunk.WriteString("\n\n")
			}
			currentChunk.WriteString(para)
			currentTokens = overlapTokens + paraTokens
		} else {
			if currentChunk.Len() > 0 {
				currentChunk.WriteString("\n\n")
			}
			currentChunk.WriteString(para)
			currentTokens += paraTokens
		}
		addLines(u)
	}

	// Add final chunk. Size filtering and IDs are handled by finalize.
	if currentChunk.Len() > 0 {
		flush()
	}

	return chunks, nil
}

// fitParagraphs returns the units with every one whose text exceeds
// MaxChunkTokens broken into smaller pieces, in order. Each piece keeps the
// source lines of the unit it came from.
func (c *StructureAwareChunker) fitParagraphs(units []unit) []unit {
	var out []unit
	for _, u := range units {
		for _, piece := range c.fitText(u.text) {
			out = append(out, unit{text: piece, first: u.first, last: u.last})
		}
	}
	return out
}

// fitText breaks text into pieces that each fit within MaxChunkTokens. It packs
// whole sentences first and falls back to words only for a sentence that is
// itself too long. A piece is never cut in the middle of a word.
func (c *StructureAwareChunker) fitText(text string) []string {
	if c.tokenCounter.Count(text) <= c.config.MaxChunkTokens {
		return []string{text}
	}
	return c.pack(c.splitSentences(text), " ")
}

// pack greedily joins units with sep while the result stays within the cap. A
// unit that is too long by itself is split by fitOversize.
func (c *StructureAwareChunker) pack(units []string, sep string) []string {
	var out []string
	var cur string
	flush := func() {
		if cur != "" {
			out = append(out, cur)
			cur = ""
		}
	}
	for _, u := range units {
		if c.tokenCounter.Count(u) > c.config.MaxChunkTokens {
			flush()
			out = append(out, c.fitOversize(u)...)
			continue
		}
		candidate := u
		if cur != "" {
			candidate = cur + sep + u
		}
		if c.tokenCounter.Count(candidate) <= c.config.MaxChunkTokens {
			cur = candidate
			continue
		}
		flush()
		cur = u
	}
	flush()
	return out
}

// fitOversize splits a unit that is over the cap by itself. Text with spaces is
// packed word by word. A single word over the cap, such as a long URL or an
// encoded blob, is split by runes, so the recursion always ends.
func (c *StructureAwareChunker) fitOversize(unit string) []string {
	words := strings.Fields(unit)
	if len(words) > 1 {
		return c.pack(words, " ")
	}
	return c.splitRunes(unit)
}

// splitRunes cuts a single over-long token into pieces that each fit the cap.
func (c *StructureAwareChunker) splitRunes(s string) []string {
	var out []string
	var cur []rune
	for _, r := range s {
		next := string(append(append([]rune{}, cur...), r))
		if len(cur) > 0 && c.tokenCounter.Count(next) > c.config.MaxChunkTokens {
			out = append(out, string(cur))
			cur = []rune{r}
			continue
		}
		cur = append(cur, r)
	}
	if len(cur) > 0 {
		out = append(out, string(cur))
	}
	return out
}

// splitSentences splits text after sentence-ending punctuation.
func (c *StructureAwareChunker) splitSentences(text string) []string {
	var out []string
	last := 0
	for _, loc := range c.sentenceEnd.FindAllStringIndex(text, -1) {
		out = append(out, strings.TrimSpace(text[last:loc[1]]))
		last = loc[1]
	}
	if rest := strings.TrimSpace(text[last:]); rest != "" {
		out = append(out, rest)
	}
	return out
}

// createChunk creates a chunk with metadata. The ID is assigned by finalize,
// once the document's chunks are known. first and last are its source lines.
func (c *StructureAwareChunker) createChunk(doc parser.Document, content string, sectionTitle string, first, last int) Chunk {
	tokenCount := c.tokenCounter.Count(content)

	// Build section path
	sectionPath := sectionTitle
	if sectionTitle == "Document" || sectionTitle == "HTML" || sectionTitle == "Text" {
		sectionPath = ""
	}

	return Chunk{
		DocID:       doc.ID,
		ParentID:    "",
		Content:     content,
		TokenCount:  tokenCount,
		SectionPath: sectionPath,
		StartOffset: 0, // Would need more precise tracking
		EndOffset:   len(content),
		StartLine:   first,
		EndLine:     last,
		Metadata: map[string]string{
			"source": doc.Metadata["source"],
		},
	}
}

// getOverlapText returns the last words of text, about OverlapTokens worth.
// The word count is an estimate; chunkParagraphs checks the real token count.
func (c *StructureAwareChunker) getOverlapText(text string) string {
	if c.config.OverlapTokens <= 0 {
		return ""
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	overlapWords := int(float64(c.config.OverlapTokens) / 1.3)
	if overlapWords > len(words) {
		overlapWords = len(words)
	}

	return strings.Join(words[len(words)-overlapWords:], " ")
}

// unit is one paragraph of a document, or one piece of a paragraph that was
// too long for a chunk. first and last are the 1-based, end-inclusive source
// lines it came from. 0 means the source has no line numbers.
type unit struct {
	text  string
	first int
	last  int
}

// normalizeNewlines converts CRLF and lone CR to LF, so line numbers count the
// same way for every file.
func normalizeNewlines(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.ReplaceAll(text, "\r", "\n")
}

// splitUnits splits text into paragraphs: runs of lines separated by blank
// lines. Line 0 of text is source line firstLine. Pass firstLine 0 when the
// text has no source lines, and every unit then has first and last 0.
func splitUnits(text string, firstLine int) []unit {
	lines := strings.Split(normalizeNewlines(text), "\n")

	var units []unit
	var group []int // indexes into lines of the current paragraph
	flush := func() {
		if len(group) == 0 {
			return
		}
		parts := make([]string, 0, len(group))
		for _, i := range group {
			parts = append(parts, lines[i])
		}
		if body := strings.TrimSpace(strings.Join(parts, "\n")); body != "" {
			u := unit{text: body}
			if firstLine > 0 {
				u.first = firstLine + group[0]
				u.last = firstLine + group[len(group)-1]
			}
			units = append(units, u)
		}
		group = group[:0]
	}
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		group = append(group, i)
	}
	flush()
	return units
}

// splitParagraphs splits text into paragraph texts. See splitUnits.
func splitParagraphs(text string) []string {
	var out []string
	for _, u := range splitUnits(text, 0) {
		out = append(out, u.text)
	}
	return out
}
