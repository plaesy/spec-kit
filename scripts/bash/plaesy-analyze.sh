#!/bin/bash

# Plaesy AI-Optimized Project Analyzer - Enhanced Version
# Comprehensive project analysis with AI-friendly documentation generation
# Usage: ./plaesy-analyze.sh [project_path] [--no-graph] [--force]
#   --no-graph  skip dependency graph build entirely
#   --force     force a full graph rebuild even if no source files changed
#               (default: graph rebuild is skipped when nothing changed since
#               the last run)

set -euo pipefail

# Colors for output
readonly GREEN='\033[0;32m'
readonly BLUE='\033[0;34m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m'

# Global variables
BUILD_GRAPH=1
FORCE_GRAPH=0
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_PATH=""
for arg in "$@"; do
    case "$arg" in
        --no-graph) BUILD_GRAPH=0 ;;
        --force) FORCE_GRAPH=1 ;;
        *) [[ -z "$PROJECT_PATH" ]] && PROJECT_PATH="$arg" ;;
    esac
done
PROJECT_PATH="${PROJECT_PATH:-$(pwd)}"
ANALYSIS_DIR="$PROJECT_PATH/.plaesy/memory/analysis"
MEMORY_DIR="$PROJECT_PATH/.plaesy/memory"
SCRIPTS_DIR="$PROJECT_PATH/scripts"
FRAMEWORK_VERSION="$(cat "$SCRIPT_DIR/../../VERSION" 2>/dev/null || echo "0.0.1")"

