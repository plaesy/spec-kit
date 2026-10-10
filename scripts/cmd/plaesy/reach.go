package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/plaesy/spec-kit/internal/reach"
	"github.com/spf13/cobra"
)

func init() { register(newReachCmd()) }

func newReachCmd() *cobra.Command {
	var (
		repoPath   string
		outDir     string
		platform   string
		query      string
		action     string
		limit      int
		index      bool
		timeout    string
		cacheTTL   string
		noCache    bool
		verbose    bool
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "reach [platform] [query]",
		Short: "Access external data sources across multiple platforms",
		Long: `Access external data from web, social media, code hosting, video platforms,
RSS feeds, and search engines with multi-backend routing, caching, rate limiting, and automatic fallback.

Platforms:
  web           Read any web page (Jina Reader)
  youtube       YouTube transcripts and search (yt-dlp)
  github        GitHub repos, issues, PRs (gh CLI)
  rss           RSS/Atom feed parsing (builtin + feedparser)
  twitter       Twitter/X search and tweets (twitter-cli, OpenCLI)
  reddit        Reddit search and posts (OpenCLI, rdt-cli)
  bilibili      Bilibili search and video info (bili-cli)
  xiaohongshu   XiaoHongShu search (OpenCLI, xiaohongshu-mcp)
  linkedin      LinkedIn profiles and search (mcp-linkedin, Jina Reader)
  exa           Semantic web search (Exa via MCP)
  v2ex          V2EX forum topics (V2EX API)
  xueqiu        Xueqiu stock forum (Xueqiu API)
  xiaoyuzhou    Xiaoyuzhou podcasts (xiaoyuzhou-mcp)
  bosszhipin    Boss直聘 job search (boss-cdp)
  facebook      Facebook (OpenCLI)
  instagram     Instagram (OpenCLI)

Examples:
  plaesy reach web "https://example.com"
  plaesy reach youtube "golang tutorial" --action search
  plaesy reach github "plaesy/spec-kit" --action view
  plaesy reach rss "https://example.com/feed.xml"
  plaesy reach twitter "golang" --action search --limit 10
  plaesy reach exa "best practices for go testing" --num-results 5
  plaesy reach configure twitter         # Configure credentials
  plaesy reach list                      # List all platforms`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if index {
				return runReachIndex(cmd.Context(), repoPath, outDir)
			}

			if len(args) == 0 {
				return cmd.Help()
			}

			if args[0] == "configure" {
				if len(args) < 2 {
					return fmt.Errorf("configure requires a platform: plaesy reach configure <platform>")
				}
				return runReachConfigure(cmd.Context(), args[1])
			}

			platform = args[0]
			query = ""
			if len(args) > 1 {
				query = args[1]
			}

			return runReachQuery(cmd.Context(), platform, query, action, limit, timeout, cacheTTL, noCache, verbose, jsonOutput)
		},
	}

	cmd.Flags().StringVar(&repoPath, "path", ".", "project directory")
	cmd.Flags().StringVar(&outDir, "outdir", ".plaesy/analysis", "output directory")
	cmd.Flags().StringVar(&action, "action", "", "action for platform (search, view, etc.)")
	cmd.Flags().IntVar(&limit, "limit", 10, "max results for search actions")
	cmd.Flags().BoolVar(&index, "index", false, "rebuild semantic index (for internal use)")
	cmd.Flags().StringVar(&timeout, "timeout", "", "query timeout (e.g. 30s, 1m)")
	cmd.Flags().StringVar(&cacheTTL, "cache-ttl", "", "cache TTL (e.g. 5m, 1h)")
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "disable caching for this query")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "verbose output")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	// Subcommands
	cmd.AddCommand(newReachConfigureCmd())
	cmd.AddCommand(newReachListCmd())

	return cmd
}

