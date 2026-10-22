// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
	"time"
)

// ExaSearchAdapter handles Exa semantic search.
type ExaSearchAdapter struct {
	BaseAdapter
	apiKey string
}

func NewExaSearchAdapter() *ExaSearchAdapter {
	return &ExaSearchAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformExaSearch,
			backends:     []Backend{"exa-mcp", "exa-api"},
			requiresAuth: true,
		},
	}
}

func (e *ExaSearchAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "exa-mcp":
		return e.executeMCP(ctx, query, opts)
	case "exa-api":
		return e.executeAPI(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for exa", backend)
	}
}

func (e *ExaSearchAdapter) executeMCP(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	args := []string{"mcporter", "call", "exa", "search", query}

	// Add options
	if numResults := opts["num_results"]; numResults != "" {
		args = append(args, "--num-results", numResults)
	}
	if includeDomains := opts["include_domains"]; includeDomains != "" {
		args = append(args, "--include-domains", includeDomains)
	}
	if excludeDomains := opts["exclude_domains"]; excludeDomains != "" {
		args = append(args, "--exclude-domains", excludeDomains)
	}
	if category := opts["category"]; category != "" {
		args = append(args, "--category", category)
	}

	return e.runCommand(ctx, "exa-mcp", args...)
}

func (e *ExaSearchAdapter) executeAPI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	if e.apiKey == "" {
		return BackendResult{
			Backend:   "exa-api",
			Platform:  PlatformExaSearch,
			Success:   false,
			Error:     "Exa API key not configured",
			Timestamp: time.Now(),
		}, nil
	}

	// Use Exa REST API directly
	// This is a simplified implementation
	args := []string{"curl", "-s", "-X", "POST", "https://api.exa.ai/search",
		"-H", "Content-Type: application/json",
		"-H", fmt.Sprintf("Authorization: Bearer %s", e.apiKey),
		"-d", fmt.Sprintf(`{"query": "%s", "numResults": 10}`, query),
	}
	return e.runCommand(ctx, "exa-api", args...)
}

func (e *ExaSearchAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "exa-mcp":
		return e.checkCommand(ctx, backend, "mcporter", "--version"), nil
	case "exa-api":
		if e.apiKey == "" {
			return BackendHealth{
				Backend:     "exa-api",
				Platform:    PlatformExaSearch,
				Available:   false,
				LastChecked: time.Now(),
				Error:       "API key not configured",
			}, nil
		}
		return e.checkCommand(ctx, backend, "curl", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}

func (e *ExaSearchAdapter) Configure(creds map[string]string) error {
	if apiKey, ok := creds["api_key"]; ok {
		e.apiKey = apiKey
	}
	return nil
}
