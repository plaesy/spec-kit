package validate

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// ConstitutionResult carries the findings of Constitution: what the document
// declares, so a caller can report the active scope and not only pass/fail.
type ConstitutionResult struct {
	Path             string
	Version          string
	ActiveDimensions []string
	Rules            int // rows in the Universal Rules table (§4)
	NonNegotiables   int // rows in the Non-Negotiables table (§5)
}

// knownDimensions is the closed set of dimensions the framework routes
// (/assess:{dimension}, /implement:{dimension}, ...). A row outside this set
// would be invisible to every prompt, so it is a validation failure rather
// than a silently ignored line.
//
// "operations" is the 9th dimension (deployment, release, reliability/SLO,
// observability, runbooks, IaC), split out of "technical", which otherwise
// claimed infrastructure/deployment in its scope. The sibling rows are kept
// deliberately: an Operations row in a constitution that does not ship the
// matching prompts/{assess,implement,fix,optimize,loop,improve}/operations.md
// files is a known transitional state, not a reason to reject the document.
var knownDimensions = []string{
	"technical", "design", "business", "product",
	"marketing", "legal", "financial", "management",
	"operations",
}

var (
	isoDateRE   = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	ruleIDRE    = regexp.MustCompile(`^(EV|QG|DOC|SC|RTM|NN)-\d+$`)
	frontmatter = regexp.MustCompile(`(?m)^---\r?\n`)
)

// Constitution validates a generated project constitution
// (.plaesy/memory/constitution.md). It enforces the contract the template
// states about itself, so a half-filled or drifted document cannot be read as
// governing authority:
//
//  1. the file exists (an absent constitution is unratified, not a pass)
//  2. frontmatter carries title, version, ratified, last_amended, active_dimensions
//  3. ratified/last_amended are ISO 8601 dates
//  4. no unfilled {{ ... }} placeholder survives (same rule as the OOXML validators)
//  5. every §1 Active Dimensions row is a known dimension marked exactly yes/no,
//     and at least one is active
//  6. frontmatter active_dimensions equals the set of §1 rows marked yes — the
//     template names the table as the source, so a disagreement is a defect
//  7. the frontmatter version equals the last row of the amendment log
//  8. rule IDs (EV/QG/DOC/SC/RTM/NN-NN) are unique across the document
//
// Every problem found is reported at once, so one run surfaces the whole list
// instead of making the author fix them one round-trip at a time.
func Constitution(path string) (*ConstitutionResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("constitution not found at %s: not ratified yet — run /start (Phase 0) to generate it from templates/constitution.template.md", path)
	}
	body := string(data)

	res := &ConstitutionResult{Path: path}
	var problems []string

	front, rest, ok := splitFrontmatter(body)
	if !ok {
		return nil, fmt.Errorf("constitution %s has no YAML frontmatter (expected a leading '---' block with title, version, ratified, last_amended, active_dimensions)", path)
	}
	fields := parseFrontmatter(front)
	for _, key := range []string{"title", "version", "ratified", "last_amended", "active_dimensions"} {
		if strings.TrimSpace(fields[key]) == "" {
			problems = append(problems, fmt.Sprintf("frontmatter key %q is missing or empty", key))
		}
	}
	for _, key := range []string{"ratified", "last_amended"} {
		if val := strings.Trim(fields[key], `"'`); val != "" && !isoDateRE.MatchString(val) {
			problems = append(problems, fmt.Sprintf("frontmatter %s = %q is not an ISO 8601 date (YYYY-MM-DD)", key, val))
		}
	}
	res.Version = strings.Trim(fields["version"], `"'`)

	if leftover := jinjaTagRE.FindAllString(body, -1); len(leftover) > 0 {
		problems = append(problems, fmt.Sprintf("unfilled template placeholders remain: %s", summarize(leftover)))
	}

	tableActive, tableErr := parseActiveDimensions(rest)
	if tableErr != nil {
		problems = append(problems, tableErr.Error())
	}
	res.ActiveDimensions = tableActive
	if len(tableActive) == 0 && tableErr == nil {
		problems = append(problems, "no active dimension: mark at least one §1 row 'yes' (a constitution that governs nothing is not a constitution)")
	}

	declared := parseDimensionList(fields["active_dimensions"])
	if len(declared) == 0 && strings.TrimSpace(fields["active_dimensions"]) == "" {
		// already reported as a missing key above
	} else if tableErr == nil {
		if got, want := declared, tableActive; !equalStrings(got, want) {
			problems = append(problems, fmt.Sprintf("frontmatter active_dimensions [%s] does not match the §1 rows marked yes [%s] — the table is the source; make them agree",
				strings.Join(got, ", "), strings.Join(want, ", ")))
		}
	}

	if v := lastLoggedVersion(rest); v != "" && res.Version != "" && v != res.Version {
		problems = append(problems, fmt.Sprintf("version drift: frontmatter says %q but the last amendment-log row says %q — the frontmatter must match the log", res.Version, v))
	}

	res.Rules, res.NonNegotiables = countRuleTables(rest, &problems)

	if len(problems) > 0 {
		sort.Strings(problems)
		return res, fmt.Errorf("constitution %s failed validation:\n  - %s", path, strings.Join(problems, "\n  - "))
	}
	return res, nil
}

