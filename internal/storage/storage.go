package storage

type DB struct {
	ShortURL map[string]string
}

func NewDBConn() *DB {
	return &DB{
		ShortURL: make(map[string]string),
	}
}

func (db *DB) GetURL(shortUrl string) string {
	return db.ShortURL[shortUrl]
}

func (db *DB) SetShortURL(shortUrl string, url string) {
	db.ShortURL[shortUrl] = url
}
