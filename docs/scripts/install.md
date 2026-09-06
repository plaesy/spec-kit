# 📦 install.sh

**System-wide Plaesy Spec-Kit installation and setup script.**

**Priority:** 🟢 **HIGH** - Required for framework installation

## 🎯 Purpose

Automated installation script that sets up Plaesy Spec-Kit system-wide with proper PATH configuration, dependency checking, and validation. This is the primary method for installing Plaesy on any system.

## 📋 Usage

### Quick Installation
```bash
# One-line installation (recommended)
curl -s https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh | bash

# Or download and run manually
curl -O https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh
chmod +x install.sh
./install.sh
```

### Custom Installation
```bash
# Install to custom directory
INSTALL_DIR="$HOME/.local/plaesy" ./install.sh

# Install with verbose output
VERBOSE=1 ./install.sh

# Install without validation (not recommended)
SKIP_VALIDATION=1 ./install.sh
```

### Development Installation
```bash
# Install from local repository
cd /path/to/spec-kit
./scripts/bash/install.sh

# Install development version
DEV_MODE=1 ./scripts/bash/install.sh
```

## 🔧 Key Features

### Cross-Platform Support
- **Linux**: Ubuntu, Debian, CentOS, RHEL, Fedora, Arch Linux
- **macOS**: Homebrew and manual installation
- **Windows**: WSL support (Windows Subsystem for Linux)
- **Universal**: Works with most Unix-like systems

### Dependency Management
- **Required Tools**: git, curl, jq, find, sed, awk
- **Optional Tools**: bat (for enhanced output), tree (for visualization)
- **Package Managers**: apt, yum, dnf, pacman, zypper, brew
- **Validation**: Checks for required tools before installation

### Directory Structure Creation
```bash
# Creates standardized directory structure:
~/.plaesy/
├── scripts/
│   ├── bash/
│   │   ├── plaesy-init.sh
│   │   ├── config-manager.sh
│   │   ├── plaesy-clean.sh
│   │   ├── plaesy-analyze.sh
│   │   └── [other scripts]
│   ├── powershell/
│   └── configs/
│       └── platform.json
├── templates/
│   ├── instructions/
│   ├── prompts/
│   ├── chatmodes/
│   ├── ai-headers/
│   └── agent-file-template.md
└── memory/
    └── analysis/
```

### Safe Installation
- **Non-Destructive**: Won't overwrite existing configurations
- **Backup Creation**: Backs up existing Plaesy installations
- **Validation**: Tests installation after completion
- **Rollback Support**: Easy to uninstall if needed

## 🚀 Installation Process

### Step 1: Environment Validation
```bash
# Checks for:
✓ Required tools (git, curl)
✓ Package manager availability
✓ Write permissions in home directory
✓ Network connectivity for downloads
```

### Step 2: Directory Setup
```bash
# Creates directories:
mkdir -p ~/.plaesy/{scripts/{bash,powershell},configs,templates/{instructions,prompts,chatmodes,ai-headers},memory/analysis}

# Sets permissions:
chmod 755 ~/.plaesy
chmod +x ~/.plaesy/scripts/bash/*.sh
```

### Step 3: Script Installation
```bash
# Downloads core scripts:
- plaesy-init.sh
- config-manager.sh
- plaesy-clean.sh
- plaesy-analyze.sh
- [and other essential scripts]

# Downloads configuration:
- platform.json
- agent-file-template.md
```

### Step 4: PATH Configuration
```bash
# Updates shell configuration files:
- ~/.bashrc (for bash)
- ~/.zshrc (for zsh)
- ~/.profile (fallback)

# Adds to PATH:
export PATH="$HOME/.local/bin:$PATH"
```

### Step 5: Wrapper Creation
```bash
# Creates executable wrapper:
~/.local/bin/plaesy -> ~/.plaesy/scripts/bash/plaesy-init.sh

# Makes executable:
chmod +x ~/.local/bin/plaesy
```

### Step 6: Validation
```bash
# Tests installation:
✓ plaesy command available
✓ Scripts executable
✓ Configuration valid
✓ PATH updated correctly
```

## 📊 Installation Examples

