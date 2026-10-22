package eval

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGoldenDocsLabels checks that every labelled file exists and contains the
// keyword recorded for its query. A label that points at a file which does not
// discuss the query would make the evaluation measure nothing useful.
func TestGoldenDocsLabels(t *testing.T) {
	// Tests run in scripts/internal/search/eval, four levels below the repo root.
	repoRoot := filepath.Join("..", "..", "..", "..")
	goldenPath := filepath.Join(repoRoot, "scripts", "testdata", "golden_docs.jsonl")

	f, err := os.Open(goldenPath)
	if err != nil {
		t.Skipf("golden set not found: %v", err)
	}
	defer f.Close()

	type label struct {
		ID          string            `json:"id"`
		ExpectedIDs []string          `json:"expected_ids"`
		Metadata    map[string]string `json:"metadata"`
	}

	var checked, paraphrase, indonesian int
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var l label
		if err := json.Unmarshal([]byte(line), &l); err != nil {
			t.Fatalf("parse label line %q: %v", line, err)
		}
		if len(l.ExpectedIDs) != 1 {
			t.Errorf("%s: want exactly one labelled file, got %d", l.ID, len(l.ExpectedIDs))
			continue
		}
		keyword := l.Metadata["keyword"]
		if keyword == "" {
			t.Errorf("%s: missing metadata.keyword", l.ID)
			continue
		}

		body, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(l.ExpectedIDs[0])))
		if err != nil {
			t.Errorf("%s: labelled file %s: %v", l.ID, l.ExpectedIDs[0], err)
			continue
		}
		if !strings.Contains(strings.ToLower(string(body)), strings.ToLower(keyword)) {
			t.Errorf("%s: keyword %q not found in %s", l.ID, keyword, l.ExpectedIDs[0])
		}
		checked++
		if l.Metadata["style"] == "paraphrase" {
			paraphrase++
		}
		if l.Metadata["lang"] == "id" {
			indonesian++
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("read golden set: %v", err)
	}
	// Floors from the expansion task: enough labels to separate a 0.05 MRR
	// difference from noise, with paraphrases and Indonesian queries included.
	if checked < 100 {
		t.Errorf("only %d labels checked; the golden set should have at least 100", checked)
	}
	if paraphrase < 20 {
		t.Errorf("only %d paraphrase labels; want at least 20", paraphrase)
	}
	if indonesian < 10 {
		t.Errorf("only %d Indonesian labels; want at least 10", indonesian)
	}
}
