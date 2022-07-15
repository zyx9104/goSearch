package gen

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/z-y-x233/goSearch/pkg/db"
	"github.com/z-y-x233/goSearch/pkg/engine"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/protobuf/pb"
	"github.com/z-y-x233/goSearch/pkg/tools"
	"google.golang.org/protobuf/proto"
)

type obj struct {
	id  uint32
	key []byte
	val []byte
}

var (
	com sync.WaitGroup
)

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
			dbID := uid
			val, err = tools.Encode(val)
			tools.HandleError(err)
			err = engine.DocDB[dbID].Set(key, val)
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

func getWords(ch <-chan string) (res []string) {
	for s := range ch {
		res = append(res, s)
	}
	return
}

type mp struct {
	Id  uint64
	Uid uint32
}

func parseDoc(doc *pb.DocIndex, ch chan *mp) {
	words := getWords(engine.Seg.CutForSearch(doc.Text, true))
	for _, word := range words {
		ch <- &mp{Id: tools.Str2Uint64(word), Uid: doc.Id}
	}
}

func consume(ht map[uint64][]uint32, itemCh <-chan *mp) {
	com.Add(1)
	for item := range itemCh {
		ht[item.Id] = append(ht[item.Id], item.Uid)
	}
	com.Done()
}

func prodDoc(docCh chan *pb.DocIndex, itemCh chan *mp) {
	var wg sync.WaitGroup
	for doc := range docCh {
		wg.Add(1)
		go func(doc *pb.DocIndex) {
			parseDoc(doc, itemCh)
			wg.Done()
		}(doc)
	}
	wg.Wait()
	close(itemCh)
}

func BuildInvIdx(start, end int) {

	logger.Infoln("===========================start build===================================")
	logger.Infoln("start:", start, "end:", end)
	itemCh := make(chan *mp, 2000)
	ht := make(map[uint64][]uint32, 10000)
	readBuf := make([]uint32, 0)
	docCh := make(chan *pb.DocIndex, 10000)

	go prodDoc(docCh, itemCh)
	go consume(ht, itemCh)

	t := time.Now()
	rdbufSize := 2500
	readData := 0
	totalTime := time.Second * 0
	for i := start; i < end; i++ {
		id := uint32(i)
		readBuf = append(readBuf, id)
		if len(readBuf) == rdbufSize {
			var wg sync.WaitGroup
			n := len(readBuf)
			wg.Add(n)

			for _, item := range readBuf {
				go func(item uint32) {
					key := tools.U32ToBytes(item)
					data, found := engine.DocDB.Get(key, engine.Buckets[item%engine.BoltBucketSize])
					if found {
						doc := &pb.DocIndex{}
						proto.Unmarshal(data, doc)
						docCh <- doc
						// logger.Infoln(doc)
					}
					wg.Done()
				}(item)
			}
			wg.Wait()
			readData += rdbufSize
			totalTime += time.Since(t)
			logger.Infof("read %v item, total data: %v read time: %v, per time: %v, total time: %v", rdbufSize, readData, time.Since(t), time.Since(t)/time.Duration(rdbufSize), totalTime)
			t = time.Now()
			readBuf = readBuf[:0]
		}
	}

	if len(readBuf) > 0 {
		var wg sync.WaitGroup
		n := len(readBuf)
		wg.Add(n)
		for _, item := range readBuf {
			go func(item uint32) {
				key := tools.U32ToBytes(item)
				data, found := engine.DocDB.Get(key, engine.Buckets[item%engine.BoltBucketSize])
				if found {
					doc := &pb.DocIndex{}
					proto.Unmarshal(data, doc)
					docCh <- doc
					// logger.Infoln(doc)
				}
				wg.Done()
			}(item)
		}
		wg.Wait()
		readData += len(readBuf)
		totalTime += time.Since(t)
		logger.Infof("read %v item, total data: %v read time: %v, per time: %v, total time: %v", rdbufSize, readData, time.Since(t), time.Since(t)/time.Duration(rdbufSize), totalTime)
	}

	close(docCh)
	com.Wait()
	type invS struct {
		Id  uint64
		Key []byte
		Val []byte
	}
	writeBuf := make([]*invS, 0)
	writeBufSize := 2500
	writeData := 0
	writeTime := time.Second * 0
	wt := time.Now()
	for key, val := range ht {
		invItem := &pb.InvIndex{Key: key, Ids: val}
		data, err := proto.Marshal(invItem)
		if err != nil {
			logger.Panic(err)
		}
		writeBuf = append(writeBuf, &invS{Id: key, Key: tools.U64ToBytes(key), Val: data})
		if len(writeBuf) >= writeBufSize {
			var wg sync.WaitGroup
			wg.Add(len(writeBuf))
			for _, item := range writeBuf {
				go func(item *invS) {
					// _, found := engine.BoltInvDB.Get(item.Key, engine.InvBuckets[item.Id%engine.BoltBucketSize])
					// if !found {
					engine.InvDB.MulSet(item.Key, item.Val, engine.InvBuckets[item.Id%engine.BoltBucketSize])
					// }
					wg.Done()
				}(item)
			}
			wg.Wait()
			writeData += len(writeBuf)
			writeTime += time.Since(wt)
			logger.Infof("write %v item, total data: %v,  write time: %v, per time: %v, total time: %v", len(writeBuf), writeData, time.Since(wt), time.Since(wt)/time.Duration(writeBufSize), writeTime)
			wt = time.Now()
			writeBuf = writeBuf[:0]
		}
	}
	if len(writeBuf) > 0 {
		var wg sync.WaitGroup
		wg.Add(len(writeBuf))
		for _, item := range writeBuf {
			go func(item *invS) {
				// _, found := engine.BoltInvDB.Get(item.Key, engine.InvBuckets[item.Id%engine.BoltBucketSize])
				// if !found {
				engine.InvDB.MulSet(item.Key, item.Val, engine.InvBuckets[item.Id%engine.BoltBucketSize])
				// }
				wg.Done()
			}(item)
		}
		wg.Wait()
		writeData += len(writeBuf)
		writeTime += time.Since(wt)
		logger.Infof("write %v item, total data: %v,  write time: %v, per time: %v, total time: %v", len(writeBuf), writeData, time.Since(wt), time.Since(wt)/time.Duration(writeBufSize), writeTime)
	}
	logger.Infoln("===========================build inv index done===================================")
}

