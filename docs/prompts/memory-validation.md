# Memory Self-Containment Validation

Ensure your project memory stays **project-local** (in `.plaesy/memory/`) and never leaks external references that won't survive on another machine.

---

## Why This Matters

When you commit and push your project:
- ❌ External paths like `~/.claude/projects/` don't exist on other machines → broken references
- ❌ `/tmp/` files get deleted on restart → lost data
- ✅ `.plaesy/memory/` is versioned in git → available everywhere

**Self-Containment Rule**: Memory must be **complete and standalone** inside `.plaesy/memory/`.

---

## Validate Your Memory

### Bash
```bash
./scripts/bash/plaesy-validate-memory.sh
```

### PowerShell
```powershell
.\scripts\powershell\plaesy-validate-memory.ps1
```

### Output
```
[INFO] Scanning .plaesy/memory for external references...

[✓] Self-containment validated ✓
[✓] Memory is project-local and git-safe
```

If issues found:
```
[✗] External references in: .plaesy/memory/overview.md
  Lines with '~/.claude/projects':
    42: See details in ~/.claude/projects/myproject/memory/
```

---

## Common Issues & Fixes

### Issue: `~/.claude/projects/...` reference found
**Cause**: Linking to global auto-memory instead of copying content

**Fix**:
1. Open the file mentioned in error
2. Copy the referenced content into `.plaesy/memory/`
3. Replace external path with internal link: `[[filename.md]]`
4. Re-run validator

### Issue: `/tmp/...` or absolute home path found
**Cause**: Temporary file that doesn't belong in persistent memory

**Fix**:
1. If content is valuable: copy into `.plaesy/memory/`
2. If not: remove the reference
3. Re-run validator

### Issue: `CLAUDE_CODE_DIR` or environment variable reference
**Cause**: Hardcoded path that's machine-specific

**Fix**:
1. Replace with relative path from repo root
2. Example: `$CLAUDE_CODE_DIR/config` → `scripts/configs/`
3. Re-run validator

---

## Validation Rules (What Gets Scanned)

Validator checks for these **external path patterns**:

| Pattern | What It Is | Fix |
|---------|-----------|-----|
| `~/.claude/projects` | Global auto-memory | Copy content into `.plaesy/memory/` |
| `~/.claude/plans` | Plan-mode files | Copy content in, use internal links |
| `~/.claude/` | Any Claude config | Not repo-safe, copy content in |
| `/tmp/` | Temp directory | Content may be deleted, move to repo |
| `C:\Users\..\.claude\` | Windows global memory | Copy content in |
| `/Users/.../.claude/` | macOS global memory | Copy content in |
| `CLAUDE_CODE_DIR` | Environment variable | Use relative paths instead |

---

## When to Run

### After `/save` command
Automatic ✓ (validation is part of save protocol)

### Before `git commit`
Manual check (recommended):
```bash
./scripts/bash/plaesy-validate-memory.sh && git add .plaesy/memory/ && git commit -m "Update memory"
```

### In CI/CD pipeline
Fail build if validation fails:
```bash
./scripts/bash/plaesy-validate-memory.sh || exit 1
```

---

## Self-Containment Checklist

Before committing memory changes:

- [ ] No `~/.claude/` paths in `.plaesy/memory/**/*.md`
- [ ] No `/tmp/` or temp directory paths
- [ ] No absolute home paths (e.g., `C:\Users\username\`, `/Users/username/`)
- [ ] No external links pointing outside the repo
- [ ] All cross-references use internal paths: `[[topic.md]]`
- [ ] Validation script passes: `./scripts/bash/plaesy-validate-memory.sh` → exit 0

---

## FAQ

**Q: Can I reference `.claude/commands/` or `.claude/roles/` in memory?**

A: Yes, those are inside the repo. Reference them as: `see .claude/commands/evolve.md`

**Q: What if I want to document an external resource (like a GitHub issue)?**

A: Document it by **content** not path:
- ❌ "See ~/.claude/projects/X/research/issue-123.md"
- ✅ "Feature: Rate limiting (GitHub issue #42 - implements exponential backoff)"

**Q: Does validation run automatically?**

A: Partially:
- ✅ `/save` command includes validation step (mandatory, blocks save with external refs)
- ❌ Git commit does NOT block (but you can add pre-commit hook if desired)
- Manual: Run validator anytime with script above

**Q: How do I know if validation passed?**

A: Exit code:
```bash
./scripts/bash/plaesy-validate-memory.sh
echo $?  # 0 = pass, 1 = fail
```

---

## Integration with `/save`

When you run `/save` command, validation is **automatic**:

1. Memory files are updated to `.plaesy/memory/`
2. Validation scan runs (**MANDATORY**)
3. If external paths found → save STOPS, reports, asks user to fix
4. If clean → save completes ✓

You should never need to manually run validator after `/save`, but you can anytime to spot-check.
