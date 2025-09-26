package cmd

import (
	"wg-drill-client/install"

	"github.com/spf13/cobra"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install wg-drill-client",
	Long:  `Install wg-drill-client.`,
	Run: func(cmd *cobra.Command, args []string) {
		install.Install()
	},
}
