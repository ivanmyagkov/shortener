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
	query := `SELECT base_url FROM urls WHERE short_url=$1`
	D.db.QueryRowContext(ctx, query, shortURL).Scan(&baseURL)
	if baseURL == "" {
		return "", interfaces.ErrNotFound
	}
	return baseURL, nil
}

func (D *Storage) GetAllURLsByUserID(userID string) ([]interfaces.ModelURL, error) {
	var modelURL []interfaces.ModelURL
	var model interfaces.ModelURL
	selectStmt, err := D.db.Prepare("SELECT short_url, base_url FROM users_url RIGHT JOIN urls u on users_url.url_id=u.id WHERE user_id=$1;")
	if err != nil {
		return nil, err
	}
	defer selectStmt.Close()

	rows, err := selectStmt.Query(userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&model.ShortURL, &model.BaseURL)
		if err != nil {
			return nil, err
		}
		modelURL = append(modelURL, model)
	}

	return modelURL, nil
}

func (D *Storage) SetShortURL(userID, shortURL, baseURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var urlID int
	query := `INSERT INTO urls (base_url, short_url) VALUES ($1, $2) RETURNING id `
	D.db.QueryRowContext(ctx, query, baseURL, shortURL).Scan(&urlID)
	if urlID != 0 {
		query = `INSERT INTO users_url (user_id, url_id) VALUES ($1, $2);`
		_, err := D.db.ExecContext(ctx, query, userID, urlID)
		if err != nil {
			return err
		}
	} else {
		var userURLID int
		querySelect := `SELECT id FROM urls WHERE base_url = $1;`
		D.db.QueryRowContext(ctx, querySelect, baseURL).Scan(&userURLID)
		query = `INSERT INTO users_url (user_id, url_id) VALUES ($1, $2);`
		_, err := D.db.ExecContext(ctx, query, userID, urlID)
		if err != nil {
			return interfaces.ErrAlreadyExists
		}
	}
	return nil
}

func (D *Storage) Ping() error {
	return D.db.Ping()
}

func createTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS urls (
		id serial primary key,
		base_url text not null unique,
		short_url text not null 
	);
	CREATE TABLE IF NOT EXISTS users_url(
	  user_id text not null,
	  url_id int not null references urls(id)
	);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil

}

func (D *Storage) Close() {
	D.db.Close()
}
