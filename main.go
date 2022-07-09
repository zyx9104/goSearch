package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"log"
)

func init() {
	viper.SetConfigFile("./config.json")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln("load config failed:", err)
	}
}

func main() {
	logger.Init()
	logger.Log.Infoln("done")
	logger.Log.WithFields(logrus.Fields{
		"prefix": "main",
		"animal": "walrus",
		"number": 8,
	}).Debug("Started observing beach")

	logger.Log.WithFields(logrus.Fields{
		"prefix":      "sensor",
		"temperature": -4,
	}).Info("Temperature changes")
}
