package designlint

import "testing"

func TestPatternGradientTextWarns(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"style.css": "h1 {\n  background: linear-gradient(90deg, #f00, #00f);\n  -webkit-background-clip: text;\n  color: transparent;\n}\n",
	})
	res := lintDir(t, dir, Options{})
	if got := res.Count(RuleGradientText, SevWarn); got != 1 {
		t.Fatalf("want 1 %s WARN, got %d: %+v", RuleGradientText, got, res.Findings)
	}
}

func TestPatternGradientTextIgnoresPlainColour(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"style.css": "h1 { color: #111; background: #fff; }\n",
	})
	res := lintDir(t, dir, Options{})
	if got := res.Count(RuleGradientText, SevWarn); got != 0 {
		t.Fatalf("want 0 %s findings, got %d", RuleGradientText, got)
	}
}

func TestPatternSideStripeWarnsOverOnePixel(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"style.css": ".card { border-left: 4px solid #333; }\n.note { border-inline-start: 3px solid red; }\n",
	})
	res := lintDir(t, dir, Options{})
	if got := res.Count(RuleSideStripe, SevWarn); got != 2 {
		t.Fatalf("want 2 %s WARN, got %d: %+v", RuleSideStripe, got, res.Findings)
	}
}

func TestPatternSideStripeAllowsHairlineAndFullBorder(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"style.css": ".row { border-left: 1px solid #ccc; }\n.card { border-top: 4px solid #333; }\n",
	})
	res := lintDir(t, dir, Options{})
	if got := res.Count(RuleSideStripe, SevWarn); got != 0 {
		t.Fatalf("want 0 %s findings, got %d", RuleSideStripe, got)
	}
}

func TestPatternGlowShadowWarnsOnZeroOffsetColour(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"style.css": ".btn { box-shadow: 0 0 24px rgba(99, 102, 241, 0.6); }\n",
	})
	res := lintDir(t, dir, Options{})
	if got := res.Count(RuleGlowShadow, SevWarn); got != 1 {
		t.Fatalf("want 1 %s WARN, got %d: %+v", RuleGlowShadow, got, res.Findings)
	}
}

func TestPatternGlowShadowAllowsOffsetShadow(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"style.css": ".card { box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08); }\n",
	})
	res := lintDir(t, dir, Options{})
	if got := res.Count(RuleGlowShadow, SevWarn); got != 0 {
		t.Fatalf("want 0 %s findings, got %d", RuleGlowShadow, got)
	}
}

func TestPatternHardShadowWarnsOnZeroBlur(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"style.css": ".tile { box-shadow: 4px 4px 0 #000; }\n.soft { box-shadow: 4px 4px 12px #000; }\n",
	})
	res := lintDir(t, dir, Options{})
	if got := res.Count(RuleHardShadow, SevWarn); got != 1 {
		t.Fatalf("want 1 %s WARN (soft shadow is allowed), got %d: %+v", RuleHardShadow, got, res.Findings)
	}
}

func TestPatternCommentedOutDeclarationIsIgnored(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"style.css": "/* .old { border-left: 6px solid red; } */\n.new { color: #111; }\n",
	})
	res := lintDir(t, dir, Options{})
	if got := res.Count(RuleSideStripe, SevWarn); got != 0 {
		t.Fatalf("want commented rule ignored, got %d", got)
	}
}
