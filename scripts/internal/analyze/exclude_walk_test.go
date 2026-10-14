package analyze

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// walkProject and the fingerprint walk must agree on what is excluded, so the
// fingerprint never tracks a folder the generated files do not describe.
func TestWalkProjectHonoursGitignoreAndSharedExclusions(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		".gitignore":                    "*.g.dart\n",
		"lib/main.dart":                 "void main() {}\n",
		"lib/model.g.dart":              "class M {}\n",
		"target/out.dart":               "class T {}\n",
		"ios/Pods/Pod.swift":            "class P {}\n",
		"ios/Flutter/ephemeral/h.swift": "class H {}\n",
	}
	for rel, body := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	res, err := walkProject(root)
	if err != nil {
		t.Fatalf("walkProject: %v", err)
	}
	var got []string
	for _, f := range res.files {
		got = append(got, f.relPath)
	}
	sort.Strings(got)
	want := []string{".gitignore", "lib/main.dart"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("walkProject files = %v\nwant %v", got, want)
	}
}
