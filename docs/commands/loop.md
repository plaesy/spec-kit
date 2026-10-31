# plaesy loop

Runs the autonomous loop driver for the current project.

## Usage

```bash
plaesy loop
plaesy loop stop-hook
```

## What it does

- `plaesy loop` runs iterations until the run ends. Each iteration runs the step set by `"step"` in `.plaesy/loop.json`:
  - `{"kind":"command","command":"<cmd>","args":[...],"timeout_seconds":N}` runs the command with the step prompt on stdin. The step passes when the command exits 0, and its stdout is saved to `.plaesy/loop/jobs/step-<n>.out`. A command that is not on PATH is a failed step that names the command.
  - `{"kind":"openai","base_url":"...","model":"...","api_key_env":"ENV_NAME"}` sends the step prompt as one chat message to an OpenAI-compatible endpoint (`base_url` includes the version prefix, for example `.../v1`). `api_key_env` names the environment variable that holds the key; the key is never read from or written to the file. A missing variable or a non-2xx response is a failed step with the reason.
  - Both kinds use the provider types of `plaesy delegate` (`cli` and `openai`).
  - The `command` kind is an example shape only; no CLI flags are verified for any tool. Write the command your tool needs.
  - With no `step`, the older `"step_command"` shell command runs. With neither, every iteration is not clean, and the run stops after `max_consecutive_failures` failed cycles in a row (default 2).
- A run that stops writes a report to `.plaesy/loop/stop-report.md` and sets `autonomous_loop.status` in `.plaesy/state.json`.
- `plaesy loop stop-hook` reads the Stop hook input from stdin. While the state is `running` and `stop_hook_active` is false, it prints a message to stderr and exits 2, which keeps the turn going. When `stop_hook_active` is true, it exits 0, so the turn can end once.

## Known gaps

These are tracked as backlog tasks, not yet fixed:

- The Stop hook is not yet installed by `plaesy init`.
