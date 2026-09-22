package analyze

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// fileCounts mirrors count_file_types(): a single pass classifying every
// project file into source/doc/config buckets by extension.
type fileCounts struct {
	SourceCode    int
	Documentation int
	Configuration int
}

var sourceCodeExtRE = regexp.MustCompile(`\.(js|jsx|ts|tsx|py|go|java|dart|sh|bash|ps1|psm1|rb|php|rs|c|cc|cpp|h|hpp|cs|kt|kts|swift)$`)
var docExtRE = regexp.MustCompile(`\.(md|txt|rst|adoc)$`)
var cfgExtRE = regexp.MustCompile(`\.(json|yaml|yml|toml|ini|xml|cfg)$`)

func countFileTypes(files []fileInfo) fileCounts {
	var fc fileCounts
	for _, f := range files {
		name := strings.ToLower(filepath.Base(f.relPath))
		switch {
		case sourceCodeExtRE.MatchString(name):
			fc.SourceCode++
		case docExtRE.MatchString(name):
			fc.Documentation++
		case cfgExtRE.MatchString(name):
			fc.Configuration++
		}
	}
	return fc
}

// directoryDescription mirrors generate_directory_description().
func directoryDescription(relPath string) string {
	switch {
	case relPath == "lib":
		return "Source code library directory"
	case strings.HasPrefix(relPath, "lib/screens"):
		return "UI screen components"
	case relPath == "lib/models":
		return "Data models and entities"
	case relPath == "lib/services":
		return "Business logic and API services"
	case relPath == "lib/utils":
		return "Utility functions and helpers"
	case relPath == "test":
		return "Test files and unit tests"
	case relPath == "android":
		return "Android platform specific code"
	case relPath == "ios":
		return "iOS platform specific code"
	case relPath == "assets":
		return "Static assets (images, fonts, etc.)"
	case relPath == "assets/images":
		return "Image assets"
	case relPath == "assets/data":
		return "Data files"
	case relPath == "docs":
		return "Documentation files"
	case relPath == "scripts":
		return "Build and utility scripts"
	default:
		return "Project directory"
	}
}

// fileType mirrors get_file_type().
func fileType(relPath string) string {
	switch {
	case strings.HasSuffix(relPath, ".dart"):
		return "dart"
	case strings.HasSuffix(relPath, ".js"), strings.HasSuffix(relPath, ".jsx"):
		return "javascript"
	case strings.HasSuffix(relPath, ".ts"), strings.HasSuffix(relPath, ".tsx"):
		return "typescript"
	case strings.HasSuffix(relPath, ".py"):
		return "python"
	case strings.HasSuffix(relPath, ".go"):
		return "go"
	case strings.HasSuffix(relPath, ".java"):
		return "java"
	case strings.HasSuffix(relPath, ".json"):
		return "json"
	case strings.HasSuffix(relPath, ".yaml"), strings.HasSuffix(relPath, ".yml"):
		return "yaml"
	case strings.HasSuffix(relPath, ".md"):
		return "markdown"
	default:
		return "text"
	}
}

// filePurpose mirrors get_file_purpose().
func filePurpose(relPath string) string {
	base := filepath.Base(relPath)
	switch {
	case base == "pubspec.yaml":
		return "Flutter/Dart project configuration"
	case base == "package.json":
		return "Node.js project configuration"
	case base == "requirements.txt":
		return "Python dependencies"
	case base == "go.mod":
		return "Go module configuration"
	case base == "Cargo.toml":
		return "Rust project configuration"
	case strings.HasPrefix(base, "README"):
		return "Project documentation"
	case base == "LICENSE":
		return "Project license"
	case strings.HasPrefix(base, "main."), strings.HasPrefix(base, "index."):
		return "Application entry point"
	case strings.Contains(base, ".config."), strings.HasSuffix(base, ".config"):
		return "Configuration file"
	case base == ".gitignore":
		return "Git ignore rules"
	case strings.HasPrefix(base, ".env"):
		return "Environment variables"
	default:
		return "Project file"
	}
}

