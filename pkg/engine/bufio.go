package engine

import (
	"sync"
	"time"

	"github.com/z-y-x233/goSearch/pkg/db"
	"github.com/z-y-x233/goSearch/pkg/logger"
)

type BufReader struct {
	BufSize int
	Buf     []*ReadObj
	ObjCh   chan []byte
	readCh  chan *ReadObj
	readWg  sync.WaitGroup
	objWg   sync.WaitGroup
	db      *db.BoltDb

	debugCh chan int
}

func NewBufReader(db *db.BoltDb) *BufReader {
	return &BufReader{
		Buf: make([]*ReadObj, 0),
		db:  db,
	}
}

func (r *BufReader) Read(obj *ReadObj) {
	r.readWg.Add(1)
	r.readCh <- obj
}
func (r *BufReader) Start(readBufSize, objBufSize int) {
	r.BufSize = ReadBufSize
	r.ObjCh = make(chan []byte, objBufSize)
	r.readCh = make(chan *ReadObj, readBufSize)
	r.debugCh = make(chan int)
	r.readWg = sync.WaitGroup{}
	r.objWg = sync.WaitGroup{}
	r.objWg.Add(1)

	go func() {
		t := time.Now()
		total := 0
		totalTime := time.Second * 0
		for len := range r.debugCh {
			total += len
			totalTime += time.Since(t)
			logger.Debugf(
				"read %v data, total data: %v, read time: %v, per time: %v, total time: %v, avg time: %v",
				len, total, time.Since(t), time.Since(t)/time.Duration(len), totalTime, totalTime/time.Duration(total),
			)
			t = time.Now()
		}
	}()

	go func() {
		for item := range r.readCh {
			r.readWg.Done()
			r.Buf = append(r.Buf, item)
			if len(r.Buf) >= r.BufSize {
				r.mulRead()
				r.debugCh <- len(r.Buf)
				r.Buf = r.Buf[:0]
			}
		}

		if len(r.Buf) > 0 {
			r.mulRead()
			r.debugCh <- len(r.Buf)
			r.Buf = r.Buf[:0]
		}
		r.objWg.Done()
	}()
}

func (r *BufReader) GetData() (res [][]byte) {
	t := time.Now()
	r.readWg.Wait()
	close(r.readCh)
	tt := t
	go func() {
		r.objWg.Wait()
		close(r.ObjCh)
	}()
	for item := range r.ObjCh {
		logger.Debugf("get one obj time: %v", time.Since(t))
		t = time.Now()
		r.objWg.Done()
		res = append(res, item)
	}
	logger.Debugf("get data time: %v", time.Since(tt))
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

type BufWriter struct {
	BufSize int
	Buf     []*WriteObj
	ObjCh   chan []byte
	writeCh chan *WriteObj
	writeWg sync.WaitGroup
	objWg   sync.WaitGroup
	db      *db.BoltDb

	debugCh chan int
}

func NewBufWriter(db *db.BoltDb) *BufWriter {
	return &BufWriter{
		BufSize: ReadBufSize,
		Buf:     make([]*WriteObj, 0),
		writeCh: make(chan *WriteObj, WriteBufSize*100),
		writeWg: sync.WaitGroup{},
		objWg:   sync.WaitGroup{},
		db:      db,
	}
}

func (w *BufWriter) Write(obj *WriteObj) {
	w.writeWg.Add(1)
	w.writeCh <- obj
}
func (w *BufWriter) Start() {
	w.writeCh = make(chan *WriteObj, WriteBufSize*100)
	w.writeWg = sync.WaitGroup{}
	w.objWg = sync.WaitGroup{}
	w.debugCh = make(chan int)
	w.objWg.Add(1)

	go func() {
		t := time.Now()
		total := 0
		totalTime := time.Second * 0
		for len := range w.debugCh {
			total += len
			totalTime += time.Since(t)
			logger.Debugf(
				"write %v data, total data: %v, write time: %v, per time: %v, total time: %v, avg time: %v",
				len, total, time.Since(t), time.Since(t)/time.Duration(len), totalTime, totalTime/time.Duration(total),
			)
			t = time.Now()
		}
	}()

	go func() {
		for item := range w.writeCh {
			w.writeWg.Done()
			w.Buf = append(w.Buf, item)
			if len(w.Buf) >= w.BufSize {
				w.mulWrite()
				w.debugCh <- len(w.Buf)
				w.Buf = w.Buf[:0]
			}
		}

		if len(w.Buf) > 0 {
			w.mulWrite()
			w.debugCh <- len(w.Buf)
			w.Buf = w.Buf[:0]
		}
		w.objWg.Done()
	}()
}

func (w *BufWriter) Wait() {
	w.writeWg.Wait()
	close(w.writeCh)
	w.objWg.Wait()
}

func (w *BufWriter) mulWrite() {
	var wg sync.WaitGroup
	wg.Add(len(w.Buf))
	w.objWg.Add(1)
	for _, item := range w.Buf {
		go func(item *WriteObj) {
			w.db.MulSet(item.Key, item.Val, item.Bucket)
			wg.Done()
		}(item)
	}
	wg.Wait()
	w.objWg.Done()
}
