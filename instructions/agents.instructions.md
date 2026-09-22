At the beginning of every AI session, read `.plaesy/instructions/plaesy.md`

## Usage Example

This file is copied verbatim as the platform's core entry point (`CLAUDE.md`,
`AGENTS.md`, etc.) by `plaesy init` — it is never edited per-project. A session
picks it up automatically:

```
$ plaesy init my-project
# → writes CLAUDE.md / AGENTS.md containing this file's single line
$ cd my-project && claude
# → the AI tool reads CLAUDE.md first, which redirects it to
#   .plaesy/instructions/plaesy.md for the full protocol
```