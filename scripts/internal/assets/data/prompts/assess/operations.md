---
description: "Assess deployment, release, observability, runbooks, IaC"
subagent: true
---

# `/assess:operations` command instructions

Scope fixed to **operations**: Deployment, release, reliability/SLOs, observability, infrastructure as code, runbooks.

- **Loads**: `.plaesy/instructions/assess-operations.md` — apply its **Persona
  Protocol** first, then follow the full protocol in `/assess`.
- **Delegates to**: `.plaesy/roles/devops.md`, `.plaesy/roles/sre.md`

The technical/operations boundary is in the Boundary table of the loaded file —
use it, do not re-derive it here.
