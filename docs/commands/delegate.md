# `delegate`

Send a bounded sub-task to a model provider configured under `delegate.providers`.
Only enabled providers run. A billable provider (for example the Anthropic API) also
needs `--confirm`, and it is refused without it.

```bash
plaesy delegate --list                          # providers, enabled and billable state
plaesy delegate --provider claude-cli --prompt "summarise docs/reference.md"
plaesy delegate --chain local --prompt "..."    # named fallback chain
plaesy delegate report                          # token usage per provider and chain
```

Usage is logged to `.plaesy/analysis/delegate-usage.jsonl`. A provider with a
`daily_token_budget` is refused once the day's estimate would exceed it.

See [CLI Reference](../reference.md).
