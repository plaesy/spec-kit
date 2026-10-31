# `validate`

```bash

plaesy validate [target|file] [args...] [flags]
```

One verb for every check. Three ways to call it:

```bash

plaesy validate                      # every project-wide check, with a summary
plaesy constitution validate         # one target, with its own flags
plaesy validate deck.pptx            # a document, dispatched by extension
```

With no arguments it runs each project-wide check — constitution, memory,
Markdown, and assumptions — and prints what ran, what was skipped, and why. A check that does not
apply (no constitution yet, no Markdown config) is reported as `[SKIP]` with the
reason, never silently passed over: an omitted check that prints nothing reads as
a check that passed.

| Target | Scope | What it checks |
|---|---|---|
| `constitution` | project-wide | The generated constitution against the contract its template states |
| `memory` | project-wide | `.plaesy/memory/*.md` for external, non-self-contained references |
| `markdown` | project-wide | Markdown style, structure, and image/link accessibility |
| `docx` | file | `.docx` OOXML structure |
| `pptx` | file | `.pptx` OOXML structure, slide count, title placeholders |
| `xlsx` | file | `.xlsx` OOXML structure, sheet names |
| `assumptions` | project-wide | `ASSUMED —` tags left by autonomous runs, optionally gated by age |
| `design` | path | Produced frontend output against the design rules: placeholder image hosts, WCAG 2.2 AA contrast, render evidence |
| `evals` | repository | Case files under `evals/`: phrase checks, lint fixtures, and model briefs only when `PLAESY_EVAL_RUNNER` is set |

`plaesy validate --list` prints the same table from the binary, so it cannot drift
from the code.

The three Office targets validate the container directly: the file must be a ZIP
holding `[Content_Types].xml` and its main part, every XML part must contain a
well-formed *root element* (bare text and empty parts are rejected), and the
document-specific rules follow — a docx needs at least one paragraph or table and
no leftover `{{ }}` template tags, a pptx's slide count comes from its
`sldIdLst` (and a missing title placeholder is a warning, not a failure), an
xlsx needs at least one sheet and its worksheets must parse. The headless-render
check the shell versions performed with LibreOffice is not available in this Go
port, and the command prints a notice saying so rather than doing less silently.

A `.md` argument routes to `markdown`; `.docx`/`.pptx`/`.xlsx` route to their own
validators. Extra expectations (a slide count, expected sheet names) are only
available through the explicit target form, because with several files there is
no unambiguous owner for a trailing argument.

Anything else is an error naming the valid targets, rather than a guess.

## Legacy spellings

`plaesy constitution validate`, `validate-memory`, `validate-markdown`,
`validate-docx`, `validate-pptx`, and `validate-xlsx` still run and print where
they moved. They are deprecated and will be removed; the subcommand form is the
supported one.

## `constitution validate`

```bash

plaesy constitution validate [path]
```

Validate a generated project constitution (default
`.plaesy/memory/constitution.md`) against the contract
`templates/constitution.template.md` states about itself:

- frontmatter carries `title`, `version`, `ratified`, `last_amended`,
  `active_dimensions`; dates are ISO 8601
- no unfilled `{{ ... }}` placeholder survives
- every §1 Active Dimensions row is a known dimension marked exactly `yes`/`no`,
  and at least one is active
- frontmatter `active_dimensions` matches the §1 rows marked `yes` (the table is
  the source of the active scope)
- the frontmatter version matches the last amendment-log row
- rule IDs (`EV-`/`QG-`/`DOC-`/`SC-`/`RTM-`/`NN-`) are unique

Prints the declared version, active dimensions, and rule/non-negotiable counts.
Exits non-zero and lists every problem at once if the constitution is unfilled,
unratified, or internally inconsistent. Run it after `/start` (Phase 1, sub-step 1.0) and after
any amendment.

## `memory validate`

```bash

plaesy memory validate
```

Scan `.plaesy/memory/*.md` for external, non-self-contained references
(`~/.claude/`, `/tmp/`, `C:\Users\...\.claude`). Reports files, patterns, and
line numbers. Exits non-zero if issues found.

## `markdown validate`

```bash

plaesy markdown validate [paths...] [flags]
```

Lint Markdown with the linter built into this binary. No Node, no npm, no
`node_modules` — for this repository and for any project that installed the
framework. The rule IDs and config keys are markdownlint's, so
`.markdownlint.json` stays portable.

25 rules are implemented: MD001, MD003, MD004, MD007, MD009, MD010, MD012,
MD013, MD022, MD024, MD025, MD026, MD029, MD031, MD032, MD036, MD040, MD041,
MD046, MD047, MD048, MD058, plus the three GitHub accessibility rules GHA001
(alt text), GHA002 (alt text that is not just a file name) and GHA003 (link text
that describes its destination). `plaesy markdown validate --list-rules` prints
the set with descriptions; a clean run means "no implemented rule fired", not
"markdownlint would be happy".

| Flag | Effect |
|------|--------|
| `--config <path>` | Config file (default `<repo>/.markdownlint.json`) |
| `--baseline <path>` | Ratchet baseline (default `<repo>/.markdownlint-baseline.json`) |
| `--no-baseline` | Ignore the baseline; require zero violations |
| `--summary` | Print only the per-rule histogram |
| `--max-files <n>` | How many offending files to detail (0 = all, default 20) |
| `--fix` | Apply the machine-fixable rules in place |
| `--write-baseline` | Re-measure and record the current total as the new ceiling |
| `--list-rules` | List implemented rules and exit |

Two behaviours are deliberate and differ from the npm tooling:

