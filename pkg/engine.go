package pkg

import (
	"fmt"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/wangbin/jiebago"
	"google.golang.org/protobuf/proto"

	"github.com/spf13/viper"
	"github.com/z-y-x233/goSearch/pkg/db/badgerDb"
	"github.com/z-y-x233/goSearch/pkg/db/boltDb"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/protobuf/pb"
	"github.com/z-y-x233/goSearch/pkg/tools"
)

var (
	docDB     []*badgerDb.BadgerDb
	invDB     [][]*badgerDb.BadgerDb
	boltDocDB *boltDb.BoltDb
	buckets   [][]byte
)

const (

	// BadgerShard badgerDB的分库数
	BadgerShard = 10

	// BoltBucketSize boltDB中桶的数量
	BoltBucketSize = 100
)

func Init() {
	logger.Infoln("========================== open doc database ==========================")
	docDB = make([]*badgerDb.BadgerDb, 0)
	for i := 0; i < BadgerShard; i++ {
		options := badger.DefaultOptions(fmt.Sprintf("%s_%d", path.Join(viper.GetString("db.doc.dir"), viper.GetString("db.doc.filename")), i))
		dbi := badgerDb.Open(options)
		// defer dbi.Close()
		docDB = append(docDB, dbi)
	}
	logger.Infoln("========================== open doc done ==========================")

	logger.Infoln("========================== open bolt doc database ==========================")

	database := path.Join(viper.GetString("db.doc.dir"), viper.GetString("db.doc.bolt"))

	boltDocDB, _ = boltDb.Open(database)

	buckets = make([][]byte, 0)

	for i := 0; i < BoltBucketSize; i++ {
		bucketName := tools.U32ToBytes(uint32(i))
		buckets = append(buckets, bucketName)
		err := boltDocDB.CreateBucketIfNotExist(bucketName)
		if err != nil {
			return
		}
	}
	logger.Infoln("========================== open bolt doc done ==========================")

	// invDBnum := 4
	// invDB = make([][]*badgerDb.BadgerDb, invDBnum)
	// logger.Infoln("========================== open invIdx database ==========================")
	// for i := 0; i < invDBnum; i++ {
	// 	invDB[i] = make([]*badgerDb.BadgerDb, 0)
	// 	filename := fmt.Sprintf("db.invIndex.filename%d", i+1)
	// 	dataname := path.Join(viper.GetString("db.invIndex.dir"), viper.GetString(filename))
	// 	logger.Infoln("open database:", dataname)
	// 	for j := 0; j < BadgerShard; j++ {
	// 		options := badger.DefaultOptions(fmt.Sprintf("%s_%d", dataname, j))
	// 		dbi := badgerDb.Open(options)
	// 		defer dbi.Close()
	// 		invDB[i] = append(invDB[i], dbi)
	// 		invDB[i][j].Get([]byte{1})
	// 	}
	// 	logger.Infoln(dataname, "open done")

	// }
	// logger.Infoln("========================== open invIdx done ==========================")
}

type obj struct {
	id  uint32
	key []byte
	val []byte
}

func EncodeData() {
	t := time.Now()
	writeBuf := make([]obj, 0)

	for i := 24000000; i < viper.GetInt("db.last_index"); i++ {
		id := uint32(i)
		key := tools.U32ToBytes(id)
		data, _ := docDB[id%BadgerShard].Get(key)
		writeBuf = append(writeBuf, obj{id: id, key: key, val: data})
		// val, _ := tools.Encode(data)
		// bytes += len(data) - len(val)
		// enDocDB[id%BadgerShard].Set(key, val)

		if i%100000 == 0 {
			var wg sync.WaitGroup
			n := len(writeBuf)

			wg.Add(n)
			for _, item := range writeBuf {
				go func(item obj) {
					boltDocDB.MulSet(item.key, item.val, buckets[item.id%BoltBucketSize])
					wg.Done()
				}(item)
			}
			wg.Wait()
			writeBuf = writeBuf[:0]
			logger.Infoln("index:", i, "encode", n, "data", "time:", time.Since(t))
			t = time.Now()
		}
	}
	var wg sync.WaitGroup
	n := len(writeBuf)
	wg.Add(n)
	for _, item := range writeBuf {
		go func(item obj) {
			boltDocDB.MulSet(item.key, item.val, buckets[item.id%BoltBucketSize])
			wg.Done()
		}(item)
	}
	wg.Wait()
	logger.Infoln("encode", n, "data", "time:", time.Since(t))
}

