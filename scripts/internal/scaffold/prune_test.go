package scaffold

import (
	"path/filepath"
	"strings"
	"testing"
)

// pruneRoots decides which trees may be deleted from, so its negative branches
// matter as much as its positive one. Each case below is a way a platform can
// fail to be prunable, and every one of them has to end with "return the roles
// root and nothing else".

func pruneConfig(platforms map[string]map[string]string, promptSource string) *PlatformConfig {
	cfg := &PlatformConfig{
		Plaesy: PlaesySection{
			BaseDirectory: ".plaesy",
			Mapping: map[string]MappingEntry{
				"prompts": {Value: promptSource},
			},
		},
		Platforms: make(map[string]PlatformEntry, len(platforms)),
	}
	for id, mapping := range platforms {
		cfg.Platforms[id] = PlatformEntry{Name: id, Mapping: mapping}
	}
	return cfg
}

func promptRootLabels(t *testing.T, roots []pruneRoot) string {
	t.Helper()
	var labels []string
	for _, r := range roots {
		labels = append(labels, r.label)
	}
	return strings.Join(labels, ",")
}

func TestPruneRootsOnlyOffersTheMirrorWhenThePlatformProvesItOwnsTheDirectory(t *testing.T) {
	home := "/home/repo"
	target := "/proj"
	base := "/proj/.plaesy"

	exclusive := map[string]string{"prompts": ".kilo/commands", "prune_prompts": "true"}
	shared := map[string]string{"prompts": ".cursor/rules", "prune_prompts": "true"}

	cases := []struct {
		name     string
		platform string
		cfg      *PlatformConfig
		want     string
	}{
		{
			name:     "no platform named",
			platform: "",
			cfg:      pruneConfig(map[string]map[string]string{"kilo_code": exclusive}, "prompts/*"),
			want:     "agent roles",
		},
		{
			name:     "platform is not in the config",
			platform: "ghost",
			cfg:      pruneConfig(map[string]map[string]string{"kilo_code": exclusive}, "prompts/*"),
			want:     "agent roles",
		},
		{
			name:     "platform maps no prompts directory",
			platform: "kilo_code",
			cfg:      pruneConfig(map[string]map[string]string{"kilo_code": {"core": "AGENTS.md"}}, "prompts/*"),
			want:     "agent roles",
		},
		{
			name:     "destination is shared with hand-written files",
			platform: "kilo_code",
			cfg:      pruneConfig(map[string]map[string]string{"kilo_code": shared}, "prompts/*"),
			want:     "agent roles",
		},
		{
			name:     "destination is exclusive but nobody claimed it",
			platform: "kilo_code",
			cfg: pruneConfig(map[string]map[string]string{
				"kilo_code": {"prompts": ".kilo/commands"},
			}, "prompts/*"),
			want: "agent roles",
		},
		{
			name:     "the claim is not a boolean",
			platform: "kilo_code",
			cfg: pruneConfig(map[string]map[string]string{
				"kilo_code": {"prompts": ".kilo/commands", "prune_prompts": "maybe"},
			}, "prompts/*"),
			want: "agent roles",
		},
		{
			name:     "the claim is explicitly false",
			platform: "kilo_code",
			cfg: pruneConfig(map[string]map[string]string{
				"kilo_code": {"prompts": ".kilo/commands", "prune_prompts": "false"},
			}, "prompts/*"),
			want: "agent roles",
		},
		{
			name:     "there is no prompt source to compare against",
			platform: "kilo_code",
			cfg:      pruneConfig(map[string]map[string]string{"kilo_code": exclusive}, ""),
			want:     "agent roles",
		},
		{
			name:     "an exclusive destination is offered",
			platform: "kilo_code",
			cfg:      pruneConfig(map[string]map[string]string{"kilo_code": exclusive}, "prompts/*"),
			want:     "agent roles,prompt mirror",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			roots := pruneRoots(home, target, base, tc.platform, tc.cfg)
			if got := promptRootLabels(t, roots); got != tc.want {
				t.Errorf("pruneRoots labels = %q, want %q", got, tc.want)
			}
		})
	}
}

