package graph

import (
	"reflect"
	"testing"
)

func TestSymbolsOfPython(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{name: "no symbols", rel: "x.py", body: "x = 1\n", want: nil},
		{name: "def", rel: "x.py", body: "def foo(a):\n    pass\n", want: []string{"foo"}},
		{name: "indented def", rel: "x.py", body: "    def foo(self):\n", want: []string{"foo"}},
		{name: "class", rel: "x.py", body: "class Bar:\n", want: []string{"Bar"}},
		{
			name: "def and class mixed",
			rel:  "x.py",
			body: "class A:\n    def m(self):\n",
			want: []string{"A", "m"},
		},
		{name: "decorated def body line is not re-matched", rel: "x.py", body: "x = 1  # def fake():\n", want: nil},
		{name: "empty file", rel: "x.py", body: "", want: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := symbolsOf(tc.rel, tc.body); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("symbolsOf() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSymbolsOfGo(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{name: "no symbols", rel: "x.go", body: "package main\n", want: nil},
		{name: "plain func", rel: "x.go", body: "func foo(a int) {\n}\n", want: []string{"foo"}},
		{name: "method drops receiver", rel: "x.go", body: "func (r *T) Bar(x int) {\n}\n", want: []string{"Bar"}},
		{name: "value receiver", rel: "x.go", body: "func (t T) Baz() {\n}\n", want: []string{"Baz"}},
		{
			name: "funcs in order",
			rel:  "x.go",
			body: "func a() {\n}\nfunc (s *S) b() {\n}\n",
			want: []string{"a", "b"},
		},
		{name: "func type declaration is not a symbol", rel: "x.go", body: "type F func()\n", want: nil},
		{name: "indented func is not matched", rel: "x.go", body: "  func foo() {\n", want: nil},
		{name: "anonymous func call is not matched", rel: "x.go", body: "go func() {\n}()\n", want: nil},
		{name: "empty file", rel: "x.go", body: "", want: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := symbolsOf(tc.rel, tc.body); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("symbolsOf() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSymbolsOfMarkdown(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{name: "no headings", rel: "x.md", body: "just text\n", want: nil},
		{name: "h1 h2 h3", rel: "x.md", body: "# One\n## Two\n### Three\n", want: []string{"One", "Two", "Three"}},
		{name: "h4 is ignored", rel: "x.md", body: "#### Four\n", want: nil},
		{name: "hash without space ignored", rel: "x.md", body: "#NoSpace\n", want: nil},
		{name: "closing hashes kept", rel: "x.md", body: "## Title ##\n", want: []string{"Title ##"}},
		{name: "hash inside text not a heading", rel: "x.md", body: "a # b\n", want: nil},
		{name: "empty heading body ignored", rel: "x.md", body: "#\n", want: nil},
		{name: "empty file", rel: "x.md", body: "", want: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := symbolsOf(tc.rel, tc.body); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("symbolsOf() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStripMDHeading(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"# One", "One"},
		{"### Deep", "Deep"},
		{"#\tTabbed", "Tabbed"},
		{"#   Padded   ", "Padded"},
		{"#", ""},
		{"#NoSpace", "NoSpace"},
		{"plain", "plain"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := stripMDHeading(tc.in); got != tc.want {
				t.Fatalf("stripMDHeading(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSymbolsOfUnsupportedExtension(t *testing.T) {
	for _, rel := range []string{"x.txt", "x.scala", "x.lua", "Makefile", "x", "x.MD.txt"} {
		t.Run(rel, func(t *testing.T) {
			if got := symbolsOf(rel, "# Heading\nfunction foo {\n"); got != nil {
				t.Fatalf("symbolsOf(%q) = %v, want nil", rel, got)
			}
		})
	}
}