### Standard Installation
```bash
$ curl -s https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/bash/install.sh | bash
🔍 Detecting environment...
✓ Operating System: Linux
✓ Shell: bash
✓ Package Manager: apt

📦 Checking dependencies...
✓ git: found
✓ curl: found
✓ jq: found
⚠️ bat: not found (optional)

📁 Creating directory structure...
✓ Created ~/.plaesy/
✓ Created ~/.plaesy/scripts/
✓ Created ~/.plaesy/templates/
...

📥 Downloading Plaesy files...
✓ Downloaded plaesy-init.sh
✓ Downloaded config-manager.sh
✓ Downloaded platform.json
...

🔧 Setting up PATH...
✓ Added to ~/.bashrc
✓ Wrapper created: ~/.local/bin/plaesy

✅ Installation complete!
🎯 Next steps:
  1. Run: source ~/.bashrc
  2. Test: plaesy --help
  3. Initialize: plaesy init . --ai claude
```

### Development Installation
```bash
$ cd /path/to/spec-kit
$ ./scripts/bash/install.sh
🔧 Development mode detected
📁 Using local repository instead of downloads
✓ Scripts linked from local repository
✓ Configuration linked from local repository
✅ Development installation complete!
```

### Custom Installation Directory
```bash
$ INSTALL_DIR="$HOME/tools/plaesy" ./install.sh
📁 Installing to custom directory: /home/user/tools/plaesy
✓ Created custom directory structure
✓ Scripts installed to custom location
⚠️ PATH not updated automatically
🎯 Add to PATH: export PATH="$HOME/tools/plaesy/scripts/bash:$PATH"
```

## 🪟 Windows PowerShell Installation

### Automated Installation (Recommended)
```powershell
# Run installer directly from web
iwr -useb https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/powershell/install.ps1 | iex

# Or download and run manually
iwr -useb https://raw.githubusercontent.com/plaesy/spec-kit/main/scripts/powershell/install.ps1 -OutFile install.ps1
.\install.ps1
```

### Manual Installation
```powershell
# Clone or download the repository
git clone https://github.com/plaesy/spec-kit.git
cd spec-kit

# Run PowerShell installer
.\scripts\powershell\install.ps1

# Or run specific scripts directly
.\scripts\powershell\plaesy-init.ps1 -Target . -AI claude_code
```

### Windows Installation Features
- **Batch File Creation**: Creates `plaesy.bat` in `%USERPROFILE%\.local\bin\`
- **PATH Configuration**: Automatically adds to Windows PATH
- **PowerShell Scripts**: Full PowerShell equivalents for all bash scripts
- **Cross-Platform**: Works on Windows 10/11 with PowerShell 5.1+
- **Professional Output**: Clean text output without emojis

### Available Commands After Installation
```powershell
# Windows PowerShell CLI
plaesy init              # Interactive AI selection in current directory
plaesy init <directory>  # Interactive AI selection in specified directory
plaesy init <directory> -AI <platform>    # Use specific AI platform in directory
plaesy analyze           # Analyze current project structure and generate documentation
plaesy clean <directory> # Clean specified directory (default: current)
plaesy upgrade           # Upgrade framework
plaesy status            # Check installation status and system information
plaesy repair            # Fix missing components and scripts
plaesy uninstall         # Remove Plaesy Spec-Kit completely
```

### Windows Validation
```powershell
# Test installation
plaesy --help

# Check installation location
Get-Command plaesy

# Validate framework
plaesy status

# Test analyze command
plaesy analyze
```

### Windows Troubleshooting

#### PowerShell Execution Policy
```powershell
# Set execution policy for current user
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Or bypass for single script
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

#### PATH Issues
```powershell
# Check if plaesy.bat is in PATH
where plaesy

# Manually add to PATH (temporary)
$env:PATH = "$env:PATH;$env:USERPROFILE\.local\bin"

# Add to PATH permanently
[Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";$env:USERPROFILE\.local\bin", "User")
```

#### Windows Defender Issues
If Windows Defender blocks the installation:
```powershell
# Add exclusion for Plaesy directory
Add-MpPreference -ExclusionPath "$env:USERPROFILE\.plaesy"
Add-MpPreference -ExclusionPath "$env:USERPROFILE\.local\bin"
```

## 🛡️ Post-Installation Validation

