// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
)

// BilibiliAdapter handles Bilibili access.
type BilibiliAdapter struct {
	BaseAdapter
}

func NewBilibiliAdapter() *BilibiliAdapter {
	return &BilibiliAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformBilibili,
			backends:     []Backend{"bili-cli", "opencli", "bili-api"},
			requiresAuth: false,
		},
	}
}

func (b *BilibiliAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "bili-cli":
		return b.executeBiliCLI(ctx, query, opts)
	case "opencli":
		return b.executeOpenCLI(ctx, query, opts)
	case "bili-api":
		return b.executeBiliAPI(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for bilibili", backend)
	}
}

func (b *BilibiliAdapter) executeBiliCLI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	args := []string{"bili"}
	action := opts["action"]
	if action == "" {
		action = "search"
	}
	args = append(args, action)
	if query != "" {
		args = append(args, query)
	}
	if limit := opts["limit"]; limit != "" {
		args = append(args, "--limit", limit)
	}
	return b.runCommand(ctx, "bili-cli", args...)
}

func (b *BilibiliAdapter) executeOpenCLI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	args := []string{"opencli", "bilibili"}
	action := opts["action"]
	if action != "" {
		args = append(args, action)
	}
	if query != "" {
		args = append(args, query)
	}
	return b.runCommand(ctx, "opencli", args...)
}

func (b *BilibiliAdapter) executeBiliAPI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	// Use Bilibili public API directly via curl
	// This is a fallback when bili-cli is not available
	action := opts["action"]
	if action == "" {
		action = "search"
	}

	var url string
	switch action {
	case "search":
		url = fmt.Sprintf("https://api.bilibili.com/x/web-interface/search/type?keyword=%s&page=1", query)
	case "video":
		url = fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", query)
	default:
		return BackendResult{}, fmt.Errorf("unknown action %s for bili-api", action)
	}

	return b.runCommand(ctx, "bili-api", "curl", "-s", url)
}

func (b *BilibiliAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "bili-cli":
		return b.checkCommand(ctx, backend, "bili", "--version"), nil
	case "opencli":
		return b.checkCommand(ctx, backend, "opencli", "--version"), nil
	case "bili-api":
		return b.checkCommand(ctx, backend, "curl", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}
