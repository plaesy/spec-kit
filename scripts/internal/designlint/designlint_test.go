package designlint

import (
	"math"
	"os"
	"path/filepath"
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
	cases := map[string]rgb{
		"#fff":               {255, 255, 255},
		"#1a1a1a":            {26, 26, 26},
		"rgb(10, 20, 30)":    {10, 20, 30},
		"rgba(10,20,30,0.5)": {10, 20, 30},
		"  #ABCDEF  ":        {0xab, 0xcd, 0xef},
	}
	for in, want := range cases {
		got, ok := parseColor(in)
		if !ok || got != want {
			t.Errorf("parseColor(%q) = %v, %v; want %v", in, got, ok, want)
		}
	}
	for _, in := range []string{"inherit", "transparent", "rgb(300, 0, 0)"} {
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
		"index.html": "<div data-stat=\"99.99%\" style=\"width: 50%\">ok</div>\n",
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

func TestWalkSkipsNodeModulesAndOtherExtensions(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"node_modules/pkg/index.html": "<img src=\"https://picsum.photos/x\">",
		"readme.md":                   "https://picsum.photos/x",
		"src/page.tsx":                "<img src=\"https://placehold.co/600\" />",
	})
	res := lintDir(t, dir, Options{})
	if res.Count(RuleImageHost, SevFail) != 1 {
		t.Fatalf("only src/page.tsx should be linted, got %+v", res.Findings)
	}
	if res.Files != 1 {
		t.Fatalf("Files = %d, want 1", res.Files)
	}
}
