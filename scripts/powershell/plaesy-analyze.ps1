# Plaesy AI-Optimized Project Analyzer - Enhanced Version (PowerShell)
# Comprehensive project analysis with AI-friendly documentation generation
# Usage: ./plaesy-analyze.ps1 [project_path] [-NoGraph] [-Force]
#   -NoGraph  skip dependency graph build entirely
#   -Force    force a full graph rebuild even if no source files changed
#             (default: graph rebuild is skipped when nothing changed since
#             the last run)

param(
    [string]$ProjectPath = ".",
    [switch]$NoGraph,
    [switch]$Force
)

# Configuration
$AnalysisDir = Join-Path $ProjectPath ".plaesy/memory/analysis"
$MemoryDir = Join-Path $ProjectPath ".plaesy/memory"
$ScriptsDir = Join-Path $ProjectPath "scripts"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$FrameworkVersion = if (Test-Path (Join-Path $ScriptDir "..\..\VERSION")) {
    (Get-Content (Join-Path $ScriptDir "..\..\VERSION")).Trim()
} else {
    "0.0.1"
}

# Create directories
New-Item -ItemType Directory -Force -Path $AnalysisDir | Out-Null
New-Item -ItemType Directory -Force -Path $MemoryDir | Out-Null
New-Item -ItemType Directory -Force -Path $ScriptsDir | Out-Null

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
function Get-AllFrameworks {
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
    return $result
}

