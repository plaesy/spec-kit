// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
)

// InstagramAdapter handles Instagram access.
type InstagramAdapter struct {
	BaseAdapter
	cookie string
}

func NewInstagramAdapter() *InstagramAdapter {
	return &InstagramAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformInstagram,
			backends:     []Backend{"opencli"},
			requiresAuth: true,
		},
	}
}

func (i *InstagramAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "opencli":
		return i.executeOpenCLI(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for instagram", backend)
	}
}

func (i *InstagramAdapter) executeOpenCLI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	args := []string{"opencli", "instagram"}
	action := opts["action"]
	if action != "" {
		args = append(args, action)
	}
	if query != "" {
		args = append(args, query)
	}
	if i.cookie != "" {
		args = append(args, "--cookie", i.cookie)
	}
	return i.runCommand(ctx, "opencli", args...)
}

func (i *InstagramAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "opencli":
		return i.checkCommand(ctx, backend, "opencli", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}

func (i *InstagramAdapter) Configure(creds map[string]string) error {
	if cookie, ok := creds["cookie"]; ok {
		i.cookie = cookie
	}
	return nil
}