// isKeyFile mirrors the `find` pattern list used to build key_files in
// generate_project_structure_json().
func isKeyFile(relPath string) bool {
	base := filepath.Base(relPath)
	switch base {
	case "package.json", "pubspec.yaml", "requirements.txt", "go.mod", "Cargo.toml",
		"main.dart", "main.js", "main.py", "main.go", "index.js":
		return true
	}
	if strings.HasSuffix(base, ".md") || strings.HasPrefix(base, "README") || base == "LICENSE" {
		return true
	}
	if strings.HasSuffix(base, ".json") || strings.HasSuffix(base, ".yaml") || strings.HasSuffix(base, ".yml") {
		return true
	}
	return false
}

// --- project.json ---

type projectSummary struct {
	Name              string `json:"name"`
	Type              string `json:"type"`
	Confidence        string `json:"confidence"`
	Description       string `json:"description"`
	Purpose           string `json:"purpose"`
	Complexity        string `json:"complexity"`
	Classification    string `json:"classification"`
	TotalFiles        int    `json:"total_files"`
	AnalysisTimestamp string `json:"analysis_timestamp"`
}

type technologyStack struct {
	PrimaryLanguages []string `json:"primary_languages"`
	Frameworks       []string `json:"frameworks"`
	DevelopmentTools []string `json:"development_tools"`
	BuildSystems     []string `json:"build_systems"`
}

type structureFiles struct {
	Total         int `json:"total"`
	Code          int `json:"code"`
	Documentation int `json:"documentation"`
	Configuration int `json:"configuration"`
}

type structureSummary struct {
	Files structureFiles `json:"files"`
}

type analysisFiles struct {
	ProjectStructure string `json:"project_structure"`
	ContextFile      string `json:"context_file"`
	Overview         string `json:"overview"`
	GeneratedBy      string `json:"generated_by"`
}

type projectJSON struct {
	ProjectSummary     projectSummary   `json:"project_summary"`
	FrameworksDetected []FrameworkHit   `json:"frameworks_detected"`
	AIInsights         aiInsights       `json:"ai_insights"`
	TechnologyStack    technologyStack  `json:"technology_stack"`
	Structure          structureSummary `json:"structure"`
	FrameworkVersion   string           `json:"framework_version"`
	AnalysisFiles      analysisFiles    `json:"analysis_files"`
}

// projectClassification mirrors the multi-framework classification block in
// generate_project_json(): single | multi-framework | full-stack |
// full-stack-with-infrastructure | infrastructure-focused | mobile-backend.
func projectClassification(frameworks []FrameworkHit) string {
	if len(frameworks) <= 1 {
		return "single"
	}

	var hasFrontend, hasBackend, hasMobile, hasDevops bool
	for _, h := range frameworks {
		switch h.Framework {
		case "nextjs", "react", "vue", "nuxt", "svelte", "astro", "remix", "gatsby", "express", "nestjs":
			hasFrontend = true
		case "django", "flask", "fastapi", "laravel", "rails", "springboot", "actix", "rocket", "gin", "echo", "fiber":
			hasBackend = true
		case "flutter":
			hasMobile = true
		case "docker", "docker-compose", "terraform", "kustomize", "vagrant":
			hasDevops = true
		}
	}

	switch {
	case hasFrontend && hasBackend:
		return "full-stack"
	case hasDevops && (hasFrontend || hasBackend):
		return "full-stack-with-infrastructure"
	case hasDevops:
		return "infrastructure-focused"
	case hasMobile && (hasBackend || hasFrontend):
		return "mobile-backend"
	default:
		return "multi-framework"
	}
}

func complexityFor(totalFiles int) string {
	switch {
	case totalFiles > 50:
		return "Large"
	case totalFiles > 20:
		return "Medium"
	default:
		return "Small"
	}
}

func timestamp() string {
	return time.Now().Format("2006-01-02T15:04:05-07:00")
}

func frameworkNames(hits []FrameworkHit) []string {
	names := make([]string, len(hits))
	for i, h := range hits {
		names[i] = h.Framework
	}
	return names
}

