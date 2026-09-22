---
description: 'Instructions for writing Rust code following idiomatic Rust practices and community standards'
applyTo: '**/*.rs,**/Cargo.toml,**/Cargo.lock'
---

# Rust Development Instructions

Follow idiomatic Rust: [The Rust Book](https://doc.rust-lang.org/book/), [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/), [Rust Style Guide](https://doc.rust-lang.org/nightly/style-guide/).

## General
- Simple, clear, idiomatic code - clarity over cleverness, let the type system do the work
- Make illegal states unrepresentable via types; prefer compile-time guarantees over runtime checks
- Small, focused functions; early return to reduce nesting
- Document public items (`///`); crate-level docs (`//!`) explain purpose and usage

## Naming
- **Crates/modules**: snake_case, short, no redundant prefixes; module name should not repeat the crate name
- **Types/traits**: UpperCamelCase; traits named for capability (`Read`, `Iterator`) or `-able`/`-er` where it reads naturally
- **Functions/variables**: snake_case; `new`/`with_*` for constructors, `from_*`/`try_from_*` for conversions, `as_*`/`to_*`/`into_*` per the standard borrow/consume convention
- **Constants/statics**: SCREAMING_SNAKE_CASE
- **Generics**: single uppercase letter (`T`, `E`) for simple cases, descriptive UpperCamelCase (`Item`, `Output`) when meaning isn't obvious from context

## Code Style and Formatting
- **Formatting**: `rustfmt` always (project `rustfmt.toml` for overrides), `clippy` clean before commit
- **Comments**: `///` for public API docs (first line is a one-sentence summary), `//` for implementation notes explaining why, not what
- **Error handling**: `Result<T, E>` for recoverable errors, never `unwrap()`/`expect()` outside tests/prototypes without justification; `?` operator for propagation; `panic!` only for unrecoverable programmer errors (invariant violations)

## Architecture and Project Structure
- **Workspace/crate layout**: binaries in `src/bin/` or `cmd/`-style separate crates; reusable logic in a library crate (`src/lib.rs`); `mod.rs` or file-per-module (`foo.rs` + `foo/`) consistently within a project
- **Visibility**: default to private; `pub(crate)` for crate-internal sharing before reaching for `pub`; minimize public surface area
- **Dependency management**: `Cargo.toml` with pinned or range-constrained versions per project policy; `cargo update` deliberately, not blindly; `cargo audit` for known vulnerabilities

## Type Safety and Language Features
- **Ownership/borrowing**: prefer borrowing (`&T`/`&mut T`) over cloning; use `Cow<'_, T>` when a function may or may not need to own data; lifetimes explicit only when inference can't resolve them
- **Types**: newtype pattern to prevent primitive obsession (`struct UserId(u64)` not a bare `u64`); enums with data over boolean/string flags; `Option<T>` instead of sentinel values
- **Traits/generics**: accept `impl Trait` or generics for flexibility at call sites, return concrete types unless dynamic dispatch is required (`Box<dyn Trait>`); default trait methods to reduce boilerplate; `derive` (`Debug`, `Clone`, `PartialEq`, `Serialize`) instead of hand-writing when possible

## Concurrency
- **Threads**: `std::thread` for OS threads, scoped threads (`thread::scope`) to borrow non-`'static` data safely; `Send`/`Sync` bounds communicate thread-safety at the type level - don't fight them with `unsafe impl` unless verified sound
- **Async**: `tokio` (or project's chosen runtime) for I/O-bound async; `async fn` returns a future - never block inside one; `.await` points are yield points, keep critical sections short
- **Synchronization**: channels (`std::sync::mpsc`, `tokio::sync::mpsc`) for message passing over shared state; `Arc<Mutex<T>>` for shared mutable state, keep lock scope minimal; `RwLock` for read-heavy workloads

## Error Handling Patterns
- **Creating**: custom error enums implementing `std::error::Error` for libraries; `thiserror` to reduce boilerplate; `anyhow` for application-level error aggregation where the caller won't match on variants
- **Propagation**: `?` to bubble up with `From`/`Into` conversions between error types; add context at boundaries (`.context()` with `anyhow`, or wrap in a domain error); don't both log and return - choose one per layer

## API Design
- **Public APIs**: builder pattern for structs with many optional fields; `impl Default` where a sensible zero-config exists; avoid `pub` fields on structs with invariants to maintain - use accessor methods
- **Serialization**: `serde` with `#[derive(Serialize, Deserialize)]`; explicit `#[serde(rename = "...")]` for wire-format mismatches; validate after deserializing untrusted input

## Performance
- **Allocation**: avoid unnecessary `clone()`/`to_owned()`; prefer iterators over collecting into intermediate `Vec`s; `Vec::with_capacity` when size is known ahead of time
- **Profiling**: `cargo flamegraph` or `perf` for hot-path analysis; benchmark with `criterion` before optimizing; measure before assuming `unsafe` or hand-rolled code is needed

## Testing
- **Organization**: unit tests in `#[cfg(test)] mod tests` next to the code; integration tests in `tests/`; doc tests (`///` code blocks) for public API examples that double as documentation
- **Writing**: `#[test]` per case or table-driven with a loop/macro for many similar cases; `#[should_panic]` for expected-panic cases; `assert_eq!`/`assert!` with descriptive messages
- **Tools**: `cargo test`, `cargo nextest` for faster/parallel runs, `cargo tarpaulin` or `cargo llvm-cov` for coverage

## Security
- **Unsafe code**: minimize `unsafe` blocks; document the invariant that makes each one sound (`// SAFETY: ...`); isolate in small, well-tested modules; prefer safe abstractions from vetted crates over hand-rolled `unsafe`
- **Input validation**: validate all external input at the boundary; strong typing (newtypes, enums) to make invalid states unrepresentable downstream; `cargo audit`/`cargo deny` in CI for dependency vulnerabilities

## Documentation
- **Code**: `///` on all public items starting with a one-sentence summary; `# Examples`, `# Panics`, `# Errors` sections on functions where relevant; `cargo doc` must build without warnings
- **README**: setup instructions (`cargo build`/`cargo run`), MSRV (minimum supported Rust version) if pinned, usage examples, feature flags documented

## Tools and Workflow
- **Essential tools**: `cargo fmt`, `cargo clippy -- -D warnings`, `cargo test`, `cargo audit`, `cargo doc`
- **Practices**: `clippy` clean and formatted before commit; pre-commit hooks for fmt/clippy; focused/atomic commits; review diffs before committing

## Common Pitfalls to Avoid
Overusing `unwrap()`/`expect()` in non-test code; fighting the borrow checker with excessive `clone()` instead of restructuring; `unsafe` without a documented safety justification; blocking calls inside `async fn`; ignoring `clippy` lints; stringly-typed data where an enum or newtype belongs; panicking on recoverable errors.

## Usage Example

```rust
#[derive(Debug, thiserror::Error)]
enum UserError {
    #[error("user {0} not found")]
    NotFound(UserId),
}

async fn find_user(id: UserId, repo: &impl UserRepository) -> Result<User, UserError> {
    repo.find(id).await?.ok_or(UserError::NotFound(id))
}
```
