# `eval`

```bash

plaesy eval --golden-set <file> [flags]
```

Run the search evaluation harness to measure retrieval quality (Recall@k,
MRR, nDCG) against a golden set, across one or more retrieval modes.

| Flag | Default | Description |
|---|---|---|
| `--golden-set` | `""` | Path to golden set JSONL file (required) |
| `--output` | `""` (stdout) | Output file path |
| `--format` | `json` | Output format: `json`, `markdown`, `csv` |
| `--modes` | `vector,bm25,hybrid` | Comma-separated modes to evaluate |
| `--top-k` | `50` | Number of results to retrieve per query |
| `--timeout` | `30s` | Per-query timeout |
| `--warmup` | `0` | Number of warmup queries |
| `--parallel` | `false` | Run queries in parallel |
| `--workers` | `4` | Max parallel workers |
| `--k-values` | `1,3,5,10,20,50` | Comma-separated k values for recall@k |
| `--seed` | `42` | Random seed |

**Example**: `plaesy eval --golden-set testdata/golden.jsonl --format markdown`

See [CLI Reference](../reference.md).