func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// generateProjectJSON mirrors generate_project_json().
func generateProjectJSON(analysisDir, projectPath string, ctx *runContext) error {
	fc := ctx.fileCounts
	totalFiles := len(ctx.walk.files)
	projectName := filepath.Base(strings.TrimRight(filepath.Clean(projectPath), string(filepath.Separator)))

	projectType := ctx.projectType.Framework
	confidence := ctx.projectType.Confidence
	insights := generateAIInsights(projectType, totalFiles)

	pj := projectJSON{
		ProjectSummary: projectSummary{
			Name:              projectName,
			Type:              projectType,
			Confidence:        confidence,
			Description:       insights.Overview,
			Purpose:           "AI-optimized development project",
			Complexity:        complexityFor(totalFiles),
			Classification:    projectClassification(ctx.frameworks),
			TotalFiles:        totalFiles,
			AnalysisTimestamp: timestamp(),
		},
		FrameworksDetected: ctx.frameworks,
		AIInsights:         insights,
		TechnologyStack: technologyStack{
			PrimaryLanguages: ctx.languages,
			Frameworks:       frameworkNames(ctx.frameworks),
			DevelopmentTools: ctx.devTools,
			BuildSystems:     ctx.buildSystems,
		},
		Structure: structureSummary{
			Files: structureFiles{
				Total:         totalFiles,
				Code:          fc.SourceCode,
				Documentation: fc.Documentation,
				Configuration: fc.Configuration,
			},
		},
		FrameworkVersion: ctx.frameworkVersion,
		AnalysisFiles: analysisFiles{
			ProjectStructure: "project.structure.json",
			ContextFile:      "../context.md",
			Overview:         "overview.md",
			GeneratedBy:      "Plaesy Spec-Kit v" + ctx.frameworkVersion,
		},
	}

	return writeJSON(filepath.Join(analysisDir, "project.json"), pj)
}

// --- project.structure.json ---

type dirEntry struct {
	FileCount   int    `json:"file_count"`
	Description string `json:"description"`
}

type keyFileEntry struct {
	SizeBytes int64  `json:"size_bytes"`
	Type      string `json:"type"`
	Purpose   string `json:"purpose"`
}

type fileTypeCounts struct {
	SourceCode    int `json:"source_code"`
	Documentation int `json:"documentation"`
	Configuration int `json:"configuration"`
}

type relatedFiles struct {
	ProjectSummary string `json:"project_summary"`
	ContextFile    string `json:"context_file"`
	Overview       string `json:"overview"`
}

type projectStructureJSON struct {
	Directories       map[string]dirEntry     `json:"directories"`
	KeyFiles          map[string]keyFileEntry `json:"key_files"`
	FileTypes         fileTypeCounts          `json:"file_types"`
	AnalysisTimestamp string                  `json:"analysis_timestamp"`
	RelatedFiles      relatedFiles            `json:"related_files"`
	GeneratedBy       string                  `json:"generated_by"`
}

// generateProjectStructureJSON mirrors generate_project_structure_json().
func generateProjectStructureJSON(analysisDir string, ctx *runContext) error {
	directories := make(map[string]dirEntry)
	for _, d := range ctx.walk.dirs {
		directories[d.relPath] = dirEntry{
			FileCount:   d.fileCount,
			Description: directoryDescription(d.relPath),
		}
	}

	keyFiles := make(map[string]keyFileEntry)
	for _, f := range ctx.walk.files {
		if !isKeyFile(f.relPath) {
			continue
		}
		keyFiles[f.relPath] = keyFileEntry{
			SizeBytes: f.size,
			Type:      fileType(f.relPath),
			Purpose:   filePurpose(f.relPath),
		}
	}

	fc := ctx.fileCounts
	psj := projectStructureJSON{
		Directories: directories,
		KeyFiles:    keyFiles,
		FileTypes: fileTypeCounts{
			SourceCode:    fc.SourceCode,
			Documentation: fc.Documentation,
			Configuration: fc.Configuration,
		},
		AnalysisTimestamp: timestamp(),
		RelatedFiles: relatedFiles{
			ProjectSummary: "project.json",
			ContextFile:    "../context.md",
			Overview:       "overview.md",
		},
		GeneratedBy: "Plaesy Spec-Kit v" + ctx.frameworkVersion,
	}

	return writeJSON(filepath.Join(analysisDir, "project.structure.json"), psj)
}

