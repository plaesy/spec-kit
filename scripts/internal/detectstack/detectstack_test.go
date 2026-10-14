package detectstack

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// mappingFixture is a compact stand-in for instructions/mapping.json: an
// always_load list, keyword categories, one extension category and one
// filename category.
const mappingFixture = `{
  "mappings": {
    "always_load": ["plaesy.instructions.md", "context-engineering.instructions.md"],
    "languages": {
      "go": {
        "file": "go.instructions.md",
        "keywords": ["golang", "go lang", "go module"],
        "extensions": [".go"]
      },
      "java": {
        "file": "java.instructions.md",
        "keywords": ["java", "jvm", "maven", "gradle"]
      }
    },
    "frameworks": {
      "react": {
        "file": "reactjs.instructions.md",
        "keywords": ["react", "jsx", "tsx", "usestate"]
      },
      "nextjs": {
        "file": "nextjs.instructions.md",
        "keywords": ["next.js", "nextjs"],
        "filenames": ["next.config.js", "next.config.mjs"]
      }
    }
  }
}`

// newPlaesyRoot creates plaesyRoot/instructions/mapping.json from body.
func newPlaesyRoot(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "instructions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir instructions: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mapping.json"), []byte(body), 0o644); err != nil {
		t.Fatalf("write mapping.json: %v", err)
	}
	return root
}

// writeFile creates path (and its parents) with content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// has reports whether got contains want.
func has(got []string, want string) bool {
	for _, g := range got {
		if g == want {
			return true
		}
	}
	return false
}

func TestWordBoundaryRE(t *testing.T) {
	tests := []struct {
		name    string
		keyword string
		text    string
		want    bool
	}{
		{name: "exact word matches", keyword: "golang", text: "module needs golang", want: true},
		{name: "word boundary blocks a longer word", keyword: "java", text: "javascript only", want: false},
		{name: "shorter keyword still matches as its own word", keyword: "java", text: "java and scala", want: true},
		{name: "matching is case insensitive", keyword: "React", text: "uses REACT hooks", want: true},
		{name: "multi word keyword", keyword: "go lang", text: "written in go lang", want: true},
		{name: "multi word keyword needs all words adjacent", keyword: "go lang", text: "go and lang separated", want: false},
		{name: "substring inside punctuation still bounded", keyword: "sql", text: "a=b;sql=1", want: true},
		{name: "underscore is a word character so no match inside", keyword: "go", text: "cargo_tool", want: false},
		{name: "regex metacharacters are quoted, not interpreted", keyword: "c#", text: "written in csharp", want: false},
		// A trailing \b needs a word character right after a non-word one, and
		// "#" is a non-word character, so \bc#\b matched nothing anywhere: the
		// csharp entry's "c#" keyword in the real mapping.json could never fire,
		// and csharp detection leaned silently on its other keywords and its
		// extensions. A boundary belongs only where there is a word edge to
		// anchor to.
		{name: "a keyword ending in a non-word character matches on its own", keyword: "c#", text: "c# c#; written in c#", want: true},
		{name: "a keyword ending in a non-word character is still bounded on the left", keyword: "c#", text: "abc#", want: false},
		{name: "dot is literal", keyword: "next.js", text: "uses nextjs not next.js", want: true},
		{name: "dot keyword does not match any char", keyword: "next.js", text: "uses nextxjs", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := wordBoundaryRE(tt.keyword)
			got := re.MatchString(strings.ToLower(tt.text))
			if got != tt.want {
				t.Errorf("wordBoundaryRE(%q).MatchString(%q) = %v, want %v (pattern %s)",
					tt.keyword, tt.text, got, tt.want, re)
			}
		})
	}
}

