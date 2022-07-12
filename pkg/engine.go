package pkg

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/spf13/viper"
	"github.com/z-y-x233/goSearch/pkg/db/badgerDb"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/protobuf/pb"
	"github.com/z-y-x233/goSearch/pkg/tools"
	"google.golang.org/protobuf/proto"
	"os"
)

const (
	//Shard is the default engine shard
	Shard = 100

	// BadgerShard badgerDB的分库数
	BadgerShard = 10
)

func ParseData() {
	var docDb []*badgerDb.BadgerDb
	for i := 0; i < BadgerShard; i++ {
		options := badger.DefaultOptions(fmt.Sprintf("%s_%d", viper.GetString("db.doc_dir")+string(os.PathSeparator)+viper.GetString("db.doc_name"), i))
		dbi := badgerDb.Open(options)
		defer dbi.Close()
		docDb = append(docDb, dbi)
	}
	n := 10
	uid := uint32(1)
	for i := 0; i < n; i++ {
		//filename := fmt.Sprintf("../wukong_release/wukong_100m_%d.csv", i)
		filename := fmt.Sprintf("D:\\wukong_release\\wukong_100m_%d.csv", i)
		data := tools.ReadCsv(filename)
		logger.Infoln("load:", filename)
		for _, line := range data {
			key := tools.U32ToBytes(uid)
			url := line[0]
			text := line[1]
			obj := &pb.DocIndex{Id: uid, Url: url, Text: text}

			val, err := proto.Marshal(obj)
			if err != nil {
				logger.Panic(err)
			}
			hs := tools.Str2Uint64(url)
			dbID := hs % BadgerShard
			err = docDb[dbID].Set(key, val)
			if err != nil {
				logger.Panic(err)
			}
			uid++
		}

	}
	logger.Debug("done ", "last uid: ", uid)

}

func Init() {

}
