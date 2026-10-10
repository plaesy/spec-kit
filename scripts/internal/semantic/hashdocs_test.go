package semantic

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/graph"
)

// Ruby drift fixture: a documented class, a documented def, a def documented
// with a "=begin"/"=end" block, and two symbols that must stay undocumented —
// one with nothing above it, one with only a blank line above it.
const rubyDriftFixture = `#!/usr/bin/env ruby
# frozen_string_literal: true

# Account is a single user's stored balance.
class Account
  # Loads the stored balance for the given id.
  def balance(id)
    0
  end

  def undocumented_helper
    1
  end
end

=begin
Persists an account to durable storage.
Writes atomically so a crash cannot truncate the ledger.
=end
def persist(account)
  account
end

module Ledger
end
`

// Shell drift fixture: both declaration forms symbols.go recognises — the
// "function name" form and the "name() {" brace form — plus an undocumented
// brace-form function. The shebang and the shellcheck directive sit above a
// blank line so they belong to nothing.
const shDriftFixture = `#!/usr/bin/env bash
# shellcheck shell=bash
set -euo pipefail

# build compiles the project into dist/.
function build {
  echo build
}

# run_tests executes the unit suite.
run_tests() {
  echo test
}

helper() {
  echo helper
}
`

// PowerShell drift fixture: a "<# ... #>" block comment, a "#" comment line,
// and an undocumented function.
const psDriftFixture = `<#
Publishes the built artifacts to the target environment.
Assumes the caller already authenticated.
#>
function Publish-Artifacts {
  Write-Host "publish"
}

# Reads the deployment target from the environment.
function Get-DeployTarget {
  return $env:DEPLOY_TARGET
}

function Remove-StaleCaches {
  Write-Host "stale"
}
`

// TestRubyDocComments_Drift is hashdocs.go's Ruby half of the guard
// TestDocExtractorNamesMatchGraphSymbols describes for JS/TS and Python, and
// needs to exist per language: the extractor carries a private copy of
// symbols.go's rubySymbols regexes, so an edit there silently kills Ruby doc
// enrichment while every other test still passes.
//
// Both directions are asserted, and the fixture pins graph.Node.Symbols
// exactly. A one-way check is not enough — breaking a regex makes the extractor
// produce *fewer* docs, and the few that remain are still legitimately keyed, so
// a subset-only check passes while enrichment has quietly rotted.
func TestRubyDocComments_Drift(t *testing.T) {
	hashDriftCheck(t, "svc/accounts.rb", rubyDriftFixture, rubyDocComments, hashDriftWant{
		symbols: []string{"Account", "balance", "undocumented_helper", "persist", "Ledger"},
		docs:    []string{"Account", "balance", "persist"},
	})
}

// TestShDocComments_Drift is the Shell half of the same guard (shSymbols, the
// ".sh" case of symbolsOf). See TestRubyDocComments_Drift for why it is
// per-language and two-directional.
func TestShDocComments_Drift(t *testing.T) {
	hashDriftCheck(t, "scripts/build.sh", shDriftFixture, shDocComments, hashDriftWant{
		symbols: []string{"build", "run_tests", "helper"},
		docs:    []string{"build", "run_tests"},
	})
}

// TestPsDocComments_Drift is the PowerShell half of the same guard (the ".ps1"
// case of symbolsOf). See TestRubyDocComments_Drift for why it is per-language
// and two-directional.
func TestPsDocComments_Drift(t *testing.T) {
	hashDriftCheck(t, "scripts/deploy.ps1", psDriftFixture, psDocComments, hashDriftWant{
		symbols: []string{"Publish-Artifacts", "Get-DeployTarget", "Remove-StaleCaches"},
		docs:    []string{"Publish-Artifacts", "Get-DeployTarget"},
	})
}

// hashDriftWant is the fixture's own account of what graph and the extractor
// must agree on, so a drift error can say which side moved.
type hashDriftWant struct {
	symbols []string
	docs    []string
}

