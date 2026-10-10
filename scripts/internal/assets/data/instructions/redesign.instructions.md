---
name: redesign-existing-projects
description: Upgrades existing websites and apps to premium quality. Audits current design, identifies generic AI patterns, and applies high-end design standards without breaking functionality.
---

# Redesign Skill

## How This Works

When applied to an existing project, follow this sequence:

1. **Scan** — Read the codebase. Identify the framework, styling method (Tailwind, vanilla CSS, styled-components, etc.), and current design patterns.
2. **Diagnose** — Run through the audit below. List every generic pattern, weak point, and missing state you find.
3. **Fix** — Apply targeted upgrades working with the existing stack. Do not rewrite from scratch. Improve what's there.

## Design Audit

### Keystone: Write the Reason

Before calling a redesign done, write a one-line reason for every major
decision: why this color, why this layout, why this typography, why this
spacing, why cards (or no cards), why this illustration or icon. If the reason
does not fit in one line, the decision is not settled and needs revisiting.
(Convention: anti-slop R-31, github.com/miqdadbadjuber/anti-slop, retrieved
2026-09-29.)

### Typography

Check for these problems and fix them:

- **[T-01] Browser default fonts or Inter everywhere.** Replace with a font that has character. Good options: `Geist`, `Outfit`, `Cabinet Grotesk`, `Satoshi`. For editorial/creative projects, pair a
  serif header with a sans-serif body.
- **[T-02] Headlines lack presence.** Increase size for display text, tighten letter-spacing, reduce line-height. Headlines should feel heavy and intentional.
- **[T-03] Body text too wide.** Limit paragraph width to roughly 65 characters. Increase line-height for readability.
- **[T-04] Only Regular (400) and Bold (700) weights used.** Introduce Medium (500) and SemiBold (600) for more subtle hierarchy.
- **[T-05] Numbers in proportional font.** Use a monospace font or enable tabular figures (`font-variant-numeric: tabular-nums`) for data-heavy interfaces.
- **[T-06] Missing letter-spacing adjustments.** Use negative tracking for large headers, positive tracking for small caps or labels.
- **[T-07] All-caps subheaders everywhere.** Try lowercase italics, sentence case, or small-caps instead.
- **[T-08] Orphaned words.** Single words sitting alone on the last line. Fix with `text-wrap: balance` or `text-wrap: pretty`.

### Color and Surfaces

- **[C-01] Pure `#000000` background.** Replace with off-black, dark charcoal, or tinted dark (`#0a0a0a`, `#121212`, or a dark navy).
- **[C-02] Oversaturated accent colors.** Keep saturation below 80%. Desaturate accents so they blend with neutrals instead of screaming.
- **[C-03] More than one accent color.** Pick one. Remove the rest. Consistency beats variety.
- **[C-04] Mixing warm and cool grays.** Stick to one gray family. Tint all grays with a consistent hue (warm or cool, not both).
- **[C-05] Purple/blue "AI gradient" aesthetic.** This is the most common AI design fingerprint. Replace with neutral bases and a single, considered accent.
- **[C-06] Generic `box-shadow`.** Tint shadows to match the background hue. Use colored shadows (e.g., dark blue shadow on a blue background) instead of pure black at low opacity.
- **[C-07] Flat design with zero texture.** Add subtle noise, grain, or micro-patterns to backgrounds. Pure flat vectors feel sterile.
- **[C-08] Perfectly even gradients.** Break the uniformity with radial gradients, noise overlays, or mesh gradients instead of standard linear 45-degree fades.
- **[C-09] Inconsistent lighting direction.** Audit all shadows to ensure they suggest a single, consistent light source.
- **[C-10] Random dark sections in a light mode page (or vice versa).** A single dark-background section breaking an otherwise light page looks like a copy-paste accident. Either commit to a full
  dark mode or keep a consistent background tone throughout. If contrast is needed, use a slightly darker shade of the same palette — not a sudden jump to `#111` in the middle of a cream page.