type wtObj struct {
	Key []byte
	Val []byte
	Bkt []byte
}

func mulWrite(db *db.BoltDb, writeBuf []*wtObj) {
	var wg sync.WaitGroup
	wg.Add(len(writeBuf))
	for _, item := range writeBuf {
		go func(item *wtObj) {
			db.MulSet(item.Key, item.Val, item.Bkt)
			wg.Done()
		}(item)
	}
	wg.Wait()
}

func MergeIndex(fileID int) {
	var boltInvdb *db.BoltDb
	dir := viper.GetString("db.invIndex.dir")
	filename := viper.GetString(fmt.Sprintf("db.invIndex.bolt%d", fileID))
	database := path.Join(dir, filename)
	boltInvdb, _ = db.Open(database, true)

	logger.Infoln("============================================ open database:", database, "==============================================")
	for i := 0; i < engine.BoltBucketSize; i++ {
		bucketName := tools.U32ToBytes(uint32(i))
		err := boltInvdb.CreateBucketIfNotExist(bucketName)
		if err != nil {
			logger.Panic(err)
		}
	}
	logger.Infoln("============================================ open done ==============================================")

	writeBuf := make([]*wtObj, 0)
	writeBufSize := 2500
	totalTime := time.Second * 0
	totalData := 0
	t := time.Now()

	for i := 0; i < engine.BoltBucketSize; i++ {
		logger.Infoln("merge bucket:", i)
		bucketName := tools.U32ToBytes(uint32(i))
		vals, err := boltInvdb.GetVals(bucketName)
		if err != nil {
			logger.Panic(err)
		}
		for _, data := range vals {
			invItem := &pb.InvIndex{}
			proto.Unmarshal(data, invItem)
			key := tools.U64ToBytes(invItem.Key)
			buf, found := engine.InvDB.Get(key, bucketName)
			if found {
				extItem := &pb.InvIndex{}
				proto.Unmarshal(buf, extItem)
				invItem.Ids = append(invItem.Ids, extItem.Ids...)
				// engine.BoltInvDB.Delete(key, bucketName)
			}
			buf, _ = proto.Marshal(invItem)
			writeBuf = append(writeBuf, &wtObj{Key: key, Val: buf, Bkt: bucketName})
			if len(writeBuf) >= writeBufSize {
				mulWrite(engine.InvDB, writeBuf)
				totalData += len(writeBuf)
				totalTime += time.Since(t)
				logger.Infof("write data: %v, total data:%v, write time: %v, per time: %v, total time: %v, avg time: %v",
					len(writeBuf), totalData, time.Since(t), time.Since(t)/time.Duration(len(writeBuf)), totalTime, totalTime/time.Duration(totalData),
				)
				t = time.Now()
				writeBuf = writeBuf[:0]
			}
		}
		if len(writeBuf) > 0 {
			mulWrite(engine.InvDB, writeBuf)
			totalData += len(writeBuf)
			totalTime += time.Since(t)
			logger.Infof("write data: %v, total data:%v, write time: %v, per time: %v, total time: %v, avg time: %v",
				len(writeBuf), totalData, time.Since(t), time.Since(t)/time.Duration(len(writeBuf)), totalTime, totalTime/time.Duration(totalData),
			)
			t = time.Now()
			writeBuf = writeBuf[:0]
		}
		logger.Infoln("bucket:", i, "merge done")
	}

}

func getDoc(docID uint32) *pb.DocIndex {
	doc := &pb.DocIndex{}
	data, _ := engine.DocDB.Get(tools.U32ToBytes(docID), engine.Buckets[docID%engine.BoltBucketSize])
	proto.Unmarshal(data, doc)
	return doc
}

func set(ids []uint32) []uint32 {
	ht := make(map[uint32]bool, len(ids))
	for _, id := range ids {
		ht[id] = true
	}
	ids = ids[:0]
	for k, _ := range ht {
		ids = append(ids, k)
	}
	return ids
}

func AllIdxs() {
	// writeBuf := make([]*wtObj, 0)
	// writeBufSize := 2500
	stat := make(map[uint64]int, 1000)
	for i := 0; i < engine.BoltBucketSize; i++ {
		logger.Infoln("print bucket:", i)
		bucketName := tools.U32ToBytes(uint32(i))
		vals, err := engine.InvDB.GetVals(bucketName)
		if err != nil {
			logger.Panic(err)
		}
		for _, item := range vals {
			invItem := &pb.InvIndex{}
			proto.Unmarshal(item, invItem)
			stat[invItem.Key] = len(invItem.Ids)

		}

	}

	file := "./pkg/data/wordIdx.txt"
	f, _ := os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0664)
	wt := bufio.NewWriter(f)
	for k, v := range stat {
		wt.WriteString(fmt.Sprintf("%d %d\n", k, v))
	}
	wt.Flush()
	// logger.Infoln("data: ", stat)
	logger.Infoln(len(stat))
}
