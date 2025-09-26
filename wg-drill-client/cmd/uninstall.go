package cmd

import (
	"wg-drill-client/install"

	"github.com/spf13/cobra"
)

var UnInstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "UnInstall wg-drill-client",
	Long:  `UnInstall wg-drill-client.`,
	Run: func(cmd *cobra.Command, args []string) {
		install.UnInstall()
	},
}
