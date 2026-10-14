package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/plaesy/spec-kit/internal/common"
	"github.com/plaesy/spec-kit/internal/installer"
	"github.com/spf13/cobra"
)

func init() {
	register(newInstallCmd())
	register(newUninstallCmd())
	register(newRepairCmd())
	register(newUpgradeCmd())
}

func newInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install the plaesy CLI to a well-known bin directory",
		// No arguments: without this, cobra's ArbitraryArgs default makes a
		// typo look like a successful install.
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			common.PrintBanner("Plaesy Constitution Kit", "installer")

			dst, err := installer.Install()
			if err != nil {
				common.LogError("install failed: %v", err)
				return err
			}

			common.LogSuccess("installed to %s", dst)

			if homeDir, err := installer.PlaesyHomeDir(); err == nil {
				common.LogInfo("Plaesy home ready at %s", homeDir)
			}

			dir, err := installer.InstallDir()
			if err == nil && !installer.OnPath(dir) {
				fmt.Println()
				fmt.Println(installer.PathInstructions(dir))
			}

			return nil
		},
	}
}

func newUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the installed plaesy binary and its Plaesy home",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := installer.InstalledPath()
			if err != nil {
				return err
			}
			appDir, err := installer.AppDataDir()
			if err != nil {
				return err
			}
			_, statErr := os.Stat(path)
			_, homeErr := os.Stat(appDir)
			if os.IsNotExist(statErr) && os.IsNotExist(homeErr) {
				fmt.Printf("plaesy is not installed at %s\n", path)
				return nil
			}

			confirmed := installer.Confirm(os.Stdin, fmt.Sprintf("Remove %s and %s? [y/N]: ", path, appDir))
			if !confirmed {
				fmt.Println("Aborted; nothing was removed.")
				return nil
			}

			removed, err := installer.Uninstall()
			if err != nil {
				common.LogError("uninstall failed: %v", err)
				return err
			}
			common.LogSuccess("removed %s and %s", removed, appDir)
			return nil
		},
	}
}

func newRepairCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "repair",
		Short: "Restore the installed binary and bundled assets from the running build",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rep, err := installer.Repair()
			if err != nil {
				common.LogError("repair failed: %v", err)
				return err
			}
			if rep.BinaryCopied {
				common.LogSuccess("binary restored at %s", rep.BinaryPath)
			} else {
				common.LogInfo("binary at %s already matches this build", rep.BinaryPath)
			}
			common.LogSuccess("Plaesy home refreshed at %s", rep.HomeDir)
			printPathHint(rep)
			return nil
		},
	}
}

func newUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Replace the installed binary with the running (newer) build",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rep, err := installer.Upgrade()
			if errors.Is(err, installer.ErrRunningFromInstall) {
				common.LogInfo("%v", err)
				return nil
			}
			if err != nil {
				common.LogError("upgrade failed: %v", err)
				return err
			}
			if rep.BinaryCopied {
				common.LogSuccess("upgraded %s to this build (version %s)", rep.BinaryPath, common.Version)
			} else {
				common.LogInfo("already up to date: %s matches this build (version %s)", rep.BinaryPath, common.Version)
			}
			common.LogSuccess("Plaesy home refreshed at %s", rep.HomeDir)
			printPathHint(rep)
			return nil
		},
	}
}

// printPathHint mirrors the hint `plaesy install` prints when the install
// directory is missing from PATH.
func printPathHint(rep installer.Report) {
	if rep.OnPath {
		return
	}
	if dir, err := installer.InstallDir(); err == nil {
		fmt.Println()
		fmt.Println(installer.PathInstructions(dir))
	}
}
