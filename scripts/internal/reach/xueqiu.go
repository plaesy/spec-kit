// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
)

// XueqiuAdapter handles Xueqiu (雪球) stock forum access.
type XueqiuAdapter struct {
	BaseAdapter
}

func NewXueqiuAdapter() *XueqiuAdapter {
	return &XueqiuAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformXueqiu,
			backends:     []Backend{"xueqiu-api"},
			requiresAuth: false,
		},
	}
}

func (x *XueqiuAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "xueqiu-api":
		return x.executeAPI(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for xueqiu", backend)
	}
}

func (x *XueqiuAdapter) executeAPI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	action := opts["action"]
	if action == "" {
		action = "search"
	}

	var url string
	switch action {
	case "search":
		url = fmt.Sprintf("https://xueqiu.com/stock/search.json?q=%s", query)
	case "stock":
		url = fmt.Sprintf("https://stock.xueqiu.com/v5/stock/quote.json?symbol=%s", query)
	case "timeline":
		url = fmt.Sprintf("https://xueqiu.com/v4/statuses/user_timeline.json?user_id=%s", query)
	case "comments":
		url = fmt.Sprintf("https://xueqiu.com/query/v1/symbol/search/status.json?q=%s", query)
	default:
		return BackendResult{}, fmt.Errorf("unknown action %s for xueqiu", action)
	}

	return x.runCommand(ctx, "xueqiu-api", "curl", "-s", "-H", "User-Agent: Mozilla/5.0", url)
}

func (x *XueqiuAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "xueqiu-api":
		return x.checkCommand(ctx, backend, "curl", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}
