---
description: "Dynamic date system for Context7 citations and time-based validation"
applyTo: "**/*.md"
triggers: ["citation", "context7", "date variable", "year validation"]
---

# Date System

## Date Variables (Auto-Replacement)

- `{{CURRENT_DATE}}` - Current date (YYYY-MM-DD)
- `{{CURRENT_YEAR}}` - Current year (YYYY)
- `{{CURRENT_MONTH}}` - Current month (MM)
- `{{CURRENT_DAY}}` - Current day (DD)

## Usage Examples

### Citation Format
```markdown
"Source: Context7 (/library/id) - Retrieved {{CURRENT_DATE}}"
```

### Validation
```markdown
"✅ Using current data from {{CURRENT_YEAR}}"
"❌ Forbidden: Using pre-{{CURRENT_YEAR}} data"
```

---

## Implementation Protocol

### For AI Assistants
1. Replace ALL hardcoded dates with variables
2. Use `{{CURRENT_DATE}}` for citations
3. Use `{{CURRENT_YEAR}}` for year comparisons
4. Verify current system date before execution
5. Use actual current date if variables mismatch

### Validation Rules
```markdown
✅ Correct: "Retrieved {{CURRENT_DATE}}"
✅ Correct: "data from {{CURRENT_YEAR}}"
❌ Incorrect: "Retrieved 2025-10-11" (hardcoded)
❌ Incorrect: "data from 2025" (hardcoded)
```

---

## Date Variable Reference

| Variable | Format | Auto-Replaces |
|----------|--------|---------------|
| `{{CURRENT_DATE}}` | YYYY-MM-DD | Daily |
| `{{CURRENT_YEAR}}` | YYYY | Yearly |
| `{{CURRENT_MONTH}}` | MM | Monthly |
| `{{CURRENT_DAY}}` | DD | Daily |

---

## Usage in Instructions

When writing instructions or specifications:

```markdown
## Context7 Citation Protocol

### Required Format
- [ ] Citation: "Source: Context7 (/library/id) - Retrieved {{CURRENT_DATE}}"
- [ ] Year validation: Use data from {{CURRENT_YEAR}} only
- [ ] Current date check: Verify {{CURRENT_DATE}} matches system date
```

---

## Future-Proof Benefits

- **Automatic Updates**: No manual date changes needed
- **Year Transitions**: Variables update automatically on January 1st
- **Backward Compatibility**: Old citations remain valid
- **Consistency**: Same format across all years (2026, 2027, 2028+)

---

*Date System v1.0.0*  
*Framework: Spec Kit Constitutional Development Framework*
