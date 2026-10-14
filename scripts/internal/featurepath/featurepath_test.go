package featurepath

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/plaesy/spec-kit/internal/common"
)

// The featurepath functions reach git only through common.GetRepoRoot,
// common.GetCurrentBranch and a bare exec.Command("git", "checkout", "-b", …),
// none of which take an injected runner, and creating a throwaway git
// repository is not allowed in these tests. The seam used instead is the one
// the code already depends on: the `git` binary found on PATH.
//
// installGitShim puts a copy of this test binary at the front of PATH under
// the name `git` (git.exe on Windows). TestMain intercepts an invocation
// carrying shimEnv and answers exactly the two read-only queries the package
// makes, plus the branch creation, so the package under test stays untouched
// and the repository is never written to. Any other invocation exits non-zero
// rather than falling through to the test suite.
//
// The shim is fail-closed and every test verifies with a canary that it is
// really the one answering (see installGitShim); a failed canary skips the
// test instead of letting the real git run.

const (
	shimEnv       = "PLAESY_TEST_GIT_SHIM"
	shimRootEnv   = "PLAESY_TEST_GIT_ROOT"
	shimBranchEnv = "PLAESY_TEST_GIT_BRANCH"
	shimFailEnv   = "PLAESY_TEST_GIT_FAIL" // "", "root", "branch", "no-head", "checkout"
	shimLogEnv    = "PLAESY_TEST_GIT_LOG"
)

var (
	shimDir       string // directory holding the fake git executable
	shimErr       error  // why the shim is unusable, if it is
	shimOnce      sync.Once
	shimCanaryErr error
)

// shimProbeRoot is the repository root the canary claims; the real git would
// answer with something else, which is how the canary tells them apart.
const shimProbeRoot = "/plaesy-gitshim-probe-root"

func TestMain(m *testing.M) {
	if os.Getenv(shimEnv) == "1" {
		// Running as the fake git: never fall through to m.Run(), which would
		// recurse into the whole test suite.
		os.Exit(runGitShim())
	}
	shimDir, shimErr = buildGitShim()
	code := m.Run()
	if shimDir != "" {
		_ = os.RemoveAll(shimDir)
	}
	os.Exit(code)
}

// buildGitShim copies the test binary into a temp directory under the name
// `git` so the OS can execute it as the fake git.
func buildGitShim() (dir string, err error) {
	self, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate the test binary: %w", err)
	}
	dir, err = os.MkdirTemp("", "plaesy-gitshim")
	if err != nil {
		return "", fmt.Errorf("create the shim directory: %w", err)
	}
	binary := filepath.Join(dir, "git")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	data, err := os.ReadFile(self)
	if err != nil {
		return dir, fmt.Errorf("read the test binary: %w", err)
	}
	if err := os.WriteFile(binary, data, 0o755); err != nil {
		return dir, fmt.Errorf("write the shim: %w", err)
	}
	return dir, nil
}

// runGitShim answers the git invocations the package under test makes.
func runGitShim() int {
	args := os.Args[1:]
	fail := os.Getenv(shimFailEnv)
	failThis := func(what string) bool { return fail == what }

	switch {
	case len(args) >= 2 && args[0] == "rev-parse" && args[1] == "--show-toplevel":
		if failThis("root") {
			return 1
		}
		fmt.Println(os.Getenv(shimRootEnv))
		return 0
	case len(args) >= 2 && args[0] == "rev-parse" && args[1] == "--abbrev-ref":
		// "branch" means the branch cannot be determined at all: both the
		// primary query and the fallback fail. "no-head" is the unborn-branch
		// case — the primary query has no revision to resolve, but
		// `git branch --show-current` still answers.
		if failThis("branch") || failThis("no-head") {
			return 1
		}
		fmt.Println(os.Getenv(shimBranchEnv))
		return 0
	case len(args) >= 2 && args[0] == "branch" && args[1] == "--show-current":
		if failThis("branch") {
			return 1
		}
		fmt.Println(os.Getenv(shimBranchEnv))
		return 0
	case len(args) >= 3 && args[0] == "checkout" && args[1] == "-b":
		if failThis("checkout") {
			return 1
		}
		if log := os.Getenv(shimLogEnv); log != "" {
			f, err := os.OpenFile(log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				return 1
			}
			defer f.Close()
			if _, err := fmt.Fprintf(f, "checkout %s\n", args[2]); err != nil {
				return 1
			}
		}
		return 0
	default:
		// Fail closed: an unexpected command must not reach real git, and it
		// must not run the test suite either.
		fmt.Fprintf(os.Stderr, "git shim: unsupported invocation %q\n", args)
		return 1
	}
}

