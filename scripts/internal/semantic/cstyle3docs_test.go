package semantic

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/graph"
)

// Unit tests for the C-family, Swift, and PHP extractors in cstyle3docs.go,
// plus a drift test per language that checks the names they key against a real
// graph.Build (see the bottom of this file and the reasoning in drift_test.go).

func TestCFamilyDocComments_StructAndSameLineBraceFunction(t *testing.T) {
	src := `/**
 * Point is a 2D coordinate pair.
 */
struct Point {
  int x;
  int y;
};

/* add sums two integers. */
int add(int a, int b) {
  return a + b;
}

int undocumented(int a) {
  return a;
}
`
	docs := cFamilyDocComments(src)
	if want := "Point is a 2D coordinate pair."; docs["Point"] != want {
		t.Errorf("docs[Point] = %q, want %q", docs["Point"], want)
	}
	if want := "add sums two integers."; docs["add"] != want {
		t.Errorf("docs[add] = %q, want %q", docs["add"], want)
	}
	if _, ok := docs["undocumented"]; ok {
		t.Errorf("undocumented should have no doc comment, got %q", docs["undocumented"])
	}
}

// The C-family shape that breaks a naive line scanner: the return type shares
// the line with the name, the "{" opens on the *next* line, and the doc comment
// ("///", the C-family answer to Swift's DocC style) sits above the whole
// declaration. The extracted text must not keep the third slash.
func TestCFamilyDocComments_TripleSlashDocOnAllmanStyleDeclaration(t *testing.T) {
	src := `/// scaleAll multiplies a coordinate by factor.
double scaleAll(struct Point p, double factor)
{
  return p.x * factor;
}
`
	docs := cFamilyDocComments(src)
	want := "scaleAll multiplies a coordinate by factor."
	if docs["scaleAll"] != want {
		t.Errorf("docs[scaleAll] = %q, want %q", docs["scaleAll"], want)
	}
	if strings.Contains(docs["scaleAll"], "/") {
		t.Errorf("docs[scaleAll] = %q, want the /// marker stripped, not left in the text", docs["scaleAll"])
	}
}

// A "///" line whose text legitimately starts with "/" must survive: only the
// marker is stripped, not a slash out of the comment body.
func TestCFamilyDocComments_TripleSlashBodyStartingWithSlash(t *testing.T) {
	src := `/// /etc/plaesy.conf is the config this reads.
int loadConfig(void)
{
  return 0;
}
`
	docs := cFamilyDocComments(src)
	want := "/etc/plaesy.conf is the config this reads."
	if docs["loadConfig"] != want {
		t.Errorf("docs[loadConfig] = %q, want %q", docs["loadConfig"], want)
	}
}

// A doc comment above a C++ method inside a class belongs to the method, and a
// documented class gets its own comment — the two must not collide.
func TestCFamilyDocComments_ClassAndDocumentedMethod(t *testing.T) {
	src := `class Buffer {
  /** Grows the buffer to hold at least size bytes. */
  void reserve(int size) {
    if (size > 0) {
      capacity = size;
    }
  }
};

/** Buffer stores a growable byte buffer. */
class Ring {
};
`
	docs := cFamilyDocComments(src)
	if want := "Grows the buffer to hold at least size bytes."; docs["reserve"] != want {
		t.Errorf("docs[reserve] = %q, want %q", docs["reserve"], want)
	}
	if want := "Buffer stores a growable byte buffer."; docs["Ring"] != want {
		t.Errorf("docs[Ring] = %q, want %q", docs["Ring"], want)
	}
	// "if (size > 0) {" and "capacity = size;" are not declarations, so they
	// must not be keyed even though they sit inside documented ones.
	for _, name := range []string{"if", "capacity", "Buffer"} {
		if _, ok := docs[name]; ok {
			t.Errorf("extractor keyed a doc comment under %q, which cFamilySymbols never extracts", name)
		}
	}
}

