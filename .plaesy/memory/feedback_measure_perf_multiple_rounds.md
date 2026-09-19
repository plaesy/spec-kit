---
name: feedback_measure_perf_multiple_rounds
description: A single before/after timing run overstates performance wins in this repo — always take multiple back-to-back rounds
metadata:
  type: feedback
---

Always measure performance with multiple back-to-back rounds (before/after alternating,
not just one run each) before reporting a percentage improvement in an `/optimize`
report.

**Why**: During [[perf_detect_stack_extension_scan_2026_09_16]] a single cold-cache
timing (5.1s → 3.35s, bash detect-stack.sh) was reported as "~34% faster." Re-measured
with 5 alternating warm-cache rounds, the real difference was inside noise (~1.1-1.5s
both ways) — the first number was cache-warming artifact, not the optimization's effect.
The PowerShell side of the same change *did* hold up under repeated measurement (~50-60%
faster, consistent across 4 rounds), so the lesson isn't "don't trust timing," it's
"one sample isn't enough to tell noise from a real win."

**How to apply**: Before writing a completion report with a % improvement number, run
before and after at least 3-5 times each, interleaved, on a warm cache. If the ranges
overlap, report "no measurable difference at this scale" rather than picking the two
most favorable numbers. Correcting an already-sent overstated claim (like this session
did) is better than leaving it standing, but getting it right the first time is better
still.
