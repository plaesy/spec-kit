// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
)

// V2EXAdapter handles V2EX forum access.
type V2EXAdapter struct {
	BaseAdapter
}

func NewV2EXAdapter() *V2EXAdapter {
	return &V2EXAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformV2EX,
			backends:     []Backend{"v2ex-api"},
			requiresAuth: false,
		},
	}
}

func (v *V2EXAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "v2ex-api":
		return v.executeAPI(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for v2ex", backend)
	}
}

func (v *V2EXAdapter) executeAPI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	action := opts["action"]
	if action == "" {
		action = "topics"
	}

	var url string
	switch action {
	case "topics":
		url = "https://www.v2ex.com/api/topics/latest.json"
	case "nodes":
		url = "https://www.v2ex.com/api/nodes/all.json"
	case "node":
		url = fmt.Sprintf("https://www.v2ex.com/api/nodes/show.json?name=%s", query)
	case "topic":
		url = fmt.Sprintf("https://www.v2ex.com/api/topics/show.json?id=%s", query)
	case "user":
		url = fmt.Sprintf("https://www.v2ex.com/api/members/show.json?username=%s", query)
	default:
		return BackendResult{}, fmt.Errorf("unknown action %s for v2ex", action)
	}

	return v.runCommand(ctx, "v2ex-api", "curl", "-s", url)
}

func (v *V2EXAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "v2ex-api":
		return v.checkCommand(ctx, backend, "curl", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}
