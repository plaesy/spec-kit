# Tool Development Specification: {{TOOL_NAME}}

**Created**: {{DATE}}
**Status**: {{DOC_STATUS|draft}}
**Type**: Command-line tool or library
**Source request**: {{SOURCE_REQUEST|verbatim text that prompted this specification}}
**Spec owner**: {{SPEC_OWNER}}
**Constitution dimensions in force**: {{ACTIVE_DIMENSIONS|Technical}}
**Related specs**: {{RELATED_SPEC_IDS|spec directory numbers this specification depends on}}

## Working constraints

Each constraint below binds the delivered tool, not this document. Every one is
either satisfied by a named section further down or it is not.

- The tool does one thing, and that one thing is named in a single sentence.
- The same capability is reachable from a library call and from a command, so
  scripting does not require reimplementing anything.
- Standard streams carry the contract: results on stdout, diagnostics on stderr.
- Every failure has a distinct exit code and a message that names the cause.
- Installation is one command on every supported platform, with no build step
  required for a user who only wants to run the tool.
- Behaviour is identical across supported platforms, including path, encoding,
  and line-ending handling.
- Documentation ships with the tool: an un-documented flag is a defect, not a gap.

## Tool Overview

### Core purpose

{{CORE_PURPOSE|one paragraph: what this tool does and the problem it removes}}

### Users and environments

| Aspect | Definition | Consequence if wrong |
|--------|------------|---------------------|
| Primary users | {{PRIMARY_USERS|who runs this regularly}} | {{PRIMARY_USERS_RISK}} |
| Secondary users | {{SECONDARY_USERS|who runs this occasionally}} | {{SECONDARY_USERS_RISK}} |
| Execution environment | {{EXECUTION_ENVIRONMENTS\|developer workstation, CI runner, or server}} | {{ENVIRONMENT_RISK}} |
| Invocation style | {{INVOCATION_STYLE\|interactive, scripted, or both}} | {{INVOCATION_RISK}} |
| Alternatives today | {{CURRENT_ALTERNATIVES\|what people use before this exists}} | {{ALTERNATIVES_RISK}} |

### Benefits

| # | Benefit | How it is measured | Baseline today |
|---|---------|--------------------|----------------|
| 1 | {{BENEFIT_1\|time or effort removed}} | {{BENEFIT_1_METRIC}} | {{BENEFIT_1_BASELINE}} |
| 2 | {{BENEFIT_2}} | {{BENEFIT_2_METRIC}} | {{BENEFIT_2_BASELINE}} |
| 3 | {{BENEFIT_3}} | {{BENEFIT_3_METRIC}} | {{BENEFIT_3_BASELINE}} |

## CLI Contract

### Commands

| Command | Purpose | Reads | Writes | Exit on success |
|---------|---------|-------|--------|-----------------|
| {{COMMAND_1}} | {{COMMAND_1_PURPOSE}} | {{COMMAND_1_INPUTS}} | {{COMMAND_1_OUTPUTS}} | {{COMMAND_1_EXIT_OK\|0}} |
| {{COMMAND_2}} | {{COMMAND_2_PURPOSE}} | {{COMMAND_2_INPUTS}} | {{COMMAND_2_OUTPUTS}} | {{COMMAND_2_EXIT_OK}} |
| {{COMMAND_3}} | {{COMMAND_3_PURPOSE}} | {{COMMAND_3_INPUTS}} | {{COMMAND_3_OUTPUTS}} | {{COMMAND_3_EXIT_OK}} |

### Flags

