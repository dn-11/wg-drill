package cmd

import (
	"wg-drill-client/daemon"

	"github.com/spf13/cobra"
)

var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start daemon service for wg-drill-client",
	Long:  `Start daemon service for wg-drill-client`,
	Run: func(cmd *cobra.Command, args []string) {
		daemon.Run()
	},
}
