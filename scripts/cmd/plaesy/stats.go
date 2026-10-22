package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/plaesy/spec-kit/internal/graph"
	"github.com/plaesy/spec-kit/internal/reach"
	"github.com/plaesy/spec-kit/internal/semantic"
	"github.com/spf13/cobra"
)

func init() { register(newStatsCmd()) }

func newStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show search index and reach cache statistics",
		Long: `Display statistics about the local search index and reach cache including:
  - Vector store size
  - Text index size
  - Document store size
  - Reach cache statistics (entries, hits, misses, hit rate)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStats(cmd.Context())
		},
	}
	return cmd
}

func runStats(ctx context.Context) error {
	opts := graph.Options{RepoPath: ".", OutDir: ".plaesy/analysis"}
	paths, err := graph.ResolvePaths(opts)
	if err != nil {
		return err
	}

	sem, err := semantic.IndexStats(opts, paths)
	if err != nil {
		return err
	}
	fmt.Println("=== plaesy stats ===")
	fmt.Println("\n--- Semantic Index (built by plaesy analyze) ---")
	fmt.Printf("  Path: %s\n", paths.OutFull)
	fmt.Printf("  Documents: %d\n", sem.Documents)
	fmt.Printf("  Last built: %s\n", formatBuilt(sem.LastBuilt))
	fmt.Printf("  Stale: %t\n", sem.Stale)
	fmt.Printf("  Checkpoint present: %t\n", sem.CheckpointPresent)

	// Reach cache stats (if available)
	cm := reach.NewConfigManager("")
	config, err := cm.Load()
	if err == nil && config.CacheDir != "" {
		cache, err := reach.NewTieredCache(
			config.CacheMemorySize,
			config.CacheDir,
			config.CacheSizeMB,
			config.DefaultCacheTTL,
		)
		if err == nil {
			cs := cache.Stats()
			fmt.Println("\n--- Reach Cache ---")
			fmt.Printf("  Entries: %d\n", cs.Entries)
			fmt.Printf("  Hits: %d\n", cs.Hits)
			fmt.Printf("  Misses: %d\n", cs.Misses)
			fmt.Printf("  Hit rate: %.1f%%\n", cs.HitRate*100)
			fmt.Printf("  Memory size: %d MB\n", config.CacheMemorySize)
			fmt.Printf("  Disk size limit: %d MB\n", config.CacheSizeMB)
			cache.Close()
		}
	}

	return nil
}

func getDirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// formatBuilt renders the last-build time, or "never" when no index was built.
func formatBuilt(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format(time.RFC3339)
}
