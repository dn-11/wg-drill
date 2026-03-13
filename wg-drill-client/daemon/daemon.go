package daemon

import (
	"fmt"
	"io"
	"math/rand"
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
	ifaces      []string
	lock        sync.RWMutex
	MinRandPort int
	MaxRandPort int
	ifaceStart  map[string]time.Time
}

func checkPort(port int) (bool, error) {
	addr := fmt.Sprintf(":%d", port)
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return false, err
	} else {
		conn.Close()
	}
	return true, nil
}

func newDaemon() *daemon {
	d := &daemon{}
	d.ifaces = config.Drill.Iface
	d.ifaceStart = make(map[string]time.Time)
	now := time.Now()
	for _, iface := range d.ifaces {
		d.ifaceStart[iface] = now
	}
	min := config.Drill.MinRandPort
	max := config.Drill.MaxRandPort
	if min <= 0 || max <= 0 || max < min {
		min = 40000
		max = 65535
	}
	d.MinRandPort = min
	d.MaxRandPort = max
	return d
}

func getEndpoint(endpoint *net.UDPAddr, pubkey string) (*net.UDPAddr, error) {
	encoded := url.QueryEscape(pubkey)
	requestUrl := "http://" + endpoint.String() + "/?pubkey=" + encoded
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
	//fmt.Println(string(body))
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
			delete(d.ifaceStart, iface)
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
		client, err := wgctrl.New()
		if err != nil {
			time.Sleep(time.Duration(config.Drill.Interval) * time.Second)
			continue
		}

		timeoutDur := time.Duration(config.Drill.Timeout) * time.Second
		startupGrace := 30 * time.Second
		if timeoutDur*2 > startupGrace {
			startupGrace = timeoutDur * 2
		}

		// Snapshot ifaces + ifaceStart under write lock, then process unlocked.
		d.lock.Lock()
		ifaces := append([]string(nil), d.ifaces...)
		startMap := make(map[string]time.Time, len(ifaces))
		now := time.Now()
		for _, iface := range ifaces {
			t, ok := d.ifaceStart[iface]
			if !ok {
				t = now
				d.ifaceStart[iface] = t
			}
			startMap[iface] = t
		}
		d.lock.Unlock()

		for _, iface := range ifaces {
			ifaceStart := startMap[iface]

			device, err := client.Device(iface)
			if err != nil {
				continue
			}

			var stunendpoint *net.UDPAddr
			for _, peer := range device.Peers {
				if peer.AllowedIPs == nil || len(peer.AllowedIPs) == 0 {
					stunendpoint = peer.Endpoint
				}
			}

			if stunendpoint != nil {
				fmt.Println("Interface:", iface, "Using STUN endpoint:", stunendpoint.String())
			} else {
				fmt.Println("Interface:", iface, "Using STUN endpoint: <nil>")
			}

			deviceConfig := wgtypes.Config{
				PrivateKey:   &device.PrivateKey,
				ReplacePeers: false,
				Peers:        []wgtypes.PeerConfig{},
			}
			if device.FirewallMark > 0 {
				deviceConfig.FirewallMark = &device.FirewallMark
			}

			for _, peer := range device.Peers {
				if (peer.AllowedIPs == nil || len(peer.AllowedIPs) == 0) && peer.Endpoint != nil {
					if peer.LastHandshakeTime.IsZero() {
						if time.Since(ifaceStart) < startupGrace {
							continue
						}
					} else if time.Since(peer.LastHandshakeTime) <= timeoutDur {
						continue
					}

					min := d.MinRandPort
					max := d.MaxRandPort
					newPort := rand.Intn(max-min+1) + min
					for {
						ok, _ := checkPort(newPort)
						if ok {
							break
						}
						newPort = rand.Intn(max-min+1) + min
					}
					deviceConfig.ListenPort = &newPort
				} else {
					if stunendpoint == nil {
						continue
					}
					addr, err := getEndpoint(stunendpoint, peer.PublicKey.String())
					if err != nil {
						continue
					}
					peerConfig := wgtypes.PeerConfig{
						PublicKey:  peer.PublicKey,
						UpdateOnly: true,
						Endpoint:   addr,
					}
					deviceConfig.Peers = append(deviceConfig.Peers, peerConfig)
				}
			}

			if err := client.ConfigureDevice(iface, deviceConfig); err != nil {
				continue
			}
		}

		client.Close()
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
						d.ifaceStart[iface] = time.Now()
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
