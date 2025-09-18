#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
LLMS_FILE="$ROOT_DIR/llms.txt"

source "$ROOT_DIR/scripts/common.sh"

echo "Validate llms.txt" "Check links and exclusion rules"

validate_not_root

if [[ ! -f "$LLMS_FILE" ]]; then
  print_error "llms.txt not found at repository root."
fi

print_step "Checking format sections and last updated line"
grep -qE "^# " "$LLMS_FILE" || print_error "Missing H1 header."
grep -qE "^> " "$LLMS_FILE" || print_warning "Missing blockquote summary (optional)."
grep -qE "^Last updated: [0-9]{4}-[0-9]{2}-[0-9]{2}$" "$LLMS_FILE" || print_warning "Missing or invalid 'Last updated' line."

print_step "Ensuring no links to excluded internal paths"
if grep -E "\]\((\.plaesy|\.github/(copilot-instructions\.md|(chatmodes|commands)/))" -n "$LLMS_FILE"; then
  print_error "Found links to excluded internal paths in llms.txt."
fi

print_step "Validating that all relative links exist"
status=0
while IFS= read -r link; do
  rel="${link#*(}"
  rel="${rel%%)*}"
  # Skip non-file links (http/https) and anchors
  if [[ "$rel" =~ ^https?:// ]] || [[ "$rel" == *#* ]]; then
    continue
  fi
  path="$ROOT_DIR/$rel"
  if [[ ! -e "$path" ]]; then
    echo "Broken link: $rel" >&2
    status=1
  fi
done < <(grep -oE "\[[^\]]+\]\([^\)]+\)" "$LLMS_FILE")

if [[ $status -ne 0 ]]; then
  print_error "One or more links in llms.txt are broken."
fi

print_success "llms.txt validation passed."
