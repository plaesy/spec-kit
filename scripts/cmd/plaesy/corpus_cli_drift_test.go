package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Three prompt files instructed `plaesy generate-template` and
// `plaesy generate-pipeline`. Neither command has ever existed. Both prompts
// asserted the command worked ("cross-platform, single entry point"), so an
// obedient agent either stopped on a raw cobra error or hand-wrote files it
// believed the command had produced. Nothing caught it: every existing drift
// test reads docs/reference.md, and every one of these bugs lived in a file
// none of them open.
//
// This test walks the whole corpus instead of one file.

// corpusRoots are the directories whose Markdown can teach a command.
var corpusRoots = []string{
	"prompts",
	"instructions",
	"agents",
	"checklists",
	"docs",
}

// corpusFiles are the individual Markdown files at the repo root.
var corpusFiles = []string{
	"AGENTS.md",
	"README.md",
}

// cliInvocation matches a `plaesy <words>` mention inside backticks. Requiring
// the backticks is what keeps prose out: a sentence that merely talks about the
// CLI is not an instruction, and flagging those would bury the real hits.
var cliInvocation = regexp.MustCompile("`plaesy ([a-z][a-z0-9-]*(?: [a-z][a-z0-9-]*)*)([^`]*)`")

// knownNonCommands are words that follow `plaesy` in prose but are not
// subcommands. Each needs a reason, because an unexplained entry here is how
// this test quietly stops catching things.
var knownNonCommands = map[string]string{
	"spec-kit":              "the product/repo name, not a subcommand",
	"PLAESY_ROOT":           "environment variable",
	"PLAESY_HOME":           "environment variable",
	"PLAESY_IMAGE_PROVIDER": "environment variable",
}

// renamed are former command names that documentation still shows, always
// inside an explicit rename narrative ("it was `plaesy update-agent-context`").
// They are not defects — the docs are correct about history.
//
// Keys are the written path, and are matched longest-first, so allowlisting
// `config detect` cannot hide a hypothetical future `plaesy detect`. Each entry
// names where it is documented: an unexplained entry here is exactly how this
// test quietly stops catching things, so adding one is a deliberate act.
var renamed = map[string]string{
	"update-agent-context":     "docs/reference.md:157 — now `context update`",
	"task-manage":              "docs/reference.md:503 — now `tasks`",
	"validate-constitution":    "docs/reference.md:631 — Deprecated alias for `validate constitution`",
	"validate-memory":          "docs/reference.md:631 — Deprecated alias for `validate memory`",
	"validate-markdown":        "docs/reference.md:631 — Deprecated alias for `validate markdown`",
	"create-new-feature":       "docs/scripts/features.md:17 — now `features create`",
	"get-feature-paths":        "docs/scripts/features.md:18 — now `features paths`",
	"check-task-prerequisites": "docs/scripts/features.md:18 — now `features validate`",
	"generate-image":           "docs/scripts/images-create.md:14 — now `images create`",
	"config detect":            "docs/reference.md:19,383 — now `platforms detect`",
	"config list":              "docs/reference.md:383 — now `platforms list`",
	"config get-platform":      "docs/reference.md:383 — now `platforms get`",
	"detect-stack":             "docs/reference.md:432 — now `stack detect`",
	"create image":             "docs/scripts/images-create.md:13 — now `images create`",
}

