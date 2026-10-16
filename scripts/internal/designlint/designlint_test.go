package designlint

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// writeFiles creates name → content under a fresh temp dir and returns its path.
func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func lintDir(t *testing.T, dir string, opts Options) Result {
	t.Helper()
	res, err := Lint([]string{dir}, opts)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	return res
}

func TestContrastRatioMatchesWCAGReference(t *testing.T) {
	// Black on white is the maximum, 21:1. #777 on white is the near miss that
	// the AA gate must reject (4.48:1).
	if got := contrastRatio(rgb{0, 0, 0}, rgb{255, 255, 255}); math.Abs(got-21) > 0.01 {
		t.Fatalf("black on white = %.3f, want 21", got)
	}
	got := contrastRatio(rgb{0x77, 0x77, 0x77}, rgb{255, 255, 255})
	if got >= 4.5 || got < 4.4 {
		t.Fatalf("#777 on white = %.3f, want just under 4.5", got)
	}
}

func TestParseColorForms(t *testing.T) {
	cases := map[string]colour{
		"#fff":               {rgb{255, 255, 255}, 1},
		"#1a1a1a":            {rgb{26, 26, 26}, 1},
		"rgb(10, 20, 30)":    {rgb{10, 20, 30}, 1},
		"rgba(10,20,30,0.5)": {rgb{10, 20, 30}, 0.5},
		"  #ABCDEF  ":        {rgb{0xab, 0xcd, 0xef}, 1},
		"#0008":              {rgb{0, 0, 0}, 0x88 / 255.0},
		"white":              {rgb{255, 255, 255}, 1},
	}
	for in, want := range cases {
		got, ok := parseColor(in)
		if !ok || got.c != want.c || math.Abs(got.a-want.a) > 1e-9 {
			t.Errorf("parseColor(%q) = %+v, %v; want %+v", in, got, ok, want)
		}
	}
	for _, in := range []string{"inherit", "currentColor", "rgb(300, 0, 0)", "hsl(0 0% 50%)"} {
		if _, ok := parseColor(in); ok {
			t.Errorf("parseColor(%q) should not resolve", in)
		}
	}
}

func TestImageHostFailsOnPlaceholderHost(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"index.html": "<html><body><img src=\"https://picsum.photos/seed/x/800/600\"></body></html>\n",
	})
	res := lintDir(t, dir, Options{})
	if res.Count(RuleImageHost, SevFail) != 1 {
		t.Fatalf("want one %s FAIL, got %+v", RuleImageHost, res.Findings)
	}
	if !res.Failed() {
		t.Fatal("placeholder host must fail the run")
	}
}

func TestImageHostPassesOnOwnAssets(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"index.html": "<img src=\"assets/hero.jpg\" alt=\"Interior of the bakery\">\n",
	})
	if res := lintDir(t, dir, Options{}); res.Failed() {
		t.Fatalf("own asset must pass, got %+v", res.Findings)
	}
}

func TestFigureWarnsOnInventedStatInVisibleText(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"index.html": "<p>We are 99.99% uptime and cost $100.00.</p>\n<script>var x = \"50%\";</script>\n",
	})
	res := lintDir(t, dir, Options{})
	if got := res.Count(RuleFigure, SevWarn); got != 2 {
		t.Fatalf("want 2 %s WARN (script text must be ignored), got %d: %+v", RuleFigure, got, res.Findings)
	}
	if res.Failed() {
		t.Fatal("invented-figure is a WARN, not a FAIL")
	}
}

func TestFigureIgnoresAttributesAndCSS(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"index.html": "<div data-stat=\"99.99%\" class=\"x\">ok</div>\n",
		"style.css":  "p { width: 50%; }\n",
	})
	if res := lintDir(t, dir, Options{}); res.Count(RuleFigure, SevWarn) != 0 {
		t.Fatalf("attributes and CSS must not trigger, got %+v", res.Findings)
	}
}

