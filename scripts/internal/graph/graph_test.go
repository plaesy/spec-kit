package graph

import "testing"

func TestToSlash(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"forward already", "a/b/c", "a/b/c"},
		{"backslash", "a\\b\\c", "a/b/c"},
		{"mixed", "a\\b/c", "a/b/c"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toSlash(tt.in); got != tt.want {
				t.Errorf("toSlash(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestExcludeDirSegment(t *testing.T) {
	tests := []struct {
		name string
		seg  string
		want bool
	}{
		{"empty", "", false},
		{"dotdir", ".git", true},
		{"node_modules", "node_modules", true},
		{"dist", "dist", true},
		{"build", "build", true},
		{"vendor", "vendor", true},
		{"pycache", "__pycache__", true},
		{"normal src", "src", false},
		{"dotted file", "main.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := excludeDirSegment(tt.seg); got != tt.want {
				t.Errorf("excludeDirSegment(%q) = %v, want %v", tt.seg, got, tt.want)
			}
		})
	}
}

func TestNodeTypeOf(t *testing.T) {
	tests := []struct {
		name, rel, want string
	}{
		{"agents dir", "agents/foo.md", "agent"},
		{"instructions dir", "instructions/x.md", "instruction"},
		{"checklists dir", "checklists/y.md", "checklist"},
		{"templates dir", "templates/z.md", "template"},
		{"scripts dir", "scripts/a.sh", "script"},
		{"docs dir", "docs/b.md", "doc"},
		{"prompts dir", "prompts/c.md", "prompt"},
		{"testing dir", "testing/d.md", "testing"},
		{"go source", "src/main.go", "source"},
		{"python source", "app.py", "source"},
		{"markdown other", "readme.md", "other"},
		{"text other", "notes.txt", "other"},
		{"root go", "main.go", "source"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nodeTypeOf(tt.rel); got != tt.want {
				t.Errorf("nodeTypeOf(%q) = %q, want %q", tt.rel, got, tt.want)
			}
		})
	}
}

func TestGetArchitectureLayer(t *testing.T) {
	tests := []struct {
		name, rel, want string
	}{
		{"api folder", "api/handler.go", "API"},
		{"controllers folder", "controllers/c.go", "API"},
		{"service file", "business/svc.go", "Service"},
		{"usecase file", "use_cases/uc.py", "Service"},
		{"repo folder", "repositories/r.go", "Data"},
		{"model file", "models/m.go", "Data"},
		{"components folder", "components/widgets/w.go", "UI"},
		{"utils folder", "utils/helpers/h.go", "Utility"},
		{"config file", "config/validator.go", "Utility"},
		{"unmatched", "misc/readme.md", ""},
		{"api in path", "src/api/routes/r.go", "API"},
		{"data in path", "app/data/persistence/x.go", "Data"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getArchitectureLayer(tt.rel); got != tt.want {
				t.Errorf("getArchitectureLayer(%q) = %q, want %q", tt.rel, got, tt.want)
			}
		})
	}
}

func TestExtOf(t *testing.T) {
	tests := []struct {
		name, rel, want string
	}{
		{"go", "a/b.go", ".go"},
		{"md", "readme.md", ".md"},
		{"no ext", "Makefile", ""},
		{"uppercase", "README.MD", ".md"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extOf(tt.rel); got != tt.want {
				t.Errorf("extOf(%q) = %q, want %q", tt.rel, got, tt.want)
			}
		})
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("a", "b"); got != "a" {
		t.Errorf("firstNonEmpty(a,b) = %q, want a", got)
	}
	if got := firstNonEmpty("", "b"); got != "b" {
		t.Errorf("firstNonEmpty(,b) = %q, want b", got)
	}
}

func TestIsAllDots(t *testing.T) {
	if !isAllDots("...") {
		t.Error("isAllDots(...) = false, want true")
	}
	// A bare "." is also all dots, and resolvePyImport rejects it for the same
	// reason it rejects "...": neither names a module.
	if !isAllDots(".") {
		t.Error("isAllDots(.) = false, want true")
	}
	if isAllDots("..a") {
		t.Error("isAllDots(..a) = true, want false")
	}
	if isAllDots("") {
		t.Error("isAllDots() = true, want false")
	}
}

func TestBaseName(t *testing.T) {
	tests := []struct {
		name, rel, want string
	}{
		{"nested", "a/b/c.go", "c.go"},
		{"root", "main.go", "main.go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := baseName(tt.rel); got != tt.want {
				t.Errorf("baseName(%q) = %q, want %q", tt.rel, got, tt.want)
			}
		})
	}
}

func TestNodeStructFields(t *testing.T) {
	n := Node{ID: "x", Label: "x", Type: "source", Degree: 3, Community: 2}
	if n.ID != "x" || n.Degree != 3 || n.Community != 2 {
		t.Fatalf("Node fields not set correctly: %+v", n)
	}
	if n.Symbols != nil {
		t.Errorf("default Symbols should be nil, got %v", n.Symbols)
	}
}

func TestEdgeStructFields(t *testing.T) {
	e := Edge{Source: "a", Target: "b", Type: "imports", Confidence: "EXTRACTED"}
	if e.Source != "a" || e.Target != "b" || e.Type != "imports" || e.Confidence != "EXTRACTED" {
		t.Fatalf("Edge fields not set correctly: %+v", e)
	}
}