// A prototype declaration ("int area();") ends in ";" and has no brace, so
// internal/graph's cLikeFuncHits never records it as a symbol. Keying a doc
// comment to it would produce a doc nodeText can never look up.
func TestCFamilyDocComments_PrototypeAndControlFlowAreNotKeys(t *testing.T) {
	src := `class Shape {
  /* Not a definition: this is only a prototype. */
  int area();
};

void run(void)
{
  if (ready) {
    while (true) {
    }
  }
}
`
	docs := cFamilyDocComments(src)
	for _, name := range []string{"area", "if", "while", "run"} {
		if _, ok := docs[name]; ok {
			t.Errorf("extractor keyed a doc comment under %q, but graph extracts no such symbol here", name)
		}
	}
	if _, ok := docs["Shape"]; ok {
		t.Errorf("Shape has no comment above it, got %q", docs["Shape"])
	}
}

// A C++20 attribute line ("[[nodiscard]]", "__attribute__((...))") between the
// doc comment and the declaration needs no special handling here, because such a
// line is never a declaration line for internal/graph either: every C-family
// name regex starts with an identifier or the class/struct keyword, so a line
// starting with "[" or "_" matches nothing and Node.Symbols has no entry to look
// up. Pinning that here keeps the asymmetry with Swift/PHP attributes visible
// and prevents a "fix" that would start keying names graph never extracts.
func TestCFamilyDocComments_AttributeLineIsNotADeclarationForGraphEither(t *testing.T) {
	src := `/** Doubles a, discarding the result if unused. */
[[nodiscard]] int doubled(int a)
{
  return a * 2;
}
`
	if docs := cFamilyDocComments(src); len(docs) != 0 {
		t.Errorf("expected no docs: graph extracts no symbol from an attribute line, got %v", sortedKeys(docs))
	}
}

// A blank line detaches a comment from the declaration below it, the same rule
// Go's doc-comment grammar and jsCommentBlock use — otherwise a comment
// belonging to some earlier statement would be merged into the next
// declaration's docs.
func TestCFamilyDocComments_BlankLineDetachesComment(t *testing.T) {
	src := `/* freeHelper releases a helper handle. */

int detachedHelper(int a)
{
  return a;
}
`
	docs := cFamilyDocComments(src)
	if len(docs) != 0 {
		t.Errorf("expected no docs for a comment separated by a blank line, got %v", sortedKeys(docs))
	}
}

func TestCFamilyDocComments_NoComments_ReturnsNil(t *testing.T) {
	docs := cFamilyDocComments("int main(int argc, char **argv) {\n  return 0;\n}\n")
	if docs != nil {
		t.Errorf("expected nil docs for a file with no comments, got %v", docs)
	}
}

func TestCFamilyDocComments_UnterminatedBlockComment_ReturnsNil(t *testing.T) {
	src := "int value = 1;\n*/\nint add(int a, int b)\n{\n  return a + b;\n}\n"
	if docs := cFamilyDocComments(src); docs != nil {
		t.Errorf("expected nil for a dangling */ with no opening /*, got %v", docs)
	}
}

func TestSwiftDocComments_FuncAndType(t *testing.T) {
	src := `/// Loads a user record from the store by id.
func loadUser(id: String) -> User? {
  return store.load(id)
}

/** Account holds the current user's profile. */
class Account {
  /// Renders the account panel.
  func render() {}

  func undocumentedRefresh() {}
}
`
	docs := swiftDocComments(src)
	if want := "Loads a user record from the store by id."; docs["loadUser"] != want {
		t.Errorf("docs[loadUser] = %q, want %q", docs["loadUser"], want)
	}
	if want := "Account holds the current user's profile."; docs["Account"] != want {
		t.Errorf("docs[Account] = %q, want %q", docs["Account"], want)
	}
	if want := "Renders the account panel."; docs["render"] != want {
		t.Errorf("docs[render] = %q, want %q", docs["render"], want)
	}
	for _, name := range []string{"undocumentedRefresh", "User", "store"} {
		if _, ok := docs[name]; ok {
			t.Errorf("extractor keyed a doc comment under %q, which swiftSymbols never extracts", name)
		}
	}
}

