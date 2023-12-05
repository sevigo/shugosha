package db

import (
	"errors"

	"github.com/dgraph-io/badger/v4"
	"github.com/sevigo/shugosha/pkg/model"
)

// BadgerDB is an implementation of the DB interface using Badger.
type BadgerDB struct {
	db *badger.DB
}

func NewBadgerDB(dbPath string) (*BadgerDB, error) {
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &BadgerDB{db: db}, nil
}

func (b *BadgerDB) Get(key string) ([]byte, error) {
	var valCopy []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return model.ErrDBKeyNotFound
			}
			return err
		}
		valCopy, err = item.ValueCopy(nil)
		return err
	})

	if err == model.ErrDBKeyNotFound {
		return nil, model.ErrDBKeyNotFound
	}
	return valCopy, err
}
func (b *BadgerDB) Set(key string, value []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

func (b *BadgerDB) Close() error {
	return b.db.Close()
}
