package designlint

import (
	"fmt"
	"strings"
)

// lintTells runs the house-rule checks that a text scan can decide: emoji used as icons,
// and index numbering in visible HTML text. Both are WARN, because a brief can ask for a
// sequence or a symbol. The rules come from taste-skill's house rules (MIT; see
// THIRD_PARTY_NOTICES.md).
func lintTells(res *Result, file, text string, html bool) {
	for _, loc := range reEmoji.FindAllStringIndex(text, -1) {
		res.Findings = append(res.Findings, Finding{
			File: file, Line: lineOf(text, loc[0]), Rule: RuleEmoji, Severity: SevWarn,
			Message: fmt.Sprintf("symbol %q used as an icon or decoration: draw icons from one set; emoji and glyphs do not share a stroke", text[loc[0]:loc[1]]),
		})
	}
	if !html {
		return
	}
	visible := reHTMLNoise.ReplaceAllStringFunc(text, func(m string) string {
		return strings.Repeat("\n", strings.Count(m, "\n"))
	})
	visible = reTag.ReplaceAllStringFunc(visible, func(m string) string {
		return strings.Repeat(" ", len(m))
	})
	for _, loc := range reIndexNumber.FindAllStringIndex(visible, -1) {
		res.Findings = append(res.Findings, Finding{
			File: file, Line: lineOf(visible, loc[0]), Rule: RuleIndexNumber, Severity: SevWarn,
			Message: fmt.Sprintf("index numbering %q: keep it only if the content really is a sequence the reader needs", visible[loc[0]:loc[1]]),
		})
	}
}