// Access-level and other modifiers sit between the comment and the keyword, and
// swiftSymbols keeps them, so the keyed name must still be the bare function
// name.
func TestSwiftDocComments_ModifiersDoNotChangeTheKey(t *testing.T) {
	src := `/// A documented public static entry point.
public static func makeSession(user: User) -> Session {
  return Session(user: user)
}
`
	docs := swiftDocComments(src)
	want := "A documented public static entry point."
	if docs["makeSession"] != want {
		t.Errorf("docs[makeSession] = %q, want %q", docs["makeSession"], want)
	}
}

func TestSwiftDocComments_BlankLineDetachesComment(t *testing.T) {
	src := `/// DetachedHelper is not documented as far as the graph is concerned.

func detachedHelper() {
}
`
	if docs := swiftDocComments(src); len(docs) != 0 {
		t.Errorf("expected no docs for a comment separated by a blank line, got %v", sortedKeys(docs))
	}
}

// Swift attributes sit between the "///" comment and the declaration, and
// "@available" above public API is close to universal. Treating the attribute
// line as the end of the comment block drops the documentation of every
// annotated declaration — the Rust "#[derive(Debug)]" defect, in Swift spelling.
// The attribute must be stepped over, and must not appear in the doc text.
func TestSwiftDocComments_AttributesBetweenDocAndDeclaration(t *testing.T) {
	src := `/// Starts a session, when the platform supports it.
@available(iOS 15.0, *)
@discardableResult
public func startSession(user: User) -> Session {
  return Session(user: user)
}

/** CachedImage holds a decoded image. */
@objc(CachedImageObjc)
final class CachedImage {
  /// Decodes the image at path.
  @MainActor
  func decode(at path: String) -> UIImage? {
    return nil
  }
}

/** A multi-line availability note.
 *
 * Deprecated since 1.0.
 */
@available(
    *,
    deprecated: "1.0",
    message: "use startSession(user:) instead"
)
func deprecatedEntryPoint() {
}
`
	docs := swiftDocComments(src)
	if want := "Starts a session, when the platform supports it."; docs["startSession"] != want {
		t.Errorf("docs[startSession] = %q, want %q", docs["startSession"], want)
	}
	if want := "CachedImage holds a decoded image."; docs["CachedImage"] != want {
		t.Errorf("docs[CachedImage] = %q, want %q", docs["CachedImage"], want)
	}
	if want := "Decodes the image at path."; docs["decode"] != want {
		t.Errorf("docs[decode] = %q, want %q", docs["decode"], want)
	}
	if want := "A multi-line availability note.\nDeprecated since 1.0."; docs["deprecatedEntryPoint"] != want {
		t.Errorf("docs[deprecatedEntryPoint] = %q, want %q", docs["deprecatedEntryPoint"], want)
	}
	for _, name := range []string{"available", "discardableResult", "objc", "MainActor"} {
		if _, ok := docs[name]; ok {
			t.Errorf("attribute %q leaked into the doc set", name)
		}
	}
	if strings.Contains(docs["startSession"], "iOS") || strings.Contains(docs["deprecatedEntryPoint"], "deprecated:") {
		t.Errorf("attribute text leaked into the doc text: %q / %q", docs["startSession"], docs["deprecatedEntryPoint"])
	}
}

// A blank line still ends the walk, so an attribute run cannot bridge a
// paragraph break and pull an unrelated comment down onto the declaration.
func TestSwiftDocComments_BlankLineBeforeAttributesStillDetaches(t *testing.T) {
	src := `/// DetachedHelper is not documented as far as the graph is concerned.

@available(iOS 15.0, *)
func detachedHelper() {
}
`
	if docs := swiftDocComments(src); len(docs) != 0 {
		t.Errorf("expected no docs for a comment separated by a blank line, got %v", sortedKeys(docs))
	}
}

