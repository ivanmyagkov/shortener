package storage

import (
	"sync"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
)

type DB struct {
	sync.Mutex
	ShortURL map[string]string
}

func NewDBConn() *DB {
	return &DB{
		ShortURL: make(map[string]string),
	}
}

func (db *DB) GetURL(shortURL string) (string, error) {
	db.Mutex.Lock()
	if v, ok := db.ShortURL[shortURL]; ok {
		return v, nil
	}
	db.Mutex.Unlock()
	return "", interfaces.ErrNotFound
}

func (db *DB) SetShortURL(shortURL string, URL string) error {
	db.Mutex.Lock()
	if _, ok := db.ShortURL[shortURL]; ok {
		return interfaces.ErrAlreadyExists
	}
	db.ShortURL[shortURL] = URL
	db.Mutex.Unlock()
	return nil

}
func (db *DB) Close() {
	db.ShortURL = nil

}
