---
name: token-saver
description: >-
  Reduce token use in agent sessions without losing information: short answers that
  keep code and paths exact, fewer repeated file reads, and compact command output
  through scripts/compact_output.py. Use when a session is long or costly, when a
  command prints a large log or test run, or when the user asks to cut tokens.
---

# token-saver

Measure first, then cut. Tokens go to three places: the model's replies, the files it
reads, and the command output it reads. Cut each one without changing what is said.

## Replies

- Answer first. `[result] [reason]. [next step].` No greeting, no restating the
  question, no recap of what was just done.
- One idea per sentence. Short words over long ones. Keep *not*, *never*, *only*,
  and *no* exactly, because they change the meaning.
- Keep code, commands, paths, file names, numbers, units, and error messages exactly
  as they are. Never shorten them.
- Do not drop a caveat that changes what the reader should do. Security warnings,
  irreversible steps, and open questions keep full sentences.
- Do not rewrite the user's own prompts or quoted text.

## Reads

- Search before reading. Use grep or a symbol search to find the line, then read only
  that range.
- Do not read a file again in the same session unless it changed. Note what you need
  from it the first time.
- Read a file by range (`offset` and `limit`) when it is longer than the part you need.

## Command output

- Quiet flags first: `-q`, `--silent`, `--no-color`, `--quiet`.
- For output longer than about 200 lines, pipe it through `scripts/compact_output.py`:
  `<command> 2>&1 | python scripts/compact_output.py`. It strips colour codes, collapses
  repeated lines, keeps the first and last lines of long runs, and reports how many lines
  it removed. It keeps every line that mentions an error, failure, warning, or panic.
- Never pipe output through the script when the exact text matters, for example a
  diff you must apply or a value you must copy.

## Check

After a change, compare the size of the output before and after:
`wc -c` on the raw and compacted text. Report the saving in bytes, not a guessed token
count.

## Sources

Approach from two public projects: [rtk](https://github.com/rtk-ai/rtk), which
filters command output before it reaches the model, and
[caveman](https://github.com/JuliusBrussee/caveman), which keeps replies short while
leaving code exact. Reported savings from those projects are theirs, not measured here.
