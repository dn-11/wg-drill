package wg

import (
	"wg-drill-server/config"

	"github.com/spf13/viper"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func StartWg() error {
	// Placeholder for wg-related utility functions
	ifaceName := config.Drill.Iface

	wg := &netlink.Wireguard{
		LinkAttrs: netlink.LinkAttrs{Name: ifaceName},
	}
	err := netlink.LinkAdd(wg)
	if err != nil {
		return err
	}

	key, err := wgtypes.ParseKey(config.WireGuard.PrivateKey)
	if err != nil {
		return err
	}
	client, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer client.Close()

	cfg := wgtypes.Config{
		PrivateKey: &key,
		Peers:      []wgtypes.PeerConfig{},
	}

	for _, peerPubKeyStr := range config.WireGuard.PeerPubkeys {
		peerKey, err := wgtypes.ParseKey(peerPubKeyStr)
		if err != nil {
			return err
		}
		peerCfg := wgtypes.PeerConfig{
			PublicKey: peerKey,
			// Additional peer settings can be added here
		}
		cfg.Peers = append(cfg.Peers, peerCfg)
	}

	if err := client.ConfigureDevice(ifaceName, cfg); err != nil {
		return err
	}
	return nil
}

func StopWg() error {
	ifaceName := config.Drill.Iface

	client, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer client.Close()

	device, err := client.Device(ifaceName)

	var peers []string

	for _, peer := range device.Peers {
		_ = append(peers, peer.PublicKey.String())
	}
	config.WireGuard.PeerPubkeys = peers
	viper.Set("wireguard.peer_pubkeys", peers)

	l, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return err
	}
	err = netlink.LinkDel(l)
	if err != nil {
		return err
	}
	return nil
}
