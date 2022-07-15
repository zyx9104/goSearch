package db

import (
	"github.com/boltdb/bolt"
)

type BoltDb struct {
	db   *bolt.DB
	path string
}

func Open(path string, readOnly bool) (*BoltDb, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{ReadOnly: readOnly})

	if err != nil {
		return nil, err
	}
	return &BoltDb{
		db:   db,
		path: path,
	}, nil
}

// CreateBucket 创建一个桶
func (s *BoltDb) CreateBucket(bucketName []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(bucketName)
		return err
	})
}

// DeleteBucket 删除一个桶
func (s *BoltDb) DeleteBucket(bucketName []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(bucketName)
	})
}

// CreateBucketIfNotExist 如果桶不存在则创建
func (s *BoltDb) CreateBucketIfNotExist(bucketName []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
}

// Get -> if key not exist will return nil.
func (s *BoltDb) Get(key []byte, bucketName []byte) ([]byte, bool) {
	var buffer []byte

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		buffer = b.Get(key)
		return nil
	})

	if err != nil {
		return nil, false
	}
	if buffer == nil {
		return nil, false
	}
	return buffer, true
}

func (s *BoltDb) Set(key []byte, value []byte, bucketName []byte) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		err := b.Put(key, value)
		return err
	})

	return err
}

func (s *BoltDb) GetVals(bucketName []byte) (vals [][]byte, err error) {

	err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			vals = append(vals, v)
		}
		return nil
	})

	return
}

func (s *BoltDb) MulSet(key []byte, value []byte, bucketName []byte) error {
	err := s.db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		err := b.Put(key, value)
		return err
	})

	return err
}

// Delete 删除
func (s *BoltDb) Delete(key []byte, bucketName []byte) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Delete(key)
	})

	return err
}

// Close 关闭
func (s *BoltDb) Close() error {
	return s.db.Close()
}
