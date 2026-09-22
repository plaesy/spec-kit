# Plaesy AI-Optimized Project Analyzer - Enhanced Version (PowerShell)
# Comprehensive project analysis with AI-friendly documentation generation
# Usage: ./plaesy-analyze.ps1 [project_path] [-NoGraph] [-Force] [-IfChanged]
#   -NoGraph  skip dependency graph build entirely
#   -Force    force a full graph rebuild even if no source files changed
#             (default: graph rebuild is skipped when nothing changed since
#             the last run)
#   -IfChanged  skip all analysis regeneration if project fingerprint matches
#               the last run (file count + newest mtime + framework version)

param(
    [string]$ProjectPath = ".",
    [switch]$NoGraph,
    [switch]$Force,
    [switch]$IfChanged
)

# Configuration
$AnalysisDir = Join-Path $ProjectPath ".plaesy/analysis"
$MemoryDir = Join-Path $ProjectPath ".plaesy/memory"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$FrameworkVersion = if (Test-Path (Join-Path $ScriptDir ".." ".." "VERSION")) {
    (Get-Content (Join-Path $ScriptDir ".." ".." "VERSION")).Trim()
} else {
    "0.0.1"
}

# Create required directories (only analysis and memory, not unused scripts dir)
New-Item -ItemType Directory -Force -Path $AnalysisDir | Out-Null
New-Item -ItemType Directory -Force -Path $MemoryDir | Out-Null

# Cached detection results (computed once in Main, reused by all generators).
# Order matters: AllFrameworksCache must be set before ProjectTypeCache
# (Get-ProjectType reads the cached frameworks string when present).
$script:AllFrameworksCache = $null
$script:ProjectTypeCache = $null
$script:AllLanguagesCache = $null
$script:FileTypeCounts = $null

# Logging functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

