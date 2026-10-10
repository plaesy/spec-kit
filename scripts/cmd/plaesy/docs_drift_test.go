package main

// The constitution's documentation bar says every CLI command appears in
// docs/reference.md. Nothing enforced it, which is how six `validate-*` rows sat
// in that table for a release after the commands were gone, and how a row kept
// pointing at a source file that had been renamed. Documentation drift is
// invisible until someone reads the wrong file, so it gets a test.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// repoDocs locates docs/reference.md by walking up from the package directory.
// Skips when it is absent, so the module still builds and tests outside a
// repository checkout.
func repoDocs(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		candidate := filepath.Join(dir, "docs", "reference.md")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("docs/reference.md not found: not a repository checkout")
	return ""
}

// builtins are cobra's own commands. They are not part of this project's
// surface, so the documentation bar does not apply to them.
func isBuiltin(name string) bool {
	return name == "help" || name == "completion"
}

// documented reports whether a command belongs in the documentation bar, using
// the same predicate cobra uses to decide what `--help` lists: a deprecated
// alias is deliberately absent from the help text, so requiring a row for it
// would document a command the project is asking people to stop using.
func documented(cmd *cobra.Command) bool {
	return !cmd.Hidden && cmd.Deprecated == "" && !isBuiltin(cmd.Name())
}

// indexRow matches one row of the Command Index table:
// | `name` | description | `source` |
var indexRow = regexp.MustCompile("(?m)^\\| `([^`]+)` \\|[^|]*\\|([^|]*)\\|")

// indexRows parses only the Command Index table. indexRow on its own matches
// every three-column table in the document — the flag tables (`--no-graph`,
// `--level`), the `config` subcommand table, the `clean --level` table — so
// using it to decide "is this a stale command row?" reports every flag in the
// reference as a command that no longer exists. The table is delimited by its
// own heading.
func indexRows(t *testing.T, body string) map[string]string {
	t.Helper()
	const heading = "## Command Index"
	start := strings.Index(body, heading)
	if start < 0 {
		t.Fatalf("%s has no %q heading, so the Command Index cannot be checked", heading, heading)
	}
	rest := body[start+len(heading):]
	end := strings.Index(rest, "\n## ")
	if end < 0 {
		end = len(rest)
	}
	rows := map[string]string{}
	for _, m := range indexRow.FindAllStringSubmatch(rest[:end], -1) {
		rows[m[1]] = m[2]
	}
	if len(rows) == 0 {
		t.Fatalf("no rows parsed out of the Command Index table after %q: its format changed, and every check against it would pass on nothing", heading)
	}
	return rows
}

// codeSpan matches an inline code span, which is how every command name in the
// reference is written.
var codeSpan = regexp.MustCompile("`([^`\n]+)`")

func TestEveryCommandIsInTheReferenceIndex(t *testing.T) {
	docs := repoDocs(t)
	body, err := os.ReadFile(docs)
	if err != nil {
		t.Fatal(err)
	}

	rows := indexRows(t, string(body))

	root := rootForTest()
	for _, cmd := range root.Commands() {
		if !documented(cmd) {
			continue
		}
		if _, ok := rows[cmd.Name()]; !ok {
			t.Errorf("command %q is missing a row in the Command Index table of %s", cmd.Name(), docs)
		}
	}
}

func TestEverySubcommandIsMentionedInTheReference(t *testing.T) {
	docs := repoDocs(t)
	body, err := os.ReadFile(docs)
	if err != nil {
		t.Fatal(err)
	}
	// A subcommand counts as documented when some backtick-quoted span in the
	// reference contains its name. Matching on spans rather than on an exact
	// `name` avoids a formatting rule disguised as a documentation rule: the
	// reference writes `### `trim run`` in a heading, `| `validate` |` in a
	// subcommand table, and `| `llm-queue --path <file>` |` in a flag-bearing
	// row, and all three are documented. A command that appears nowhere still
	// fails, which is the bar that matters.
	tokens := map[string]bool{}
	for _, m := range codeSpan.FindAllStringSubmatch(string(body), -1) {
		for _, token := range strings.Fields(m[1]) {
			tokens[token] = true
		}
	}
	root := rootForTest()
	for _, cmd := range root.Commands() {
		if !documented(cmd) {
			continue
		}
		for _, sub := range cmd.Commands() {
			if !documented(sub) {
				continue
			}
			if !tokens[sub.Name()] {
				t.Errorf("subcommand %q of %q is not mentioned in %s", sub.Name(), cmd.Name(), docs)
			}
		}
	}
}

