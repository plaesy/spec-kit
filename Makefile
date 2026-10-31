# GNU make on Windows ignores SHELL and runs recipes with cmd.exe, where `rm` does not exist.
ifeq ($(OS),Windows_NT)
RM_GIT = powershell -NoProfile -Command "if (Test-Path .git) { Remove-Item -Recurse -Force .git }"
else
RM_GIT = rm -rf .git
endif

# Release asset name that install.sh looks for on this machine. The root build is named the
# same way, so it is the file the curl installer would have shipped.
ifeq ($(OS),Windows_NT)
EXE = .exe
ASSET = plaesy-windows-amd64.exe
else
EXE =
ASSET = plaesy-$(shell uname -s | tr A-Z a-z)-$(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
endif
PLAESY = ./$(ASSET)

# Same build as release.yml: cargo release build of the crate, copied to the release asset name.
build:
	cd ../cli/crates/plaesy && cargo build --release
	cp ../cli/crates/plaesy/target/release/plaesy$(EXE) $(ASSET)

# The one command for local work: rebuilds the root binary, installs it over the real installed
# plaesy through install.sh (the release asset is served from dist/ with SHA256SUMS), then
# reloads the generated .plaesy/ and mirrors with that root binary and checks them.
reload: build
	rm -rf dist && mkdir -p dist
	cp $(ASSET) dist/$(ASSET)
	cd dist && sha256sum $(ASSET) > SHA256SUMS
	PLAESY_BASE_URL="file://$$(cd dist && pwd -W 2>/dev/null || pwd)" bash install.sh
	if [ -n "$$LOCALAPPDATA" ]; then cmp "$$LOCALAPPDATA/Plaesy/bin/plaesy$(EXE)" $(ASSET); fi
	$(PLAESY) reload
	$(PLAESY) reload --ai claude --prune --prune-apply
	$(PLAESY) reload --ai kilo --prune --prune-apply
	$(PLAESY) markdown validate
	echo "reload ok: installed plaesy is the root build, reloaded with it"

test:
	cd ../cli/crates/plaesy && cargo test

# Local development only: symlink every skill (a folder under skills/ with a SKILL.md) into the
# agent skill folder, so `git pull` updates the linked skills in place. Override with
# AGENT_SKILLS=$HOME/.codex/skills. Needs bash and, on Windows, developer mode for symlinks.
AGENT_SKILLS ?= $(HOME)/.claude/skills
link-skills: SHELL := bash
link-skills: export MSYS := winsymlinks:nativestrict
link-skills:
	mkdir -p "$(AGENT_SKILLS)"
	for d in skills/*/; do \
	  [ -f "$$d/SKILL.md" ] || continue; \
	  ln -sfn "$$(cd "$$d" && pwd)" "$(AGENT_SKILLS)/$$(basename "$$d")"; \
	done
	echo "linked skills into $(AGENT_SKILLS)"

# Removes only the symlinks that point into this repository.
unlink-skills: SHELL := bash
unlink-skills:
	repo="$$(pwd -P)"; \
	for l in "$(AGENT_SKILLS)"/*; do \
	  [ -L "$$l" ] || continue; \
	  case "$$(readlink "$$l")" in "$$repo"/*) rm "$$l" ;; esac; \
	done
	echo "unlinked skills from $(AGENT_SKILLS)"

# Case files for the instructions (phrase checks and lint fixtures) and the offline design
# lint. Model briefs run only when PLAESY_EVAL_RUNNER is set; otherwise they are SKIP.
evals: build
	$(PLAESY) validate evals

# Removes only generated files. It never stages or commits: `git add -Af` here
# swallowed every unrelated working-tree change and committed it as "Initial
# Commit", which is how the repo ended up with three commits by that name.
clean:
	git add -A && git commit -m "Initial Commit" && rm -rf .claude .kilo .plaesy AGENTS.md CLAUDE.md

claude: build
	$(PLAESY) init . --ai claude && $(PLAESY) reload && $(PLAESY) analyze && claude


kilo: build
	$(PLAESY) init . --ai kilo && $(PLAESY) reload && $(PLAESY) analyze && kilo --auto

detect: build
	$(PLAESY) platforms detect

# DANGER: rewrites local history from scratch, force-pushes it over main, AND
# deletes every GitHub Release (with its tag and uploaded assets) plus any
# leftover bare tag on the remote before re-tagging at VERSION. Everything
# here is irreversible once it reaches GitHub:
#   1. delete every Release (gh release delete --cleanup-tag removes its tag too)
#   2. delete any remaining bare tag the step above didn't own
#   3. wipe local .git, commit a fresh "Initial Commit", force-push it to main
#   4. tag that commit v<VERSION> and push the tag
# Step 4's tag push fires .github/workflows/release.yml on the remote, which
# builds the cross-platform binaries and recreates the GitHub Release — so a
# plain `make reset` now reproduces a full release from scratch, not just a
# clean history.
resetf: test
	-gh release list -R plaesy/spec-kit --limit 1000 --json tagName -q '.[].tagName' | xargs -r -I{} gh release delete {} -R plaesy/spec-kit --yes --cleanup-tag
	-git ls-remote --tags git@github.com:plaesy/spec-kit.git | awk '{print $$2}' | while read -r ref; do git push git@github.com:plaesy/spec-kit.git --delete "$$ref"; done
	$(RM_GIT) && git init && git remote add origin git@github.com:plaesy/spec-kit.git && git checkout -b main && git add -A && git commit -m "Initial Commit" && git push origin main -f && git tag "v$$(cat VERSION)" && git push origin "v$$(cat VERSION)" && sudo plaesy uninstall

reset:
	$(RM_GIT) && git init && git remote add origin git@github.com:plaesy/spec-kit.git && git checkout -b main && git add -A && git commit -m "Initial Commit"
