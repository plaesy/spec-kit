# Instruction case files

Checks declared against the instruction files and the design lint. Read the names
carefully: most of these checks test wording and the lint, not model behaviour. Only
the brief cases test a model, and they run only when a runner is configured.

```text
make evals                       # go tests: case files, fixtures, runner parsing
plaesy validate evals            # the same cases, one line each, with a skip notice
plaesy validate design <path>    # design lint on any produced frontend folder
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
PLAESY_EVAL_RUNNER='"C:\tools\my runner.exe" --fast' plaesy validate evals
PLAESY_EVAL_RUNNER='["C:\\tools\\my runner.exe", "--fast"]' plaesy validate evals
```

The first is a command line in which a quoted group is one argument. Backslashes are
kept, so an unescaped Windows path works. The second is a JSON array of strings, which
is unambiguous. Either way, a runner that exits 0 but writes no lintable file is a
FAIL, not a pass.

Without the variable, every brief is reported `SKIP`, and the summary says that no
model behaviour was checked. A skip is not a pass. The repository does not bundle a
model, and CI does not run brief cases.

## Adding a case

1. Decide what must be true: a wording that must be present or absent, or an output
   that must lint a certain way.
2. For wording, add a `static` entry in the instruction's case file. Keep the phrase
   specific enough that a reader cannot satisfy it by accident.
3. For output, add a fixture folder and an `outputs` entry with the rule IDs that must
   appear. Rule IDs are in `internal/designlint`.
