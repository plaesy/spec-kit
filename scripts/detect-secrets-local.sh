#!/usr/bin/env bash
# Local Secrets Detection Script
# Part of Plaesy Constitutional Framework

set -e

echo "🔒 Scanning for secrets and sensitive data..."

# Load whitelist patterns if file exists
whitelist_file=".plaesyignore"

if [ -f "$whitelist_file" ]; then
    echo "📋 Loading secrets whitelist..."
    whitelist_patterns=$(cat "$whitelist_file" | grep -v '^#' | grep -v '^$' | tr '\n' '|' | sed 's/|$//')

    # Scan for secrets with more specific patterns
    secret_results=$(grep -rI \
        -E "(password\s*=\s*[\"'][^\"']{3,}[\"']|api[_-]?key\s*=\s*[\"'][^\"']{10,}[\"']|secret\s*=\s*[\"'][^\"']{10,}[\"'])" \
        --include="*.go" --include="*.js" --include="*.py" --include="*.dart" \
        --exclude-dir=node_modules --exclude-dir=vendor --exclude-dir=dist \
        . 2>/dev/null || true)

    # Filter out whitelisted patterns if any found
    if [ -n "$secret_results" ] && [ -n "$whitelist_patterns" ]; then
        secret_matches=$(echo "$secret_results" | grep -vE "$whitelist_patterns" || true)
    else
        secret_matches="$secret_results"
    fi
else
    # No whitelist, scan with specific patterns
    secret_matches=$(grep -rI \
        -E "(password\s*=\s*[\"'][^\"']{3,}[\"']|api[_-]?key\s*=\s*[\"'][^\"']{10,}[\"']|secret\s*=\s*[\"'][^\"']{10,}[\"'])" \
        --include="*.go" --include="*.js" --include="*.py" --include="*.dart" \
        --exclude-dir=node_modules --exclude-dir=vendor --exclude-dir=dist \
        . 2>/dev/null || true)
fi

# Check if we found any potential secrets
if [ -n "$secret_matches" ]; then
    echo "❌ Potential hardcoded secrets detected"

    # Count matches and show only file paths, not full content
    num_files=$(echo "$secret_matches" | cut -d: -f1 | sort -u | wc -l)
    num_matches=$(echo "$secret_matches" | wc -l)

    echo "📊 Found $num_matches issues in $num_files files"
    echo "📁 Affected files:"
    echo "$secret_matches" | cut -d: -f1 | sort -u | sed 's/^/  • /'
    echo ""
    echo "💡 Use environment variables or secure credential storage"
    exit 1
fi

echo "✅ No secrets detected"
