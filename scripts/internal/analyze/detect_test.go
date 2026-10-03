package analyze

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestFileExistsAndDirExists(t *testing.T) {
	root := newFixture(t, map[string]string{
		"file.txt":     "hi",
		"dir/inner.md": "# hi",
	})
	cases := []struct {
		name          string
		path          string
		wantFileExist bool
		wantDirExist  bool
	}{
		{"regular file", "file.txt", true, false},
		{"directory", "dir", false, true},
		{"nested file", filepath.Join("dir", "inner.md"), true, false},
		{"missing", "nope.txt", false, false},
		{"dot directory", ".git", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := filepath.Join(root, filepath.FromSlash(tc.path))
			if got := fileExists(p); got != tc.wantFileExist {
				t.Errorf("fileExists(%q) = %v, want %v", tc.path, got, tc.wantFileExist)
			}
			if got := dirExists(p); got != tc.wantDirExist {
				t.Errorf("dirExists(%q) = %v, want %v", tc.path, got, tc.wantDirExist)
			}
		})
	}
}

func TestAnyFileMatches(t *testing.T) {
	root := newFixture(t, map[string]string{
		"app.csproj":       "<Project/>",
		"sub/other.csproj": "<Project/>",
		"readme.md":        "# hi",
	})
	cases := []struct {
		name    string
		pattern string
		want    bool
	}{
		{"root glob hit", "*.csproj", true},
		{"single-level glob reaches into a subdirectory", "*/*.csproj", true},
		{"glob is not recursive beyond one level", "*/*/*.csproj", false},
		{"literal name", "readme.md", true},
		{"no match", "*.tf", false},
		{"metacharacters not honoured in literal lookup", "readme[.md", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := anyFileMatches(root, tc.pattern); got != tc.want {
				t.Errorf("anyFileMatches(%q) = %v, want %v", tc.pattern, got, tc.want)
			}
		})
	}
}

func TestFileContains(t *testing.T) {
	root := newFixture(t, map[string]string{
		"package.json":  `{"dependencies":{"react":"18"}}`,
		"nested/a.yaml": "apiVersion: apps/v1",
		"empty.txt":     "",
	})
	cases := []struct {
		name     string
		substr   string
		patterns []string
		want     bool
	}{
		{"literal pattern hit", "react", []string{"package.json"}, true},
		{"literal pattern miss", "vue", []string{"package.json"}, false},
		{"missing file tolerated", "anything", []string{"absent.json"}, false},
		{"root glob does not reach into subdirectories", "apiVersion", []string{"*.yaml"}, false},
		{"nested glob pattern hit", "apiVersion", []string{"nested/*.yaml"}, true},
		{"multiple patterns any hit", "react", []string{"absent.json", "package.json"}, true},
		{"all patterns miss", "react", []string{"absent.json", "nested/a.yaml"}, false},
		{"empty file does not match a non-empty needle", "zzz", []string{"empty.txt"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := fileContains(root, tc.substr, tc.patterns...); got != tc.want {
				t.Errorf("fileContains(%q, %v) = %v, want %v", tc.substr, tc.patterns, got, tc.want)
			}
		})
	}
}

func TestHas(t *testing.T) {
	hits := []FrameworkHit{{"react", "high"}, {"docker", "medium"}}
	cases := []struct {
		name string
		hits []FrameworkHit
		want bool
	}{
		{"present", hits, true},
		{"present with different confidence still matches", []FrameworkHit{{"react", "low"}}, true},
		{"absent", []FrameworkHit{{"vue", "high"}}, false},
		{"empty list", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := has(tc.hits, "react"); got != tc.want {
				t.Errorf("has(%v, \"react\") = %v, want %v", tc.hits, got, tc.want)
			}
		})
	}
}

