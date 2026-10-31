#!/usr/bin/env python3
"""Run one skill's evals headlessly and grade each expected item.

Usage:
    python scripts/run_skill_evals.py skills/spec [--case ID] [--runner kilo|claude] [--model MODEL]

For each case in <skill>/evals/evals.json:
  1. Runs `runner (kilo by default, free model; or claude -p) in a fresh scratch folder, with
     the skill folder named in the prompt, and asks it to follow the skill for the case query.
  2. Lists and reads the files the run wrote in the scratch folder.
  3. Runs a second `claude -p` as a grader that checks each expected item against
     the run's reply and files, and returns JSON.

Results are written to <skill>/evals/results.json. Only the standard library is used.
A case passes only when every expected item is judged pass by the grader. A case
whose run or grading fails to parse is recorded as ERROR, never as pass.
"""

import argparse
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from datetime import date
from pathlib import Path

MAX_FILE_CHARS = 20000


def run_model(prompt, cwd, add_dir, runner, model, timeout):
    """Sends one prompt to the chosen runner and returns its text reply."""
    if runner == "kilo":
        exe = shutil.which("kilo") or shutil.which("kilo.cmd")
        if not exe:
            raise RuntimeError("kilo is not on PATH")
        # Multi-line prompts break through the Windows .cmd shim, so the prompt goes in a file.
        prompt_file = Path(cwd) / ".eval-prompt.md"
        prompt_file.write_text(prompt, encoding="utf-8")
        # The message comes before -f, because -f takes every following argument as a file.
        cmd = [exe, "run", "-m", model, "--format", "default",
               "Follow the attached instructions exactly. Reply with the requested output only.",
               "-f", str(prompt_file)]
        # kilo run has no add-dir flag; the skill folder is named by absolute path in the prompt.
    else:
        cmd = [
            "claude", "-p",
            "--permission-mode", "acceptEdits",
            "--allowedTools", "WebSearch", "WebFetch",
            "--add-dir", str(add_dir),
            "--output-format", "text",
        ]
        if model:
            cmd += ["--model", model]
        # The prompt goes on stdin: long grader prompts exceed the Windows command-line limit.
        proc = subprocess.run(
            cmd, cwd=cwd, input=prompt, capture_output=True, text=True, encoding="utf-8",
            errors="replace", timeout=timeout,
        )
        if proc.returncode != 0:
            raise RuntimeError(f"{runner} exited {proc.returncode}: {proc.stderr.strip()[:500]}")
        return proc.stdout
    proc = subprocess.run(
        cmd, cwd=cwd, capture_output=True, text=True, encoding="utf-8",
        errors="replace", timeout=timeout,
    )
    if proc.returncode != 0:
        raise RuntimeError(f"{runner} exited {proc.returncode}: {proc.stderr.strip()[:500]}")
    return proc.stdout


def list_files(root):
    found = []
    for path in sorted(Path(root).rglob("*")):
        if path.is_file():
            found.append(path.relative_to(root).as_posix())
    return found


def read_files(root, names):
    parts = []
    for name in names:
        text = (Path(root) / name).read_text(encoding="utf-8", errors="replace")
        if len(text) > MAX_FILE_CHARS:
            text = text[:MAX_FILE_CHARS] + "\n[truncated]"
        parts.append(f"--- FILE: {name} ---\n{text}")
    return "\n\n".join(parts)


FILE_BLOCK = re.compile(r"<<<FILE (?P<path>[^>\n]+)>>>\n(?P<body>.*?)\n?<<<END>>>", re.S)


def write_file_blocks(reply, root):
    """Writes each <<<FILE path>>> block in the reply under root. Paths outside root are refused."""
    root = Path(root).resolve()
    for m in FILE_BLOCK.finditer(reply):
        target = (root / m.group("path").strip()).resolve()
        if root not in target.parents:
            continue
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(m.group("body"), encoding="utf-8")


def run_prompt(skill_dir, query):
    skill_md = (skill_dir / "SKILL.md").as_posix()
    return (
        f"Follow the skill defined in {skill_md}. Read that file first, then follow it "
        "exactly for the task below. This is a headless run with no file tools to rely on: "
        "give every file the skill asks for as its own block in your reply, in this form:\n"
        "<<<FILE path/relative/to/folder>>>\n<full file content>\n<<<END>>>\n"
        "Use the paths the skill gives. Do not show file contents in any other form than these "
        "blocks. The runner writes each block to disk. You cannot ask "
        "the user questions interactively, so list any questions after all the file blocks.\n\n"
        "Task:\n" + query
    )


