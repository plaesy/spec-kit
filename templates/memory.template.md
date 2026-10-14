---
title: "Memory & Knowledge Index"
description: "Central index for project memory, guidance, and reference"
updatedAt: "{{UPDATED_AT|<ISO-8601 timestamp, e.g. 2026-01-31T09:00:00Z>}}"
---

# Memory Index & Navigation

**Project:** {{PROJECT_NAME}}
**Last Updated:** {{UPDATED_AT|<ISO-8601 timestamp, e.g. 2026-01-31T09:00:00Z>}}
**Purpose:** Central index for project memory, guidance, and reference

---

## 🎯 Core Reference

This index starts empty and must stay that way: it is a navigation surface, not a
container. Add a link here only once the file it points at exists.

- Project overview → create `memory/overview.md`, then link it here as
  `- [Project Overview](memory/overview.md) — Complete framework summary, architecture, key concepts`

---

## 📋 Guidance & Feedback

### Established Rules (Follow These)

{{RULE}} — {{WHY|the reason this rule exists: a past bug it prevents, an invariant it protects}}

<!-- One bullet per rule, imperative, each with its own "why". A rule with no
     stated reason is a demand the agent cannot generalise from — see
     instructions/how-to-create-instructionsmd.instructions.md. Add a rule here
     only once it has actually been learned, not preemptively. -->

---

## 📌 Reference Information

### Platform-Agnostic Paths (After plaesy init)

- Instructions → `.plaesy/instructions/[name].md` (flat structure)
- Agents → `.plaesy/roles/[name].md` (flat structure, platform-agnostic)
- Scripts → `.plaesy/scripts/{powershell,bash}/`
- Configs → `.plaesy/scripts/configs/`
- Analysis → `.plaesy/analysis/`
- Tasks → `.plaesy/tasks/{backlog,todo,doing,done,blocked}/`

---

## ⚡ Quick Lookup

**"How do I reference an instruction file?"**
→ Use `.plaesy/instructions/[name].md` (flat, no subfolders)

**"Where are agents stored?"**
→ `.plaesy/roles/[name].md` (platform-agnostic, like instructions in .plaesy/instructions/)

---

**Status:** Memory index created
**Next:** Add project-specific entries as needed
