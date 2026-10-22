package cleaner

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/plaesy/spec-kit/internal/config"
)

// fixtureConfig is a small platform.json whose declaration order
// (zeta, alpha, beta, nullmap, generic_ai) deliberately differs from sorted
// order, so detection order is observable.
const fixtureConfig = `{
  "version": "1.0",
  "description": "cleaner fixture",
  "plaesy": {
    "base_directory": ".plaesy",
    "core_directories": ["memory", "instructions"],
    "project_directories": [".plaesy/specs"],
    "mapping": {
      "core": {"value": "instructions/agents.instructions.md"},
      "instructions": {"value": "instructions/*", "excludes": ["agents.instructions.md"]}
    }
  },
  "platforms": {
    "zeta": {
      "name": "Zeta",
      "detection": ["zeta/marker"],
      "mapping": {
        "core": "ZETA.md",
        "instructions": "zeta/instructions",
        "prompts": "zeta/commands",
        "agents": "zeta/roles"
      }
    },
    "alpha": {
      "name": "Alpha",
      "detection": ["alpha/marker", "alpha/other"],
      "mapping": {
        "core": "ALPHA.md",
        "instructions": "alpha/instructions",
        "prompts": "alpha/commands"
      }
    },
    "beta": {
      "name": "Beta",
      "detection": ["beta/marker"],
      "mapping": {
        "core": "AGENTS.md"
      }
    },
    "nullmap": {
      "name": "NullMap",
      "detection": ["nullmap/marker"],
      "mapping": {
        "core": "null",
        "instructions": "nullmap/instructions"
      }
    },
    "generic_ai": {
      "name": "Generic AI",
      "detection": [],
      "mapping": {
        "core": "AI-INSTRUCTIONS.md",
        "instructions": "ai-config/instructions"
      }
    }
  }
}`

// writeFixtureConfig writes fixtureConfig into dir and returns its path.
func writeFixtureConfig(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "platform.json")
	if err := os.WriteFile(path, []byte(fixtureConfig), 0o644); err != nil {
		t.Fatalf("write fixture config: %v", err)
	}
	return path
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func exists(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	t.Fatalf("stat %s: %v", path, err)
	return false
}

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(data) != want {
		t.Fatalf("%s = %q, want %q", path, data, want)
	}
}

func newCleaner(t *testing.T, target, cfgPath string, level Level, aiChoice string) *Cleaner {
	t.Helper()
	c, err := New(Options{
		TargetDir:   target,
		Level:       level,
		AIChoice:    aiChoice,
		AutoConfirm: true,
		Backup:      true,
	}, cfgPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// captureOutput redirects os.Stdout and os.Stderr for the duration of fn and
// returns what each received. common.LogInfo/LogSuccess/PrintPlan write to
// os.Stdout, LogWarning/LogError to os.Stderr.
func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()
	origOut, origErr := os.Stdout, os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout, os.Stderr = wOut, wErr

	outCh := make(chan string, 1)
	errCh := make(chan string, 1)
	go func() {
		var b strings.Builder
		buf := make([]byte, 4096)
		for {
			n, e := rOut.Read(buf)
			b.Write(buf[:n])
			if e != nil {
				break
			}
		}
		outCh <- b.String()
	}()
	go func() {
		var b strings.Builder
		buf := make([]byte, 4096)
		for {
			n, e := rErr.Read(buf)
			b.Write(buf[:n])
			if e != nil {
				break
			}
		}
		errCh <- b.String()
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				wOut.Close()
				wErr.Close()
				os.Stdout, os.Stderr = origOut, origErr
				panic(r)
			}
		}()
		fn()
	}()

	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = origOut, origErr
	return <-outCh, <-errCh
}

// snapshot records every entry under root (relative path -> content or
// "<dir>") so a test can assert that nothing at all changed.
func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			out[rel] = "<dir>"
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[rel] = string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return out
}

func backupDirs(t *testing.T, target string) []string {
	t.Helper()
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatalf("readdir %s: %v", target, err)
	}
	var out []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".plaesy-backup-") {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

func TestValidLevel(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"safe", true},
		{"thorough", true},
		{"complete", true},
		{"", false},
		{"Safe", false},
		{"SAFE", false},
		{" safe", false},
		{"safe ", false},
		{"safest", false},
		{"generic", false},
	}
	for _, tc := range cases {
		if got := ValidLevel(tc.in); got != tc.want {
			t.Errorf("ValidLevel(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestNewResolvesTargetAndDefaults(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "project")
	mustMkdir(t, target)
	cfgPath := writeFixtureConfig(t, root)

	t.Run("absolute target kept", func(t *testing.T) {
		c := newCleaner(t, target, cfgPath, "", "")
		if c.Opts.TargetDir != target {
			t.Errorf("TargetDir = %q, want %q", c.Opts.TargetDir, target)
		}
		if c.Opts.Level != LevelSafe {
			t.Errorf("Level = %q, want safe", c.Opts.Level)
		}
		if c.ConfigPath != cfgPath {
			t.Errorf("ConfigPath = %q, want %q", c.ConfigPath, cfgPath)
		}
		if c.Config == nil || len(c.Config.Platforms) != 5 {
			t.Errorf("Config not loaded: %+v", c.Config)
		}
	})

	t.Run("relative target made absolute", func(t *testing.T) {
		// New resolves through os.Stat on the process cwd, so the relative
		// form is exercised from a copy of the cwd: the target is passed as
		// "./<base>" relative to the test process, which lives in the package
		// directory, so instead only the resolution contract is checked here.
		c := newCleaner(t, ".", cfgPath, LevelThorough, "")
		if !filepath.IsAbs(c.Opts.TargetDir) {
			t.Errorf("TargetDir = %q, want absolute", c.Opts.TargetDir)
		}
		if c.Opts.TargetDir != filepath.Clean(".") && !strings.HasSuffix(c.Opts.TargetDir, "cleaner") {
			t.Logf("cwd-derived target: %s", c.Opts.TargetDir)
		}
	})

	t.Run("missing target is left unresolved", func(t *testing.T) {
		// os.Stat fails, so the abs conversion at cleaner.go:69 is skipped and
		// the caller's relative string survives into Opts.TargetDir.
		c := newCleaner(t, "does-not-exist-relative", cfgPath, LevelComplete, "alpha")
		if c.Opts.TargetDir != "does-not-exist-relative" {
			t.Errorf("TargetDir = %q, want the unresolved relative string", c.Opts.TargetDir)
		}
	})

	t.Run("target that is a file is left unresolved", func(t *testing.T) {
		file := filepath.Join(root, "not-a-dir.txt")
		mustWrite(t, file, "x")
		c := newCleaner(t, file, cfgPath, LevelSafe, "alpha")
		if c.Opts.TargetDir != file {
			t.Errorf("TargetDir = %q, want %q", c.Opts.TargetDir, file)
		}
	})
}

func TestNewConfigErrors(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "project")
	mustMkdir(t, target)

	t.Run("missing config", func(t *testing.T) {
		_, err := New(Options{TargetDir: target}, filepath.Join(root, "nope.json"))
		if err == nil {
			t.Fatal("expected error for missing config")
		}
		if !strings.Contains(err.Error(), "platform configuration file not found") {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		bad := filepath.Join(root, "bad.json")
		mustWrite(t, bad, "{not json")
		_, err := New(Options{TargetDir: target}, bad)
		if err == nil || !strings.Contains(err.Error(), "invalid JSON syntax") {
			t.Fatalf("error = %v, want invalid JSON syntax", err)
		}
	})
}

func TestValidateEnvironment(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)

	t.Run("writable directory passes and leaves no probe", func(t *testing.T) {
		target := filepath.Join(root, "ok")
		mustMkdir(t, target)
		c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
		if err := c.ValidateEnvironment(); err != nil {
			t.Fatalf("ValidateEnvironment: %v", err)
		}
		if exists(t, filepath.Join(target, ".plaesy-clean-write-test")) {
			t.Error("probe file .plaesy-clean-write-test was left behind")
		}
	})

	t.Run("missing directory", func(t *testing.T) {
		target := filepath.Join(root, "absent")
		c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
		err := c.ValidateEnvironment()
		if err == nil || !strings.Contains(err.Error(), "does not exist") {
			t.Fatalf("error = %v, want does not exist", err)
		}
		if !strings.Contains(err.Error(), target) {
			t.Errorf("error %q does not name the target", err)
		}
	})

	t.Run("target is a file", func(t *testing.T) {
		file := filepath.Join(root, "plain.txt")
		mustWrite(t, file, "x")
		c := newCleaner(t, file, cfgPath, LevelSafe, "alpha")
		err := c.ValidateEnvironment()
		// "exists but is not a directory", not "does not exist". The target is
		// right there; reporting it as missing sends the operator looking for
		// a path problem that is not the one they have. This test pinned the
		// older message, so the misdiagnosis was covered by a test.
		if err == nil || !strings.Contains(err.Error(), "not a directory") {
			t.Fatalf("error = %v, want it to say the target exists but is not a directory", err)
		}
		if strings.Contains(err.Error(), "does not exist") {
			t.Errorf("error = %v, want it not to claim a present target is missing", err)
		}
	})

	t.Run("probe path occupied by a directory reports not writable", func(t *testing.T) {
		// os.Create over an existing directory fails on every supported
		// platform, which is the only portable way to reach the
		// "not writable" branch.
		target := filepath.Join(root, "occupied")
		mustMkdir(t, target)
		mustMkdir(t, filepath.Join(target, ".plaesy-clean-write-test"))
		c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
		err := c.ValidateEnvironment()
		if err == nil || !strings.Contains(err.Error(), "is not writable") {
			t.Fatalf("error = %v, want is not writable", err)
		}
		// The probe it could not create is left in place (no cleanup on error).
		if !exists(t, filepath.Join(target, ".plaesy-clean-write-test")) {
			t.Error("pre-existing probe path was removed on the error path")
		}
	})

	t.Run("unreadable directory", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows ignores the owner-read bit on directories")
		}
		target := filepath.Join(root, "unreadable")
		mustMkdir(t, target)
		if err := os.Chmod(target, 0o000); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(target, 0o755) })
		if os.Geteuid() == 0 {
			t.Skip("root bypasses directory permissions")
		}
		c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
		err := c.ValidateEnvironment()
		if err == nil || !strings.Contains(err.Error(), "is not readable") {
			t.Fatalf("error = %v, want is not readable", err)
		}
	})
}

