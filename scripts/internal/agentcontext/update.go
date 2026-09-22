// Package agentcontext ports update-agent-context.sh: it syncs the
// per-agent context file (CLAUDE.md, GEMINI.md, .github/copilot-instructions.md,
// .cursor/rules/specify-rules.mdc, QWEN.md, AGENTS.md) with the tech stack
// extracted from the current feature's plan.md.
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
	featureDir := filepath.Join(repoRoot, "specs", branch)
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

	run := func(key string) error {
		t, ok := tg[key]
		if !ok {
			return fmt.Errorf("unknown agent type %q (expected claude|gemini|copilot|cursor|qwen|opencode)", key)
		}
		return updateAgentFile(t, repoRoot, branch, pf)
	}

	switch agentType {
	case "":
		anyExisted := false
		for _, key := range orderedKeys {
			t := tg[key]
			if _, err := os.Stat(t.file); err == nil {
				anyExisted = true
				if err := updateAgentFile(t, repoRoot, branch, pf); err != nil {
					return err
				}
			}
		}
		if !anyExisted {
			if err := updateAgentFile(tg["claude"], repoRoot, branch, pf); err != nil {
				return err
			}
		}
	default:
		if err := run(agentType); err != nil {
			return err
		}
	}

	fmt.Println()
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
	fmt.Println()
	fmt.Println("Usage: plaesy update-agent-context [claude|gemini|copilot|cursor|qwen|opencode]")
	return nil
}

func commandsFor(lang string) string {
	switch {
	case strings.Contains(lang, "Python"):
		return "cd src && pytest && ruff check ."
	case strings.Contains(lang, "Rust"):
		return "cargo test && cargo clippy"
	case strings.Contains(lang, "JavaScript"), strings.Contains(lang, "TypeScript"):
		return "npm test && npm run lint"
	default:
		if lang == "" {
			return "# Add commands for your stack"
		}
		return "# Add commands for " + lang
	}
}

func structureFor(projectType string) string {
	if strings.Contains(projectType, "web") {
		return "backend/\nfrontend/\ntests/"
	}
	return "src/\ntests/"
}

func updateAgentFile(t agentTarget, repoRoot, branch string, pf planFields) error {
	fmt.Printf("Updating %s context file: %s\n", t.name, t.file)

	if _, err := os.Stat(t.file); err != nil {
		return createAgentFile(t, repoRoot, branch, pf)
	}
	return patchAgentFile(t, branch, pf)
}

func createAgentFile(t agentTarget, repoRoot, branch string, pf planFields) error {
	fmt.Printf("Creating new %s context file...\n", t.name)

	template := filepath.Join(repoRoot, ".plaesy", "templates", "agent-file-template.md")
	data, err := os.ReadFile(template)
	if err != nil {
		return fmt.Errorf("template not found: %s", template)
	}
	content := string(data)

	content = strings.ReplaceAll(content, "[PROJECT NAME]", filepath.Base(repoRoot))
	content = strings.ReplaceAll(content, "[DATE]", time.Now().Format("2006-01-02"))
	content = strings.ReplaceAll(content, "[EXTRACTED FROM ALL PLAN.MD FILES]",
		fmt.Sprintf("- %s + %s (%s)", pf.Lang, pf.Framework, branch))
	content = strings.ReplaceAll(content, "[ACTUAL STRUCTURE FROM PLANS]", structureFor(pf.ProjectType))
	content = strings.ReplaceAll(content, "[ONLY COMMANDS FOR ACTIVE TECHNOLOGIES]", commandsFor(pf.Lang))
	content = strings.ReplaceAll(content, "[LANGUAGE-SPECIFIC, ONLY FOR LANGUAGES IN USE]",
		fmt.Sprintf("%s: Follow standard conventions", pf.Lang))
	content = strings.ReplaceAll(content, "[LAST 3 FEATURES AND WHAT THEY ADDED]",
		fmt.Sprintf("- %s: Added %s + %s", branch, pf.Lang, pf.Framework))

	if err := os.MkdirAll(filepath.Dir(t.file), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(t.file, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("✅ %s context file updated successfully\n", t.name)
	return nil
}

var (
	activeTechRE   = regexp.MustCompile(`(?s)## Active Technologies\n(.*?)\n\n`)
	recentChangeRE = regexp.MustCompile(`(?s)## Recent Changes\n(.*?)(\n\n|$)`)
	lastUpdatedRE  = regexp.MustCompile(`Last updated: \d{4}-\d{2}-\d{2}`)
)

func patchAgentFile(t agentTarget, branch string, pf planFields) error {
	fmt.Printf("Updating existing %s context file...\n", t.name)

	data, err := os.ReadFile(t.file)
	if err != nil {
		return err
	}
	content := string(data)

	// Preserve manual additions block verbatim; everything else may be rewritten.
	manualBlock := ""
	if start := strings.Index(content, "<!-- MANUAL ADDITIONS START -->"); start >= 0 {
		if end := strings.Index(content, "<!-- MANUAL ADDITIONS END -->"); end >= 0 {
			manualBlock = content[start : end+len("<!-- MANUAL ADDITIONS END -->")]
		}
	}

	if m := activeTechRE.FindStringSubmatch(content); m != nil {
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
		var lines []string
		for _, l := range strings.Split(strings.TrimSpace(m[1]), "\n") {
			if l != "" {
				lines = append(lines, l)
			}
		}
		lines = append([]string{fmt.Sprintf("- %s: Added %s + %s", branch, pf.Lang, pf.Framework)}, lines...)
		if len(lines) > 3 {
			lines = lines[:3]
		}
		content = recentChangeRE.ReplaceAllString(content, "## Recent Changes\n"+strings.Join(lines, "\n")+"\n\n")
	}

	content = lastUpdatedRE.ReplaceAllString(content, "Last updated: "+time.Now().Format("2006-01-02"))

	if manualBlock != "" {
		if start := strings.Index(content, "<!-- MANUAL ADDITIONS START -->"); start >= 0 {
			if end := strings.Index(content, "<!-- MANUAL ADDITIONS END -->"); end >= 0 {
				content = content[:start] + manualBlock + content[end+len("<!-- MANUAL ADDITIONS END -->"):]
			}
		}
	}

	if err := os.WriteFile(t.file, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("✅ %s context file updated successfully\n", t.name)
	return nil
}
