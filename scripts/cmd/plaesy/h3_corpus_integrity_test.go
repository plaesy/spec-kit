package main

// H3 (docs/assessment/2026-09-27-prompt-corpus.md, task T001, Block 4): three
// of its seven enforcement checks had no guard anywhere in this package. The
// other four were already covered elsewhere (routing_drift_test.go,
// mirror_parity_test.go, dead_link_test.go) — this file adds only the gap:
//   - no duplicated H2 heading within a single prompt file
//   - every @role mention in agents/*.agents.md resolves to a real agent file
//   - a sub-prompt's frontmatter description matches its parent's Usage
//     Format table entry for that selector

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func promptsRoot(t *testing.T) string {
	t.Helper()
	root := corpusRoot(t)
	dir := filepath.Join(root, "prompts")
	if _, err := os.Stat(dir); err != nil {
		t.Skip("prompts/ not found: not a repository checkout")
	}
	return dir
}

// TestNoDuplicateH2HeadingInAPromptFile pins the structural half of the H3
// brief: two `## ` headings with the same text in one file mean one of them
// is either a copy-paste leftover or the sign of an accidentally duplicated
// section tail (the exact defect class C2 fixed in prompts/create/tasks.md).
// Headings are only counted outside fenced code blocks — several prompts
// (fix.md, doc.md) embed a literal report template with its own `## `
// headings inside a ```-fence, and those are example output, not structure.
func TestNoDuplicateH2HeadingInAPromptFile(t *testing.T) {
	root := promptsRoot(t)
	h2 := regexp.MustCompile(`^##\s+(.+?)\s*$`)
	fence := regexp.MustCompile("^```")

	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		lines := strings.Split(string(raw), "\n")
		inFence := false
		seen := map[string]int{}
		var order []string
		for _, line := range lines {
			if fence.MatchString(strings.TrimRight(line, "\r")) {
				inFence = !inFence
				continue
			}
			if inFence {
				continue
			}
			if m := h2.FindStringSubmatch(strings.TrimRight(line, "\r")); m != nil {
				heading := m[1]
				if _, ok := seen[heading]; !ok {
					order = append(order, heading)
				}
				seen[heading]++
			}
		}
		for _, heading := range order {
			if seen[heading] > 1 {
				rel, _ := filepath.Rel(root, p)
				t.Errorf("prompts/%s: heading %q appears %d times (outside code fences) — one is likely a duplicated section tail",
					filepath.ToSlash(rel), heading, seen[heading])
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// agentHandles lists agents/*.agents.md basenames (without the .agents.md
// suffix) — the set of names a `@role` mention is allowed to resolve to.
func agentHandles(t *testing.T, root string) map[string]bool {
	t.Helper()
	dir := filepath.Join(root, "agents")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skip("agents/ not found: not a repository checkout")
	}
	handles := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".agents.md") {
			continue
		}
		handles[strings.TrimSuffix(e.Name(), ".agents.md")] = true
	}
	return handles
}

// TestEveryRoleMentionInAgentsResolvesToAnAgentFile checks Agent Authoring
// Contract §6 ("every @role mention must resolve to an existing
// agents/<role>.agents.md") the way the contract itself is scoped: within
// agents/*.agents.md, where `@name` is unambiguously a role reference. `@`
// appears with other meanings elsewhere in the corpus (npm scopes, decorators
// inside code fences), so this guard is intentionally not corpus-wide.
func TestEveryRoleMentionInAgentsResolvesToAnAgentFile(t *testing.T) {
	root := corpusRoot(t)
	handles := agentHandles(t, root)
	mention := regexp.MustCompile(`@([a-z][a-z-]*)`)

	dir := filepath.Join(root, "agents")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".agents.md") {
			continue
		}
		self := strings.TrimSuffix(e.Name(), ".agents.md")
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range mention.FindAllStringSubmatch(string(raw), -1) {
			name := m[1]
			if !handles[name] {
				t.Errorf("agents/%s: @%s does not resolve to any agents/%s.agents.md",
					e.Name(), name, name)
			}
			_ = self
		}
	}
}

// usageFormatEntry is one `/{fam}:{dim} ... # description` line from a
// parent command's Usage Format block.
var usageFormatEntry = regexp.MustCompile(`^/([a-z]+):([a-z][a-z-]*(?::[a-z][a-z-]*)?)(?:[^#]*)#\s*(.+?)\s*$`)

// TestSubPromptDescriptionMatchesParentUsageTable checks the last H3 gap:
// prompts/{fam}.md documents each `/{fam}:{dim}` selector with a one-line
// description in its Usage Format block, and prompts/{fam}/{dim}.md carries
// the same text in its own frontmatter `description`. The two are meant to
// be the same sentence read from two places — a mismatch means one was
// edited without the other, which is exactly the single-source-of-truth
// drift this corpus has been fixing all session.
func TestSubPromptDescriptionMatchesParentUsageTable(t *testing.T) {
	root := promptsRoot(t)
	descField := regexp.MustCompile(`(?m)^description:\s*"([^"]*)"`)

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		family := strings.TrimSuffix(e.Name(), ".md")
		subdir := filepath.Join(root, family)
		if info, err := os.Stat(subdir); err != nil || !info.IsDir() {
			continue // no dimension-scoped sub-prompts for this family
		}
		if family == "loop" {
			// prompts/loop.md's only `/loop:technical` line lives inside a
			// "Loop Configuration" example block illustrating flag syntax, not
			// a per-dimension description table like assess/fix/implement/
			// optimize/improve have one line per dimension for. Comparing it
			// 1:1 against prompts/loop/technical.md's real description is a
			// false positive, not drift — verified against the source 2026-09-29.
			continue
		}

		parentRaw, err := os.ReadFile(filepath.Join(root, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		parentDesc := map[string]string{}
		for _, line := range strings.Split(string(parentRaw), "\n") {
			line = strings.TrimSpace(strings.TrimRight(line, "\r"))
			if m := usageFormatEntry.FindStringSubmatch(line); m != nil {
				fam, dim, desc := m[1], m[2], m[3]
				if fam != family {
					continue
				}
				if _, exists := parentDesc[dim]; !exists {
					parentDesc[dim] = desc
				}
			}
		}

		for dim, wantDesc := range parentDesc {
			subPath := filepath.Join(subdir, strings.ReplaceAll(dim, ":", string(filepath.Separator))+".md")
			subRaw, err := os.ReadFile(subPath)
			if err != nil {
				// A referenced selector with no backing file is
				// routing_drift_test.go's job (TestRoutingTableReferencesResolve);
				// this test only compares descriptions where both sides exist.
				continue
			}
			m := descField.FindStringSubmatch(string(subRaw))
			if m == nil {
				t.Errorf("prompts/%s/%s.md: no frontmatter description to compare against parent Usage Format entry %q",
					family, dim, wantDesc)
				continue
			}
			gotDesc := m[1]
			if gotDesc != wantDesc {
				t.Errorf("prompts/%s.md documents /%s:%s as %q, but prompts/%s/%s.md's description says %q",
					family, family, dim, wantDesc, family, dim, gotDesc)
			}
		}
	}
}
