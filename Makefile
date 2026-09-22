build:
	cd scripts && go build -o ../plaesy ./cmd/plaesy

reload: build
	git add -A && git commit -m "Initial Commit" && rm -rf .claude .plaesy CLAUDE.md && ./plaesy init . --ai claude_code && git restore .plaesy && claude
# 	rm -rf .claude .plaesy CLAUDE.md && ./plaesy init . --ai opencode

detect: build
	./plaesy config detect-platform

reset:
	rm -rf .git && git init && git remote add origin git@github.com:plaesy/spec-kit.git && git checkout -b main && git add -A && git commit -m "Initial Commit" 
# 	&& git push origin main -f