func TestDefaultFontWarnsOnInterFirst(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"style.css": "body { font-family: Inter, sans-serif; }\nh1 { font-family: 'Geist', sans-serif; }\n",
	})
	res := lintDir(t, dir, Options{})
	if got := res.Count(RuleFont, SevWarn); got != 1 {
		t.Fatalf("want 1 %s WARN (Geist is allowed), got %d: %+v", RuleFont, got, res.Findings)
	}
}

func TestContrastFailsNearMiss(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"style.css": ".note { color: #777; background: #fff; }\n",
	})
	res := lintDir(t, dir, Options{})
	if res.Count(RuleContrast, SevFail) != 1 {
		t.Fatalf("#777 on #fff must fail AA normal text, got %+v", res.Findings)
	}
	if res.ContrastChecked != 1 {
		t.Fatalf("ContrastChecked = %d, want 1", res.ContrastChecked)
	}
}

func TestContrastPassesAndUsesTokens(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"style.css": ":root { --ink: #1a1a1a; --paper: #f5f5f0; }\n.body { color: var(--ink); background-color: var(--paper); }\n",
	})
	res := lintDir(t, dir, Options{})
	if res.Failed() {
		t.Fatalf("token pair must pass, got %+v", res.Findings)
	}
	if res.ContrastChecked != 1 {
		t.Fatalf("token pair should be checked, ContrastChecked = %d", res.ContrastChecked)
	}
}

func TestContrastLargeTextUsesThreeToOne(t *testing.T) {
	// #808080 on white is 3.95:1: below 4.5 for body text, above 3 for large text.
	body := writeFiles(t, map[string]string{"a.css": "p { color: #808080; background: #fff; font-size: 16px; }\n"})
	if !lintDir(t, body, Options{}).Failed() {
		t.Fatal("3.95:1 body text must fail")
	}
	large := writeFiles(t, map[string]string{"a.css": "h1 { color: #808080; background: #fff; font-size: 32px; }\n"})
	if lintDir(t, large, Options{}).Failed() {
		t.Fatal("3.95:1 at 32px must pass the 3:1 large-text rule")
	}
	bold := writeFiles(t, map[string]string{"a.css": "h2 { color: #808080; background: #fff; font-size: 19px; font-weight: 700; }\n"})
	if lintDir(t, bold, Options{}).Failed() {
		t.Fatal("3.95:1 at 19px bold must pass the 3:1 large-text rule")
	}
}

func TestContrastSkipsUnresolvableRules(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.css": ".hero { color: #fff; background: linear-gradient(#000, #111); }\n.x { color: inherit; background: #000; }\n",
	})
	res := lintDir(t, dir, Options{})
	if res.Failed() {
		t.Fatalf("unresolvable rules must not fail, got %+v", res.Findings)
	}
	if res.ContrastSkipped != 2 || res.ContrastChecked != 0 {
		t.Fatalf("skipped=%d checked=%d, want 2 and 0", res.ContrastSkipped, res.ContrastChecked)
	}
}

// M-1: the verdict must not depend on map iteration. Two files define --ink
// differently (light and dark theme). The conflict is reported, and the contrast
// that uses --ink is skipped, so the verdict is the same on every run.
func TestTokenConflictIsDeterministicAndSkipped(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.css": ":root{--ink:#111}\n.p{color:var(--ink);background:#fff}\n",
		"b.css": ":root[data-theme=\"dark\"]{--ink:#fff}\n",
	})
	first := lintDir(t, dir, Options{})
	if first.Failed() {
		t.Fatalf("conflicting token must be skipped, not failed: %+v", first.Findings)
	}
	if first.Count(RuleToken, SevWarn) != 1 || first.ContrastSkipped != 1 || first.ContrastChecked != 0 {
		t.Fatalf("want one token WARN and one skipped check, got %+v (skipped %d, checked %d)",
			first.Findings, first.ContrastSkipped, first.ContrastChecked)
	}
}

