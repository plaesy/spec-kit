package analyze

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// files builds a []fileInfo from slash-separated relative paths.
func files(paths ...string) []fileInfo {
	out := make([]fileInfo, 0, len(paths))
	for _, p := range paths {
		out = append(out, fileInfo{relPath: p, absPath: filepath.FromSlash(p), size: int64(len(p))})
	}
	return out
}

// dirsOf builds a []dirInfo from slash-separated relative paths.
func dirsOf(paths ...string) []dirInfo {
	out := make([]dirInfo, 0, len(paths))
	for _, p := range paths {
		out = append(out, dirInfo{relPath: p})
	}
	return out
}

func TestCountFileTypes(t *testing.T) {
	cases := []struct {
		name  string
		files []fileInfo
		want  fileCounts
	}{
		{
			name:  "empty input",
			files: nil,
			want:  fileCounts{},
		},
		{
			name: "source, docs and config are classified independently",
			files: files(
				"main.go", "lib/a.py", "app.tsx", "run.sh", "Model.java", "lib.rs",
				"README.md", "notes.txt", "guide.rst", "CHANGELOG.adoc",
				"package.json", "config.yaml", "pyproject.toml", "setup.cfg", "pom.xml", "app.ini",
			),
			want: fileCounts{SourceCode: 6, Documentation: 4, Configuration: 6},
		},
		{
			name:  "extension matching is case insensitive",
			files: files("README.MD", "Main.GO", "CONFIG.YAML"),
			want:  fileCounts{SourceCode: 1, Documentation: 1, Configuration: 1},
		},
		{
			name:  "unknown extensions are ignored entirely",
			files: files("logo.png", "Makefile", "Dockerfile", "LICENSE", "data.csv"),
			want:  fileCounts{},
		},
		{
			name:  "extensionless files are ignored",
			files: files("Makefile", "Rakefile", "Jenkinsfile"),
			want:  fileCounts{},
		},
		{
			name:  "only the base name is considered, not the directory",
			files: files("docs/json/main.py"),
			want:  fileCounts{SourceCode: 1},
		},
		{
			name:  "a directory-like path ending in a tracked extension counts as that type",
			files: files("vendor/lib.go"),
			want:  fileCounts{SourceCode: 1},
		},
		{
			name:  "source wins when an extension could match several buckets",
			files: files("a.json.js"),
			want:  fileCounts{SourceCode: 1},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := countFileTypes(tc.files); got != tc.want {
				t.Errorf("countFileTypes() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestDirectoryDescription(t *testing.T) {
	cases := []struct {
		relPath string
		want    string
	}{
		{"lib", "Source code library directory"},
		{"lib/screens", "UI screen components"},
		{"lib/screens/home", "UI screen components"},
		{"lib/models", "Data models and entities"},
		{"lib/services", "Business logic and API services"},
		{"lib/utils", "Utility functions and helpers"},
		{"test", "Test files and unit tests"},
		{"android", "Android platform specific code"},
		{"ios", "iOS platform specific code"},
		{"assets", "Static assets (images, fonts, etc.)"},
		{"assets/images", "Image assets"},
		{"assets/data", "Data files"},
		{"docs", "Documentation files"},
		{"scripts", "Build and utility scripts"},
		{"src", "Project directory"},
		// The lib/screens rule is a raw prefix match, so a sibling directory
		// with a longer name is misclassified too.
		{"lib/screens_extra", "UI screen components"},
		{"", "Project directory"},
	}
	for _, tc := range cases {
		t.Run(tc.relPath, func(t *testing.T) {
			if got := directoryDescription(tc.relPath); got != tc.want {
				t.Errorf("directoryDescription(%q) = %q, want %q", tc.relPath, got, tc.want)
			}
		})
	}
}

func TestFileType(t *testing.T) {
	cases := []struct {
		relPath string
		want    string
	}{
		{"lib/main.dart", "dart"},
		{"src/app.js", "javascript"},
		{"src/app.jsx", "javascript"},
		{"src/app.ts", "typescript"},
		{"src/app.tsx", "typescript"},
		{"scripts/tool.py", "python"},
		{"cmd/main.go", "go"},
		{"src/Main.java", "java"},
		{"data/config.json", "json"},
		{"k8s/deploy.yaml", "yaml"},
		{"k8s/deploy.yml", "yaml"},
		{"docs/README.md", "markdown"},
		{"src/main.rs", "text"},
		{"src/logo.png", "text"},
		{"Makefile", "text"},
		{"", "text"},
	}
	for _, tc := range cases {
		t.Run(tc.relPath, func(t *testing.T) {
			if got := fileType(tc.relPath); got != tc.want {
				t.Errorf("fileType(%q) = %q, want %q", tc.relPath, got, tc.want)
			}
		})
	}
}

func TestFilePurpose(t *testing.T) {
	cases := []struct {
		name    string
		relPath string
		want    string
	}{
		{"pubspec.yaml", "pubspec.yaml", "Flutter/Dart project configuration"},
		{"nested package.json", "web/package.json", "Node.js project configuration"},
		{"requirements.txt", "requirements.txt", "Python dependencies"},
		{"go.mod", "go.mod", "Go module configuration"},
		{"Cargo.toml", "crates/app/Cargo.toml", "Rust project configuration"},
		{"README.md", "README.md", "Project documentation"},
		{"README without extension", "README", "Project documentation"},
		{"nested README", "docs/README.rst", "Project documentation"},
		{"LICENSE", "LICENSE", "Project license"},
		{"main.py", "main.py", "Application entry point"},
		// The entry-point rule needs a dot after the name, so an extensionless
		// "main" is not recognised.
		{"extensionless main", "cmd/main", "Project file"},
		{"index.js", "src/index.js", "Application entry point"},
		{"config file middle dot", "vite.config.js", "Configuration file"},
		{"config file suffix", "tsconfig.config", "Configuration file"},
		{"gitignore", ".gitignore", "Git ignore rules"},
		{"env file", ".env", "Environment variables"},
		{"env example", ".env.example", "Environment variables"},
		{"unclassified", "src/widget.tsx", "Project file"},
		// Ordering matters: the entry-point rules are evaluated before the
		// generic ".config." rule, so main.config.js is an entry point.
		{"entry point beats configuration file", "main.config.js", "Application entry point"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := filePurpose(tc.relPath); got != tc.want {
				t.Errorf("filePurpose(%q) = %q, want %q", tc.relPath, got, tc.want)
			}
		})
	}
}

func TestIsKeyFile(t *testing.T) {
	cases := []struct {
		name    string
		relPath string
		want    bool
	}{
		{"package.json", "package.json", true},
		{"pubspec.yaml", "pubspec.yaml", true},
		{"requirements.txt", "requirements.txt", true},
		{"go.mod", "go.mod", true},
		{"Cargo.toml", "Cargo.toml", true},
		{"main.dart", "lib/main.dart", true},
		{"main.js", "main.js", true},
		{"main.py", "app/main.py", true},
		{"main.go", "cmd/main.go", true},
		{"index.js", "web/index.js", true},
		{"any markdown", "docs/guide.md", true},
		{"README prefix", "sub/README", true},
		{"LICENSE", "LICENSE", true},
		{"any json", "data/records.json", true},
		{"any yaml", "ci/pipeline.yaml", true},
		{"any yml", "ci/pipeline.yml", true},
		{"plain source", "src/widget.ts", false},
		{"image", "assets/logo.png", false},
		{"lockfile", "package-lock.json", true},
		{"text file", "notes.txt", false},
		{"empty path", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isKeyFile(tc.relPath); got != tc.want {
				t.Errorf("isKeyFile(%q) = %v, want %v", tc.relPath, got, tc.want)
			}
		})
	}
}

func TestProjectClassification(t *testing.T) {
	hits := func(names ...string) []FrameworkHit {
		out := make([]FrameworkHit, 0, len(names))
		for _, n := range names {
			out = append(out, FrameworkHit{Framework: n, Confidence: "high"})
		}
		return out
	}
	cases := []struct {
		name       string
		frameworks []FrameworkHit
		want       string
	}{
		{"no frameworks", nil, "single"},
		{"one framework", hits("go"), "single"},
		{"frontend and backend", hits("nextjs", "django"), "full-stack"},
		// Two backends without a frontend are not "full-stack"; only the
		// frontend+backend pairing is.
		{"two backends only", hits("gin", "echo"), "multi-framework"},
		{"frontend plus devops", hits("react", "docker"), "full-stack-with-infrastructure"},
		{"devops only", hits("docker", "terraform"), "infrastructure-focused"},
		{"mobile plus backend", hits("flutter", "gin"), "mobile-backend"},
		{"no recognised roles", hits("nodejs", "python"), "multi-framework"},
		{"kustomize alone counts as devops", hits("spec-kit", "kustomize"), "infrastructure-focused"},
		// A single mobile hit short-circuits before the mobile rules run.
		{"single mobile framework", hits("flutter"), "single"},
		// hasDevops is evaluated before hasMobile, so a mobile + devops project
		// is reported as infrastructure-focused rather than mobile-backend.
		{"mobile plus devops prefers infrastructure", hits("flutter", "docker"), "infrastructure-focused"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := projectClassification(tc.frameworks); got != tc.want {
				t.Errorf("projectClassification(%v) = %q, want %q", tc.frameworks, got, tc.want)
			}
		})
	}
}

func TestComplexityFor(t *testing.T) {
	cases := []struct {
		totalFiles int
		want       string
	}{
		{0, "Small"},
		{1, "Small"},
		{20, "Small"},
		{21, "Medium"},
		{50, "Medium"},
		{51, "Large"},
		{10000, "Large"},
	}
	for _, tc := range cases {
		t.Run(strconv.Itoa(tc.totalFiles), func(t *testing.T) {
			if got := complexityFor(tc.totalFiles); got != tc.want {
				t.Errorf("complexityFor(%d) = %q, want %q", tc.totalFiles, got, tc.want)
			}
		})
	}
}

func TestFrameworkNames(t *testing.T) {
	cases := []struct {
		name string
		hits []FrameworkHit
		want []string
	}{
		{"empty", nil, []string{}},
		{"single", []FrameworkHit{{"go", "medium"}}, []string{"go"}},
		{"many", []FrameworkHit{{"nextjs", "high"}, {"react", "high"}, {"docker", "medium"}},
			[]string{"nextjs", "react", "docker"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := frameworkNames(tc.hits)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("frameworkNames(%v) = %v, want %v", tc.hits, got, tc.want)
			}
		})
	}
}

