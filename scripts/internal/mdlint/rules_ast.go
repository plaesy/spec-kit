package mdlint

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// This file holds the AST rules: decisions that need the parsed document rather
// than the raw lines.

// ---- MD001 heading-increment ---------------------------------------------------

type MD001Options struct {
	HeadingLevels []int `json:"heading_levels"`
}

func checkMD001(d *doc, rs *ruleSet) []violation {
	_ = opt[MD001Options](rs, "MD001")
	previous := 0
	var out []violation
	d.walk(func(node ast.Node) bool {
		heading, ok := node.(*ast.Heading)
		if !ok {
			return true
		}
		level := heading.Level
		if previous > 0 && level > previous+1 {
			out = append(out, violation{"MD001", d.lineOf(node), 1,
				fmt.Sprintf("heading level jumps from h%d to h%d", previous, level)})
		}
		previous = level
		return false
	})
	return out
}

// ---- MD003 heading-style -------------------------------------------------------

type MD003Options struct {
	Style string `json:"style"`
}

func checkMD003(d *doc, rs *ruleSet) []violation {
	o := opt[MD003Options](rs, "MD003")
	style := o.Style
	if style == "" {
		style = "consistent"
	}
	var first string
	var out []violation
	d.walk(func(node ast.Node) bool {
		heading, ok := node.(*ast.Heading)
		if !ok {
			return true
		}
		kind := "atx"
		if !strings.HasPrefix(strings.TrimSpace(headingSource(d, heading)), "#") {
			kind = "setext"
		}
		if style == "consistent" {
			if first == "" {
				first = kind
			}
			if kind != first {
				out = append(out, violation{"MD003", d.lineOf(node), 1,
					"heading style " + kind + " is inconsistent with " + first})
			}
			return false
		}
		if kind != style {
			out = append(out, violation{"MD003", d.lineOf(node), 1, "heading style " + kind + " (style: " + style + ")"})
		}
		return false
	})
	return out
}

func headingSource(d *doc, heading *ast.Heading) string {
	seg := heading.Lines()
	if seg == nil || seg.Len() == 0 {
		return ""
	}
	return string(d.source[seg.At(0).Start:seg.At(0).Stop])
}

// ---- MD024 no-duplicate-heading ------------------------------------------------

type MD024Options struct {
	SiblingsOnly       *bool `json:"siblings_only"`
	AllowDifferentText *bool `json:"-"`
}

func checkMD024(d *doc, rs *ruleSet) []violation {
	o := opt[MD024Options](rs, "MD024")
	siblingsOnly := o.SiblingsOnly == nil || *o.SiblingsOnly
	seen := map[string]bool{}
	// path holds the current ancestor chain by heading level, so two headings
	// can be told apart by the sections they sit in.
	path := map[int]string{}
	var out []violation
	d.walk(func(node ast.Node) bool {
		heading, ok := node.(*ast.Heading)
		if !ok {
			return true
		}
		text := strings.ToLower(strings.TrimSpace(headingText(heading, d.source)))
		key := text
		msg := "duplicate heading \"" + text + "\" (this level already has it)"
		if siblingsOnly {
			// Keying on level and text alone compares every heading of a level
			// against every other, so a `### Subcommands` under one section
			// collided with the `### Subcommands` under the next. Siblings are
			// the headings that share a parent, so the ancestor chain is part
			// of the key.
			var anc []string
			for lvl := 1; lvl < heading.Level; lvl++ {
				if t, ok := path[lvl]; ok {
					anc = append(anc, t)
				}
			}
			key = strings.Join(append(anc, text), "\x00")
			msg = "duplicate heading \"" + text + "\" (a sibling under the same parent already uses it)"
		}
		path[heading.Level] = text
		if seen[key] {
			out = append(out, violation{"MD024", d.lineOf(node), 1, msg})
			return false
		}
		seen[key] = true
		return false
	})
	return out
}

// headingText concatenates a heading's inline text. The source bytes are
// required: goldmark text segments store offsets, not strings.
func headingText(heading *ast.Heading, source []byte) string {
	return nodeText(heading, source)
}

// ---- MD025 single-title --------------------------------------------------------

func checkMD025(d *doc, rs *ruleSet) []violation {
	o := opt[MD025Options](rs, "MD025")
	level := 1
	if o.Level != nil {
		level = *o.Level
	}
	if o.FrontMatterTitle != "" {
		frontMatterTitlePattern = o.FrontMatterTitle
	}
	frontMatterCounts := false
	if _, ok := d.frontMatterTitle(); ok {
		frontMatterCounts = frontMatterTitlePattern != "^$"
	}
	count := 0
	var out []violation
	if frontMatterCounts {
		count++
	}
	d.walk(func(node ast.Node) bool {
		heading, ok := node.(*ast.Heading)
		if !ok {
			return true
		}
		// goldmark parses YAML front matter as body markdown, so every `#`
		// comment in it is a real ast.Heading. Counting those reports the
		// front matter as headings -- 9 MD025s in templates/design.template.md,
		// all of them YAML comments, with no fix available from inside the
		// template since `#` is the only comment sigil YAML has. lineOf is
		// 1-based; inFrontMatter takes a 0-based index.
		if d.inFrontMatter(d.lineOf(node) - 1) {
			return false
		}
		if heading.Level == level {
			count++
			if count > 1 && !frontMatterCounts {
				out = append(out, violation{"MD025", d.lineOf(node), 1,
					fmt.Sprintf("multiple h%d headings (a document may have only one)", level)})
			}
		}
		return false
	})
	if count == 0 && !frontMatterCounts {
		out = append(out, violation{"MD025", 1, 1,
			fmt.Sprintf("no h%d heading in the document", level)})
	}
	return out
}

