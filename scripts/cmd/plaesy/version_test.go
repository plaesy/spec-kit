package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/common"
)

// `plaesy --version` prints the build version through the root command's
// Version field, which Cobra provides. There is deliberately no `version`
// subcommand: a second spelling for one fact is one more thing to keep in step.
func TestVersionFlagPrintsTheBuildVersion(t *testing.T) {
	var out bytes.Buffer
	root := rootForTest()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("plaesy --version: %v", err)
	}
	if !strings.Contains(out.String(), common.Version) {
		t.Errorf("plaesy --version printed %q, want it to contain %q", out.String(), common.Version)
	}
}
