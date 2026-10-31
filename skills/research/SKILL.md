---
name: research
description: >-
  Research a question with cited sources before a decision: find the sources, open
  them, record each claim with its URL and access date, and label anything
  unverified. Use when the user asks to research, compare options, check facts,
  find current prices or versions, or when a spec, plan, or decision needs
  evidence.
---

# research

Answer one question with evidence the reader can check. Research feeds a decision,
so each claim must be traceable to a source and dated.

## Rules

- Every external claim has a source URL and the date it was accessed:
  `[claim] — Source: [URL], accessed YYYY-MM-DD`.
- Open the source before citing it. A search snippet alone is not a source.
- A claim you cannot verify is labeled `UNKNOWN` (no answer found) or `ASSUMED`
  (a default used on purpose). Never present it as fact.
- Prefer primary sources: the vendor's documentation, the specification, the
  regulator, the paper. Use secondary sources (blogs, aggregators) only when no
  primary source exists, and say so.
- Dates matter. For prices, versions, and laws, record the date the page was read,
  and say when the source was last updated if it shows that.
- Research does not decide. It lays out the options, their trade-offs, and the
  evidence. The decision is made by the person or the plan that uses it.

## Workflow

1. **State the question** in one sentence, and list what the answer must decide.
2. **Search** with at least three different queries: the topic, the vendor or
   standard name, and the specific claim. Note the sources that look primary.
3. **Open and read** the best sources. Record the claims that answer the question.
4. **Check the claims** against a second source where the claim matters. Note
   disagreements instead of picking one silently.
5. **Write the findings** in the format below.
6. **Report** the answer, the strength of the evidence, and what is still unknown.

## Output

```markdown
# Research: <question>

Question: <one sentence>
Decision it informs: <spec, plan, or choice>

## Findings

- <claim> — Source: <URL>, accessed <YYYY-MM-DD>
- <claim> — Source: <URL>, accessed <YYYY-MM-DD> (secondary; no primary source found)

## Options compared

| Option | Evidence for | Evidence against | Cost or limit |
| --- | --- | --- | --- |

## Unknown and assumed

- UNKNOWN: <question>
- ASSUMED — <default>: <why>
```
