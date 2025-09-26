package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "wg-drill-client",
	Short: "",
	Long:  `A WireGuard tools to make connect under nat for dn11.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func init() {
	RootCmd.AddCommand(DaemonCmd)
	RootCmd.AddCommand(UpCmd)
	RootCmd.AddCommand(DownCmd)
	RootCmd.AddCommand(ShowCmd)
	RootCmd.AddCommand(InstallCmd)
	RootCmd.AddCommand(UnInstallCmd)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
	}
}