func TestDetectAllPlatforms(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	cfg := loadFixture(t, cfgPath)

	cases := []struct {
		name    string
		present []string
		want    []string
	}{
		{"no markers falls back to generic_ai", nil, []string{"generic_ai"}},
		{"single marker", []string{"alpha/marker"}, []string{"alpha"}},
		{"declaration order not sorted order", []string{"alpha/marker", "zeta/marker"}, []string{"zeta", "alpha"}},
		{"second detection pattern matches", []string{"alpha/other"}, []string{"alpha"}},
		{"three platforms", []string{"zeta/marker", "alpha/marker", "beta/marker"}, []string{"zeta", "alpha", "beta"}},
		{"platform with null mapping still detected", []string{"nullmap/marker"}, []string{"nullmap"}},
		{"empty-detection platform is never detected", []string{"ai-config"}, []string{"generic_ai"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			target := filepath.Join(root, strings.ReplaceAll(tc.name, " ", "_"))
			mustMkdir(t, target)
			for _, p := range tc.present {
				mustWrite(t, filepath.Join(target, p), "marker")
			}
			got, err := DetectAllPlatforms(cfg, cfgPath, target)
			if err != nil {
				t.Fatalf("DetectAllPlatforms: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("DetectAllPlatforms = %v, want %v", got, tc.want)
			}
		})
	}

	t.Run("one platform listed once despite several matching patterns", func(t *testing.T) {
		target := filepath.Join(root, "dup")
		mustWrite(t, filepath.Join(target, "alpha", "marker"), "m")
		mustWrite(t, filepath.Join(target, "alpha", "other"), "m")
		if got, err := DetectAllPlatforms(cfg, cfgPath, target); err != nil || !reflect.DeepEqual(got, []string{"alpha"}) {
			t.Errorf("DetectAllPlatforms = %v, want [alpha]", got)
		}
	})

	t.Run("unreadable configPath falls back to sorted map keys", func(t *testing.T) {
		// ListPlatforms re-reads configPath; when that fails the function falls
		// back to sorted map keys, so detection order silently changes.
		target := filepath.Join(root, "sorted")
		mustWrite(t, filepath.Join(target, "alpha", "marker"), "m")
		mustWrite(t, filepath.Join(target, "zeta", "marker"), "m")
		got, err := DetectAllPlatforms(cfg, filepath.Join(root, "missing.json"), target)
		if err != nil {
			t.Fatalf("DetectAllPlatforms: %v", err)
		}
		if !reflect.DeepEqual(got, []string{"alpha", "zeta"}) {
			t.Errorf("DetectAllPlatforms = %v, want sorted [alpha zeta]", got)
		}
	})
}

func TestDetectAllPlatformsNilConfigPanics(t *testing.T) {
	// A nil *config.PlatformConfig is not guarded: ListPlatforms survives it
	// but the cfg.Platforms lookup at cleaner.go:133 dereferences nil.
	defer func() {
		if r := recover(); r == nil {
			t.Error("DetectAllPlatforms(nil, ...) returned normally, want a nil dereference panic")
		}
	}()
	DetectAllPlatforms(nil, filepath.Join(t.TempDir(), "platform.json"), t.TempDir())
}

func TestResolvePlatforms(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)

	t.Run("AIChoice wins and splits on whitespace", func(t *testing.T) {
		target := filepath.Join(root, "a")
		mustWrite(t, filepath.Join(target, "alpha", "marker"), "m")
		c := newCleaner(t, target, cfgPath, LevelSafe, "  alpha\tbeta  zeta ")
		want := []string{"alpha", "beta", "zeta"}
		if got, err := c.ResolvePlatforms(); err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("ResolvePlatforms = %v, want %v", got, want)
		}
	})

	t.Run("AIChoice is not validated against the config", func(t *testing.T) {
		c := newCleaner(t, filepath.Join(root, "b"), cfgPath, LevelSafe, "does-not-exist")
		if got, err := c.ResolvePlatforms(); err != nil || !reflect.DeepEqual(got, []string{"does-not-exist"}) {
			t.Errorf("ResolvePlatforms = %v, want the unknown platform verbatim", got)
		}
	})

	t.Run("auto-detect when AIChoice is empty", func(t *testing.T) {
		target := filepath.Join(root, "c")
		mustWrite(t, filepath.Join(target, "zeta", "marker"), "m")
		c := newCleaner(t, target, cfgPath, LevelSafe, "")
		if got, err := c.ResolvePlatforms(); err != nil || !reflect.DeepEqual(got, []string{"zeta"}) {
			t.Errorf("ResolvePlatforms = %v, want [zeta]", got)
		}
	})

	t.Run("no detection yields generic_ai not generic", func(t *testing.T) {
		// The ["generic"] fallback in ResolvePlatforms is unreachable:
		// DetectAllPlatforms already returns ["generic_ai"] for no matches.
		c := newCleaner(t, filepath.Join(root, "d"), cfgPath, LevelSafe, "")
		if got, err := c.ResolvePlatforms(); err != nil || !reflect.DeepEqual(got, []string{"generic_ai"}) {
			t.Errorf("ResolvePlatforms = %v, want [generic_ai]", got)
		}
	})
}

func TestMappingTargetPath(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	c := newCleaner(t, root, cfgPath, LevelSafe, "alpha")

	cases := []struct {
		platform, mappingType, want string
	}{
		{"alpha", "core", "ALPHA.md"},
		{"alpha", "agents", ""},
		{"nullmap", "core", ""},
		{"nullmap", "instructions", "nullmap/instructions"},
		{"unknown", "core", ""},
		{"generic_ai", "core", "AI-INSTRUCTIONS.md"},
	}
	for _, tc := range cases {
		if got := c.mappingTargetPath(tc.platform, tc.mappingType); got != tc.want {
			t.Errorf("mappingTargetPath(%q,%q) = %q, want %q", tc.platform, tc.mappingType, got, tc.want)
		}
	}
}

