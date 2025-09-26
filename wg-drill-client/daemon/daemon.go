package daemon

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
	"wg-drill-client/config"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const SocketPath = "/var/run/wg-drill-client.sock"

type daemon struct {
	ifaces []string
	lock   sync.RWMutex
}

func newDaemon() *daemon {
	d := &daemon{}
	d.ifaces = config.Drill.Iface
	return d
}

func getEndpoint(pubkey string) (*net.UDPAddr, error) {
	encoded := url.QueryEscape(pubkey)
	requestUrl := "https://" + config.Server.Endpoint + encoded
	resp, err := http.Get(requestUrl)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, nil
	}
	body, err := io.ReadAll(resp.Body)
	addr, err := net.ResolveUDPAddr("udp", string(body))
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (d *daemon) Sync() {
	client, err := wgctrl.New()
	if err != nil {
		panic(err)
	}
	for {
		for _, iface := range d.ifaces {
			device, err := client.Device(iface)
			if err != nil {
				fmt.Printf("Failed to get device %s for %s: %s\n", iface, iface, err)
				continue
			}
			var peers []wgtypes.PeerConfig
			for _, peer := range device.Peers {
				addr, err := getEndpoint(peer.PublicKey.String())
				if err != nil {
					fmt.Printf("Failed to get endpoint for %s: %s\n", peer.PublicKey.String(), err)
					continue
				}
				peers = append(peers, wgtypes.PeerConfig{
					Endpoint: addr,
				})
			}
			config := wgtypes.Config{
				Peers: peers,
			}
			err = client.ConfigureDevice(iface, config)
			if err != nil {
				fmt.Printf("Failed to configure device %s: %s\n", iface, err)
				continue
			}

		}
		time.Sleep(time.Duration(config.Drill.Interval) * time.Second)
	}
	client.Close()
}

func (d *daemon) commu() { // 与CLI通信
	os.Remove(SocketPath)
	ln, err := net.Listen("unix", SocketPath)
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 1024)
			n, _ := c.Read(buf)
			cmd := string(buf[:n])
			params := strings.Fields(cmd)
			message := ""
			switch params[0] {
			case "up": //todo
				if len(params) != 2 {
					message += "Usage: up <interface>\n"
					return
				} else {
					d.lock.Lock()

					d.lock.Unlock()
				}
			case "down": //todo
				if len(params) != 2 {
					message += "Usage: down <interface>\n"
					return
				} else {
					d.lock.Lock()

					d.lock.Unlock()
				}
			case "show": //todo
				d.lock.RLock()

				d.lock.RUnlock()
			default:
				message += "Unknown command\n"
			}
			//message += "\n"
			c.Write([]byte(message))
		}(conn)
	}
}

func Run() {
	config.Init()
	d := newDaemon()
	go d.Sync()
	go d.commu()
	select {}
}
