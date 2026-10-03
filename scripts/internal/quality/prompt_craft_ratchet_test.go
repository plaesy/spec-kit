package quality

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// PromptCraftBaselineName is the recorded floor this ratchet enforces.
const PromptCraftBaselineName = "prompt-craft-baseline.json"

var updatePromptCraftBaseline = flag.Bool("update-prompt-craft-baseline", false,
	"re-measure prompt craft and rewrite "+PromptCraftBaselineName+" with the result")

// Every other guard in this package measures FORM: markdown violations, legacy
// marker counts, mirror parity, metadata counts. None of them measures whether a
// prompt is any good. That gap is why a clean lint run could be reported as
// "prompt quality 100/100" without anything having been falsified.
//
// The criteria encoded here come from published prompt-engineering guidance
// (Anthropic's context-engineering and prompt-engineering write-ups):
//   - structure: a prompt declares an objective and an output contract, so the
//     agent knows what it is producing, not just what to do;
//   - no restated intent: a section named "Protocol" that merely repeats the
//     Objective misleads an agent skimming for the procedure;
//   - canonical examples: worked examples are the highest-leverage part of a
//     prompt, so their coverage is tracked and may not shrink.
//
// Counts are recorded in prompt-craft-baseline.json and may only improve.
// Pass -update-prompt-craft-baseline to re-record after a deliberate pass.

// routers dispatch to other prompts and hold no procedure of their own, so the
// protocol/example criteria do not apply to them.
var promptRouters = map[string]bool{
	"create.md": true,
	"spec.md":   true,
}

// outputContractSections are the headings this corpus has used to say "here is
// what you produce". Accepting any of them keeps the check honest about intent
// rather than mandating one spelling.
var outputContractSections = []string{
	"output format", "output contract", "completion format",
	"completion report", "definition of done", "report format",
}

// headingRe matches a heading at ANY depth. It originally matched only `##`,
// which undercounted optimize.md as having no worked example because its
// example sits under a deeper heading -- a real example, missed by the
// instrument. The same blindness applied to the objective and output-contract
// checks: any prompt nesting those below h2 would have been reported as
// missing them. A metric that only looks one level down is not measuring the
// thing it claims.
var headingRe = regexp.MustCompile(`^#{1,6}\s+(.+?)\s*$`)

// normalize reduces prose to comparable form: lowercase, markdown punctuation
// and placeholder tokens removed, whitespace collapsed. It is deliberately
// lossy -- the question is only whether two sections say the same thing.
func normalize(s string) string {
	s = strings.ToLower(s)
	s = regexp.MustCompile(`\{\{[^}]*\}\}`).ReplaceAllString(s, " ")
	s = regexp.MustCompile("[`*_#>|-]").ReplaceAllString(s, " ")
	return strings.Join(strings.Fields(s), " ")
}

// section returns the body of the first h2 whose heading matches want.
func section(lines []string, want func(string) bool) (string, bool) {
	start := -1
	for i, l := range lines {
		m := headingRe.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		if start >= 0 {
			return strings.Join(lines[start:i], "\n"), true
		}
		if want(strings.ToLower(m[1])) {
			start = i + 1
		}
	}
	if start >= 0 {
		return strings.Join(lines[start:], "\n"), true
	}
	return "", false
}

type promptMetrics struct {
	Prompts        int      `json:"prompts"`
	WithObjective  int      `json:"with_objective"`
	WithOutput     int      `json:"with_output_contract"`
	WithExample    int      `json:"with_worked_example"`
	RestatedInFile []string `json:"protocol_restates_objective,omitempty"`
}

func measurePrompts(t *testing.T) promptMetrics {
	t.Helper()
	dir := "../../../prompts"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read prompts dir: %v", err)
	}
	var m promptMetrics
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") || promptRouters[name] {
			continue
		}
		m.Prompts++
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		lines := strings.Split(string(raw), "\n")

		obj, hasObj := section(lines, func(h string) bool { return h == "objective" })
		if hasObj {
			m.WithObjective++
		}
		_, hasOut := section(lines, func(h string) bool {
			for _, c := range outputContractSections {
				if strings.Contains(h, c) {
					return true
				}
			}
			return false
		})
		if hasOut {
			m.WithOutput++
		}
		_, hasEx := section(lines, func(h string) bool { return strings.Contains(h, "example") })
		if hasEx {
			m.WithExample++
		}

		// The defect this check exists for: a "Protocol" section that says
		// nothing the Objective did not already say. Substring containment in
		// either direction catches the common "same sentence, shorter" shape.
		if hasObj {
			if proto, ok := section(lines, func(h string) bool { return h == "protocol" }); ok {
				p, o := normalize(proto), normalize(obj)
				if p != "" && (strings.Contains(o, p) || strings.Contains(p, o)) {
					m.RestatedInFile = append(m.RestatedInFile, name)
				}
			}
		}
	}
	return m
}

func TestPromptCraftDoesNotRegress(t *testing.T) {
	got := measurePrompts(t)
	path := filepath.Join("prompt-craft-baseline.json")

	if *updatePromptCraftBaseline {
		writePromptCraftBaseline(t, path, got)
		t.Logf("recorded: %d prompts, %d objective, %d output contract, %d example, restated=%v",
			got.Prompts, got.WithObjective, got.WithOutput, got.WithExample, got.RestatedInFile)
		return
	}

	prev := readPromptCraftBaseline(t, path)
	if got.WithObjective < prev.WithObjective {
		t.Errorf("prompts with an Objective: %d, was %d", got.WithObjective, prev.WithObjective)
	}
	if got.WithOutput < prev.WithOutput {
		t.Errorf("prompts with an output contract: %d, was %d", got.WithOutput, prev.WithOutput)
	}
	if got.WithExample < prev.WithExample {
		t.Errorf("prompts with a worked example: %d, was %d", got.WithExample, prev.WithExample)
	}
	if len(got.RestatedInFile) > len(prev.RestatedInFile) {
		t.Errorf("protocol-restates-objective: %d file(s) %v, was %d %v",
			len(got.RestatedInFile), got.RestatedInFile,
			len(prev.RestatedInFile), prev.RestatedInFile)
	}
}

// Without a recorded baseline there is no ratchet, and a bar that only exists in
// prose is a wish. Failing loudly here is deliberate: an absent baseline must not
// be read as a passing one.
func readPromptCraftBaseline(t *testing.T, path string) promptMetrics {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v\n\nWithout a recorded baseline there is no ratchet, and a bar "+
			"that only exists in prose is a wish. Record one with:\n\n"+
			"    go test ./internal/quality -run TestPromptCraftDoesNotRegress -update-prompt-craft-baseline",
			path, err)
	}
	var b promptMetrics
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if b.Prompts == 0 && !*updatePromptCraftBaseline {
		t.Fatalf("%s records no prompts: the ratchet would pass on an empty measurement", path)
	}
	return b
}

func writePromptCraftBaseline(t *testing.T, path string, m promptMetrics) {
	t.Helper()
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("encode %s: %v", PromptCraftBaselineName, err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write %s: %v", PromptCraftBaselineName, err)
	}
}