# Function to detect all frameworks in the project
# Cached result is stored in $script:AllFrameworksCache so downstream
# generators reuse a single detection pass instead of re-scanning.
function Get-AllFrameworks {
    if ($script:AllFrameworksCache -ne $null) { return $script:AllFrameworksCache }
    $frameworks = @()
    $confidences = @()

    # JavaScript/TypeScript frameworks
    if (Test-Path (Join-Path $ProjectPath "package.json")) {
        $packageContent = Get-Content (Join-Path $ProjectPath "package.json") -Raw
        if ($packageContent -match '"next"') {
            $frameworks += "nextjs"
            $confidences += "high"
        }
        if ($packageContent -match '"react"') {
            $frameworks += "react"
            $confidences += "high"
        }
        if ($packageContent -match '"vue"') {
            $frameworks += "vue"
            $confidences += "high"
        }
        if ($packageContent -match '"svelte"') {
            $frameworks += "svelte"
            $confidences += "high"
        }
        if ($packageContent -match '"astro"') {
            $frameworks += "astro"
            $confidences += "high"
        }
        if ($packageContent -match '"remix"') {
            $frameworks += "remix"
            $confidences += "high"
        }
        if ($packageContent -match '"gatsby"') {
            $frameworks += "gatsby"
            $confidences += "high"
        }
        if ($packageContent -match '"nuxt"') {
            $frameworks += "nuxt"
            $confidences += "high"
        }
        if ($packageContent -match '"express"') {
            $frameworks += "express"
            $confidences += "high"
        }
        if ($packageContent -match '"nestjs"') {
            $frameworks += "nestjs"
            $confidences += "high"
        }
        if ($packageContent -match '"angular"') {
            $frameworks += "angular"
            $confidences += "high"
        }
        if ($frameworks.Count -eq 0) {
            $frameworks += "nodejs"
            $confidences += "medium"
        }
    }
  
    # Check for additional JS/TS config files
    if ((Test-Path (Join-Path $ProjectPath "next.config.js")) -or (Test-Path (Join-Path $ProjectPath "next.config.mjs"))) {
        if ("nextjs" -notin $frameworks) {
            $frameworks += "nextjs"
            $confidences += "high"
        }
    }
    if ((Test-Path (Join-Path $ProjectPath "nuxt.config.js")) -or (Test-Path (Join-Path $ProjectPath "nuxt.config.ts"))) {
        if ("nuxt" -notin $frameworks) {
            $frameworks += "nuxt"
            $confidences += "high"
        }
    }
    if (Test-Path (Join-Path $ProjectPath "svelte.config.js")) {
        if ("svelte" -notin $frameworks) {
            $frameworks += "svelte"
            $confidences += "high"
        }
    }
    if ((Test-Path (Join-Path $ProjectPath "astro.config.mjs")) -or (Test-Path (Join-Path $ProjectPath "astro.config.ts"))) {
        if ("astro" -notin $frameworks) {
            $frameworks += "astro"
            $confidences += "high"
        }
    }
    if (Test-Path (Join-Path $ProjectPath "remix.config.js")) {
        if ("remix" -notin $frameworks) {
            $frameworks += "remix"
            $confidences += "high"
        }
    }
    if (Test-Path (Join-Path $ProjectPath "gatsby-config.js")) {
        if ("gatsby" -notin $frameworks) {
            $frameworks += "gatsby"
            $confidences += "high"
        }
    }
    if ((Test-Path (Join-Path $ProjectPath "vite.config.js")) -or (Test-Path (Join-Path $ProjectPath "vite.config.ts"))) {
        if ("vite" -notin $frameworks) {
            $frameworks += "vite"
            $confidences += "medium"
        }
    }

    # Python frameworks
    if (Test-Path (Join-Path $ProjectPath "pyproject.toml")) {
        $pyprojectContent = Get-Content (Join-Path $ProjectPath "pyproject.toml") -Raw
        if ($pyprojectContent -match "django") {
            $frameworks += "django"
            $confidences += "high"
        }
        if ($pyprojectContent -match "fastapi") {
            $frameworks += "fastapi"
            $confidences += "high"
        }
        if ($pyprojectContent -match "poetry") {
            $frameworks += "poetry"
            $confidences += "medium"
        }
        if ("python" -notin $frameworks -and "django" -notin $frameworks -and "fastapi" -notin $frameworks -and "poetry" -notin $frameworks) {
            $frameworks += "python"
            $confidences += "medium"
        }
    }
    if (Test-Path (Join-Path $ProjectPath "requirements.txt")) {
        $reqContent = Get-Content (Join-Path $ProjectPath "requirements.txt") -Raw
        if ($reqContent -match "django") {
            $frameworks += "django"
            $confidences += "high"
        }
        if ($reqContent -match "fastapi") {
            $frameworks += "fastapi"
            $confidences += "high"
        }
        if ($reqContent -match "flask") {
            $frameworks += "flask"
            $confidences += "high"
        }
        if ("python" -notin $frameworks -and "django" -notin $frameworks -and "fastapi" -notin $frameworks -and "flask" -notin $frameworks) {
            $frameworks += "python"
            $confidences += "medium"
        }
    }
    if (Test-Path (Join-Path $ProjectPath "Pipfile")) {
        if ("pipenv" -notin $frameworks) {
            $frameworks += "pipenv"
            $confidences += "high"
        }
    }
    if (Test-Path (Join-Path $ProjectPath "manage.py")) {
        if ("django" -notin $frameworks) {
            $frameworks += "django"
            $confidences += "high"
        }
    }

    # Mobile frameworks
    if (Test-Path (Join-Path $ProjectPath "pubspec.yaml")) {
        if ("flutter" -notin $frameworks) {
            $frameworks += "flutter"
            $confidences += "high"
        }
    }

    # Go frameworks
    if (Test-Path (Join-Path $ProjectPath "go.mod")) {
        $goModContent = Get-Content (Join-Path $ProjectPath "go.mod") -Raw
        if ($goModContent -match "gin-gonic") {
            $frameworks += "gin"
            $confidences += "high"
        }
        if ($goModContent -match "labstack/echo") {
            $frameworks += "echo"
            $confidences += "high"
        }
        if ($goModContent -match "gofiber") {
            $frameworks += "fiber"
            $confidences += "high"
        }
        if ("go" -notin $frameworks -and "gin" -notin $frameworks -and "echo" -notin $frameworks -and "fiber" -notin $frameworks) {
            $frameworks += "go"
            $confidences += "medium"
        }
    }

    # Rust frameworks
    if (Test-Path (Join-Path $ProjectPath "Cargo.toml")) {
        $cargoContent = Get-Content (Join-Path $ProjectPath "Cargo.toml") -Raw
        if ($cargoContent -match "actix-web") {
            $frameworks += "actix"
            $confidences += "high"
        }
        if ($cargoContent -match "rocket") {
            $frameworks += "rocket"
            $confidences += "high"
        }
        if ($cargoContent -match "axum") {
            $frameworks += "axum"
            $confidences += "high"
        }
        if ("rust" -notin $frameworks -and "actix" -notin $frameworks -and "rocket" -notin $frameworks -and "axum" -notin $frameworks) {
            $frameworks += "rust"
            $confidences += "medium"
        }
    }

    # Java frameworks
    if (Test-Path (Join-Path $ProjectPath "pom.xml")) {
        $pomContent = Get-Content (Join-Path $ProjectPath "pom.xml") -Raw
        if ($pomContent -match "spring-boot") {
            $frameworks += "springboot"
            $confidences += "high"
        }
        if ($pomContent -match "spring") {
            $frameworks += "spring"
            $confidences += "high"
        }
        if ("maven" -notin $frameworks -and "springboot" -notin $frameworks -and "spring" -notin $frameworks) {
            $frameworks += "maven"
            $confidences += "medium"
        }
    }
    if ((Test-Path (Join-Path $ProjectPath "build.gradle")) -or (Test-Path (Join-Path $ProjectPath "build.gradle.kts"))) {
        $gradleFiles = Get-ChildItem -Path $ProjectPath -Filter "build.gradle*"
        $hasSpringBoot = $false
        $hasKtor = $false

        foreach ($file in $gradleFiles) {
            $content = Get-Content $file.FullName -Raw
            if ($content -match "org.springframework.boot") {
                $hasSpringBoot = $true
            }
            if ($content -match "io.ktor") {
                $hasKtor = $true
            }
        }

        if ($hasSpringBoot) {
            $frameworks += "springboot"
            $confidences += "high"
        }
        if ($hasKtor) {
            $frameworks += "ktor"
            $confidences += "high"
        }
        if ("gradle" -notin $frameworks -and "springboot" -notin $frameworks -and "ktor" -notin $frameworks) {
            $frameworks += "gradle"
            $confidences += "medium"
        }
    }

    # Ruby frameworks
    if (Test-Path (Join-Path $ProjectPath "Gemfile")) {
        $gemfileContent = Get-Content (Join-Path $ProjectPath "Gemfile") -Raw
        if ($gemfileContent -match "rails") {
            $frameworks += "rails"
            $confidences += "high"
        }
        if ($gemfileContent -match "sinatra") {
            $frameworks += "sinatra"
            $confidences += "high"
        }
        if ("ruby" -notin $frameworks -and "rails" -notin $frameworks -and "sinatra" -notin $frameworks) {
            $frameworks += "ruby"
            $confidences += "medium"
        }
    }
    if (Test-Path (Join-Path $ProjectPath "config/application.rb")) {
        if ("rails" -notin $frameworks) {
            $frameworks += "rails"
            $confidences += "high"
        }
    }

    # PHP frameworks
    if (Test-Path (Join-Path $ProjectPath "composer.json")) {
        $composerContent = Get-Content (Join-Path $ProjectPath "composer.json") -Raw
        if ($composerContent -match "laravel/framework") {
            $frameworks += "laravel"
            $confidences += "high"
        }
        if ($composerContent -match "symfony") {
            $frameworks += "symfony"
            $confidences += "high"
        }
        if ("php" -notin $frameworks -and "laravel" -notin $frameworks -and "symfony" -notin $frameworks) {
            $frameworks += "php"
            $confidences += "medium"
        }
    }
    if (Test-Path (Join-Path $ProjectPath "wp-config.php")) {
        if ("wordpress" -notin $frameworks) {
            $frameworks += "wordpress"
            $confidences += "high"
        }
    }

    # C#/.NET frameworks
    if (Get-ChildItem -Path $ProjectPath -Filter "*.csproj" -ErrorAction SilentlyContinue) {
        $csprojFiles = Get-ChildItem -Path $ProjectPath -Filter "*.csproj"
        $hasAspNetCore = $false

        foreach ($file in $csprojFiles) {
            $content = Get-Content $file.FullName -Raw
            if ($content -match "Microsoft.AspNetCore") {
                $hasAspNetCore = $true
                break
            }
        }

        if ($hasAspNetCore) {
            $frameworks += "aspnet"
            $confidences += "high"
        }
        if ("dotnet" -notin $frameworks -and "aspnet" -notin $frameworks) {
            $frameworks += "dotnet"
            $confidences += "medium"
        }
    }

    # Swift frameworks
    if (Test-Path (Join-Path $ProjectPath "Package.swift")) {
        if ("swift" -notin $frameworks) {
            $frameworks += "swift"
            $confidences += "high"
        }
    }

    # Container/DevOps
    if (Test-Path (Join-Path $ProjectPath "Dockerfile")) {
        if ("docker" -notin $frameworks) {
            $frameworks += "docker"
            $confidences += "medium"
        }
    }
    if ((Test-Path (Join-Path $ProjectPath "docker-compose.yml")) -or (Test-Path (Join-Path $ProjectPath "docker-compose.yaml"))) {
        if ("docker-compose" -notin $frameworks) {
            $frameworks += "docker-compose"
            $confidences += "medium"
        }
    }
    if ((Test-Path (Join-Path $ProjectPath "Terrafile")) -or (Test-Path (Join-Path $ProjectPath "main.tf"))) {
        if ("terraform" -notin $frameworks) {
            $frameworks += "terraform"
            $confidences += "high"
        }
    }

    # Configuration management
    if (Test-Path (Join-Path $ProjectPath "Vagrantfile")) {
        if ("vagrant" -notin $frameworks) {
            $frameworks += "vagrant"
            $confidences += "high"
        }
    }
    if (Test-Path (Join-Path $ProjectPath "kustomization.yaml")) {
        if ("kustomize" -notin $frameworks) {
            $frameworks += "kustomize"
            $confidences += "high"
        }
    }

    # Check for Plaesy framework itself
    if (Test-Path (Join-Path $ProjectPath "README.md")) {
        $readmeContent = Get-Content (Join-Path $ProjectPath "README.md") -Raw
        if ($readmeContent -match "Plaesy Spec-Kit") {
            $frameworks += "spec-kit"
            $confidences += "high"
        }
    }

    # If no frameworks found, default to generic
    if ($frameworks.Count -eq 0) {
        $frameworks += "generic"
        $confidences += "low"
    }

    # Return formatted string with all frameworks
    $result = ""
    for ($i = 0; $i -lt $frameworks.Count; $i++) {
        if ($i -gt 0) { $result += "," }
        $result += $frameworks[$i] + ":" + $confidences[$i]
    }
    $script:AllFrameworksCache = $result
    return $result
}

