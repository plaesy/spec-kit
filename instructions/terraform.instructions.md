---
description: 'Terraform Conventions and Guidelines'
applyTo: '**/*.tf'
---

# Terraform Conventions

## General Instructions
- Use Terraform to provision/manage infrastructure, under version control

## Security
- Latest stable Terraform and provider versions; update regularly for security patches
- Store secrets securely (AWS Secrets Manager, SSM Parameter Store); rotate credentials regularly, automate rotation where possible
- Reference secrets via env vars pointing at Secrets Manager/SSM, keeping sensitive values out of state files
- Never commit credentials/API keys/passwords/certificates/state to version control - use `.gitignore`
- Mark sensitive variables `sensitive = true` to prevent display in plan/apply output
- IAM roles/policies with least privilege
- Security groups and network ACLs to control network access
- Private subnets by default; public subnets only for resources needing direct internet access (load balancers, NAT gateways)
- Encrypt data at rest and in transit (EBS, S3, RDS encryption; TLS between services)
- Regularly scan configs for vulnerabilities (`trivy`, `tfsec`, `checkov`)

## Modularity
- Separate projects per major infrastructure component - less complexity, faster plan/apply, independent dev/deploy, less risk of unrelated-resource changes
- Modules to avoid duplication: encapsulate related resources, simplify complex configs, avoid circular dependencies
  - Only use modules where they add value - not for single resources, avoid excessive nesting
- `output` blocks for info useful to other modules/users; mark `sensitive = true` if the output contains sensitive data

## Maintainability
- Prioritize readability, clarity, maintainability
- Comment complex configurations and non-obvious design decisions
- Concise, efficient, idiomatic configs
- No hard-coded values - use variables, with sensible defaults where appropriate
- Data sources for existing-resource info instead of manual config - reduces errors, stays current, adapts across environments
  - Don't use data sources for resources created in the same config (use outputs instead); remove unnecessary ones (they slow plan/apply)
- `locals` for values used multiple times, for consistency

## Style and Formatting
- Descriptive, consistently-named resources/variables/outputs
- Follow the Terraform Style Guide (2-space indentation)
- Group related resources in the same file with consistent naming (`providers.tf`, `variables.tf`, `network.tf`, `ecs.tf`, `mariadb.tf`)
- `depends_on` at the start of resource definitions (dependency relationships clear); use only when necessary to avoid circular deps
- `for_each`/`count` at the start of resource definitions (after `depends_on` if present) - `for_each` for collections, `count` for numeric iteration
- `lifecycle` blocks at the end of resource definitions
- Alphabetize providers/variables/data sources/resources/outputs per file, for navigation
- Group related attributes within blocks: required before optional, commented per section, blank-line separated, alphabetized within each section
- Blank lines between logical config sections
- `terraform fmt` for automatic formatting, `terraform validate` for syntax/validity, `tflint` for style violations - run `tflint` regularly

## Documentation
- `description` and `type` on all variables/outputs, with clear concise descriptions and appropriate types (`string`, `number`, `bool`, `list`, `map`)
- Comment configurations where it adds value - purpose of resources/variables, complex decisions; avoid redundant comments
- `README.md` per project: overview, structure, setup/usage instructions
- `terraform-docs` for automatic documentation generation

## Testing
- Tests for config functionality, `.tftest.hcl` extension
- Cover both positive and negative scenarios
- Idempotent tests, safe to run repeatedly without side effects
