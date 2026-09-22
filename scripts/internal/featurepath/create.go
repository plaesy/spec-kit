// Package featurepath implements the feature-branch workflow helpers ported
// from create-new-feature.sh, get-feature-paths.sh and
// check-task-prerequisites.sh.
package featurepath

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/plaesy/spec-kit/internal/common"
)

// CreateResult mirrors the fields emitted by create-new-feature.sh.
type CreateResult struct {
	BranchName string
	SpecFile   string
	FeatureNum string
}

var (
	leadingNumberRE = regexp.MustCompile(`^[0-9]+`)
	nonAlnumRE      = regexp.MustCompile(`[^a-z0-9]+`)
)

// CreateNewFeature mirrors create-new-feature.sh: computes the next feature
// number, creates a NNN-slug branch, and seeds specs/<branch>/spec.md from
// templates/spec.template.md.
func CreateNewFeature(description string) (*CreateResult, error) {
	description = strings.TrimSpace(description)
	if description == "" {
		return nil, fmt.Errorf("feature description is required")
	}

	repoRoot, err := common.GetRepoRoot()
	if err != nil {
		return nil, err
	}

	specsDir := filepath.Join(repoRoot, "specs")
	if err := os.MkdirAll(specsDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating specs dir: %w", err)
	}

	highest := 0
	entries, _ := os.ReadDir(specsDir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		match := leadingNumberRE.FindString(e.Name())
		if match == "" {
			continue
		}
		n, err := strconv.Atoi(match)
		if err == nil && n > highest {
			highest = n
		}
	}

	featureNum := fmt.Sprintf("%03d", highest+1)

	slug := strings.ToLower(description)
	slug = nonAlnumRE.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	words := []string{}
	for _, w := range strings.Split(slug, "-") {
		if w == "" {
			continue
		}
		words = append(words, w)
		if len(words) == 3 {
			break
		}
	}

	branchName := featureNum + "-" + strings.Join(words, "-")

	if out, err := exec.Command("git", "checkout", "-b", branchName).CombinedOutput(); err != nil {
		return nil, fmt.Errorf("git checkout -b %s: %w: %s", branchName, err, string(out))
	}

	featureDir := filepath.Join(specsDir, branchName)
	if err := os.MkdirAll(featureDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating feature dir: %w", err)
	}

	template := filepath.Join(repoRoot, "templates", "spec.template.md")
	specFile := filepath.Join(featureDir, "spec.md")
	if data, err := os.ReadFile(template); err == nil {
		if err := os.WriteFile(specFile, data, 0o644); err != nil {
			return nil, fmt.Errorf("writing spec.md: %w", err)
		}
	} else {
		if err := os.WriteFile(specFile, nil, 0o644); err != nil {
			return nil, fmt.Errorf("creating spec.md: %w", err)
		}
	}

	return &CreateResult{BranchName: branchName, SpecFile: specFile, FeatureNum: featureNum}, nil
}
