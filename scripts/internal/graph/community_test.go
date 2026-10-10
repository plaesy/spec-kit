package graph

import (
	"reflect"
	"sort"
	"testing"
)

func TestComputeDegree(t *testing.T) {
	tests := []struct {
		name    string
		nodeIDs []string
		edges   []Edge
		want    map[string]int
	}{
		{
			name:    "isolated nodes get zero",
			nodeIDs: []string{"a", "b"},
			want:    map[string]int{"a": 0, "b": 0},
		},
		{
			name:    "each edge counts once for both endpoints",
			nodeIDs: []string{"a", "b", "c"},
			edges:   []Edge{{Source: "a", Target: "b"}, {Source: "b", Target: "c"}},
			want:    map[string]int{"a": 1, "b": 2, "c": 1},
		},
		{
			name:    "self loop adds two to the same node",
			nodeIDs: []string{"a"},
			edges:   []Edge{{Source: "a", Target: "a"}},
			want:    map[string]int{"a": 2},
		},
		{
			name:    "parallel edges accumulate",
			nodeIDs: []string{"a", "b"},
			edges:   []Edge{{Source: "a", Target: "b", Type: "imports"}, {Source: "a", Target: "b", Type: "mentions"}},
			want:    map[string]int{"a": 2, "b": 2},
		},
		{
			name:    "edge to unknown node creates implicit entry",
			nodeIDs: []string{"a"},
			edges:   []Edge{{Source: "a", Target: "z"}},
			want:    map[string]int{"a": 1, "z": 1},
		},
		{
			name:    "nil inputs yield empty map",
			nodeIDs: nil,
			edges:   nil,
			want:    map[string]int{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeDegree(tc.nodeIDs, tc.edges)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("computeDegree() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBuildAdjacency(t *testing.T) {
	t.Run("undirected both directions", func(t *testing.T) {
		adj := buildAdjacency([]Edge{{Source: "a", Target: "b"}, {Source: "b", Target: "c"}})
		for node, want := range map[string][]string{"a": {"b"}, "b": {"a", "c"}, "c": {"b"}} {
			if !reflect.DeepEqual(adj[node], want) {
				t.Errorf("adj[%q] = %v, want %v", node, adj[node], want)
			}
		}
	})

	t.Run("no edges yields no keys", func(t *testing.T) {
		if adj := buildAdjacency(nil); len(adj) != 0 {
			t.Errorf("buildAdjacency(nil) = %v, want empty", adj)
		}
	})

	t.Run("repeated endpoint accumulates duplicates", func(t *testing.T) {
		adj := buildAdjacency([]Edge{{Source: "a", Target: "b"}, {Source: "c", Target: "b"}})
		if !reflect.DeepEqual(adj["b"], []string{"a", "c"}) {
			t.Errorf("adj[b] = %v, want [a c]", adj["b"])
		}
	})
}

func TestLabelPropagation(t *testing.T) {
	tests := []struct {
		name    string
		nodeIDs []string
		edges   []Edge
		want    map[string]string
	}{
		{
			name:    "single node keeps own label",
			nodeIDs: []string{"only"},
			want:    map[string]string{"only": "only"},
		},
		{
			name:    "path collapses to one community",
			nodeIDs: []string{"a", "b", "c"},
			edges:   []Edge{{Source: "a", Target: "b"}, {Source: "b", Target: "c"}},
			want:    map[string]string{"a": "b", "b": "b", "c": "b"},
		},
		{
			name:    "star converges on hub label",
			nodeIDs: []string{"a", "b", "c"},
			edges:   []Edge{{Source: "a", Target: "b"}, {Source: "a", Target: "c"}},
			want:    map[string]string{"a": "b", "b": "b", "c": "b"},
		},
		{
			name:    "disconnected components stay separate",
			nodeIDs: []string{"a", "b", "d", "e"},
			edges:   []Edge{{Source: "a", Target: "b"}, {Source: "d", Target: "e"}},
			want:    map[string]string{"a": "b", "b": "b", "d": "e", "e": "e"},
		},
		{
			name:    "isolated node keeps its own label",
			nodeIDs: []string{"a", "b", "z"},
			edges:   []Edge{{Source: "a", Target: "b"}},
			want:    map[string]string{"a": "b", "b": "b", "z": "z"},
		},
		{
			name:    "node absent from adjacency keeps own label",
			nodeIDs: []string{"z"},
			want:    map[string]string{"z": "z"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := labelPropagation(tc.nodeIDs, buildAdjacency(tc.edges))
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("labelPropagation() = %v, want %v", got, tc.want)
			}
		})
	}

	t.Run("deterministic tie-break picks lexicographically smallest label", func(t *testing.T) {
		got := labelPropagation([]string{"a", "b", "c"}, buildAdjacency([]Edge{
			{Source: "a", Target: "b"},
			{Source: "a", Target: "c"},
		}))
		if got["a"] != "b" {
			t.Fatalf("label[a] = %q, want %q (lex smallest of b/c)", got["a"], "b")
		}
	})

	t.Run("repeated runs are stable", func(t *testing.T) {
		ids := []string{"a", "b", "c", "d", "e", "f"}
		edges := []Edge{
			{Source: "a", Target: "b"}, {Source: "c", Target: "d"}, {Source: "b", Target: "c"},
		}
		first := labelPropagation(ids, buildAdjacency(edges))
		for i := 0; i < 5; i++ {
			if got := labelPropagation(ids, buildAdjacency(edges)); !reflect.DeepEqual(got, first) {
				t.Fatalf("run %d = %v, want %v", i, got, first)
			}
		}
	})
}

func TestAssignCommunityIDs(t *testing.T) {
	tests := []struct {
		name    string
		nodeIDs []string
		label   map[string]string
		want    map[string]int
	}{
		{
			name:    "bigger community gets lower id regardless of label order",
			nodeIDs: []string{"a", "b", "c", "d"},
			label:   map[string]string{"a": "La", "b": "La", "c": "La", "d": "Zb"},
			want:    map[string]int{"a": 0, "b": 0, "c": 0, "d": 1},
		},
		{
			name:    "equal sizes fall back to label ascending",
			nodeIDs: []string{"a", "b", "zzz"},
			label:   map[string]string{"a": "La", "b": "Lb", "zzz": "Lz"},
			want:    map[string]int{"a": 0, "b": 1, "zzz": 2},
		},
		{
			name:    "all-singleton labels ordered by label",
			nodeIDs: []string{"a", "b", "c", "d"},
			label:   map[string]string{"a": "Lb", "b": "La", "c": "Ld", "d": "Lc"},
			want:    map[string]int{"a": 1, "b": 0, "c": 3, "d": 2},
		},
		{
			name:    "node with no label maps to the empty-string community",
			nodeIDs: []string{"a", "b"},
			label:   map[string]string{"a": "x"},
			want:    map[string]int{"a": 1, "b": 0},
		},
		{
			name:    "no nodes yields empty map",
			nodeIDs: nil,
			label:   nil,
			want:    map[string]int{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := assignCommunityIDs(tc.nodeIDs, tc.label)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("assignCommunityIDs() = %v, want %v", got, tc.want)
			}
		})
	}

	t.Run("ids are contiguous from zero", func(t *testing.T) {
		ids := []string{"n1", "n2", "n3", "n4", "n5"}
		label := map[string]string{"n1": "x", "n2": "x", "n3": "y", "n4": "z", "n5": "z"}
		got := assignCommunityIDs(ids, label)
		seen := map[int]bool{}
		for _, id := range ids {
			seen[got[id]] = true
		}
		var keys []int
		for k := range seen {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		if !reflect.DeepEqual(keys, []int{0, 1, 2}) {
			t.Fatalf("community ids = %v, want [0 1 2]", keys)
		}
	})
}
