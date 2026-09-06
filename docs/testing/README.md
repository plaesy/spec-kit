# Testing Framework

This directory contains comprehensive testing scripts for the Plaesy Spec-Kit framework.

## Scripts

### Bash Testing
- **File**: `bash/run.sh`
- **Usage**: `./testing/bash/run.sh [options]`
- **Description**: Testing script for Unix/Linux/macOS environments
- **Dependencies**: none beyond built-in bash commands

#### Options:
- `-h, --help`: Show help message
- `-v, --verbose`: Enable verbose output
- `-q, --quiet`: Suppress non-error output
- `--syntax-only`: Run only syntax checks
- `--no-functional`: Skip functional tests
- `--integration-only`: Run only integration tests

#### Test Categories:
- **Static Tests**: Directory, config, permissions, syntax validation
- **Functional Tests**: Script functionality validation
- **Integration Tests**: Install, init, analyze, end-to-end workflow
- **Utility Tests**: Clean, feature creation, context management
- **Platform Tests**: AI platform detection and configuration
- **Mapping Tests**: `plaesy init` output mapping, validated against real script execution (no simulation) across all 19 supported AI platforms

Run `./testing/bash/run.sh` for the current pass/fail count — the script prints `Total Tests: N` at the end; don't rely on a number in this doc, it drifts as tests are added.

- **Required Commands**: bash, curl, git, find, grep, sed, awk (built-in Unix commands)

#### Examples:
```bash
# Run all tests
./testing/bash/run.sh

# Run only syntax checks
./testing/bash/run.sh --syntax-only

# Run only integration tests
./testing/bash/run.sh --integration-only

# Run with verbose output
./testing/bash/run.sh --verbose

# Show help
./testing/bash/run.sh --help
```

### PowerShell Testing
- **File**: `powershell/run.ps1`
- **Usage**: `.\testing\powershell\run.ps1 [options]`
- **Description**: Testing script for Windows environments
- **Dependencies**: none beyond standard PowerShell cmdlets

#### Options:
- `-Help`: Show help message
- `-Verbose`: Enable verbose output
- `-Quiet`: Suppress non-error output
- `-SyntaxOnly`: Run only syntax checks
- `-NoFunctional`: Skip functional tests
- `-IntegrationOnly`: Run only integration tests

#### Test Categories:
- **Static Tests**: Directory, config, permissions, syntax validation
- **Functional Tests**: Script functionality validation
- **Integration Tests**: Install, init, analyze, end-to-end workflow
- **Utility Tests**: Clean, feature creation, context management, helper utilities
- **Platform Tests**: AI platform detection and configuration
- **Mapping Tests**: `plaesy init` output mapping, validated against real script execution (no simulation) across all 19 supported AI platforms

Run `.\testing\powershell\run.ps1` for the current pass/fail count.

- **Required Commands**: PowerShell, curl, git, find, grep, sed, awk (built-in/standard commands)

#### Examples:
```powershell
# Run all tests
.\testing\powershell\run.ps1

# Run only syntax checks
.\testing\powershell\run.ps1 -SyntaxOnly

# Run only integration tests
.\testing\powershell\run.ps1 -IntegrationOnly

# Run with verbose output
.\testing\powershell\run.ps1 -Verbose

# Show help
.\testing\powershell\run.ps1 -Help
```

## Test Categories

### 1. Structure Tests
- Directory structure validation
- Configuration file existence checks
- Script permission validation

### 2. Syntax Tests
- JSON configuration syntax validation
- Bash script syntax validation (Linux/macOS)
- PowerShell script syntax validation (Windows)

### 3. System Tests
- Required command availability
- Git repository status checks
- Version consistency validation

### 4. Functional Tests
- Script execution validation
- Help functionality testing
- Basic functionality verification

### 5. Platform Tests
- AI platform configuration validation
- Platform detection pattern testing
- File mapping configuration verification

### 6. Mapping Tests
- Plaesy init output mapping validation with **REAL script execution**
- Platform-specific file structure creation verification
- Core file, instructions, prompts, and chatmodes mapping verification
- Content validation for AI platform-specific outputs
- **NO SIMULATION** - tests execute actual plaesy-init script

## Exit Codes

- `0`: All tests passed
- `1`: One or more tests failed

## Output Format

The testing framework uses color-coded output:
- **[INFO]**: Blue informational messages
- **[PASS]**: Green successful tests
- **[FAIL]**: Red failed tests
- **[SKIP]**: Yellow skipped tests
- **[TEST]**: Blue test announcements

## Integration with CI/CD

These scripts are designed to be easily integrated into CI/CD pipelines:

### GitHub Actions Example
```yaml
- name: Run framework tests
  run: ./testing/bash/run.sh --syntax-only
```

### Azure Pipelines Example
```yaml
- script: ./testing/bash/run.sh
  displayName: 'Run Plaesy Framework Tests'
```

## Troubleshooting

### Common Issues

1. **Permission Denied**: Ensure the bash script is executable:
   ```bash
   chmod +x testing/bash/run.sh
   ```

2. **Missing Commands**: Framework has **ZERO external dependencies** - uses only built-in commands:
   ```bash
   # Required commands (built-in to bash/Unix systems):
   # bash, curl, git, find, grep, sed, awk
   # All are typically pre-installed on modern systems

   # Required commands (PowerShell/Windows systems):
   # PowerShell, curl, git, find, grep, sed, awk
   # All are standard or available through built-in PowerShell cmdlets
   ```

3. **PowerShell Execution Policy**: Set appropriate execution policy:
   ```powershell
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   ```

### Test Results Interpretation

- **Failed tests**: Indicate serious issues that should be resolved
- **Skipped tests**: Usually due to missing optional dependencies
- **Passed tests**: Everything is working correctly

For more information about the Plaesy Spec-Kit framework, see the main [README.md](../README.md) file.