| Flag | Short | Type | Default | Applies to | Description |
|------|-------|------|---------|------------|-------------|
| --help | -h | flag | n/a | all | {{HELP_FLAG_BEHAVIOUR}} |
| --version | -V | flag | n/a | all | {{VERSION_FLAG_BEHAVIOUR}} |
| --verbose | -v | flag | false | {{VERBOSE_SCOPE}} | {{VERBOSE_FLAG_BEHAVIOUR}} |
| --quiet | -q | flag | false | {{QUIET_SCOPE}} | {{QUIET_FLAG_BEHAVIOUR}} |
| --format | -f | string | {{DEFAULT_FORMAT\|json}} | {{FORMAT_SCOPE}} | {{FORMAT_FLAG_DESCRIPTION\|accepted values}} |
| --output | -o | path | stdout | {{OUTPUT_SCOPE}} | {{OUTPUT_FLAG_DESCRIPTION}} |
| --config | -c | path | {{DEFAULT_CONFIG_PATH}} | all | {{CONFIG_FLAG_DESCRIPTION}} |
| {{EXTRA_FLAG}} | {{EXTRA_FLAG_SHORT}} | {{EXTRA_FLAG_TYPE}} | {{EXTRA_FLAG_DEFAULT}} | {{EXTRA_FLAG_SCOPE}} | {{EXTRA_FLAG_DESCRIPTION}} |

### Stream contract

| Stream | Contents | Rule |
|--------|----------|------|
| stdout | {{STDOUT_CONTENT\|machine-readable results only}} | {{STDOUT_RULE\|empty when there is no result, never mixed with logs}} |
| stderr | {{STDERR_CONTENT\|diagnostics, progress, warnings, errors}} | {{STDERR_RULE\|always human-readable}} |
| Exit code | {{EXIT_CODE_CONTRACT}} | {{EXIT_CODE_RULE}} |

### Exit codes

| Code | Meaning | Retryable | Example condition |
|------|---------|-----------|-------------------|
| 0 | Success | n/a | {{EXIT_0_CONDITION}} |
| 1 | {{EXIT_1_MEANING\|general failure}} | {{EXIT_1_RETRY}} | {{EXIT_1_CONDITION}} |
| 2 | {{EXIT_2_MEANING\|usage error}} | no | {{EXIT_2_CONDITION}} |
| 3 | {{EXIT_3_MEANING\|input validation failure}} | no | {{EXIT_3_CONDITION}} |
| 4 | {{EXIT_4_MEANING\|partial success}} | {{EXIT_4_RETRY}} | {{EXIT_4_CONDITION}} |
| {{EXIT_EXTRA_CODE}} | {{EXIT_EXTRA_MEANING}} | {{EXIT_EXTRA_RETRY}} | {{EXIT_EXTRA_CONDITION}} |

### Invocation examples

```bash
# Minimal invocation
{{TOOL_NAME}} {{EXAMPLE_1_ARGUMENTS}}

# Explicit input and output
{{TOOL_NAME}} --format {{EXAMPLE_2_FORMAT}} --output {{EXAMPLE_2_OUTPUT}} {{EXAMPLE_2_INPUT}}

# Reading from a stream
cat {{EXAMPLE_3_INPUT}} | {{TOOL_NAME}} --format {{EXAMPLE_3_FORMAT}}

# Batch processing over a directory
{{TOOL_NAME}} --recursive --pattern "{{EXAMPLE_4_PATTERN}}" {{EXAMPLE_4_DIRECTORY}}

# Configuration inspection
{{TOOL_NAME}} config show
```

## Library Interface

| Element | Signature or shape | Stability |
|---------|--------------------|-----------|
| {{LIBRARY_ENTRY_1}} | {{LIBRARY_ENTRY_1_SIGNATURE}} | {{LIBRARY_ENTRY_1_STABILITY\|public and versioned}} |
| {{LIBRARY_ENTRY_2}} | {{LIBRARY_ENTRY_2_SIGNATURE}} | {{LIBRARY_ENTRY_2_STABILITY}} |
| {{LIBRARY_ENTRY_3}} | {{LIBRARY_ENTRY_3_SIGNATURE}} | {{LIBRARY_ENTRY_3_STABILITY}} |

- Module and package layout: {{PACKAGE_LAYOUT}}
- Public surface commitment: {{API_STABILITY_POLICY|what may change in a major version only}}
- Versioning scheme: {{VERSIONING_SCHEME|which scheme, and what each increment means}}
- Minimum supported runtime version: {{MIN_RUNTIME_VERSION}}
- Deprecation policy: {{LIBRARY_DEPRECATION_POLICY}}

