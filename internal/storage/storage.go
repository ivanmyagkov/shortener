package storage

type DB struct {
	ShortURL map[string]string
}

func NewDBConn() *DB {
	return &DB{
		ShortURL: make(map[string]string),
	}
}

func (db *DB) GetURL(shortURL string) string {
	return db.ShortURL[shortURL]
}

func (db *DB) SetShortURL(shortURL string, URL string) {
	db.ShortURL[shortURL] = URL
}
