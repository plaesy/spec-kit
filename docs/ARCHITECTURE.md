# 🏗️ Plaesy Spec-Kit Architecture & Reference Paths

**Complete guide to file locations, copying mechanisms, and correct reference patterns.**

---

## 🎯 Architecture Overview

Plaesy uses a **master-copy architecture**:

```
┌─────────────────────────────────────────┐
│     SPEC-KIT REPOSITORY (Master)        │
│  (github.com/plaesy/spec-kit)           │
├─────────────────────────────────────────┤
│ • prompts/                              │
│ • instructions/                         │
│ • templates/                            │
│ • scripts/                              │
│ • docs/                                 │
└─────────────────────────────────────────┘
         │ plaesy init │
         ↓
┌─────────────────────────────────────────┐
│   PROJECT REPOSITORY (Self-Contained)   │
│  (user's project, any git repo)         │
├─────────────────────────────────────────┤
│ • CLAUDE.md               (root)         │
│ • AGENTS.md               (root)         │
│ • .plaesy/                              │
│   ├── memory/             (flat struct) │
│   │   ├── plaesy.md      (universal)   │
│   │   ├── overview.md    (project)     │
│   │   ├── workflows.md                 │
│   │   └── [topics].md    (detailed)    │
│   ├── roles/              (flat struct) │
│   │   ├── designer.md    (chatmodes)   │
│   │   ├── ba.md                        │
│   │   └── [name].md                    │
│   ├── analysis/           (graphs)      │
│   ├── scripts/            (tools)       │
│   ├── tasks/              (backlog)     │
│   ├── context.md          (session)     │
│   └── memory.md           (index/TOC)   │
│ • .cursor/rules/          (Cursor)      │
│ • .github/copilot-instructions.md      │
│ • ... (other platform paths)            │
└─────────────────────────────────────────┘
```

---

## 📁 Spec-Kit Structure (Master/Reference)

### `/prompts/` - Workflow Definitions

| File | Purpose | Copied To | Context |
|------|---------|-----------|---------|
| `start.md` | Feature discovery & planning | Platform-specific | Spec-kit master |
| `clarify.md` | Specification clarification | Platform-specific | Spec-kit master |
| `research.md` | Deep research & exploration | Platform-specific | Spec-kit master |
| `design.md` | UI/UX design generation | Platform-specific | Spec-kit master |
| `flow.md` | System design & architecture | Platform-specific | Spec-kit master |
| `implement.md` | Code implementation (TDD) | Platform-specific | Spec-kit master |
| `assess.md` | Quality assessment & review | Platform-specific | Spec-kit master |
| `optimize.md` | Performance optimization | Platform-specific | Spec-kit master |
| `fix.md` | Bug fixing & debugging | Platform-specific | Spec-kit master |
| `continue.md` | Resume interrupted work | Platform-specific | Spec-kit master |
| `save.md` | Memory persistence & export | Platform-specific | Spec-kit master |
| `doc.md` | Documentation generation | Platform-specific | Spec-kit master |
| `evolve.md` | Autonomous loop orchestration | Platform-specific | Spec-kit master |

**Note**: Prompts are platform-specific. When copied to project, they go to platform-specific locations:
- Claude Code → `.claude/commands/`
- Cursor → `.cursor/rules/`
- GitHub Copilot → `.github/prompts/`
- etc.

### `/instructions/` - Development Guidance

**These files ARE copied to projects:**

