# 📊 plaesy-analyze.sh

**AI-optimized project structure analysis and documentation generation.**

**Priority:** 🔴 **CRITICAL** - Must be run first when encountering any Plaesy project

## 🎯 Purpose

Analyzes project structure and generates AI-optimized documentation for comprehensive project understanding. This script creates structured data that AI assistants use to understand project context, architecture, and development patterns.

## 🚀 KEY INFORMATION FOR AI ASSISTANTS

### ✅ **Standalone Script - NO Prerequisites Required**
- **Works independently** - No other scripts need to be run first
- **Self-contained** - Complete analysis in a single execution
- **No dependencies** - Just run and get full analysis
- **Cross-platform** - Available for both Bash and PowerShell

### ✅ **Simple Usage**
```bash
# Linux/macOS
./scripts/bash/plaesy-analyze.sh

# Windows PowerShell
./scripts/powershell/plaesy-analyze.ps1

# Or if Plaesy CLI is installed
plaesy analyze

# Skip the (slow) dependency graph build on large repos
./scripts/bash/plaesy-analyze.sh --no-graph
./scripts/powershell/plaesy-analyze.ps1 -NoGraph
```

## 📋 Usage

## 📁 Output Files

The script generates the project summary, structure, memory files **and the dependency graph** in a single run:

### 1. `project.json` - AI-Optimized Project Summary
**Purpose**: Complete project overview for AI understanding
**Contains**:
- Project type detection (50+ project types)
- Technology stack analysis
- File structure categorization
- Development workflow patterns
- AI insights and recommendations
- Project complexity assessment

### 2. `project.structure.json` - Detailed Project Structure
**Purpose**: Comprehensive file and directory analysis
**Contains**:
- Directory importance levels and purposes
- File type analysis and counting
- Key file identification and significance
- Project conventions detection
- Structural patterns

### 3. `context.md` - AI Session Context File
**Purpose**: Current session context for AI assistants (≤100 lines)
**Location**: `.plaesy/context.md`
**Contains**:
- Project overview and type
- Current task and objectives
- Technology stack summary
- Key directories and files
- Development workflow patterns

### 4. `memory.md` - Project Overview
**Purpose**: Project overview for persistent information (≤300 lines)
**Location**: `.plaesy/memory/memory.md`
**Contains**:
- Detailed project analysis
- Technology stack information
- Architecture insights
- Development patterns detected
- AI recommendations and insights

### 5. Dependency Graph (from `plaesy-graph`, built automatically)

Analyze also builds the dependency graph as part of its output. See [plaesy-graph.md](plaesy-graph.md) for details on the graph schema and how to query it.

- `project.graph.json` — nodes (`{id, label, type, group, degree, community, symbols}`) + edges (`{source, target, type, confidence}`)
- `project.html` — self-contained interactive force-directed visualization
- `reports.md` — graph report: communities, god nodes, orphan files, suggested questions

> **Note:** the graph build scans every file (references/calls/imports/mentions), so on large repos it is the slowest step. Skip it with `--no-graph` (Bash) or `-NoGraph` (PowerShell).

## 🎯 What AI Assistants Should Do

### **Standard Analysis Workflow:**
1. **Run the script**: `plaesy analyze` (standalone, no prerequisites)
2. **Read project.json**: Get project overview and insights
3. **Read project.structure.json**: Understand detailed structure
4. **Read context.md**: Get current session context (≤100 lines)
5. **Read memory.md**: Get project overview (≤300 lines)
6. **Read project.graph.json / reports.md**: Understand cross-file dependencies, communities, and high-degree files
7. **Proceed with development**: Use analysis for context-aware work

### **Quality Assurance Workflow:**
```bash
# Optional: Validate analysis quality
ls -la .plaesy/memory/analysis/  # Verify all files generated (summary + graph)
cat .plaesy/memory/analysis/project.json | jq '.project_summary.type'  # Verify project type
```

