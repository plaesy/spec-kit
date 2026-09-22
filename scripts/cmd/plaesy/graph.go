package main

import (
	"fmt"
	"os"
	"time"

	"github.com/plaesy/spec-kit/internal/graph"
	"github.com/spf13/cobra"
)

func init() { register(newGraphCmd()) }

func newGraphCmd() *cobra.Command {
	var (
		repoPath        string
		outDir          string
		query           string
		fuzzyQuery      string
		searchDepth     int
		pathFrom        string
		pathTo          string
		explainNode     string
		impactNode      string
		impactDepth     int
		impactVisualize bool
		semanticQueue   bool
		businessLogic   bool
		generatePaths   bool
		applySemantic   string
		watch           bool
		watchInterval   int
		ifChanged       bool
	)

	cmd := &cobra.Command{
		Use:   "graph [project_path]",
		Short: "Build (or query) a lightweight knowledge graph of a project",
		Long: `Scans a project's source/doc/config files, extracts structural references
(markdown links, dot-sourcing/imports, calls) plus plain-text "mentions" and
per-file symbols, runs community detection, and emits project.graph.json,
project.html (a force-directed viz) and reports.md under --outdir.

Also invoked by "plaesy analyze" for the combined analyze+graph pipeline, but
works standalone: "plaesy graph <project_path>".`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				repoPath = args[0]
			}
			opts := graph.Options{RepoPath: repoPath, OutDir: outDir}
			paths, err := graph.ResolvePaths(opts)
			if err != nil {
				return err
			}

			switch {
			case watch:
				return runWatch(opts, paths, watchInterval)
			case explainNode != "":
				g, err := loadOrError(paths)
				if err != nil {
					return err
				}
				graph.DoExplain(g, explainNode, os.Stdout)
			case impactNode != "":
				g, err := loadOrError(paths)
				if err != nil {
					return err
				}
				if impactVisualize {
					fmt.Fprintf(os.Stderr, "[plaesy-graph] Generating impact visualization for '%s'...\n", impactNode)
					if err := graph.DoImpactVisualize(impactNode, paths.OutFull); err != nil {
						return err
					}
					fmt.Fprintln(os.Stderr, "[plaesy-graph] Impact visualization written to impact-visualization.html")
				} else {
					graph.DoImpactCheck(g, impactNode, impactDepth, os.Stdout)
				}
			case pathFrom != "" && pathTo != "":
				g, err := loadOrError(paths)
				if err != nil {
					return err
				}
				graph.DoPathQuery(g, pathFrom, pathTo, os.Stdout)
			case businessLogic:
				g, err := loadOrError(paths)
				if err != nil {
					return err
				}
				count, err := graph.DoBusinessLogic(g, paths.OutFull)
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "[plaesy-graph] Generating business logic queue with %d inferred edges...\n", count)
				fmt.Fprintf(os.Stderr, "[plaesy-graph] Business logic queue written to %s\n", paths.OutFull+string(os.PathSeparator)+"business-logic-queue.json")
			case generatePaths:
				g, err := loadOrError(paths)
				if err != nil {
					return err
				}
				fmt.Fprintln(os.Stderr, "[plaesy-graph] Generating learning paths via topological analysis...")
				if err := graph.DoGeneratePaths(g, paths.OutFull); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "[plaesy-graph] Learning paths written to %s\n", paths.OutFull+string(os.PathSeparator)+"learning-paths.json")
			case fuzzyQuery != "":
				g, err := loadOrError(paths)
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "[plaesy-graph] Fuzzy search for '%s' (search depth: %d):\n", fuzzyQuery, searchDepth)
				graph.DoFuzzyQuery(g, fuzzyQuery, os.Stderr)
				fmt.Fprintln(os.Stderr)
			case semanticQueue:
				g, err := loadOrError(paths)
				if err != nil {
					return err
				}
				count, err := graph.DoSemanticQueue(g, paths.OutFull)
				if err != nil {
					return err
				}
				if count == 0 {
					fmt.Fprintln(os.Stderr, "[plaesy-graph] No INFERRED edges to review - nothing queued.")
					return nil
				}
				fmt.Fprintf(os.Stderr, "[plaesy-graph] Wrote %d INFERRED edges to %s\n", count, paths.OutFull+string(os.PathSeparator)+"semantic-queue.json")
				fmt.Fprintln(os.Stderr, "[plaesy-graph] Ask your AI assistant to explain each pair, save the result as")
				fmt.Fprintln(os.Stderr, "  a JSON array of {source,target,rationale}, then run --apply-semantic <file>.")
			case applySemantic != "":
				g, err := loadOrError(paths)
				if err != nil {
					return err
				}
				if _, err := os.Stat(applySemantic); err != nil {
					fmt.Fprintf(os.Stderr, "[plaesy-graph] Annotations file not found: %s\n", applySemantic)
					return nil
				}
				applied, err := graph.DoApplySemantic(g, applySemantic, paths.OutFull)
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "[plaesy-graph] Applied rationale to %d edge(s). reports.md updated.\n", applied)
			case query != "":
				g, err := loadOrError(paths)
				if err != nil {
					return err
				}
				graph.DoQuery(g, query, os.Stdout)
			case ifChanged:
				current, err := graph.SourceFingerprint(opts, paths)
				if err != nil {
					return err
				}
				stored := ""
				if b, err := os.ReadFile(paths.FingerprintFile); err == nil {
					stored = string(b)
				}
				if _, err := os.Stat(paths.ProjectJSON); err == nil && stored != "" && stored == current {
					fmt.Fprintln(os.Stderr, "[plaesy-graph] No source changes detected since last build - skipping rebuild.")
					fmt.Fprintf(os.Stderr, "[plaesy-graph] Output: %s\n", paths.OutFull)
					return nil
				}
				return runBuild(opts, paths)
			default:
				return runBuild(opts, paths)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&repoPath, "path", ".", "project directory to scan")
	cmd.Flags().StringVar(&outDir, "outdir", ".plaesy/analysis", "output directory (relative to --path)")
	cmd.Flags().StringVar(&query, "query", "", "keyword query against node ids")
	cmd.Flags().StringVar(&fuzzyQuery, "fuzzy-query", "", "fuzzy/substring search against node ids")
	cmd.Flags().IntVar(&searchDepth, "search-depth", 2, "search depth for --fuzzy-query (display only)")
	cmd.Flags().StringVar(&pathFrom, "path-query-from", "", "shortest-path query: source node")
	cmd.Flags().StringVar(&pathTo, "path-query-to", "", "shortest-path query: target node")
	cmd.Flags().StringVar(&explainNode, "explain", "", "explain one node's edges and symbols")
	cmd.Flags().StringVar(&impactNode, "impact-check", "", "list nodes impacted by changing this node")
	cmd.Flags().IntVar(&impactDepth, "impact-depth", 2, "BFS depth for --impact-check")
	cmd.Flags().BoolVar(&impactVisualize, "impact-visualize", false, "write impact-visualization.html instead of printing")
	cmd.Flags().BoolVar(&semanticQueue, "semantic-queue", false, "export INFERRED edges to semantic-queue.json")
	cmd.Flags().BoolVar(&businessLogic, "business-logic", false, "export INFERRED edges to business-logic-queue.json")
	cmd.Flags().BoolVar(&generatePaths, "generate-paths", false, "write a learning-paths.json stub")
	cmd.Flags().StringVar(&applySemantic, "apply-semantic", "", "merge a {source,target,rationale}[] annotations file into reports.md")
	cmd.Flags().BoolVar(&watch, "watch", false, "rebuild automatically on source changes")
	cmd.Flags().IntVar(&watchInterval, "watch-interval", 3, "poll interval in seconds for --watch")
	cmd.Flags().BoolVar(&ifChanged, "if-changed", false, "rebuild only if source files changed since last build")

	return cmd
}