func TestSwiftDocComments_NoComments_ReturnsNil(t *testing.T) {
	if docs := swiftDocComments("func noop() {}\n"); docs != nil {
		t.Errorf("expected nil docs for a file with no comments, got %v", docs)
	}
}

func TestPHPDocComments_ClassAndMethod(t *testing.T) {
	src := `<?php

/** AccountRepository reads account records from the database. */
class AccountRepository
{
    /**
     * Finds one account by id.
     *
     * @param string $id the account id
     */
    public function find($id)
    {
        return $this->rows[$id] ?? null;
    }

    private function undocumentedDelete($id)
    {
        unset($this->rows[$id]);
    }
}

// FormatAccount renders an account as a JSON string.
function formatAccount(array $account): string
{
    return json_encode($account);
}
`
	docs := phpDocComments(src)
	if want := "AccountRepository reads account records from the database."; docs["AccountRepository"] != want {
		t.Errorf("docs[AccountRepository] = %q, want %q", docs["AccountRepository"], want)
	}
	if want := "Finds one account by id.\n@param string $id the account id"; docs["find"] != want {
		t.Errorf("docs[find] = %q, want %q", docs["find"], want)
	}
	if want := "FormatAccount renders an account as a JSON string."; docs["formatAccount"] != want {
		t.Errorf("docs[formatAccount] = %q, want %q", docs["formatAccount"], want)
	}
	for _, name := range []string{"undocumentedDelete", "json_encode", "rows"} {
		if _, ok := docs[name]; ok {
			t.Errorf("extractor keyed a doc comment under %q, which phpSymbols never extracts", name)
		}
	}
}

// interface/trait declarations take the same PHPDoc block as a class.
func TestPHPDocComments_InterfaceAndTrait(t *testing.T) {
	src := `<?php

/** Storage persists account records. */
interface Storage
{
}

/** MemoryStorage keeps records for the process lifetime. */
trait MemoryStorage
{
}
`
	docs := phpDocComments(src)
	if want := "Storage persists account records."; docs["Storage"] != want {
		t.Errorf("docs[Storage] = %q, want %q", docs["Storage"], want)
	}
	if want := "MemoryStorage keeps records for the process lifetime."; docs["MemoryStorage"] != want {
		t.Errorf("docs[MemoryStorage] = %q, want %q", docs["MemoryStorage"], want)
	}
}

// PHP 8 attributes are conventionally written between the docblock and the
// declaration, so a docblock above an attributed method must still be found —
// and the attribute itself must not become the doc text.
func TestPHPDocComments_AttributesBetweenDocAndDeclaration(t *testing.T) {
	src := `<?php

class AccountController
{
    /**
     * Removes an account permanently.
     */
    #[Route('/accounts/{id}', methods: ['DELETE'])]
    public function purge($id)
    {
        unset($this->rows[$id]);
    }

    /**
     * Lists accounts, newest first.
     */
    #[Route(
        '/accounts'
    )]
    #[MapRequestPayload]
    public function index()
    {
        return array_keys($this->rows);
    }
}
`
	docs := phpDocComments(src)
	if want := "Removes an account permanently."; docs["purge"] != want {
		t.Errorf("docs[purge] = %q, want %q", docs["purge"], want)
	}
	if want := "Lists accounts, newest first."; docs["index"] != want {
		t.Errorf("docs[index] = %q, want %q", docs["index"], want)
	}
	if strings.Contains(docs["purge"], "Route") || strings.Contains(docs["index"], "MapRequestPayload") {
		t.Errorf("attribute text leaked into the doc text: %q / %q", docs["purge"], docs["index"])
	}
}

// A blank line still ends the walk, so an attribute run cannot bridge a
// paragraph break and pull an unrelated comment onto the declaration.
func TestPHPDocComments_BlankLineBeforeAttributesStillDetaches(t *testing.T) {
	src := `<?php

// detachedHelper does something.

#[Route('/x')]
function detachedHelper()
{
    return 1;
}
`
	if docs := phpDocComments(src); len(docs) != 0 {
		t.Errorf("expected no docs for a comment separated by a blank line, got %v", sortedKeys(docs))
	}
}

