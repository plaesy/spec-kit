// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"fmt"
	"os/exec"
)

// GitHubAdapter handles GitHub repository access.
type GitHubAdapter struct {
	BaseAdapter
}

func NewGitHubAdapter() *GitHubAdapter {
	return &GitHubAdapter{
		BaseAdapter: BaseAdapter{
			name:         PlatformGitHub,
			backends:     []Backend{"gh-cli", "git"},
			requiresAuth: false, // Works for public repos without auth
		},
	}
}

// Execute fetches GitHub data.
func (g *GitHubAdapter) Execute(ctx context.Context, backend Backend, query string, opts map[string]string) (BackendResult, error) {
	switch backend {
	case "gh-cli":
		return g.executeGhCli(ctx, query, opts)
	case "git":
		return g.executeGit(ctx, query, opts)
	default:
		return BackendResult{}, fmt.Errorf("unknown backend %s for github", backend)
	}
}

func (g *GitHubAdapter) executeGhCli(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	// Parse query: can be "owner/repo", "owner/repo#issue", "owner/repo/pull/123", etc.
	args := []string{"gh"}

	action := opts["action"]
	if action == "" {
		action = "view"
	}

	args = append(args, "repo", action, query)

	// Add flags based on action
	switch action {
	case "view":
		args = append(args, "--json", "name,description,url,stargazerCount,forkCount,primaryLanguage,issues,latestRelease")
	case "list-issues":
		args = append(args, "--json", "number,title,state,labels,createdAt,updatedAt")
	case "list-prs":
		args = append(args, "--json", "number,title,state,labels,createdAt,updatedAt")
	case "list-releases":
		args = append(args, "--json", "tagName,name,createdAt,isLatest,isPrerelease")
	}

	return g.runCommand(ctx, "gh-cli", args...)
}

func (g *GitHubAdapter) executeGit(ctx context.Context, query string, opts map[string]string) (BackendResult, error) {
	// Fallback: git ls-remote for basic info
	return g.runCommand(ctx, "git", "git", "ls-remote", "https://github.com/"+query)
}

// Health checks if gh CLI is available.
func (g *GitHubAdapter) Health(ctx context.Context, backend Backend) (BackendHealth, error) {
	switch backend {
	case "gh-cli":
		return g.checkCommand(ctx, backend, "gh", "--version"), nil
	case "git":
		return g.checkCommand(ctx, backend, "git", "--version"), nil
	default:
		return BackendHealth{}, fmt.Errorf("unknown backend %s", backend)
	}
}

// Configure sets up GitHub authentication.
func (g *GitHubAdapter) Configure(creds map[string]string) error {
	if token, ok := creds["token"]; ok && token != "" {
		// Configure gh auth
		cmd := exec.Command("gh", "auth", "setup-git")
		cmd.Env = append(cmd.Env, "GH_TOKEN="+token)
		return cmd.Run()
	}
	return nil
}
