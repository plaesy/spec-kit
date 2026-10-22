// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
)

// XiaoyuzhouAdapter handles Xiaoyuzhou (小宇宙) podcast access.
type XiaoyuzhouAdapter struct {
	BaseAdapter
	whisperKey string
}

func NewXiaoyuzhouAdapter() *XiaoyuzhouAdapter {
	return &XiaoyuzhouAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformXiaoyuzhou,
			backends:     []Backend{"xiaoyuzhou-mcp", "xiaoyuzhou-api"},
			requiresAuth: true,
		},
	}
}

func (x *XiaoyuzhouAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "xiaoyuzhou-mcp":
		return x.executeMCP(ctx, query, opts)
	case "xiaoyuzhou-api":
		return x.executeAPI(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for xiaoyuzhou", backend)
	}
}

func (x *XiaoyuzhouAdapter) executeMCP(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	action := opts["action"]
	if action == "" {
		action = "search"
	}

	args := []string{"mcporter", "call", "xiaoyuzhou", action, query}
	if x.whisperKey != "" {
		args = append(args, "--whisper-key", x.whisperKey)
	}
	return x.runCommand(ctx, "xiaoyuzhou-mcp", args...)
}

func (x *XiaoyuzhouAdapter) executeAPI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	action := opts["action"]
	if action == "" {
		action = "search"
	}

	var url string
	switch action {
	case "search":
		url = fmt.Sprintf("https://www.xiaoyuzhoufm.com/api/search?q=%s", query)
	case "podcast":
		url = fmt.Sprintf("https://www.xiaoyuzhoufm.com/api/podcast/%s", query)
	case "episode":
		url = fmt.Sprintf("https://www.xiaoyuzhoufm.com/api/episode/%s", query)
	default:
		return BackendResult{}, fmt.Errorf("unknown action %s for xiaoyuzhou", action)
	}

	return x.runCommand(ctx, "xiaoyuzhou-api", "curl", "-s", url)
}

func (x *XiaoyuzhouAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "xiaoyuzhou-mcp":
		return x.checkCommand(ctx, backend, "mcporter", "--version"), nil
	case "xiaoyuzhou-api":
		return x.checkCommand(ctx, backend, "curl", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}

func (x *XiaoyuzhouAdapter) Configure(creds map[string]string) error {
	if key, ok := creds["whisper_key"]; ok {
		x.whisperKey = key
	}
	return nil
}