// runBuild mirrors build_graph()'s top-level orchestration: scan, save
// project.graph.json/reports.md/project.html, print the summary, and record
// a fingerprint for --if-changed/--watch.
func runBuild(opts graph.Options, paths graph.Paths) error {
	g, err := graph.Build(opts, paths)
	if err != nil {
		return err
	}
	if err := graph.SaveJSON(g, paths.ProjectJSON); err != nil {
		return err
	}
	if err := graph.WriteReport(g, paths.OutFull, ""); err != nil {
		return err
	}
	if err := graph.WriteHTML(g, paths.OutFull); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "[plaesy-graph] Done. Nodes: %d  Edges: %d\n", len(g.Nodes), len(g.Edges))
	fmt.Fprintf(os.Stderr, "[plaesy-graph] Output: %s\n", paths.OutFull)

	if fp, err := graph.SourceFingerprint(opts, paths); err == nil {
		_ = os.WriteFile(paths.FingerprintFile, []byte(fp), 0o644)
	}
	return nil
}

// loadOrError mirrors require_graph(): the query subcommands need an
// already-built graph on disk.
func loadOrError(paths graph.Paths) (*graph.Graph, error) {
	if _, err := os.Stat(paths.ProjectJSON); err != nil {
		return nil, fmt.Errorf("no graph found in %s. Run without flags first to build it", paths.OutFull)
	}
	return graph.LoadJSON(paths.ProjectJSON)
}

// runWatch mirrors do_watch(): build once, then poll the source fingerprint
// and rebuild whenever it changes.
func runWatch(opts graph.Options, paths graph.Paths, intervalSeconds int) error {
	fmt.Fprintf(os.Stderr, "[plaesy-graph] Watch mode: rebuilding now, then polling every %ds. Ctrl+C to stop.\n", intervalSeconds)
	if err := runBuild(opts, paths); err != nil {
		return err
	}
	lastFP, err := graph.SourceFingerprint(opts, paths)
	if err != nil {
		return err
	}
	for {
		time.Sleep(time.Duration(intervalSeconds) * time.Second)
		fp, err := graph.SourceFingerprint(opts, paths)
		if err != nil {
			return err
		}
		if fp != lastFP {
			fmt.Fprintln(os.Stderr, "[plaesy-graph] Change detected, rebuilding...")
			if err := runBuild(opts, paths); err != nil {
				return err
			}
			lastFP, err = graph.SourceFingerprint(opts, paths)
			if err != nil {
				return err
			}
		}
	}
}
