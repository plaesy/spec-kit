---
description: "Design tokens, visual consistency, WCAG, dark mode"
---

# `/optimize:design` command instructions

This is the **design**-scoped entry point into `/optimize`. Follow the full protocol
defined in `/optimize`, with scope fixed to **design**: Design tokens, visual consistency, WCAG, dark mode.

Same 8-dimension scope set as `/assess:{scope}` (see `/assess`) — treat
`$ARGUMENTS` (e.g. `--focus <sub-area>`) as narrowing within this dimension, not
as a dimension selector; do not broaden scope to other dimensions unless the user
explicitly asks with a comma-separated scope on `/optimize` itself.