# Ensure directories exist
mkdir -p "$ANALYSIS_DIR" "$MEMORY_DIR"
mkdir -p "$SCRIPTS_DIR"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Function to detect all frameworks in the project
detect_all_frameworks() {
    local frameworks=()
    local confidences=()

    # JavaScript/TypeScript frameworks
    if [[ -f "$PROJECT_PATH/package.json" ]]; then
        if grep -q "next" "$PROJECT_PATH/package.json"; then
            frameworks+=("nextjs")
            confidences+=("high")
        fi
        if grep -q "react" "$PROJECT_PATH/package.json"; then
            frameworks+=("react")
            confidences+=("high")
        fi
        if grep -q "vue" "$PROJECT_PATH/package.json"; then
            frameworks+=("vue")
            confidences+=("high")
        fi
        if grep -q "svelte" "$PROJECT_PATH/package.json"; then
            frameworks+=("svelte")
            confidences+=("high")
        fi
        if grep -q "astro" "$PROJECT_PATH/package.json"; then
            frameworks+=("astro")
            confidences+=("high")
        fi
        if grep -q "remix" "$PROJECT_PATH/package.json"; then
            frameworks+=("remix")
            confidences+=("high")
        fi
        if grep -q "gatsby" "$PROJECT_PATH/package.json"; then
            frameworks+=("gatsby")
            confidences+=("high")
        fi
        if grep -q "nuxt" "$PROJECT_PATH/package.json"; then
            frameworks+=("nuxt")
            confidences+=("high")
        fi
        if grep -q "express" "$PROJECT_PATH/package.json"; then
            frameworks+=("express")
            confidences+=("high")
        fi
        if grep -q "nestjs" "$PROJECT_PATH/package.json"; then
            frameworks+=("nestjs")
            confidences+=("high")
        fi
        if [[ ${#frameworks[@]} -eq 0 ]]; then
            frameworks+=("nodejs")
            confidences+=("medium")
        fi
    fi

    # Check for additional JS/TS config files
    if [[ -f "$PROJECT_PATH/next.config.js" ]] || [[ -f "$PROJECT_PATH/next.config.mjs" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " nextjs " ]]; then
            frameworks+=("nextjs")
            confidences+=("high")
        fi
    fi
    if [[ -f "$PROJECT_PATH/nuxt.config.js" ]] || [[ -f "$PROJECT_PATH/nuxt.config.ts" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " nuxt " ]]; then
            frameworks+=("nuxt")
            confidences+=("high")
        fi
    fi
    if [[ -f "$PROJECT_PATH/svelte.config.js" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " svelte " ]]; then
            frameworks+=("svelte")
            confidences+=("high")
        fi
    fi
    if [[ -f "$PROJECT_PATH/astro.config.mjs" ]] || [[ -f "$PROJECT_PATH/astro.config.ts" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " astro " ]]; then
            frameworks+=("astro")
            confidences+=("high")
        fi
    fi
    if [[ -f "$PROJECT_PATH/remix.config.js" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " remix " ]]; then
            frameworks+=("remix")
            confidences+=("high")
        fi
    fi
    if [[ -f "$PROJECT_PATH/gatsby-config.js" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " gatsby " ]]; then
            frameworks+=("gatsby")
            confidences+=("high")
        fi
    fi
    if [[ -f "$PROJECT_PATH/vite.config.js" ]] || [[ -f "$PROJECT_PATH/vite.config.ts" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " vite " ]]; then
            frameworks+=("vite")
            confidences+=("medium")
        fi
    fi

    # Python frameworks
    if [[ -f "$PROJECT_PATH/pyproject.toml" ]]; then
        if grep -q "django" "$PROJECT_PATH/pyproject.toml" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " django " ]]; then
                frameworks+=("django")
                confidences+=("high")
            fi
        fi
        if grep -q "fastapi" "$PROJECT_PATH/pyproject.toml" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " fastapi " ]]; then
                frameworks+=("fastapi")
                confidences+=("high")
            fi
        fi
        if grep -q "poetry" "$PROJECT_PATH/pyproject.toml" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " poetry " ]]; then
                frameworks+=("poetry")
                confidences+=("medium")
            fi
        fi
        if [[ ${#frameworks[@]} -eq 0 ]]; then
            frameworks+=("python")
            confidences+=("medium")
        fi
    fi
    if [[ -f "$PROJECT_PATH/requirements.txt" ]]; then
        if grep -q "django" "$PROJECT_PATH/requirements.txt" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " django " ]]; then
                frameworks+=("django")
                confidences+=("high")
            fi
        fi
        if grep -q "fastapi" "$PROJECT_PATH/requirements.txt" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " fastapi " ]]; then
                frameworks+=("fastapi")
                confidences+=("high")
            fi
        fi
        if grep -q "flask" "$PROJECT_PATH/requirements.txt" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " flask " ]]; then
                frameworks+=("flask")
                confidences+=("high")
            fi
        fi
        if [[ ${#frameworks[@]} -eq 0 ]]; then
            frameworks+=("python")
            confidences+=("medium")
        fi
    fi
    if [[ -f "$PROJECT_PATH/Pipfile" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " pipenv " ]]; then
            frameworks+=("pipenv")
            confidences+=("high")
        fi
    fi
    if [[ -f "$PROJECT_PATH/manage.py" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " django " ]]; then
            frameworks+=("django")
            confidences+=("high")
        fi
    fi

    # Mobile frameworks
    if [[ -f "$PROJECT_PATH/pubspec.yaml" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " flutter " ]]; then
            frameworks+=("flutter")
            confidences+=("high")
        fi
    fi

    # Go frameworks
    if [[ -f "$PROJECT_PATH/go.mod" ]]; then
        if grep -q "gin-gonic" "$PROJECT_PATH/go.mod" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " gin " ]]; then
                frameworks+=("gin")
                confidences+=("high")
            fi
        fi
        if grep -q "labstack/echo" "$PROJECT_PATH/go.mod" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " echo " ]]; then
                frameworks+=("echo")
                confidences+=("high")
            fi
        fi
        if grep -q "gofiber" "$PROJECT_PATH/go.mod" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " fiber " ]]; then
                frameworks+=("fiber")
                confidences+=("high")
            fi
        fi
        if [[ ${#frameworks[@]} -eq 0 ]]; then
            frameworks+=("go")
            confidences+=("medium")
        fi
    fi

    # Rust frameworks
    if [[ -f "$PROJECT_PATH/Cargo.toml" ]]; then
        if grep -q "actix-web" "$PROJECT_PATH/Cargo.toml" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " actix " ]]; then
                frameworks+=("actix")
                confidences+=("high")
            fi
        fi
        if grep -q "rocket" "$PROJECT_PATH/Cargo.toml" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " rocket " ]]; then
                frameworks+=("rocket")
                confidences+=("high")
            fi
        fi
        if grep -q "axum" "$PROJECT_PATH/Cargo.toml" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " axum " ]]; then
                frameworks+=("axum")
                confidences+=("high")
            fi
        fi
        if [[ ${#frameworks[@]} -eq 0 ]]; then
            frameworks+=("rust")
            confidences+=("medium")
        fi
    fi

    # Java frameworks
    if [[ -f "$PROJECT_PATH/pom.xml" ]]; then
        if grep -q "spring-boot" "$PROJECT_PATH/pom.xml" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " springboot " ]]; then
                frameworks+=("springboot")
                confidences+=("high")
            fi
        fi
        if grep -q "spring" "$PROJECT_PATH/pom.xml" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " spring " ]]; then
                frameworks+=("spring")
                confidences+=("high")
            fi
        fi
        if [[ ${#frameworks[@]} -eq 0 ]]; then
            frameworks+=("maven")
            confidences+=("medium")
        fi
    fi
    if [[ -f "$PROJECT_PATH/build.gradle" ]] || [[ -f "$PROJECT_PATH/build.gradle.kts" ]]; then
        if grep -q "org.springframework.boot" "$PROJECT_PATH/build.gradle"* 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " springboot " ]]; then
                frameworks+=("springboot")
                confidences+=("high")
            fi
        fi
        if grep -q "io.ktor" "$PROJECT_PATH/build.gradle"* 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " ktor " ]]; then
                frameworks+=("ktor")
                confidences+=("high")
            fi
        fi
        if [[ ${#frameworks[@]} -eq 0 ]]; then
            frameworks+=("gradle")
            confidences+=("medium")
        fi
    fi

    # Ruby frameworks
    if [[ -f "$PROJECT_PATH/Gemfile" ]]; then
        if grep -q "rails" "$PROJECT_PATH/Gemfile" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " rails " ]]; then
                frameworks+=("rails")
                confidences+=("high")
            fi
        fi
        if grep -q "sinatra" "$PROJECT_PATH/Gemfile" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " sinatra " ]]; then
                frameworks+=("sinatra")
                confidences+=("high")
            fi
        fi
        if [[ ${#frameworks[@]} -eq 0 ]]; then
            frameworks+=("ruby")
            confidences+=("medium")
        fi
    fi
    if [[ -f "$PROJECT_PATH/config/application.rb" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " rails " ]]; then
            frameworks+=("rails")
            confidences+=("high")
        fi
    fi

    # PHP frameworks
    if [[ -f "$PROJECT_PATH/composer.json" ]]; then
        if grep -q "laravel/framework" "$PROJECT_PATH/composer.json" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " laravel " ]]; then
                frameworks+=("laravel")
                confidences+=("high")
            fi
        fi
        if grep -q "symfony" "$PROJECT_PATH/composer.json" 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " symfony " ]]; then
                frameworks+=("symfony")
                confidences+=("high")
            fi
        fi
        if [[ ${#frameworks[@]} -eq 0 ]]; then
            frameworks+=("php")
            confidences+=("medium")
        fi
    fi
    if [[ -f "$PROJECT_PATH/wp-config.php" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " wordpress " ]]; then
            frameworks+=("wordpress")
            confidences+=("high")
        fi
    fi

    # C#/.NET frameworks
    if [[ -f "$PROJECT_PATH/*.csproj" ]]; then
        if grep -q "Microsoft.AspNetCore" "$PROJECT_PATH"/*.csproj 2>/dev/null; then
            if [[ ! " ${frameworks[*]} " =~ " aspnet " ]]; then
                frameworks+=("aspnet")
                confidences+=("high")
            fi
        fi
        if [[ ${#frameworks[@]} -eq 0 ]]; then
            frameworks+=("dotnet")
            confidences+=("medium")
        fi
    fi

    # Swift frameworks
    if [[ -f "$PROJECT_PATH/Package.swift" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " swift " ]]; then
            frameworks+=("swift")
            confidences+=("high")
        fi
    fi

    # Container/DevOps
    if [[ -f "$PROJECT_PATH/Dockerfile" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " docker " ]]; then
            frameworks+=("docker")
            confidences+=("medium")
        fi
    fi
    if [[ -f "$PROJECT_PATH/docker-compose.yml" ]] || [[ -f "$PROJECT_PATH/docker-compose.yaml" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " docker-compose " ]]; then
            frameworks+=("docker-compose")
            confidences+=("medium")
        fi
    fi
    if [[ -f "$PROJECT_PATH/Terrafile" ]] || [[ -f "$PROJECT_PATH/main.tf" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " terraform " ]]; then
            frameworks+=("terraform")
            confidences+=("high")
        fi
    fi

    # Configuration management
    if [[ -f "$PROJECT_PATH/Vagrantfile" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " vagrant " ]]; then
            frameworks+=("vagrant")
            confidences+=("high")
        fi
    fi
    if [[ -f "$PROJECT_PATH/kustomization.yaml" ]]; then
        if [[ ! " ${frameworks[*]} " =~ " kustomize " ]]; then
            frameworks+=("kustomize")
            confidences+=("high")
        fi
    fi

    # Check for Plaesy framework itself
    if [[ -f "$PROJECT_PATH/README.md" ]] && grep -q "Plaesy Spec-Kit" "$PROJECT_PATH/README.md"; then
        if [[ ! " ${frameworks[*]} " =~ " spec-kit " ]]; then
            frameworks+=("spec-kit")
            confidences+=("high")
        fi
    fi

    # Return results
    echo "${frameworks[*]}:${confidences[*]}"
}

# Function to detect project type (legacy function for backward compatibility)
detect_project_type() {
    local frameworks_confidence="${FRAMEWORKS_CACHE:-$(detect_all_frameworks)}"
    local frameworks_str="${frameworks_confidence%:*}"
    local confidences_str="${frameworks_confidence#*:}"

    # Convert to arrays
    local frameworks=($frameworks_str)
    local confidences=($confidences_str)

    # If no frameworks found, return generic
    if [[ ${#frameworks[@]} -eq 0 ]]; then
        echo "generic:low"
        return
    fi

    # If only one framework, return it
    if [[ ${#frameworks[@]} -eq 1 ]]; then
        echo "${frameworks[0]}:${confidences[0]}"
        return
    fi

    # Multiple frameworks - prioritize by confidence and type
    local primary_framework=""
    local primary_confidence="low"
    local primary_type=""

    # Priority: web frameworks > mobile > backend > devops
    local priority_frameworks=("nextjs" "react" "vue" "nuxt" "svelte" "astro" "remix" "gatsby" "django" "flask" "fastapi" "laravel" "rails" "express" "nestjs" "springboot" "flutter" "actix" "rocket" "gin" "echo" "fiber" "rust" "go" "java" "kotlin" "nodejs" "python" "php" "ruby" "swift")

    for priority in "${priority_frameworks[@]}"; do
        for i in "${!frameworks[@]}"; do
            if [[ "${frameworks[$i]}" == "$priority" ]]; then
                echo "${frameworks[$i]}:${confidences[$i]}"
                return
            fi
        done
    done

    # Fallback to first detected framework
    echo "${frameworks[0]}:${confidences[0]}"
}

# Function to get all frameworks with their confidences
get_frameworks_list() {
    local frameworks_confidence="${FRAMEWORKS_CACHE:-$(detect_all_frameworks)}"
    local frameworks_str="${frameworks_confidence%:*}"
    local confidences_str="${frameworks_confidence#*:}"

    # Convert to arrays
    local frameworks=($frameworks_str)
    local confidences=($confidences_str)

    # Build JSON array
    local result="["
    for i in "${!frameworks[@]}"; do
        if [[ $i -gt 0 ]]; then
            result+=","
        fi
        result+="{\"framework\":\"${frameworks[$i]}\",\"confidence\":\"${confidences[$i]}\"}"
    done
    result+="]"

    echo "$result"
}

# Function to detect all programming languages
detect_all_languages() {
    local languages=()

    # Single batched pass: one find + one awk instead of 18 separate
    # `find | wc -l` spawns (each spawn costs ~110-200ms on MSYS/Windows).
    # awk regexes mirror the -name globs used previously exactly.
    local detected
    detected=$(find "$PROJECT_PATH" -type f \
        \( -name "*.js" -o -name "*.jsx" -o -name "*.ts" -o -name "*.tsx" -o \
        -name "*.py" -o -name "*.go" -o -name "*.java" -o -name "*.dart" -o \
        -name "*.cpp" -o -name "*.cc" -o -name "*.cxx" -o -name "*.c" -o \
        -name "*.h" -o -name "*.php" -o -name "*.rb" -o -name "*.rs" -o \
        -name "*.swift" -o -name "*.kt" -o -name "*.scala" -o -name "*.sh" -o \
        -name "*.html" -o -name "*.css" -o -name "*.scss" -o -name "*.sass" \) \
        ! -path "*/.*/*" ! -path "*/node_modules/*" \
        -print 2>/dev/null |
        awk '
            { f = $0; sub(/^.*\//, "", f) }
            f ~ /\.(js|jsx)$/   { js++ }
            f ~ /\.(ts|tsx)$/   { ts++ }
            f ~ /\.py$/         { py++ }
            f ~ /\.go$/         { go++ }
            f ~ /\.java$/       { java++ }
            f ~ /\.dart$/       { dart++ }
            f ~ /\.(cpp|cc|cxx)$/ { cpp++ }
            f ~ /\.c$/          { c++ }
            f ~ /\.h$/          { h++ }
            f ~ /\.php$/        { php++ }
            f ~ /\.rb$/         { rb++ }
            f ~ /\.rs$/         { rs++ }
            f ~ /\.swift$/      { swift++ }
            f ~ /\.kt$/         { kt++ }
            f ~ /\.scala$/      { scala++ }
            f ~ /\.sh$/         { sh++ }
            f ~ /\.html$/       { html++ }
            f ~ /\.(css|scss|sass)$/ { css++ }
            END {
                if (js>0) print "JavaScript"
                if (ts>0) print "TypeScript"
                if (py>0) print "Python"
                if (go>0) print "Go"
                if (java>0) print "Java"
                if (dart>0) print "Dart"
                if (cpp>0) print "C++"
                if (c>0) print "C"
                if (h>0) print "C/C++ Headers"
                if (php>0) print "PHP"
                if (rb>0) print "Ruby"
                if (rs>0) print "Rust"
                if (swift>0) print "Swift"
                if (kt>0) print "Kotlin"
                if (scala>0) print "Scala"
                if (sh>0) print "Shell"
                if (html>0) print "HTML"
                if (css>0) print "CSS"
            }')

    # Add detected languages in canonical order
    while IFS= read -r lang; do
        [[ -n "$lang" ]] && languages+=("$lang")
    done <<< "$detected"

    # If no languages found, default to JavaScript
    if [[ ${#languages[@]} -eq 0 ]]; then
        languages=("JavaScript")
    fi

    # Convert array to comma-separated string
    local result=$(IFS=','; echo "${languages[*]}")
    echo "$result"
}

# Function to detect primary language (most files)
detect_primary_language() {
    local languages_str=$(detect_all_languages)
    # Return first language as primary
    echo "$languages_str" | cut -d, -f1
}

# Function to detect development tools
detect_development_tools() {
    local tools=()

    # Version control
    if [[ -d "$PROJECT_PATH/.git" ]]; then
        tools+=("Git")
    fi
    if [[ -d "$PROJECT_PATH/.svn" ]]; then
        tools+=("Subversion")
    fi
    if [[ -f "$PROJECT_PATH/.hg" ]]; then
        tools+=("Mercurial")
    fi

    # Package managers
    if [[ -f "$PROJECT_PATH/package.json" ]] || [[ -f "$PROJECT_PATH/package-lock.json" ]] || [[ -f "$PROJECT_PATH/yarn.lock" ]] || [[ -f "$PROJECT_PATH/pnpm-lock.yaml" ]]; then
        tools+=("npm/yarn/pnpm")
    fi
    if [[ -f "$PROJECT_PATH/requirements.txt" ]] || [[ -f "$PROJECT_PATH/Pipfile" ]] || [[ -f "$PROJECT_PATH/pyproject.toml" ]]; then
        tools+=("pip/poetry")
    fi
    if [[ -f "$PROJECT_PATH/go.mod" ]]; then
        tools+=("Go Modules")
    fi
    if [[ -f "$PROJECT_PATH/Cargo.toml" ]]; then
        tools+=("Cargo")
    fi
    if [[ -f "$PROJECT_PATH/pom.xml" ]] || [[ -f "$PROJECT_PATH/build.gradle" ]] || [[ -f "$PROJECT_PATH/build.gradle.kts" ]]; then
        tools+=("Maven/Gradle")
    fi
    if [[ -f "$PROJECT_PATH/Gemfile" ]]; then
        tools+=("Bundler")
    fi
    if [[ -f "$PROJECT_PATH/composer.json" ]]; then
        tools+=("Composer")
    fi
    if [[ -f "$PROJECT_PATH/pubspec.yaml" ]]; then
        tools+=("Pub")
    fi

    # CI/CD tools
    if [[ -f "$PROJECT_PATH/.github/workflows" ]] || [[ -d "$PROJECT_PATH/.github/workflows" ]]; then
        tools+=("GitHub Actions")
    fi
    if [[ -f "$PROJECT_PATH/.gitlab-ci.yml" ]]; then
        tools+=("GitLab CI")
    fi
    if [[ -f "$PROJECT_PATH/Jenkinsfile" ]]; then
        tools+=("Jenkins")
    fi
    if [[ -f "$PROJECT_PATH/azure-pipelines.yml" ]]; then
        tools+=("Azure Pipelines")
    fi

    # Testing frameworks
    if [[ -f "$PROJECT_PATH/jest.config.js" ]] || [[ -f "$PROJECT_PATH/jest.config.json" ]] || grep -q "jest" "$PROJECT_PATH/package.json" 2>/dev/null; then
        tools+=("Jest")
    fi
    if [[ -f "$PROJECT_PATH/vitest.config.js" ]] || [[ -f "$PROJECT_PATH/vitest.config.ts" ]]; then
        tools+=("Vitest")
    fi
    if grep -q "pytest" "$PROJECT_PATH/requirements.txt" 2>/dev/null || [[ -f "$PROJECT_PATH/pytest.ini" ]] || [[ -d "$PROJECT_PATH/tests" ]]; then
        tools+=("pytest")
    fi
    if [[ -f "$PROJECT_PATH/go.mod" ]] && grep -q "test" "$PROJECT_PATH/go.mod" 2>/dev/null; then
        tools+=("Go test")
    fi

    # Linting and formatting
    if [[ -f "$PROJECT_PATH/.eslintrc.js" ]] || [[ -f "$PROJECT_PATH/.eslintrc.json" ]] || [[ -f "$PROJECT_PATH/.eslintrc.yml" ]] || grep -q "eslint" "$PROJECT_PATH/package.json" 2>/dev/null; then
        tools+=("ESLint")
    fi
    if [[ -f "$PROJECT_PATH/.prettierrc" ]] || [[ -f "$PROJECT_PATH/.prettierrc.json" ]] || [[ -f "$PROJECT_PATH/.prettierrc.js" ]] || grep -q "prettier" "$PROJECT_PATH/package.json" 2>/dev/null; then
        tools+=("Prettier")
    fi
    if [[ -f "$PROJECT_PATH/pylintrc" ]] || [[ -f "$PROJECT_PATH/.flake8" ]] || grep -q "flake8" "$PROJECT_PATH/requirements.txt" 2>/dev/null; then
        tools+=("Python Linters")
    fi
    if [[ -f "$PROJECT_PATH/.golangci.yml" ]] || [[ -f "$PROJECT_PATH/golangci.yml" ]]; then
        tools+=("golangci-lint")
    fi
    if [[ -f "$PROJECT_PATH/rustfmt.toml" ]] || [[ -f "$PROJECT_PATH/.rustfmt.toml" ]]; then
        tools+=("rustfmt")
    fi
    if [[ -f "$PROJECT_PATH/clippy.toml" ]] || [[ -f "$PROJECT_PATH/.clippy.toml" ]]; then
        tools+=("Clippy")
    fi

    # Docker and containerization
    if [[ -f "$PROJECT_PATH/Dockerfile" ]] || [[ -f "$PROJECT_PATH/docker-compose.yml" ]] || [[ -f "$PROJECT_PATH/docker-compose.yaml" ]]; then
        tools+=("Docker")
    fi
    if [[ -f "$PROJECT_PATH/kubernetes" ]] || [[ -d "$PROJECT_PATH/kubernetes" ]] || [[ -f "$PROJECT_PATH/*.yaml" ]] && grep -q "apiVersion:" "$PROJECT_PATH"/*.yaml 2>/dev/null; then
        tools+=("Kubernetes")
    fi

    # If no tools found, default to basic
    if [[ ${#tools[@]} -eq 0 ]]; then
        tools=("Manual")
    fi

    # Convert array to comma-separated string
    local result=$(IFS=','; echo "${tools[*]}")
    echo "$result"
}

# Function to detect build systems
detect_build_systems() {
    local systems=()

    # JavaScript/TypeScript build tools
    if [[ -f "$PROJECT_PATH/package.json" ]]; then
        if grep -q "webpack" "$PROJECT_PATH/package.json" 2>/dev/null || [[ -f "$PROJECT_PATH/webpack.config.js" ]] || [[ -f "$PROJECT_PATH/webpack.config.ts" ]]; then
            systems+=("Webpack")
        fi
        if grep -q "vite" "$PROJECT_PATH/package.json" 2>/dev/null || [[ -f "$PROJECT_PATH/vite.config.js" ]] || [[ -f "$PROJECT_PATH/vite.config.ts" ]]; then
            systems+=("Vite")
        fi
        if grep -q "rollup" "$PROJECT_PATH/package.json" 2>/dev/null || [[ -f "$PROJECT_PATH/rollup.config.js" ]] || [[ -f "$PROJECT_PATH/rollup.config.ts" ]]; then
            systems+=("Rollup")
        fi
        if grep -q "parcel" "$PROJECT_PATH/package.json" 2>/dev/null || [[ -f "$PROJECT_PATH/parcelrc" ]] || [[ -f "$PROJECT_PATH/.parcelrc" ]]; then
            systems+=("Parcel")
        fi
        if grep -q "esbuild" "$PROJECT_PATH/package.json" 2>/dev/null || [[ -f "$PROJECT_PATH/esbuild.js" ]] || [[ -f "$PROJECT_PATH/esbuild.config.js" ]]; then
            systems+=("esbuild")
        fi
        if grep -q "turbo" "$PROJECT_PATH/package.json" 2>/dev/null || [[ -f "$PROJECT_PATH/turbo.json" ]]; then
            systems+=("Turbopack")
        fi
        # Next.js has its own build system
        if grep -q "next" "$PROJECT_PATH/package.json" 2>/dev/null; then
            systems+=("Next.js")
        fi
        # Create React App
        if [[ -f "$PROJECT_PATH/public/index.html" ]] && grep -q "react-scripts" "$PROJECT_PATH/package.json" 2>/dev/null; then
            systems+=("Create React App")
        fi
    fi

    # Python build systems
    if [[ -f "$PROJECT_PATH/pyproject.toml" ]]; then
        if grep -q "setuptools" "$PROJECT_PATH/pyproject.toml" 2>/dev/null; then
            systems+=("setuptools")
        fi
        if grep -q "poetry" "$PROJECT_PATH/pyproject.toml" 2>/dev/null; then
            systems+=("Poetry")
        fi
        if grep -q "hatch" "$PROJECT_PATH/pyproject.toml" 2>/dev/null; then
            systems+=("Hatch")
        fi
        if grep -q "pdm" "$PROJECT_PATH/pyproject.toml" 2>/dev/null; then
            systems+=("PDM")
        fi
    fi
    if [[ -f "$PROJECT_PATH/setup.py" ]]; then
        systems+=("setuptools")
    fi
    if [[ -f "$PROJECT_PATH/Makefile" ]] && grep -q "python" "$PROJECT_PATH/Makefile" 2>/dev/null; then
        systems+=("Make")
    fi

    # Java build systems
    if [[ -f "$PROJECT_PATH/pom.xml" ]]; then
        systems+=("Maven")
    fi
    if [[ -f "$PROJECT_PATH/build.gradle" ]] || [[ -f "$PROJECT_PATH/build.gradle.kts" ]]; then
        systems+=("Gradle")
    fi
    if [[ -f "$PROJECT_PATH/build.xml" ]]; then
        systems+=("Ant")
    fi

    # Go build systems
    if [[ -f "$PROJECT_PATH/go.mod" ]]; then
        systems+=("Go Modules")
    fi
    if [[ -f "$PROJECT_PATH/Makefile" ]] && grep -q "go" "$PROJECT_PATH/Makefile" 2>/dev/null; then
        systems+=("Make")
    fi

    # Rust build systems
    if [[ -f "$PROJECT_PATH/Cargo.toml" ]]; then
        systems+=("Cargo")
    fi

    # C/C++ build systems
    if [[ -f "$PROJECT_PATH/CMakeLists.txt" ]]; then
        systems+=("CMake")
    fi
    if [[ -f "$PROJECT_PATH/Makefile" ]]; then
        systems+=("Make")
    fi
    if [[ -f "$PROJECT_PATH/meson.build" ]]; then
        systems+=("Meson")
    fi
    if [[ -f "$PROJECT_PATH/configure.ac" ]] || [[ -f "$PROJECT_PATH/Makefile.am" ]]; then
        systems+=("Autotools")
    fi

    # Ruby build systems
    if [[ -f "$PROJECT_PATH/Gemfile" ]]; then
        systems+=("Bundler")
    fi
    if [[ -f "$PROJECT_PATH/Rakefile" ]]; then
        systems+=("Rake")
    fi

    # PHP build systems
    if [[ -f "$PROJECT_PATH/composer.json" ]]; then
        systems+=("Composer")
    fi

    # Mobile build systems
    if [[ -f "$PROJECT_PATH/pubspec.yaml" ]]; then
        systems+=("Pub")
    fi
    if [[ -d "$PROJECT_PATH/android" ]] && [[ -f "$PROJECT_PATH/android/build.gradle" ]]; then
        systems+=("Gradle (Android)")
    fi
    if [[ -d "$PROJECT_PATH/ios" ]] && [[ -f "$PROJECT_PATH/ios/Runner.xcodeproj/project.pbxproj" ]]; then
        systems+=("Xcode")
    fi

    # Container and infrastructure
    if [[ -f "$PROJECT_PATH/Dockerfile" ]]; then
        systems+=("Docker")
    fi
    if [[ -f "$PROJECT_PATH/docker-compose.yml" ]] || [[ -f "$PROJECT_PATH/docker-compose.yaml" ]]; then
        systems+=("Docker Compose")
    fi
    if [[ -f "$PROJECT_PATH/Terrafile" ]] || [[ -f "$PROJECT_PATH/main.tf" ]]; then
        systems+=("Terraform")
    fi

    # If no build systems found, default to Manual
    if [[ ${#systems[@]} -eq 0 ]]; then
        systems=("Manual")
    fi

    # Convert array to comma-separated string
    local result=$(IFS=','; echo "${systems[*]}")
    echo "$result"
}

# Function to generate AI insights
generate_ai_insights() {
    local project_type="$1"
    local total_files="$2"
    local overview=""
    local recommendations=()

    case "$project_type" in
        # JavaScript/TypeScript frameworks
        "nextjs")
            overview="Next.js application with React framework featuring SSR/SSG capabilities"
            recommendations+=("Implement comprehensive SEO optimization with metadata API")
            recommendations+=("Add performance monitoring with Next.js Analytics")
            recommendations+=("Consider incremental static regeneration for dynamic content")
            ;;
        "react")
            overview="React application with component-based architecture and hooks system"
            recommendations+=("Consider adding TypeScript for better type safety")
            recommendations+=("Implement comprehensive unit tests with Jest and React Testing Library")
            recommendations+=("Add state management solution like Redux Toolkit or Zustand")
            ;;
        "vue")
            overview="Vue.js application with reactive component system and progressive framework"
            recommendations+=("Consider Vue 3 Composition API for better code organization")
            recommendations+=("Add TypeScript support with vue-tsc")
            recommendations+=("Implement Vue Router for single page applications")
            ;;
        "nuxt")
            overview="Nuxt.js application providing Vue.js framework with SSR capabilities"
            recommendations+=("Optimize for search engines with proper meta tags")
            recommendations+=("Implement static site generation for better performance")
            recommendations+=("Add Nuxt modules for enhanced functionality")
            ;;
        "svelte")
            overview="Svelte application with compile-time optimizations and reactive components"
            recommendations+=("Add TypeScript support for better development experience")
            recommendations+=("Implement SvelteKit for full-stack applications")
            recommendations+=("Consider adding unit tests with Vitest")
            ;;
        "astro")
            overview="Astro application focused on content-rich sites with minimal JavaScript"
            recommendations+=("Optimize for performance with island architecture")
            recommendations+=("Implement proper SEO and meta management")
            recommendations+=("Add Markdown/MDX content processing capabilities")
            ;;
        "remix")
            overview="Remix web framework with focus on web fundamentals and nested routes"
            recommendations+=("Implement proper error boundaries and loading states")
            recommendations+=("Add database integration with Prisma or similar")
            recommendations+=("Optimize for progressive enhancement")
            ;;
        "gatsby")
            overview="Gatsby static site generator with React and GraphQL data layer"
            recommendations+=("Optimize build performance with Gatsby caching")
            recommendations+=("Implement proper image optimization with gatsby-plugin-image")
            recommendations+=("Add SEO optimization with gatsby-plugin-react-helmet")
            ;;
        "express")
            overview="Express.js web application with minimal Node.js framework"
            recommendations+=("Add comprehensive error handling middleware")
            recommendations+=("Implement proper logging with Winston or Morgan")
            recommendations+=("Consider adding TypeScript support")
            ;;
        "nestjs")
            overview="NestJS application with TypeScript and architectural patterns"
            recommendations+=("Implement comprehensive validation with class-validator")
            recommendations+=("Add OpenAPI/Swagger documentation")
            recommendations+=("Consider microservices architecture with NestJS modules")
            ;;
        "vite")
            overview="Vite-powered development environment with fast build tooling"
            recommendations+=("Configure proper build optimization for production")
            recommendations+=("Add plugin ecosystem for enhanced functionality")
            recommendations+=("Implement HMR optimization for development workflow")
            ;;
        # Python frameworks
        "django")
            overview="Django web application with batteries-included framework"
            recommendations+=("Add comprehensive API documentation with Django REST framework")
            recommendations+=("Implement proper caching strategies with Redis")
            recommendations+=("Add background task processing with Celery")
            ;;
        "fastapi")
            overview="FastAPI application with high-performance async web framework"
            recommendations+=("Add comprehensive OpenAPI/Swagger documentation")
            recommendations+=("Implement dependency injection for better testing")
            recommendations+=("Add async database operations with SQLAlchemy 2.0")
            ;;
        "flask")
            overview="Flask web application with lightweight and extensible framework"
            recommendations+=("Add SQLAlchemy for database management")
            recommendations+=("Implement proper application factory pattern")
            recommendations+=("Add Flask extensions for enhanced functionality")
            ;;
        "poetry")
            overview="Python project managed with Poetry for dependency management"
            recommendations+=("Configure proper dependency groups for development")
            recommendations+=("Add comprehensive pytest configuration")
            recommendations+=("Implement pre-commit hooks with poetry hooks")
            ;;
        "pipenv")
            overview="Python project with Pipenv for virtual environment and dependency management"
            recommendations+=("Configure proper Pipfile for development dependencies")
            recommendations+=("Add pipenv scripts for common tasks")
            ;;
        # Mobile frameworks
        "flutter")
            overview="Flutter application with cross-platform mobile development framework"
            recommendations+=("Add comprehensive unit tests with flutter test")
            recommendations+=("Add effective linter configuration with flutter analyze")
            recommendations+=("Consider adding integration tests for critical user flows")
            ;;
        # Go frameworks
        "gin")
            overview="Gin web framework for Go with high-performance HTTP router"
            recommendations+=("Add comprehensive middleware for authentication and logging")
            recommendations+=("Implement proper error handling and validation")
            recommendations+=("Add database integration with GORM or sqlx")
            ;;
        "echo")
            overview="Echo web framework for Go with high performance and extensibility"
            recommendations+=("Implement proper middleware stack")
            recommendations+=("Add comprehensive error handling")
            recommendations+=("Consider adding database integration")
            ;;
        "fiber")
            overview="Fiber web framework for Go inspired by Express.js"
            recommendations+=("Add comprehensive middleware configuration")
            recommendations+=("Implement proper error handling")
            recommendations+=("Add database integration for full-stack applications")
            ;;
        # Rust frameworks
        "actix")
            overview="Actix Web framework for Rust with high-performance actor model"
            recommendations+=("Implement proper middleware for authentication")
            recommendations+=("Add comprehensive error handling")
            recommendations+=("Consider adding database integration with sqlx")
            ;;
        "rocket")
            overview="Rocket web framework for Rust with type-safe and fast development"
            recommendations+=("Implement proper database integration with Diesel")
            recommendations+=("Add comprehensive testing strategies")
            recommendations+=("Configure proper async/await patterns")
            ;;
        "axum")
            overview="Axum web framework for Rust built on Tokio with async support"
            recommendations+=("Implement proper async database integration")
            recommendations+=("Add comprehensive middleware for logging and tracing")
            recommendations+=("Consider adding tower middleware ecosystem")
            ;;
        "rust")
            overview="Rust application with focus on memory safety and performance"
            recommendations+=("Add comprehensive unit and integration tests")
            recommendations+=("Implement proper error handling with Result types")
            recommendations+=("Consider adding async runtime optimization")
            ;;
        # Java frameworks
        "springboot")
            overview="Spring Boot application with enterprise-grade framework"
            recommendations+=("Add comprehensive Spring Security configuration")
            recommendations+=("Implement proper database integration with Spring Data JPA")
            recommendations+=("Add RESTful API documentation with OpenAPI")
            ;;
        "spring")
            overview="Spring framework application with dependency injection and AOP"
            recommendations+=("Configure proper Spring context management")
            recommendations+=("Add comprehensive testing with Spring Test")
            recommendations+=("Implement proper MVC patterns")
            ;;
        "maven")
            overview="Java project managed with Maven build system"
            recommendations+=("Configure proper dependency management")
            recommendations+=("Add comprehensive plugin configuration")
            recommendations+=("Implement proper build lifecycle")
            ;;
        "gradle")
            overview="Java project managed with Gradle build system"
            recommendations+=("Configure proper build scripts with Kotlin DSL")
            recommendations+=("Add comprehensive plugin ecosystem")
            recommendations+=("Implement proper dependency management")
            ;;
        "ktor")
            overview="Ktor framework for Kotlin with asynchronous web applications"
            recommendations+=("Add comprehensive structured logging")
            recommendations+=("Implement proper content negotiation")
            recommendations+=("Add database integration with Exposed")
            ;;
        # Ruby frameworks
        "rails")
            overview="Ruby on Rails application with convention over configuration"
            recommendations+=("Add comprehensive testing with RSpec or Minitest")
            recommendations+=("Implement proper background jobs with Sidekiq")
            recommendations+=("Add API versioning strategies")
            ;;
        "sinatra")
            overview="Sinatra DSL for Ruby with lightweight web applications"
            recommendations+=("Add comprehensive middleware configuration")
            recommendations+=("Implement proper RESTful API patterns")
            recommendations+=("Consider adding ActiveRecord for database")
            ;;
        # PHP frameworks
        "laravel")
            overview="Laravel PHP framework with elegant syntax and comprehensive ecosystem"
            recommendations+=("Add comprehensive API documentation with Laravel API resources")
            recommendations+=("Implement proper queue system with Redis and Horizon")
            recommendations+=("Add comprehensive testing with PHPUnit and Laravel Dusk")
            ;;
        "symfony")
            overview="Symfony PHP framework with reusable components and MVC architecture"
            recommendations+=("Configure proper service container and dependency injection")
            recommendations+=("Add comprehensive form validation and security")
            recommendations+=("Implement proper API platform integration")
            ;;
        "wordpress")
            overview="WordPress content management system with plugin and theme development"
            recommendations+=("Implement proper WordPress coding standards")
            recommendations+=("Add comprehensive security practices")
            recommendations+=("Consider adding custom post types and taxonomies")
            ;;
        # C#/.NET frameworks
        "aspnet")
            overview="ASP.NET Core application with cross-platform web framework"
            recommendations+=("Add comprehensive authentication and authorization")
            recommendations+=("Implement proper dependency injection patterns")
            recommendations+=("Add Entity Framework Core for database operations")
            ;;
        "dotnet")
            overview=".NET application with managed runtime and comprehensive framework"
            recommendations+=("Configure proper NuGet package management")
            recommendations+=("Add comprehensive unit testing with xUnit or NUnit")
            recommendations+=("Implement proper async/await patterns")
            ;;
        # Swift
        "swift")
            overview="Swift application with safe and performant programming language"
            recommendations+=("Add comprehensive unit tests with XCTest")
            recommendations+=("Implement proper error handling with Result types")
            recommendations+=("Consider adding async/await patterns")
            ;;
        # DevOps/Infrastructure
        "docker")
            overview="Docker containerized application with containerization platform"
            recommendations+=("Add multi-stage builds for optimization")
            recommendations+=("Implement proper health checks")
            recommendations+=("Consider adding Docker Compose for orchestration")
            ;;
        "docker-compose")
            overview="Docker Compose application with multi-container orchestration"
            recommendations+=("Configure proper networking and volumes")
            recommendations+=("Add environment-specific configurations")
            recommendations+=("Implement proper service dependencies")
            ;;
        "terraform")
            overview="Terraform infrastructure as code with cloud resource management"
            recommendations+=("Implement proper state management strategies")
            recommendations+=("Add comprehensive variable validation")
            recommendations+=("Configure proper remote state backends")
            ;;
        "vagrant")
            overview="Vagrant development environment with virtualization management"
            recommendations+=("Configure proper provisioning scripts")
            recommendations+=("Add multi-machine setup if needed")
            recommendations+=("Implement proper network configuration")
            ;;
        "kustomize")
            overview="Kustomize Kubernetes configuration management with template-free customization"
            recommendations+=("Configure proper overlay structure")
            recommendations+=("Add comprehensive resource organization")
            recommendations+=("Implement proper secret management")
            ;;
        # Plaesy
        "spec-kit")
            overview="Plaesy Spec-Kit framework for AI-assisted development"
            recommendations+=("Add more AI platform integrations")
            ;;
        *)
            overview="Generic project structure"
            recommendations+=("Add README.md with project documentation")
            recommendations+=("Consider adding automated testing")
            recommendations+=("Implement proper version control practices")
            ;;
    esac

    if [[ $total_files -lt 10 ]]; then
        recommendations+=("Expand project with additional features and modules")
    fi

    local rec_json=$(printf '"%s",' "${recommendations[@]}" | sed 's/,$//')

    echo "{\"overview\": \"$overview\", \"recommendations\": [${rec_json:-}]}"
}


# Function to generate comprehensive project.json
generate_project_json() {
    log_info "Generating comprehensive project.json..."

    local type_confidence="${PROJECT_TYPE_CACHE:-$(detect_project_type)}"
    local project_type="${type_confidence%%:*}"
    local confidence="${type_confidence#*:}"
    local total_files=$(find "$PROJECT_PATH" -type f ! -path "*/.*/*" ! -path "*/node_modules/*" | wc -l)
    local project_name="${PROJECT_PATH%/}"
    project_name="${project_name##*/}"
    local all_languages="${ALL_LANGUAGES_CACHE:-$(detect_all_languages)}"
    local development_tools=$(detect_development_tools)
    local build_systems=$(detect_build_systems)

    # Get all frameworks detected
    local frameworks_list=$(get_frameworks_list)
    # Extract framework names from the JSON array once (no jq dependency,
    # no per-item spawns): [{"framework":"x",...},{"framework":"y",...}]
    local frameworks_json_bodies
    frameworks_json_bodies=$(printf '%s' "$frameworks_list" | grep -o '"framework"[^}]*' || true)
    local frameworks=()
    while IFS= read -r fw_body; do
        [[ -n "$fw_body" ]] || continue
        local fw="${fw_body#*\"framework\":}"
        fw="${fw#*\"}"
        fw="${fw%%\"*}"
        [[ -n "$fw" ]] && frameworks+=("$fw")
    done <<< "$frameworks_json_bodies"
    local frameworks_count=${#frameworks[@]}

    # Generate AI insights for primary framework
    local ai_insights=$(generate_ai_insights "$project_type" "$total_files")
    local overview=""
    if [[ "$ai_insights" =~ \"overview\"[[:space:]]*:[[:space:]]*\"([^\"]*)\" ]]; then
        overview="${BASH_REMATCH[1]}"
    fi

    # Determine complexity
    local complexity="Small"
    if [[ $total_files -gt 50 ]]; then
        complexity="Large"
    elif [[ $total_files -gt 20 ]]; then
        complexity="Medium"
    fi

    # Determine project type classification
    local project_classification="single"
    if [[ $frameworks_count -gt 1 ]]; then
        project_classification="multi-framework"

        # Check for specific multi-framework patterns
        local has_frontend=false
        local has_backend=false
        local has_mobile=false
        local has_devops=false

        # Parse frameworks to detect patterns
        for framework in "${frameworks[@]}"; do
            case "$framework" in
                "nextjs"|"react"|"vue"|"nuxt"|"svelte"|"astro"|"remix"|"gatsby"|"express"|"nestjs")
                    has_frontend=true
                    ;;
                "django"|"flask"|"fastapi"|"laravel"|"rails"|"springboot"|"actix"|"rocket"|"gin"|"echo"|"fiber")
                    has_backend=true
                    ;;
                "flutter")
                    has_mobile=true
                    ;;
                "docker"|"docker-compose"|"terraform"|"kustomize"|"vagrant")
                    has_devops=true
                    ;;
            esac
        done

        # Set classification based on detected pattern
        if [[ "$has_frontend" == true && "$has_backend" == true ]]; then
            project_classification="full-stack"
        elif [[ "$has_devops" == true && ("$has_frontend" == true || "$has_backend" == true) ]]; then
            project_classification="full-stack-with-infrastructure"
        elif [[ "$has_devops" == true ]]; then
            project_classification="infrastructure-focused"
        elif [[ "$has_mobile" == true && ("$has_backend" == true || "$has_frontend" == true) ]]; then
            project_classification="mobile-backend"
        fi
    fi

    # Count file types: one find + awk instead of three separate finds.
    local sc_count doc_count cfg_count
    sc_count=0; doc_count=0; cfg_count=0
    while IFS= read -r fname; do
        case "$fname" in
            *.js|*.jsx|*.ts|*.tsx|*.py|*.go|*.java|*.dart) sc_count=$((sc_count + 1)) ;;
            *.md|*.txt) doc_count=$((doc_count + 1)) ;;
            *.json|*.yaml|*.yml|*.toml|*.ini) cfg_count=$((cfg_count + 1)) ;;
        esac
    done < <(find "$PROJECT_PATH" -type f ! -path "*/.*/*" ! -path "*/node_modules/*" -printf '%f\n' 2>/dev/null)

    # Build frameworks array for JSON (pure bash, no jq)
    local frameworks_json="["
    local first_fw=true
    for framework in "${frameworks[@]}"; do
        if [[ "$first_fw" = true ]]; then
            first_fw=false
        else
            frameworks_json+=","
        fi
        frameworks_json+="\"$framework\""
    done
    frameworks_json+="]"

    # Generate project.json
    cat > "$ANALYSIS_DIR/project.json" << EOF
{
    "project_summary": {
        "name": "$project_name",
        "type": "$project_type",
        "confidence": "$confidence",
        "description": "$overview",
        "purpose": "AI-optimized development project",
        "complexity": "$complexity",
        "classification": "$project_classification",
        "total_files": $total_files,
        "analysis_timestamp": "$(date -Iseconds)"
    },
    "frameworks_detected": $frameworks_list,
    "ai_insights": $ai_insights,
    "technology_stack": {
        "primary_languages": ["${all_languages//,/\",\"}"],
        "frameworks": $frameworks_json,
        "development_tools": ["${development_tools//,/\",\"}"],
        "build_systems": ["${build_systems//,/\",\"}"]
    },
    "structure": {
        "files": {
            "total": $total_files,
            "code": $sc_count,
            "documentation": $doc_count,
            "configuration": $cfg_count
        }
    },
    "framework_version": "$FRAMEWORK_VERSION",
    "analysis_files": {
        "project_structure": "project.structure.json",
        "context_file": "../context.md",
        "overview": "../memory/memory.md",
        "generated_by": "Plaesy Spec-Kit v$FRAMEWORK_VERSION"
    }
}
EOF

    log_success "Comprehensive project.json generated"
}

