---
applyTo: '**/components/**,**/*.tsx,**/*.jsx,**/*.vue,**/*.svelte,**/*.css,**/*.scss,**/*.pptx,**/storyboard/**,.plaesy/memory/design.md,**/*.fig'
description: "Core UI/UX design principles — hierarchy, consistency, feedback, affordance, progressive disclosure, Fitts's law, information architecture, interaction/navigation design, usability validation, accessibility-first defaults — for any medium (code, storyboards, slide decks, Figma, static mockups), not only shipped code."
---

# UI/UX Design Principles

The core rule: **the interface should tell the user what's possible and what just
happened, without making them think about it.** Every principle below protects one
half of that — what's possible (affordance, hierarchy, consistency, information
architecture) or what happened (feedback) — or removes friction from acting on it
(Fitts's law, progressive disclosure, navigation design) — or verifies it actually
worked (usability validation). This file is about principles to apply *while
designing or building*; for a scored WCAG compliance audit of what's already
shipped, see `.plaesy/instructions/assess-design.md` — that file checks contrast ratios and
tab order after the fact, this one is about not creating the violation in the
first place.

**Medium-agnostic.** These principles are about the *interaction model*, not the
implementation substrate — they apply equally to a coded component, a Figma
mockup, a storyboard panel (`/create:storyboard`), or a presentation-deck prototype
(`.pptx`), not only to shipped `.tsx`/`.css`. A storyboard panel that shows a "Save"
button with no visible feedback state has the same defect as the JSX example below
— the medium changed, the usability failure didn't.

Sources (retrieved 2026-09-27): [UXPin — 14 Essential UI Design Principles](https://www.uxpin.com/studio/blog/ui-design-principles/),
[UXPin — UX Design Principles: 16 Rules](https://www.uxpin.com/studio/blog/ux-design-principles/),
[Nielsen Norman Group — 10 Usability Heuristics](https://www.nngroup.com/articles/ten-usability-heuristics/),
[Toptal — Heuristic Principles for Mobile Interfaces](https://www.toptal.com/designers/usability-testing/mobile-heuristic-principles).

## Visual hierarchy

Size, weight, color, and position should tell the user what to look at first, second,
third — deliberately, not as a side effect of whatever order elements were added in
the code. — A form where the "Cancel" button is visually heavier (larger, higher
contrast) than "Submit" guides users toward the action they didn't come to take;
hierarchy should match intent, not markup order.

## Consistency

The same action should look and behave the same way everywhere in the product. — If
"Delete" is a red button with a confirmation dialog on one screen and a plain text
link with no confirmation on another, a user who learned the safe pattern on the
first screen has no reason to expect danger on the second — consistency is what lets
learned behavior transfer.

## Feedback for every action

Every user action gets an immediate, perceptible response — success, failure, or
"in progress" — never silence. — A submit button that does nothing visible for 2
seconds while a request is in flight gets clicked again (a duplicate submission) or
abandoned (the user assumes it's broken); a spinner or disabled-state change costs one
line and prevents both.

```jsx
// BAD: no feedback — user can't tell if the click registered
<button onClick={submit}>Save</button>

// GOOD: state reflects what's actually happening
<button onClick={submit} disabled={isSaving}>
  {isSaving ? "Saving…" : "Save"}
</button>
```

## Affordance

An element's appearance should suggest what it does, before the user has to try
it. — Flat, unstyled text that's actually a clickable link (no underline, no color
difference, no cursor change) forces the user to hover-and-guess across the whole page
to find what's interactive; a clickable element should look clickable without
requiring a hover to prove it.

## Progressive disclosure

Show the common path by default; reveal advanced options only when asked for. — A
signup form with 15 fields visible at once (most of them optional edge cases) reads as
more effortful than it is; showing 3 required fields plus an "Advanced options"
toggle keeps the default path fast for the 90% case without removing capability for
the 10% that needs it.

## Fitts's law — target size and distance

The time (and error rate) to hit a target grows with how far away and how small it
is. — A 16px "delete" icon squeezed next to a 16px "edit" icon invites mis-taps,
especially on touch; on mobile specifically, keep primary touch targets at least
44×48px and give destructive actions extra spacing from adjacent safe actions, not
less.

## Accessibility-first defaults (not a full audit — a baseline)

- Never ship an interactive element with no visible **focus state** — a button, link,
  or input that only shows focus via default browser outline (or none at all) after a
  CSS reset silently breaks keyboard navigation for every user who doesn't use a
  mouse.
- Never encode meaning by color alone — a form error shown only as a red border, with
  no icon or text, is invisible to colorblind users and to anyone who can't
  distinguish the specific red used.
- Default to sufficient contrast (roughly 4.5:1 for body text) when picking a color
  pair, rather than picking colors for brand appeal first and checking contrast
  later — checking after the fact means redesigning, checking before means it was
  never wrong.

```css
/* BAD: focus state removed for a "cleaner" look — breaks keyboard nav */
button:focus { outline: none; }

/* GOOD: a visible, deliberate focus style */
button:focus-visible { outline: 2px solid #2563eb; outline-offset: 2px; }
```

## Information architecture

Group and label content the way users think about it, not the way the org chart or
database schema is structured. — A nav menu with "Products," "Solutions," and
"Platform" as three separate top-level items that all lead to the same feature set
reflects internal team boundaries, not how a visitor searches; users abandon a
search when the label they expected (e.g. "Pricing") isn't where they expected it.
Card-sort or just ask "what would a first-time user call this?" before naming a
section.

## Interaction & navigation design

A user should always know where they are, how they got there, and how to get back,
without using the browser back button as the only escape hatch. — A multi-step
wizard with no step indicator and no way to jump back to step 1 without restarting
forces users to either complete it blind or abandon it; breadcrumbs, step counters,
and a visible "exit" are not decoration, they're the user's mental map of the flow.
Every flow needs an explicit path for the user who changed their mind.

## Usability validation (closing the loop)

A design or interaction pattern is a hypothesis about how users will behave, not a
fact — validate it before treating it as settled. — Shipping a redesigned checkout
flow without comparing completion rate against the old one means a regression goes
unnoticed until support tickets pile up; even a lightweight check (5-user hallway
test, a `/create:storyboard` walkthrough reviewed against the task it's supposed to
support, or an A/B flag) catches the mismatch between "looks right" and "works
right" before it reaches everyone. State what you expect the user to be able to do
after seeing the design, and check that assumption explicitly — don't let "it looks
clean" substitute for "a new user got through it."

## Applying these under real constraints

- These principles trade off against ship speed and against each other — e.g.
  progressive disclosure adds one more interaction (the toggle) that a fully-flat form
  doesn't need. Apply proportionally to how often the advanced path is actually used,
  not by default on every form.
- When a stakeholder or brand guideline pushes against an accessibility-first default
  (e.g. a brand color that fails contrast), treat that as a real requirements conflict
  to resolve explicitly (adjust the shade, add a secondary cue) — not a principle to
  quietly drop.
- When reviewing UI code, name which principle is at stake and the concrete usability
  failure it causes (as in the examples above), not just "this feels off" — that's
  what makes the finding actionable for whoever fixes it.
