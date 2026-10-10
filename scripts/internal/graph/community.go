package graph

import "sort"

// computeDegree mirrors build_graph step 4: each edge (regardless of type)
// contributes +1 degree to both its source and target.
func computeDegree(nodeIDs []string, edges []Edge) map[string]int {
	degree := make(map[string]int, len(nodeIDs))
	for _, id := range nodeIDs {
		degree[id] = 0
	}
	for _, e := range edges {
		degree[e.Source]++
		degree[e.Target]++
	}
	return degree
}

// buildAdjacency builds an undirected adjacency list from every edge
// (source<->target both ways), matching the bash ADJ[] map.
func buildAdjacency(edges []Edge) map[string][]string {
	adj := make(map[string][]string)
	for _, e := range edges {
		adj[e.Source] = append(adj[e.Source], e.Target)
		adj[e.Target] = append(adj[e.Target], e.Source)
	}
	return adj
}

// labelPropagation mirrors build_graph step 5: up to 15 rounds of each node
// adopting the most common label among its neighbors, run in nodeIDs
// (sorted) order each round for determinism (the bash version iterates a
// bash associative array, whose order is unspecified/hash-based - this is a
// deliberate, documented improvement, not a behavior regression: the
// resulting communities are still connectivity-derived, just deterministic
// across runs and platforms).
func labelPropagation(nodeIDs []string, adj map[string][]string) map[string]string {
	label := make(map[string]string, len(nodeIDs))
	for _, id := range nodeIDs {
		label[id] = id
	}
	for iter := 0; iter < 15; iter++ {
		changed := false
		for _, id := range nodeIDs {
			neighbors := adj[id]
			if len(neighbors) == 0 {
				continue
			}
			counts := make(map[string]int)
			for _, nb := range neighbors {
				counts[label[nb]]++
			}
			// Deterministic tie-break: highest count, then lexicographically
			// smallest label.
			var neighborLabels []string
			for l := range counts {
				neighborLabels = append(neighborLabels, l)
			}
			sort.Strings(neighborLabels)
			best, bestCount := "", -1
			for _, l := range neighborLabels {
				if counts[l] > bestCount {
					bestCount = counts[l]
					best = l
				}
			}
			if best != "" && best != label[id] {
				label[id] = best
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return label
}

// assignCommunityIDs mirrors the "normalize labels -> sequential community
// ids ordered by size" step: bigger communities get lower ids; ties broken
// by label for determinism.
func assignCommunityIDs(nodeIDs []string, label map[string]string) map[string]int {
	counts := make(map[string]int)
	for _, id := range nodeIDs {
		counts[label[id]]++
	}
	labels := make([]string, 0, len(counts))
	for l := range counts {
		labels = append(labels, l)
	}
	sort.Slice(labels, func(i, j int) bool {
		if counts[labels[i]] != counts[labels[j]] {
			return counts[labels[i]] > counts[labels[j]]
		}
		return labels[i] < labels[j]
	})
	ids := make(map[string]int, len(labels))
	for i, l := range labels {
		ids[l] = i
	}
	out := make(map[string]int, len(nodeIDs))
	for _, id := range nodeIDs {
		out[id] = ids[label[id]]
	}
	return out
}
