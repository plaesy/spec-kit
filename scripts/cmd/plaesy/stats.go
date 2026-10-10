package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/plaesy/spec-kit/internal/reach"
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
	indexPath := ".plaesy/search"
	vectorPath := indexPath + "/vectorstore"
	textPath := indexPath + "/textindex"
	storePath := indexPath + "/store"

	fmt.Println("=== plaesy stats ===")
	fmt.Printf("Index directory: %s\n\n", indexPath)

	// Vector store stats
	if _, err := os.Stat(vectorPath); err == nil {
		fmt.Println("--- Vector Store ---")
		fmt.Printf("  Path: %s\n", vectorPath)
		fmt.Printf("  Size: %d bytes\n", getDirSize(vectorPath))
	} else {
		fmt.Println("--- Vector Store ---")
		fmt.Printf("  Not initialized (run 'plaesy search index')\n")
	}

	// Text index stats
	if _, err := os.Stat(textPath); err == nil {
		fmt.Println("\n--- Text Index (Bleve) ---")
		fmt.Printf("  Path: %s\n", textPath)
		fmt.Printf("  Size: %d bytes\n", getDirSize(textPath))
	} else {
		fmt.Println("\n--- Text Index (Bleve) ---")
		fmt.Printf("  Not initialized (run 'plaesy search index')\n")
	}

	// Document store stats
	if _, err := os.Stat(storePath); err == nil {
		fmt.Println("\n--- Document Store ---")
		fmt.Printf("  Path: %s\n", storePath)
		fmt.Printf("  Size: %d bytes\n", getDirSize(storePath))
	} else {
		fmt.Println("\n--- Document Store ---")
		fmt.Printf("  Not initialized (run 'plaesy search index')\n")
	}

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
