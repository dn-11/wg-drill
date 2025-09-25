package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "wg-natdrill-plugin",
	Short: "",
	Long:  `A WireGuard tools to make connect under nat for dn11.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func init() {
	RootCmd.AddCommand(DownCmd)
	RootCmd.AddCommand(DaemonCmd)
	RootCmd.AddCommand(UpCmd)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println("root")
	}
}
