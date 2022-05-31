package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
)

type Storage struct {
	db *sql.DB
}

func NewDB(psqlConn string) (*Storage, error) {
	db, err := sql.Open("postgres", psqlConn)
	if err != nil {
		return nil, interfaces.ErrDBConn
	}

	if err = db.Ping(); err != nil {
		return nil, interfaces.ErrPingDB
	}
	log.Println("Connected to DB!")
	if err = createTable(db); err != nil {
		return nil, interfaces.ErrCreateTable
	}
	return &Storage{
		db: db,
	}, nil
}

func (D *Storage) GetURL(shortURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var baseURL string
	query := `SELECT base_url FROM urls WHERE short_url=$1`
	err := D.db.QueryRowContext(ctx, query, shortURL).Scan(&baseURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", interfaces.ErrNotFound
		}
		return "", err
	}

	return baseURL, nil
}

func (D *Storage) GetAllURLsByUserID(userID string) ([]interfaces.ModelURL, error) {
	var modelURL []interfaces.ModelURL
	var model interfaces.ModelURL
	rows, err := D.db.Query("SELECT short_url, base_url FROM users_url RIGHT JOIN urls u on users_url.url_id=u.id WHERE user_id=$1;", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		return nil, err
	}
	for rows.Next() {
		if err = rows.Scan(&model.ShortURL, &model.BaseURL); err != nil {
			return nil, err
		}
		modelURL = append(modelURL, model)
	}

	return modelURL, nil
}

func (D *Storage) SetShortURL(userID, shortURL, baseURL string) error {
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()
	var urlID int
	query := `INSERT INTO urls (base_url, short_url) VALUES ($1, $2) RETURNING id `
	err := D.db.QueryRow(query, baseURL, shortURL).Scan(&urlID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			query = `INSERT INTO users_url (user_id, url_id) VALUES ($1, $2);`
			_, err := D.db.Exec(query, userID, urlID)
			if err != nil {
				return err
			}
		}
		return err
	}

	var userURLID int
	querySelect := `SELECT id FROM urls WHERE base_url = $1;`
	err = D.db.QueryRow(querySelect, baseURL).Scan(&userURLID)
	if errors.Is(err, sql.ErrNoRows) {
		return err
	}
	query = `INSERT INTO users_url (user_id, url_id) VALUES ($1, $2) ;`
	_, err = D.db.Exec(query, userID, userURLID)
	if err != nil {
		return interfaces.ErrAlreadyExists
	}
	log.Println(err)
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
	  user_id text not null ,
	  url_id int not null  references urls(id),
	  CONSTRAINT unique_url UNIQUE (user_id, url_id)
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil

}

func (D *Storage) Close() {
	D.db.Close()
}
