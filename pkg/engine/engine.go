package engine

import (
	"fmt"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/z-y-x233/goSearch/pkg/db"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/model"
	"github.com/z-y-x233/goSearch/pkg/proto/pb"
	"github.com/z-y-x233/goSearch/pkg/tools"
	"github.com/z-y-x233/goSearch/pkg/tree"
	"google.golang.org/protobuf/proto"
)

type Engine struct {
	DocDB *db.BoltDb
	InvDB *db.BoltDb
	wg    sync.WaitGroup
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

type Options struct {
	DocPath  string
	InvPath  string
	ReadOnly bool
}

var (
	Buckets [][]byte
)

const (
	ReadBufSize    = 200
	WriteBufSize   = 2500
	BoltBucketSize = 100
	MaxResultSize  = 1000
)

func Init() {
	Buckets = make([][]byte, 0)
	for i := 0; i < BoltBucketSize; i++ {
		bucketName := tools.U32ToBytes(uint32(i))
		Buckets = append(Buckets, bucketName)
	}
}

func (o *Options) String() string {
	return fmt.Sprintf(
		"DocPath: %s, InvPath: %s, ReadOnly: %v",
		o.DocPath, o.InvPath, o.ReadOnly)
}

func DefaultOptions() *Options {
	return &Options{
		DocPath:  path.Join(viper.GetString("db.doc.dir"), viper.GetString("db.doc.bolt")),
		InvPath:  path.Join(viper.GetString("db.invIndex.dir"), viper.GetString("db.invIndex.bolt")),
		ReadOnly: false,
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

	e.DocDB, err = db.Open(option.DocPath, option.ReadOnly)
	tools.HandleError(fmt.Sprintf("open %s failed:", option.DocPath), err)

	for i := 0; i < BoltBucketSize; i++ {
		bucketName := tools.U32ToBytes(uint32(i))
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

	logger.Infoln("========================== Init Done ==========================")
	e.wg.Done()
	return e
}

func (e *Engine) Close() {
	logger.Infoln("Close Database")
	e.InvDB.Close()
	e.DocDB.Close()
}

func (e *Engine) Wait() {
	e.wg.Wait()
}

func DefaultEngine() *Engine {
	return NewEngine(DefaultOptions())
}

func (e *Engine) GetDoc(docID uint32) *pb.DocIndex {
	doc := &pb.DocIndex{}
	data, _ := e.DocDB.Get(tools.U32ToBytes(docID), Buckets[docID%BoltBucketSize])
	proto.Unmarshal(data, doc)
	return doc
}

func (e *Engine) GetDocs(docIDs model.Docs) []*pb.DocIndex {
	docs := []*pb.DocIndex{}
	docReader := NewBufReader(e.DocDB)
	docReader.Start(len(docIDs)*2, MaxResultSize*3/2)

	for _, doc := range docIDs {
		key := tools.U32ToBytes(doc.Id)
		bucket := Buckets[doc.Id%BoltBucketSize]
		docReader.Read(&ReadObj{Key: key, Bucket: bucket})
	}
	data := docReader.GetData()
	for _, item := range data {
		doc := &pb.DocIndex{}
		err := proto.Unmarshal(item, doc)
		tools.HandleError("unmarshal failed:", err)
		docs = append(docs, doc)
	}
	return docs
}

func (e *Engine) GetDocsByUid(docIDs []uint32) []*pb.DocIndex {
	docs := []*pb.DocIndex{}
	for _, uid := range docIDs {
		doc := e.GetDoc(uid)
		docs = append(docs, doc)
	}
	return docs
}

func (e *Engine) GetInvItems(word uint64) *pb.InvIndex {
	key := tools.U64ToBytes(word)
	bucket := Buckets[word%BoltBucketSize]
	buf, f := e.InvDB.Get(key, bucket)
	if !f {
		return nil
	}
	item := &pb.InvIndex{}
	err := proto.Unmarshal(buf, item)
	tools.HandleError("unmarshal fail:", err)
	return item
}

func (e *Engine) Query(q string) (res model.Docs) {
	words := tools.WordCut(q)
	wordMap := make(map[uint64]*pb.InvIndex, len(words))
	invChan := make(chan *pb.InvIndex, 20)
	logger.Infoln(words)
	tt := time.Now()
	t := time.Now()
	var wg sync.WaitGroup
	wg.Add(len(words))
	for _, word := range words {
		key := tools.Str2Uint64(word)
		go func() {
			item := e.GetInvItems(key)
			if item != nil {
				invChan <- item
			}
			wg.Done()
		}()
	}
	wg.Wait()
	close(invChan)
	for item := range invChan {
		wordMap[item.Key] = item
	}
	parseTime := time.Since(t)
	var size uint64 = 0
	if len(words) > 0 {
		size = tools.Str2Uint64(words[0])
	}
	docScore := make(map[uint32]float64, tools.WordIds[size])
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
			docScore[doc.Id] += tools.IDF(item.Key, len(docScore)) * tools.R(int(doc.Cnt))
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

func (e *Engine) FliterResult(docs []*pb.DocIndex, fliterWord []string) (res []*pb.DocIndex) {
	acTrie := tree.NewAC()
	acTrie.Build(fliterWord)
	for _, doc := range docs {
		if !acTrie.Find(doc.Text) {
			res = append(res, doc)
		}
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
