package chunker

import "strings"

// SimpleTokenCounter estimates tokens as words * 1.3. Tests use it where the
// exact tokenizer does not matter. Production code always passes the embedding
// model's tokenizer.
type SimpleTokenCounter struct{}

// Count estimates the token count of text.
func (c *SimpleTokenCounter) Count(text string) int {
	return int(float64(len(strings.Fields(text))) * 1.3)
}
