package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Dead relative links are the quietest defect class in this corpus. They render
// as ordinary blue text, they never fail a build, and an agent following one
// concludes the file was deleted rather than that the link is wrong.
//
// A sweep of the corpus found 596 relative links with dead targets across 13
// files, in four shapes:
//
//	docs/instructions/README.md      37 links wrote ./<n>.instructions.md as if
//	                                the files sat beside the doc; they are two
//	                                levels up in instructions/
//	prompts/{doc,fix,implement,optimize}.md
//	                                8 links wrapped the href in backticks, so
//	                                the backticks became part of the URL
//	docs/reference.md:202            pointed at docs/scripts/create.md, which
//	                                does not exist; the /create prompt is the
//	                                only spec of that command
//	README.md, docs/README.md        claimed a top-level testing/ directory
//	                                that has never existed
//
// The exempt targets below are not noise. Each is a link an author is being
// told to write, not one the corpus ships, and each is listed with the file
// that justifies it so the exemption is reviewable rather than a blanket skip.

var mdRelativeLink = regexp.MustCompile(`\[([^\]]*)\]\(([^)\s]+)\)`)

var (
	// Target carries a fill-in, so it resolves when the skeleton is filled
	// and cannot resolve in the repository. That is correct, not broken.
	targetHasFillIn = regexp.MustCompile(`\{\{|\{[a-z][a-z0-9_.-]*\}`)

	mdFenceOpen = regexp.MustCompile("^" + `\s*` + "```")

	// linkDirs are the corpus roots that carry prose and cross-references.
	// .plaesy is excluded because it is generated and mirrors the source, and
	// docs/assessment is excluded because a dated report must keep quoting the
	// broken commands it was written to record.
	linkDirs = []string{"prompts", "instructions", "agents", "checklists", "templates", "docs"}

	linkSkipDir = map[string]bool{".git": true, "node_modules": true, "assessment": true}

	// exampleTargets are placeholder syntax being described in prose. A test
	// that flagged these would be flagging the documentation of the very
	// syntax it is checking.
	exampleTargets = map[string]string{
		"path":                  "instructions/context-engineering.instructions.md:114 — describes a `[text](path)` link form",
		"url":                   "templates/README.md:140 — shows `[X](url)` as the counter-example to migration",
		"file.md":               "instructions/plaesy.instructions.md:121 — specifies the index line format",
		"overview.md":           "instructions/plaesy.instructions.md:122 — worked example, already caveated as valid only once the file exists",
		"memory/overview.md":    "templates/memory.template.md:20-21 — tells the author to create it, then link it",
		"decisions/overview.md": "instructions/plaesy.instructions.md:141 — worked example, already caveated as valid only once the file exists",
		"decisions.md":          "templates/context.template.md:27 — resolves correctly once the template is rendered into .plaesy/, where decisions.md is a sibling",
	}
)

type deadLink struct {
	line         int
	text, target string
}

// findDeadRelativeLinks returns the relative links in body whose target does
// not exist relative to the file's own directory.
func findDeadRelativeLinks(t *testing.T, absPath, root, body string) []deadLink {
	t.Helper()
	dir := filepath.Dir(absPath)
	var dead []deadLink
	inFence := false
	for i, line := range strings.Split(body, "\n") {
		if mdFenceOpen.MatchString(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		for _, m := range mdRelativeLink.FindAllStringSubmatch(line, -1) {
			text, raw := m[1], m[2]
			if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") ||
				strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "mailto:") {
				continue
			}
			target := raw
			if i := strings.IndexByte(target, '#'); i >= 0 {
				target = target[:i]
			}
			if target == "" {
				continue
			}
			if targetHasFillIn.MatchString(raw) {
				continue
			}
			if strings.ContainsAny(raw, "`* ") {
				// A backticked or spaced href is malformed markdown;
				// resolve it anyway so the message names the bug.
				raw = strings.Trim(raw, "`")
				target = strings.TrimSpace(strings.SplitN(target, " ", 2)[0])
			}
			if why, ok := exampleTargets[target]; ok {
				_ = why
				continue
			}
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(target))); err == nil {
				continue
			}
			if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(target))); err == nil {
				continue
			}
			dead = append(dead, deadLink{i + 1, text, raw})
		}
	}
	return dead
}

func TestNoDeadRelativeMarkdownLinks(t *testing.T) {
	docs := repoDocs(t)
	if docs == "" {
		t.Skip("not a repository checkout")
	}
	root := filepath.Dir(filepath.Dir(docs))

	type row struct {
		rel  string
		dead []deadLink
	}
	var all []row
	checked := 0
	for _, d := range linkDirs {
		base := filepath.Join(root, d)
		if _, err := os.Stat(base); err != nil {
			continue
		}
		err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if linkSkipDir[info.Name()] {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".md") {
				return nil
			}
			body, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			checked++
			if dead := findDeadRelativeLinks(t, path, root, string(body)); len(dead) > 0 {
				all = append(all, row{filepath.ToSlash(rel), dead})
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", d, err)
		}
	}
	if checked == 0 {
		t.Fatal("no markdown scanned; the walk is wrong, which would make this test vacuous")
	}

	sort.Slice(all, func(i, j int) bool {
		if len(all[i].dead) != len(all[j].dead) {
			return len(all[i].dead) > len(all[j].dead)
		}
		return all[i].rel < all[j].rel
	})
	for _, r := range all {
		for _, d := range r.dead {
			t.Errorf("%s:%d: link [%s](%s) has no target. A relative link that resolves "+
				"to nothing tells a reader the file was deleted, and tells an agent the "+
				"same thing. Either point it at the real path or drop the link.", r.rel, d.line, d.text, d.target)
		}
	}
	t.Logf("%d markdown file(s) scanned, %d with dead relative links", checked, len(all))
}

