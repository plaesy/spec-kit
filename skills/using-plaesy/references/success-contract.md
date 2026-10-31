---
name: success-contract
description: The task checklist rule. Every task keeps a contracts.json checklist in its own folder. plaesy tasks validate runs each item and moves the task to done when every item is passed. Load it before starting any task and before claiming completion.
---

# Success Contract

Every task has a checklist: the conditions that say the work is done. The checklist is `contracts.json` in the task's folder, and `plaesy tasks validate` runs it.

> More see: `.plaesy/instructions/tasks.md`

## 1. Where it lives

```text
.plaesy/tasks/<status>/<task>/
├── spec.md         what the task is for
├── plan.md         planning
└── contracts.json  the checklist
```

The checklist sits in the task folder, next to the task. `<status>` is the folder the task sits in: `backlog`, `todo`, `doing`, `done`, or `blocked`. `blocked` is only for a task waiting on something outside the work (another task, an external service); an unfinished task stays in `doing`.

## 2. The file

```json
{
  "id": "implement-login-form",
  "checklist": [
    {
      "id": "C1",
      "check": "login tests pass",
      "run": "cd app && go test ./login -count=1",
      "expected": "pass",
      "status": "failed",
      "histories": [
        {"at": "2026-10-22T10:15:00Z", "actual": "exit code 1", "run": "cd app && go test ./login -count=1"}
      ]
    }
  ]
}
```

