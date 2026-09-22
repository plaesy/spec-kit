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
var mappingTypes = []string{"core", "instructions", "prompts", "chatmodes"}

// plaesyDirs mirrors the bash script's plaesy_dirs array. docs/ is
// deliberately excluded, as it may hold project documentation.
var plaesyDirs = []string{".plaesy", "specs"}

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
	info, err := os.Stat(c.Opts.TargetDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("target directory '%s' does not exist", c.Opts.TargetDir)
	}
	f, err := os.Open(c.Opts.TargetDir)
	if err != nil {
		return fmt.Errorf("target directory '%s' is not readable", c.Opts.TargetDir)
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
func DetectAllPlatforms(cfg *config.PlatformConfig, configPath, targetDir string) []string {
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
			if _, err := os.Stat(filepath.Join(targetDir, pattern)); err == nil {
				found = append(found, name)
				break
			}
		}
	}
	if len(found) == 0 {
		return []string{"generic_ai"}
	}
	return found
}

// ResolvePlatforms returns the platforms to act on: AIChoice split on
// whitespace if set, else auto-detection, else ["generic"] — mirroring the
// bash script's repeated "$AI_CHOICE else detect_all_platforms" pattern.
func (c *Cleaner) ResolvePlatforms() []string {
	if c.Opts.AIChoice != "" {
		return strings.Fields(c.Opts.AIChoice)
	}
	platforms := DetectAllPlatforms(c.Config, c.ConfigPath, c.Opts.TargetDir)
	if len(platforms) == 0 {
		return []string{"generic"}
	}
	return platforms
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
	PlaesyDirs      []string          // found top-level Plaesy dirs (.plaesy, specs)
	PlatformEntries map[string][]string // platform -> found target paths (display form, "/" suffixed for dirs)
	Empty           bool              // true if nothing was found to remove
}

// BuildPlan mirrors show_cleanup_plan's analysis phase (without printing).
func (c *Cleaner) BuildPlan() Plan {
	platforms := c.ResolvePlatforms()
	plan := Plan{Platforms: platforms, PlatformEntries: map[string][]string{}}

	for _, dir := range plaesyDirs {
		if info, err := os.Stat(filepath.Join(c.Opts.TargetDir, dir)); err == nil && info.IsDir() {
			plan.PlaesyDirs = append(plan.PlaesyDirs, dir)
		}
	}

	for _, platform := range platforms {
		if platform == "generic" || platform == "generic_ai" {
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
		if platform == "generic" || platform == "generic_ai" {
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

	for _, platform := range c.ResolvePlatforms() {
		if platform == "generic" || platform == "generic_ai" {
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
// core/instructions/prompts/chatmodes targets, then prunes now-empty parent
// directories of those targets. Returns the number of items removed.
func (c *Cleaner) Remove() (int, error) {
	result := removeResult{}

	common.LogInfo("Removing plaesy directories based on platform.json mapping...")
	common.LogInfo("Removing Plaesy framework directories...")
	for _, dir := range plaesyDirs {
		full := filepath.Join(c.Opts.TargetDir, dir)
		if info, err := os.Stat(full); err == nil && info.IsDir() {
			if err := os.RemoveAll(full); err != nil {
				result.failed = append(result.failed, dir)
				fmt.Printf("  \u274c Failed to remove: %s\n", dir)
			} else {
				result.removed++
				fmt.Printf("  \u2705 Removed Plaesy framework: %s\n", dir)
			}
		}
	}

	platforms := c.ResolvePlatforms()
	for _, platform := range platforms {
		if platform == "generic" || platform == "generic_ai" {
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

	if len(result.failed) > 0 {
		common.LogWarning("Failed to remove some items: %s", strings.Join(result.failed, " "))
	}
	if result.removed > 0 {
		common.LogSuccess("Successfully removed %d items", result.removed)
	} else {
		common.LogWarning("No items were removed")
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
