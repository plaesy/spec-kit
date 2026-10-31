# `mcp`

```bash

plaesy mcp [--root <project>]
```

Run a read-only [Model Context Protocol](https://modelcontextprotocol.io) server over stdio for the project's `.plaesy/` folder. Any MCP client can read the project's framework state (memory, decisions, tasks, instructions) without the files being pasted into the chat.

Tools exposed:

- `plaesy_list` — list files under `.plaesy/` (relative path and size in bytes). Optional `path` argument.
- `plaesy_read` — read one text file under `.plaesy/` (`md`, `json`, `yaml`, `yml`, `txt`, `csv`, `toml`). Results are capped at 512 KB.
- `plaesy_write` — create or overwrite one text file under `.plaesy/` with `path` and `content`. Missing folders are created. Content is capped at 512 KB and written atomically.

Guarantees:

- Writes only through `plaesy_write`, and only text types listed above. Nothing is deleted.
- `analysis/` is refused for writes because `plaesy analyze` generates it.
- Paths are confined to `.plaesy/`. Absolute paths, `..` escapes and symlinks that point outside the folder are refused. A write never follows a symlink.
- stdout carries only protocol messages.

Registration: `plaesy init` and `plaesy reload` add a `plaesy` entry to the project's `.mcp.json` (`"command": "plaesy", "args": ["mcp"]`) and keep other entries in that file. The `plaesy` binary must be on `PATH` for the client to start the server.

## Protocol

- Transport: stdio, one JSON-RPC 2.0 message per line.
- Protocol version: negotiated. A supported version requested by the client is echoed back. Supported: `2025-11-25`, `2025-06-18`, `2025-03-26`, `2024-11-05`. For any other version the server answers with `2025-11-25` and the client decides whether to disconnect.
- Advertised capability: `tools` only (no resources, prompts or sampling).
- Errors: malformed requests (unknown tool, missing or mistyped argument, bad params) return a JSON-RPC error with code `-32602`. Unknown methods return `-32601`. Failures while running a valid tool call (file missing, path refused, size limit) return a normal result with `isError: true`.
- Tool annotations: `plaesy_list` and `plaesy_read` are `readOnlyHint: true`. `plaesy_write` is `destructiveHint: true` because it replaces existing content.
- Server info: name `plaesy`, version as reported by `plaesy --version`.
- Methods: `initialize`, `notifications/initialized`, `tools/list`, `tools/call`.

### Tool arguments

| Tool | Arguments | Required |
|------|-----------|----------|
| `plaesy_list` | `path`: folder relative to `.plaesy/` (empty = root) | none |
| `plaesy_read` | `path`: file relative to `.plaesy/`, e.g. `memory.md` | `path` |
| `plaesy_write` | `path`: file relative to `.plaesy/`, e.g. `memory/overview.md`; `content`: full new file content | `path`, `content` |

## Smoke test

Check the server responds without starting a client:

```bash
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"check","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | plaesy mcp
```

A healthy server replies to `initialize` with `serverInfo` and to `tools/list` with the three tools above.

## Troubleshooting

- **`plaesy install` fails with "Access is denied" on Windows.** A running `plaesy mcp` process holds the installed `plaesy.exe` open. Close the MCP client session (or stop the process) and run the install again.
- **Client shows no `plaesy` server.** Check that `plaesy` is on `PATH` and that `.mcp.json` contains the `plaesy` entry.
- **A tool returns an error for a path.** Paths must be relative and stay inside `.plaesy/`; `analysis/` is read-only for MCP writes.

See [CLI Reference](../reference.md).