| Category | Files | Copied To | When Loaded |
|----------|-------|-----------|------------|
| **Root Core** | `agents.instructions.md` → `CLAUDE.md` | Project root | Every Claude Code session |
| **Universal Core** | `plaesy.instructions.md` → `.plaesy/memory/plaesy.md` | `.plaesy/memory/` | Via mapping.json always_load |
| **Always-Load** | All 21 instructions (assess, clarify, fix, etc + plaesy-trim, plaesy-graph, etc) | `.plaesy/memory/` | Via mapping.json always_load |
| **Framework-Specific** | react.js, next.js, angular, spring-boot, nestjs, rails, flutter, etc | `.plaesy/memory/` | If detected in code |
| **Language-Specific** | java, go, rust, csharp, python, sql, etc | `.plaesy/memory/` | If detected in code |
| **Cross-Cutting** | testing-strategy, security-audit, tech-validation, performance-baseline, devops, kubernetes, terraform, git, tdd-enforcement | `.plaesy/memory/` | If detected/needed |
| **Methodology** | brainstorming-techniques, changelog, markdown | `.plaesy/memory/` | If detected |

---

## 🔄 Reference Path Rules

### ✅ CORRECT Reference Patterns

**In Spec-Kit Prompts** (master references):
```markdown
<!-- References to other prompts in spec-kit -->
> **See also**: [plaesy.md](../instructions/plaesy.instructions.md) for core guidelines

<!-- References to instructions (after copy to project) -->
**Follow shared protocols**: `.plaesy/memory/quality-gates.md` → `.plaesy/memory/error-recovery.md`
```

**In Spec-Kit Instructions** (master files):
```markdown
<!-- Cross-references to other instructions (after copy) -->
**Follow**: `.plaesy/memory/quality-gates.md` → Quality Enforcement

<!-- DO NOT reference self -->
✅ File does not reference itself

<!-- Example for other files referencing shared-protocols -->
Example: `.plaesy/memory/quality-gates.md` → Quality Enforcement
```

**In Project Documentation** (after plaesy init):
```markdown
<!-- All references point to project copies -->
See `.plaesy/memory/plaesy.md` for all core guidelines
Follow `.plaesy/memory/quality-gates.md` for validation
```

---

### ❌ INCORRECT Reference Patterns

```markdown
<!-- WRONG: References to spec-kit paths in instruction files -->
✗ "Follow instructions/quality-gates.instructions.md"
✗ "See prompts/start.md for examples"

<!-- WRONG: Self-references -->
✗ In error-recovery.instructions.md: "Follow instructions/error-recovery.instructions.md"

<!-- WRONG: Outdated paths -->
✗ ".plaesy/memory/quality-gates.md" (should be flat: .plaesy/memory/quality-gates.md)
✗ "Follow prompts/quality-gates.md" (prompts don't go to .plaesy/memory/, they're platform-specific)
```

---

## 🎯 Copy Mechanism: How Files Move to Projects

### 1. Instructions Copy (`plaesy init`)

**Source**: `instructions/*.instructions.md`  
**Destination**: `.plaesy/memory/`  
**Filename transformation**:
- `instructions/quality-gates.instructions.md` → `.plaesy/memory/quality-gates.md`
- `instructions/error-recovery.instructions.md` → `.plaesy/memory/error-recovery.md`
- `instructions/plaesy.instructions.md` → `.plaesy/memory/plaesy.md`

**Selection Logic**:
```
1. Detect project technologies (package.json, go.mod, Cargo.toml, etc)
2. Load instructions/mapping.json
3. Find entries for always_load → copy all
4. Find entries matching detected tech → copy matched only
5. Result: Lightweight, project-specific knowledge base
```

**Effect**: Project developers have access to only relevant instructions + always-load core. Files stored flat in `.plaesy/memory/` (no subfolders).

### 2. Scripts Copy (`plaesy init`)

**Source**: `scripts/powershell/` and `scripts/bash/`  
**Destination**: `.plaesy/scripts/{{powershell|bash}}/`  
**Files copied**: All automation scripts (plaesy-trim.ps1/sh, plaesy-graph.ps1/sh, etc.)

**Filename transformation**:
- `scripts/powershell/plaesy-trim.ps1` → `.plaesy/scripts/powershell/plaesy-trim.ps1`
- `scripts/bash/plaesy-trim.sh` → `.plaesy/scripts/bash/plaesy-trim.sh`
- `scripts/configs/` → `.plaesy/scripts/configs/` (platform.json, plaesy-trim-rules.json, etc.)

