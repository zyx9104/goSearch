package engine

import (
	"path"
	"sync"
	"time"

	"github.com/spf13/viper"
	"github.com/wangbin/jiebago"
	"github.com/z-y-x233/goSearch/pkg/db"
	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/protobuf/pb"
	"github.com/z-y-x233/goSearch/pkg/tools"
	"google.golang.org/protobuf/proto"
)

var (
	DocDB      *db.BoltDb
	InvDB      *db.BoltDb
	Buckets    [][]byte
	InvBuckets [][]byte

	Seg jiebago.Segmenter
)

type ReadObj struct {
	Key    []byte
	Bucket []byte
}

type WriteObj struct {
	ReadObj
	Val []byte
}

const (
	ReadBufSize    = 1500
	WriteBufSize   = 2500
	BoltBucketSize = 100
)

type BufReader struct {
	BufSize int
	Buf     []*ReadObj
	ObjCh   chan []byte
	readCh  chan *ReadObj
	readWg  sync.WaitGroup
	objWg   sync.WaitGroup
	db      *db.BoltDb
}

func Init(invID int) {

	Seg.LoadDictionary("./pkg/data/dict.txt")

	logger.Infoln("========================== open bolt doc database ==========================")
	database := path.Join(viper.GetString("db.doc.dir"), viper.GetString("db.doc.bolt"))

	DocDB, _ = db.Open(database, false)

	Buckets = make([][]byte, 0)

	for i := 0; i < BoltBucketSize; i++ {
		bucketName := tools.U32ToBytes(uint32(i))
		Buckets = append(Buckets, bucketName)
		err := DocDB.CreateBucketIfNotExist(bucketName)
		if err != nil {
			return
		}
	}
	logger.Infoln("========================== open bolt doc done ==========================")

	logger.Infoln("========================== open bolt inv database ==========================")

	dir := viper.GetString("db.invIndex.dir")
	filename := viper.GetString("db.invIndex.bolt")
	database = path.Join(dir, filename)

	InvDB, _ = db.Open(database, false)

	InvBuckets = make([][]byte, 0)

	for i := 0; i < BoltBucketSize; i++ {
		bucketName := tools.U32ToBytes(uint32(i))
		InvBuckets = append(InvBuckets, bucketName)
		err := InvDB.CreateBucketIfNotExist(bucketName)
		if err != nil {
			return
		}
	}
	logger.Infoln("========================== open bolt inv done ==========================")

}

func NewBufReader(db *db.BoltDb) *BufReader {
	return &BufReader{
		BufSize: ReadBufSize,
		Buf:     make([]*ReadObj, 0),
		ObjCh:   make(chan []byte, ReadBufSize*1000),
		readCh:  make(chan *ReadObj, ReadBufSize*1000),
		readWg:  sync.WaitGroup{},
		objWg:   sync.WaitGroup{},
		db:      db,
	}
}

func (r *BufReader) Read(obj *ReadObj) {
	r.readWg.Add(1)
	// logger.Info("readWg add")
	r.readCh <- obj
}
func (r *BufReader) Start() {
	r.objWg.Add(1)
	for item := range r.readCh {
		r.readWg.Done()
		// logger.Info("readWg done")
		r.Buf = append(r.Buf, item)
		if len(r.Buf) >= r.BufSize {
			r.mulRead()
			r.Buf = r.Buf[:0]
		}
	}

	if len(r.Buf) > 0 {
		r.mulRead()
		r.Buf = r.Buf[:0]
	}
	r.objWg.Done()
}

func (r *BufReader) GetData() (res [][]byte) {
	r.readWg.Wait()
	close(r.readCh)

	go func() {
		r.objWg.Wait()
		close(r.ObjCh)
	}()
	for item := range r.ObjCh {
		r.objWg.Done()
		res = append(res, item)
	}
	return res
}

func (r *BufReader) mulRead() {
	var wg sync.WaitGroup
	wg.Add(len(r.Buf))
	r.objWg.Add(1)
	for _, item := range r.Buf {
		go func(item *ReadObj) {
			data, found := r.db.Get(item.Key, item.Bucket)
			if found {
				r.objWg.Add(1)
				r.ObjCh <- data
			}
			wg.Done()
		}(item)
	}
	wg.Wait()
	r.objWg.Done()
}

func InvMarshal(item *pb.InvIndex) (data []byte, err error) {
	data, err = proto.Marshal(item)
	return
}

func DocMarshal(item *pb.DocIndex) (data []byte, err error) {
	data, err = proto.Marshal(item)
	return
}

func InvUnmarshal(data []byte) (item *pb.InvIndex, err error) {
	item = &pb.InvIndex{}
	err = proto.Unmarshal(data, item)
	return
}

func DocUnmarshal(data []byte) (item *pb.DocIndex, err error) {
	item = &pb.DocIndex{}
	err = proto.Unmarshal(data, item)
	return
}

type BufWriter struct {
}

func getWords(ch <-chan string) (words []string) {
	for word := range ch {
		words = append(words, word)
	}
	return
}

func WordCut(q string) []string {
	ch := Seg.CutForSearch(q, true)
	words := getWords(ch)

	return words
}

func getDoc(docID uint32) *pb.DocIndex {
	doc := &pb.DocIndex{}
	data, _ := DocDB.Get(tools.U32ToBytes(docID), Buckets[docID%BoltBucketSize])
	proto.Unmarshal(data, doc)
	return doc
}

func GetDocs(docIDs []uint32) []*pb.DocIndex {
	docs := []*pb.DocIndex{}
	for _, uid := range docIDs {
		docs = append(docs, getDoc(uid))
	}
	return docs
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

func Query(q string) []*pb.DocIndex {

	words := WordCut(q)
	docs := []*pb.DocIndex{}
	logger.Infoln(words)
	ids := []uint32{}
	tt := time.Now()
	invReader := NewBufReader(InvDB)
	docReader := NewBufReader(DocDB)
	go invReader.Start()
	go docReader.Start()

	for _, word := range words {
		id := tools.Str2Uint64(word)
		key := tools.U64ToBytes(id)
		bucket := Buckets[id%BoltBucketSize]
		invReader.Read(&ReadObj{Key: key, Bucket: bucket})
	}

	invData := invReader.GetData()
	for _, item := range invData {
		// logger.Infoln(item)
		inv, err := InvUnmarshal(item)
		if err != nil {
			logger.Panic("unmarshal failed:", err)
		}
		ids = append(ids, inv.Ids...)
	}
	ids = set(ids)
	// ids = ids[:100000]
	logger.Infoln("read idx time:", time.Since(tt), "words:", len(words), "ids:", len(ids))
	tt = time.Now()
	for _, uid := range ids {
		key := tools.U32ToBytes(uid)
		bucket := Buckets[uid%BoltBucketSize]
		docReader.Read(&ReadObj{Key: key, Bucket: bucket})
	}
	// docReader.GetData()
	ut := time.Second * 0
	rtt := time.Now()
	docData := docReader.GetData()
	rt := time.Since(rtt)
	for _, item := range docData {
		utt := time.Now()
		doc, err := DocUnmarshal(item)
		ut += time.Since(utt)
		if err != nil {
			logger.Panic(err)
		}
		docs = append(docs, doc)
	}

	logger.Infoln("read doc time:", rt, "unmarshal time:", ut, "total time:", time.Since(tt))
	return docs
}

func Search(word string) {

}
