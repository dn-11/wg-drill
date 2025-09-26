package cmd

import (
	"fmt"
	"wg-drill-server/util"

	"github.com/spf13/cobra"
)

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Add a WireGuard interface to the running daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Usage: up <interface>")
		}
		util.CommuDaemon("up " + args[0])
		return nil
	},
}
