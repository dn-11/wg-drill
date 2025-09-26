package cmd

import (
	"fmt"
	"wg-drill-server/util"

	"github.com/spf13/cobra"
)

var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "down server",
	Long:  `down exchanging endpoint with peer.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Usage: down <interface>")
		}
		util.CommuDaemon("down " + args[0])
		return nil
	},
}
