package config

// DelegateConfig selects which providers `plaesy delegate` may call. Only
// providers with Enabled set run. A provider with Billable set (it charges per
// call) also needs --confirm on each run.
type DelegateConfig struct {
	Providers []DelegateProviderConfig `json:"providers"`
	// Chains names an ordered list of providers to try. A call falls through a
	// chain only on a classified transient or rate-limit error.
	Chains map[string][]string `json:"chains,omitempty"`
}

// DelegateProviderConfig describes one provider entry.
//
// Type selects the implementation: "cli" (runs Command with the prompt on
// stdin), "anthropic" (Messages API over HTTP), or "ollama" (local HTTP).
// Unused fields for a type are ignored.
type DelegateProviderConfig struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Enabled   bool     `json:"enabled"`
	Billable  bool     `json:"billable"`
	Command   string   `json:"command,omitempty"`     // cli: executable name or path
	Args      []string `json:"args,omitempty"`        // cli: arguments before the prompt is read from stdin
	Model     string   `json:"model,omitempty"`       // anthropic, ollama
	Endpoint  string   `json:"endpoint,omitempty"`    // anthropic, ollama: base URL
	APIKeyEnv string   `json:"api_key_env,omitempty"` // anthropic: environment variable that holds the key
	// DailyTokenBudget caps input plus output tokens per UTC day for a billable
	// provider. 0 means no budget is configured: a billable call still needs --confirm.
	DailyTokenBudget int `json:"daily_token_budget,omitempty"`
}

// DefaultDelegateConfig returns the built-in provider list. The defaults are
// ASSUMED values: CLIs run on the user's own subscription (not billable), the
// Anthropic API is billable, and kilo and the others start disabled until the
// user turns them on.
func DefaultDelegateConfig() *DelegateConfig {
	return &DelegateConfig{
		Providers: []DelegateProviderConfig{
			{Name: "claude-cli", Type: "cli", Enabled: true, Billable: false, Command: "claude", Args: []string{"-p"}},
			{Name: "kilo-cli", Type: "cli", Enabled: false, Billable: false, Command: "kilo", Args: []string{"run"}},
			{Name: "anthropic", Type: "anthropic", Enabled: false, Billable: true, Model: "claude-sonnet-5-5", APIKeyEnv: "ANTHROPIC_API_KEY"},
			{Name: "ollama", Type: "ollama", Enabled: false, Billable: false, Model: "llama3.2", Endpoint: "http://localhost:11434"},
		},
	}
}