# Function to get primary framework (highest confidence)
# Uses cached AllFrameworks result when available.
function Get-ProjectType {
    $frameworksStr = Get-AllFrameworks
    $frameworkEntries = $frameworksStr -split ","

    $primaryFramework = "generic"
    $primaryConfidence = "low"

    foreach ($entry in $frameworkEntries) {
        $parts = $entry -split ":"
        $framework = $parts[0]
        $confidence = $parts[1]

        # Prioritize higher confidence
        if ($confidence -eq "high" -and $primaryConfidence -ne "high") {
            $primaryFramework = $framework
            $primaryConfidence = $confidence
        }
        elseif ($confidence -eq "high" -and $primaryConfidence -eq "high") {
            # If both high confidence, prefer front-end frameworks
            if ($framework -in @("nextjs", "react", "vue", "angular", "svelte", "astro")) {
                $primaryFramework = $framework
                $primaryConfidence = $confidence
            }
        }
    }

    return "$primaryFramework`:$primaryConfidence"
}

# Function to detect all programming languages
# Cached result is stored in $script:AllLanguagesCache so downstream
# generators reuse a single detection pass instead of re-scanning.
function Get-AllLanguages {
    if ($script:AllLanguagesCache -ne $null) { return $script:AllLanguagesCache }

    $languageCounts = @{
        js = 0
        ts = 0
        py = 0
        go = 0
        java = 0
        dart = 0
        cpp = 0
        c = 0
        h = 0
        php = 0
        rb = 0
        rs = 0
        swift = 0
        kt = 0
        scala = 0
        sh = 0
        html = 0
        css = 0
    }

    Get-ChildItem -Path $ProjectPath -Recurse -File -ErrorAction SilentlyContinue | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    } | ForEach-Object {
        switch ($_.Extension.ToLowerInvariant()) {
            ".js" { $languageCounts.js++; break }
            ".jsx" { $languageCounts.js++; break }
            ".ts" { $languageCounts.ts++; break }
            ".tsx" { $languageCounts.ts++; break }
            ".py" { $languageCounts.py++; break }
            ".go" { $languageCounts.go++; break }
            ".java" { $languageCounts.java++; break }
            ".dart" { $languageCounts.dart++; break }
            ".cpp" { $languageCounts.cpp++; break }
            ".cc" { $languageCounts.cpp++; break }
            ".cxx" { $languageCounts.cpp++; break }
            ".c" { $languageCounts.c++; break }
            ".h" { $languageCounts.h++; break }
            ".php" { $languageCounts.php++; break }
            ".rb" { $languageCounts.rb++; break }
            ".rs" { $languageCounts.rs++; break }
            ".swift" { $languageCounts.swift++; break }
            ".kt" { $languageCounts.kt++; break }
            ".scala" { $languageCounts.scala++; break }
            ".sh" { $languageCounts.sh++; break }
            ".html" { $languageCounts.html++; break }
            ".css" { $languageCounts.css++; break }
            ".scss" { $languageCounts.css++; break }
            ".sass" { $languageCounts.css++; break }
        }
    }

    $languages = @()
    if ($languageCounts.js -gt 0) { $languages += "JavaScript" }
    if ($languageCounts.ts -gt 0) { $languages += "TypeScript" }
    if ($languageCounts.py -gt 0) { $languages += "Python" }
    if ($languageCounts.go -gt 0) { $languages += "Go" }
    if ($languageCounts.java -gt 0) { $languages += "Java" }
    if ($languageCounts.dart -gt 0) { $languages += "Dart" }
    if ($languageCounts.cpp -gt 0) { $languages += "C++" }
    if ($languageCounts.c -gt 0) { $languages += "C" }
    if ($languageCounts.h -gt 0) { $languages += "C/C++ Headers" }
    if ($languageCounts.php -gt 0) { $languages += "PHP" }
    if ($languageCounts.rb -gt 0) { $languages += "Ruby" }
    if ($languageCounts.rs -gt 0) { $languages += "Rust" }
    if ($languageCounts.swift -gt 0) { $languages += "Swift" }
    if ($languageCounts.kt -gt 0) { $languages += "Kotlin" }
    if ($languageCounts.scala -gt 0) { $languages += "Scala" }
    if ($languageCounts.sh -gt 0) { $languages += "Shell" }
    if ($languageCounts.html -gt 0) { $languages += "HTML" }
    if ($languageCounts.css -gt 0) { $languages += "CSS" }

    if ($languages.Count -eq 0) {
        $languages += "JavaScript"
    }

    $result = $languages -join ", "
    $script:AllLanguagesCache = $result
    return $result
}

