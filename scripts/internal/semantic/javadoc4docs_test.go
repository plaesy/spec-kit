package semantic

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/plaesy/spec-kit/internal/graph"
)

// This file covers the four Javadoc/KDoc-shaped extractors: Dart (.dart),
// Java (.java), Kotlin (.kt/.kts) and C# (.cs). Per-language unit tests pin the
// doc-comment grammar; the drift test at the bottom pins the one coupling this
// package cannot express in the type system — the names nodeText looks docs up
// by must be the names graph.Node.Symbols actually contains.
//
// Assertions on "///" line-doc text are Contains, not exact equality: jsDocText
// strips exactly two leading slashes, so the text keeps a stray "/" in front of
// it ("/ Loads the user."). That is a known cosmetic artifact of the shared
// helper, and pinning the artifact here would make these tests fail the moment
// whoever owns jsdocs.go tightens it. "/** ... */" blocks come out clean and
// are asserted exactly.

// Dart: "///" line doc above the class and above a typed method, an
// undocumented method that must stay out of the doc set.
func TestDartDocComments_ClassAndMethod(t *testing.T) {
	src := `/// Stores user records for the app.
class UserRepo {
  /** Loads one user by id. */
  Future<User> loadUser(String id) {
    return store.get(id);
  }

  void evict(String id) {
    store.remove(id);
  }
}
`
	docs := dartDocComments(src)
	if got := docs["loadUser"]; got != "Loads one user by id." {
		t.Errorf("docs[loadUser] = %q, want %q", got, "Loads one user by id.")
	}
	if got := docs["UserRepo"]; got != "Stores user records for the app." {
		t.Errorf("docs[UserRepo] = %q, want the /// line doc with no leftover slash", got)
	}
	if _, ok := docs["evict"]; ok {
		t.Errorf("evict is undocumented, got %q", docs["evict"])
	}
	if len(docs) != 2 {
		t.Errorf("docs = %v, want exactly the 2 documented symbols", sortedKeys(docs))
	}
}

// dartSymbols keys off `class`, optionally preceded by `abstract`, so the
// modifier must not change the extracted name.
func TestDartDocComments_AbstractClassKey(t *testing.T) {
	docs := dartDocComments("/** Base repository. */\nabstract class BaseRepo {\n}\n")
	if got := docs["BaseRepo"]; got != "Base repository." {
		t.Errorf("docs[BaseRepo] = %q, want %q", got, "Base repository.")
	}
}

func TestDartDocComments_NoComments_ReturnsNil(t *testing.T) {
	if docs := dartDocComments("class Plain {\n  void noop() {\n  }\n}\n"); docs != nil {
		t.Errorf("want nil for a file with no doc comments, got %v", docs)
	}
}

// Java: Javadoc above the class and above a method, plus the Allman brace
// style cLikeFuncHits also accepts (signature line, "{" on the next line).
func TestJavaDocComments_ClassAndMethod(t *testing.T) {
	src := `/** Authenticates users against the account store. */
public class UserService {
  /** Finds every account. */
  public List<User> findAll()
  {
    return repo.all();
  }

  public void flush() {
    repo.clear();
  }
}
`
	docs := javaDocComments(src)
	if got := docs["UserService"]; got != "Authenticates users against the account store." {
		t.Errorf("docs[UserService] = %q", got)
	}
	if got := docs["findAll"]; got != "Finds every account." {
		t.Errorf("docs[findAll] = %q, want %q (Allman brace on the next line)", got, "Finds every account.")
	}
	if _, ok := docs["flush"]; ok {
		t.Errorf("flush is undocumented, got %q", docs["flush"])
	}
	if len(docs) != 2 {
		t.Errorf("docs = %v, want exactly the 2 documented symbols", sortedKeys(docs))
	}
}

// symJavaClassRE covers interface and enum as well as class; all three must
// key under the bare type name.
func TestJavaDocComments_InterfaceAndEnumKeys(t *testing.T) {
	src := `/** A repository of accounts. */
interface AccountRepo {
  /** A billing state. */
  enum BillingState {
    OPEN
  }
}
`
	docs := javaDocComments(src)
	if got := docs["AccountRepo"]; got != "A repository of accounts." {
		t.Errorf("docs[AccountRepo] = %q", got)
	}
	if got := docs["BillingState"]; got != "A billing state." {
		t.Errorf("docs[BillingState] = %q", got)
	}
}