// A row that names a source file which no longer exists sends the reader to a
// dead end and hides the fact that the command moved.
func TestCommandIndexSourceFilesExist(t *testing.T) {
	docs := repoDocs(t)
	root := rootForTest()
	known := map[string]bool{}
	for _, cmd := range root.Commands() {
		if documented(cmd) {
			known[cmd.Name()] = true
		}
	}

	body, err := os.ReadFile(docs)
	if err != nil {
		t.Fatal(err)
	}
	repoRoot := filepath.Dir(filepath.Dir(docs))

	for name, srcs := range indexRows(t, string(body)) {
		if !known[name] {
			continue // a row for a command that no longer exists is covered by TestEveryIndexRowIsARealCommand
		}
		for _, field := range strings.Split(srcs, "+") {
			src := strings.TrimSpace(strings.Trim(strings.TrimSpace(field), "`"))
			if src == "" || !strings.HasSuffix(src, ".go") {
				continue
			}
			if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(src))); err != nil {
				t.Errorf("Command Index row for %q points at %s, which does not exist", name, src)
			}
		}
	}
}

// The other direction. The tests above walk the CLI and ask what the table is
// missing, so a row left behind by a deleted or renamed command was invisible:
// the loop skips unknown names rather than failing on them. The comment there
// said "covered above" — nothing covered it, and the six `validate-*` rows that
// outlived their commands are the case it should have caught.
func TestEveryIndexRowIsARealCommand(t *testing.T) {
	docs := repoDocs(t)
	body, err := os.ReadFile(docs)
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, cmd := range rootForTest().Commands() {
		if documented(cmd) {
			known[cmd.Name()] = true
		}
	}

	for name := range indexRows(t, string(body)) {
		if !known[name] {
			t.Errorf("Command Index row %q has no matching command — the row outlived the command", name)
		}
	}
}

// The constitution makes "the README command list is in sync with
// docs/reference.md" a binding threshold, and nothing checked it. The two lists
// are hand-maintained, duplicated, and — the reason this is a test and not a
// convention — already worded differently: the bash block says
// `plaesy validate [target]` and the PowerShell block says `plaesy validate`.
// That is fine, since the blocks differ in shell syntax by design; what must
// hold is that both name the same set of commands.
func TestReadmeCommandListMatchesTheReferenceIndex(t *testing.T) {
	docs := repoDocs(t)
	readme := filepath.Join(filepath.Dir(filepath.Dir(docs)), "README.md")
	body, err := os.ReadFile(readme)
	if err != nil {
		t.Fatalf("read README: %v", err)
	}

	docBody, err := os.ReadFile(docs)
	if err != nil {
		t.Fatal(err)
	}
	rows := map[string]bool{}
	for name := range indexRows(t, string(docBody)) {
		rows[name] = true
	}

	// `plaesy <name>` at the start of a line inside a fenced block, with the
	// name taken up to the first space so `<directory>` and `[target]` do not
	// become part of it.
	readmeCmd := regexp.MustCompile("(?m)^plaesy ([a-z][a-z-]*)")
	blocks := fencedBlocks(string(body))
	if len(blocks) < 2 {
		t.Fatalf("expected a bash and a PowerShell command list in %s, found %d fenced blocks", readme, len(blocks))
	}
	for i, b := range blocks {
		if !strings.Contains(b, "plaesy ") {
			continue // prose or a non-command block
		}
		names := map[string]bool{}
		for _, m := range readmeCmd.FindAllStringSubmatch(b, -1) {
			names[m[1]] = true
		}
		if len(names) == 0 {
			continue
		}
		for name := range names {
			if !rows[name] {
				t.Errorf("README block %d lists `plaesy %s`, which has no row in the Command Index of %s", i, name, docs)
			}
		}
		for name := range rows {
			if !names[name] {
				t.Errorf("README block %d omits `plaesy %s`, which the Command Index of %s lists", i, name, docs)
			}
		}
	}
}

// fencedBlocks returns the contents of every ``` fenced block, in order.
func fencedBlocks(s string) []string {
	// \r?\n, not \n: on a Windows checkout with core.autocrlf (the GitHub
	// Actions windows-latest default, and this repo has no .gitattributes
	// forcing LF), README.md's line endings are "```bash\r\n...", and a
	// bare \n right after the fence tag never matches a \r\n file — every
	// block in the file would silently come back as zero blocks, not as a
	// parse of the wrong content, which is why this was windows-only.
	var out []string
	re := regexp.MustCompile("(?s)```[a-z]*\r?\n(.*?)```")
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1])
	}
	return out
}