func TestBuildPlan(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)

	t.Run("empty project", func(t *testing.T) {
		target := filepath.Join(root, "empty")
		mustMkdir(t, target)
		c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
		plan := c.BuildPlan()
		if !plan.Empty {
			t.Errorf("Empty = false, plan = %+v", plan)
		}
		if len(plan.PlatformEntries) != 0 {
			t.Errorf("PlatformEntries = %v, want empty", plan.PlatformEntries)
		}
	})

	t.Run("dirs and mapped entries are labelled", func(t *testing.T) {
		target := filepath.Join(root, "full")
		mustMkdir(t, filepath.Join(target, ".plaesy", "specs"))
		mustMkdir(t, filepath.Join(target, "specs"))
		mustWrite(t, filepath.Join(target, "ALPHA.md"), "a")
		mustMkdir(t, filepath.Join(target, "alpha", "instructions"))
		mustWrite(t, filepath.Join(target, "keepme.txt"), "k")

		c := newCleaner(t, target, cfgPath, LevelThorough, "alpha")
		plan := c.BuildPlan()
		if want := []string{".plaesy", "specs"}; !reflect.DeepEqual(plan.PlaesyDirs, want) {
			t.Errorf("PlaesyDirs = %v, want %v", plan.PlaesyDirs, want)
		}
		want := []string{"ALPHA.md (file)", "alpha/instructions/ (directory)"}
		if got := plan.PlatformEntries["alpha"]; !reflect.DeepEqual(got, want) {
			t.Errorf("PlatformEntries[alpha] = %v, want %v", got, want)
		}
		if plan.Empty {
			t.Error("Empty = true, want false")
		}
		if got := plan.Platforms; !reflect.DeepEqual(got, []string{"alpha"}) {
			t.Errorf("Platforms = %v, want [alpha]", got)
		}
	})

	t.Run("null mapping and missing targets are skipped", func(t *testing.T) {
		target := filepath.Join(root, "nulls")
		mustWrite(t, filepath.Join(target, "nullmap", "instructions", "x.md"), "x")
		c := newCleaner(t, target, cfgPath, LevelSafe, "nullmap")
		plan := c.BuildPlan()
		if got := plan.PlatformEntries["nullmap"]; !reflect.DeepEqual(got, []string{"nullmap/instructions/ (directory)"}) {
			t.Errorf("PlatformEntries[nullmap] = %v", got)
		}
	})

	t.Run("generic platforms are skipped but still reported", func(t *testing.T) {
		target := filepath.Join(root, "generic")
		mustWrite(t, filepath.Join(target, "AI-INSTRUCTIONS.md"), "a")
		for _, ai := range []string{"generic", "generic_ai"} {
			c := newCleaner(t, target, cfgPath, LevelSafe, ai)
			plan := c.BuildPlan()
			if !reflect.DeepEqual(plan.Platforms, []string{ai}) {
				t.Errorf("Platforms = %v, want [%s]", plan.Platforms, ai)
			}
			if len(plan.PlatformEntries) != 0 {
				t.Errorf("PlatformEntries = %v, want empty for %s", plan.PlatformEntries, ai)
			}
			if !plan.Empty {
				t.Errorf("Empty = false for %s", ai)
			}
		}
	})

	t.Run("plaesy mapping targets are never planned", func(t *testing.T) {
		// plaesy.mapping.core is instructions/agents.instructions.md; the
		// cleaner only ever reads platforms.*.mapping, so this file is never
		// reported even when it exists.
		target := filepath.Join(root, "structmap")
		mustWrite(t, filepath.Join(target, "instructions", "agents.instructions.md"), "x")
		c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
		plan := c.BuildPlan()
		if !plan.Empty {
			t.Errorf("plan not empty: %+v", plan)
		}
	})

	// The level decides what the plan offers to delete, so the plan is not the
	// same document at every level: safe must not list specs/, because the
	// command's help reserves it for thorough.
	t.Run("level decides what the plan offers to delete", func(t *testing.T) {
		target := filepath.Join(root, "levels")
		mustMkdir(t, filepath.Join(target, ".plaesy"))
		mustMkdir(t, filepath.Join(target, "specs"))
		mustWrite(t, filepath.Join(target, "ALPHA.md"), "a")

		safe := newCleaner(t, target, cfgPath, LevelSafe, "alpha").BuildPlan()
		thorough := newCleaner(t, target, cfgPath, LevelThorough, "alpha").BuildPlan()
		complete := newCleaner(t, target, cfgPath, LevelComplete, "alpha").BuildPlan()

		if want := []string{".plaesy"}; !reflect.DeepEqual(safe.PlaesyDirs, want) {
			t.Errorf("safe PlaesyDirs = %v, want %v: specs/ is the user's own work", safe.PlaesyDirs, want)
		}
		for name, plan := range map[string]Plan{"thorough": thorough, "complete": complete} {
			if want := []string{".plaesy", "specs"}; !reflect.DeepEqual(plan.PlaesyDirs, want) {
				t.Errorf("%s PlaesyDirs = %v, want %v", name, plan.PlaesyDirs, want)
			}
		}
		if !reflect.DeepEqual(safe.PlatformEntries, thorough.PlatformEntries) {
			t.Errorf("the platform mapping should not depend on the level: %v vs %v",
				safe.PlatformEntries, thorough.PlatformEntries)
		}
	})
}

func TestPrintPlan(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)

	build := func(t *testing.T, level Level, dryRun, backup bool, ai string) (Plan, *Cleaner) {
		t.Helper()
		target := filepath.Join(root, "plan-"+string(level))
		mustMkdir(t, filepath.Join(target, ".plaesy"))
		mustWrite(t, filepath.Join(target, "ALPHA.md"), "a")
		mustMkdir(t, filepath.Join(target, "alpha", "instructions"))
		c, err := New(Options{TargetDir: target, Level: level, DryRun: dryRun, Backup: backup, AIChoice: ai}, cfgPath)
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		return c.BuildPlan(), c
	}

	levels := map[Level]string{
		LevelSafe:     "Safe cleanup: Framework files only, preserve user code",
		LevelThorough: "Thorough cleanup: Framework + specs, preserve user code",
		LevelComplete: "Complete cleanup: Everything Plaesy-related (DANGEROUS)",
	}
	for lvl, want := range levels {
		t.Run("level "+string(lvl), func(t *testing.T) {
			plan, c := build(t, lvl, false, false, "alpha")
			out, errOut := captureOutput(t, func() { c.PrintPlan(plan) })
			if errOut != "" {
				t.Errorf("unexpected stderr: %s", errOut)
			}
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q:\n%s", want, out)
			}
			if !strings.Contains(out, "Cleanup Plan (Level: "+string(lvl)+", Platforms: alpha)") {
				t.Errorf("missing header:\n%s", out)
			}
			if !strings.Contains(out, "  - .plaesy/") {
				t.Errorf("missing plaesy dir entry:\n%s", out)
			}
			if !strings.Contains(out, "  - ALPHA.md (file)") {
				t.Errorf("missing file entry:\n%s", out)
			}
			if !strings.Contains(out, "  - Your source code (src/, lib/, components/, etc.)") {
				t.Errorf("missing preserved section:\n%s", out)
			}
			if strings.Contains(out, "[DRY RUN MODE]") {
				t.Errorf("dry run banner printed with DryRun=false:\n%s", out)
			}
			if strings.Contains(out, "Backup will be created") {
				t.Errorf("backup note printed with Backup=false:\n%s", out)
			}
		})
	}

	t.Run("dry run banner and suppressed backup note", func(t *testing.T) {
		plan, c := build(t, LevelSafe, true, true, "alpha")
		out, _ := captureOutput(t, func() { c.PrintPlan(plan) })
		if !strings.Contains(out, "[DRY RUN MODE] No files will be actually removed.") {
			t.Errorf("missing dry-run banner:\n%s", out)
		}
		if strings.Contains(out, "Backup will be created before removal") {
			t.Errorf("backup note printed during dry run:\n%s", out)
		}
	})

	t.Run("backup note when enabled", func(t *testing.T) {
		plan, c := build(t, LevelSafe, false, true, "alpha")
		out, _ := captureOutput(t, func() { c.PrintPlan(plan) })
		if !strings.Contains(out, "  - Backup will be created before removal") {
			t.Errorf("missing backup note:\n%s", out)
		}
	})

	t.Run("no platform entries prints both the per-platform and the generic line", func(t *testing.T) {
		// foundAnyPlatform is driven by the size of PlatformEntries, not by the
		// resolved platform list, so a listed platform with no matched targets
		// still gets the "Generic platform" placeholder appended.
		target := filepath.Join(root, "noentries")
		mustMkdir(t, filepath.Join(target, ".plaesy"))
		c := newCleaner(t, target, cfgPath, LevelSafe, "beta")
		plan := c.BuildPlan()
		out, _ := captureOutput(t, func() { c.PrintPlan(plan) })
		if !strings.Contains(out, "  - No platform-specific files found for beta") {
			t.Errorf("missing per-platform empty line:\n%s", out)
		}
		if !strings.Contains(out, "  - Generic platform: No specific files to remove") {
			t.Errorf("missing generic placeholder:\n%s", out)
		}
	})

	t.Run("generic placeholder when only generic platforms resolved", func(t *testing.T) {
		target := filepath.Join(root, "onlygeneric")
		mustMkdir(t, target)
		c := newCleaner(t, target, cfgPath, LevelSafe, "generic")
		plan := c.BuildPlan()
		out, _ := captureOutput(t, func() { c.PrintPlan(plan) })
		if !strings.Contains(out, "\nPlatform-Specific Files:\n  - Generic platform: No specific files to remove") {
			t.Errorf("missing generic placeholder:\n%s", out)
		}
	})

	t.Run("unknown level prints no level description", func(t *testing.T) {
		plan, c := build(t, Level("bogus"), false, false, "alpha")
		out, _ := captureOutput(t, func() { c.PrintPlan(plan) })
		if strings.Contains(out, "cleanup:") {
			t.Errorf("level description printed for unknown level:\n%s", out)
		}
		if !strings.Contains(out, "Cleanup Plan (Level: bogus") {
			t.Errorf("header missing:\n%s", out)
		}
	})
}