func TestJavaDocComments_NoComments_ReturnsNil(t *testing.T) {
	if docs := javaDocComments("public class Plain {\n  void noop() {\n  }\n}\n"); docs != nil {
		t.Errorf("want nil for a file with no doc comments, got %v", docs)
	}
}

// Kotlin: KDoc is "/** ... */" (same shape as Javadoc), and the "fun" keyword
// wins over the type regex exactly as kotlinSymbols' `continue` does.
func TestKotlinDocComments_FunctionAndClass(t *testing.T) {
	src := `/**
 * Loads a user by id from the local store.
 */
class UserRepo {
  /** Fetches one user record. */
  suspend fun loadUser(id: String): User {
    return store.get(id)
  }

  fun evict(id: String) {
    cache.remove(id)
  }
}
`
	docs := kotlinDocComments(src)
	if got := docs["UserRepo"]; got != "Loads a user by id from the local store." {
		t.Errorf("docs[UserRepo] = %q", got)
	}
	if got := docs["loadUser"]; got != "Fetches one user record." {
		t.Errorf("docs[loadUser] = %q", got)
	}
	if _, ok := docs["evict"]; ok {
		t.Errorf("evict is undocumented, got %q", docs["evict"])
	}
	if len(docs) != 2 {
		t.Errorf("docs = %v, want exactly the 2 documented symbols", sortedKeys(docs))
	}
}

// symKotlinTypeRE accepts a modifier run (data/open/sealed/...) and `object`,
// and symKotlinFunRE accepts suspend/override/... on a function — the keys must
// not include the modifiers.
func TestKotlinDocComments_ModifierKeys(t *testing.T) {
	src := `/** A value holder. */
data class UserRepo {
  /** Overridden loader. */
  override fun loadUser(id: String): User = store.get(id)
}

/** A singleton registry. */
object Registry
`
	docs := kotlinDocComments(src)
	for name, want := range map[string]string{
		"UserRepo": "A value holder.",
		"loadUser": "Overridden loader.",
		"Registry": "A singleton registry.",
	} {
		if got := docs[name]; got != want {
			t.Errorf("docs[%s] = %q, want %q", name, got, want)
		}
	}
}

func TestKotlinDocComments_NoComments_ReturnsNil(t *testing.T) {
	if docs := kotlinDocComments("class Plain {\n  fun noop() {\n  }\n}\n"); docs != nil {
		t.Errorf("want nil for a file with no doc comments, got %v", docs)
	}
}

// C#: the dominant convention is "///" XML doc lines; "/** ... */" works too
// because it is the same block-comment shape.
func TestCSharpDocComments_ClassAndMethod(t *testing.T) {
	src := `/// <summary>Authenticates users against the account store.</summary>
public class UserService {
  /// <summary>Finds every account.</summary>
  public List<User> FindAll() {
    return repo.All();
  }

  public void Flush() {
    repo.Clear();
  }
}
`
	docs := csharpDocComments(src)
	if got := docs["FindAll"]; got != "Finds every account." {
		t.Errorf("docs[FindAll] = %q, want %q — the XML doc tags should be gone", got, "Finds every account.")
	}
	if got := docs["UserService"]; got != "Authenticates users against the account store." {
		t.Errorf("docs[UserService] = %q", got)
	}
	if _, ok := docs["Flush"]; ok {
		t.Errorf("Flush is undocumented, got %q", docs["Flush"])
	}
	if len(docs) != 2 {
		t.Errorf("docs = %v, want exactly the 2 documented symbols", sortedKeys(docs))
	}
}

func TestCSharpDocComments_BlockDocAndStructKey(t *testing.T) {
	src := `/** A point in 2D space. */
public struct Point {
  /** Sums two points. */
  public static Point Add(Point a, Point b) {
    return new Point();
  }
}
`
	docs := csharpDocComments(src)
	if got := docs["Point"]; got != "A point in 2D space." {
		t.Errorf("docs[Point] = %q", got)
	}
	if got := docs["Add"]; got != "Sums two points." {
		t.Errorf("docs[Add] = %q", got)
	}
}

func TestCSharpDocComments_NoComments_ReturnsNil(t *testing.T) {
	if docs := csharpDocComments("public class Plain {\n  void Noop() {\n  }\n}\n"); docs != nil {
		t.Errorf("want nil for a file with no doc comments, got %v", docs)
	}
}