**Effect**: Project has self-contained tools for token compression, knowledge graph analysis, and workflow automation.

### 3. Prompts Copy (Platform-Specific)

**Source**: `prompts/*.md`  
**Destination**: Platform-dependent
- Claude Code: `.claude/commands/`
- Cursor: `.cursor/rules/`
- GitHub Copilot: `.github/prompts/`
- Codeium: `.codeium/prompts/`
- etc.

**Selection Logic**:
```
1. Detect platform (check for .claude/, .cursor/, .github/, etc)
2. Copy ALL prompts to platform-specific location
3. Result: Prompts available in that platform's context
```

**Effect**: Project developers have access to all workflow prompts on their platform.

### 4. Chatmodes Copy (`plaesy init`)

**Source**: `chatmodes/*.chatmode.md`  
**Destination**: `.plaesy/roles/` (platform-agnostic, all platforms)  
**Filename transformation**:
- `chatmodes/designer.chatmode.md` → `.plaesy/roles/designer.md`
- `chatmodes/ba.chatmode.md` → `.plaesy/roles/ba.md`
- `chatmodes/pm.chatmode.md` → `.plaesy/roles/pm.md`
- etc. (22 chatmode files total)

**Selection Logic**:
```
1. Copy ALL 22 chatmode files to .plaesy/roles/
2. Stored flat (no subfolders) like instructions
3. Same format for all platforms (unlike prompts which are platform-specific)
```

**Effect**: 
- All specialized chatmodes available for `/assess` delegation (design review, business analysis, marketing, legal, etc.)
- Single source of truth across all platforms
- Follows same pattern as instructions → `.plaesy/memory/`

### 5. Tasks Structure Creation (`plaesy init`)

**Destination**: `.plaesy/tasks/`  
**Structure**:
```
.plaesy/tasks/
├── backlog/     # Unscheduled ideas, features, bugs
├── todo/        # Ready to start (next in queue)
├── doing/       # In active work
├── done/        # Completed and validated
└── blocked/     # Waiting on dependency
```

**Effect**: Project has autonomous workflow support for `/evolve` command and manual task management via `plaesy-task-manage.sh` / `.ps1`

### 5. Analysis Directory (`plaesy init` + tools)

**Destination**: `.plaesy/analysis/`  
**Contents**:
- `project.graph.json` — Knowledge graph structure (from plaesy-graph)
- `project.html` — Interactive knowledge graph visualization
- `reports.md` — Plain-language summary of relationships
- `token-stats.json` — Token savings tracked by plaesy-trim
- Other analysis outputs as generated

**Effect**: Analysis outputs isolated from memory, enabling easy cleanup/regeneration.

### 6. Memory & Loop State Templates (`plaesy init`)

**Files created**:
- `.plaesy/context.md` — Current session template (max 100 lines)
- `.plaesy/memory.md` — Index/TOC of memory files (max 50 lines)
- `.plaesy/state.json` — Autonomous loop configuration

**Effect**: Project ready for autonomous loops with proper governance and memory management

---

## 📝 Reference Paths Cheat Sheet

| Location | Reference To | Example | Notes |
|----------|-------------|---------|-------|
| **Spec-kit prompts** | Instructions (after copy) | `.plaesy/memory/quality-gates.md` | Platform-agnostic |
| **Spec-kit prompts** | Chatmodes (after copy) | `.plaesy/roles/designer.md` | For specialized delegation |
| **Spec-kit instructions** | Other instructions (after copy) | `.plaesy/memory/quality-gates.md` | NO self-reference |
| **Spec-kit instructions** | Chatmodes (after copy) | `.plaesy/roles/ba.md` | For delegation |
| **Spec-kit docs** | Instructions (after copy) | `.plaesy/memory/plaesy.md` | For project context |
| **Spec-kit docs** | Prompts (spec-kit) | `../prompts/start.md` or `prompts/start.md` | For documentation context |
| **Spec-kit docs** | Chatmodes (spec-kit) | `../chatmodes/designer.chatmode.md` | For documentation context |
| **Project context** | Instructions | `.plaesy/memory/[name].md` | Always use this (flat structure) |
| **Project context** | Chatmodes | `.plaesy/roles/[name].md` | Always use this (flat structure) |
| **Project context** | Prompts | Platform-specific path | Not needed; prompts are in platform context |

