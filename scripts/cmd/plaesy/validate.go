package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/mdlint"
	"github.com/plaesy/spec-kit/internal/validate"
	"github.com/spf13/cobra"
)

// validateTarget describes one validator: its name under `plaesy validate`, and a
// constructor for a fresh command instance. A constructor rather than one shared
// instance because the subcommand and its pre-1.0 alias must not share a flag set
// — two commands over one FlagSet redefine flags and panic.
type validateTarget struct {
	name string
	// repoWide is true when the validator can check a whole project with no
	// arguments, which is what bare `plaesy validate` runs.
	repoWide bool
	build    func() *cobra.Command
}

// The pre-1.0 spelling was one top-level command per validator (six of them,
// all the same verb). They collapse into `plaesy validate <target>`; the old
// names stay as deprecated aliases so existing docs, scripts, and muscle memory
// keep working. Cobra hides a deprecated command from help and announces the
// replacement when it runs, so the help text shows one verb while nothing breaks.
var validateTargets = []validateTarget{
	{name: "constitution", repoWide: true, build: newConstitutionCmd},
	{name: "memory", repoWide: true, build: newMemoryCmd},
	{name: "markdown", repoWide: true, build: newMarkdownCmd},
	{name: "assumptions", repoWide: true, build: newAssumptionsCmd},
	{name: "docx", build: newDocxCmd},
	{name: "pptx", build: newPptxCmd},
	{name: "xlsx", build: newXlsxCmd},
}

func init() {
	register(newValidateCmd())
	for _, target := range validateTargets {
		register(legacyValidateAlias(target))
	}
}

// newValidateCmd is the single entry point for every check. Three shapes:
//
//	plaesy validate                    every project-wide check, in one run
//	plaesy validate <target> [args]    one validator, with its own flags
//	plaesy validate <file.docx|...>    an OOXML document, dispatched by extension
func newValidateCmd() *cobra.Command {
	var list bool
	cmd := &cobra.Command{
		Use:   "validate [target|file] [args...]",
		Short: "Validate a constitution, memory, Markdown, or an OOXML document",
		Long: `Run every check, or one of them.

With no arguments, runs each project-wide check — constitution, memory, and
Markdown — and prints a summary naming what ran, what was skipped, and why:

  plaesy validate

Name a target to run just that one, with its own flags:

  plaesy validate constitution [path]
  plaesy validate memory
  plaesy validate markdown [paths...] [--no-baseline] [--fix] [--summary]
  plaesy validate docx <file.docx>
  plaesy validate pptx <file.pptx> [expected_slide_count]
  plaesy validate xlsx <file.xlsx> [expected_sheet_name...]

Or pass a document and let the extension pick the validator:

  plaesy validate deck.pptx

Extra expectations (a slide count, sheet names) are only available through the
explicit target form, because with several files there is no way to know which
argument belongs to which.

The pre-1.0 spellings — plaesy validate-constitution, validate-memory,
validate-markdown, validate-docx, validate-pptx, validate-xlsx — still run and
print where they moved. They are deprecated and will be removed.`,
		Example: `  plaesy validate
  plaesy validate --list
  plaesy validate constitution
  plaesy validate markdown --summary
  plaesy validate xlsx report.xlsx`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if list {
				printValidateTargets(cmd)
				return nil
			}
			if len(args) > 0 {
				return runValidateFiles(args)
			}
			return runValidateProject()
		},
	}
	cmd.Flags().BoolVar(&list, "list", false, "list the validation targets and exit")
	for _, target := range validateTargets {
		cmd.AddCommand(target.build())
	}
	return cmd
}

// legacyValidateAlias republishes a validator under its old top-level name.
func legacyValidateAlias(target validateTarget) *cobra.Command {
	cmd := target.build()
	cmd.Use = "validate-" + cmd.Use
	cmd.Deprecated = fmt.Sprintf("use 'plaesy validate %s' instead", target.name)
	return cmd
}

