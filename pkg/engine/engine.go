package engine

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/wangbin/jiebago"
	"github.com/z-y-x233/goSearch/pkg/db"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/model"
	"github.com/z-y-x233/goSearch/pkg/protobuf/pb"
	"github.com/z-y-x233/goSearch/pkg/tools"
	"google.golang.org/protobuf/proto"
)

type Engine struct {
	DocDB   *db.BoltDb
	InvDB   *db.BoltDb
	Buckets [][]byte
	Seg     jiebago.Segmenter
	WordIds map[uint64]int

	InvReader *BufReader
	InvWriter *BufWriter
	DocReader *BufReader
	DocWriter *BufWriter

	wg sync.WaitGroup
}

type ReadObj struct {
	Key    []byte
	Bucket []byte
}

type WriteObj struct {
	Key    []byte
	Val    []byte
	Bucket []byte
}

const (
	ReadBufSize    = 200
	WriteBufSize   = 2500
	BoltBucketSize = 100
	MaxResultSize  = 1000
)

type Options struct {
	DocPath     string
	InvPath     string
	DictPath    string
	WordIdsPath string
	ReadOnly    bool
}

func (o *Options) String() string {
	return fmt.Sprintf(
		"DocPath: %s, InvPath: %s, DictPath: %s, ReadOnly: %v, WordIdsPath: %v",
		o.DocPath, o.InvPath, o.DictPath, o.ReadOnly, o.WordIdsPath)
}

func DefaultOptions() *Options {
	return &Options{
		DocPath:     path.Join(viper.GetString("db.doc.dir"), viper.GetString("db.doc.bolt")),
		InvPath:     path.Join(viper.GetString("db.invIndex.dir"), viper.GetString("db.invIndex.bolt")),
		DictPath:    viper.GetString("db.dict"),
		WordIdsPath: viper.GetString("db.wordIds"),
		ReadOnly:    false,
	}
}

func (e *Engine) LoadWordIds(path string) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0664)
	tools.HandleError("load word ids failed:", err)
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = ' '
	lines, err := reader.ReadAll()
	tools.HandleError("read failed:", err)
	for _, line := range lines {
		word, _ := strconv.ParseUint(line[0], 10, 0)
		ids, _ := strconv.Atoi(line[1])
		e.WordIds[word] = ids
	}
}

func NewEngine(option *Options) *Engine {
	var (
		e   *Engine = &Engine{}
		err error
	)
	logger.Infoln("========================== Init Engine ==========================")
	logger.Infoln("Options:", option)
	logger.Infoln("========================== open database ==========================")
	e.wg = sync.WaitGroup{}
	e.wg.Add(1)
	err = e.Seg.LoadDictionary(option.DictPath)
	tools.HandleError(fmt.Sprintf("load %s failed:", option.DictPath), err)
	e.DocDB, err = db.Open(option.DocPath, option.ReadOnly)
	tools.HandleError(fmt.Sprintf("open %s failed:", option.DocPath), err)
	e.Buckets = make([][]byte, 0)
	e.WordIds = map[uint64]int{}
	for i := 0; i < BoltBucketSize; i++ {
		bucketName := tools.U32ToBytes(uint32(i))
		e.Buckets = append(e.Buckets, bucketName)
		if !option.ReadOnly {
			err := e.DocDB.CreateBucketIfNotExist(bucketName)
			tools.HandleError(fmt.Sprintf("create doc %d bucket failed:", i), err)
		}

	}

	e.InvDB, err = db.Open(option.InvPath, option.ReadOnly)
	tools.HandleError(fmt.Sprintf("open %s failed:", option.InvPath), err)

	for i := 0; i < BoltBucketSize; i++ {
		bucketName := tools.U32ToBytes(uint32(i))
		if !option.ReadOnly {
			err := e.InvDB.CreateBucketIfNotExist(bucketName)
			tools.HandleError(fmt.Sprintf("create inv %d bucket failed:", i), err)
		}
	}
	e.InvReader = NewBufReader(e.InvDB)
	e.InvWriter = NewBufWriter(e.InvDB)
	e.DocReader = NewBufReader(e.DocDB)
	e.DocWriter = NewBufWriter(e.DocDB)
	logger.Infoln("========================== load word ids ==========================")
	e.LoadWordIds(option.WordIdsPath)
	logger.Infoln("========================== Init Done ==========================")
	e.wg.Done()
	return e
}

func (e *Engine) Close() {
	e.InvDB.Close()
	e.DocDB.Close()
}

func (e *Engine) Wait() {
	e.wg.Wait()
}

func DefaultEngine() *Engine {
	return NewEngine(DefaultOptions())
}

func getWords(ch <-chan string) (words []string) {
	for word := range ch {
		words = append(words, word)
	}
	return
}

func (e *Engine) WordCutForInv(q string) []string {
	//不区分大小写
	q = strings.ToLower(q)
	//移除所有的标点符号
	q = tools.RemovePunctuation(q)
	//移除所有的空格
	q = tools.RemoveSpace(q)
	ch := e.Seg.CutForSearch(q, true)
	words := getWords(ch)

	return words
}

