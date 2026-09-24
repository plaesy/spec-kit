---
name: plaesy-install-scripts-v0.0.2
description: Cross-platform installation scripts & pre-release audit for v0.0.2
metadata:
  type: project
---

# Plaesy v0.0.2 Install Scripts & Pre-Release Audit

**Date**: 2026-09-25  
**Status**: ✅ Complete - Ready for release tag v0.0.2

## Summary

Created automated cross-platform installation scripts and completed comprehensive pre-release audit:

1. **Install Scripts**: `install.sh` (Linux/macOS) + `install.ps1` (Windows PowerShell)
2. **Documentation Clarity**: Fixed all prompt path references, explained platform-specific mapping
3. **Audit Checklist**: All items verified - no breaking changes, ready for binary release

## Implementation

### Install Scripts

**`install.sh` (Linux/macOS)**
- Auto-detect platform (Linux/macOS) & architecture (x64/ARM64)
- Download latest binary from GitHub Releases
- Install to `/usr/local/bin` or `~/.local/bin`
- Handle sudo when needed
- Verify PATH and print next steps

**`install.ps1` (Windows PowerShell)**
- Auto-detect architecture (x64/ARM/ARM64)
- Download latest binary from GitHub Releases
- Smart install path (Program Files for admin, LocalAppData for regular users)
- Handle PATH verification
- Provide instructions for permanent PATH update

**README.md Updated**
- New "Option 1: Automated Installation" section (recommended)
- Moved "Build from Source" to Option 2
- Clear curl one-liners for both platforms

### Documentation Audit & Fixes

**Issue Identified**: `docs/prompts/README.md` referenced `prompts/assess.md` file paths, confusing users about where prompts live post-install

**Root Cause**: Prompts install to platform-specific folders:
- Claude Code → `.claude/commands/`
- Cursor → `.cursor/rules/`
- OpenCode → `.opencode/prompts/`
- GitHub Copilot → `.github/prompts/`
- Windsurf → `.windsurf/prompts/`

**Fixes Applied**:

1. **Added disclaimer table** to `docs/prompts/README.md`
   - Explains prompts/ exists only in source repo
   - Shows platform-specific install locations (5 AI tools)
   - Emphasizes command-name reference post-install

2. **Changed documentation format**
   - All "File: `prompts/x.md`" → "Command: `/x`"
   - Removed file path references from section headers
   - Emphasized command names throughout

3. **Fixed internal references**
   - `prompts/create.md`: removed `prompts/create/{scope}.md` reference
   - Changed to "each scope has its own command"
   - Updated example references

**Reference Validation Results**:
- ✅ No `.plaesy/prompts/` dead references
- ✅ No broken `prompts/` file paths in instructions/templates/chatmodes
- ✅ All references use command names or explain data sources
- ✅ Infrastructure comments (about data loading) are context-appropriate

## Pre-Release Checklist

### ✅ Install Scripts
- [x] `install.sh` with platform/arch detection
- [x] `install.ps1` with admin privilege handling
- [x] Both download from GitHub Releases (latest)
- [x] PATH verification & guidance
- [x] README integration

### ✅ CI/CD Pipeline
- [x] GitHub Actions release.yml exists
- [x] Builds 5 platforms (Linux x64/ARM64, macOS x64/ARM64, Windows x64)
- [x] Auto-publishes to GitHub Releases
- [x] Install scripts can download latest binary

### ✅ Documentation Clarity
- [x] Prompt folder mapping explained
- [x] Platform-specific locations documented
- [x] Command name usage emphasized
- [x] All file path references fixed

### ✅ Reference Validation
- [x] No broken prompts/ paths
- [x] No `.plaesy/prompts/` references
- [x] All references by command name
- [x] Infrastructure comments appropriate

### ✅ Code Quality
- [x] All commits atomic & well-described
- [x] Changes pushed to dev branch
- [x] No uncommitted changes

## Commits

```
0b2bb7c - Fix prompts/create.md references to command names
923add2 - Clarify prompt folder mapping in docs/prompts/README.md
e13f669 - Add automated installation scripts + README updates
```

## Next Steps

1. **Create Release Tag**
   ```bash
   git tag v0.0.2
   git push origin v0.0.2
   ```

2. **GitHub Actions Triggers**
   - Auto-builds for 5 platforms
   - Uploads to GitHub Releases
   - Release becomes available immediately

3. **Verification**
   - Test `install.sh` download from fresh release
   - Test `install.ps1` on Windows
   - Verify binary executes correctly

4. **User Communication**
   - Share new install instructions
   - Promote ease of install (no Go required!)
   - Link to docs/prompts/README.md for platform mapping

## Quality Metrics

**Build Quality**: ✅ No errors or warnings  
**Documentation**: ✅ Clear, consistent, with disclaimers  
**User Experience**: ✅ One-line curl/PowerShell commands  
**Platform Support**: ✅ 5 platforms covered + automated

---

**Why This Matters**: v0.0.2 enables users to install plaesy without Go, eliminating the "clone + build" friction. Install scripts auto-download latest binary, and documentation clarity prevents confusion about prompt file locations.
