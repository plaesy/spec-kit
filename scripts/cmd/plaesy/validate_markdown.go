package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/mdlint"
	"github.com/spf13/cobra"
)

func newMarkdownCmd() *cobra.Command {
	var (
		configPath    string
		baselinePath  string
		noBaseline    bool
		summaryOnly   bool
		listRules     bool
		fix           bool
		writeBaseline bool
		worst         int
	)
	cmd := &cobra.Command{
		Use:   "markdown [paths...]",
		Short: "Lint Markdown with the built-in Go linter (no Node, no npm)",
		Long: `Lint Markdown using the same binary that installs the framework.

Implements the markdownlint rule set on top of goldmark with the same rule IDs and
config keys, plus the three GitHub accessibility rules (GHA001-GHA003) that the
npm-only ruleset provided. Node and npm are not required — for this repository or
for a project that installed it.

With no PATH, every .md file under the repository root is scanned, skipping
` + strings.Join(mdlint.DefaultIgnoreDirs, ", ") + `.

By default the run is compared against a ratchet baseline
(.markdownlint-baseline.json): the command fails when the violation total rises
above the recorded ceiling, so pre-existing violations do not block unrelated work
while new violations do. Pass --no-baseline to require zero violations.

Unknown rule names and unknown option keys in the config are hard errors, not
warnings: a config that half-loads is a check that reports success while checking
nothing.`,
		Example: `  plaesy validate markdown
  plaesy validate markdown docs/ --summary
  plaesy validate markdown --no-baseline          # require zero
  plaesy validate markdown --fix                  # apply machine fixable rules
  plaesy validate markdown --list-rules           # what is implemented
  plaesy validate markdown --write-baseline       # re-measure the ratchet`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if listRules {
				printRuleList()
				return nil
			}

			root, err := common.GetRepoRoot()
			if err != nil {
				root = "."
			}
			if configPath == "" {
				configPath = filepath.Join(root, mdlint.DefaultConfigName)
			}
			cfg, err := mdlint.LoadConfig(configPath)
			if err != nil {
				fmt.Println("[✗]", err)
				return err
			}

			paths := args
			if len(paths) == 0 {
				paths = []string{root}
			}

			if fix {
				changed, fixErr := mdlint.Fix(root, paths, cfg, nil)
				if fixErr != nil {
					fmt.Println("[✗]", fixErr)
					return fixErr
				}
				if !changed {
					fmt.Println("[OK] nothing to fix")
					return nil
				}
				fmt.Println("[✓] applied machine-fixable rules (MD009, MD010, MD012, MD022, MD031, MD032, MD047, MD058)")
			}

			res, err := mdlint.Lint(root, paths, cfg, nil)
			if err != nil {
				fmt.Println("[✗]", err)
				return err
			}

			fmt.Printf("[INFO] scanned %d markdown file(s), %d with violations\n\n", res.Scanned, res.Files)

			if baselinePath == "" {
				baselinePath = filepath.Join(root, mdlint.DefaultBaselineName)
			}
			if writeBaseline {
				before := res.Total
				if err := mdlint.WriteBaseline(baselinePath, res); err != nil {
					fmt.Println("[✗]", err)
					return err
				}
				fmt.Printf("[✓] recorded a new ceiling of %d violation(s) in %s\n", before, baselinePath)
				fmt.Println("Only re-measure when the rise is understood: a ratchet that is re-set on every")
				fmt.Println("red build is not a ratchet.")
				return nil
			}
			var baseline *mdlint.Baseline
			if !noBaseline {
				baseline, err = mdlint.LoadBaseline(baselinePath)
				if err != nil {
					fmt.Println("[✗]", err)
					return err
				}
			}

			if res.Total == 0 {
				fmt.Println("[OK] no markdown violations")
				return nil
			}

			if !summaryOnly {
				// 0 means "all", as the flag help says, and the loop below
				// already treats a non-positive limit as unbounded. Clamping
				// <= 0 to 20 here made that contract unreachable, so a full
				// report was silently truncated to the default 20 files while
				// the histogram still counted every file — leaving violations
				// counted but not locatable. Negative is not a documented
				// value; treat it as "all" rather than inventing a cap.
				limit := worst
				if limit < 0 {
					limit = 0
				}
				shown := 0
				for _, f := range res.Violations {
					if limit > 0 && shown >= limit {
						break
					}
					fmt.Printf("%s (%d)\n", f.Path, f.Total)
					for _, v := range f.Detail {
						fmt.Printf("  %s:%d:%d %s — %s\n", f.Path, v.Line, v.Column, v.Rule, v.Message)
					}
					shown++
				}
				if remaining := len(res.Violations) - shown; remaining > 0 {
					fmt.Printf("\n… and %d more file(s) (see --summary)\n", remaining)
				}
				fmt.Println()
			}

			fmt.Printf("Total: %d violation(s)\n", res.Total)
			fmt.Printf("By rule:%s\n", res.Histogram())

			if res.Exceeds(baseline) {
				fmt.Println()
				if baseline == nil {
					fmt.Printf("[✗] %d violation(s) and --no-baseline requires zero\n", res.Total)
					return fmt.Errorf("markdown lint found %d violation(s) with --no-baseline", res.Total)
				}
				fmt.Printf("[✗] %d violations exceeds the baseline ceiling of %d (%s)\n", res.Total, baseline.MaxViolations, filepath.Base(baselinePath))
				fmt.Println("Either fix them (plaesy validate-markdown --fix, then by hand), or record the new")
				fmt.Println("ceiling in the baseline file — never delete the baseline to make a run pass.")
				return fmt.Errorf("markdown lint ratchet broken: %d > %d", res.Total, baseline.MaxViolations)
			}
			if baseline != nil {
				headroom := baseline.MaxViolations - res.Total
				switch {
				case headroom == 0:
					fmt.Printf("\n[OK] exactly at the ratchet ceiling of %d — any new violation fails. "+
						"Lower it with --write-baseline as you fix things. Measured %s\n",
						baseline.MaxViolations, baseline.MeasuredAt)
				case headroom < 10:
					fmt.Printf("\n[OK] within the ratchet ceiling of %d (%d to spare) — measured %s\n",
						baseline.MaxViolations, headroom, baseline.MeasuredAt)
				default:
					fmt.Printf("\n[OK] within the ratchet ceiling of %d (%d headroom) — measured %s\n",
						baseline.MaxViolations, headroom, baseline.MeasuredAt)
				}
			} else {
				fmt.Println()
				fmt.Println("[OK] no baseline configured, so this run required zero violations")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&configPath, "config", "", "config file (default: <repo>/"+mdlint.DefaultConfigName+")")
	cmd.Flags().StringVar(&baselinePath, "baseline", "", "ratchet baseline file (default: <repo>/"+mdlint.DefaultBaselineName+")")
	cmd.Flags().BoolVar(&noBaseline, "no-baseline", false, "ignore the baseline and require zero violations")
	cmd.Flags().BoolVar(&summaryOnly, "summary", false, "print only the per-rule histogram")
	cmd.Flags().BoolVar(&listRules, "list-rules", false, "list the implemented rules and exit")
	cmd.Flags().BoolVar(&fix, "fix", false, "apply the machine-fixable rules in place")
	cmd.Flags().BoolVar(&writeBaseline, "write-baseline", false, "re-measure and record the current total as the new ceiling")
	cmd.Flags().IntVar(&worst, "max-files", 20, "how many offending files to detail (0 = all)")
	return cmd
}

func printRuleList() {
	ids := mdlint.RuleIDs()
	sort.Strings(ids)
	fmt.Printf("Implemented markdownlint rules (%d):\n\n", len(ids))
	for _, id := range ids {
		fmt.Printf("  %-8s %s\n", id, mdlint.RuleDescription(id))
	}
	fmt.Println()
	fmt.Println("Not implemented (markdownlint has more): MD005, MD006, MD008, MD011, MD014, MD018,")
	fmt.Println("MD019, MD020, MD021, MD023, MD027, MD028, MD030, MD033, MD034, MD035, MD037, MD038,")
	fmt.Println("MD039, MD042, MD044, MD045, MD049, MD050, MD051, MD052, MD053, MD054, MD055, MD056,")
	fmt.Println("MD059, MD060.")
	fmt.Println()
	fmt.Println("A clean run means \"no implemented rule fired\", not \"markdownlint would be happy\".")
	if _, err := os.Stat(mdlint.DefaultConfigName); err == nil {
		fmt.Printf("\nConfig in use: %s\n", mdlint.DefaultConfigName)
	}
}
