package reach

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// The adapters shell out to curl, gh, yt-dlp and similar tools. Tests must not
// reach the network or a real tool, so each name is a copy of this test binary
// placed first on PATH. When a copy runs, TestMain sees that its name is not
// the test binary, and acts as the tool: it prints REACH_FAKE_OUT, exits with
// REACH_FAKE_EXIT, and appends its argv to REACH_FAKE_LOG as one JSON line.

// fakeCommandNames lists every external command an adapter can invoke.
var fakeCommandNames = []string{
	"curl", "gh", "git", "opencli", "bili", "bili-cli", "yt-dlp",
	"youtube-transcript-api", "feedparser", "rdt-cli", "bird", "twitter-cli",
	"xhs-cli", "xiaohongshu-mcp", "mcp-linkedin", "exa-mcp", "boss-cdp", "boss",
	"xiaoyuzhou-mcp", "npx", "uvx", "mcp-server", "mcporter", "python3",
	"rdt", "twitter", "xhs",
}

// fakeBinDir holds the copies, set up once by TestMain.
var fakeBinDir string

func TestMain(m *testing.M) {
	if name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe"); !strings.HasSuffix(name, ".test") {
		os.Exit(runFakeCommand(name, os.Args[1:]))
	}

	dir, err := os.MkdirTemp("", "reach-fakebin")
	if err != nil {
		fmt.Fprintln(os.Stderr, "reach tests: creating fake bin dir:", err)
		os.Exit(1)
	}
	fakeBinDir = dir

	self, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "reach tests: locating test binary:", err)
		os.Exit(1)
	}
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	for _, name := range fakeCommandNames {
		if err := copyFile(self, filepath.Join(dir, name+ext)); err != nil {
			fmt.Fprintln(os.Stderr, "reach tests: copying fake", name, err)
			os.Exit(1)
		}
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// runFakeCommand behaves as the named external tool.
func runFakeCommand(name string, args []string) int {
	if logPath := os.Getenv("REACH_FAKE_LOG"); logPath != "" {
		line, _ := json.Marshal(map[string]any{"cmd": name, "args": args})
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err == nil {
			f.Write(append(line, '\n'))
			f.Close()
		}
	}
	if path := os.Getenv("REACH_FAKE_STDIN"); path != "" {
		data, _ := io.ReadAll(os.Stdin)
		os.WriteFile(path, data, 0o600)
	}
	fmt.Fprint(os.Stdout, os.Getenv("REACH_FAKE_OUT"))
	if raw := os.Getenv("REACH_FAKE_ERR"); raw != "" {
		fmt.Fprint(os.Stderr, raw)
	}
	code, err := strconv.Atoi(os.Getenv("REACH_FAKE_EXIT"))
	if err != nil {
		return 0
	}
	return code
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// fakeTool makes every external command succeed with stdout and returns the
// path of the call log. Tests read the log to check the argv an adapter built.
func fakeTool(t *testing.T, stdout string) string {
	t.Helper()
	return fakeToolExit(t, stdout, 0)
}

// fakeToolExit is fakeTool with a chosen exit code, so failure paths run too.
func fakeToolExit(t *testing.T, stdout string, exit int) string {
	t.Helper()
	logPath := filepath.Join(t.TempDir(), "calls.jsonl")
	t.Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("REACH_FAKE_OUT", stdout)
	t.Setenv("REACH_FAKE_ERR", "")
	t.Setenv("REACH_FAKE_EXIT", strconv.Itoa(exit))
	t.Setenv("REACH_FAKE_LOG", logPath)
	return logPath
}

// noTools points PATH at an empty directory, so every external command is
// missing. It exercises the branch where a backend is not installed.
func noTools(t *testing.T) {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
}

// calls reads the call log written by the fake tools.
func calls(t *testing.T, logPath string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatal(err)
	}
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var c map[string]any
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			t.Fatalf("bad call log line %q: %v", line, err)
		}
		out = append(out, c)
	}
	return out
}