# Function to detect primary language (most files)
function Get-PrimaryLanguage {
    $languagesStr = Get-AllLanguages
    # Return first language as primary
    return ($languagesStr -split ",")[0]
}

# Function to generate AI insights
function Get-AIInsights {
    param([string]$ProjectType, [int]$TotalFiles)

    $overview = ""
    $recommendations = @()

    switch ($ProjectType) {
        {$_ -in @("react", "nextjs")} {
            $overview = "Well-structured React project with component-based architecture"
            $recommendations += "Consider adding TypeScript for better type safety"
            $recommendations += "Add comprehensive unit tests with Jest"
        }
        "vue" {
            $overview = "Vue.js application with reactive component system"
            $recommendations += "Consider Vue 3 Composition API for better code organization"
        }
        "django" {
            $overview = "Django web application with MVC architecture"
            $recommendations += "Add API documentation with Django REST framework"
        }
        "flask" {
            $overview = "Lightweight Flask web application"
            $recommendations += "Consider adding SQLAlchemy for database management"
        }
        "go" {
            $overview = "Go application with efficient concurrency support"
            $recommendations += "Consider adding comprehensive benchmarks"
        }
        "flutter" {
            $overview = "Flutter application with cross-platform mobile development framework"
            $recommendations += "Consider adding comprehensive unit tests with flutter test"
            $recommendations += "Add effective linter configuration with flutter analyze"
            $recommendations += "Consider adding integration tests for critical user flows"
        }
        "spec-kit" {
            $overview = "Plaesy Spec-Kit framework for AI-assisted development"
            $recommendations += "Add more AI platform integrations"
        }
        default {
            $overview = "Generic project structure"
            $recommendations += "Add README.md with project documentation"
            $recommendations += "Consider adding automated testing"
        }
    }

    if ($TotalFiles -lt 10) {
        $recommendations += "Expand project with additional features and modules"
    }

    return @{
        overview = $overview
        recommendations = $recommendations
    }
}

# Function to generate AI insights for multi-framework projects
function Get-MultiFrameworkAIInsights {
    param([array]$FrameworksList, [string]$Classification, [int]$TotalFiles)

    $overview = ""
    $recommendations = @()

    switch ($Classification) {
        "full-stack" {
            $overview = "Full-stack application with both frontend and backend frameworks"
            $recommendations += "Implement consistent API contracts between frontend and backend"
            $recommendations += "Set up shared TypeScript types for better type safety"
            $recommendations += "Configure CORS properly for cross-origin requests"
            $recommendations += "Implement comprehensive error handling across the stack"
            $recommendations += "Consider using authentication middleware that works across frameworks"
        }
        "full-stack-with-infrastructure" {
            $overview = "Full-stack application with containerized infrastructure deployment"
            $recommendations += "Implement comprehensive container orchestration strategies"
            $recommendations += "Set up environment-specific configuration management"
            $recommendations += "Implement health checks for all services"
            $recommendations += "Configure logging and monitoring across the stack"
            $recommendations += "Set up automated deployment pipelines with CI/CD"
        }
        "multi-framework" {
            $overview = "Multi-framework project with diverse technology stack"
            $recommendations += "Establish consistent coding standards across frameworks"
            $recommendations += "Implement shared testing strategies"
            $recommendations += "Create unified build and deployment processes"
            $recommendations += "Document inter-framework communication patterns"
        }
        default {
            $overview = "Multi-technology project"
            $recommendations += "Ensure proper integration between different technologies"
        }
    }

    # Add framework-specific recommendations
    $hasFrontend = $false
    $hasBackend = $false
    $hasMobile = $false
    $hasContainer = $false

    foreach ($fw in $FrameworksList) {
        if ($fw.framework -in @("react", "vue", "angular", "svelte", "nextjs", "nuxt", "astro", "gatsby")) {
            $hasFrontend = $true
            $recommendations += "Consider implementing component library for UI consistency"
        }
        if ($fw.framework -in @("django", "flask", "fastapi", "express", "nestjs", "springboot", "rails", "laravel", "actix", "rocket", "axum")) {
            $hasBackend = $true
            $recommendations += "Implement comprehensive API documentation (OpenAPI/Swagger)"
        }
        if ($fw.framework -eq "flutter") {
            $hasMobile = $true
            $recommendations += "Consider using responsive design patterns for mobile compatibility"
        }
        if ($fw.framework -in @("docker", "docker-compose")) {
            $hasContainer = $true
            $recommendations += "Optimize container images for production deployment"
        }
    }

    if ($hasFrontend -and $hasBackend) {
        $recommendations += "Implement comprehensive integration tests between frontend and backend"
    }

    if ($TotalFiles -lt 20) {
        $recommendations += "Expand project with additional features and modules"
    }

    # Remove duplicate recommendations
    $recommendations = $recommendations | Sort-Object -Unique

    return @{
        overview = $overview
        recommendations = $recommendations
        multi_framework_details = @{
            classification = $Classification
            frontend_detected = $hasFrontend
            backend_detected = $hasBackend
            mobile_detected = $hasMobile
            container_detected = $hasContainer
            total_frameworks = $FrameworksList.Count
        }
    }
}

