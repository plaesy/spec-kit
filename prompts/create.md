---
description: "Generate real assets (images, diagrams, tasks, storyboards, boilerplate templates) from specifications — router for /create:{scope}"
---
# `/create` command instructions

⚡ **Run with**: standard (single-agent; no multi-agent fan-out needed for one asset call)

## Usage Format

```bash
# Asset generation
/create:images "a flat-style empty-state illustration for an empty inbox"
/create:storyboard "User signup flow: landing → email → verify → dashboard"
/create:diagram "Architecture: frontend → API gateway → microservices → database"
/create:tasks "requirements.md" --format markdown

# Boilerplate templates
/create:template:api "REST API for e-commerce product catalog with search"
/create:template:infra "AWS infrastructure: VPC, RDS PostgreSQL, S3, CloudFront"
/create:template:ci "GitHub Actions CI/CD pipeline with security scanning"
/create:template:project "Monorepo with React frontend + Node backend"
```

`/create` is a **router**, same shape as `/assess`/`/improve`/`/fix`: it has no
scope-less behavior of its own. Today it has eight scopes:

### Asset Generation (Real Binary/File Output)
| Scope                | Purpose                                                                                                              |
| -------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `/create:images` | Turn a text description (or a design-spine/brandkit entry) into an actual saved image file, not just a prompt string |
| `/create:storyboard` | Generate a visual storyboard (sequential narrative panels) from a user journey/interaction flow — produces real image assets per panel, not descriptions |
| `/create:diagram` | Generate visual diagrams (architecture, flowchart, ERD, sequence, swimlane, mindmap) from natural language or code context — produces SVG and Mermaid markdown files |
| `/create:tasks` | Generate hierarchical task backlog (epics → user stories → acceptance criteria → technical tasks) from requirements specifications — produces structured Markdown + JSON backlog files |

### Boilerplate Templates (Code Scaffolding)
| Scope                | Purpose                                                                                                              |
| -------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `/create:template:api` | Generate API boilerplate (OpenAPI 3.2.0 YAML, GraphQL SDL) with modular components, pagination patterns, authentication schemas |
| `/create:template:infra` | Generate infrastructure-as-code (Terraform .tf, CloudFormation YAML) with modular structure, remote state config, policy enforcement |
| `/create:template:ci` | Generate CI/CD pipeline configs (GitHub Actions .yml, GitLab CI .yml) with security scanning, staged deployments, compliance gates |
| `/create:template:project` | Generate project scaffolding (monorepo structure, package.json, tsconfig.json, ESLint/Prettier config) ready for install |

More scopes (e.g. `/create:audio`, `/create:video`) are added the same way if
the project ever needs them — this file stays a thin router; each scope has its
own command (e.g., `/create:images`, `/create:template:api`).

## Why This Exists

Every other Plaesy prompt (`/implement:design`, `/improve:design`, `/doc`) can
describe what an asset *should* look like, but none of them actually call an
image-generation API and write bytes to disk — they stop at a text description
and leave the human to go generate it by hand. `/create:images` closes that
gap: given a description (typed by a user, or handed to it programmatically by
another prompt), it produces a real image file in the project and reports its
path, so the calling context can reference it immediately (in an `<img>`, a
Figma upload, a doc).

## Callable By Other Prompts

Any prompt that needs a concrete asset mid-run invokes the appropriate `/create:{scope}` directly with a structured call instead of re-describing generation itself:

- **Images**: `/implement:design` building a component that needs an icon/illustration, `/improve:design` replacing an outdated asset, `/doc` illustrating a concept
- **Diagrams**: `/implement:technical` documenting system architecture, `/assess:technical` analyzing existing systems, `/doc` explaining processes/data models
- **Tasks**: `/implement` decomposing requirements into sprint-ready stories, `/doc` generating task documentation from specifications

See "Programmatic Invocation" sections in the respective scope documentation.

## Anti-Patterns (NEVER Do These)

- ❌ Return only a text prompt and call it "done" — the point of this command is
  a saved file; if generation truly cannot run (no provider configured), say so
  explicitly and hand back the prompt as a documented fallback, don't blur the two
- ❌ Silently pick a different provider than the one configured — fail loudly and
  tell the user what to configure
- ❌ Generate an asset that ignores the project's `brandkit.instructions.md` /
  `.plaesy/memory/design-spine.md` when either exists — this is asset generation
  *for the project*, not generic stock art
- ❌ Overwrite an existing asset file without confirmation — write to a new
  path/version unless the user explicitly asked to replace one

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
