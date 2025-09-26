package cmd

import (
	"wg-drill-server/daemon"

	"github.com/spf13/cobra"
)

var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run natdrill daemon",
	Run: func(cmd *cobra.Command, args []string) {
		daemon.Run()
	},
}
