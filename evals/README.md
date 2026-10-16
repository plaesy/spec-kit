# Behavioural evals

Cases that check whether instruction files say what they must say, and whether
frontend output meets the design rules. Run them with:

```text
make evals                       # go tests: cases, fixtures, runner plumbing
plaesy validate evals            # the same cases, printed one per line
plaesy validate design <path>    # design lint on any produced frontend folder
```

## Layout

- `cases/<instruction>.json`: one file per instruction it tests. Unknown keys are
  rejected, so a misspelt key cannot silently drop a check.
- `fixtures/<name>/`: small frontend folders, used as known-good and known-bad
  output for the lint.

## Case kinds

| Kind | What it checks | Needs an LLM |
|---|---|---|
| `static` | The instruction text contains `requires` phrases and none of the `forbids` phrases. | No |
| `outputs` | A fixture folder lints as `expect` (`pass` = no FAIL; `fail` = at least one FAIL), and the rules in `rules` appear. | No |
| `briefs` | A model writes a frontend folder for the brief, and the folder lints clean. | Yes, optional |

## Running brief cases

Brief cases need a model. The repository does not bundle one. Set
`PLAESY_EVAL_RUNNER` to a command. It is called as
`<command> <brief-id> <output-dir>` with the brief text in `PLAESY_EVAL_BRIEF`.
It must write the frontend files into `<output-dir>`. Without the variable, every
brief is reported `SKIP`. A skip is not a pass.

```text
PLAESY_EVAL_RUNNER="./scripts/my-runner" plaesy validate evals
```

## Adding a case

1. Add the phrase to the instruction, or decide that the phrase must not appear.
2. Add a `static` entry in the instruction's case file that names it.
3. For a behaviour that shows up in output, add a fixture and an `outputs` entry
   with the rule IDs that must appear. Rule IDs are in `internal/designlint`.
