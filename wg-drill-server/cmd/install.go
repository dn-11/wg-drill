package cmd

import (
	"wg-drill-server/install"

	"github.com/spf13/cobra"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install wg-drill-server",
	Long:  `Install wg-drill-server.`,
	Run: func(cmd *cobra.Command, args []string) {
		install.Install()
	},
}
