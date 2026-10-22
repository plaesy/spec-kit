package embedder

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/config"
)

// fakeModel stands in for a loaded rembed model. It records what it was asked
// to embed, so the tests can check the prefixes the adapter adds, and it can be
// told to fail or to return a wrong number of vectors.
type fakeModel struct {
	dim      int
	tokens   int
	embedErr error
	// vecs, when set, replaces the default of one vector per input text.
	vecs   [][]float32
	seen   [][]string
	closed bool
}

func (f *fakeModel) Embed(_ context.Context, texts []string) ([][]float32, error) {
	f.seen = append(f.seen, append([]string(nil), texts...))
	if f.embedErr != nil {
		return nil, f.embedErr
	}
	if f.vecs != nil {
		return f.vecs, nil
	}
	out := make([][]float32, len(texts))
	for i := range texts {
		out[i] = make([]float32, f.dim)
	}
	return out, nil
}

func (f *fakeModel) Tokenize(text string) []int64 {
	if f.tokens > 0 {
		return make([]int64, f.tokens)
	}
	return make([]int64, len(strings.Fields(text)))
}

func (f *fakeModel) Dim() int { return f.dim }

func (f *fakeModel) Close() error {
	f.closed = true
	return nil
}

// withFakeModel replaces loadModel for one test and restores it afterwards.
// The returned fake is the model the adapter will use.
func withFakeModel(t *testing.T, m *fakeModel) {
	t.Helper()
	orig := loadModel
	loadModel = func(string) (vectorModel, error) { return m, nil }
	t.Cleanup(func() { loadModel = orig })
}

func newFakeEmbedder(t *testing.T, model string, m *fakeModel) *RembedEmbedder {
	t.Helper()
	withFakeModel(t, m)
	e, err := NewRembedEmbedder(Config{Type: "rembed", Model: model, Dimension: m.dim})
	if err != nil {
		t.Fatalf("NewRembedEmbedder: %v", err)
	}
	return e
}

func TestNewRembedEmbedder_LoadErrorIsWrapped(t *testing.T) {
	cause := errors.New("network down")
	orig := loadModel
	loadModel = func(string) (vectorModel, error) { return nil, cause }
	t.Cleanup(func() { loadModel = orig })

	_, err := NewRembedEmbedder(Config{Type: "rembed", Model: "intfloat/multilingual-e5-base"})
	if !errors.Is(err, cause) {
		t.Fatalf("err = %v, want it to wrap the load error", err)
	}
	if !strings.Contains(err.Error(), "intfloat/multilingual-e5-base") {
		t.Errorf("err = %q, want the model id in the message", err)
	}
}

func TestNewRembedEmbedder_DimensionMismatchClosesTheModel(t *testing.T) {
	m := &fakeModel{dim: 384}
	withFakeModel(t, m)

	_, err := NewRembedEmbedder(Config{Type: "rembed", Model: "x/small", Dimension: 768})
	if err == nil {
		t.Fatal("expected a dimension mismatch error")
	}
	if !strings.Contains(err.Error(), "384") || !strings.Contains(err.Error(), "768") {
		t.Errorf("err = %q, want both 384 and 768 named", err)
	}
	if !m.closed {
		t.Error("the model was not closed after the mismatch; it would leak")
	}
}

func TestNewRembedEmbedder_ZeroDimensionSkipsTheCheck(t *testing.T) {
	m := &fakeModel{dim: 384}
	withFakeModel(t, m)
	if _, err := NewRembedEmbedder(Config{Type: "rembed", Model: "x/small", Dimension: 0}); err != nil {
		t.Fatalf("Dimension 0 means unchecked, got %v", err)
	}
	if m.closed {
		t.Error("a successful load must not close the model")
	}
}

func TestEmbedQuery_AddsTheQueryPrefix(t *testing.T) {
	m := &fakeModel{dim: 3}
	e := newFakeEmbedder(t, "intfloat/multilingual-e5-base", m)

	vec, err := e.EmbedQuery(context.Background(), "binary search")
	if err != nil {
		t.Fatalf("EmbedQuery: %v", err)
	}
	if len(vec) != 3 {
		t.Errorf("vector length = %d, want 3", len(vec))
	}
	if got := m.seen[0][0]; got != "query: binary search" {
		t.Errorf("embedded %q, want the query prefix added", got)
	}
}