# Function to count file types across project (single pass, cached result)
# Returns hashtable with: @{ code=<count>, doc=<count>, config=<count> }
function Get-FileTypeCounts {
    if ($script:FileTypeCounts -ne $null) { return $script:FileTypeCounts }

    $codeFiles = 0
    $docFiles = 0
    $configFiles = 0

    Get-ChildItem -Path $ProjectPath -Recurse -File | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    } | ForEach-Object {
        $ext = $_.Extension.ToLower()

        if ($ext -in @('.js', '.jsx', '.ts', '.tsx', '.py', '.go', '.java', '.dart', '.sh', '.bash', '.ps1', '.psm1', '.rb', '.php', '.rs', '.c', '.cc', '.cpp', '.h', '.hpp', '.cs', '.kt', '.kts', '.swift')) {
            $codeFiles++
        }
        elseif ($ext -in @('.md', '.txt', '.rst', '.adoc')) {
            $docFiles++
        }
        elseif ($ext -in @('.json', '.yaml', '.yml', '.toml', '.ini', '.xml', '.cfg')) {
            $configFiles++
        }
    }

    $script:FileTypeCounts = @{ code = $codeFiles; doc = $docFiles; config = $configFiles }
    return $script:FileTypeCounts
}

# Function to detect development tools (mirrors bash detect_development_tools)
function Get-DevelopmentTools {
    $tools = @()

    # Version control
    if (Test-Path (Join-Path $ProjectPath ".git")) { $tools += "Git" }
    if (Test-Path (Join-Path $ProjectPath ".svn")) { $tools += "Subversion" }
    if (Test-Path (Join-Path $ProjectPath ".hg")) { $tools += "Mercurial" }

    # Package managers
    if (Test-Path (Join-Path $ProjectPath "package.json")) { $tools += "npm/yarn/pnpm" }
    if (Test-Path (Join-Path $ProjectPath "requirements.txt")) { $tools += "pip/poetry" }
    if (Test-Path (Join-Path $ProjectPath "go.mod")) { $tools += "Go Modules" }
    if (Test-Path (Join-Path $ProjectPath "Cargo.toml")) { $tools += "Cargo" }
    if (Test-Path (Join-Path $ProjectPath "pom.xml")) { $tools += "Maven/Gradle" }
    if (Test-Path (Join-Path $ProjectPath "Gemfile")) { $tools += "Bundler" }
    if (Test-Path (Join-Path $ProjectPath "composer.json")) { $tools += "Composer" }
    if (Test-Path (Join-Path $ProjectPath "pubspec.yaml")) { $tools += "Pub" }

    # CI/CD tools
    if ((Test-Path (Join-Path $ProjectPath ".github/workflows")) -or (Test-Path (Join-Path $ProjectPath ".github/workflows"))) { $tools += "GitHub Actions" }
    if (Test-Path (Join-Path $ProjectPath ".gitlab-ci.yml")) { $tools += "GitLab CI" }
    if (Test-Path (Join-Path $ProjectPath "Jenkinsfile")) { $tools += "Jenkins" }
    if (Test-Path (Join-Path $ProjectPath "azure-pipelines.yml")) { $tools += "Azure Pipelines" }

    # Testing frameworks
    if (Test-Path (Join-Path $ProjectPath "jest.config.js")) { $tools += "Jest" }
    if (Test-Path (Join-Path $ProjectPath "vitest.config.js")) { $tools += "Vitest" }
    if (Test-Path (Join-Path $ProjectPath "pytest.ini")) { $tools += "pytest" }

    # Linting and formatting
    if (Test-Path (Join-Path $ProjectPath ".eslintrc.js")) { $tools += "ESLint" }
    if (Test-Path (Join-Path $ProjectPath ".prettierrc")) { $tools += "Prettier" }

    # Docker and containerization
    if (Test-Path (Join-Path $ProjectPath "Dockerfile")) { $tools += "Docker" }

    # If no tools found, default to basic
    if ($tools.Count -eq 0) { $tools += "Manual" }

    return $tools
}

# Function to detect build systems (mirrors bash detect_build_systems)
function Get-BuildSystems {
    $systems = @()

    # JavaScript/TypeScript build tools
    if (Test-Path (Join-Path $ProjectPath "package.json")) {
        if (Test-Path (Join-Path $ProjectPath "webpack.config.js")) { $systems += "Webpack" }
        if (Test-Path (Join-Path $ProjectPath "vite.config.js")) { $systems += "Vite" }
        if (Test-Path (Join-Path $ProjectPath "rollup.config.js")) { $systems += "Rollup" }
        if (Test-Path (Join-Path $ProjectPath "esbuild.js")) { $systems += "esbuild" }
        if (Test-Path (Join-Path $ProjectPath "turbo.json")) { $systems += "Turbopack" }
    }

    # Python build systems
    if (Test-Path (Join-Path $ProjectPath "pyproject.toml")) { $systems += "Poetry" }
    if (Test-Path (Join-Path $ProjectPath "setup.py")) { $systems += "setuptools" }
    if (Test-Path (Join-Path $ProjectPath "Makefile")) { $systems += "Make" }

    # Java build systems
    if (Test-Path (Join-Path $ProjectPath "pom.xml")) { $systems += "Maven" }
    if (Test-Path (Join-Path $ProjectPath "build.gradle")) { $systems += "Gradle" }
    if (Test-Path (Join-Path $ProjectPath "build.xml")) { $systems += "Ant" }

    # Go build systems
    if (Test-Path (Join-Path $ProjectPath "go.mod")) { $systems += "Go Modules" }

    # Rust build systems
    if (Test-Path (Join-Path $ProjectPath "Cargo.toml")) { $systems += "Cargo" }

    # C/C++ build systems
    if (Test-Path (Join-Path $ProjectPath "CMakeLists.txt")) { $systems += "CMake" }
    if (Test-Path (Join-Path $ProjectPath "Makefile")) { $systems += "Make" }
    if (Test-Path (Join-Path $ProjectPath "meson.build")) { $systems += "Meson" }

    # Ruby build systems
    if (Test-Path (Join-Path $ProjectPath "Gemfile")) { $systems += "Bundler" }
    if (Test-Path (Join-Path $ProjectPath "Rakefile")) { $systems += "Rake" }

    # PHP build systems
    if (Test-Path (Join-Path $ProjectPath "composer.json")) { $systems += "Composer" }

    # Mobile build systems
    if (Test-Path (Join-Path $ProjectPath "pubspec.yaml")) { $systems += "Pub" }

    # Container and infrastructure
    if (Test-Path (Join-Path $ProjectPath "Dockerfile")) { $systems += "Docker" }
    if (Test-Path (Join-Path $ProjectPath "docker-compose.yml")) { $systems += "Docker Compose" }
    if (Test-Path (Join-Path $ProjectPath "main.tf")) { $systems += "Terraform" }

    # If no build systems found, default to Manual
    if ($systems.Count -eq 0) { $systems += "Manual" }

    return $systems
}

