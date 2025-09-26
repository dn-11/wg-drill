package config

import (
	"errors"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var Server struct {
	Endpoint string
}

var Drill struct {
	Enable   bool
	Iface    []string
	Interval int
}

const file = "/etc/wg-drill-client/config.toml"

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

	if viper.IsSet("server.endpoint") == false {
		panic(errors.New("server.endpoint is not set in config file"))
	}

	update()

	viper.OnConfigChange(func(e fsnotify.Event) {
		update()
	})

	viper.WatchConfig()

}

func update() {
	Server.Endpoint = viper.GetString("server.endpoint")
	Drill.Enable = viper.GetBool("drill.enable")
	Drill.Iface = viper.GetStringSlice("drill.iface")
	Drill.Interval = viper.GetInt("drill.interval")
}
