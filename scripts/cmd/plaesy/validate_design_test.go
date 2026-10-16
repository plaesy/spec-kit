package main

import (
	"os"
	"testing"
)

// quiet runs fn with stdout discarded, so command output does not flood go test.
func quiet(t *testing.T, fn func() error) error {
	t.Helper()
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer devnull.Close()
	stdout := os.Stdout
	os.Stdout = devnull
	defer func() { os.Stdout = stdout }()
	return fn()
}

func TestValidateDesignCleanFixturePasses(t *testing.T) {
	cmd := newDesignCmd()
	cmd.SetArgs([]string{"../../../evals/fixtures/frontend-clean"})
	if err := quiet(t, cmd.Execute); err != nil {
		t.Fatalf("clean fixture must pass design lint: %v", err)
	}
}

func TestValidateDesignPlaceholderFixtureFails(t *testing.T) {
	cmd := newDesignCmd()
	cmd.SetArgs([]string{"../../../evals/fixtures/frontend-placeholder-image"})
	if err := quiet(t, cmd.Execute); err == nil {
		t.Fatal("placeholder-host fixture must fail design lint")
	}
}

func TestValidateDesignWarnPassesUnlessStrict(t *testing.T) {
	warnOnly := "../../../evals/fixtures/frontend-invented-figures"

	cmd := newDesignCmd()
	cmd.SetArgs([]string{warnOnly})
	if err := quiet(t, cmd.Execute); err != nil {
		t.Fatalf("WARN-only fixture must pass without --strict: %v", err)
	}

	strict := newDesignCmd()
	strict.SetArgs([]string{warnOnly, "--strict"})
	if err := quiet(t, strict.Execute); err == nil {
		t.Fatal("WARN-only fixture must fail with --strict")
	}
}

func TestValidateEvalsRepositoryPasses(t *testing.T) {
	cmd := newEvalsCmd()
	cmd.SetArgs([]string{})
	if err := quiet(t, cmd.Execute); err != nil {
		t.Fatalf("repository evals must pass: %v", err)
	}
}
