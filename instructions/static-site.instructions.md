---
description: 'Static website (plain HTML/CSS/JS or a static-site generator) development, performance, SEO, and deployment standards'
applyTo: '**/*.html, **/index.html, **/astro.config.*, **/eleventy.config.*, **/.eleventy.js, **/_config.yml, **/hugo.toml, **/hugo.yaml, **/config.toml'
---

# Static Site Instructions

Guidance for a site that ships as pre-rendered HTML/CSS/JS with no application
server at runtime — a personal site, portfolio, landing page, or blog — whether
hand-written or built with a static-site generator (Astro, Eleventy, Hugo,
Jekyll, plain HTML).

## Project Context

- No backend process serves requests at runtime; the build step (if any)
  produces a folder of static files an edge/CDN host serves directly
- Pick a generator only when templating/content-reuse actually earns its
  complexity — a handful of hand-written HTML pages sharing a `<head>` partial
  via the host's include mechanism (or a tiny build script) is a legitimate,
  simpler choice than pulling in a generator for a 3-page personal site
- Common generators and when they fit: **Astro** (component islands, mixed
  content + interactivity), **Eleventy** (content-heavy, minimal opinion),
  **Hugo** (large content volume, build speed matters), **Jekyll** (GitHub
  Pages native support, blog-first)

## Development Standards

### Structure & Content
- One canonical source per page — if using a generator, content lives in its
  content directory (Markdown/frontmatter), not hand-duplicated across pages
- Use semantic HTML5 landmarks (`<header>`, `<nav>`, `<main>`, `<footer>`,
  `<article>`) instead of generic `<div>` soup — this is both an accessibility
  and an SEO signal, not just style
- Every page ships its own `<title>` and `<meta name="description">` — never a
  site-wide default copy-pasted across pages

### Performance
- Ship no more CSS/JS than the page uses — avoid a site-wide bundle when a
  handful of pages don't share behavior; static sites have no runtime excuse
  for unused code shipping to every visitor
- Serve images already sized/compressed for their largest rendered dimension,
  in a modern format (WebP/AVIF with a fallback) — do not rely on the browser
  to downscale a full-resolution source
- Set `width`/`height` (or `aspect-ratio`) on every `<img>` to prevent layout
  shift; `loading="lazy"` on below-the-fold images
- Self-host fonts (or use `font-display: swap`) rather than a render-blocking
  third-party font request with no fallback strategy
- A static site with a Lighthouse performance score below ~90 almost always
  means unoptimized images or unused JS/CSS shipped globally — check those
  first before anything more exotic

### SEO & Discoverability
- Ship `robots.txt` and an XML sitemap (generators: usually a plugin; hand-built:
  a small build-time script) — do not hand-maintain a sitemap by memory
- Use Open Graph + Twitter Card meta tags on every page meant to be shared,
  with an actual per-page image, not one shared placeholder
- Canonical URLs (`<link rel="canonical">`) when the same content is reachable
  at more than one path (e.g. with/without trailing slash)
- Structured data (JSON-LD) for content types that benefit from it (articles,
  person/portfolio, FAQ) — only where it's factually accurate, never to game
  rich results

### Accessibility
- Follow `.plaesy/roles/accessibility.md` — color contrast, focus-visible
  states, alt text that describes function not just appearance, keyboard
  navigability of any interactive component (nav toggle, carousel, etc.)
- A static site has no excuse for missing alt text or unlabeled form
  controls — there's no dynamic content to explain away the gap

### Deployment
- Prefer a host matched to the generator's zero-config path (Netlify, Vercel,
  Cloudflare Pages, GitHub Pages) over a hand-rolled server for a purely
  static output — a static site does not need a VM/container unless it has
  requirements beyond static hosting (custom headers logic, edge functions)
- Set cache headers appropriately: long-lived immutable caching for
  hashed/fingerprinted assets, short/no-cache for `index.html` and other
  entry points, so deploys are visible immediately without a cache-bust dance
- Enable HTTPS (default on all major static hosts) and a redirect from the
  bare/`www` variant that isn't canonical to the one that is, not both live

### Testing and Validation
- Validate HTML (e.g. via the W3C validator or an equivalent linter) as part
  of the build, not as an occasional manual check
- Run Lighthouse (or an equivalent) for performance/accessibility/SEO scores
  before calling a page done — treat a regression here the same as a failing
  test, not a "nice to have"
- Check broken-link status for internal links at minimum before publishing;
  external links periodically, since they rot independently of your changes

## Common Pitfalls

- ❌ Pulling in a full generator/framework for a 2-3 page personal site where
  hand-written HTML plus a build-time include for shared `<head>`/nav would be
  simpler and have zero framework upgrade surface
- ❌ Shipping a site-wide CSS/JS bundle that most pages don't need
- ❌ Missing per-page meta description/OG image, so every shared link looks
  identical regardless of which page was shared
- ❌ Unsized images causing layout shift, or full-resolution source images
  served untouched
