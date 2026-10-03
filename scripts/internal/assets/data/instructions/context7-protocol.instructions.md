---
description: "Single source for the Context7 documentation-lookup protocol - when it is mandatory, the exact tool sequence, the required citation form, and the fallback when it is unavailable"
applyTo: "**/*"
---

# Context7 Protocol

Context7 resolves a library ID to current, version-matched documentation. It is
the authority for "what does this API actually look like in the version we are
installing" — the question general model knowledge answers confidently and
sometimes wrongly.

This file is the only place the protocol is stated. Commands that require it
(`/implement`, `/fix`) point here rather than restating the sequence, because a
restated sequence drifts: one command ends up citing, another does not, and a
reader cannot tell which rule is current.

## When It Is Mandatory

| Situation | Required |
|-----------|----------|
| Adding or upgrading any third-party dependency | yes |
| Using an API whose signature may have changed since training | yes |
| Answering "is this library deprecated / what is the current version" | yes |
| Editing only the project's own code, no new dependency | no |
| Reading a framework config file whose format is not in doubt | no |
| The library is in the "Always Available" table below | no — cite the primary source directly |

"Not mandatory" means *not required*, not *forbidden*. When a lookup is cheap and
the answer will be acted on, do it.

## The Sequence

1. `mcp__context7__resolve-library-id` for **every** technology involved — one
   call per library, not per project. A stack of three technologies is three
   resolve calls.
2. `mcp__context7__get-library-docs` (tokens=500) against each resolved ID.
3. Cite the result inline, in the form below.
4. On failure, follow the fallback chain. Never silently proceed as if the
   lookup had succeeded.

## Required Citation Form

```text
Implementation based on Context7 (/library/id) — Retrieved {{CURRENT_DATE}}
```

The citation carries three obligations: the resolved library ID (so a reader can
re-resolve it), the retrieval date (so a stale answer is visible), and the word
Context7 (so it is distinguishable from a general-knowledge claim). A finding,
decision, or implementation line that used Context7 and carries no citation is
unverifiable and should be treated as uncited.

## Fallback Chain

Context7 being unavailable degrades the *sourcing*, never the requirement to
source. Work down this list and record which rung you landed on:

1. **Local cache** — `~/.context7/cache`, if populated. Say that you used a
   cached copy and give its date; a cache entry from six months ago is not a
   current answer and must not be presented as one.
2. **Primary documentation** — the library's own docs, or its repository. Cite
   the URL. This is the preferred fallback: it is the same authority Context7
   indexes.
3. **The fallback tables** in `.plaesy/instructions/tech-validation.md`, which map a
   technology to a maintained source when Context7 is down.
4. **General knowledge** — acceptable only for stable, unversioned facts, and
   it must be labelled as general knowledge rather than presented as verified.
   Anything version-dependent stops here: if the answer depends on which version
   is current, you do not know it, and the honest output is the question, not a
   guess.
5. **Apply the most conservative documented default and label it** — when the
   answer changes what gets built but no source above resolved it, pick the
   behavior most consistent with the library's last known-stable release,
   label it `ASSUMED — <default>, unverifiable at Context7 fallback rung 5`,
   and record it in `.plaesy/context.md` so the next session (or an
   `ASSUMED`-tag audit) can revisit it. This stops only for a rule-8 hard-stop
   category (`.plaesy/instructions/plaesy.md`) — e.g. the unresolved API
   behavior governs a destructive/irreversible operation.

## What Context7 Is Not For

- **Not a licence to skip reading the code.** Context7 documents the library as
  published; your installed version may be patched, vendored, or wrapped. Read
  what you actually depend on.
- **Not a substitute for the project's own instructions.** Stack conventions,
  testing strategy, and the constitution's thresholds live in
  `.plaesy/memory/constitution.md` and the `*-design-principles` files, and
  Context7 says nothing about them.
- **Not applicable to private or internal packages.** There is no public ID to
  resolve; read the package's own source and types instead.

## Always Available (No Lookup Needed)

| Standard | Source |
|----------|--------|
| **OWASP Top 10** | `https://owasp.org/www-project-top-ten/` |
| **WCAG 2.1** | `https://www.w3.org/WAI/WCAG21/quickref/` |
| **REST API Best Practices** | `https://restfulapi.net` |
| **GraphQL Spec** | `https://spec.graphql.org` |
| **Semantic Versioning** | `https://semver.org` |

These are specifications, not libraries: they have no version to resolve and no
release cycle, so a Context7 call would return nothing. Cite them directly.
