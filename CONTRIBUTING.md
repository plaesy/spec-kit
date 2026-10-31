# Contributing to Spec-Kit

Thanks for your interest in contributing! This project is a **business validation pipeline** (niche → idea → validation → make money verdict) with optional YouTube automation.

## Before You Start

- Search existing [issues](https://github.com/plaesy/spec-kit/issues) and [pull requests](https://github.com/plaesy/spec-kit/pulls) to avoid duplicate work.
- For a non-trivial change, open an issue first to discuss the approach.

## Development Setup

### 1. Fork & Clone
```bash
git clone https://github.com/your-username/spec-kit.git
cd spec-kit
```

### 2. Install Dependencies
```bash
# Python 3.11+ for skills and oracle harness
pip install -q pytest pyyaml

# Optional: Go 1.22+ for plaesy CLI (if modifying CLI)
# cd scripts && go build ./... && go vet ./... && go test ./...
```

### 3. Run Tests Before PR
```bash
# All skill evals
for skill in skills/*/evals; do
  python -m pytest "$skill" -v
done

# Oracle harness replay (critical skills)
python scripts/oracle_runner.py --skill all --replay

# Full CI simulation
python -m pytest skills/*/evals/ -v --tb=short
python scripts/oracle_runner.py --skill all --replay
```

---

## Contribution Types

| Type | Prefix | Example |
|------|--------|---------|
| New Skill | `feat(skill-name):` | `feat(market-research): add TAM/SAM/SOM calculator` |
| Bug Fix | `fix(skill-name):` | `fix(economics-engine): handle negative margin edge case` |
| Documentation | `docs:` | `docs: update CLI-CONTRACT for youtube-automation` |
| Test/Eval | `test:` | `test(youtube-automation): add findability score edge cases` |
| Refactor | `refactor:` | `refactor(fallback_chains): extract base Tool class` |
| CI/CD | `ci:` | `ci: add oracle replay to workflow` |

---

## Skill Development Guidelines

### 1. Every Skill Must Have

```
skills/your-skill/
├── SKILL.md                    # Frontmatter + documentation
├── main.py                     # Entry point (or skill.py)
├── evals/
│   ├── evals.json              # Test definitions with assertions
│   └── results.json            # Evidence: 100% pass
├── oracle/
│   └── goldens.json            # Golden tests for critical functions
├── references/                 # Templates, data files (optional)
└── tests/                      # Offline pytest suite (no API keys)
    └── test_*.py
```

### 2. SKILL.md Requirements

**Frontmatter (required):**
```yaml
---
name: your-skill
category: productivity
description: One sentence, trigger first, ends with period, ≤60 chars
---
```

**Body must include:**
- `## When to Use` — explicit triggers
- `## Input` — parameters with types
- `## Process` — numbered steps
- `## Output` — format specification
- `## Quality Gates` — blocking criteria
- `## Evals` — reference to evals/

### 3. Eval Requirements

**evals.json** — Test definitions:
```json
{
  "skill": "your-skill",
  "version": "1.0",
  "tests": [
    {
      "name": "test_name",
      "description": "What this tests",
      "input": {...},
      "expected_output": {...},
      "assertions": ["assertion1", "assertion2"]
    }
  ]
}
```

**results.json** — Evidence (must show 100% pass):
```json
{
  "skill": "your-skill",
  "version": "1.0",
  "run_date": "YYYY-MM-DD",
  "results": [
    {"test": "test_name", "status": "passed", "output": {...}}
  ],
  "summary": {"total": N, "passed": N, "failed": 0, "pass_rate": 1.0}
}
```

### 3. Oracle Golden Tests (for critical skills)

For skills with verdict/decision logic (economics-engine, pipeline, youtube-automation, founder, buyer-discovery, compliance-health, plan-compiler, assess):

```json
{
  "skill": "your-skill",
  "version": "1.0",
  "engine": "python",
  "goldens": [
    {
      "name": "critical_function",
      "description": "What it tests",
      "input": {...},
      "expected_output": {...},
      "tolerance": {"numeric_field": 0.001}
    }
  ]
}
```

Run: `python scripts/oracle_runner.py --skill your-skill --replay`

---

## Code Style

### Python
- Type hints required for all public functions
- Docstrings for all public classes/functions (Google style)
- Max line length: 100 chars
- Use `pathlib.Path` not `os.path`

```python
def calculate_margin(price: float, cost: float) -> float:
    """Calculate contribution margin per sale.
    
    Args:
        price: Price per sale (P)
        cost: Direct cost per sale (V)
    
    Returns:
        Margin per sale (m = P - V)
    """
    return price - cost
```

### Markdown
- No trailing whitespace
- Tables use pipes with header separator
- Code blocks specify language
- Links use relative paths when possible

### Commit Messages
Follow Conventional Commits:
```
<type>(<scope>): <subject>

<body>

<footer>
```

Examples:
```
feat(economics-engine): add segment mix sensitivity analysis

Adds segment_mix sensitivity to test coach-heavy vs B2B-heavy mixes.
Updates evals with 2 new golden tests.

Closes #42
```

---

## Adding a New Skill

1. **Create skill directory**
   ```bash
   mkdir -p skills/new-skill/{evals,oracle,references,tests}
   ```

2. **Write SKILL.md** (follow template)

3. **Implement main.py** with CLI entry point

4. **Write evals/evals.json** with ≥3 test cases

5. **Run evals locally**
   ```bash
   python -m pytest skills/new-skill/evals/ -v
   ```

6. **Create oracle/goldens.json** (if critical logic)

7. **Run oracle harness**
   ```bash
   python scripts/oracle_runner.py --skill new-skill --record --golden test_name
   python scripts/oracle_runner.py --skill new-skill --replay
   ```

8. **Add to using-plaesy/SKILL.md** skill registry

10. **Update plugin manifests** (`.claude-plugin`, `.codex-plugin`, `.agents/plugins`)

11. **Add to CI matrix** in `.github/workflows/ci.yml`

12. **Write docs** in `docs/` if needed

---

## Running Tests Locally

```bash
# All evals
for skill in skills/*/evals; do
  echo "=== $(basename $(dirname $skill)) ==="
  python -m pytest "$skill" -v
done

# All oracle replays
for skill in economics-engine pipeline youtube-automation founder buyer-discovery compliance-health plan-compiler assess; do
  echo "=== $skill ==="
  python scripts/oracle_runner.py --skill "$skill" --replay
done

# Full CI simulation
python -m pytest skills/*/evals/ -v --tb=short
python scripts/oracle_runner.py --skill all --replay
```

---

## Code Review Checklist

- [ ] SKILL.md frontmatter valid (desc ≤60 chars, category set)
- [ ] `## When to Use` section present
- [ ] evals.json has ≥3 test cases
- [ ] results.json shows 100% pass
- [ ] Oracle goldens added (if critical logic)
- [ ] Oracle replay passes
- [ ] Offline pytest suite in tests/ (no API keys)
- [ ] Plugin manifests updated
- [ ] CI matrix updated
- [ ] Docs updated (ARCHITECTURE.md, DESIGN.md, CLI-CONTRACT.md if CLI changes)
- [ ] CHANGELOG.md entry added
- [ ] Commit follows Conventional Commits
- [ ] No API keys/secrets in code
- [ ] Fail-soft chains for external calls
- [ ] Quality gates documented

---

## Release Process

Maintainers only:

1. Update version in all plugin manifests
2. Update `CHANGELOG.md`
3. Tag release: `git tag -a v1.x.x -m "Release v1.x.x"`
3. Push tag: `git push origin v1.x.x`
4. GitHub Actions builds and publishes artifacts

---

## Questions?

Open an issue or join the discussion in [Discord](https://discord.gg/example).