func TestLoadMapping(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		missing    bool
		wantAlways []string
		wantCats   []string
		wantErr    bool
		check      func(t *testing.T, cats map[string]map[string]entry)
	}{
		{
			name: "always_load and categories", body: mappingFixture,
			wantAlways: []string{"plaesy.instructions.md", "context-engineering.instructions.md"},
			wantCats:   []string{"frameworks", "languages"},
			check: func(t *testing.T, cats map[string]map[string]entry) {
				go1, ok := cats["languages"]["go"]
				if !ok {
					t.Fatal("categories[\"languages\"][\"go\"] missing")
				}
				if go1.File != "go.instructions.md" {
					t.Errorf("file = %q", go1.File)
				}
				if !reflect.DeepEqual(go1.Extensions, []string{".go"}) {
					t.Errorf("extensions = %v", go1.Extensions)
				}
				if !reflect.DeepEqual(go1.Keywords, []string{"golang", "go lang", "go module"}) {
					t.Errorf("keywords = %v", go1.Keywords)
				}
				// A category member with no extensions/filenames leaves them nil.
				if cats["languages"]["java"].Extensions != nil {
					t.Error("extensions should be nil when absent")
				}
			},
		},
		{
			name: "no always_load key", body: `{"mappings": {"languages": {"go": {"file": "go.md"}}}}`,
			wantAlways: nil, wantCats: []string{"languages"},
		},
		{
			name: "empty always_load", body: `{"mappings": {"always_load": [], "x": {"a": {"file": "a.md"}}}}`,
			wantAlways: []string{}, wantCats: []string{"x"},
		},
		{
			name: "no mappings key at all", body: `{"description": "nothing here"}`,
			wantAlways: nil, wantCats: nil,
		},
		{
			name: "unreadable patterns and triggers are ignored", body: `{"mappings": {
			  "x": {"a": {"file": "a.md", "patterns": ["^never$"], "triggers": ["t"]}}
			}}`,
			wantAlways: nil, wantCats: []string{"x"},
		},
		{
			name:     "missing file",
			missing:  true,
			wantErr:  true,
			wantCats: nil,
		},
		{
			name:    "malformed json",
			body:    `{"mappings": {`,
			wantErr: true,
		},
		{
			name:    "always_load of the wrong type",
			body:    `{"mappings": {"always_load": {"a": "b"}}}`,
			wantErr: true,
		},
		{
			// A sibling that is not an object is metadata and is skipped, not
			// decoded as a category. The real mapping.json has "scope_load" in
			// exactly this shape, and decoding it as a category is what made
			// `plaesy detect-stack` fail on every project.
			name:     "sibling list is skipped, not decoded as a category",
			body:     `{"mappings": {"scope_load": ["a.instructions.md"], "frameworks": {"nextjs": {"file": "nextjs.instructions.md"}}}}`,
			wantCats: []string{"frameworks"},
		},
		{
			name:     "sibling string is skipped, not decoded as a category",
			body:     `{"mappings": {"note": "a.instructions.md"}}`,
			wantCats: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "mapping.json")
			if !tt.missing {
				writeFile(t, path, tt.body)
			}

			always, cats, err := loadMapping(path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got always=%v categories=%v", always, cats)
				}
				return
			}
			if err != nil {
				t.Fatalf("loadMapping: %v", err)
			}
			if !reflect.DeepEqual(always, tt.wantAlways) {
				t.Errorf("always_load = %#v, want %#v", always, tt.wantAlways)
			}
			var gotCats []string
			for k := range cats {
				gotCats = append(gotCats, k)
			}
			sort.Strings(gotCats)
			sort.Strings(tt.wantCats)
			if len(gotCats) != len(tt.wantCats) {
				t.Fatalf("categories = %v, want %v", gotCats, tt.wantCats)
			}
			for i := range gotCats {
				if gotCats[i] != tt.wantCats[i] {
					t.Errorf("categories = %v, want %v", gotCats, tt.wantCats)
					break
				}
			}
			if _, present := cats["always_load"]; present {
				t.Error("always_load must not appear as a category")
			}
			if tt.check != nil {
				tt.check(t, cats)
			}
		})
	}
}

// TestRealMappingJsonLoadsTheRepositoryMapping runs the parser against the
// repository's own instructions/mapping.json.
//
// "mappings" carries "scope_load" — the per-dimension assess instruction files a
// scope loads — next to the detection categories, as a plain string array.
// loadMapping used to decode every non-always_load member as a map of entries,
// so that array made the file fail to parse and `plaesy detect-stack` returned
// the error for every project, emitting neither always_load nor any category.
// A sibling that is not an object is metadata now, and skipped.
func TestRealMappingJsonLoadsTheRepositoryMapping(t *testing.T) {
	path := repoMappingPath(t)
	always, cats, err := loadMapping(path)
	if err != nil {
		t.Fatalf("the repository's own mapping.json must load: %v", err)
	}
	if len(always) == 0 {
		t.Error("always_load decoded empty")
	}
	if len(cats) == 0 {
		t.Error("mapping.json parsed but no categories were decoded")
	}
	if got := cats["languages"]["go"].File; got != "go.instructions.md" {
		t.Errorf("languages.go.file = %q, want go.instructions.md", got)
	}
	if _, isCategory := cats["scope_load"]; isCategory {
		t.Error("scope_load is a list of scoped instruction files, not a detection category")
	}
	if _, present := cats["always_load"]; present {
		t.Error("always_load must not appear as a category")
	}
}

