package main

import (
	"strings"
	"testing"
)

// TestResourceValidateRuns checks that every validator is reachable as plaesy <target>
// validate, the resource-first form.
func TestResourceValidateRuns(t *testing.T) {
	root := rootForTest()
	for _, target := range validateTargets {
		cmd, _, err := root.Find([]string{target.name, "validate"})
		if err != nil || cmd == nil || cmd.Name() != "validate" {
			t.Errorf("plaesy %s validate is not a command: %v", target.name, err)
			continue
		}
		if cmd.Parent() == nil || cmd.Parent().Name() != target.name {
			t.Errorf("validate under %q, want %q", cmd.Parent().Name(), target.name)
		}
	}
}

// TestLegacyValidateNamesNewLocation checks that the old spellings still resolve and
// that each one is deprecated with the resource-first name it moved to.
func TestLegacyValidateNamesNewLocation(t *testing.T) {
	root := rootForTest()
	for _, target := range validateTargets {
		old, _, err := root.Find([]string{"validate", target.name})
		if err != nil || old == nil || old.Name() != target.name {
			t.Errorf("plaesy validate %s no longer resolves: %v", target.name, err)
			continue
		}
		want := "plaesy " + target.name + " validate"
		if !strings.Contains(old.Deprecated, want) {
			t.Errorf("plaesy validate %s deprecation = %q, want it to name %q", target.name, old.Deprecated, want)
		}
		if !target.legacy {
			continue
		}
		pre, _, err := root.Find([]string{"validate-" + target.name})
		if err != nil || pre == nil || pre.Name() != "validate-"+target.name {
			t.Errorf("validate-%s no longer resolves: %v", target.name, err)
			continue
		}
		if !strings.Contains(pre.Deprecated, want) {
			t.Errorf("validate-%s deprecation = %q, want it to name %q", target.name, pre.Deprecated, want)
		}
	}
}
