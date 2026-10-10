// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
	"strings"
)

// LinkedInAdapter handles LinkedIn access.
type LinkedInAdapter struct {
	BaseAdapter
	mcpConfig string
}

func NewLinkedInAdapter() *LinkedInAdapter {
	return &LinkedInAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformLinkedIn,
			backends:     []Backend{"mcp-linkedin", "jina-reader"},
			requiresAuth: true,
		},
	}
}

func (l *LinkedInAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "mcp-linkedin":
		return l.executeMCPLinkedIn(ctx, query, opts)
	case "jina-reader":
		return l.executeJinaReader(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for linkedin", backend)
	}
}

func (l *LinkedInAdapter) executeMCPLinkedIn(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	// mcp-server-linkedin is an MCP server for LinkedIn
	action := opts["action"]
	if action == "" {
		action = "search"
	}

	args := []string{"mcporter", "call", "linkedin", action, query}
	return l.runCommand(ctx, "mcp-linkedin", args...)
}

func (l *LinkedInAdapter) executeJinaReader(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	// Fallback: use Jina Reader to scrape public LinkedIn pages
	if !strings.HasPrefix(query, "http") {
		query = "https://www.linkedin.com/" + query
	}
	jinaURL := "https://r.jina.ai/" + query
	return l.runCommand(ctx, "jina-reader", "curl", "-s", "-L", jinaURL)
}

func (l *LinkedInAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "mcp-linkedin":
		return l.checkCommand(ctx, backend, "mcporter", "--version"), nil
	case "jina-reader":
		return l.checkCommand(ctx, backend, "curl", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}

func (l *LinkedInAdapter) Configure(creds map[string]string) error {
	if config, ok := creds["mcp_config"]; ok {
		l.mcpConfig = config
	}
	return nil
}