- **[C-11] Empty, flat sections with no visual depth.** Sections that are just text on a plain background feel unfinished. Add high-quality background imagery (blurred, overlaid, or masked), subtle
  patterns, or ambient gradients. Use reliable placeholder sources like `https://picsum.photos/seed/{name}/1920/1080` when real assets are not available. Experiment with background images behind hero
  sections, feature blocks, or CTAs — even a subtle full-width photo at low opacity adds presence.

### Layout

- **[L-01] Everything centered and symmetrical.** Break symmetry with offset margins, mixed aspect ratios, or left-aligned headers over centered content.
- **[L-02] Three equal card columns as feature row.** This is the most generic AI layout. Replace with a 2-column zig-zag, asymmetric grid, horizontal scroll, or masonry layout.
- **[L-03] Using `height: 100vh` for full-screen sections.** Replace with `min-height: 100dvh` to prevent layout jumping on mobile browsers (iOS Safari viewport bug).
- **[L-04] Complex flexbox percentage math.** Replace with CSS Grid for reliable multi-column structures.
- **[L-05] No max-width container.** Add a container constraint (around 1200-1440px) with auto margins so content doesn't stretch edge-to-edge on wide screens.
- **[L-06] Cards of equal height forced by flexbox.** Allow variable heights or use masonry when content varies in length.
- **[L-07] Uniform border-radius on everything.** Vary the radius: tighter on inner elements, softer on containers.
- **[L-08] No overlap or depth.** Elements sit flat next to each other. Use negative margins to create layering and visual depth.
- **[L-09] Symmetrical vertical padding.** Top and bottom padding are always identical. Adjust optically — bottom padding often needs to be slightly larger.
- **[L-10] Dashboard always has a left sidebar.** Try top navigation, a floating command menu, or a collapsible panel instead.
- **[L-11] Missing whitespace.** Double the spacing. Let the design breathe. Dense layouts work for data dashboards, not for marketing pages.
- **[L-12] Buttons not bottom-aligned in card groups.** When cards have different content lengths, CTAs end up at random heights. Pin buttons to the bottom of each card so they form a clean
  horizontal line regardless of content above.
- **[L-13] Feature lists starting at different vertical positions.** In pricing tables or comparison cards, the list of features should start at the same Y position across all columns. Use consistent
  spacing above the list or fixed-height title/price blocks.
- **[L-14] Inconsistent vertical rhythm in side-by-side elements.** When placing cards, columns, or panels next to each other, align shared elements (titles, descriptions, prices, buttons) across all
  items. Misaligned baselines make the layout look broken.
- **[L-15] Mathematical alignment that looks optically wrong.** Centering by the math doesn't always look centered to the eye. Icons next to text, play buttons in circles, or text in buttons often
  need 1-2px optical adjustments to feel right.

### Interactivity and States

- **[I-01] No hover states on buttons.** Add background shift, slight scale, or translate on hover.
- **[I-02] No active/pressed feedback.** Add a subtle `scale(0.98)` or `translateY(1px)` on press to simulate a physical click.
- **[I-03] Instant transitions with zero duration.** Add smooth transitions (200-300ms) to all interactive elements.
- **[I-04] Missing focus ring.** Ensure visible focus indicators for keyboard navigation. This is an accessibility requirement, not optional.
- **[I-05] No loading states.** Replace generic circular spinners with skeleton loaders that match the layout shape.
- **[I-06] No empty states.** An empty dashboard showing nothing is a missed opportunity. Design a composed "getting started" view.
- **[I-07] No error states.** Add clear, inline error messages for forms. Do not use `window.alert()`.
- **[I-08] Dead links.** Buttons that link to `#`. Either link to real destinations or visually disable them.
- **[I-09] No indication of current page in navigation.** Style the active nav link differently so users know where they are.
- **[I-10] Scroll jumping.** Anchor clicks jump instantly. Add `scroll-behavior: smooth`.
- **[I-11] Animations using `top`, `left`, `width`, `height`.** Switch to `transform` and `opacity` for GPU-accelerated, smooth animation.

### Content

