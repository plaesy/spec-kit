// Package aiheaders ports scripts/bash/inject-ai-headers.sh: it injects
// platform-specific YAML front-matter headers into prompt/chatmode/
// instructions files for a chosen AI platform.
package aiheaders

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// DefaultHeader is the fallback header file used when no platform-specific
// header exists.
const DefaultHeader = "manual.header.yaml"

// ValidPlatforms mirrors the bash script's validate_ai_platform list.
var ValidPlatforms = []string{
	"copilot", "cursor", "windsurf", "claude", "chatgpt", "gemini",
	"trae-ai", "qwen-code", "codex-cli", "opencode-cli", "local-ai", "manual",
}

// ValidatePlatform returns true if platform is a recognized AI platform.
func ValidatePlatform(platform string) bool {
	for _, p := range ValidPlatforms {
		if platform == p {
			return true
		}
	}
	return false
}

// Options mirrors the CLI flags accepted by the bash script.
type Options struct {
	AIPlatform string
	TargetDir  string
	HeadersDir string // defaults to <repo-root>/templates/ai-headers
	DryRun     bool
	Force      bool
	Merge      bool
	Backup     bool
	ListOnly   bool
	Patterns   []string
	Excludes   []string

	// Stdout/Stderr allow tests (and callers) to capture output; default to
	// os.Stdout/os.Stderr when nil.
	Stdout io.Writer
	Stderr io.Writer
}

func (o *Options) out() io.Writer {
	if o.Stdout != nil {
		return o.Stdout
	}
	return os.Stdout
}

func (o *Options) err() io.Writer {
	if o.Stderr != nil {
		return o.Stderr
	}
	return os.Stderr
}

func (o *Options) logInfo(format string, args ...any) {
	fmt.Fprintf(o.err(), "[INFO] "+format+"\n", args...)
}

func (o *Options) logSuccess(format string, args ...any) {
	fmt.Fprintf(o.err(), "[SUCCESS] "+format+"\n", args...)
}

func (o *Options) logWarning(format string, args ...any) {
	fmt.Fprintf(o.err(), "[WARNING] "+format+"\n", args...)
}

func (o *Options) logError(format string, args ...any) {
	fmt.Fprintf(o.err(), "[ERROR] "+format+"\n", args...)
}

// Result summarizes a run, mirroring the bash script's final counters.
type Result struct {
	Processed int
	Skipped   int
}

