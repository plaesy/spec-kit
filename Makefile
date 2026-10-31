# Spec-Kit Makefile
# Run: make help

.PHONY: help evals oracle test clean install lint docs sim validate

# Default target
help:
	@echo "Spec-Kit - Business Validation Pipeline"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Core Targets:"
	@echo "  make evals          - Run all skill evals (9 core skills)"
	@echo "  make oracle         - Run oracle harness for 9 core skills"
	@echo "  make test           - Run evals + oracle (full test suite)"
	@echo "  make clean          - Clean generated artifacts"
	@echo ""
	@echo "Skill Targets (run specific skill):"
	@echo "  make evals-ideator"
	@echo "  make evals-buyer-discovery"
	@echo "  make evals-compliance-health"
	@echo "  make evals-economics-engine"
	@echo "  make evals-plan-compiler"
	@echo "  make evals-assess"
	@echo "  make evals-pipeline"
	@echo "  make evals-youtube-automation"
	@echo "  make evals-founder"
	@echo ""
	@echo "  make oracle-ideator"
	@echo "  make oracle-buyer-discovery"
	@echo "  make oracle-compliance-health"
	@echo "  make oracle-economics-engine"
	@echo "  make oracle-plan-compiler"
	@echo "  make oracle-assess"
	@echo "  make oracle-pipeline"
	@echo "  make oracle-youtube-automation"
	@echo "  make oracle-founder"
	@echo ""
	@echo "Simulation:"
	@echo "  make sim           - Run full ai-health-influencer simulation"
	@echo "  make sim-eval      - Evaluate simulation results"
	@echo ""
	@echo "Maintenance:"
	@echo "  make lint          - Check code style (pycodestyle, ruff)"
	@echo "  make docs          - Validate documentation"
	@echo "  make install       - Install dependencies"
	@echo "  make clean         - Remove .plaesy/, __pycache__, *.pyc"
	@echo "  make validate      - Validate all skills structure"

# Run all skill evals (9 core skills)
evals:
	@echo "Running all skill evaluations..."
	@python scripts/validate_evals.py

# Run oracle harness for 9 core skills
oracle:
	@echo "Running oracle harness replay..."
	@for skill in ideator buyer-discovery compliance-health economics-engine plan-compiler assess pipeline youtube-automation founder; do \
		echo "=== $$skill ==="; \
		python scripts/oracle_runner.py --skill $$skill --replay; \
	done

# Full test suite: evals + oracle
test: evals oracle

# Individual skill evals
evals-%:
	@skill=$*; \
	if [ -d "skills/$$skill/evals" ]; then \
		echo "=== $$skill ==="; \
		python scripts/validate_evals.py $$skill; \
	else \
		echo "No evals for $$skill"; \
	fi

# Individual oracle
oracle-%:
	@skill=$*; \
	if [ -d "skills/$$skill/oracle" ]; then \
		echo "=== $$skill ==="; \
		python scripts/oracle_runner.py --skill $$skill --replay; \
	else \
		echo "No oracle for $$skill"; \
	fi

# Run full ai-health-influencer simulation
sim:
	@echo "Running full pipeline simulation for ai-health-influencer..."
	@python scripts/run_simulation.py

# Evaluate simulation results
sim-eval:
	@echo "Evaluating simulation results..."
	@python scripts/oracle_runner.py --skill pipeline --replay

# Install dependencies
install:
	pip install -q pytest pyyaml
	@for skill in $$(find skills -name "requirements.txt" -exec dirname {} \; | sort -u); do \
		if [ -f "$$skill/requirements.txt" ]; then \
			pip install -q -r "$$skill/requirements.txt"; \
		fi; \
	done

# Lint check
lint:
	@which ruff >/dev/null 2>&1 && ruff check skills/ scripts/ || \
	which pycodestyle >/dev/null 2>&1 && pycodestyle skills/ scripts/ || \
	echo "No linter installed (install ruff or pycodestyle)"

# Validate documentation
docs:
	@echo "Checking documentation..."
	@for f in docs/ARCHITECTURE.md docs/DESIGN.md docs/CLI-CONTRACT.md CHANGELOG.md CONTRIBUTING.md; do \
		if [ ! -f "$$f" ]; then echo "Missing: $$f"; exit 1; fi; \
	done
	@echo "All docs present"

# Clean generated artifacts
clean:
	rm -rf .plaesy/
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -name "*.pyc" -delete 2>/dev/null || true
	find . -name "*.pyo" -delete 2>/dev/null || true
	find . -name ".pytest_cache" -exec rm -rf {} + 2>/dev/null || true

# Quick validation of all skills structure
validate:
	@echo "Validating all skills structure..."
	@python scripts/validate_skills.py