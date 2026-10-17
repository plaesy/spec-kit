package contract

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Anchor is the result of checking the freeze against git history.
//
// What it proves when Verified is true: the frozen snapshot is byte-equal (as JSON) to
// the snapshot in the first committed version of the file that carries one; that first
// version records no results; and every amendment in every committed version is still
// present, unchanged, in the current file (the amendment log is append-only).
//
// What it does not prove: that the first frozen version was committed before the work
// started (git records when a commit was made, not whether the work came before it);
// that the contract was not replaced by a new file under another name; or anything about
// the work itself. Those remain self-attested.
type Anchor struct {
	Verified bool
	Commit   string // short hash of the first committed version with a frozen snapshot
	Note     string // what was checked, or why it could not be
}

// VerifyAnchor checks the current contract against its git history. It returns an
// Anchor and the issues the check raised. The issue severity depends on opts: a freeze
// that cannot be verified is a warning, and an error under --gate or --require-git.
func VerifyAnchor(path string, cur Contract, opts Options) (Anchor, []Issue) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return unverified(opts, "cannot resolve the contract path: "+err.Error())
	}
	top, err := gitOut(filepath.Dir(abs), "rev-parse", "--show-toplevel")
	if err != nil {
		return unverified(opts, "not in a git repository: the freeze cannot be verified")
	}
	top = filepath.FromSlash(top)
	rel, err := filepath.Rel(top, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return unverified(opts, "contract is outside the git repository: the freeze cannot be verified")
	}
	relSlash := filepath.ToSlash(rel)
	if _, err := gitOut(top, "ls-files", "--error-unmatch", "--", relSlash); err != nil {
		return unverified(opts, "the contract is not committed to git yet: commit it with its frozen snapshot before work starts")
	}
	logOut, err := gitOut(top, "log", "--reverse", "--format=%H", "--", relSlash)
	if err != nil {
		return unverified(opts, "cannot read the contract's git history: "+err.Error())
	}

	var versions []Contract
	var shas []string
	for _, sha := range strings.Fields(logOut) {
		raw, err := gitOut(top, "show", sha+":"+relSlash)
		if err != nil {
			continue
		}
		c, err := Parse([]byte(raw))
		if err != nil {
			continue // an earlier version that was not a valid contract cannot be an anchor
		}
		versions = append(versions, c)
		shas = append(shas, sha)
	}

	first := -1
	for i, v := range versions {
		if v.Frozen != nil {
			first = i
			break
		}
	}
	if first < 0 {
		return unverified(opts, "no committed version of the contract holds a frozen snapshot: commit the frozen contract before work starts")
	}

	var issues []Issue
	add := func(format string, args ...any) {
		issues = append(issues, Issue{Severity: SevError, Message: fmt.Sprintf(format, args...)})
	}
	commit := shas[first]
	anchored := versions[first]
	for _, it := range anchored.Items {
		if it.Result != "" {
			add("the first frozen version (commit %s) already records results: the contract was filled in after the work", short(commit))
			break
		}
	}
	if !sameJSON(anchored.Frozen, cur.Frozen) {
		add("frozen snapshot differs from the first committed version (commit %s): the freeze was rewritten; restore it and record each change as an amendment", short(commit))
	}
	for i, v := range versions {
		for j, a := range v.Amendments {
			if j >= len(cur.Amendments) || !sameJSON(a, cur.Amendments[j]) {
				add("amendment %d recorded in commit %s was removed or changed: the amendment log is append-only", j+1, short(shas[i]))
			}
		}
	}
	if len(issues) > 0 {
		return Anchor{Verified: false, Commit: short(commit), Note: "git freeze check failed"}, issues
	}
	return Anchor{
		Verified: true,
		Commit:   short(commit),
		Note: fmt.Sprintf("frozen snapshot matches the first committed version (commit %s); %d committed version(s), amendment log append-only",
			short(commit), len(versions)),
	}, nil
}

// unverified reports a freeze that could not be checked. It is a warning, and an error
// under --gate or --require-git, where an unchecked freeze must not pass.
func unverified(opts Options, note string) (Anchor, []Issue) {
	sev := SevWarning
	if opts.Gate || opts.RequireGit {
		sev = SevError
	}
	return Anchor{Verified: false, Note: note},
		[]Issue{{Severity: sev, Message: "freeze NOT VERIFIED: " + note}}
}

func sameJSON(a, b any) bool {
	x, errA := json.Marshal(a)
	y, errB := json.Marshal(b)
	return errA == nil && errB == nil && string(x) == string(y)
}

func short(sha string) string {
	if len(sha) > 9 {
		return sha[:9]
	}
	return sha
}

// gitOut runs git in dir and returns trimmed stdout. An error includes stderr.
func gitOut(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
