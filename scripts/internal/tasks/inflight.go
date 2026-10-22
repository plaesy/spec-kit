package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InFlightHeading is the section of .plaesy/context.md that lists active tasks.
const InFlightHeading = "## In-Flight Tasks"

// inFlightStatuses are the statuses that count as in flight: next up, and in progress.
var inFlightStatuses = map[string]bool{"todo": true, "doing": true}

// RenderInFlight returns the In-Flight section body for the given tasks, one line per
// task, in the order given. Tasks in any other status are left out.
func RenderInFlight(refs []TaskRef) string {
	var b strings.Builder
	for _, r := range refs {
		if !inFlightStatuses[r.Status] {
			continue
		}
		fmt.Fprintf(&b, "- [%s] %s (%s): %s\n", r.Spec, r.Name, r.Status, r.Title)
	}
	if b.Len() == 0 {
		return "- none\n"
	}
	return b.String()
}

// ReplaceSection replaces the body under heading with body, up to the next "## "
// heading or the end of the file. The heading line itself is kept. If the heading is
// missing, the section is appended. Line endings of content are preserved.
func ReplaceSection(content, heading, body string) string {
	eol := "\n"
	if strings.Contains(content, "\r\n") {
		eol = "\r\n"
	}
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	bodyLines := strings.Split(strings.TrimRight(body, "\n"), "\n")

	start := -1
	for i, line := range lines {
		if strings.TrimRight(line, " \t") == heading {
			start = i
			break
		}
	}
	if start < 0 {
		trimmed := strings.TrimRight(strings.Join(lines, "\n"), "\n")
		return trimmed + "\n\n" + heading + "\n" + strings.Join(bodyLines, "\n") + "\n"
	}

	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}

	var out []string
	out = append(out, lines[:start+1]...)
	out = append(out, bodyLines...)
	out = append(out, "")
	out = append(out, lines[end:]...)
	result := strings.Join(out, "\n")
	result = strings.TrimRight(result, "\n") + "\n"
	return strings.ReplaceAll(result, "\n", eol)
}

// SyncInFlight rewrites the In-Flight Tasks section of <projectRoot>/.plaesy/context.md
// from the task files under specsDir. It reports whether the file changed. Only that
// section is replaced; the rest of the file is kept byte for byte.
func SyncInFlight(projectRoot string, refs []TaskRef) (changed bool, err error) {
	path := filepath.Join(projectRoot, ".plaesy", "context.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	updated := ReplaceSection(string(data), InFlightHeading, RenderInFlight(refs))
	if updated == string(data) {
		return false, nil
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return false, err
	}
	return true, nil
}
