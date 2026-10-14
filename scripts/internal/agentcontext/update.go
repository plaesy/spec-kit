// Package agentcontext ports update-agent-context.sh: it syncs an EXISTING
// per-agent context file (CLAUDE.md, GEMINI.md, .github/copilot-instructions.md,
// .cursor/rules/specify-rules.mdc, QWEN.md, AGENTS.md) with the tech stack
// extracted from the current feature's plan.md. It patches, it does not
// scaffold: a missing context file is an error, not an invitation to invent one
// from a template (the agent-file template was removed on 2026-09-25).
package agentcontext

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/plaesy/spec-kit/internal/common"
)

type agentTarget struct {
	key  string
	file string
	name string
}

func targets(repoRoot string) map[string]agentTarget {
	return map[string]agentTarget{
		"claude":   {"claude", filepath.Join(repoRoot, "CLAUDE.md"), "Claude Code"},
		"gemini":   {"gemini", filepath.Join(repoRoot, "GEMINI.md"), "Gemini CLI"},
		"copilot":  {"copilot", filepath.Join(repoRoot, ".github", "copilot-instructions.md"), "GitHub Copilot"},
		"cursor":   {"cursor", filepath.Join(repoRoot, ".cursor", "rules", "specify-rules.mdc"), "Cursor IDE"},
		"qwen":     {"qwen", filepath.Join(repoRoot, "QWEN.md"), "Qwen Code"},
		"opencode": {"opencode", filepath.Join(repoRoot, "AGENTS.md"), "opencode"},
	}
}

var orderedKeys = []string{"claude", "gemini", "copilot", "cursor", "qwen", "opencode"}

type planFields struct {
	Lang        string
	Framework   string
	DB          string
	ProjectType string
}

var (
	langRE      = regexp.MustCompile(`^\*\*Language/Version\*\*:\s*(.+)$`)
	frameworkRE = regexp.MustCompile(`^\*\*Primary Dependencies\*\*:\s*(.+)$`)
	dbRE        = regexp.MustCompile(`^\*\*Storage\*\*:\s*(.+)$`)
	projTypeRE  = regexp.MustCompile(`^\*\*Project Type\*\*:\s*(.+)$`)
)

func extractPlanFields(planPath string) (planFields, error) {
	f, err := os.Open(planPath)
	if err != nil {
		return planFields{}, err
	}
	defer f.Close()

	var pf planFields
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if pf.Lang == "" {
			if m := langRE.FindStringSubmatch(line); m != nil && !strings.Contains(m[1], "NEEDS CLARIFICATION") {
				pf.Lang = strings.TrimSpace(m[1])
			}
		}
		if pf.Framework == "" {
			if m := frameworkRE.FindStringSubmatch(line); m != nil && !strings.Contains(m[1], "NEEDS CLARIFICATION") {
				pf.Framework = strings.TrimSpace(m[1])
			}
		}
		if pf.DB == "" {
			if m := dbRE.FindStringSubmatch(line); m != nil && !strings.Contains(m[1], "N/A") && !strings.Contains(m[1], "NEEDS CLARIFICATION") {
				pf.DB = strings.TrimSpace(m[1])
			}
		}
		if pf.ProjectType == "" {
			if m := projTypeRE.FindStringSubmatch(line); m != nil {
				pf.ProjectType = strings.TrimSpace(m[1])
			}
		}
	}
	return pf, scanner.Err()
}

