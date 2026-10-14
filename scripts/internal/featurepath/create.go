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
// number, creates a NNN-slug branch, and seeds .plaesy/specs/<branch>/spec.md from
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

	specsDir := filepath.Join(repoRoot, ".plaesy", "specs")
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

	specFile := filepath.Join(featureDir, "spec.md")
	if err := seedSpecFromTemplate(repoRoot, specFile); err != nil {
		return nil, err
	}

	return &CreateResult{BranchName: branchName, SpecFile: specFile, FeatureNum: featureNum}, nil
}

// specTemplateName is the file seeded into .plaesy/specs/<branch>/spec.md.
const specTemplateName = "spec.template.md"

// specTemplateCandidates lists, in order, where the spec template may live
// relative to the repository root.
//
// The installed location comes first: `plaesy init` copies templates into
// <repo>/.plaesy/templates (see scaffold.copyTemplates). The bare
// <repo>/templates path is the spec-kit source tree itself, which is only
// correct when the framework is dogfooding its own repository.
//
// This ordering matters: reading <repo>/templates unconditionally made every
// installed project silently produce a zero-byte spec.md, because the read
// failed and the error was discarded.
func specTemplateCandidates(repoRoot string) []string {
	return []string{
		filepath.Join(repoRoot, ".plaesy", "templates", specTemplateName),
		filepath.Join(repoRoot, "templates", specTemplateName),
	}
}

// seedSpecFromTemplate copies the spec template to specFile.
//
// A missing template is a hard error: silently writing an empty spec.md
// produced artifacts that looked complete but carried no content, and the
// caller had no signal that anything was wrong.
func seedSpecFromTemplate(repoRoot, specFile string) error {
	var tried []string
	for _, template := range specTemplateCandidates(repoRoot) {
		tried = append(tried, template)
		data, err := os.ReadFile(template)
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("reading spec template %s: %w", template, err)
			}
			continue
		}
		if err := os.WriteFile(specFile, data, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", specFile, err)
		}
		return nil
	}
	return fmt.Errorf(
		"spec template %s not found; looked in:\n  %s\nrun `plaesy init <dir> --ai <platform>` to install the framework before creating a feature",
		specTemplateName, strings.Join(tried, "\n  "),
	)
}
