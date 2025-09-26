package util

import (
	"os"
	"strings"
)

func JudgePlatform() string {
	if data, err := os.ReadFile("/proc/1/comm"); err == nil {
		name := strings.TrimSpace(string(data))
		if strings.EqualFold(name, "procd") {
			return "procd"
		} else if strings.EqualFold(name, "systemd") {
			return "systemd"
		}
	}
	return "unknown"
}
