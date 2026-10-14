// Package quality holds the repository's own quality gates: the checks that run
// inside `go test ./...` and therefore inside the only CI job. It has no
// non-test source on purpose — a gate that ships as a library would tempt
// someone to make it optional.
package quality

const (
	// BaselineName is the file, next to scripts/go.mod, that records the
	// measured per-package coverage every package must stay at or above.
	BaselineName = "coverage-baseline.json"

	// ModulePath is this module, as declared in go.mod.
	ModulePath = "github.com/plaesy/spec-kit"

	// SelfPackage is this package. It is excluded from the ratchet: measuring
	// it would mean the ratchet runs itself, and `go test` inside `go test`
	// recurses until the machine gives up.
	SelfPackage = ModulePath + "/internal/quality"

	// measureTimeout bounds one package's test run. The per-package runs
	// include the repository's slower suites, so this is deliberately far
	// larger than any single suite.
	measureTimeout = 20 // minutes
)