// GetHeaderFile resolves the header YAML file for a platform/file-type,
// mirroring get_header_file()'s fallback order:
//  1. <headersDir>/<platform>.<ftype>.yaml
//  2. <headersDir>/<platform>.header.yaml
//  3. <headersDir>/<DefaultHeader>
func GetHeaderFile(headersDir, platform, ftype string, warn func(string, ...any)) string {
	if ftype != "" {
		candidate := filepath.Join(headersDir, platform+"."+ftype+".yaml")
		if fileExists(candidate) {
			return candidate
		}
	}

	platformHeader := filepath.Join(headersDir, platform+".header.yaml")
	if fileExists(platformHeader) {
		return platformHeader
	}

	if warn != nil {
		if ftype != "" {
			warn("No specific header for %s (type: %s), using default header", platform, ftype)
		} else {
			warn("No specific header for %s, using default header", platform)
		}
	}
	return filepath.Join(headersDir, DefaultHeader)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// FindPromptFiles walks targetDir and returns files matching patterns
// (shell-style filename globs, matched against the base name) and not
// matching any exclude glob against the full relative path, mirroring
// find_prompt_files().
func FindPromptFiles(targetDir string, patterns, excludes []string) ([]string, error) {
	info, err := os.Stat(targetDir)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("target directory does not exist: %s", targetDir)
	}

	if len(patterns) == 0 {
		patterns = []string{"*.prompt.md", "*.chatmode.md", "*.instructions.md"}
	}

	var results []string
	err = filepath.Walk(targetDir, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if fi.IsDir() {
			return nil
		}
		base := filepath.Base(path)

		matched := false
		for _, p := range patterns {
			if ok, _ := filepath.Match(p, base); ok {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}

		for _, e := range excludes {
			// bash used `-not -path "*/<exclude>"`, i.e. the exclude glob is
			// matched against the tail of the path after a "/".
			if ok, _ := filepath.Match("*/"+e, filepath.ToSlash(path)); ok {
				return nil
			}
			if ok, _ := filepath.Match(e, base); ok {
				return nil
			}
		}

		results = append(results, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

var frontMatterDelim = "---"

// splitFrontMatter returns (frontMatterLines, bodyStartLine, ok). bodyStartLine
// is the 1-based line number of the closing "---" (matching bash's
// `header_end` from `grep -n '^---$' | sed -n '2p'`).
func splitFrontMatter(lines []string) (fm []string, closingLine int, ok bool) {
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != frontMatterDelim {
		return nil, 0, false
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == frontMatterDelim {
			return lines[1:i], i + 1, true
		}
	}
	return nil, 0, false
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Preserve trailing-newline semantics closely enough: split on \n, keep
	// a trailing empty element if the file ends with \n (matches bash's
	// line-based tools operating on text files).
	content := string(data)
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	return lines, nil
}

var keyLineRE = regexp.MustCompile(`(?i)^[[:space:]]*(platform|type)[[:space:]]*:`)

// HasHeader mirrors has_header(): true if the file has YAML front-matter
// (with platform:/type: keys, or just any closing '---'), or if the first
// 20 lines contain a recognizable legacy header marker.
func HasHeader(path string) bool {
	lines, err := readLines(path)
	if err != nil {
		return false
	}

	fm, _, ok := splitFrontMatter(lines)
	if ok {
		for _, l := range fm {
			if keyLineRE.MatchString(l) {
				return true
			}
		}
		// front-matter present even without explicit keys
		return true
	}

	limit := 20
	if limit > len(lines) {
		limit = len(lines)
	}
	legacyRE := regexp.MustCompile(`(Constitutional.*Framework|AI.*Configuration|# .*Header)`)
	for i := 0; i < limit; i++ {
		if legacyRE.MatchString(lines[i]) {
			return true
		}
	}
	return false
}

var frontMatterKeyRE = func(key string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)^[[:space:]]*` + regexp.QuoteMeta(key) + `[[:space:]]*:[[:space:]]*(.*)$`)
}

// GetFrontMatterKey extracts a key's value from a file's YAML front-matter,
// mirroring get_front_matter_key(). Returns "" if not present.
func GetFrontMatterKey(path, key string) string {
	lines, err := readLines(path)
	if err != nil {
		return ""
	}
	fm, _, ok := splitFrontMatter(lines)
	if !ok {
		return ""
	}
	re := frontMatterKeyRE(key)
	for _, l := range fm {
		if m := re.FindStringSubmatch(l); m != nil {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

var descriptionRE = regexp.MustCompile(`(?i)^[[:space:]]*description[[:space:]]*:[[:space:]]*"?([^"]*)"?[[:space:]]*$`)

// GetFileDescription mirrors get_file_description(): pulls description from
// the target file's own front-matter, falling back to a filename-derived
// default keyed by extension.
func GetFileDescription(targetFile string) string {
	if lines, err := readLines(targetFile); err == nil {
		if fm, _, ok := splitFrontMatter(lines); ok {
			for _, l := range fm {
				if m := descriptionRE.FindStringSubmatch(l); m != nil {
					return m[1]
				}
			}
		}
	}

	base := filepath.Base(targetFile)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	// bash strips a single trailing ".ext" via `sed 's/\.[^.]*$//'`; mirror
	// that (not multi-suffix stripping) so "*.prompt.md" -> "*.prompt".
	switch {
	case strings.HasSuffix(targetFile, ".prompt.md"):
		return "Prompt: " + strings.TrimSuffix(base, ".prompt.md")
	case strings.HasSuffix(targetFile, ".chatmode.md"):
		return "Chat mode: " + strings.TrimSuffix(base, ".chatmode.md")
	case strings.HasSuffix(targetFile, ".instructions.md"):
		return "Instructions: " + strings.TrimSuffix(base, ".instructions.md")
	default:
		return "Configuration: " + name
	}
}

var headerContentKeyRE = regexp.MustCompile(`^header_content:[[:space:]]*\|[[:space:]]*$`)

// GetYAMLHeaderContent extracts the `header_content: |` block scalar from a
// header YAML file and replaces the {{DESCRIPTION}} placeholder using
// targetFile (if non-empty), mirroring get_yaml_header_content(). Only the
// minimal YAML block-scalar shape used by templates/ai-headers/*.yaml is
// supported (no full YAML parser is used, per stdlib-only constraint).
func GetYAMLHeaderContent(yamlFile, targetFile string) (string, error) {
	f, err := os.Open(yamlFile)
	if err != nil {
		return "", fmt.Errorf("YAML header file not found or not readable: %s", yamlFile)
	}
	defer f.Close()

	var content []string
	inBlock := false
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !inBlock {
			if headerContentKeyRE.MatchString(line) {
				inBlock = true
			}
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			trimmed := strings.TrimLeft(line, " \t")
			content = append(content, trimmed)
			continue
		}
		// first non-indented line ends the block scalar
		inBlock = false
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	result := strings.Join(content, "\n")
	if targetFile != "" && result != "" {
		result = strings.ReplaceAll(result, "{{DESCRIPTION}}", GetFileDescription(targetFile))
	}
	return result, nil
}

// createBackup mirrors create_backup(): copies file to file.backup.<timestamp>.
func createBackup(path string, o *Options) error {
	backup := fmt.Sprintf("%s.backup.%s", path, time.Now().Format("20060102_150405"))
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	mode := os.FileMode(0o644)
	if err == nil {
		mode = info.Mode()
	}
	if err := os.WriteFile(backup, data, mode); err != nil {
		return err
	}
	o.logInfo("Created backup: %s", backup)
	return nil
}

// removeExistingHeader mirrors remove_existing_header(): strips a leading
// front-matter block (up to and including the second '---' line, plus the
// two blank lines inject_header adds after it), if present.
func removeExistingHeader(path string, o *Options) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	_, closingLine, ok := splitFrontMatter(lines)
	if !ok {
		return nil
	}
	// bash: tail -n "+$((header_end + 2))" -- skip the closing '---' line and
	// one line after it (header_end is 1-based line number of closing ---).
	skip := closingLine + 1 // 0-based index of first line to skip past header_end
	if skip > len(lines) {
		skip = len(lines)
	}
	rest := strings.Join(lines[skip:], "\n")
	if err := os.WriteFile(path, []byte(rest), filePerm(path)); err != nil {
		return err
	}
	o.logInfo("Removed existing header from %s", filepath.Base(path))
	return nil
}

func filePerm(path string) os.FileMode {
	if info, err := os.Stat(path); err == nil {
		return info.Mode()
	}
	return 0o644
}

// injectHeader mirrors inject_header(): prepends header content plus two
// blank lines to the file.
func injectHeader(path, headerContent string, o *Options) error {
	original, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	mode := filePerm(path)

	var b strings.Builder
	b.WriteString(headerContent)
	b.WriteString("\n\n\n")
	b.Write(original)

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), mode); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// mergeFrontMatter mirrors merge_front_matter(): merges header front-matter
// keys into the file's existing front-matter, keeping existing values and
// only appending keys the file doesn't already have.
func mergeFrontMatter(path, headerContent string, o *Options) (bool, error) {
	filename := filepath.Base(path)

	lines, err := readLines(path)
	if err != nil {
		return false, err
	}
	existingFM, closingLine, ok := splitFrontMatter(lines)
	if !ok || len(existingFM) == 0 {
		o.logWarning("Cannot merge front-matter for %s: missing existing or header front-matter", filename)
		return false, nil
	}

	headerLines := strings.Split(headerContent, "\n")
	headerFM, _, hok := splitFrontMatter(headerLines)
	if !hok || len(headerFM) == 0 {
		o.logWarning("Cannot merge front-matter for %s: missing existing or header front-matter", filename)
		return false, nil
	}

	existingKeys, existingOrder := parseFMKeys(existingFM)
	headerKeys, headerOrder := parseFMKeys(headerFM)

	var merged []string
	for _, k := range existingOrder {
		merged = append(merged, k+": "+existingKeys[k])
	}
	for _, k := range headerOrder {
		if _, exists := existingKeys[k]; !exists {
			merged = append(merged, k+": "+headerKeys[k])
		}
	}

	if len(merged) == 0 {
		o.logWarning("Merged front-matter is empty for %s", filename)
		return false, nil
	}

	if o.DryRun {
		o.logInfo("[DRY RUN] Would merge front-matter into %s:\n%s", filename, strings.Join(merged, "\n"))
		return true, nil
	}

	if o.Backup {
		if err := createBackup(path, o); err != nil {
			return false, err
		}
	}

	rest := lines[closingLine:] // lines after the closing '---'
	var b strings.Builder
	b.WriteString("---\n")
	for _, l := range merged {
		b.WriteString(l)
		b.WriteString("\n")
	}
	b.WriteString("---\n")
	b.WriteString(strings.Join(rest, "\n"))

	if err := os.WriteFile(path, []byte(b.String()), filePerm(path)); err != nil {
		return false, err
	}
	o.logSuccess("Merged front-matter into %s", filename)
	return true, nil
}