- **Unknown rule names and unknown option keys are hard errors.** A config that
  half-loads is a check that reports success while checking nothing. `extends`
  is rejected outright — there is no registry to resolve it against, so the
  rules a project needs must live in the config file. `$`-prefixed keys are
  treated as human documentation and ignored.
- **The default run is a ratchet.** Pre-existing debt should not block unrelated
  work, but new violations must fail. The baseline records a ceiling and the
  per-rule counts behind it; a run passes while the total stays at or below the
  ceiling and fails when it rises. `TestRepoMarkdownStaysWithinRatchet` in
  `scripts/internal/mdlint` runs the same check inside `go test ./...`, which is
  the only CI job — that is what keeps Markdown linting alive without a second
  pipeline to forget. Lower the ceiling when you fix things; raising it to make
  a red build green defeats the point.

The ratchet is measured over the whole tree, so linting a single file compares
that file against the repository-wide ceiling. Pass `--no-baseline` when you
want zero tolerance for one file.

Machine-fixable rules: MD009, MD010, MD012, MD022, MD031, MD032, MD047, MD058.

`--fix` derives its decisions from the same predicates the corresponding check
reports on, and `internal/mdlint/fix_test.go` lints the output of every fix pass
and fails when a fixable rule still fires. That test is not ceremony: the fixer
once asked whether the wrong line continued the block above it, so a list below a
heading never got its blank line and the command reported `[OK] nothing to fix` on
a file the linter had just listed 20 fixable violations for. A fix pass that
changes nothing and a document that is already clean look identical from the
outside, so the pass is checked by what it leaves behind.

## `assumptions validate`

```bash

plaesy assumptions validate [--review-days <n>]
```

Scans the corpus for `ASSUMED —` tags: the label an autonomous run writes
instead of stopping to ask (`instructions/plaesy.instructions.md` rule 2). By
default it is an audit, not a gate — it always lists what it found and exits
`0`. It becomes a gate only when `.plaesy/state.json`'s
`assumptions_review_days` is a positive integer (the same opt-in convention as
`autonomous_loop.checkpoint_interval`), in which case a tag whose file has had
no commit in over that many days fails the run — surfacing it for a batched,
asynchronous human review rather than blocking the autonomous run that wrote
it in the first place.

| Flag | Effect |
|------|--------|
| `--review-days <n>` | Override `.plaesy/state.json`'s `assumptions_review_days` for this run (`0` = use the config, or informational-only if unset) |

## `docx validate` / `pptx validate` / `xlsx validate`

```bash

plaesy docx validate <file.docx>
plaesy pptx validate <file.pptx> [expected_slide_count]
plaesy xlsx validate <file.xlsx> [expected_sheet_name...]
```

Validate OOXML structure without rendering. Reports paragraph/table/slide/sheet
counts and structural issues.

## `design validate`

```bash
plaesy design validate <path...> [--evidence <dir>] [--strict]
```

Lints produced HTML, CSS, JSX, TSX, Vue and Svelte files without a browser. FAIL
findings: pictures from placeholder hosts (`DL-IMG-HOST`); text colour pairs below
WCAG 2.2 AA (`DL-CONTRAST`, 4.5:1 normal text, 3:1 large text); and, with
`--evidence`, no screenshot at 1440, 768 or 390 px (`DL-VIEWPORTS`). WARN findings:
placeholder-looking figures in visible text (`DL-INVENTED-FIGURE`), Inter, Arial or
Roboto as the first font (`DL-DEFAULT-FONT`), and a token declared with different
values in different files (`DL-TOKEN`). WARN fails the run only with `--strict`.

What is read: CSS rules (including `@media` blocks), inline `style=""` attributes, and
JSX `style={{ }}` object literals. What is not read: utility classes (Tailwind and
similar), CSS-in-JS helpers, `style={variable}` and Vue `:style` bindings. Those are
counted as skipped, and the summary prints the count.

Colours: `#rgb`, `#rgba`, `#rrggbb`, `#rrggbbaa`, `rgb()` and `rgba()` with 0-255
channels, the keywords `black`, `white` and `transparent`, and `var(--token)` for a
token declared in the same run. Text with alpha is composited over its background.
Sizes: px, rem, em, % and `clamp()` (the smallest bound). em and % are taken against
a 16px root, and the finding says so. A rule with a gradient, an image, an inherited
colour, a background from elsewhere in the cascade, or an unreadable size is skipped
and counted. The verdict is the same for the same input on every run.

`--evidence` names a directory of screenshots. Each file name must contain the width
as its own number (`home-768.png` counts; `1768.png` does not).

## `evals validate`

```bash
plaesy evals validate
PLAESY_EVAL_RUNNER='["<command>", "<arg>"]' plaesy evals validate
```

Runs every case file under `evals/cases`. Three kinds of case:

- **static**: checks that an instruction contains required wording and none of the
  forbidden wording. It confirms that the words are present. It does not check meaning.
- **outputs**: lints a hand-written fixture folder under `evals/fixtures` and checks the
  expected result. It tests the lint on known inputs. The fixtures are not model output.
- **briefs**: the only check of model behaviour. It needs `PLAESY_EVAL_RUNNER`, run as
  `<command> <brief-id> <output-dir>` with the brief in `PLAESY_EVAL_BRIEF`. The runner
  must write at least one lintable file, with no FAIL findings. Without the variable
  every brief is reported `SKIP`, the summary says that no model behaviour was checked,
  and a skip is never a pass. `PLAESY_EVAL_RUNNER` is a JSON array, or a command line
  where a "quoted group" is one argument; backslashes are kept, so Windows paths work.

See `evals/README.md`.

See [CLI Reference](../reference.md).
