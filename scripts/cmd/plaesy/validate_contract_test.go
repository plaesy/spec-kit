package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/plaesy/spec-kit/internal/contract"
)

const contractFixtures = "../../../evals/fixtures/contracts/"

func runContract(t *testing.T, args ...string) error {
	t.Helper()
	cmd := newContractCmd()
	cmd.SetArgs(args)
	return quiet(t, cmd.Execute)
}

func TestValidateContractValidOpenPasses(t *testing.T) {
	if err := runContract(t, contractFixtures+"valid-open.contract.md"); err != nil {
		t.Fatalf("a valid frozen contract must pass the schema check: %v", err)
	}
}

// gateFromFixture commits the fixture's targets as an open, frozen contract, then writes the
// fixture's results on top, uncommitted, as the phase would at its gate. The freeze is then
// verifiable against git, which --gate requires.
func gateFromFixture(t *testing.T, fixture string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(contractFixtures, fixture))
	if err != nil {
		t.Fatal(err)
	}
	c, err := contract.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	dir := gitTemp(t)
	open := c
	open.Status = "open"
	open.Amendments, open.Waivers = nil, nil
	for i := range open.Items {
		open.Items[i].Result, open.Items[i].Actual, open.Items[i].Evidence, open.Items[i].Reason = "", "", "", ""
	}
	path := filepath.Join(dir, "fix-empty-slug.contract.md")
	if err := os.WriteFile(path, []byte(contractText(t, open)), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommit(t, dir, "freeze before work")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestValidateContractGatePassesWhenBlockingMet(t *testing.T) {
	path := gateFromFixture(t, "gate-pass.contract.md")
	if err := runContract(t, path, "--gate"); err != nil {
		t.Fatalf("a closed contract with every result met must pass the gate: %v", err)
	}
}

func TestValidateContractGateFailsWhenBlockingNotMet(t *testing.T) {
	path := gateFromFixture(t, "gate-fail.contract.md")
	if err := runContract(t, path, "--gate"); err == nil {
		t.Fatal("a blocking item not met must fail the gate")
	}
}

// Without git, the gate cannot verify the freeze, so it refuses, and says so.
func TestValidateContractGateRefusesAnUnverifiableFreeze(t *testing.T) {
	if err := runContract(t, contractFixtures+"gate-pass.contract.md", "--gate"); err == nil {
		t.Fatal("a gate run on a file outside git must refuse: the freeze is not verified")
	}
}

// contractText renders c as a contract file for the test fixtures.
func contractText(t *testing.T, c contract.Contract) string {
	t.Helper()
	body, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return "# Success Contract\n\n```json\n" + string(body) + "\n```\n"
}

func TestValidateContractRejectsVagueContract(t *testing.T) {
	if err := runContract(t, contractFixtures+"invalid-vague.contract.md"); err == nil {
		t.Fatal("vague words and non-measurable targets must fail")
	}
}

func TestValidateContractGateNeedsResults(t *testing.T) {
	// The valid-open fixture has no results: the schema check passes, the gate must not.
	if err := runContract(t, contractFixtures+"valid-open.contract.md", "--gate"); err == nil {
		t.Fatal("a gate run with no recorded results must fail")
	}
}

// gitTemp creates a repository in a temp dir for the freeze-anchor CLI tests.
func gitTemp(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH: the freeze anchor cannot be exercised here")
	}
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"}, {"config", "user.email", "test@example.invalid"}, {"config", "user.name", "test"},
		{"config", "commit.gpgsign", "false"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return dir
}

func gitCommit(t *testing.T, dir, msg string) {
	t.Helper()
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-q", "-m", msg}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
}

const contractHead = "# Success Contract\n\n```json\n"
const contractTail = "\n```\n"

// B-1 CLI reproduction: a blocking item removed after freeze, with a reason, must make
// `validate contract --gate` exit non-zero. The review saw this exit 0.
func TestValidateContractBlockingRemovalFailsGateFromTheCLI(t *testing.T) {
	dir := gitTemp(t)
	frozen := `{"id":"fix-x","phase":"fix","dimension":"technical","status":"open",
	"frozen":{"items":[
	  {"id":"C1","metric":"failing tests","target":"zero","method":"go test ./x/","blocking":true},
	  {"id":"C2","metric":"coverage","target":">= 90","method":"go test -cover ./x/","blocking":false}]},
	"items":[
	  {"id":"C1","kpi":"k","metric":"failing tests","target":"zero","method":"go test ./x/","blocking":true},
	  {"id":"C2","kpi":"k","metric":"coverage","target":">= 90","method":"go test -cover ./x/","blocking":false}],
	"amendments":[],"waivers":[]}`
	path := filepath.Join(dir, "fix-x.contract.md")
	if err := os.WriteFile(path, []byte(contractHead+frozen+contractTail), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommit(t, dir, "freeze")
	closed := `{"id":"fix-x","phase":"fix","dimension":"technical","status":"closed",
	"frozen":{"items":[
	  {"id":"C1","metric":"failing tests","target":"zero","method":"go test ./x/","blocking":true},
	  {"id":"C2","metric":"coverage","target":">= 90","method":"go test -cover ./x/","blocking":false}]},
	"items":[
	  {"id":"C2","kpi":"k","metric":"coverage","target":">= 90","method":"go test -cover ./x/","blocking":false,"result":"met","actual":"92.4","evidence":"cmd:go test -cover ./x/"}],
	"amendments":[{"item":"C1","field":"removed","reason":"scope cut"}],"waivers":[]}`
	if err := os.WriteFile(path, []byte(contractHead+closed+contractTail), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runContract(t, path, "--gate"); err == nil {
		t.Fatal("removing a blocking item with a reason must fail the gate from the CLI")
	}
}

// M-4 CLI reproduction: a closed contract with a blocking not_met must fail even without --gate.
func TestValidateContractClosedStatusFailsWithoutGateFlagFromTheCLI(t *testing.T) {
	dir := t.TempDir()
	closed := `{"id":"fix-y","phase":"fix","dimension":"technical","status":"closed",
	"frozen":{"items":[{"id":"C1","metric":"m","target":"zero","method":"go test ./y/","blocking":true}]},
	"items":[{"id":"C1","kpi":"k","metric":"m","target":"zero","method":"go test ./y/","blocking":true,"result":"not_met","evidence":"cmd:go test ./y/"}],
	"amendments":[],"waivers":[]}`
	path := filepath.Join(dir, "fix-y.contract.md")
	if err := os.WriteFile(path, []byte(contractHead+closed+contractTail), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runContract(t, path); err == nil {
		t.Fatal("a closed contract with a blocking not_met must fail even without --gate")
	}
}