func TestTimestampFormat(t *testing.T) {
	re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(Z|[+-]\d{2}:\d{2})$`)
	got := timestamp()
	if !re.MatchString(got) {
		t.Errorf("timestamp() = %q, does not match RFC3339-ish layout", got)
	}
}

func TestWriteJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	payload := map[string]any{"b": 2, "a": []int{1, 2}}
	if err := writeJSON(path, payload); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}
	got := readFileString(t, path)
	mustContain(t, got, "\n    \"a\": [")
	mustContain(t, got, "\n    \"b\": 2")
	// Keys are emitted in sorted order, so the file is deterministic.
	if strings.Index(got, `"a"`) > strings.Index(got, `"b"`) {
		t.Errorf("expected deterministic (sorted) key order in:\n%s", got)
	}
}

func TestWriteJSONMarshalError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.json")
	// Channels cannot be marshalled, which exercises the error branch.
	if err := writeJSON(path, make(chan int)); err == nil {
		t.Fatal("expected a marshal error for an unencodable value")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("no file should be written when marshalling fails")
	}
}

// testRunContext builds a fully populated runContext for the generators: a
// three-file Go project that also looks containerized.
func testRunContext() *runContext {
	fl := files("main.go", "README.md", "package.json")
	fc := countFileTypes(fl)
	return &runContext{
		walk: walkResult{
			files:      fl,
			dirs:       dirsOf("lib", "tests"),
			newestUnix: 1,
		},
		frameworks:       []FrameworkHit{{"go", "medium"}, {"docker", "medium"}},
		projectType:      FrameworkHit{"go", "medium"},
		languages:        []string{"Go"},
		devTools:         []string{"Git", "Go Modules"},
		buildSystems:     []string{"Go Modules"},
		fileCounts:       fc,
		frameworkVersion: "1.2.3",
	}
}

func TestGenerateProjectJSON(t *testing.T) {
	project := newFixture(t, map[string]string{"main.go": "package main"})
	analysisDir := filepath.Join(project, ".plaesy", "analysis")
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		t.Fatalf("mkdir analysis dir: %v", err)
	}
	ctx := testRunContext()

	if err := generateProjectJSON(analysisDir, project, ctx); err != nil {
		t.Fatalf("generateProjectJSON: %v", err)
	}
	raw := readFileString(t, filepath.Join(analysisDir, "project.json"))

	var got projectJSON
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("project.json is not valid JSON: %v\n%s", err, raw)
	}

	if got.ProjectSummary.Name != filepath.Base(project) {
		t.Errorf("name = %q, want the project directory base name %q", got.ProjectSummary.Name, filepath.Base(project))
	}
	if got.ProjectSummary.Type != "go" || got.ProjectSummary.Confidence != "medium" {
		t.Errorf("type/confidence = %q/%q, want go/medium", got.ProjectSummary.Type, got.ProjectSummary.Confidence)
	}
	if want := generateAIInsights("go", len(ctx.walk.files)).Overview; got.ProjectSummary.Description != want {
		t.Errorf("description = %q, want the AI insights overview %q", got.ProjectSummary.Description, want)
	}
	if got.ProjectSummary.Purpose != "AI-optimized development project" {
		t.Errorf("purpose = %q", got.ProjectSummary.Purpose)
	}
	if got.ProjectSummary.Complexity != "Small" {
		t.Errorf("complexity = %q, want Small for %d files", got.ProjectSummary.Complexity, len(ctx.walk.files))
	}
	// "go" is not part of any role list in projectClassification, so a Go +
	// Docker project is classified purely on its devops hit.
	if got.ProjectSummary.Classification != "infrastructure-focused" {
		t.Errorf("classification = %q, want infrastructure-focused", got.ProjectSummary.Classification)
	}
	if got.ProjectSummary.TotalFiles != len(ctx.walk.files) {
		t.Errorf("total_files = %d, want %d", got.ProjectSummary.TotalFiles, len(ctx.walk.files))
	}
	if got.ProjectSummary.AnalysisTimestamp == "" {
		t.Error("analysis_timestamp must be populated")
	}
	if !reflect.DeepEqual(got.FrameworksDetected, ctx.frameworks) {
		t.Errorf("frameworks_detected = %+v, want %+v", got.FrameworksDetected, ctx.frameworks)
	}
	if got.AIInsights.Overview != generateAIInsights("go", len(ctx.walk.files)).Overview {
		t.Errorf("ai_insights.overview = %q, want the generated insights overview", got.AIInsights.Overview)
	}
	if !reflect.DeepEqual(got.TechnologyStack.PrimaryLanguages, []string{"Go"}) {
		t.Errorf("primary_languages = %v", got.TechnologyStack.PrimaryLanguages)
	}
	if !reflect.DeepEqual(got.TechnologyStack.Frameworks, []string{"go", "docker"}) {
		t.Errorf("frameworks = %v", got.TechnologyStack.Frameworks)
	}
	if !reflect.DeepEqual(got.TechnologyStack.DevelopmentTools, []string{"Git", "Go Modules"}) {
		t.Errorf("development_tools = %v", got.TechnologyStack.DevelopmentTools)
	}
	if !reflect.DeepEqual(got.TechnologyStack.BuildSystems, []string{"Go Modules"}) {
		t.Errorf("build_systems = %v", got.TechnologyStack.BuildSystems)
	}
	if got.Structure.Files.Total != len(ctx.walk.files) ||
		got.Structure.Files.Code != ctx.fileCounts.SourceCode ||
		got.Structure.Files.Documentation != ctx.fileCounts.Documentation ||
		got.Structure.Files.Configuration != ctx.fileCounts.Configuration {
		t.Errorf("structure = %+v, want total %d and counts %+v",
			got.Structure.Files, len(ctx.walk.files), ctx.fileCounts)
	}
	if got.FrameworkVersion != "1.2.3" {
		t.Errorf("framework_version = %q", got.FrameworkVersion)
	}
	if got.AnalysisFiles.ProjectStructure != "project.structure.json" ||
		got.AnalysisFiles.ContextFile != "../context.md" ||
		got.AnalysisFiles.Overview != "overview.md" ||
		got.AnalysisFiles.GeneratedBy != "Plaesy Spec-Kit v1.2.3" {
		t.Errorf("analysis_files = %+v", got.AnalysisFiles)
	}
	// The file is indented with four spaces, not a compact dump.
	mustContain(t, raw, "\n    \"project_summary\": {")
}

func TestGenerateProjectStructureJSON(t *testing.T) {
	project := newFixture(t, map[string]string{"main.go": "package main"})
	analysisDir := filepath.Join(project, ".plaesy", "analysis")
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		t.Fatalf("mkdir analysis dir: %v", err)
	}
	ctx := testRunContext()

	if err := generateProjectStructureJSON(analysisDir, ctx); err != nil {
		t.Fatalf("generateProjectStructureJSON: %v", err)
	}
	raw := readFileString(t, filepath.Join(analysisDir, "project.structure.json"))

	var got projectStructureJSON
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("project.structure.json is not valid JSON: %v\n%s", err, raw)
	}

	// Only the singular "test" directory has a dedicated description; "tests"
	// falls back to the generic label.
	if want := map[string]dirEntry{
		"lib":   {FileCount: ctx.walk.dirs[0].fileCount, Description: "Source code library directory"},
		"tests": {FileCount: ctx.walk.dirs[1].fileCount, Description: "Project directory"},
	}; !reflect.DeepEqual(got.Directories, want) {
		t.Errorf("directories = %+v, want %+v", got.Directories, want)
	}

	// main.go, README.md and package.json are all key files; the size comes
	// from the walk.
	wantKey := map[string]keyFileEntry{
		"main.go":      {SizeBytes: int64(len("main.go")), Type: "go", Purpose: "Application entry point"},
		"README.md":    {SizeBytes: int64(len("README.md")), Type: "markdown", Purpose: "Project documentation"},
		"package.json": {SizeBytes: int64(len("package.json")), Type: "json", Purpose: "Node.js project configuration"},
	}
	if !reflect.DeepEqual(got.KeyFiles, wantKey) {
		t.Errorf("key_files = %+v, want %+v", got.KeyFiles, wantKey)
	}

	if got.FileTypes.SourceCode != ctx.fileCounts.SourceCode ||
		got.FileTypes.Documentation != ctx.fileCounts.Documentation ||
		got.FileTypes.Configuration != ctx.fileCounts.Configuration {
		t.Errorf("file_types = %+v, want %+v", got.FileTypes, ctx.fileCounts)
	}
	if got.RelatedFiles.ProjectSummary != "project.json" ||
		got.RelatedFiles.ContextFile != "../context.md" ||
		got.RelatedFiles.Overview != "overview.md" {
		t.Errorf("related_files = %+v", got.RelatedFiles)
	}
	if got.GeneratedBy != "Plaesy Spec-Kit v1.2.3" {
		t.Errorf("generated_by = %q", got.GeneratedBy)
	}
	if got.AnalysisTimestamp == "" {
		t.Error("analysis_timestamp must be populated")
	}
}

func TestGenerateProjectStructureJSONOmitsNonKeyFiles(t *testing.T) {
	analysisDir := t.TempDir()
	ctx := &runContext{
		walk: walkResult{
			files: files("src/widget.tsx", "data.csv", "lib/main.go"),
			dirs:  dirsOf("src"),
		},
		frameworkVersion: "0.0.1",
	}
	if err := generateProjectStructureJSON(analysisDir, ctx); err != nil {
		t.Fatalf("generateProjectStructureJSON: %v", err)
	}
	var got projectStructureJSON
	raw := readFileString(t, filepath.Join(analysisDir, "project.structure.json"))
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := got.KeyFiles["src/widget.tsx"]; ok {
		t.Error("a plain source file must not be listed as a key file")
	}
	if _, ok := got.KeyFiles["data.csv"]; ok {
		t.Error("a csv file must not be listed as a key file")
	}
	if _, ok := got.KeyFiles["lib/main.go"]; !ok {
		t.Errorf("lib/main.go should be a key file, got %+v", got.KeyFiles)
	}
	if got.GeneratedBy != "Plaesy Spec-Kit v0.0.1" {
		t.Errorf("generated_by = %q", got.GeneratedBy)
	}
}

func TestGenerateOverviewMD(t *testing.T) {
	project := newFixture(t, map[string]string{
		"main.go":                  "package main",
		".github/workflows/ci.yml": "on: push",
	})
	analysisDir := filepath.Join(project, ".plaesy", "analysis")
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		t.Fatalf("mkdir analysis dir: %v", err)
	}

	cases := []struct {
		name     string
		docs     int
		dirs     []dirInfo
		project  string
		wantSubs []string
		notSubs  []string
	}{
		{
			name:     "gaps are reported when docs, tests and CI are all missing",
			docs:     0,
			dirs:     dirsOf("lib"),
			project:  "go",
			wantSubs: []string{"(only 0 doc file(s) found)", "no test/spec directory detected"},
			notSubs:  []string{"no CI configuration detected", "No gaps detected"},
		},
		{
			name:    "nothing to recommend when docs, tests and CI are all present",
			docs:    3,
			dirs:    dirsOf("tests", "docs"),
			project: "spec-kit",
			wantSubs: []string{
				"No gaps detected against baseline checks (docs, tests, CI)",
				"Plaesy Spec-Kit framework for AI-assisted development workflow automation",
			},
			notSubs: []string{"Add comprehensive documentation", "Implement automated testing"},
		},
		{
			name:     "well documented project without tests only reports the test gap",
			docs:     5,
			dirs:     dirsOf("lib"),
			project:  "django",
			wantSubs: []string{"no test/spec directory detected"},
			notSubs:  []string{"Add comprehensive documentation", "No gaps detected"},
		},
		{
			name:     "generic description for non spec-kit projects",
			docs:     1,
			dirs:     dirsOf("spec"),
			project:  "gin",
			wantSubs: []string{"This is an AI-generated project context document for development assistance"},
			notSubs:  []string{"Plaesy Spec-Kit framework for AI-assisted development"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &runContext{
				walk:             walkResult{files: files("main.go"), dirs: tc.dirs},
				projectType:      FrameworkHit{Framework: tc.project, Confidence: "high"},
				languages:        []string{"Go", "Python"},
				fileCounts:       fileCounts{SourceCode: 1, Documentation: tc.docs, Configuration: 0},
				frameworkVersion: "1.2.3",
			}
			if err := generateOverviewMD(project, analysisDir, ctx); err != nil {
				t.Fatalf("generateOverviewMD: %v", err)
			}
			got := readFileString(t, filepath.Join(analysisDir, "overview.md"))

			for _, s := range []string{
				"# Project Analysis Overview",
				"regenerated (replaced, not appended)",
				"**Project**: " + filepath.Base(project),
				"## Technology Stack",
				"- **Languages**: Go, Python",
				"- **Framework**: " + tc.project,
				"- **Tools**: Git, Plaesy CLI",
				"## File Structure",
				"| File Type | Count |",
				"|-----------|-------|",
				"| Source Code | 1 |",
				"| Documentation | " + strconv.Itoa(tc.docs) + " |",
				"| Config | 0 |",
				"## Recommendations",
				"## Related Analysis Files",
				"`project.json`",
				"`project.structure.json`",
				"*Generated by Plaesy Spec-Kit v1.2.3*",
			} {
				mustContain(t, got, s)
			}
			for _, s := range tc.wantSubs {
				mustContain(t, got, s)
			}
			for _, s := range tc.notSubs {
				mustNotContain(t, got, s)
			}
		})
	}
}

func TestGenerateOverviewMDReplacesPreviousContent(t *testing.T) {
	project := newFixture(t, map[string]string{"main.go": "package main"})
	analysisDir := filepath.Join(project, ".plaesy", "analysis")
	if err := os.MkdirAll(analysisDir, 0o755); err != nil {
		t.Fatalf("mkdir analysis dir: %v", err)
	}
	ctx := &runContext{
		walk:             walkResult{files: files("main.go"), dirs: dirsOf("lib")},
		projectType:      FrameworkHit{Framework: "go", Confidence: "medium"},
		languages:        []string{"Go"},
		fileCounts:       fileCounts{SourceCode: 1},
		frameworkVersion: "1.0.0",
	}
	if err := generateOverviewMD(project, analysisDir, ctx); err != nil {
		t.Fatalf("first generateOverviewMD: %v", err)
	}
	if err := generateOverviewMD(project, analysisDir, ctx); err != nil {
		t.Fatalf("second generateOverviewMD: %v", err)
	}
	got := readFileString(t, filepath.Join(analysisDir, "overview.md"))
	if n := strings.Count(got, "# Project Analysis Overview"); n != 1 {
		t.Errorf("header appears %d times, want exactly 1 (content must be replaced, not appended)", n)
	}
}

func TestHasTestDir(t *testing.T) {
	cases := []struct {
		name string
		dirs []dirInfo
		want bool
	}{
		{"none", nil, false},
		{"tests", dirsOf("lib", "tests"), true},
		{"test singular", dirsOf("test"), true},
		{"spec", dirsOf("spec"), true},
		{"__tests__", dirsOf("src/__tests__"), true},
		{"case insensitive", dirsOf("TESTS"), true},
		{"unrelated", dirsOf("lib", "docs"), false},
		{"prefix match on a longer word", dirsOf("testimonials"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasTestDir(tc.dirs); got != tc.want {
				t.Errorf("hasTestDir(%+v) = %v, want %v", tc.dirs, got, tc.want)
			}
		})
	}
}

func TestHasCIConfig(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  bool
	}{
		{"github workflows", map[string]string{".github/workflows/ci.yml": "on: push"}, true},
		{"gitlab ci", map[string]string{".gitlab-ci.yml": "stages: []"}, true},
		{"travis", map[string]string{".travis.yml": "language: go"}, true},
		{"none", map[string]string{"main.go": "package main"}, false},
		{"github directory without workflows", map[string]string{".github/ISSUE_TEMPLATE.md": "x"}, false},
		{"a directory named .gitlab-ci.yml does not count", map[string]string{".gitlab-ci.yml/": ""}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newFixture(t, tc.files)
			if got := hasCIConfig(root); got != tc.want {
				t.Errorf("hasCIConfig() = %v, want %v", got, tc.want)
			}
		})
	}
}
