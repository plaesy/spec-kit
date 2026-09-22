// Package detectstack ports scripts/bash/detect-stack.sh (and its PowerShell
// twin scripts/powershell/detect-stack.ps1) to Go. It maps a target project
// to the instructions/*.instructions.md files relevant to it, driven by
// instructions/mapping.json's always_load list plus keyword/extension/
// filename categories, using only the standard library.
package detectstack

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// manifestFiles are the root-level (and recursively, monorepo) manifests
// scanned for keyword content, mirroring both shell twins' identical lists.
var manifestFiles = []string{
	"package.json", "go.mod", "pom.xml", "build.gradle",
	"Cargo.toml", "pubspec.yaml", "Gemfile", "requirements.txt",
}

// specGlobs are spec/context/analysis files whose content also feeds the
// keyword scan, mirroring both shell twins.
var specGlobs = []string{
	"specs/*/context.md",
	"specs/*/requirements.md",
	"docs/specs/*/context.md",
	"docs/specs/*/requirements.md",
}

// maxDepth mirrors both shell twins' -maxdepth 3 / -Depth 3 recursive walk
// limit, counted in path segments below the target directory.
const maxDepth = 3

// entry is one mapping.json category member. Only the fields actually read
// by the shell twins are modeled (see mapping.json's field_semantics_note).
type entry struct {
	File       string   `json:"file"`
	Keywords   []string `json:"keywords"`
	Extensions []string `json:"extensions"`
	Filenames  []string `json:"filenames"`
}

// wordBoundaryRE builds a \b...\b pattern per keyword, mirroring the bash
// twin's manual [![:alnum:]_] boundary check and the PowerShell twin's
// "\b" + Escape(kw) + "\b" regex, so short/common keywords ("java" inside
// "javascript") never match as a bare substring.
func wordBoundaryRE(keyword string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(keyword) + `\b`)
}