// installGitShim points PATH at the fake git, reporting root and branch, and
// returns the file the shim logs branch creations to. It skips the test when
// the shim cannot be made effective — verified by a canary — so the real git
// is never asked to write anything. The canary spawns a copy of the test
// binary, so it runs once per package: the shim directory is fixed, and
// PATH is only ever prefixed with it.
func installGitShim(t *testing.T, root, branch string) string {
	t.Helper()
	if shimErr != nil {
		t.Skipf("git shim unavailable: %v", shimErr)
	}
	log := filepath.Join(t.TempDir(), "git.log")
	t.Setenv(shimRootEnv, root)
	t.Setenv(shimBranchEnv, branch)
	t.Setenv(shimFailEnv, "")
	t.Setenv(shimLogEnv, log)
	t.Setenv(shimEnv, "1")
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	shimOnce.Do(func() {
		prev, hadPrev := os.LookupEnv(shimRootEnv)
		_ = os.Setenv(shimRootEnv, shimProbeRoot)
		out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if hadPrev {
			_ = os.Setenv(shimRootEnv, prev)
		} else {
			_ = os.Unsetenv(shimRootEnv)
		}
		if err != nil || strings.TrimSpace(string(out)) != shimProbeRoot {
			shimCanaryErr = fmt.Errorf("canary returned %q, %v", strings.TrimSpace(string(out)), err)
		}
	})
	if shimCanaryErr != nil {
		t.Skipf("git shim is not effective; skipping: %v", shimCanaryErr)
	}
	return log
}

// gitBranchLog returns the branch names the shim was asked to create.
func gitBranchLog(t *testing.T, log string) []string {
	t.Helper()
	data, err := os.ReadFile(log)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read shim log: %v", err)
	}
	var out []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// chdir switches the process working directory for the duration of the test
// (Go 1.20 has no t.Chdir). Callers must not use t.Parallel.
func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Errorf("restore cwd to %s: %v", prev, err)
		}
	})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestCreateNewFeatureRequiresADescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{name: "empty", description: ""},
		{name: "whitespace only", description: "   \t\n "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateNewFeature(tt.description)
			if err == nil {
				t.Fatalf("expected an error, got %+v", got)
			}
			if err.Error() != "feature description is required" {
				t.Errorf("error = %q", err)
			}
			if got != nil {
				t.Error("result must be nil on error")
			}
		})
	}
}

// defaultSpecTemplate is the body seeded for the cases that are about branch
// naming rather than template resolution.
const defaultSpecTemplate = "# Feature: template\n"

