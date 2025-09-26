package cmd

import (
	"wg-drill-client/util"

	"github.com/spf13/cobra"
)

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Start wg-drill-client on WireGuard interface",
	Long:  `Start wg-drill-client on WireGuard interface`,
	Run: func(cmd *cobra.Command, args []string) {
		util.CommuDaemon("up " + args[0])
	},
}
