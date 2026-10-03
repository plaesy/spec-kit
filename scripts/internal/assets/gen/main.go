// Command gen syncs scripts/internal/assets/data from its real sources:
// the repo-root templates/, instructions/, prompts/, agents/, checklists/
// directories, and scripts/configs. Run it with the scripts/ module as the
// working directory:
//
//	cd scripts && go run ./internal/assets/gen
//
// `make build` and `make assets` do this automatically; run it manually
// after editing any of those source directories and before a plain
// `go build`, or TestAssetsDataMatchesSource will fail.
package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var topLevelDirs = []string{"templates", "instructions", "prompts", "agents", "checklists"}

func main() {
	wd, err := os.Getwd()
	must(err)
	repoRoot := filepath.Dir(wd)
	destRoot := filepath.Join(wd, "internal", "assets", "data")

	must(os.RemoveAll(destRoot))

	for _, d := range topLevelDirs {
		must(copyTree(filepath.Join(repoRoot, d), filepath.Join(destRoot, d)))
	}
	must(copyTree(filepath.Join(wd, "configs"), filepath.Join(destRoot, "scripts", "configs")))

	fmt.Println("synced assets into", destRoot)
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