func TestDetectAllFrameworks(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  []FrameworkHit
	}{
		{
			name:  "empty project detects nothing",
			files: map[string]string{"README.md": "plain readme"},
			want:  nil,
		},
		{
			name:  "package.json with next dependency",
			files: map[string]string{"package.json": `{"dependencies":{"next":"14.0.0"}}`},
			want:  []FrameworkHit{{"nextjs", "high"}},
		},
		{
			name:  "package.json with react dependency",
			files: map[string]string{"package.json": `{"dependencies":{"react":"18.2.0"}}`},
			want:  []FrameworkHit{{"react", "high"}},
		},
		{
			name:  "package.json without any known framework falls back to nodejs medium",
			files: map[string]string{"package.json": `{"name":"plain"}`},
			want:  []FrameworkHit{{"nodejs", "medium"}},
		},
		{
			name: "package.json plus next.config.js deduplicates nextjs",
			files: map[string]string{
				"package.json":   `{"dependencies":{"next":"14.0.0"}}`,
				"next.config.js": "module.exports = {}",
			},
			want: []FrameworkHit{{"nextjs", "high"}},
		},
		{
			name:  "next.config.mjs alone still detects nextjs",
			files: map[string]string{"next.config.mjs": "export default {}"},
			want:  []FrameworkHit{{"nextjs", "high"}},
		},
		{
			name:  "nuxt.config.ts",
			files: map[string]string{"nuxt.config.ts": "export default {}"},
			want:  []FrameworkHit{{"nuxt", "high"}},
		},
		{
			name:  "svelte.config.js",
			files: map[string]string{"svelte.config.js": "export default {}"},
			want:  []FrameworkHit{{"svelte", "high"}},
		},
		{
			name:  "astro.config.ts",
			files: map[string]string{"astro.config.ts": "export default {}"},
			want:  []FrameworkHit{{"astro", "high"}},
		},
		{
			name:  "remix.config.js",
			files: map[string]string{"remix.config.js": "module.exports = {}"},
			want:  []FrameworkHit{{"remix", "high"}},
		},
		{
			name:  "gatsby-config.js",
			files: map[string]string{"gatsby-config.js": "module.exports = {}"},
			want:  []FrameworkHit{{"gatsby", "high"}},
		},
		{
			name:  "vite config alone yields medium confidence",
			files: map[string]string{"vite.config.ts": "export default {}"},
			want:  []FrameworkHit{{"vite", "medium"}},
		},
		{
			name: "requirements.txt with flask",
			files: map[string]string{
				"requirements.txt": "flask==3.0.0\nrequests==2.31.0\n",
			},
			want: []FrameworkHit{{"flask", "high"}},
		},
		{
			name:  "requirements.txt with only third party deps falls back to python",
			files: map[string]string{"requirements.txt": "requests==2.31.0\n"},
			want:  []FrameworkHit{{"python", "medium"}},
		},
		{
			name:  "pyproject.toml with django, fastapi and poetry",
			files: map[string]string{"pyproject.toml": "[tool.poetry.dependencies]\ndjango = \"5\"\nfastapi = \"0.1\"\n"},
			want:  []FrameworkHit{{"django", "high"}, {"fastapi", "high"}, {"poetry", "medium"}},
		},
		{
			name:  "pyproject.toml with neither framework falls back to python",
			files: map[string]string{"pyproject.toml": "[project]\nname = \"x\"\n"},
			want:  []FrameworkHit{{"python", "medium"}},
		},
		{
			name:  "requirements.txt with django and fastapi",
			files: map[string]string{"requirements.txt": "django==5.0\nfastapi==0.110\n"},
			want:  []FrameworkHit{{"django", "high"}, {"fastapi", "high"}},
		},
		{
			// Needle matching is case sensitive, so a capitalised requirement
			// name is not recognised.
			name:  "capitalised requirement names are missed",
			files: map[string]string{"requirements.txt": "Django==5.0\n"},
			want:  []FrameworkHit{{"python", "medium"}},
		},
		{
			name:  "pyproject.toml with poetry suppresses the generic python hit",
			files: map[string]string{"pyproject.toml": "[tool.poetry]\nname = \"x\"\n"},
			want:  []FrameworkHit{{"poetry", "medium"}},
		},
		{
			name:  "manage.py implies django",
			files: map[string]string{"manage.py": "import django"},
			want:  []FrameworkHit{{"django", "high"}},
		},
		{
			name:  "Pipfile implies pipenv",
			files: map[string]string{"Pipfile": "[packages]"},
			want:  []FrameworkHit{{"pipenv", "high"}},
		},
		{
			name:  "pubspec.yaml implies flutter",
			files: map[string]string{"pubspec.yaml": "name: app"},
			want:  []FrameworkHit{{"flutter", "high"}},
		},
		{
			name:  "go.mod with gin",
			files: map[string]string{"go.mod": "module x\nrequire github.com/gin-gonic/gin v1.9.1\n"},
			want:  []FrameworkHit{{"gin", "high"}},
		},
		{
			name:  "go.mod with echo and fiber",
			files: map[string]string{"go.mod": "module x\nrequire github.com/labstack/echo/v4 v4.11\nrequire github.com/gofiber/fiber/v2 v2.52\n"},
			want:  []FrameworkHit{{"echo", "high"}, {"fiber", "high"}},
		},
		{
			name:  "go.mod without web framework falls back to go",
			files: map[string]string{"go.mod": "module x\ngo 1.22\n"},
			want:  []FrameworkHit{{"go", "medium"}},
		},
		{
			name:  "Cargo.toml with actix-web",
			files: map[string]string{"Cargo.toml": "[dependencies]\nactix-web = \"4\"\n"},
			want:  []FrameworkHit{{"actix", "high"}},
		},
		{
			name:  "Cargo.toml with rocket",
			files: map[string]string{"Cargo.toml": "[dependencies]\nrocket = \"0.5\"\n"},
			want:  []FrameworkHit{{"rocket", "high"}},
		},
		{
			name:  "Cargo.toml with axum",
			files: map[string]string{"Cargo.toml": "[dependencies]\naxum = \"0.7\"\n"},
			want:  []FrameworkHit{{"axum", "high"}},
		},
		{
			name:  "Cargo.toml without web framework falls back to rust",
			files: map[string]string{"Cargo.toml": "[package]\nname = \"x\"\n"},
			want:  []FrameworkHit{{"rust", "medium"}},
		},
		{
			name:  "pom.xml mentioning spring-boot also matches the spring substring",
			files: map[string]string{"pom.xml": "<dependency>spring-boot-starter-web</dependency>"},
			want:  []FrameworkHit{{"springboot", "high"}, {"spring", "high"}},
		},
		{
			name:  "pom.xml without spring falls back to maven",
			files: map[string]string{"pom.xml": "<project>junit</project>"},
			want:  []FrameworkHit{{"maven", "medium"}},
		},
		{
			name:  "build.gradle with the spring boot plugin",
			files: map[string]string{"build.gradle": `id "org.springframework.boot" version "3.2.0"`},
			want:  []FrameworkHit{{"springboot", "high"}},
		},
		{
			name:  "build.gradle without a known plugin falls back to gradle",
			files: map[string]string{"build.gradle": "apply plugin: 'java'"},
			want:  []FrameworkHit{{"gradle", "medium"}},
		},
		{
			name:  "build.gradle.kts with io.ktor",
			files: map[string]string{"build.gradle.kts": `implementation("io.ktor:ktor-server")`},
			want:  []FrameworkHit{{"ktor", "high"}},
		},
		{
			name:  "Gemfile with sinatra",
			files: map[string]string{"Gemfile": "source 'x'\ngem 'sinatra'"},
			want:  []FrameworkHit{{"sinatra", "high"}},
		},
		{
			name:  "Gemfile without a web gem falls back to ruby",
			files: map[string]string{"Gemfile": "source 'x'\ngem 'rake'"},
			want:  []FrameworkHit{{"ruby", "medium"}},
		},
		{
			name:  "config/application.rb implies rails",
			files: map[string]string{"config/application.rb": "module App"},
			want:  []FrameworkHit{{"rails", "high"}},
		},
		{
			name:  "Gemfile with rails",
			files: map[string]string{"Gemfile": "source 'https://rubygems.org'\ngem 'rails'"},
			want:  []FrameworkHit{{"rails", "high"}},
		},
		{
			name:  "composer.json with laravel/framework",
			files: map[string]string{"composer.json": `{"require":{"laravel/framework":"^10"}}`},
			want:  []FrameworkHit{{"laravel", "high"}},
		},
		{
			name:  "composer.json with symfony",
			files: map[string]string{"composer.json": `{"require":{"symfony/framework-bundle":"^6"}}`},
			want:  []FrameworkHit{{"symfony", "high"}},
		},
		{
			name:  "composer.json without a framework falls back to php",
			files: map[string]string{"composer.json": `{"require":{"monolog/monolog":"^3"}}`},
			want:  []FrameworkHit{{"php", "medium"}},
		},
		{
			name:  "wp-config.php implies wordpress",
			files: map[string]string{"wp-config.php": "<?php // config"},
			want:  []FrameworkHit{{"wordpress", "high"}},
		},
		{
			name:  "csproj with aspnet",
			files: map[string]string{"app.csproj": `<PackageReference Include="Microsoft.AspNetCore.App" />`},
			want:  []FrameworkHit{{"aspnet", "high"}},
		},
		{
			name:  "csproj without aspnet falls back to dotnet",
			files: map[string]string{"app.csproj": "<Project Sdk=\"Microsoft.NET.Sdk\" />"},
			want:  []FrameworkHit{{"dotnet", "medium"}},
		},
		{
			name:  "Package.swift implies swift",
			files: map[string]string{"Package.swift": "// swift-tools-version:5.9"},
			want:  []FrameworkHit{{"swift", "high"}},
		},
		{
			name: "container and iac files are detected in source order",
			files: map[string]string{
				"Dockerfile":         "FROM alpine",
				"docker-compose.yml": "services: {}",
				"main.tf":            `resource "null_resource" "x" {}`,
				"Vagrantfile":        "Vagrant.configure",
				"kustomization.yaml": "resources: []",
			},
			want: []FrameworkHit{
				{"docker", "medium"},
				{"docker-compose", "medium"},
				{"terraform", "high"},
				{"vagrant", "high"},
				{"kustomize", "high"},
			},
		},
		{
			name:  "README mentioning Plaesy Spec-Kit",
			files: map[string]string{"README.md": "# Plaesy Spec-Kit\n"},
			want:  []FrameworkHit{{"spec-kit", "high"}},
		},
		{
			name:  "README without the marker does not detect spec-kit",
			files: map[string]string{"README.md": "# Some Project\n"},
			want:  nil,
		},
		{
			name: "multiple ecosystems are all reported in evaluation order",
			files: map[string]string{
				"package.json": `{"dependencies":{"next":"14","react":"18"}}`,
				"go.mod":       "module x\nrequire github.com/gin-gonic/gin v1.9.1",
			},
			want: []FrameworkHit{{"nextjs", "high"}, {"react", "high"}, {"gin", "high"}},
		},
		{
			name: "substrings match anywhere in package.json (react-scripts reads as react)",
			files: map[string]string{
				"package.json": `{"devDependencies":{"react-scripts":"5.0.1"}}`,
			},
			want: []FrameworkHit{{"react", "high"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newFixture(t, tc.files)
			got := detectAllFrameworks(root)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("detectAllFrameworks()\n got: %+v\nwant: %+v", got, tc.want)
			}
		})
	}
}

