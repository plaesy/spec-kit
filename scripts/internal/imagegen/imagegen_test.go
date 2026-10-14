package imagegen

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The provider functions (generateOpenAI/generateGemini) build requests to a
// hard-coded host and hand them to doRequest, which calls http.DefaultClient.Do.
// The base URL is not a parameter, so it cannot be injected without a source
// edit; but http.DefaultClient.Transport is an exported, assignable field, so a
// same-package test can redirect every request to a local httptest.Server. This
// exercises the real request-building and response-parsing code without any
// real network call and without reading a real provider key.
type fakeTransport struct {
	server *httptest.Server

	gotMethod  string
	gotURL     *url.URL
	gotHeaders http.Header
	gotBody    []byte
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	origURL := *req.URL // capture scheme/host/path/query before rewriting
	req.Body = io.NopCloser(bytes.NewReader(body))

	f.gotMethod = req.Method
	f.gotURL = &origURL
	f.gotHeaders = req.Header.Clone()
	f.gotBody = body

	// Rewrite to the local server. DefaultTransport (a distinct global) does the
	// forwarding, so there is no recursion through this transport.
	req.URL.Scheme = "http"
	req.URL.Host = f.server.Listener.Addr().String()
	return http.DefaultTransport.RoundTrip(req)
}

func withTransport(t *testing.T, tr http.RoundTripper) {
	t.Helper()
	old := http.DefaultClient.Transport
	http.DefaultClient.Transport = tr
	t.Cleanup(func() { http.DefaultClient.Transport = old })
}

// setupFake installs a fake transport and a local httptest.Server, capturing
// the request the production code issues.
func setupFake(t *testing.T, handler http.Handler) *fakeTransport {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	f := &fakeTransport{server: srv}
	withTransport(t, f)
	oldTimeout := http.DefaultClient.Timeout
	http.DefaultClient.Timeout = 5 * time.Second
	t.Cleanup(func() { http.DefaultClient.Timeout = oldTimeout })
	return f
}

func b64image(t *testing.T, plain string) string {
	t.Helper()
	return base64.StdEncoding.EncodeToString([]byte(plain))
}

func openAIBody(b64 string) map[string]any {
	return map[string]any{"data": []map[string]any{{"b64_json": b64}}}
}

func geminiBody(b64 string) map[string]any {
	return map[string]any{"predictions": []map[string]any{{"bytesBase64Encoded": b64}}}
}

func jsonOKHandler(t *testing.T, v any) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(v); err != nil {
			t.Errorf("encode response: %v", err)
		}
	})
}

// rawHandler writes status (0 => default 200) then raw body bytes.
func rawHandler(status int, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if status != 0 {
			w.WriteHeader(status)
		}
		io.WriteString(w, body)
	})
}

const (
	fakeOpenAIKey = "fake-openai-key"
	fakeGeminiKey = "fake-gemini-key"
)

func withOpenAIKey(t *testing.T) {
	t.Helper()
	t.Setenv("OPENAI_API_KEY", fakeOpenAIKey)
}

func withGeminiKey(t *testing.T) {
	t.Helper()
	t.Setenv("GEMINI_API_KEY", fakeGeminiKey)
}

func withNoKey(t *testing.T) {
	t.Helper()
	// Force both keys empty so a real key present in the environment can never be
	// used and a real network call can never happen.
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "")
}

// noImageProvider neutralizes PLAESY_IMAGE_PROVIDER from the real environment.
func noImageProvider(t *testing.T) {
	t.Helper()
	t.Setenv("PLAESY_IMAGE_PROVIDER", "")
}

// okOpenAI wires a fake openai success response and a fake key.
func okOpenAI(t *testing.T) *fakeTransport {
	t.Helper()
	withOpenAIKey(t)
	noImageProvider(t)
	return setupFake(t, jsonOKHandler(t, openAIBody(b64image(t, "bytes"))))
}

func okGemini(t *testing.T) *fakeTransport {
	t.Helper()
	withGeminiKey(t)
	noImageProvider(t)
	return setupFake(t, jsonOKHandler(t, geminiBody(b64image(t, "bytes"))))
}

func tempOut(t *testing.T, parts ...string) string {
	t.Helper()
	return filepath.Join(append([]string{t.TempDir(), "imagegen"}, parts...)...)
}

