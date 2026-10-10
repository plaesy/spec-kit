// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
)

// BossZhipinAdapter handles Boss直聘 job search.
type BossZhipinAdapter struct {
	BaseAdapter
	cookie string
}

func NewBossZhipinAdapter() *BossZhipinAdapter {
	return &BossZhipinAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformBossZhipin,
			backends:     []Backend{"boss-cdp", "boss-api"},
			requiresAuth: true,
		},
	}
}

func (b *BossZhipinAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "boss-cdp":
		return b.executeCDP(ctx, query, opts)
	case "boss-api":
		return b.executeAPI(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for bosszhipin", backend)
	}
}

func (b *BossZhipinAdapter) executeCDP(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	// boss-cdp uses Chrome DevTools Protocol for browser automation
	args := []string{"boss-cdp"}
	action := opts["action"]
	if action == "" {
		action = "search"
	}
	args = append(args, action, query)
	if b.cookie != "" {
		args = append(args, "--cookie", b.cookie)
	}
	return b.runCommand(ctx, "boss-cdp", args...)
}

func (b *BossZhipinAdapter) executeAPI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	// Direct API calls (limited)
	action := opts["action"]
	if action == "" {
		action = "search"
	}

	var url string
	switch action {
	case "search":
		url = fmt.Sprintf("https://www.zhipin.com/wapi/zpgeek/search/joblist.json?query=%s", query)
	case "job":
		url = fmt.Sprintf("https://www.zhipin.com/wapi/zpgeek/view/job/card.json?jobId=%s", query)
	default:
		return BackendResult{}, fmt.Errorf("unknown action %s for bosszhipin", action)
	}

	args := []string{"curl", "-s", "-H", "User-Agent: Mozilla/5.0"}
	if b.cookie != "" {
		args = append(args, "-H", fmt.Sprintf("Cookie: %s", b.cookie))
	}
	args = append(args, url)
	return b.runCommand(ctx, "boss-api", args...)
}

func (b *BossZhipinAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "boss-cdp":
		return b.checkCommand(ctx, backend, "boss-cdp", "--version"), nil
	case "boss-api":
		return b.checkCommand(ctx, backend, "curl", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}

func (b *BossZhipinAdapter) Configure(creds map[string]string) error {
	if cookie, ok := creds["cookie"]; ok {
		b.cookie = cookie
	}
	return nil
}