// The mirror the root describes has to point at the real source and the real
// destination, or the whole comparison is against the wrong trees.
func TestPruneRootPointsAtTheSourceAndDestination(t *testing.T) {
	cfg := pruneConfig(map[string]map[string]string{
		"kilo_code": {"prompts": ".kilo/commands", "prune_prompts": "true"},
	}, "prompts/*")

	roots := pruneRoots("/home/repo", "/proj", "/proj/.plaesy", "kilo_code", cfg)
	if len(roots) != 2 {
		t.Fatalf("pruneRoots returned %d roots, want 2", len(roots))
	}

	mirror := roots[1]
	if want := filepath.Join("/home/repo", "prompts"); mirror.srcDir != want {
		t.Errorf("mirror srcDir = %q, want %q (the trailing /* is not a directory)", mirror.srcDir, want)
	}
	if want := filepath.Join("/proj", ".kilo", "commands"); mirror.destDir != want {
		t.Errorf("mirror destDir = %q, want %q", mirror.destDir, want)
	}
	if mirror.srcStrip != ".md" || mirror.dstExt != ".md" {
		t.Errorf("mirror name rule = %q -> %q, want .md -> .md", mirror.srcStrip, mirror.dstExt)
	}

	roles := roots[0]
	if want := filepath.Join("/home/repo", "agents"); roles.srcDir != want {
		t.Errorf("roles srcDir = %q, want %q", roles.srcDir, want)
	}
	if want := filepath.Join("/proj", ".plaesy", "roles"); roles.destDir != want {
		t.Errorf("roles destDir = %q, want %q", roles.destDir, want)
	}
	// A source agent file is architect.agents.md and the copy is
	// architect.md, so the rule that connects them is not the identity one the
	// prompt mirror uses.
	if roles.srcStrip != ".agents.md" || roles.dstExt != ".md" {
		t.Errorf("roles name rule = %q -> %q, want .agents.md -> .md", roles.srcStrip, roles.dstExt)
	}
}

// withinRoot is the last thing between a bug in the candidate list and a
// deletion outside the owned tree, so it is tested on the paths that escape.
func TestWithinRootRejectsPathsOutsideTheRoot(t *testing.T) {
	root := filepath.Join("proj", ".kilo", "commands")

	for _, outside := range []string{
		filepath.Join("proj", ".kilo", "commands-backup", "x.md"),
		filepath.Join("proj", ".kilo", "other", "x.md"),
		filepath.Join("proj", ".kilo", "commands", "..", "..", "memory.md"),
		filepath.Join("elsewhere", "x.md"),
	} {
		if withinRoot(root, outside) {
			t.Errorf("withinRoot(%q, %q) = true, want false", root, outside)
		}
	}
	for _, inside := range []string{
		filepath.Join("proj", ".kilo", "commands", "x.md"),
		filepath.Join("proj", ".kilo", "commands", "create", "doc", "design.md"),
		root,
	} {
		if !withinRoot(root, inside) {
			t.Errorf("withinRoot(%q, %q) = false, want true", root, inside)
		}
	}
}

// A candidate that is not inside an owned tree is refused rather than removed,
// even when it reaches applyPrune. This is the guard that keeps a bug in the
// scan from becoming a deletion somewhere else in the project.
func TestApplyPruneRefusesCandidatesOutsideEveryOwnedTree(t *testing.T) {
	target := t.TempDir()

	outside := filepath.Join(target, "specs", "001-thing", "spec.md")
	writeFile(t, outside, "MY SPEC\n")
	ownedRoot := filepath.Join(target, ".plaesy", "roles")
	writeFile(t, filepath.Join(ownedRoot, "ghost.md"), "retired\n")

	roots := []pruneRoot{{label: "agent roles", destDir: ownedRoot, srcDir: "unused", srcStrip: ".md", dstExt: ".md"}}
	removed := applyPrune(roots, target, []string{
		"specs/001-thing/spec.md",
		filepath.Join(".plaesy", "roles", "ghost.md"),
	})

	if len(removed) != 1 {
		t.Fatalf("applyPrune removed %v, want only the file inside the owned tree", removed)
	}
	if !strings.HasSuffix(removed[0], "ghost.md") {
		t.Errorf("applyPrune removed %q, want the owned-tree file", removed[0])
	}
	if !exists(t, outside) {
		t.Error("a spec outside every owned tree was deleted")
	}
}
