package trimmer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// segment is one prose-or-code span of a file, split at ```fence lines
// (each fence line trails the segment that precedes it), mirroring the
// implicit segmentation walked by both cmd_llm_queue() and cmd_apply_llm()
// in the bash script. index is that segment's position in the walk (0-based,
// counting both prose and code segments) - the same index space
// llm-queue's "index" field and apply-llm's lookup both use.
type segment struct {
	index    int
	text     string
	wasProse bool // in_code state *before* this segment's trailing fence toggled it
}

// splitSegments walks the lines of text exactly as the bash script's
// line-at-a-time loops do: a segment ends (and a new one begins) right
// after any line matching ^```, with the fence line itself included in the
// segment that precedes it. A final trailing segment (possibly empty) is
// always emitted for whatever follows the last fence, matching the
// unconditional flush after each bash while-loop.
func splitSegments(text string) []segment {
	var segments []segment
	var buf strings.Builder
	inCode := false
	idx := 0

	lines := splitLinesKeepTrailing(text)
	for _, line := range lines {
		buf.WriteString(line)
		buf.WriteString("\n")
		if fenceRE.MatchString(line) {
			segments = append(segments, segment{index: idx, text: buf.String(), wasProse: !inCode})
			buf.Reset()
			inCode = !inCode
			idx++
		}
	}
	segments = append(segments, segment{index: idx, text: buf.String(), wasProse: !inCode})
	return segments
}

// llmQueueFile is the {file, segments:[{index,text}]} shape written by
// WriteLLMQueue and read back by ApplyLLM.
type llmQueueFile struct {
	File     string       `json:"file"`
	Segments []llmSegment `json:"segments"`
}
type llmSegment struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

// WriteLLMQueue mirrors cmd_llm_queue(): extracts every prose segment of
// at least 40 characters from the file at path into
// "<repoRoot>/.plaesy/analysis/trim-queue.json" for the calling assistant to
// rewrite, and returns the queue path plus how many segments were written.
func WriteLLMQueue(repoRoot, path string) (queuePath string, count int, err error) {
	target := path
	if _, statErr := os.Stat(target); statErr != nil {
		target = filepath.Join(repoRoot, path)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return "", 0, fmt.Errorf("path not found: %s", path)
	}
	relative, relErr := filepath.Rel(repoRoot, target)
	if relErr != nil || strings.HasPrefix(relative, "..") {
		relative = target
	}

	var entries []llmSegment
	for _, seg := range splitSegments(string(data)) {
		if seg.wasProse && len(seg.text) >= 40 {
			entries = append(entries, llmSegment{Index: seg.index, Text: seg.text})
		}
	}

	outDir := filepath.Join(repoRoot, ".plaesy", "analysis")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", 0, err
	}
	queuePath = filepath.Join(outDir, "trim-queue.json")
	out, err := json.MarshalIndent(llmQueueFile{File: relative, Segments: entries}, "", "  ")
	if err != nil {
		return "", 0, err
	}
	if err := os.WriteFile(queuePath, append(out, '\n'), 0o644); err != nil {
		return "", 0, err
	}
	return queuePath, len(entries), nil
}

// ApplyLLMResult mirrors the two summary lines printed by cmd_apply_llm().
type ApplyLLMResult struct {
	Relative string
	Before   int
	After    int
}

// ApplyLLM mirrors cmd_apply_llm(): merges the {file, segments} rewrites
// from annotationsPath back into the file at path (matching each rewrite to
// the original prose segment by index), backs the original up to
// "<file>.bak", writes the merged result, and records the before/after
// token counts to statsPath.
func ApplyLLM(repoRoot, statsPath, path, annotationsPath string) (ApplyLLMResult, error) {
	target := path
	if _, statErr := os.Stat(target); statErr != nil {
		target = filepath.Join(repoRoot, path)
	}
	original, err := os.ReadFile(target)
	if err != nil {
		return ApplyLLMResult{}, fmt.Errorf("path not found: %s", path)
	}
	annData, err := os.ReadFile(annotationsPath)
	if err != nil {
		return ApplyLLMResult{}, fmt.Errorf("annotations file not found: %s", annotationsPath)
	}
	var queue llmQueueFile
	if err := json.Unmarshal(annData, &queue); err != nil {
		return ApplyLLMResult{}, fmt.Errorf("invalid annotations JSON: %w", err)
	}
	rewrites := make(map[int]string, len(queue.Segments))
	for _, s := range queue.Segments {
		if _, exists := rewrites[s.Index]; !exists { // first match wins, matching bash's `head -1`
			rewrites[s.Index] = s.Text
		}
	}

	relative, relErr := filepath.Rel(repoRoot, target)
	if relErr != nil || strings.HasPrefix(relative, "..") {
		relative = target
	}

	var out strings.Builder
	for _, seg := range splitSegments(string(original)) {
		if seg.wasProse {
			if rewrite, ok := rewrites[seg.index]; ok {
				out.WriteString(rewrite)
				continue
			}
		}
		out.WriteString(seg.text)
	}

	before := EstimateTokens(string(original))
	after := EstimateTokens(out.String())

	if err := os.WriteFile(target+".bak", original, 0o644); err != nil {
		return ApplyLLMResult{}, err
	}
	if err := os.WriteFile(target, []byte(out.String()), 0o644); err != nil {
		return ApplyLLMResult{}, err
	}
	if statsPath != "" {
		if err := AddStatRecord(statsPath, "layer2-llm", relative, before, after); err != nil {
			return ApplyLLMResult{}, err
		}
	}

	return ApplyLLMResult{Relative: relative, Before: before, After: after}, nil
}
