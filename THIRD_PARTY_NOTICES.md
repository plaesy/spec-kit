# Third-Party Notices

Some instruction files in this repository contain text adapted from the
third-party sources below. Each affected file says so under its heading. The
adapted text has been modified (IDs, cross-references, and project-specific
rules were added or changed), so the upstream files are not a substitute for
the copies here. Where this file says "modified", that is the statement the
licences require.

## taste-skill (MIT)

- Source: <https://github.com/Leonxlnx/taste-skill>
  (`skills/redesign-skill/SKILL.md`, `skills/brandkit/SKILL.md`, and the
  verification and fullness guidance in `skills/taste-skill/SKILL.md`)
- Used in:
  - `instructions/redesign.instructions.md` (modified)
  - `instructions/brandkit.instructions.md` (modified)
  - `instructions/frontend-design.instructions.md`: the sections "The idea test",
    "Fullness" and "Verify with renders" are derived from taste-skill's
    grounding, juror-note and render-verification steps (modified)
- Licence: MIT. The upstream licence text is reproduced below, as the licence requires.

```text
MIT License

Copyright (c) 2026 Leonxlnx

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## anthropics/skills, frontend-design (Apache-2.0)

- Source: <https://github.com/anthropics/skills/tree/main/skills/frontend-design>
  (`SKILL.md`; licence file `LICENSE.txt` in the same folder)
- Used in: `instructions/frontend-design.instructions.md` (modified: the verification,
  fullness and idea-test sections above are additions to the upstream text, and
  the design-principle wording was kept with the rules of this framework)
- Licence: Apache License, Version 2.0. The full licence text is in
  [`licenses/Apache-2.0-frontend-design.txt`](licenses/Apache-2.0-frontend-design.txt),
  a byte-for-byte copy of the upstream `LICENSE.txt` (git blob
  `f433b1a53f5b830a205fd2df78e2b34974656c7b`). The same text is at
  <https://www.apache.org/licenses/LICENSE-2.0>.
- Copyright holder: the upstream `LICENSE.txt` names no copyright holder. Its
  appendix is the standard Apache-2.0 template, so no holder line is copied here,
  and none is invented. Anthropic is the upstream publisher of the repository.
- Modification statement: this framework changed the upstream text, as noted above.
  Those changes are marked in `instructions/frontend-design.instructions.md` by the
  "Adapted from" line under its heading.
