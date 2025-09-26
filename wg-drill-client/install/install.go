package install

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"wg-drill-client/util"
)

const binPath = "/usr/local/bin/wg-drill-client"
const daemonPath = "/usr/local/bin/wg-drill-client daemon"
const dirPath = "/etc/wg-drill-client"
const configPath = "/etc/wg-drill-client/config.toml"

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
Description=wg-drill-client
After=network.target

[Service]
ExecStart=` + daemonPath + `
Restart=always

[Install]
WantedBy=multi-user.target
`

		servicePath := "/etc/systemd/system/wg-drill-client.service"
		os.WriteFile(servicePath, []byte(service), 0644)
		exec.Command("systemctl", "daemon-reload").Run()
		exec.Command("systemctl", "enable", "wg-drill-client").Run()
		exec.Command("systemctl", "start", "wg-drill-client").Run()
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

		scriptPath := "/etc/init.d/wg-drill-client"
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
endpoing = "example.com:14514"

[drill]
enable = true
iface = []
interval = 10
`
	if err := os.WriteFile(configPath, []byte(configfile), 0644); err != nil {
		fmt.Printf("fail to write config file: %v\n", err)
		return
	}
	fmt.Println("config file at /etc/wg-drill-client/config.toml")
}

func UnInstall() {
	platform := util.JudgePlatform()
	switch platform {
	case "systemd":
		servicePath := "/etc/systemd/system/wg-drill-client.service"
		exec.Command("systemctl", "stop", "wg-drill-client").Run()
		exec.Command("systemctl", "disable", "wg-drill-client").Run()
		os.Remove(servicePath)
		exec.Command("systemctl", "daemon-reload").Run()
		fmt.Println("systemd uninstalled")
	case "procd":
		scriptPath := "/etc/init.d/wg-drill-client"
		exec.Command(scriptPath, "stop").Run()
		exec.Command(scriptPath, "disable").Run()
		os.Remove(scriptPath)
		fmt.Println("procd uninstalled")
	default:
		fmt.Println("unsupported platform type")
	}
	if err := os.Remove(binPath); err != nil {
		fmt.Printf("fail to remove binary file: %v\n", err)
	}
	fmt.Println("binary file removed")
	if err := os.RemoveAll(dirPath); err != nil {
		fmt.Printf("fail to remove config file: %v\n", err)
	}
	fmt.Println("config file removed")
}