func TestPHPDocComments_BlankLineDetachesComment(t *testing.T) {
	src := `<?php

// detachedHelper does something.

function detachedHelper()
{
    return 1;
}
`
	if docs := phpDocComments(src); len(docs) != 0 {
		t.Errorf("expected no docs for a comment separated by a blank line, got %v", sortedKeys(docs))
	}
}

func TestPHPDocComments_NoComments_ReturnsNil(t *testing.T) {
	if docs := phpDocComments("<?php\nfunction noop() {}\n"); docs != nil {
		t.Errorf("expected nil docs for a file with no comments, got %v", docs)
	}
}

// The *DocsFor wrappers are the seam docsFor (semantic.go) dispatches to, so
// they carry the extension lists themselves: an extension symbols.go does not
// extract symbols for must yield nil, never a doc map nothing will read.
func TestCFamilyDocsFor_ExtensionList(t *testing.T) {
	for _, ext := range []string{".c", ".h", ".cpp", ".hpp", ".cc", ".hh", ".cxx", ".C", ".H"} {
		t.Run(ext, func(t *testing.T) {
			repoRoot := t.TempDir()
			rel := "native/engine" + ext
			writeFile(t, filepath.Join(repoRoot, filepath.FromSlash(rel)), "/* add sums two integers. */\nint add(int a, int b) {\n  return a + b;\n}\n")
			n := graph.Node{ID: rel, Symbols: []string{"add"}}
			if doc := cFamilyDocsFor(repoRoot, n)["add"]; doc != "add sums two integers." {
				t.Errorf("cFamilyDocsFor(%q) = %q, want the doc comment", rel, doc)
			}
		})
	}
	for _, ext := range []string{".rs", ".cs", ".swift", ".txt", ""} {
		t.Run("skip"+ext, func(t *testing.T) {
			if docs := cFamilyDocsFor(t.TempDir(), graph.Node{ID: "native/engine" + ext}); docs != nil {
				t.Errorf("cFamilyDocsFor(%q) = %v, want nil", "native/engine"+ext, docs)
			}
		})
	}
}

func TestSwiftDocsFor_Extension(t *testing.T) {
	repoRoot := t.TempDir()
	writeFile(t, filepath.Join(repoRoot, "svc", "session.swift"), "/// Loads a user record by id.\nfunc loadUser(id: String) -> User? {\n  return store.load(id)\n}\n")
	n := graph.Node{ID: "svc/session.swift", Symbols: []string{"loadUser"}}
	if want := "Loads a user record by id."; swiftDocsFor(repoRoot, n)["loadUser"] != want {
		t.Errorf("swiftDocsFor = %q, want %q", swiftDocsFor(repoRoot, n)["loadUser"], want)
	}
	for _, ext := range []string{".rs", ".swift.bak", ".txt", ""} {
		if docs := swiftDocsFor(t.TempDir(), graph.Node{ID: "svc/session" + ext}); docs != nil {
			t.Errorf("swiftDocsFor(%q) = %v, want nil", "svc/session"+ext, docs)
		}
	}
}

func TestPHPDocsFor_Extension(t *testing.T) {
	repoRoot := t.TempDir()
	writeFile(t, filepath.Join(repoRoot, "app", "accounts.php"), "<?php\n\n/** AccountRepository reads account records. */\nclass AccountRepository\n{\n}\n")
	n := graph.Node{ID: "app/accounts.php", Symbols: []string{"AccountRepository"}}
	if want := "AccountRepository reads account records."; phpDocsFor(repoRoot, n)["AccountRepository"] != want {
		t.Errorf("phpDocsFor = %q, want %q", phpDocsFor(repoRoot, n)["AccountRepository"], want)
	}
	for _, ext := range []string{".rb", ".php.bak", ".txt", ""} {
		if docs := phpDocsFor(t.TempDir(), graph.Node{ID: "app/accounts" + ext}); docs != nil {
			t.Errorf("phpDocsFor(%q) = %v, want nil", "app/accounts"+ext, docs)
		}
	}
}