func TestConfirm(t *testing.T) {
	cases := []struct {
		name        string
		autoConfirm bool
		input       string
		want        bool
		wantOut     string
	}{
		{"auto confirm skips the prompt", true, "", true, "Auto-confirm mode: proceeding with deletion..."},
		{"y", false, "y\n", true, "Are you sure you want to continue? (y/N): "},
		{"Y uppercase", false, "Y\n", true, ""},
		{"yes", false, "yes\n", true, ""},
		{"YES uppercase", false, "YES\n", true, ""},
		{"y without trailing newline", false, "y", true, ""},
		{"padded yes", false, "  yes  \n", true, ""},
		{"n", false, "n\n", false, ""},
		{"empty line", false, "\n", false, ""},
		{"eof", false, "", false, ""},
		{"garbage", false, "maybe\n", false, ""},
		{"yEs mixed case", false, "yEs\n", true, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			inPath := filepath.Join(root, "stdin")
			if err := os.WriteFile(inPath, []byte(tc.input), 0o644); err != nil {
				t.Fatalf("write stdin: %v", err)
			}
			in, err := os.Open(inPath)
			if err != nil {
				t.Fatalf("open stdin: %v", err)
			}
			defer in.Close()
			outPath := filepath.Join(root, "stdout")
			out, err := os.Create(outPath)
			if err != nil {
				t.Fatalf("create stdout: %v", err)
			}

			got := Confirm(Options{AutoConfirm: tc.autoConfirm}, bufio.NewReader(in), out)
			out.Close()
			if got != tc.want {
				t.Errorf("Confirm = %v, want %v", got, tc.want)
			}
			written, err := os.ReadFile(outPath)
			if err != nil {
				t.Fatalf("read stdout: %v", err)
			}
			s := string(written)
			if tc.wantOut != "" && !strings.Contains(s, tc.wantOut) {
				t.Errorf("stdout = %q, want it to contain %q", s, tc.wantOut)
			}
			if !tc.autoConfirm && !strings.Contains(s, "This will permanently delete the directories and files listed above.") {
				t.Errorf("stdout = %q, missing the warning line", s)
			}
		})
	}
}

func TestCreateBackupDisabled(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	target := filepath.Join(root, "p")
	mustMkdir(t, target)

	c, err := New(Options{TargetDir: target, Backup: false}, cfgPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	dir, err := c.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}
	if dir != "" {
		t.Errorf("backup dir = %q, want empty when Backup is false", dir)
	}
	if got := backupDirs(t, target); len(got) != 0 {
		t.Errorf("backup directories created anyway: %v", got)
	}
}

func TestCreateBackup(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	target := filepath.Join(root, "p")
	mustMkdir(t, filepath.Join(target, ".plaesy", "memory"))
	mustWrite(t, filepath.Join(target, ".plaesy", "memory", "notes.md"), "mem")
	mustWrite(t, filepath.Join(target, ".plaesy", "top.txt"), "top")
	mustWrite(t, filepath.Join(target, "specs", "spec.md"), "user spec")
	mustWrite(t, filepath.Join(target, "ALPHA.md"), "alpha core")
	mustMkdir(t, filepath.Join(target, "alpha", "instructions"))
	mustWrite(t, filepath.Join(target, "alpha", "instructions", "i.md"), "i")
	mustWrite(t, filepath.Join(target, "src", "main.go"), "package main")
	mustWrite(t, filepath.Join(target, "AI-INSTRUCTIONS.md"), "generic core")

	c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
	dir, err := c.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}
	if dir == "" {
		t.Fatal("CreateBackup returned an empty path")
	}
	if filepath.Dir(dir) != target {
		t.Errorf("backup dir %q is not directly inside the target", dir)
	}
	base := filepath.Base(dir)
	if !strings.HasPrefix(base, ".plaesy-backup-") {
		t.Errorf("backup dir name %q lacks the .plaesy-backup- prefix", base)
	}
	stamp := strings.TrimPrefix(base, ".plaesy-backup-")
	if _, err := time.Parse("20060102-150405", stamp); err != nil {
		t.Errorf("backup dir timestamp %q: %v", stamp, err)
	}

	assertContent(t, filepath.Join(dir, ".plaesy", "memory", "notes.md"), "mem")
	assertContent(t, filepath.Join(dir, ".plaesy", "top.txt"), "top")
	assertContent(t, filepath.Join(dir, "ALPHA.md"), "alpha core")
	assertContent(t, filepath.Join(dir, "alpha", "instructions", "i.md"), "i")

	// safe does not delete specs/, so the backup does not carry it either.
	if exists(t, filepath.Join(dir, "specs")) {
		t.Error("specs/ was copied into the backup at level safe, which does not delete it")
	}
	if exists(t, filepath.Join(dir, "AI-INSTRUCTIONS.md")) {
		t.Error("generic_ai mapping target was backed up")
	}
	if exists(t, filepath.Join(dir, "src")) {
		t.Error("unrelated src/ was backed up")
	}

	// The originals are untouched by a backup.
	assertContent(t, filepath.Join(target, "specs", "spec.md"), "user spec")
	assertContent(t, filepath.Join(target, "ALPHA.md"), "alpha core")
}

// A backup that leaves out a directory the same run deletes is not a backup.
// specs/ is exactly that case above the safe level, and --backup is the default,
// so the promise "backup created before removal" used to break silently.
func TestCreateBackupCoversEveryDirectoryTheLevelDeletes(t *testing.T) {
	for _, level := range []Level{LevelThorough, LevelComplete} {
		t.Run(string(level), func(t *testing.T) {
			root := t.TempDir()
			cfgPath := writeFixtureConfig(t, root)
			target := filepath.Join(root, "p")
			mustWrite(t, filepath.Join(target, ".plaesy", "top.txt"), "top")
			mustWrite(t, filepath.Join(target, "specs", "001-x", "spec.md"), "user spec")
			mustWrite(t, filepath.Join(target, "ALPHA.md"), "alpha core")

			c := newCleaner(t, target, cfgPath, level, "alpha")
			dir, err := c.CreateBackup()
			if err != nil {
				t.Fatalf("CreateBackup: %v", err)
			}
			assertContent(t, filepath.Join(dir, ".plaesy", "top.txt"), "top")
			assertContent(t, filepath.Join(dir, "specs", "001-x", "spec.md"), "user spec")

			// And the run it backs up really does delete it.
			captureOutput(t, func() {
				if _, err := c.Remove(); err != nil {
					t.Fatalf("Remove: %v", err)
				}
			})
			if exists(t, filepath.Join(target, "specs")) {
				t.Error("specs/ survived a thorough/complete clean")
			}
			assertContent(t, filepath.Join(dir, "specs", "001-x", "spec.md"), "user spec")
		})
	}
}

