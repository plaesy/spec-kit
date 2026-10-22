#!/bin/sh
# plaesy design-lint: advisory design check after a file edit. Never blocks; always exits 0.
# Reads the hook JSON on stdin. Lints the files the edit names. When the payload names
# none, it lints the frontend files git reports as changed. Needs the plaesy binary at the
# project root or on PATH; without it the hook does nothing.
root=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
input=$(cat)
named=$(printf '%s' "$input" | grep -oE '"file_path"[[:space:]]*:[[:space:]]*"[^"]*"' | sed -E 's/.*:[[:space:]]*"(.*)"$/\1/')
patched=$(printf '%s' "$input" | grep -oE '\*\*\* (Add|Update) File: [^\\"]+' | sed -E 's/^\*\*\* (Add|Update) File: //')
files=$(printf '%s\n%s\n' "$named" "$patched" | sed 's/\\\\/\//g' | sed '/^$/d')
if [ -z "$files" ]; then
  files=$(git -C "$root" status --porcelain 2>/dev/null | sed -E 's/^...//')
fi
bin="$root/plaesy"
[ -x "$bin" ] || bin="$root/plaesy.exe"
[ -x "$bin" ] || bin=$(command -v plaesy 2>/dev/null)
[ -n "$bin" ] || exit 0
printf '%s\n' "$files" | sort -u | while IFS= read -r f; do
  case "$f" in *.html|*.htm|*.css|*.scss|*.jsx|*.tsx|*.vue|*.svelte) ;; *) continue ;; esac
  p="$f"
  case "$p" in /*|?:*) ;; *) p="$root/$f" ;; esac
  [ -f "$p" ] || continue
  "$bin" validate design "$p" 2>&1 | grep -E '^(FAIL|WARN) '
done
exit 0
