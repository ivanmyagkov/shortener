package storage

import (
	"context"
	"database/sql"
	"log"
	"time"

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
	if err = createTable(db); err != nil {
		log.Fatal(err)
	}
	return &Storage{
		db: db,
	}
}

func (D *Storage) GetURL(shortURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var baseURL string
	log.Println(shortURL)
	query := `SELECT base_url FROM urls WHERE short_url=$1`
	D.db.QueryRowContext(ctx, query, shortURL).Scan(&baseURL)
	log.Println(baseURL)
	if baseURL == "" {
		return "", interfaces.ErrNotFound
	}
	return baseURL, nil
}

func (D *Storage) GetAllURLsByUserID(userID string) ([]interfaces.ModelURL, error) {

	panic("implement me")
}

func (D *Storage) SetShortURL(userID, shortURL, baseURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user int
	query := `SELECT 1 FROM urls WHERE short_url=$1 and user_id = $2 `
	D.db.QueryRowContext(ctx, query, shortURL, userID).Scan(&user)
	if user == 0 {
		query := `INSERT INTO urls (user_id, base_url,short_url) VALUES ($1,$2,$3) `
		_, err := D.db.ExecContext(ctx, query, userID, baseURL, shortURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (D *Storage) Ping() error {
	return D.db.Ping()
}

func createTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS urls (
				user_id text not null,
				base_url text not null,
				short_url text not null);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil

}

func (D *Storage) Close() {
	D.db.Close()
}
