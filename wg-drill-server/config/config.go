package config

import (
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

//var Server struct {
//	ListenAddr string
//	ListenPort int
//}

var Drill struct {
	Enable   bool
	Iface    string
	Interval int
}

var WireGuard struct {
	PrivateKey  string
	ListenPort  int
	PeerPubkeys []string
}

const file = "/etc/wg-drill-server/config.toml"

func Init() {
	if _, err := os.Stat(file); err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("Error checking config file: %v", err)
		}
	}
	viper.SetConfigFile(file)
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	viper.SetDefault("drill.enable", true)
	viper.SetDefault("drill.interval", 10)
	viper.SetDefault("wireguard.privatekey", WireGuard.PrivateKey)
	viper.SetDefault("wireguard.listenport", WireGuard.ListenPort)

	update()

	viper.OnConfigChange(func(e fsnotify.Event) {
		update()
	})

	viper.WatchConfig()

}

func update() {

	Drill.Enable = viper.GetBool("drill.enable")
	Drill.Iface = viper.GetString("drill.iface")
	Drill.Interval = viper.GetInt("drill.interval")
}