func TestCreateNewFeature(t *testing.T) {
	tests := []struct {
		name string
		// specs are pre-created under <root>/.plaesy/specs.
		specs      []string
		files      []string // pre-created non-directory entries in the specs dir
		template   string   // spec.template.md content; "" = use defaultSpecTemplate
		templateAt string   // "source" (default) | "installed" | "both"
		noTemplate bool     // skip seeding: CreateNewFeature must then fail
		desc       string
		wantBranch string
	}{
		{
			name: "first feature in an empty tree", desc: "Add user login",
			wantBranch: "001-add-user-login",
		},
		{
			name: "description is lowercased", desc: "Add User Login",
			wantBranch: "001-add-user-login",
		},
		{
			name: "only three words are kept", desc: "add user login with oauth sso",
			wantBranch: "001-add-user-login",
		},
		{
			name: "runs of non alphanumerics collapse to one dash", desc: "Fix   the__bug--now",
			wantBranch: "001-fix-the-bug",
		},
		{
			name: "leading and trailing punctuation is trimmed", desc: "!!! ship it !!!",
			wantBranch: "001-ship-it",
		},
		{
			name: "surrounding whitespace is trimmed", desc: "   add login   ",
			wantBranch: "001-add-login",
		},
		{
			name: "numbers inside the description are kept", desc: "Support OAuth2 logins",
			wantBranch: "001-support-oauth2-logins",
		},
		{
			// Documents the slug's ASCII-only rule: accented letters are not
			// transliterated, they are dropped, which can leave a single
			// letter standing in for a word.
			name: "non ascii letters are dropped", desc: "Añadir café",
			wantBranch: "001-a-adir-caf",
		},
		{
			// Degenerate input: nothing survives slugging, so the branch is
			// just the number and a separator. It still satisfies
			// common.CheckFeatureBranch's ^[0-9]{3}- pattern.
			name: "description with no usable characters", desc: "***",
			wantBranch: "001-",
		},
		{
			name:       "numbering continues past the highest existing feature",
			specs:      []string{"001-alpha", "007-gamma", "notanumber", "12-numeric"},
			wantBranch: "013-add-login",
			desc:       "add login",
		},
		{
			name:  "non directory entries in the specs dir are ignored",
			files: []string{"042-file.md", "notanumber"},
			desc:  "add login", wantBranch: "001-add-login",
		},
		{
			name:     "template content is copied into spec.md",
			template: "# Feature: template\n", desc: "add login",
			wantBranch: "001-add-login",
		},
		{
			// The installed location, as produced by `plaesy init`. This is
			// the case that regressed: the template used to be read from
			// <root>/templates only, so every installed project silently
			// produced a zero-byte spec.md.
			name:       "template is read from the installed .plaesy location",
			templateAt: "installed", template: "# Installed\n",
			desc: "add login", wantBranch: "001-add-login",
		},
		{
			// Both locations present: the installed copy wins, matching the
			// order in specTemplateCandidates.
			name:       "installed template takes precedence over the source-tree path",
			templateAt: "both", template: "# Installed\n",
			desc: "add login", wantBranch: "001-add-login",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			log := installGitShim(t, root, "main")
			for _, s := range tt.specs {
				if err := os.MkdirAll(filepath.Join(root, ".plaesy", "specs", s), 0o755); err != nil {
					t.Fatalf("seed spec dir: %v", err)
				}
			}
			for _, f := range tt.files {
				writeFile(t, filepath.Join(root, ".plaesy", "specs", f), "x")
			}
			// Every case except the ones that opt out seeds a template, so
			// the slug/branch cases exercise what they are about instead of
			// failing early on a missing template.
			if !tt.noTemplate {
				body := tt.template
				if body == "" {
					body = defaultSpecTemplate
				}
				at := tt.templateAt
				if at == "" {
					at = "source"
				}
				if at == "source" || at == "both" {
					writeFile(t, filepath.Join(root, "templates", "spec.template.md"), body)
				}
				if at == "installed" || at == "both" {
					writeFile(t, filepath.Join(root, ".plaesy", "templates", "spec.template.md"), body)
				}
				tt.template = body
			}

			res, err := CreateNewFeature(tt.desc)
			if err != nil {
				t.Fatalf("CreateNewFeature: %v", err)
			}
			if res.BranchName != tt.wantBranch {
				t.Errorf("branch = %q, want %q", res.BranchName, tt.wantBranch)
			}
			wantNum := res.BranchName[:3]
			if res.FeatureNum != wantNum {
				t.Errorf("feature number = %q, want the branch prefix %q", res.FeatureNum, wantNum)
			}
			wantSpec := filepath.Join(root, ".plaesy", "specs", tt.wantBranch, "spec.md")
			if res.SpecFile != wantSpec {
				t.Errorf("spec file = %q, want %q", res.SpecFile, wantSpec)
			}

			// The branch name reaches git, and only one branch is created.
			if got := gitBranchLog(t, log); len(got) != 1 || got[0] != "checkout "+tt.wantBranch {
				t.Errorf("git was asked for %v, want exactly [checkout %s]", got, tt.wantBranch)
			}

			// The feature directory and spec.md are created, carrying the
			// template's content.
			got, err := os.ReadFile(wantSpec)
			if err != nil {
				t.Fatalf("spec.md not created: %v", err)
			}
			if string(got) != tt.template {
				t.Errorf("spec.md content = %q, want %q", got, tt.template)
			}
		})
	}
}

