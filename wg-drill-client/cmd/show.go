package cmd

import (
	"wg-drill-client/util"

	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show wg-drill-client status on WireGuard interface",
	Long:  `Show wg-drill-client status on WireGuard interface`,
	Run: func(cmd *cobra.Command, args []string) {
		util.CommuDaemon("show")
	},
}