// parseFMKeys parses "key: value" lines (first ": " splits key/value),
// preserving first-seen order, mirroring the awk logic in merge_front_matter.
func parseFMKeys(lines []string) (map[string]string, []string) {
	vals := map[string]string{}
	var order []string
	for _, l := range lines {
		idx := strings.Index(l, ": ")
		var key, val string
		if idx < 0 {
			key = strings.TrimSpace(l)
			val = ""
		} else {
			key = strings.TrimSpace(l[:idx])
			val = l[idx+2:]
		}
		if key == "" {
			continue
		}
		if _, seen := vals[key]; !seen {
			order = append(order, key)
		}
		vals[key] = val
	}
	return vals, order
}

// processFile mirrors process_file(): backup/force/inject flow for a single
// file (non-merge path).
func processFile(path, headerContent string, o *Options) (processed bool, err error) {
	filename := filepath.Base(path)

	if o.DryRun {
		if HasHeader(path) && !o.Force {
			o.logInfo("[DRY RUN] Would skip %s (already has header, use --force to overwrite)", filename)
		} else {
			o.logInfo("[DRY RUN] Would inject header into %s", filename)
		}
		return true, nil
	}

	if HasHeader(path) && !o.Force {
		o.logWarning("Skipping %s (already has header, use --force to overwrite)", filename)
		return false, nil
	}

	if o.Backup {
		if err := createBackup(path, o); err != nil {
			return false, err
		}
	}

	if o.Force && HasHeader(path) {
		if err := removeExistingHeader(path, o); err != nil {
			return false, err
		}
	}

	if err := injectHeader(path, headerContent, o); err != nil {
		return false, err
	}
	o.logSuccess("Injected header into %s", filename)
	return true, nil
}