func TestGenerateValidation(t *testing.T) {
	// Validation (prompt/out) runs before directory creation, so a missing
	// required argument must not even attempt to make the output dir.
	tests := []struct {
		name string
		opts Options
	}{
		{"both empty", Options{}},
		{"empty prompt", Options{Out: tempOut(t, "x", "y.png")}},
		{"empty out", Options{Prompt: "p"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Generate(tt.opts)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "prompt and out are required") {
				t.Errorf("error %q, want 'prompt and out are required'", err)
			}
		})
	}
}

func TestGenerateProviderResolution(t *testing.T) {
	t.Run("opts openai", func(t *testing.T) {
		f := okOpenAI(t)
		out := tempOut(t, "a", "img.png")
		path, err := Generate(Options{Prompt: "p", Provider: "openai", Out: out})
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		if path != out {
			t.Errorf("returned path %q, want %q", path, out)
		}
		if f.gotURL.Scheme != "https" || f.gotURL.Host != "api.openai.com" {
			t.Errorf("url %s, want https://api.openai.com", f.gotURL)
		}
		if f.gotURL.Path != "/v1/images/generations" {
			t.Errorf("path %q", f.gotURL.Path)
		}
	})

	t.Run("opts gemini", func(t *testing.T) {
		f := okGemini(t)
		out := tempOut(t, "a", "img.png")
		_, err := Generate(Options{Prompt: "p", Provider: "gemini", Out: out})
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		if f.gotURL.Host != "generativelanguage.googleapis.com" {
			t.Errorf("host %q", f.gotURL.Host)
		}
		if !strings.EqualFold(f.gotURL.Query().Get("key"), fakeGeminiKey) {
			t.Errorf("query key %q", f.gotURL.Query().Get("key"))
		}
		if f.gotHeaders.Get("Authorization") != "" {
			t.Errorf("gemini should send no Authorization header (key is in the URL), got %q", f.gotHeaders.Get("Authorization"))
		}
	})

	t.Run("env openai", func(t *testing.T) {
		f := okOpenAI(t)
		t.Setenv("PLAESY_IMAGE_PROVIDER", "openai")
		out := tempOut(t, "b", "img.png")
		_, err := Generate(Options{Prompt: "p", Out: out})
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		if f.gotURL.Host != "api.openai.com" {
			t.Errorf("host %q", f.gotURL.Host)
		}
	})

	t.Run("env gemini", func(t *testing.T) {
		f := okGemini(t)
		t.Setenv("PLAESY_IMAGE_PROVIDER", "gemini")
		out := tempOut(t, "b", "img.png")
		_, err := Generate(Options{Prompt: "p", Out: out})
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		if f.gotURL.Host != "generativelanguage.googleapis.com" {
			t.Errorf("host %q", f.gotURL.Host)
		}
	})

	t.Run("default to openai when no key and no provider", func(t *testing.T) {
		// No key configured: generateOpenAI must short-circuit before any HTTP,
		// so no fake transport is installed.
		withNoKey(t)
		noImageProvider(t)
		_, err := Generate(Options{Prompt: "p", Out: tempOut(t, "c", "img.png")})
		if !strings.Contains(err.Error(), "OPENAI_API_KEY is not set") {
			t.Errorf("error %q, want OPENAI_API_KEY is not set", err)
		}
	})

	t.Run("unknown provider via opts", func(t *testing.T) {
		withNoKey(t)
		noImageProvider(t)
		out := tempOut(t, "deep", "nest", "img.png")
		dir := filepath.Dir(out)
		_, err := Generate(Options{Prompt: "p", Provider: "bogus", Out: out})
		if err == nil {
			t.Fatal("expected unknown-provider error")
		}
		if !strings.Contains(err.Error(), `unknown provider "bogus"`) {
			t.Errorf("error %q", err)
		}
		if !strings.Contains(err.Error(), "supported: openai, gemini") {
			t.Errorf("error %q", err)
		}
		// MkdirAll runs before the provider switch, so the parent dir is created
		// even though the provider is invalid.
		if _, statErr := os.Stat(dir); statErr != nil {
			t.Errorf("expected parent dir %q to exist before the provider error", dir)
		}
	})

	t.Run("unknown provider via env", func(t *testing.T) {
		withNoKey(t)
		t.Setenv("PLAESY_IMAGE_PROVIDER", "bogus")
		_, err := Generate(Options{Prompt: "p", Out: tempOut(t, "c", "img.png")})
		if !strings.Contains(err.Error(), `unknown provider "bogus"`) {
			t.Errorf("error %q", err)
		}
	})

	t.Run("opts provider beats env", func(t *testing.T) {
		f := okOpenAI(t)
		t.Setenv("PLAESY_IMAGE_PROVIDER", "gemini")
		out := tempOut(t, "d", "img.png")
		_, err := Generate(Options{Prompt: "p", Provider: "openai", Out: out})
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		if f.gotURL.Host != "api.openai.com" {
			t.Errorf("opts.Provider should win over env, got host %q", f.gotURL.Host)
		}
	})
}

