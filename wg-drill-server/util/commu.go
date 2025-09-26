package util

import (
	"fmt"
	"net"
	"wg-drill-server/daemon"
)

func CommuDaemon(cmd string) {
	conn, err := net.Dial("unix", daemon.SocketPath)
	if err != nil {
		fmt.Printf("Error connecting to daemon socket: %v\n", err)
		return
	}
	defer conn.Close()
	_, _ = conn.Write([]byte(cmd + "\n"))
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	println(string(buf[:n]))
}
