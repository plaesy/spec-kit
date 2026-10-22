package mdlint

import (
	"bytes"
	"strings"
)

// fixSource applies the machine-fixable rules to raw source. It works on text
// rather than on violations so a single pass converges: trailing whitespace, hard
// tabs, repeated blank lines, missing blank lines around headings/fences/lists/
// tables, and the final newline.
func fixSource(rs *ruleSet, src []byte) []byte {
	normalized := bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))
	lines := strings.Split(string(normalized), "\n")
	// A trailing newline leaves an empty final element that is not a real line.
	trailingNewline := false
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
		trailingNewline = true
	}
	if len(lines) == 0 {
		return src
	}

	enabled := map[string]bool{}
	for _, r := range rs.rules {
		enabled[r.id] = r.fixable
	}
	d := &doc{lines: lines}
	remask(d, lines)

	if enabled["MD010"] {
		o := opt[MD010Options](rs, "MD010")
		includeCode := o.IncludeCodeBlocks == nil || *o.IncludeCodeBlocks
		spaces := 4
		if o.SpacesPerTab != nil {
			spaces = *o.SpacesPerTab
		}
		for i := range lines {
			if d.inFrontMatter(i) || (d.inCode(i) && !includeCode) {
				continue
			}
			if strings.ContainsRune(lines[i], '\t') {
				lines[i] = strings.ReplaceAll(lines[i], "\t", strings.Repeat(" ", spaces))
			}
		}
	}

	if enabled["MD009"] {
		o := opt[MD009Options](rs, "MD009")
		allowed := 0
		if o.BrSpaces != nil {
			allowed = *o.BrSpaces
		}
		for i := range lines {
			if d.inFrontMatter(i) {
				continue
			}
			trimmed := strings.TrimRight(lines[i], " \t")
			if extra := len(lines[i]) - len(trimmed); extra > allowed {
				lines[i] = trimmed
			}
		}
	}

	if enabled["MD012"] {
		o := opt[MD012Options](rs, "MD012")
		max := 1
		if o.Maximum != nil {
			max = *o.Maximum
		}
		lines = collapseBlankRuns(lines, d, max)
	}

	// Spacing rules need the masks recomputed after each rewrite pass.
	remask(d, lines)

	if enabled["MD022"] {
		o := opt[MD022Options](rs, "MD022")
		above, below := 1, 1
		if o.LinesAbove != nil {
			above = *o.LinesAbove
		}
		if o.LinesBelow != nil {
			below = *o.LinesBelow
		}
		lines = applySpacing(lines, d, func() spacingPlan { return fixMD022(d, above, below) })
		remask(d, lines)
	}
	if enabled["MD031"] {
		lines = applySpacing(lines, d, func() spacingPlan { return fixMD031(d, 1, 1) })
		remask(d, lines)
	}
	if enabled["MD058"] {
		lines = applySpacing(lines, d, func() spacingPlan { return fixMD058(d, 1, 1) })
		remask(d, lines)
	}
	if enabled["MD032"] {
		lines = applySpacing(lines, d, func() spacingPlan { return fixMD032(d, 1, 1) })
		remask(d, lines)
	}

	out := strings.Join(lines, "\n")
	if enabled["MD047"] {
		out = strings.TrimRight(out, "\n") + "\n"
		trailingNewline = true
	}
	_ = trailingNewline
	return []byte(out)
}

// remask points the doc's line view at the current working lines and rebuilds
// every mask from it.
//
// The ensure* passes return a new slice rather than mutating in place, so without
// this the masks keep the length they had when the document was read while the
// loops iterate a longer slice — and the first direct index into a mask runs off
// the end. That is not theoretical: `--fix` panicked on the first real file in
// this repository, after the unit tests had passed on two-line fixtures where no
// pass ever grew the document.
func remask(d *doc, lines []string) {
	d.lines = lines
	d.markFrontMatter()
	d.markFences()
	d.markHeadings()
	d.markLists()
	d.markTables()
}

// collapseBlankRuns trims runs of blank lines down to maximum, outside code blocks.
func collapseBlankRuns(lines []string, d *doc, maximum int) []string {
	out := make([]string, 0, len(lines))
	run := 0
	for i, line := range lines {
		if !d.inFrontMatter(i) && strings.TrimSpace(line) == "" {
			run++
			if run > maximum {
				continue
			}
		} else {
			run = 0
		}
		out = append(out, line)
	}
	return out
}

// spacingPlan names the line indexes a spacing fix must separate: insert a blank
// line before each index in `before`, and after each index in `after`.
//
// A plan is computed from the same condition the corresponding rule reports on.
// That sharing is the point. The fixer used to answer "does this line continue
// the block above it?" with a predicate of its own — and it asked the question
// about the wrong line: it looked at the line a blank was about to be inserted
// before and asked whether *that* line looked like the start of a construct. A
// list item below a heading therefore always looked like a continuation of
// something, so the blank line MD022 and MD032 both ask for was never inserted.
// `--fix` printed "nothing to fix" on a file the linter had just listed 20
// fixable violations for, and the two halves disagreed in the one way nobody
// notices: a document that is equally clean before and after the fix.
type spacingPlan struct {
	before map[int]bool
	after  map[int]bool
}

func newSpacingPlan() spacingPlan {
	return spacingPlan{before: map[int]bool{}, after: map[int]bool{}}
}

