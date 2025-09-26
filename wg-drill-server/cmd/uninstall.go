package cmd

import (
	"wg-drill-server/install"

	"github.com/spf13/cobra"
)

var UnInstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall wg-drill-server",
	Long:  `Uninstall wg-drill-server.`,
	Run: func(cmd *cobra.Command, args []string) {
		install.UnInstall()
	},
}