func TestDetectProjectType(t *testing.T) {
	cases := []struct {
		name string
		hits []FrameworkHit
		want FrameworkHit
	}{
		{"no hits yields generic low", nil, FrameworkHit{"generic", "low"}},
		{"single hit is returned verbatim", []FrameworkHit{{"go", "medium"}}, FrameworkHit{"go", "medium"}},
		{
			name: "priority list beats evaluation order",
			hits: []FrameworkHit{{"docker", "medium"}, {"django", "high"}, {"react", "high"}},
			want: FrameworkHit{"react", "high"},
		},
		{
			name: "backend priority beats generic language hits",
			hits: []FrameworkHit{{"python", "medium"}, {"fastapi", "high"}},
			want: FrameworkHit{"fastapi", "high"},
		},
		{
			name: "falls back to the first hit when nothing is prioritised",
			hits: []FrameworkHit{{"spec-kit", "high"}, {"kustomize", "high"}},
			want: FrameworkHit{"spec-kit", "high"},
		},
		{
			name: "confidence travels with the winning framework",
			hits: []FrameworkHit{{"docker", "medium"}, {"vue", "high"}},
			want: FrameworkHit{"vue", "high"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := detectProjectType(tc.hits); got != tc.want {
				t.Errorf("detectProjectType(%+v) = %+v, want %+v", tc.hits, got, tc.want)
			}
		})
	}
}