## Configuration

| Key | Type | Default | Environment override | Effect |
|-----|------|---------|----------------------|--------|
| {{CONFIG_KEY_1}} | {{CONFIG_KEY_1_TYPE}} | {{CONFIG_KEY_1_DEFAULT}} | {{CONFIG_KEY_1_ENV}} | {{CONFIG_KEY_1_EFFECT}} |
| {{CONFIG_KEY_2}} | {{CONFIG_KEY_2_TYPE}} | {{CONFIG_KEY_2_DEFAULT}} | {{CONFIG_KEY_2_ENV}} | {{CONFIG_KEY_2_EFFECT}} |

- File format and location: {{CONFIG_FORMAT|file format}} at {{CONFIG_LOCATION|default path}}
- Precedence, highest first: {{CONFIG_PRECEDENCE|flag, environment, project file, user file, built-in default}}
- Validation: {{CONFIG_VALIDATION|what happens to an unknown key or an unparsable value}}
- Secrets in configuration: {{CONFIG_SECRETS_POLICY}}

## Input and Output Handling

| Concern | Decision | Rationale |
|---------|----------|-----------|
| Accepted input formats | {{INPUT_FORMATS}} | {{INPUT_FORMATS_RATIONALE}} |
| Rejected input | {{INPUT_REJECTION_RULE}} | {{INPUT_REJECTION_RATIONALE}} |
| Size and count limits | {{INPUT_LIMITS}} | {{INPUT_LIMITS_RATIONALE}} |
| Encoding handling | {{ENCODING_HANDLING}} | {{ENCODING_HANDLING_RATIONALE}} |
| Line-ending handling | {{LINE_ENDING_HANDLING}} | {{LINE_ENDING_HANDLING_RATIONALE}} |
| Output serialisation | {{OUTPUT_SERIALISATION}} | {{OUTPUT_SERIALISATION_RATIONALE}} |
| Partial failure reporting | {{PARTIAL_FAILURE_REPORTING}} | {{PARTIAL_FAILURE_RATIONALE}} |
| Determinism | {{DETERMINISM_RULE\|same input and config produce byte-identical output}} | {{DETERMINISM_RATIONALE}} |

## Extensibility

| Extension point | Contract | Who may register | Compatibility promise |
|-----------------|----------|------------------|------------------------|
| {{EXTENSION_1}} | {{EXTENSION_1_CONTRACT}} | {{EXTENSION_1_PROVIDERS}} | {{EXTENSION_1_COMPATIBILITY}} |
| {{EXTENSION_2}} | {{EXTENSION_2_CONTRACT}} | {{EXTENSION_2_PROVIDERS}} | {{EXTENSION_2_COMPATIBILITY}} |

- Discovery mechanism: {{EXTENSION_DISCOVERY|how extensions are found}}
- Isolation and failure behaviour: {{EXTENSION_ISOLATION|what happens when an extension misbehaves}}
- Version compatibility: {{EXTENSION_COMPATIBILITY_RULE}}

## Distribution and Installation

| Target platform | Architectures | Package | Install command | Verified by |
|-----------------|---------------|---------|-----------------|-------------|
| {{PLATFORM_1}} | {{PLATFORM_1_ARCHITECTURES}} | {{PLATFORM_1_PACKAGE}} | {{PLATFORM_1_INSTALL}} | {{PLATFORM_1_VERIFICATION}} |
| {{PLATFORM_2}} | {{PLATFORM_2_ARCHITECTURES}} | {{PLATFORM_2_PACKAGE}} | {{PLATFORM_2_INSTALL}} | {{PLATFORM_2_VERIFICATION}} |
| {{PLATFORM_3}} | {{PLATFORM_3_ARCHITECTURES}} | {{PLATFORM_3_PACKAGE}} | {{PLATFORM_3_INSTALL}} | {{PLATFORM_3_VERIFICATION}} |

- Binary size budget: {{BINARY_SIZE_BUDGET}}
- Container image, if published: {{CONTAINER_IMAGE}}
- Build from source: {{BUILD_FROM_SOURCE_COMMAND}}
- Checksums and signing: {{SIGNING_PROCEDURE|what a user verifies before running a download}}
- Release channel and cadence: {{RELEASE_CADENCE}}

