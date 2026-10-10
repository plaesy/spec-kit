package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureStdoutStderr redirects the process streams for the duration of fn and
// returns what was written to each. The logging helpers write straight to
// os.Stdout/os.Stderr rather than taking a writer, so swapping the streams is
// the only way to see what a caller sees. The targets are temp files rather
// than pipes: a pipe read returns whatever happens to have arrived, which
// truncates multi-line output like the banner.
func captureStdoutStderr(t *testing.T, fn func()) (stdout, stderr string) {
	t.Helper()
	outF, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatalf("temp stdout: %v", err)
	}
	errF, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatalf("temp stderr: %v", err)
	}
	prevOut, prevErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outF, errF
	func() {
		defer func() {
			os.Stdout, os.Stderr = prevOut, prevErr
			outF.Close()
			errF.Close()
		}()
		fn()
	}()

	read := func(f *os.File) string {
		data, err := os.ReadFile(f.Name())
		if err != nil {
			t.Fatalf("read %s: %v", f.Name(), err)
		}
		return string(data)
	}
	return read(outF), read(errF)
}

func TestLogLevelsGoToTheExpectedStream(t *testing.T) {
	cases := []struct {
		name       string
		call       func()
		wantStream string
		wantLevel  string
	}{
		{"info goes to stdout", func() { LogInfo("hello %s", "world") }, "stdout", "INFO"},
		{"success goes to stdout", func() { LogSuccess("done") }, "stdout", "SUCCESS"},
		{"warning goes to stderr", func() { LogWarning("careful") }, "stderr", "WARNING"},
		{"error goes to stderr", func() { LogError("broken: %v", 42) }, "stderr", "ERROR"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr := captureStdoutStderr(t, tc.call)
			got := stdout
			other := stderr
			if tc.wantStream == "stderr" {
				got, other = stderr, stdout
			}
			if !strings.Contains(got, tc.wantLevel) || !strings.Contains(got, "hello world") &&
				!strings.Contains(got, "done") && !strings.Contains(got, "careful") && !strings.Contains(got, "broken: 42") {
				t.Errorf("expected output missing from %s: %q", tc.wantStream, got)
			}
			if other != "" {
				t.Errorf("the other stream received output: %q", other)
			}
		})
	}
}

func TestLogDebugIsGatedByDebugMode(t *testing.T) {
	prev := DebugMode
	t.Cleanup(func() { DebugMode = prev })

	DebugMode = false
	stdout, stderr := captureStdoutStderr(t, func() { LogDebug("invisible %d", 1) })
	if stdout != "" || stderr != "" {
		t.Errorf("LogDebug printed with DebugMode off: %q / %q", stdout, stderr)
	}

	DebugMode = true
	stdout, _ = captureStdoutStderr(t, func() { LogDebug("visible %d", 1) })
	if !strings.Contains(stdout, "DEBUG") || !strings.Contains(stdout, "visible 1") {
		t.Errorf("LogDebug with DebugMode on = %q", stdout)
	}
}

func TestLogFileAppendsEveryLevel(t *testing.T) {
	prevFile, prevDebug := LogFile, DebugMode
	t.Cleanup(func() { LogFile, DebugMode = prevFile, prevDebug })

	logPath := filepath.Join(t.TempDir(), "nested", "plaesy.log")
	LogFile = logPath
	DebugMode = true

	captureStdoutStderr(t, func() {
		LogInfo("first")
		LogWarning("second")
		LogDebug("third")
		LogInfo("fourth")
	})

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("log has %d lines, want 4 (append, not truncate):\n%s", len(lines), data)
	}
	if !strings.Contains(lines[0], "first") || !strings.Contains(lines[3], "fourth") {
		t.Errorf("log content out of order:\n%s", data)
	}
	// The file line carries the same level tag and the process name as the
	// console line, so a log can be read on its own.
	if !strings.Contains(lines[1], "WARNING") || !strings.Contains(lines[2], "DEBUG") {
		t.Errorf("log lines missing their level:\n%s", data)
	}
	if !strings.Contains(lines[0], filepath.Base(os.Args[0])) {
		t.Errorf("log lines missing the program name:\n%s", data)
	}
}

func TestLogFileFailureDoesNotPanic(t *testing.T) {
	prev := LogFile
	t.Cleanup(func() { LogFile = prev })

	// A path that cannot be opened: a directory, not a file. Logging must not
	// take the command down over a diagnostic.
	LogFile = t.TempDir()
	captureStdoutStderr(t, func() { LogInfo("still printed") })
}

// TestLogRuntimeStringWithPercentIsNotCorrupted pins the calling convention
// these helpers require. The first parameter is a format string, so a message
// assembled at runtime must be passed as an argument to a constant "%s" —
// never as the format itself.
//
// Passing it directly looks harmless right up until the message contains a
// percent sign, which is legal in a filename on every OS plaesy targets. Then
// Sprintf consumes the following text as a verb and the user sees
// "report%!f(MISSING)inal.md" instead of the file that was just copied. Every
// one of the 52 call sites in internal/scaffold and cmd/plaesy interpolated a
// path this way; `go vet` reports them as "non-constant format string".
//
// This test covers the helper's half of the contract. The call sites' half is
// enforced by `go vet`, which `go test` already runs before compiling — a bad
// call site fails the build of its own package rather than passing quietly.
func TestLogRuntimeStringWithPercentIsNotCorrupted(t *testing.T) {
	// %20 is a width verb, and a trailing "100%" is a truncated one: both are
	// mangled by the old form, and both survive "%s" intact.
	const want = "  ✓ report%20final.md copied (100% done)"

	prevDebug := DebugMode
	DebugMode = true
	t.Cleanup(func() { DebugMode = prevDebug })

	for _, tc := range []struct {
		name string
		call func()
	}{
		{"info", func() { LogInfo("%s", want) }},
		{"success", func() { LogSuccess("%s", want) }},
		{"warning", func() { LogWarning("%s", want) }},
		{"error", func() { LogError("%s", want) }},
		{"debug", func() { LogDebug("%s", want) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr := captureStdoutStderr(t, tc.call)
			if !strings.Contains(stdout+stderr, want) {
				t.Errorf("message containing %% was mangled:\nstdout: %q\nstderr: %q", stdout, stderr)
			}
		})
	}
}

func TestPrintBanner(t *testing.T) {
	stdout, _ := captureStdoutStderr(t, func() { PrintBanner("Title", "Subtitle") })
	for _, want := range []string{purple, "█", "Title", "Subtitle", colorReset} {
		if !strings.Contains(stdout, want) {
			t.Errorf("banner missing %q:\n%s", want, stdout)
		}
	}
	if strings.Count(stdout, "Title") != 1 || strings.Count(stdout, "Subtitle") != 1 {
		t.Errorf("title or subtitle printed more than once:\n%s", stdout)
	}

	// Empty title and subtitle print the logo alone, with no stray blank label.
	stdout, _ = captureStdoutStderr(t, func() { PrintBanner("", "") })
	if strings.Contains(stdout, "🏛️") {
		t.Errorf("empty title still printed a label:\n%s", stdout)
	}
}