func TestCreateBackupNestedFileTargetCreatesParents(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "p")
	mustWrite(t, filepath.Join(target, ".github", "copilot-instructions.md"), "copilot")

	nested := `{
  "version": "1.0",
  "platforms": {
    "gh": {"name":"GH","detection":["github-marker"],"mapping":{"core":".github/copilot-instructions.md"}}
  }
}`
	nestedPath := filepath.Join(root, "nested.json")
	mustWrite(t, nestedPath, nested)

	c := newCleaner(t, target, nestedPath, LevelSafe, "gh")
	dir, err := c.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}
	assertContent(t, filepath.Join(dir, ".github", "copilot-instructions.md"), "copilot")
}

func TestCreateBackupMkdirFails(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	target := filepath.Join(root, "p")
	mustMkdir(t, target)

	// Occupy the two timestamps the backup name could take with regular
	// files, so os.MkdirAll cannot create the backup directory.
	now := time.Now()
	for _, d := range []time.Time{now, now.Add(time.Second)} {
		name := ".plaesy-backup-" + d.Format("20060102-150405")
		if exists(t, filepath.Join(target, name)) {
			t.Skipf("backup name %s already exists", name)
		}
		mustWrite(t, filepath.Join(target, name), "blocker")
	}

	c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
	out, errOut := captureOutput(t, func() {
		dir, err := c.CreateBackup()
		if err == nil {
			t.Fatalf("CreateBackup succeeded with path %q, want an error", dir)
		}
	})
	if !strings.Contains(out, "Creating backup:") {
		t.Errorf("stdout = %q, want the Creating backup line", out)
	}
	if errOut != "" {
		t.Errorf("unexpected stderr: %s", errOut)
	}
}

func TestCreateBackupSkipsGenericPlatforms(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	for _, ai := range []string{"generic", "generic_ai"} {
		t.Run(ai, func(t *testing.T) {
			target := filepath.Join(root, ai)
			mustWrite(t, filepath.Join(target, "AI-INSTRUCTIONS.md"), "a")
			mustMkdir(t, filepath.Join(target, "ai-config", "instructions"))
			c := newCleaner(t, target, cfgPath, LevelSafe, ai)
			dir, err := c.CreateBackup()
			if err != nil {
				t.Fatalf("CreateBackup: %v", err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("readdir: %v", err)
			}
			if len(entries) != 0 {
				t.Errorf("backup for %s contains %d entries, want none", ai, len(entries))
			}
		})
	}
}

func TestCreateBackupPlaesyCopyFails(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only Windows is probed for a read denial")
	}
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	target := filepath.Join(root, "p")
	ctx := filepath.Join(target, ".plaesy", "memory", "ctx.md")
	mustWrite(t, ctx, "ctx")
	denyACL(t, ctx, "R")

	c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
	dir, err := c.CreateBackup()
	if err == nil {
		t.Fatalf("CreateBackup succeeded with %q, want a copy error", dir)
	}
	if dir != "" {
		t.Errorf("CreateBackup returned %q alongside the error, want an empty path", dir)
	}
	// The partially built backup directory is not cleaned up on failure.
	if got := backupDirs(t, target); len(got) != 1 {
		t.Errorf("backup directories = %v, want the partial one to remain", got)
	}
}

func TestCreateBackupMappedDirCopyFails(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only Windows is probed for a read denial")
	}
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	target := filepath.Join(root, "p")
	inner := filepath.Join(target, "alpha", "instructions", "i.md")
	mustWrite(t, inner, "i")
	mustWrite(t, filepath.Join(target, "ALPHA.md"), "a")
	denyACL(t, inner, "R")

	c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
	dir, err := c.CreateBackup()
	if err == nil {
		t.Fatalf("CreateBackup succeeded with %q, want a copy error", dir)
	}
	if dir != "" {
		t.Errorf("CreateBackup returned %q alongside the error, want an empty path", dir)
	}
}

func TestCreateBackupNoSourceIsStillCreated(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	target := filepath.Join(root, "p")
	mustMkdir(t, target)

	c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
	dir, err := c.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir %s: %v", dir, err)
	}
	if len(entries) != 0 {
		t.Errorf("empty backup contains %d entries", len(entries))
	}
}

// claudeProject populates a target with a claude_code-shaped layout plus user
// files that must survive.
func claudeProject(t *testing.T, target string) {
	t.Helper()
	mustWrite(t, filepath.Join(target, ".plaesy", "memory", "ctx.md"), "ctx")
	mustWrite(t, filepath.Join(target, "specs", "001-feature", "spec.md"), "spec")
	mustWrite(t, filepath.Join(target, "CLAUDE.md"), "claude core")
	mustWrite(t, filepath.Join(target, ".claude", "commands", "plan.md"), "plan cmd")
	mustWrite(t, filepath.Join(target, ".claude", "instructions", "i.md"), "i")
	mustWrite(t, filepath.Join(target, ".claude", "roles", "r.md"), "r")
	mustWrite(t, filepath.Join(target, "src", "main.go"), "package main")
	mustWrite(t, filepath.Join(target, "README.md"), "readme")
}

const claudeConfig = `{
  "version": "1.0",
  "plaesy": {"base_directory": ".plaesy"},
  "platforms": {
    "claude_code": {
      "name": "Claude Code",
      "detection": ["CLAUDE.md"],
      "mapping": {
        "core": "CLAUDE.md",
        "instructions": ".claude/instructions",
        "prompts": ".claude/commands",
        "agents": ".claude/roles"
      }
    }
  }
}`

func TestRemoveRespectsTheLevel(t *testing.T) {
	root := t.TempDir()
	cfgPath := filepath.Join(root, "claude.json")
	mustWrite(t, cfgPath, claudeConfig)

	type result struct {
		removed  int
		leftover []string
	}
	var first result
	for i, lvl := range []Level{LevelSafe, LevelThorough, LevelComplete} {
		t.Run(string(lvl), func(t *testing.T) {
			target := filepath.Join(root, string(lvl))
			claudeProject(t, target)
			c := newCleaner(t, target, cfgPath, lvl, "claude_code")
			before := snapshot(t, target)

			var removed int
			out, _ := captureOutput(t, func() {
				n, err := c.Remove()
				if err != nil {
					t.Fatalf("Remove: %v", err)
				}
				removed = n
			})

			// specs/ is the user's own feature work. safe says "framework files
			// only, preserve user code", so it must survive there and be removed
			// only once the level is raised.
			specsKept := lvl == LevelSafe
			wantRemoved := 7
			if specsKept {
				wantRemoved = 6
			}
			if removed != wantRemoved {
				t.Errorf("Remove = %d, want %d\n%s", removed, wantRemoved, out)
			}
			for _, gone := range []string{
				".plaesy", "CLAUDE.md", ".claude",
				filepath.Join(".claude", "commands"),
				filepath.Join(".claude", "instructions"),
				filepath.Join(".claude", "roles"),
			} {
				if exists(t, filepath.Join(target, gone)) {
					t.Errorf("%s still exists after Remove (level %s)", gone, lvl)
				}
			}
			if specsKept {
				assertContent(t, filepath.Join(target, "specs", "001-feature", "spec.md"), "spec")
			} else if exists(t, filepath.Join(target, "specs")) {
				t.Errorf("specs still exists after Remove (level %s)", lvl)
			}
			if !strings.Contains(out, "Removed empty parent directory: .claude") {
				t.Errorf("stdout missing the .claude prune line:\n%s", out)
			}
			assertContent(t, filepath.Join(target, "src", "main.go"), "package main")
			assertContent(t, filepath.Join(target, "README.md"), "readme")

			after := snapshot(t, target)
			delete(before, ".plaesy")
			delete(after, ".plaesy")
			if specsKept {
				delete(before, "specs")
				delete(after, "specs")
			}
			gotLeft := leftoverOf(after, before)
			if specsKept {
				// specs/ and its contents are in both snapshots, so nothing is
				// reported as newly created either way.
				for _, entry := range gotLeft {
					if strings.HasPrefix(entry, "specs") {
						t.Errorf("%s reported as created, want the untouched specs/ tree", entry)
					}
				}
			}
			if i == 0 {
				first = result{removed: removed, leftover: gotLeft}
			} else if !reflect.DeepEqual(gotLeft, first.leftover) {
				t.Errorf("level %s differs from safe in what survives: %v vs %v",
					lvl, gotLeft, first.leftover)
			}
		})
	}
}

