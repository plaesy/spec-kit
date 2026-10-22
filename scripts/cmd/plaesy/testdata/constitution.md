---
title: "Test Project Constitution"
version: 1.0.0
ratified: 2026-01-15
last_amended: 2026-01-15
active_dimensions: [technical, product]
---

# Test Project Constitution

Fixture for `TestBareValidate*` in `cmd/plaesy`. It is committed here rather
than read from `.plaesy/memory/constitution.md` on purpose: that path holds
generated, untracked project state, so a test that reads it fails on any clean
checkout and after `make clean` — the suite has to be hermetic.

This document satisfies the contract in `internal/validate/constitution.go`:
frontmatter keys present, ISO dates, no unfilled template placeholders, at least
one active dimension, frontmatter `active_dimensions` equal to the section 1 rows
marked `yes`, frontmatter `version` equal to the last Amendment Log row, and
unique rule IDs.

## 1. Active Dimensions

| Dimension | Active | Rationale |
|---|---|---|
| technical | yes | the fixture exercises code-facing checks |
| design | no | out of scope for the fixture |
| business | no | out of scope for the fixture |
| product | yes | the fixture exercises a user-facing check |
| marketing | no | out of scope for the fixture |
| legal | no | out of scope for the fixture |
| financial | no | out of scope for the fixture |
| management | no | out of scope for the fixture |

## 2. Rules

| ID | Rule | Level | Detection Anchor |
|---|---|---|---|
| EV-1 | Every claim carries a source and a retrieval date | blocking | an unsourced assertion in the output |
| QG-1 | The build is green before a phase is called done | blocking | non-zero exit from the build |
| DOC-1 | Public behaviour is documented | advisory | a changed exported symbol without a doc edit |

## 3. Non-Negotiables

| ID | Non-Negotiable | Detection Anchor |
|---|---|---|
| NN-1 | Never commit a credential | a secret-shaped literal in a tracked file |

## 4. Amendment Log

| Version | Date | Change |
|---|---|---|
| 1.0.0 | 2026-01-15 | initial ratification |