---

## 🔍 Why This Architecture?

### Problem Solved
1. **Self-Contained Projects** - Each project has only its relevant instructions (CLAUDE.md rule)
2. **Platform Independence** - Prompts work across Claude Code, Cursor, GitHub Copilot, etc.
3. **No Broken Links** - After copy, all internal references point to actual files
4. **Scalability** - New instructions don't slow down small projects

### Key Rule: `.plaesy/memory/` is Canonical (Flat Structure)

After `plaesy init`:
- ✅ Instructions are in `.plaesy/memory/` (flat, no subfolders)
- ✅ All references point to this location
- ✅ Project is self-contained (no external deps)

---

## ✅ Checklist: Reference Correctness

Before committing files, verify:

### Spec-Kit Instruction Files
- [ ] No self-references (file doesn't reference itself)
- [ ] Cross-references use `.plaesy/memory/` paths (flat structure, no subfolders)
- [ ] All references are to files that will exist after copy

### Spec-Kit Script Files
- [ ] All config references use `.plaesy/scripts/configs/` (for project context)
- [ ] No references to spec-kit `scripts/` paths in usage examples
- [ ] Usage examples show `.plaesy/scripts/` paths

### Spec-Kit Prompt Files
- [ ] All instruction references use `.plaesy/memory/` paths (flat structure)
- [ ] All prompt cross-references are relative (e.g., `[name.md](name.md)`)
- [ ] Frontmatter includes proper description

### Spec-Kit Documentation
- [ ] Instructions documented in `docs/instructions/README.md`
- [ ] Prompts documented in `docs/prompts/README.md`
- [ ] Scripts documented in `docs/scripts/README.md`
- [ ] Architecture clear for maintainers

### Project Context (After `plaesy init`)
- [ ] All instruction references point to `.plaesy/memory/` (flat, no subfolders)
- [ ] All chatmode references point to `.plaesy/roles/` (flat, no subfolders)
- [ ] All script references point to `.plaesy/scripts/`
- [ ] All prompt references point to platform-specific paths
- [ ] All analysis references point to `.plaesy/analysis/`
- [ ] No references to `instructions/`, `chatmodes/`, or `scripts/` (spec-kit paths)
- [ ] Index files created at root level (memory.md, context.md, state.json)
- [ ] All instruction files in `.plaesy/memory/` directly (flat: quality-gates.md, not quality-gates/index.md)
- [ ] All chatmode files in `.plaesy/roles/` directly (flat: designer.md, not designer/index.md)
- [ ] Tasks structure created at `.plaesy/tasks/`
- [ ] Analysis outputs at `.plaesy/analysis/`

---

## 🚀 Quick Reference: Fix Checklist

**If documentation references are broken:**

1. **Find broken reference**:
   ```bash
   grep -r "instructions/.*\.instructions\.md" docs/ prompts/
   grep -r "prompts/shared/" .
   ```

2. **Identify correct path**:
   - Instruction → `.plaesy/memory/[name].md` (flat structure)
   - Prompt → Relative or platform-specific
   - Self-ref → Remove entirely

3. **Update and verify**:
   ```bash
   # Find all references
   grep -r "\.plaesy/memory/" prompts/ | wc -l
   # Should match: number of instruction references (no /knowledge/ subfolder)
   ```

---

**Last updated**: 2026-09-06  
**Version**: 1.0  
**Maintainer**: Plaesy Team