// hashDriftCheck builds a real graph over a one-file fixture with graph.Build
// and asserts the three properties the extractor has to hold for nodeText's
// lookup to work:
//
//	(1) graph.Node.Symbols is exactly the fixture's declared symbols,
//	(2) every symbol the fixture documents gets a doc comment,
//	(3) every keyed doc name is a real graph symbol, and no undocumented symbol
//	    is keyed at all.
func hashDriftCheck(t *testing.T, rel, content string, extract func(string) map[string]string, want hashDriftWant) {
	t.Helper()

	repoRoot := t.TempDir()
	outDir := t.TempDir()
	path := filepath.Join(repoRoot, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := graph.Options{RepoPath: repoRoot, OutDir: outDir}
	paths, err := graph.ResolvePaths(opts)
	if err != nil {
		t.Fatalf("ResolvePaths: %v", err)
	}
	paths.OutFull = outDir

	g, err := graph.Build(opts, paths)
	if err != nil {
		t.Fatalf("graph.Build: %v", err)
	}

	var symbols []string
	for _, n := range g.Nodes {
		if filepath.ToSlash(n.ID) == rel {
			symbols = n.Symbols
			break
		}
	}
	if symbols == nil {
		t.Fatalf("graph.Build produced no node for %s (ids: %v)", rel, nodeIDs(g))
	}
	if !reflect.DeepEqual(symbols, want.symbols) {
		t.Fatalf("graph.Node.Symbols for %s = %v, want %v — the fixture and "+
			"internal/graph/symbols.go disagree, so this test's own premise is wrong", rel, symbols, want.symbols)
	}
	symbolSet := make(map[string]bool, len(symbols))
	for _, s := range symbols {
		symbolSet[s] = true
	}

	docs := extract(content)

	// (2) every documented symbol must have been recognised.
	for _, name := range want.docs {
		if docs[name] == "" {
			t.Errorf("extractor found no doc comment for documented symbol %q; got %v. "+
				"If graph.Build still extracts it (%v) but this package does not, the copied "+
				"regex has drifted from internal/graph/symbols.go and enrichment is silently dead",
				name, sortedKeys(docs), symbols)
		}
	}

	// (3) nothing keyed under a name graph does not extract, and no doc for a
	// symbol the fixture leaves undocumented.
	for name := range docs {
		if !symbolSet[name] {
			t.Errorf("extractor keyed a doc comment under %q, but graph.Build extracted %v — "+
				"nodeText would never look this comment up", name, symbols)
		}
		if !containsString(want.docs, name) {
			t.Errorf("extractor produced a doc for undocumented symbol %q; the fixture documents only %v",
				name, want.docs)
		}
	}
}

// The keys must be exactly the names internal/graph's rubySymbols puts into
// graph.Node.Symbols, or nodeText's lookup would never match. The expected keys
// below are the ones asserted by graph's own TestSymbolsOfRuby (def, class,
// "self." method, predicate name) plus the "::" namespace form its regex allows.
func TestRubyDocComments_KeysMatchGraphSymbols(t *testing.T) {
	src := `# loads a user
def load_user(id)
end

# finds a user
def self.find(id)
end

# is the account usable
def valid?
end

# a repository
class UserRepo
end

# namespaced
class Api::V1::Account
end
`
	docs := rubyDocComments(src)
	for _, name := range []string{"load_user", "self.find", "valid?", "UserRepo", "Api::V1::Account"} {
		if docs[name] == "" {
			t.Errorf("docs = %v, want an entry for %q (rubySymbols yields that name)", keysOf(docs), name)
		}
	}
	if len(docs) != 5 {
		t.Errorf("docs = %v, want exactly the 5 declared symbols", keysOf(docs))
	}
}

// "=begin"/"=end" is Ruby's other comment form: bare prose with no "#" on any
// line, which is why it needs its own scan. Blank lines inside the block are
// kept as paragraph breaks and each line is joined with a newline, so nodeText's
// firstLine() takes the summary line.
func TestRubyDocComments_BeginEndBlock(t *testing.T) {
	src := `=begin
Persists an account to durable storage.

Writes atomically so a crash cannot truncate the ledger.
=end
def persist(account)
  account
end

# a documented class
class Account
end
`
	docs := rubyDocComments(src)
	want := "Persists an account to durable storage.\nWrites atomically so a crash cannot truncate the ledger."
	if docs["persist"] != want {
		t.Errorf("docs[persist] = %q, want %q", docs["persist"], want)
	}
	if docs["Account"] != "a documented class" {
		t.Errorf("docs[Account] = %q, want %q", docs["Account"], "a documented class")
	}
}

// An "=end" with no "=begin" above it must contribute nothing rather than pair
// with a distant opener, the same way pyDocstring drops an unterminated
// docstring. An empty block likewise contributes nothing.
func TestRubyDocComments_BeginEndWithoutBegin_IsNotADocComment(t *testing.T) {
	unterminated := "=end\ndef persist(account)\n  account\nend\n"
	if docs := rubyDocComments(unterminated); docs["persist"] != "" {
		t.Errorf("docs[persist] = %q, want no doc from an =end with no =begin above it", docs["persist"])
	}

	empty := "=begin\n=end\ndef persist(account)\n  account\nend\n"
	if docs := rubyDocComments(empty); docs["persist"] != "" {
		t.Errorf("docs[persist] = %q, want no doc from an empty =begin/=end block", docs["persist"])
	}
}

// A blank line detaches a comment from the declaration below it — the same rule
// go/parser's doc-comment grammar applies — otherwise a comment belonging to
// some earlier statement would be merged into the next declaration's docs.
func TestRubyDocComments_BlankLineDetachesComment(t *testing.T) {
	src := `# hash_password hashes the stored password.

# detachedComment is not a doc comment for the method below it.
def detached
end
`
	docs := rubyDocComments(src)
	want := "detachedComment is not a doc comment for the method below it."
	if docs["detached"] != want {
		t.Errorf("docs[detached] = %q, want only the contiguous comment %q", docs["detached"], want)
	}
	if detached := rubyDocComments("# orphan comment\n\ndef orphan\nend\n"); detached["orphan"] != "" {
		t.Errorf("a comment separated by a blank line must not be attributed, got %q", detached["orphan"])
	}
}

func TestRubyDocComments_NoComments_ReturnsNil(t *testing.T) {
	if docs := rubyDocComments("x = 1\n\ndef f(a)\n  a\nend\n"); docs != nil {
		t.Errorf("expected nil docs for a file with no doc comments, got %v", docs)
	}
}

// Both declaration forms symbols.go recognises for Shell: the "function name"
// form and the "name() {" brace form.
func TestShDocComments_KeysMatchGraphSymbols(t *testing.T) {
	src := `# compiles the project
function build {
  :
}

# runs the suite
run_tests() {
  :
}

# a bracketed name is still extracted by shSymbols
function noop {
  :
}
`
	docs := shDocComments(src)
	for _, name := range []string{"build", "run_tests", "noop"} {
		if docs[name] == "" {
			t.Errorf("docs = %v, want an entry for %q (shSymbols yields that name)", keysOf(docs), name)
		}
	}
	if len(docs) != 3 {
		t.Errorf("docs = %v, want exactly the 3 declared symbols", keysOf(docs))
	}
}

// A comment above a call site must not be inherited by the next declaration,
// and a blank line detaches a comment as it does in every other extractor.
func TestShDocComments_BlankLineDetachesComment(t *testing.T) {
	src := `# hashPassword hashes the stored password.

# detached is not a doc comment for the function below it.
run_tests() {
  :
}
`
	docs := shDocComments(src)
	want := "detached is not a doc comment for the function below it."
	if docs["run_tests"] != want {
		t.Errorf("docs[run_tests] = %q, want only the contiguous comment %q", docs["run_tests"], want)
	}
	if orphan := shDocComments("# orphan comment\n\nnoop() {\n  :\n}\n"); orphan["noop"] != "" {
		t.Errorf("a comment separated by a blank line must not be attributed, got %q", orphan["noop"])
	}
}

func TestShDocComments_NoComments_ReturnsNil(t *testing.T) {
	if docs := shDocComments("echo hi\n\nnoop() {\n  :\n}\n"); docs != nil {
		t.Errorf("expected nil docs for a file with no doc comments, got %v", docs)
	}
}

// A shebang is an interpreter directive, not prose. Ruby and Shell scripts both
// open with one, and it sits directly above the first declaration often enough
// that attributing it would put "/usr/bin/env bash" into the embedding as that
// function's description.
func TestHashDocComments_ShebangIsNeverADocComment(t *testing.T) {
	cases := []struct {
		name    string
		extract func(string) map[string]string
		src     string
		key     string
	}{
		{
			name:    "ruby bare shebang",
			extract: rubyDocComments,
			src:     "#!/usr/bin/env ruby\ndef build\nend\n",
			key:     "build",
		},
		{
			name:    "ruby shebang then real comment",
			extract: rubyDocComments,
			src:     "#!/usr/bin/env ruby\n# Builds the project.\ndef build\nend\n",
			key:     "build",
		},
		{
			name:    "shell bare shebang",
			extract: shDocComments,
			src:     "#!/usr/bin/env bash\nfunction build {\n  :\n}\n",
			key:     "build",
		},
		{
			name:    "shell shebang then real comment",
			extract: shDocComments,
			src:     "#!/usr/bin/env bash\n# Builds the project.\nfunction build {\n  :\n}\n",
			key:     "build",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			docs := tc.extract(tc.src)
			if strings.Contains(docs[tc.key], "/usr/bin/env") {
				t.Errorf("docs[%q] = %q, want no interpreter path in the doc comment", tc.key, docs[tc.key])
			}
			want := ""
			if strings.Contains(tc.src, "Builds the project.") {
				want = "Builds the project."
			}
			if docs[tc.key] != want {
				t.Errorf("docs[%q] = %q, want %q", tc.key, docs[tc.key], want)
			}
		})
	}
}