// Update mirrors update-agent-context.sh's main entry point. agentType is
// one of claude|gemini|copilot|cursor|qwen|opencode, or "" to update every
// context file that already exists (falling back to creating CLAUDE.md if
// none exist yet).
func Update(agentType string) error {
	repoRoot, err := common.GetRepoRoot()
	if err != nil {
		return err
	}
	branch, err := common.GetCurrentBranch()
	if err != nil {
		return err
	}
	featureDir := filepath.Join(repoRoot, ".plaesy", "specs", branch)
	planPath := filepath.Join(featureDir, "plan.md")

	if _, err := os.Stat(planPath); err != nil {
		return fmt.Errorf("no plan.md found at %s", planPath)
	}

	fmt.Printf("=== Updating agent context files for feature %s ===\n", branch)

	pf, err := extractPlanFields(planPath)
	if err != nil {
		return fmt.Errorf("reading plan.md: %w", err)
	}

	tg := targets(repoRoot)

	changedAny := false
	run := func(key string) error {
		t, ok := tg[key]
		if !ok {
			return fmt.Errorf("unknown agent type %q (expected claude|gemini|copilot|cursor|qwen|opencode)", key)
		}
		changed, err := updateAgentFile(t, branch, pf)
		changedAny = changedAny || changed
		return err
	}

	switch agentType {
	case "":
		anyExisted := false
		for _, key := range orderedKeys {
			t := tg[key]
			if _, err := os.Stat(t.file); err == nil {
				anyExisted = true
				changed, err := updateAgentFile(t, branch, pf)
				if err != nil {
					return err
				}
				changedAny = changedAny || changed
			}
		}
		if !anyExisted {
			names := make([]string, 0, len(orderedKeys))
			for _, key := range orderedKeys {
				names = append(names, tg[key].file)
			}
			return fmt.Errorf("no agent context file found — create one of %s first; plaesy context update only updates existing files", strings.Join(names, ", "))
		}
	default:
		if err := run(agentType); err != nil {
			return err
		}
	}

	fmt.Println()
	// The summary below is rendered from the plan file, not from what was
	// written, so printing it on a run that changed nothing claimed
	// "Added language: Rust" for a file that was left alone.
	if !changedAny {
		fmt.Println("No changes: every existing context file was already up to date.")
	} else {
		fmt.Println("Summary of changes:")
		if pf.Lang != "" {
			fmt.Printf("- Added language: %s\n", pf.Lang)
		}
		if pf.Framework != "" {
			fmt.Printf("- Added framework: %s\n", pf.Framework)
		}
		if pf.DB != "" && pf.DB != "N/A" {
			fmt.Printf("- Added database: %s\n", pf.DB)
		}
	}
	fmt.Println()
	fmt.Println("Usage: plaesy context update [claude|gemini|copilot|cursor|qwen|opencode]")
	return nil
}

// updateAgentFile syncs an EXISTING agent context file with the current
// feature's plan.md. It never creates one: the agent-file template was removed
// on 2026-09-25, so scaffolding a context file is the user's call (copy an
// existing CLAUDE.md/AGENTS.md, or let the AI platform create its own). A
// missing target is reported as an error rather than silently skipped — a run
// that changed nothing must not report success.
func updateAgentFile(t agentTarget, branch string, pf planFields) (bool, error) {
	if _, err := os.Stat(t.file); err != nil {
		return false, fmt.Errorf("no %s context file at %s: create it first (plaesy context update only updates an existing file)", t.name, t.file)
	}

	fmt.Printf("Updating %s context file: %s\n", t.name, t.file)
	return patchAgentFile(t, branch, pf)
}

var (
	activeTechRE   = regexp.MustCompile(`(?s)## Active Technologies\n(.*?)\n\n`)
	recentChangeRE = regexp.MustCompile(`(?s)## Recent Changes\n(.*?)(\n\n|$)`)
	lastUpdatedRE  = regexp.MustCompile(`Last updated: \d{4}-\d{2}-\d{2}`)
)

// recentChangeEntry renders the "what this branch added" line. The parts are
// joined rather than formatted positionally because a plan may legitimately
// leave Language blank (a NEEDS CLARIFICATION marker) — the old positional
// format wrote "Added  + Gin", with the gap where the language should be.
func recentChangeEntry(branch string, pf planFields) string {
	added := make([]string, 0, 2)
	if pf.Lang != "" {
		added = append(added, pf.Lang)
	}
	if pf.Framework != "" {
		added = append(added, pf.Framework)
	}
	if len(added) == 0 {
		return fmt.Sprintf("- %s: Updated plan", branch)
	}
	return fmt.Sprintf("- %s: Added %s", branch, strings.Join(added, " + "))
}