func ParseData(start, end uint32) {

	n := 256
	uid := uint32(1)

	for i := 0; i < n; i++ {
		filename := fmt.Sprintf("../wukong_release/wukong_100m_%d.csv", i)
		// filename := fmt.Sprintf("D:\\wukong_release\\wukong_100m_%d.csv", i)
		data := tools.ReadCsv(filename)
		for _, line := range data {
			key := tools.U32ToBytes(uid)
			url := line[0]
			text := line[1]
			obj := &pb.DocIndex{Id: uid, Url: url, Text: text}

			val, err := proto.Marshal(obj)
			if err != nil {
				logger.Panic(err)
			}
			dbID := uid % BadgerShard
			val, err = tools.Encode(val)
			tools.HandleError(err)
			err = docDB[dbID].Set(key, val)
			tools.HandleError(err)
			uid++
			if uid == end {
				break
			}
		}
		if uid == end {
			break
		}
		runtime.GC()
	}
	logger.Info("done ", "last uid: ", uid)

}

func BuildInvIdx(start, end int) {

	logger.Debug("================ start build ===============")

	// maxIdx := viper.GetInt("db.last_index")

	var seg jiebago.Segmenter

	seg.LoadDictionary("./pkg/data/dict.txt")

	// getWords := func(ch <-chan string) (res []string) {
	// 	for s := range ch {
	// 		res = append(res, s)
	// 	}
	// 	return
	// }
	// maxIdx = 300000
	invIdx := make(map[string]*pb.InvIndex, 100000)
	t1 := time.Now()
	// t := time.Now()
	unmarshalTime := time.Second * 0
	wordCutTime := time.Second * 0
	statWordTime := time.Second * 0
	totalLength := 0

	// for i := start; i < end; i++ {
	// 	ut := time.Now()
	// 	doc := &pb.DocIndex{}
	// 	k := tools.U32ToBytes(uint32(i))
	// 	data, found := docDB[i%BadgerShard].Get(k)
	// 	if !found {
	// 		logger.Debugln("id:", i, "not found")
	// 		continue
	// 	}
	// 	err := proto.Unmarshal(data, doc)
	// 	logger.Debug(doc)
	// 	if err != nil {
	// 		logger.Panic(err)
	// 	}
	// 	unmarshalTime += time.Since(ut)
	// 	wct := time.Now()
	// 	ch := seg.CutForSearch(doc.Text, true)
	// 	words := getWords(ch)
	// 	wordCutTime += time.Since(wct)
	// 	totalLength += len([]rune(doc.Text))
	// 	swt := time.Now()
	// 	wordCount := make(map[string]int, 1000)
	// 	for _, word := range words {
	// 		wordCount[word]++
	// 	}

	// 	for word, cnt := range wordCount {
	// 		key := tools.Str2Uint64(word)
	// 		if _, ok := invIdx[word]; !ok {
	// 			invIdx[word] = &pb.InvIndex{Ids: make([]*pb.Item, 0), Key: key}
	// 		}

	// 		invIdx[word].Ids = append(invIdx[word].Ids, &pb.Item{Id: uint32(i), Count: uint32(cnt)})
	// 	}
	// 	statWordTime += time.Since(swt)
	// 	if i%100000 == 0 {
	// 		logger.Infoln("load 10w data", "words:", len(invIdx), "time:", time.Since(t))
	// 		t = time.Now()
	// 	}
	// }
	// var buckets [][]byte = make([][]byte, BoltBucketSize)
	// invIdxDb, err := boltDb.Open(viper.GetString("db.inv_index_dir") + string(os.PathSeparator) + viper.GetString("db.inv_index_name"))
	// if err != nil {
	// 	logger.Panic(err)
	// }
	// for i := 0; i < BoltBucketSize; i++ {
	// 	bucketName := tools.U32ToBytes(uint32(i))
	// 	buckets[i] = bucketName
	// 	err := invIdxDb.CreateBucketIfNotExist(bucketName)
	// 	if err != nil {
	// 		logger.Panic(err)
	// 	}
	// }
	var invDB []*badgerDb.BadgerDb
	for i := 0; i < BadgerShard; i++ {
		docDB[i].Close()
		options := badger.DefaultOptions(fmt.Sprintf("%s_%d", path.Join(viper.GetString("db.invIndex.dir"), viper.GetString("db.invIndex.filename1")), i))
		dbi := badgerDb.Open(options)
		defer dbi.Close()
		invDB = append(invDB, dbi)
	}

	logger.Infoln("totalLength:", totalLength, "avgLength", totalLength/2000000, "total words:", len(invIdx), "parse data time:", time.Since(t1))
	logger.Infoln("unmarshalTime:", unmarshalTime, "wordCutTime:", wordCutTime, "statWordTime:", statWordTime)
	t1 = time.Now()
	cnt := 0
	existed := 0
	mergeTime := time.Second * 0
	for w, val := range invIdx {
		key := tools.U64ToBytes(val.Key)
		dbID := val.Key % BadgerShard
		// buf, found := invIdxDb.Get(key, buckets[val.Key%BoltBucketSize])
		buf, found := invDB[dbID].Get(key)
		if found {
			existed++
			invidx := &pb.InvIndex{}
			proto.Unmarshal(buf, invidx)
			mt := time.Now()
			merge(val, invidx)
			mergeTime += time.Since(mt)
		}
		data, err := proto.Marshal(val)
		if err != nil {
			logger.Panic(err)
		}
		// invIdxDb.Set(key, data, buckets[val.Key%BoltBucketSize])
		invDB[dbID].Set(key, data)
		logger.Debugln(w, val)

		delete(invIdx, w)
		cnt++
		if cnt == 50000 {
			runtime.GC()
			cnt = 0
		}
	}
	logger.Infoln("existed words:", existed, "merge time:", mergeTime, "write time:", time.Since(t1))
}