// ---- MD036 no-emphasis-as-heading ----------------------------------------------

type MD036Options struct {
	Punctuation *string `json:"punctuation"`
}

// defaultMD036Punctuation matches markdownlint's own default EXACTLY
// (".,;:!?。，；：！？"). The trailing ? and ？ are part of upstream's set and were
// missing from an earlier version of this constant here, which silently failed
// to exempt a question-shaped label such as "**How might we {{X}}?**". A
// default that is a subset of upstream's is still a wrong default: it reports
// violations a conformant markdownlint run would not.
const defaultMD036Punctuation = ".,;:!?。，；：！？"

// startsOwnParagraph reports whether line i is the first line of its own
// paragraph, which is what MD036 actually asks about: emphasis standing in for
// a heading is a paragraph that is *only* emphasis, not a line that happens to
// hold nothing but emphasis.
//
// A line with text above it and no blank between is a paragraph continuation
// ("Intro sentence\n*emphasis*"), and a line directly under a list item is a
// lazy continuation of that item's paragraph. Both are ordinary prose. The
// block-level lines that do end a paragraph are a heading, a fence delimiter, a
// table row, a blockquote and a thematic break.
func startsOwnParagraph(d *doc, i int) bool {
	if i == 0 {
		return true
	}
	if d.heading[i-1] > 0 || d.fenceLines[i-1] || d.table[i-1] {
		return true
	}
	prev := strings.TrimSpace(d.lines[i-1])
	if prev == "" || strings.HasPrefix(prev, ">") {
		return true
	}
	return prev == "---" || prev == "***" || prev == "___"
}

func checkMD036(d *doc, rs *ruleSet) []violation {
	// MD036Options.Punctuation used to be declared and parsed and then dropped
	// on the floor (`_ = opt[MD036Options](rs, "MD036")`), so setting it in
	// .markdownlint.json changed nothing while appearing to work. An option
	// that fails to apply is indistinguishable from one that did.
	punctuation := defaultMD036Punctuation
	if o := opt[MD036Options](rs, "MD036"); o.Punctuation != nil {
		punctuation = *o.Punctuation
	}
	var out []violation
	// A whole paragraph made of a single emphasized run, immediately followed by
	// a blank line, is a heading written with emphasis.
	for i := 0; i < d.n(); i++ {
		if d.inFrontMatter(i) || d.inCode(i) || d.heading[i] > 0 {
			continue
		}
		trimmed := strings.TrimSpace(d.lines[i])
		if !strings.HasPrefix(trimmed, "*") || !strings.HasSuffix(trimmed, "*") || len(trimmed) < 4 {
			continue
		}
		inner := strings.TrimSpace(strings.Trim(trimmed, "*"))
		if inner == "" || strings.ContainsAny(inner, "*") {
			continue
		}
		if r := []rune(inner); punctuation != "" && strings.ContainsRune(punctuation, r[len(r)-1]) {
			continue
		}
		if !startsOwnParagraph(d, i) {
			continue
		}
		if i+1 < d.n() && strings.TrimSpace(d.lines[i+1]) == "" {
			out = append(out, violation{"MD036", i + 1, 1,
				"emphasis used instead of a heading"})
		}
	}
	return out
}

// ---- MD041 first-line-heading --------------------------------------------------

func checkMD041(d *doc, rs *ruleSet) []violation {
	if d.n() == 0 {
		return nil
	}
	// Front matter is skipped: the first content line decides.
	start := d.fmEnd
	for start < d.n() && strings.TrimSpace(d.lines[start]) == "" {
		start++
	}
	if start >= d.n() {
		return nil
	}
	line := d.lines[start]
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") {
		return nil
	}
	if d.heading[start] > 0 {
		return nil
	}
	if isSetextUnderline(trimmed) {
		return nil
	}
	return []violation{{"MD041", start + 1, 1, "first content line is not a top-level heading"}}
}

// ---- MD046 code-block-style ----------------------------------------------------

type MD046Options struct {
	Style string `json:"style"`
}

func checkMD046(d *doc, rs *ruleSet) []violation {
	o := opt[MD046Options](rs, "MD046")
	style := o.Style
	if style == "" {
		style = "consistent"
	}
	var out []violation
	for i := 0; i < d.n(); i++ {
		if d.inFrontMatter(i) || d.inCode(i) {
			continue
		}
		// An indented code block is four spaces at the start of a line that is
		// not inside a list.
		if strings.HasPrefix(d.lines[i], "    ") && strings.TrimSpace(d.lines[i]) != "" {
			prevBlank := i == 0 || strings.TrimSpace(d.lines[i-1]) == ""
			prevList := i > 0 && d.list[i-1]
			if prevBlank && !prevList {
				out = append(out, violation{"MD046", i + 1, 1, "indented code block (style: " + style + ")"})
			}
		}
	}
	return out
}

// ---- MD048 code-fence-style ----------------------------------------------------

type MD048Options struct {
	Style string `json:"style"`
}

func checkMD048(d *doc, rs *ruleSet) []violation {
	o := opt[MD048Options](rs, "MD048")
	style := o.Style
	if style == "" {
		style = "backtick"
	}
	var out []violation
	for i := 0; i < d.n(); i++ {
		if !d.fenceLines[i] {
			continue
		}
		trimmed := strings.TrimLeft(d.lines[i], " ")
		if trimmed == "" {
			continue
		}
		if style == "tilde" && trimmed[0] != '~' {
			out = append(out, violation{"MD048", i + 1, 1, "code fence style is backtick (style: tilde)"})
		}
		if style == "backtick" && trimmed[0] == '~' {
			out = append(out, violation{"MD048", i + 1, 1, "code fence style is tilde (style: backtick)"})
		}
	}
	return out
}