- **[CO-01] Generic names like "John Doe" or "Jane Smith".** Use diverse, realistic-sounding names.
- **[CO-02] Fake round numbers like `99.99%`, `50%`, `$100.00`.** Use organic, messy data: `47.2%`, `$99.00`, `+1 (312) 847-1928`.
- **[CO-03] Placeholder company names like "Acme Corp", "Nexus", "SmartFlow".** Invent contextual, believable brand names.
- **[CO-04] AI copywriting cliches.** Never use "Elevate", "Seamless", "Unleash", "Next-Gen", "Game-changer", "Delve", "Tapestry", or "In the world of...". Write plain, specific language.
- **[CO-05] Exclamation marks in success messages.** Remove them. Be confident, not loud.
- **[CO-06] "Oops!" error messages.** Be direct: "Connection failed. Please try again."
- **[CO-07] Passive voice.** Use active voice: "We couldn't save your changes" instead of "Mistakes were made."
- **[CO-08] All blog post dates identical.** Randomize dates to appear real.
- **[CO-09] Same avatar image for multiple users.** Use unique assets for every distinct person.
- **[CO-10] Lorem Ipsum.** Never use placeholder latin text. Write real draft copy.
- **[CO-11] Title Case On Every Header.** Use sentence case instead.

### Component Patterns

- **[CP-01] Generic card look (border + shadow + white background).** Remove the border, or use only background color, or use only spacing. Cards should exist only when elevation communicates
  hierarchy.
- **[CP-02] Always one filled button + one ghost button.** Add text links or tertiary styles to reduce visual noise.
- **[CP-03] Pill-shaped "New" and "Beta" badges.** Try square badges, flags, or plain text labels.
- **[CP-04] Accordion FAQ sections.** Use a side-by-side list, searchable help, or inline progressive disclosure.
- **[CP-05] 3-card carousel testimonials with dots.** Replace with a masonry wall, embedded social posts, or a single rotating quote.
- **[CP-06] Pricing table with 3 towers.** Highlight the recommended tier with color and emphasis, not just extra height.
- **[CP-07] Modals for everything.** Use inline editing, slide-over panels, or expandable sections instead of popups for simple actions.
- **[CP-08] Avatar circles exclusively.** Try squircles or rounded squares for a less generic look.
- **[CP-09] Light/dark toggle always a sun/moon switch.** Use a dropdown, system preference detection, or integrate it into settings.
- **[CP-10] Footer link farm with 4 columns.** Simplify. Focus on main navigational paths and legally required links.

### Iconography

- **[IC-01] Lucide or Feather icons exclusively.** These are the "default" AI icon choice. Use Phosphor, Heroicons, or a custom set for differentiation.
- **[IC-02] Rocketship for "Launch", shield for "Security".** Replace cliche metaphors with less obvious icons (bolt, fingerprint, spark, vault).
- **[IC-03] Inconsistent stroke widths across icons.** Audit all icons and standardize to one stroke weight.
- **[IC-04] Missing favicon.** Always include a branded favicon.
- **[IC-05] Stock "diverse team" photos.** Use real team photos, candid shots, or a consistent illustration style instead of uncanny stock imagery.

### Code Quality

- **[CQ-01] Div soup.** Use semantic HTML: `<nav>`, `<main>`, `<article>`, `<aside>`, `<section>`.
- **[CQ-02] Inline styles mixed with CSS classes.** Move all styling to the project's styling system.
- **[CQ-03] Hardcoded pixel widths.** Use relative units (`%`, `rem`, `em`, `max-width`) for flexible layouts.
- **[CQ-04] Missing alt text on images.** Describe image content for screen readers. Never leave `alt=""` or `alt="image"` on meaningful images.
- **[CQ-05] Arbitrary z-index values like `9999`.** Establish a clean z-index scale in the theme/variables.
- **[CQ-06] Commented-out dead code.** Remove all debug artifacts before shipping.
- **[CQ-07] Import hallucinations.** Check that every import actually exists in `package.json` or the project dependencies.
- **[CQ-08] Missing meta tags.** Add proper `<title>`, `description`, `og:image`, and social sharing meta tags.

