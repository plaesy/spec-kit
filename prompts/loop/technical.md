---
description: "Code quality, architecture, performance, security, tests, infrastructure, deployment"
---

# `/loop:technical` command instructions

This is the **technical**-scoped entry point into `/loop`. Follow the full protocol
defined in `/loop`, with scope fixed to **technical**: Code quality, architecture, performance, security, tests, infrastructure, deployment.

Same 8-dimension scope set as `/assess:{scope}` (see `/assess`) — treat
`$ARGUMENTS` (e.g. `--focus <sub-area>`) as narrowing within this dimension, not
as a dimension selector; do not broaden scope to other dimensions unless the user
explicitly asks with a comma-separated scope on `/loop` itself.
