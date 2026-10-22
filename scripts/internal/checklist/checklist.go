// Package checklist reads and runs the contracts.json checklist of a task. Each item
// has a run command that proves it and an expected value its output must reach. Its status is
// pending until it first passes, passed while its last run passed, and failed while its
// last run failed. Every failed run is kept in the item's histories. A task is done
// when every item is passed.
package checklist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/plaesy/spec-kit/internal/expected"
)

// Status values an item can have.
const (
	StatusPending = "pending"
	StatusPassed  = "passed"
	StatusFailed  = "failed"
)

// History is one failed run of an item: when it failed, why, and the command that ran.
type History struct {
	At     string `json:"at"`
	Actual string `json:"actual"`
	Run    string `json:"run"`
}

// Item is one condition of the task.
type Item struct {
	ID        string    `json:"id"`
	Check     string    `json:"check"`
	Run       string    `json:"run"`
	Expected  string    `json:"expected"`
	Status    string    `json:"status"`
	Histories []History `json:"histories"`
}

// File is the contents of a task's contracts.json.
type File struct {
	ID        string `json:"id"`
	Checklist []Item `json:"checklist"`
}

// Runner runs one command and reports its exit code and standard output.
type Runner func(run string) (exit int, stdout string, err error)

// Load reads a contracts.json. Unknown fields are an error, so a misspelt key cannot
// silently drop a check.
func Load(path string) (File, error) {
	var f File
	raw, err := os.ReadFile(path)
	if err != nil {
		return f, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return f, fmt.Errorf("%s: %w", path, err)
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		return f, fmt.Errorf("%s: content after the checklist object", path)
	}
	for _, it := range f.Checklist {
		if it.ID == "" || it.Run == "" || it.Expected == "" {
			return f, fmt.Errorf("%s: item %q needs id, run and expected", path, it.ID)
		}
	}
	return f, nil
}

// Save writes the checklist back, replacing the file in one step.
func Save(path string, f File) error {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(filepath.Dir(path), filepath.Base(path)+".tmp")
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Run runs every item's command and sets its status. A failed run is appended to the
// item's histories with its message. A passed run sets the status without history.
func Run(f *File, run Runner, now time.Time) {
	for i := range f.Checklist {
		it := &f.Checklist[i]
		msg, ok := runOne(*it, run)
		if ok {
			it.Status = StatusPassed
			continue
		}
		it.Status = StatusFailed
		it.Histories = append(it.Histories, History{At: now.UTC().Format(time.RFC3339), Actual: msg, Run: it.Run})
	}
}

// runOne returns the failure message and false, or an empty message and true.
func runOne(it Item, run Runner) (string, bool) {
	exit, out, err := run(it.Run)
	if err != nil {
		return err.Error(), false
	}
	if exit != 0 {
		return fmt.Sprintf("exit code %d", exit), false
	}
	result, _, reason := expected.Judge(it.Expected, lastLine(out))
	if result == expected.ResultMet {
		return "", true
	}
	if reason == "" {
		reason = fmt.Sprintf("result %s for expected %q", result, it.Expected)
	}
	return reason, false
}

// Done reports whether every item of the checklist is passed.
func Done(f File) bool {
	if len(f.Checklist) == 0 {
		return false
	}
	for _, it := range f.Checklist {
		if it.Status != StatusPassed {
			return false
		}
	}
	return true
}

// Progress returns how many items are passed, and how many items there are in all.
func Progress(f File) (passed, total int) {
	for _, it := range f.Checklist {
		if it.Status == StatusPassed {
			passed++
		}
	}
	return passed, len(f.Checklist)
}

func lastLine(out string) string {
	lines := strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if s := strings.TrimSpace(lines[i]); s != "" {
			return s
		}
	}
	return ""
}
