package main

import (
	"github.com/spf13/viper"
	"github.com/wangbin/jiebago"
	"github.com/z-y-x233/goSearch/pkg"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"log"
)

var (
	seg           jiebago.Segmenter
	wordFilterMap map[string]bool
)

func init() {
	viper.SetConfigFile("./config.json")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	seg.LoadDictionary("./pkg/data/dict.txt")

	if err != nil {
		log.Fatal("load config failed:", err)
	}
	err = logger.Init()
	if err != nil {
		log.Fatal("init logger failed:", err)

	}

	wordFilterMap = make(map[string]bool)
	wordFilterMap["了"] = true
	wordFilterMap["的"] = true
	wordFilterMap["么"] = true
	wordFilterMap["呢"] = true
	wordFilterMap["和"] = true
	wordFilterMap["与"] = true
	wordFilterMap["于"] = true
	wordFilterMap["吗"] = true
	wordFilterMap["吧"] = true
	wordFilterMap["呀"] = true
	wordFilterMap["啊"] = true
	wordFilterMap["哎"] = true
	wordFilterMap["是"] = true
	wordFilterMap["人"] = true
	wordFilterMap["名"] = true
	wordFilterMap["在"] = true
	wordFilterMap["不"] = true
	wordFilterMap["被"] = true
	wordFilterMap["有"] = true
	wordFilterMap["无"] = true
	wordFilterMap["都"] = true
	wordFilterMap["也"] = true
	wordFilterMap["这"] = true
	wordFilterMap["是"] = true
	wordFilterMap["好"] = true
	wordFilterMap["【"] = true
	wordFilterMap["】"] = true
	wordFilterMap["《"] = true
	wordFilterMap["》"] = true
	wordFilterMap["，"] = true
	wordFilterMap["。"] = true
	wordFilterMap["？"] = true
	wordFilterMap["！"] = true
	wordFilterMap["、"] = true
	wordFilterMap["；"] = true
	wordFilterMap["："] = true
	wordFilterMap["（"] = true
	wordFilterMap["）"] = true
	wordFilterMap["什么"] = true
	wordFilterMap["\""] = true
	wordFilterMap["”"] = true
	wordFilterMap["‘"] = true
	wordFilterMap["“"] = true
	wordFilterMap["’"] = true
	wordFilterMap[","] = true

}

func print(ch <-chan string) (ss []string) {

	for w := range ch {
		ss = append(ss, w)
	}
	return ss
}

func main() {
	logger.Info("start")
	pkg.ParseData()
	logger.Info("done")

}
