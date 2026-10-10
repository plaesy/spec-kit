// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
	"strings"
	"time"

	"os/exec"
)

// TwitterAdapter handles Twitter/X access.
type TwitterAdapter struct {
	BaseAdapter
	authToken string
	ct0       string
}

func NewTwitterAdapter() *TwitterAdapter {
	return &TwitterAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformTwitter,
			backends:     []Backend{"twitter-cli", "opencli", "bird"},
			requiresAuth: true,
		},
	}
}

// Execute fetches Twitter data.
func (t *TwitterAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "twitter-cli":
		return t.executeTwitterCLI(ctx, query, opts)
	case "opencli":
		return t.executeOpenCLI(ctx, query, opts)
	case "bird":
		return t.executeBird(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for twitter", backend)
	}
}

func (t *TwitterAdapter) executeTwitterCLI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	// twitter-cli commands: tweet, user, search, timeline
	action := opts["action"]
	if action == "" {
		action = "search"
	}

	args := []string{"twitter"}

	// Set auth env vars
	env := []string{}
	if t.authToken != "" {
		env = append(env, "TWITTER_AUTH_TOKEN="+t.authToken)
	}
	if t.ct0 != "" {
		env = append(env, "TWITTER_CT0="+t.ct0)
	}

	switch action {
	case "tweet":
		args = append(args, "tweet", "view", query)
	case "user":
		args = append(args, "user", "view", query)
	case "search":
		args = append(args, "search", query)
		if limit := opts["limit"]; limit != "" {
			args = append(args, "--limit", limit)
		}
	case "timeline":
		args = append(args, "timeline")
	default:
		return BackendResult{}, fmt.Errorf("unknown action %s for twitter-cli", action)
	}

	return t.runCommandWithEnv(ctx, "twitter-cli", env, args...)
}

func (t *TwitterAdapter) executeOpenCLI(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	// OpenCLI uses browser session
	args := []string{"opencli", "twitter"}
	action := opts["action"]
	if action != "" {
		args = append(args, action)
	}
	if query != "" {
		args = append(args, query)
	}
	return t.runCommand(ctx, "opencli", args...)
}

func (t *TwitterAdapter) executeBird(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	// bird is another Twitter CLI
	args := []string{"bird"}
	action := opts["action"]
	if action == "" {
		action = "search"
	}
	args = append(args, action, query)
	return t.runCommand(ctx, "bird", args...)
}

func (t *TwitterAdapter) runCommandWithEnv(ctx context.Context, backend Backend, env []string, args ...string) (BackendResult, error) {
	start := time.Now()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Env = append(cmd.Environ(), env...)
	output, err := cmd.CombinedOutput()
	latency := time.Since(start)

	result := BackendResult{
		Backend:   backend,
		Platform:  t.name,
		Timestamp: time.Now(),
		Latency:   latency,
	}

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("%v: %s", err, string(output))
		return result, nil
	}

	result.Success = true
	result.Data = strings.TrimSpace(string(output))
	return result, nil
}

// Health checks if backends are available.
func (t *TwitterAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "twitter-cli":
		return t.checkCommand(ctx, backend, "twitter", "--version"), nil
	case "opencli":
		return t.checkCommand(ctx, backend, "opencli", "--version"), nil
	case "bird":
		return t.checkCommand(ctx, backend, "bird", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}

// Configure sets up Twitter authentication.
func (t *TwitterAdapter) Configure(creds map[string]string) error {
	if authToken, ok := creds["auth_token"]; ok {
		t.authToken = authToken
	}
	if ct0, ok := creds["ct0"]; ok {
		t.ct0 = ct0
	}
	return nil
}