# Function to generate project structure JSON
generate_project_structure_json() {
    log_info "Generating detailed project.structure.json..."

    # Generate directories structure
    local directories_json=""
    local key_files_json=""

    # Get all directories (excluding hidden and system dirs)
    local dirs=$(find "$PROJECT_PATH" -type d ! -path "*/.*/*" ! -name ".*" ! -path "*/node_modules/*" | sort)

    # One pass over every file to map directory -> direct file count.
    # Replaces one `find | wc -l` spawn per directory (~200ms each on MSYS).
    # Loaded into an associative array for O(1) lookup below instead of
    # re-scanning the whole list per directory (was O(dirs * unique_dirs)).
    local -A file_counts_map
    while IFS=$'\t' read -r count_dir count; do
        [[ -n "$count_dir" ]] && file_counts_map["$count_dir"]=$count
    done < <(find "$PROJECT_PATH" -type f ! -path "*/.*/*" ! -path "*/node_modules/*" -printf '%h\n' 2>/dev/null | awk '{ c[$0]++ } END { for (d in c) print d "\t" c[d] }')

    # Build directories JSON
    directories_json="{"
    local first_dir=true
    while IFS= read -r dir; do
        local relative_path="${dir#$PROJECT_PATH/}"
        if [[ "$relative_path" != "." ]] && [[ "$relative_path" != "$PROJECT_PATH" ]]; then
            if [[ "$first_dir" = true ]]; then
                first_dir=false
            else
                directories_json+=","
            fi
            local file_count=${file_counts_map["$dir"]:-0}
            generate_directory_description "$relative_path"
            directories_json+="
        \"$relative_path\": {
            \"file_count\": $file_count,
            \"description\": \"$GDD_OUT\"
        }"
        fi
    done <<< "$dirs"
    directories_json+="
    }"

    # Get key files: one find pass with size via -printf (replaces per-file
    # `stat` spawns, ~200ms each on MSYS).
    local key_files
    key_files=$(find "$PROJECT_PATH" -type f \( \
        -name "package.json" -o -name "pubspec.yaml" -o -name "requirements.txt" -o -name "go.mod" -o -name "Cargo.toml" -o \
        -name "*.md" -o -name "README*" -o -name "LICENSE" -o \
        -name "*.json" -o -name "*.yaml" -o -name "*.yml" -o \
        -name "main.dart" -o -name "main.js" -o -name "main.py" -o -name "main.go" -o -name "index.js" \
        \) ! -path "*/.*/*" ! -path "*/node_modules/*" -printf '%s\t%p\n' 2>/dev/null | sort -t$'\t' -k2)

    # Build key files JSON
    key_files_json="{"
    local first_file=true
    while IFS=$'\t' read -r file_size file; do
        local relative_path="${file#$PROJECT_PATH/}"
        if [[ "$first_file" = true ]]; then
            first_file=false
        else
            key_files_json+=","
        fi
        get_file_type "$relative_path"
        get_file_purpose "$relative_path"
        key_files_json+="
        \"$relative_path\": {
            \"size_bytes\": $file_size,
            \"type\": \"$GFT_OUT\",
            \"purpose\": \"$GFP_OUT\"
        }"
    done <<< "$key_files"
    key_files_json+="
    }"

    # Count file types: one find + awk instead of three separate finds.
    local sc_count doc_count cfg_count
    sc_count=0; doc_count=0; cfg_count=0
    while IFS= read -r fname; do
        case "$fname" in
            *.js|*.jsx|*.ts|*.tsx|*.py|*.go|*.java|*.dart) sc_count=$((sc_count + 1)) ;;
            *.md|*.txt) doc_count=$((doc_count + 1)) ;;
            *.json|*.yaml|*.yml|*.toml|*.ini) cfg_count=$((cfg_count + 1)) ;;
        esac
    done < <(find "$PROJECT_PATH" -type f ! -path "*/.*/*" ! -path "*/node_modules/*" -printf '%f\n' 2>/dev/null)

    cat > "$ANALYSIS_DIR/project.structure.json" << EOF
{
    "directories": $directories_json,
    "key_files": $key_files_json,
    "file_types": {
        "source_code": $sc_count,
        "documentation": $doc_count,
        "configuration": $cfg_count
    },
    "analysis_timestamp": "$(date -Iseconds)",
    "related_files": {
        "project_summary": "project.json",
        "context_file": "../context.md",
        "overview": "../memory.md"
    },
    "generated_by": "Plaesy Spec-Kit v$FRAMEWORK_VERSION"
}
EOF

    log_success "Detailed project.structure.json generated"
}

