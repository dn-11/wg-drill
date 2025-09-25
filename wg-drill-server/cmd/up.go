package cmd

import (
	"bufio"
	"fmt"
	"net"

	"github.com/spf13/cobra"
)

var UpCmd = &cobra.Command{
	Use:   "up",
	Short: "Add a WireGuard interface to the running daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Please provide a interface name")
		}
		//addIface := args[0]
		conn, err := net.Dial("unix", "/var/run/wg-natdrill.sock")
		if err != nil {
			return fmt.Errorf("connect to daemon socket failed: %w", err)
		}
		defer func() { _ = conn.Close() }()
		if _, err := fmt.Fprintf(conn, "ADD %s\n", args[0]); err != nil {
			return fmt.Errorf("send command failed: %w", err)
		}
		reader := bufio.NewReader(conn)
		resp, _ := reader.ReadString('\n')
		fmt.Print(resp)
		return nil
	}
}