// M-1 repeat-run: identical input must give identical output, every time, and
// the order of the path arguments must not matter.
func TestLintIsStableAcrossRunsAndPathOrder(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.css":       ":root{--ink:#111;--paper:#fff}\n.p{color:var(--ink);background:#fff}\n.q{color:#777;background:var(--paper)}\n",
		"b.css":       ":root[data-theme=\"dark\"]{--ink:#fff}\n.r{color:#808080;background:#fff;font-size:1.1em}\n",
		"c/page.html": "<p style=\"color:#777777; background:#ffffff\">x</p>\n<p>99.99%</p>\n",
		"d/app.jsx":   "export const A = () => <p style={{ color: '#777', backgroundColor: '#fff' }}>x</p>;\n",
	})
	want := lintDir(t, dir, Options{})
	for i := 0; i < 30; i++ {
		got := lintDir(t, dir, Options{})
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("run %d differs:\nwant %v\n got %v", i, fmt.Sprint(want), fmt.Sprint(got))
		}
	}
	files, _ := collectFiles([]string{filepath.Join(dir, "a.css"), filepath.Join(dir, "b.css")})
	swapped, _ := collectFiles([]string{filepath.Join(dir, "b.css"), filepath.Join(dir, "a.css")})
	if !reflect.DeepEqual(files, swapped) {
		t.Fatalf("path order changed the file list: %v vs %v", files, swapped)
	}
}

// M-2: inline style attributes and JSX style objects are checked.
func TestInlineStyleAttributeIsChecked(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"index.html": "<p style=\"color:#777777; background:#ffffff\">Payment is due.</p>\n",
	})
	res := lintDir(t, dir, Options{})
	if res.Count(RuleContrast, SevFail) != 1 {
		t.Fatalf("inline #777 on #fff must FAIL, got %+v (checked %d)", res.Findings, res.ContrastChecked)
	}
}

func TestJSXStyleObjectIsChecked(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"page.tsx": "export const P = () => <p style={{ color: '#777777', backgroundColor: '#ffffff', fontSize: 16 }}>x</p>;\n",
	})
	res := lintDir(t, dir, Options{})
	if res.Count(RuleContrast, SevFail) != 1 {
		t.Fatalf("JSX #777 on #fff must FAIL, got %+v", res.Findings)
	}
}

func TestNonLiteralStyleIsCountedAsSkipped(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"page.jsx": "export const P = () => <p style={styles.body}>x</p>;\n<div :style=\"x\"></div>\n",
	})
	res := lintDir(t, dir, Options{})
	if res.Failed() || res.ContrastSkipped != 2 {
		t.Fatalf("non-literal and Vue bindings must be skipped and counted, got skipped=%d findings=%+v", res.ContrastSkipped, res.Findings)
	}
}

// M-3: translucent text is composited over its background before the check.
// firstRule returns the message of the first finding with the given rule, or "".
func firstRule(res Result, rule string) string {
	for _, f := range res.Findings {
		if f.Rule == rule {
			return f.Message
		}
	}
	return ""
}

func TestTranslucentTextIsCompositedAndFails(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.css": ".n{color:rgba(0,0,0,0.4); background-color:#fff}\n",
	})
	res := lintDir(t, dir, Options{})
	if res.Count(RuleContrast, SevFail) != 1 {
		t.Fatalf("rgba(0,0,0,.4) on white composites to about #999 (2.85:1) and must FAIL, got %+v", res.Findings)
	}
	if !strings.Contains(firstRule(res, RuleContrast), "opaque") {
		t.Fatalf("message should say the colour is translucent: %q", firstRule(res, RuleContrast))
	}
}

func TestTranslucentTextThatPassesIsNotFlagged(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.css": ".n{color:rgba(0,0,0,0.9); background-color:#fff}\n",
	})
	if res := lintDir(t, dir, Options{}); res.Failed() {
		t.Fatalf("rgba(0,0,0,.9) on white is about 16:1 and must pass, got %+v", res.Findings)
	}
}

