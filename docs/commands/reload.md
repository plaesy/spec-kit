# `reload`

```bash

plaesy reload [flags]
```

Refresh the generated parts of .plaesy/ from the current sources.

'plaesy init' never overwrites a file that already exists, so a project
initialised before an instruction or template was edited keeps the stale copy
permanently: repair and upgrade are stubs, so there was no command that would
bring it forward. This is that command.

Reload overwrites generated files and leaves everything else alone. It will not
overwrite memory.md, context.md or state.json, and it never writes specs/,
tasks/, analysis/ or memory/. A file with no counterpart in the sources is left
in place, so a hand-written template survives.

Every run refreshes the role mirror in .plaesy/roles/, with or without --ai.
The framework's own instructions delegate to those files, so they must match
the sources whatever AI platform the project uses.

That last rule is right for files you wrote and wrong for files reload wrote.
A prompt renamed in prompts/ leaves a complete, still-invokable command
behind in the mirror, so --prune lists orphans in the tool-owned trees (the
agent roles and the platform prompt mirror) and --prune-apply removes them.
It touches nothing else: no .plaesy/ content, no memory, and never a
destination the platform shares with hand-written files. Deletion is opt-in
twice because these directories are gitignored and version control cannot undo
it.

The AI platform is not guessed -- it is not recorded anywhere on disk, so a
wrong guess would write a second platform's files into the project. Pass --ai
to refresh the platform core file, prompts and the platform's own agent
directory as well (for example .claude/agents/). The prompt mirror is only
pruned for the platform you name, for the same reason. The .plaesy/roles/
mirror is not platform-specific, so it does not need --ai.

Usage:
  plaesy reload [directory] [flags]

See [CLI Reference](../reference.md).
