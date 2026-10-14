package graph

import (
	"reflect"
	"testing"
)

func TestSymbolsOfRust(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{
			name: "plain fn",
			rel:  "x.rs",
			body: "fn add(a: i32, b: i32) -> i32 {\n}\n",
			want: []string{"add"},
		},
		{
			name: "pub async fn and struct/enum/trait",
			rel:  "x.rs",
			body: "pub async fn fetch() {}\nstruct User {}\nenum Status {}\ntrait Repo {}\n",
			want: []string{"fetch", "User", "Status", "Repo"},
		},
		{name: "no match", rel: "x.rs", body: "let x = 1;\n", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := symbolsOf(tt.rel, tt.body)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("symbolsOf(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

func TestSymbolsOfDart(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{
			name: "class and typed method",
			rel:  "x.dart",
			body: "class UserRepo {\n  Future<User> loadUser(String id) {\n  }\n}\n",
			want: []string{"UserRepo", "loadUser"},
		},
		{
			name: "control flow is not a method",
			rel:  "x.dart",
			body: "void run() {\n  if (ready) {\n  }\n  for (var i = 0; i < 10; i++) {\n  }\n}\n",
			want: []string{"run"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := symbolsOf(tt.rel, tt.body)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("symbolsOf(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

func TestSymbolsOfCLikeAllmanBrace(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{
			name: "java allman-style brace on next line",
			rel:  "x.java",
			body: "public class UserService\n{\n  public List<User> findAll()\n  {\n  }\n}\n",
			want: []string{"UserService", "findAll"},
		},
		{
			name: "csharp allman-style brace on next line",
			rel:  "x.cs",
			body: "public class Widget\n{\n  public void Render()\n  {\n  }\n}\n",
			want: []string{"Widget", "Render"},
		},
		{
			name: "interface prototype ending in semicolon is not a function",
			rel:  "x.h",
			body: "class Shape {\n  int area();\n};\n",
			want: []string{"Shape"},
		},
		{
			name: "blank line between signature and brace is still matched",
			rel:  "x.cpp",
			body: "int add(int a, int b)\n\n{\n}\n",
			want: []string{"add"},
		},
		{
			name: "order preserved: function before class in mixed content",
			rel:  "x.java",
			body: "int standalone()\n{\n}\nclass Later {\n}\n",
			want: []string{"standalone", "Later"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := symbolsOf(tt.rel, tt.body)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("symbolsOf(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

func TestSymbolsOfJava(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{
			name: "class and method",
			rel:  "x.java",
			body: "public class UserService {\n  public List<User> findAll() {\n  }\n}\n",
			want: []string{"UserService", "findAll"},
		},
		{
			name: "control flow is not a method",
			rel:  "x.java",
			body: "void run() {\n  while (true) {\n  }\n}\n",
			want: []string{"run"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := symbolsOf(tt.rel, tt.body)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("symbolsOf(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

func TestSymbolsOfKotlin(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{
			name: "fun and class",
			rel:  "x.kt",
			body: "class UserRepo {\n  suspend fun loadUser(id: String): User {\n  }\n}\n",
			want: []string{"UserRepo", "loadUser"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := symbolsOf(tt.rel, tt.body)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("symbolsOf(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

func TestSymbolsOfCSharp(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{
			name: "class and method",
			rel:  "x.cs",
			body: "public class UserService {\n  public List<User> FindAll() {\n  }\n}\n",
			want: []string{"UserService", "FindAll"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := symbolsOf(tt.rel, tt.body)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("symbolsOf(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

func TestSymbolsOfCFamily(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{
			name: "struct and function",
			rel:  "x.c",
			body: "struct Point {\n};\nint add(int a, int b) {\n}\n",
			want: []string{"Point", "add"},
		},
		{
			name: "cpp class and method",
			rel:  "x.cpp",
			body: "class Widget {\n  void render() {\n  }\n};\n",
			want: []string{"Widget", "render"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := symbolsOf(tt.rel, tt.body)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("symbolsOf(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

func TestSymbolsOfSwift(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{
			name: "func and class",
			rel:  "x.swift",
			body: "class UserRepo {\n  func loadUser(id: String) -> User {\n  }\n}\n",
			want: []string{"UserRepo", "loadUser"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := symbolsOf(tt.rel, tt.body)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("symbolsOf(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

func TestSymbolsOfRuby(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{
			name: "def and class",
			rel:  "x.rb",
			body: "class UserRepo\n  def load_user(id)\n  end\nend\n",
			want: []string{"UserRepo", "load_user"},
		},
		{
			name: "self dot method and predicate name",
			rel:  "x.rb",
			body: "def self.find(id)\nend\ndef valid?\nend\n",
			want: []string{"self.find", "valid?"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := symbolsOf(tt.rel, tt.body)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("symbolsOf(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

func TestSymbolsOfPHP(t *testing.T) {
	tests := []struct {
		name string
		rel  string
		body string
		want []string
	}{
		{
			name: "class and function",
			rel:  "x.php",
			body: "class UserRepo {\n  public function loadUser($id) {\n  }\n}\n",
			want: []string{"UserRepo", "loadUser"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := symbolsOf(tt.rel, tt.body)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("symbolsOf(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}