func runReachQuery(ctx context.Context, platformStr, query, action string, limit int, timeout, cacheTTL string, noCache, verbose, jsonOutput bool) error {
	platform := reach.Platform(platformStr)
	if !isValidPlatform(platform) {
		return fmt.Errorf("unknown platform %q. Run 'plaesy reach list' for available platforms", platformStr)
	}

	// Load configuration
	cm := reach.NewConfigManager("")
	config, err := cm.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Create cache
	var cache reach.CacheBackend
	if config.CacheDir != "" {
		cache, err = reach.NewTieredCache(
			config.CacheMemorySize,
			config.CacheDir,
			config.CacheSizeMB,
			config.DefaultCacheTTL,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create cache: %v\n", err)
		}
	}

	// Create rate limiter
	rateLimiter := reach.NewPlatformRateLimiter(reach.DefaultRateLimits())

	// Create logger
	logger := &CLILogger{verbose: verbose}

	// Create registry with all features
	registry := reach.NewRegistry(config,
		reach.WithRegistryCache(cache),
		reach.WithRegistryRateLimiter(rateLimiter),
		reach.WithRegistryLogger(logger),
	)

	// Load saved credentials
	if err := registry.LoadConfig(ctx); err != nil {
		return fmt.Errorf("loading credentials: %w", err)
	}

	// Prepare options
	opts := make(map[string]string)
	if action != "" {
		opts["action"] = action
	}
	if limit > 0 {
		opts["limit"] = fmt.Sprintf("%d", limit)
	}
	if timeout != "" {
		opts["timeout"] = timeout
	}
	if cacheTTL != "" {
		opts["cache_ttl"] = cacheTTL
	}
	if noCache {
		opts["cache"] = "false"
	}

	// Execute query
	if verbose {
		fmt.Fprintf(os.Stderr, "[plaesy-reach] Querying %s via %s...\n", platform, query)
	}
	start := time.Now()
	result, err := registry.Query(ctx, platform, query, opts)
	elapsed := time.Since(start)

	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("backend %s failed: %s", result.Backend, result.Error)
	}

	if jsonOutput {
		output := map[string]any{
			"platform":  platform,
			"backend":   result.Backend,
			"success":   result.Success,
			"data":      result.Data,
			"latency":   result.Latency.String(),
			"timestamp": result.Timestamp.Format(time.RFC3339),
		}
		jsonData, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonData))
	} else {
		if verbose {
			fmt.Fprintf(os.Stderr, "[plaesy-reach] Success via %s (%.2fs)\n", result.Backend, elapsed.Seconds())
		}
		fmt.Println(result.Data)
	}

	return nil
}

func runReachConfigure(ctx context.Context, platformStr string) error {
	platform := reach.Platform(platformStr)
	if !isValidPlatform(platform) {
		return fmt.Errorf("unknown platform %q", platformStr)
	}

	keys := reach.CredentialKeys(platform)
	if len(keys) == 0 {
		fmt.Printf("Platform %s does not require configuration (zero-config).\n", platform)
		return nil
	}

	fmt.Printf("Configuring %s...\n", platform)
	fmt.Printf("Required credentials: %s\n", strings.Join(keys, ", "))
	fmt.Println("Enter values (press Enter to skip):")

	creds := make(map[string]string)
	for _, key := range keys {
		fmt.Printf("  %s: ", key)
		var value string
		fmt.Scanln(&value)
		if value != "" {
			creds[key] = value
		}
	}

	cm := reach.NewConfigManager("")
	config, err := cm.Load()
	if err != nil {
		return err
	}

	pc := config.Platforms[platform]
	pc.Enabled = true
	if pc.Credentials == nil {
		pc.Credentials = make(map[string]string)
	}
	for k, v := range creds {
		pc.Credentials[k] = v
	}
	config.Platforms[platform] = pc

	if err := cm.Save(config); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("✓ Configuration saved for %s\n", platform)
	return nil
}

func runReachList(cmd *cobra.Command, args []string) error {
	fmt.Println("Available platforms:")
	for _, p := range reach.AllPlatforms {
		zeroConfig := false
		for _, zc := range reach.ZeroConfigPlatforms {
			if zc == p {
				zeroConfig = true
				break
			}
		}
		status := ""
		if zeroConfig {
			status = " (zero-config)"
		} else {
			status = " (requires auth)"
		}
		fmt.Printf("  %s%s\n", p, status)
	}
	return nil
}

func newReachConfigureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure <platform>",
		Short: "Configure credentials for a platform",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReachConfigure(cmd.Context(), args[0])
		},
	}
	return cmd
}

func newReachListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available platforms",
		Args:  cobra.NoArgs,
		RunE:  runReachList,
	}
	return cmd
}

func runReachIndex(ctx context.Context, repoPath, outDir string) error {
	return fmt.Errorf("use 'plaesy search --index' to rebuild semantic index")
}

func isValidPlatform(p reach.Platform) bool {
	for _, valid := range reach.AllPlatforms {
		if valid == p {
			return true
		}
	}
	return false
}

// CLILogger implements the Logger interface for CLI output.
type CLILogger struct {
	verbose bool
}

func (l *CLILogger) Debug(msg string, args ...any) {
	if l.verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s %v\n", msg, args)
	}
}

func (l *CLILogger) Info(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "[INFO] %s %v\n", msg, args)
}

func (l *CLILogger) Warn(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "[WARN] %s %v\n", msg, args)
}

func (l *CLILogger) Error(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "[ERROR] %s %v\n", msg, args)
}
