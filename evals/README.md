# Instruction case files

Checks declared against the instruction files and the design lint. Read the names
carefully: most of these checks test wording and the lint, not model behaviour. Only
the brief cases test a model, and they run only when a runner is configured.

```text
make evals                       # go tests: case files, fixtures, runner parsing
plaesy evals validate            # the same cases, one line each, with a skip notice
plaesy design validate <path>    # design lint on any produced frontend folder
```

## Layout

- `cases/<instruction>.json`: one file per instruction it tests. Unknown keys are
  rejected, so a misspelt key cannot silently drop a check.
- `fixtures/<name>/`: small frontend folders, hand-written, with known results for
  the lint. They are not model output.

## Case kinds

| Kind | What it checks | What it does not check | Runs by default |
|---|---|---|---|
| `static` | The instruction text contains `requires` phrases and none of the `forbids` phrases. | Meaning. A phrase in a sentence that says the opposite still passes. | Yes |
| `outputs` | A fixture folder lints as `expect` (`pass`: no FAIL; `fail`: at least one FAIL), and the rules in `rules` appear. | Any model. The fixtures are written by hand. | Yes |
| `briefs` | A model writes a frontend folder for the brief, and the folder lints clean with at least one file. | Anything unless a runner is set. | No: reported as SKIP |

Only `briefs` tests what a model produces, and it needs `PLAESY_EVAL_RUNNER`.

## Running brief cases

Set `PLAESY_EVAL_RUNNER` to a command. It is called as
`<command> <brief-id> <output-dir>` with the brief text in `PLAESY_EVAL_BRIEF`. It
must write the frontend files into `<output-dir>`. Two forms are accepted:

```text
PLAESY_EVAL_RUNNER='"C:\tools\my runner.exe" --fast' plaesy evals validate
PLAESY_EVAL_RUNNER='["C:\\tools\\my runner.exe", "--fast"]' plaesy evals validate
```

The first is a command line in which a quoted group is one argument. Backslashes are
kept, so an unescaped Windows path works. The second is a JSON array of strings, which
is unambiguous. Either way, a runner that exits 0 but writes no lintable file is a
FAIL, not a pass.

Without the variable, every brief is reported `SKIP`, and the summary says that no
model behaviour was checked. A skip is not a pass. The repository does not bundle a
model, and CI does not run brief cases.

## Trials and consistency

Model output varies between runs, so one passing brief shows little. Guidance for brief
cases (Anthropic, "Demystifying evals for AI agents",
https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents, retrieved 2026-10-17):

- **Run several trials per brief.** Each trial is one attempt. Report **pass@k** (at least
  one success in k) only when one good result is enough. Report **pass^k** (all k succeed)
  for behaviour that must be consistent every time, which is the case for a Success
  Contract's blocking items.
- **Grade the outcome, not the path.** Check the files produced and whether they lint, not
  the exact sequence of tool calls. Path-based graders reject valid approaches.
- **Read transcripts before trusting a failing grader.** A failure must be explainable from
  the transcript. If it is not, the grader is wrong.
- **Start from real failures.** 20 to 50 cases taken from failures you have actually seen
  are a sound start. Convert each user-reported failure into a case.

**Gap:** `PLAESY_EVAL_RUNNER` runs each brief once. The repository has no k-trial runner and
no pass^k report yet, so brief results today are single-trial and must not be read as
consistency results.

## Adding a case

1. Decide what must be true: a wording that must be present or absent, or an output
   that must lint a certain way.
2. For wording, add a `static` entry in the instruction's case file. Keep the phrase
   specific enough that a reader cannot satisfy it by accident.
3. For output, add a fixture folder and an `outputs` entry with the rule IDs that must
   appear. Rule IDs are in `internal/designlint`.