func merge(a, b *pb.InvIndex) {
	// mp := make(map[uint32]uint32, 0)
	// for _, v := range a.Ids {
	// 	mp[v.Id] += v.Count
	// }
	// for _, v := range b.Ids {
	// 	mp[v.Id] += v.Count
	// }
	// a.Ids = a.Ids[:0]
	// for k, v := range mp {
	// 	a.Ids = append(a.Ids, &pb.Item{Id: k, Count: v})
	// }
}

func MergeIndex(start, end int) {

	// getDoc := func(docID uint32) *pb.DocIndex {
	// 	filename := "db.doc.filename"
	// 	dataname := path.Join(viper.GetString("db.doc.dir"), viper.GetString(filename))
	// 	options := badger.DefaultOptions(fmt.Sprintf("%s_%d", dataname, docID%BadgerShard))
	// 	db := badgerDb.Open(options)
	// 	defer db.Close()
	// 	data, _ := db.Get(tools.U32ToBytes(docID))
	// 	doc := &pb.DocIndex{}
	// 	proto.Unmarshal(data, doc)
	// 	return doc
	// }

	// getDocs := func(inv *pb.InvIndex) {
	// 	for _, item := range inv.Ids {
	// 		doc := getDoc(item.Id)
	// 		logger.Info(doc)
	// 	}
	// }

	logger.Infoln("========================== open dst database ==========================")
	dstDB := make([]*badgerDb.BadgerDb, 0)
	filename := "db.invIndex.dstfile"
	dataname := path.Join(viper.GetString("db.invIndex.dir"), viper.GetString(filename))
	logger.Infoln("open database:", dataname)
	for j := 0; j < BadgerShard; j++ {
		options := badger.DefaultOptions(fmt.Sprintf("%s_%d", dataname, j))
		dbi := badgerDb.Open(options)
		defer dbi.Close()
		dstDB = append(dstDB, dbi)
	}
	logger.Infoln("========================== open done ==========================")

	logger.Infoln("========================== merge database start ==========================")

	for i := start; i < end; i++ {
		mergeTime := time.Second * 0
		iterTime := time.Second * 0
		unmarshalTime := time.Second * 0
		marshalTime := time.Second * 0
		for j := 0; j < BadgerShard; j++ {
			logger.Infof("merge database %d:%d", i, j)
			tt := time.Now()
			it, err := invDB[i][j].GetKeys()
			logger.Info("key num: ", len(it))
			if err != nil {
				logger.Panic(err)
			}
			itt := time.Now()

			for _, k := range it {
				v, found := invDB[i][j].Get(k)
				if !found {
					continue
				}
				kk := make([]byte, len(k))
				copy(kk, k)

				invItem1 := &pb.InvIndex{}
				invItem2 := &pb.InvIndex{}
				ut := time.Now()
				proto.Unmarshal(v, invItem1)
				// getDocs(invItem1)
				logger.Info(k)
				unmarshalTime += time.Since(ut)
				buf, fd := dstDB[j].Get(kk)
				if fd {
					ut := time.Now()

					proto.Unmarshal(buf, invItem2)
					unmarshalTime += time.Since(ut)
					// getDocs(invItem2)
					mt := time.Now()
					merge(invItem1, invItem2)
					mergeTime += time.Since(mt)
				}
				mt := time.Now()
				data, err := proto.Marshal(invItem1)
				if err != nil {
					logger.Panic(err)
				}
				marshalTime += time.Since(mt)
				copy(kk, k)

				err = dstDB[j].Set(kk, data)
				// getDocs(invItem1)
				data, _ = dstDB[j].Get(kk)
				proto.Unmarshal(data, invItem1)
				// getDocs(invItem1)
				if err != nil {
					logger.Info(err)
				}
			}
			keys, _ := dstDB[j].GetKeys()

			logger.Infoln("after merge dstDB", j, "key num:", len(keys))
			iterTime += time.Since(itt)
			logger.Infoln("database", i, ":", j, "merge time:", time.Since(tt))
		}
		logger.Infoln("merge time:", mergeTime, "iter time:", iterTime, "unmarshal time", unmarshalTime, "marshal time:", marshalTime)
	}
	logger.Infoln("========================== merge database done ==========================")

}

func BoltTest() {
	t := time.Now()
	var wg sync.WaitGroup

	wg.Add(100000)
	for i := 1; i < 100001; i++ {
		id := uint32(i)
		key := tools.U32ToBytes(id)
		go func() {
			boltDocDB.Get(key, buckets[i%BadgerShard])
			wg.Done()
		}()
		// val, _ := tools.Encode(data)
		// bytes += len(data) - len(val)
		// enDocDB[id%BadgerShard].Set(key, val)

	}
	wg.Wait()
	// var wg sync.WaitGroup
	n := 100000
	logger.Infoln("encode", n, "data", "time:", time.Since(t))
	t = time.Now()

	wg.Done()
	// var wg sync.WaitGroup
	// n := len(writeBuf)
	// wg.Add(n)
	// for _, item := range writeBuf {
	// 	go func(item obj) {
	// 		boltDocDB.MulSet(item.key, item.val, buckets[item.id%BoltBucketSize])
	// 		wg.Done()
	// 	}(item)
	// }
	// wg.Wait()
	logger.Infoln("encode", "1e", "data", "time:", time.Since(t))
}