// classifyFileType mirrors the bash heuristics for prompts/chatmodes/
// instructions/generic based on the file's relative path and name.
func classifyFileType(targetDir, file string) string {
	rel, err := filepath.Rel(targetDir, file)
	if err != nil {
		rel = file
	}
	rel = filepath.ToSlash(rel)

	switch {
	case regexp.MustCompile(`(^|/)prompts?(/|$)`).MatchString(rel) || strings.HasSuffix(file, ".prompt.md"):
		return "prompts"
	case regexp.MustCompile(`(^|/)chatmodes?(/|$)`).MatchString(rel) || strings.HasSuffix(file, ".chatmode.md") || strings.Contains(file, "chatmode"):
		return "chatmodes"
	case regexp.MustCompile(`(^|/)instructions?(/|$)`).MatchString(rel) || strings.HasSuffix(file, ".instruction.md") || strings.Contains(file, "instructions"):
		return "instructions"
	default:
		return "generic"
	}
}

// Run executes the full injection workflow, mirroring main() in the bash
// script. It returns the processed/skipped counts (dry-run counts what
// would happen). listOnly runs produce a zero Result after printing the
// mapping.
func Run(o *Options) (*Result, error) {
	if o.AIPlatform == "" {
		return nil, fmt.Errorf("AI platform is required")
	}
	if o.TargetDir == "" {
		return nil, fmt.Errorf("target directory is required")
	}
	if !ValidatePlatform(o.AIPlatform) {
		return nil, fmt.Errorf("invalid AI platform: %s", o.AIPlatform)
	}

	o.logInfo("Starting header injection for AI platform: %s", o.AIPlatform)

	headerFile := GetHeaderFile(o.HeadersDir, o.AIPlatform, "", o.logWarning)
	if !fileExists(headerFile) {
		return nil, fmt.Errorf("header file not found: %s", headerFile)
	}

	o.logInfo("Using header file: %s", filepath.Base(headerFile))

	files, err := FindPromptFiles(o.TargetDir, o.Patterns, o.Excludes)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		o.logWarning("No prompt file(s) found in %s", o.TargetDir)
		return &Result{}, nil
	}

	o.logInfo("Found %d prompt file(s)", len(files))

	if o.ListOnly {
		for _, file := range files {
			ftype := classifyFileType(o.TargetDir, file)
			headerForType := GetHeaderFile(o.HeadersDir, o.AIPlatform, ftype, nil)
			fmt.Fprintf(o.out(), "%s -> type=%s -> header=%s\n", file, ftype, filepath.Base(headerForType))
		}
		return &Result{}, nil
	}

	processed, skipped := 0, 0

	for _, file := range files {
		ftype := classifyFileType(o.TargetDir, file)
		headerForType := GetHeaderFile(o.HeadersDir, o.AIPlatform, ftype, o.logWarning)

		headerContent, err := GetYAMLHeaderContent(headerForType, file)
		if err != nil {
			return nil, err
		}

		if strings.TrimSpace(headerContent) == "" {
			o.logInfo("Skipping %s (header file %s is empty)", filepath.Base(file), filepath.Base(headerForType))
			skipped++
			continue
		}

		fmPlatform := GetFrontMatterKey(file, "platform")
		if fmPlatform != "" && fmPlatform != o.AIPlatform && !o.Force {
			o.logInfo("Skipping %s (front-matter platform: %s does not match target: %s)", filepath.Base(file), fmPlatform, o.AIPlatform)
			skipped++
			continue
		}
		if fmPlatform != "" && fmPlatform == o.AIPlatform && !o.Force {
			o.logInfo("Skipping %s (platform %s already configured)", filepath.Base(file), o.AIPlatform)
			skipped++
			continue
		}

		if o.Merge && HasHeader(file) && !o.Force {
			ok, err := mergeFrontMatter(file, headerContent, o)
			if err != nil {
				return nil, err
			}
			if ok {
				processed++
			} else {
				skipped++
			}
			continue
		}

		ok, err := processFile(file, headerContent, o)
		if err != nil {
			return nil, err
		}
		if ok {
			processed++
		} else {
			skipped++
		}
	}

	fmt.Fprintln(o.err())
	if o.DryRun {
		o.logInfo("DRY RUN COMPLETE")
		o.logInfo("Would process: %d files", processed)
		o.logInfo("Would skip: %d files", skipped)
		return &Result{Processed: processed, Skipped: skipped}, nil
	}

	o.logSuccess("HEADER INJECTION COMPLETE")
	o.logSuccess("Processed: %d files", processed)
	if skipped > 0 {
		o.logInfo("Skipped: %d files", skipped)
	}

	if processed > 0 {
		configFile := filepath.Join(o.TargetDir, ".plaesy-headers.json")
		config := fmt.Sprintf(`{
  "ai_platform": %q,
  "injection_timestamp": %q,
  "processed_files": %d,
  "header_file": %q,
  "constitutional_framework_version": "3.0.0"
}
`, o.AIPlatform, time.Now().UTC().Format("2006-01-02T15:04:05Z"), processed, filepath.Base(headerFile))
		if err := os.WriteFile(configFile, []byte(config), 0o644); err != nil {
			return nil, err
		}
		o.logInfo("Created configuration: %s", filepath.Base(configFile))
	}

	return &Result{Processed: processed, Skipped: skipped}, nil
}
