#!/usr/bin/env python3
"""Compact command output read from stdin, keeping every line that matters.

Usage:
    <command> 2>&1 | python scripts/compact_output.py [--max-run 12] [--keep-edge 20]

What it does, in order:
  1. Removes ANSI colour and cursor codes.
  2. Collapses runs of the same line into one line with a count.
  3. Collapses a long run of passing-test lines (ok, PASS, ✓) into one count line.
  4. Keeps every line that mentions error, fail, warn, panic, or exception.
  5. When the output is still longer than --keep-edge * 2 lines, keeps the first and
     last --keep-edge lines and replaces the middle with a count.

It never edits a line's text. It only drops lines and merges repeats, and it says how
many lines it removed on stderr.
"""

import argparse
import re
import sys

ANSI = re.compile(r"\x1b\[[0-9;?]*[A-Za-z]|\x1b\][^\x07]*\x07")
# A prefix match, not a whole word: "errors", "failures:", "panicked" and "warnings" must match too.
KEEP = re.compile(r"\b(err\w*|fail\w*|warn\w*|panic\w*|exception\w*|traceback|fatal)", re.I)
PASS = re.compile(r"(^|\s)(ok|PASS|passed|✓|✔)(\s|$|:)", re.I)


def compact(lines, max_run, keep_edge):
    # 1 and 2: strip codes, merge identical adjacent lines.
    merged = []
    for raw in lines:
        line = ANSI.sub("", raw.rstrip("\n"))
        if merged and merged[-1][0] == line:
            merged[-1][1] += 1
        else:
            merged.append([line, 1])

    # 3: collapse long runs of passing lines. 4: never drop a line that matches KEEP.
    out = []
    pass_run = []
    for line, count in merged:
        is_pass = bool(PASS.search(line)) and not KEEP.search(line)
        if is_pass:
            pass_run.append((line, count))
            continue
        if pass_run:
            out.extend(_flush_pass(pass_run, max_run))
            pass_run = []
        out.append((line, count))
    if pass_run:
        out.extend(_flush_pass(pass_run, max_run))

    text = [line if count == 1 else f"{line}  [x{count}]" for line, count in out]

    # 5: head and tail for very long output; everything with KEEP is kept inside the middle.
    if len(text) > keep_edge * 2:
        head, middle, tail = text[:keep_edge], text[keep_edge:-keep_edge], text[-keep_edge:]
        kept = [l for l in middle if KEEP.search(l)]
        removed = len(middle) - len(kept)
        return head + [f"... {removed} line(s) elided, {len(kept)} matching line(s) kept ..."] + kept + tail
    return text


def _flush_pass(run, max_run):
    if len(run) <= max_run:
        return run
    total = sum(count for _, count in run)
    first = run[0][0]
    return [(f"{first}  [... {len(run)} passing lines, {total} occurrences collapsed ...]", 1)]


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--max-run", type=int, default=12, help="collapse runs of passing lines longer than this")
    parser.add_argument("--keep-edge", type=int, default=20, help="lines kept at each end of long output")
    args = parser.parse_args()

    source = sys.stdin.read().splitlines(keepends=True)
    result = compact(source, args.max_run, args.keep_edge)
    sys.stdout.write("\n".join(result) + "\n")
    before = sum(len(l) for l in source)
    after = sum(len(l) + 1 for l in result)
    sys.stderr.write(f"compact_output: {len(source)} -> {len(result)} lines, {before} -> {after} bytes\n")


if __name__ == "__main__":
    main()
