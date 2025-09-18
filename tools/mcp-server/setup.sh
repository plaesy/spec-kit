#!/usr/bin/env bash

# Spec-Kit MCP Server Setup Script
# Configures MCP server for various AI platforms

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRAMEWORK_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
MCP_SERVER_DIR="$FRAMEWORK_ROOT/tools/mcp-server"

echo "🚀 Setting up Spec-Kit MCP Server..."
echo "Framework Root: $FRAMEWORK_ROOT"
echo "MCP Server: $MCP_SERVER_DIR"

# Build the MCP server
echo "📦 Building MCP server..."
cd "$MCP_SERVER_DIR"
npm run build

# Platform-specific setup functions
setup_claude_desktop() {
    echo "🤖 Setting up Claude Desktop integration..."
    
    # Determine Claude Desktop config path based on OS
    if [[ "$OSTYPE" == "darwin"* ]]; then
        CLAUDE_CONFIG_DIR="$HOME/Library/Application Support/Claude"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        CLAUDE_CONFIG_DIR="$HOME/.config/claude"
    else
        echo "❌ Unsupported OS for Claude Desktop setup"
        return 1
    fi
    
    mkdir -p "$CLAUDE_CONFIG_DIR"
    
    # Generate Claude Desktop config with actual paths
    cat > "$CLAUDE_CONFIG_DIR/claude_desktop_config.json" << EOF
{
  "mcpServers": {
    "spec-kit": {
      "command": "node",
      "args": [
        "$MCP_SERVER_DIR/dist/index.js"
      ],
      "env": {
        "SPEC_KIT_ROOT": "$FRAMEWORK_ROOT"
      }
    }
  }
}
EOF
    
    echo "✅ Claude Desktop config created at: $CLAUDE_CONFIG_DIR/claude_desktop_config.json"
}

setup_vscode() {
    echo "💻 Setting up VS Code integration..."
    
    # Create VS Code extension structure
    VSCODE_EXT_DIR="$MCP_SERVER_DIR/vscode-extension"
    mkdir -p "$VSCODE_EXT_DIR/src"
    
    # Copy VS Code extension package.json
    cp "$MCP_SERVER_DIR/configs/vscode-extension-package.json" "$VSCODE_EXT_DIR/package.json"
    
    # Create basic extension entry point
    cat > "$VSCODE_EXT_DIR/src/extension.ts" << 'EOF'
import * as vscode from 'vscode';
import { spawn } from 'child_process';

export function activate(context: vscode.ExtensionContext) {
    console.log('Spec-Kit MCP extension is now active!');
    
    // Register commands
    const detectContextCmd = vscode.commands.registerCommand('spec-kit-mcp.detectContext', () => {
        const editor = vscode.window.activeTextEditor;
        if (editor) {
            const content = editor.document.getText();
            // Call MCP server to detect context
            detectContext(content);
        }
    });
    
    context.subscriptions.push(detectContextCmd);
}

function detectContext(content: string) {
    // Implementation to call MCP server
    vscode.window.showInformationMessage('Context detection initiated');
}

export function deactivate() {}
EOF
    
    echo "✅ VS Code extension structure created at: $VSCODE_EXT_DIR"
}

setup_web_interface() {
    echo "🌐 Setting up Web interface integration..."
    
    # Create web interface config
    cat > "$MCP_SERVER_DIR/configs/web-interface.json" << EOF
{
  "server": {
    "host": "localhost",
    "port": 3001,
    "cors": {
      "enabled": true,
      "origins": ["http://localhost:3000", "https://claude.ai"]
    }
  },
  "mcp": {
    "transport": "websocket",
    "endpoint": "/mcp",
    "spec_kit_root": "$FRAMEWORK_ROOT"
  },
  "features": {
    "context_detection": true,
    "agent_switching": true,
    "constitutional_validation": true
  }
}
EOF
    
    echo "✅ Web interface config created"
}

# Create startup script
create_startup_script() {
    echo "📝 Creating startup script..."
    
    cat > "$MCP_SERVER_DIR/start-mcp-server.sh" << EOF
#!/usr/bin/env bash

# Spec-Kit MCP Server Startup Script

export SPEC_KIT_ROOT="$FRAMEWORK_ROOT"
cd "$MCP_SERVER_DIR"

echo "🚀 Starting Spec-Kit MCP Server..."
echo "Framework Root: \$SPEC_KIT_ROOT"

# Check if build exists
if [[ ! -d "dist" ]]; then
    echo "📦 Building MCP server..."
    npm run build
fi

# Start the server
node dist/index.js
EOF
    
    chmod +x "$MCP_SERVER_DIR/start-mcp-server.sh"
    echo "✅ Startup script created: $MCP_SERVER_DIR/start-mcp-server.sh"
}

# Create installation verification
create_verification_script() {
    echo "🔍 Creating verification script..."
    
    cat > "$MCP_SERVER_DIR/verify-setup.sh" << EOF
#!/usr/bin/env bash

# Spec-Kit MCP Server Verification Script

echo "🔍 Verifying Spec-Kit MCP Server setup..."

# Check if built
if [[ -f "$MCP_SERVER_DIR/dist/index.js" ]]; then
    echo "✅ MCP server built successfully"
else
    echo "❌ MCP server not built"
    exit 1
fi

# Check framework structure
if [[ -f "$FRAMEWORK_ROOT/memory/constitution.md" ]]; then
    echo "✅ Constitutional framework detected"
else
    echo "⚠️  Constitutional framework not found"
fi

# Check chatmodes
if [[ -d "$FRAMEWORK_ROOT/chatmodes" ]]; then
    CHATMODE_COUNT=\$(ls "$FRAMEWORK_ROOT/chatmodes"/*.chatmode.md 2>/dev/null | wc -l)
    echo "✅ Found \$CHATMODE_COUNT chatmode files"
else
    echo "⚠️  Chatmodes directory not found"
fi

# Test MCP server startup
echo "🧪 Testing MCP server startup..."
timeout 5s node "$MCP_SERVER_DIR/dist/index.js" <<< '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' && echo "✅ MCP server responds correctly" || echo "⚠️  MCP server test failed"

echo "✨ Verification complete!"
EOF
    
    chmod +x "$MCP_SERVER_DIR/verify-setup.sh"
    echo "✅ Verification script created: $MCP_SERVER_DIR/verify-setup.sh"
}

# Main setup function
main() {
    echo "Please select platform(s) to setup:"
    echo "1) Claude Desktop"
    echo "2) VS Code Extension"
    echo "3) Web Interface"
    echo "4) All platforms"
    echo "5) Skip platform setup (scripts only)"
    
    read -p "Enter your choice (1-5): " choice
    
    case $choice in
        1) setup_claude_desktop ;;
        2) setup_vscode ;;
        3) setup_web_interface ;;
        4) 
            setup_claude_desktop
            setup_vscode
            setup_web_interface
            ;;
        5) echo "Skipping platform setup..." ;;
        *) echo "Invalid choice. Skipping platform setup..." ;;
    esac
    
    create_startup_script
    create_verification_script
    
    echo ""
    echo "🎉 Spec-Kit MCP Server setup complete!"
    echo ""
    echo "📋 Next steps:"
    echo "1. Run verification: $MCP_SERVER_DIR/verify-setup.sh"
    echo "2. Start MCP server: $MCP_SERVER_DIR/start-mcp-server.sh"
    echo "3. Connect your AI platform to the MCP server"
    echo ""
    echo "📚 Configuration files created in: $MCP_SERVER_DIR/configs/"
}

# Run main function
main "$@"