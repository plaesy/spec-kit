assets:
	cd scripts && go run ./internal/assets/gen

build: assets
	cd scripts && go build -o ../plaesy ./cmd/plaesy

test:
	cd scripts && go test ./...

# Removes only generated files. It never stages or commits: `git add -Af` here
# swallowed every unrelated working-tree change and committed it as "Initial
# Commit", which is how the repo ended up with three commits by that name.
clean:
	git add -A && git commit -m "Initial Commit" && rm -rf .claude .kilo .plaesy AGENTS.md CLAUDE.md

claude: build
	./plaesy init . --ai claude && ./plaesy reload && ./plaesy analyze && claude


kilo: build
	./plaesy init . --ai kilo && ./plaesy reload && ./plaesy analyze && kilo --auto

detect: build
	./plaesy platforms detect

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
resetf: assets
	-gh release list -R plaesy/spec-kit --limit 1000 --json tagName -q '.[].tagName' | xargs -r -I{} gh release delete {} -R plaesy/spec-kit --yes --cleanup-tag
	-git ls-remote --tags git@github.com:plaesy/spec-kit.git | awk '{print $$2}' | while read -r ref; do git push git@github.com:plaesy/spec-kit.git --delete "$$ref"; done
	rm -rf .git && git init && git remote add origin git@github.com:plaesy/spec-kit.git && git checkout -b main && git add -A && git commit -m "Initial Commit" && git push origin main -f && git tag "v$$(cat VERSION)" && git push origin "v$$(cat VERSION)"

reset:	
	rm -rf .git && git init && git remote add origin git@github.com:plaesy/spec-kit.git && git checkout -b main && git add -A && git commit -m "Initial Commit"