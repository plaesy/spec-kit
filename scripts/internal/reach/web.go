// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
	"strings"
)

// WebAdapter handles web page reading via Jina Reader.
type WebAdapter struct {
	BaseAdapter
}

func NewWebAdapter() *WebAdapter {
	return &WebAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformWeb,
			backends:     []Backend{"jina-reader", "curl"},
			requiresAuth: false,
		},
	}
}

// Execute reads a web page.
func (w *WebAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "jina-reader":
		return w.executeJinaReader(ctx, query)
	case "curl":
		return w.executeCurl(ctx, query)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for web", backend)
	}
}

func (w *WebAdapter) executeJinaReader(ctx context.Context, url string) (BackendResult, error) {
	// Use Jina Reader API: https://r.jina.ai/http://<url>
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}
	jinaURL := "https://r.jina.ai/" + url
	return w.runCommand(ctx, "jina-reader", "curl", "-s", "-L", jinaURL)
}

func (w *WebAdapter) executeCurl(ctx context.Context, url string) (BackendResult, error) {
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}
	return w.runCommand(ctx, "curl", "curl", "-s", "-L", "-A", "Mozilla/5.0", url)
}

// Health checks if backends are available.
func (w *WebAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "jina-reader":
		return w.checkCommand(ctx, backend, "curl", "--version"), nil
	case "curl":
		return w.checkCommand(ctx, backend, "curl", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}
