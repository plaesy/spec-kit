# Platform-Specific Installation Troubleshooting

This guide helps resolve common issues related to platform-specific script filtering in Plaesy Spec-Kit.

## Overview

Since version 2.0, Plaesy Spec-Kit implements intelligent platform filtering during installation:
- **Linux/macOS**: Only bash scripts (*.sh) are installed
- **Windows**: Only PowerShell scripts (*.ps1) are installed

This optimization reduces installation size by ~50% and eliminates platform incompatibilities.

## Common Issues & Solutions

### 1. Wrong Platform Script Detected

**Symptoms:**
- Installation completes but CLI commands fail
- Scripts not found errors
- Platform mismatch warnings

**Diagnosis:**
```bash
# Check installed script types
find ~/.plaesy/scripts -name "*.sh" -o -name "*.ps1" | head -10

# Verify CLI wrapper paths
cat ~/.local/bin/plaesy
```

**Solutions:**
- **Linux/macOS**: Should only see .sh files
- **Windows**: Should only see .ps1 files
- If mixed: Reinstall using correct platform installer

### 2. CLI Command Not Found

**Symptoms:**
- `plaesy: command not found`
- `The term 'plaesy' is not recognized`

**Diagnosis:**
```bash
# Check if CLI wrapper exists
ls -la ~/.local/bin/plaesy*

# Check PATH configuration
echo $PATH | grep -o ~/.local/bin
```

**Solutions:**
1. **Linux/macOS**: Add to ~/.bashrc or ~/.zshrc:
   ```bash
   export PATH="$HOME/.local/bin:$PATH"
   ```

2. **Windows**: Add to PowerShell profile:
   ```powershell
   $env:PATH += ";$env:USERPROFILE\.local\bin"
   ```

3. **Reload shell**: `source ~/.bashrc` or restart terminal

### 3. Installation Script Fails

**Symptoms:**
- Permission denied errors
- Network connectivity issues
- Git clone failures

**Diagnosis:**
```bash
# Check permissions
ls -la /path/to/install.sh

# Test connectivity
curl -s https://api.github.com/repos/plaesy/spec-kit | head -5

# Check disk space
df -h ~
```

**Solutions:**
1. **Permission issues**:
   ```bash
   chmod +x install.sh
   ```

2. **Network issues**:
   - Use proxy if needed: `export https_proxy=http://proxy:port`
   - Try alternative installation method: manual download

3. **Disk space**: Ensure at least 50MB free space

### 4. Existing Installation Conflicts

**Symptoms:**
- Mixed script types after upgrade
- Version conflicts
- Incomplete filtering

**Diagnosis:**
```bash
# Check installation details
plaesy status

# Verify framework version
cat ~/.plaesy/VERSION 2>/dev/null || echo "Version file missing"

# Check for orphaned files
find ~/.plaesy -name "*.ps1" -o -name "*.sh" | sort
```

**Solutions:**
1. **Clean reinstall**:
   ```bash
   # Backup important data
   cp -r ~/.plaesy/memory/knowledge ~/.plaesy-knowledge-backup
   
   # Remove and reinstall
   rm -rf ~/.plaesy
   bash install.sh
   
   # Restore data
   cp -r ~/.plaesy-knowledge-backup ~/.plaesy/memory/knowledge
   ```

2. **Selective cleanup** (Advanced):
   ```bash
   # Remove wrong platform scripts
   find ~/.plaesy/scripts -name "*.ps1" -delete  # On Linux/macOS
   find ~/.plaesy/scripts -name "*.sh" -delete   # On Windows (PowerShell)
   ```

### 5. Project Initialization Issues

**Symptoms:**
- `plaesy init` fails
- Missing template files
- Incomplete project setup

**Diagnosis:**
```bash
# Check framework completeness
ls ~/.plaesy/templates/ | wc -l  # Should be >10 files

# Test init in empty directory
mkdir /tmp/test-project && cd /tmp/test-project
plaesy init .
```

**Solutions:**
1. **Missing templates**: Reinstall framework
2. **Permissions**: Check write permissions in target directory
3. **Path issues**: Verify `~/.plaesy` exists and is complete

### 6. Performance Issues

**Symptoms:**
- Slow CLI responses
- Installation takes too long
- High memory usage

**Diagnosis:**
```bash
# Check installation size
du -sh ~/.plaesy

# Time CLI operations
time plaesy status

# Monitor system resources
top -p $(pgrep -f plaesy) | head -10
```

**Solutions:**
1. **Large installation**: 
   - Expected size: ~2MB
   - If >10MB: Clean reinstall
   
2. **Slow responses**:
   - Check disk I/O: `iostat 1 5`
   - Verify SSD/HDD performance
   
3. **Memory issues**: Usually indicates larger problem, reinstall recommended

## Advanced Troubleshooting

### Debug Mode Installation

Enable verbose logging during installation:

**Linux/macOS:**
```bash
bash install.sh 2>&1 | tee install-debug.log
```

**Windows:**
```powershell
.\install.ps1 -Verbose | Tee-Object -FilePath install-debug.log
```

### Manual Verification

Verify platform filtering manually:

```bash
# Clone repository
git clone https://github.com/plaesy/spec-kit.git /tmp/spec-kit-test

# Check script counts
find /tmp/spec-kit-test/scripts -name "*.sh" | wc -l    # Total bash scripts
find /tmp/spec-kit-test/scripts -name "*.ps1" | wc -l   # Total PowerShell scripts

# Compare with installed version
find ~/.plaesy/scripts -name "*.sh" | wc -l    # Should match if Linux/macOS
find ~/.plaesy/scripts -name "*.ps1" | wc -l   # Should be 0 if Linux/macOS
```

### Log Files

Key log locations for debugging:

- **Installation logs**: Check terminal output during install
- **CLI logs**: `~/.plaesy/logs/` (if implemented)
- **System logs**: `/var/log/` (Linux) or Event Viewer (Windows)

## Getting Help

If issues persist:

1. **Check version compatibility**: Ensure using latest release
2. **Search issues**: [GitHub Issues](https://github.com/plaesy/spec-kit/issues)
3. **Create bug report**: Include:
   - Operating system and version
   - Installation method used
   - Complete error messages
   - Output of `plaesy status`
   - Contents of debug logs

## Prevention

To avoid future issues:

1. **Use correct installer**: Always use platform-appropriate script
2. **Regular updates**: Keep framework updated
3. **Clean environment**: Avoid manual modifications
4. **Backup data**: Before major updates, backup `~/.plaesy/memory/`

---

**Note**: This troubleshooting guide covers the new platform-optimized installation. For legacy installations (pre-2.0), consider upgrading to benefit from these improvements.