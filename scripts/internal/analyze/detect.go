// Package analyze ports scripts/bash/plaesy-analyze.sh to Go. It analyzes a
// project directory and writes AI-oriented analysis artifacts under
// <project>/.plaesy/analysis/.
package analyze

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FrameworkHit is one detected framework with a confidence level.
type FrameworkHit struct {
	Framework  string `json:"framework"`
	Confidence string `json:"confidence"`
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func anyFileMatches(dir, pattern string) bool {
	matches, _ := filepath.Glob(filepath.Join(dir, pattern))
	return len(matches) > 0
}

// fileContains reports whether any of the given files (relative to dir, glob
// patterns allowed) contain substr. Mirrors `grep -q substr file` semantics,
// tolerating a missing file (grep's 2>/dev/null fallback).
func fileContains(dir, substr string, patterns ...string) bool {
	for _, pattern := range patterns {
		var paths []string
		if strings.ContainsAny(pattern, "*?[") {
			matches, _ := filepath.Glob(filepath.Join(dir, pattern))
			paths = matches
		} else {
			paths = []string{filepath.Join(dir, pattern)}
		}
		for _, p := range paths {
			data, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			if strings.Contains(string(data), substr) {
				return true
			}
		}
	}
	return false
}

func has(hits []FrameworkHit, name string) bool {
	for _, h := range hits {
		if h.Framework == name {
			return true
		}
	}
	return false
}

// detectAllFrameworks mirrors detect_all_frameworks() in the bash script.
func detectAllFrameworks(projectPath string) []FrameworkHit {
	var hits []FrameworkHit
	add := func(name, confidence string) {
		if !has(hits, name) {
			hits = append(hits, FrameworkHit{name, confidence})
		}
	}

	// JavaScript/TypeScript frameworks
	pkgJSON := filepath.Join(projectPath, "package.json")
	if fileExists(pkgJSON) {
		before := len(hits)
		for _, pair := range []struct{ needle, name string }{
			{"next", "nextjs"}, {"react", "react"}, {"vue", "vue"},
			{"svelte", "svelte"}, {"astro", "astro"}, {"remix", "remix"},
			{"gatsby", "gatsby"}, {"nuxt", "nuxt"}, {"express", "express"},
			{"nestjs", "nestjs"},
		} {
			if fileContains(projectPath, pair.needle, "package.json") {
				add(pair.name, "high")
			}
		}
		if len(hits) == before {
			add("nodejs", "medium")
		}
	}

	// Additional JS/TS config files
	if fileExists(filepath.Join(projectPath, "next.config.js")) || fileExists(filepath.Join(projectPath, "next.config.mjs")) {
		add("nextjs", "high")
	}
	if fileExists(filepath.Join(projectPath, "nuxt.config.js")) || fileExists(filepath.Join(projectPath, "nuxt.config.ts")) {
		add("nuxt", "high")
	}
	if fileExists(filepath.Join(projectPath, "svelte.config.js")) {
		add("svelte", "high")
	}
	if fileExists(filepath.Join(projectPath, "astro.config.mjs")) || fileExists(filepath.Join(projectPath, "astro.config.ts")) {
		add("astro", "high")
	}
	if fileExists(filepath.Join(projectPath, "remix.config.js")) {
		add("remix", "high")
	}
	if fileExists(filepath.Join(projectPath, "gatsby-config.js")) {
		add("gatsby", "high")
	}
	if fileExists(filepath.Join(projectPath, "vite.config.js")) || fileExists(filepath.Join(projectPath, "vite.config.ts")) {
		add("vite", "medium")
	}

	// Python frameworks
	pyproject := filepath.Join(projectPath, "pyproject.toml")
	if fileExists(pyproject) {
		before := len(hits)
		if fileContains(projectPath, "django", "pyproject.toml") {
			add("django", "high")
		}
		if fileContains(projectPath, "fastapi", "pyproject.toml") {
			add("fastapi", "high")
		}
		if fileContains(projectPath, "poetry", "pyproject.toml") {
			add("poetry", "medium")
		}
		if len(hits) == before {
			add("python", "medium")
		}
	}
	reqTxt := filepath.Join(projectPath, "requirements.txt")
	if fileExists(reqTxt) {
		before := len(hits)
		if fileContains(projectPath, "django", "requirements.txt") {
			add("django", "high")
		}
		if fileContains(projectPath, "fastapi", "requirements.txt") {
			add("fastapi", "high")
		}
		if fileContains(projectPath, "flask", "requirements.txt") {
			add("flask", "high")
		}
		if len(hits) == before {
			add("python", "medium")
		}
	}
	if fileExists(filepath.Join(projectPath, "Pipfile")) {
		add("pipenv", "high")
	}
	if fileExists(filepath.Join(projectPath, "manage.py")) {
		add("django", "high")
	}

	// Mobile frameworks
	if fileExists(filepath.Join(projectPath, "pubspec.yaml")) {
		add("flutter", "high")
	}

	// Go frameworks
	if fileExists(filepath.Join(projectPath, "go.mod")) {
		before := len(hits)
		if fileContains(projectPath, "gin-gonic", "go.mod") {
			add("gin", "high")
		}
		if fileContains(projectPath, "labstack/echo", "go.mod") {
			add("echo", "high")
		}
		if fileContains(projectPath, "gofiber", "go.mod") {
			add("fiber", "high")
		}
		if len(hits) == before {
			add("go", "medium")
		}
	}

	// Rust frameworks
	if fileExists(filepath.Join(projectPath, "Cargo.toml")) {
		before := len(hits)
		if fileContains(projectPath, "actix-web", "Cargo.toml") {
			add("actix", "high")
		}
		if fileContains(projectPath, "rocket", "Cargo.toml") {
			add("rocket", "high")
		}
		if fileContains(projectPath, "axum", "Cargo.toml") {
			add("axum", "high")
		}
		if len(hits) == before {
			add("rust", "medium")
		}
	}

	// Java frameworks (Maven)
	if fileExists(filepath.Join(projectPath, "pom.xml")) {
		before := len(hits)
		if fileContains(projectPath, "spring-boot", "pom.xml") {
			add("springboot", "high")
		}
		if fileContains(projectPath, "spring", "pom.xml") {
			add("spring", "high")
		}
		if len(hits) == before {
			add("maven", "medium")
		}
	}
	// Java frameworks (Gradle)
	if fileExists(filepath.Join(projectPath, "build.gradle")) || fileExists(filepath.Join(projectPath, "build.gradle.kts")) {
		before := len(hits)
		if fileContains(projectPath, "org.springframework.boot", "build.gradle", "build.gradle.kts") {
			add("springboot", "high")
		}
		if fileContains(projectPath, "io.ktor", "build.gradle", "build.gradle.kts") {
			add("ktor", "high")
		}
		if len(hits) == before {
			add("gradle", "medium")
		}
	}

	// Ruby frameworks
	if fileExists(filepath.Join(projectPath, "Gemfile")) {
		before := len(hits)
		if fileContains(projectPath, "rails", "Gemfile") {
			add("rails", "high")
		}
		if fileContains(projectPath, "sinatra", "Gemfile") {
			add("sinatra", "high")
		}
		if len(hits) == before {
			add("ruby", "medium")
		}
	}
	if fileExists(filepath.Join(projectPath, "config", "application.rb")) {
		add("rails", "high")
	}

	// PHP frameworks
	if fileExists(filepath.Join(projectPath, "composer.json")) {
		before := len(hits)
		if fileContains(projectPath, "laravel/framework", "composer.json") {
			add("laravel", "high")
		}
		if fileContains(projectPath, "symfony", "composer.json") {
			add("symfony", "high")
		}
		if len(hits) == before {
			add("php", "medium")
		}
	}
	if fileExists(filepath.Join(projectPath, "wp-config.php")) {
		add("wordpress", "high")
	}

	// C#/.NET frameworks
	if anyFileMatches(projectPath, "*.csproj") {
		before := len(hits)
		if fileContains(projectPath, "Microsoft.AspNetCore", "*.csproj") {
			add("aspnet", "high")
		}
		if len(hits) == before {
			add("dotnet", "medium")
		}
	}

	// Swift
	if fileExists(filepath.Join(projectPath, "Package.swift")) {
		add("swift", "high")
	}

	// Container/DevOps
	if fileExists(filepath.Join(projectPath, "Dockerfile")) {
		add("docker", "medium")
	}
	if fileExists(filepath.Join(projectPath, "docker-compose.yml")) || fileExists(filepath.Join(projectPath, "docker-compose.yaml")) {
		add("docker-compose", "medium")
	}
	if fileExists(filepath.Join(projectPath, "Terrafile")) || fileExists(filepath.Join(projectPath, "main.tf")) {
		add("terraform", "high")
	}

	// Configuration management
	if fileExists(filepath.Join(projectPath, "Vagrantfile")) {
		add("vagrant", "high")
	}
	if fileExists(filepath.Join(projectPath, "kustomization.yaml")) {
		add("kustomize", "high")
	}

	// Plaesy framework itself
	if fileContains(projectPath, "Plaesy Spec-Kit", "README.md") {
		add("spec-kit", "high")
	}

	return hits
}

// priorityFrameworks mirrors the priority_frameworks array in
// detect_project_type().
var priorityFrameworks = []string{
	"nextjs", "react", "vue", "nuxt", "svelte", "astro", "remix", "gatsby",
	"django", "flask", "fastapi", "laravel", "rails", "express", "nestjs",
	"springboot", "flutter", "actix", "rocket", "gin", "echo", "fiber",
	"rust", "go", "java", "kotlin", "nodejs", "python", "php", "ruby", "swift",
}

// detectProjectType mirrors detect_project_type(): pick a single primary
// framework+confidence from the full hit list.
func detectProjectType(hits []FrameworkHit) FrameworkHit {
	if len(hits) == 0 {
		return FrameworkHit{"generic", "low"}
	}
	if len(hits) == 1 {
		return hits[0]
	}
	for _, p := range priorityFrameworks {
		for _, h := range hits {
			if h.Framework == p {
				return h
			}
		}
	}
	return hits[0]
}

// languageDef pairs a canonical language name with the file extensions that
// indicate it, mirroring the awk pass in detect_all_languages().
type languageDef struct {
	name string
	re   *regexp.Regexp
}

var languageDefs = []languageDef{
	{"JavaScript", regexp.MustCompile(`\.(js|jsx)$`)},
	{"TypeScript", regexp.MustCompile(`\.(ts|tsx)$`)},
	{"Python", regexp.MustCompile(`\.py$`)},
	{"Go", regexp.MustCompile(`\.go$`)},
	{"Java", regexp.MustCompile(`\.java$`)},
	{"Dart", regexp.MustCompile(`\.dart$`)},
	{"C++", regexp.MustCompile(`\.(cpp|cc|cxx)$`)},
	{"C", regexp.MustCompile(`\.c$`)},
	{"C/C++ Headers", regexp.MustCompile(`\.h$`)},
	{"PHP", regexp.MustCompile(`\.php$`)},
	{"Ruby", regexp.MustCompile(`\.rb$`)},
	{"Rust", regexp.MustCompile(`\.rs$`)},
	{"Swift", regexp.MustCompile(`\.swift$`)},
	{"Kotlin", regexp.MustCompile(`\.kt$`)},
	{"Scala", regexp.MustCompile(`\.scala$`)},
	{"Shell", regexp.MustCompile(`\.sh$`)},
	{"HTML", regexp.MustCompile(`\.html$`)},
	{"CSS", regexp.MustCompile(`\.(css|scss|sass)$`)},
}

// detectAllLanguages walks the project once and returns detected languages
// in the canonical order used by the bash script, defaulting to
// ["JavaScript"] when nothing matches.
func detectAllLanguages(files []string) []string {
	seen := make(map[string]bool)
	for _, f := range files {
		base := filepath.Base(f)
		for _, ld := range languageDefs {
			if ld.re.MatchString(base) {
				seen[ld.name] = true
			}
		}
	}
	var out []string
	for _, ld := range languageDefs {
		if seen[ld.name] {
			out = append(out, ld.name)
		}
	}
	if len(out) == 0 {
		out = []string{"JavaScript"}
	}
	return out
}

func addTool(tools *[]string, name string) {
	for _, t := range *tools {
		if t == name {
			return
		}
	}
	*tools = append(*tools, name)
}

// detectDevelopmentTools mirrors detect_development_tools().
func detectDevelopmentTools(projectPath string) []string {
	var tools []string

	if dirExists(filepath.Join(projectPath, ".git")) {
		addTool(&tools, "Git")
	}
	if dirExists(filepath.Join(projectPath, ".svn")) {
		addTool(&tools, "Subversion")
	}
	if fileExists(filepath.Join(projectPath, ".hg")) {
		addTool(&tools, "Mercurial")
	}

	if fileExists(filepath.Join(projectPath, "package.json")) || fileExists(filepath.Join(projectPath, "package-lock.json")) ||
		fileExists(filepath.Join(projectPath, "yarn.lock")) || fileExists(filepath.Join(projectPath, "pnpm-lock.yaml")) {
		addTool(&tools, "npm/yarn/pnpm")
	}
	if fileExists(filepath.Join(projectPath, "requirements.txt")) || fileExists(filepath.Join(projectPath, "Pipfile")) ||
		fileExists(filepath.Join(projectPath, "pyproject.toml")) {
		addTool(&tools, "pip/poetry")
	}
	if fileExists(filepath.Join(projectPath, "go.mod")) {
		addTool(&tools, "Go Modules")
	}
	if fileExists(filepath.Join(projectPath, "Cargo.toml")) {
		addTool(&tools, "Cargo")
	}
	if fileExists(filepath.Join(projectPath, "pom.xml")) || fileExists(filepath.Join(projectPath, "build.gradle")) ||
		fileExists(filepath.Join(projectPath, "build.gradle.kts")) {
		addTool(&tools, "Maven/Gradle")
	}
	if fileExists(filepath.Join(projectPath, "Gemfile")) {
		addTool(&tools, "Bundler")
	}
	if fileExists(filepath.Join(projectPath, "composer.json")) {
		addTool(&tools, "Composer")
	}
	if fileExists(filepath.Join(projectPath, "pubspec.yaml")) {
		addTool(&tools, "Pub")
	}

	if dirExists(filepath.Join(projectPath, ".github", "workflows")) {
		addTool(&tools, "GitHub Actions")
	}
	if fileExists(filepath.Join(projectPath, ".gitlab-ci.yml")) {
		addTool(&tools, "GitLab CI")
	}
	if fileExists(filepath.Join(projectPath, "Jenkinsfile")) {
		addTool(&tools, "Jenkins")
	}
	if fileExists(filepath.Join(projectPath, "azure-pipelines.yml")) {
		addTool(&tools, "Azure Pipelines")
	}

	if fileExists(filepath.Join(projectPath, "jest.config.js")) || fileExists(filepath.Join(projectPath, "jest.config.json")) ||
		fileContains(projectPath, "jest", "package.json") {
		addTool(&tools, "Jest")
	}
	if fileExists(filepath.Join(projectPath, "vitest.config.js")) || fileExists(filepath.Join(projectPath, "vitest.config.ts")) {
		addTool(&tools, "Vitest")
	}
	if fileContains(projectPath, "pytest", "requirements.txt") || fileExists(filepath.Join(projectPath, "pytest.ini")) ||
		dirExists(filepath.Join(projectPath, "tests")) {
		addTool(&tools, "pytest")
	}
	if fileExists(filepath.Join(projectPath, "go.mod")) && fileContains(projectPath, "test", "go.mod") {
		addTool(&tools, "Go test")
	}

	if fileExists(filepath.Join(projectPath, ".eslintrc.js")) || fileExists(filepath.Join(projectPath, ".eslintrc.json")) ||
		fileExists(filepath.Join(projectPath, ".eslintrc.yml")) || fileContains(projectPath, "eslint", "package.json") {
		addTool(&tools, "ESLint")
	}
	if fileExists(filepath.Join(projectPath, ".prettierrc")) || fileExists(filepath.Join(projectPath, ".prettierrc.json")) ||
		fileExists(filepath.Join(projectPath, ".prettierrc.js")) || fileContains(projectPath, "prettier", "package.json") {
		addTool(&tools, "Prettier")
	}
	if fileExists(filepath.Join(projectPath, "pylintrc")) || fileExists(filepath.Join(projectPath, ".flake8")) ||
		fileContains(projectPath, "flake8", "requirements.txt") {
		addTool(&tools, "Python Linters")
	}
	if fileExists(filepath.Join(projectPath, ".golangci.yml")) || fileExists(filepath.Join(projectPath, "golangci.yml")) {
		addTool(&tools, "golangci-lint")
	}
	if fileExists(filepath.Join(projectPath, "rustfmt.toml")) || fileExists(filepath.Join(projectPath, ".rustfmt.toml")) {
		addTool(&tools, "rustfmt")
	}
	if fileExists(filepath.Join(projectPath, "clippy.toml")) || fileExists(filepath.Join(projectPath, ".clippy.toml")) {
		addTool(&tools, "Clippy")
	}

	if fileExists(filepath.Join(projectPath, "Dockerfile")) || fileExists(filepath.Join(projectPath, "docker-compose.yml")) ||
		fileExists(filepath.Join(projectPath, "docker-compose.yaml")) {
		addTool(&tools, "Docker")
	}
	if dirExists(filepath.Join(projectPath, "kubernetes")) || (anyFileMatches(projectPath, "*.yaml") && fileContains(projectPath, "apiVersion:", "*.yaml")) {
		addTool(&tools, "Kubernetes")
	}

	if len(tools) == 0 {
		tools = []string{"Manual"}
	}
	return tools
}

// detectBuildSystems mirrors detect_build_systems().
func detectBuildSystems(projectPath string) []string {
	var systems []string
	add := func(name string) { addTool(&systems, name) }

	if fileExists(filepath.Join(projectPath, "package.json")) {
		if fileContains(projectPath, "webpack", "package.json") || fileExists(filepath.Join(projectPath, "webpack.config.js")) || fileExists(filepath.Join(projectPath, "webpack.config.ts")) {
			add("Webpack")
		}
		if fileContains(projectPath, "vite", "package.json") || fileExists(filepath.Join(projectPath, "vite.config.js")) || fileExists(filepath.Join(projectPath, "vite.config.ts")) {
			add("Vite")
		}
		if fileContains(projectPath, "rollup", "package.json") || fileExists(filepath.Join(projectPath, "rollup.config.js")) || fileExists(filepath.Join(projectPath, "rollup.config.ts")) {
			add("Rollup")
		}
		if fileContains(projectPath, "parcel", "package.json") || fileExists(filepath.Join(projectPath, "parcelrc")) || fileExists(filepath.Join(projectPath, ".parcelrc")) {
			add("Parcel")
		}
		if fileContains(projectPath, "esbuild", "package.json") || fileExists(filepath.Join(projectPath, "esbuild.js")) || fileExists(filepath.Join(projectPath, "esbuild.config.js")) {
			add("esbuild")
		}
		if fileContains(projectPath, "turbo", "package.json") || fileExists(filepath.Join(projectPath, "turbo.json")) {
			add("Turbopack")
		}
		if fileContains(projectPath, "next", "package.json") {
			add("Next.js")
		}
		if fileExists(filepath.Join(projectPath, "public", "index.html")) && fileContains(projectPath, "react-scripts", "package.json") {
			add("Create React App")
		}
	}

	if fileExists(filepath.Join(projectPath, "pyproject.toml")) {
		if fileContains(projectPath, "setuptools", "pyproject.toml") {
			add("setuptools")
		}
		if fileContains(projectPath, "poetry", "pyproject.toml") {
			add("Poetry")
		}
		if fileContains(projectPath, "hatch", "pyproject.toml") {
			add("Hatch")
		}
		if fileContains(projectPath, "pdm", "pyproject.toml") {
			add("PDM")
		}
	}
	if fileExists(filepath.Join(projectPath, "setup.py")) {
		add("setuptools")
	}
	if fileExists(filepath.Join(projectPath, "Makefile")) && fileContains(projectPath, "python", "Makefile") {
		add("Make")
	}

	if fileExists(filepath.Join(projectPath, "pom.xml")) {
		add("Maven")
	}
	if fileExists(filepath.Join(projectPath, "build.gradle")) || fileExists(filepath.Join(projectPath, "build.gradle.kts")) {
		add("Gradle")
	}
	if fileExists(filepath.Join(projectPath, "build.xml")) {
		add("Ant")
	}

	if fileExists(filepath.Join(projectPath, "go.mod")) {
		add("Go Modules")
	}
	if fileExists(filepath.Join(projectPath, "Makefile")) && fileContains(projectPath, "go", "Makefile") {
		add("Make")
	}

	if fileExists(filepath.Join(projectPath, "Cargo.toml")) {
		add("Cargo")
	}

	if fileExists(filepath.Join(projectPath, "CMakeLists.txt")) {
		add("CMake")
	}
	if fileExists(filepath.Join(projectPath, "Makefile")) {
		add("Make")
	}
	if fileExists(filepath.Join(projectPath, "meson.build")) {
		add("Meson")
	}
	if fileExists(filepath.Join(projectPath, "configure.ac")) || fileExists(filepath.Join(projectPath, "Makefile.am")) {
		add("Autotools")
	}

	if fileExists(filepath.Join(projectPath, "Gemfile")) {
		add("Bundler")
	}
	if fileExists(filepath.Join(projectPath, "Rakefile")) {
		add("Rake")
	}

	if fileExists(filepath.Join(projectPath, "composer.json")) {
		add("Composer")
	}

	if fileExists(filepath.Join(projectPath, "pubspec.yaml")) {
		add("Pub")
	}
	if dirExists(filepath.Join(projectPath, "android")) && fileExists(filepath.Join(projectPath, "android", "build.gradle")) {
		add("Gradle (Android)")
	}
	if dirExists(filepath.Join(projectPath, "ios")) && fileExists(filepath.Join(projectPath, "ios", "Runner.xcodeproj", "project.pbxproj")) {
		add("Xcode")
	}

	if fileExists(filepath.Join(projectPath, "Dockerfile")) {
		add("Docker")
	}
	if fileExists(filepath.Join(projectPath, "docker-compose.yml")) || fileExists(filepath.Join(projectPath, "docker-compose.yaml")) {
		add("Docker Compose")
	}
	if fileExists(filepath.Join(projectPath, "Terrafile")) || fileExists(filepath.Join(projectPath, "main.tf")) {
		add("Terraform")
	}

	if len(systems) == 0 {
		systems = []string{"Manual"}
	}
	return systems
}