func TestGenerateSize(t *testing.T) {
	t.Run("custom size", func(t *testing.T) {
		f := okOpenAI(t)
		_, err := Generate(Options{Prompt: "p", Provider: "openai", Size: "256x256", Out: tempOut(t, "a", "img.png")})
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		var body struct {
			Size string `json:"size"`
		}
		if err := json.Unmarshal(f.gotBody, &body); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if body.Size != "256x256" {
			t.Errorf("size %q, want 256x256", body.Size)
		}
	})

	t.Run("default size", func(t *testing.T) {
		f := okOpenAI(t)
		_, err := Generate(Options{Prompt: "p", Provider: "openai", Out: tempOut(t, "b", "img.png")})
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		var body struct {
			Size string `json:"size"`
		}
		if err := json.Unmarshal(f.gotBody, &body); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if body.Size != defaultSize {
			t.Errorf("size %q, want %q (defaultSize)", body.Size, defaultSize)
		}
		if defaultSize != "1024x1024" {
			t.Errorf("defaultSize = %q, want 1024x1024", defaultSize)
		}
	})
}

// TestGenerateMkdirAllError pins the "creating output directory" wrap when the
// output path sits below a regular file.
func TestGenerateMkdirAllError(t *testing.T) {
	withNoKey(t)
	noImageProvider(t)
	td := t.TempDir()
	blocker := filepath.Join(td, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(blocker, "sub", "img.png")
	_, err := Generate(Options{Prompt: "p", Out: out})
	if !strings.Contains(err.Error(), "creating output directory") {
		t.Errorf("error %q, want 'creating output directory'", err)
	}
}

// TestGenerateWriteFileError pins the "writing output file" wrap when Out points
// at an existing directory.
func TestGenerateWriteFileError(t *testing.T) {
	okOpenAI(t)
	out := tempOut(t, "dir")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := Generate(Options{Prompt: "p", Provider: "openai", Out: out})
	if !strings.Contains(err.Error(), "writing output file") {
		t.Errorf("error %q, want 'writing output file'", err)
	}
	// no image data leaked onto disk
	if _, statErr := os.Stat(out); statErr != nil {
		if !os.IsNotExist(statErr) {
			t.Errorf("stat out: %v", statErr)
		}
	}
}

func TestGenerateWritesFile(t *testing.T) {
	plain := "png-bytes"
	// The fake transport is reinstalled with the real payload for this test.
	f := setupFake(t, jsonOKHandler(t, openAIBody(b64image(t, plain))))
	withOpenAIKey(t)
	noImageProvider(t)
	out := tempOut(t, "assets", "sub", "img.png")
	path, err := Generate(Options{Prompt: "a cat", Provider: "openai", Out: out})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if path != out {
		t.Errorf("returned %q, want %q", path, out)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(got) != plain {
		t.Errorf("file content %q, want %q", got, plain)
	}
	if f.gotURL.Host != "api.openai.com" {
		t.Errorf("host %q", f.gotURL.Host)
	}
}

func TestGenerateNoKeyConfigured(t *testing.T) {
	t.Run("openai via Generate", func(t *testing.T) {
		withNoKey(t)
		noImageProvider(t)
		_, err := Generate(Options{Prompt: "p", Out: tempOut(t, "a", "img.png")})
		if !strings.Contains(err.Error(), "OPENAI_API_KEY is not set") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("openai direct", func(t *testing.T) {
		withNoKey(t)
		_, err := generateOpenAI("p", defaultSize)
		if !strings.Contains(err.Error(), "OPENAI_API_KEY is not set") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("gemini via Generate", func(t *testing.T) {
		withNoKey(t)
		t.Setenv("PLAESY_IMAGE_PROVIDER", "gemini")
		_, err := Generate(Options{Prompt: "p", Out: tempOut(t, "a", "img.png")})
		if !strings.Contains(err.Error(), "GEMINI_API_KEY is not set") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("gemini direct", func(t *testing.T) {
		withNoKey(t)
		_, err := generateGemini("p")
		if !strings.Contains(err.Error(), "GEMINI_API_KEY is not set") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("succeeds with a fake key and fake transport (no real key read)", func(t *testing.T) {
		f := okOpenAI(t)
		if _, err := Generate(Options{Prompt: "p", Provider: "openai", Out: tempOut(t, "a", "img.png")}); err != nil {
			t.Fatalf("Generate: %v", err)
		}
		// The fake key, not a real env value, is what reached the provider.
		if f.gotHeaders.Get("Authorization") != "Bearer "+fakeOpenAIKey {
			t.Errorf("authorization header %q, want bearer %q", f.gotHeaders.Get("Authorization"), fakeOpenAIKey)
		}
	})
}

func TestGenerateProviderError(t *testing.T) {
	t.Run("openai api error", func(t *testing.T) {
		withOpenAIKey(t)
		noImageProvider(t)
		out := tempOut(t, "a", "img.png")
		setupFake(t, rawHandler(401, `{"error":{"message":"Unauthorized"}}`))
		_, err := Generate(Options{Prompt: "p", Provider: "openai", Out: out})
		if !strings.Contains(err.Error(), "OpenAI image API: Unauthorized") {
			t.Errorf("error %q", err)
		}
		if _, statErr := os.Stat(out); statErr == nil {
			t.Error("output file should not be written on API error")
		}
	})

	t.Run("openai non-json 500", func(t *testing.T) {
		withOpenAIKey(t)
		noImageProvider(t)
		setupFake(t, rawHandler(500, "internal server error"))
		_, err := generateOpenAI("p", defaultSize)
		if !strings.Contains(err.Error(), "parsing OpenAI response") {
			t.Errorf("error %q, want 'parsing OpenAI response'", err)
		}
	})

	t.Run("gemini api error", func(t *testing.T) {
		withGeminiKey(t)
		noImageProvider(t)
		out := tempOut(t, "a", "img.png")
		setupFake(t, rawHandler(503, `{"error":{"message":"model overloaded"}}`))
		_, err := Generate(Options{Prompt: "p", Provider: "gemini", Out: out})
		if !strings.Contains(err.Error(), "Gemini image API: model overloaded") {
			t.Errorf("error %q", err)
		}
		if _, statErr := os.Stat(out); statErr == nil {
			t.Error("output file should not be written on API error")
		}
	})

	t.Run("gemini non-json 500", func(t *testing.T) {
		withGeminiKey(t)
		noImageProvider(t)
		setupFake(t, rawHandler(500, "oops"))
		_, err := generateGemini("p")
		if !strings.Contains(err.Error(), "parsing Gemini response") {
			t.Errorf("error %q", err)
		}
	})
}

func TestGenerateOpenAIRequestConstruction(t *testing.T) {
	f := okOpenAI(t)
	prompt := "a red dog"
	size := "512x512"
	decoded, err := generateOpenAI(prompt, size)
	if err != nil {
		t.Fatalf("generateOpenAI: %v", err)
	}
	if string(decoded) != "bytes" {
		t.Errorf("decoded %q, want %q", decoded, "bytes")
	}
	if f.gotMethod != http.MethodPost {
		t.Errorf("method %q, want POST", f.gotMethod)
	}
	if f.gotURL.Scheme != "https" || f.gotURL.Host != "api.openai.com" {
		t.Errorf("url %s", f.gotURL)
	}
	if f.gotURL.Path != "/v1/images/generations" {
		t.Errorf("path %q", f.gotURL.Path)
	}
	if got := f.gotHeaders.Get("Authorization"); got != "Bearer "+fakeOpenAIKey {
		t.Errorf("authorization %q", got)
	}
	if got := f.gotHeaders.Get("Content-Type"); got != "application/json" {
		t.Errorf("content-type %q", got)
	}
	var body struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
		Size   string `json:"size"`
		N      int    `json:"n"`
	}
	if err := json.Unmarshal(f.gotBody, &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body.Model != "gpt-image-1" {
		t.Errorf("model %q, want gpt-image-1", body.Model)
	}
	if body.Prompt != prompt {
		t.Errorf("prompt %q", body.Prompt)
	}
	if body.Size != size {
		t.Errorf("size %q, want %q", body.Size, size)
	}
	if body.N != 1 {
		t.Errorf("n %d, want 1", body.N)
	}
}

func TestGenerateOpenAIResponseParsing(t *testing.T) {
	withOpenAIKey(t)
	noImageProvider(t)

	t.Run("success", func(t *testing.T) {
		setupFake(t, jsonOKHandler(t, openAIBody(b64image(t, "png-bytes"))))
		got, err := generateOpenAI("p", defaultSize)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if string(got) != "png-bytes" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		setupFake(t, rawHandler(200, "not json"))
		_, err := generateOpenAI("p", defaultSize)
		if !strings.Contains(err.Error(), "parsing OpenAI response") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("api error with non-200 status is reported via the error field", func(t *testing.T) {
		// The code never inspects resp.StatusCode; only the JSON "error" field
		// matters. A 401 here surfaces the embedded message, not a status error.
		setupFake(t, rawHandler(401, `{"error":{"message":"Unauthorized"}}`))
		_, err := generateOpenAI("p", defaultSize)
		if !strings.Contains(err.Error(), "OpenAI image API: Unauthorized") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("empty choices: data absent", func(t *testing.T) {
		setupFake(t, rawHandler(200, `{}`))
		_, err := generateOpenAI("p", defaultSize)
		if !strings.Contains(err.Error(), "OpenAI image API: no image data returned") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("empty choices: empty data slice", func(t *testing.T) {
		setupFake(t, rawHandler(200, `{"data":[]}`))
		_, err := generateOpenAI("p", defaultSize)
		if !strings.Contains(err.Error(), "OpenAI image API: no image data returned") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("empty choices: empty b64_json", func(t *testing.T) {
		setupFake(t, rawHandler(200, `{"data":[{"b64_json":""}]}`))
		_, err := generateOpenAI("p", defaultSize)
		if !strings.Contains(err.Error(), "OpenAI image API: no image data returned") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("invalid base64 in b64_json", func(t *testing.T) {
		setupFake(t, rawHandler(200, `{"data":[{"b64_json":"!!!not-base64!!!"}]}`))
		_, err := generateOpenAI("p", defaultSize)
		if err == nil {
			t.Fatal("expected base64 decode error")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "base64") {
			t.Errorf("error %q, want a base64 error", err)
		}
	})
}

func TestGenerateGeminiRequestConstruction(t *testing.T) {
	f := okGemini(t)
	prompt := "a tree"
	decoded, err := generateGemini(prompt)
	if err != nil {
		t.Fatalf("generateGemini: %v", err)
	}
	if string(decoded) != "bytes" {
		t.Errorf("decoded %q, want %q", decoded, "bytes")
	}
	if f.gotMethod != http.MethodPost {
		t.Errorf("method %q, want POST", f.gotMethod)
	}
	if f.gotURL.Scheme != "https" || f.gotURL.Host != "generativelanguage.googleapis.com" {
		t.Errorf("url %s", f.gotURL)
	}
	if f.gotURL.Path != "/v1beta/models/imagen-3.0-generate-002:predict" {
		t.Errorf("path %q", f.gotURL.Path)
	}
	// Suspicious: the API key travels in the URL query string.
	if f.gotURL.Query().Get("key") != fakeGeminiKey {
		t.Errorf("query key %q, want %q", f.gotURL.Query().Get("key"), fakeGeminiKey)
	}
	if got := f.gotHeaders.Get("Authorization"); got != "" {
		t.Errorf("gemini must not send an Authorization header, got %q", got)
	}
	if got := f.gotHeaders.Get("Content-Type"); got != "application/json" {
		t.Errorf("content-type %q", got)
	}
	var body struct {
		Instances  []map[string]any `json:"instances"`
		Parameters map[string]any   `json:"parameters"`
	}
	if err := json.Unmarshal(f.gotBody, &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if len(body.Instances) != 1 {
		t.Fatalf("instances len %d", len(body.Instances))
	}
	if body.Instances[0]["prompt"] != prompt {
		t.Errorf("prompt %v, want %q", body.Instances[0]["prompt"], prompt)
	}
	if body.Parameters["sampleCount"] != float64(1) {
		t.Errorf("sampleCount %v, want 1", body.Parameters["sampleCount"])
	}
}

func TestGenerateGeminiResponseParsing(t *testing.T) {
	withGeminiKey(t)
	noImageProvider(t)

	t.Run("success", func(t *testing.T) {
		setupFake(t, jsonOKHandler(t, geminiBody(b64image(t, "gem-bytes"))))
		got, err := generateGemini("p")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if string(got) != "gem-bytes" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		setupFake(t, rawHandler(200, "not json"))
		_, err := generateGemini("p")
		if !strings.Contains(err.Error(), "parsing Gemini response") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("api error with non-200 status is reported via the error field", func(t *testing.T) {
		setupFake(t, rawHandler(403, `{"error":{"message":"permission denied"}}`))
		_, err := generateGemini("p")
		if !strings.Contains(err.Error(), "Gemini image API: permission denied") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("empty choices: predictions absent", func(t *testing.T) {
		setupFake(t, rawHandler(200, `{}`))
		_, err := generateGemini("p")
		if !strings.Contains(err.Error(), "Gemini image API: no image data returned") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("empty choices: empty predictions slice", func(t *testing.T) {
		setupFake(t, rawHandler(200, `{"predictions":[]}`))
		_, err := generateGemini("p")
		if !strings.Contains(err.Error(), "Gemini image API: no image data returned") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("empty choices: empty bytesBase64Encoded", func(t *testing.T) {
		setupFake(t, rawHandler(200, `{"predictions":[{"bytesBase64Encoded":""}]}`))
		_, err := generateGemini("p")
		if !strings.Contains(err.Error(), "Gemini image API: no image data returned") {
			t.Errorf("error %q", err)
		}
	})

	t.Run("invalid base64 in bytesBase64Encoded", func(t *testing.T) {
		setupFake(t, rawHandler(200, `{"predictions":[{"bytesBase64Encoded":"@@@bad@@@"}]}`))
		_, err := generateGemini("p")
		if err == nil {
			t.Fatal("expected base64 decode error")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "base64") {
			t.Errorf("error %q, want a base64 error", err)
		}
	})
}

// TestDoRequestErrorPaths covers doRequest directly with round-trippers that do
// not touch the network, since the provider functions only call doRequest for the
// transport and read-error branches.
func TestDoRequestErrorPaths(t *testing.T) {
	t.Run("transport error", func(t *testing.T) {
		withTransport(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("dial failed")
		}))
		req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(nil))
		if err != nil {
			t.Fatal(err)
		}
		_, err = doRequest(req)
		if !strings.Contains(err.Error(), "request failed") {
			t.Errorf("error %q, want 'request failed'", err)
		}
		if !strings.Contains(err.Error(), "dial failed") {
			t.Errorf("error %q, want wrapped 'dial failed'", err)
		}
	})

	t.Run("read error", func(t *testing.T) {
		withTransport(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       &failingReader{},
			}, nil
		}))
		req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(nil))
		if err != nil {
			t.Fatal(err)
		}
		_, err = doRequest(req)
		if !strings.Contains(err.Error(), "reading response") {
			t.Errorf("error %q, want 'reading response'", err)
		}
		if !strings.Contains(err.Error(), "read boom") {
			t.Errorf("error %q, want wrapped 'read boom'", err)
		}
	})

	t.Run("success and body close", func(t *testing.T) {
		body := "hello body"
		withTransport(t, roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}))
		req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(nil))
		if err != nil {
			t.Fatal(err)
		}
		got, err := doRequest(req)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if string(got) != body {
			t.Errorf("body %q, want %q", got, body)
		}
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// failingReader is an io.ReadCloser whose Read always errors, to exercise
// doRequest's "reading response" error path without any transport.
type failingReader struct{}

func (failingReader) Read(p []byte) (int, error) { return 0, errors.New("read boom") }
func (failingReader) Close() error               { return nil }