// loadMapping reads and parses instructions/mapping.json, splitting out the
// always_load array from the rest of the (variable) category set.
func loadMapping(path string) (alwaysLoad []string, categories map[string]map[string]entry, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	var raw struct {
		Mappings map[string]json.RawMessage `json:"mappings"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, nil, err
	}

	if al, ok := raw.Mappings["always_load"]; ok {
		if err := json.Unmarshal(al, &alwaysLoad); err != nil {
			return nil, nil, err
		}
	}

	categories = make(map[string]map[string]entry)
	for name, rawCat := range raw.Mappings {
		if name == "always_load" {
			continue
		}
		var cat map[string]entry
		if err := json.Unmarshal(rawCat, &cat); err != nil {
			return nil, nil, err
		}
		categories[name] = cat
	}
	return alwaysLoad, categories, nil
}

// depth returns how many path segments dir is below root (0 for root
// itself), used to enforce maxDepth the same way -maxdepth/-Depth do.
func depth(root, dir string) int {
	rel, err := filepath.Rel(root, dir)
	if err != nil || rel == "." {
		return 0
	}
	return len(strings.Split(filepath.ToSlash(rel), "/"))
}

// isPruned reports whether name (a directory) should be skipped, mirroring
// both shell twins pruning .git and node_modules from the walk.
func isPruned(name string) bool {
	return name == ".git" || name == "node_modules"
}

// Detect mirrors detect-stack.sh/.ps1: it returns the sorted, deduplicated
// list of instruction filenames relevant to targetDir, using
// plaesyRoot/instructions/mapping.json as the keyword registry.
func Detect(targetDir, plaesyRoot string) ([]string, error) {
	mappingPath := filepath.Join(plaesyRoot, "instructions", "mapping.json")

	result := make(map[string]struct{})

	alwaysLoad, categories, err := loadMapping(mappingPath)
	if err != nil {
		// Mirrors both shell twins: a missing/unreadable mapping.json is not
		// fatal, it just means no always_load / keyword matches are emitted.
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	for _, f := range alwaysLoad {
		result[f] = struct{}{}
	}

	// Collect every extension/filename to look for across every category
	// entry up front, so the recursive walk below happens exactly once
	// (mirrors both shell twins' "one walk total" optimization).
	extToFiles := make(map[string][]string)
	nameToFiles := make(map[string][]string)
	for _, cat := range categories {
		for _, e := range cat {
			for _, ext := range e.Extensions {
				extToFiles[strings.ToLower(ext)] = append(extToFiles[strings.ToLower(ext)], e.File)
			}
			for _, fn := range e.Filenames {
				nameToFiles[strings.ToLower(fn)] = append(nameToFiles[strings.ToLower(fn)], e.File)
			}
		}
	}

	var scanText strings.Builder
	directFiles := make(map[string]struct{})
	reactDetected := false

	// Root-level manifests first (fast path), matching both shell twins.
	for _, f := range manifestFiles {
		p := filepath.Join(targetDir, f)
		if b, err := os.ReadFile(p); err == nil {
			scanText.Write(b)
			scanText.WriteByte(' ')
		}
	}

	// Single recursive walk (depth-limited, .git/node_modules pruned) that
	// both gathers monorepo-nested manifest content and checks for
	// extension/filename/tsx-jsx presence, mirroring the shell twins' one
	// combined find/Get-ChildItem pass.
	manifestSet := make(map[string]struct{}, len(manifestFiles))
	for _, f := range manifestFiles {
		manifestSet[f] = struct{}{}
	}
	if info, statErr := os.Stat(targetDir); statErr == nil && info.IsDir() {
		_ = filepath.WalkDir(targetDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil //nolint:nilerr // best-effort walk, matches shells' 2>/dev/null
			}
			if path != targetDir && depth(targetDir, path) > maxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if d.IsDir() {
				if isPruned(d.Name()) {
					return filepath.SkipDir
				}
				return nil
			}

			base := d.Name()
			lower := strings.ToLower(base)

			// Nested manifests (monorepo support), skipping the root-level
			// ones already read above, plus *.csproj which has no fixed name.
			if depth(targetDir, path) >= 1 {
				if _, ok := manifestSet[base]; ok {
					if b, rErr := os.ReadFile(path); rErr == nil {
						scanText.Write(b)
						scanText.WriteByte(' ')
					}
				}
			}
			if strings.HasSuffix(lower, ".csproj") {
				if b, rErr := os.ReadFile(path); rErr == nil {
					scanText.Write(b)
					scanText.WriteByte(' ')
				}
			}

			switch {
			case strings.HasSuffix(lower, ".tsx"), strings.HasSuffix(lower, ".jsx"):
				reactDetected = true
			}
			for ext, files := range extToFiles {
				if strings.HasSuffix(lower, ext) {
					for _, f := range files {
						directFiles[f] = struct{}{}
					}
				}
			}
			if files, ok := nameToFiles[lower]; ok {
				for _, f := range files {
					directFiles[f] = struct{}{}
				}
			}
			return nil
		})
	}

	// Spec/context/analysis files.
	for _, g := range specGlobs {
		matches, _ := filepath.Glob(filepath.Join(targetDir, g))
		for _, m := range matches {
			if b, err := os.ReadFile(m); err == nil {
				scanText.Write(b)
				scanText.WriteByte(' ')
			}
		}
	}
	for _, f := range []string{"docs/project.json", ".plaesy/analysis/project.json"} {
		p := filepath.Join(targetDir, f)
		if b, err := os.ReadFile(p); err == nil {
			scanText.Write(b)
			scanText.WriteByte(' ')
		}
	}

	if reactDetected {
		scanText.WriteString(" tsx jsx react")
	}
	scanTextLower := strings.ToLower(scanText.String())

	for _, cat := range categories {
		for _, e := range cat {
			if e.File == "" {
				continue
			}
			for _, kw := range e.Keywords {
				kw = strings.TrimSpace(kw)
				if kw == "" {
					continue
				}
				if wordBoundaryRE(kw).MatchString(scanTextLower) {
					result[e.File] = struct{}{}
					break
				}
			}
		}
	}
	for f := range directFiles {
		result[f] = struct{}{}
	}

	out := make([]string, 0, len(result))
	for f := range result {
		out = append(out, f)
	}
	sort.Strings(out)
	return out, nil
}
