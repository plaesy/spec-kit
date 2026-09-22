package main

import (
	"fmt"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/validate"
	"github.com/spf13/cobra"
)

func init() {
	register(newValidateMemoryCmd())
	register(newValidateDocxCmd())
	register(newValidatePptxCmd())
	register(newValidateXlsxCmd())
}

// newValidateMemoryCmd ports scripts/bash/plaesy-validate-memory.sh.
func newValidateMemoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate-memory",
		Short: "Scan .plaesy/memory/ for external references and validate self-containment",
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
			fmt.Println("3. Re-run this script after fixes")
			return fmt.Errorf("self-containment issues detected in %d file(s)", len(res.Issues))
		},
	}
}

// newValidateDocxCmd ports scripts/bash/validate-docx.sh.
func newValidateDocxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate-docx <file.docx>",
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

// newValidatePptxCmd ports scripts/bash/validate-pptx.sh.
func newValidatePptxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate-pptx <file.pptx> [expected_slide_count]",
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

// newValidateXlsxCmd ports scripts/bash/validate-xlsx.sh.
func newValidateXlsxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate-xlsx <file.xlsx> [expected_sheet_name...]",
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