// TestCreateNewFeatureFailsWithoutATemplate is the regression test for the
// silent-empty-spec bug: when no spec template is installed, CreateNewFeature
// must return an error naming the file it looked for. It must never leave a
// zero-byte spec.md behind, which previously read as a complete artifact.
func TestCreateNewFeatureFailsWithoutATemplate(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "main")

	res, err := CreateNewFeature("add login")
	if err == nil {
		t.Fatalf("expected an error when the spec template is missing, got %+v", res)
	}
	if !strings.Contains(err.Error(), "spec.template.md not found") {
		t.Errorf("error should name the missing template, got %q", err)
	}
	// The error must list both candidate paths so the user can see that
	// `plaesy init` is the fix.
	for _, want := range []string{
		filepath.Join(".plaesy", "templates"),
		filepath.Join("templates", "spec.template.md"),
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %q, got %q", want, err)
		}
	}
	if res != nil {
		t.Error("result must be nil on error")
	}
	if _, err := os.Stat(filepath.Join(root, ".plaesy", "specs", "001-add-login", "spec.md")); err == nil {
		t.Error("an empty spec.md must not be written when the template is missing")
	}
}

// TestCreateNewFeatureReportsGitFailure covers the error wrapping when the
// branch cannot be created: nothing is written under .plaesy/specs.
func TestCreateNewFeatureReportsGitFailure(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "main")
	t.Setenv(shimFailEnv, "checkout")

	res, err := CreateNewFeature("add login")
	if err == nil {
		t.Fatalf("expected an error, got %+v", res)
	}
	if !strings.Contains(err.Error(), "git checkout -b 001-add-login") {
		t.Errorf("error should name the branch command, got %q", err)
	}
	if res != nil {
		t.Error("result must be nil on error")
	}
	if _, err := os.Stat(filepath.Join(root, ".plaesy", "specs", "001-add-login")); err == nil {
		t.Error("the feature dir must not be created when the branch creation failed")
	}
}

// TestCreateNewFeatureWithoutARepository covers the early exit: the specs
// directory is never touched when the repository root cannot be determined.
func TestCreateNewFeatureWithoutARepository(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "main")
	t.Setenv(shimFailEnv, "root")

	res, err := CreateNewFeature("add login")
	if err == nil {
		t.Fatalf("expected an error, got %+v", res)
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("error = %q", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".plaesy")); err == nil {
		t.Error("nothing should be created when the repository root is unknown")
	}
}

// TestCreateNewFeatureSpecsPathBlocked covers a tree where .plaesy/specs is
// occupied by a file: the directory cannot be created and the failure is
// reported before anything is written.
func TestCreateNewFeatureSpecsPathBlocked(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "main")
	writeFile(t, filepath.Join(root, ".plaesy", "specs"), "not a directory\n")

	res, err := CreateNewFeature("add login")
	if err == nil {
		t.Fatalf("expected an error, got %+v", res)
	}
	if !strings.Contains(err.Error(), "creating specs dir") {
		t.Errorf("error = %q, want the specs-dir failure", err)
	}
}

// TestCreateNewFeatureFeatureDirBlocked covers the same class of failure one
// step later, when the numbered branch path is occupied by a file.
func TestCreateNewFeatureFeatureDirBlocked(t *testing.T) {
	root := t.TempDir()
	installGitShim(t, root, "main")
	// Not a directory, so the numbering scan skips it as well.
	writeFile(t, filepath.Join(root, ".plaesy", "specs", "001-add-login"), "blocked\n")

	res, err := CreateNewFeature("add login")
	if err == nil {
		t.Fatalf("expected an error, got %+v", res)
	}
	if !strings.Contains(err.Error(), "creating feature dir") {
		t.Errorf("error = %q, want the feature-dir failure", err)
	}
}