# Function to generate comprehensive project.json
function New-ProjectJson {
    Write-Info "Generating comprehensive project.json..."

    $typeConfidence = Get-ProjectType
    $projectType = $typeConfidence.Split(':')[0]
    $confidence = $typeConfidence.Split(':')[1]
    $totalFiles = (Get-ChildItem -Path $ProjectPath -Recurse -File | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $projectName = Split-Path -Leaf (Resolve-Path $ProjectPath)
    $primaryLanguage = Get-PrimaryLanguage
    $allLanguages = Get-AllLanguages

    # Get all frameworks detected
    $frameworksStr = Get-AllFrameworks
    $frameworkEntries = $frameworksStr -split ","
    $frameworksList = @()
    foreach ($entry in $frameworkEntries) {
        $parts = $entry -split ":"
        $frameworksList += @{
            framework = $parts[0]
            confidence = $parts[1]
        }
    }

    # Determine project classification based on frameworks
    $classification = "single"
    if ($frameworksList.Count -gt 1) {
        $hasFrontend = $false
        $hasBackend = $false
        $hasDatabase = $false
        $hasInfrastructure = $false

        foreach ($fw in $frameworksList) {
            if ($fw.framework -in @("react", "vue", "angular", "svelte", "nextjs", "nuxt", "astro", "gatsby")) {
                $hasFrontend = $true
            }
            if ($fw.framework -in @("django", "flask", "fastapi", "express", "nestjs", "springboot", "rails", "laravel", "actix", "rocket", "axum")) {
                $hasBackend = $true
            }
            if ($fw.framework -in @("django", "flask", "fastapi", "springboot", "rails", "laravel")) {
                $hasDatabase = $true
            }
            if ($fw.framework -in @("docker", "docker-compose", "kubernetes", "terraform", "kustomize")) {
                $hasInfrastructure = $true
            }
        }

        if ($hasFrontend -and $hasBackend) {
            if ($hasInfrastructure) {
                $classification = "full-stack-with-infrastructure"
            } else {
                $classification = "full-stack"
            }
        } elseif ($hasFrontend -or $hasBackend) {
            $classification = "multi-framework"
        }
    }

    # Use multi-framework AI insights if multiple frameworks detected
    if ($classification -ne "single") {
        $aiInsights = Get-MultiFrameworkAIInsights -FrameworksList $frameworksList -Classification $classification -TotalFiles $totalFiles
    } else {
        $aiInsights = Get-AIInsights -ProjectType $projectType -TotalFiles $totalFiles
    }
    $overview = $aiInsights.overview

    # Determine complexity
    $complexity = "Small"
    if ($totalFiles -gt 50) {
        $complexity = "Large"
    }
    elseif ($totalFiles -gt 20) {
        $complexity = "Medium"
    }

    # Use cached file type counts (single-pass, computed once in Main)
    $fileTypeCounts = Get-FileTypeCounts
    $codeFiles = $fileTypeCounts.code
    $docFiles = $fileTypeCounts.doc
    $configFiles = $fileTypeCounts.config

    # Build frameworks array for technology stack
    $frameworksArray = @()
    foreach ($fw in $frameworksList) {
        $frameworksArray += $fw.framework
    }

    # Detect development tools and build systems (mirrors bash detect_development_tools/detect_build_systems)
    $devTools = Get-DevelopmentTools
    $buildSystems = Get-BuildSystems

    # Generate project.json
    $projectJson = @{
        project_summary = @{
            name = $projectName
            type = $projectType
            confidence = $confidence
            description = $overview
            purpose = "AI-optimized development project"
            complexity = $complexity
            classification = $classification
            total_files = $totalFiles
            analysis_timestamp = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
        }
        frameworks_detected = $frameworksList
        ai_insights = $aiInsights
        technology_stack = @{
            primary_languages = @($allLanguages -split ", ")
            frameworks = $frameworksArray
            development_tools = $devTools
            build_systems = $buildSystems
        }
        structure = @{
            files = @{
                total = $totalFiles
                code = $codeFiles
                documentation = $docFiles
                configuration = $configFiles
            }
        }
        framework_version = $FrameworkVersion
    }

    $jsonPath = Join-Path $AnalysisDir "project.json"
    $projectJson | ConvertTo-Json -Depth 10 | Out-File -FilePath $jsonPath -Encoding UTF8

    Write-Success "Comprehensive project.json generated"
}

# Function to generate project structure JSON
function New-ProjectStructureJson {
    Write-Info "Generating detailed project.structure.json..."

    # Get all directories (excluding hidden and system dirs)
    $dirs = Get-ChildItem -Path $ProjectPath -Recurse -Directory | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+($|[\\/])" -and $_.FullName -notmatch "node_modules"
    } | Sort-Object FullName

    # Build directories hashtable
    # One recursive file pass -> per-directory counts, instead of one
    # Get-ChildItem spawn per directory (very slow on large trees).
    $resolvedProjectPath = (Resolve-Path $ProjectPath).Path
    $fileCountsByDir = @{}
    Get-ChildItem -Path $ProjectPath -Recurse -File | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    } | ForEach-Object {
        $parent = $_.DirectoryName
        if ($fileCountsByDir.ContainsKey($parent)) {
            $fileCountsByDir[$parent]++
        } else {
            $fileCountsByDir[$parent] = 1
        }
    }

    $directories = @{}
    foreach ($dir in $dirs) {
        $relativePath = $dir.FullName.Replace($resolvedProjectPath, "").TrimStart("\", "/")
        if ($relativePath -ne "") {
            $fileCount = if ($fileCountsByDir.ContainsKey($dir.FullName)) { $fileCountsByDir[$dir.FullName] } else { 0 }
            $directories[$relativePath] = @{
                file_count = $fileCount
                description = (Get-DirectoryDescription $relativePath)
            }
        }
    }

    # Get key files
    $keyFilePatterns = @(
        "package.json", "pubspec.yaml", "requirements.txt", "go.mod", "Cargo.toml",
        "*.md", "README*", "LICENSE",
        "*.json", "*.yaml", "*.yml",
        "main.dart", "main.js", "main.py", "main.go", "index.js"
    )

    $keyFiles = @{}
    foreach ($pattern in $keyFilePatterns) {
        $files = Get-ChildItem -Path $ProjectPath -Recurse -File -Filter $pattern | Where-Object {
            $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
        }

        foreach ($file in $files) {
            $relativePath = $file.FullName.Replace($resolvedProjectPath, "").TrimStart("\", "/")
            $keyFiles[$relativePath] = @{
                size_bytes = $file.Length
                type = (Get-FileType $relativePath)
                purpose = (Get-FilePurpose $relativePath)
            }
        }
    }

    # Use cached file type counts (single-pass, computed once in Main)
    $fileTypeCounts = Get-FileTypeCounts
    $codeFiles = $fileTypeCounts.code
    $docFiles = $fileTypeCounts.doc
    $configFiles = $fileTypeCounts.config

    $structureJson = @{
        directories = $directories
        key_files = $keyFiles
        file_types = @{
            source_code = $codeFiles
            documentation = $docFiles
            configuration = $configFiles
        }
        analysis_timestamp = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    }

    $jsonPath = Join-Path $AnalysisDir "project.structure.json"
    $structureJson | ConvertTo-Json -Depth 10 | Out-File -FilePath $jsonPath -Encoding UTF8

    Write-Success "Detailed project.structure.json generated"
}

# Helper function to generate directory descriptions
function Get-DirectoryDescription {
    param([string]$DirPath)

    switch -Wildcard ($DirPath) {
        "lib" { return "Source code library directory" }
        "lib/screens*" { return "UI screen components" }
        "lib/models" { return "Data models and entities" }
        "lib/services" { return "Business logic and API services" }
        "lib/utils" { return "Utility functions and helpers" }
        "test" { return "Test files and unit tests" }
        "android" { return "Android platform specific code" }
        "ios" { return "iOS platform specific code" }
        "assets" { return "Static assets (images, fonts, etc.)" }
        "assets/images" { return "Image assets" }
        "assets/data" { return "Data files" }
        "docs" { return "Documentation files" }
        "scripts" { return "Build and utility scripts" }
        default { return "Project directory" }
    }
}

# Helper function to get file type
function Get-FileType {
    param([string]$File)

    $extension = [System.IO.Path]::GetExtension($File).ToLower()
    switch ($extension) {
        ".dart" { return "dart" }
        ".js" { return "javascript" }
        ".jsx" { return "javascript" }
        ".ts" { return "typescript" }
        ".tsx" { return "typescript" }
        ".py" { return "python" }
        ".go" { return "go" }
        ".java" { return "java" }
        ".json" { return "json" }
        ".yaml" { return "yaml" }
        ".yml" { return "yaml" }
        ".md" { return "markdown" }
        default { return "text" }
    }
}

# Helper function to get file purpose
function Get-FilePurpose {
    param([string]$File)

    $fileName = Split-Path -Leaf $File
    switch -Wildcard ($fileName) {
        "pubspec.yaml" { return "Flutter/Dart project configuration" }
        "package.json" { return "Node.js project configuration" }
        "requirements.txt" { return "Python dependencies" }
        "go.mod" { return "Go module configuration" }
        "Cargo.toml" { return "Rust project configuration" }
        "README*" { return "Project documentation" }
        "LICENSE" { return "Project license" }
        "main.*" { return "Application entry point" }
        "index.*" { return "Application entry point" }
        "*.config.*" { return "Configuration file" }
        ".gitignore" { return "Git ignore rules" }
        ".env*" { return "Environment variables" }
        default { return "Project file" }
    }
}

# Function to generate .plaesy/analysis/overview.md (always replaced, never appended)
function New-AnalysisOverviewMd {
    $overviewPath = Join-Path $AnalysisDir "overview.md"
    Write-Info "Generating analysis/overview.md (replacing previous snapshot)..."

    $allLanguages = if ($script:AllLanguagesCache -ne $null) { $script:AllLanguagesCache } else { Get-AllLanguages }
    $languagesReadable = $allLanguages
    $typeConfidence = if ($script:ProjectTypeCache -ne $null) { $script:ProjectTypeCache } else { Get-ProjectType }
    $projectType = $typeConfidence.Split(':')[0]
    $timestamp = (Get-Date).ToString("o")
    $projectName = Split-Path -Leaf (Resolve-Path $ProjectPath)
    $projectDesc = if ($projectType -eq "spec-kit") {
        "Plaesy Spec-Kit framework for AI-assisted development workflow automation"
    } else {
        "This is an AI-generated project context document for development assistance"
    }

    # Use cached file type counts (single-pass, computed once in Main)
    $fileTypeCounts = Get-FileTypeCounts
    $codeFiles = $fileTypeCounts.code
    $docFiles = $fileTypeCounts.doc
    $configFiles = $fileTypeCounts.config

    # Build recommendations from actual repo signals instead of static boilerplate.
    $recommendations = @()
    if ($docFiles -lt 3) { $recommendations += "- Add comprehensive documentation (only $docFiles doc file(s) found)" }
    $hasTestDir = Get-ChildItem -Path $ProjectPath -Recurse -Directory -ErrorAction SilentlyContinue | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules" -and $_.Name -match "^(test|tests|spec|specs|__tests__)$"
    } | Select-Object -First 1
    if (-not $hasTestDir) { $recommendations += "- Implement automated testing (no test/spec directory detected)" }
    $hasCi = (Test-Path (Join-Path $ProjectPath ".github/workflows")) -or (Test-Path (Join-Path $ProjectPath ".gitlab-ci.yml")) -or (Test-Path (Join-Path $ProjectPath ".travis.yml"))
    if (-not $hasCi) { $recommendations += "- Set up CI/CD pipeline (no CI configuration detected)" }
    if ($recommendations.Count -eq 0) { $recommendations += "- No gaps detected against baseline checks (docs, tests, CI) - keep it up." }
    $recommendationsText = $recommendations -join "`n"

    $content = @"
# Project Analysis Overview

*This file is regenerated (replaced, not appended) on every ``plaesy analyze`` run.
Manual notes belong in ``.plaesy/context.md`` / ``.plaesy/memory.md``, which analyze no longer touches.*

**Generated**: $timestamp
**Project**: $projectName
**Description**: $projectDesc

## Technology Stack
- **Languages**: $languagesReadable
- **Framework**: $projectType
- **Tools**: Git, Plaesy CLI

## File Structure
| File Type | Count |
|-----------|-------|
| Source Code | $codeFiles |
| Documentation | $docFiles |
| Config | $configFiles |

## Recommendations
$recommendationsText

## Related Analysis Files
- **Project Summary**: ``project.json`` - Complete project overview and AI insights
- **Project Structure**: ``project.structure.json`` - Detailed file and directory analysis

---
*Generated by Plaesy Spec-Kit*
"@

    $content | Out-File -FilePath $overviewPath -Encoding UTF8
    Write-Success "analysis/overview.md replaced with latest snapshot"
}

