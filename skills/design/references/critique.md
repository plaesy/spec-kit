---
name: design-critique
description: "Judge a rendered frontend like a design juror: fourteen passes on the screenshots, a weighted score (design 40%, usability 30%, creativity 20%, content 10%) and hard caps that stop a page with a visible defect from scoring well. Use after a render, before the handoff."
applyTo: "**/*.html,**/*.css,**/*.scss,**/*.tsx,**/*.jsx,**/*.vue,**/*.svelte"
---

# Design critique

The lint and the render check show what is broken. This file decides whether the page
is good. Judge the screenshots, never the code: the author reads intentions into code,
and a picture shows only what is there. Use it after `frontend-design.instructions.md`
step "Verify with renders", on the desktop and phone captures at 1440, 768 and 390 px.

Adapted from taste-skill's critique reference (MIT; see `THIRD_PARTY_NOTICES.md`).

## 1. Fourteen passes

Make each pass on its own. Do not blend them.

1. **One second, then five.** Glance and look away. Exactly one answer: a word, an image,
   an object. If the answer is "a website", there is no focal point. Then give the first
   screen five seconds as a stranger: can you say what this is, who it is for and what to
   do next?
2. **Squint.** The page should reduce to two or three large masses with clear size
   differences. Many small similar masses means nothing leads.
3. **Cover the logo.** Could this first screen belong to a competitor, or to another
   industry? If swapping the name and the noun gives an equally valid page, there is no
   idea.
4. **Edges.** Count the distinct left x-positions of text blocks. Two or three, reused on
   purpose, is composed. Six means things were placed one at a time.
5. **Space and frame.** Related things are close, unrelated things far. The gap between
   sections is clearly larger than any gap inside one. Left and right margins match. No
   section keeps its text in one half and leaves the other empty. No blank stretch is
   taller than half a screen.
6. **Type.** The jump from body to headline is at least three times. No more than five
   sizes. No headline with one orphan word. No paragraph wider than about 75 characters.
7. **Colour.** Name the proportions: which colour holds most of the surface, which
   supports, where the accent is. An accent in ten places means nothing.
8. **Images.** At least four real pictures (photographs, archival images, footage);
   drawings do not count. Each image does a job. Crops share ratios. No grey box, stretched
   photo or fake screen of bars.
9. **Sameness.** Name each section's layout in one word (split, list, full-bleed, grid,
   statement). The same word twice in a row, or three times on the page, is a template.
10. **Phone.** The mobile screen is a composition of its own, not the desktop squeezed.
11. **Fullness.** Each desktop screen is a composed poster. A narrow column of small text
    with dead ground either side is underbuilt, however austere it looks.
12. **Substance.** Every section the brief asks for delivers its information completely and
    at a readable size. The primary action is a real button on the first screen.
13. **Collisions.** Nothing lies on a headline, statement or button at any scroll position.
    Nothing runs off the edge of the phone. A row that scrolls sideways reads as cut off
    unless the cut is plainly deliberate.
14. **Competitor and live thing.** Put the nearest competitor's name on the page: does the
    idea still work? Name the one thing the visitor can do, or that differs each visit. If
    either answer is weak, it is the first problem of the round.

Passes 11–13 check that removing ornament did not remove substance.

## 2. Score

Score each criterion from 1 to 10.

- **Design (40%)**, composition, typography, colour, imagery, craft. 9–10: every screen
  holds as a poster. 7: clean and professional, little tension. 5 or less: weak hierarchy,
  accidental spacing, default type, mismatched images.
- **Usability (30%)**, orientation in three seconds, readable text, obvious actions,
  honest mobile. 9–10: you know what it is and what to do without scrolling. 6 or less:
  you have to hunt, or something breaks.
- **Creativity (20%)**, an idea you can state in one sentence and carry through. 9–10: a
  concept specific to this subject. 6 or less: the category default.
- **Content (10%)**, words and pictures that are specific and edited. 9–10: every line
  could only be about this subject. 6 or less: slogans, filler, stock.

`overall = 0.4·design + 0.3·usability + 0.2·creativity + 0.1·content`

For calibration, an honest 9 is rare. If your first score for your own work is 9, score it
again as a juror who did not build it.

## 3. Caps

These override the arithmetic. The lowest applicable cap wins.

- Any render or lint FAIL still present: overall cannot exceed **6.0**.
- A headline of three or more lines at any width, an eyebrow label, monospace labels,
  01/02/03 numbering, a decorative gradient, a shadow on a resting element, or a decorative
  hairline: overall cannot exceed **6.5**.
- Pass 3 fails: creativity cannot exceed **6**. Pass 14 idea fails: creativity cannot
  exceed **7**. No live thing: creativity cannot exceed **7.5**.
- Fewer than four real pictures: design cannot exceed **6.5**.
- A page that barely moves, or one identical fade on every block: design cannot exceed **7**.
- Two or more desktop screens mostly empty ground, or running text small and grey: design
  cannot exceed **6.5**.
- Unequal margins, a section huddled to one side, or a blank stretch taller than half a
  screen: design cannot exceed **6.5**.
- Any placeholder (grey box, lorem, "Image here"): content cannot exceed **5**.
- A section the brief asks for is missing or reduced to a few words: content cannot exceed **6**.
- No real button for the primary action on the first screen: usability cannot exceed **7**.
- The five-second pass fails: usability cannot exceed **6**.
- A section of running text with no heading naming it: usability cannot exceed **7**.
- Anything covering a headline at any scroll position: usability cannot exceed **6**.
- Content clipped off-screen, or a blank band where a reveal did not fire: usability
  cannot exceed **5**.

## 4. Writing the critique and fixing

Write the findings as: pass number, what the picture shows, the cap it triggers (if any),
and the fix. Fix the top three problems, render again, and look again. Stop after three
rounds and list what remains open in the handoff. Do not claim a clean critique for a
round that did not run.
