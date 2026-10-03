package trimmer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// splitSegments decides which spans a rewrite is allowed to touch, so its
// index space is the contract between WriteLLMQueue and ApplyLLM. The fence line
// belongs to the segment *before* it, and a final segment is always emitted.
func TestSplitSegments(t *testing.T) {
	doc := strings.Join([]string{
		"intro prose that is long enough to matter",
		"",
		"```go",
		"code := 1",
		"```",
		"",
		"outro prose",
		"",
	}, "\n")
	segs := splitSegments(doc)
	if len(segs) != 3 {
		t.Fatalf("got %d segments, want 3 (prose, code, prose):\n%#v", len(segs), segs)
	}
	if segs[0].index != 0 || !segs[0].wasProse {
		t.Errorf("segment 0 = %+v, want index 0 and wasProse", segs[0])
	}
	if !strings.HasSuffix(segs[0].text, "```go\n") {
		t.Errorf("the opening fence belongs to the segment before it: %q", segs[0].text)
	}
	if segs[1].wasProse {
		t.Error("the code segment must not be marked as prose")
	}
	if segs[1].index != 1 || segs[2].index != 2 {
		t.Errorf("indices = %d, %d, want 1 and 2", segs[1].index, segs[2].index)
	}
	if !segs[2].wasProse || !strings.Contains(segs[2].text, "outro prose") {
		t.Errorf("trailing segment = %+v", segs[2])
	}
}

func TestSplitSegmentsOnTextWithNoFence(t *testing.T) {
	segs := splitSegments("just prose\n")
	if len(segs) != 1 || !segs[0].wasProse || segs[0].text != "just prose\n" {
		t.Errorf("got %#v, want a single prose segment holding the whole text", segs)
	}
}

func TestWriteLLMQueueKeepsLongProseOnly(t *testing.T) {
	repo := t.TempDir()
	doc := filepath.Join(repo, "doc.md")
	short := "tiny\n"
	long := "This paragraph is comfortably longer than the forty character floor for queueing.\n"
	writeFile(t, doc, "```\ncode that must not be queued for rewriting at all\n```\n"+short+long)

	queuePath, count, err := WriteLLMQueue(repo, doc)
	if err != nil {
		t.Fatalf("WriteLLMQueue: %v", err)
	}
	if count != 1 {
		t.Errorf("queued %d segments, want 1 (the long prose one)", count)
	}
	if filepath.Base(queuePath) != "trim-queue.json" {
		t.Errorf("queue path = %q", queuePath)
	}
	var q llmQueueFile
	data, err := os.ReadFile(queuePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &q); err != nil {
		t.Fatalf("parse queue: %v", err)
	}
	if q.File != "doc.md" {
		t.Errorf("queue file field = %q, want the repo-relative name", q.File)
	}
	if len(q.Segments) != 1 {
		t.Fatalf("queue holds %d segments: %+v", len(q.Segments), q.Segments)
	}
	if !strings.Contains(q.Segments[0].Text, "forty character floor") {
		t.Errorf("queued the wrong text: %q", q.Segments[0].Text)
	}
	if strings.Contains(q.Segments[0].Text, "code that must not be queued") {
		t.Error("a code segment reached the rewrite queue")
	}
}

func TestWriteLLMQueueAcceptsARepoRelativePath(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "docs", "a.md"),
		"A paragraph long enough to clear the queue floor for rewriting.\n")
	_, count, err := WriteLLMQueue(repo, filepath.Join("docs", "a.md"))
	if err != nil {
		t.Fatalf("WriteLLMQueue: %v", err)
	}
	if count != 1 {
		t.Errorf("queued %d segments, want 1", count)
	}
}

func TestWriteLLMQueueOnAMissingFile(t *testing.T) {
	_, _, err := WriteLLMQueue(t.TempDir(), "absent.md")
	if err == nil {
		t.Error("queueing a file that does not exist must be an error")
	}
}

