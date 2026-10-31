# `status`

```bash

plaesy status
```

Show Plaesy CLI version and environment status. Reports:

- git availability
- resolved git repository root (uses `GetRepoRoot()` with Windows path normalization)
- Banner with version (`internal/common.Version`, set at build time)

See [CLI Reference](../reference.md).