// A non-object sibling of any shape is metadata, not a category: a mapping file
// may carry notes, counts or lists next to the categories.
func TestLoadMappingSkipsNonObjectSiblings(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "mapping.json")
	body := `{
	  "mappings": {
	    "always_load": ["core.instructions.md"],
	    "scope_load": ["assess-technical.instructions.md"],
	    "note": "a plain string sibling",
	    "count": 3,
	    "nested": {"null": null},
	    "frameworks": {"nextjs": {"file": "nextjs.instructions.md"}}
	  }
	}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write mapping: %v", err)
	}
	always, cats, err := loadMapping(path)
	if err != nil {
		t.Fatalf("loadMapping: %v", err)
	}
	if len(always) != 1 || always[0] != "core.instructions.md" {
		t.Errorf("always_load = %v", always)
	}
	if len(cats) != 2 {
		t.Errorf("categories = %v, want exactly frameworks and nested", keysOfCategories(cats))
	}
	if got := cats["frameworks"]["nextjs"].File; got != "nextjs.instructions.md" {
		t.Errorf("frameworks.nextjs.file = %q", got)
	}
}

func keysOfCategories(m map[string]map[string]entry) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func TestDepth(t *testing.T) {
	tests := []struct {
		name string
		root string
		dir  string
		want int
	}{
		{name: "root itself", root: "/a", dir: "/a", want: 0},
		{name: "one level", root: "/a", dir: "/a/b", want: 1},
		{name: "two levels", root: "/a", dir: "/a/b/c", want: 2},
		{name: "at the walk limit", root: "/a", dir: "/a/b/c/d", want: 3},
		{name: "beyond the walk limit", root: "/a", dir: "/a/b/c/d/e", want: 4},
		{name: "sibling", root: "/a", dir: "/a/b", want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := depth(tt.root, tt.dir); got != tt.want {
				t.Errorf("depth(%q, %q) = %d, want %d", tt.root, tt.dir, got, tt.want)
			}
		})
	}
}

func TestIsPruned(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{name: ".git is pruned", dir: ".git", want: true},
		{name: "node_modules is pruned", dir: "node_modules", want: true},
		{name: "anything else is walked", dir: "src", want: false},
		{name: "case matters: .GIT is not pruned", dir: ".GIT", want: false},
		{name: "nested node_modules name is pruned", dir: "node_modules", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPruned(tt.dir); got != tt.want {
				t.Errorf("isPruned(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

func TestDetectAlwaysLoadOnly(t *testing.T) {
	root := newPlaesyRoot(t, mappingFixture)
	target := t.TempDir()

	got, err := Detect(target, root)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	want := []string{"context-engineering.instructions.md", "plaesy.instructions.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Detect = %v, want %v (sorted always_load only)", got, want)
	}
}

func TestDetectMissingMappingIsNotAnError(t *testing.T) {
	tests := []struct {
		name string
		// seed replaces the mapping.json written by newPlaesyRoot.
		seed string
	}{
		{name: "no mapping.json at all"},
		{name: "mapping.json is malformed json", seed: `{"mappings": {`},
		{name: "mapping.json is a directory", seed: "dir"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "instructions")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			if tt.seed == "dir" {
				if err := os.MkdirAll(filepath.Join(dir, "mapping.json"), 0o755); err != nil {
					t.Fatalf("mkdir mapping.json: %v", err)
				}
			} else if tt.seed != "" {
				writeFile(t, filepath.Join(dir, "mapping.json"), tt.seed)
			}
			target := t.TempDir()

			got, err := Detect(target, root)
			if tt.name == "no mapping.json at all" {
				if err != nil {
					t.Fatalf("a missing mapping.json must not be an error, got %v", err)
				}
				if got != nil {
					t.Errorf("Detect = %v, want nil for a missing mapping.json", got)
				}
				return
			}
			// Malformed or unreadable mapping.json is reported, not swallowed.
			if err == nil {
				t.Fatalf("expected an error for %s, got %v", tt.name, got)
			}
		})
	}
}

func TestDetectKeywordSources(t *testing.T) {
	tests := []struct {
		name string
		// files are created relative to the target dir.
		files map[string]string
		want  []string
		notIn []string
	}{
		{
			name:  "root manifest content",
			files: map[string]string{"package.json": `{"dependencies": {"react": "^18"}}`},
			want:  []string{"reactjs.instructions.md"},
			notIn: []string{"go.instructions.md"},
		},
		{
			name:  "nested manifest content in a monorepo",
			files: map[string]string{"services/api/go.mod": "module x\n// golang\n"},
			want:  []string{"go.instructions.md"},
		},
		{
			name:  "nested csproj content without a fixed filename",
			files: map[string]string{"src/Api/Api.csproj": "<Project>jvm-ish toolchain</Project>"},
			want:  []string{"java.instructions.md"},
		},
		{
			name:  "spec context file feeds the keyword scan",
			files: map[string]string{".plaesy/specs/001-x/context.md": "We will use gradle for builds."},
			want:  []string{"java.instructions.md"},
		},
		{
			name:  "spec requirements file feeds the keyword scan",
			files: map[string]string{".plaesy/specs/002-y/requirements.md": "Must target the JVM."},
			want:  []string{"java.instructions.md"},
		},
		{
			name:  "docs/project.json feeds the keyword scan",
			files: map[string]string{"docs/project.json": `{"stack": "golang"}`},
			want:  []string{"go.instructions.md"},
		},
		{
			name:  "analysis project.json feeds the keyword scan",
			files: map[string]string{".plaesy/analysis/project.json": `{"lang": "maven project"}`},
			want:  []string{"java.instructions.md"},
		},
		{
			name:  "tsx file implies react keywords",
			files: map[string]string{"src/App.tsx": "export const App = () => null"},
			want:  []string{"reactjs.instructions.md"},
		},
		{
			name:  "jsx file implies react keywords",
			files: map[string]string{"src/App.jsx": "export const App = () => null"},
			want:  []string{"reactjs.instructions.md"},
		},
		{
			name:  "extension presence only, content irrelevant",
			files: map[string]string{"pkg/thing.go": "package pkg"},
			want:  []string{"go.instructions.md"},
		},
		{
			name:  "filename presence selects the nextjs instructions",
			files: map[string]string{"next.config.mjs": "// unrelated"},
			want:  []string{"nextjs.instructions.md"},
		},
		{
			name:  "filename matching is case insensitive",
			files: map[string]string{"NEXT.CONFIG.JS": "// unrelated"},
			want:  []string{"nextjs.instructions.md"},
		},
		{
			name: "word boundary: javascript does not match the java keyword",
			files: map[string]string{
				"package.json": `{"name": "x", "keywords": ["javascript"]}`,
			},
			notIn: []string{"java.instructions.md"},
		},
		{
			name:  "empty target matches nothing beyond always_load",
			files: map[string]string{},
			notIn: []string{"go.instructions.md", "java.instructions.md", "reactjs.instructions.md", "nextjs.instructions.md"},
		},
		{
			name:  "node_modules is pruned",
			files: map[string]string{"node_modules/react/package.json": `{"dependencies": {"react": "1"}}`},
			notIn: []string{"reactjs.instructions.md"},
		},
		{
			name:  "nested node_modules is pruned too",
			files: map[string]string{"app/node_modules/react/package.json": `{"dependencies": {"react": "1"}}`},
			notIn: []string{"reactjs.instructions.md"},
		},
		{
			name:  ".git is pruned",
			files: map[string]string{".git/hooks/package.json": `{"dependencies": {"react": "1"}}`},
			notIn: []string{"reactjs.instructions.md"},
		},
		{
			name:  "manifests below the walk limit are ignored",
			files: map[string]string{"a/b/c/d/go.mod": "// golang"},
			notIn: []string{"go.instructions.md"},
		},
		{
			name:  "files below the walk limit do not match extensions",
			files: map[string]string{"a/b/c/d/thing.go": "package pkg"},
			notIn: []string{"go.instructions.md"},
		},
		{
			name:  "csproj below the walk limit is ignored",
			files: map[string]string{"a/b/c/d/Api.csproj": "jvm"},
			notIn: []string{"java.instructions.md"},
		},
		{
			// depth() counts every path segment below the target, the file
			// name included, so a/b/go.mod sits at the limit (3) and
			// a/b/c/go.mod is already past it.
			name:  "manifest at exactly the walk limit is read",
			files: map[string]string{"a/b/go.mod": "// golang"},
			want:  []string{"go.instructions.md"},
		},
		{
			name:  "manifest one segment past the walk limit is ignored",
			files: map[string]string{"a/b/c/go.mod": "// golang"},
			notIn: []string{"go.instructions.md"},
		},
		{
			name:  "extension match at exactly the walk limit",
			files: map[string]string{"a/b/thing.go": "package pkg"},
			want:  []string{"go.instructions.md"},
		},
		{
			name:  "csproj at exactly the walk limit is read",
			files: map[string]string{"a/b/Api.csproj": "jvm"},
			want:  []string{"java.instructions.md"},
		},
		{
			name:  "csproj one segment past the walk limit is ignored",
			files: map[string]string{"a/b/c/Api.csproj": "jvm"},
			notIn: []string{"java.instructions.md"},
		},
		{
			name: "content is joined with a space so words do not merge across files",
			files: map[string]string{
				"a.txt": "go",
				"b.txt": "mod",
			},
			notIn: []string{"go.instructions.md"},
		},
		{
			name:  "case insensitive keyword match",
			files: map[string]string{"package.json": `{"note": "We use GoLang here"}`},
			want:  []string{"go.instructions.md"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newPlaesyRoot(t, mappingFixture)
			target := t.TempDir()
			for rel, content := range tt.files {
				writeFile(t, filepath.Join(target, filepath.FromSlash(rel)), content)
			}

			got, err := Detect(target, root)
			if err != nil {
				t.Fatalf("Detect: %v", err)
			}
			for _, w := range tt.want {
				if !has(got, w) {
					t.Errorf("Detect = %v, missing %q", got, w)
				}
			}
			for _, w := range tt.notIn {
				if has(got, w) {
					t.Errorf("Detect = %v, must not contain %q", got, w)
				}
			}
			// always_load is always present.
			for _, w := range []string{"plaesy.instructions.md", "context-engineering.instructions.md"} {
				if !has(got, w) {
					t.Errorf("Detect = %v, missing always_load entry %q", got, w)
				}
			}
		})
	}
}

// TestDetectOutputIsSortedAndDeduplicated covers the contract the callers
// parse: one sorted list, each file once, however many sources matched it.
func TestDetectOutputIsSortedAndDeduplicated(t *testing.T) {
	root := newPlaesyRoot(t, mappingFixture)
	target := t.TempDir()
	// react is matched three ways at once: a manifest keyword, a nested
	// manifest keyword and a spec file.
	writeFile(t, filepath.Join(target, "package.json"), `{"dependencies": {"react": "18"}}`)
	writeFile(t, filepath.Join(target, "app", "package.json"), `{"name": "useEffect demo"}`)
	writeFile(t, filepath.Join(target, ".plaesy", "specs", "001-x", "context.md"), "react and useState")

	got, err := Detect(target, root)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	want := []string{
		"context-engineering.instructions.md",
		"plaesy.instructions.md",
		"reactjs.instructions.md",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Detect = %v, want %v", got, want)
	}
}

// TestDetectMissingTargetDir covers a target that does not exist: the manifest
// reads and the directory walk are both best effort, so detection degrades to
// always_load instead of failing.
func TestDetectMissingTargetDir(t *testing.T) {
	root := newPlaesyRoot(t, mappingFixture)
	target := filepath.Join(t.TempDir(), "does-not-exist")

	got, err := Detect(target, root)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	want := []string{"context-engineering.instructions.md", "plaesy.instructions.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Detect = %v, want %v", got, want)
	}
}

// TestDetectIgnoresEntriesWithoutAFile covers category members with no "file"
// key: they contribute keywords but nothing to the output.
func TestDetectIgnoresEntriesWithoutAFile(t *testing.T) {
	body := `{"mappings": {
	  "always_load": ["base.instructions.md"],
	  "x": {"entry": {"keywords": ["golang"]}}
	}}`
	root := newPlaesyRoot(t, body)
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "go.mod"), "// golang")

	got, err := Detect(target, root)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"base.instructions.md"}) {
		t.Errorf("Detect = %v, want only the always_load entry", got)
	}
}

// TestDetectIgnoresBlankKeywordsAndFiles documents the two guards in the
// keyword loop: a blank keyword is skipped and an entry with no file never
// reaches the result set.
func TestDetectIgnoresBlankKeywords(t *testing.T) {
	body := `{"mappings": {
	  "always_load": ["base.instructions.md"],
	  "x": {"entry": {"file": "x.instructions.md", "keywords": ["   ", "", "  "]}}
	}}`
	root := newPlaesyRoot(t, body)
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "go.mod"), "golang blank keywords")

	got, err := Detect(target, root)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"base.instructions.md"}) {
		t.Errorf("Detect = %v, want blank keywords to be ignored", got)
	}
}

// repoMappingPath locates the repository's own instructions/mapping.json by
// walking up from the test's working directory, skipping when the test runs
// outside the repository.
func repoMappingPath(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Skipf("cannot determine the working directory: %v", err)
	}
	for {
		candidate := filepath.Join(dir, "instructions", "mapping.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skip("not running inside the spec-kit repository (no instructions/mapping.json above the test)")
		}
		dir = parent
	}
}