func TestTranslucentBackgroundIsSkipped(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.css": ".n{color:#777; background-color:rgba(255,255,255,0.5)}\n",
	})
	res := lintDir(t, dir, Options{})
	if res.Failed() || res.ContrastChecked != 0 || res.ContrastSkipped != 1 {
		t.Fatalf("a translucent background cannot be measured here: want skipped, got %+v", res)
	}
}

// M-4: em and rem resolve to px; a 1.5em heading is large text.
func TestEmFontSizeResolvesToLargeText(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.css": "h2{font-size:1.5em; color:#808080; background-color:#fff}\n",
	})
	res := lintDir(t, dir, Options{})
	if res.Failed() {
		t.Fatalf("1.5em = 24px is large text; 3.95:1 must pass the 3:1 rule, got %+v", res.Findings)
	}
}

func TestEmFontSizeBelowLargeIsNormalAndStates(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.css": "p{font-size:1.1em; color:#808080; background-color:#fff}\n",
	})
	res := lintDir(t, dir, Options{})
	if res.Count(RuleContrast, SevFail) != 1 {
		t.Fatalf("1.1em = 17.6px is normal text: 3.95:1 must FAIL, got %+v", res.Findings)
	}
	if !strings.Contains(firstRule(res, RuleContrast), "assuming a 16px root") {
		t.Fatalf("the em assumption must be stated: %q", res.Findings[0].Message)
	}
}

func TestClampUsesTheSmallestSize(t *testing.T) {
	// clamp(1rem, 2vw, 1.5rem): the smallest argument is 16px, so normal text applies.
	dir := writeFiles(t, map[string]string{
		"a.css": "p{font-size:clamp(1rem, 2vw, 1.5rem); color:#808080; background-color:#fff}\n",
	})
	if res := lintDir(t, dir, Options{}); res.Count(RuleContrast, SevFail) != 1 {
		t.Fatalf("clamp with a 16px minimum is normal text and must FAIL at 3.95:1, got %+v", res.Findings)
	}
}

func TestUnknownFontSizeIsSkippedNotGuessed(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.css": "p{font-size:var(--fs); color:#808080; background-color:#fff}\n",
	})
	res := lintDir(t, dir, Options{})
	if res.Failed() || res.ContrastSkipped != 1 {
		t.Fatalf("an unreadable font-size must be skipped, got %+v (skipped %d)", res.Findings, res.ContrastSkipped)
	}
}

func TestFontShorthandSizeAndWeight(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.css": "h2{font: 600 24px/1.2 Geist, sans-serif; color:#808080; background-color:#fff}\n",
	})
	if res := lintDir(t, dir, Options{}); res.Failed() {
		t.Fatalf("24px in the shorthand is large text and must pass 3.95:1, got %+v", res.Findings)
	}
}

func TestCommentedOutRuleIsIgnored(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.css": "/* p{color:#777;background:#fff} */\n",
	})
	if res := lintDir(t, dir, Options{}); res.ContrastChecked != 0 || res.Failed() {
		t.Fatalf("a commented-out rule must not be checked: %+v", res)
	}
}

func TestViewportsInfoWithoutEvidence(t *testing.T) {
	dir := writeFiles(t, map[string]string{"a.html": "<p>x</p>\n"})
	res := lintDir(t, dir, Options{})
	if res.Count(RuleViewports, SevInfo) != 1 || res.Failed() {
		t.Fatalf("no evidence must be INFO, not a failure: %+v", res.Findings)
	}
}

func TestViewportsFailWhenAWidthIsMissing(t *testing.T) {
	src := writeFiles(t, map[string]string{"a.html": "<p>x</p>\n"})
	evidence := writeFiles(t, map[string]string{
		"home-1440.png": "",
		"home-390.png":  "",
		"notes.txt":     "768",
	})
	res := lintDir(t, src, Options{EvidenceDir: evidence})
	if got := res.Count(RuleViewports, SevFail); got != 1 {
		t.Fatalf("want exactly one missing width (768; .txt does not count), got %d: %+v", got, res.Findings)
	}
	if !strings.Contains(res.Findings[len(res.Findings)-1].Message, "768") {
		t.Fatalf("finding should name 768 px: %+v", res.Findings)
	}
}

