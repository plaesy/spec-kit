#!/usr/bin/env python3
"""
Oracle Harness for Spec-Kit Skills
Records and replays golden tests against skill implementations.
Usage:
  python oracle_runner.py --skill economics-engine --record    # Record new goldens
  python oracle_runner.py --skill economics-engine --replay    # Replay against goldens
  python oracle_runner.py --skill all --replay                 # Replay all skills
"""

import json
import sys
import argparse
import subprocess
import importlib.util
from pathlib import Path
from typing import Dict, Any, List
from datetime import datetime

SPEC_KIT_ROOT = Path(__file__).parent.parent
SKILLS_DIR = SPEC_KIT_ROOT / "skills"

class OracleHarness:
    def __init__(self, skill_name: str):
        self.skill_name = skill_name
        self.skill_dir = SKILLS_DIR / skill_name
        self.oracle_dir = self.skill_dir / "oracle"
        self.goldens_file = self.oracle_dir / "goldens.json"
        self.vectors_dir = self.oracle_dir / "vectors"
        self.vectors_dir.mkdir(exist_ok=True)
        
    def load_goldens(self) -> List[Dict]:
        if not self.goldens_file.exists():
            return []
        with open(self.goldens_file) as f:
            data = json.load(f)
        return data.get("goldens", [])
    
    def record(self, golden: Dict) -> Dict:
        """Run a golden test and record the actual output as vector."""
        print(f"Recording: {golden['name']}")
        
        # Import the skill module dynamically
        skill_module = self._load_skill_module()
        if not skill_module:
            return {"error": "Could not load skill module"}
        
        # Find the handler function
        handler = getattr(skill_module, golden.get("handler", "run"), None)
        if not handler:
            return {"error": f"Handler not found: {golden.get('handler', 'run')}"}
        
        # Run with input
        input_data = golden["input"]
        try:
            output = handler(input_data)
            if hasattr(output, '__dict__'):
                output = output.__dict__
        except Exception as e:
            output = {"error": str(e), "exception_type": type(e).__name__}
        
        # Create vector record
        vector = {
            "golden_name": golden["name"],
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "input": golden["input"],
            "expected_output": golden["expected_output"],
            "actual_output": output,
            "tolerance": golden.get("tolerance", {})
        }
        
        # Save vector
        vector_file = self.vectors_dir / f"{golden['name']}.json"
        with open(vector_file, "w") as f:
            json.dump(vector, f, indent=2)
        
        print(f"  Saved vector to {vector_file}")
        return vector
    
    def replay(self) -> Dict:
        """Replay all goldens against current implementation."""
        goldens = self.load_goldens()
        if not goldens:
            return {"error": "No goldens found", "skill": self.skill_name}
        
        results = {
            "skill": self.skill_name,
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "total": len(goldens),
            "passed": 0,
            "failed": 0,
            "details": []
        }
        
        for golden in goldens:
            print(f"Replaying: {golden['name']}")
            vector = self.record(golden)
            
            if "error" in vector:
                results["failed"] += 1
                results["details"].append({
                    "name": golden["name"],
                    "status": "ERROR",
                    "error": vector["error"]
                })
                continue
            
            # Compare actual vs expected
            passed, diff = self._compare_outputs(
                golden["expected_output"],
                vector["actual_output"],
                golden.get("tolerance", {})
            )
            
            if passed:
                results["passed"] += 1
                results["details"].append({"name": golden["name"], "status": "PASS"})
                print(f"  PASS")
            else:
                results["failed"] += 1
                results["details"].append({
                    "name": golden["name"],
                    "status": "FAIL",
                    "diff": diff
                })
                print(f"  FAIL: {diff}")
        
        # Save replay results
        replay_file = self.oracle_dir / f"replay_{datetime.utcnow().strftime('%Y%m%d_%H%M%S')}.json"
        with open(replay_file, "w") as f:
            json.dump(results, f, indent=2)
        
        print(f"\nResults: {results['passed']}/{results['total']} passed")
        print(f"Saved replay to {replay_file}")
        return results
    
    def _compare_outputs(self, expected: Dict, actual: Dict, tolerance: Dict) -> tuple:
        """Compare expected vs actual with tolerance for numeric values."""
        diffs = []
        
        for key, exp_val in expected.items():
            if key not in actual:
                diffs.append(f"Missing key: {key}")
                continue
            
            act_val = actual[key]
            
            # Handle numeric tolerance
            if key in tolerance and isinstance(exp_val, (int, float)) and isinstance(act_val, (int, float)):
                tol = tolerance[key]
                if abs(exp_val - act_val) > tol:
                    diffs.append(f"{key}: expected {exp_val}, got {act_val} (tolerance {tol})")
            elif exp_val != act_val:
                diffs.append(f"{key}: expected {exp_val}, got {act_val}")
        
        # Check for unexpected keys (warn but don't fail)
        for key in actual:
            if key not in expected:
                diffs.append(f"Unexpected key: {key} = {actual[key]}")
        
        return len(diffs) == 0, diffs
    
    def _load_skill_module(self):
        """Dynamically load the skill's main module."""
        # Try common entry points
        for entry in ["main.py", "skill.py", f"{self.skill_name}.py", "run.py"]:
            module_path = self.skill_dir / entry
            if module_path.exists():
                spec = importlib.util.spec_from_file_location(self.skill_name, module_path)
                module = importlib.util.module_from_spec(spec)
                spec.loader.exec_module(module)
                return module
        
        # Try to find any .py file
        for py_file in self.skill_dir.glob("*.py"):
            spec = importlib.util.spec_from_file_location(self.skill_name, py_file)
            module = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(module)
            return module
        
        return None


def main():
    parser = argparse.ArgumentParser(description="Oracle Harness for Spec-Kit Skills")
    parser.add_argument("--skill", required=True, help="Skill name or 'all'")
    parser.add_argument("--record", action="store_true", help="Record new goldens (run with --golden)")
    parser.add_argument("--replay", action="store_true", help="Replay goldens against current implementation")
    parser.add_argument("--golden", help="Specific golden name to record")
    args = parser.parse_args()
    
    if args.skill == "all":
        skills = [d.name for d in SKILLS_DIR.iterdir() if d.is_dir() and (d / "oracle" / "goldens.json").exists()]
        for skill in skills:
            print(f"\n=== {skill} ===")
            harness = OracleHarness(skill)
            if args.replay:
                harness.replay()
        return
    
    harness = OracleHarness(args.skill)
    
    if args.record and args.golden:
        goldens = harness.load_goldens()
        golden = next((g for g in goldens if g["name"] == args.golden), None)
        if golden:
            harness.record(golden)
        else:
            print(f"Golden not found: {args.golden}")
    elif args.replay:
        harness.replay()
    else:
        parser.print_help()


if __name__ == "__main__":
    main()