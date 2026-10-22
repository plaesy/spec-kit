package checklist

import (
	"errors"
	"testing"
	"time"
)

var fixedNow = time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)

func TestRunKeepsFailuresInHistories(t *testing.T) {
	f := File{ID: "t", Checklist: []Item{
		{ID: "C1", Run: "ok", Expected: "pass"},
		{ID: "C2", Run: "count", Expected: ">= 3"},
		{ID: "C3", Run: "broken", Expected: "pass"},
	}}
	run := func(run string) (int, string, error) {
		switch run {
		case "ok":
			return 0, "", nil
		case "count":
			return 0, "5\n", nil
		case "broken":
			return 0, "", errors.New("no shell")
		}
		return 1, "", nil
	}

	Run(&f, run, fixedNow)

	if f.Checklist[0].Status != StatusPassed || f.Checklist[1].Status != StatusPassed {
		t.Fatalf("C1 and C2 must be passed: %+v", f.Checklist)
	}
	c3 := f.Checklist[2]
	if c3.Status != StatusFailed || len(c3.Histories) != 1 {
		t.Fatalf("C3 must be failed with one history: %+v", c3)
	}
	h := c3.Histories[0]
	if h.Actual != "no shell" || h.Run != "broken" || h.At != "2026-10-22T12:00:00Z" {
		t.Errorf("history = %+v", h)
	}
	if Done(f) {
		t.Errorf("a task with a failed item must not be done")
	}
}

func TestFailuresAccumulateAndDoneWhenAllPass(t *testing.T) {
	f := File{ID: "t", Checklist: []Item{{ID: "C1", Run: "x", Expected: "pass"}}}
	fail := func(string) (int, string, error) { return 1, "", nil }
	pass := func(string) (int, string, error) { return 0, "", nil }

	Run(&f, fail, fixedNow)
	Run(&f, fail, fixedNow.Add(time.Hour))
	if len(f.Checklist[0].Histories) != 2 {
		t.Fatalf("two failed runs must leave two histories: %+v", f.Checklist[0].Histories)
	}
	Run(&f, pass, fixedNow.Add(2*time.Hour))
	if !Done(f) {
		t.Fatalf("after the check passes the task must be done")
	}
	if len(f.Checklist[0].Histories) != 2 {
		t.Errorf("a pass must not erase earlier failures")
	}
	if passed, total := Progress(f); passed != 1 || total != 1 {
		t.Errorf("progress = %d/%d", passed, total)
	}
}