### Strategic Omissions (What AI Typically Forgets)

- **[SO-01] No legal links.** Add privacy policy and terms of service links in the footer.
- **[SO-02] No "back" navigation.** Dead ends in user flows. Every page needs a way back.
- **[SO-03] No custom 404 page.** Design a helpful, branded "page not found" experience.
- **[SO-04] No form validation.** Add client-side validation for emails, required fields, and format checks.
- **[SO-05] No "skip to content" link.** Essential for keyboard users. Add a hidden skip-link.
- **[SO-06] No cookie consent.** If required by jurisdiction, add a compliant consent banner.

## Upgrade Techniques

When upgrading a project, pull from these high-impact techniques to replace generic patterns:

### Typography Upgrades

- **Variable font animation.** Interpolate weight or width on scroll or hover for text that feels alive.
- **Outlined-to-fill transitions.** Text starts as a stroke outline and fills with color on scroll entry or interaction.
- **Text mask reveals.** Large typography acting as a window to video or animated imagery behind it.

### Layout Upgrades

- **Broken grid / asymmetry.** Elements that deliberately ignore column structure — overlapping, bleeding off-screen, or offset with calculated randomness.
- **Whitespace maximization.** Aggressive use of negative space to force focus on a single element.
- **Parallax card stacks.** Sections that stick and physically stack over each other during scroll.
- **Split-screen scroll.** Two halves of the screen sliding in opposite directions.

### Motion Upgrades

- **Smooth scroll with inertia.** Decouple scrolling from browser defaults for a heavier, cinematic feel.
- **Staggered entry.** Elements cascade in with slight delays, combining Y-axis translation with opacity fade. Never mount everything at once.
- **Spring physics.** Replace linear easing with spring-based motion for a natural, weighty feel on all interactive elements.
- **Scroll-driven reveals.** Content entering through expanding masks, wipes, or draw-on SVG paths tied to scroll progress.

### Surface Upgrades

- **True glassmorphism.** Go beyond `backdrop-filter: blur`. Add a 1px inner border and a subtle inner shadow to simulate edge refraction.
- **Spotlight borders.** Card borders that illuminate dynamically under the cursor.
- **Grain and noise overlays.** A fixed, pointer-events-none overlay with subtle noise to break digital flatness.
- **Colored, tinted shadows.** Shadows that carry the hue of the background rather than using generic black.

## Fix Priority

Apply changes in this order for maximum visual impact with minimum risk:

1. **Font swap** — biggest instant improvement, lowest risk
2. **Color palette cleanup** — remove clashing or oversaturated colors
3. **Hover and active states** — makes the interface feel alive
4. **Layout and spacing** — proper grid, max-width, consistent padding
5. **Replace generic components** — swap cliche patterns for modern alternatives
6. **Add loading, empty, and error states** — makes it feel finished
7. **Polish typography scale and spacing** — the premium final touch

## Rules

- Work with the existing tech stack. Do not migrate frameworks or styling libraries.
- Do not break existing functionality. Test after every change.
- Before importing any new library, check the project's dependency file first.
- If the project uses Tailwind, check the version (v3 vs v4) before modifying config.
- If the project has no framework, use vanilla CSS.
- Keep changes reviewable and focused. Small, targeted improvements over big rewrites.
- Every ID'd finding you act on gets a one-line reason in your report (see Keystone above) — a fix without a stated reason is not distinguishable from a guess.

## Usage Example

```text
1. Scan: Next.js + Tailwind v3, three equal-width feature cards, Inter everywhere.
2. Diagnose: generic 3-card layout, pure #000 background, no hover states,
   purple/blue AI gradient hero.
3. Fix (Fix Priority order):
   - Swap Inter → Geist for headings
   - Replace #000 → #0a0a0a, drop the purple/blue gradient for a single
     desaturated accent
   - Add hover/active states to all buttons (200ms transition)
   - Break the 3-card row into a 2-column zig-zag
→ Same Tailwind stack, no functionality changed, tested after each step.
```
