package contract

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckDirSkipsWhenThereAreNoContracts(t *testing.T) {
	res, err := CheckDir(t.TempDir())
	if err != nil || len(res) != 0 {
		t.Fatalf("no .plaesy/contracts must be an empty result, not an error: %v %+v", err, res)
	}
}

func TestCheckDirGatesClosedContractsOnly(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, filepath.FromSlash(Dir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	open := freshFrozen()
	closed := base()
	closed.ID = "closed-one"
	closed.Status = "closed"
	closed.Items[0].Result, closed.Items[0].Actual, closed.Items[0].Evidence = ResultNotMet, "1", "cmd:go test ./x/"
	closed.Items[1].Result, closed.Items[1].Actual, closed.Items[1].Evidence = ResultMet, "92", "cmd:go test -cover ./x/"
	writeAt(t, filepath.Join(dir, "a-open.contract.md"), open)
	writeAt(t, filepath.Join(dir, "b-closed.contract.md"), closed)

	res, err := CheckDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("want 2 contracts checked, got %d", len(res))
	}
	if res[0].Gate || !res[1].Gate {
		t.Fatalf("only the closed contract must be gated: %+v", res)
	}
	if res[0].Report.Errors() != 0 {
		t.Fatalf("the open contract must pass its schema check: %+v", res[0].Report.Issues)
	}
	if res[1].Report.Errors() == 0 {
		t.Fatalf("the closed contract with a blocking not_met must fail at the gate: %+v", res[1].Report.Issues)
	}
}

func writeAt(t *testing.T, path string, c Contract) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contractText(t, c)), 0o644); err != nil {
		t.Fatal(err)
	}
}
