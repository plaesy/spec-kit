---
applyTo: '*'
description: "Git workflows, branching strategies, commit discipline, and version control best practices for consistent team development."
---

# Git Workflow and Version Control Instructions

## Purpose
Establish consistent git practices: disciplined commits, clear branching strategies, atomic changes, and maintainable version history.

## Core Principles

### 1. Commit Discipline (Atomic Changes)
- **One logical change per commit**: Each commit should represent a single, testable, reviewable change
- **Self-contained commits**: Commit must not break the build or introduce known regressions
- **Meaningful commit messages**: 
  - Format: `<type>: <subject>` (50 chars max for subject)
  - Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `style`, `perf`
  - Example: `fix: prevent XSS vulnerability in user input sanitization`
  - Body: Explain WHY, not WHAT (the diff shows what changed)
  - Reference issues: `Closes #123` or `Relates to #456`

### 2. Branching Strategy (Git Flow Inspired)
- **main**: Production-ready code only
  - Protected branch; requires PR review + passing tests
  - Tagged with semantic versions (v1.0.0)
  - Hotfixes branched from main when needed
  
- **Integration branch** (develop, dev, or project-specific): Pre-release work
  - Pre-release, feature complete, testing in progress
  - Merge from feature branches
  - Release branches cut from here
  - **Convention**: Use `develop` (Git Flow standard) or document chosen name in project README/CLAUDE.md
  
- **Feature branches**: From integration branch or main
  - Format: `feature/short-description` or `feat/short-description`
  - Example: `feature/user-authentication`, `feat/api-rate-limiting`
  - Delete after merge (keep history clean)
  
- **Release branches**: Pre-release stabilization
  - Format: `release/v1.0.0`
  - Merge back to main AND integration branch after release
  
- **Hotfix branches**: Emergency production fixes
  - Format: `hotfix/issue-number` or `hotfix/critical-issue`
  - Branch from main, merge back to main AND integration branch

### 3. Pull Request Standards
- **Clear title and description**: Explain what changed and why
- **Link to issues**: Reference related issues/tickets
- **Keep PRs focused**: Ideally <400 lines changed, <10 files affected
- **Request reviews**: At least one approval before merge
- **Squash/rebase on merge**: Keep main/develop history clean (unless specific commits need preservation)
- **Delete branch after merge**: GitHub default; avoid stale branches

### 4. Merge Conflict Resolution
- **Always communicate**: If conflicts likely, discuss with team before merging
- **Resolve locally**: Pull latest main, resolve conflicts in your branch, verify tests pass
- **Never accept "ours" without review**: Understand both sides of the conflict
- **Test after resolving**: Conflicts can introduce silent bugs

### 5. Tags and Versioning
- **Semantic Versioning**: MAJOR.MINOR.PATCH (e.g., v1.2.3)
  - MAJOR: Breaking API changes
  - MINOR: Backward-compatible new features
  - PATCH: Bug fixes, no new features
- **Tag on main only**: `git tag v1.0.0 && git push origin v1.0.0`
- **Annotated tags**: Include message
  - `git tag -a v1.0.0 -m "Release v1.0.0: feature descriptions and bug fixes"`

### 6. Commit History Health
- **Avoid force-push to shared branches**: Never `git push --force` to main/develop
- **Rebase feature branches locally**: Keep feature branch history clean before PR
  - `git rebase develop` (then squash commits if needed)
- **No merge commits in feature branches**: Use `git rebase` or squash on merge
- **Keep .gitignore up-to-date**: 
  - Exclude: node_modules, build artifacts, .env files, IDE settings
  - Review before each commit: `git status` shows untracked files

### 7. Security and Secrets
- **NEVER commit secrets**: API keys, tokens, passwords, private keys
- **Use environment files**: .env (in .gitignore), or CI/CD secrets
- **Scan before push**: Use git hooks (`pre-commit` hook) or tools like Gitleaks
- **If secrets leaked**: 
  - Revoke/rotate immediately
  - Use `git filter-branch` or BFG Repo-Cleaner to remove from history
  - Force-push only after confirmation (secrets no longer valid)

### 8. Collaboration Practices
- **Pull before push**: Always `git pull` before `git push` to avoid conflicts
- **Communicate branch work**: Let team know which branch you're working on
- **Code review checklist**:
  - Does it solve the problem as described?
  - Is it secure? (No SQL injection, XSS, hardcoded secrets, etc.)
  - Is it testable? (Tests included or updated?)
  - Is it maintainable? (Clear naming, no unnecessary complexity)
  - Does it follow project conventions?
- **Respond to review feedback**: Timely, thoughtful responses

### 9. Local Development Workflow
Replace `$INTEGRATION_BRANCH` with your project's integration branch name (e.g., `develop`, `dev`, `staging`).

```bash
# Ensure feature branch exists and switch to it
git checkout -b feature/my-feature origin/$INTEGRATION_BRANCH  # Create new
# OR
git checkout feature/my-feature  # Switch to existing

# Make commits
git add <files>
git commit -m "feat: add user authentication"

# Before pushing, sync with latest integration branch
git fetch origin
git rebase origin/$INTEGRATION_BRANCH

# Push to remote
git push origin feature/my-feature

# After PR merge, clean up
git checkout $INTEGRATION_BRANCH
git pull origin $INTEGRATION_BRANCH
git branch -d feature/my-feature
git push origin --delete feature/my-feature
```

**Note**: Substitute `$INTEGRATION_BRANCH` with actual branch name for your project.

### 10. Avoiding Common Mistakes
- **Mistake**: Committing large binary files (use Git LFS if necessary)
  - **Fix**: Check `.gitattributes`, use `git lfs track "*.psd"` for large files
  
- **Mistake**: Working directly on main or integration branch (develop/dev/staging)
  - **Fix**: Always create feature branch, never direct commits to protected branches
  
- **Mistake**: Huge commits with multiple unrelated changes
  - **Fix**: Split into atomic commits with clear messages
  
- **Mistake**: Unclear commit messages ("fixed stuff", "wip")
  - **Fix**: Descriptive messages explaining context and reasoning
  
- **Mistake**: Not testing before pushing
  - **Fix**: Local tests passing + CI tests passing before merge

## Integration with Plaesy Workflow
- **Before `/implement`**: Ensure feature branch exists and checked out
  - Create if missing: `git checkout -b feature/description origin/integration-branch`
  - Reuse existing: `git checkout feature/description`
  - Never `/implement` on main or integration branch (develop/dev/staging)
  - Follow `.plaesy/memory/quality-gates.md` for pre-implementation checks
- **During `/implement`**: Regular atomic commits with clear messages
  - Reference `.plaesy/memory/authoring-invariants.md` for commit discipline
  - Run tests per `.plaesy/memory/testing-strategy.md` before pushing
- **Before PR**: Rebase on latest integration branch, squash if needed
- **After merge**: Update local integration branch with pull, delete feature branch
- **On `/save`**: Reference commit SHA or PR number in memory

## Tools and Automation
- **Pre-commit hooks**: Prevent secrets, enforce linting
- **CI/CD integration**: Auto-test on push, block merges if tests fail
- **GitHub/GitLab features**: Branch protection, auto-delete after merge, require reviews
- **Git aliases**: Create shortcuts for common commands
  ```bash
  git config --global alias.co checkout
  git config --global alias.br branch
  git config --global alias.ci commit
  git config --global alias.st status
  ```
