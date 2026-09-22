package main

import (
	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/imagegen"
	"github.com/spf13/cobra"
)

func init() { register(newGenerateImageCmd()) }

func newGenerateImageCmd() *cobra.Command {
	var prompt, provider, size, out string

	cmd := &cobra.Command{
		Use:   "generate-image",
		Short: "Generate an image asset via a configured provider API",
		RunE: func(cmd *cobra.Command, args []string) error {
			if prompt == "" || out == "" {
				return cmd.Usage()
			}

			path, err := imagegen.Generate(imagegen.Options{
				Prompt:   prompt,
				Provider: provider,
				Size:     size,
				Out:      out,
			})
			if err != nil {
				common.LogError("%s", err)
				return err
			}

			cmd.Println(path)
			return nil
		},
	}

	cmd.Flags().StringVar(&prompt, "prompt", "", "Image prompt text (required)")
	cmd.Flags().StringVar(&provider, "provider", "", "Image provider: openai (default) or gemini")
	cmd.Flags().StringVar(&size, "size", "", "Image size, e.g. 1024x1024")
	cmd.Flags().StringVar(&out, "out", "", "Output file path (required)")

	return cmd
}