// subtreeHasFlag reports whether cmd or any of its descendants declares the
// named flag.
func (idx *commandIndex) subtreeHasFlag(cmd *cobra.Command, name string) bool {
	found := false
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if found {
			return
		}
		if c.Flags().Lookup(name) != nil {
			found = true
			return
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(cmd)
	return found
}

// isRenamed reports whether the leading words of a written invocation are a
// documented former name. Longest match wins, so a two-word old name is
// preferred over a one-word one.
func isRenamed(words []string) bool {
	for n := len(words); n >= 1; n-- {
		if _, ok := renamed[strings.Join(words[:n], " ")]; ok {
			return true
		}
	}
	return false
}

// commandIndex flattens the live cobra tree into lookup maps.
type commandIndex struct {
	// parents maps a top-level command name to itself, for the "is this a real
	// command" check.
	parents map[string]*cobra.Command
	// byPath maps "stack detect" to that command, for full-path validation.
	byPath map[string]*cobra.Command
	// aliases maps every accepted spelling (name + aliases) to the canonical
	// path, so a doc using a legitimate alias is not reported as a typo.
	aliases map[string]string
}

func newCommandIndex(root *cobra.Command) *commandIndex {
	idx := &commandIndex{
		parents: map[string]*cobra.Command{},
		byPath:  map[string]*cobra.Command{},
		aliases: map[string]string{},
	}
	var walk func(cmd *cobra.Command, prefix string)
	walk = func(cmd *cobra.Command, prefix string) {
		if cmd.Hidden || cmd.Deprecated != "" {
			return
		}
		if prefix == "" {
			idx.parents[cmd.Name()] = cmd
		}
		path := cmd.Name()
		if prefix != "" {
			path = prefix + " " + cmd.Name()
		}
		idx.byPath[path] = cmd
		// The full path is the primary key. Registering only cmd.Name() here
		// would map "create" -> "features create" and leave "features create"
		// itself unresolvable, which reports every real subcommand as a typo.
		idx.aliases[path] = path
		if prefix == "" {
			idx.aliases[cmd.Name()] = path
		} else {
			idx.aliases[prefix+" "+cmd.Name()] = path
		}
		for _, a := range cmd.Aliases {
			if prefix == "" {
				idx.aliases[a] = path
			} else {
				idx.aliases[prefix+" "+a] = path
			}
		}
		for _, sub := range cmd.Commands() {
			walk(sub, path)
		}
	}
	for _, cmd := range root.Commands() {
		walk(cmd, "")
	}
	return idx
}

// resolve takes the words after `plaesy` and returns the deepest command path
// that actually exists, plus the words that were left over as arguments.
//
// Longest-prefix matching is what makes this usable. The corpus legitimately
// writes `plaesy trim run go build` and `plaesy context update claude`, where
// `go build` and `claude` are arguments, not subcommands. Validating the whole
// string as a command name reports all of those as typos, and a check that
// cries wolf 20 times is a check people delete. So: walk down the tree as far
// as it goes, and judge only the first word that does not exist.
func (idx *commandIndex) resolve(words []string) (cmd *cobra.Command, path string, rest []string) {
	if len(words) == 0 {
		return nil, "", nil
	}
	// A bare group (`plaesy config`) is a real command; it just takes no
	// arguments of its own.
	if c, ok := idx.parents[words[0]]; ok {
		return c, words[0], words[1:]
	}
	for i := 1; i <= len(words); i++ {
		candidate := strings.Join(words[:i], " ")
		c, ok := idx.byPath[candidate]
		if !ok {
			break
		}
		cmd, path = c, candidate
	}
	return cmd, path, words[len(strings.Fields(path)):]
}

// corpusMarkdownPaths returns every Markdown file the walk should cover.
func corpusMarkdownPaths(t *testing.T) []string {
	t.Helper()
	root := repoDocs(t)
	if root == "" {
		t.Skip("not a repository checkout")
	}
	repoRoot := filepath.Dir(filepath.Dir(root))

	var paths []string
	for _, dir := range corpusRoots {
		base := filepath.Join(repoRoot, filepath.FromSlash(dir))
		filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				name := d.Name()
				// .plaesy/ and node_modules-style trees under a corpus dir are
				// generated mirrors, not sources; reading them doubles every
				// finding and adds no coverage.
				if name == ".plaesy" || name == "node_modules" || name == ".git" {
					return filepath.SkipDir
				}
				// docs/assessment/ holds dated audit reports. They quote
				// commands that no longer exist precisely because they are
				// records of what was broken, and rewriting history would make
				// the report lie about what it found.
				if dir == "docs" && name == "assessment" {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, ".md") {
				paths = append(paths, path)
			}
			return nil
		})
	}
	for _, f := range corpusFiles {
		if _, err := os.Stat(filepath.Join(repoRoot, f)); err == nil {
			paths = append(paths, filepath.Join(repoRoot, f))
		}
	}
	sort.Strings(paths)
	return paths
}

