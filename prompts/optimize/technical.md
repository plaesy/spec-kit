---
description: "Performance, code quality, database, backend/frontend"
---

# `/optimize:technical` command instructions

This is the **technical**-scoped entry point into `/optimize`. Follow the full protocol
defined in `/optimize`, with scope fixed to **technical**: Performance, code quality, database, backend/frontend.

Same 8-dimension scope set as `/assess:{scope}` (see `/assess`) — treat
`$ARGUMENTS` (e.g. `--focus <sub-area>`) as narrowing within this dimension, not
as a dimension selector; do not broaden scope to other dimensions unless the user
explicitly asks with a comma-separated scope on `/optimize` itself.
