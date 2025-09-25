package util

import (
	"encoding/base32"
	"encoding/base64"
	"errors"
	"log"
	"os/exec"
	"path/filepath"

	"golang.zx2c4.com/wireguard/wgctrl"
)

var scriptPath string

func SetEndpoint(iface string) error {
	wgClient, err := wgctrl.New()
	if err != nil {
		return err
	}
	wgDevice, err := wgClient.Device(iface)
	if err != nil {
		return err
	}
	if len(wgDevice.Peers) < 1 {
		log.Println("no peers found")
		return errors.New("no peers found")
	}
	for _, peer := range wgDevice.Peers {
		pubkey, err := Base64ToBase32(peer.PublicKey.String())
		if err != nil {
			return err
		}
		filename := "example.sh"
		scriptPath := filepath.Join("/", "etc", "natdrill", filename)
		cmd := exec.Command("/bin/sh", scriptPath, pubkey, peer.Endpoint.String())
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func Base64ToBase32(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(data), nil
}
