// Package cleaner ports scripts/bash/plaesy-clean.sh (and its PowerShell
// counterpart): safe removal of Plaesy Constitution Kit framework files and
// directories from a project, driven by scripts/configs/platform.json.
package cleaner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/config"
)

// Level is a cleanup aggressiveness level, mirroring CLEANUP_LEVEL.
type Level string

const (
	LevelSafe     Level = "safe"
	LevelThorough Level = "thorough"
	LevelComplete Level = "complete"
)

// ValidLevel reports whether s is one of the three recognized levels.
func ValidLevel(s string) bool {
	switch Level(s) {
	case LevelSafe, LevelThorough, LevelComplete:
		return true
	}
	return false
}

// mappingTypes mirrors the bash script's fixed mapping_types array.
var mappingTypes = []string{"core", "instructions", "prompts", "agents"}

// plaesyDirs mirrors the bash script's plaesy_dirs array. docs/ is
// deliberately excluded, as it may hold project documentation.
var plaesyDirs = []string{".plaesy", "specs"}

// dirsForLevel returns the top-level directories this level may remove.
//
// specs/ holds the user's authored feature documents (`plaesy create feature`
// writes there), and the command's own help reserves it for `thorough`:
// "safe: Remove framework files only, preserve user code". The level used to be
// read only to print a banner, so the default `plaesy clean` destroyed specs
// with no extra confirmation — the one thing the default level promises to
// spare. docs/ stays excluded at every level for the same reason.
func dirsForLevel(level Level) []string {
	if level == LevelThorough || level == LevelComplete {
		return plaesyDirs
	}
	// A fresh one-element slice rather than plaesyDirs[:1], whose spare capacity
	// would let a caller's append write into the package-level list.
	return []string{plaesyDirs[0]}
}

// includesGenericAI reports whether this level also removes the fallback
// platform's own files (AI-INSTRUCTIONS.md, ai-config/*). Those are named
// "Various" in platform.json and are the least certainly-Plaesy thing in a
// project, so they go only at `complete`.
func includesGenericAI(level Level) bool {
	return level == LevelComplete
}

// Options mirrors the parsed CLI flags of plaesy-clean.sh.
type Options struct {
	TargetDir   string
	AutoConfirm bool
	DryRun      bool
	Backup      bool
	Level       Level
	AIChoice    string
	Verbose     bool
}

// Cleaner drives a clean operation against a resolved target directory and
// loaded platform configuration.
type Cleaner struct {
	Opts       Options
	Config     *config.PlatformConfig
	ConfigPath string
}

