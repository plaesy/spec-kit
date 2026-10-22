package retriever

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/plaesy/spec-kit/internal/search/chunker"
	"github.com/plaesy/spec-kit/internal/search/store"
)

// ParentChildRetriever answers a query with parent sections instead of chunks.
//
// The search matches small child chunks, which are precise. Each matched child
// is then mapped to its parent section: every chunk of the same document that
// shares its SectionPath. The parent's full text is returned as the context,
// and the matched children are returned as citations with their offsets.
//
// This is not the sibling expansion in HybridRetriever (expandParentChild).
// That mode appends sibling chunks to the list. This mode returns one entry per
// section, so no section appears twice.
type ParentChildRetriever struct {
	hybrid *HybridRetriever
}

// NewParentChildRetriever wraps a HybridRetriever. The wrapper needs the
// document store (to read a section's chunks) and does its own grouping, so the
// hybrid's sibling expansion is not used.
func NewParentChildRetriever(h *HybridRetriever) *ParentChildRetriever {
	return &ParentChildRetriever{hybrid: h}
}

// ParentSection is one matched section: its full text, and the child chunks that
// matched inside it.
type ParentSection struct {
	DocID       string
	Path        string // the document's path, as stored in the chunk metadata
	SectionPath string // the section's heading path, "" when the chunk has none
	Content     string // every chunk of the section, in document order
	StartLine   int    // 1-based source line of the section's first chunk, 0 when unknown
	EndLine     int    // 1-based, end-inclusive; 0 when unknown
	Score       float64
	Children    []ChildCitation // matched children, best first
}

// ChildCitation points at one matched child chunk inside its parent section.
type ChildCitation struct {
	ChunkID     string
	StartOffset int // character offset in the original document
	EndOffset   int
	Score       float64
}

// Retrieve returns the matched parent sections, best first, at most TopK of them.
// It needs the document store; without one it returns an error rather than
// guessing at section text.
func (p *ParentChildRetriever) Retrieve(ctx context.Context, query string, opts QueryOptions) ([]ParentSection, error) {
	if p.hybrid.docStore == nil {
		return nil, fmt.Errorf("parent-child retrieval needs a document store")
	}
	children, err := p.hybrid.search(ctx, query, opts)
	if err != nil {
		return nil, err
	}

	// Load each document once; several children often share one.
	docs := make(map[string]*store.ProcessedDocument)
	sections := make(map[string]*ParentSection)
	var order []string

	for _, child := range children {
		if child.DocID == "" {
			continue
		}
		doc, ok := docs[child.DocID]
		if !ok {
			procDoc, err := p.hybrid.docStore.GetProcessed(ctx, child.DocID)
			if err != nil || procDoc == nil {
				// A child whose document cannot be read has no section to return.
				docs[child.DocID] = nil
				continue
			}
			doc = procDoc
			docs[child.DocID] = doc
		}
		if doc == nil {
			continue
		}

		chunk, found := findChunk(doc.Chunks, child.ID)
		if !found {
			continue
		}

		key := child.DocID + "\x00" + chunk.SectionPath
		section, exists := sections[key]
		if !exists {
			section = &ParentSection{
				DocID:       child.DocID,
				Path:        child.Metadata["path"],
				SectionPath: chunk.SectionPath,
				Content:     sectionText(doc.Chunks, chunk.SectionPath),
			}
			section.StartLine, section.EndLine = sectionLines(doc.Chunks, chunk.SectionPath)
			sections[key] = section
			order = append(order, key)
		}
		if child.Score > section.Score {
			section.Score = child.Score
		}
		section.Children = append(section.Children, ChildCitation{
			ChunkID:     chunk.ID,
			StartOffset: chunk.StartOffset,
			EndOffset:   chunk.EndOffset,
			Score:       child.Score,
		})
	}

	out := make([]ParentSection, 0, len(order))
	for _, key := range order {
		s := sections[key]
		sort.SliceStable(s.Children, func(i, j int) bool { return s.Children[i].Score > s.Children[j].Score })
		out = append(out, *s)
	}
	// Ties are broken by section key, so the order does not depend on map iteration.
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].DocID+"|"+out[i].SectionPath < out[j].DocID+"|"+out[j].SectionPath
	})

	if top := p.hybrid.config.TopK; top > 0 && len(out) > top {
		out = out[:top]
	}
	return out, nil
}

// findChunk returns the chunk with the given ID.
func findChunk(chunks []chunker.Chunk, id string) (chunker.Chunk, bool) {
	for _, c := range chunks {
		if c.ID == id {
			return c, true
		}
	}
	return chunker.Chunk{}, false
}

// sectionLines returns the first start line and the last end line over the chunks
// of one section, skipping chunks that have no line numbers. Both are 0 when no
// chunk has them.
func sectionLines(chunks []chunker.Chunk, sectionPath string) (int, int) {
	start, end := 0, 0
	for _, c := range chunks {
		if c.SectionPath != sectionPath || c.StartLine == 0 {
			continue
		}
		if start == 0 || c.StartLine < start {
			start = c.StartLine
		}
		if c.EndLine > end {
			end = c.EndLine
		}
	}
	return start, end
}

// sectionText joins every chunk of one section, in document order, separated by
// a blank line. Chunks keep their order in the stored document, so the join does
// not reorder them.
func sectionText(chunks []chunker.Chunk, sectionPath string) string {
	var parts []string
	for _, c := range chunks {
		if c.SectionPath == sectionPath {
			parts = append(parts, c.Content)
		}
	}
	return strings.Join(parts, "\n\n")
}