// A blank line detaches a comment from the declaration below it, the same rule
// Go's doc-comment grammar and jsCommentBlock use — otherwise a comment
// belonging to some earlier statement would be attributed to the next
// declaration's docs.
func TestJavadocStyleDocs_BlankLineDetachesComment(t *testing.T) {
	src := `/** Not a doc comment for evict(). */

void evict(String id) {
}
`
	if got := javaDocComments(src)["evict"]; got != "" {
		t.Errorf("docs[evict] = %q, want nothing: the comment is separated by a blank line", got)
	}
	if got := dartDocComments(src)["evict"]; got != "" {
		t.Errorf("dart docs[evict] = %q, want nothing", got)
	}
}

// An unterminated "*/" with no "/*" above it must not be paired with a distant
// block comment — jsCommentBlock gives up, and so must these extractors.
func TestJavadocStyleDocs_DanglingBlockEndIsNotADocComment(t *testing.T) {
	src := "const x = 1;\n*/\nclass Orphan {\n}\n"
	if docs := javaDocComments(src); docs != nil {
		t.Errorf("want nil for a dangling */, got %v", docs)
	}
	if docs := csharpDocComments(src); docs != nil {
		t.Errorf("want nil for a dangling */, got %v", docs)
	}
	if docs := kotlinDocComments(src); docs != nil {
		t.Errorf("want nil for a dangling */, got %v", docs)
	}
}

// The canonical one-symbol shapes, asserted exactly. These mirror the rows
// TestDocsFor_WiredExtensions uses for .dart/.java/.kt/.kts/.cs: a file whose
// whole content is a doc comment plus one declaration, where the doc text must
// come back with no leftover comment marker. The "///" rows are the point —
// jsDocText strips exactly two slashes, so without this file's per-language
// third-slash fix-up every Dart and C# doc would come back as "/ Target docs."
func TestJavadocStyleDocs_MinimalFileShapes(t *testing.T) {
	cases := []struct {
		lang    string
		content string
		extract func(string) map[string]string
		want    string
	}{
		{lang: "dart", content: "/// Target docs.\nclass Target {}\n", extract: dartDocComments, want: "Target docs."},
		{lang: "java", content: "/** Target docs. */\nclass Target {}\n", extract: javaDocComments, want: "Target docs."},
		{lang: "kotlin", content: "/** Target docs. */\nclass Target\n", extract: kotlinDocComments, want: "Target docs."},
		{lang: "csharp", content: "/// Target docs.\nclass Target { }\n", extract: csharpDocComments, want: "Target docs."},
	}

	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			docs := tc.extract(tc.content)
			if len(docs) != 1 {
				t.Fatalf("docs = %v, want exactly one entry", sortedKeys(docs))
			}
			for name, got := range docs {
				if got != tc.want {
					t.Errorf("docs[%q] = %q, want %q — a comment marker leaked into the doc text", name, got, tc.want)
				}
			}
		})
	}
}

// An annotation between the doc comment and the declaration must not detach
// them. This is the dominant real-world shape in Java (@Override above an
// overriding method is close to universal) and C# (every MVC action carries an
// attribute, and C#'s doc convention is "///" lines), so refusing to step over
// it silently drops most members' documentation. The annotation must also be
// left out of the doc text rather than embedded as if it were documentation,
// and an annotation with no doc comment above it must stay undocumented.
func TestJavaDocComments_AnnotationBetweenDocAndDeclaration(t *testing.T) {
	src := `class UserService {
  /** Reloads accounts from the remote store. */
  @Override
  @SuppressWarnings("unchecked")
  public void reload() {
    repo.reload();
  }

  @Deprecated
  public void flush() {
    repo.clear();
  }
}
`
	docs := javaDocComments(src)
	if got := docs["reload"]; got != "Reloads accounts from the remote store." {
		t.Errorf("docs[reload] = %q, want the Javadoc above the two annotations", got)
	}
	if _, ok := docs["flush"]; ok {
		t.Errorf("@Deprecated alone is not a doc comment, got %q", docs["flush"])
	}
	if len(docs) != 1 {
		t.Errorf("docs = %v, want only reload", sortedKeys(docs))
	}
}

func TestDartDocComments_AnnotationBetweenDocAndDeclaration(t *testing.T) {
	src := `class UserRepo {
  /// Reloads records from the remote store.
  @override
  void reload() {
    store.reload();
  }

  @visibleForTesting
  void evict(String id) {
    store.remove(id);
  }
}
`
	docs := dartDocComments(src)
	if got := docs["reload"]; got != "Reloads records from the remote store." {
		t.Errorf("docs[reload] = %q, want the /// line doc above @override", got)
	}
	if _, ok := docs["evict"]; ok {
		t.Errorf("@visibleForTesting alone is not a doc comment, got %q", docs["evict"])
	}
}

