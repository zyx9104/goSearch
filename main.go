package main

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
	"github.com/z-y-x233/goSearch/pkg/engine"
	"github.com/z-y-x233/goSearch/pkg/logger"
)

var (
	wordFilterMap map[string]bool
)

func init() {
	viper.SetConfigFile("./config.json")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()

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
	logger.Infoln("========================== Process Start ==========================")

	// s, _ := strconv.Atoi(os.Args[1])
	// e, _ := strconv.Atoi(os.Args[2])
	// // s, e := 1, viper.GetInt("db.last_index")
	// id, _ := strconv.Atoi(os.Args[3])
	e := engine.DefaultEngine()
	e.Wait()
	q := os.Args[1]

	logger.Info("Start Search")
	t := time.Now()
	docs := e.Search(q)
	st := time.Since(t)
	for i, doc := range docs {
		if i >= 10 {
			break
		}
		logger.Infoln(doc.Id, doc.Text, (len(doc.Text)+len(doc.Url)+4)/4)
		// fmt.Println(doc.Id, doc.Text)

	}
	logger.Info("Search Done time:", st)
	// gen.GenWordIds()
	logger.Infoln("========================== Process Done ==========================")
}
