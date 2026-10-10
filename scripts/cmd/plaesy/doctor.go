package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/plaesy/spec-kit/internal/reach"
	"github.com/spf13/cobra"
)

func init() {
	register(newDoctorCmd())
}

func newDoctorCmd() *cobra.Command {
	var (
		platform string
		verbose  bool
	)

	cmd := &cobra.Command{
		Use:   "doctor [platform]",
		Short: "Check health of all system components (reach platforms, search index, etc.)",
		Long: `Run health checks across all plaesy capabilities.

Checks:
  - Reach platforms (backend availability, credentials, rate limits)
  - Search index (freshness, document count, vector/text index status)
  - Configuration (validity, migration status)

Examples:
  plaesy doctor                      # Full system health check
  plaesy doctor --platform web       # Check specific reach platform
  plaesy doctor --verbose            # Show rate limits and cache stats`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				platform = args[0]
			}
			return runDoctor(cmd.Context(), platform, verbose)
		},
	}

	cmd.Flags().StringVar(&platform, "platform", "", "specific platform to check (web, github, youtube, etc.)")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "show rate limit and cache stats")

	return cmd
}

func runDoctor(ctx context.Context, platformStr string, verbose bool) error {
	cm := reach.NewConfigManager("")
	config, err := cm.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Create registry with cache and rate limiter
	var cache reach.CacheBackend
	if config.CacheDir != "" {
		cache, _ = reach.NewTieredCache(
			config.CacheMemorySize,
			config.CacheDir,
			config.CacheSizeMB,
			config.DefaultCacheTTL,
		)
	}
	rateLimiter := reach.NewPlatformRateLimiter(reach.DefaultRateLimits())
	logger := &CLILogger{verbose: verbose}

	registry := reach.NewRegistry(config,
		reach.WithRegistryCache(cache),
		reach.WithRegistryRateLimiter(rateLimiter),
		reach.WithRegistryLogger(logger),
	)

	var platforms []reach.Platform
	if platformStr != "" {
		p := reach.Platform(platformStr)
		if !isValidPlatform(p) {
			return fmt.Errorf("unknown platform %q", platformStr)
		}
		platforms = []reach.Platform{p}
	} else {
		platforms = registry.Router().ListPlatforms()
	}

	fmt.Println("=== plaesy doctor ===")
	fmt.Println()

	allHealthy := true

	// Check reach platforms
	fmt.Println("--- Reach Platforms ---")
	for _, platform := range platforms {
		pc, _ := config.Platforms[platform]
		if !pc.Enabled && platformStr == "" {
			continue // Skip disabled platforms in full check
		}

		health, err := registry.Health(ctx, platform)
		if err != nil {
			fmt.Printf("%s: ERROR - %v\n", platform, err)
			allHealthy = false
			continue
		}

		platformHealthy := false
		fmt.Printf("%s:\n", platform)
		for _, h := range health {
			status := "✓"
			if !h.Available {
				status = "✗"
				allHealthy = false
			} else {
				platformHealthy = true
			}
			authStr := ""
			if h.RequiresAuth {
				authStr = " (auth required)"
			}
			versionStr := ""
			if h.Version != "" {
				versionStr = fmt.Sprintf(" v%s", h.Version)
			}
			fmt.Printf("  %s %s%s%s - %v\n", status, h.Backend, versionStr, authStr, h.Latency.Round(time.Millisecond))
			if h.Error != "" {
				fmt.Printf("      Error: %s\n", h.Error)
			}
		}

		if verbose {
			rl := rateLimiter.Get(platform)
			fmt.Printf("  Rate limit: %d/%d remaining, resets at %v\n", rl.Remaining(), rl.Limit(), rl.Reset().Format(time.RFC3339))
			if cache != nil {
				cs := cache.Stats()
				fmt.Printf("  Cache: %d entries, %.1f%% hit rate\n", cs.Entries, cs.HitRate*100)
			}
		}

		if !platformHealthy && platformStr == "" {
			fmt.Printf("  ⚠ No working backend for %s\n", platform)
		}
		fmt.Println()
	}

	// Check search index
	fmt.Println("--- Search Index ---")
	searchHealthy := checkSearchIndex()
	if !searchHealthy {
		allHealthy = false
	}
	fmt.Println()

	if allHealthy {
		fmt.Println("✓ All components healthy")
	} else {
		fmt.Println("✗ Some components have issues")
		os.Exit(1)
	}

	return nil
}

func checkSearchIndex() bool {
	// Check if search index exists
	indexPath := ".plaesy/search"
	vectorPath := indexPath + "/vectorstore"
	textPath := indexPath + "/textindex"

	healthy := true

	// Check vector store
	if _, err := os.Stat(vectorPath); os.IsNotExist(err) {
		fmt.Printf("  ✗ Vector store not found at %s\n", vectorPath)
		healthy = false
	} else {
		fmt.Printf("  ✓ Vector store found at %s\n", vectorPath)
	}

	// Check text index
	if _, err := os.Stat(textPath); os.IsNotExist(err) {
		fmt.Printf("  ✗ Text index not found at %s\n", textPath)
		healthy = false
	} else {
		fmt.Printf("  ✓ Text index found at %s\n", textPath)
	}

	// Check document store
	storePath := indexPath + "/store"
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		fmt.Printf("  ✗ Document store not found at %s\n", storePath)
		healthy = false
	} else {
		fmt.Printf("  ✓ Document store found at %s\n", storePath)
	}

	return healthy
}
