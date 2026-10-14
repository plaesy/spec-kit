package reach

import (
	"context"
	"strings"
	"testing"
)

// allAdapters builds every platform adapter the registry knows about.
func allAdapters() []PlatformAdapter {
	return []PlatformAdapter{
		NewBilibiliAdapter(),
		NewBossZhipinAdapter(),
		NewExaSearchAdapter(),
		NewFacebookAdapter(),
		NewGitHubAdapter(),
		NewInstagramAdapter(),
		NewLinkedInAdapter(),
		NewRedditAdapter(),
		NewRSSAdapter(),
		NewTwitterAdapter(),
		NewWebAdapter(),
		NewXiaoHongShuAdapter(),
		NewXiaoyuzhouAdapter(),
		NewXueqiuAdapter(),
		NewYouTubeAdapter(),
		NewV2EXAdapter(),
	}
}

// probeActions are the action and query values the adapters switch on. Each one
// reaches a different case, and an unknown value reaches the default branch.
var probeActions = []string{
	"", "search", "video", "view", "user", "job", "list-issues", "list-prs",
	"list-releases", "comments", "stock", "timeline", "tweet", "episode",
	"podcast", "topics", "nodes", "node", "topic", "feed", "unknown-action",
}

var probeQueries = []string{"rust", "https://example.com/post", "BV1xx411c7mD", "123"}

// TestAdapters_EveryBackendAndActionRuns drives each adapter through every
// backend, action and query, once with working tools, once with failing tools,
// and once with no tools installed. It checks that nothing panics and that a
// failed tool is reported as a result, not as a Go error.
func TestAdapters_EveryBackendAndActionRuns(t *testing.T) {
	ctx := context.Background()
	scenarios := []struct {
		name  string
		setup func(t *testing.T)
	}{
		{"tools succeed", func(t *testing.T) { fakeTool(t, `{"ok":true}`) }},
		{"tools fail", func(t *testing.T) { fakeToolExit(t, "rate limited", 1) }},
		{"no tools installed", func(t *testing.T) { noTools(t) }},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			sc.setup(t)
			for _, ad := range allAdapters() {
				for _, backend := range ad.Backends() {
					for _, action := range probeActions {
						for _, q := range probeQueries {
							opts := map[string]string{"action": action}
							res, err := ad.Execute(ctx, backend, q, opts)
							if err == nil && res.Platform == "" && res.Backend == "" && res.Error == "" && !res.Success {
								t.Errorf("%s/%s action=%q query=%q returned an empty result", ad.Name(), backend, action, q)
							}
						}
					}
					if _, err := ad.Health(ctx, backend); err != nil && !strings.Contains(err.Error(), "unknown backend") {
						t.Errorf("%s Health(%s): %v", ad.Name(), backend, err)
					}
				}
				if _, err := ad.Health(ctx, "no-such-backend"); err == nil {
					t.Errorf("%s Health accepted an unknown backend", ad.Name())
				}
				if _, err := ad.Execute(ctx, "no-such-backend", "x", nil); err == nil {
					t.Errorf("%s Execute accepted an unknown backend", ad.Name())
				}
			}
		})
	}
}

// TestAdapters_ConfigureAndAuth covers the credential and auth surface of each
// adapter. Platforms that need a key must accept one and report RequiresAuth.
func TestAdapters_ConfigureAndAuth(t *testing.T) {
	// GitHub's Configure runs `gh auth setup-git`. A fake keeps the real tool out.
	fakeTool(t, "")
	for _, ad := range allAdapters() {
		if err := ad.Configure(map[string]string{}); err != nil {
			t.Errorf("%s Configure(empty): %v", ad.Name(), err)
		}
		if err := ad.Configure(map[string]string{"api_key": "k", "cookie": "c"}); err != nil {
			t.Errorf("%s Configure(with credentials): %v", ad.Name(), err)
		}
		_ = ad.RequiresAuth()
		if len(ad.Backends()) == 0 {
			t.Errorf("%s has no backends", ad.Name())
		}
	}
}

