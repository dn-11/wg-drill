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
			var stunendpoint *net.UDPAddr
			for _, peer := range device.Peers {
				if peer.AllowedIPs == nil || len(peer.AllowedIPs) == 0 {
					stunendpoint = peer.Endpoint
				}
			}
			fmt.Println("Interface:", iface, "Using STUN endpoint:", stunendpoint.String())
			deviceConfig := wgtypes.Config{
				PrivateKey:   &device.PrivateKey,
				ReplacePeers: false,
				Peers:        []wgtypes.PeerConfig{},
			}

			if device.FirewallMark > 0 {
				deviceConfig.FirewallMark = &device.FirewallMark
			}

			for _, peer := range device.Peers { // 遍历所有peer
				if peer.AllowedIPs == nil || len(peer.AllowedIPs) == 0 { //检测是否卡nat,如果lasthandshake超过timeout重新设置listenport
					if time.Since(peer.LastHandshakeTime) > time.Duration(config.Drill.Timeout)*time.Second {
						min := d.MinRandPort
						max := d.MaxRandPort
						//if min <= 0 || max <= 0 || max < min {
						//	min = 40000
						//	max = 65535
						//}
						newPort := rand.Intn(max-min+1) + min
						for {
							ok, _ := checkPort(newPort)
							if ok {
								break
							} else {
								newPort = rand.Intn(max-min+1) + min
							}
						}
						deviceConfig.ListenPort = &newPort
					}
				} else { //更新endpoint
					addr, err := getEndpoint(stunendpoint, peer.PublicKey.String())
					//fmt.Println(addr, err)
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

					deviceConfig.Peers = append(deviceConfig.Peers, peerConfig)
				}

			}
			err = client.ConfigureDevice(iface, deviceConfig)
			if err != nil {
				//fmt.Printf("Failed to configure device %s for %s: %s\n", iface, iface, err)
				continue
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
