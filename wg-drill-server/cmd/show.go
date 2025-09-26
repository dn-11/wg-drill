package cmd

import (
	"wg-drill-server/util"

	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show running iface",
	Long:  `Show running interface in wg-drill-server.`,
	Run: func(cmd *cobra.Command, args []string) {
		util.CommuDaemon("show")
	},
}
