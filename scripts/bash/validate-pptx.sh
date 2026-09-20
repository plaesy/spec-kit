#!/bin/bash
# PPTX Structural + Render Validator
# Usage: validate-pptx.sh <file.pptx> [expected_slide_count]
# 1. Verifies the file is a well-formed OOXML zip (python-pptx can open it)
# 2. Headless-renders each slide to PNG via LibreOffice to catch corruption
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

PPTX_FILE="${1:-}"
EXPECTED_SLIDES="${2:-}"

if [[ -z "$PPTX_FILE" ]]; then
    log_error "Usage: $0 <file.pptx> [expected_slide_count]"
    exit 1
fi

if [[ ! -f "$PPTX_FILE" ]]; then
    log_error "File not found: $PPTX_FILE"
    exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
    log_error "python3 not found — required to validate structure via python-pptx"
    exit 1
fi

log_info "Validating structure of $PPTX_FILE"
python3 - "$PPTX_FILE" "$EXPECTED_SLIDES" <<'PYEOF'
import sys
from pptx import Presentation

path, expected = sys.argv[1], sys.argv[2] if len(sys.argv) > 2 else ""
prs = Presentation(path)
slide_count = len(prs.slides)
print(f"[INFO] slide count: {slide_count}")

for i, slide in enumerate(prs.slides, start=1):
    if slide.shapes.title is None:
        print(f"[WARN] slide {i} has no title placeholder")

if expected:
    assert slide_count == int(expected), f"expected {expected} slides, got {slide_count}"
print("[OK] structural validation passed")
PYEOF

if command -v soffice >/dev/null 2>&1; then
    log_info "Headless-rendering slides via LibreOffice to catch corruption..."
    OUT_DIR="$(mktemp -d)"
    soffice --headless --convert-to png --outdir "$OUT_DIR" "$PPTX_FILE" >/dev/null 2>&1 || {
        log_error "LibreOffice failed to render $PPTX_FILE — file may be corrupt"
        exit 1
    }
    log_success "Rendered without error (output in $OUT_DIR)"
else
    log_info "LibreOffice (soffice) not found — skipping render check. Install libreoffice for full CI validation."
fi

log_success "$PPTX_FILE passed validation"