// A node whose file is missing on disk must degrade to nil rather than fail —
// doc comments are best-effort signal, never a reason to fail indexing.
func TestCFamilySwiftPHPDocsFor_FileMissingOnDisk_ReturnsNil(t *testing.T) {
	repoRoot := t.TempDir()
	if docs := cFamilyDocsFor(repoRoot, graph.Node{ID: "native/gone.cpp"}); docs != nil {
		t.Errorf("cFamilyDocsFor on a missing file = %v, want nil", docs)
	}
	if docs := swiftDocsFor(repoRoot, graph.Node{ID: "svc/gone.swift"}); docs != nil {
		t.Errorf("swiftDocsFor on a missing file = %v, want nil", docs)
	}
	if docs := phpDocsFor(repoRoot, graph.Node{ID: "app/gone.php"}); docs != nil {
		t.Errorf("phpDocsFor on a missing file = %v, want nil", docs)
	}
}

// Drift tests. Same contract as TestDocExtractorNamesMatchGraphSymbols in
// drift_test.go, restated per language here because cstyle3docs.go adds three
// more private copies of internal/graph/symbols.go's regexes: the keys must be
// exactly the names Node.Symbols carries, or nodeText's lookup silently never
// fires and enrichment rots with nothing reporting it.
//
// Both directions are asserted, and both are needed. Checking only that every
// keyed name appears in Node.Symbols is not sufficient: breaking a copied regex
// makes the extractor produce *fewer* docs, and the few that remain are still
// legitimately keyed, so a subset-only check passes while enrichment has
// quietly stopped working.
//
//  1. every documented symbol in the fixture gets a doc   (catches drift),
//  2. every keyed doc name is a real graph symbol           (catches mis-keying),
//  3. an undocumented fixture symbol gets no doc           (catches over-reach).
//
// cfamilySwiftPHPDriftCheck is named for the three languages it covers so it
// cannot collide with another language agent's helper in this package.
func TestCFamilyDocExtractorNamesMatchGraphSymbols(t *testing.T) {
	// .hh and .cxx are deliberately absent: symbols.go extracts symbols for
	// them, but graph's file scan (includeExtRE) never collects those two
	// extensions, so graph.Build cannot produce a node to compare against.
	for _, rel := range []string{"native/engine.c", "native/engine.h", "native/engine.cpp", "native/engine.hpp", "native/engine.cc"} {
		t.Run(rel, func(t *testing.T) {
			cfamilySwiftPHPDriftCheck(t, rel, cfamilyDriftFixture, cFamilyDocComments,
				[]string{"Point", "add", "scaleAll", "reserve"})
		})
	}
}

func TestSwiftDocExtractorNamesMatchGraphSymbols(t *testing.T) {
	cfamilySwiftPHPDriftCheck(t, "svc/session.swift", swiftDriftFixture, swiftDocComments,
		[]string{"loadUser", "Account", "render", "startSession"})
}

func TestPHPDocExtractorNamesMatchGraphSymbols(t *testing.T) {
	cfamilySwiftPHPDriftCheck(t, "app/accounts.php", phpDriftFixture, phpDocComments,
		[]string{"AccountRepository", "find", "formatAccount", "purge"})
}

