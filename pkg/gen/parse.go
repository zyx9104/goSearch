package gen

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/viper"
	"github.com/z-y-x233/goSearch/pkg/db"
	"github.com/z-y-x233/goSearch/pkg/engine"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/protobuf/pb"
	"github.com/z-y-x233/goSearch/pkg/tools"
	"google.golang.org/protobuf/proto"
)

func BuildInvIdx(start, end, id int) {

	logger.Infoln("===========================start build===================================")
	logger.Infoln("start:", start, "end:", end)

	// e := engine.DefaultEngine()
	o := engine.DefaultOptions()
	s := fmt.Sprintf("db.invIndex.bolt%d", id)
	o.InvPath = path.Join(viper.GetString("db.invIndex.dir"), viper.GetString(s))
	e := engine.NewEngine(o)
	e.Wait()
	e.DocReader.Start(1000000, 1000000)
	b := time.Now()
	for i := start; i < end; i++ {
		key := tools.U32ToBytes(uint32(i))
		bucket := e.Buckets[i%engine.BoltBucketSize]
		e.DocReader.Read(&engine.ReadObj{Key: key, Bucket: bucket})

	}
	docData := e.DocReader.GetData()

	logger.Infof("from %v to %v, read time: %v", start, end, time.Since(b))

	invMap := make(map[uint64][]*pb.InvItem, 1000000)
	parseDoc := func(doc *pb.DocIndex) {
		wordMap := make(map[string]int, 100)
		words := e.WordCutForInv(doc.Text)
		for _, word := range words {
			wordMap[word]++
		}
		for word, cnt := range wordMap {
			key := tools.Str2Uint64(word)
			invMap[key] = append(invMap[key], &pb.InvItem{Id: doc.Id, Cnt: int32(cnt)})
		}
	}
	logger.Infoln("Start parse doc")
	dt := time.Now()
	cnt := 0
	totalData := 0
	totalTime := time.Second * 0
	t := time.Now()
	for _, data := range docData {
		doc := &pb.DocIndex{}
		proto.Unmarshal(data, doc)
		parseDoc(doc)
		cnt++
		if cnt == 100000 {
			totalData += 100000
			totalTime += time.Since(t)
			logger.Infof(
				"parse 10w data, total time: %v, total data: %v, parse time: %v, per time: %v, avg time: %v",
				totalTime, totalData, time.Since(t), time.Since(t)/time.Duration(100000), totalTime/time.Duration(totalData),
			)
			cnt = 0
			t = time.Now()
		}
	}
	logger.Infoln("Parse doc time:", time.Since(dt))

	logger.Infoln("Start write inv index")
	e.InvWriter.Start()

	wt := time.Now()
	for k, v := range invMap {
		key := tools.U64ToBytes(k)
		bucket := e.Buckets[k%engine.BoltBucketSize]
		val, err := proto.Marshal(&pb.InvIndex{Key: k, Items: v})
		tools.HandleError("marshal inv failed:", err)
		e.InvWriter.Write(&engine.WriteObj{Key: key, Val: val, Bucket: bucket})
		delete(invMap, k)
	}
	e.InvWriter.Wait()
	logger.Infof("write time: %v", time.Since(wt))
	logger.Infoln("===========================build inv index done===================================")
}

func MergeIndex() {
	e := engine.DefaultEngine()
	e.Wait()
	e.InvWriter.Start()
	t := time.Now()
	for i := 1; i <= 1; i++ {
		path := path.Join(viper.GetString("db.invIndex.dir"), viper.GetString(fmt.Sprintf("db.invIndex.bolt%d", i)))
		db, _ := db.Open(path, false)
		for j := 0; j < engine.BoltBucketSize; j++ {
			bucket := e.Buckets[j]
			vals, _ := db.GetVals(bucket)
			for _, item := range vals {
				invItem := &pb.InvIndex{}
				proto.Unmarshal(item, invItem)

				buf, f := e.InvDB.Get(tools.U64ToBytes(invItem.Key), bucket)
				if f {
					temp := &pb.InvIndex{}
					proto.Unmarshal(buf, temp)
					invItem.Items = append(invItem.Items, temp.Items...)
				}
				data, _ := proto.Marshal(invItem)
				e.InvWriter.Write(&engine.WriteObj{Key: tools.U64ToBytes(invItem.Key), Val: data, Bucket: bucket})
			}
		}
		e.InvWriter.Wait()
		logger.Infoln("merge", i, "done", "time:", time.Since(t))
		t = time.Now()
		e.InvWriter.Start()
	}
	e.InvWriter.Wait()
}

func GenWordIds() {
	e := engine.DefaultEngine()
	wc := make(map[uint64]int, 8000000)
	for i := 0; i < engine.BoltBucketSize; i++ {
		bucket := e.Buckets[i]
		vals, err := e.InvDB.GetVals(bucket)
		tools.HandleError("", err)
		for _, bytes := range vals {
			inv := &pb.InvIndex{}
			proto.Unmarshal(bytes, inv)
			wc[inv.Key] = len(inv.Items)
		}
		logger.Infoln("bucket", i, "done")
	}
	file, _ := os.OpenFile("./pkg/data/word_ids.txt", os.O_CREATE|os.O_RDWR, 0664)
	wt := bufio.NewWriter(file)
	for k, v := range wc {
		wt.WriteString(fmt.Sprintf("%v %v\n", k, v))
	}
	wt.Flush()
}
