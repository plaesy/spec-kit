---
description: "Generate infrastructure-as-code templates (Terraform .tf or CloudFormation YAML) with 2026 best practices — produces real, deployable IaC files, not descriptions"
---

# `/create:template:infra` command instructions

This is the **template:infra**-scoped entry point into `/create`. It produces real, deployable infrastructure-as-code files (Terraform .tf or CloudFormation YAML), not descriptions or templates.

## Usage Format

```bash
/create:template:infra "AWS VPC with public/private subnets, NAT gateway, and security groups"
/create:template:infra "S3 bucket with encryption, versioning, and lifecycle policies" --tool terraform --out infrastructure/s3-module
/create:template:infra --for .plaesy/memory/infra-spine.md --component "vpc-with-rds" --environment prod
/create:template:infra "Multi-environment Kubernetes cluster on EKS" --tool cloudformation --region us-east-1 --params environment=staging,instance_type=t3.medium
```

| Flag | Default | Meaning |
|---|---|---|
| `--tool` | inferred from project config or `terraform` | IaC tool: `terraform`, `cloudformation`, `opentofu`, `pulumi`, `cdk` |
| `--out` | `infrastructure/<slugified-description>` | Output directory for generated files (relative to project root) |
| `--for` | — | Path to an infra-spine/spec file to pull component context from instead of freeform description |
| `--component` | — | With `--for`, which infrastructure component/module to generate (e.g., `vpc-with-rds`, `eks-cluster`) |
| `--environment` | `dev` | Target deployment environment: `dev`, `staging`, `prod` — affects variable defaults, tagging, security policies |
| `--cloud` | `aws` | Cloud provider: `aws`, `azure`, `gcp` — shapes provider config and service availability |
| `--region` | inferred from project or `us-east-1` | Cloud region (AWS) or location (Azure/GCP) for resource deployment |
| `--params` | — | Comma-separated `key=value` pairs for environment-specific parameters (e.g., `instance_type=t3.large,enable_monitoring=true`) |
| `--with-state-backend` | `true` | Generate remote state backend config (S3+DynamoDB for Terraform, or cloud-native equivalents) |
| `--with-validation` | `true` | Include pre-deployment validation scripts (Checkov, terraform validate, cfn-lint) |
| `--with-tests` | `true` | Include integration test files (terratest, cfn-lint policies, Pulumi tests) |
| `--modularity` | `auto` | Module strategy: `auto` (infer from scope), `flat` (single file), `modular` (composable modules) |

## Protocol

### Step 1: Resolve the Infrastructure Brief

- If `--for`/`--component` given: read that file, extract the component's technical description, resource dependencies, and compliance requirements. If missing entirely, ask the user for a one-line infrastructure brief rather than guessing.
- Otherwise the CLI/programmatic argument *is* the brief.
- Layer in project infrastructure standards when they exist:
  - `.plaesy/memory/infra-spine.md` — approved architectures, naming conventions, security policies, module catalog
  - `.plaesy/memory/cloud-config.md` (or project env vars) — cloud provider, regions, tagging standards
  - `.plaesy/roles/security.md` / `.plaesy/roles/devops.md` — compliance requirements, audit trails, encryption policies
- **Infer modularity**: Single component → `flat` file. Multi-tier architecture → `modular` with nested structure.
- **Infer cloud provider** from `--cloud` flag or project config; default to AWS.
- Compose the final generation context:
  ```
  {brief} — {tool} format, {environment}-grade infrastructure, {cloud} services, 
  remote state backend, security scanning, comprehensive variable parameterization, 
  outputs for cross-stack references, inline documentation
  ```

### Step 2: Resolve the Tool & Provider

Read the tool/provider config, in this order, stop at first match:
1. `--tool` CLI flag
2. Env var `PLAESY_IAC_TOOL` (`terraform` | `cloudformation` | `opentofu` | `pulumi` | `cdk`)
3. `.plaesy/scripts/configs/iac-provider.json` → `{"tool": "terraform", "cloud": "aws"}`
4. Default: `terraform` + `aws`

**Validate tool availability**:
- For Terraform/OpenTofu: confirm `terraform` or `tofu` CLI is in `$PATH`
- For CloudFormation: confirm AWS CLI is installed
- For Pulumi: confirm `pulumi` CLI is available
- For AWS CDK: confirm `cdk` CLI and Node.js runtime

