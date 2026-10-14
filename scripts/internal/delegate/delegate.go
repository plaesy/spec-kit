// Package delegate sends a bounded prompt to another model through a provider
// chosen in .plaesy/config.json (delegate.providers). Only enabled providers
// run, and a billable provider also needs the caller to confirm.
package delegate

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
)

// Request is one prompt for a provider.
type Request struct {
	Prompt          string
	MaxOutputTokens int           // 0 means the provider default; CLI providers ignore it
	Timeout         time.Duration // 0 means no timeout beyond the context
}

// Response is what a provider returns for one call. Token counts are nil when
// the provider does not report them (the CLI providers).
type Response struct {
	Provider     string        `json:"provider"`
	Text         string        `json:"text"`
	InputTokens  *int          `json:"input_tokens"`
	OutputTokens *int          `json:"output_tokens"`
	WallMillis   int64         `json:"wall_ms"`
	WallTime     time.Duration `json:"-"`
}

// Provider is one way to reach a model. Complete must honour ctx: a timeout
// or cancellation must stop the work, including any child process.
type Provider interface {
	Name() string
	Complete(ctx context.Context, req Request) (Response, error)
}

// Select finds the named provider in cfg and checks that it may run now.
// confirm reports whether the caller passed --confirm.
func Select(cfg *config.DelegateConfig, name string, confirm bool) (config.DelegateProviderConfig, error) {
	if cfg == nil {
		return config.DelegateProviderConfig{}, fmt.Errorf("no delegate configuration")
	}
	for _, p := range cfg.Providers {
		if p.Name != name {
			continue
		}
		if !p.Enabled {
			return p, fmt.Errorf("provider %q is disabled; set delegate.providers[].enabled to true in .plaesy/config.json", name)
		}
		if p.Billable && !confirm {
			return p, fmt.Errorf("provider %q is billable (it charges per call); pass --confirm to run it", name)
		}
		return p, nil
	}
	return config.DelegateProviderConfig{}, fmt.Errorf("unknown provider %q (configured: %s)", name, strings.Join(Names(cfg), ", "))
}

// Names lists the configured provider names in config order.
func Names(cfg *config.DelegateConfig) []string {
	if cfg == nil {
		return nil
	}
	out := make([]string, 0, len(cfg.Providers))
	for _, p := range cfg.Providers {
		out = append(out, p.Name)
	}
	return out
}

// Build turns a provider entry into a Provider. getenv reads the API key for
// the anthropic type; it is injected so tests never touch the real environment.
func Build(p config.DelegateProviderConfig, getenv func(string) string) (Provider, error) {
	switch p.Type {
	case "cli":
		if p.Command == "" {
			return nil, fmt.Errorf("provider %q: cli type needs a command", p.Name)
		}
		return &cliProvider{name: p.Name, command: p.Command, args: p.Args}, nil
	case "anthropic":
		if p.Model == "" {
			return nil, fmt.Errorf("provider %q: anthropic type needs a model", p.Name)
		}
		envName := p.APIKeyEnv
		if envName == "" {
			envName = "ANTHROPIC_API_KEY"
		}
		endpoint := p.Endpoint
		if endpoint == "" {
			endpoint = "https://api.anthropic.com"
		}
		return &anthropicProvider{name: p.Name, model: p.Model, endpoint: endpoint, keyEnv: envName, getenv: getenv}, nil
	case "ollama":
		if p.Model == "" {
			return nil, fmt.Errorf("provider %q: ollama type needs a model", p.Name)
		}
		endpoint := p.Endpoint
		if endpoint == "" {
			endpoint = "http://localhost:11434"
		}
		return &ollamaProvider{name: p.Name, model: p.Model, endpoint: endpoint}, nil
	default:
		return nil, fmt.Errorf("provider %q: unknown type %q (supported: cli, anthropic, ollama)", p.Name, p.Type)
	}
}

// Run applies the timeout, calls the provider, and stamps the wall time.
func Run(ctx context.Context, p Provider, req Request) (Response, error) {
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}
	start := time.Now()
	resp, err := p.Complete(ctx, req)
	if err != nil {
		return Response{}, err
	}
	resp.Provider = p.Name()
	resp.WallTime = time.Since(start)
	resp.WallMillis = resp.WallTime.Milliseconds()
	return resp, nil
}
