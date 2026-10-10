package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// security-and-owasp and security-audit each carry half of one taxonomy: the
// first states the coding rule per category, the second the audit criteria.
// For two files to be "one taxonomy" they have to agree on what A04 means.
//
// They did not. security-audit used OWASP Top 10 (2021) verbatim, while
// security-and-owasp merged A01 with A10 and A05 with A06 and had no A04 or
// A09 at all — so A01 meant "access control and SSRF" in one file and "access
// control" in the other, and an audit report citing A01 resolved differently
// depending on which file the reader opened.
//
// The fix was to align the merged file to 2021, because 2021 is the external
// standard with fixed numbering. This test keeps them aligned.

// owaspHeading matches a `### N. Axx: Name` section heading in either file.
var owaspHeading = regexp.MustCompile(`(?mi)^#{2,4}\s*(?:\d+(?:\.\d+)?\.?\s*)?A(\d{2})\s*[:.\-]\s*([^\n]+)$`)

func owaspCategories(t *testing.T, path string) map[string]string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	out := map[string]string{}
	for _, m := range owaspHeading.FindAllStringSubmatch(string(data), -1) {
		code := m[1]
		// The first heading wins. A01 legitimately appears again inside a
		// cross-reference line; the section heading is the definition.
		if _, seen := out[code]; !seen {
			out[code] = strings.TrimSpace(strings.Trim(m[2], "*_` "))
		}
	}
	return out
}

func TestOWASPTaxonomyIsIdenticalInBothSecurityFiles(t *testing.T) {
	docs := repoDocs(t)
	if docs == "" {
		t.Skip("not a repository checkout")
	}
	root := filepath.Dir(filepath.Dir(docs))

	codingPath := filepath.Join(root, "instructions", "security-and-owasp.instructions.md")
	auditPath := filepath.Join(root, "instructions", "security-audit.instructions.md")

	coding := owaspCategories(t, codingPath)
	audit := owaspCategories(t, auditPath)

	if len(audit) == 0 {
		t.Fatalf("no OWASP headings parsed out of %s — the heading pattern changed, and this test is now vacuous", auditPath)
	}

	var missing, extra, mismatched []string

	// Categories security-and-owasp deliberately does not carry a coding rule
	// for. This is a declared omission, not a gap: each entry says where the
	// category is handled instead, so a reader following "A04" from the audit
	// file lands somewhere real. Adding a category to security-audit without
	// either adding it here or declaring it here fails the test.
	declaredOmissions := map[string]string{
		"A04": "Insecure Design — owned by security-design-principles.instructions.md; a design concern with no code-level rule",
		"A09": "Security Logging and Monitoring Failures — audited by security-audit.instructions.md; an operations concern with no code-level rule",
	}

	for code, name := range audit {
		got, ok := coding[code]
		if !ok {
			if reason, declared := declaredOmissions["A"+code]; declared {
				_ = reason
				continue
			}
			missing = append(missing, code+": "+name)
			continue
		}
		if !sameCategory(got, name) {
			mismatched = append(mismatched, code+": audit says "+quote(name)+", coding says "+quote(got))
		}
	}
	for code, name := range coding {
		if _, ok := audit[code]; !ok {
			extra = append(extra, code+": "+name)
		}
	}

	sort.Strings(missing)
	sort.Strings(extra)
	sort.Strings(mismatched)

	for _, m := range missing {
		t.Errorf("security-and-owasp has no section for %s, but security-audit audits it. "+
			"Either add a coding rule here, or declare the omission in declaredOmissions "+
			"with a reason, so a reader following the number lands somewhere real.", m)
	}
	for _, e := range extra {
		t.Errorf("security-and-owasp has a section for %s that security-audit does not audit. "+
			"Add it to the audit file or remove the section.", e)
	}
	for _, m := range mismatched {
		t.Errorf("OWASP category name disagrees between the two files — %s. "+
			"One number, one meaning, or the two files are not one taxonomy.", m)
	}
}

// sameCategory compares two category names after normalising formatting only:
// case, punctuation, and "&" versus "and". Formatting is not a taxonomy
// difference; an extra category name is.
//
// The normaliser deliberately does NOT accept a prefix relationship. Allowing
// "broken access control & server-side request forgery" to match "broken access
// control" is precisely the defect this test exists to catch — and a lenient
// comparison here made the test pass on the exact bug it was written for, which
// is worse than having no test.
func sameCategory(a, b string) bool {
	norm := func(s string) string {
		s = strings.ToLower(s)
		// A parenthetical is a gloss on the name, not part of it:
		// "Injection (SQL, NoSQL, Command)" names the same category as
		// "Injection", and requiring both files to carry the same gloss would
		// make this test a typography check.
		s = parenGloss.ReplaceAllString(s, " ")
		s = strings.NewReplacer("&", " and ", "-", " ", "/", " ").Replace(s)
		return strings.Join(strings.Fields(s), " ")
	}
	return norm(a) == norm(b)
}

// parenGloss strips a trailing or inline parenthetical from a category name.
var parenGloss = regexp.MustCompile(`\s*\([^)]*\)`)

func quote(s string) string { return "\"" + s + "\"" }
