# Changelog and Release Management Instructions

## Purpose
Maintain a consistent, user-friendly changelog that documents version history, tracks breaking changes, and communicates impact to users and stakeholders.

## Core Principles

### 1. Changelog Format (Keep a Changelog Standard)
- **File location**: `CHANGELOG.md` in project root
- **Format**: Markdown, structured by version with dates
- **Semantic Versioning**: MAJOR.MINOR.PATCH (v1.2.3)
  - MAJOR: Breaking changes (incompatible API changes)
  - MINOR: New features (backward compatible)
  - PATCH: Bug fixes (backward compatible)

### 2. Changelog Structure per Version

```markdown
# Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]
### Added
- New features in development

### Changed
- Changes to existing functionality

### Deprecated
- Features marked for removal in future versions

### Removed
- Features or APIs removed in this version

### Fixed
- Bug fixes

### Security
- Security vulnerability fixes

## [1.2.0] - 2026-01-15
### Added
- New feature: User authentication with OAuth 2.0
- New API endpoint for batch operations
- Support for dark mode

### Changed
- Improved database query performance (50% faster)
- Updated authentication library to v3.1.0
- Refactored internal caching mechanism

### Fixed
- Fixed memory leak in connection pool
- Corrected timezone handling for international dates

### Security
- Patched XSS vulnerability in user input sanitization (CVE-2025-1234)

## [1.1.0] - 2025-11-20
### Added
- Webhook support for real-time notifications

### Fixed
- Fixed race condition in concurrent request handling

## [1.0.0] - 2025-09-01
- Initial release
```

### 3. Category Definitions

- **[Unreleased]**: Pending features, changes, and fixes for next release
  - Updated as work progresses
  - Moved to versioned section during release

- **Added**: New features or functionality
  - User-facing improvements
  - New APIs, endpoints, or commands
  - New configuration options

- **Changed**: Modifications to existing functionality
  - Updated dependencies (with version numbers)
  - Changed behavior (with rationale)
  - Performance improvements with metrics (e.g., "50% faster")
  - UI/UX improvements

- **Deprecated**: Features marked for future removal
  - Include removal timeline: "Will be removed in v2.0.0"
  - Migration path or alternative: "Use `newFeature()` instead"

- **Removed**: Features, APIs, or support dropped
  - Include replacement if applicable
  - Version when removed

- **Fixed**: Bug fixes and corrections
  - Be specific about what was fixed
  - Reference issue numbers if applicable

- **Security**: Security patches and vulnerability fixes
  - Always include CVE or security advisory reference
  - Severity level if applicable
  - Patch vs. upgrade recommendation

### 4. Writing Guidelines

#### Clarity and Specificity
- ✅ **GOOD**: "Fixed XSS vulnerability in user profile editor allowing arbitrary script injection"
- ❌ **BAD**: "Fixed security issue"

- ✅ **GOOD**: "Improved API response time by 40% through connection pooling optimization"
- ❌ **BAD**: "Performance improvements"

- ✅ **GOOD**: "Dropped support for Node.js 12; minimum version is now Node.js 16"
- ❌ **BAD**: "Removed old Node version support"

#### User-Centric Language
- Write for end users, not just developers
- Explain impact: "Users can now..." or "Breaking change for..."
- Avoid technical jargon unless necessary

#### Examples
```markdown
### Added
- New `--dry-run` flag for safer batch operations (see docs/batch.md)
- Support for PostgreSQL 14+ with JSON operators
- Webhook retry logic with exponential backoff (configurable)

### Changed
- Updated authentication to use OpenID Connect (breaking change)
  - Migration guide: docs/migration/oauth-to-oidc.md
- Improved search performance by 60% through Elasticsearch integration
- API rate limits increased from 100/min to 1000/min

### Removed
- Removed deprecated REST API v1 (v2 available for 12+ months)
- Dropped Python 2 support (EOL since 2020)

### Security
- Patched SQL injection vulnerability in advanced search (CVE-2025-5678)
- Upgraded bcrypt to v5.0.0 for stronger password hashing
```

### 5. Release Workflow

#### Before Release
1. **Audit [Unreleased] section**
   - Review all pending entries
   - Verify accuracy and completeness
   - Consolidate if needed

2. **Create release branch** (if using Git Flow)
   ```bash
   git checkout -b release/v1.2.0
   ```

