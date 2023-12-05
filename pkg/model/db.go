package model

import "errors"

// DB is an interface for abstracting database operations.
//
//go:generate go run github.com/vektra/mockery/v2@v2 --name=DB --filename=db.go --output=../../mocks/
type DB interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Close() error
}

// ErrDBKeyNotFound is used when a key is not found in the database.
var ErrDBKeyNotFound = errors.New("key not found in the database")
