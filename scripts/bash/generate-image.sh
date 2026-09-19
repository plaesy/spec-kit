#!/usr/bin/env bash
# Generate an image asset via a configured provider API and save it to disk.
# Used by prompts/generate/images.md (/generate:images). Requires curl + either
# `jq` (preferred) or `python3` (fallback) to build/parse JSON, and `base64`.

set -euo pipefail
IFS=$' \n\t'

PROVIDER="${PLAESY_IMAGE_PROVIDER:-openai}"
SIZE="1024x1024"
OUT=""
PROMPT=""

usage() {
    echo "Usage: $0 --prompt \"<text>\" [--provider openai|gemini] [--size 1024x1024] --out <path>" >&2
    exit 1
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --prompt) PROMPT="$2"; shift 2 ;;
        --provider) PROVIDER="$2"; shift 2 ;;
        --size) SIZE="$2"; shift 2 ;;
        --out) OUT="$2"; shift 2 ;;
        *) echo "Unknown argument: $1" >&2; usage ;;
    esac
done

[[ -z "$PROMPT" || -z "$OUT" ]] && usage

mkdir -p "$(dirname "$OUT")"

case "$PROVIDER" in
    openai)
        if [[ -z "${OPENAI_API_KEY:-}" ]]; then
            echo "ERROR: OPENAI_API_KEY is not set. Export it, e.g.:" >&2
            echo "  export OPENAI_API_KEY=sk-..." >&2
            exit 2
        fi
        RESPONSE=$(curl -sS https://api.openai.com/v1/images/generations \
            -H "Authorization: Bearer ${OPENAI_API_KEY}" \
            -H "Content-Type: application/json" \
            -d "$(python3 - "$PROMPT" "$SIZE" <<'PY'
import json, sys
print(json.dumps({"model": "gpt-image-1", "prompt": sys.argv[1], "size": sys.argv[2], "n": 1}))
PY
)")
        ERR=$(python3 - "$RESPONSE" <<'PY'
import json, sys
d = json.loads(sys.argv[1])
print(d.get("error", {}).get("message", "") if isinstance(d, dict) else "")
PY
)
        if [[ -n "$ERR" ]]; then
            echo "ERROR: OpenAI image API: $ERR" >&2
            exit 3
        fi
        python3 - "$RESPONSE" "$OUT" <<'PY'
import json, base64, sys
d = json.loads(sys.argv[1])
b64 = d["data"][0]["b64_json"]
with open(sys.argv[2], "wb") as f:
    f.write(base64.b64decode(b64))
PY
        ;;
    gemini)
        if [[ -z "${GEMINI_API_KEY:-}" ]]; then
            echo "ERROR: GEMINI_API_KEY is not set. Export it, e.g.:" >&2
            echo "  export GEMINI_API_KEY=..." >&2
            exit 2
        fi
        RESPONSE=$(curl -sS "https://generativelanguage.googleapis.com/v1beta/models/imagen-3.0-generate-002:predict?key=${GEMINI_API_KEY}" \
            -H "Content-Type: application/json" \
            -d "$(python3 - "$PROMPT" <<'PY'
import json, sys
print(json.dumps({"instances": [{"prompt": sys.argv[1]}], "parameters": {"sampleCount": 1}}))
PY
)")
        ERR=$(python3 - "$RESPONSE" <<'PY'
import json, sys
d = json.loads(sys.argv[1])
print(d.get("error", {}).get("message", "") if isinstance(d, dict) else "")
PY
)
        if [[ -n "$ERR" ]]; then
            echo "ERROR: Gemini image API: $ERR" >&2
            exit 3
        fi
        python3 - "$RESPONSE" "$OUT" <<'PY'
import json, base64, sys
d = json.loads(sys.argv[1])
b64 = d["predictions"][0]["bytesBase64Encoded"]
with open(sys.argv[2], "wb") as f:
    f.write(base64.b64decode(b64))
PY
        ;;
    *)
        echo "ERROR: Unknown provider '$PROVIDER' (supported: openai, gemini)" >&2
        exit 4
        ;;
esac

if [[ ! -s "$OUT" ]]; then
    echo "ERROR: $OUT was not written or is empty" >&2
    exit 5
fi

echo "$OUT"