// --- overview.md ---

// generateOverviewMD mirrors generate_analysis_overview_md(): always
// replaces (never appends) .plaesy/analysis/overview.md.
func generateOverviewMD(projectPath, analysisDir string, ctx *runContext) error {
	languagesReadable := strings.Join(ctx.languages, ", ")
	projectType := ctx.projectType.Framework

	projectDesc := "This is an AI-generated project context document for development assistance"
	if projectType == "spec-kit" {
		projectDesc = "Plaesy Spec-Kit framework for AI-assisted development workflow automation"
	}

	projectName := filepath.Base(strings.TrimRight(filepath.Clean(projectPath), string(filepath.Separator)))

	fc := ctx.fileCounts

	var recs []string
	if fc.Documentation < 3 {
		recs = append(recs, fmt.Sprintf("- Add comprehensive documentation (only %d doc file(s) found)", fc.Documentation))
	}
	if !hasTestDir(ctx.walk.dirs) {
		recs = append(recs, "- Implement automated testing (no test/spec directory detected)")
	}
	if !hasCIConfig(projectPath) {
		recs = append(recs, "- Set up CI/CD pipeline (no CI configuration detected)")
	}
	if len(recs) == 0 {
		recs = append(recs, "- No gaps detected against baseline checks (docs, tests, CI) — keep it up.")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# Project Analysis Overview\n\n")
	fmt.Fprintf(&b, "*This file is regenerated (replaced, not appended) on every `plaesy analyze` run.\n")
	fmt.Fprintf(&b, "Manual notes belong in `.plaesy/context.md` / `.plaesy/memory.md`, which analyze no longer touches.*\n\n")
	fmt.Fprintf(&b, "**Generated**: %s\n", timestamp())
	fmt.Fprintf(&b, "**Project**: %s\n", projectName)
	fmt.Fprintf(&b, "**Description**: %s\n\n", projectDesc)
	fmt.Fprintf(&b, "## Technology Stack\n")
	fmt.Fprintf(&b, "- **Languages**: %s\n", languagesReadable)
	fmt.Fprintf(&b, "- **Framework**: %s\n", projectType)
	fmt.Fprintf(&b, "- **Tools**: Git, Plaesy CLI\n\n")
	fmt.Fprintf(&b, "## File Structure\n")
	fmt.Fprintf(&b, "| File Type | Count |\n")
	fmt.Fprintf(&b, "|-----------|-------|\n")
	fmt.Fprintf(&b, "| Source Code | %d |\n", fc.SourceCode)
	fmt.Fprintf(&b, "| Documentation | %d |\n", fc.Documentation)
	fmt.Fprintf(&b, "| Config | %d |\n\n", fc.Configuration)
	fmt.Fprintf(&b, "## Recommendations\n%s\n\n", strings.Join(recs, "\n"))
	fmt.Fprintf(&b, "## Related Analysis Files\n")
	fmt.Fprintf(&b, "- **Project Summary**: `project.json` - Complete project overview and AI insights\n")
	fmt.Fprintf(&b, "- **Project Structure**: `project.structure.json` - Detailed file and directory analysis\n\n")
	fmt.Fprintf(&b, "---\n*Generated by Plaesy Spec-Kit v%s*\n", ctx.frameworkVersion)

	return os.WriteFile(filepath.Join(analysisDir, "overview.md"), []byte(b.String()), 0o644)
}

func hasTestDir(dirs []dirInfo) bool {
	for _, d := range dirs {
		base := strings.ToLower(filepath.Base(d.relPath))
		if strings.HasPrefix(base, "test") || strings.HasPrefix(base, "spec") || base == "__tests__" {
			return true
		}
	}
	return false
}

func hasCIConfig(projectPath string) bool {
	if dirExists(filepath.Join(projectPath, ".github", "workflows")) {
		return true
	}
	return fileExists(filepath.Join(projectPath, ".gitlab-ci.yml")) || fileExists(filepath.Join(projectPath, ".travis.yml"))
}