func leftoverOf(after, before map[string]string) []string {
	var out []string
	for k := range after {
		if _, ok := before[k]; !ok {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func TestRemovePrunesEmptyParents(t *testing.T) {
	root := t.TempDir()
	cfgPath := filepath.Join(root, "cursor.json")
	mustWrite(t, cfgPath, `{
  "version": "1.0",
  "platforms": {
    "cursor_ai": {
      "name": "Cursor",
      "detection": [".cursorrules"],
      "mapping": {
        "core": ".cursor/rules/plaesy.mdc",
        "instructions": ".cursor/instructions",
        "prompts": ".cursor/rules"
      }
    }
  }
}`)
	target := filepath.Join(root, "p")
	mustWrite(t, filepath.Join(target, ".cursor", "rules", "plaesy.mdc"), "mdc")
	mustMkdir(t, filepath.Join(target, ".cursor", "instructions"))
	mustWrite(t, filepath.Join(target, ".cursorrules"), "rules")

	// A second, non-empty parent must survive.
	mustWrite(t, filepath.Join(target, ".cursor", "user", "keep.txt"), "k")

	c := newCleaner(t, target, cfgPath, LevelComplete, "cursor_ai")
	var removed int
	captureOutput(t, func() {
		var err error
		removed, err = c.Remove()
		if err != nil {
			t.Fatalf("Remove: %v", err)
		}
	})
	if removed != 3 {
		t.Errorf("Remove = %d, want 3 (.cursor/rules, .cursor/instructions, .cursorrules)", removed)
	}
	if exists(t, filepath.Join(target, ".cursor", "rules")) {
		t.Error(".cursor/rules survived")
	}
	if !exists(t, filepath.Join(target, ".cursor", "user", "keep.txt")) {
		t.Error("user file was removed")
	}
}

func TestRemovePrunesNowEmptyUserParent(t *testing.T) {
	root := t.TempDir()
	cfgPath := filepath.Join(root, "p.json")
	mustWrite(t, cfgPath, `{
  "version": "1.0",
  "platforms": {
    "odd": {
      "name": "Odd",
      "detection": ["odd-marker"],
      "mapping": {"core": "tools/plaesy/only-file.md"}
    }
  }
}`)
	target := filepath.Join(root, "p")
	// tools/plaesy contains nothing but a mapped file, so both tools/plaesy
	// and tools end up empty and are pruned.
	mustWrite(t, filepath.Join(target, "tools", "plaesy", "only-file.md"), "f")

	c := newCleaner(t, target, cfgPath, LevelSafe, "odd")
	var removed int
	captureOutput(t, func() {
		var err error
		removed, err = c.Remove()
		if err != nil {
			t.Fatalf("Remove: %v", err)
		}
	})
	// tools/plaesy holds nothing but a mapped file, so it is pruned, but the
	// prune is not recursive: tools/ itself is not a parent of any mapping
	// target and stays behind.
	if removed != 2 {
		t.Errorf("Remove = %d, want 2 (the file and the empty tools/plaesy parent)", removed)
	}
	if exists(t, filepath.Join(target, "tools", "plaesy")) {
		t.Error("tools/plaesy survived")
	}
	if !exists(t, filepath.Join(target, "tools")) {
		t.Error("tools/ was pruned; only the direct parent of a target is")
	}
}

func TestRemoveNoMatches(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	target := filepath.Join(root, "p")
	mustWrite(t, filepath.Join(target, "README.md"), "r")
	mustWrite(t, filepath.Join(target, "src", "main.go"), "pkg")

	c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
	before := snapshot(t, target)
	var removed int
	out, errOut := captureOutput(t, func() {
		var err error
		removed, err = c.Remove()
		if err != nil {
			t.Fatalf("Remove: %v", err)
		}
	})
	if removed != 0 {
		t.Errorf("Remove = %d, want 0", removed)
	}
	if !strings.Contains(errOut, "No items were removed") {
		t.Errorf("stderr = %q, want the No items were removed warning", errOut)
	}
	if strings.Contains(out, "Successfully removed") {
		t.Errorf("stdout = %q, want no success line", out)
	}
	if !reflect.DeepEqual(snapshot(t, target), before) {
		t.Error("target changed although nothing matched")
	}
}

func TestRemoveMultiplePlatforms(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	target := filepath.Join(root, "p")
	mustWrite(t, filepath.Join(target, "ALPHA.md"), "a")
	mustWrite(t, filepath.Join(target, "AGENTS.md"), "b")
	mustWrite(t, filepath.Join(target, "alpha", "commands", "c.md"), "c")
	mustWrite(t, filepath.Join(target, "beta-only.txt"), "b2")

	c := newCleaner(t, target, cfgPath, LevelThorough, "alpha beta")
	var removed int
	out, _ := captureOutput(t, func() {
		var err error
		removed, err = c.Remove()
		if err != nil {
			t.Fatalf("Remove: %v", err)
		}
	})
	if removed != 4 {
		t.Errorf("Remove = %d, want 4 (ALPHA.md, alpha/commands, the empty alpha/ parent, AGENTS.md)\n%s", removed, out)
	}
	if exists(t, filepath.Join(target, "ALPHA.md")) || exists(t, filepath.Join(target, "AGENTS.md")) {
		t.Error("core files survived")
	}
	if exists(t, filepath.Join(target, "alpha", "commands")) {
		t.Error("alpha/commands survived")
	}
	// alpha/commands was the only entry of alpha/, so alpha/ is pruned too.
	if exists(t, filepath.Join(target, "alpha")) {
		t.Error("empty alpha/ parent survived")
	}
	if !strings.Contains(out, "Removed empty parent directory: alpha") {
		t.Errorf("stdout missing the parent prune line:\n%s", out)
	}
	if !strings.Contains(out, "Successfully removed 4 items") {
		t.Errorf("stdout missing the statistics line:\n%s", out)
	}
}

func TestRemoveGenericAIOnlyAtComplete(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)

	// AI-INSTRUCTIONS.md and ai-config/ are mapped by the "various" fallback
	// platform, so nothing in them is certainly Plaesy's own. They survive until
	// the level says "remove everything Plaesy-related".
	cases := []struct {
		level       Level
		platform    string
		wantRemoved int
		wantKept    []string
	}{
		{LevelSafe, "generic_ai", 1, []string{"AI-INSTRUCTIONS.md", "ai-config"}},
		{LevelThorough, "generic_ai", 1, []string{"AI-INSTRUCTIONS.md", "ai-config"}},
		{LevelComplete, "generic_ai", 4, nil},
		{LevelComplete, "generic", 1, []string{"AI-INSTRUCTIONS.md", "ai-config"}},
	}
	for _, tc := range cases {
		name := string(tc.level) + "/" + tc.platform
		t.Run(name, func(t *testing.T) {
			target := filepath.Join(root, name)
			mustWrite(t, filepath.Join(target, "AI-INSTRUCTIONS.md"), "a")
			mustMkdir(t, filepath.Join(target, "ai-config", "instructions"))
			mustWrite(t, filepath.Join(target, ".plaesy", "x.md"), "x")

			c := newCleaner(t, target, cfgPath, tc.level, tc.platform)
			var removed int
			captureOutput(t, func() {
				var err error
				removed, err = c.Remove()
				if err != nil {
					t.Fatalf("Remove: %v", err)
				}
			})
			if removed != tc.wantRemoved {
				t.Errorf("Remove = %d, want %d", removed, tc.wantRemoved)
			}
			for _, kept := range tc.wantKept {
				if !exists(t, filepath.Join(target, kept)) {
					t.Errorf("%s was removed at level %s", kept, tc.level)
				}
			}
			if exists(t, filepath.Join(target, ".plaesy", "x.md")) {
				t.Error(".plaesy should have been removed at every level")
			}
			if tc.wantRemoved == 4 {
				for _, gone := range []string{"AI-INSTRUCTIONS.md", "ai-config"} {
					if exists(t, filepath.Join(target, gone)) {
						t.Errorf("%s survived a complete clean", gone)
					}
				}
			}
		})
	}
}

func TestRemoveUnknownPlatformRemovesOnlyPlaesyDirs(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	// A platform that is not in platform.json has no mapping to follow, so only
	// the top-level framework directories are candidates.
	c := newCleaner(t, filepath.Join(root, "p"), cfgPath, LevelThorough, "not-a-platform")
	target := c.Opts.TargetDir
	mustWrite(t, filepath.Join(target, "ALPHA.md"), "a")
	mustWrite(t, filepath.Join(target, ".plaesy", "x.md"), "x")
	mustMkdir(t, filepath.Join(target, "specs"))

	var removed int
	captureOutput(t, func() {
		var err error
		removed, err = c.Remove()
		if err != nil {
			t.Fatalf("Remove: %v", err)
		}
	})
	if removed != 2 {
		t.Errorf("Remove = %d, want 2 (.plaesy and specs)", removed)
	}
	if !exists(t, filepath.Join(target, "ALPHA.md")) {
		t.Error("ALPHA.md was removed for an unknown platform")
	}
}

// denyACL adds a deny ACE for Everyone on path and removes it again when the
// test ends. The read-only attribute alone is not enough to make removal or
// copying fail: Go's os.Remove/RemoveAll clear it and retry, which is why the
// failure branches need a real ACL denial.
func denyACL(t *testing.T, path, rights string) {
	t.Helper()
	icacls, err := exec.LookPath("icacls")
	if err != nil {
		t.Skipf("icacls not available: %v", err)
	}
	if out, err := exec.Command(icacls, path, "/deny", "*S-1-1-0:("+rights+")").CombinedOutput(); err != nil {
		t.Skipf("icacls deny on %s: %v: %s", path, err, out)
	}
	t.Cleanup(func() {
		_ = exec.Command(icacls, path, "/remove:d", "*S-1-1-0").Run()
	})
}

func denyWrite(t *testing.T, path string) { denyACL(t, path, "W") }

func TestRemovePartialFailure(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only Windows is probed for a removal denial")
	}
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	target := filepath.Join(root, "p")
	ro := filepath.Join(target, "ALPHA.md")
	mustWrite(t, ro, "a")
	if err := os.Chmod(ro, 0o444); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(ro, 0o666) })
	denyWrite(t, ro)
	mustWrite(t, filepath.Join(target, "alpha", "commands", "c.md"), "c")

	c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
	var removed int
	var removeErr error
	out, errOut := captureOutput(t, func() {
		removed, removeErr = c.Remove()
	})
	// A partial failure used to be downgraded to a warning and reported as
	// nil, so `plaesy clean` printed "completed!" and exited 0 with the
	// denied files still in place. The successful removals must still be
	// counted and reported; the failure must still reach the caller.
	if removeErr == nil {
		t.Error("Remove must return an error for a partial failure, got nil")
	}
	if removed != 2 {
		t.Errorf("Remove = %d, want 2 (alpha/commands and the empty alpha/ parent)", removed)
	}
	if !exists(t, ro) {
		t.Error("denied ALPHA.md was removed anyway")
	}
	if !strings.Contains(out, "Failed to remove: ALPHA.md") {
		t.Errorf("stdout = %q, want the failure line", out)
	}
	if !strings.Contains(errOut, "Failed to remove some items: ALPHA.md") {
		t.Errorf("stderr = %q, want the failure warning", errOut)
	}
	if !strings.Contains(out, "Removed platform alpha prompts: alpha/commands") {
		t.Errorf("stdout = %q, want the successful removal line", out)
	}
}

