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

var ErrKeyNotFound = errors.New("Key not found")
