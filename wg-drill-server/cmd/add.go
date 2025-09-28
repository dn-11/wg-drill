package cmd

import (
	"fmt"
	"wg-drill-server/util"

	"github.com/spf13/cobra"
)

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a WireGuard peer to the running daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Usage: add <interface>")
		}
		util.CommuDaemon("add " + args[0])
		return nil
	},
}
