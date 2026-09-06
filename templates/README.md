# Templates

Reusable document skeletons for the Spec-Kit workflow. `template-registry.json` is the machine-readable registry (used by tooling to resolve `{template}` references); this file is the human index.

## Core workflow

| Template | Purpose |
|----------|---------|
| [spec.template.md](spec.template.md) | Feature specification |
| [plan.template.md](plan.template.md) | Implementation plan |
| [tasks.template.md](tasks.template.md) | Task breakdown |
| [spec-structure-folder.template.md](spec-structure-folder.template.md) | Recommended `specs/<feature>/` folder layout |
| [task-structure-guide.template.md](task-structure-guide.template.md) | Epic/story/checklist folder structure and templates |
| [task-readme.template.md](task-readme.template.md) / [task-status.template.json](task-status.template.json) | Per-task README and status file |
| [status.template.md](status.template.md) | Project status dashboard |
| [agent-file-template.md](agent-file-template.md) | AI agent context file |

## Decisions & governance

| Template | Purpose |
|----------|---------|
| [adr.template.md](adr.template.md) | Architecture Decision Record |
| [rtm.template.md](rtm.template.md) | Requirements traceability matrix |
| [change-request.template.md](change-request.template.md) | Change request |
| [project-charter.template.md](project-charter.template.md) | Project charter |
| [raci-matrix.template.md](raci-matrix.template.md) | RACI matrix |
| [stakeholder-register.template.md](stakeholder-register.template.md) | Stakeholder register |
| [risk-register.template.md](risk-register.template.md) | Risk register |

## Requirements & design docs

| Template | Purpose |
|----------|---------|
| [brd.template.md](brd.template.md) | Business requirements |
| [prd.template.md](prd.template.md) | Product requirements |
| [srs.template.md](srs.template.md) | Software requirements spec |
| [fsd.template.md](fsd.template.md) | Functional spec |
| [icd.template.md](icd.template.md) | Interface control document |
| [dfd-and-context.template.md](dfd-and-context.template.md) | Data flow / context diagram |
| [brainstorming.template.md](brainstorming.template.md) | Brainstorming session output |
| [research.template.md](research.template.md) | Research doc |
| [application.template.md](application.template.md) / [service.template.md](service.template.md) / [mobile-app.template.md](mobile-app.template.md) / [tool.template.md](tool.template.md) | Component-type-specific specs |

## Security, privacy & compliance

| Template | Purpose |
|----------|---------|
| [threat-model.template.md](threat-model.template.md) | STRIDE/LINDDUN threat model |
| [security-assessment.template.md](security-assessment.template.md) | Security assessment |
| [ci-cd-security.template.md](ci-cd-security.template.md) | Pipeline security review |
| [dpia.template.md](dpia.template.md) | Data protection impact assessment |
| [data-governance.template.md](data-governance.template.md) | Data governance plan |
| [sbom-template.json](sbom-template.json) | Software bill of materials |

## Testing, quality & operations

| Template | Purpose |
|----------|---------|
| [test-plan.template.md](test-plan.template.md) / [test-strategy.template.md](test-strategy.template.md) | Test planning |
| [performance-testing-strategy.template.md](performance-testing-strategy.template.md) | Performance test strategy |
| [vv-plan.template.md](vv-plan.template.md) | Verification & validation plan |
| [iso25010-quality-model.template.md](iso25010-quality-model.template.md) | ISO/IEC 25010 quality model |
| [sqap.template.md](sqap.template.md) | Software quality assurance plan |
| [runbook.template.md](runbook.template.md) / [monitoring-playbook.template.md](monitoring-playbook.template.md) | Operations runbooks |
| [orr-checklist.template.md](orr-checklist.template.md) | Operational readiness review |
| [sla.template.md](sla.template.md) | Service level agreement |
| [deployment-guide.template.md](deployment-guide.template.md) / [migration-plan.template.md](migration-plan.template.md) | Deployment & migration |
| [release-plan.template.md](release-plan.template.md) | Release plan |
| [scmp.template.md](scmp.template.md) | Software configuration management plan |
| [spmp.template.md](spmp.template.md) | Software project management plan |

## Communication & documentation

| Template | Purpose |
|----------|---------|
| [communication-plan.template.md](communication-plan.template.md) | Communication plan |
| [api-documentation.template.md](api-documentation.template.md) | API docs |
| [integration-examples.template.md](integration-examples.template.md) | Integration examples |
| [user-documentation.template.md](user-documentation.template.md) / [user-training-guide.template.md](user-training-guide.template.md) | End-user docs |

## Platform-specific assets

- [ai-headers/](ai-headers/) — AI-provider-specific header injection for prompts/chatmodes/instructions (see [ai-headers/README.md](ai-headers/README.md))
- [cloud/](cloud/) — Terraform/Kubernetes/ARM starter configs (see [cloud/README.md](cloud/README.md))
- [ai-automation-workflow.yaml](ai-automation-workflow.yaml) — CI automation workflow template

## Conventions

- Placeholders use `{curly_braces}` or `[SQUARE_BRACKETS]` — pick one style per file and stay consistent within it, don't mix.
- Keep templates skeleton-first: headings, tables, and checklist items over prose. Prose belongs in the filled-out document, not the template.
- Adding a template: give it a unique `*.template.md`/`.template.json` name (no ad-hoc `-template.md` variants), register it in `template-registry.json`, and add a row above.