def grade_prompt(case, reply, files_text):
    expected = "\n".join(f"{i + 1}. {item}" for i, item in enumerate(case["expected"]))
    return (
        "You are grading an agent run against expected behaviour. Judge only what the reply "
        "and the files show. If something is not shown, it is not met.\n\n"
        f"Task given to the agent:\n{case['query']}\n\n"
        f"Expected behaviour (judge each item separately):\n{expected}\n\n"
        f"Agent reply:\n{reply}\n\n"
        f"Files written by the agent:\n{files_text or '(none)'}\n\n"
        "Return only JSON, with no prose around it, in this shape:\n"
        '{"items": [{"expected": "<item text>", "pass": true, "reason": "<one line>"}]}\n'
    )


def parse_grade(text):
    match = re.search(r"\{.*\}", text, re.S)
    if not match:
        raise ValueError("grader returned no JSON: " + " ".join(text.split())[:300])
    data = json.loads(match.group(0))
    items = data["items"]
    return [{"expected": i["expected"], "pass": bool(i["pass"]), "reason": i.get("reason", "")} for i in items]


def run_case(skill_dir, case, runner, model, timeout):
    scratch = Path(tempfile.mkdtemp(prefix=f"skill-eval-{case['id']}-"))
    try:
        reply = run_model(run_prompt(skill_dir, case["query"]), scratch, skill_dir, runner, model, timeout)
        write_file_blocks(reply, scratch)
        files = [f for f in list_files(scratch) if f != ".eval-prompt.md"]
        files_text = read_files(scratch, files) if files else ""
        grade_text = run_model(
            grade_prompt(case, reply, files_text), scratch, skill_dir, runner, model, timeout
        )
        items = parse_grade(grade_text)
        status = "PASS" if items and all(i["pass"] for i in items) else "FAIL"
        # A case whose expected items say the agent writes files must have written files.
        # Content in the reply alone is not a file, so the case cannot pass without one.
        if not files and any(re.search(r"\bWrites?\b", e, re.I) and not re.search(r"\bnot\b|\bdoes not\b|\bdon't\b", e, re.I)
                             for e in case["expected"]):
            status = "FAIL"
            items.append({"expected": "files written to disk", "pass": False,
                          "reason": "no file was written; the reply contained content only"})
        return {"id": case["id"], "status": status, "items": items, "files": files, "reply": reply[-2000:], "files_text": files_text[:8000]}
    except Exception as exc:  # recorded, never counted as pass
        return {"id": case["id"], "status": "ERROR", "error": str(exc)}
    finally:
        shutil.rmtree(scratch, ignore_errors=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("skill", help="skill folder, for example skills/spec")
    parser.add_argument("--case", help="run only this case id")
    parser.add_argument("--runner", choices=["claude", "kilo"], default="kilo")
    parser.add_argument("--model", default=None, help="default: kilo/kilo-auto/free for kilo, claude default otherwise")
    parser.add_argument("--timeout", type=int, default=900, help="seconds per claude call")
    args = parser.parse_args()

    skill_dir = Path(args.skill).resolve()
    evals_file = skill_dir / "evals" / "evals.json"
    if not evals_file.exists():
        print(f"no evals file at {evals_file}", file=sys.stderr)
        return 2
    cases = json.loads(evals_file.read_text(encoding="utf-8"))["cases"]
    if args.case:
        cases = [c for c in cases if c["id"] == args.case]
        if not cases:
            print(f"no case {args.case}", file=sys.stderr)
            return 2

    model = args.model or ("kilo/kilo-auto/free" if args.runner == "kilo" else None)
    results = []
    for case in cases:
        print(f"running {case['id']} ...", flush=True)
        result = run_case(skill_dir, case, args.runner, model, args.timeout)
        print(f"  {result['status']}", flush=True)
        results.append(result)

    # Keep results of cases that were not run this time; replace only the ones that ran.
    results_file = skill_dir / "evals" / "results.json"
    kept = []
    if results_file.exists():
        previous = json.loads(results_file.read_text(encoding="utf-8")).get("results", [])
        ran = {r["id"] for r in results}
        kept = [r for r in previous if r["id"] not in ran]
    out = {
        "skill": skill_dir.name,
        "run_date": date.today().isoformat(),
        "runner": args.runner,
        "model": model or "default",
        "results": kept + results,
    }
    results_file.write_text(json.dumps(out, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    failed = [r for r in results if r["status"] != "PASS"]
    print(f"{len(results) - len(failed)}/{len(results)} passed; results in {skill_dir / 'evals' / 'results.json'}")
    return 0 if not failed else 1


if __name__ == "__main__":
    sys.exit(main())
