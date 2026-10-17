package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/designlint"
	"github.com/spf13/cobra"
)

func newDesignCmd() *cobra.Command {
	var (
		evidence string
		strict   bool
	)
	cmd := &cobra.Command{
		Use:   "design <path...>",
		Short: "Lint produced frontend output against the design rules (offline, no browser)",
		Long: `Check produced HTML, CSS, JSX, TSX, Vue or Svelte files against the design rules
in instructions/frontend-design and instructions/redesign.

Deterministic checks, with no browser and no network:

  DL-IMG-HOST        pictures from placeholder hosts (picsum.photos and similar)      FAIL
  DL-CONTRAST        text colour pairs below WCAG 2.2 AA (4.5:1; 3:1 for large text)  FAIL
  DL-VIEWPORTS       with --evidence: a screenshot at 1440, 768 and 390 px            FAIL
  DL-INVENTED-FIGURE placeholder-looking figures in visible text (99.99%, $100.00)    WARN
  DL-DEFAULT-FONT    Inter, Arial, Roboto, system-ui... as the first family           WARN

Covered: CSS rules (including @media blocks), inline style="" attributes, and JSX
style={{ }} object literals. Not covered: utility classes (Tailwind and similar),
CSS-in-JS helpers, style={variable} and Vue :style bindings. Those are counted as
skipped, not checked.

Colours: #rgb, #rgba, #rrggbb, #rrggbbaa, rgb() and rgba() with 0-255 channels, the
keywords black, white and transparent, and var(--token) where the token is declared
in the run. A text colour with alpha is composited over its background. A token
declared with different values in different files is a conflict: it is not used,
and a WARN names it. Font sizes in px, rem, em, % and clamp() (the smallest bound)
are read; em and % are taken against a 16px root, and the finding says so.
Gradients, images, inherited colours and unreadable sizes are skipped and counted.

--evidence names a directory of screenshots. Each file name must contain the width
as its own number. Without it, viewport coverage is reported as INFO and not checked.

WARN findings do not fail the run unless --strict is given. The verdict is the same
for the same input on every run.`,
		Example: `  plaesy validate design dist/
  plaesy validate design src/ --evidence .plaesy/renders
  plaesy validate design index.html --strict`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := designlint.Lint(args, designlint.Options{EvidenceDir: evidence})
			if err != nil {
				fmt.Println("[✗]", err)
				return err
			}
			for _, f := range res.Findings {
				loc := f.File
				if f.Line > 0 {
					loc = fmt.Sprintf("%s:%d", f.File, f.Line)
				}
				fmt.Printf("%-4s %s  %s  %s\n", f.Severity, loc, f.Rule, f.Message)
			}
			fails, warns := countSeverity(res, designlint.SevFail), countSeverity(res, designlint.SevWarn)
			fmt.Printf("\n[INFO] %d file(s); contrast checked %d, skipped %d\n", res.Files, res.ContrastChecked, res.ContrastSkipped)
			fmt.Printf("%d FAIL, %d WARN\n", fails, warns)
			if fails > 0 || (strict && warns > 0) {
				// Cobra prints the error once, so the failure is not repeated here.
				return fmt.Errorf("design lint found %d FAIL and %d WARN (strict=%t)", fails, warns, strict)
			}
			fmt.Println("[OK] design lint passed")
			return nil
		},
	}
	cmd.Flags().StringVar(&evidence, "evidence", "", "directory of screenshots, one per required width")
	cmd.Flags().BoolVar(&strict, "strict", false, "treat WARN findings as failures")
	return cmd
}

func countSeverity(res designlint.Result, sev designlint.Severity) int {
	n := 0
	for _, f := range res.Findings {
		if f.Severity == sev {
			n++
		}
	}
	return n
}
