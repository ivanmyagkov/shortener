package storage

import "github.com/ivanmyagkov/shortener.git/internal/interfaces"

type DB struct {
	ShortURL map[string]string
}

func NewDBConn() *DB {
	return &DB{
		ShortURL: make(map[string]string),
	}
}

func (db *DB) GetURL(shortURL string) (string, error) {
	if v, ok := db.ShortURL[shortURL]; ok {
		return v, nil
	}
	return "", interfaces.ErrNotFound
}

func (db *DB) SetShortURL(shortURL string, URL string) error {
	if _, ok := db.ShortURL[shortURL]; ok {
		return interfaces.ErrAlreadyExists
	}
	db.ShortURL[shortURL] = URL
	return nil

}
func (db *DB) Close() {
	db.ShortURL = nil

}
