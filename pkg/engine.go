package pkg

const (
	//Shard is the default engine shard
	Shard = 100

	// BadgerShard badgerDB的分库数
	BadgerShard = 100
)

// func ParseData() {
// 	var docDb []*badgerDb.BadgerDb
// 	for i := 0; i < 1; i++ {
// 		options := badger.DefaultOptions(fmt.Sprintf("%s_%d", viper.GetString("db.doc_dir")+string(os.PathSeparator)+viper.GetString("db.doc_name"), i))
// 		dbi := badgerDb.Open(options)
// 		defer dbi.Close()
// 		docDb = append(docDb, dbi)
// 	}
// 	n := 1
// 	for i := 0; i < n; i++ {
// 		filename := fmt.Sprintf("../wukong_release/wukong_100m_%d.csv", i)
// 		data := tools.ReadCsv(filename)
// 		line := data[0]
// 		// for i, line := range data {
// 		key := tools.U32ToBytes(uint32(i))
// 		url := line[0]
// 		text := line[1]
// 		obj := &pb.DocIndex{Id: uint32(i), Url: url, Text: text}
// 		val, err := proto.Marshal(obj)
// 		if err != nil {
// 			fmt.Println("====================================", err)
// 		}
// 		docDb[0].Set(key, val)
// 		// }

// 	}
// 	logger.Debug("done")

// }

func Init() {

}
