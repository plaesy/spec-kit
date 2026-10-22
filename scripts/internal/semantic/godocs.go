package semantic

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// goDocComments maps each top-level function/method/type name declared in a
// Go source file to its doc comment (the "// Foo does X" block immediately
// above the declaration), if any. Uses go/parser instead of a regex: Go doc
// comments have a precise, well-specified grammar (a contiguous comment
// block directly preceding the declaration, no blank line in between) that
// the standard library already parses correctly, so there's no reason to
// reimplement it with a heuristic the way symbols.go does for the 14 other
// languages it supports.
//
// Returns (nil, err) only for a file that fails to parse (e.g. invalid Go
// syntax, or the path doesn't exist) — callers treat that as "no docs for
// this node" rather than a fatal error, since embedding is best-effort
// signal, not a correctness-critical path.
func goDocComments(content string) (map[string]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	docs := make(map[string]string)
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Doc != nil && d.Name != nil {
				if text := strings.TrimSpace(d.Doc.Text()); text != "" {
					docs[d.Name.Name] = text
				}
			}
		case *ast.GenDecl:
			// Doc comment can be on the GenDecl itself (single `type Foo
			// struct{...}` form) or on the individual Spec (grouped `type (
			// Foo struct{...} )` form) — check both.
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name == nil {
					continue
				}
				doc := ts.Doc
				if doc == nil {
					doc = d.Doc
				}
				if doc != nil {
					if text := strings.TrimSpace(doc.Text()); text != "" {
						docs[ts.Name.Name] = text
					}
				}
			}
		}
	}
	if len(docs) == 0 {
		return nil, nil
	}
	return docs, nil
}

// firstLine returns just the first line of a (possibly multi-line) doc
// comment: enough semantic signal for embedding without dragging an entire
// paragraph (and its unrelated detail) into the vector.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