- `checklist`: the items of the task, in order.
- Each item has an `id`, a `check` (what is proven), a `run` (a shell command, run from the project root), and an `expected` (what the command's last output line must reach).
- `status` is one of:
  - `pending`: never run to a pass yet.
  - `passed`: the last run passed.
  - `failed`: the last run failed. The reason is in `histories`.
- `histories`: every failed run, with `at` (time), `actual` (why it failed), and `run` (the command that ran). A pass does not erase earlier failures.

## 3. Expected values

| Expected | The run passes when |
|---|---|
| `pass` | the command exits 0 |
| `zero`, `none` | the output's last line is `0` or `none` |
| `yes`, `present`, `absent` | the last line is the matching word |
| `all`, `every` | the last line is `N of N` |
| `>= 9`, `<= 2 s`, `>= 80%` | the last line is a number that meets the comparison, with the same unit |

A run must print its measurement as its last line, for example `wc -l < file`, not `wc -l file`.

## 4. Running it

```bash
plaesy tasks validate                         # every task with a contracts.json
plaesy tasks validate implement-login-form    # one task
```

Each item's `run` is executed. A passing run sets the item to `passed`. A failing run sets it to `failed` and adds a history entry. When every item is `passed`, the task moves to `done`. A task that is not finished is never moved to `done`.

## 5. Rules

1. Write the checklist before the work, so the expected values are set before the result is known. Change an expected value only by writing down why, in the task's `spec.md`.
2. Do not report a task complete while any item is not `passed`. This is constitution NN-08.
3. An item is passed only by its `run`. Do not set `status` by hand.
4. Do not delete `histories`. They are the record of what failed and when.

---

## Spec-Kit Pipeline Success Contracts

These contract templates apply to each stage of the spec-kit business validation pipeline. Use them when creating tasks for pipeline stages.

### Contract Template: Idea Discovery (Stage 1)

```json
{
  "id": "ideator-<topic>",
  "checklist": [
    {"id": "C1", "check": "Generated 8-12 distinct options", "run": "jq '.options | length' output.json", "expected": ">= 8"},
    {"id": "C2", "check": "Each option has name, description, assumption, cheapest_test, label", "run": "jq '.options[] | has(\"name\") and has(\"description\") and has(\"assumption\") and has(\"cheapest_test\") and has(\"label\")' output.json", "expected": "all"},
    {"id": "C3", "check": "Labels are FACT/INFERRED/SPECULATIVE only", "run": "jq '.options[].label | IN(\"FACT\",\"INFERRED\",\"SPECULATIVE\")' output.json", "expected": "all"},
    {"id": "C4", "check": "At least one contrarian option", "run": "jq '.options[] | select(.description | test(\"contrarian|against|opposite\";\"i\")) | length' output.json", "expected": ">= 1"},
    {"id": "C5", "check": "No duplicate assumptions", "run": "jq '.options | map(.assumption) | unique | length == length' output.json", "expected": "true"},
    {"id": "C6", "check": "Candidate files written to .plaesy/ideas/<topic>/idea.md", "run": "ls .plaesy/ideas/*/idea.md | wc -l", "expected": ">= 8"},
    {"id": "C7", "check": "Index updated in .plaesy/ideas.md", "run": "grep -c 'status: candidate' .plaesy/ideas.md", "expected": ">= 8"}
  ]
}
```

### Contract Template: Buyer Discovery (Stage 5)

```json
{
  "id": "buyer-discovery-<topic>",
  "checklist": [
    {"id": "C1", "check": "≥5 real conversations recorded", "run": "jq '.conversation_count' buyers-summary.json", "expected": ">= 5"},
    {"id": "C2", "check": "≥3 'would pay YES'", "run": "jq '.would_pay_yes' buyers-summary.json", "expected": ">= 3"},
    {"id": "C3", "check": "Weighted score ≥ 60%", "run": "jq '.weighted_score' buyers-summary.json", "expected": ">= 0.6"},
    {"id": "C4", "check": "All conversations have date, role, questions, answers, would_pay, price_point", "run": "jq '.conversations[] | has(\"date\") and has(\"role\") and has(\"questions\") and has(\"answers\") and has(\"would_pay\") and has(\"price_point\")' buyers-real.md", "expected": "all"},
    {"id": "C5", "check": "Segment Validation Matrix completed", "run": "jq '.segments | keys | length' segment-matrix.json", "expected": ">= 3"},
    {"id": "C6", "check": "buyers.md output in founder format", "run": "grep -c '| Date | Person' buyers.md", "expected": ">= 1"}
  ]
}
```

### Contract Template: Economics Verdict (Stage 9)

```json
{
  "id": "economics-engine-<topic>",
  "checklist": [
    {"id": "C1", "check": "All 4 inputs present (P, V, F, S)", "run": "jq 'has(\"P\") and has(\"V\") and has(\"F\") and has(\"S\")' economics-inputs.json", "expected": "true"},
    {"id": "C2", "check": "S has stated basis", "run": "jq '.S_basis | length > 0' economics-inputs.json", "expected": "true"},
    {"id": "C3", "check": "Margin calculated: m = P - V", "run": "jq '.m == (.P - .V)' economics-output.json", "expected": "true"},
    {"id": "C4", "check": "Break-even: B = ceil(F / m)", "run": "jq '.B == ((.F / .m) | ceil)' economics-output.json", "expected": "true"},
    {"id": "C5", "check": "Verdict follows rules: Incomplete | Not yet | Profitable", "run": "jq '.verdict | IN(\"Incomplete\",\"Not yet\",\"Profitable\")' economics-output.json", "expected": "true"},
    {"id": "C6", "check": "If Not yet: failing number named", "run": "jq '.failing_number | length > 0' economics-output.json", "expected": "true"},
    {"id": "C7", "check": "If Profitable: margin of safety stated", "run": "jq '.margin_of_safety' economics-output.json", "expected": "present"},
    {"id": "C8", "check": "Sensitivity analysis run on price, cost, volume, CAC, fixed_cost, segment_mix", "run": "jq '.sensitivity.scenarios | length' economics-output.json", "expected": ">= 6"},
    {"id": "C9", "check": "economics.md written with all sections", "run": "grep -c 'Price P =' .plaesy/ideas/<topic>/economics.md", "expected": ">= 1"}
  ]
}
```

### Contract Template: Pivot Decision Gate (Stage 9→10)

```json
{
  "id": "pivot-gate-<topic>",
  "checklist": [
    {"id": "C1", "check": "Economics verdict is Not yet (gate triggered)", "run": "jq '.verdict' economics-output.json", "expected": "\"Not yet\""},
    {"id": "C2", "check": "Segment analysis completed", "run": "jq 'has(\"segment_analysis\")' pivot-options.json", "expected": "true"},
    {"id": "C3", "check": "Segment Validation Matrix run", "run": "jq 'has(\"segment_matrix\")' pivot-options.json", "expected": "true"},
    {"id": "C4", "check": "4 pivot options presented with projected verdicts", "run": "jq '.options | length' pivot-options.json", "expected": "4"},
    {"id": "C5", "check": "User decision documented in plan.md", "run": "grep -c '## Pivot Analysis' .plaesy/ideas/<topic>/plan.md", "expected": ">= 1"},
    {"id": "C6", "check": "Chosen pivot has validation requirements listed", "run": "grep -A5 'Validation Required' .plaesy/ideas/<topic>/plan.md | grep -c 'validation'", "expected": ">= 1"}
  ]
}
```

### Contract Template: Plan Compilation (Stage 14)

```json
{
  "id": "plan-compiler-<topic>",
  "checklist": [
    {"id": "C1", "check": "All 11 required stage files exist", "run": "ls .plaesy/ideas/<topic>/{idea,board,competitors,buyers,pricing,offer,economics,marketing,brand,ops,launch}.md | wc -l", "expected": "11"},
    {"id": "C2", "check": "Verdict at top of plan.md", "run": "grep -c '\\*\\*Verdict:\\*\\*' .plaesy/ideas/<topic>/plan.md", "expected": "1"},
    {"id": "C3", "check": "Driving numbers at top (margin, B, S, MoS)", "run": "grep -c 'Margin.*Break-even.*Expected.*MoS' .plaesy/ideas/<topic>/plan.md", "expected": "1"},
    {"id": "C4", "check": "Board votes summary included", "run": "grep -c 'Board votes' .plaesy/ideas/<topic>/plan.md", "expected": "1"},
    {"id": "C5", "check": "One-page summary generated", "run": "grep -c '## Summary' .plaesy/ideas/<topic>/plan.md", "expected": "1"},
    {"id": "C6", "check": "All stage files concatenated in order", "run": "grep -c '### ' .plaesy/ideas/<topic>/plan.md", "expected": "11"},
    {"id": "C7", "check": "If verdict=Not yet: Pivot Analysis section exists", "run": "grep -c '## Pivot Analysis' .plaesy/ideas/<topic>/plan.md", "expected": "1"},
    {"id": "C8", "check": "Index updated in .plaesy/ideas.md with verdict + date", "run": "grep '<topic>' .plaesy/ideas.md | grep -c 'verdict:'", "expected": "1"}
  ]
}
```

### Contract Template: Assessment (Stage 15)

```json
{
  "id": "assess-<topic>-<dimension>",
  "checklist": [
    {"id": "C1", "check": "Grade is A+/A/B+/B/C/D", "run": "jq '.grade | IN(\"A+\",\"A\",\"B+\",\"B\",\"C\",\"D\")' assessment-output.json", "expected": "true"},
    {"id": "C2", "check": "Score 0-100", "run": "jq '.score >= 0 and .score <= 100' assessment-output.json", "expected": "true"},
    {"id": "C3", "check": "All component scores present with weights", "run": "jq '.component_scores | length' assessment-output.json", "expected": "5"},
    {"id": "C4", "check": "Top 5 findings with sources", "run": "jq '.top_5_findings | length' assessment-output.json", "expected": "5"},
    {"id": "C5", "check": "Recommendations actionable", "run": "jq '.recommendations | length' assessment-output.json", "expected": ">= 1"},
    {"id": "C6", "check": "Confidence basis is HIGH/MEDIUM/LOW", "run": "jq '.confidence_basis | IN(\"HIGH\",\"MEDIUM\",\"LOW\")' assessment-output.json", "expected": "true"},
    {"id": "C7", "check": "Next phase command specified", "run": "jq '.next_phase | length > 0' assessment-output.json", "expected": "true"},
    {"id": "C8", "check": "Output format has all canonical sections", "run": "grep -c '## Overall Score\\|## Category Scores\\|## Findings\\|## Recommendations\\|## Confidence Basis\\|## Next Phase' assessment-<dimension>.md", "expected": "6"}
  ]
}
```

### Contract Template: Full Pipeline

```json
{
  "id": "pipeline-<topic>",
  "checklist": [
    {"id": "C1", "check": "All 15 stages completed", "run": "jq '.stages_completed' pipeline-output.json", "expected": "15"},
    {"id": "C2", "check": "Verdict produced: Incomplete | Not yet | Profitable", "run": "jq '.verdict | IN(\"Incomplete\",\"Not yet\",\"Profitable\")' pipeline-output.json", "expected": "true"},
    {"id": "C3", "check": "plan.md exists with verdict", "run": "test -f .plaesy/ideas/<topic>/plan.md && grep -q 'Verdict' .plaesy/ideas/<topic>/plan.md", "expected": "pass"},
    {"id": "C4", "check": "ideas.md index updated", "run": "grep '<topic>' .plaesy/ideas.md | grep -q 'verdict:'", "expected": "pass"},
    {"id": "C5", "check": "If Not yet: Pivot Analysis documented", "run": "jq '.verdict' pipeline-output.json | grep -q 'Not yet' && grep -q '## Pivot Analysis' .plaesy/ideas/<topic>/plan.md", "expected": "pass"},
    {"id": "C6", "check": "All Spec-Kit gates passed (S1-S9)", "run": "jq '.spec_kit_gates | all(.status == \"PASS\")' pipeline-output.json", "expected": "true"},
    {"id": "C7", "check": "Assessments run for business, marketing, financial", "run": "jq '.assessments_run' pipeline-output.json", "expected": "3"},
    {"id": "C8", "check": "YouTube Automation completed if invoked", "run": "jq '.youtube_automation.completed' pipeline-output.json", "expected": "true"}
  ]
}
```

### Contract Template: YouTube Automation (Stage 16)

```json
{
  "id": "youtube-automation-<topic>",
  "checklist": [
    {"id": "C1", "check": "Research sources >= 3 with URLs", "run": "jq '.research_sources | length' youtube-output.json", "expected": ">= 3"},
    {"id": "C2", "check": "All script claims grounded or marked assumption", "run": "jq '.fact_check.grounded' youtube-output.json", "expected": "true"},
    {"id": "C3", "check": "Script has hook, body, CTA", "run": "jq '.script.has_hook and .script.has_body and .script.has_cta' youtube-output.json", "expected": "true"},
    {"id": "C4", "check": "Visuals count >= 5 with credits", "run": "jq '.visuals | length' youtube-output.json", "expected": ">= 5"},
    {"id": "C5", "check": "Voice SRT captions synced", "run": "jq '.voice.srt_synced' youtube-output.json", "expected": "true"},
    {"id": "C6", "check": "Render valid MP4, duration within 10%", "run": "jq '.render.valid and (.render.duration_ratio >= 0.9 and .render.duration_ratio <= 1.1)' youtube-output.json", "expected": "true"},
    {"id": "C7", "check": "Findability Score >= 80", "run": "jq '.findability_score >= 80' youtube-output.json", "expected": "true"},
    {"id": "C8", "check": "Thumbnail < 2MB, correct aspect ratio", "run": "jq '.thumbnail.size_mb < 2 and (.thumbnail.aspect_ratio == \"16:9\" or .thumbnail.aspect_ratio == \"9:16\")' youtube-output.json", "expected": "true"},
    {"id": "C9", "check": "At least 1 thumbnail style generated", "run": "jq '.thumbnail.styles | length' youtube-output.json", "expected": ">= 1"},
    {"id": "C10", "check": "Output files exist (video.mp4, thumbnail.jpg, captions.srt, script.md, metadata.json)", "run": "ls output/library/<video-id>/{video.mp4,thumbnail.jpg,captions.srt,script.md,metadata.json} | wc -l", "expected": "5"}
  ]
}
```