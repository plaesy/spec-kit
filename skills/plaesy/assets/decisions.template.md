---
title: "Decisions Index"
description: "Index of project decisions with traceability and access timestamps"
updatedAt: "{{UPDATED_AT|<ISO-8601 timestamp, e.g. 2026-01-31T09:00:00Z>}}"
---

# Decisions Index

**Project:** {{PROJECT_NAME}}
**Last Updated:** {{UPDATED_AT|<ISO-8601 timestamp, e.g. 2026-01-31T09:00:00Z>}}
**Purpose:** Index of project decisions with traceability and access timestamps

---

## How This Index Works

This file is an index only — it links to decision topic files under
`.plaesy/decisions/`. Each entry records:

- **Topic** — the decision subject (matches the filename)
- **Decision** — the chosen outcome in one line
- **References** — links to source material that informed or supports the decision
- **Accessed** — timestamp when the reference was last consulted

Add a link here only once the file it points at exists. Keep descriptions under
100 chars. No nested bullets or sub-links — flat list only.

---

## Active Decisions

<!-- One line per decision topic file. Format:

- DECISION_TITLE → decisions/TOPIC.md — ONE_LINE_DECISION · DATE

Each entry MUST include at least one reference link and an access timestamp.
A decision with no recorded reference or access time is incomplete.
-->

- {{DECISION_TITLE}} → create `decisions/{{TOPIC}}.md`, then link it here as
  `- [{{DECISION_TITLE}}](decisions/{{TOPIC}}.md) — {{ONE_LINE_DECISION}} · {{DATE}}`

---

## Superseded / Deprecated Decisions

<!-- Move entries here when a decision is replaced. Keep the original file;
     only the index entry changes. -->

- {{SUPERSEDED_TITLE}} → `decisions/{{TOPIC}}.md` (superseded by {{NEW_DECISION_LINK}})

---

## Reference Information

**Decision files**: `.plaesy/decisions/{TOPIC}.md` (flat structure, one topic per file)
**Index**: `.plaesy/decisions.md` (this file)
**Related**: `.plaesy/memory.md` — knowledge base for patterns learned from decisions

---

**Status:** Decisions index created
**Next:** Add project-specific decision entries as they arise; every link above must resolve.
