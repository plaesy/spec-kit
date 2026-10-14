// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
)

// FacebookAdapter handles Facebook access.
type FacebookAdapter struct {
	BaseAdapter
	cookie string
}

func NewFacebookAdapter() *FacebookAdapter {
	return &FacebookAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformFacebook,
			backends:     []Backend{"opencli"},
			requiresAuth: true,
		},
	}
}

func (f *FacebookAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "opencli":
		return f.executeOpenCLI(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for facebook", backend)
	}
}

func (f *FacebookAdapter) executeOpenCLI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	args := []string{"opencli", "facebook"}
	action := opts["action"]
	if action != "" {
		args = append(args, action)
	}
	if query != "" {
		args = append(args, query)
	}
	if f.cookie != "" {
		args = append(args, "--cookie", f.cookie)
	}
	return f.runCommand(ctx, "opencli", args...)
}

func (f *FacebookAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "opencli":
		return f.checkCommand(ctx, backend, "opencli", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}

func (f *FacebookAdapter) Configure(creds map[string]string) error {
	if cookie, ok := creds["cookie"]; ok {
		f.cookie = cookie
	}
	return nil
}