func TestGetPathsReport(t *testing.T) {
	t.Run("on a feature branch", func(t *testing.T) {
		root := t.TempDir()
		installGitShim(t, root, "001-add-login")

		got := GetPathsReport()
		featureDir := filepath.Join(root, ".plaesy", "specs", "001-add-login")
		want := strings.Join([]string{
			"REPO_ROOT: " + root,
			"BRANCH: 001-add-login",
			"FEATURE_DIR: " + featureDir,
			"FEATURE_SPEC: " + filepath.Join(featureDir, "spec.md"),
			"IMPL_PLAN: " + filepath.Join(featureDir, "plan.md"),
			"TASKS: " + filepath.Join(featureDir, "tasks.md"),
			"",
		}, "\n")
		if got != want {
			t.Errorf("GetPathsReport =\n%s\nwant\n%s", got, want)
		}
	})

	t.Run("not on a feature branch", func(t *testing.T) {
		root := t.TempDir()
		installGitShim(t, root, "main")

		got := GetPathsReport()
		if !strings.Contains(got, "BRANCH: main\n") {
			t.Errorf("report should still name the current branch:\n%s", got)
		}
		if !strings.HasSuffix(got, "INFO: Not on a feature branch (format: XXX-feature-name)\n") {
			t.Errorf("report should end with the feature-branch notice:\n%s", got)
		}
		// The paths are still reported, relative to the current branch.
		if !strings.Contains(got, "FEATURE_DIR: "+filepath.Join(root, ".plaesy", "specs", "main")) {
			t.Errorf("report should still list the derived paths:\n%s", got)
		}
	})

	t.Run("three digit branch is a feature branch", func(t *testing.T) {
		root := t.TempDir()
		installGitShim(t, root, "999-a")
		if strings.Contains(GetPathsReport(), "Not on a feature branch") {
			t.Error("999-a satisfies the NNN- pattern and must not be reported as off-branch")
		}
	})

	t.Run("branch without the numeric prefix is not a feature branch", func(t *testing.T) {
		root := t.TempDir()
		installGitShim(t, root, "feature/001-add-login")
		if !strings.Contains(GetPathsReport(), "Not on a feature branch") {
			t.Error("feature/001-add-login does not start with NNN- and must be reported as off-branch")
		}
	})

	fallbackTests := []struct {
		name string
		fail string
	}{
		{name: "no git repository at all", fail: "root"},
		{name: "branch query fails", fail: "branch"},
	}
	for _, tt := range fallbackTests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			cwd := t.TempDir()
			installGitShim(t, root, "main")
			t.Setenv(shimFailEnv, tt.fail)
			chdir(t, cwd)

			// The fallback's REPO_ROOT comes from a real os.Getwd(), not from
			// the pre-chdir `cwd` string: on macOS, a t.TempDir() under
			// /var/folders/... is reached through /var, a symlink to
			// /private/var, and Getwd (unlike a bare string comparison)
			// returns the kernel's resolved, symlink-free view —
			// "/private/var/folders/...". Re-deriving "want" from Getwd after
			// the chdir is what makes this assertion OS-independent.
			resolvedCwd, err := os.Getwd()
			if err != nil {
				t.Fatalf("getwd after chdir: %v", err)
			}

			got := GetPathsReport()
			// The fallback reports the working directory, which is the only
			// location it can still name.
			want := strings.Join([]string{
				"REPO_ROOT: " + resolvedCwd,
				"BRANCH: unknown",
				"FEATURE_DIR: Not available",
				"FEATURE_SPEC: Not available",
				"IMPL_PLAN: Not available",
				"TASKS: Not available",
				"INFO: Unable to determine feature paths",
				"",
			}, "\n")
			if got != want {
				t.Errorf("GetPathsReport =\n%s\nwant\n%s", got, want)
			}
		})
	}
}

// seedFeature lays out a feature directory under root and returns the paths
// CheckTaskPrerequisites works from.
func seedFeature(t *testing.T, root, branch string, files ...string) string {
	t.Helper()
	featureDir := filepath.Join(root, ".plaesy", "specs", branch)
	if err := os.MkdirAll(featureDir, 0o755); err != nil {
		t.Fatalf("mkdir feature dir: %v", err)
	}
	for _, f := range files {
		writeFile(t, filepath.Join(featureDir, f), "content\n")
	}
	return featureDir
}

