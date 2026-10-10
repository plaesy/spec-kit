// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
)

// RedditAdapter handles Reddit access.
type RedditAdapter struct {
	BaseAdapter
	cookie string
}

func NewRedditAdapter() *RedditAdapter {
	return &RedditAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformReddit,
			backends:     []Backend{"opencli", "rdt-cli"},
			requiresAuth: true,
		},
	}
}

func (r *RedditAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "opencli":
		return r.executeOpenCLI(ctx, query, opts)
	case "rdt-cli":
		return r.executeRdtCLI(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for reddit", backend)
	}
}

func (r *RedditAdapter) executeOpenCLI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	args := []string{"opencli", "reddit"}
	action := opts["action"]
	if action != "" {
		args = append(args, action)
	}
	if query != "" {
		args = append(args, query)
	}
	return r.runCommand(ctx, "opencli", args...)
}

func (r *RedditAdapter) executeRdtCLI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	args := []string{"rdt"}
	action := opts["action"]
	if action == "" {
		action = "search"
	}
	args = append(args, action)
	if query != "" {
		args = append(args, query)
	}
	if r.cookie != "" {
		args = append(args, "--cookie", r.cookie)
	}
	return r.runCommand(ctx, "rdt-cli", args...)
}

func (r *RedditAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "opencli":
		return r.checkCommand(ctx, backend, "opencli", "--version"), nil
	case "rdt-cli":
		return r.checkCommand(ctx, backend, "rdt", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}

func (r *RedditAdapter) Configure(creds map[string]string) error {
	if cookie, ok := creds["cookie"]; ok {
		r.cookie = cookie
	}
	return nil
}