func cfamilySwiftPHPDriftCheck(t *testing.T, rel, content string, extract func(string) map[string]string, wantDocs []string) {
	t.Helper()

	repoRoot := t.TempDir()
	outDir := t.TempDir()
	path := filepath.Join(repoRoot, filepath.FromSlash(rel))
	writeFile(t, path, content)

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
		t.Fatalf("graph.Build produced no symbols for %s (ids: %v)", rel, nodeIDs(g))
	}
	symbolSet := make(map[string]bool, len(symbols))
	for _, s := range symbols {
		symbolSet[s] = true
	}

	docs := extract(content)

	// (1) every documented symbol must have been recognised.
	for _, want := range wantDocs {
		if _, ok := docs[want]; !ok {
			t.Errorf("extractor found no doc comment for documented symbol %q; got %v. "+
				"If graph.Build still extracts it (%v) but this package does not, the copied "+
				"regex has drifted from internal/graph/symbols.go and enrichment is silently dead",
				want, sortedKeys(docs), symbols)
		}
		if !symbolSet[want] {
			t.Errorf("fixture symbol %q is documented but absent from graph.Node.Symbols %v — "+
				"the fixture and internal/graph/symbols.go disagree", want, symbols)
		}
	}

	// (2) nothing may be keyed under a name graph does not extract,
	// (3) and an undocumented symbol must stay out of the doc set.
	for name := range docs {
		if !symbolSet[name] {
			t.Errorf("extractor keyed a doc comment under %q, but graph.Build extracted %v — "+
				"nodeText would never look this comment up", name, symbols)
		}
		if !containsString(wantDocs, name) {
			t.Errorf("extractor produced a doc for undocumented symbol %q; the fixture documents only %v",
				name, wantDocs)
		}
	}
}

// cfamilyDriftFixture deliberately mixes shapes that are easy to get wrong for
// C-family: a documented struct, a "/* */" comment above a same-line-brace
// function, a "///" comment above a declaration whose brace opens on the next
// line, a documented method inside a class, an undocumented class, an
// undocumented function, a prototype, and control flow. Anything the extractor
// recognises here must also survive graph.Build's own extraction, and the two
// undocumented symbols must stay out of the doc set.
const cfamilyDriftFixture = `/**
 * Point is a 2D coordinate pair.
 */
struct Point {
  int x;
  int y;
};

/* add sums two integers. */
int add(int a, int b) {
  return a + b;
}

/// scaleAll multiplies a coordinate by factor.
double scaleAll(struct Point p, double factor)
{
  return p.x * factor;
}

class Buffer {
  /** Grows the buffer to hold at least size bytes. */
  void reserve(int size) {
    if (size > 0) {
      capacity = size;
    }
  }

  int area();
};

int undocumented(int a) {
  return a;
}
`

// swiftDriftFixture mixes Swift's two doc styles ("///" and "/** */"), a method
// inside a type, an attribute-decorated declaration, and two undocumented
// funcs. The annotated declaration is in the fixture on purpose: without it a
// regression that drops every attributed declaration's documentation would leave
// this test green.
const swiftDriftFixture = `import Foundation

/// Loads a user record from the store by id.
func loadUser(id: String) -> User? {
  return store.load(id)
}

/// Starts a session, when the platform supports it.
@available(iOS 15.0, *)
@discardableResult
public func startSession(user: User) -> Session {
  return Session(user: user)
}

/** Account holds the current user's profile. */
class Account {
  /// Renders the account panel.
  func render() {}

  func undocumentedRefresh() {}
}

func undocumentedHelper() -> Int {
  return 0
}
`

// phpDriftFixture mixes a documented class, a documented method with a
// multi-line PHPDoc block, an attribute-decorated documented method, a
// documented plain function, and an undocumented method. The attributed method
// is in the fixture on purpose: without it a regression that drops every
// attributed method's documentation would leave this test green.
const phpDriftFixture = `<?php

/** AccountRepository reads account records from the database. */
class AccountRepository
{
    /**
     * Finds one account by id.
     *
     * @param string $id the account id
     */
    public function find($id)
    {
        return $this->rows[$id] ?? null;
    }

    /**
     * Removes an account permanently.
     */
    #[Route('/accounts/{id}', methods: ['DELETE'])]
    public function purge($id)
    {
        unset($this->rows[$id]);
    }

    private function undocumentedDelete($id)
    {
        unset($this->rows[$id]);
    }
}

// FormatAccount renders an account as a JSON string.
function formatAccount(array $account): string
{
    return json_encode($account);
}
`