func TestPriorityFrameworksCoverage(t *testing.T) {
	// Every framework name that the generator can emit as a project type must
	// either be in the priority list or intentionally absent (tooling-only
	// names like docker/terraform). This locks the priority list shape.
	for _, name := range []string{"nextjs", "react", "vue", "django", "flutter", "go", "rust", "swift"} {
		if !hasString(priorityFrameworks, name) {
			t.Errorf("priorityFrameworks is missing %q", name)
		}
	}
	if len(priorityFrameworks) == 0 {
		t.Fatal("priorityFrameworks must not be empty")
	}
	seen := map[string]bool{}
	for _, p := range priorityFrameworks {
		if seen[p] {
			t.Errorf("duplicate entry %q in priorityFrameworks", p)
		}
		seen[p] = true
	}
}

func TestDetectAllLanguages(t *testing.T) {
	cases := []struct {
		name  string
		files []string
		want  []string
	}{
		{"no recognisable extension defaults to JavaScript", []string{"Makefile", "LICENSE"}, []string{"JavaScript"}},
		{"empty input defaults to JavaScript", nil, []string{"JavaScript"}},
		{"single go file", []string{"cmd/main.go"}, []string{"Go"}},
		{"python", []string{"app.py"}, []string{"Python"}},
		{"dart", []string{"lib/main.dart"}, []string{"Dart"}},
		{"java and kotlin", []string{"A.java", "B.kt"}, []string{"Java", "Kotlin"}},
		{"c, c++ and headers are separate languages in canonical order", []string{"a.c", "b.cpp", "c.h"}, []string{"C++", "C", "C/C++ Headers"}},
		{"css family counts once", []string{"a.css", "b.scss", "c.sass"}, []string{"CSS"}},
		{"canonical order is independent of input order", []string{"z.sh", "a.py", "m.go"}, []string{"Python", "Go", "Shell"}},
		{"tsx and tsx-adjacent extensions", []string{"a.ts", "b.tsx", "c.js", "d.jsx"}, []string{"JavaScript", "TypeScript"}},
		{"case sensitive extension matching", []string{"Main.GO"}, []string{"JavaScript"}},
		{"directory names are ignored, only the base name counts", []string{"go/notes.txt"}, []string{"JavaScript"}},
		{"html and markdown-adjacent types", []string{"index.html", "style.css"}, []string{"HTML", "CSS"}},
		{"rb, rs, swift, scala, php", []string{"a.rb", "b.rs", "c.swift", "d.scala", "e.php"},
			[]string{"PHP", "Ruby", "Rust", "Swift", "Scala"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := detectAllLanguages(tc.files); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("detectAllLanguages(%v) = %v, want %v", tc.files, got, tc.want)
			}
		})
	}
}