// printValidateTargets writes to the command's output stream rather than to
// stdout directly, so `validate --list | grep` and a caller redirecting output
// both see the list.
func printValidateTargets(cmd *cobra.Command) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Validation targets:\n\n")
	for _, target := range validateTargets {
		scope := "needs a file"
		if target.repoWide {
			scope = "project-wide"
		}
		fmt.Fprintf(out, "  %-14s %-14s %s\n", target.name, scope, target.build().Short)
	}
	fmt.Fprint(out, "\nBare 'plaesy validate' runs the project-wide ones. A .md file is\n"+
		"routed to the markdown target; a .docx/.pptx/.xlsx to its own.\n")
}

// runValidateProject runs every project-wide check and reports all of them, so a
// failure in the constitution does not hide the state of the memory files.
func runValidateProject() error {
	root, err := common.GetRepoRoot()
	if err != nil {
		root = "."
	}
	constitution := filepath.Join(root, ".plaesy", "memory", "constitution.md")

	checks := []projectCheck{
		{
			name: "constitution",
			skip: skipIfMissing(constitution, "no constitution yet — generate one with /start"),
			run: func() error {
				c := newConstitutionCmd()
				return c.RunE(c, []string{constitution})
			},
		},
		{
			name: "memory",
			// A project that has not run /start has no .plaesy/memory yet. The
			// explicit `validate memory` still reports that as an error, because
			// the user asked for a directory that is not there; the bare run skips
			// it, because "check my project" on a fresh project is a question with
			// an answer, not a failure.
			skip: skipIfMissing(filepath.Join(root, ".plaesy", "memory"), "no .plaesy/memory yet — generate one with /start"),
			run: func() error {
				c := newMemoryCmd()
				return c.RunE(c, nil)
			},
		},
		{
			name: "markdown",
			skip: markdownSkipReason(root),
			run: func() error {
				c := newMarkdownCmd()
				return c.RunE(c, nil)
			},
		},
	}

	fmt.Printf("[INFO] Validating project at %s\n", root)
	passed, failed := 0, 0
	for _, c := range checks {
		fmt.Printf("\n--- %s ---\n", c.name)
		if c.skip != "" {
			fmt.Printf("[SKIP] %s: %s\n", c.name, c.skip)
			continue
		}
		if err := c.run(); err != nil {
			failed++
			continue
		}
		passed++
	}

	fmt.Println()
	fmt.Printf("[INFO] %d passed, %d failed", passed, failed)
	if skipped := countSkipped(checks); skipped > 0 {
		fmt.Printf(", %d skipped", skipped)
	}
	fmt.Println()
	if failed > 0 {
		return fmt.Errorf("%d of %d project checks failed", failed, passed+failed)
	}
	return nil
}

// projectCheck is one entry in the bare `plaesy validate` run.
type projectCheck struct {
	name string
	// skip is non-empty when the check does not apply to this project, and says
	// why. A silently omitted check reads as a passing one.
	skip string
	run  func() error
}

// countSkipped reports how many checks declared themselves inapplicable.
func countSkipped(checks []projectCheck) int {
	n := 0
	for _, c := range checks {
		if c.skip != "" {
			n++
		}
	}
	return n
}

func skipIfMissing(path, reason string) string {
	if _, err := os.Stat(path); err != nil {
		return reason
	}
	return ""
}

// markdownSkipReason keeps `plaesy validate` usable in a project that never
// opted into Markdown linting. With no config and no baseline the lint would
// demand zero violations across the whole tree, which is a different check than
// the one the user asked for.
func markdownSkipReason(root string) string {
	_, cfgErr := os.Stat(filepath.Join(root, mdlint.DefaultConfigName))
	_, baseErr := os.Stat(filepath.Join(root, mdlint.DefaultBaselineName))
	if cfgErr != nil && baseErr != nil {
		return "no " + mdlint.DefaultConfigName + " and no " + mdlint.DefaultBaselineName +
			" — run 'plaesy validate markdown' to opt in"
	}
	return ""
}

