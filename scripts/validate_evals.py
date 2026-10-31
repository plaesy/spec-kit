#!/usr/bin/env python3
"""
Validate evals for a skill
"""

import json
import sys
import os

def validate_skill_evals(skill_name):
    """Validate evals for a single skill"""
    evals_path = f"skills/{skill_name}/evals/evals.json"
    results_path = f"skills/{skill_name}/evals/results.json"
    
    if not os.path.exists(evals_path):
        print(f"MISSING: {evals_path}")
        return False
    
    if not os.path.exists(results_path):
        print(f"MISSING: {results_path}")
        return False
    
    with open(evals_path) as f:
        evals = json.load(f)
    
    with open(results_path) as f:
        results = json.load(f)
    
    # Check that all tests in evals have corresponding results
    eval_tests = {t['name'] for t in evals.get('tests', [])}
    result_tests = {r['test'] for r in results.get('results', [])}
    
    missing = eval_tests - result_tests
    if missing:
        print(f"MISSING RESULTS: {skill_name}: {missing}")
        return False
    
    extra = result_tests - eval_tests
    if extra:
        print(f"EXTRA RESULTS: {skill_name}: {extra}")
        # Not an error, just warning
        print(f"WARNING: Extra results for {skill_name}: {extra}")
    
    # Check all results have status = passed
    for r in results.get('results', []):
        if r.get('status') != 'passed':
            print(f"FAILED: {skill_name}/{r['test']}: status={r.get('status')}")
            return False
    
    summary = results.get('summary', {})
    if summary.get('pass_rate', 0) != 1.0:
        print(f"NOT 100% PASS: {skill_name} pass_rate={summary.get('pass_rate')}")
        return False
    
    print(f"OK: {skill_name} ({len(results.get('results', []))} tests passed)")
    return True

def main():
    skills_dir = 'skills'
    core_skills = [
        'ideator',
        'buyer-discovery',
        'compliance-health',
        'economics-engine',
        'plan-compiler',
        'assess',
        'pipeline',
        'youtube-automation',
        'founder',
    ]
    
    all_ok = True
    for skill in core_skills:
        if not validate_skill_evals(skill):
            all_ok = False
    
    if not all_ok:
        sys.exit(1)
    
    print("\nAll evals passed!")

if __name__ == '__main__':
    main()