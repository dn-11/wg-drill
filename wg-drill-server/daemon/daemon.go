// Package daemon 启动一个守护进程，监听来自客户端的请求
// 启动

package daemon

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"wg-drill-server/config"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const SocketPath = "/var/run/wg-drill-server.sock"

type daemon struct {
	iface        string
	pubkeytoaddr map[string]*net.UDPAddr
	lock         sync.RWMutex
	port         int
}

func (d *daemon) getIface() (*wgtypes.Device, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	device, err := client.Device(d.iface)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func newDaemon() *daemon {
	d := &daemon{}
	d.iface = config.Drill.Iface
	device, err := d.getIface()
	if err != nil {
		fmt.Printf("Error getting device: %s\n", err)
		panic(err)
	}
	d.port = device.ListenPort
	d.pubkeytoaddr = make(map[string]*net.UDPAddr)
	err = d.initPeer()
	if err != nil {
		panic(err)
	}
	d.port = device.ListenPort
	return d
}

func (d *daemon) initPeer() error {
	device, err := d.getIface()
	if err != nil {
		return err
	}
	for _, peer := range device.Peers {
		d.pubkeytoaddr[peer.PublicKey.String()] = nil
	}
	return nil
}

func (d *daemon) addPeer(pubkey string) error {

	client, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer client.Close()

	d.lock.Lock()
	defer d.lock.Unlock()

	key, err := wgtypes.ParseKey(pubkey)
	if err != nil {
		return err
	}
	peerConfig := wgtypes.PeerConfig{
		PublicKey: key,
	}

	deviceConfig := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerConfig},
	}
	return client.ConfigureDevice(d.iface, deviceConfig)

}

func (d *daemon) removePeer(pubkey string) error {
	client, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer client.Close()

	d.lock.Lock()
	defer d.lock.Unlock()

	key, err := wgtypes.ParseKey(pubkey)
	if err != nil {
		return err
	}

	peerConfig := wgtypes.PeerConfig{
		PublicKey: key,
		Remove:    true,
	}

	deviceConfig := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerConfig},
	}

	delete(d.pubkeytoaddr, pubkey)

	return client.ConfigureDevice(d.iface, deviceConfig)
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
			case "add":
				if len(params) != 2 {
					message += "Usage: add <pubkey>\n"
					return
				} else {
					err := d.addPeer(params[1])
					if err != nil {
						message += "Error: " + err.Error() + "\n"
					} else {
						message += "Added peer " + params[1] + "\n"
					}
				}
			case "del":
				if len(params) != 2 {
					message += "Usage: del <pubkey>\n"
					return
				} else {
					err := d.removePeer(params[1])
					if err != nil {
						message += "Error: " + err.Error() + "\n"
					} else {
						message += "Removed peer " + params[1] + "\n"
					}
				}
			case "show":
				d.lock.RLock()
				message += fmt.Sprintf("%-44s\t%s\n", "PublicKey", "Endpoint")
				for key, addr := range d.pubkeytoaddr {
					endpoint := "<none>"
					if addr != nil {
						endpoint = addr.String()
					}
					message += fmt.Sprintf("%-44s\t%s\n", key, endpoint)
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

func (d *daemon) update() { //update peer endpoint periodically
	for {
		device, err := d.getIface()
		if err != nil {
			time.Sleep(time.Duration(config.Drill.Interval) * time.Second)
			continue
		}
		for {
			d.lock.RLock()
			for _, peer := range device.Peers {
				d.pubkeytoaddr[peer.PublicKey.String()] = peer.Endpoint

			}
			d.lock.RUnlock()
			break
		}
		time.Sleep(time.Duration(config.Drill.Interval) * time.Second)
	}
}

func (d *daemon) handler(w http.ResponseWriter, r *http.Request) {
	pubkey := r.URL.Query().Get("pubkey")

	addr := d.pubkeytoaddr[pubkey]
	fmt.Println(pubkey)
	if addr == nil || (addr.IP == nil && addr.Port == 0) {
		http.Error(w, "Not Found", http.StatusNotFound)
	} else {
		fmt.Fprintf(w, addr.String())
	}
	return
}

func (d *daemon) server() {

	http.HandleFunc("/", d.handler)
	fmt.Println("start server at", d.port)
	http.ListenAndServe(fmt.Sprintf(":%d", d.port), nil)
}

func Run() {
	config.Init()
	d := newDaemon()
	go d.commu()
	go d.update()
	go d.server()
	fmt.Println("Running wg-drill-server daemon...")
	select {}
}
