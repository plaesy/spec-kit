package tasks

import (
	"os"
	"path/filepath"
)

// ContractFileName is the checklist a task folder carries, next to its spec.md.
const ContractFileName = "contracts.json"

// ContractRef is one task folder that carries a contracts.json.
type ContractRef struct {
	Spec   string
	Status string
	Task   string
	Path   string
}

// ListContracts returns every task folder under tasksDir that holds a contracts.json,
// ordered by status in Statuses order, then by name. A non-empty spec keeps only the
// tasks of that spec.
func ListContracts(tasksDir, spec string) ([]ContractRef, error) {
	var refs []ContractRef
	for _, status := range Statuses {
		names, _, err := ListTasks(tasksDir, status)
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			path := filepath.Join(tasksDir, status, name, ContractFileName)
			if _, err := os.Stat(path); err != nil {
				continue
			}
			taskSpec := SpecOf(tasksDir, status, name)
			if spec != "" && taskSpec != spec {
				continue
			}
			refs = append(refs, ContractRef{Spec: taskSpec, Status: status, Task: name, Path: path})
		}
	}
	return refs, nil
}