func relToRepo(t *testing.T, path string) string {
	t.Helper()
	root := repoDocs(t)
	rel, err := filepath.Rel(filepath.Dir(filepath.Dir(root)), path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// TestEveryCommandInTheCorpusExists is the check that would have caught
// generate-template, generate-pipeline, detect-stack, config detect-platform,
// and the three get-mapping-* names.
func TestEveryCommandInTheCorpusExists(t *testing.T) {
	idx := newCommandIndex(rootForTest())

	var findings []string
	for _, path := range corpusMarkdownPaths(t) {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		rel := relToRepo(t, path)

		// Line numbers make the finding actionable without a re-search.
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			for _, m := range cliInvocation.FindAllStringSubmatch(line, -1) {
				written := strings.TrimSpace(m[1])
				if written == "" {
					continue
				}
				words := strings.Fields(written)
				if _, ok := knownNonCommands[words[0]]; ok {
					continue
				}
				if isRenamed(words) {
					continue
				}
				cmd, _, _ := idx.resolve(words)
				if cmd != nil {
					continue
				}
				// Nothing resolved, not even the first word. Report the first
				// word, because that is the part that is wrong; the rest is
				// usually an argument the author appended.
				findings = append(findings,
					fmt.Sprintf("%s:%d: `plaesy %s` — %q is not a command; run `plaesy --help`",
						rel, i+1, written, words[0]))
			}
		}
	}

	for _, f := range findings {
		t.Error(f)
	}
}

// TestEveryFlagInTheCorpusExists is the half of the check that catches missing
// flag documentation. `stack detect --install` shipped and was absent from the
// reference's flag table; the existing docs drift test matches subcommand
// *names* in backtick spans and never looks at flags, so nothing noticed.
//
// A flag is attributed to the deepest command that resolved in the same span,
// because that is the command the author meant. `plaesy stack detect --install`
// attributes --install to `stack detect`, not to `stack`.
func TestEveryFlagInTheCorpusExists(t *testing.T) {
	idx := newCommandIndex(rootForTest())
	flagRef := regexp.MustCompile(`(^|\s)(--[a-z][a-z0-9-]*)`)

	var findings []string
	seen := map[string]bool{} // one finding per (command, flag), not per line

	for _, path := range corpusMarkdownPaths(t) {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		rel := relToRepo(t, path)

		for i, line := range strings.Split(string(data), "\n") {
			for _, m := range cliInvocation.FindAllStringSubmatch(line, -1) {
				written := strings.TrimSpace(m[1])
				if written == "" {
					continue
				}
				words := strings.Fields(written)
				if _, ok := knownNonCommands[words[0]]; ok {
					continue
				}
				if isRenamed(words) {
					continue
				}
				cmd, resolvedPath, _ := idx.resolve(words)
				if cmd == nil {
					// A dead command is TestEveryCommandInTheCorpusExists's
					// finding. Attributing flags to it would double-report.
					continue
				}
				// Only the trailing text of the span can hold flags; the words
				// already consumed as the command path are excluded.
				consumed := len(strings.Fields(resolvedPath))
				tail := strings.Join(words[consumed:], " ") + m[2]

				for _, fm := range flagRef.FindAllStringSubmatch(tail, -1) {
					flag := fm[2]
					// pflag's Lookup takes the bare name. Passing "--ai" misses
					// every flag and makes this test report the whole corpus
					// as broken, which is how a check gets ignored.
					name := strings.TrimLeft(flag, "-")
					// --help is cobra's own, added at execution time and
					// therefore absent from Flags() here.
					if name == "help" {
						continue
					}
					if cmd.Flags().Lookup(name) != nil {
						continue
					}
					// A global flag on the root is legal on any subcommand.
					if root := rootForTest(); root.PersistentFlags().Lookup(name) != nil {
						continue
					}
					// The corpus writes `plaesy features --json` for what is
					// really `plaesy features create --json`: a group followed
					// by one of its children's flags. That is imprecise, not
					// dead, so it only counts as a finding when no descendant
					// of the named command has the flag at all.
					if idx.subtreeHasFlag(cmd, name) {
						continue
					}
					key := resolvedPath + " " + flag
					if seen[key] {
						continue
					}
					seen[key] = true
					findings = append(findings, fmt.Sprintf(
						"%s:%d: `plaesy %s %s` — %s is not a flag of `plaesy %s`",
						rel, i+1, resolvedPath, flag, flag, resolvedPath))
				}
			}
		}
	}

	for _, f := range findings {
		t.Error(f)
	}
}
