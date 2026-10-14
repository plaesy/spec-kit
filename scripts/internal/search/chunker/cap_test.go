package chunker

import (
	"context"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/search/parser"
)

// wordCounter counts whitespace-separated words exactly, so the cap can be
// checked without estimation error.
type wordCounter struct{}

func (wordCounter) Count(text string) int { return len(strings.Fields(text)) }

// runeCounter counts characters, to exercise splitting of a single long token.
type runeCounter struct{}

func (runeCounter) Count(text string) int { return len([]rune(text)) }

func chunkPlain(t *testing.T, cfg Config, counter TokenCounter, content string) []Chunk {
	t.Helper()
	c := NewStructureAwareChunker(cfg, counter)
	doc := parser.Document{ID: "capdoc-000001", Path: "notes.txt", Content: content}
	chunks, err := c.Chunk(context.Background(), doc)
	if err != nil {
		t.Fatalf("Chunk: %v", err)
	}
	return chunks
}

func capConfig(max int) Config {
	cfg := DefaultConfig()
	cfg.MaxChunkTokens = max
	cfg.MinChunkTokens = 1
	cfg.OverlapTokens = 4
	return cfg
}

// A paragraph with no blank line, such as a table or a long code block, used to
// become one chunk of any size. Every chunk must now respect the cap.
func TestCap_OversizedParagraphIsSplit(t *testing.T) {
	const max = 12
	var words []string
	for i := 0; i < 200; i++ {
		words = append(words, "tok")
	}
	content := strings.Join(words, " ")

	chunks := chunkPlain(t, capConfig(max), wordCounter{}, content)
	if len(chunks) < 2 {
		t.Fatalf("expected the paragraph to be split, got %d chunk(s)", len(chunks))
	}
	for i, ch := range chunks {
		if n := (wordCounter{}).Count(ch.Content); n > max {
			t.Errorf("chunk %d has %d tokens, cap is %d", i, n, max)
		}
		if ch.TokenCount > max {
			t.Errorf("chunk %d TokenCount %d exceeds cap %d", i, ch.TokenCount, max)
		}
	}
}

// Splitting should prefer sentence boundaries, so a chunk does not start or stop
// in the middle of a sentence when a sentence boundary was available.
func TestCap_SplitsOnSentenceBoundaries(t *testing.T) {
	const max = 10
	sentence := "Alpha beta gamma delta epsilon zeta."
	content := strings.Repeat(sentence+" ", 8)

	chunks := chunkPlain(t, capConfig(max), wordCounter{}, content)
	if len(chunks) < 2 {
		t.Fatalf("expected several chunks, got %d", len(chunks))
	}
	for i, ch := range chunks {
		if !strings.HasSuffix(strings.TrimSpace(ch.Content), ".") {
			t.Errorf("chunk %d does not end at a sentence boundary: %q", i, ch.Content)
		}
	}
}

// Splitting must not lose text. Every distinct word of the input must still
// appear in some chunk.
func TestCap_NoTextIsLost(t *testing.T) {
	const max = 7
	var b strings.Builder
	for i := 0; i < 60; i++ {
		b.WriteString("word")
		b.WriteString(strings.Repeat("x", i%5))
		b.WriteString(" ")
		if i%9 == 8 {
			b.WriteString("End. ")
		}
	}
	content := b.String()

	chunks := chunkPlain(t, capConfig(max), wordCounter{}, content)
	joined := ""
	for _, ch := range chunks {
		joined += " " + ch.Content
	}
	for _, w := range strings.Fields(content) {
		if !strings.Contains(joined, w) {
			t.Errorf("word %q was lost by chunking", w)
		}
	}
}

// A single token longer than the cap is split by runes. The split must end, and
// each piece must fit the cap.
func TestCap_SingleOverlongTokenIsSplitByRunes(t *testing.T) {
	const max = 8
	token := strings.Repeat("ab", 25) // 50 runes, no spaces

	chunks := chunkPlain(t, capConfig(max), runeCounter{}, token)
	if len(chunks) < 2 {
		t.Fatalf("expected the token to be split, got %d chunk(s)", len(chunks))
	}
	var rebuilt strings.Builder
	for i, ch := range chunks {
		if n := (runeCounter{}).Count(ch.Content); n > max {
			t.Errorf("piece %d has %d runes, cap is %d", i, n, max)
		}
		rebuilt.WriteString(ch.Content)
	}
	if rebuilt.String() != token {
		t.Errorf("pieces do not rebuild the original token")
	}
}

// The overlap carried into the next chunk must not push that chunk past the cap.
func TestCap_OverlapDoesNotExceedCap(t *testing.T) {
	const max = 10
	cfg := capConfig(max)
	cfg.OverlapTokens = 9

	para := strings.Repeat("one two three four five six seven ", 3)
	content := para + "\n\n" + para

	for _, ch := range chunkPlain(t, cfg, wordCounter{}, content) {
		if n := (wordCounter{}).Count(ch.Content); n > max {
			t.Errorf("chunk with overlap has %d tokens, cap is %d", n, max)
		}
	}
}
