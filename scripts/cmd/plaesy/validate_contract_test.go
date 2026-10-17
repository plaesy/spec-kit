package main

import "testing"

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

func TestValidateContractGatePassesWhenBlockingMet(t *testing.T) {
	if err := runContract(t, contractFixtures+"gate-pass.contract.md", "--gate"); err != nil {
		t.Fatalf("a closed contract with every result met must pass the gate: %v", err)
	}
}

func TestValidateContractGateFailsWhenBlockingNotMet(t *testing.T) {
	if err := runContract(t, contractFixtures+"gate-fail.contract.md", "--gate"); err == nil {
		t.Fatal("a blocking item not met must fail the gate")
	}
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