// patchAgentFile rewrites the tracked sections of t.file and reports whether
// the file actually changed. The three patterns are all optional matches, so
// "nothing to do" has two distinct causes that used to look identical: a file
// already in sync (legitimate, and idempotent) and a file with none of the
// sections at all (nothing in the framework creates them, so this is the
// normal case for a user's own CLAUDE.md/AGENTS.md). The second is an error;
// the first is reported, not claimed as an update.
func patchAgentFile(t agentTarget, branch string, pf planFields) (bool, error) {
	fmt.Printf("Updating existing %s context file...\n", t.name)

	data, err := os.ReadFile(t.file)
	if err != nil {
		return false, err
	}
	content := string(data)
	matched := false

	// Preserve manual additions block verbatim; everything else may be rewritten.
	manualBlock := ""
	if start := strings.Index(content, "<!-- MANUAL ADDITIONS START -->"); start >= 0 {
		if end := strings.Index(content, "<!-- MANUAL ADDITIONS END -->"); end >= 0 {
			manualBlock = content[start : end+len("<!-- MANUAL ADDITIONS END -->")]
		}
	}

	if m := activeTechRE.FindStringSubmatch(content); m != nil {
		matched = true
		existing := m[1]
		var additions []string
		if pf.Lang != "" && !strings.Contains(existing, pf.Lang) {
			additions = append(additions, fmt.Sprintf("- %s + %s (%s)", pf.Lang, pf.Framework, branch))
		}
		if pf.DB != "" && pf.DB != "N/A" && !strings.Contains(existing, pf.DB) {
			additions = append(additions, fmt.Sprintf("- %s (%s)", pf.DB, branch))
		}
		if len(additions) > 0 {
			newBlock := existing + "\n" + strings.Join(additions, "\n")
			content = strings.Replace(content, m[0], "## Active Technologies\n"+newBlock+"\n\n", 1)
		}
	}

	if m := recentChangeRE.FindStringSubmatch(content); m != nil {
		matched = true
		var lines []string
		for _, l := range strings.Split(strings.TrimSpace(m[1]), "\n") {
			if l != "" {
				lines = append(lines, l)
			}
		}
		entry := recentChangeEntry(branch, pf)
		// One line per branch, updated in place. Prepending unconditionally —
		// which is what this did — meant every re-run on the same branch added
		// another identical line, and the keep-last-3 cap silently evicted
		// *other* branches' real history to hide the duplicates. The result
		// depended on how many times a command happened to be run.
		replaced := false
		prefix := "- " + branch + ":"
		for i, l := range lines {
			if strings.HasPrefix(l, prefix) {
				lines[i] = entry
				replaced = true
				break
			}
		}
		if !replaced {
			lines = append([]string{entry}, lines...)
		}
		if len(lines) > 3 {
			lines = lines[:3]
		}
		content = recentChangeRE.ReplaceAllString(content, "## Recent Changes\n"+strings.Join(lines, "\n")+"\n\n")
	}

	content = lastUpdatedRE.ReplaceAllString(content, "Last updated: "+time.Now().Format("2006-01-02"))
	if lastUpdatedRE.MatchString(string(data)) {
		matched = true
	}

	if manualBlock != "" {
		if start := strings.Index(content, "<!-- MANUAL ADDITIONS START -->"); start >= 0 {
			if end := strings.Index(content, "<!-- MANUAL ADDITIONS END -->"); end >= 0 {
				content = content[:start] + manualBlock + content[end+len("<!-- MANUAL ADDITIONS END -->"):]
			}
		}
	}

	if !matched {
		// None of the three patterns found anything, so the file is one this
		// command has never been able to update. It used to be written back
		// byte-identically and reported as "updated successfully" — the very
		// "a run that changed nothing never reports success" rule that
		// updateAgentFile documents, on the file that function is about. A
		// context file without these sections is the normal case: nothing in
		// the framework creates them, so a user's own CLAUDE.md/AGENTS.md
		// never has them.
		return false, fmt.Errorf("no patchable sections in %s: expected a '## Active Technologies' or '## Recent Changes' heading and a 'Last updated: YYYY-MM-DD' line; the file was left untouched", t.file)
	}

	if content == string(data) {
		fmt.Printf("➖ %s context file already up to date for %s (nothing to write)\n", t.name, branch)
		return false, nil
	}

	if err := os.WriteFile(t.file, []byte(content), 0o644); err != nil {
		return false, err
	}
	fmt.Printf("✅ %s context file updated successfully\n", t.name)
	return true, nil
}
