package main

import (
	"github.com/plaesy/spec-kit/internal/analyze"
	"github.com/spf13/cobra"
)

func init() { register(newAnalyzeCmd()) }

func newAnalyzeCmd() *cobra.Command {
	var noGraph bool
	var noIndex bool
	var force bool
	var ifChanged bool

	cmd := &cobra.Command{
		Use:   "analyze [project_path]",
		Short: "Generate AI-optimized project analysis under .plaesy/analysis/",
		Long: `Generate AI-optimized project analysis under .plaesy/analysis/.

Also refreshes the semantic-search index that "plaesy search" queries, so an
analyze run is followed by a usable search. Use --no-index to skip that step
(it loads an embedding model and re-embeds every symbol-bearing node), or
"plaesy search --index" to rebuild the index alone.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectPath string
			if len(args) == 1 {
				projectPath = args[0]
			}
			return analyze.Run(analyze.Options{
				ProjectPath: projectPath,
				NoGraph:     noGraph,
				NoIndex:     noIndex,
				Force:       force,
				IfChanged:   ifChanged,
				Ctx:         cmd.Context(),
			})
		},
	}

	cmd.Flags().BoolVar(&noGraph, "no-graph", false, "skip dependency graph build entirely")
	cmd.Flags().BoolVar(&noIndex, "no-index", false, "skip the semantic-search index build (`plaesy search --index` still builds it on demand)")
	cmd.Flags().BoolVar(&force, "force", false, "force a full regeneration even if the project fingerprint is unchanged")
	cmd.Flags().BoolVar(&ifChanged, "if-changed", false, "legacy no-op flag: the fingerprint fast path is always on by default")
	_ = cmd.Flags().MarkHidden("if-changed")

	return cmd
}