3. **Update version numbers**
   - package.json (Node.js)
   - setup.py or pyproject.toml (Python)
   - Cargo.toml (Rust)
   - build.gradle (Java/Kotlin)
   - Chart version if using Helm
   - VERSION file (if exists)

4. **Update CHANGELOG.md**
   - Move [Unreleased] content to versioned section
   - Add release date (YYYY-MM-DD format)
   - Create new [Unreleased] section

5. **Commit and tag**
   ```bash
   git add CHANGELOG.md package.json
   git commit -m "chore: bump version to 1.2.0"
   git tag -a v1.2.0 -m "Release v1.2.0"
   ```

6. **Final checks**
   - All tests passing
   - Documentation updated
   - No uncommitted changes

#### After Release
1. **Merge release branch** to main and develop (Git Flow)
   ```bash
   git checkout main
   git merge release/v1.2.0
   git checkout develop
   git merge release/v1.2.0
   git branch -d release/v1.2.0
   ```

2. **Push release** to remote
   ```bash
   git push origin main develop --tags
   ```

3. **Publish release notes**
   - GitHub: Create Release from tag, copy changelog content
   - Announcement: Email, Slack, blog post, etc.

### 6. Version Bumping Rules

#### Semantic Versioning Decision Tree
```
Current version: 1.5.3

Are there breaking changes?
  ├─ YES → MAJOR: 2.0.0
  └─ NO → Continue...
         Are there new features?
           ├─ YES → MINOR: 1.6.0
           └─ NO → PATCH: 1.5.4
```

#### Examples
- **1.0.0 → 2.0.0** (MAJOR)
  - Removed deprecated API endpoints
  - Changed database schema incompatibly
  - Dropped support for older versions
  
- **1.0.0 → 1.1.0** (MINOR)
  - Added new API endpoint
  - Added new configuration option
  - Added new CLI command
  
- **1.0.0 → 1.0.1** (PATCH)
  - Fixed bug in calculation
  - Fixed memory leak
  - Updated dependency patch version

### 7. Deprecation Policy

When marking features for removal:
1. **Announce clearly**: "Deprecated in v1.2.0, will be removed in v2.0.0"
2. **Provide migration path**: Link to docs or show alternative
3. **Timeline**: Minimum 2-3 releases before removal
4. **Warnings**: Add deprecation warnings in code/logs

Example:
```markdown
### Deprecated
- Deprecated `AuthV1` class in favor of `AuthV2` (OAuth 2.0)
  - Removal timeline: v2.0.0 (approximately 6 months)
  - Migration guide: docs/migration/auth-v1-to-v2.md
  - Use: `from auth import AuthV2` instead of `from auth import AuthV1`
```

### 8. Common Mistakes to Avoid

- ❌ **Vague entries**: "Improvements" instead of specific changes
- ❌ **Missing breaking changes**: Always highlight incompatibilities
- ❌ **Inconsistent dates**: Use YYYY-MM-DD format consistently
- ❌ **No CVE references**: Security fixes must include vulnerability ID
- ❌ **Forgetting [Unreleased]**: Always maintain unreleased section
- ❌ **Outdated CHANGELOG**: Update during development, not after release

### 9. Integration with Plaesy Workflow

#### During `/implement` Phase
- Each PR should include CHANGELOG.md update
- Add entries to [Unreleased] section
- Use same categories and format

#### During `/optimize` Phase
- Review entries for clarity
- Update performance metrics in [Changed] section
- Reference optimization techniques used

#### During `/fix` Phase
- Add security fixes to [Security] section with CVE
- Add bug fixes to [Fixed] section with issue reference
- Note hotfix vs. regular patch

#### During `/save` Phase
- Commit CHANGELOG.md changes
- Include changelog entries in memory context

#### Before Release (`/doc` phase or manual)
- Finalize [Unreleased] section
- Create version tag
- Publish release notes

### 10. Tools and Automation

- **Standard-version**: Automate versioning and changelog generation
  ```bash
  npx standard-version  # Generates changelog from conventional commits
  ```

- **conventional-commits**: Enforce commit message format for auto-changelog
  - Commit format: `type: description`
  - Auto-generates changelog from commits

- **GitHub Actions**: Automate changelog generation
  - Trigger on tagged releases
  - Auto-create GitHub Release from CHANGELOG.md content

- **Release Notes Generators**: Tools like Release Drafter
  - Automatically collects PR descriptions into changelog
  - Requires PR labels to categorize entries
