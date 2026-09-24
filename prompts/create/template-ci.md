---
description: "Generate production-ready CI/CD pipeline configuration files (GitHub Actions YAML, GitLab CI) with 2026 security, performance, and compliance best practices"
---

# `/create:template:ci` command instructions

This is the **template:ci**-scoped entry point into `/create`. It produces real saved pipeline configuration files (`.github/workflows/ci-cd.yml` for GitHub, `.gitlab-ci.yml` for GitLab), not descriptions.

## Usage Format

```bash
/create:template:ci --platform github --variant standard
/create:template:ci --platform gitlab --variant comprehensive --compliance gdpr,soc2
/create:template:ci --platform github --variant minimal --out .github/workflows/ci-cd.yml
/create:template:ci --for .plaesy/memory/infrastructure.md --platform auto
```

| Flag | Default | Meaning |
|---|---|---|
| `--platform` | `auto` (detect from repo) | Target CI/CD platform: `github` (GitHub Actions), `gitlab` (GitLab CI), or `auto` to detect from repo presence |
| `--variant` | `standard` | Template scope: `minimal` (fast linting + unit tests), `standard` (+ staging deploy), `comprehensive` (+ security scanning, compliance, monitoring) |
| `--compliance` | inferred from `.plaesy/instructions/` | Compliance frameworks: comma-separated list from `gdpr`, `soc2`, `iso27001`, `pci-dss`, `owasp` |
| `--security-level` | `high` | Security scanning depth: `basic` (linting only), `high` (SAST + dependency scan), `critical` (+ DAST, container scan, SBOM) |
| `--for` | — | Path to infrastructure/deployment spec file to pull environment/service context from |
| `--out` | `.github/workflows/ci-cd.yml` (GitHub) or `.gitlab-ci.yml` (GitLab) | Where the pipeline config is written, relative to project root |
| `--concurrency` | `true` | Enable concurrency groups and caching to optimize runtime |
| `--stages` | inferred from variant | Explicit pipeline stages to include (default: Source, Build, Test, Security, Compliance, Staging, Production, Monitoring) |

## Protocol

### Step 1: Resolve Pipeline Requirements

- If `--for` given: read that file, extract deployment targets, environments (staging/production), and compliance requirements. If missing, ask the user to clarify deployment strategy.
- If `--compliance` given: validate each framework name against known set (gdpr, soc2, iso27001, pci-dss, owasp). Flag unknown values and continue with known ones.
- Otherwise use `--platform` and `--variant` to infer requirements.
- **Infer variant features** (if not explicitly specified):
  - `minimal`: linting, unit tests, Docker build (if Dockerfile exists), deploy to single target
  - `standard`: + integration tests, staging environment, health checks
  - `comprehensive`: + SAST (CodeQL/Snyk), DAST, container scanning, SBOM generation, compliance validation, monitoring setup, blue-green deployments
- Layer in project deployment context when it exists:
  - `.plaesy/instructions/quality-gates.md` — test coverage thresholds, linting rules, test categories (unit/integration/E2E)
  - `.plaesy/memory/infrastructure.md` — service names, environment variables, deployment strategy (blue-green, canary, rolling)
  - `.plaesy/roles/security.md` — compliance requirements, secret management patterns, audit logging
- Check `.plaesy/instructions/error-recovery.md` for incident/rollback procedures to embed in post-deployment stages.

### Step 2: Resolve the Platform

Read the platform config, in this order, stop at first match:
1. `--platform` flag (github | gitlab | auto)
2. `.plaesy/scripts/configs/ci-platform.json` → `{"platform": "github"}`
3. Detect from repo: presence of `.github/workflows/`, `.gitlab-ci.yml`, or git remote URL domain
4. Default: `github`

If `auto` and detection fails, stop and ask the user to specify `--platform github` or `--platform gitlab` explicitly.

### Step 3: Generate Pipeline Configuration

Run the `plaesy generate-pipeline` command to compose and write the configuration file:

```bash
plaesy generate-pipeline \
  --platform github \
  --variant standard \
  --compliance gdpr,soc2 \
  --security-level high \
  --concurrency true \
  --out .github/workflows/ci-cd.yml
```

This is a single cross-platform command — no separate bash/PowerShell variant. The command:
- Generates complete YAML with all required stages, jobs, environment variables, secrets blocks
- Embeds GitHub-specific markers for reusable workflows, OIDC, matrix strategies, conditional execution
- References security tools matching the security-level (CodeQL, Snyk, Trixy, etc.)
- Embeds compliance checks (GDPR data handling, audit logging, retention policies)
- Includes post-deployment monitoring setup (health checks, alerts, observability)

### Step 4: Generate Supporting Files

Write companion files to the same directory as the pipeline config:

**secrets.list**: A YAML file documenting all secrets required by the pipeline:
```yaml
secrets:
  - name: OPENAI_API_KEY
    description: "API key for code analysis tools"
    scope: "build"
  - name: SENTRY_DSN
    description: "Error tracking endpoint"
    scope: "production-deployment"
  - name: DB_CONNECTION_STRING
    description: "Production database URL (read-only replica for tests)"
    scope: "testing"
```

**setup-instructions.md**: A README with:
- Prerequisites (GitHub Actions org settings, GitLab runner config, required permissions)
- Step-by-step secret configuration (per-environment)
- Deployment procedures (triggering pipelines, approval gates)
- Troubleshooting guide (common failure modes, recovery steps)
- Monitoring and alerting setup (where to find logs, how to configure notifications)
- Compliance validation checklist

### Step 5: Validate & Report

