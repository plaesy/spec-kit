---
description: 'Excel (.xlsx) generation and automation standards using openpyxl/XlsxWriter'
applyTo: '**/*.xlsx, **/reports/**, **/spreadsheets/**, **/*excel*.py, **/*xlsx*.py'
---

# Excel Automation Instructions

Generate, read, and modify `.xlsx` files programmatically. Library choice depends on the operation — pick per scenario, not by default habit.

## Project Context
- **`openpyxl`**: read/modify existing workbooks, template-based reporting, cell styling, formulas, charts
- **`XlsxWriter`**: write-only generation of new large workbooks (better performance/memory for write-heavy, large-dataset cases)
- **`pandas`** (with `openpyxl`/`xlsxwriter` engine): tabular data in/out when the sheet is essentially a DataFrame dump
- Cloud target (OneDrive/SharePoint, not a local file): **Microsoft Graph API** (`/me/drive/items/{id}/workbook` range/cell endpoints support direct cell/range writes without downloading the file) — requires Azure AD app registration and OAuth2; reserve for cloud-native integrations, not local generation

## Dependencies
```
openpyxl>=3.1        # template-fill, modify-existing, styling, formulas
XlsxWriter>=3.2       # write-only, large from-scratch datasets (pick per scenario, not both by default)
pandas>=2.0           # optional — only if the data is naturally a DataFrame
```
- Pin exact versions in the project's `requirements.txt`/`pyproject.toml`
- Optional: LibreOffice (`soffice` on PATH) for headless render validation — see `plaesy validate-xlsx`; not required for structural validation, only for the render check

## Development Standards

### Choosing the Right Tool
- Use `openpyxl` when **modifying an existing file** or filling a pre-designed template — it is the only one of the two that can load and edit existing workbooks
- Use `XlsxWriter` when **writing a new large dataset from scratch** — it streams to disk and is significantly faster/lighter than openpyxl for this case
- Never default to `openpyxl` for large write-only jobs without measuring — memory use is roughly 50x the resulting file size (a 50MB output can mean ~2.5GB RAM) because the full workbook model is held in memory

### Memory and Performance
- When only reading, open with `load_workbook(path, read_only=True)` — near-instant load and drastically lower memory footprint; never use read-only mode if you also need to write back to the same object
- When writing very large sheets with openpyxl, use `Workbook(write_only=True)` with `ws.append()` row-by-row instead of cell-by-cell assignment
- Disable external link resolution (`data_only`/`keep_links` as appropriate) when only cached values are needed, to avoid openpyxl trying to open linked worksheets
- Batch cell writes; avoid per-cell style objects — reuse `NamedStyle`/style instances instead of constructing a new `Font`/`Fill` per cell

### Template-Based Reporting
- Load a pre-designed `.xlsx` template and fill data into named ranges or known cell coordinates rather than building formatting from scratch — preserves corporate styling, conditional formatting, and existing formulas
- Keep data and formatting concerns separate: a data-writing pass, then (only if needed) a formatting pass

### Formulas and Data Integrity
- Write formulas as strings (`ws['A1'] = '=SUM(B1:B10)'`); do not attempt to compute Excel formulas in Python — let Excel/LibreOffice evaluate on open
- Set explicit cell number formats (`cell.number_format`) for dates, currency, and percentages instead of relying on Excel's autodetection
- Validate that formula ranges match actual written data bounds before saving

### Testing and Validation
- Assert on sheet names, header rows, and a sample of data cells after generation — not just successful save
- For files intended to be reopened and edited by users, round-trip test: write → reload with `openpyxl` → verify structure survives
- Immediately after generating the file, run it through `plaesy validate-xlsx <file.xlsx> [expected_sheet_name...]` — do this as a normal step of the generation task, not something gated behind a pipeline

### Security
- Never enable or write VBA macros (`.xlsm`) unless explicitly required and reviewed
- Treat any formula-like user input written into cells as a CSV/formula-injection risk (e.g. values starting with `=`, `+`, `-`, `@`) — prefix with a safe character or quote as text when the source is untrusted

## Starter Template

```python
"""Template-first XLSX report skeleton — pick openpyxl (template-fill/read)
or XlsxWriter (large write-only) per the Development Standards above."""
from openpyxl import load_workbook

TEMPLATE_PATH = "template.xlsx"   # styling/formulas owned by the template
DATA_SHEET = "Data"
HEADER_ROW = 1


def fill_template(data_rows: list[dict], columns: list[str], output_path: str):
    wb = load_workbook(TEMPLATE_PATH)   # not read_only: we write back
    ws = wb[DATA_SHEET]

    for col_idx, name in enumerate(columns, start=1):
        assert ws.cell(row=HEADER_ROW, column=col_idx).value == name, \
            f"template header mismatch at column {col_idx}"

    for row_offset, row in enumerate(data_rows, start=1):
        for col_idx, name in enumerate(columns, start=1):
            ws.cell(row=HEADER_ROW + row_offset, column=col_idx, value=row[name])

    wb.save(output_path)


def write_large_dataset(data_rows: list[dict], columns: list[str], output_path: str):
    """Write-only path for large, from-scratch datasets."""
    import xlsxwriter

    wb = xlsxwriter.Workbook(output_path, {"constant_memory": True})
    ws = wb.add_worksheet(DATA_SHEET)
    ws.write_row(0, 0, columns)
    for r, row in enumerate(data_rows, start=1):
        ws.write_row(r, 0, [row[c] for c in columns])
    wb.close()


if __name__ == "__main__":
    rows = [{"name": "Alice", "score": 92}, {"name": "Bob", "score": 81}]
    fill_template(rows, columns=["name", "score"], output_path="report.xlsx")
```

## Implementation Process
1. Decide read vs. write-heavy vs. template-fill scenario, then pick `openpyxl` or `XlsxWriter` accordingly
2. Load template or initialize workbook per the chosen mode (read_only/write_only as applicable)
3. Write data in a batched, row-oriented pass
4. Apply formatting/formulas in a separate pass
5. Validate structure (sheets, headers, sample cells) before returning the file as done, then run `validate-xlsx.sh`/`validate-xlsx.ps1` on the output file to confirm it opens correctly