**If tool not available**: stop, tell the user exactly which tool to install (with download link), and offer the infrastructure brief from Step 1 as a manual fallback.

### Step 3: Generate IaC Template

Run the `plaesy generate-template` command (cross-platform, single entry point):

```bash
plaesy generate-template \
  --prompt "<composed context>" \
  --tool terraform \
  --cloud aws \
  --environment prod \
  --out infrastructure/vpc-module \
  --with-state-backend \
  --with-validation \
  --with-tests
```

**Expected output structure** (for `--tool terraform`, `--modularity auto`):
```
infrastructure/vpc-module/
├── main.tf                     # Core resource definitions
├── variables.tf                # Input variables with descriptions, defaults, constraints
├── outputs.tf                  # Output blocks for cross-module references
├── terraform.tfvars            # Environment-specific variable values
├── backends.tf                 # Remote state config (S3+DynamoDB for prod)
├── locals.tf                   # Computed values, common tags
├── data.tf                     # Data sources (existing infrastructure references)
├── validation.tf               # Variable validation rules (HCL 1.2+)
└── modules/
    ├── vpc/
    │   ├── main.tf
    │   ├── variables.tf
    │   └── outputs.tf
    └── security-groups/
        ├── main.tf
        ├── variables.tf
        └── outputs.tf
```

**Expected output structure** (for `--tool cloudformation`):
```
infrastructure/vpc-stack/
├── template.yaml               # Main CloudFormation template
├── parameters.json             # Parameter default values
├── nested-vpc.yaml             # Nested stack for VPC (optional)
├── validation-policy.json      # CloudFormation policy for compliance
└── tests/
    └── template-test.yaml      # cfn-lint policies
```

### Step 4: Validate & Test

