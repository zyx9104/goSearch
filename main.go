package main

import (
	"log"

	"github.com/spf13/viper"
	"github.com/wangbin/jiebago"
	"github.com/z-y-x233/goSearch/pkg/logger"
)

var seg jiebago.Segmenter

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

}

func print(ch <-chan string) (ss []string) {

	for w := range ch {
		ss = append(ss, w)
	}
	return ss
}

func main() {
	s := print(seg.CutAll("小明硕士毕业于中国科学院计算所，后在日本京都大学深造"))
	logger.Debug(s)
}
