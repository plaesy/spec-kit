reload:
	git add -A && git commit -m "Initial Commit" && rm -rf .claude .plaesy CLAUDE.md && ./scripts/bash/plaesy-init.sh . --ai claude_code && git restore .plaesy && claude
# 	rm -rf .claude .plaesy CLAUDE.md && ./scripts/bash/plaesy-init.sh . --ai opencode

detect:
	bash scripts/bash/config-manager.sh detect-platform

reset:
	rm -rf .git && git init && git remote add origin git@github.com:plaesy/spec-kit.git && git checkout -b main && git add -A && git commit -m "Initial Commit" 
# 	&& git push origin main -f