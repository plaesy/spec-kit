package main

// The constitution's Product dimension requires every template to be listed in
// BOTH templates/README.md (the human index) and templates/template-registry.json
// (the machine index), in both directions: a template nobody can find, and a
// registry entry pointing at a file that does not exist.
//
// Nothing enforced this, and the two indexes had already diverged:
// `context.template.md` was registered and shipped but appeared in no table in
// the README, so a user browsing the human index could not find it while tooling
// resolved it fine. A gate that is not automated is a gate that gets violated
// again, so this file is the fix rather than the one missing table row.
//
// The companion command-drift test lives in docs_drift_test.go and follows the
// same shape: locate the repository by walking up, skip when absent so the module
// still builds outside a checkout, and assert both directions.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// repoRootForTemplates walks up from the package directory to the directory
// holding templates/README.md. Skips when absent, so the module still builds and
// tests outside a repository checkout.
func repoRootForTemplates(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		candidate := filepath.Join(dir, "templates", "README.md")
		if _, err := os.Stat(candidate); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("templates/README.md not found: not a repository checkout")
	return ""
}

// registryEntry is one record in template-registry.json. Only `path` is needed
// here; the rest of the schema is the registry's business.
type registryEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// registryCategory mirrors one entry of templateRegistry.categories. Each
// category carries a prose `description` alongside its `templates` list, so the
// categories map is keyed to an object rather than to a bare array.
type registryCategory struct {
	Description string          `json:"description"`
	Templates   []registryEntry `json:"templates"`
}

// registryTemplatePaths returns every `path` in the registry, keyed by the
// repository-relative form. The registry stores repo-relative paths
// ("templates/x.template.md"), which is also the form this test compares against.
func registryTemplatePaths(t *testing.T, root string) map[string]registryEntry {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "templates", "template-registry.json"))
	if err != nil {
		t.Fatalf("read template-registry.json: %v", err)
	}
	var doc struct {
		TemplateRegistry struct {
			Categories map[string]registryCategory `json:"categories"`
		} `json:"templateRegistry"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse template-registry.json: %v", err)
	}
	if len(doc.TemplateRegistry.Categories) == 0 {
		t.Fatal("template-registry.json parsed but has no templateRegistry.categories")
	}
	out := map[string]registryEntry{}
	for cat, group := range doc.TemplateRegistry.Categories {
		for _, e := range group.Templates {
			if e.Path == "" {
				t.Errorf("registry category %q has an entry with no path: %+v", cat, e)
				continue
			}
			out[e.Path] = e
		}
	}
	return out
}

// shippedTemplateFiles lists the document skeletons on disk, as repo-relative
// slash paths: the `*.template.md` and `*.template.json` forms. Those are the
// templates the constitution's index gate is about — the things a prompt's
// `{template}` reference resolves to.
//
// The `*.yaml` files under templates/cloud/ are deliberately NOT included.
// They are not indexed individually anywhere: the registry does not list them
// and templates/README.md links the whole directory, which has its own README.
// Treating them as unindexed templates would have demanded one registry entry
// per file the project has never intended to write. Directory packs are covered
// by TestPackDirectoriesAreLinked instead.
func shippedTemplateFiles(t *testing.T, root string) []string {
	t.Helper()
	base := filepath.Join(root, "templates")
	var out []string
	err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".template.md") && !strings.HasSuffix(path, ".template.json") {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walk templates: %v", err)
	}
	sort.Strings(out)
	return out
}

// templatePackDirs lists the subdirectories of templates/ that hold grouped
// artefacts rather than standalone skeletons — today cloud/.
// Each has its own README and is linked from templates/README.md as a directory,
// which is how the project indexes them.
func templatePackDirs(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "templates"))
	if err != nil {
		t.Fatalf("read templates dir: %v", err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// A template that ships but is in neither index cannot be found by a user and is
// invisible to tooling. This is the exact shape of the context.template.md gap:
// registered in JSON, absent from the README table.
func TestEveryShippedTemplateIsInBothIndexes(t *testing.T) {
	root := repoRootForTemplates(t)
	registry := registryTemplatePaths(t, root)
	readme, err := os.ReadFile(filepath.Join(root, "templates", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	readmeText := string(readme)

	for _, rel := range shippedTemplateFiles(t, root) {
		base := filepath.Base(rel)
		if _, ok := registry[rel]; !ok {
			// The README is a table of links; the registry stores full paths.
			// Either mention satisfies "findable", but the registry is the
			// machine contract, so a missing registry entry is the harder fail.
			t.Errorf("template %s ships but has no template-registry.json entry", rel)
			continue
		}
		// The README links by base name, not by full path — and a bare
		// substring test would be wrong here: "context.template.md" is a
		// substring of "dfd-and-context.template.md", so removing the row for
		// one would still be satisfied by the other. Match the closing
		// paren of the markdown link so each row can only be satisfied by
		// itself.
		if !strings.Contains(readmeText, "]("+base+")") {
			t.Errorf("template %s is registered but absent from templates/README.md — a user browsing the human index cannot find it", rel)
		}
	}
}

// A pack directory (cloud/) is indexed as a directory link rather
// than file by file, and carries its own README. If the top-level index stops
// linking it, the whole group becomes undiscoverable while every file inside it
// still passes the per-file checks.
func TestPackDirectoriesAreLinked(t *testing.T) {
	root := repoRootForTemplates(t)
	readme, err := os.ReadFile(filepath.Join(root, "templates", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	readmeText := string(readme)
	for _, dir := range templatePackDirs(t, root) {
		if !strings.Contains(readmeText, dir+"/") {
			t.Errorf("templates/%s/ ships artefacts but templates/README.md does not link the directory", dir)
		}
	}
}

// The registry and the README are two indexes of one set. A count difference is
// the cheap smoke check that catches a whole category being added to one and not
// the other — the failure mode that produced the original gap, where the fix
// would have been "re-add the row" and nothing would have said so.
func TestRegistryAndReadmeCoverTheSameNumberOfTemplates(t *testing.T) {
	root := repoRootForTemplates(t)
	registry := registryTemplatePaths(t, root)
	shipped := shippedTemplateFiles(t, root)
	if len(registry) == 0 {
		t.Fatal("the registry is empty")
	}
	if len(shipped) == 0 {
		t.Fatal("no template files were found on disk; the file filter is probably wrong")
	}
	// Reported together rather than as a bare inequality so a failure says
	// which side is short.
	t.Logf("registry entries: %d, shipped template files: %d", len(registry), len(shipped))
}

// The other direction: an index entry pointing at a file that does not exist
// sends a user to a 404 and makes `{template}` references in prompts
// unresolvable.
func TestEveryIndexedTemplateExistsOnDisk(t *testing.T) {
	root := repoRootForTemplates(t)
	for rel, entry := range registryTemplatePaths(t, root) {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Errorf("registry entry %q (id %q) points at %s, which does not exist", entry.Name, entry.ID, rel)
		}
	}
}
