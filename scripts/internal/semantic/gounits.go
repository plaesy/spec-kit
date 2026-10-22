package semantic

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// Unit is one declaration in a source file: a function, a method, or a type.
// Start and End are 1-based line numbers, and End is inclusive. A doc comment
// that sits directly above the declaration is part of the range, so Start is
// its first line. A free-standing comment (separated by a blank line, or not
// attached to any declaration) is not a unit.
type Unit struct {
	Name  string
	Kind  string // "func", "method", or "type"
	Doc   string // first line of the doc comment, "" when there is none
	Start int
	End   int
}

// goUnits lists the declarations in a Go source file with their exact line
// ranges. It uses go/ast, not a regex, because a block end cannot be guessed
// from the text. A file that fails to parse returns an error, and the caller
// falls back to file-level indexing.
func goUnits(content string) ([]Unit, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	line := func(p token.Pos) int { return fset.Position(p).Line }
	var units []Unit
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name == nil {
				continue
			}
			kind := "func"
			if d.Recv != nil {
				kind = "method"
			}
			start := d.Pos()
			if d.Doc != nil {
				start = d.Doc.Pos()
			}
			units = append(units, Unit{
				Name:  d.Name.Name,
				Kind:  kind,
				Doc:   firstLine(strings.TrimSpace(d.Doc.Text())),
				Start: line(start),
				End:   line(d.End()),
			})
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			// A single `type Foo ...` owns the GenDecl, so its doc and its
			// start belong to the declaration. In a grouped `type ( ... )`
			// each spec is its own unit, and only a doc on the spec counts.
			grouped := d.Lparen.IsValid()
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name == nil {
					continue
				}
				doc, start := ts.Doc, ts.Pos()
				if !grouped {
					doc, start = d.Doc, d.Pos()
				}
				if doc != nil {
					start = doc.Pos()
				}
				end := ts.End()
				if !grouped {
					end = d.End()
				}
				units = append(units, Unit{
					Name:  ts.Name.Name,
					Kind:  "type",
					Doc:   firstLine(strings.TrimSpace(doc.Text())),
					Start: line(start),
					End:   line(end),
				})
			}
		}
	}
	return units, nil
}