func TestKotlinDocComments_AnnotationBetweenDocAndDeclaration(t *testing.T) {
	src := `class UserRepo {
  /** Reloads records from the remote store. */
  @Throws(IOException::class)
  suspend fun reload() {
    store.reload()
  }

  @JvmStatic
  fun evict(id: String) {
    cache.remove(id)
  }
}
`
	docs := kotlinDocComments(src)
	if got := docs["reload"]; got != "Reloads records from the remote store." {
		t.Errorf("docs[reload] = %q, want the KDoc above @Throws", got)
	}
	if _, ok := docs["evict"]; ok {
		t.Errorf("@JvmStatic alone is not a doc comment, got %q", docs["evict"])
	}
}

// C# stacks attributes, so this also pins the multi-line attribute *list* (two
// consecutive lines) rather than just one.
func TestCSharpDocComments_AttributesBetweenDocAndDeclaration(t *testing.T) {
	src := `public class UserService {
  /// <summary>Reloads accounts from the remote store.</summary>
  [HttpGet("api/accounts/reload")]
  [ProducesResponseType(StatusCodes.Status200OK)]
  public string Reload() {
    return repo.Reload();
  }

  [Obsolete]
  public void Flush() {
    repo.Clear();
  }
}
`
	docs := csharpDocComments(src)
	if got := docs["Reload"]; got != "Reloads accounts from the remote store." {
		t.Errorf("docs[Reload] = %q, want the /// doc above the two attributes, tags stripped", got)
	}
	for _, marker := range []string{"HttpGet", "ProducesResponseType"} {
		if strings.Contains(docs["Reload"], marker) {
			t.Errorf("docs[Reload] = %q, want the attribute %q left out of the doc text", docs["Reload"], marker)
		}
	}
	if _, ok := docs["Flush"]; ok {
		t.Errorf("[Obsolete] alone is not a doc comment, got %q", docs["Flush"])
	}
}

// The canonical multi-line C# XML doc form. Its first line is a bare
// "<summary>", and nodeText embeds only firstLine — so if the tags survived,
// every C# member documented this way would be embedded as "<summary>" and the
// feature would do nothing for the language where it matters most.
func TestCSharpDocComments_MultiLineXMLDocDropsTagLines(t *testing.T) {
	src := `public class UserService {
  /// <summary>
  /// Reloads accounts from the remote store.
  /// </summary>
  public string Reload() {
    return repo.Reload();
  }

  /// <summary>Finds every account.</summary>
  /// <param name="id">the account id</param>
  public string FindById(string id) {
    return repo.Find(id);
  }
}
`
	docs := csharpDocComments(src)
	if got := docs["Reload"]; got != "Reloads accounts from the remote store." {
		t.Errorf("docs[Reload] = %q, want the tag lines dropped", got)
	}
	if got := docs["FindById"]; got != "Finds every account.\nthe account id" {
		t.Errorf("docs[FindById] = %q, want tags dropped and the wrapped text kept", got)
	}
}