### Update mechanism

| Behaviour | Decision |
|-----------|----------|
| Version check | {{UPDATE_CHECK_BEHAVIOUR}} |
| Automatic update | {{UPDATE_MECHANISM\|opt-in, opt-out, or none}} |
| Configuration migration across versions | {{CONFIG_MIGRATION_BEHAVIOUR}} |
| Rollback to a previous version | {{VERSION_ROLLBACK_BEHAVIOUR}} |

## Documentation

| Artefact | Contents | Owner | Status |
|----------|----------|-------|--------|
| README | {{README_CONTENTS}} | {{README_OWNER}} | {{README_STATUS\|not started}} |
| CLI reference | {{CLI_REFERENCE_CONTENTS}} | {{CLI_REFERENCE_OWNER}} | {{CLI_REFERENCE_STATUS}} |
| Configuration guide | {{CONFIG_GUIDE_CONTENTS}} | {{CONFIG_GUIDE_OWNER}} | {{CONFIG_GUIDE_STATUS}} |
| Examples | {{EXAMPLES_CONTENTS}} | {{EXAMPLES_OWNER}} | {{EXAMPLES_STATUS}} |
| Troubleshooting | {{TROUBLESHOOTING_CONTENTS}} | {{TROUBLESHOOTING_OWNER}} | {{TROUBLESHOOTING_STATUS}} |
| Library guide | {{LIBRARY_GUIDE_CONTENTS}} | {{LIBRARY_GUIDE_OWNER}} | {{LIBRARY_GUIDE_STATUS}} |
| Extension guide | {{EXTENSION_GUIDE_CONTENTS}} | {{EXTENSION_GUIDE_OWNER}} | {{EXTENSION_GUIDE_STATUS}} |

## Reliability and Security

| Concern | Decision | Verified by |
|---------|----------|-------------|
| Input validation | {{RELIABILITY_INPUT_VALIDATION}} | {{RELIABILITY_INPUT_VERIFICATION}} |
| File system safety | {{RELIABILITY_FS_SAFETY\|path traversal, symlinks, and permissions}} | {{RELIABILITY_FS_VERIFICATION}} |
| Network access | {{RELIABILITY_NETWORK\|whether the tool makes any, and to where}} | {{RELIABILITY_NETWORK_VERIFICATION}} |
| Atomicity | {{RELIABILITY_ATOMICITY\|partial writes and interrupted runs}} | {{RELIABILITY_ATOMICITY_VERIFICATION}} |
| Dependency policy | {{DEPENDENCY_POLICY}} | {{DEPENDENCY_VERIFICATION}} |
| Telemetry | {{TELEMETRY_POLICY\|opt-in, opt-out, or none, and what is collected}} | {{TELEMETRY_VERIFICATION}} |
| Resource ceiling | {{RESOURCE_CEILING}} | {{RESOURCE_VERIFICATION}} |

## Testing Strategy

| Level | Scope | Target | Tools |
|-------|-------|--------|-------|
| Unit | Core processing and utilities | {{TEST_COVERAGE_MIN\|90}}% | {{UNIT_TEST_TOOLS}} |
| Command | Each command, flag, and exit code | {{COMMAND_TEST_COVERAGE}} | {{COMMAND_TEST_TOOLS}} |
| Golden file | Output compared byte-for-byte against a fixture | {{GOLDEN_FIXTURE_COUNT}} fixtures | {{GOLDEN_TEST_TOOLS}} |
| Integration | File input, configuration, plugins | {{INTEGRATION_COVERAGE_TARGET}} | {{INTEGRATION_TEST_TOOLS}} |
| End to end | Real invocation of the built binary | {{E2E_SCENARIO_COUNT}} scenarios | {{E2E_TEST_TOOLS}} |
| Cross-platform | The same suite on every supported platform | {{PLATFORM_MATRIX}} | {{PLATFORM_TEST_TOOLS}} |
| Performance | Large inputs, peak memory, cold start | {{PERFORMANCE_PROFILE}} | {{PERFORMANCE_TEST_TOOLS}} |

