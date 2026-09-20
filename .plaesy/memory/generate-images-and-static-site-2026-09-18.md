---
title: "New /generate:images command + static-site framework gap closed"
date: "2026-09-18"
---

## What Happened

User (working on a personal static website project consuming this spec-kit) asked
two things this session:

1. Noticed there was no `/generate:images` prompt for UI/design asset generation.
2. Later reported an `/improve` run surfaced a framework gap: no
   instructions/template for "static site" as an artifact type.

Both were genuine framework gaps (not project findings) — addressed directly in
the spec-kit source.

## Changes Made

### 1. New `/generate:images` command
- `prompts/generate.md` — thin router (mirrors `/assess`, `/improve`, `/fix`
  shape), one scope today: `images`.
- `prompts/generate/images.md` — full protocol: resolve brief (reads
  `.plaesy/memory/design-spine.md` / `instructions/brandkit.instructions.md` /
  `.plaesy/roles/designer.md` + `accessibility.md` when present) → resolve
  provider (env var `PLAESY_IMAGE_PROVIDER` or `.plaesy/scripts/configs/image-provider.json`,
  default `openai`) → call the generator script → validate output file is
  non-empty → report path + alt text + reproducible prompt. Also defines a
  **programmatic invocation contract** so other prompts (`/implement:design`,
  `/improve:design`, `/doc`) can call it mid-run instead of just describing an
  asset in prose.
- `scripts/bash/generate-image.sh` + `scripts/powershell/generate-image.ps1` —
  actually call an image-gen API (OpenAI `gpt-image-1` images endpoint, or
  Gemini/Imagen as an alternate `--provider`), base64-decode, write the PNG.
  **Requires `OPENAI_API_KEY` or `GEMINI_API_KEY` env var** — if unset, the
  prompt stops and tells the user exactly what to export, returning the
  composed prompt text as a manual fallback rather than pretending to succeed.
- `docs/prompts/README.md` nav table updated with the new command.

### 2. Static-site framework gap closed
- New `instructions/static-site.instructions.md` — covers plain HTML/CSS/JS and
  generators (Astro, Eleventy, Hugo, Jekyll): structure, performance (image
  sizing/format, unused CSS/JS, font loading), SEO (sitemap, robots.txt, OG
  tags, canonical URLs, JSON-LD), accessibility, deployment (Netlify/
  Vercel/Cloudflare Pages/GitHub Pages, cache headers), testing (HTML validation,
  Lighthouse, broken-link checks).
- Registered in `instructions/mapping.json` under `frameworks.static-site`
  (keywords: eleventy/11ty/hugo/jekyll/"static site generator"; filenames:
  `astro.config.{mjs,ts}`, `eleventy.config.js`, `.eleventy.js`, `_config.yml`,
  `hugo.{toml,yaml}`, `config.toml`).
- `prompts/improve.md` Artifact Mode "Identify the Artifact Type" table got a
  new row pointing Artifact-Mode `/improve` runs at `static-site.instructions.md`
  for this artifact type, so it's no longer reported as a gap.

## Known Limitation (Not Yet Fixed)

Plain hand-written HTML with **no** generator config file (e.g. a bare 2-3 page
personal site with just `index.html`, no `astro.config.*`/`_config.yml`/etc.)
will **not** trigger auto-detection in `mapping.json`/`detect-stack.*` — there's
no reliable filename/keyword signal for "plain HTML site" the way there is for a
named generator. The instructions file's own `applyTo: '**/*.html'` frontmatter
still makes it usable manually/by an editor that reads that field, but
`plaesy init`'s auto-copy won't pick it up on its own for a config-less site.
If this comes up again, the fix would be a low-priority `"extensions": [".html"]`
detection entry — deliberately not added yet since `.html` is too broad a signal
(would also match e.g. templates inside a non-static-site framework project) and
needs a smarter heuristic (e.g. "no other framework detected AND an index.html
exists at repo root") that neither detect-stack script currently supports.

## Not Yet Committed

Same as the rest of this session's backlog — see `context.md` `## Next`. These
changes are new/uncommitted files on top of the still-pending colon-scope and
stack-detection batches.