func TestRemovePlaesyDirFailure(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only Windows is probed for a removal denial")
	}
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	target := filepath.Join(root, "p")
	plaesyDir := filepath.Join(target, ".plaesy")
	mustWrite(t, filepath.Join(plaesyDir, "memory", "ctx.md"), "ctx")
	if err := os.Chmod(plaesyDir, 0o500); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(plaesyDir, 0o755) })
	denyWrite(t, plaesyDir)
	mustWrite(t, filepath.Join(target, "ALPHA.md"), "a")

	c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
	var removed int
	var removeErr error
	out, errOut := captureOutput(t, func() {
		removed, removeErr = c.Remove()
	})
	// A partial failure used to be downgraded to a warning and reported as
	// nil, so `plaesy clean` printed "completed!" and exited 0 with the
	// denied files still in place. The successful removals must still be
	// counted and reported; the failure must still reach the caller.
	if removeErr == nil {
		t.Error("Remove must return an error for a partial failure, got nil")
	}
	if removed != 1 {
		t.Errorf("Remove = %d, want 1 (only ALPHA.md)", removed)
	}
	if !exists(t, filepath.Join(plaesyDir, "memory", "ctx.md")) {
		t.Error("denied .plaesy was removed anyway")
	}
	if !strings.Contains(out, "Failed to remove: .plaesy") {
		t.Errorf("stdout = %q, want the framework-dir failure line", out)
	}
	if !strings.Contains(errOut, "Failed to remove some items: .plaesy") {
		t.Errorf("stderr = %q, want the failure warning", errOut)
	}
}

func TestRemoveMappedDirFailure(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only Windows is probed for a removal denial")
	}
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	target := filepath.Join(root, "p")
	mappedDir := filepath.Join(target, "alpha", "instructions")
	mustWrite(t, filepath.Join(mappedDir, "i.md"), "i")
	if err := os.Chmod(mappedDir, 0o500); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(mappedDir, 0o755) })
	denyWrite(t, mappedDir)
	mustWrite(t, filepath.Join(target, "ALPHA.md"), "a")

	c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
	var removed int
	var removeErr error
	out, errOut := captureOutput(t, func() {
		removed, removeErr = c.Remove()
	})
	// A partial failure used to be downgraded to a warning and reported as
	// nil, so `plaesy clean` printed "completed!" and exited 0 with the
	// denied files still in place. The successful removals must still be
	// counted and reported; the failure must still reach the caller.
	if removeErr == nil {
		t.Error("Remove must return an error for a partial failure, got nil")
	}
	if removed != 1 {
		t.Errorf("Remove = %d, want 1 (only ALPHA.md)", removed)
	}
	if !exists(t, filepath.Join(mappedDir, "i.md")) {
		t.Error("denied mapped directory was removed anyway")
	}
	if !strings.Contains(out, "Failed to remove: alpha/instructions") {
		t.Errorf("stdout = %q, want the failure line", out)
	}
	if !strings.Contains(errOut, "Failed to remove some items: alpha/instructions") {
		t.Errorf("stderr = %q, want the failure warning", errOut)
	}
}

