// Package daemon exposes Run to start the background workers and unix control server.
package daemon

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"wg-natdrill/config"
	"wg-natdrill/util"
)

const socketPath = "/var/run/wg-natdrill.sock"

var (
	started = make(map[string]bool)
	mu      sync.Mutex
)

func startIfaceWorker(iface string, interval time.Duration) bool {
	mu.Lock()
	if started[iface] {
		mu.Unlock()
		return false
	}
	started[iface] = true
	mu.Unlock()
	go func() {
		for {
			if err := util.SetEndpoint(iface); err != nil {
				log.Printf("SetEndpoint error for %s: %v", iface, err)
			}
			time.Sleep(interval)
		}
	}()
	log.Printf("started worker for iface: %s", iface)
	return true
}

func scanConfigLoop(interval time.Duration) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		ifaces := config.GetViper().GetStringSlice("ifaces")
		for _, iface := range ifaces {
			startIfaceWorker(iface, interval)
		}
		<-ticker.C
	}
}

func handleConn(c net.Conn, interval time.Duration) {
	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("close conn error: %v", err)
		}
	}()
	reader := bufio.NewReader(c)
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) >= 2 && strings.ToUpper(fields[0]) == "ADD" {
		iface := fields[1]
		if iface == "" {
			if _, err := fmt.Fprintln(c, "ERR missing iface"); err != nil {
				log.Printf("write response error: %v", err)
			}
			return
		}
		newStarted := startIfaceWorker(iface, interval)
		if newStarted {
			if _, err := fmt.Fprintln(c, "OK"); err != nil {
				log.Printf("write response error: %v", err)
			}
		} else {
			if _, err := fmt.Fprintln(c, "EXISTS"); err != nil {
				log.Printf("write response error: %v", err)
			}
		}
		return
	}
	if _, err := fmt.Fprintln(c, "ERR unknown command"); err != nil {
		log.Printf("write response error: %v", err)
	}
}

func listenAndServeUnix(sock string, interval time.Duration) error {
	if err := os.MkdirAll(filepath.Dir(sock), 0755); err != nil {
		return err
	}
	if st, err := os.Stat(sock); err == nil && (st.Mode()&os.ModeSocket) != 0 {
		// If socket exists, try connect to detect a running daemon
		if conn, err := net.DialTimeout("unix", sock, 200*time.Millisecond); err == nil {
			_ = conn.Close()
			return fmt.Errorf("daemon already running (socket %s)", sock)
		}
		_ = os.Remove(sock)
	}
	ln, err := net.Listen("unix", sock)
	if err != nil {
		return err
	}
	_ = os.Chmod(sock, 0660)
	log.Printf("control socket listening on %s", sock)
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			return err
		}
		go handleConn(conn, interval)
	}
}

// Run starts the daemon and blocks until SIGINT/SIGTERM.
func Run() {
	config.Init()
	intervalSec := config.GetViper().GetInt("interval")
	if intervalSec <= 0 {
		intervalSec = 60
	}
	interval := time.Duration(intervalSec) * time.Second

	// start from config
	for _, iface := range config.GetViper().GetStringSlice("ifaces") {
		startIfaceWorker(iface, interval)
	}
	// periodic scan for new ifaces in config
	go scanConfigLoop(interval)
	// control server
	go func() {
		if err := listenAndServeUnix(socketPath, interval); err != nil {
			log.Fatalf("control server error: %v", err)
		}
	}()

	// graceful shutdown
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	_ = os.Remove(socketPath)
	log.Println("daemon exiting")
}
