---
description: 'Word (.docx) generation and automation standards using python-docx'
applyTo: '**/*.docx, **/documents/**, **/*word*.py, **/*docx*.py'
---

# Word Automation Instructions

Generate and manipulate `.docx` files programmatically using `python-docx` (actively maintained, full OOXML control for logic-heavy document generation).

## Project Context
- Library: `python-docx` — paragraphs, runs, styles, tables, headers/footers, sections
- For template-driven generation with placeholder tags (`{{ field }}`), prefer `docxtpl` (python-docx-template, Jinja2-based) on top of python-docx rather than assembling documents purely with the low-level API
- Known unsupported-for-manipulation features: tracked changes, content controls, certain chart types, native footnotes — these are preserved on load/save but cannot be created/edited programmatically; footnotes require direct XML manipulation as a workaround
- Cloud target (OneDrive/SharePoint, not a local file): Microsoft Graph has no native document-editing endpoint for Word — upload/download the `.docx` via Graph's file endpoints and still edit it locally with `python-docx`/`docxtpl`; requires Azure AD app registration and OAuth2, so reserve it for cloud-native integrations, not local generation

## Dependencies
```
python-docx>=1.1      # core generation/manipulation
docxtpl>=0.19          # only if using Jinja2 tag-based template filling
```
- Pin exact versions in the project's `requirements.txt`/`pyproject.toml`
- Optional: LibreOffice (`soffice` on PATH) for headless render validation — see `.plaesy/scripts/bash/validate-docx.sh` / `.plaesy/scripts/powershell/validate-docx.ps1`; not required for structural validation, only for the render check

## Development Standards

### Template-First Design
- Design the Word template (styles, headers/footers, section breaks) in Word itself, then let Python fill dynamic content — do not attempt to replicate corporate formatting purely through generation code
- With `docxtpl`, define Jinja2 placeholders inside the template `.docx` and pass a context dict — this keeps layout/design in the hands of whoever owns the template, decoupled from code
- With raw `python-docx`, apply existing paragraph/character styles from the template (`doc.styles`) rather than manually setting font/size/color on every run

### Document Structure
- Use heading styles (`Heading 1`, `Heading 2`, …) for document structure instead of manually bolding/enlarging text — this keeps the generated TOC, navigation pane, and accessibility tooling functional
- Build tables via `add_table` with a defined style, not manually bordered text blocks
- Keep one clear content-to-document mapping: a data model in → one section/paragraph out, so generation stays traceable when the template changes

### Bulk Generation
- For batch document generation (e.g., one contract per customer), loop over data rows and save each output with a unique, collision-safe filename — do not mutate and re-save the same `Document` object across iterations; reload the template fresh per item
- Keep the data source (CSV/DB/JSON) and the template separate from the generation script so either can change independently

### Known Limitations — Design Around Them
- Do not promise tracked-changes or content-control manipulation — these are read/preserve only in python-docx
- Do not rely on native footnote creation — either avoid footnotes in generated output or implement via direct `lxml` manipulation of the document XML, clearly commented as a workaround
- Chart types beyond what python-docx exposes should be rendered as images and inserted with `add_picture`, same pattern as PowerPoint generation

### Testing and Validation
- Assert on paragraph count, heading text, and table structure after generation, not just successful save
- Round-trip test for templates: fill → reload → verify placeholder text was fully replaced (no leftover `{{ }}` tags)
- Immediately after generating the file, run it through `.plaesy/scripts/bash/validate-docx.sh <file.docx>` (bash) or `.plaesy/scripts/powershell/validate-docx.ps1 -DocxFile <file.docx>` (PowerShell) — do this as a normal step of the generation task, not something gated behind a pipeline

### Security
- Never embed macros (`.docm`) unless explicitly required and reviewed
- Sanitize any user-supplied content inserted into `docxtpl`/Jinja2 context — treat it as data, never render user input as a Jinja2 template string itself (avoid `Template(user_input)`)

## Starter Template

```python
"""Template-first DOCX generator skeleton — docxtpl for tag-based filling,
raw python-docx below for structural/logic-heavy generation."""
from docx import Document

TEMPLATE_PATH = "template.docx"   # styles/headers/footers owned by the template


def build_from_template(sections: list[dict], output_path: str):
    doc = Document(TEMPLATE_PATH)   # inherits styles, headers, footers
    for section in sections:
        doc.add_heading(section["heading"], level=section.get("level", 1))
        for paragraph in section["paragraphs"]:
            doc.add_paragraph(paragraph, style="Body Text")
        if "table" in section:
            rows, cols = section["table"]["rows"], section["table"]["cols"]
            table = doc.add_table(rows=len(rows) + 1, cols=cols)
            table.style = "Light Grid Accent 1"
            for c, header in enumerate(section["table"]["headers"]):
                table.cell(0, c).text = header
            for r, row in enumerate(rows, start=1):
                for c, value in enumerate(row):
                    table.cell(r, c).text = str(value)
    doc.save(output_path)


def build_from_docxtpl(context: dict, output_path: str):
    """Preferred path when the template already has {{ jinja2 }} tags."""
    from docxtpl import DocxTemplate

    tpl = DocxTemplate(TEMPLATE_PATH)
    tpl.render(context)   # never render user input as the template itself
    tpl.save(output_path)


if __name__ == "__main__":
    build_from_template(
        sections=[{"heading": "Summary", "paragraphs": ["Generated report body."]}],
        output_path="report.docx",
    )
```

## Implementation Process
1. Design the `.docx` template (styles, placeholders) in Word or reuse an approved one
2. Choose `docxtpl` for tag-based filling or raw `python-docx` for structural/logic-heavy generation
3. Populate content per document, reloading the template fresh for each item in batch runs
4. Validate structure (headings, tables, no leftover placeholders) before saving
5. Save with a unique filename per generated document, then run `validate-docx.sh`/`validate-docx.ps1` on the output file to confirm it opens correctly
