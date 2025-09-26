package config

import (
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var Server struct {
	ListenAddr string
	ListenPort int
}

var Drill struct {
	Enable   bool
	Iface    []string
	Interval int
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

	viper.SetDefault("server.listenaddr", "0.0.0.0")
	viper.SetDefault("server.listenaddr", 14514)
	viper.SetDefault("drill.enable", true)
	viper.SetDefault("drill.interval", 10)

	update()

	viper.OnConfigChange(func(e fsnotify.Event) {
		update()
	})

	viper.WatchConfig()

}

func update() {
	Server.ListenAddr = viper.GetString("server.listenaddr")
	Server.ListenPort = viper.GetInt("server.listenport")
	Drill.Enable = viper.GetBool("drill.enable")
	Drill.Iface = viper.GetStringSlice("drill.iface")
	Drill.Interval = viper.GetInt("drill.interval")
}
