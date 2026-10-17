package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/config"
)

const testKey = "sk-test-SECRET-value-123"

func TestDoctor_DelegateProvidersNoKeys(t *testing.T) {
	// Create a config with an anthropic provider that has an API key
	cfg := &config.DelegateConfig{
		Providers: []config.DelegateProviderConfig{
			{Name: "claude-cli", Type: "cli", Enabled: true, Billable: false, Command: "claude", Args: []string{"-p"}},
			{Name: "kilo-cli", Type: "cli", Enabled: false, Billable: false, Command: "kilo", Args: []string{"run"}},
			{Name: "anthropic", Type: "anthropic", Enabled: true, Billable: true, Model: "claude-sonnet-5-5", APIKeyEnv: "ANTHROPIC_API_KEY"},
			{Name: "ollama", Type: "ollama", Enabled: false, Billable: false, Model: "llama3.2", Endpoint: "http://localhost:11434"},
		},
	}

	// Capture output using os.Pipe
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdout := os.Stdout
	os.Stdout = w
	printDelegateStatus(cfg)
	w.Close()
	os.Stdout = oldStdout

	var buf strings.Builder
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	// Check that all providers are listed
	if !strings.Contains(output, "claude-cli") {
		t.Errorf("output missing claude-cli: %s", output)
	}
	if !strings.Contains(output, "kilo-cli") {
		t.Errorf("output missing kilo-cli: %s", output)
	}
	if !strings.Contains(output, "anthropic") {
		t.Errorf("output missing anthropic: %s", output)
	}
	if !strings.Contains(output, "ollama") {
		t.Errorf("output missing ollama: %s", output)
	}

	// Check that statuses are correct
	if !strings.Contains(output, "cli") {
		t.Errorf("output missing cli type: %s", output)
	}
	if !strings.Contains(output, "anthropic") {
		t.Errorf("output missing anthropic type: %s", output)
	}
	if !strings.Contains(output, "ollama") {
		t.Errorf("output missing ollama type: %s", output)
	}

	// CRITICAL: Check that no API key is printed
	if strings.Contains(output, testKey) {
		t.Errorf("API key leaked in doctor output: %s", output)
	}
	if strings.Contains(output, "SECRET") {
		t.Errorf("SECRET leaked in doctor output: %s", output)
	}

	// Check that the key env var name is not printed as a value
	if strings.Contains(output, "ANTHROPIC_API_KEY") {
		t.Errorf("API key env var name leaked in doctor output: %s", output)
	}

	t.Logf("Doctor delegate output:\n%s", output)
}

func TestDoctor_DelegateProvidersDisabled(t *testing.T) {
	cfg := &config.DelegateConfig{
		Providers: []config.DelegateProviderConfig{
			{Name: "disabled-provider", Type: "cli", Enabled: false, Billable: false, Command: "foo"},
		},
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdout := os.Stdout
	os.Stdout = w
	printDelegateStatus(cfg)
	w.Close()
	os.Stdout = oldStdout

	var buf strings.Builder
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if !strings.Contains(output, "disabled-provider") {
		t.Errorf("output missing disabled-provider: %s", output)
	}
	if !strings.Contains(output, "disabled") {
		t.Errorf("output missing disabled status: %s", output)
	}

	t.Logf("Doctor delegate output (disabled):\n%s", output)
}

func TestDoctor_DelegateProvidersBillableNeedsConfirm(t *testing.T) {
	cfg := &config.DelegateConfig{
		Providers: []config.DelegateProviderConfig{
			{Name: "paid", Type: "anthropic", Enabled: true, Billable: true, Model: "m"},
		},
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdout := os.Stdout
	os.Stdout = w
	printDelegateStatus(cfg)
	w.Close()
	os.Stdout = oldStdout

	var buf strings.Builder
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if !strings.Contains(output, "paid") {
		t.Errorf("output missing paid: %s", output)
	}
	if !strings.Contains(output, "needs --confirm") {
		t.Errorf("output missing needs --confirm: %s", output)
	}

	t.Logf("Doctor delegate output (billable):\n%s", output)
}