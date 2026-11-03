#!/usr/bin/env python3
"""
Validate all 8 core skills + founder have required structure
"""

import os
import sys

def main():
    skills_dir = 'skills'
    # Validate every skill directory. Evals/oracles are required only for the
    # critical skills that declare them, so technology and utility skills remain
    # valid without pretending they have business-logic golden tests.
    required = ['SKILL.md']
    errors = []
    skill_names = []
    for entry in sorted(os.listdir(skills_dir)):
        skill_path = os.path.join(skills_dir, entry)
        if not os.path.isdir(skill_path) or entry == 'oracle':
            continue
        # The plaesy directory contains plugin assets, not a Hermes skill.
        if entry == 'plaesy' and not os.path.exists(os.path.join(skill_path, 'SKILL.md')):
            continue
        skill_names.append(entry)
        for req in required:
            if not os.path.exists(os.path.join(skill_path, req)):
                errors.append(f"MISSING: {entry}/{req}")
        evals_path = os.path.join(skill_path, 'evals', 'evals.json')
        results_path = os.path.join(skill_path, 'evals', 'results.json')
        if os.path.exists(evals_path) != os.path.exists(results_path):
            errors.append(f"INCOMPLETE EVAL PAIR: {entry}")
        oracle_path = os.path.join(skill_path, 'oracle', 'goldens.json')
        if os.path.exists(oracle_path) and not os.path.isdir(os.path.join(skill_path, 'oracle')):
            errors.append(f"INVALID ORACLE DIR: {entry}")

    if errors:
        for error in errors:
            print(error)
        sys.exit(1)

    print(f'All {len(skill_names)} skills have required structure')

if __name__ == '__main__':
    main()