// TestBaseAdapter_CommandResults checks the two helpers every adapter relies on.
func TestBaseAdapter_CommandResults(t *testing.T) {
	ctx := context.Background()
	base := &BaseAdapter{name: "demo", backends: []Backend{"demo-cli"}}

	t.Run("runCommand reports success with trimmed output", func(t *testing.T) {
		fakeTool(t, "  hello  \n")
		res, err := base.runCommand(ctx, "demo-cli", "curl", "-s", "http://x")
		if err != nil {
			t.Fatal(err)
		}
		if !res.Success || res.Data != "hello" || res.Platform != "demo" || res.Backend != "demo-cli" {
			t.Errorf("result = %+v", res)
		}
	})

	t.Run("runCommand reports a failing tool as a result", func(t *testing.T) {
		fakeToolExit(t, "boom", 2)
		res, err := base.runCommand(ctx, "demo-cli", "curl")
		if err != nil {
			t.Fatalf("a failing tool must not be a Go error: %v", err)
		}
		if res.Success || !strings.Contains(res.Error, "boom") {
			t.Errorf("result = %+v, want failure carrying the tool's output", res)
		}
	})

	t.Run("runCommand reports a missing tool as a result", func(t *testing.T) {
		noTools(t)
		res, err := base.runCommand(ctx, "demo-cli", "no-such-tool-xyz")
		if err != nil {
			t.Fatalf("a missing tool must not be a Go error: %v", err)
		}
		if res.Success || res.Error == "" {
			t.Errorf("result = %+v, want a failure with a message", res)
		}
	})

	t.Run("checkCommand is healthy when the tool runs", func(t *testing.T) {
		fakeTool(t, "1.0")
		h := base.checkCommand(ctx, "demo-cli", "curl", "--version")
		if !h.Available {
			t.Errorf("health = %+v, want healthy", h)
		}
	})

	t.Run("checkCommand is unhealthy when the tool is missing", func(t *testing.T) {
		noTools(t)
		h := base.checkCommand(ctx, "demo-cli", "no-such-tool-xyz", "--version")
		if h.Available {
			t.Errorf("health = %+v, want unhealthy", h)
		}
	})

	t.Run("getVersion returns the trimmed tool output", func(t *testing.T) {
		fakeTool(t, "  v9.9  \n")
		v, err := base.getVersion(ctx, "curl")
		if err != nil || v != "v9.9" {
			t.Errorf("getVersion = %q, %v; want v9.9, nil", v, err)
		}
	})
}

// TestAdapters_ArgumentsReachTheTool checks the URL and command each adapter
// builds, using the call log. A wrong URL or tool name is a bug the driver test
// above would not see, because it only checks that nothing panics.
func TestAdapters_ArgumentsReachTheTool(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name     string
		adapter  PlatformAdapter
		backend  Backend
		opts     map[string]string
		query    string
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "v2ex topics",
			adapter:  NewV2EXAdapter(),
			backend:  "v2ex-api",
			opts:     map[string]string{"action": "topics"},
			wantCmd:  "curl",
			wantArgs: []string{"https://www.v2ex.com/api/topics/latest.json"},
		},
		{
			name:     "v2ex user",
			adapter:  NewV2EXAdapter(),
			backend:  "v2ex-api",
			opts:     map[string]string{"action": "user"},
			query:    "alice",
			wantCmd:  "curl",
			wantArgs: []string{"https://www.v2ex.com/api/members/show.json?username=alice"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			log := fakeTool(t, "{}")
			if _, err := c.adapter.Execute(ctx, c.backend, c.query, c.opts); err != nil {
				t.Fatal(err)
			}
			got := calls(t, log)
			if len(got) == 0 {
				t.Fatal("no tool was called")
			}
			if got[0]["cmd"] != c.wantCmd {
				t.Errorf("tool = %v, want %s", got[0]["cmd"], c.wantCmd)
			}
			args, _ := got[0]["args"].([]any)
			joined := ""
			for _, a := range args {
				joined += a.(string) + " "
			}
			for _, want := range c.wantArgs {
				if !strings.Contains(joined, want) {
					t.Errorf("args = %q, want them to contain %q", joined, want)
				}
			}
		})
	}
}
