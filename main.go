package main

import (
	"log"

	"github.com/spf13/viper"
	"github.com/z-y-x233/goSearch/pkg/logger"
)

func init() {
	viper.SetConfigFile("./config.json")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln("load config failed:", err)
	}
	err = logger.Init()
	if err != nil {
		log.Fatalln("init logger failed:", err)

	}
}

func main() {

}