# Helper function to generate directory descriptions.
# Sets the global GDD_OUT instead of echoing, so callers avoid a subshell
# fork (each $(...) fork costs ~110-200ms on MSYS/Windows).
GDD_OUT=""
generate_directory_description() {
    local dir_path="$1"
    case "$dir_path" in
        "lib") GDD_OUT="Source code library directory" ;;
        "lib/screens"*) GDD_OUT="UI screen components" ;;
        "lib/models") GDD_OUT="Data models and entities" ;;
        "lib/services") GDD_OUT="Business logic and API services" ;;
        "lib/utils") GDD_OUT="Utility functions and helpers" ;;
        "test") GDD_OUT="Test files and unit tests" ;;
        "android") GDD_OUT="Android platform specific code" ;;
        "ios") GDD_OUT="iOS platform specific code" ;;
        "assets") GDD_OUT="Static assets (images, fonts, etc.)" ;;
        "assets/images") GDD_OUT="Image assets" ;;
        "assets/data") GDD_OUT="Data files" ;;
        "docs") GDD_OUT="Documentation files" ;;
        "scripts") GDD_OUT="Build and utility scripts" ;;
        *) GDD_OUT="Project directory" ;;
    esac
}

# Helper function to get file type.
# Sets the global GFT_OUT instead of echoing (see generate_directory_description).
GFT_OUT=""
get_file_type() {
    local file="$1"
    case "$file" in
        *.dart) GFT_OUT="dart" ;;
        *.js|*.jsx) GFT_OUT="javascript" ;;
        *.ts|*.tsx) GFT_OUT="typescript" ;;
        *.py) GFT_OUT="python" ;;
        *.go) GFT_OUT="go" ;;
        *.java) GFT_OUT="java" ;;
        *.json) GFT_OUT="json" ;;
        *.yaml|*.yml) GFT_OUT="yaml" ;;
        *.md) GFT_OUT="markdown" ;;
        *) GFT_OUT="text" ;;
    esac
}

