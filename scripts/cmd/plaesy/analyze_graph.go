package main

import (
	"fmt"
	"os"

	"github.com/plaesy/spec-kit/internal/graph"
)

// graphQuery holds the graph query and export flags of `plaesy analyze`. When
// any of them is set, analyze answers the query from the graph already on disk
// and does not regenerate the analysis.
type graphQuery struct {
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
}

// active reports whether any query or export flag was given.
func (q graphQuery) active() bool {
	return q.query != "" || q.fuzzyQuery != "" || q.pathFrom != "" || q.pathTo != "" ||
		q.explainNode != "" || q.impactNode != "" || q.semanticQueue || q.businessLogic ||
		q.generatePaths || q.applySemantic != ""
}

// runGraphQuery answers one query or export from the graph on disk.
func runGraphQuery(opts graph.Options, paths graph.Paths, q graphQuery) error {
	switch {
	case q.explainNode != "":
		g, err := loadOrError(paths)
		if err != nil {
			return err
		}
		graph.DoExplain(g, q.explainNode, os.Stdout)
	case q.impactNode != "":
		g, err := loadOrError(paths)
		if err != nil {
			return err
		}
		if q.impactVisualize {
			fmt.Fprintf(os.Stderr, "[plaesy-analyze] Generating impact visualization for '%s'...\n", q.impactNode)
			if err := graph.DoImpactVisualize(q.impactNode, paths.OutFull); err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "[plaesy-analyze] Impact visualization written to impact-visualization.html")
		} else {
			graph.DoImpactCheck(g, q.impactNode, q.impactDepth, os.Stdout)
		}
	case q.pathFrom != "" && q.pathTo != "":
		g, err := loadOrError(paths)
		if err != nil {
			return err
		}
		graph.DoPathQuery(g, q.pathFrom, q.pathTo, os.Stdout)
	case q.businessLogic:
		g, err := loadOrError(paths)
		if err != nil {
			return err
		}
		count, err := graph.DoBusinessLogic(g, paths.OutFull)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "[plaesy-analyze] Generating business logic queue with %d inferred edges...\n", count)
		fmt.Fprintf(os.Stderr, "[plaesy-analyze] Business logic queue written to %s\n", paths.OutFull+string(os.PathSeparator)+"business-logic-queue.json")
	case q.generatePaths:
		g, err := loadOrError(paths)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "[plaesy-analyze] Generating learning paths via topological analysis...")
		if err := graph.DoGeneratePaths(g, paths.OutFull); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "[plaesy-analyze] Learning paths written to %s\n", paths.OutFull+string(os.PathSeparator)+"learning-paths.json")
	case q.fuzzyQuery != "":
		g, err := loadOrError(paths)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "[plaesy-analyze] Fuzzy search for '%s' (search depth: %d):\n", q.fuzzyQuery, q.searchDepth)
		graph.DoFuzzyQuery(g, q.fuzzyQuery, os.Stderr)
		fmt.Fprintln(os.Stderr)
	case q.semanticQueue:
		g, err := loadOrError(paths)
		if err != nil {
			return err
		}
		count, err := graph.DoSemanticQueue(g, paths.OutFull)
		if err != nil {
			return err
		}
		if count == 0 {
			fmt.Fprintln(os.Stderr, "[plaesy-analyze] No INFERRED edges to review - nothing queued.")
			return nil
		}
		fmt.Fprintf(os.Stderr, "[plaesy-analyze] Wrote %d INFERRED edges to %s\n", count, paths.OutFull+string(os.PathSeparator)+"semantic-queue.json")
		fmt.Fprintln(os.Stderr, "[plaesy-analyze] Ask your AI assistant to explain each pair, save the result as")
		fmt.Fprintln(os.Stderr, "  a JSON array of {source,target,rationale}, then run --apply-semantic <file>.")
	case q.applySemantic != "":
		g, err := loadOrError(paths)
		if err != nil {
			return err
		}
		if _, err := os.Stat(q.applySemantic); err != nil {
			// A missing input file is a user error, not an empty result.
			return fmt.Errorf("annotations file not found: %s: %w", q.applySemantic, err)
		}
		applied, err := graph.DoApplySemantic(g, q.applySemantic, paths.OutFull)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "[plaesy-analyze] Applied rationale to %d edge(s). reports.md updated.\n", applied)
	case q.query != "":
		g, err := loadOrError(paths)
		if err != nil {
			return err
		}
		graph.DoQuery(g, q.query, os.Stdout)
	}
	return nil
}

// loadOrError requires an already-built graph on disk for the query flags.
func loadOrError(paths graph.Paths) (*graph.Graph, error) {
	if _, err := os.Stat(paths.ProjectJSON); err != nil {
		return nil, fmt.Errorf("no graph found in %s. Run `plaesy analyze` without query flags first to build it", paths.OutFull)
	}
	return graph.LoadJSON(paths.ProjectJSON)
}