func TestAddTool(t *testing.T) {
	cases := []struct {
		name string
		seed []string
		add  string
		want []string
	}{
		{"appends to nil", nil, "Git", []string{"Git"}},
		{"appends new", []string{"Git"}, "Docker", []string{"Git", "Docker"}},
		{"duplicate ignored", []string{"Git", "Docker"}, "Git", []string{"Git", "Docker"}},
		{"duplicate of only entry ignored", []string{"Git"}, "Git", []string{"Git"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := append([]string(nil), tc.seed...)
			addTool(&got, tc.add)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("addTool(%v, %q) = %v, want %v", tc.seed, tc.add, got, tc.want)
			}
		})
	}
}

func TestDetectDevelopmentTools(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{"empty project falls back to Manual", map[string]string{}, []string{"Manual"}},
		{"git directory", map[string]string{".git/": ""}, []string{"Git"}},
		{"subversion directory", map[string]string{".svn/": ""}, []string{"Subversion"}},
		{"mercurial is detected from a .hg FILE not a directory", map[string]string{".hg": "hg"}, []string{"Mercurial"}},
		{"npm ecosystem from package-lock", map[string]string{"package-lock.json": "{}"}, []string{"npm/yarn/pnpm"}},
		{"python tooling from Pipfile", map[string]string{"Pipfile": "[packages]"}, []string{"pip/poetry"}},
		{"go modules and go test", map[string]string{
			"go.mod": "module x\nrequire github.com/stretchr/testify v1.8.0",
		}, []string{"Go Modules", "Go test"}},
		{"cargo", map[string]string{"Cargo.toml": "[package]"}, []string{"Cargo"}},
		{"maven and gradle", map[string]string{"pom.xml": "<project/>"}, []string{"Maven/Gradle"}},
		{"bundler", map[string]string{"Gemfile": "source 'x'"}, []string{"Bundler"}},
		{"composer", map[string]string{"composer.json": "{}"}, []string{"Composer"}},
		{"pub", map[string]string{"pubspec.yaml": "name: a"}, []string{"Pub"}},
		{"github actions from workflows directory", map[string]string{".github/workflows/": ""}, []string{"GitHub Actions"}},
		{"gitlab ci", map[string]string{".gitlab-ci.yml": "stages: []"}, []string{"GitLab CI"}},
		{"jenkins", map[string]string{"Jenkinsfile": "pipeline {}"}, []string{"Jenkins"}},
		{"azure pipelines", map[string]string{"azure-pipelines.yml": "trigger: []"}, []string{"Azure Pipelines"}},
		{"jest from package.json dependency", map[string]string{
			"package.json": `{"devDependencies":{"jest":"29"}}`,
		}, []string{"npm/yarn/pnpm", "Jest"}},
		{"vitest from config file", map[string]string{"vitest.config.ts": "export default {}"}, []string{"Vitest"}},
		{"pytest from a tests directory", map[string]string{"tests/": ""}, []string{"pytest"}},
		{"pytest from pytest.ini", map[string]string{"pytest.ini": "[pytest]"}, []string{"pytest"}},
		{"golangci-lint from .golangci.yml", map[string]string{".golangci.yml": "linters: []"}, []string{"golangci-lint"}},
		{"rustfmt and clippy", map[string]string{
			"rustfmt.toml": "edition = \"2021\"",
			"clippy.toml":  "too_many_arguments_threshold = 7",
		}, []string{"rustfmt", "Clippy"}},
		{"eslint and prettier from package.json", map[string]string{
			"package.json": `{"devDependencies":{"eslint":"8","prettier":"3"}}`,
		}, []string{"npm/yarn/pnpm", "ESLint", "Prettier"}},
		{"python linters from flake8 requirement", map[string]string{
			"requirements.txt": "flake8==6.0.0",
		}, []string{"pip/poetry", "Python Linters"}},
		{"docker", map[string]string{"Dockerfile": "FROM alpine"}, []string{"Docker"}},
		{"kubernetes from a kubernetes directory", map[string]string{"kubernetes/": ""}, []string{"Kubernetes"}},
		{"kubernetes sniffed from apiVersion in a root yaml", map[string]string{
			"deploy.yaml": "apiVersion: apps/v1\nkind: Deployment",
		}, []string{"Kubernetes"}},
		{"combined project reports every tool in detection order", map[string]string{
			".git/":              "",
			"package.json":       `{"devDependencies":{"jest":"29","eslint":"8"}}`,
			"go.mod":             "module x",
			".github/workflows/": "",
		}, []string{"Git", "npm/yarn/pnpm", "Go Modules", "GitHub Actions", "Jest", "ESLint"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newFixture(t, tc.files)
			if got := detectDevelopmentTools(root); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("detectDevelopmentTools()\n got: %v\nwant: %v", got, tc.want)
			}
		})
	}
}