// runValidateFiles dispatches documents by extension so the common case needs no
// target name. Extra per-file expectations are not threaded through here: with
// several files there is no unambiguous owner for a trailing argument, and
// guessing would validate the wrong thing silently.
func runValidateFiles(args []string) error {
	// Markdown is included: a file argument means "check this file", and a
	// reader who types `plaesy validate README.md` means the linter, not a
	// usage error.
	byExt := map[string]string{
		".md": "markdown", ".markdown": "markdown",
		".docx": "docx", ".pptx": "pptx", ".xlsx": "xlsx",
	}
	failed := 0
	for _, arg := range args {
		ext := strings.ToLower(filepath.Ext(arg))
		name, ok := byExt[ext]
		if !ok {
			return fmt.Errorf("cannot tell what to validate from %q: pass a target (%s) or a .md/.docx/.pptx/.xlsx file",
				arg, strings.Join(validateTargetNames(), ", "))
		}
		for _, target := range validateTargets {
			if target.name != name {
				continue
			}
			c := target.build()
			if err := c.RunE(c, []string{arg}); err != nil {
				failed++
			}
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d document(s) failed validation", failed, len(args))
	}
	return nil
}

func validateTargetNames() []string {
	names := make([]string, 0, len(validateTargets))
	for _, t := range validateTargets {
		names = append(names, t.name)
	}
	return names
}

// newConstitutionCmd checks a generated project constitution against the
// contract templates/constitution.template.md states about itself: frontmatter
// completeness, no unfilled placeholders, a coherent active-dimension scope
// (frontmatter must agree with the §1 table), version/amendment-log agreement,
// and unique rule IDs. This is the deterministic gate behind the constitution's
// own "every threshold is measured, not asserted" rule — an unvalidated
// constitution silently widens or narrows the scope of every later phase.
func newConstitutionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "constitution [path]",
		Short: "Validate a generated constitution",
		Long: `Validate a generated project constitution (default
.plaesy/memory/constitution.md) against the contract
templates/constitution.template.md states about itself:

- frontmatter carries title, version, ratified, last_amended,
  active_dimensions; dates are ISO 8601
- no unfilled {{ ... }} placeholder survives
- every §1 Active Dimensions row is a known dimension marked exactly yes/no,
  and at least one is active
- frontmatter active_dimensions matches the §1 rows marked yes (the table is
  the source of the active scope)
- the frontmatter version matches the last amendment-log row
- rule IDs (EV-/QG-/DOC-/SC-/RTM-/NN-) are unique

Prints the declared version, active dimensions, and rule/non-negotiable counts.
Exits non-zero and lists every problem at once if the constitution is unfilled,
unratified, or internally inconsistent. Run it after /start Phase 0 and after
any amendment.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) == 1 {
				path = args[0]
			} else {
				repoRoot, err := common.GetRepoRoot()
				if err != nil {
					repoRoot = "."
				}
				path = filepath.Join(repoRoot, ".plaesy", "memory", "constitution.md")
			}

			fmt.Printf("[INFO] Validating constitution: %s\n\n", path)

			res, err := validate.Constitution(path)
			if res != nil {
				fmt.Printf("[INFO] version: %s | active dimensions: %v | rules: %d | non-negotiables: %d\n\n",
					res.Version, res.ActiveDimensions, res.Rules, res.NonNegotiables)
			}
			if err != nil {
				fmt.Println("[✗]", err)
				fmt.Println()
				fmt.Println("To fix:")
				fmt.Println("1. Regenerate from templates/constitution.template.md (plaesy init) and fill every {{PLACEHOLDER}}")
				fmt.Println("2. Mark each §1 dimension row yes/no; the table is the source of the active scope")
				fmt.Println("3. Mirror the 'yes' rows into active_dimensions in the frontmatter")
				fmt.Println("4. Make the frontmatter version match the last amendment-log row, then re-run")
				return err
			}

			fmt.Println("[OK] constitutional validation passed")
			fmt.Printf("[✓] %s is internally consistent and fully filled\n", path)
			return nil
		},
	}
}

// newMemoryCmd ports scripts/bash/plaesy-validate-memory.sh.
func newMemoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "memory",
		Short: "Scan .plaesy/memory/ for external references and validate self-containment",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := common.GetRepoRoot()
			if err != nil {
				repoRoot = "."
			}

			fmt.Printf("[INFO] Scanning %s for external references...\n\n", ".plaesy/memory")

			res, err := validate.Memory(repoRoot)
			if err != nil {
				fmt.Println("[✗]", err)
				return err
			}

			for _, issue := range res.Issues {
				fmt.Printf("[✗] External references in: %s\n", issue.File)
				for _, pat := range issue.Patterns {
					fmt.Printf("  Lines with '%s':\n", pat)
					for _, line := range issue.Lines[pat] {
						fmt.Printf("    %s\n", line)
					}
				}
				fmt.Println()
			}

			fmt.Println()
			fmt.Println("[INFO] Scan complete:")
			fmt.Printf("  Files checked: %d\n", res.FilesChecked)
			fmt.Printf("  External refs found: %d\n", len(res.Issues))
			fmt.Println()

			if len(res.Issues) == 0 {
				fmt.Println("[✓] Self-containment validated ✓")
				fmt.Println("[✓] Memory is project-local and git-safe")
				return nil
			}

			fmt.Println("[!] Self-containment issues detected")
			fmt.Println()
			fmt.Println("To fix:")
			fmt.Println("1. For content in ~/.claude/projects: Copy it into .plaesy/memory/")
			fmt.Println("2. For external links: Replace with internal .plaesy/memory/ references (flat structure, no subfolders)")
			fmt.Println("3. Re-run this command after fixes")
			return fmt.Errorf("self-containment issues detected in %d file(s)", len(res.Issues))
		},
	}
}

// newAssumptionsCmd scans the corpus for `ASSUMED —` tags
// (instructions/plaesy.instructions.md rule 2 — the label an autonomous run
// writes instead of stopping to ask). It is an audit, not a gate, by default:
// it always lists what it found and exits 0. It becomes a gate only when
// `.plaesy/state.json`'s `assumptions_review_days` is a positive integer —
// the same opt-in convention as `autonomous_loop.checkpoint_interval` — in
// which case a tag whose file has had no commit in over that many days fails
// the run, surfacing it for a batched, asynchronous human review rather than
// blocking the autonomous run that wrote it in the first place.
func newAssumptionsCmd() *cobra.Command {
	var reviewDaysFlag int
	cmd := &cobra.Command{
		Use:   "assumptions",
		Short: "Audit ASSUMED — tags left by autonomous runs, optionally gated by age",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := common.GetRepoRoot()
			if err != nil {
				repoRoot = "."
			}

			reviewDays := reviewDaysFlag
			if reviewDays == 0 {
				reviewDays = readAssumptionsReviewDays(repoRoot)
			}

			fmt.Printf("[INFO] Scanning %s for ASSUMED — tags...\n\n", repoRoot)

			res, err := validate.Assumptions(repoRoot, reviewDays)
			if err != nil {
				fmt.Println("[✗]", err)
				return err
			}

			for _, t := range res.Tags {
				fmt.Printf("  %s:%d — %s\n", t.File, t.Line, t.Text)
			}

			fmt.Println()
			fmt.Println("[INFO] Scan complete:")
			fmt.Printf("  Files scanned: %d\n", res.FilesScanned)
			fmt.Printf("  ASSUMED tags found: %d\n", len(res.Tags))

			if reviewDays <= 0 {
				fmt.Println("  Review window: off (informational only — set " +
					"assumptions_review_days in .plaesy/state.json, or pass --review-days, to gate on age)")
				fmt.Println()
				fmt.Println("[✓] Audit complete (no gate configured)")
				return nil
			}

			fmt.Printf("  Review window: %d day(s)\n", reviewDays)
			fmt.Printf("  Stale (file uncommitted past window): %d\n\n", len(res.Stale))

			if len(res.Stale) == 0 {
				fmt.Println("[✓] No stale assumptions")
				return nil
			}

			fmt.Println("[!] Stale assumptions found:")
			for _, t := range res.Stale {
				fmt.Printf("  %s:%d — %s\n", t.File, t.Line, t.Text)
			}
			return fmt.Errorf("%d assumption(s) exceeded the %d-day review window", len(res.Stale), reviewDays)
		},
	}
	cmd.Flags().IntVar(&reviewDaysFlag, "review-days", 0,
		"override .plaesy/state.json's assumptions_review_days for this run (0 = use the config, or informational-only if unset)")
	return cmd
}

// readAssumptionsReviewDays reads `assumptions_review_days` from
// .plaesy/state.json. Any absence or parse failure is treated as "off" (0) —
// the same fail-open default as checkpoint_interval, since a config typo
// should never silently turn an informational audit into a blocking gate.
func readAssumptionsReviewDays(repoRoot string) int {
	data, err := os.ReadFile(filepath.Join(repoRoot, ".plaesy", "state.json"))
	if err != nil {
		return 0
	}
	var parsed struct {
		AssumptionsReviewDays int `json:"assumptions_review_days"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return 0
	}
	return parsed.AssumptionsReviewDays
}

