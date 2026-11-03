import json
import subprocess
import sys
from pathlib import Path


def run_case(case_id):
    result = subprocess.run(
        [sys.executable, "scripts/run_deterministic_evals.py", "founder", "--case", case_id],
        capture_output=True, text=True, check=False
    )
    return result


def test_profit_case_is_deterministic_and_passes():
    result = run_case("profitable-with-basis")
    assert result.returncode == 0, result.stdout + result.stderr
    assert "PASS: profitable-with-basis" in result.stdout


def test_negative_margin_has_no_meaningful_break_even():
    result = run_case("not-yet-negative-margin")
    assert result.returncode == 0, result.stdout + result.stderr
    assert "PASS: not-yet-negative-margin" in result.stdout


def test_all_founder_cases_run_without_model():
    result = subprocess.run(
        [sys.executable, "scripts/run_deterministic_evals.py", "founder"],
        capture_output=True, text=True, check=False
    )
    assert result.returncode == 0, result.stdout + result.stderr
    assert "6/6 passed" in result.stdout


def test_results_use_legacy_compatible_test_field():
    results = json.loads(Path("skills/founder/evals/results.json").read_text())
    ids = {item.get("test", item.get("id")) for item in results["results"]}
    assert "profitable-with-basis" in ids
