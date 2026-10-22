package designlint

import "testing"

func TestTellEmojiAsIconWarns(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"index.html": "<!doctype html><html><body><h2>Fast</h2><p>✨ Ship faster 🚀</p></body></html>",
	})
	res := lintDir(t, dir, Options{})
	if got := res.Count(RuleEmoji, SevWarn); got != 2 {
		t.Fatalf("want 2 %s WARN, got %d: %+v", RuleEmoji, got, res.Findings)
	}
}

func TestTellIndexNumberingWarnsInVisibleText(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"index.html": "<html><body><span>01 / 03</span><span>(002)</span><p>Step 2</p></body></html>",
	})
	res := lintDir(t, dir, Options{})
	if got := res.Count(RuleIndexNumber, SevWarn); got != 3 {
		t.Fatalf("want 3 %s WARN, got %d: %+v", RuleIndexNumber, got, res.Findings)
	}
}

func TestTellIndexNumberingIgnoresScriptAndDates(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"index.html": "<html><body><p>Printed 2026-10-20</p><script>const label = \"01 / 03\";</script></body></html>",
	})
	res := lintDir(t, dir, Options{})
	if got := res.Count(RuleIndexNumber, SevWarn); got != 0 {
		t.Fatalf("want 0 %s findings (script text and dates are not numbering), got %d", RuleIndexNumber, got)
	}
}

func TestTellCleanPageHasNoTells(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"style.css": "h1 { font-size: 48px; }\n",
		"index.html": "<html><body><h1>Invoices paid on time</h1><p>Write an invoice in under a minute.</p></body></html>",
	})
	res := lintDir(t, dir, Options{})
	for _, rule := range []string{RuleEmoji, RuleIndexNumber} {
		if got := res.Count(rule, SevWarn); got != 0 {
			t.Fatalf("want 0 %s findings on a clean page, got %d", rule, got)
		}
	}
}
