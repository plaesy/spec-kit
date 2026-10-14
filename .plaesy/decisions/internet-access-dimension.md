---
title: "Internet Access as a tenth dimension (superseded)"
description: "Superseded on 2026-10-14 by remove-internet-access-dimension. Kept for the record of why the row was briefly accepted."
updatedAt: "2026-10-14T05:40:00.000Z"
---

# Decision: Internet Access as a tenth dimension (superseded)

**Status:** Superseded by [remove-internet-access-dimension](remove-internet-access-dimension.md).

## Context

`templates/constitution.template.md` shipped an "Internet Access" row (§1), but the validator
accepted only nine dimensions, so `/start` output failed `plaesy validate constitution`.

## What was done

The validator accepted `internet_access` as a tenth dimension, with the §1 label normalized
to `internet_access`. The gate was `plaesy doctor`.

## Why it was reversed

`plaesy reach` is a set of tools for fetching external data. It is not a deliverable that
needs a dimension of its own. Its health check already lives in `plaesy doctor`.

## References

- `.plaesy/memory/constitution.md` §9 amendment log, versions 1.1.0, 1.1.1, and 2.0.0, accessed 2026-10-14