### **No Additional Steps Required:**
- ❌ Don't run prerequisite scripts
- ❌ Don't check for dependencies
- ❌ Don't run setup scripts
- ✅ Just run plaesy-analyze and start working
- ✅ Analysis is self-contained and complete

## 🔧 Key Features

### Project Type Detection
Automatically detects 50+ project types including:
- **Frontend**: React, Vue, Angular, Next.js, Nuxt.js
- **Backend**: Django, Flask, Express, FastAPI, Spring Boot
- **Mobile**: Flutter, React Native, iOS Native, Android Native
- **Infrastructure**: Docker, Kubernetes, Terraform
- **Data Science**: Jupyter, Machine Learning, Analytics
- **Enterprise**: Microservices, Full-stack, DevOps
- **Specialized**: Healthcare, Finance, Gaming, Blockchain

### AI-Optimized Analysis
- **Smart categorization**: Files grouped by purpose and importance
- **Pattern detection**: Identifies development workflows and conventions
- **Context generation**: Creates AI-readable project summaries
- **Insights extraction**: Provides actionable recommendations

### Comprehensive Coverage
- **File analysis**: All file types counted and categorized
- **Directory mapping**: Key directories identified with purpose
- **Technology detection**: Languages, frameworks, and tools identified
- **Configuration analysis**: Build systems and deployment patterns detected

## 🚀 AI Assistant Integration

### Critical Information Provided
1. **Project Context**: What type of project and its purpose
2. **Technology Stack**: Languages, frameworks, and tools used
3. **Structure Analysis**: How the project is organized
4. **Development Patterns**: Workflow and conventions used
5. **Key Files**: Important files and their purposes

### Usage in AI Workflow
```bash
# AI assistant workflow:
1. plaesy analyze          # Generate project understanding
2. Read .plaesy/analysis/project.json     # Get project summary
3. Read .plaesy/analysis/project.structure.json  # Get detailed structure
4. Read .plaesy/context.md               # Get session context
5. Read .plaesy/memory.md             # Get project knowledge
6. Proceed with context-aware development
```

## 📊 Project Categories Detected

### Web Development
- `react` - React.js applications
- `vue` - Vue.js applications
- `angular` - Angular applications
- `nextjs` - Next.js full-stack React
- `nuxtjs` - Nuxt.js Vue.js framework
- `express` - Express.js Node.js backend
- `django` - Django Python framework
- `flask` - Flask Python framework
- `fastapi` - FastAPI modern Python

### Mobile Development
- `flutter` - Flutter cross-platform
- `react-native` - React Native framework
- `ios-native` - iOS Swift/Objective-C
- `android-native` - Android Kotlin/Java

### System Languages
- `go` - Go high-performance backend
- `rust` - Rust systems programming
- `java` - Java enterprise applications
- `dotnet` - .NET framework applications
- `python` - Python applications
- `ruby` - Ruby on Rails applications
- `php` - PHP web applications

### Infrastructure & DevOps
- `docker` - Docker containerization
- `kubernetes` - K8s orchestration
- `terraform` - Infrastructure as code
- `devops` - CI/CD and automation

### Data & Machine Learning
- `machine-learning` - ML model development
- `data-science` - Data analysis with Jupyter
- `jupyter-data-science` - Jupyter notebooks
- `analytics` - Statistical analysis

### Enterprise & Business
- `microservices-architecture` - Distributed systems
- `enterprise-fullstack` - Large-scale applications
- `financial-systems` - Finance and banking
- `healthcare-systems` - Medical applications

### Specialized
- `spec-kit` - Plaesy framework itself
- `game-development` - Unity/Unreal games
- `blockchain` - Web3 and smart contracts
- `iot-systems` - Internet of Things

## 🎯 AI Insights Generated

### Project Analysis
- **Complexity assessment**: Small, medium, or large-scale
- **Maturity level**: Basic, developing, or comprehensive
- **Automation level**: Low, medium, or high automation
- **Organization quality**: Simple, organized, or well-organized