func TestDetectBuildSystems(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{"empty project falls back to Manual", map[string]string{}, []string{"Manual"}},
		{"webpack vite and next from package.json", map[string]string{
			"package.json": `{"devDependencies":{"webpack":"5","vite":"4","next":"14"}}`,
		}, []string{"Webpack", "Vite", "Next.js"}},
		{"rollup parcel esbuild and turbopack from package.json", map[string]string{
			"package.json": `{"devDependencies":{"rollup":"4","parcel":"2","esbuild":"0.20","turbo":"2"}}`,
		}, []string{"Rollup", "Parcel", "esbuild", "Turbopack"}},
		{"parcel via .parcelrc", map[string]string{
			"package.json": `{"name":"x"}`,
			".parcelrc":    "{}",
		}, []string{"Parcel"}},
		{"create react app needs public/index.html too", map[string]string{
			"package.json":      `{"dependencies":{"react-scripts":"5.0.1"}}`,
			"public/index.html": "<html></html>",
		}, []string{"Create React App"}},
		{"create react app without the html marker falls back to Manual", map[string]string{
			"package.json": `{"dependencies":{"react-scripts":"5.0.1"}}`,
		}, []string{"Manual"}},
		{"webpack config file alone (the whole block needs package.json) falls back to Manual", map[string]string{
			"webpack.config.js": "module.exports = {}",
		}, []string{"Manual"}},
		{"pyproject build backends are detected by substring", map[string]string{
			"pyproject.toml": "[build-system]\nrequires = [\"poetry-core\", \"hatchling\"]\n[tool.pdm]",
		}, []string{"Poetry", "Hatch", "PDM"}},
		{"setup.py", map[string]string{"setup.py": "from setuptools import setup"}, []string{"setuptools"}},
		{"pyproject setuptools backend", map[string]string{
			"pyproject.toml": "[build-system]\nrequires = [\"setuptools>=68\"]",
		}, []string{"setuptools"}},
		{"Makefile mentioning python", map[string]string{
			"setup.py": "from setuptools import setup",
			"Makefile": "build:\n\tpython setup.py sdist",
		}, []string{"setuptools", "Make"}},
		{"go makefile detection", map[string]string{
			"go.mod":   "module x",
			"Makefile": "build:\n\tgo build ./...",
		}, []string{"Go Modules", "Make"}},
		{"a generic Makefile alone still counts as Make", map[string]string{"Makefile": "all:\n\techo hi"},
			[]string{"Make"}},
		{"Make with go.mod is not duplicated", map[string]string{
			"go.mod":   "module x",
			"Makefile": "all:\n\tgo test ./...",
		}, []string{"Go Modules", "Make"}},
		{"maven", map[string]string{"pom.xml": "<project/>"}, []string{"Maven"}},
		{"gradle kts", map[string]string{"build.gradle.kts": "plugins {}"}, []string{"Gradle"}},
		{"ant", map[string]string{"build.xml": "<project/>"}, []string{"Ant"}},
		{"cmake and meson", map[string]string{
			"CMakeLists.txt": "project(x)",
			"meson.build":    "project('x')",
		}, []string{"CMake", "Meson"}},
		{"autotools", map[string]string{"configure.ac": "AC_INIT"}, []string{"Autotools"}},
		{"cargo", map[string]string{"Cargo.toml": "[package]"}, []string{"Cargo"}},
		{"rake", map[string]string{"Rakefile": "task :default"}, []string{"Rake"}},
		{"bundler from Gemfile", map[string]string{"Gemfile": "source 'x'"}, []string{"Bundler"}},
		{"flutter android gradle", map[string]string{
			"pubspec.yaml":         "name: a",
			"android/build.gradle": "plugins {}",
		}, []string{"Pub", "Gradle (Android)"}},
		{"flutter ios xcode", map[string]string{
			"pubspec.yaml":                         "name: a",
			"ios/Runner.xcodeproj/project.pbxproj": "// pbx",
		}, []string{"Pub", "Xcode"}},
		{"ios directory without the pbxproj file is not Xcode", map[string]string{
			"pubspec.yaml": "name: a",
			"ios/":         "",
		}, []string{"Pub"}},
		{"docker stack", map[string]string{
			"Dockerfile":         "FROM alpine",
			"docker-compose.yml": "services: {}",
		}, []string{"Docker", "Docker Compose"}},
		{"terraform", map[string]string{"Terrafile": ""}, []string{"Terraform"}},
		{"composer", map[string]string{"composer.json": "{}"}, []string{"Composer"}},
		{"a package.json with no known bundler yields Manual", map[string]string{
			"package.json": `{"name":"x"}`,
		}, []string{"Manual"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newFixture(t, tc.files)
			got := detectBuildSystems(root)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("detectBuildSystems()\n got: %v\nwant: %v", got, tc.want)
			}
		})
	}
}
