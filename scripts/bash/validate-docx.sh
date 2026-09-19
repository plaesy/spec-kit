#!/bin/bash
# DOCX Structural + Render Validator
# Usage: validate-docx.sh <file.docx>
# 1. Verifies the file is a well-formed OOXML zip (python-docx can open it)
# 2. Checks for leftover unfilled Jinja2 placeholder tags ({{ }})
# 3. Headless-renders the document via LibreOffice to catch corruption
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

DOCX_FILE="${1:-}"

if [[ -z "$DOCX_FILE" ]]; then
    log_error "Usage: $0 <file.docx>"
    exit 1
fi

if [[ ! -f "$DOCX_FILE" ]]; then
    log_error "File not found: $DOCX_FILE"
    exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
    log_error "python3 not found — required to validate structure via python-docx"
    exit 1
fi

log_info "Validating structure of $DOCX_FILE"
python3 - "$DOCX_FILE" <<'PYEOF'
import re
import sys
from docx import Document

path = sys.argv[1]
doc = Document(path)

paragraph_count = len(doc.paragraphs)
table_count = len(doc.tables)
print(f"[INFO] paragraphs: {paragraph_count}, tables: {table_count}")

if paragraph_count == 0 and table_count == 0:
    raise AssertionError("document has no paragraphs or tables")

leftover = [p.text for p in doc.paragraphs if re.search(r"\{\{.*?\}\}", p.text)]
if leftover:
    raise AssertionError(f"leftover unfilled template tags found: {leftover}")

print("[OK] structural validation passed")
PYEOF

if command -v soffice >/dev/null 2>&1; then
    log_info "Headless-rendering document via LibreOffice to catch corruption..."
    OUT_DIR="$(mktemp -d)"
    soffice --headless --convert-to pdf --outdir "$OUT_DIR" "$DOCX_FILE" >/dev/null 2>&1 || {
        log_error "LibreOffice failed to render $DOCX_FILE — file may be corrupt"
        exit 1
    }
    log_success "Rendered without error (output in $OUT_DIR)"
else
    log_info "LibreOffice (soffice) not found — skipping render check. Install libreoffice for full CI validation."
fi

log_success "$DOCX_FILE passed validation"
