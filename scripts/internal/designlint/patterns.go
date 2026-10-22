package designlint

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Rule identifiers for the pattern checks. They are stable, like the rules above.
const (
	RuleGradientText = "DL-GRADIENT-TEXT"
	RuleSideStripe   = "DL-SIDE-STRIPE"
	RuleGlowShadow   = "DL-GLOW-SHADOW"
	RuleHardShadow   = "DL-HARD-SHADOW"
	RuleEmoji        = "DL-EMOJI"
	RuleIndexNumber  = "DL-INDEX-NUMBER"
)

// The pattern checks read declarations the way the other CSS checks do: text, with
// comments blanked out. Each hit is a WARN, because each one is a default that a
// brief can justify. The frontend-design rules name these as tells of a generated page.
var (
	reGradientTextClip = regexp.MustCompile(`(?i)(?:-webkit-)?background-clip\s*:\s*text\b|backgroundClip\s*:\s*['"]text['"]`)
	reSideStripe       = regexp.MustCompile(`(?i)border-(?:left|right|inline-start|inline-end)\s*:\s*([0-9]*\.?[0-9]+)px`)
	reGlowShadow       = regexp.MustCompile(`(?i)box-shadow\s*:\s*0(?:px)?\s+0(?:px)?\s+([0-9]*\.?[0-9]+)px\s+(?:-?[0-9.]+px\s+)?(?:rgba?\(|hsla?\(|#)`)
	reEmoji            = regexp.MustCompile(`[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}]`)
	reIndexNumber      = regexp.MustCompile(`\b0[1-9]\s*/\s*0[1-9]\b|\(0\d{2}\)|\bStep\s+[0-9]+\b`)
	reHardShadow       = regexp.MustCompile(`(?i)box-shadow\s*:\s*(-?[0-9]*\.?[0-9]+)px\s+(-?[0-9]*\.?[0-9]+)px\s+0(?:px)?\s+(?:-?[0-9.]+px\s+)?(?:rgba?\(|hsla?\(|#)`)
)

// blankCSSComments replaces each /* ... */ comment with spaces, keeping newlines and
// byte offsets, so line numbers from the match still point at the right line.
func blankCSSComments(text string) string {
	return reCSSComment.ReplaceAllStringFunc(text, func(m string) string {
		return strings.Map(func(r rune) rune {
			if r == '\n' {
				return '\n'
			}
			return ' '
		}, m)
	})
}

// lintPatterns runs the four pattern checks over one file's text.
func lintPatterns(res *Result, file, text string) {
	text = blankCSSComments(text)

	for _, loc := range reGradientTextClip.FindAllStringIndex(text, -1) {
		res.Findings = append(res.Findings, Finding{
			File: file, Line: lineOf(text, loc[0]), Rule: RuleGradientText, Severity: SevWarn,
			Message: "gradient text (background-clip: text): emphasis should come from weight or size; keep a gradient only if the brief pins it",
		})
	}

	for _, m := range reSideStripe.FindAllStringSubmatchIndex(text, -1) {
		px, err := strconv.ParseFloat(text[m[2]:m[3]], 64)
		if err != nil || px <= 1 {
			continue // a 1px hairline is a divider, not a stripe
		}
		res.Findings = append(res.Findings, Finding{
			File: file, Line: lineOf(text, m[0]), Rule: RuleSideStripe, Severity: SevWarn,
			Message: fmt.Sprintf("side-stripe border of %gpx on one edge: use a full border, a background tint, or no accent", px),
		})
	}

	for _, m := range reGlowShadow.FindAllStringSubmatchIndex(text, -1) {
		res.Findings = append(res.Findings, Finding{
			File: file, Line: lineOf(text, m[0]), Rule: RuleGlowShadow, Severity: SevWarn,
			Message: "zero-offset coloured halo (box-shadow with no x/y offset): shadows need an offset and a soft blur to read as depth",
		})
	}

	for _, m := range reHardShadow.FindAllStringSubmatchIndex(text, -1) {
		x, errX := strconv.ParseFloat(text[m[2]:m[3]], 64)
		y, errY := strconv.ParseFloat(text[m[4]:m[5]], 64)
		if errX != nil || errY != nil || (x == 0 && y == 0) {
			continue
		}
		res.Findings = append(res.Findings, Finding{
			File: file, Line: lineOf(text, m[0]), Rule: RuleHardShadow, Severity: SevWarn,
			Message: "hard offset shadow (zero blur): a block shadow is a costume unless the whole design is built on it",
		})
	}
}
