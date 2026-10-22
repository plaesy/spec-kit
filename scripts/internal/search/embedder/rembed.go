package embedder

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rostamlabs/rembed"
)

// RembedEmbedder runs a BERT-family encoder in pure Go. There is no cgo, no
// ONNX Runtime, and no server to run. The model is fetched from the Hugging
// Face Hub on first use and cached under $REMBED_CACHE, so the user never
// installs anything by hand. It works on Windows, Linux, and macOS.
type RembedEmbedder struct {
	model         vectorModel
	queryPrefix   string
	passagePrefix string
}

// vectorModel is the part of a rembed model this package uses. It exists so
// tests can replace the real model, which downloads weights and is slow.
type vectorModel interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	Tokenize(text string) []int64
	Dim() int
	Close() error
}

// loadModel loads a model by Hugging Face id. Tests replace it with a fake.
var loadModel = func(name string) (vectorModel, error) {
	m, err := rembed.Load(name)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// NewRembedEmbedder loads cfg.Model through rembed. The first call downloads
// the model; later calls read the cache. A configured Dimension that does not
// match the model is an error, because mixing dimensions would corrupt the
// index without any warning.
func NewRembedEmbedder(cfg Config) (*RembedEmbedder, error) {
	if cfg.Model == "" {
		return nil, errors.New("rembed embedder: model is required (a Hugging Face id such as intfloat/multilingual-e5-base)")
	}

	model, err := loadModel(cfg.Model)
	if err != nil {
		return nil, fmt.Errorf("rembed embedder: loading %q: %w", cfg.Model, err)
	}

	if cfg.Dimension > 0 && model.Dim() != cfg.Dimension {
		_ = model.Close()
		return nil, fmt.Errorf("rembed embedder: model %q produces %d-dimensional vectors, config expects %d", cfg.Model, model.Dim(), cfg.Dimension)
	}

	qp, pp := prefixesFor(cfg.Model)
	return &RembedEmbedder{model: model, queryPrefix: qp, passagePrefix: pp}, nil
}

// prefixesFor returns the prefixes a model family expects on queries and on
// passages. rembed embeds exactly the text it is given and adds none, so the
// prefixes must be added here. The e5 family (multilingual-e5-base and the
// other e5 sizes) uses "query: " and "passage: ". Other models get none.
func prefixesFor(model string) (query, passage string) {
	if strings.Contains(strings.ToLower(model), "e5") {
		return "query: ", "passage: "
	}
	return "", ""
}

// EmbedQuery returns the L2-normalised vector for one search query.
func (e *RembedEmbedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	vecs, err := e.model.Embed(ctx, []string{e.queryPrefix + query})
	if err != nil {
		return nil, err
	}
	if len(vecs) != 1 {
		return nil, fmt.Errorf("rembed embedder: expected 1 query vector, got %d", len(vecs))
	}
	return vecs[0], nil
}

// CountTokens returns the model token count of text. rembed's tokenizer stops
// at the model's maximum length, so a result equal to that maximum means "at
// least this long", which is enough for every cap below the maximum.
// rembed's Tokenize is documented as unstable API; this is the one place that
// depends on it.
func (e *RembedEmbedder) CountTokens(text string) int {
	return len(e.model.Tokenize(text))
}

// EmbedPassages returns one L2-normalised vector per passage.
func (e *RembedEmbedder) EmbedPassages(ctx context.Context, passages []string) ([][]float32, error) {
	prefixed := make([]string, len(passages))
	for i, p := range passages {
		prefixed[i] = e.passagePrefix + p
	}
	return e.model.Embed(ctx, prefixed)
}

// Close releases the model's memory.
func (e *RembedEmbedder) Close() error { return e.model.Close() }
