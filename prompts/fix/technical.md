---
description: "Code bugs, runtime errors, integration failures"
---

# `/fix:technical` command instructions

This is the **technical**-scoped entry point into `/fix`. Follow the full protocol
defined in `/fix`, with scope fixed to **technical**: Code bugs, runtime errors, integration failures.

Same 8-dimension scope set as `/assess:{scope}` (see `/assess`) — treat
`$ARGUMENTS` (e.g. `--focus <sub-area>`) as narrowing within this dimension, not
as a dimension selector; do not broaden scope to other dimensions unless the user
explicitly asks with a comma-separated scope on `/fix` itself.
