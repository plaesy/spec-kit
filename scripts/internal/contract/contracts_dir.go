package contract

import (
	"os"
	"path/filepath"
	"sort"
)

// Dir is where a repository keeps its contracts, relative to the repository root.
const Dir = ".plaesy/contracts"

// DirResult is the outcome of checking one contract file found under Dir.
type DirResult struct {
	Report Report
	Gate   bool // the file is closed, so the gate is checked
}

// CheckDir checks every *.contract.md under root/.plaesy/contracts. A closed contract is
// checked against the gate; an open one against its schema and freeze. It returns nil when
// there are no contracts, so a repository without any is not an error.
func CheckDir(root string) ([]DirResult, error) {
	paths, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(Dir), "*.contract.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	var out []DirResult
	for _, p := range paths {
		gate := false
		if c, err := readContract(p); err == nil && c.Status == "closed" {
			gate = true
		}
		rep, err := Check(p, Options{Gate: gate, RepoRoot: root})
		if err != nil {
			return out, err
		}
		out = append(out, DirResult{Report: rep, Gate: gate})
	}
	return out, nil
}

// readContract parses the contract file at path. It is used to read the status before the check.
func readContract(path string) (Contract, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, err
	}
	return Parse(raw)
}
