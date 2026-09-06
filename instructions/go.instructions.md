---
description: 'Instructions for writing Go code following idiomatic Go practices and community standards'
applyTo: '**/*.go,**/go.mod,**/go.sum'
---

# Go Development Instructions

Follow idiomatic Go: [Effective Go](https://go.dev/doc/effective_go), [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), [Google's Go Style Guide](https://google.github.io/styleguide/go/).

## General
- Simple, clear, idiomatic code - clarity over cleverness, least surprise
- Left-aligned happy path (minimize indentation), return early to reduce nesting
- Make the zero value useful
- Document exported types/functions/methods/packages
- Use Go modules for dependency management

## Naming
- **Packages**: lowercase, single-word, no underscores/hyphens/mixedCaps; name for what it provides, not contains; avoid generic names (`util`, `common`, `base`); singular not plural
- **Variables/functions**: mixedCaps/MixedCaps, short but descriptive; single-letter only for very short scopes (loop indices); exported = capital first letter; avoid stuttering (`http.Server` not `http.HTTPServer`)
- **Interfaces**: `-er` suffix where possible (`Reader`, `Writer`); single-method interfaces named after the method; keep small and focused
- **Constants**: MixedCaps exported / mixedCaps unexported; group in `const` blocks; consider typed constants for type safety

## Code Style and Formatting
- **Formatting**: `gofmt` always, `goimports` for imports, no hard line-length limit but keep readable, blank lines between logical groups
- **Comments**: complete sentences starting with the name of the thing described; package comments start with "Package [name]"; `//` for most, `/* */` sparingly (mainly package docs); document why, not what, unless the what is complex
- **Error handling**: check immediately after the call; don't discard with `_` without documenting why; wrap with `fmt.Errorf`+`%w`; custom error types for specific-error checks; error as last return value, named `err`; lowercase message, no trailing punctuation

## Architecture and Project Structure
- **Package organization**: standard Go project layout; `main` packages in `cmd/`; reusable packages in `pkg/` or `internal/`; `internal/` for non-exportable packages; group related functionality; avoid circular deps
- **Dependency management**: Go modules (`go.mod`/`go.sum`); minimal deps, updated regularly for security; `go mod tidy`; vendor only when necessary

## Type Safety and Language Features
- **Types**: define for meaning/type safety; struct tags for JSON/XML/DB mapping; explicit conversions; check the second return value on type assertions
- **Pointers vs values**: pointers for large structs or when modifying the receiver; values for small structs/immutability; be consistent within a type's method set; consider the zero value when choosing
- **Interfaces/composition**: accept interfaces, return concrete types; keep interfaces small (1-3 methods); use embedding for composition; define interfaces where used, not where implemented; don't export unless necessary

## Concurrency
- **Goroutines**: don't spawn in libraries - let the caller control concurrency; always know how it exits; `sync.WaitGroup`/channels to wait; avoid leaks via cleanup
- **Channels**: for goroutine communication - share memory by communicating, not the reverse; close from the sender side; buffered when capacity is known; `select` for non-blocking ops
- **Synchronization**: `sync.Mutex` for shared state (keep critical sections small); `sync.RWMutex` for many readers; prefer channels over mutexes when possible; `sync.Once` for one-time init

## Error Handling Patterns
- **Creating**: `errors.New` for static, `fmt.Errorf` for dynamic, custom types for domain-specific, exported vars for sentinel errors, `errors.Is`/`errors.As` for checking
- **Propagation**: add context going up the stack; don't both log and return (choose one); handle at the appropriate level; consider structured errors for debugging

## API Design
- **HTTP handlers**: `http.HandlerFunc` for simple, `http.Handler` for stateful; middleware for cross-cutting concerns; correct status codes/headers; graceful error handling
- **JSON APIs**: struct tags for marshaling control; validate input; pointers for optional fields; `json.RawMessage` for delayed parsing; handle JSON errors

## Performance
- **Memory**: minimize hot-path allocations; reuse objects (`sync.Pool`); value receivers for small structs; preallocate slices when size known; avoid unnecessary string conversions
- **Profiling**: `pprof`; benchmark critical paths; profile before optimizing; algorithmic improvements first; `testing.B` for benchmarks

## Testing
- **Organization**: same-package (white-box) tests; `_test` package suffix for black-box; `_test.go` filename suffix, next to the code tested
- **Writing**: table-driven for multiple cases; `Test_functionName_scenario` naming; `t.Run` subtests; test success and error cases; use `testify` sparingly
- **Helpers**: `t.Helper()` on helper functions; fixtures for complex setup; `testing.TB` for functions shared by tests/benchmarks; `t.Cleanup()` for resource cleanup

## Security
- **Input validation**: validate all external input; strong typing to prevent invalid states; sanitize before SQL; careful with user-input file paths; escape per context (HTML/SQL/shell)
- **Cryptography**: stdlib crypto packages only, never roll your own; `crypto/rand` for randomness; bcrypt (or similar) for passwords; TLS for network communication

## Documentation
- **Code**: document all exported symbols, starting with the symbol name; examples where helpful; keep close to code, update when code changes
- **README**: clear setup instructions, dependencies/requirements, usage examples, config options, troubleshooting section

## Tools and Workflow
- **Essential tools**: `go fmt`, `go vet`, `golangci-lint`, `go test`, `go mod`, `go generate`
- **Practices**: tests before commit; pre-commit hooks for format/lint; focused/atomic commits; meaningful messages; review diffs before committing

## Common Pitfalls to Avoid
Not checking errors; ignoring race conditions; goroutine leaks; skipping `defer` for cleanup; concurrent map writes; confusing nil interfaces with nil pointers; unclosed resources (files, connections); unnecessary globals; over-using `interface{}`; ignoring a type's zero value.