package install

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"wg-drill-server/util"
)

var binPath = ""
var scriptPath = ""
var daemonPath = ""

const dirPath = "/etc/wg-drill-server"
const configPath = "/etc/wg-drill-server/config.toml"

func initPlatform() {
	systemPlatform := util.JudgePlatform()
	if systemPlatform == "procd" {
		binPath = "/usr/sbin/wg-drill-server"
		scriptPath = "/etc/init.d/wg-drill-server"
		daemonPath = binPath + " daemon"
	} else if systemPlatform == "systemd" {
		binPath = "/usr/local/bin/wg-drill-server"
		scriptPath = "/etc/systemd/system/wg-drill-server.service"
		daemonPath = binPath + " daemon"
	}
}

func Install() {
	file, err := exec.LookPath(os.Args[0])
	if err != nil && !errors.Is(err, exec.ErrDot) {
		fmt.Printf("fail to get binary file path: %v\n", err)
		return
	}
	absFile, err := filepath.Abs(file)
	if err != nil {
		fmt.Printf("fail to get binary file path: %v\n", err)
		return
	}
	fmt.Printf("file at: %v\n", absFile)

	originFp, err := os.Open(absFile)
	if err != nil {
		fmt.Printf("file to opne binary file: %v\n", err)
		return
	}
	defer originFp.Close()

	if _, err := os.Stat(binPath); err == nil {
		if err := os.Remove(binPath); err != nil {
			fmt.Printf("fail to remove former file: %v\n", err)
			return
		}
	}

	fp, err := os.OpenFile(binPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		fmt.Printf("fail to write target path: %v\n", err)
		return
	}
	defer fp.Close()
	if _, err = io.Copy(fp, originFp); err != nil {
		_ = os.Remove(binPath)
		fmt.Printf("fail to copy binary file: %v\n", err)
		return
	}
	fmt.Printf("copy binary file to %s", binPath)

	platform := util.JudgePlatform()
	switch platform {
	case "systemd":

		service := `[Unit]
Description=wg-drill-server
After=network.target

[Service]
ExecStart=` + daemonPath + `
Restart=always

[Install]
WantedBy=multi-user.target
`

		servicePath := "/etc/systemd/system/wg-drill-server.service"
		os.WriteFile(servicePath, []byte(service), 0644)
		exec.Command("systemctl", "daemon-reload").Run()
		exec.Command("systemctl", "enable", "wg-drill-server").Run()
		exec.Command("systemctl", "start", "wg-drill-server").Run()
		fmt.Println("systemd installed")
	case "procd":

		initScript := `#!/bin/sh /etc/rc.common
START=99
USE_PROCD=1

start_service() {
	procd_open_instance
	procd_set_param command ` + daemonPath + `
	procd_close_instance
}
`

		os.WriteFile(scriptPath, []byte(initScript), 0755)
		exec.Command("chmod", "+x", scriptPath).Run()
		exec.Command(scriptPath, "enable").Run()
		exec.Command(scriptPath, "start").Run()
		fmt.Println("procd installed")
	default:
		fmt.Println("unsupported platform type")
	}

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		fmt.Printf("make directory failed: %v\n", err)
		return
	}

	configfile := `
[server]
listenaddr = "0.0.0.0"
listenport = 14514

[drill]
enable = true
iface = []
interval = 10
`
	if err := os.WriteFile(configPath, []byte(configfile), 0644); err != nil {
		fmt.Printf("fail to write config file: %v\n", err)
		return
	}
	fmt.Println("config file at /etc/wg-drill-server/config.toml")
}

func UnInstall() {
	initPlatform()
	if err := os.Remove(scriptPath); err != nil {
		fmt.Printf("fail to remove init script: %v\n", err)
	} else {
		fmt.Println("script removed")
	}
	if err := os.Remove(binPath); err != nil {
		fmt.Printf("fail to remove binary file: %v\n", err)
	} else {
		fmt.Println("binary file removed")
	}
	if err := os.RemoveAll(dirPath); err != nil {
		fmt.Printf("fail to remove config file: %v\n", err)
	} else {
		fmt.Println("config file removed")
	}
}