func TestEmbedPassages_AddsThePassagePrefixToEachText(t *testing.T) {
	m := &fakeModel{dim: 3}
	e := newFakeEmbedder(t, "intfloat/multilingual-e5-base", m)

	vecs, err := e.EmbedPassages(context.Background(), []string{"alpha", "beta"})
	if err != nil {
		t.Fatalf("EmbedPassages: %v", err)
	}
	if len(vecs) != 2 {
		t.Fatalf("got %d vectors, want 2", len(vecs))
	}
	want := []string{"passage: alpha", "passage: beta"}
	for i := range want {
		if m.seen[0][i] != want[i] {
			t.Errorf("text %d = %q, want %q", i, m.seen[0][i], want[i])
		}
	}
}

// Passages and queries use different prefixes, so mixing them up would change
// results silently. This pins both at once.
func TestPrefixes_QueryAndPassageDiffer(t *testing.T) {
	m := &fakeModel{dim: 3}
	e := newFakeEmbedder(t, "intfloat/multilingual-e5-base", m)
	_, _ = e.EmbedQuery(context.Background(), "same text")
	_, _ = e.EmbedPassages(context.Background(), []string{"same text"})
	if m.seen[0][0] == m.seen[1][0] {
		t.Errorf("query and passage were embedded identically: %q", m.seen[0][0])
	}
}

func TestEmbedPassages_NonE5ModelGetsNoPrefix(t *testing.T) {
	m := &fakeModel{dim: 3}
	e := newFakeEmbedder(t, "BAAI/bge-m3", m)
	_, _ = e.EmbedQuery(context.Background(), "q")
	_, _ = e.EmbedPassages(context.Background(), []string{"p"})
	if m.seen[0][0] != "q" || m.seen[1][0] != "p" {
		t.Errorf("non-e5 model got prefixes: query=%q passage=%q", m.seen[0][0], m.seen[1][0])
	}
}

func TestEmbedQuery_ModelErrorPropagates(t *testing.T) {
	cause := errors.New("tokenizer failed")
	m := &fakeModel{dim: 3, embedErr: cause}
	e := newFakeEmbedder(t, "intfloat/multilingual-e5-base", m)
	if _, err := e.EmbedQuery(context.Background(), "q"); !errors.Is(err, cause) {
		t.Errorf("err = %v, want the model error", err)
	}
}

func TestEmbedPassages_ModelErrorPropagates(t *testing.T) {
	cause := errors.New("tokenizer failed")
	m := &fakeModel{dim: 3, embedErr: cause}
	e := newFakeEmbedder(t, "intfloat/multilingual-e5-base", m)
	if _, err := e.EmbedPassages(context.Background(), []string{"p"}); !errors.Is(err, cause) {
		t.Errorf("err = %v, want the model error", err)
	}
}

// A query must produce exactly one vector. Zero or several would be silently
// wrong for the search that uses it, so both are errors.
func TestEmbedQuery_WrongVectorCountIsAnError(t *testing.T) {
	for _, n := range []int{0, 2} {
		vecs := make([][]float32, n)
		for i := range vecs {
			vecs[i] = []float32{1, 0, 0}
		}
		m := &fakeModel{dim: 3, vecs: vecs}
		e := newFakeEmbedder(t, "intfloat/multilingual-e5-base", m)
		if _, err := e.EmbedQuery(context.Background(), "q"); err == nil {
			t.Errorf("%d vectors for one query: want an error", n)
		}
	}
}

