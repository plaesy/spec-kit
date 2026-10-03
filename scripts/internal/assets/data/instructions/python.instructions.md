---
description: 'Python development standards and best practices - PEP 8 style, typing, packaging, testing, async, and the anti-patterns that bite'
applyTo: '**/*.py,**/pyproject.toml,**/setup.py,**/setup.cfg,**/requirements.txt'
---

# Python Development Instructions

Readable, explicit Python. PEP 8 is the floor, not the ceiling: it fixes the
formatting so the reviewer spends attention on the logic instead.

Sources (retrieved 2026-09-27): [PEP 8 — Style Guide for Python Code](https://peps.python.org/pep-0008/),
[PEP 20 — The Zen of Python](https://peps.python.org/pep-0020/),
[PEP 257 — Docstring Conventions](https://peps.python.org/pep-0257/),
[PEP 484 — Type Hints](https://peps.python.org/pep-0484/),
[`typing` module reference](https://docs.python.org/3/library/typing.html),
[Python Developer's Guide](https://docs.python.org/3/devguide/).
Version-sensitive syntax (`type X = ...`, PEP 695 generics, `except*`) is called
out with the version that introduced it — check the project's minimum supported
version before using it.

## Project Context

- Follow PEP 8; a project with its own agreed style takes precedence (PEP 8 says so
  explicitly: "consistency within a project is more important")
- Type hints on public surfaces are expected; the runtime does not enforce them
  (they are for type checkers, IDEs and linters)
- Managed environments and declared dependencies — no `pip install` into a system
  interpreter as part of a deploy
- The project's own formatter/linter config (black, ruff, flake8) is the source of
  truth for mechanical style; do not hand-format around it

## Development Standards

### Formatting and layout

- 4 spaces per indentation level, never tabs mixed with spaces
- 79-character lines (72 for docstrings/comments) unless the project has agreed
  otherwise; wrap inside parentheses/brackets rather than with backslashes
- Two blank lines around top-level defs/classes, one blank line between methods
- Imports at the top, one per line, grouped: stdlib, then third party, then local,
  blank line between groups; absolute imports by default
- No trailing whitespace; one statement per line (no `if x: y; z`)
- No wildcard imports (`from x import *`) — they hide the real namespace

### Naming

- Modules/packages: short, lowercase, underscores only if they aid readability
- Classes: CapWords; exceptions: CapWords + `Error` suffix
- Functions/variables: lowercase_with_underscores; constants: UPPER_CASE
- Never single-letter `l`, `O`, or `I` — use `L` instead of `l`
- `_single_leading_underscore` = internal; `__dunder__` only for documented
  protocol names, never invent one

### Types and typing

- Annotate function parameters and return values on public APIs
- `Optional[X]` is not the same as "has a default" — a parameter that may be
  passed `None` needs the annotation regardless of whether it has a default
- Prefer `X | None` and builtin generics (`list[str]`, `dict[str, int]`) when the
  minimum supported Python version allows them; `TypeAlias` is deprecated in favour
  of the `type` statement on 3.12+ (docs.python.org `typing`, retrieved 2026-09-27)
- Use `NewType` to make a distinct, cheaply-checked identifier type (e.g. `UserId`
  vs raw `int`) — it is enforced statically only, never at runtime
- `Any` disables checking in both directions; prefer `object` when the value may be
  any type and you still want safety
- Type checkers are optional tooling — never change runtime behaviour based on an
  annotation

### Errors and control flow

- `raise X from Y` when replacing an error; `from None` only when the detail is
  transferred into the new message
- Derive from `Exception`, never directly from `BaseException`
- Catch the narrowest exception that can actually be raised; a bare `except:`
  swallows `KeyboardInterrupt` and `SystemExit`
- Keep the `try` block to the minimum that can fail; use `else` for the success path
  so a bug in the handler is not masked as a `KeyError`
- `is None` / `is not None`, not `== None` or `not x is None`
- `isinstance()`, never `type(x) == int`
- Define real exception classes for callers to catch, instead of returning
  sentinel error strings; hierarchy shaped by what callers need to distinguish
- Use `def`, not `f = lambda x: ...`, when the name is meaningful (tracebacks and
  `repr` show the name)

### Resources and side effects

- `with` for anything that needs closing (files, sockets, DB connections, locks)
- Context managers should be a separate, clearly named function if they do more than
  acquire/release — `with conn:` hides what `__exit__` actually does
- No module-level mutable state or import-time I/O; work belongs behind a function
  so importing is cheap and side-effect free

### Data modelling

- `dataclasses` for value objects; `NamedTuple` when immutability plus tuple
  behaviour is wanted; `Enum` over magic strings for a closed set of states
- `__slots__` on hot, high-cardinality classes only
- Prefer comprehensions and generators for simple transforms; do not compress
  readable code into a nested comprehension
- `''.join([...])` over repeated `+=` in loops — the in-place optimisation is
  CPython-specific and fragile (PEP 8)

### Modules and packages

- `src/` layout with an explicit package marker; keep `__init__.py` thin — re-export
  a deliberate public API with `__all__`, do not import submodules for side effects
- Relative imports inside a package, absolute from a top-level script
- One responsibility per module; a module that needs a "misc" name is two modules

### Concurrency and async

- `asyncio` for I/O concurrency; `multiprocessing` for CPU-bound work; threads for
  blocking I/O that has no async client
- `async def` all the way down — never block the event loop with sync I/O; if you
  must, push it to a thread executor deliberately
- Structured concurrency: every task is awaited, cancelled, or explicitly
  fire-and-forget with a documented reason
- `Lock` (or `asyncio.Lock`) around shared state; prefer a queue or a channel over a
  shared counter wherever the shape allows
- Timers and background tasks need a shutdown path — no task outliving the process

### Testing

- `pytest` as the default runner; tests in `tests/`, named `test_*.py`
- One behaviour per test; arrange/act/assert in visible order; a failing assert
  message that names the actual behaviour, not the mechanism
- Parametrize instead of copy-pasting near-identical tests
- Fixtures over module-level setup; `tmp_path` over writing into the repo
- `monkeypatch` for env/global state; always restore — use the fixture, do not
  hand-roll setUp/tearDown for it
- Assert on behaviour through the public surface; do not assert on log text
- Coverage measured and enforced in CI, not assumed — an unrun check must not
  report as passing

### Security

- Never `eval`, `exec`, or `pickle` on untrusted input; `yaml.load` without a safe
  loader is the same class of bug
- Parameterise SQL — never string-format user input into a query
- `secrets` for tokens/passwords (not `random`); `hashlib.scrypt`/`pbkdf2_hmac` or
  a vetted password library for passwords
- Validate and coerce at the boundary; do not trust that a caller sent the right type
- Never log secrets, tokens, or full request bodies containing PII
- `pip install` is arbitrary code execution — pin, review, and audit the lockfile
- Subprocess: pass argument lists, never `shell=True` with interpolated input

### Tooling and workflow

- `ruff` (lint + format) or the project's existing black/flake8/mypy configuration;
  run them in pre-commit and in CI
- Type-check with mypy/pyright on the modules that benefit; keep it strict where it
  is already adopted, do not start a half-strict config
- One commit per logical change, tests in the same commit, messages that say why

## Common Pitfalls to Avoid

Mutable default arguments (`def f(x=[])`); broad `except:`; bare `except
Exception` around a whole function; shadowing builtins (`list`, `id`, `type`);
late-binding closures in loops; `is` on ints/strings; mutating a dict while
iterating it; forgetting `await`; sync I/O inside `async def`; circular imports
worked around with function-level imports everywhere; unpickling untrusted data;
`os.path` string concatenation instead of `pathlib`; catching and returning `None`
to signal failure.

## Usage Example

```python
from collections.abc import Sequence
from dataclasses import dataclass, field

@dataclass(frozen=True)
class Order:
    id: str
    lines: Sequence[str] = field(default_factory=tuple)

def line_total(order: Order) -> int:
    """Return the character total for one order. Raises ValueError when malformed."""
    if not order.id:
        raise ValueError("order is missing an id")
    return sum(len(line) for line in order.lines)
```

## References

- PEP 8 — https://peps.python.org/pep-0008/
- PEP 20 — https://peps.python.org/pep-0020/
- PEP 257 — https://peps.python.org/pep-0257/
- PEP 484 — https://peps.python.org/pep-0484/
- `typing` reference — https://docs.python.org/3/library/typing.html
- Developer's Guide — https://docs.python.org/3/devguide/
