// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
)

// XiaoHongShuAdapter handles XiaoHongShu (Little Red Book) access.
type XiaoHongShuAdapter struct {
	BaseAdapter
	cookie string
}

func NewXiaoHongShuAdapter() *XiaoHongShuAdapter {
	return &XiaoHongShuAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformXiaoHongShu,
			backends:     []Backend{"opencli", "xiaohongshu-mcp", "xhs-cli"},
			requiresAuth: true,
		},
	}
}

func (x *XiaoHongShuAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "opencli":
		return x.executeOpenCLI(ctx, query, opts)
	case "xiaohongshu-mcp":
		return x.executeMCP(ctx, query, opts)
	case "xhs-cli":
		return x.executeXHSCLI(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for xiaohongshu", backend)
	}
}

func (x *XiaoHongShuAdapter) executeOpenCLI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	args := []string{"opencli", "xiaohongshu"}
	action := opts["action"]
	if action != "" {
		args = append(args, action)
	}
	if query != "" {
		args = append(args, query)
	}
	return x.runCommand(ctx, "opencli", args...)
}

func (x *XiaoHongShuAdapter) executeMCP(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	// xiaohongshu-mcp is an MCP server
	// This would connect via MCP protocol
	args := []string{"mcporter", "call", "xiaohongshu", "search", query}
	if x.cookie != "" {
		args = append(args, "--cookie", x.cookie)
	}
	return x.runCommand(ctx, "xiaohongshu-mcp", args...)
}

func (x *XiaoHongShuAdapter) executeXHSCLI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	args := []string{"xhs"}
	action := opts["action"]
	if action == "" {
		action = "search"
	}
	args = append(args, action, query)
	if x.cookie != "" {
		args = append(args, "--cookie", x.cookie)
	}
	return x.runCommand(ctx, "xhs-cli", args...)
}

func (x *XiaoHongShuAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "opencli":
		return x.checkCommand(ctx, backend, "opencli", "--version"), nil
	case "xiaohongshu-mcp":
		return x.checkCommand(ctx, backend, "mcporter", "--version"), nil
	case "xhs-cli":
		return x.checkCommand(ctx, backend, "xhs", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}

func (x *XiaoHongShuAdapter) Configure(creds map[string]string) error {
	if cookie, ok := creds["cookie"]; ok {
		x.cookie = cookie
	}
	return nil
}