func TestNewEmptyTargetDirUsesCwd(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	c, err := New(Options{}, cfgPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// cleaner.go:66 turns an empty TargetDir into "." and then resolves it
	// against the process cwd.
	if c.Opts.TargetDir != filepath.Clean(wd) {
		t.Errorf("TargetDir = %q, want the cwd %q", c.Opts.TargetDir, wd)
	}
	if c.Opts.Level != LevelSafe {
		t.Errorf("Level = %q, want safe", c.Opts.Level)
	}
}

func TestRemoveIgnoresDryRunFlag(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	target := filepath.Join(root, "p")
	mustWrite(t, filepath.Join(target, "ALPHA.md"), "a")
	mustMkdir(t, filepath.Join(target, "specs"))

	c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
	c.Opts.DryRun = true
	captureOutput(t, func() {
		if _, err := c.Remove(); err != nil {
			t.Fatalf("Remove: %v", err)
		}
	})
	// The dry-run guarantee lives in cmd/plaesy/clean.go:162, not in the
	// package: Remove deletes even with DryRun set.
	if exists(t, filepath.Join(target, "ALPHA.md")) {
		t.Error("ALPHA.md survived a direct Remove call with DryRun=true")
	}
}

func TestDryRunLeavesEverythingUntouched(t *testing.T) {
	root := t.TempDir()
	cfgPath := filepath.Join(root, "claude.json")
	mustWrite(t, cfgPath, claudeConfig)
	target := filepath.Join(root, "p")
	claudeProject(t, target)

	// The dry-run flow as cmd/plaesy/clean.go performs it.
	c := newCleaner(t, target, cfgPath, LevelThorough, "claude_code")
	c.Opts.DryRun = true
	before := snapshot(t, target)

	if err := c.ValidateEnvironment(); err != nil {
		t.Fatalf("ValidateEnvironment: %v", err)
	}
	plan := c.BuildPlan()
	captureOutput(t, func() { c.PrintPlan(plan) })
	if plan.Empty {
		t.Fatal("plan is empty, nothing would be tested")
	}
	// clean.go:156 and clean.go:162 both gate on !DryRun.
	if c.Opts.Backup && !c.Opts.DryRun {
		if _, err := c.CreateBackup(); err != nil {
			t.Fatalf("CreateBackup: %v", err)
		}
	}
	if !c.Opts.DryRun {
		if _, err := c.Remove(); err != nil {
			t.Fatalf("Remove: %v", err)
		}
	}

	if after := snapshot(t, target); !reflect.DeepEqual(after, before) {
		t.Errorf("dry run modified the tree:\nbefore=%v\nafter=%v", before, after)
	}
	if got := backupDirs(t, target); len(got) != 0 {
		t.Errorf("dry run created backup directories: %v", got)
	}
}

func TestBackupThenRemoveRoundTrip(t *testing.T) {
	root := t.TempDir()
	cfgPath := filepath.Join(root, "claude.json")
	mustWrite(t, cfgPath, claudeConfig)
	target := filepath.Join(root, "p")
	claudeProject(t, target)

	c := newCleaner(t, target, cfgPath, LevelComplete, "claude_code")
	dir, err := c.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}
	captureOutput(t, func() {
		if _, err := c.Remove(); err != nil {
			t.Fatalf("Remove: %v", err)
		}
	})

	assertContent(t, filepath.Join(dir, "CLAUDE.md"), "claude core")
	assertContent(t, filepath.Join(dir, ".claude", "commands", "plan.md"), "plan cmd")
	assertContent(t, filepath.Join(dir, ".plaesy", "memory", "ctx.md"), "ctx")
	if exists(t, filepath.Join(target, "CLAUDE.md")) {
		t.Error("original survived the removal")
	}
	if exists(t, filepath.Join(target, ".plaesy")) {
		t.Error(".plaesy survived the removal")
	}
	if exists(t, filepath.Join(target, "specs")) {
		t.Error("specs survived the removal")
	}
	// The backup directory itself is not a mapping target, so it survives.
	if !exists(t, dir) {
		t.Error("backup directory was removed by Remove")
	}
	if !exists(t, filepath.Join(target, "src", "main.go")) {
		t.Error("src/ was removed")
	}
}

func TestBackupIsNotCleanedOnSecondRun(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	target := filepath.Join(root, "p")
	mustWrite(t, filepath.Join(target, ".plaesy", "x.md"), "x")

	c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
	dir, err := c.CreateBackup()
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}
	captureOutput(t, func() {
		if _, err := c.Remove(); err != nil {
			t.Fatalf("Remove: %v", err)
		}
	})
	if !exists(t, filepath.Join(dir, ".plaesy", "x.md")) {
		t.Error("backup contents were removed along with .plaesy")
	}
}

func TestUniqueSorted(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"nil", nil, nil},
		{"empty", []string{}, nil},
		{"dedupes and sorts", []string{"b", "a", "b", "c", "a"}, []string{"a", "b", "c"}},
		{"preserves nothing", []string{"z"}, []string{"z"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := uniqueSorted(tc.in); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("uniqueSorted(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src.txt")
	mustWrite(t, src, "hello")
	dst := filepath.Join(root, "nested", "dst.txt")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}
	assertContent(t, dst, "hello")

	t.Run("missing source", func(t *testing.T) {
		if err := copyFile(filepath.Join(root, "nope.txt"), filepath.Join(root, "out.txt")); err == nil {
			t.Error("copyFile succeeded for a missing source")
		}
	})

	t.Run("mode is carried over", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows has no POSIX permission bits")
		}
		ro := filepath.Join(root, "ro.txt")
		mustWrite(t, ro, "x")
		if err := os.Chmod(ro, 0o400); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		out := filepath.Join(root, "ro-copy.txt")
		if err := copyFile(ro, out); err != nil {
			t.Fatalf("copyFile: %v", err)
		}
		info, err := os.Stat(out)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if info.Mode().Perm() != 0o400 {
			t.Errorf("copied mode = %v, want 0400", info.Mode().Perm())
		}
	})
}

func TestCopyTree(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	mustWrite(t, filepath.Join(src, "a.txt"), "a")
	mustWrite(t, filepath.Join(src, "deep", "deeper", "b.txt"), "b")
	mustMkdir(t, filepath.Join(src, "empty"))

	dst := filepath.Join(root, "dst")
	if err := copyTree(src, dst); err != nil {
		t.Fatalf("copyTree: %v", err)
	}
	assertContent(t, filepath.Join(dst, "a.txt"), "a")
	assertContent(t, filepath.Join(dst, "deep", "deeper", "b.txt"), "b")
	if !exists(t, filepath.Join(dst, "empty")) {
		t.Error("empty source directory was not recreated")
	}

	t.Run("missing source", func(t *testing.T) {
		if err := copyTree(filepath.Join(root, "nope"), filepath.Join(root, "out")); err == nil {
			t.Error("copyTree succeeded for a missing source")
		}
	})
}

// loadFixture parses a config file for the package-level functions that take a
// *config.PlatformConfig.
func loadFixture(t *testing.T, path string) *config.PlatformConfig {
	t.Helper()
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	return cfg
}

// A detection marker that cannot be stat'd is not the same fact as a marker
// that is absent, and the difference decides which files get deleted.
//
// The old code dropped the platform on any stat error, fell out of the loop
// with nothing found, and returned the "generic_ai" fallback — a wrong
// deletion set, reported as detection.
//
// The unreadable-parent is a regular file rather than a directory, so stat on
// the child returns ENOTDIR. That is a different errno from ENOENT, which is
// what makes os.IsNotExist report false for it, and it happens identically on
// every platform — unlike a permission-denied parent, which root ignores and
// Windows does not model the same way. This is the same shape as the
// ".plaesy/specs replaced by a file" case featurepath/list.go already handles.
func TestDetectAllPlatformsReportsAnUnreadableMarkerInsteadOfFallingBack(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Windows reports this as ERROR_PATH_NOT_FOUND, which Go surfaces as
		// ENOENT, so os.IsNotExist reports true and the case is genuinely
		// indistinguishable from an absent marker there. Asserting otherwise
		// would be asserting a platform behaviour that does not exist; the
		// class is listed in the Windows bug notes for that reason.
		t.Skip("a blocking file is reported as PATH_NOT_FOUND on Windows, so the " +
			"unreadable-marker case cannot be provoked there")
	}
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)
	cfg := loadFixture(t, cfgPath)

	target := filepath.Join(root, "project")
	mustMkdir(t, target)
	// "alpha/marker" is the first detection pattern. Make "alpha" a file so
	// the stat of "alpha/marker" fails with something other than "not found".
	if err := os.WriteFile(filepath.Join(target, "alpha"), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}

	got, err := DetectAllPlatforms(cfg, cfgPath, target)
	if err == nil {
		t.Fatalf("DetectAllPlatforms = %v with no error; an unreadable marker "+
			"must be reported, not silently folded into the generic_ai fallback "+
			"that decides the deletion set", got)
	}
	if !strings.Contains(err.Error(), "alpha") {
		t.Errorf("error %q does not name the platform or marker it could not read", err)
	}
}

// The framework-dir loop used to be `if info, err := os.Stat(full); err == nil
// && info.IsDir()`, which cannot tell "not there" from "there but not a
// directory" and skipped both. The switch keeps them apart, and the not-a-
// directory case is reachable from a test: plaesyDirs[0] is a regular file.
func TestRemoveSkipsAFrameworkPathThatIsNotADirectory(t *testing.T) {
	root := t.TempDir()
	cfgPath := writeFixtureConfig(t, root)

	target := filepath.Join(root, "project")
	mustMkdir(t, target)
	// A regular file where the first framework directory is expected.
	mustWrite(t, filepath.Join(target, plaesyDirs[0]), "not a directory")

	c := newCleaner(t, target, cfgPath, LevelSafe, "alpha")
	removed, err := c.Remove()
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if removed != 0 {
		t.Errorf("Remove reported %d removals, want 0: a regular file is not a "+
			"framework directory and must not be counted as one", removed)
	}
	if _, err := os.Stat(filepath.Join(target, plaesyDirs[0])); err != nil {
		t.Errorf("Remove deleted %s, which is a file it should have skipped: %v",
			plaesyDirs[0], err)
	}
}