// TestNoDanglingMarkdownAnchors is the second half of the cross-reference check.
// A dead target was only half the problem: a link can point at a file that
// exists and still land nowhere, because the `#fragment` inside it matches no
// heading. 17 such links shipped in two files --
//
//	docs/prompts/README.md        8 rows of its Quick Navigation table pointed
//	                              at #creatediagram, #createtasks, #create*
//	                              for /create: sub-command sections the document
//	                              never had. The 10 top-level rows worked, which
//	                              is why the table looked fine.
//	docs/templates/README.md      9 category links pointed at a taxonomy --
//	                              "Project Specification Templates" and so on --
//	                              that exists in no file, not even the registry
//
// The slug function below is the subtle part. GitHub turns *each* whitespace
// character into its own dash, so `#### 3.8.2 Authentication & Authorization`
// slugs to `382-authentication--authorization` with a double dash. Collapsing
// runs, which is the obvious implementation, makes that correct anchor look
// dangling -- which is how this scan first reported a false positive and had to
// be fixed before it could be trusted to report a real one.
//
// Hex colour literals and links inside code spans are excluded: `#121212` in a
// palette table and `[FAQ](#frequently-asked-questions)` in a syntax example are
// not cross-references, and a check that flags them trains people to ignore it.
func TestNoDanglingMarkdownAnchors(t *testing.T) {
	docs := repoDocs(t)
	if docs == "" {
		t.Skip("not a repository checkout")
	}
	root := filepath.Dir(filepath.Dir(docs))

	anchorCache := map[string]map[string]bool{}
	anchorsOf := func(path string) map[string]bool {
		if a, ok := anchorCache[path]; ok {
			return a
		}
		set := map[string]bool{}
		raw, err := os.ReadFile(path)
		if err != nil {
			anchorCache[path] = set
			return set
		}
		inFence := false
		for _, line := range strings.Split(string(raw), "\n") {
			if mdFenceOpen.MatchString(line) {
				inFence = !inFence
				continue
			}
			if inFence {
				continue
			}
			if m := dlHeading.FindStringSubmatch(line); m != nil {
				set[ghSlug(m[1])] = true
			}
		}
		anchorCache[path] = set
		return set
	}

	hexLiteral := regexp.MustCompile(`[0-9a-fA-F]{3,8}\z`)
	checked := 0
	var problems []string
	for _, d := range linkDirs {
		base := filepath.Join(root, d)
		if _, err := os.Stat(base); err != nil {
			continue
		}
		err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if linkSkipDir[info.Name()] {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".md") {
				return nil
			}
			raw, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			rel = filepath.ToSlash(rel)
			inFence := false
			for i, line := range strings.Split(string(raw), "\n") {
				if mdFenceOpen.MatchString(line) {
					inFence = !inFence
					continue
				}
				if inFence {
					continue
				}
				prose := codeSpan.ReplaceAllStringFunc(line, func(s string) string {
					return strings.Repeat(" ", len(s))
				})
				for _, m := range mdRelativeLink.FindAllStringSubmatch(prose, -1) {
					raw2 := m[2]
					if strings.HasPrefix(raw2, "http://") || strings.HasPrefix(raw2, "https://") ||
						strings.HasPrefix(raw2, "mailto:") {
						continue
					}
					pathPart, frag, found := strings.Cut(raw2, "#")
					if !found || frag == "" || hexLiteral.MatchString(frag) {
						continue
					}
					if !anchorFragment.MatchString(frag) {
						continue
					}
					target := anchorsOf(path)
					if pathPart != "" {
						tp := osPathJoin(filepath.Dir(path), pathPart)
						if _, err := os.Stat(tp); err != nil {
							continue // dead target; TestNoDeadRelativeMarkdownLinks owns that
						}
						target = anchorsOf(tp)
					}
					checked++
					if !target[frag] {
						problems = append(problems, fmt.Sprintf(
							"%s:%d: [%s](%s) resolves to a file that exists, but no heading in it "+
								"slugs to #%s", rel, i+1, m[1], raw2, frag))
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", d, err)
		}
	}
	if checked == 0 {
		t.Fatal("no anchors checked; the walk is wrong, which would make this test vacuous")
	}
	sort.Strings(problems)
	for _, p := range problems {
		t.Error(p)
	}
	t.Logf("%d anchor fragment(s) checked, %d dangling", checked, len(problems))
}

var (
	dlHeading        = regexp.MustCompile(`^#{1,6}\s+(.*)$`)
	anchorFragment   = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	mdNonWordOrSpace = regexp.MustCompile(`[^\w\s-]`)
	mdWhitespace     = regexp.MustCompile(`[\s_]`)
	mdLeadDash       = regexp.MustCompile(`^-+|-+$`)
)

// ghSlug reproduces GitHub's heading slug. Each whitespace character becomes its
// own dash; runs are not collapsed. That distinction is the whole reason this
// function exists in this shape.
func ghSlug(heading string) string {
	s := strings.ToLower(strings.TrimSpace(heading))
	s = mdNonWordOrSpace.ReplaceAllString(s, "")
	s = mdWhitespace.ReplaceAllString(s, "-")
	return mdLeadDash.ReplaceAllString(s, "")
}

func osPathJoin(dir, rel string) string {
	if strings.HasPrefix(rel, "/") {
		return filepath.FromSlash(rel)
	}
	return filepath.Join(dir, filepath.FromSlash(rel))
}