// splitFrontmatter returns the YAML block between the leading '---' fences and
// the remainder of the document.
func splitFrontmatter(body string) (front, rest string, ok bool) {
	if !strings.HasPrefix(body, "---") {
		return "", body, false
	}
	parts := frontmatter.Split(body, 3)
	if len(parts) < 3 {
		return "", body, false
	}
	// parts[0] is empty (the first fence), parts[1] the block, parts[2] the body.
	return parts[1], parts[2], true
}

func parseFrontmatter(front string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(front, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		out[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}
	return out
}

// parseDimensionList reads `active_dimensions: [technical, business]`.
func parseDimensionList(raw string) []string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if p := strings.ToLower(strings.Trim(strings.TrimSpace(part), `"'`)); p != "" {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

// parseActiveDimensions reads the §1 table and returns the dimensions marked
// yes. A returned error means the table itself is unusable (missing section,
// malformed row, unknown dimension, or a cell that is not exactly yes/no).
func parseActiveDimensions(body string) ([]string, error) {
	known := map[string]bool{}
	for _, d := range knownDimensions {
		known[d] = true
	}

	_, rows, err := tableAfterHeading(body, "Active Dimensions")
	if err != nil {
		return nil, err
	}

	var active []string
	for _, cells := range rows {
		if len(cells) < 2 {
			return nil, fmt.Errorf("§1 Active Dimensions row has %d cell(s), need at least a dimension and an Active cell: %q", len(cells), strings.Join(cells, " | "))
		}
		name := strings.ToLower(cells[0])
		flag := strings.ToLower(cells[1])
		if !known[name] {
			return nil, fmt.Errorf("§1 Active Dimensions row %q is an unknown dimension (known: %s) — a row outside the closed set is invisible to every prompt",
				cells[0], strings.Join(knownDimensions, ", "))
		}
		if flag != "yes" && flag != "no" {
			return nil, fmt.Errorf("§1 Active Dimensions cell for %q must be yes or no, got %q", cells[0], cells[1])
		}
		if flag == "yes" {
			active = append(active, name)
		}
	}
	sort.Strings(active)
	return active, nil
}

// lastLoggedVersion returns the version in the final data row of the amendment
// log, or "" when the log has no Version column or no data rows.
func lastLoggedVersion(body string) string {
	header, rows, err := tableAfterHeading(body, "Amendment Log")
	if err != nil {
		return ""
	}
	idx := -1
	for i, cell := range header {
		if strings.EqualFold(strings.TrimSpace(cell), "version") {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ""
	}
	last := rows[len(rows)-1]
	if idx >= len(last) {
		return ""
	}
	return strings.TrimSpace(last[idx])
}

// countRuleTables counts the §4 rule rows and §5 non-negotiable rows and
// reports any duplicate rule ID.
func countRuleTables(body string, problems *[]string) (rules, nonNegotiables int) {
	seen := map[string]bool{}
	for _, cells := range allRuleRows(body) {
		if len(cells) == 0 {
			continue
		}
		id := strings.TrimSpace(cells[0])
		if !ruleIDRE.MatchString(id) {
			continue
		}
		if seen[id] {
			*problems = append(*problems, fmt.Sprintf("duplicate rule ID %q — IDs are cited by phases and findings, so they must be unique", id))
			continue
		}
		seen[id] = true
		if strings.HasPrefix(id, "NN-") {
			nonNegotiables++
		} else {
			rules++
		}
	}
	return rules, nonNegotiables
}

// allRuleRows returns every table data row in the document whose first cell
// looks like a rule ID, plus the header row of each table it appears in, so
// callers needing a header can pair them positionally.
func allRuleRows(body string) [][]string {
	var out [][]string
	lines := strings.Split(body, "\n")
	inTable := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") {
			if !inTable {
				out = append(out, splitRow(trimmed))
				inTable = true
				continue
			}
			if isSeparatorRow(trimmed) {
				continue
			}
			out = append(out, splitRow(trimmed))
			continue
		}
		inTable = false
	}
	return out
}

// tableAfterHeading returns the header cells and the data rows of the first
// markdown table that follows a heading mentioning name. The header and the
// data are returned separately so callers cannot accidentally validate the
// header row as if it were content.
func tableAfterHeading(body, name string) (header []string, data [][]string, err error) {
	lines := strings.Split(body, "\n")
	lower := strings.ToLower(name)
	found := false
	inTable := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !found {
			if strings.HasPrefix(trimmed, "#") && strings.Contains(strings.ToLower(trimmed), lower) {
				found = true
			}
			continue
		}
		if strings.HasPrefix(trimmed, "|") {
			if !inTable {
				inTable = true
				header = splitRow(trimmed)
				continue
			}
			if isSeparatorRow(trimmed) {
				continue
			}
			data = append(data, splitRow(trimmed))
			continue
		}
		if inTable {
			break // the first table after the heading is the one we mean
		}
	}
	if !found {
		return nil, nil, fmt.Errorf("no heading containing %q found — the constitution must keep its '## 1. Active Dimensions' section", name)
	}
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("no data rows in the table under %q", name)
	}
	return header, data, nil
}