# Compute a fingerprint of the project: file count + newest mtime + framework version.
# Used by -IfChanged to skip regeneration when nothing changed since last run.
function Get-AnalyzeFingerprint {
    $extPatterns = @('*.md', '*.ps1', '*.sh', '*.js', '*.jsx', '*.ts', '*.tsx',
        '*.py', '*.go', '*.dart', '*.java', '*.kt', '*.kts', '*.swift',
        '*.c', '*.h', '*.cc', '*.cpp', '*.hpp', '*.cs', '*.rs', '*.rb', '*.php',
        '*.json', '*.yaml', '*.yml', '*.toml', '*.xml', '*.ini', '*.cfg')
    $files = Get-ChildItem -Path $ProjectPath -Recurse -File -Include $extPatterns -ErrorAction SilentlyContinue | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }
    $count = $files.Count
    $maxMtime = if ($files.Count -gt 0) { ($files | Sort-Object LastWriteTime -Descending | Select-Object -First 1).LastWriteTime.ToString('o') } else { "0" }
    return "$count|$maxMtime|$FrameworkVersion"
}

# Check if the project has changed since the last analysis run.
# Returns $true if unchanged (safe to skip), $false if changed (must regenerate).
function Test-AnalysisUnchanged {
    $fpFile = Join-Path $AnalysisDir ".analysis-fingerprint"
    $currentFp = Get-AnalyzeFingerprint
    if (Test-Path $fpFile) {
        $lastFp = (Get-Content $fpFile -Raw).Trim()
        if ($currentFp -eq $lastFp) {
            return $true
        }
    }
    # Changed or no prior fingerprint — save and return $false
    $currentFp | Out-File -FilePath $fpFile -Encoding UTF8 -NoNewline
    return $false
}

