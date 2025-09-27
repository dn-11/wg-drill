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
	requestUrl := "http://" + config.Server.Endpoint + "/?pubkey=" + encoded
	resp, err := http.Get(requestUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, nil
	}
	body, err := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	addr, err := net.ResolveUDPAddr("udp", string(body))
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (d *daemon) RemoveIface(iface string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	for i, v := range d.ifaces {
		if v == iface {
			d.ifaces = append(d.ifaces[:i], d.ifaces[i+1:]...)
			break
		}
	}
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func (d *daemon) Sync() {
	for {
		d.lock.RLock()
		client, err := wgctrl.New()
		if err != nil {
			//fmt.Printf("Failed to open wgctrl: %s\n", err)
			d.lock.RUnlock()
			time.Sleep(time.Duration(config.Drill.Interval) * time.Second)
			continue
		}
		for _, iface := range d.ifaces {
			device, err := client.Device(iface)
			if err != nil {
				//fmt.Printf("Failed to get device %s for %s: %s\n", iface, iface, err)
				continue
			}
			for _, peer := range device.Peers {
				addr, err := getEndpoint(peer.PublicKey.String())
				if err != nil {
					//fmt.Printf("Failed to get endpoint for %s: %s\n", peer.PublicKey.String(), err)
					continue
				}
				//fmt.Printf("Found peer %s with endpoint %s\n", peer.PublicKey, addr.String())
				peerConfig := wgtypes.PeerConfig{
					PublicKey:  peer.PublicKey,
					UpdateOnly: true,
					Endpoint:   addr,
				}
				deviceConfig := wgtypes.Config{
					PrivateKey:   &device.PrivateKey,
					ReplacePeers: false,
					Peers:        []wgtypes.PeerConfig{peerConfig},
				}
				if device.FirewallMark > 0 {
					deviceConfig.FirewallMark = &device.FirewallMark
				}
				err = client.ConfigureDevice(iface, deviceConfig)
				if err != nil {
					//fmt.Printf("Failed to configure device %s for %s: %s\n", iface, iface, err)
					continue
				}
			}

		}
		client.Close()
		d.lock.RUnlock()
		time.Sleep(time.Duration(config.Drill.Interval) * time.Second)
	}
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
			case "up":
				if len(params) != 2 {
					message += "Usage: up <interface>\n"
					return
				} else {
					for _, iface := range params[1:] {
						if contains(d.ifaces, iface) {
							message += iface + " already exists\n"
							continue
						}
						d.lock.Lock()
						d.ifaces = append(d.ifaces, iface)
						message += "append:" + iface + "\n"
						d.lock.Unlock()

					}
				}
			case "down":
				if len(params) != 2 {
					message += "Usage: down <interface>\n"
					return
				} else {
					for _, iface := range params[1:] {
						d.RemoveIface(iface)
						message += "remove:" + iface + "\n"
					}
				}
			case "show": //todo
				d.lock.RLock()
				message += "Interfaces:\n"
				for _, iface := range d.ifaces {
					message += "  " + iface + "\n"
				}
				d.lock.RUnlock()
			default:
				message += "Unknown command\n"
			}
			//message += "\n"
			_, _ = c.Write([]byte(message))
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
