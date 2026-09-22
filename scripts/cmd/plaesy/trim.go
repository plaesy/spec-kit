package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/plaesy/spec-kit/internal/trimmer"
	"github.com/spf13/cobra"
)

func init() { register(newTrimCmd()) }

// trimPaths bundles the repo-root-relative paths plaesy-trim.sh derives up
// front (RULES_PATH, GRAPH_PATH, STATS_PATH), resolved once per invocation.
type trimPaths struct {
	repoRoot string
	rules    *trimmer.Rules
	graph    string
	stats    string
}

func resolveTrimPaths() trimPaths {
	repoRoot, err := os.Getwd()
	if err != nil {
		repoRoot = "."
	}
	// Best-effort "framework root": the source tree this binary was built
	// from, when running via `go run`/`go build` from within the repo.
	// Unlike plaesy-trim.sh (a script that always knows its own directory),
	// a compiled binary can be installed anywhere, so this is a fallback
	// only - see trimmer.LoadRules for the final baked-in default.
	rulesPath := trimmer.ResolveRulesPath(repoRoot, repoRoot)
	return trimPaths{
		repoRoot: repoRoot,
		rules:    trimmer.LoadRules(rulesPath),
		graph:    filepath.Join(repoRoot, ".plaesy", "analysis", "project.graph.json"),
		stats:    filepath.Join(repoRoot, ".plaesy", "memory", "token-stats.json"),
	}
}

func newTrimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trim",
		Short: "Token/context compression for command output, memory and instruction files",
		Long: `Token/context compression for the plaesy spec-kit repo (or any project it is run against).

Layer 1 (command output): trim run <command...>              - per-tool aware dedupe/truncate/filter
Layer 2 (memory/instruction files): trim compress --path ...  - heuristic denser style (fast, local)
  llm-queue / apply-llm                                         - LLM-quality rewrite via the calling
                                                                   assistant, no API key required
Report: trim report                                           - cumulative savings across layers`,
	}

	cmd.AddCommand(newTrimRunCmd(), newTrimCompressCmd(), newTrimLLMQueueCmd(), newTrimApplyLLMCmd(), newTrimReportCmd())
	return cmd
}

func newTrimRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "run -- <command...>",
		Short:              "Layer 1: run a command and compress its output",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// DisableFlagParsing means args may still include a leading "--";
			// cobra itself only strips that when flag parsing is enabled.
			if len(args) > 0 && args[0] == "--" {
				args = args[1:]
			}
			if len(args) == 0 {
				return fmt.Errorf("usage: plaesy trim run <command...>")
			}
			paths := resolveTrimPaths()
			result, err := trimmer.Run(paths.rules, paths.stats, args)
			if err != nil {
				return err
			}
			fmt.Print(result.Output)
			return nil
		},
	}
}

func newTrimCompressCmd() *cobra.Command {
	var path, level string
	var recurse, dryRun bool
	c := &cobra.Command{
		Use:   "compress",
		Short: "Layer 2: fast heuristic prose compression",
		RunE: func(cmd *cobra.Command, args []string) error {
			if path == "" {
				return fmt.Errorf("usage: plaesy trim compress --path <file|dir> [--level lite|full|ultra] [--recurse] [--dry-run]")
			}
			paths := resolveTrimPaths()
			results, err := trimmer.CompressPath(paths.rules, paths.repoRoot, paths.graph, paths.stats, path, level, recurse, dryRun)
			if err != nil {
				return err
			}
			for _, r := range results {
				fmt.Print(r.String())
			}
			return nil
		},
	}
	c.Flags().StringVar(&path, "path", "", "file or directory to compress")
	c.Flags().StringVar(&level, "level", "", "lite|full|ultra (auto-resolved from the graph when omitted)")
	c.Flags().BoolVar(&recurse, "recurse", false, "recurse into subdirectories (directory targets only)")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "report savings without writing changes")
	return c
}

func newTrimLLMQueueCmd() *cobra.Command {
	var path string
	c := &cobra.Command{
		Use:   "llm-queue",
		Short: "Layer 2 (LLM mode): export prose segments to rewrite",
		RunE: func(cmd *cobra.Command, args []string) error {
			if path == "" {
				return fmt.Errorf("usage: plaesy trim llm-queue --path <file>")
			}
			paths := resolveTrimPaths()
			queuePath, count, err := trimmer.WriteLLMQueue(paths.repoRoot, path)
			if err != nil {
				return err
			}
			relative, relErr := filepath.Rel(paths.repoRoot, path)
			if relErr != nil {
				relative = path
			}
			fmt.Printf("Wrote %d prose segment(s) from %s to %s\n", count, relative, queuePath)
			fmt.Println("Next: have the calling assistant rewrite each segment.text denser (same meaning, no code/commands touched),")
			fmt.Println("save as {file, segments:[{index, text}]} JSON, then run:")
			fmt.Printf("  plaesy trim apply-llm --path %s --annotations <annotations.json>\n", path)
			return nil
		},
	}
	c.Flags().StringVar(&path, "path", "", "file to export prose segments from")
	return c
}

func newTrimApplyLLMCmd() *cobra.Command {
	var path, annotations string
	c := &cobra.Command{
		Use:   "apply-llm",
		Short: "Layer 2 (LLM mode): merge rewritten segments back",
		RunE: func(cmd *cobra.Command, args []string) error {
			if path == "" || annotations == "" {
				return fmt.Errorf("usage: plaesy trim apply-llm --path <file> --annotations <annotations.json>")
			}
			paths := resolveTrimPaths()
			result, err := trimmer.ApplyLLM(paths.repoRoot, paths.stats, path, annotations)
			if err != nil {
				return err
			}
			fmt.Printf("%s [llm]\n  %d tokens -> %d tokens  (%.1f%% saved)\n", result.Relative, result.Before, result.After, trimmer.PctSaved(result.Before, result.After))
			return nil
		},
	}
	c.Flags().StringVar(&path, "path", "", "file to apply rewritten segments to")
	c.Flags().StringVar(&annotations, "annotations", "", "JSON file of rewritten segments")
	return c
}

func newTrimReportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Cumulative savings across layers",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := resolveTrimPaths()
			out, err := trimmer.Report(paths.stats)
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}
}
