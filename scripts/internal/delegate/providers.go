package delegate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const defaultMaxOutputTokens = 1024

// maxErrorBody caps how much of an error response is echoed into an error
// message, so a large response cannot flood the output.
const maxErrorBody = 300

// cliProvider runs a local CLI with the prompt on stdin, for example
// `claude -p` or `kilo run`. The CLI does its own auth, so this provider
// never sees a key. It reports no token counts.
type cliProvider struct {
	name    string
	command string
	args    []string
}

func (p *cliProvider) Name() string { return p.name }

// Capabilities reports what a CLI provider can take. A CLI takes text only.
func (p *cliProvider) Capabilities() []string { return []string{CapText} }

func (p *cliProvider) Complete(ctx context.Context, req Request) (Response, error) {
	cmd := exec.CommandContext(ctx, p.command, p.args...)
	cmd.Dir = req.Dir
	cmd.Stdin = strings.NewReader(req.Prompt)
	// Without a bound, a grandchild holding stdout can keep Wait blocked after
	// the context kills the direct child.
	cmd.WaitDelay = 2 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// A CLI that is missing, times out, or exits non-zero may not have
		// reached a model, so the next provider in a chain may still answer.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return Response{}, &Error{Class: Transient, Err: fmt.Errorf("provider %q: stopped (%w)", p.name, ctxErr)}
		}
		return Response{}, &Error{Class: Transient, Err: fmt.Errorf("provider %q: %s failed: %v: %s", p.name, p.command, err, strings.TrimSpace(stderr.String()))}
	}
	return Response{Text: strings.TrimSpace(stdout.String())}, nil
}

// anthropicProvider calls the Messages API. The key is read from an
// environment variable for each call and goes only into the request header.
// It is never written to a Response, an error, or a log.
type anthropicProvider struct {
	name     string
	model    string
	endpoint string
	keyEnv   string
	getenv   func(string) string
	client   *http.Client
}

func (p *anthropicProvider) Name() string { return p.name }

// Capabilities reports what the Messages API provider can take. This provider
// sends text only.
func (p *anthropicProvider) Capabilities() []string { return []string{CapText} }

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (p *anthropicProvider) Complete(ctx context.Context, req Request) (Response, error) {
	key := ""
	if p.getenv != nil {
		key = p.getenv(p.keyEnv)
	}
	if key == "" {
		// A missing key is configuration. Another provider will not fix it.
		return Response{}, &Error{Class: Terminal, Err: fmt.Errorf("provider %q: environment variable %s is not set", p.name, p.keyEnv)}
	}
	maxTokens := req.MaxOutputTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxOutputTokens
	}
	body, err := json.Marshal(anthropicRequest{
		Model:     p.model,
		MaxTokens: maxTokens,
		Messages:  []anthropicMessage{{Role: "user", Content: req.Prompt}},
	})
	if err != nil {
		return Response{}, &Error{Class: Terminal, Err: fmt.Errorf("provider %q: encoding request: %w", p.name, err)}
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(p.endpoint, "/")+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return Response{}, &Error{Class: Terminal, Err: fmt.Errorf("provider %q: building request: %w", p.name, err)}
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("x-api-key", key)

	var out anthropicResponse
	status, header, raw, err := doJSON(p.client, httpReq, &out)
	if err != nil {
		return Response{}, wrapTransport(p.name, status, err)
	}
	if status < 200 || status > 299 {
		return Response{}, httpError(p.name, status, header, raw)
	}
	var text strings.Builder
	for _, c := range out.Content {
		if c.Type == "text" {
			text.WriteString(c.Text)
		}
	}
	in, outTok := out.Usage.InputTokens, out.Usage.OutputTokens
	return Response{Text: text.String(), InputTokens: &in, OutputTokens: &outTok}, nil
}

// doJSON sends httpReq and decodes a 2xx body into into. It returns the status,
// the response headers, and the raw body, so the caller can classify a non-2xx
// response. A status of 0 means the request never got a response.
func doJSON(client *http.Client, httpReq *http.Request, into any) (int, http.Header, []byte, error) {
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, resp.Header, nil, fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		if err := json.Unmarshal(raw, into); err != nil {
			return resp.StatusCode, resp.Header, raw, fmt.Errorf("decoding response: %w", err)
		}
	}
	return resp.StatusCode, resp.Header, raw, nil
}

// wrapTransport classifies an error from doJSON. No response means the network
// or the server is unreachable, which is transient. A response that could not
// be decoded is terminal, because retrying it elsewhere would repeat the call.
func wrapTransport(name string, status int, err error) error {
	class := Terminal
	if status == 0 {
		class = Transient
	}
	return &Error{Class: class, Err: fmt.Errorf("provider %q: %w", name, err)}
}

// httpError classifies a non-2xx response. 429 is rate limiting, honoured with
// Retry-After. 5xx is transient. Other 4xx responses (bad key, bad request,
// unknown model) are terminal, because the same request will fail elsewhere.
func httpError(name string, status int, header http.Header, raw []byte) error {
	err := fmt.Errorf("provider %q: HTTP %d: %s", name, status, truncate(raw))
	switch {
	case status == http.StatusTooManyRequests:
		return &Error{Class: RateLimited, RetryAfter: retryAfter(header), Err: err}
	case status >= 500:
		return &Error{Class: Transient, Err: err}
	default:
		return &Error{Class: Terminal, Err: err}
	}
}

// retryAfter reads the Retry-After header as seconds or as an HTTP date. It
// returns 0 when the header is absent or unusable.
func retryAfter(h http.Header) time.Duration {
	v := strings.TrimSpace(h.Get("Retry-After"))
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}

// ollamaProvider calls a local Ollama server. It is not billable: the model
// runs on the user's machine.
type ollamaProvider struct {
	name     string
	model    string
	endpoint string
	client   *http.Client
}

func (p *ollamaProvider) Name() string { return p.name }

// Capabilities reports what the Ollama provider can take. This provider sends
// text only.
func (p *ollamaProvider) Capabilities() []string { return []string{CapText} }

type ollamaRequest struct {
	Model   string        `json:"model"`
	Prompt  string        `json:"prompt"`
	Stream  bool          `json:"stream"`
	Options ollamaOptions `json:"options"`
}

type ollamaOptions struct {
	NumPredict int `json:"num_predict"`
}

type ollamaResponse struct {
	Response        string `json:"response"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
}

func (p *ollamaProvider) Complete(ctx context.Context, req Request) (Response, error) {
	maxTokens := req.MaxOutputTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxOutputTokens
	}
	body, err := json.Marshal(ollamaRequest{Model: p.model, Prompt: req.Prompt, Stream: false, Options: ollamaOptions{NumPredict: maxTokens}})
	if err != nil {
		return Response{}, &Error{Class: Terminal, Err: fmt.Errorf("provider %q: encoding request: %w", p.name, err)}
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(p.endpoint, "/")+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return Response{}, &Error{Class: Terminal, Err: fmt.Errorf("provider %q: building request: %w", p.name, err)}
	}
	httpReq.Header.Set("content-type", "application/json")

	var out ollamaResponse
	status, header, raw, err := doJSON(p.client, httpReq, &out)
	if err != nil {
		return Response{}, wrapTransport(p.name, status, err)
	}
	if status < 200 || status > 299 {
		return Response{}, httpError(p.name, status, header, raw)
	}
	in, outTok := out.PromptEvalCount, out.EvalCount
	return Response{Text: strings.TrimSpace(out.Response), InputTokens: &in, OutputTokens: &outTok}, nil
}

func truncate(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > maxErrorBody {
		return s[:maxErrorBody] + "..."
	}
	return s
}