# Helper function to get file purpose.
# Sets the global GFP_OUT instead of echoing (see generate_directory_description).
GFP_OUT=""
get_file_purpose() {
    local file="$1"
    local basename="${file##*/}"
    case "$basename" in
        pubspec.yaml) GFP_OUT="Flutter/Dart project configuration" ;;
        package.json) GFP_OUT="Node.js project configuration" ;;
        requirements.txt) GFP_OUT="Python dependencies" ;;
        go.mod) GFP_OUT="Go module configuration" ;;
        Cargo.toml) GFP_OUT="Rust project configuration" ;;
        README*) GFP_OUT="Project documentation" ;;
        LICENSE) GFP_OUT="Project license" ;;
        main.*|index.*) GFP_OUT="Application entry point" ;;
        *.config.*|*.config) GFP_OUT="Configuration file" ;;
        .gitignore) GFP_OUT="Git ignore rules" ;;
        .env*) GFP_OUT="Environment variables" ;;
        *) GFP_OUT="Project file" ;;
    esac
}

# Function to generate context.md
generate_context_md() {
    if [ -f "$MEMORY_DIR/context.md" ]; then
        log_info "context.md already exists, preserving existing content (not overwritten)"
        return 0
    fi

    log_info "Generating AI-optimized context.md..."

    local all_languages="${ALL_LANGUAGES_CACHE:-$(detect_all_languages)}"
    # Convert comma-separated to readable format
    local languages_readable=$(echo "$all_languages" | sed 's/,/, /g')
    local type_confidence="${PROJECT_TYPE_CACHE:-$(detect_project_type)}"
    local project_type=$(echo "$type_confidence" | cut -d: -f1)
    local project_name=$(basename "$PROJECT_PATH")

    # Generate context based on project type
    local project_desc=""
    local ai_context=""

    case "$project_type" in
        "spec-kit")
            project_desc="Plaesy Spec-Kit framework for AI-assisted development workflow automation"
            ai_context="Framework development and enhancement with advanced AI integration capabilities"
            ;;
        *)
            project_desc="This is an AI-generated project context document for development assistance"
            ai_context="General development tasks with AI-powered workflow optimization"
            ;;
    esac

    cat > "$MEMORY_DIR/context.md" << EOF