// m-14: a width must be its own number. 1768 and 3900 are not 768 and 390.
func TestViewportMatchesWholeNumberOnly(t *testing.T) {
	src := writeFiles(t, map[string]string{"a.html": "<p>x</p>\n"})
	evidence := writeFiles(t, map[string]string{
		"1440.png": "", "1768.png": "", "3900.png": "",
	})
	res := lintDir(t, src, Options{EvidenceDir: evidence})
	if got := res.Count(RuleViewports, SevFail); got != 2 {
		t.Fatalf("1768 and 3900 must not satisfy 768 and 390: want 2 FAIL, got %d: %+v", got, res.Findings)
	}
}

func TestViewportsPassWhenAllWidthsPresent(t *testing.T) {
	src := writeFiles(t, map[string]string{"a.html": "<p>x</p>\n"})
	evidence := writeFiles(t, map[string]string{
		"1440.png": "", "768.webp": "", "390.jpg": "",
	})
	if res := lintDir(t, src, Options{EvidenceDir: evidence}); res.Failed() {
		t.Fatalf("all widths present must pass, got %+v", res.Findings)
	}
}

func TestViewportsFailOnMissingDirectory(t *testing.T) {
	src := writeFiles(t, map[string]string{"a.html": "<p>x</p>\n"})
	res := lintDir(t, src, Options{EvidenceDir: filepath.Join(src, "nope")})
	if !res.Failed() {
		t.Fatal("a missing evidence directory must fail")
	}
}

// m-15: an unreadable evidence directory is reported as such, not as missing widths.
func TestUnreadableEvidenceIsReportedAsItself(t *testing.T) {
	src := writeFiles(t, map[string]string{"a.html": "<p>x</p>\n"})
	res := lintDir(t, src, Options{EvidenceDir: filepath.Join(src, "missing")})
	for _, f := range res.Findings {
		if strings.Contains(f.Message, "does not exist") {
			return
		}
	}
	t.Fatalf("missing directory must say so: %+v", res.Findings)
}

// m-18: build output is skipped only at the top of the target; source folders are linted.
func TestWalkSkipsBuildOutputAtTopLevelOnly(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"node_modules/pkg/index.html": "<img src=\"https://picsum.photos/x\">",
		"readme.md":                   "https://picsum.photos/x",
		"src/page.tsx":                "<img src=\"https://placehold.co/600\" />",
		"src/build/tool.html":         "<img src=\"https://placehold.co/1\" />",
		"dist/bundle.html":            "<img src=\"https://placehold.co/2\" />",
	})
	res := lintDir(t, dir, Options{})
	if res.Count(RuleImageHost, SevFail) != 2 {
		t.Fatalf("src/page.tsx and src/build/tool.html must be linted; dist/ and node_modules/ skipped: %+v", res.Findings)
	}
	if res.Files != 2 {
		t.Fatalf("Files = %d, want 2", res.Files)
	}
}

func TestTargetItselfNamedBuildIsLinted(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"build/index.html": "<img src=\"https://placehold.co/1\" />",
	})
	res := lintDir(t, filepath.Join(dir, "build"), Options{})
	if res.Count(RuleImageHost, SevFail) != 1 {
		t.Fatalf("an explicit target named build/ must be linted: %+v", res.Findings)
	}
}

// m-16: a colour the lint cannot read is counted as skipped, never silently dropped.
func TestUnreadableColourIsCountedAsSkipped(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.css": ".x{color:hsl(0 0% 50%); background:#fff}\n",
	})
	res := lintDir(t, dir, Options{})
	if res.Failed() || res.ContrastSkipped != 1 || res.ContrastChecked != 0 {
		t.Fatalf("hsl() text must be skipped and counted: %+v (skipped %d)", res.Findings, res.ContrastSkipped)
	}
}
