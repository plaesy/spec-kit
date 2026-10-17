package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// successContractBlock is the one pointer every phase prompt carries. It is kept
// identical in every prompt, so no prompt can drift to a weaker or different rule.
// To change the rule, change this constant and the same text in every prompt, then
// run plaesy reload so the embedded copies follow.
const successContractBlock = "## Success Contract\n\n" +
	"Before any work begins, after any routing or persona step above, write the contract, freeze it and commit it: rule in\n" +
	"`.plaesy/instructions/success-contract.md`, template in `.plaesy/templates/success-contract.template.md`, check with\n" +
	"`plaesy validate contract <file>`. At the end, compare every item with its result (Gate 0 and the Contract comparison\n" +
	"table in `.plaesy/instructions/quality-gates.md`). Do not report completion while a blocking item is unmet or unmeasured\n" +
	"(constitution NN-08).\n"

// Every phase prompt, including the per-dimension variants, carries the block exactly once.
func TestEveryPromptCarriesTheSuccessContractBlockOnce(t *testing.T) {
	root := corpusRoot(t)
	var checked int
	err := filepath.WalkDir(filepath.Join(root, "prompts"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		text := strings.ReplaceAll(string(raw), "\r\n", "\n")
		checked++
		if n := strings.Count(text, successContractBlock); n != 1 {
			rel, _ := filepath.Rel(root, path)
			t.Errorf("%s: the Success Contract block must appear exactly once, found %d (copy it verbatim from the reference block)", rel, n)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked < 70 {
		t.Fatalf("expected the full prompt corpus, checked only %d files", checked)
	}
}

// The rule is enforced where completion is decided: Gate 0 and the contract comparison.
func TestQualityGatesEnforceTheContract(t *testing.T) {
	root := corpusRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "instructions", "quality-gates.instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, want := range []string{
		"**Gate 0 — Success Contract (precondition, blocking).**",
		"### Contract comparison (required, part of Gate 0)",
		"0. Success Contract:",
		"plaesy validate contract <file> --gate",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("quality-gates must contain %q", want)
		}
	}
	// The numbered gates keep their numbers: other files cite them.
	for _, want := range []string{"1. **Build**", "2. **Tests**", "7. **Accessibility**"} {
		if !strings.Contains(text, want) {
			t.Errorf("quality-gates must keep the numbered gate %q", want)
		}
	}
}

// The global file carries only the pointer, not the spec. It must stay short and must link.
func TestPlaesyInstructionsPointsToTheContractWithoutPastingIt(t *testing.T) {
	root := corpusRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "instructions", "plaesy.instructions.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	start := strings.Index(text, "### Success Contract (Mandatory)")
	if start < 0 {
		t.Fatal("plaesy.instructions.md must carry the mandatory Success Contract pointer")
	}
	section := text[start:]
	// The section ends at the next heading of any level (## or ###).
	if end := strings.Index(section[1:], "\n#"); end >= 0 {
		section = section[:end+1]
	}
	if n := strings.Count(section, "\n"); n > 12 {
		t.Errorf("the global pointer is %d lines; keep it short and move the spec to success-contract.instructions.md", n)
	}
	if !strings.Contains(section, "success-contract.md") || !strings.Contains(section, "quality-gates.md") {
		t.Error("the global pointer must link the contract file and the gate")
	}
	if !strings.Contains(text, "| `success-contract.md` |") {
		t.Error("plaesy.instructions.md's scope-load table must have a trigger row for success-contract.md")
	}
}

// The contract spec must exist in the scope-load list, so it is installed and reachable.
func TestSuccessContractIsScopeLoaded(t *testing.T) {
	root := corpusRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "instructions", "mapping.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"success-contract.instructions.md"`) {
		t.Fatal("mapping.json scope_load must list success-contract.instructions.md")
	}
}

// Agents and checklists that decide completion point at the contract rule.
func TestCompletionRolesPointAtTheContract(t *testing.T) {
	root := corpusRoot(t)
	for _, rel := range []string{
		"agents/verifier.agents.md",
		"agents/critic.agents.md",
		"agents/qa.agents.md",
		"checklists/story-done.checklist.md",
		"checklists/story-draft.checklist.md",
	} {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), "success-contract.md") {
			t.Errorf("%s must reference the Success Contract rule", rel)
		}
	}
}
