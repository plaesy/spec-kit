package graph

import (
	"reflect"
	"testing"
)

func TestSymbolsOfPowerShell(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{name: "no function lines", rel: "x.ps1", body: "Write-Host 'hi'\n", want: nil},
		{
			name: "single function",
			rel:  "x.ps1",
			body: "function Get-Thing {\n}\n",
			want: []string{"Get-Thing"},
		},
		{
			name: "indented function is still matched",
			rel:  "x.ps1",
			body: "  \tfunction  DoIt  {\n}\n",
			want: []string{"DoIt"},
		},
		{
			name: "multiple functions keep line order",
			rel:  "x.ps1",
			body: "function A {\n}\n\nfunction B {\n}\n",
			want: []string{"A", "B"},
		},
		{
			name: "filter keyword is not a function",
			rel:  "x.ps1",
			body: "filter Thing {\n}\n",
			want: nil,
		},
		{name: "empty file", rel: "x.ps1", body: "", want: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := symbolsOf(tc.rel, tc.body); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("symbolsOf() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSymbolsOfSh(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{name: "no symbols", rel: "x.sh", body: "echo hi\n", want: nil},
		{name: "function keyword form", rel: "x.sh", body: "function build {\n}\n", want: []string{"build"}},
		{name: "brace form", rel: "x.sh", body: "build() {\n}\n", want: []string{"build"}},
		{
			name: "both forms on different lines",
			rel:  "x.sh",
			body: "a() {\n}\nfunction b {\n}\n",
			want: []string{"a", "b"},
		},
		{name: "function keyword wins over brace form on same line", rel: "x.sh", body: "function c {\n", want: []string{"c"}},
		{name: "brace form with args is not a symbol", rel: "x.sh", body: "a() {\n", want: []string{"a"}},
		{name: "call site is ignored", rel: "x.sh", body: "build all\n", want: nil},
		{name: "leading whitespace brace form", rel: "x.sh", body: "\tbuild()  {\n", want: []string{"build"}},
		{name: "empty file", rel: "x.sh", body: "", want: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := symbolsOf(tc.rel, tc.body); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("symbolsOf() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSymbolsOfJS(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{name: "no symbols", rel: "x.js", body: "const a = 1\n", want: nil},
		{name: "function declaration", rel: "x.js", body: "function foo(a) {\n}\n", want: []string{"foo"}},
		{name: "exported function", rel: "x.js", body: "export function foo(a) {\n", want: []string{"foo"}},
		{name: "class", rel: "x.js", body: "class Bar {\n", want: []string{"Bar"}},
		{name: "exported class", rel: "x.js", body: "export class Baz {\n", want: []string{"Baz"}},
		{
			name: "func before class in mixed content",
			rel:  "x.ts",
			body: "class A {\n}\nfunction b() {\n}\n",
			want: []string{"A", "b"},
		},
		{name: "arrow function is not matched", rel: "x.ts", body: "const f = () => 1\n", want: nil},
		{name: "ts/jsx/tsx extensions all route to js", rel: "x.jsx", body: "function j() {\n", want: []string{"j"}},
		{name: "empty file", rel: "x.tsx", body: "", want: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := symbolsOf(tc.rel, tc.body); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("symbolsOf() = %v, want %v", got, tc.want)
			}
		})
	}
}
