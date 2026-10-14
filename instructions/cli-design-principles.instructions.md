---
applyTo: '**/cmd/**,**/cli/**,**/*.cli.ts,**/*.cli.go'
description: 'Core CLI/developer-experience design principles (stream discipline, exit codes, machine-readable output, safe destructive commands, consistent flags, help text) to apply when designing or reviewing a command-line tool.'
---

# CLI Design Principles

The golden rule: **a CLI is a script's API before it's a human's tool.** Every
principle below exists to make the tool predictable enough for a script to depend on,
while staying readable enough for a human running it by hand.

Sources (retrieved 2026-09-27): [Command Line Interface Guidelines (clig.dev)](https://clig.dev/),
[Atlassian — 10 design principles for delightful CLIs](https://www.atlassian.com/blog/it-teams/10-design-principles-for-delightful-clis),
[Pigweed — CLI style guide](https://pigweed.dev/docs/style/cli.html),
[DEV Community — 14 great tips to make amazing CLI applications](https://dev.to/wesen/14-great-tips-to-make-amazing-cli-applications-3gp3).

## Respect the streams: stdout is data, stderr is everything else

Send the actual result to `stdout`; send logs, progress, and errors to `stderr`; use
the exit code to say whether it worked. — A tool that prints a progress spinner and
its final result both to `stdout` corrupts every script that pipes its output into
`jq` or another command — the consumer can't tell the real payload from noise mixed
into the same stream.

```text
BAD:  stdout: "Fetching...\nDone.\n{\"id\": 42}"   # progress text corrupts the JSON
GOOD: stderr: "Fetching...\nDone."                 # progress/diagnostics
      stdout: "{\"id\": 42}"                        # only the payload
```

## Exit codes are a contract, not decoration

Return `0` only on success, and use distinct non-zero codes for distinct failure
categories, consistently across every subcommand. — A tool that returns `1` for
"file not found," "network timeout," and "invalid input" alike forces every calling
script to parse stderr text to know what actually went wrong; a script checking
`if [ $? -ne 0 ]` at least gets pass/fail right, but only distinct codes let it branch
on *why*.

## Confirm before anything destructive — never require an interactive prompt to be safe

- A destructive action needs either a **dry-run** mode that shows what would happen
  with no side effects, or an explicit non-interactive confirmation flag
  (`--force`/`--yes`), and it must fail safely (do nothing) when neither is given in a
  non-interactive context. — A `plaesy clean` that deletes files the moment it's
  invoked, with no `--dry-run` and no required `--yes`, turns one fat-fingered Enter
  key (or one CI script that forgot the flag) into unrecoverable data loss.
- Don't gate safety *only* behind an interactive prompt. — A tool that always pauses
  for a `y/N` prompt before deleting is unusable in CI/non-interactive contexts,
  which pushes people toward wrapping it in `yes | tool` — defeating the safety
  entirely instead of giving it a real non-interactive escape hatch.

```text
BAD:  $ plaesy clean
      Deleting 214 files...   # no warning, no way to preview, no flag needed

GOOD: $ plaesy clean --dry-run
      Would delete 214 files (12.4MB). Re-run with --yes to actually delete.
      $ plaesy clean --yes
      Deleted 214 files (12.4MB).
```

## Default to human-first output, offer a machine-readable mode explicitly

Print readable, formatted text by default; support `--json` (or `--output json`) for
scripts, and never mix the two in the same stream. — A tool with no `--json` forces
every script that wants to consume its output to regex-parse a human-oriented format
that can change wording between versions with no warning; an explicit machine-readable
mode is a real, stable contract instead of an accidental one.

## Give error messages a next step, not just a diagnosis

State what went wrong, why, and what to do about it — not just that something failed. —
`Error: failed` tells the user nothing they didn't already know (the command failed;
that's why they're reading the error); `Error: config file not found at ./plaesy.yaml
— run 'plaesy init' to create one` tells them the cause and the exact next command.

## Keep flag names and behavior consistent across every subcommand

Pick one name per concept (`--verbose`, not `-V` in one subcommand and `--debug` in
another) and use it identically everywhere in the tool. — A user who learns
`--dry-run` on `clean` shouldn't have to relearn it as `--preview` on `migrate`; the
same flag name doing the same thing everywhere is what makes a multi-command CLI feel
like one tool instead of several glued together.

## Make `--help` progressive, not exhaustive by default

Show a short, scannable summary by default (`tool cmd --help` → usage + common flags);
put the exhaustive reference (every flag, every edge case) behind that same flag
rather than dumping both at once. — A `--help` that prints 200 lines for a command
with 3 commonly-used flags buries the ones a user actually needs under flags they'll
almost never touch; a short default with a clear pointer to more detail respects both
the new user and the power user.

## Applying these under real constraints

- Not every internal or debug-only subcommand needs the full treatment (dry-run mode,
  `--json`, exhaustive help) — apply these proportionally to how often a command is
  scripted or run by someone other than its author, not uniformly to every internal
  tool.
- Backward compatibility of flags/output format is itself a promise once scripts
  depend on it — treat renaming a flag or changing the default output shape as a
  breaking change requiring a deprecation window, the same way an API would.
- When reviewing a CLI command, name the specific script-breaking or user-harming
  failure a missing principle allows (mixed streams corrupting a pipe, an
  unrecoverable delete with no confirmation, an unparseable error) rather than "this
  isn't user-friendly."
