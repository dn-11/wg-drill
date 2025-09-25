package config

import (
	"log"

	"github.com/spf13/viper"
)

var config *viper.Viper

func Init() {
	config = viper.New()
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.AddConfigPath("/etc/natdrill")

	if err := config.ReadInConfig(); err != nil {
		log.Fatalf("fail to read config file: %v", err)
	}

}

func GetViper() *viper.Viper {
	return config
}