func TestApplyLLMMergesRewritesAndBacksUp(t *testing.T) {
	repo := t.TempDir()
	doc := filepath.Join(repo, "doc.md")
	original := "Prose paragraph one that the assistant rewrites.\n```\nkeep this code\n```\nProse paragraph two that stays.\n"
	writeFile(t, doc, original)
	stats := filepath.Join(repo, "stats.json")

	annotations := filepath.Join(repo, "annotations.json")
	writeFile(t, annotations, `{"file":"doc.md","segments":[{"index":0,"text":"Rewritten opening.\n"}]}`)

	res, err := ApplyLLM(repo, stats, doc, annotations)
	if err != nil {
		t.Fatalf("ApplyLLM: %v", err)
	}
	if res.Relative != "doc.md" {
		t.Errorf("relative = %q", res.Relative)
	}
	backup, err := os.ReadFile(doc + ".bak")
	if err != nil {
		t.Fatalf("no backup: %v", err)
	}
	if string(backup) != original {
		t.Errorf("backup = %q, want the original", backup)
	}
	merged, err := os.ReadFile(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Rewritten opening.", "keep this code", "Prose paragraph two that stays."} {
		if !strings.Contains(string(merged), want) {
			t.Errorf("merged file is missing %q:\n%s", want, merged)
		}
	}
	if strings.Contains(string(merged), "paragraph one") {
		t.Errorf("the rewritten segment survived:\n%s", merged)
	}
	records := readStats(t, stats)
	if len(records) != 1 || records[0].Layer != "layer2-llm" {
		t.Errorf("stats = %+v, want one layer2-llm record", records)
	}
}

// A rewrite indexed onto a code segment is ignored: the assistant is supposed to
// be given prose only, and honouring a code index would let a malformed
// annotation file rewrite a code block.
func TestApplyLLMIgnoresARewriteIndexedOntoCode(t *testing.T) {
	repo := t.TempDir()
	doc := filepath.Join(repo, "doc.md")
	writeFile(t, doc, "Prose.\n```\noriginal code\n```\n")
	annotations := filepath.Join(repo, "annotations.json")
	writeFile(t, annotations, `{"segments":[{"index":1,"text":"rewritten code\n"}]}`)

	if _, err := ApplyLLM(repo, "", doc, annotations); err != nil {
		t.Fatalf("ApplyLLM: %v", err)
	}
	got, err := os.ReadFile(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "original code") {
		t.Errorf("a code segment was rewritten:\n%s", got)
	}
}

func TestApplyLLMFirstRewriteForAnIndexWins(t *testing.T) {
	repo := t.TempDir()
	doc := filepath.Join(repo, "doc.md")
	writeFile(t, doc, "Prose segment.\n")
	annotations := filepath.Join(repo, "annotations.json")
	writeFile(t, annotations, `{"segments":[{"index":0,"text":"first\n"},{"index":0,"text":"second\n"}]}`)

	if _, err := ApplyLLM(repo, "", doc, annotations); err != nil {
		t.Fatalf("ApplyLLM: %v", err)
	}
	got, err := os.ReadFile(doc)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(got)) != "first" {
		t.Errorf("got %q, want the first rewrite for a duplicated index", got)
	}
}

func TestApplyLLMErrors(t *testing.T) {
	for _, tc := range []struct {
		name        string
		missingDoc  bool
		annotations string
		writeAnn    bool
	}{
		{name: "the file does not exist", missingDoc: true, annotations: "{}", writeAnn: true},
		{name: "the annotations file does not exist"},
		{name: "the annotations are not json", annotations: "{oops", writeAnn: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// A fresh directory per case: "the annotations file does not exist"
			// is only meaningful if no earlier case left one behind.
			repo := t.TempDir()
			doc := filepath.Join(repo, "doc.md")
			if !tc.missingDoc {
				writeFile(t, doc, "Prose.\n")
			}
			ann := filepath.Join(repo, "annotations.json")
			if tc.writeAnn {
				writeFile(t, ann, tc.annotations)
			}
			path := doc
			if tc.missingDoc {
				path = "absent.md"
			}
			if _, err := ApplyLLM(repo, "", path, ann); err == nil {
				t.Error("want an error")
			}
		})
	}
}