### Quick Validation
```bash
# Test installation
plaesy --help

# Check installation location
which plaesy
ls -la ~/.plaesy/

# Validate configuration
~/.plaesy/scripts/bash/config-manager.sh validate-config
```

### Comprehensive Validation
```bash
#!/bin/bash
# validate-installation.sh

echo "🔍 Validating Plaesy installation..."

# Check command availability
if command -v plaesy >/dev/null 2>&1; then
    echo "✅ plaesy command available"
    echo "📍 Location: $(which plaesy)"
else
    echo "❌ plaesy command not found"
    echo "💡 Try: source ~/.bashrc or restart terminal"
fi

# Check directory structure
echo ""
echo "📁 Directory structure:"
if [ -d ~/.plaesy ]; then
    echo "✅ ~/.plaesy/ exists"
    ls -la ~/.plaesy/scripts/bash/ | head -5
else
    echo "❌ ~/.plaesy/ missing"
fi

# Check configuration
echo ""
echo "⚙️ Configuration:"
if ~/.plaesy/scripts/bash/config-manager.sh validate-config >/dev/null 2>&1; then
    echo "✅ Configuration valid"
else
    echo "❌ Configuration invalid"
fi

# Test core functionality
echo ""
echo "🧪 Functionality test:"
if ~/.plaesy/scripts/bash/config-manager.sh list-platforms >/dev/null 2>&1; then
    echo "✅ Core scripts working"
else
    echo "❌ Core scripts not working"
fi

echo ""
echo "🎉 Validation complete!"
```

## 🔧 Environment Variables

### Installation Options
```bash
# Custom installation directory
INSTALL_DIR="/custom/path" ./install.sh

# Verbose output
VERBOSE=1 ./install.sh

# Skip validation (not recommended)
SKIP_VALIDATION=1 ./install.sh

# Development mode
DEV_MODE=1 ./install.sh

# Force overwrite existing installation
FORCE=1 ./install.sh
```

### Runtime Configuration
```bash
# Plaesy home directory
export PLAESA_HOME="$HOME/.plaesy"

# Custom scripts directory
export PLAESA_SCRIPTS_DIR="$HOME/.plaesy/scripts/bash"

# Debug mode
export PLAESA_DEBUG=1
```

## 🆘 Troubleshooting

### Common Issues

#### Permission Denied
```bash
# Symptoms:
Permission denied during installation

# Solutions:
# Check permissions on home directory
ls -la ~

# Fix permissions if needed
chmod 755 ~

# Use sudo only if absolutely necessary
sudo ./install.sh
```

#### Missing Dependencies
```bash
# Symptoms:
jq: command not found during installation

# Solutions:
# Install missing dependencies manually
sudo apt-get install jq  # Ubuntu/Debian
sudo yum install jq      # CentOS/RHEL
brew install jq         # macOS

# Or download jq directly
wget https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64
chmod +x jq
sudo mv jq /usr/local/bin/
```

#### PATH Not Updated
```bash
# Symptoms:
plaesy: command not found after installation

# Solutions:
# Reload shell configuration
source ~/.bashrc
# or
source ~/.zshrc

# Manual PATH addition
export PATH="$HOME/.local/bin:$PATH"

# Add to shell config permanently
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Debug Mode

```bash
# Enable debug mode
DEBUG=1 ./install.sh

# Or run with bash debugging
bash -x install.sh

# Check installation log
cat install.log 2>/dev/null || echo "No log file found"
```

## 📞 Support

### Documentation Resources
- **Main Documentation**: [../../README.md](../../README.md)
- **Scripts Documentation**: [./README.md](./README.md)
- **Troubleshooting**: Framework troubleshooting guide

### Getting Help
1. **Check this guide first** - Most issues are covered here
2. **Run validation** - Use validation commands to ensure success
3. **Check GitHub Issues** - [github.com/plaesy/spec-kit/issues](https://github.com/plaesy/spec-kit/issues)
4. **Join Discussions** - [github.com/plaesy/spec-kit/discussions](https://github.com/plaesy/spec-kit/discussions)

---

**📦 This installation script provides a complete, automated setup for Plaesy Spec-Kit with cross-platform support and comprehensive validation.**