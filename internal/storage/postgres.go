package storage

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
)

type Storage struct {
	db *sql.DB
}

func NewDB(psqlConn string) *Storage {
	db, err := sql.Open("postgres", psqlConn)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to DB!")
	return &Storage{
		db: db,
	}
}

func (D *Storage) GetURL(shortURL string) (string, error) {
	panic("implement me")
}

func (D *Storage) GetAllURLsByUserID(userID string) ([]interfaces.ModelURL, error) {
	panic("implement me")
}

func (D *Storage) SetShortURL(userID, shortURL, baseURL string) error {
	panic("implement me")
}

func (D *Storage) Ping() error {
	return D.db.Ping()
}

func (D *Storage) Close() {
	D.db.Close()
}