# Function to get primary framework (highest confidence)
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
function Get-AllLanguages {
    $languages = @()

    # Count different language files
    $jsCount = ((Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.js" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count + (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.jsx" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count)
    $tsCount = ((Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.ts" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count + (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.tsx" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count)
    $pyCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.py" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $goCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.go" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $javaCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.java" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $dartCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.dart" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $cppCount = ((Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.cpp" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count + (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.cc" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count + (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.cxx" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count)
    $cCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.c" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $hCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.h" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $phpCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.php" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $rbCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.rb" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $rsCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.rs" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $swiftCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.swift" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $ktCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.kt" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $scalaCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.scala" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $shCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.sh" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $htmlCount = (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.html" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count
    $cssCount = ((Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.css" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count + (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.scss" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count + (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter "*.sass" | Where-Object {
        $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
    }).Count)

    # Add languages that have files
    if ($jsCount -gt 0) { $languages += "JavaScript" }
    if ($tsCount -gt 0) { $languages += "TypeScript" }
    if ($pyCount -gt 0) { $languages += "Python" }
    if ($goCount -gt 0) { $languages += "Go" }
    if ($javaCount -gt 0) { $languages += "Java" }
    if ($dartCount -gt 0) { $languages += "Dart" }
    if ($cppCount -gt 0) { $languages += "C++" }
    if ($cCount -gt 0) { $languages += "C" }
    if ($hCount -gt 0) { $languages += "C/C++ Headers" }
    if ($phpCount -gt 0) { $languages += "PHP" }
    if ($rbCount -gt 0) { $languages += "Ruby" }
    if ($rsCount -gt 0) { $languages += "Rust" }
    if ($swiftCount -gt 0) { $languages += "Swift" }
    if ($ktCount -gt 0) { $languages += "Kotlin" }
    if ($scalaCount -gt 0) { $languages += "Scala" }
    if ($shCount -gt 0) { $languages += "Shell" }
    if ($htmlCount -gt 0) { $languages += "HTML" }
    if ($cssCount -gt 0) { $languages += "CSS" }

    # If no languages found, default to JavaScript
    if ($languages.Count -eq 0) {
        $languages += "JavaScript"
    }

    # Convert array to comma-separated string
    return $languages -join ", "
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

    # Count file types
    $codePatterns = @("*.js", "*.jsx", "*.ts", "*.tsx", "*.py", "*.go", "*.java")
    $codeFiles = 0
    foreach ($pattern in $codePatterns) {
        $codeFiles += (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter $pattern | Where-Object {
            $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
        }).Count
    }

    $docPatterns = @("*.md", "*.txt")
    $docFiles = 0
    foreach ($pattern in $docPatterns) {
        $docFiles += (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter $pattern | Where-Object {
            $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
        }).Count
    }

    $configPatterns = @("*.json", "*.yaml", "*.yml", "*.toml", "*.ini")
    $configFiles = 0
    foreach ($pattern in $configPatterns) {
        $configFiles += (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter $pattern | Where-Object {
            $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
        }).Count
    }

    # Build frameworks array for technology stack
    $frameworksArray = @()
    foreach ($fw in $frameworksList) {
        $frameworksArray += $fw.framework
    }

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
            development_tools = @("Git")
            build_systems = @("Manual")
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

    # Count file types
    $codePatterns = @("*.js", "*.jsx", "*.ts", "*.tsx", "*.py", "*.go", "*.java", "*.dart")
    $codeFiles = 0
    foreach ($pattern in $codePatterns) {
        $codeFiles += (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter $pattern | Where-Object {
            $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
        }).Count
    }

    $docPatterns = @("*.md", "*.txt")
    $docFiles = 0
    foreach ($pattern in $docPatterns) {
        $docFiles += (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter $pattern | Where-Object {
            $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
        }).Count
    }

    $configPatterns = @("*.json", "*.yaml", "*.yml", "*.toml", "*.ini")
    $configFiles = 0
    foreach ($pattern in $configPatterns) {
        $configFiles += (Get-ChildItem -Path $ProjectPath -Recurse -File -Filter $pattern | Where-Object {
            $_.FullName -notmatch "[\\/]\.[^\\/]+[\\/]" -and $_.FullName -notmatch "node_modules"
        }).Count
    }

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

# Function to generate context.md
function New-ContextMd {
    $contextPath = Join-Path $MemoryDir "context.md"
    if (Test-Path $contextPath) {
        Write-Info "context.md already exists, preserving existing content (not overwritten)"
        return
    }

    Write-Info "Generating AI-optimized context.md..."

    $allLanguages = Get-AllLanguages
    # Convert comma-separated to readable format
    $languagesReadable = $allLanguages -replace ", ", ", "

    $contextContent = @"
# Project Context

## Project Overview
This is an AI-generated project context document for development assistance.

## Current Task Context
Analysis generated for AI-assisted development workflow.

## Technology Stack
- **Languages**: $languagesReadable
- **Tools**: Git
- **Framework**: Generic

## Key Information
- Total files analyzed during project scan
- Directory structure documented
- Development patterns identified

---
*Generated by Plaesy Spec-Kit*
"@

    $contextContent | Out-File -FilePath $contextPath -Encoding UTF8

    Write-Success "AI-optimized context.md generated (max 100 lines)"
}

# Function to generate overview.md
function New-OverviewMd {
    $overviewPath = Join-Path $MemoryDir "overview.md"
    if (Test-Path $overviewPath) {
        Write-Info "overview.md already exists, preserving existing content (not overwritten)"
        return
    }

    Write-Info "Generating project overview.md..."

    $allLanguages = Get-AllLanguages
    $languagesReadable = $allLanguages -replace ", ", ", "
    $typeConfidence = Get-ProjectType
    $projectType = $typeConfidence.Split(':')[0]

    $overviewContent = @"
# Project Overview

## Project Summary
This project has been analyzed by Plaesy Spec-Kit for AI-assisted development.

## Technology Stack
- **Languages**: $languagesReadable
- **Framework**: $projectType
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

---
*Generated by Plaesy Spec-Kit*
"@

    $overviewContent | Out-File -FilePath $overviewPath -Encoding UTF8

    Write-Success "Comprehensive overview.md generated (max 300 lines)"
}

# Function to generate project scripts
function New-ProjectScripts {
    Write-Info "Generating project scripts..."

    $testRunnerContent = @"
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
"@

    $testRunnerPath = Join-Path $ScriptsDir "test-runner.sh"
    $testRunnerContent | Out-File -FilePath $testRunnerPath -Encoding UTF8

    # Make it executable (on Unix systems)
    if ($IsLinux -or $IsMacOS) {
        chmod +x $testRunnerPath
    }

    Write-Success "Project scripts generated"
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

    # Run comprehensive analysis functions
    New-ProjectJson
    New-ProjectStructureJson
    New-ContextMd
    New-OverviewMd

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
    Write-Info "Memory files (in $MemoryDir):"
    Write-Info "   - context.md - AI session context (max 100 lines)"
    Write-Info "   - overview.md - Project overview (max 300 lines)"
}

# Run main function
Main