- **Syntax validation**: Run `terraform validate` or `aws cloudformation validate-template`
- **Security scanning**: Run Checkov against the template (`checkov -f infrastructure/...`)
- **Policy enforcement**: Validate against organizational standards (e.g., "all S3 buckets must have encryption enabled")
- **Test execution**: If `--with-tests` given, run integration tests (terratest, Pulumi automation tests)
- Report any validation errors verbatim (don't paraphrase); stop if critical failures present

### Step 5: Index & Document

Write a `.plaesy/memory/infra-templates/{component-name}.md` index file containing:
- Brief description (copied from Step 1)
- Architecture diagram reference (if diagram exists in `.plaesy/memory/diagrams/`)
- Input variables with descriptions and constraints
- Output exports (for cross-module/cross-stack usage)
- Deployment prerequisites (AWS account setup, IAM roles, API keys)
- Usage examples (how to invoke this template via Terraform Cloud, CloudFormation console, or CLI)
- Security & compliance notes
- Links to related templates/modules

Update `.plaesy/memory/infra-templates.md` to index all generated templates.

## Programmatic Invocation (Called By Other Prompts)

A prompt that needs infrastructure mid-run (`/implement:infra`, `/improve:product`, `/doc`) calls this protocol directly:

```
CALL /create:template:infra
  brief: "<one-line infrastructure description>"
  tool: "<terraform | cloudformation | opentofu | pulumi | cdk>"
  cloud: "<aws | azure | gcp>"
  environment: "<dev | staging | prod>"
  out: "<path the calling prompt will reference>"
  with_state_backend: <true | false>
  with_validation: <true | false>
RETURNS
  directory: <written directory path, or null if generation was skipped>
  files: {main_file, variables_file, outputs_file, …}
  index_file: <path to .plaesy/memory/infra-templates/{component-name}.md>
  template_context: <final composed prompt, for reproducibility>
  validation_report: {passed: boolean, errors: […], warnings: […]}
```

If tool is not available, the call returns `directory: null` and `template_context` still populated — the calling prompt must treat this as "template pending, manual generation needed" and say so in its own output, never silently drop the template or fabricate paths that don't exist.

## Anti-Patterns (NEVER Do These)

- ❌ Fabricate or hallucinate file paths without `plaesy generate-template` actually having written files there
- ❌ Omit remote state backend configuration for shared/production deployments — single-machine state is not production-ready
- ❌ Skip input variable parameterization — hardcode values instead of using `variables.tf` with descriptions and defaults
- ❌ Omit outputs blocks — other modules and downstream automation need to reference created resource IDs and endpoints
- ❌ Generate a single flat template for a multi-tier architecture — use modular structure to reduce blast radius and enable reuse
- ❌ Skip security scanning (Checkov, policy-as-code) — unvalidated templates risk compliance violations and misconfigurations
- ❌ Ignore naming conventions and tagging standards — infrastructure that doesn't follow organizational standards creates operational chaos
- ❌ Generate CloudFormation templates without Export fields for cross-stack references (or use Fn::GetStackOutput for June 2026+ features)
- ❌ Omit inline documentation (descriptions on variables, comments on non-obvious logic) — future maintainers and automation tools need clarity
- ❌ Generate templates that don't integrate with CI/CD pipelines (no `terraform cloud` config, no CloudFormation stack set definitions)
- ❌ Retry provider/validation calls in a loop on failure — surface the error once, stop, and let the user fix the root cause
- ❌ Skip data source declarations for existing infrastructure — always enable consumption of already-deployed resources without re-provisioning

## Integration Points

### Version Control Ready
- Output structure immediately supports git-based workflows with clear module versioning (SemVer tags per `.plaesy/memory/infra-spine.md`)
- `.gitignore` includes state files, lock files, and local secrets
- Commit message templates suggest infrastructure changes with ticket/RFC references

### Deployment Integration
Generated templates are immediately executable:
- **Terraform**: `terraform plan` → `terraform apply` workflow; Terraform Cloud integration for state locking and run queue
- **CloudFormation**: `aws cloudformation deploy --template-file template.yaml --parameter-overrides` ready to run
- **Output consumption**: Explicit output blocks enable downstream automation (CI/CD to retrieve endpoints, IDs, DNS names)

### Compliance & Audit
- Security scanning hooks (Checkov) run pre-deployment
- Policy-as-code prevents non-compliant resources (no unencrypted storage, no overly permissive IAM)
- Audit trails enabled (CloudTrail, Terraform logs) for traceability
- Tagging standards applied automatically (environment, owner, cost-center tags)

### Testing & Validation
- Integration tests (terratest, cfn-lint policies) run on generated templates
- Drift detection enabled (CloudFormation stack drift monitoring)
- State locking prevents concurrent applies

## Success Criteria

A generated infrastructure template is complete when:
- ✅ Core resource definitions present (main.tf or template.yaml)
- ✅ Input variables fully parameterized with descriptions and type constraints
- ✅ Output blocks defined for cross-module/cross-stack references
- ✅ Remote state backend configured (prod-grade deployments)
- ✅ Security scanning passed (Checkov, policy validation)
- ✅ Syntax validation passed (`terraform validate`, `cfn validate`)
- ✅ Naming conventions and tagging standards applied
- ✅ Inline documentation present (variable descriptions, comment blocks)
- ✅ Integration tests present (if `--with-tests` given)
- ✅ Index file written to `.plaesy/memory/infra-templates/{component-name}.md`
- ✅ `.plaesy/memory/infra-templates.md` updated to index the new template
- ✅ Output directory reported with file list

---

**Follow shared protocols**: `.plaesy/instructions/quality-gates.md` → `.plaesy/instructions/error-recovery.md`

**Research sources**:
- [HashiCorp Terraform Best Practices](https://developer.hashicorp.com/terraform/language/values/outputs) (Remote state, modularity, outputs)
- [AWS CloudFormation Best Practices](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/best-practices.html) (Parameters, outputs, nested stacks)
- [Terraform vs. Pulumi vs. OpenTofu: 2026 Comparison](https://www.frugaltesting.com/blog/terraform-vs-pulumi-vs-opentofu-best-iac-tools-for-cloud-automation-in-2026)
- [Infrastructure as Code Best Practices (2026)](https://cloudforge.cloud/blog/infrastructure-as-code-best-practices) (Modularity, testing, policy-as-code, state management)
- [Checkov: Infrastructure-as-Code Scanning](https://www.checkov.io/) (Security and compliance validation)