// PowerShell's block comment form, in both the multi-line and single-line
// spellings. The single-line form needs its own branch in hashDocText: the
// scan has to accept the "<#"-prefixed line as a comment, and the text has to
// drop the closing "#>" that the multi-line form puts on its own line.
func TestPsDocComments_BlockComment(t *testing.T) {
	multi := `<#
Publishes the built artifacts.
Assumes the caller already authenticated.
#>
function Publish-Artifacts {
  Write-Host "publish"
}
`
	docs := psDocComments(multi)
	want := "Publishes the built artifacts.\nAssumes the caller already authenticated."
	if docs["Publish-Artifacts"] != want {
		t.Errorf("docs[Publish-Artifacts] = %q, want %q", docs["Publish-Artifacts"], want)
	}

	single := "<# Reads the deployment target. #>\nfunction Get-DeployTarget {\n  return 1\n}\n"
	docs = psDocComments(single)
	if docs["Get-DeployTarget"] != "Reads the deployment target." {
		t.Errorf("docs[Get-DeployTarget] = %q, want %q", docs["Get-DeployTarget"], "Reads the deployment target.")
	}
}

// A "#>" line with no "<#" above it is a dangling delimiter, not a comment block
// that can be attributed to the declaration below — pairing it with a distant
// opener would attach an unrelated paragraph.
func TestPsDocComments_DanglingBlockClose_IsNotADocComment(t *testing.T) {
	src := "Write-Host 1\n#>\nfunction Publish-Artifacts {\n  Write-Host \"publish\"\n}\n"
	if docs := psDocComments(src); docs != nil {
		t.Errorf("expected nil for a dangling #> with no opening <#, got %v", docs)
	}
}