// newDocxCmd ports scripts/bash/validate-docx.sh.
func newDocxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "docx <file.docx>",
		Short: "Validate a .docx file's OOXML structure",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			fmt.Printf("[INFO] Validating structure of %s\n", path)

			res, err := validate.Docx(path)
			if res != nil {
				fmt.Printf("[INFO] paragraphs: %d, tables: %d\n", res.Paragraphs, res.Tables)
			}
			if err != nil {
				fmt.Println("[✗]", err)
				return err
			}
			fmt.Println("[OK] structural validation passed")

			fmt.Println(validate.NoRenderCheckNotice())
			fmt.Printf("[✓] %s passed validation\n", path)
			return nil
		},
	}
}

// newPptxCmd ports scripts/bash/validate-pptx.sh.
func newPptxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pptx <file.pptx> [expected_slide_count]",
		Short: "Validate a .pptx file's OOXML structure",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			expected := 0
			if len(args) == 2 {
				if _, err := fmt.Sscanf(args[1], "%d", &expected); err != nil {
					return fmt.Errorf("invalid expected_slide_count %q: %w", args[1], err)
				}
			}

			fmt.Printf("[INFO] Validating structure of %s\n", path)

			res, err := validate.Pptx(path, expected)
			if res != nil {
				fmt.Printf("[INFO] slide count: %d\n", res.SlideCount)
				for _, n := range res.SlidesNoTitle {
					fmt.Printf("[WARN] slide %d has no title placeholder\n", n)
				}
			}
			if err != nil {
				fmt.Println("[✗]", err)
				return err
			}
			fmt.Println("[OK] structural validation passed")

			fmt.Println(validate.NoRenderCheckNotice())
			fmt.Printf("[✓] %s passed validation\n", path)
			return nil
		},
	}
}

// newXlsxCmd ports scripts/bash/validate-xlsx.sh.
func newXlsxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "xlsx <file.xlsx> [expected_sheet_name...]",
		Short: "Validate a .xlsx file's OOXML structure",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			expectedSheets := args[1:]

			fmt.Printf("[INFO] Validating structure of %s\n", path)

			res, err := validate.Xlsx(path, expectedSheets)
			if res != nil {
				fmt.Printf("[INFO] sheets: %v\n", res.Sheets)
			}
			if err != nil {
				fmt.Println("[✗]", err)
				return err
			}
			fmt.Println("[OK] structural validation passed")

			fmt.Println(validate.NoRenderCheckNotice())
			fmt.Printf("[✓] %s passed validation\n", path)
			return nil
		},
	}
}