// The adapter does not renormalise: rembed returns L2-normalised vectors, and
// the real-model test checks that. Here the fake's vector must come back as is.
func TestEmbedQuery_ReturnsTheModelVectorUnchanged(t *testing.T) {
	m := &fakeModel{dim: 2, vecs: [][]float32{{3, 4}}}
	e := newFakeEmbedder(t, "intfloat/multilingual-e5-base", m)
	vec, err := e.EmbedQuery(context.Background(), "q")
	if err != nil {
		t.Fatalf("EmbedQuery: %v", err)
	}
	if vec[0] != 3 || vec[1] != 4 {
		t.Errorf("vector = %v, want [3 4] unchanged", vec)
	}
}

// An empty batch is passed through to the model, which returns whatever it
// returns for no input. The adapter adds no check of its own here.
func TestEmbedPassages_EmptyInputIsPassedThrough(t *testing.T) {
	m := &fakeModel{dim: 3}
	e := newFakeEmbedder(t, "intfloat/multilingual-e5-base", m)
	vecs, err := e.EmbedPassages(context.Background(), nil)
	if err != nil {
		t.Fatalf("EmbedPassages(nil): %v", err)
	}
	if len(vecs) != 0 {
		t.Errorf("got %d vectors for no input, want 0", len(vecs))
	}
	if len(m.seen[0]) != 0 {
		t.Errorf("model was asked for %v, want an empty batch", m.seen[0])
	}
}

func TestCountTokens_DelegatesToTheModelTokenizer(t *testing.T) {
	m := &fakeModel{dim: 3, tokens: 7}
	e := newFakeEmbedder(t, "intfloat/multilingual-e5-base", m)
	if n := e.CountTokens("anything at all"); n != 7 {
		t.Errorf("CountTokens = %d, want 7 from the model tokenizer", n)
	}
}

func TestClose_ClosesTheModel(t *testing.T) {
	m := &fakeModel{dim: 3}
	e := newFakeEmbedder(t, "intfloat/multilingual-e5-base", m)
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if !m.closed {
		t.Error("Close did not reach the model")
	}
}

// The factory must build a working embedder from config, using the same path
// the search command uses.
func TestNewEmbedder_RembedBuildsAWorkingEmbedder(t *testing.T) {
	m := &fakeModel{dim: 768}
	withFakeModel(t, m)
	emb, err := NewEmbedder(&config.EmbedderConfig{Type: "rembed", Model: "intfloat/multilingual-e5-base", Dimension: 768})
	if err != nil {
		t.Fatalf("NewEmbedder: %v", err)
	}
	if _, err := emb.EmbedQuery(context.Background(), "q"); err != nil {
		t.Fatalf("EmbedQuery through the factory: %v", err)
	}
}

// The factory's rejection names rembed, so a user with an old config knows what
// to change it to.
func TestNewEmbedder_UnsupportedTypeNamesRembed(t *testing.T) {
	_, err := NewEmbedder(&config.EmbedderConfig{Type: "onnx", Model: "x", Dimension: 768})
	if err == nil || !strings.Contains(err.Error(), "rembed") {
		t.Errorf("err = %v, want a message naming rembed", err)
	}
}

// The prefix rule matches case-insensitively, so a configured id such as
// "Intfloat/E5-Small" still gets its prefixes.
func TestPrefixesFor_IsCaseInsensitive(t *testing.T) {
	q, p := prefixesFor("Intfloat/E5-Small")
	if q != "query: " || p != "passage: " {
		t.Errorf("prefixesFor(upper-case e5) = (%q, %q), want the e5 prefixes", q, p)
	}
	q, p = prefixesFor("BAAI/bge-m3")
	if q != "" || p != "" {
		t.Errorf("prefixesFor(bge-m3) = (%q, %q), want none", q, p)
	}
}

// The default loader is not replaced here. A name that is not an org/name model
// id is rejected before rembed touches the network, so the loader's error path
// can run without downloading anything. Only the success path needs the real
// model, and TestRembed_RealModel covers that behind PLAESY_TEST_REMBED=1.
func TestLoadModel_RejectsANameThatIsNotAModelID(t *testing.T) {
	m, err := loadModel("no-slash-name")
	if err == nil {
		t.Fatal("expected an error for a name that is not a model id")
	}
	if m != nil {
		t.Errorf("model = %v, want nil on error", m)
	}
}