# Main execution function
function Main {
    Write-Info "Starting comprehensive project analysis..."
    Write-Info "Project path: $ProjectPath"
    Write-Info "Analysis directory: $AnalysisDir"

    if (-not (Test-Path $ProjectPath -PathType Container)) {
        Write-Host "[ERROR] Directory '$ProjectPath' does not exist" -ForegroundColor Red
        exit 1
    }

    # -IfChanged fast path: skip all regeneration if project fingerprint matches
    # the last run's fingerprint. -Force overrides this (forces full regeneration).
    if ($IfChanged -and (-not $Force) -and (Test-AnalysisUnchanged)) {
        Write-Success "Analysis unchanged since last run (-IfChanged). Skipping regeneration."
        Write-Info "Analysis files (in $AnalysisDir):"
        Write-Info "   - project.json - AI-optimized project summary (cached)"
        Write-Info "   - project.structure.json - Detailed project structure (cached)"
        Write-Info "   - overview.md - Analysis snapshot (cached)"
        if (-not $NoGraph) {
            Write-Info "   - project.graph.json - Dependency graph (cached)"
            Write-Info "   - project.html - Interactive graph visualization (cached)"
            Write-Info "   - reports.md - Graph report (cached)"
        }
        return
    }

    # Compute detection results once and reuse across all generators.
    # Order matters: AllFrameworksCache must be set before ProjectTypeCache
    # (Get-ProjectType reads the cached frameworks string when present).
    $script:AllLanguagesCache = Get-AllLanguages
    $script:AllFrameworksCache = Get-AllFrameworks
    $script:ProjectTypeCache = Get-ProjectType
    $script:FileTypeCounts = Get-FileTypeCounts

    # Run comprehensive analysis functions
    New-ProjectJson
    New-ProjectStructureJson
    New-AnalysisOverviewMd

    # Dependency graph (part of analyze output)
    if (-not $NoGraph) {
        if ($Force) {
            Write-Info "Building dependency graph (forced)..."
            try { & (Join-Path $ScriptDir "plaesy-graph.ps1") -Path $ProjectPath }
            catch { Write-Host "[WARNING] Graph build skipped/failed: $($_.Exception.Message)" -ForegroundColor Yellow }
        } else {
            Write-Info "Building dependency graph (incremental: only rebuilds if source changed)..."
            try { & (Join-Path $ScriptDir "plaesy-graph.ps1") -Path $ProjectPath -IfChanged }
            catch { Write-Host "[WARNING] Graph build skipped/failed: $($_.Exception.Message)" -ForegroundColor Yellow }
        }
    }

    Write-Success "Comprehensive analysis completed!"
    Write-Info "Generated files:"
    Write-Info "Analysis files (in $AnalysisDir):"
    Write-Info "   - project.json - AI-optimized project summary"
    Write-Info "   - project.structure.json - Detailed project structure"
    if (-not $NoGraph) {
        Write-Info "   - project.graph.json - Dependency graph (nodes + edges)"
        Write-Info "   - project.html - Interactive graph visualization"
        Write-Info "   - reports.md - Graph report (communities, god nodes, orphans)"
    }
    Write-Info "   - overview.md - Analysis snapshot (replaced every run)"
    Write-Info "No topic memory files are generated by analyze."
}

# Run main function
Main