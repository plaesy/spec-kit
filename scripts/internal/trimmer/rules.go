// Package trimmer is a Go port of scripts/bash/plaesy-trim.sh: token/context
// compression for the plaesy spec-kit repo (or any project it is run
// against). It offers three layers:
//
//	Layer 1 (command output): Run(...)                - per-tool aware
//	                                                      dedupe/truncate/filter
//	                                                      of a subprocess's output
//	Layer 2 (memory/instruction files): CompressPath(...) - heuristic denser
//	                                                          prose style (fast, local)
//	                             WriteLLMQueue / ApplyLLM  - LLM-quality rewrite
//	                                                          performed by the
//	                                                          calling assistant
//	                                                          (no API key needed)
//	Report: Report(...)                                - cumulative savings
//	                                                      across layers
//
// Only the Go standard library is used.
package trimmer

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Layer1Rule mirrors one entry of rules.json's "layer1_commands" object.
type Layer1Rule struct {
	MaxLines          int    `json:"max_lines"`
	DedupeConsecutive bool   `json:"dedupe_consecutive"`
	Mode              string `json:"mode,omitempty"`
}

// FillerPattern mirrors one entry of rules.json's "layer2_filler_patterns"
// array: a regex (Go RE2 syntax, same subset bash's `sed -E` accepts for
// these patterns) and its replacement.
type FillerPattern struct {
	Find    string `json:"find"`
	Replace string `json:"replace"`
}

// DegreeThresholds mirrors rules.json's "layer2_degree_thresholds".
type DegreeThresholds struct {
	LiteMinDegree int `json:"lite_min_degree"`
	FullMinDegree int `json:"full_min_degree"`
}

// Rules mirrors the full shape of configs/plaesy-trim-rules.json.
type Rules struct {
	Description           string                `json:"description"`
	Version               string                `json:"version"`
	Layer1Commands        map[string]Layer1Rule `json:"layer1_commands"`
	Layer2FillerPatterns  []FillerPattern       `json:"layer2_filler_patterns"`
	Layer2DegreeThreshold DegreeThresholds      `json:"layer2_degree_thresholds"`
}

// defaultLayer1Rule mirrors rules.json's "_default" entry.
const (
	defaultMaxLines = 40
)

// defaultRules is the built-in fallback, matching
// scripts/configs/plaesy-trim-rules.json byte-for-byte in content (not
// necessarily key order), used when no rules.json can be found on disk at
// all (neither a project override nor a framework-shipped copy). Having a
// baked-in fallback means the trim command always works even when installed
// standalone, unlike the bash script which hard-requires a rules.json file
// to exist relative to itself.
func defaultRules() *Rules {
	return &Rules{
		Description: "Rules for plaesy-trim: Layer 1 (command output) thresholds and Layer 2 (prose compression) filler patterns.",
		Version:     "1.0.0",
		Layer1Commands: map[string]Layer1Rule{
			"_default":     {MaxLines: defaultMaxLines, DedupeConsecutive: true},
			"git status":   {MaxLines: 25, DedupeConsecutive: true},
			"git push":     {MaxLines: 10, DedupeConsecutive: true},
			"git log":      {MaxLines: 30, DedupeConsecutive: false},
			"npm install":  {MaxLines: 15, DedupeConsecutive: true},
			"npm test":     {MaxLines: 50, DedupeConsecutive: true, Mode: "test-results"},
			"cargo test":   {MaxLines: 40, DedupeConsecutive: true, Mode: "test-results"},
			"cargo build":  {MaxLines: 20, DedupeConsecutive: true},
			"pytest":       {MaxLines: 50, DedupeConsecutive: true, Mode: "test-results"},
			"dotnet build": {MaxLines: 25, DedupeConsecutive: true},
			"dotnet test":  {MaxLines: 40, DedupeConsecutive: true, Mode: "test-results"},
			"docker build": {MaxLines: 20, DedupeConsecutive: true},
		},
		Layer2FillerPatterns: []FillerPattern{
			{Find: `\bIt is important to note that\b`, Replace: ""},
			{Find: `\bPlease note that\b`, Replace: ""},
			{Find: `\bYou should\b`, Replace: ""},
			{Find: `\bYou will need to\b`, Replace: "Need to"},
			{Find: `\bin order to\b`, Replace: "to"},
			{Find: `\bmake sure to\b`, Replace: ""},
			{Find: `\bshould be considered as\b`, Replace: "is"},
			{Find: `\bthere is a need to\b`, Replace: "must"},
			{Find: `\bas well as\b`, Replace: "and"},
			{Find: `\bdue to the fact that\b`, Replace: "because"},
		},
		Layer2DegreeThreshold: DegreeThresholds{LiteMinDegree: 5, FullMinDegree: 1},
	}
}

// ResolveRulesPath mirrors the bash script's RULES_PATH resolution: a
// project override at .plaesy/scripts/configs/plaesy-trim-rules.json wins;
// otherwise fall back to the framework's own copy alongside this source
// tree (best-effort, since a compiled binary has no fixed "script
// directory" the way plaesy-trim.sh does — see LoadRules for the final
// baked-in fallback). frameworkRoot may be "".
func ResolveRulesPath(repoRoot, frameworkRoot string) string {
	projectPath := filepath.Join(repoRoot, ".plaesy", "scripts", "configs", "plaesy-trim-rules.json")
	if fileExists(projectPath) {
		return projectPath
	}
	if frameworkRoot != "" {
		fwPath := filepath.Join(frameworkRoot, "scripts", "configs", "plaesy-trim-rules.json")
		if fileExists(fwPath) {
			return fwPath
		}
	}
	return projectPath
}

// LoadRules reads and parses the rules file at path, falling back to the
// built-in defaultRules() when the file does not exist or fails to parse.
func LoadRules(path string) *Rules {
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultRules()
	}
	var r Rules
	if err := json.Unmarshal(data, &r); err != nil {
		return defaultRules()
	}
	if r.Layer1Commands == nil {
		r.Layer1Commands = defaultRules().Layer1Commands
	}
	return &r
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