### Development Workflow
- **Version control**: Git usage patterns
- **Build systems**: Make, npm, Cargo, Maven, etc.
- **Testing structure**: Test organization patterns
- **Documentation level**: Documentation completeness

### Recommendations
- **Missing elements**: Suggested additions
- **Improvement areas**: Optimization opportunities
- **Best practices**: Recommended patterns
- **Tool suggestions**: Helpful tools and integrations

## 📝 Example Output

### project.json Structure
```json
{
  "project_summary": {
    "name": "my-project",
    "type": "react",
    "description": "React.js web application with component-based architecture",
    "purpose": "Modern web application with interactive user interface",
    "complexity": "Medium"
  },
  "ai_insights": {
    "overview": "Well-structured React project with comprehensive component architecture",
    "development_workflow": "Node.js/npm ecosystem with script-based tasks",
    "key_patterns": ["test_driven_development", "documentation_first", "container_ready"],
    "recommendations": ["Consider adding more integration tests", "Add API documentation"]
  },
  "structure": {
    "files": {
      "total": 45,
      "code": 23,
      "documentation": 8,
      "automation": 4,
      "configuration": 10
    }
  },
  "technology_stack": {
    "primary_languages": ["JavaScript", "TypeScript", "HTML", "CSS"],
    "development_tools": ["Git", "NPM"],
    "build_systems": ["NPM Scripts", "Webpack"],
    "deployment": "Docker containers"
  }
}
```

## 🔍 Advanced Features

### File Type Detection
- **Source files**: 15+ programming languages
- **Configuration files**: JSON, YAML, TOML, INI
- **Documentation**: Markdown, text, PDF
- **Build files**: Makefiles, Dockerfiles, CI/CD configs
- **Assets**: Images, fonts, media files

### Directory Analysis
- **Purpose identification**: Each directory's role in project
- **Importance levels**: Critical, high, medium, low priority
- **Structure patterns**: Standard vs. custom organization
- **Content analysis**: File types and counts per directory

### Pattern Recognition
- **Conventions detection**: Coding and project conventions
- **Workflow identification**: Development workflow patterns
- **Integration points**: External service integrations
- **Architecture hints**: System architecture clues

## 🛡️ Error Handling

### Robust Analysis
- **Permission handling**: Graceful handling of permission issues
- **Large projects**: Efficient processing of large codebases
- **Edge cases**: Handling of empty or unusual project structures
- **Encoding support**: Multiple file encoding support

### Fallback Behavior
- **Unknown types**: Graceful handling of unrecognized file types
- **Minimal projects**: Works with single-file projects
- **Corrupted data**: Recovery from partial analysis
- **Resource limits**: Memory and processing optimization

## 📚 Integration Examples

### For AI Assistants
```bash
# Standard AI workflow
./bash/plaesy-analyze.sh

# Read analysis results
cat .plaesy/memory/analysis/project.json | jq '.project_summary'
cat .plaesy/memory/analysis/project.structure.json | jq '.key_directories'

# Use insights for development
# - Understand project type and technology stack
# - Identify key files and directories
# - Follow detected conventions
# - Apply recommended improvements
```

### For Development Teams
```bash
# Project onboarding
./bash/plaesy-analyze.sh

# Review project insights
cat .plaesy/memory/analysis/project.json | jq '.ai_insights'

# Identify areas for improvement
cat .plaesy/memory/analysis/project.json | jq '.ai_insights.recommendations'
```

## 🆘 Troubleshooting

### Common Issues
```bash
# Permission denied
chmod +x ./bash/plaesy-analyze.sh

# No analysis directory created
mkdir -p .plaesy/memory/analysis

# JSON parsing errors
# Check for special characters in file names
# Ensure proper file permissions

# Large project timeout
# Script is optimized for performance
# May take time on very large projects
```

### Debug Mode
```bash
# Enable verbose output
set -x
./bash/plaesy-analyze.sh
set +x
```

---

**📊 This script is the foundation of AI understanding in Plaesy projects. Always run it first to provide AI assistants with comprehensive project context.**