func TestCheckTaskPrerequisites(t *testing.T) {
	t.Run("not a git repository", func(t *testing.T) {
		root := t.TempDir()
		installGitShim(t, root, "main")
		t.Setenv(shimFailEnv, "root")

		res, err := CheckTaskPrerequisites()
		if err == nil {
			t.Fatalf("expected an error, got %+v", res)
		}
		if !strings.Contains(err.Error(), "not a git repository") {
			t.Errorf("error = %q", err)
		}
	})

	t.Run("not on a feature branch", func(t *testing.T) {
		root := t.TempDir()
		installGitShim(t, root, "main")

		res, err := CheckTaskPrerequisites()
		if err == nil {
			t.Fatalf("expected an error, got %+v", res)
		}
		if !strings.Contains(err.Error(), "not on a feature branch (current: main)") {
			t.Errorf("error = %q", err)
		}
	})

	t.Run("feature directory missing", func(t *testing.T) {
		root := t.TempDir()
		installGitShim(t, root, "001-a")

		res, err := CheckTaskPrerequisites()
		if err == nil {
			t.Fatalf("expected an error, got %+v", res)
		}
		featureDir := filepath.Join(root, ".plaesy", "specs", "001-a")
		if !strings.Contains(err.Error(), "feature directory not found: "+featureDir) ||
			!strings.Contains(err.Error(), "run /start first") {
			t.Errorf("error = %q, want the feature-dir message with the /start hint", err)
		}
	})

	t.Run("feature path is a file not a directory", func(t *testing.T) {
		root := t.TempDir()
		installGitShim(t, root, "001-a")
		writeFile(t, filepath.Join(root, ".plaesy", "specs", "001-a"), "not a dir\n")

		if _, err := CheckTaskPrerequisites(); err == nil ||
			!strings.Contains(err.Error(), "feature directory not found") {
			t.Errorf("error = %v, want the feature-dir message", err)
		}
	})

	t.Run("plan.md missing", func(t *testing.T) {
		root := t.TempDir()
		installGitShim(t, root, "001-a")
		featureDir := seedFeature(t, root, "001-a", "spec.md")

		res, err := CheckTaskPrerequisites()
		if err == nil {
			t.Fatalf("expected an error, got %+v", res)
		}
		if !strings.Contains(err.Error(), "plan.md not found in "+featureDir) ||
			!strings.Contains(err.Error(), "templates/plan.template.md") {
			t.Errorf("error = %q, want the plan.md message with the template hint", err)
		}
	})

	t.Run("plan.md is a directory", func(t *testing.T) {
		root := t.TempDir()
		installGitShim(t, root, "001-a")
		seedFeature(t, root, "001-a")
		if err := os.MkdirAll(filepath.Join(root, ".plaesy", "specs", "001-a", "plan.md"), 0o755); err != nil {
			t.Fatalf("mkdir plan.md: %v", err)
		}

		// os.Stat only checks existence, so a directory passes this gate.
		if _, err := CheckTaskPrerequisites(); err != nil &&
			!strings.Contains(err.Error(), "plan.md not found") {
			t.Errorf("error = %v, want either success or the plan.md message", err)
		}
	})
}

