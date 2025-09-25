package cmd

import (
	"github.com/spf13/cobra"
)

var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "Start server",
	Long:  `Start exchanging endpoint with peer.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
