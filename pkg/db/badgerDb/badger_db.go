package badgerDb

import (
	"log"
	"sync"

	"github.com/dgraph-io/badger/v3"
)

type BadgerDb struct {
	db   *badger.DB
	path string
	sync.Mutex
}

// OpenWithDefault open a badgerDB with default setting
func OpenWithDefault(path string) *BadgerDb {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		log.Fatal(err)
	}
	return &BadgerDb{
		db:   db,
		path: path,
	}
}

func Open(option badger.Options) *BadgerDb {
	db, err := badger.Open(option)
	if err != nil {
		log.Fatal(err)
	}
	return &BadgerDb{
		db:   db,
		path: option.Dir,
	}
}

func (b *BadgerDb) GetKeys() (keys [][]byte, er error) {

	err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			keys = append(keys, k)
		}
		return nil
	})

	if err != nil {
		log.Println("Failed to iterator keys from the cache.", "error", err)
		return nil, err
	}
	return keys, nil
}

// Get if not find bool will return false
func (b *BadgerDb) Get(key []byte) ([]byte, bool) {
	result := make([]byte, 0)
	err := b.db.View(func(txn *badger.Txn) error {
		value, err := txn.Get(key)
		if err != nil {
			return err
		}
		return value.Value(func(val []byte) error {
			result = append(result, val...)
			return nil
		})
	})
	if err != nil {
		return nil, false
	}
	return result, true
}

func (b *BadgerDb) Set(key []byte, value []byte) error {
	b.Lock()
	defer b.Unlock()
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (b *BadgerDb) Delete(key []byte) error {
	b.Lock()
	defer b.Unlock()
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (b *BadgerDb) Close() error {
	return b.db.Close()
}
