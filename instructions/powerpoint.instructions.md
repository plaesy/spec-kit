---
description: 'PowerPoint (.pptx) generation and automation standards using python-pptx'
applyTo: '**/*.pptx, **/slides/**, **/presentations/**, **/*ppt*.py'
---

# PowerPoint Automation Instructions

Generate and manipulate `.pptx` files programmatically using `python-pptx` (MIT, no Microsoft Office installation required).

## Project Context
- Library: `python-pptx` — full OOXML control over slides, shapes, text, images, tables, charts
- Alternative when native Office is available: COM automation (`pywin32`) or Office Add-ins (Office.js) — use only when the task requires live editing inside a running PowerPoint instance
- Alternative when the target lives in OneDrive/SharePoint (not a local file): **Microsoft Graph API** (`/me/drive/items/{id}/workbook` family for Excel; for PowerPoint, Graph has no native slide-editing endpoint — upload/download the `.pptx` via Graph's file endpoints and still edit it locally with `python-pptx`) — requires Azure AD app registration and OAuth2, so reserve it for cloud-native integrations, not local generation
- For high-volume consulting-grade decks (KPI dashboards, SWOT, timelines) at scale, consider a rendering API layer on top instead of hand-rolling every shape

## Dependencies
```
python-pptx>=1.0.2      # core generation/manipulation
Pillow>=10.0             # only if resizing/compressing images before embedding
```
- Pin exact versions in the project's `requirements.txt`/`pyproject.toml`; `python-pptx` 1.0.x targets Python 3.8+ and current OOXML — verify against the project's Python version before pinning
- Optional: LibreOffice (`soffice` on PATH) for headless render validation — see `.plaesy/scripts/bash/validate-pptx.sh` / `.plaesy/scripts/powershell/validate-pptx.ps1`; not required for structural validation, only for the render check

## Development Standards

### Template-First Design
- Start from an existing `.pptx` template (`Presentation('template.pptx')`) instead of building slides from a blank deck — inherit theme, fonts, and master layouts
- Define placeholder shapes (title, body, picture, chart) in the template and fill them programmatically rather than positioning every shape from scratch
- Use `slide_layouts` from the template's slide master; never hardcode colors/fonts that already exist in the theme

### Text and Layout Safety
- Keep text within safe bounds of its placeholder/textbox — validate length or enable `word_wrap` and `auto_size = MSO_AUTO_SIZE.TEXT_TO_FIT_SHAPE` to avoid overflow
- Set explicit units via `Emu`/`Inches`/`Pt` — never assume default measurement units
- Avoid overlapping shapes; compute positions relative to slide width/height (`prs.slide_width`, `prs.slide_height`) so layouts stay resolution-independent

### Charts and Data
- Native chart support covers column, bar, line, pie, and area only — waterfall, funnel, treemap, radar, and sunburst are **not supported**; render these as images (matplotlib/plotly export) and insert via `add_picture` instead of forcing an unsupported chart type
- Use `CategoryChartData`/`chart.replace_data()` to update chart data without rebuilding the chart object
- Prefer native tables (`add_table`) over image-based tables so content remains editable and accessible

### Images and Assets
- Compress/resize images before embedding to keep file size reasonable — `python-pptx` does not recompress inserted images
- Reference external assets by path only at build time; never leave dangling references in the output file

### Reusable Generation Code
- Build a thin helper layer (functions per slide type: title slide, content slide, chart slide) rather than repeating 60+ lines of shape/formatting code per slide
- Separate content data (JSON/YAML/dict) from layout code so decks can be regenerated from new data without touching generation logic
- Validate the deck's structure (slide count, required placeholders present) before returning it as "done" — treat this as a quality gate, not an afterthought

### Testing and Validation
- Immediately after generating the file, run it through `.plaesy/scripts/bash/validate-pptx.sh <file.pptx> [expected_slide_count]` (bash) or `.plaesy/scripts/powershell/validate-pptx.ps1 -PptxFile <file.pptx> [-ExpectedSlides <int>]` (PowerShell) — do this as a normal step of the generation task, not something gated behind a pipeline. It always runs structural validation (opens the file via `python-pptx`, checks slide count/placeholders) and additionally headless-renders via LibreOffice when `soffice` is on PATH, catching corruption that the object model alone would miss
- For one-off/manual generation, running the script once right after `build()` is enough — no need to wire it into CI unless the generation runs repeatedly or unattended
- Assert on shape counts, placeholder text, and slide count in tests — not just "file was written without exception"

### Security
- Never embed macros (`.pptm`) unless explicitly required and reviewed — plain `.pptx` has no executable content
- Sanitize any user-supplied text inserted into shapes to avoid breaking XML structure (python-pptx escapes this automatically via the API — never string-concatenate raw XML from user input)

## Starter Template

```python
"""Template-first PPTX generator skeleton — copy and adapt per project."""
from pptx import Presentation
from pptx.util import Emu
from pptx.enum.text import MSO_AUTO_SIZE

TEMPLATE_PATH = "template.pptx"   # theme/master owned by design, not code
LAYOUT_TITLE = 0
LAYOUT_CONTENT = 1


def add_title_slide(prs: Presentation, title: str, subtitle: str):
    slide = prs.slides.add_slide(prs.slide_layouts[LAYOUT_TITLE])
    slide.shapes.title.text = title
    slide.placeholders[1].text = subtitle
    return slide


def add_content_slide(prs: Presentation, title: str, bullets: list[str]):
    slide = prs.slides.add_slide(prs.slide_layouts[LAYOUT_CONTENT])
    slide.shapes.title.text = title
    body = slide.placeholders[1].text_frame
    body.word_wrap = True
    body.auto_size = MSO_AUTO_SIZE.TEXT_TO_FIT_SHAPE
    body.clear()
    for i, bullet in enumerate(bullets):
        p = body.paragraphs[0] if i == 0 else body.add_paragraph()
        p.text = bullet
    return slide


def validate(prs: Presentation, expected_slide_count: int):
    assert len(prs.slides) == expected_slide_count, "slide count mismatch"
    for slide in prs.slides:
        assert slide.shapes.title is not None, "missing title placeholder"


def build(data: dict, output_path: str):
    prs = Presentation(TEMPLATE_PATH)
    add_title_slide(prs, data["title"], data["subtitle"])
    for section in data["sections"]:
        add_content_slide(prs, section["heading"], section["bullets"])
    validate(prs, expected_slide_count=1 + len(data["sections"]))
    prs.save(output_path)


if __name__ == "__main__":
    build(
        data={
            "title": "Q3 Review",
            "subtitle": "Prepared by Team",
            "sections": [{"heading": "Highlights", "bullets": ["Point A", "Point B"]}],
        },
        output_path="output.pptx",
    )
```

## Implementation Process
1. Load or create the base `Presentation` from a template
2. Map input data to slide layouts and placeholders
3. Populate text, tables, charts, and images per slide
4. Run structural validation (slide/placeholder counts)
5. Save, then immediately run `validate-pptx.sh`/`validate-pptx.ps1` on the output file to confirm it opens correctly and renders without corruption
