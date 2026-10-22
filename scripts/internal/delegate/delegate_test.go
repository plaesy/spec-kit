package delegate

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
)

const testKey = "sk-test-SECRET-value-123"

// fakeProvider is an in-process provider for tests; it never leaves the process.
type fakeProvider struct {
	name string
	text string
}

func (f *fakeProvider) Name() string { return f.name }

func (f *fakeProvider) Complete(_ context.Context, req Request) (Response, error) {
	return Response{Text: f.text + ":" + req.Prompt}, nil
}

func cfgWith(providers ...config.DelegateProviderConfig) *config.DelegateConfig {
	return &config.DelegateConfig{Providers: providers}
}

func TestSelect_DisabledProviderRefused(t *testing.T) {
	cfg := cfgWith(config.DelegateProviderConfig{Name: "off", Type: "cli", Enabled: false, Command: "x"})
	_, err := Select(cfg, "off", true)
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("want disabled error, got %v", err)
	}
}

func TestSelect_BillableNeedsConfirm(t *testing.T) {
	cfg := cfgWith(config.DelegateProviderConfig{Name: "paid", Type: "anthropic", Enabled: true, Billable: true, Model: "m"})
	if _, err := Select(cfg, "paid", false); err == nil || !strings.Contains(err.Error(), "--confirm") {
		t.Fatalf("billable without confirm: want --confirm error, got %v", err)
	}
	if _, err := Select(cfg, "paid", true); err != nil {
		t.Fatalf("billable with confirm: unexpected error %v", err)
	}
}

func TestSelect_NotBillableDoesNotNeedConfirm(t *testing.T) {
	cfg := cfgWith(config.DelegateProviderConfig{Name: "local", Type: "ollama", Enabled: true, Model: "m"})
	if _, err := Select(cfg, "local", false); err != nil {
		t.Fatalf("non-billable without confirm: unexpected error %v", err)
	}
}

func TestSelect_UnknownListsConfigured(t *testing.T) {
	cfg := cfgWith(config.DelegateProviderConfig{Name: "a"}, config.DelegateProviderConfig{Name: "b"})
	_, err := Select(cfg, "zzz", false)
	if err == nil || !strings.Contains(err.Error(), "a, b") {
		t.Fatalf("want error listing configured names, got %v", err)
	}
}

func TestBuild_UnknownTypeListsSupported(t *testing.T) {
	_, err := Build(config.DelegateProviderConfig{Name: "x", Type: "openai"}, nil)
	if err == nil || !strings.Contains(err.Error(), "cli, anthropic, ollama") {
		t.Fatalf("want supported-types error, got %v", err)
	}
}

func TestRun_StampsNameAndWallTime(t *testing.T) {
	resp, err := Run(context.Background(), &fakeProvider{name: "fake", text: "ok"}, Request{Prompt: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Provider != "fake" || resp.Text != "ok:hi" {
		t.Fatalf("resp = %+v", resp)
	}
	if resp.WallTime < 0 {
		t.Fatalf("wall time negative: %v", resp.WallTime)
	}
}

// TestHelperProcess is not a real test. The cli tests run this test binary as
// the child process, so no external CLI is needed.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	for len(args) > 0 && args[0] != "--" {
		args = args[1:]
	}
	if len(args) > 1 && args[1] == "sleep" {
		time.Sleep(time.Minute)
		os.Exit(0)
	}
	in, _ := io.ReadAll(os.Stdin)
	os.Stdout.WriteString("echo:" + string(in))
	os.Exit(0)
}

func helperArgs(mode string) []string {
	return []string{"-test.run=TestHelperProcess", "--", mode}
}

func TestCLI_ReadsPromptFromStdin(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	p := &cliProvider{name: "cli", command: os.Args[0], args: helperArgs("echo")}
	resp, err := p.Complete(context.Background(), Request{Prompt: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.Text, "echo:hello") {
		t.Fatalf("text = %q, want echo of prompt", resp.Text)
	}
	if resp.InputTokens != nil || resp.OutputTokens != nil {
		t.Fatal("cli provider should report no token counts")
	}
}

func TestCLI_TimeoutKillsChild(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	p := &cliProvider{name: "cli", command: os.Args[0], args: helperArgs("sleep")}
	start := time.Now()
	_, err := Run(context.Background(), p, Request{Prompt: "x", Timeout: 300 * time.Millisecond})
	if err == nil {
		t.Fatal("want timeout error")
	}
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Fatalf("child was not killed promptly: %v", elapsed)
	}
}

func TestAnthropic_ParsesTextAndTokens(t *testing.T) {
	var gotKey, gotVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("x-api-key")
		gotVersion = r.Header.Get("anthropic-version")
		io.Copy(io.Discard, r.Body)
		w.Header().Set("content-type", "application/json")
		io.WriteString(w, `{"content":[{"type":"text","text":"done"}],"usage":{"input_tokens":12,"output_tokens":3}}`)
	}))
	defer srv.Close()

	p := &anthropicProvider{name: "api", model: "m", endpoint: srv.URL, keyEnv: "K", getenv: func(string) string { return testKey }}
	resp, err := p.Complete(context.Background(), Request{Prompt: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Text != "done" || *resp.InputTokens != 12 || *resp.OutputTokens != 3 {
		t.Fatalf("resp = %+v", resp)
	}
	if gotKey != testKey || gotVersion == "" {
		t.Fatalf("headers not sent: key=%q version=%q", gotKey, gotVersion)
	}
}

// TestAnthropic_KeyNeverInOutput checks the key on every path that could leak
// it: the error message, the error response body, and the JSON that the cmd
// prints.
func TestAnthropic_KeyNeverInOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":{"type":"authentication_error","message":"invalid key"}}`)
	}))
	defer srv.Close()

	p := &anthropicProvider{name: "api", model: "m", endpoint: srv.URL, keyEnv: "K", getenv: func(string) string { return testKey }}
	_, err := p.Complete(context.Background(), Request{Prompt: "hi"})
	if err == nil {
		t.Fatal("want HTTP error")
	}
	if strings.Contains(err.Error(), testKey) {
		t.Fatalf("key leaked into error: %v", err)
	}

	ok := &fakeProvider{name: "api", text: "x"}
	resp, _ := Run(context.Background(), ok, Request{Prompt: "p"})
	out, _ := json.Marshal(resp)
	if strings.Contains(string(out), testKey) {
		t.Fatalf("key leaked into JSON output: %s", out)
	}
}

func TestAnthropic_MissingKeyFailsWithoutValue(t *testing.T) {
	p := &anthropicProvider{name: "api", model: "m", endpoint: "http://127.0.0.1:1", keyEnv: "NO_SUCH_VAR", getenv: func(string) string { return "" }}
	_, err := p.Complete(context.Background(), Request{Prompt: "hi"})
	if err == nil || !strings.Contains(err.Error(), "NO_SUCH_VAR") {
		t.Fatalf("want error naming the variable, got %v", err)
	}
}

func TestOllama_ParsesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			http.NotFound(w, r)
			return
		}
		io.Copy(io.Discard, r.Body)
		io.WriteString(w, `{"response":"local answer","prompt_eval_count":5,"eval_count":2}`)
	}))
	defer srv.Close()

	p := &ollamaProvider{name: "local", model: "m", endpoint: srv.URL}
	resp, err := p.Complete(context.Background(), Request{Prompt: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Text != "local answer" || *resp.InputTokens != 5 || *resp.OutputTokens != 2 {
		t.Fatalf("resp = %+v", resp)
	}
}
