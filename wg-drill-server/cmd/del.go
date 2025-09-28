package cmd

import (
	"fmt"
	"wg-drill-server/util"

	"github.com/spf13/cobra"
)

var DelCmd = &cobra.Command{
	Use:   "del",
	Short: "del a WireGuard peer to the running daemon",
	Long:  `del exchanging endpoint with peer.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Usage: del <interface>")
		}
		util.CommuDaemon("del " + args[0])
		return nil
	},
}