# Project Context

## Project Overview
**$project_desc**

**Project**: $project_name

## Current Task Context
$ai_context

## Technology Stack
- **Languages**: $languages_readable
- **Tools**: Git, Plaesy CLI

## Key Information
- Total files analyzed: $(find "$PROJECT_PATH" -type f ! -path "*/.*/*" | wc -l)
- Directory structure documented in .plaesy/memory/analysis/
- Development patterns and templates available in docs/
- Scripts available for automation in scripts/

## Framework Components
- Scripts: Automation and utility scripts
- Templates: AI-optimized prompt templates
- Instructions: Technology-specific guidance
- Chat Modes: AI role configurations
- Checklists: Quality assurance frameworks

## Related Analysis Files
- **Project Summary**: \`analysis/project.json\` - Complete project overview, framework detection, and AI insights
- **Project Structure**: \`analysis/project.structure.json\` - Detailed directory structure and file analysis
- **Project Overview**: \`memory/memory.md\` - Comprehensive project documentation and recommendations

## Quick Access to Project Information
- **Framework Type & Details**: See \`analysis/project.json\` → \`project_summary\` → \`type\`
- **Detected Frameworks**: See \`analysis/project.json\` → \`frameworks_detected\`
- **Technology Stack**: See \`analysis/project.json\` → \`technology_stack\`

---
*Generated by Plaesy Spec-Kit v$FRAMEWORK_VERSION*
EOF

    log_success 'AI-optimized context.md generated (max 100 lines)'
}

# Function to generate memory.md
generate_overview_md() {
    if [ -f "$MEMORY_DIR/memory.md" ]; then
        log_info "memory.md already exists, preserving existing content (not overwritten)"
        return 0
    fi

    log_info "Generating project memory.md..."

    local all_languages="${ALL_LANGUAGES_CACHE:-$(detect_all_languages)}"
    local languages_readable=$(echo "$all_languages" | sed 's/,/, /g')
    local type_confidence="${PROJECT_TYPE_CACHE:-$(detect_project_type)}"
    local project_type=$(echo "$type_confidence" | cut -d: -f1)

    cat > "$MEMORY_DIR/memory.md" << EOF
# Project Overview

## Project Summary
This project has been analyzed by Plaesy Spec-Kit for AI-assisted development.

## Technology Stack
- **Languages**: $languages_readable
- **Framework**: $project_type
- **Tools**: Git

## File Structure
| File Type | Count |
|-----------|-------|
| Source Code | 0 |
| Documentation | 0 |

## Recommendations
- Add comprehensive documentation
- Implement automated testing
- Set up CI/CD pipeline

## Related Analysis Files
- **Project Summary**: \`analysis/project.json\` - Complete project overview and AI insights
- **Project Structure**: \`analysis/project.structure.json\` - Detailed file and directory analysis
- **Context File**: \`context.md\` - AI session context and task information

## Analysis Metadata
- Generated by: Plaesy Spec-Kit v$FRAMEWORK_VERSION
- Analysis timestamp: $(date -Iseconds)
- Total files analyzed: $(find "$PROJECT_PATH" -type f ! -path "*/.*/*" | wc -l)

---
*Generated by Plaesy Spec-Kit v$FRAMEWORK_VERSION*
EOF

    log_success 'Comprehensive memory.md generated (max 300 lines)'
}

# Function to generate project scripts
generate_project_scripts() {
    log_info "Generating project scripts..."

    cat > "$SCRIPTS_DIR/test-runner.sh" << 'EOF'
#!/bin/bash
echo "Running tests..."
if [[ -f "package.json" ]]; then
    npm test
elif [[ -f "requirements.txt" ]]; then
    python -m pytest
elif [[ -f "go.mod" ]]; then
    go test ./...
else
    echo "No test framework detected"
fi
echo "Test execution completed"
EOF

    chmod +x "$SCRIPTS_DIR/test-runner.sh"
    log_success "Project scripts generated"
}

# Main execution function
main() {
    log_info "Starting comprehensive project analysis..."
    log_info "Project path: $PROJECT_PATH"
    log_info "Analysis directory: $ANALYSIS_DIR"

    # Run comprehensive analysis functions
    # Detect languages & project type once; reuse across generators (each
    # call spawns find/awk ~0.3-0.5s on MSYS/Windows).
    ALL_LANGUAGES_CACHE=$(detect_all_languages)
    PROJECT_TYPE_CACHE=$(detect_project_type)
    FRAMEWORKS_CACHE=$(detect_all_frameworks)

    generate_project_json
    generate_project_structure_json
    generate_context_md
    generate_overview_md

    # Dependency graph (part of analyze output)
    if [[ "$BUILD_GRAPH" -eq 1 ]]; then
        if [[ "$FORCE_GRAPH" -eq 1 ]]; then
            log_info "Building dependency graph (forced)..."
            bash "$SCRIPT_DIR/plaesy-graph.sh" --path "$PROJECT_PATH" || log_info "Graph build skipped/failed; summary files remain valid."
        else
            log_info "Building dependency graph (incremental: only rebuilds if source changed)..."
            bash "$SCRIPT_DIR/plaesy-graph.sh" --path "$PROJECT_PATH" --if-changed || log_info "Graph build skipped/failed; summary files remain valid."
        fi
    fi

    log_success 'Comprehensive analysis completed!'
    log_info 'Generated files:'
    log_info 'Analysis files (in '$ANALYSIS_DIR'):'
    log_info '   - project.json - AI-optimized project summary'
    log_info '   - project.structure.json - Detailed project structure'
    if [[ "$BUILD_GRAPH" -eq 1 ]]; then
        log_info '   - project.graph.json - Dependency graph (nodes + edges)'
        log_info '   - project.html - Interactive graph visualization'
        log_info '   - reports.md - Graph report (communities, god nodes, orphans)'
    fi
    log_info 'Memory files (in '$MEMORY_DIR'):'
    log_info '   - context.md - AI session context (max 100 lines)'
    log_info '   - memory.md - Project overview (max 300 lines)'
}

# Run main function
main "$@"