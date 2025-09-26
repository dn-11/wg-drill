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
)

const SocketPath = "/var/run/wg-drill-server.sock"

type daemon struct {
	ifaces       map[string]info
	pubkeytoaddr map[string]*net.UDPAddr
	lock         sync.RWMutex
}

type info []string

func newDaemon() *daemon {
	d := &daemon{}
	d.ifaces = make(map[string]info)
	for _, iface := range config.Drill.Iface {
		d.ifaces[iface] = info{}
	}
	d.pubkeytoaddr = make(map[string]*net.UDPAddr)
	return d
}

func (d *daemon) getpubkeys(iface string) ([]string, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	device, err := client.Device(iface)
	if err != nil {
		return nil, err
	}
	var pubkeys []string
	for _, peer := range device.Peers {
		pubkeys = append(pubkeys, peer.PublicKey.String())
	}
	return pubkeys, nil
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
					d.lock.Lock()
					for _, iface := range params[1:] {
						pubkeys, err := d.getpubkeys(iface)
						if err != nil {
							message += fmt.Sprintf("Failed to get pubkeys for interface %s: %v\n", iface, err)
						} else {
							d.ifaces[iface] = info(pubkeys)
							message += fmt.Sprintf("Interface %s added with %d peers\n", iface, len(pubkeys))
						}
					}
					d.lock.Unlock()
				}
			case "down":
				if len(params) != 2 {
					message += "Usage: down <interface>\n"
					return
				} else {
					d.lock.Lock()
					for _, iface := range params[1:] {
						pubkeys := d.ifaces[iface]
						for _, pubkey := range pubkeys {
							delete(d.pubkeytoaddr, pubkey)
							message += fmt.Sprintf("Removed pubkey %s from tracking\n", pubkey)
						}
						delete(d.ifaces, iface)
					}
					d.lock.Unlock()
				}
			case "show":
				d.lock.RLock()
				for iface, pubkeys := range d.ifaces {
					message += fmt.Sprintf("Interface: %s\n", iface)
					for _, pubkey := range pubkeys {
						addr := d.pubkeytoaddr[pubkey]
						message += fmt.Sprintf("  Pubkey %s Address: %s\n", pubkey, addr)
					}
				}
				d.lock.RUnlock()
			default:
				message += "Unknown command\n"
			}
			//message += "\n"
			c.Write([]byte(message))
		}(conn)
	}
}

func (d *daemon) update() {
	for {
		for iface, _ := range d.ifaces {
			d.lock.Lock()
			client, err := wgctrl.New()
			if err != nil {
				continue
			}
			device, err := client.Device(iface)
			if err != nil {
				continue
			}
			var pubkeys []string
			for _, peer := range device.Peers {
				pubkeys = append(pubkeys, peer.PublicKey.String())
			}
			d.ifaces[iface] = pubkeys
			for _, peer := range device.Peers {
				d.pubkeytoaddr[peer.PublicKey.String()] = peer.Endpoint
			}
			client.Close()
			d.lock.Unlock()
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
	fmt.Println("服务器启动，监听端口 ", config.Server.ListenPort)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Server.ListenPort), nil)
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
