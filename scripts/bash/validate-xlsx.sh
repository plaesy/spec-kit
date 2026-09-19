#!/bin/bash
# XLSX Structural + Render Validator
# Usage: validate-xlsx.sh <file.xlsx> [expected_sheet_name...]
# 1. Verifies the file is a well-formed OOXML zip (openpyxl can open it)
# 2. Headless-renders the workbook via LibreOffice to catch corruption
#    that the object model alone would not surface

set -euo pipefail
IFS=$' \n\t'

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${CYAN}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }

XLSX_FILE="${1:-}"
shift || true
EXPECTED_SHEETS=("$@")

if [[ -z "$XLSX_FILE" ]]; then
    log_error "Usage: $0 <file.xlsx> [expected_sheet_name...]"
    exit 1
fi

if [[ ! -f "$XLSX_FILE" ]]; then
    log_error "File not found: $XLSX_FILE"
    exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
    log_error "python3 not found — required to validate structure via openpyxl"
    exit 1
fi

log_info "Validating structure of $XLSX_FILE"
python3 - "$XLSX_FILE" "${EXPECTED_SHEETS[@]:-}" <<'PYEOF'
import sys
from openpyxl import load_workbook

path = sys.argv[1]
expected_sheets = sys.argv[2:]

wb = load_workbook(path, read_only=True)
sheet_names = wb.sheetnames
print(f"[INFO] sheets: {sheet_names}")

if not sheet_names:
    raise AssertionError("workbook has no sheets")

for name in expected_sheets:
    assert name in sheet_names, f"expected sheet '{name}' not found in {sheet_names}"

for name in sheet_names:
    ws = wb[name]
    _ = ws.max_row, ws.max_column

print("[OK] structural validation passed")
PYEOF

if command -v soffice >/dev/null 2>&1; then
    log_info "Headless-rendering workbook via LibreOffice to catch corruption..."
    OUT_DIR="$(mktemp -d)"
    soffice --headless --convert-to pdf --outdir "$OUT_DIR" "$XLSX_FILE" >/dev/null 2>&1 || {
        log_error "LibreOffice failed to render $XLSX_FILE — file may be corrupt"
        exit 1
    }
    log_success "Rendered without error (output in $OUT_DIR)"
else
    log_info "LibreOffice (soffice) not found — skipping render check. Install libreoffice for full CI validation."
fi

log_success "$XLSX_FILE passed validation"
