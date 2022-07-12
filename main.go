package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
	"github.com/wangbin/jiebago"
	"github.com/z-y-x233/goSearch/pkg/db/badgerDb"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/protobuf/pb/model"
	"github.com/z-y-x233/goSearch/pkg/tools"
	"google.golang.org/protobuf/proto"
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
	// logger.Debug("start")
	// pkg.ParseData()
	options := badger.DefaultOptions(fmt.Sprintf("%s_%d", viper.GetString("db.doc_dir")+string(os.PathSeparator)+viper.GetString("db.doc_name"), 0))
	dbi := badgerDb.Open(options)
	defer dbi.Close()

	docc := &model.DocIndex{Id: 1, Url: "sss.pp", Text: "test db proto"}
	logger.Debug(docc)

	buf, err := proto.Marshal(docc)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Debug(buf)

	dbi.Set(tools.U32ToBytes(docc.Id), buf)
	obj, found := dbi.Get(tools.U32ToBytes(docc.Id))

	if !found {
		logger.Debug("not found key 0")
	}
	doc := &model.DocIndex{}
	err = proto.Unmarshal(obj, doc)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Debug(doc)
}
