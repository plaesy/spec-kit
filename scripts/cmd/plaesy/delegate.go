package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
	"github.com/plaesy/spec-kit/internal/delegate"
	"github.com/spf13/cobra"
)

func init() { register(newDelegateCmd()) }

// analysisRel is the analysis directory, relative to the project, where the
// delegate usage log is kept.
var analysisRel = filepath.Join(".plaesy", "analysis")

// delegateOutput is the JSON printed for one call: the answer plus every
// attempt the chain made, so the caller can see which provider answered.
type delegateOutput struct {
	delegate.Response
	Attempts []delegate.Attempt `json:"attempts"`
}

// newDelegateCmd sends one bounded prompt to a configured provider, or to a
// named chain of providers. Which providers may run is set in .plaesy/config.json
// under delegate.
func newDelegateCmd() *cobra.Command {
	var repoPath, provider, chainName, prompt, out string
	var maxTokens int
	var timeout time.Duration
	var confirm, list bool

	cmd := &cobra.Command{
		Use:   "delegate",
		Short: "Send a prompt to a configured model provider or chain (cli, anthropic, ollama)",
		Long: `Send one bounded prompt to a provider listed in .plaesy/config.json under
delegate.providers, or to a named chain under delegate.chains. Only enabled
providers run. A billable provider (one that charges per call) also needs
--confirm, and a daily token budget is checked before its call.

  plaesy delegate --list
  plaesy delegate --provider claude-cli --prompt "summarise this diff"
  plaesy delegate --chain review --prompt "review this change"
  plaesy delegate --provider anthropic --prompt "..." --confirm
  plaesy delegate report

A chain falls through only on a transient error or a rate limit (after
Retry-After, up to 60s). Any other error stops the chain. The output is JSON
on stdout, or a file with --out. API keys are read from the environment
variable named in the provider entry and are never printed.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadDelegateConfig(repoPath)
			if err != nil {
				return err
			}
			if list {
				return printDelegateList(cfg)
			}
			if prompt == "" || (provider == "") == (chainName == "") {
				return fmt.Errorf("--prompt is required, and exactly one of --provider or --chain (use --list to see providers)")
			}
			names, err := chainNames(cfg, provider, chainName)
			if err != nil {
				return err
			}

			usagePath := delegate.UsagePath(filepath.Join(repoPath, analysisRel))
			recs, err := delegate.ReadUsage(usagePath)
			if err != nil {
				return err
			}
			resp, attempts, err := delegate.RunChain(cmd.Context(), cfg, names, delegate.Request{
				Prompt:          prompt,
				MaxOutputTokens: maxTokens,
				Timeout:         timeout,
			}, delegate.ChainOptions{
				Chain:   chainName,
				Confirm: confirm,
				Build:   delegate.EnvBuild(),
				Check: func(entry config.DelegateProviderConfig, req delegate.Request) error {
					return delegate.CheckBudget(entry, req, recs, time.Now())
				},
			})
			if err != nil {
				return err
			}
			if err := delegate.AppendUsage(usagePath, usageRecord(cfg, resp, chainName)); err != nil {
				return err
			}

			data, err := json.MarshalIndent(delegateOutput{Response: resp, Attempts: attempts}, "", "  ")
			if err != nil {
				return fmt.Errorf("encoding output: %w", err)
			}
			if out != "" {
				if err := os.WriteFile(out, append(data, '\n'), 0o644); err != nil {
					return fmt.Errorf("writing %s: %w", out, err)
				}
				return nil
			}
			fmt.Fprintln(os.Stdout, string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&repoPath, "path", ".", "project directory whose .plaesy/config.json is read")
	cmd.Flags().StringVar(&provider, "provider", "", "provider name from delegate.providers")
	cmd.Flags().StringVar(&chainName, "chain", "", "chain name from delegate.chains")
	cmd.Flags().StringVar(&prompt, "prompt", "", "prompt to send")
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 0, "max output tokens (0 = provider default; CLI providers ignore it)")
	cmd.Flags().DurationVar(&timeout, "timeout", 120*time.Second, "stop each call after this long")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "allow a billable provider to run")
	cmd.Flags().StringVar(&out, "out", "", "write the JSON result to this file instead of stdout")
	cmd.Flags().BoolVar(&list, "list", false, "list configured providers and whether each may run")

	cmd.AddCommand(newDelegateReportCmd())
	return cmd
}

// newDelegateReportCmd prints total usage per provider and per chain from the
// usage log.
func newDelegateReportCmd() *cobra.Command {
	var repoPath string
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Show delegate usage totals per provider and per chain",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			recs, err := delegate.ReadUsage(delegate.UsagePath(filepath.Join(repoPath, analysisRel)))
			if err != nil {
				return err
			}
			if len(recs) == 0 {
				fmt.Fprintln(os.Stdout, "no delegate calls recorded")
				return nil
			}
			byProvider, byChain := delegate.Summarize(recs)
			printTotals(os.Stdout, "PROVIDER", byProvider)
			fmt.Fprintln(os.Stdout)
			printTotals(os.Stdout, "CHAIN", byChain)
			return nil
		},
	}
	cmd.Flags().StringVar(&repoPath, "path", ".", "project directory whose usage log is read")
	return cmd
}

// chainNames returns the provider names to try: one name for --provider, or the
// ordered list for --chain.
func chainNames(cfg *config.DelegateConfig, provider, chainName string) ([]string, error) {
	if provider != "" {
		return []string{provider}, nil
	}
	names, ok := cfg.Chains[chainName]
	if !ok {
		known := make([]string, 0, len(cfg.Chains))
		for k := range cfg.Chains {
			known = append(known, k)
		}
		sort.Strings(known)
		return nil, fmt.Errorf("unknown chain %q (configured: %s)", chainName, strings.Join(known, ", "))
	}
	return names, nil
}

// usageRecord builds the log entry for an answered call. Billable is looked up
// from the config entry that answered.
func usageRecord(cfg *config.DelegateConfig, resp delegate.Response, chain string) delegate.UsageRecord {
	billable := false
	for _, p := range cfg.Providers {
		if p.Name == resp.Provider {
			billable = p.Billable
		}
	}
	return delegate.UsageRecord{
		At:           time.Now().UTC().Format(time.RFC3339),
		Provider:     resp.Provider,
		Chain:        chain,
		Billable:     billable,
		InputTokens:  resp.InputTokens,
		OutputTokens: resp.OutputTokens,
		WallMs:       resp.WallMillis,
	}
}

// printTotals writes one usage table. Calls without token counts are shown in
// their own column, so an unknown count is never read as zero.
func printTotals(w *os.File, label string, totals map[string]delegate.Totals) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "%s\tCALLS\tINPUT_TOKENS\tOUTPUT_TOKENS\tNO_TOKEN_COUNT\tWALL_MS\n", label)
	for _, k := range delegate.SortedKeys(totals) {
		t := totals[k]
		name := k
		if name == "" {
			name = "(direct)"
		}
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%d\n", name, t.Calls, t.InputTokens, t.OutputTokens, t.UnknownTokens, t.WallMs)
	}
	tw.Flush()
}

// loadDelegateConfig reads the delegate section of .plaesy/config.json under
// repoPath. A missing file or missing section falls back to the defaults.
func loadDelegateConfig(repoPath string) (*config.DelegateConfig, error) {
	path := filepath.Join(repoPath, config.UnifiedConfigPath())
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return config.DefaultDelegateConfig(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var uc config.UnifiedConfig
	if err := json.Unmarshal(data, &uc); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if uc.Delegate == nil {
		return config.DefaultDelegateConfig(), nil
	}
	return uc.Delegate, nil
}

// delegateStatus says whether a provider may run now. It is shared by --list and
// by plaesy doctor, so both report the same state.
func delegateStatus(p config.DelegateProviderConfig) string {
	switch {
	case !p.Enabled:
		return "disabled"
	case p.Billable:
		return "ready (needs --confirm)"
	default:
		return "ready"
	}
}

// printDelegateList shows each provider and whether a call would be allowed
// now. It prints names and flags only, never keys.
func printDelegateList(cfg *config.DelegateConfig) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tENABLED\tBILLABLE\tSTATUS")
	for _, p := range cfg.Providers {
		fmt.Fprintf(w, "%s\t%s\t%t\t%t\t%s\n", p.Name, p.Type, p.Enabled, p.Billable, delegateStatus(p))
	}
	return w.Flush()
}

// printDelegateStatus lists the configured providers for plaesy doctor. It is
// informational: a disabled provider is not a health failure.
func printDelegateStatus(cfg *config.DelegateConfig) {
	for _, p := range cfg.Providers {
		fmt.Printf("  %-14s %-10s %s\n", p.Name, p.Type, delegateStatus(p))
	}
}