// applySpacing inserts the blank lines a plan asks for, above first, then below.
//
// The plan is recomputed for the second pass. Line indexes are positions, and the
// first pass moves every position below its insertions — a plan computed once and
// applied twice inserts a second blank line in the gap the first pass just filled,
// which is how "# Title" ended up with two blank lines under it.
func applySpacing(lines []string, d *doc, plan func() spacingPlan) []string {
	if p := plan(); len(p.before) > 0 {
		lines = insertBlanksBefore(lines, d, p.before)
		remask(d, lines)
	}
	if p := plan(); len(p.after) > 0 {
		lines = insertBlanksAfter(lines, d, p.after)
	}
	return lines
}

// insertBlanksBefore inserts a blank line ahead of each flagged line, unless the
// line above it is already blank or belongs to front matter or a code block.
func insertBlanksBefore(lines []string, d *doc, at map[int]bool) []string {
	out := make([]string, 0, len(lines)+len(at))
	for i, line := range lines {
		if i > 0 && at[i] && !d.inFrontMatter(i) && !d.inCode(i-1) &&
			strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		out = append(out, line)
	}
	return out
}

// insertBlanksAfter inserts a blank line below each flagged line, unless the
// line below it is already blank or is past the end of the document.
func insertBlanksAfter(lines []string, d *doc, at map[int]bool) []string {
	out := make([]string, 0, len(lines)+len(at))
	for i, line := range lines {
		out = append(out, line)
		if i+1 < len(lines) && at[i] && !d.inFrontMatter(i+1) && !d.inCode(i+1) &&
			strings.TrimSpace(lines[i+1]) != "" {
			out = append(out, "")
		}
	}
	return out
}

// fixMD022 mirrors checkMD022: a blank line above every heading, and one below
// the line the heading ends on — its underline, for a setext heading.
func fixMD022(d *doc, linesAbove, linesBelow int) spacingPlan {
	plan := newSpacingPlan()
	if linesAbove > 0 {
		for i := 0; i < d.n(); i++ {
			if d.heading[i] == 0 {
				continue
			}
			// Nothing precedes the first heading in a document, so there is
			// nothing to separate.
			firstContent := i == 0 || (d.fmEnd > 0 && i == d.fmEnd)
			if !firstContent && strings.TrimSpace(d.lines[i-1]) != "" {
				plan.before[i] = true
			}
		}
	}
	if linesBelow > 0 {
		for i := 0; i < d.n(); i++ {
			if d.heading[i] == 0 {
				continue
			}
			end := i
			if d.headingEnd[i] > end {
				end = d.headingEnd[i]
			}
			if end+1 < d.n() && strings.TrimSpace(d.lines[end+1]) != "" && d.heading[end+1] == 0 {
				plan.after[end] = true
			}
		}
	}
	return plan
}

// fixMD031 mirrors checkMD031: a blank line before the fence that opens a block
// and after the one that closes it. The fence that opens a block is never
// separated from the code below it.
func fixMD031(d *doc, linesAbove, linesBelow int) spacingPlan {
	plan := newSpacingPlan()
	for i := 0; i < d.n(); i++ {
		if !d.fenceLines[i] {
			continue
		}
		if linesAbove > 0 && i > 0 && !d.fenceLines[i-1] &&
			strings.TrimSpace(d.lines[i-1]) != "" && !d.inCode(i-1) {
			plan.before[i] = true
		}
		if linesBelow <= 0 || d.fenceOpens[i] {
			continue
		}
		if i+1 < d.n() && !d.fenceLines[i+1] && strings.TrimSpace(d.lines[i+1]) != "" {
			plan.after[i] = true
		}
	}
	return plan
}

// fixMD032 mirrors checkMD032, including the rule's own definition of a list
// that continues from the line above: a nested item, a wrapped item, a line
// closing a blockquote, or the far side of a fence.
func fixMD032(d *doc, linesAbove, linesBelow int) spacingPlan {
	plan := newSpacingPlan()
	for i := 0; i < d.n(); i++ {
		if !d.list[i] {
			continue
		}
		if linesAbove > 0 && i > 0 {
			prev := d.lines[i-1]
			continues := d.list[i-1] || isIndented(prev) ||
				strings.HasSuffix(strings.TrimRight(prev, " "), ">") || d.fence[i-1] || d.fenceLines[i-1]
			if !continues && strings.TrimSpace(prev) != "" && !d.inCode(i-1) {
				plan.before[i] = true
			}
		}
		if linesBelow <= 0 || i+1 >= d.n() {
			continue
		}
		next := strings.TrimSpace(d.lines[i+1])
		follows := d.list[i+1] || isIndented(d.lines[i+1]) ||
			strings.HasPrefix(next, "- ") || strings.HasPrefix(next, "* ") ||
			strings.HasPrefix(next, "> ") || d.fenceLines[i+1] || d.fence[i+1]
		if !follows && next != "" {
			plan.after[i] = true
		}
	}
	return plan
}

// fixMD058 mirrors checkMD058: one blank line above the first row of a table and
// one below its last row.
func fixMD058(d *doc, linesAbove, linesBelow int) spacingPlan {
	plan := newSpacingPlan()
	seen := map[int]bool{}
	for i := 0; i < d.n(); i++ {
		if !d.table[i] || seen[i] {
			continue
		}
		if linesAbove > 0 && i > 0 && !d.table[i-1] &&
			strings.TrimSpace(d.lines[i-1]) != "" && !d.inCode(i-1) {
			plan.before[i] = true
		}
		j := i
		for j+1 < d.n() && d.table[j+1] {
			j++
		}
		seen[j] = true
		if linesBelow > 0 && j+1 < d.n() && !d.fenceLines[j+1] &&
			strings.TrimSpace(d.lines[j+1]) != "" {
			plan.after[j] = true
		}
	}
	return plan
}
