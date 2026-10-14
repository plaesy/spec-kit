---
title: "Remove the internet_access dimension"
description: "The constitution governs nine dimensions. plaesy reach is tooling; its health check stays in plaesy doctor, outside the constitution."
updatedAt: "2026-10-14T05:40:00.000Z"
---

# Decision: Remove the internet_access dimension

## Context

An "Internet Access" dimension was added on 2026-10-14 so the template row would validate
(see [internet-access-dimension](internet-access-dimension.md), now superseded). On review,
`plaesy reach` is a set of tools for fetching external data. The work it does is not a
deliverable that a dimension would assess, and its health check is already a top-level command.

## Decision

- The validator's closed set is the nine dimensions again: technical, design, business, product,
  marketing, legal, financial, management, operations.
- The Internet Access row and its three reach bullets are removed from the template and from the
  project constitution.
- Platform health is checked with `plaesy doctor`, which the project already runs. It is not a
  constitution rule.
- The constitution moves to 2.0.0 (MAJOR), because it narrows the governed scope.

## Options considered

1. Keep the dimension and rename it `category`. Rejected: it keeps the same governance role
   under another name.
2. Fold it into `technical` or `operations`. Rejected: `technical` assesses code and `operations`
   assesses deployment; neither fits external data sources.

## References

- `scripts/internal/validate/constitution.go` (`knownDimensions`), accessed 2026-10-14
- `templates/constitution.template.md` §1, accessed 2026-10-14
- `scripts/cmd/plaesy/doctor.go` (top-level health check), accessed 2026-10-14
