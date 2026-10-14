package semantic

import "testing"

func TestGoDocComments_Func(t *testing.T) {
	src := `package pkg

// ValidateUser checks the given credentials against the user store.
// Returns false on any mismatch.
func ValidateUser(user, pass string) bool { return true }

func Undocumented() {}
`
	docs, err := goDocComments(src)
	if err != nil {
		t.Fatalf("goDocComments: %v", err)
	}
	want := "ValidateUser checks the given credentials against the user store.\nReturns false on any mismatch."
	if docs["ValidateUser"] != want {
		t.Errorf("docs[ValidateUser] = %q, want %q", docs["ValidateUser"], want)
	}
	if _, ok := docs["Undocumented"]; ok {
		t.Errorf("Undocumented should have no doc comment, got %q", docs["Undocumented"])
	}
}

func TestGoDocComments_SingleTypeDecl(t *testing.T) {
	src := `package pkg

// User represents an authenticated account.
type User struct {
	Name string
}
`
	docs, err := goDocComments(src)
	if err != nil {
		t.Fatalf("goDocComments: %v", err)
	}
	if docs["User"] == "" {
		t.Error("expected a doc comment for User")
	}
}

func TestGoDocComments_GroupedTypeDecl(t *testing.T) {
	src := `package pkg

type (
	// Account is a billing account.
	Account struct{}

	// Session is a login session.
	Session struct{}
)
`
	docs, err := goDocComments(src)
	if err != nil {
		t.Fatalf("goDocComments: %v", err)
	}
	if docs["Account"] == "" {
		t.Error("expected a doc comment for Account")
	}
	if docs["Session"] == "" {
		t.Error("expected a doc comment for Session")
	}
}

func TestGoDocComments_NoComments_ReturnsNilNoError(t *testing.T) {
	docs, err := goDocComments("package pkg\n\nfunc F() {}\n")
	if err != nil {
		t.Fatalf("goDocComments: %v", err)
	}
	if docs != nil {
		t.Errorf("expected nil docs for a file with no doc comments, got %v", docs)
	}
}

func TestGoDocComments_InvalidSyntax_ReturnsError(t *testing.T) {
	if _, err := goDocComments("this is not valid go {{{"); err == nil {
		t.Fatal("expected a parse error for invalid Go syntax")
	}
}

func TestFirstLine(t *testing.T) {
	cases := map[string]string{
		"one line":             "one line",
		"first\nsecond\nthird": "first",
		"":                     "",
	}
	for in, want := range cases {
		if got := firstLine(in); got != want {
			t.Errorf("firstLine(%q) = %q, want %q", in, got, want)
		}
	}
}
