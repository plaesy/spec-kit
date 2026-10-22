package main

import (
	"time"

	"github.com/plaesy/spec-kit/internal/analyze"
	"github.com/plaesy/spec-kit/internal/graph"
	"github.com/spf13/cobra"
)

func init() { register(newAnalyzeCmd()) }

func newAnalyzeCmd() *cobra.Command {
	var noGraph bool
	var noIndex bool
	var force bool
	var ifChanged bool
	var watch bool
	var watchInterval int
	var q graphQuery

	cmd := &cobra.Command{
		Use:   "analyze [project_path]",
		Short: "Analyze a project: documentation, dependency graph, and search indexes",
		Long: `Generate AI-optimized project analysis under .plaesy/analysis/.

Also refreshes the semantic-search index that "plaesy search" queries,
so an analyze run is followed by a usable search. Use --no-index to skip that
step (it loads an embedding model and re-embeds every symbol-bearing node).

With --watch, analyze runs once and then again whenever source files change.

The query and export flags (--query, --explain, --impact-check, and the rest)
answer from the graph already on disk and do not regenerate the analysis. Run
analyze once without them first.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectPath string
			if len(args) == 1 {
				projectPath = args[0]
			}
			if q.active() {
				opts := graph.Options{RepoPath: projectPath, OutDir: ".plaesy/analysis"}
				if opts.RepoPath == "" {
					opts.RepoPath = "."
				}
				paths, err := graph.ResolvePaths(opts)
				if err != nil {
					return err
				}
				return runGraphQuery(opts, paths, q)
			}
			opts := analyze.Options{
				ProjectPath: projectPath,
				NoGraph:     noGraph,
				NoIndex:     noIndex,
				Force:       force,
				IfChanged:   ifChanged,
				Ctx:         cmd.Context(),
			}
			if watch {
				return analyze.Watch(opts, time.Duration(watchInterval)*time.Second)
			}
			return analyze.Run(opts)
		},
	}

	cmd.Flags().BoolVar(&noGraph, "no-graph", false, "skip dependency graph build entirely")
	cmd.Flags().BoolVar(&noIndex, "no-index", false, "skip the semantic-search index build (`plaesy analyze` still builds it on demand)")
	cmd.Flags().BoolVar(&force, "force", false, "force a full regeneration even if the project fingerprint is unchanged")
	cmd.Flags().BoolVar(&ifChanged, "if-changed", false, "legacy no-op flag: the fingerprint fast path is always on by default")
	_ = cmd.Flags().MarkHidden("if-changed")
	cmd.Flags().BoolVar(&watch, "watch", false, "re-run analyze automatically when source files change")
	cmd.Flags().IntVar(&watchInterval, "watch-interval", 3, "poll interval in seconds for --watch")

	cmd.Flags().StringVar(&q.query, "query", "", "keyword query against node ids")
	cmd.Flags().StringVar(&q.fuzzyQuery, "fuzzy-query", "", "fuzzy/substring search against node ids")
	cmd.Flags().IntVar(&q.searchDepth, "search-depth", 2, "search depth for --fuzzy-query (display only)")
	cmd.Flags().StringVar(&q.pathFrom, "path-query-from", "", "shortest-path query: source node")
	cmd.Flags().StringVar(&q.pathTo, "path-query-to", "", "shortest-path query: target node")
	cmd.Flags().StringVar(&q.explainNode, "explain", "", "explain one node's edges and symbols")
	cmd.Flags().StringVar(&q.impactNode, "impact-check", "", "list nodes impacted by changing this node")
	cmd.Flags().IntVar(&q.impactDepth, "impact-depth", 2, "BFS depth for --impact-check")
	cmd.Flags().BoolVar(&q.impactVisualize, "impact-visualize", false, "write impact-visualization.html instead of printing")
	cmd.Flags().BoolVar(&q.semanticQueue, "semantic-queue", false, "export INFERRED edges to semantic-queue.json")
	cmd.Flags().BoolVar(&q.businessLogic, "business-logic", false, "export INFERRED edges to business-logic-queue.json")
	cmd.Flags().BoolVar(&q.generatePaths, "generate-paths", false, "write a learning-paths.json stub")
	cmd.Flags().StringVar(&q.applySemantic, "apply-semantic", "", "merge a {source,target,rationale}[] annotations file into reports.md")

	return cmd
}