- [ ] {{TEST_SCENARIO_1|e.g. invalid flag exits with the documented usage code}}
- [ ] {{TEST_SCENARIO_2}}
- [ ] {{TEST_SCENARIO_3}}
- [ ] {{TEST_SCENARIO_4}}
- [ ] {{TEST_SCENARIO_5}}

## Success Metrics

| Metric | Target | Measured by | Baseline |
|--------|--------|-------------|----------|
| Runtime on the reference workload | {{RUNTIME_TARGET}} | {{RUNTIME_MEASUREMENT}} | {{RUNTIME_BASELINE}} |
| Peak memory on the reference workload | {{MEMORY_TARGET}} | {{MEMORY_MEASUREMENT}} | {{MEMORY_BASELINE}} |
| Installation success rate | {{INSTALL_SUCCESS_TARGET}} | {{INSTALL_MEASUREMENT}} | {{INSTALL_BASELINE}} |
| Error rate | {{ERROR_RATE_TARGET}} | {{ERROR_RATE_MEASUREMENT}} | {{ERROR_RATE_BASELINE}} |
| Documentation coverage of flags | {{DOC_COVERAGE_TARGET}} | {{DOC_COVERAGE_MEASUREMENT}} | {{DOC_COVERAGE_BASELINE}} |
| Test coverage | {{TEST_COVERAGE_MIN\|90}}% | {{COVERAGE_MEASUREMENT}} | {{COVERAGE_BASELINE}} |

## Delivery Stages

- [ ] {{STAGE_1|foundation: project structure, command skeleton, test harness}} — {{STAGE_1_TARGET}}
- [ ] {{STAGE_2|core processing, input and output handling, configuration}} — {{STAGE_2_TARGET}}
- [ ] {{STAGE_3|documentation, packaging, and release pipeline}} — {{STAGE_3_TARGET}}
- [ ] {{STAGE_4|extensions, platform matrix, and community feedback}} — {{STAGE_4_TARGET}}

## Open Questions

| # | Question | Why it matters | Owner | Blocking what | Resolution date |
|---|----------|----------------|-------|----------------|-----------------|
| Q1 | {{QUESTION_1}} | {{QUESTION_1_IMPACT}} | {{QUESTION_1_OWNER}} | {{QUESTION_1_BLOCKS}} | {{QUESTION_1_DUE}} |
| Q2 | {{QUESTION_2}} | {{QUESTION_2_IMPACT}} | {{QUESTION_2_OWNER}} | {{QUESTION_2_BLOCKS}} | {{QUESTION_2_DUE}} |
| Q3 | {{QUESTION_3}} | {{QUESTION_3_IMPACT}} | {{QUESTION_3_OWNER}} | {{QUESTION_3_BLOCKS}} | {{QUESTION_3_DUE}} |
| Q4 | {{QUESTION_4}} | {{QUESTION_4_IMPACT}} | {{QUESTION_4_OWNER}} | {{QUESTION_4_BLOCKS}} | {{QUESTION_4_DUE}} |

## Completion Checklist

Nothing below is true when this file is created. Each box is ticked only when the
thing it names has actually happened; an unticked box is the honest state.

- [ ] Every section above is filled in, not merely headed.
- [ ] Every `{{PLACEHOLDER}}` token in this document has been replaced.
- [ ] Every command and flag has a documented description and an exit code.
- [ ] Every supported platform has been exercised, not assumed.
- [ ] Every output format has a golden fixture.
- [ ] Every assumption is labelled as fact, inferred, speculative, or unknown.
- [ ] Every threshold used here matches the active-dimension values in the constitution.

## Handover

| Item | Value |
|------|-------|
| Next artifact | design.template.md |
| Reviewers | {{REVIEWERS}} |
| Approval required from | {{APPROVERS}} |
| Estimated timeline | {{ESTIMATED_TIMELINE}} |
| Distribution strategy | {{DISTRIBUTION_STRATEGY}} |
