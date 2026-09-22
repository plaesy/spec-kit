package main

import (
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
		RunE: func(cmd *cobra.Command, args []string) error {
			common.PrintBanner("Plaesy Constitution Kit", "installer")

			dst, err := installer.Install()
			if err != nil {
				common.LogError("install failed: %v", err)
				return err
			}

			common.LogSuccess("installed to %s", dst)

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
		Short: "Remove the installed plaesy CLI binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := installer.InstalledPath()
			if err != nil {
				return err
			}
			if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
				fmt.Printf("plaesy is not installed at %s\n", path)
				return nil
			}

			confirmed := installer.Confirm(os.Stdin, fmt.Sprintf("Remove %s? [y/N]: ", path))
			if !confirmed {
				fmt.Println("Aborted; nothing was removed.")
				return nil
			}

			removed, err := installer.Uninstall()
			if err != nil {
				common.LogError("uninstall failed: %v", err)
				return err
			}
			common.LogSuccess("removed %s", removed)
			return nil
		},
	}
}

func newRepairCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "repair",
		Short: "Repair the plaesy installation (not yet implemented)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(installer.NotImplementedMessage)
			return nil
		},
	}
}

func newUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the plaesy installation (not yet implemented)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(installer.NotImplementedMessage)
			return nil
		},
	}
}