func TestCheckTaskPrerequisitesAvailableDocs(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		// dirs are created empty under the feature dir.
		dirs     []string
		wantDocs []string
	}{
		{
			name: "plan only", files: []string{"plan.md"},
			wantDocs: nil,
		},
		{
			name: "research only", files: []string{"plan.md", "research.md"},
			wantDocs: []string{"research.md"},
		},
		{
			name: "data model only", files: []string{"plan.md", "data-model.md"},
			wantDocs: []string{"data-model.md"},
		},
		{
			name: "quickstart only", files: []string{"plan.md", "quickstart.md"},
			wantDocs: []string{"quickstart.md"},
		},
		{
			name: "empty contracts dir counts as absent", files: []string{"plan.md"},
			dirs:     []string{"contracts"},
			wantDocs: nil,
		},
		{
			name: "populated contracts dir is listed", files: []string{"plan.md"},
			dirs:     []string{"contracts"},
			wantDocs: []string{"contracts/"},
		},
		{
			name:  "all docs, in the report's order",
			files: []string{"plan.md", "quickstart.md", "data-model.md", "research.md", "spec.md", "tasks.md"},
			dirs:  []string{"contracts"},
			// contracts/ is empty here, so it is the one entry left out.
			wantDocs: []string{"research.md", "data-model.md", "quickstart.md"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			installGitShim(t, root, "001-a")
			featureDir := seedFeature(t, root, "001-a", tt.files...)
			for _, d := range tt.dirs {
				if err := os.MkdirAll(filepath.Join(featureDir, d), 0o755); err != nil {
					t.Fatalf("mkdir %s: %v", d, err)
				}
			}
			// Populate a listed directory so the expectation holds for the
			// "populated contracts dir" case.
			for _, d := range tt.dirs {
				if tt.wantDocs != nil && contains(tt.wantDocs, d+"/") {
					writeFile(t, filepath.Join(featureDir, d, "api.md"), "x\n")
				}
			}

			res, err := CheckTaskPrerequisites()
			if err != nil {
				t.Fatalf("CheckTaskPrerequisites: %v", err)
			}
			if res.FeatureDir != featureDir {
				t.Errorf("feature dir = %q, want %q", res.FeatureDir, featureDir)
			}
			if len(res.AvailableDocs) != len(tt.wantDocs) {
				t.Fatalf("available docs = %v, want %v", res.AvailableDocs, tt.wantDocs)
			}
			for i := range tt.wantDocs {
				if res.AvailableDocs[i] != tt.wantDocs[i] {
					t.Errorf("available docs = %v, want %v", res.AvailableDocs, tt.wantDocs)
					break
				}
			}
		})
	}
}

func TestPrereqResultTextReport(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		dirs  []string
		want  string
	}{
		{
			name: "nothing present",
			want: strings.Join([]string{
				"FEATURE_DIR:/repo/.plaesy/specs/001-a",
				"AVAILABLE_DOCS:",
				"  ✗ research.md",
				"  ✗ data-model.md",
				"  ✗ contracts/",
				"  ✗ quickstart.md",
				"",
			}, "\n"),
		},
		{
			name:  "files present, contracts dir empty",
			files: []string{"research.md", "quickstart.md"},
			dirs:  []string{"contracts"},
			want: strings.Join([]string{
				"FEATURE_DIR:/repo/.plaesy/specs/001-a",
				"AVAILABLE_DOCS:",
				"  ✓ research.md",
				"  ✗ data-model.md",
				"  ✗ contracts/",
				"  ✓ quickstart.md",
				"",
			}, "\n"),
		},
		{
			name:  "all present",
			files: []string{"research.md", "data-model.md", "quickstart.md"},
			dirs:  []string{"contracts"},
			want: strings.Join([]string{
				"FEATURE_DIR:/repo/.plaesy/specs/001-a",
				"AVAILABLE_DOCS:",
				"  ✓ research.md",
				"  ✓ data-model.md",
				"  ✓ contracts/",
				"  ✓ quickstart.md",
				"",
			}, "\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			fp := &common.FeaturePaths{
				RepoRoot:      "/repo",
				CurrentBranch: "001-a",
				FeatureDir:    "/repo/.plaesy/specs/001-a",
				Research:      filepath.Join(dir, "research.md"),
				DataModel:     filepath.Join(dir, "data-model.md"),
				ContractsDir:  filepath.Join(dir, "contracts"),
				Quickstart:    filepath.Join(dir, "quickstart.md"),
			}
			for _, f := range tt.files {
				var target string
				switch f {
				case "research.md":
					target = fp.Research
				case "data-model.md":
					target = fp.DataModel
				case "quickstart.md":
					target = fp.Quickstart
				}
				writeFile(t, target, "content\n")
			}
			for _, d := range tt.dirs {
				p := filepath.Join(dir, d)
				if err := os.MkdirAll(p, 0o755); err != nil {
					t.Fatalf("mkdir %s: %v", d, err)
				}
			}
			// The contracts dir is reported as present only when non-empty.
			for _, w := range strings.Split(tt.want, "\n") {
				if strings.Contains(w, "✓ contracts/") {
					writeFile(t, filepath.Join(fp.ContractsDir, "api.md"), "x\n")
				}
			}

			r := &PrereqResult{FeatureDir: fp.FeatureDir}
			if got := r.TextReport(fp); got != tt.want {
				t.Errorf("TextReport =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
