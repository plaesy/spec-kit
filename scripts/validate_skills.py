#!/usr/bin/env python3
"""
Validate all 8 core skills + founder have required structure
"""

import os
import sys

def main():
    skills_dir = 'skills'
    # Only check the 9 skills that should have evals + oracle
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
    required = [
        'SKILL.md',
        'evals/evals.json',
        'evals/results.json',
        'oracle/goldens.json'
    ]
    
    errors = []
    for skill in core_skills:
        skill_path = os.path.join(skills_dir, skill)
        if not os.path.isdir(skill_path):
            errors.append(f"MISSING DIR: {skill}")
            continue
        for req in required:
            req_path = os.path.join(skill_path, req)
            if not os.path.exists(req_path):
                errors.append(f"MISSING: {skill}/{req}")
    
    if errors:
        for e in errors:
            print(e)
        sys.exit(1)
    
    print('All 9 core skills have required structure')

if __name__ == '__main__':
    main()