// New resolves opts.TargetDir to an absolute path and loads platform.json,
// mirroring main()'s upfront TARGET_DIR resolution in the bash script.
func New(opts Options, configPath string) (*Cleaner, error) {
	if opts.TargetDir == "" {
		opts.TargetDir = "."
	}
	if info, err := os.Stat(opts.TargetDir); err == nil && info.IsDir() {
		abs, err := filepath.Abs(opts.TargetDir)
		if err != nil {
			return nil, err
		}
		opts.TargetDir = abs
	}
	if opts.Level == "" {
		opts.Level = LevelSafe
	}

	if configPath == "" {
		p, err := config.DefaultConfigPath()
		if err != nil {
			return nil, err
		}
		configPath = p
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	return &Cleaner{Opts: opts, Config: cfg, ConfigPath: configPath}, nil
}

// ValidateEnvironment mirrors validate_clean_environment: the target
// directory must exist, be readable and be writable.
func (c *Cleaner) ValidateEnvironment() error {
	// Stat first and keep the three failures apart. `err != nil || !info.IsDir()`
	// reports every one of them as "does not exist": a directory that is locked
	// or whose parent denies access is present, and telling the operator it is
	// missing sends them looking for a path problem while the real one is a
	// lock. This is the preflight for a command that deletes things, so a
	// misdiagnosis here is the operator acting on the wrong belief.
	if info, err := os.Stat(c.Opts.TargetDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("target directory '%s' does not exist", c.Opts.TargetDir)
		}
		return fmt.Errorf("cannot access target directory '%s': %w", c.Opts.TargetDir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("target '%s' exists but is not a directory", c.Opts.TargetDir)
	}
	f, err := os.Open(c.Opts.TargetDir)
	if err != nil {
		return fmt.Errorf("target directory '%s' is not readable: %w", c.Opts.TargetDir, err)
	}
	f.Close()
	probe := filepath.Join(c.Opts.TargetDir, ".plaesy-clean-write-test")
	wf, err := os.Create(probe)
	if err != nil {
		return fmt.Errorf("target directory '%s' is not writable", c.Opts.TargetDir)
	}
	wf.Close()
	os.Remove(probe)
	return nil
}

// DetectAllPlatforms mirrors detect_all_platforms: every platform in
// platform.json whose detection pattern exists relative to targetDir,
// falling back to "generic_ai" when none match. Detection is checked in the
// platform declaration order of platform.json.
//
// It returns an error because it cannot answer the question without one. A
// marker that exists but cannot be read is not evidence that the platform is
// absent, and the previous signature had nowhere to put that: a locked marker
// dropped the platform, the loop fell through to "generic_ai", and `plaesy
// clean` went on to build a deletion set for the wrong platform and report
// success. Reading a marker is best-effort when the answer is only "is this
// project an AI project at all"; it is not best-effort when the answer decides
// which files get deleted.
func DetectAllPlatforms(cfg *config.PlatformConfig, configPath, targetDir string) ([]string, error) {
	names, err := cfg.ListPlatforms(configPath)
	if err != nil || len(names) == 0 {
		names = make([]string, 0, len(cfg.Platforms))
		for n := range cfg.Platforms {
			names = append(names, n)
		}
		sort.Strings(names)
	}

	var found []string
	for _, name := range names {
		p, ok := cfg.Platforms[name]
		if !ok || len(p.Detection) == 0 {
			continue
		}
		for _, pattern := range p.Detection {
			// Only an absent marker means the platform is not there. A marker
			// that exists but cannot be read is a different fact, and dropping
			// the platform here falls through to the generic_ai default — so a
			// locked file produces a cleanup set for the wrong platform, with
			// no error to indicate it.
			if _, err := os.Stat(filepath.Join(targetDir, pattern)); err == nil {
				found = append(found, name)
				break
			} else if !os.IsNotExist(err) {
				return nil, fmt.Errorf("cannot read the %s detection marker %q: %w",
					name, pattern, err)
			}
		}
	}
	if len(found) == 0 {
		return []string{"generic_ai"}, nil
	}
	return found, nil
}

// ResolvePlatforms returns the platforms to act on: AIChoice split on
// whitespace if set, else auto-detection, else ["generic"] — mirroring the
// bash script's repeated "$AI_CHOICE else detect_all_platforms" pattern.
//
// The error is from auto-detection only. An explicit --ai choice needs no
// filesystem access, so it never fails.
func (c *Cleaner) ResolvePlatforms() ([]string, error) {
	if c.Opts.AIChoice != "" {
		return strings.Fields(c.Opts.AIChoice), nil
	}
	platforms, err := DetectAllPlatforms(c.Config, c.ConfigPath, c.Opts.TargetDir)
	if err != nil {
		return nil, err
	}
	if len(platforms) == 0 {
		return []string{"generic"}, nil
	}
	return platforms, nil
}

// mappingTargetPath resolves a platform's mapping target for mappingType
// relative to base, mirroring get-mapping-value's platform.*.mapping.* path
// (bash's $CONFIG_MANAGER get-mapping-value $platform $mappingType).
func (c *Cleaner) mappingTargetPath(platform, mappingType string) string {
	p, ok := c.Config.Platforms[platform]
	if !ok {
		return ""
	}
	v := p.Mapping[mappingType]
	if v == "null" {
		return ""
	}
	return v
}

// Plan is the result of analyzing what a clean run would remove.
type Plan struct {
	Platforms       []string
	PlaesyDirs      []string            // found top-level Plaesy dirs (.plaesy, specs)
	PlatformEntries map[string][]string // platform -> found target paths (display form, "/" suffixed for dirs)
	Empty           bool                // true if nothing was found to remove
	// Err is why the plan is empty when it is empty for a reason other than
	// "there is nothing to remove" — chiefly a detection marker that could not
	// be read. Callers must check it before treating Empty as a clean bill of
	// health, because an unreadable marker produces an empty plan that looks
	// identical to a project with nothing to clean.
	Err error
}

// BuildPlan mirrors show_cleanup_plan's analysis phase (without printing).
func (c *Cleaner) BuildPlan() Plan {
	platforms, err := c.ResolvePlatforms()
	if err != nil {
		return Plan{Empty: true, Err: err}
	}
	plan := Plan{Platforms: platforms, PlatformEntries: map[string][]string{}}

	for _, dir := range dirsForLevel(c.Opts.Level) {
		if info, err := os.Stat(filepath.Join(c.Opts.TargetDir, dir)); err == nil && info.IsDir() {
			plan.PlaesyDirs = append(plan.PlaesyDirs, dir)
		}
	}

	for _, platform := range platforms {
		if platform == "generic" {
			continue
		}
		if platform == "generic_ai" && !includesGenericAI(c.Opts.Level) {
			continue
		}
		var entries []string
		for _, mt := range mappingTypes {
			target := c.mappingTargetPath(platform, mt)
			if target == "" {
				continue
			}
			full := filepath.Join(c.Opts.TargetDir, target)
			info, err := os.Stat(full)
			if err != nil {
				continue
			}
			if info.IsDir() {
				entries = append(entries, target+"/ (directory)")
			} else {
				entries = append(entries, target+" (file)")
			}
		}
		if len(entries) > 0 {
			plan.PlatformEntries[platform] = entries
		}
	}

	plan.Empty = len(plan.PlaesyDirs) == 0 && len(plan.PlatformEntries) == 0
	return plan
}

// PrintPlan renders the plan the way show_cleanup_plan does.
func (c *Cleaner) PrintPlan(plan Plan) {
	if c.Opts.DryRun {
		fmt.Println("[DRY RUN MODE] No files will be actually removed.")
		fmt.Println()
	}

	fmt.Printf("Cleanup Plan (Level: %s, Platforms: %s)\n", c.Opts.Level, strings.Join(plan.Platforms, ", "))
	fmt.Println(strings.Repeat("=", 67))

	switch c.Opts.Level {
	case LevelSafe:
		fmt.Println("Safe cleanup: Framework files only, preserve user code")
	case LevelThorough:
		fmt.Println("Thorough cleanup: Framework + specs, preserve user code")
	case LevelComplete:
		fmt.Println("Complete cleanup: Everything Plaesy-related (DANGEROUS)")
	}

	fmt.Println("\nPlaesy Framework Directories:")
	for _, dir := range plan.PlaesyDirs {
		fmt.Printf("  - %s/\n", dir)
	}

	foundAnyPlatform := len(plan.PlatformEntries) > 0
	for _, platform := range plan.Platforms {
		if platform == "generic" {
			continue
		}
		if platform == "generic_ai" && !includesGenericAI(c.Opts.Level) {
			continue
		}
		fmt.Printf("\nPlatform-Specific Files (%s):\n", platform)
		fmt.Printf("  Based on platform.json mapping for %s:\n", platform)
		entries := plan.PlatformEntries[platform]
		if len(entries) == 0 {
			fmt.Printf("  - No platform-specific files found for %s\n", platform)
			continue
		}
		for _, e := range entries {
			fmt.Printf("  - %s\n", e)
		}
	}
	if !foundAnyPlatform {
		fmt.Println("\nPlatform-Specific Files:")
		fmt.Println("  - Generic platform: No specific files to remove")
	}

	fmt.Println("\nWhat will be preserved:")
	fmt.Println("  - Your source code (src/, lib/, components/, etc.)")
	fmt.Println("  - Git repository (.git/)")
	fmt.Println("  - User-generated files not in Plaesy directories")
	fmt.Println("  - Configuration files (.env, .gitignore, etc.)")
	if c.Opts.Level != LevelThorough && c.Opts.Level != LevelComplete {
		fmt.Println("  - specs/ (raise the level to thorough or complete to remove it)")
	}

	if c.Opts.Backup && !c.Opts.DryRun {
		fmt.Println("  - Backup will be created before removal")
	}
	fmt.Println()
}

// Confirm mirrors confirm_removal: auto-confirms when AutoConfirm is set,
// otherwise prompts stdin for a y/N answer. Returns false when the user
// declined (caller should abort without error, matching `exit 0`).
func Confirm(opts Options, in *bufio.Reader, out *os.File) bool {
	if opts.AutoConfirm {
		fmt.Fprintln(out, "Auto-confirm mode: proceeding with deletion...")
		return true
	}
	fmt.Fprintln(out, "This will permanently delete the directories and files listed above.")
	fmt.Fprint(out, "Are you sure you want to continue? (y/N): ")
	line, _ := in.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

// removeResult accumulates outcome of a removal batch, mirroring the bash
// removed_count / failed_removals bookkeeping.
type removeResult struct {
	removed int
	failed  []string
}

// CreateBackup mirrors create_backup: copies .plaesy and the resolved
// platforms' mapped files/dirs into TARGET_DIR/.plaesy-backup-<timestamp>.
// No-op (returns "", nil) if Backup is disabled.
func (c *Cleaner) CreateBackup() (string, error) {
	if !c.Opts.Backup {
		return "", nil
	}
	backupDir := filepath.Join(c.Opts.TargetDir, ".plaesy-backup-"+time.Now().Format("20060102-150405"))
	common.LogInfo("Creating backup: %s", backupDir)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", err
	}

	src := filepath.Join(c.Opts.TargetDir, ".plaesy")
	if info, err := os.Stat(src); err == nil && info.IsDir() {
		if err := copyTree(src, filepath.Join(backupDir, ".plaesy")); err != nil {
			return "", err
		}
	}

	// Everything else this level may delete has to be in the backup too. specs/
	// was not: thorough and above delete it, so the run destroyed the user's
	// feature documents while the "backup created before removal" promise stood.
	for _, dir := range dirsForLevel(c.Opts.Level) {
		if dir == ".plaesy" {
			continue
		}
		src := filepath.Join(c.Opts.TargetDir, dir)
		if info, err := os.Stat(src); err == nil && info.IsDir() {
			if err := copyTree(src, filepath.Join(backupDir, dir)); err != nil {
				return "", err
			}
			common.LogInfo("Backed up Plaesy directory: %s", dir)
		}
	}

	platforms, err := c.ResolvePlatforms()
	if err != nil {
		return "", err
	}
	for _, platform := range platforms {
		if platform == "generic" {
			continue
		}
		if platform == "generic_ai" && !includesGenericAI(c.Opts.Level) {
			continue
		}
		for _, mt := range mappingTypes {
			target := c.mappingTargetPath(platform, mt)
			if target == "" {
				continue
			}
			full := filepath.Join(c.Opts.TargetDir, target)
			info, err := os.Stat(full)
			if err != nil {
				continue
			}
			dest := filepath.Join(backupDir, target)
			if info.IsDir() {
				if err := copyTree(full, dest); err != nil {
					return "", err
				}
				common.LogInfo("Backed up platform %s directory: %s", platform, target)
			} else {
				if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
					return "", err
				}
				if err := copyFile(full, dest); err != nil {
					return "", err
				}
				common.LogInfo("Backed up platform %s file: %s", platform, target)
			}
		}
	}

	common.LogSuccess("Backup created: %s", backupDir)
	return backupDir, nil
}

