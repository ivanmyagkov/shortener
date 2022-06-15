package storage

import (
	"sync"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
)

type DB struct {
	sync.Mutex
	Storage  map[string]string
	ShortURL map[string][]interfaces.ModelURL
}

func NewDBConn() *DB {
	return &DB{
		Storage:  make(map[string]string),
		ShortURL: make(map[string][]interfaces.ModelURL),
	}
}

func (db *DB) GetURL(shortURL string) (string, error) {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.Storage[shortURL]; ok {
		return db.Storage[shortURL], nil
	}
	return "", interfaces.ErrNotFound
}

func (db *DB) GetAllURLsByUserID(userID string) ([]interfaces.ModelURL, error) {
	var ok bool
	if _, ok = db.ShortURL[userID]; ok {
		return db.ShortURL[userID], nil
	}
	return nil, interfaces.ErrNotFound
}
func (db *DB) DelBatchShortURLs(tasks []interfaces.Task) error {

	return nil
}
func (db *DB) SetShortURL(userID, shortURL, URL string) error {
	db.Lock()
	defer db.Unlock()
	modelURL := interfaces.ModelURL{
		ShortURL: shortURL,
		BaseURL:  URL,
	}
	if _, ok := db.ShortURL[userID]; ok {
		for _, val := range db.ShortURL[userID] {
			if val.ShortURL == shortURL {
				return interfaces.ErrAlreadyExists
			}
		}
	}
	db.ShortURL[userID] = append(db.ShortURL[userID], modelURL)
	db.Storage[modelURL.ShortURL] = modelURL.BaseURL
	return nil
}
func (db *DB) Ping() error {
	return nil
}
func (db *DB) Close() error {
	db.ShortURL = nil
	return nil

}
