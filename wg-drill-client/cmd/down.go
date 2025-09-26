package cmd

import (
	"wg-drill-client/util"

	"github.com/spf13/cobra"
)

var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop wg-drill-client on WireGuard interface",
	Long:  `Stop wg-drill-client on WireGuard interface`,
	Run: func(cmd *cobra.Command, args []string) {
		util.CommuDaemon("down " + args[0])
	},
}
