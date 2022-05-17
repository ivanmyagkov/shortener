package storage

import (
	"log"
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
	db.Lock()
	defer db.Unlock()
	if v, ok := db.ShortURL[shortURL]; ok {
		return v, nil
	}
	return "", interfaces.ErrNotFound
}

func (db *DB) SetShortURL(shortURL string, URL string) error {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.ShortURL[shortURL]; ok {
		log.Println(interfaces.ErrAlreadyExists)
		return nil
	}
	db.ShortURL[shortURL] = URL

	return nil

}
func (db *DB) Close() {
	db.ShortURL = nil

}