// Remove mirrors remove_plaesy_directories: removes the top-level Plaesy
// framework directories, then each resolved platform's mapped
// core/instructions/prompts/agents targets, then prunes now-empty parent
// directories of those targets. Returns the number of items removed.
func (c *Cleaner) Remove() (int, error) {
	result := removeResult{}

	common.LogInfo("Removing plaesy directories based on platform.json mapping...")
	common.LogInfo("Removing Plaesy framework directories...")
	for _, dir := range dirsForLevel(c.Opts.Level) {
		full := filepath.Join(c.Opts.TargetDir, dir)
		// `err == nil && info.IsDir()` skipped a directory that exists but
		// cannot be inspected, and the run went on to report success. For a
		// command whose job is deletion, "I could not look at this" and "there
		// is nothing here" must not produce the same outcome.
		info, statErr := os.Stat(full)
		switch {
		case statErr != nil && os.IsNotExist(statErr):
			continue
		case statErr != nil:
			return 0, fmt.Errorf("cannot access %s: %w", full, statErr)
		case !info.IsDir():
			continue
		}
		if err := os.RemoveAll(full); err != nil {
			result.failed = append(result.failed, dir)
			fmt.Printf("  ❌ Failed to remove: %s\n", dir)
		} else {
			result.removed++
			fmt.Printf("  ✅ Removed Plaesy framework: %s\n", dir)
		}
	}

	platforms, err := c.ResolvePlatforms()
	if err != nil {
		return 0, err
	}
	for _, platform := range platforms {
		if platform == "generic" {
			continue
		}
		if platform == "generic_ai" && !includesGenericAI(c.Opts.Level) {
			continue
		}
		common.LogInfo("Removing platform-specific files for: %s", platform)

		var parentDirs []string
		for _, mt := range mappingTypes {
			target := c.mappingTargetPath(platform, mt)
			if target == "" {
				continue
			}
			full := filepath.Join(c.Opts.TargetDir, target)
			info, err := os.Stat(full)
			if err != nil {
				continue
			}
			if info.IsDir() {
				if err := os.RemoveAll(full); err != nil {
					result.failed = append(result.failed, target)
					fmt.Printf("  \u274c Failed to remove: %s\n", target)
				} else {
					result.removed++
					fmt.Printf("  \u2705 Removed platform %s %s: %s\n", platform, mt, target)
				}
			} else {
				if err := os.Remove(full); err != nil {
					result.failed = append(result.failed, target)
					fmt.Printf("  \u274c Failed to remove: %s\n", target)
				} else {
					result.removed++
					fmt.Printf("  \u2705 Removed platform %s %s: %s\n", platform, mt, target)
				}
			}

			if parent := filepath.Dir(target); parent != "." {
				parentDirs = append(parentDirs, parent)
			}
		}

		for _, parent := range uniqueSorted(parentDirs) {
			full := filepath.Join(c.Opts.TargetDir, parent)
			entries, err := os.ReadDir(full)
			if err == nil && len(entries) == 0 {
				if err := os.Remove(full); err == nil {
					result.removed++
					fmt.Printf("  \u2705 Removed empty parent directory: %s\n", parent)
				}
			}
		}
	}

	if result.removed > 0 {
		common.LogSuccess("Successfully removed %d items", result.removed)
	} else {
		common.LogWarning("No items were removed")
	}

	if len(result.failed) > 0 {
		// Downgrading per-item failures to a warning and returning nil made
		// `plaesy clean` print "completed!" and exit 0 with every mapped
		// file still in place — the likeliest real failure is a Windows file
		// lock or a read-only git-tracked file. The warning stays for the
		// human reading the terminal; the error is for the caller.
		common.LogWarning("Failed to remove some items: %s", strings.Join(result.failed, " "))
		return result.removed, fmt.Errorf("failed to remove %d item(s): %s", len(result.failed), strings.Join(result.failed, " "))
	}

	return result.removed, nil
}

func uniqueSorted(items []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, it := range items {
		if !seen[it] {
			seen[it] = true
			out = append(out, it)
		}
	}
	sort.Strings(out)
	return out
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	mode := os.FileMode(0o644)
	if err == nil {
		mode = info.Mode()
	}
	return os.WriteFile(dst, data, mode)
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}