func (e *Engine) WordCut(q string) (res []string) {

	//不区分大小写
	q = strings.ToLower(q)
	//移除所有的标点符号
	q = tools.RemovePunctuation(q)
	//移除所有的空格
	q = tools.RemoveSpace(q)

	ch := e.Seg.CutForSearch(q, true)
	words := getWords(ch)
	wordSet := make(map[string]int, 10)
	length := 0
	for _, word := range words {
		wordSet[word]++
		if wordSet[word] == 1 {
			key := tools.Str2Uint64(word)
			length += e.WordIds[key]
		}
	}
	logger.Info(wordSet)
	for word := range wordSet {
		key := tools.Str2Uint64(word)
		logger.Infoln(word, len([]rune(word)), e.WordIds[key])
		if len([]rune(word)) == 1 && e.WordIds[key] >= 10000 {
			continue
		}
		if len(words) >= 5 && e.WordIds[key]*2 >= length {
			continue
		}
		res = append(res, word)
	}
	if len(res) == 0 {
		res = append(res, words[0])
	}
	return res
}

func (e *Engine) getDoc(docID uint32) *pb.DocIndex {
	doc := &pb.DocIndex{}
	data, _ := e.DocDB.Get(tools.U32ToBytes(docID), e.Buckets[docID%BoltBucketSize])
	proto.Unmarshal(data, doc)
	return doc
}

func (e *Engine) GetDocs(docIDs model.Docs) []*pb.DocIndex {
	docs := []*pb.DocIndex{}
	for _, doc := range docIDs {
		docs = append(docs, e.getDoc(doc.Id))
	}
	return docs
}

func (e *Engine) GetInvItems(word uint64) *pb.InvIndex {
	key := tools.U64ToBytes(word)
	bucket := e.Buckets[word%BoltBucketSize]
	buf, f := e.InvDB.Get(key, bucket)
	if !f {
		return nil
	}
	item := &pb.InvIndex{}
	err := proto.Unmarshal(buf, item)
	tools.HandleError("unmarshal fail:", err)
	return item
}

func (e *Engine) idf(word uint64, N int) float64 {

	return math.Log(float64(N+1)) / math.Log(float64(float64(e.WordIds[word])+0.5))

}

func (e *Engine) r(word uint64, tf int) float64 {
	var (
		k1 float64 = 1.2
		b  float64 = 0.75
	)
	up := k1 * float64(tf)
	down := k1*(1-b) + float64(tf)
	return up / down
}

func (e *Engine) Query(q string) (res model.Docs) {
	words := e.WordCut(q)
	wordMap := make(map[uint64]*pb.InvIndex, len(words))
	logger.Infoln(words)
	tt := time.Now()
	t := time.Now()
	for _, word := range words {
		key := tools.Str2Uint64(word)
		item := e.GetInvItems(key)
		if item != nil {
			wordMap[key] = item
		}
	}
	parseTime := time.Since(t)

	docScore := make(map[uint32]float64, e.WordIds[tools.Str2Uint64(words[0])])
	t = time.Now()
	for _, item := range wordMap {
		for _, item2 := range item.Items {
			if _, ok := docScore[item2.Id]; !ok {
				docScore[item2.Id] = 0
			}
		}
	}
	st := time.Since(t)
	logger.Infoln("total time:", time.Since(tt), "find time:", parseTime, "find docs:", len(docScore), "stat time:", st)
	t = time.Now()
	logger.Infoln("Start cal doc score")

	for _, item := range wordMap {
		for _, doc := range item.Items {
			docScore[doc.Id] += e.idf(item.Key, len(docScore)) * e.r(item.Key, int(doc.Cnt))
		}
	}
	logger.Infoln("Cal score time:", time.Since(t))
	t = time.Now()
	for uid, score := range docScore {
		res = append(res, &model.SliceItem{Id: uid, Score: score})
	}
	sort.Sort(res)
	logger.Infoln("Sort time:", time.Since(t))
	if len(res) > MaxResultSize {
		res = res[:MaxResultSize]
	}
	return res
}

func (e *Engine) Search(q string) []*pb.DocIndex {
	slices := e.Query(q)
	t := time.Now()
	res := e.GetDocs(slices)
	logger.Infoln("Get docs time:", time.Since(t))
	return res
}
func (e *Engine) ParseDoc() {
	path := path.Join(viper.GetString("db.invIndex.dir"), viper.GetString("db.invIndex.bolt3"))
	db2, _ := db.Open(path, false)
	write := NewBufWriter(db2)
	write.Start()
	t := time.Now()
	for i := 0; i < BoltBucketSize; i++ {
		logger.Infoln("parse bucket:", i)
		bucketName := tools.U32ToBytes(uint32(i))
		db2.CreateBucketIfNotExist(bucketName)
		vals, err := e.InvDB.GetVals(bucketName)
		tools.HandleError("GetDoc error:", err)
		for _, item := range vals {
			invItem := &pb.InvIndex{}
			proto.Unmarshal(item, invItem)
			write.Write(&WriteObj{Key: tools.U64ToBytes(invItem.Key), Val: item, Bucket: bucketName})
		}
	}
	write.Wait()
	logger.Infoln("total time:", time.Since(t))
}