- Confirm the pipeline config file exists and is valid YAML (parse it; surface syntax errors verbatim, don't paraphrase)
- If supporting files generated, confirm they exist and are readable
- Report:
  - Pipeline config file path written (e.g., `.github/workflows/ci-cd.yml`)
  - Platform and variant used (e.g., "GitHub Actions, standard variant")
  - Stages included (e.g., "8 stages: Source → Build → Test → Security → Compliance → Staging → Production → Monitoring")
  - Compliance frameworks enabled (if any)
  - Supporting files written (secrets.list, setup-instructions.md)
  - Next steps: "To activate: 1. Export secrets per secrets.list. 2. Commit .github/workflows/ci-cd.yml. 3. Push to main branch. 4. Verify first run in GitHub Actions tab."

## Programmatic Invocation (Called By Other Prompts)

A prompt that needs a CI/CD pipeline template mid-run (`/implement`, `/improve:product`, `/doc`) calls this protocol directly:

```
CALL /create:template:ci
  platform: "github" | "gitlab" | "auto"
  variant: "minimal" | "standard" | "comprehensive"
  compliance: ["gdpr", "soc2"] (optional; omit if none needed)
  security_level: "basic" | "high" | "critical" (optional; defaults to "high")
  out: "<path for pipeline config>"
RETURNS
  config_file: <written file path, or null if generation was skipped — see below>
  platform_used: <"github" or "gitlab">
  stages: [<stage list>]
  secrets_file: <path to secrets.list>
  instructions_file: <path to setup-instructions.md>
  config_content: <YAML content for reproducibility>
```

If provider/CLI not available, the call returns `config_file: null` and `config_content` still populated — the calling prompt must treat this as "pipeline pending, manual setup needed" and say so in its own output, never silently drop the config or fabricate paths that don't exist.

## Anti-Patterns (NEVER Do These)

- ❌ Return only a YAML snippet or description of what the pipeline should do; the point of this command is a working config file ready to commit
- ❌ Hardcode API keys, database passwords, or environment-specific values directly in the YAML — use GitHub Secrets / GitLab CI Variables instead
- ❌ Generate a generic template that ignores `.plaesy/instructions/quality-gates.md`, `.plaesy/memory/infrastructure.md`, or compliance requirements — templates must be project-specific
- ❌ Omit security scanning (SAST, dependency scan, secrets detection) from any variant except `minimal` — security is non-negotiable in 2026
- ❌ Use branch names or tags instead of full commit SHAs for action/workflow includes — this breaks reproducibility and enables supply-chain attacks
- ❌ Include manual approval gates without documenting approval procedures in setup-instructions.md
- ❌ Generate separate pipelines for staging vs. production without shared base jobs/reusable workflows — violates DRY
- ❌ Skip health checks, smoke tests, and automated rollback on deployment failure
- ❌ Omit SBOM (CycloneDX/SPDX format) generation or artifact scanning when `comprehensive` variant requested
- ❌ Embed personally identifiable data (user names, email addresses) in logs or pipeline output
- ❌ Use environment variables for secrets instead of platform-native Secret/Variable types (GitHub Secrets, GitLab CI Variables)
- ❌ Generate a pipeline without documenting which tools/credentials are required (always produce secrets.list)
- ❌ Retry failed deployments in a loop without manual intervention — surface errors once and require explicit re-run

## Design-Spine Integration

If the project has a `.plaesy/memory/infrastructure.md` with defined environments/services:

1. **Before generating**, check for deployment targets (staging, production, etc.)
2. If found, extract environment variables, service names, database configurations
3. Embed these into job-level env blocks and deployment targets
4. If deployment strategy defined (blue-green, canary), use it; otherwise default to rolling deployment
5. If a service **missing** from infrastructure spec but user requests it, note in setup-instructions.md as a configuration gap

## Success Criteria

A CI/CD pipeline template is complete when:
- ✅ Pipeline config file exists and is valid YAML (passes `yamllint`)
- ✅ Matches requested platform (GitHub Actions or GitLab CI syntax)
- ✅ Includes all 8+ requested stages (or fewer if `--stages` specified)
- ✅ Security scanning integrated (SAST, dependency scan, container scan per variant)
- ✅ Compliance checks present (audit logging, data handling per framework)
- ✅ Secrets documented in secrets.list (not embedded in YAML)
- ✅ Environment-specific jobs separated (staging, production with distinct approval gates)
- ✅ Health checks and smoke tests present
- ✅ Post-deployment monitoring/alerting configured
- ✅ Reusable workflows/jobs defined (DRY principle observed)
- ✅ setup-instructions.md includes prerequisites, secret setup, deployment procedures, troubleshooting
- ✅ Ready to commit and push without modification

## Sources & References

- GitHub Actions Best Practices 2026: https://dev.to/asifthewebguy/cicd-pipeline-best-practices-a-production-ready-guide-for-2026-5fon
- Building Production-Ready CI/CD 2026: https://medium.com/@krishnafattepurkar/building-a-production-ready-ci-cd-pipeline-the-complete-2026-guide-b3d6a661ecd8
- GitLab CI Best Practices 2026: https://oneuptime.com/blog/post/2026-01-27-gitlab-ci-performance/
- Continuous Integration Best Practices: https://about.gitlab.com/topics/ci-cd/continuous-integration-best-practices/
- DevOps Trends 2026: https://www.refontelearning.com/blog/devops-engineering-in-2026-top-ci-cd-tools-trends-and-best-practices-github-actions-vs-jenkins
- CI/CD Best Practices Guide: https://www.harness.io/blog/ci-cd-best-practices

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`