// PowerShell's comment-based help puts tag keywords (.SYNOPSIS, .PARAMETERS)
// in the same block as the prose, the way JSDoc puts @param in the same block.
// They are kept verbatim and nodeText's firstLine() decides what reaches the
// embedding — the same trade jsdocs.go makes for @param lines — so the
// behaviour is pinned here rather than left to whichever line happens to come
// first in someone's file.
func TestPsDocComments_TagLinesKeptVerbatim(t *testing.T) {
	src := `<#
.SYNOPSIS
Publishes the built artifacts.
.PARAMETER Target
The environment to publish to.
#>
function Publish-Artifacts {
  Write-Host "publish"
}
`
	docs := psDocComments(src)
	want := ".SYNOPSIS\nPublishes the built artifacts.\n.PARAMETER Target\nThe environment to publish to."
	if docs["Publish-Artifacts"] != want {
		t.Errorf("docs[Publish-Artifacts] = %q, want %q", docs["Publish-Artifacts"], want)
	}
	if first := firstLine(docs["Publish-Artifacts"]); first != ".SYNOPSIS" {
		t.Errorf("firstLine = %q, want %q — tag lines are kept, so the first line is whatever "+
			"the author wrote first", first, ".SYNOPSIS")
	}
}

// PowerShell is a Windows-first language, so CRLF is the common case on disk:
// a stray "\r" must not end up inside the doc text or stop the declaration from
// being recognised.
func TestPsDocComments_CRLFFile(t *testing.T) {
	src := strings.Join([]string{
		"# Reads the deployment target.",
		"function Get-DeployTarget {",
		"  return 1",
		"}",
		"",
	}, "\r\n")

	docs := psDocComments(src)
	if docs["Get-DeployTarget"] != "Reads the deployment target." {
		t.Errorf("docs[Get-DeployTarget] = %q, want %q", docs["Get-DeployTarget"], "Reads the deployment target.")
	}
}