// splitRow splits one markdown table row into trimmed cells. A backslash-escaped
// pipe (`\|`) is a literal pipe inside a cell, not a column separator — required
// because placeholder defaults such as {{VAR|default}} are written escaped when
// they appear in a table cell.
func splitRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")

	var cells []string
	var cur strings.Builder
	escaped := false
	for _, r := range line {
		switch {
		case escaped:
			if r != '|' {
				cur.WriteRune('\\')
			}
			cur.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == '|':
			cells = append(cells, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	if escaped {
		cur.WriteRune('\\')
	}
	cells = append(cells, strings.TrimSpace(cur.String()))
	return cells
}

func isSeparatorRow(line string) bool {
	for _, cell := range splitRow(line) {
		if cell != "" && strings.Trim(cell, "-: ") != "" {
			return false
		}
	}
	return true
}

func trimmed(s string) string { return strings.TrimSpace(s) }

// summarize dedupes placeholders and caps the list, so validating a raw
// template (dozens of placeholders by design) still prints a readable error.
func summarize(in []string) string {
	const max = 10
	uniq := dedupe(in)
	if len(uniq) <= max {
		return strings.Join(uniq, ", ")
	}
	return fmt.Sprintf("%s, … and %d more", strings.Join(uniq[:max], ", "), len(uniq)-max)
}

// equalStrings compares two string slices element-wise (Go 1.20 has no slices.Equal).
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
