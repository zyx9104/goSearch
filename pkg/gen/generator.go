package gen

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"

	"github.com/z-y-x233/goSearch/pkg/db"
	"github.com/z-y-x233/goSearch/pkg/engine"
	"github.com/z-y-x233/goSearch/pkg/log"
	"github.com/z-y-x233/goSearch/pkg/proto/pb"
	"github.com/z-y-x233/goSearch/pkg/tools"
	"github.com/z-y-x233/goSearch/pkg/tree"
)

func BuildInvIdx(start, end, id int) {

	log.Infoln("===========================start build===================================")
	log.Infoln("start:", start, "end:", end)

	// e := engine.DefaultEngine()
	o := engine.DefaultOptions()
	s := fmt.Sprintf("db.invIndex.bolt%d", id)
	o.InvPath = path.Join(viper.GetString("db.invIndex.dir"), viper.GetString(s))
	e := engine.NewEngine(o)
	e.Wait()
	docReader := engine.NewBufReader(e.DocDB)
	docReader.Start(1000000, 1000000)
	b := time.Now()
	for i := start; i < end; i++ {
		key := tools.U32ToBytes(uint32(i))
		bucket := engine.Buckets[i%engine.BoltBucketSize]
		docReader.Read(&engine.ReadObj{Key: key, Bucket: bucket})

	}
	docData := docReader.GetData()

	log.Infof("from %v to %v, read time: %v", start, end, time.Since(b))

	invMap := make(map[uint64][]*pb.InvItem, 1000000)
	parseDoc := func(doc *pb.DocIndex) {
		wordMap := make(map[string]int, 100)
		words := tools.WordCutForInv(doc.Text)
		for _, word := range words {
			wordMap[word]++
		}
		for word, cnt := range wordMap {
			key := tools.Str2Uint64(word)
			invMap[key] = append(invMap[key], &pb.InvItem{Id: doc.Id, Cnt: int32(cnt)})
		}
	}
	log.Infoln("Start parse doc")
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
			log.Infof(
				"parse 10w data, total time: %v, total data: %v, parse time: %v, per time: %v, avg time: %v",
				totalTime, totalData, time.Since(t), time.Since(t)/time.Duration(100000), totalTime/time.Duration(totalData),
			)
			cnt = 0
			t = time.Now()
		}
	}
	log.Infoln("Parse doc time:", time.Since(dt))

	log.Infoln("Start write inv index")
	invWriter := engine.NewBufWriter(e.InvDB)
	invWriter.Start()

	wt := time.Now()
	for k, v := range invMap {
		key := tools.U64ToBytes(k)
		bucket := engine.Buckets[k%engine.BoltBucketSize]
		val, err := proto.Marshal(&pb.InvIndex{Key: k, Items: v})
		tools.HandleError("marshal inv failed:", err)
		invWriter.Write(&engine.WriteObj{Key: key, Val: val, Bucket: bucket})
		delete(invMap, k)
	}
	invWriter.Wait()
	log.Infof("write time: %v", time.Since(wt))
	log.Infoln("===========================build inv index done===================================")
}

func MergeIndex() {
	e := engine.DefaultEngine()
	e.Wait()
	invWriter := engine.NewBufWriter(e.InvDB)

	invWriter.Start()
	t := time.Now()
	for i := 1; i <= 1; i++ {
		path := path.Join(viper.GetString("db.invIndex.dir"), viper.GetString(fmt.Sprintf("db.invIndex.bolt%d", i)))
		db, _ := db.Open(path, false)
		for j := 0; j < engine.BoltBucketSize; j++ {
			bucket := engine.Buckets[j]
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
				invWriter.Write(&engine.WriteObj{Key: tools.U64ToBytes(invItem.Key), Val: data, Bucket: bucket})
			}
		}
		invWriter.Wait()
		log.Infoln("merge", i, "done", "time:", time.Since(t))
		t = time.Now()
		invWriter.Start()
	}
	invWriter.Wait()
}

func GenWordIds() {
	e := engine.DefaultEngine()
	wc := make(map[uint64]int, 8000000)
	for i := 0; i < engine.BoltBucketSize; i++ {
		bucket := engine.Buckets[i]
		vals, err := e.InvDB.GetVals(bucket)
		tools.HandleError("", err)
		for _, bytes := range vals {
			inv := &pb.InvIndex{}
			proto.Unmarshal(bytes, inv)
			wc[inv.Key] = len(inv.Items)
		}
		log.Infoln("bucket", i, "done")
	}
	file, _ := os.OpenFile("./pkg/data/word_ids.txt", os.O_CREATE|os.O_RDWR, 0664)
	wt := bufio.NewWriter(file)
	for k, v := range wc {
		wt.WriteString(fmt.Sprintf("%v %v\n", k, v))
	}
	wt.Flush()
}

func GenSearchHistory() {
	e := engine.DefaultEngine()
	e.Wait()
	n := 300000
	buf := 2500

	tr := tree.NewTrie()
	ids := []uint32{}
	for i := 0; i < n; i++ {
		uid := uint32(rand.Int31n(101483886))
		ids = append(ids, uid)
	}
	t := time.Now()
	docs := e.GetDocsByUid(ids)
	cnt := 0
	log.Info("get docs time:", time.Since(t))
	t = time.Now()
	for _, doc := range docs {
		tr.Insert(doc.Text)
		cnt++
		if cnt == buf {
			log.Debugln("insert", cnt, "query", "insert time:", time.Since(t))
			t = time.Now()
			cnt = 0
		}
	}
	filenema := viper.GetString("db.searchHistory")
	// ss := tr.RelatedSearch("", tr.Size)
	// for _, line := range ss {
	// 	logger.Debug(line)
	// }
	tr.Save(filenema)
	log.Info("save time: ", time.Since(t))
}
