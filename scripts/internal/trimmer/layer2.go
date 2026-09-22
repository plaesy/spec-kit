package trimmer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	fenceRE        = regexp.MustCompile("^```")
	multiSpaceRE   = regexp.MustCompile(` {2,}`)
	leadingSpaceRE = regexp.MustCompile(`^[ \t]+`)
	stopwordRE     = regexp.MustCompile(`(?i)\b(that|which|very|really|basically|simply)\b[ \t]*`)
	blankRunRE     = regexp.MustCompile(`\n{3,}`)
)

// graphNode is the subset of a project.graph.json node needed to resolve
// the compression level from a file's community degree.
type graphNode struct {
	ID     string `json:"id"`
	Degree int    `json:"degree"`
}
type graphFile struct {
	Nodes []graphNode `json:"nodes"`
}

// getNodeDegree mirrors get_node_degree(): looks up a file's degree in
// project.graph.json at graphPath by relative path, matching either
// forward- or backslash-separated ids (the bash version only produced
// Windows-style ids as a literal-text-match hack; parsing the JSON properly
// here lets both separators match).
func getNodeDegree(graphPath, relative string) (int, bool) {
	data, err := os.ReadFile(graphPath)
	if err != nil {
		return 0, false
	}
	var g graphFile
	if err := json.Unmarshal(data, &g); err != nil {
		return 0, false
	}
	slash := filepath.ToSlash(relative)
	back := strings.ReplaceAll(relative, "/", `\`)
	for _, n := range g.Nodes {
		if n.ID == relative || n.ID == slash || n.ID == back {
			return n.Degree, true
		}
	}
	return 0, false
}

// ResolveLevel mirrors resolve_level(): an explicit requested level always
// wins; otherwise the file's graph degree picks lite/full/ultra via the
// rules' thresholds, defaulting to "full" when no graph data is available.
func ResolveLevel(rules *Rules, graphPath, relative, requested string) string {
	if requested != "" {
		return requested
	}
	degree, ok := getNodeDegree(graphPath, relative)
	if !ok {
		return "full"
	}
	t := rules.Layer2DegreeThreshold
	switch {
	case degree >= t.LiteMinDegree:
		return "lite"
	case degree >= t.FullMinDegree:
		return "full"
	default:
		return "ultra"
	}
}

// compiledFillerPatterns lazily compiles a Rules' filler-pattern regexes;
// invalid patterns (should not occur with well-formed rules.json) are
// skipped, mirroring bash's `sed ... || printf original` no-op fallback.
func compiledFillerPatterns(rules *Rules) []struct {
	re      *regexp.Regexp
	replace string
} {
	out := make([]struct {
		re      *regexp.Regexp
		replace string
	}, 0, len(rules.Layer2FillerPatterns))
	for _, p := range rules.Layer2FillerPatterns {
		re, err := regexp.Compile("(?i)" + p.Find)
		if err != nil {
			continue
		}
		out = append(out, struct {
			re      *regexp.Regexp
			replace string
		}{re, p.Replace})
	}
	return out
}

// applyLevel mirrors apply_level(): applies filler-pattern substitution at
// lite/full/ultra, whitespace normalization at full/ultra, and stopword
// stripping at ultra, always finishing by collapsing 3+ consecutive
// newlines down to a single blank line.
func applyLevel(piece, level string, rules *Rules) string {
	if level == "lite" || level == "full" || level == "ultra" {
		for _, fp := range compiledFillerPatterns(rules) {
			piece = fp.re.ReplaceAllString(piece, fp.replace)
		}
	}
	if level == "full" || level == "ultra" {
		piece = multiSpaceRE.ReplaceAllString(piece, " ")
		piece = replacePerLine(piece, func(line string) string {
			return leadingSpaceRE.ReplaceAllString(line, "")
		})
	}
	if level == "ultra" {
		piece = stopwordRE.ReplaceAllString(piece, "")
	}
	piece = blankRunRE.ReplaceAllString(piece, "\n\n")
	return piece
}

func replacePerLine(text string, f func(string) string) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		lines[i] = f(l)
	}
	return strings.Join(lines, "\n")
}

// CompressProse mirrors compress_prose(): applyLevel() is run on every
// prose span, leaving fenced ```code blocks (fence lines included)
// untouched.
func CompressProse(text, level string, rules *Rules) string {
	lines := splitLinesKeepTrailing(text)
	var out strings.Builder
	var buf strings.Builder
	inCode := false

	flushProse := func() {
		out.WriteString(applyLevel(buf.String(), level, rules))
		buf.Reset()
	}

	for _, line := range lines {
		if fenceRE.MatchString(line) {
			if inCode {
				buf.WriteString(line)
				buf.WriteString("\n")
				out.WriteString(buf.String())
				buf.Reset()
				inCode = false
			} else {
				flushProse()
				buf.WriteString(line)
				buf.WriteString("\n")
				inCode = true
			}
			continue
		}
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	if inCode {
		out.WriteString(buf.String())
	} else {
		flushProse()
	}
	return out.String()
}

// splitLinesKeepTrailing mirrors bash's `while IFS= read -r line || [[ -n
// "$line" ]]` over a here-string: splits on "\n" without discarding a
// trailing empty line the way splitLines does, since compress_prose must
// reproduce the original file's trailing newline exactly.
func splitLinesKeepTrailing(text string) []string {
	if text == "" {
		return nil
	}
	trimmed := strings.TrimSuffix(text, "\n")
	return strings.Split(trimmed, "\n")
}

// FileCompressResult reports one file's compression outcome, mirroring the
// two lines printed by compress_one_file().
type FileCompressResult struct {
	Relative string
	Level    string
	Before   int
	After    int
}

func (r FileCompressResult) String() string {
	return fmt.Sprintf("%s [%s]\n  %d tokens -> %d tokens  (%s%% saved)\n", r.Relative, r.Level, r.Before, r.After, formatPct(PctSaved(r.Before, r.After)))
}

func formatPct(p float64) string {
	return fmt.Sprintf("%.1f", p)
}

// CompressOneFile mirrors compress_one_file(): compress a single file at
// requestedLevel (auto-resolved via the graph when ""), and unless dryRun,
// back it up to "<file>.bak" and overwrite it in place, recording the
// before/after token counts to statsPath.
func CompressOneFile(rules *Rules, repoRoot, graphPath, statsPath, file, requestedLevel string, dryRun bool) (FileCompressResult, error) {
	original, err := os.ReadFile(file)
	if err != nil {
		return FileCompressResult{}, err
	}
	relative, err := filepath.Rel(repoRoot, file)
	if err != nil || strings.HasPrefix(relative, "..") {
		relative = file
	}
	level := ResolveLevel(rules, graphPath, relative, requestedLevel)
	compressed := CompressProse(string(original), level, rules)

	before := EstimateTokens(string(original))
	after := EstimateTokens(compressed)
	result := FileCompressResult{Relative: relative, Level: level, Before: before, After: after}

	if !dryRun {
		if err := os.WriteFile(file+".bak", original, 0o644); err != nil {
			return result, err
		}
		if err := os.WriteFile(file, []byte(compressed), 0o644); err != nil {
			return result, err
		}
		if statsPath != "" {
			if err := AddStatRecord(statsPath, "layer2", relative, before, after); err != nil {
				return result, err
			}
		}
	}
	return result, nil
}

// CompressPath mirrors cmd_compress(): file or directory dispatch. For a
// directory, only *.md files are compressed - direct children only, unless
// recurse is set.
func CompressPath(rules *Rules, repoRoot, graphPath, statsPath, path, level string, recurse, dryRun bool) ([]FileCompressResult, error) {
	target := path
	if _, err := os.Stat(target); err != nil {
		target = filepath.Join(repoRoot, path)
	}
	info, err := os.Stat(target)
	if err != nil {
		return nil, fmt.Errorf("path not found: %s", path)
	}

	var files []string
	if info.IsDir() {
		if recurse {
			err = filepath.Walk(target, func(p string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !fi.IsDir() && strings.EqualFold(filepath.Ext(p), ".md") {
					files = append(files, p)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			entries, err := os.ReadDir(target)
			if err != nil {
				return nil, err
			}
			for _, e := range entries {
				if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), ".md") {
					files = append(files, filepath.Join(target, e.Name()))
				}
			}
		}
	} else {
		files = []string{target}
	}

	results := make([]FileCompressResult, 0, len(files))
	for _, f := range files {
		r, err := CompressOneFile(rules, repoRoot, graphPath, statsPath, f, level, dryRun)
		if err != nil {
			return results, err
		}
		results = append(results, r)
	}
	return results, nil
}