func TestPsDocComments_NoComments_ReturnsNil(t *testing.T) {
	if docs := psDocComments("Write-Host 'hi'\nfunction Get-Config {\n  return $null\n}\n"); docs != nil {
		t.Errorf("expected nil docs for a file with no doc comments, got %v", docs)
	}
}

// hashCommentLine is the one predicate all three languages share, so its
// exclusions are tested directly rather than only through a language.
func TestHashCommentLine(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"# a comment", true},
		{"#comment without a space", true},
		{"\t# indented comment", true},
		{"## an RDoc heading", true},
		{"#!/usr/bin/env bash", false},
		{"#!/usr/bin/env ruby", false},
		{"", false},
		{"build() {", false},
		{"<# block", false},
		// "#>" is a "#" comment line as far as this predicate is concerned —
		// it means nothing in Ruby or Shell, and PowerShell's block closer is
		// matched by hashBlock *before* this predicate runs in both
		// hashCommentLinesAbove and hashDocText.
		{"#> close", true},
	}
	for _, tc := range cases {
		if got := hashCommentLine(strings.TrimSpace(tc.line)); got != tc.want {
			t.Errorf("hashCommentLine(%q) = %t, want %t", tc.line, got, tc.want)
		}
	}
}

// hashDocText is the shared cleaner, so its failure modes are tested here: a
// range that is not all comment content yields nothing (rather than embedding
// code as prose), a block comment's body is kept verbatim, and a blank line
// only survives where a blank line is meaningful.
func TestHashDocText(t *testing.T) {
	if got := hashDocText([]string{"# first line", "echo code", "# last line"}, hashNoBlock); got != "" {
		t.Errorf("hashDocText with a code line = %q, want \"\"", got)
	}

	block := []string{"<#", "first line", "second line", "#>"}
	want := "first line\nsecond line"
	if got := hashDocText(block, hashPSBlock); got != want {
		t.Errorf("hashDocText(block) = %q, want %q", got, want)
	}

	// A blank line ends a "#" run (see hashCommentLinesAbove), so a range like
	// this never reaches the cleaner — treating it as unattributable keeps the
	// "comment text only" guarantee true even if the scan changes. Blank lines
	// are dropped inside a block comment instead, where they are paragraph
	// breaks rather than separators.
	if got := hashDocText([]string{"# first line", "", "# second line"}, hashNoBlock); got != "" {
		t.Errorf("hashDocText with a blank line in a run = %q, want \"\"", got)
	}
	paragraphs := []string{"<#", "first line", "", "second line", "#>"}
	if got := hashDocText(paragraphs, hashPSBlock); got != want {
		t.Errorf("hashDocText with a blank line in a block = %q, want %q", got, want)
	}

	if got := hashDocText(nil, hashNoBlock); got != "" {
		t.Errorf("hashDocText(nil) = %q, want \"\"", got)
	}
}