// The queue is written to one fixed path for the whole project, so the real
// workflow is: queue file A, an assistant rewrites the JSON, then apply it to
// some --path. The rewrites are keyed only by segment index, so applying one
// file's answer to another overwrote unrelated prose wherever the indexes
// happened to line up — and reported a successful apply. The queue records
// which file it came from; nothing read that field.
func TestApplyLLMRefusesAnnotationsQueuedForAnotherFile(t *testing.T) {
	repo := t.TempDir()
	queued := filepath.Join(repo, "docs", "a.md")
	other := filepath.Join(repo, "docs", "b.md")
	long := "This paragraph is comfortably longer than the forty character floor for queueing.\n"
	writeFile(t, queued, long)
	writeFile(t, other, long)

	queuePath, count, err := WriteLLMQueue(repo, queued)
	if err != nil {
		t.Fatalf("WriteLLMQueue: %v", err)
	}
	if count == 0 {
		t.Fatal("fixture produced no queued segments, so the test proves nothing")
	}
	// The assistant's answer to a.md, still carrying a's identity.
	annPath := filepath.Join(repo, "annotations.json")
	var queue llmQueueFile
	data, err := os.ReadFile(queuePath)
	if err != nil {
		t.Fatalf("read queue: %v", err)
	}
	if err := json.Unmarshal(data, &queue); err != nil {
		t.Fatalf("parse queue: %v", err)
	}
	queue.Segments = []llmSegment{{Index: 0, Text: "REWRITTEN FOR A\n"}}
	out, _ := json.Marshal(queue)
	writeFile(t, annPath, string(out))

	if _, err := ApplyLLM(repo, "", other, annPath); err == nil {
		t.Fatal("ApplyLLM applied one file's rewrites to another file")
	}
	got, _ := os.ReadFile(other)
	if strings.Contains(string(got), "REWRITTEN FOR A") {
		t.Errorf("the mismatched rewrite was spliced in anyway:\n%s", got)
	}
	if _, err := os.Stat(other + ".bak"); !os.IsNotExist(err) {
		t.Error("a refused apply must not write a .bak")
	}
}

// The same file reached two ways — once via an absolute path, once relative to
// the repo — is the same file and must still apply.
func TestApplyLLMAcceptsTheSameFileSpelledDifferently(t *testing.T) {
	repo := t.TempDir()
	doc := filepath.Join(repo, "docs", "a.md")
	long := "This paragraph is comfortably longer than the forty character floor for queueing.\n"
	writeFile(t, doc, long)

	queuePath, _, err := WriteLLMQueue(repo, doc)
	if err != nil {
		t.Fatalf("WriteLLMQueue: %v", err)
	}
	var queue llmQueueFile
	data, _ := os.ReadFile(queuePath)
	if err := json.Unmarshal(data, &queue); err != nil {
		t.Fatalf("parse queue: %v", err)
	}
	queue.Segments = []llmSegment{{Index: 0, Text: "REWRITTEN\n"}}
	ann, _ := json.Marshal(queue)
	annPath := filepath.Join(repo, "annotations.json")
	writeFile(t, annPath, string(ann))

	rel, err := filepath.Rel(repo, doc)
	if err != nil {
		t.Fatalf("rel: %v", err)
	}
	if rel == doc {
		t.Fatalf("fixture is not inside the repo, so the relative spelling is not a second spelling: %q", rel)
	}
	// Spell the same file the other way round: the queue holds the
	// repo-relative path, --path is given the absolute one. No chdir — that
	// makes t.TempDir cleanup fail on Windows.
	if _, err := ApplyLLM(repo, "", doc, annPath); err != nil {
		t.Fatalf("ApplyLLM on the same file spelled absolutely: %v", err)
	}
	_ = rel
	got, _ := os.ReadFile(doc)
	if !strings.Contains(string(got), "REWRITTEN") {
		t.Errorf("the rewrite was not applied:\n%s", got)
	}
}