// TestJavadocStyleDocExtractors_NamesMatchGraphSymbols guards the coupling the
// type system cannot: this package carries private copies of
// internal/graph/symbols.go's symbol-name regexes (see javadoc4docs.go), and
// nodeText looks a doc comment up by the name in graph.Node.Symbols. Break a
// copied regex and the extractors quietly stop recognising names while every
// other test here still passes — a false negative inside an anti-duplication
// check, with nothing reporting it.
//
// Both directions are required. Checking only that every keyed doc name is a
// real graph symbol is not enough: breaking a regex makes an extractor produce
// *fewer* docs, and the survivors are still legitimately keyed, so a subset
// check passes while enrichment has rotted. Each case therefore names the
// symbols its fixture documents and the test requires all of them:
//
//  1. every documented symbol in the fixture gets a doc  (catches drift),
//  2. every keyed doc name is a real graph symbol         (catches mis-keying),
//  3. the undocumented fixture symbol gets no doc         (catches over-reach).
//
// The bare cases carry no comments at all and require an empty doc set, while
// still going through graph.Build so a file with no doc comments is proven to
// yield symbols without enrichment rather than an empty node.
func TestJavadocStyleDocExtractors_NamesMatchGraphSymbols(t *testing.T) {
	cases := []struct {
		lang    string
		rel     string
		content string
		extract func(string) map[string]string
		want    []string
	}{
		{lang: "dart", rel: "lib/user_repo.dart", content: dartDriftFixture, extract: dartDocComments, want: []string{"UserRepo", "loadUser", "reload"}},
		{lang: "java", rel: "svc/user_service.java", content: javaDriftFixture, extract: javaDocComments, want: []string{"UserService", "findAll", "reload"}},
		{lang: "kotlin", rel: "repo/user_repo.kt", content: kotlinDriftFixture, extract: kotlinDocComments, want: []string{"UserRepo", "loadUser", "reload"}},
		{lang: "kotlin script", rel: "repo/tasks.gradle.kts", content: kotlinDriftFixture, extract: kotlinDocComments, want: []string{"UserRepo", "loadUser", "reload"}},
		{lang: "csharp", rel: "app/user_service.cs", content: csharpDriftFixture, extract: csharpDocComments, want: []string{"UserService", "FindAll", "Reload"}},

		{lang: "dart bare", rel: "lib/plain.dart", content: "class Plain {\n  void noop() {\n  }\n}\n", extract: dartDocComments},
		{lang: "java bare", rel: "svc/plain.java", content: "public class Plain {\n  @Override\n  void noop() {\n  }\n}\n", extract: javaDocComments},
		{lang: "kotlin bare", rel: "repo/plain.kt", content: "class Plain {\n  @JvmStatic\n  fun noop() {\n  }\n}\n", extract: kotlinDocComments},
		{lang: "csharp bare", rel: "app/plain.cs", content: "public class Plain {\n  void Noop() {\n  }\n}\n", extract: csharpDocComments},
	}

	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			repoRoot := t.TempDir()
			outDir := t.TempDir()
			path := filepath.Join(repoRoot, filepath.FromSlash(tc.rel))
			writeFile(t, path, tc.content)

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
				if filepath.ToSlash(n.ID) == tc.rel {
					symbols = n.Symbols
					break
				}
			}
			if symbols == nil {
				t.Fatalf("graph.Build produced no node for %s (ids: %v)", tc.rel, nodeIDs(g))
			}
			symbolSet := make(map[string]bool, len(symbols))
			for _, s := range symbols {
				symbolSet[s] = true
			}

			docs := tc.extract(tc.content)

			// (1) every documented symbol must have been recognised.
			for _, want := range tc.want {
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
				if !containsString(tc.want, name) {
					t.Errorf("extractor produced a doc for undocumented symbol %q; the fixture documents only %v",
						name, tc.want)
				}
			}
		})
	}
}

// Each fixture documents a class and two members (one of them behind an
// annotation, which is the common case in Java/C# and must not detach the doc),
// and leaves one annotated member undocumented, so the drift test has a symbol
// it must NOT key a doc for. Every declaration here is deliberately written so
// graph's own cLikeFuncHits accepts it (same-line or Allman "{", never a bare
// "{}" body).
const dartDriftFixture = `/// Stores user records for the app.
class UserRepo {
  /** Loads one user by id. */
  Future<User> loadUser(String id) {
    return store.get(id);
  }

  /// Reloads records from the remote store.
  @override
  void reload() {
    store.reload();
  }

  @override
  void evict(String id) {
    store.remove(id);
  }
}
`

const javaDriftFixture = `/** Authenticates users against the account store. */
public class UserService {
  /** Finds every account. */
  public List<User> findAll() {
    return repo.all();
  }

  /** Reloads accounts from the remote store. */
  @Override
  public void reload() {
    repo.reload();
  }

  @Override
  public void flush() {
    repo.clear();
  }
}
`

const kotlinDriftFixture = `/**
 * Loads a user by id from the local store.
 */
class UserRepo {
  /** Fetches one user record. */
  suspend fun loadUser(id: String): User {
    return store.get(id)
  }

  /** Reloads records from the remote store. */
  @Throws(IOException::class)
  suspend fun reload() {
    store.reload()
  }

  @JvmStatic
  fun evict(id: String) {
    cache.remove(id)
  }
}
`

const csharpDriftFixture = `/// <summary>Authenticates users against the account store.</summary>
public class UserService {
  /// <summary>Finds every account.</summary>
  public List<User> FindAll() {
    return repo.All();
  }

  /// <summary>Reloads accounts from the remote store.</summary>
  [HttpGet("api/accounts/reload")]
  public string Reload() {
    return repo.Reload();
  }

  [Obsolete]
  public void Flush() {
    repo.Clear();
  }
}
`
