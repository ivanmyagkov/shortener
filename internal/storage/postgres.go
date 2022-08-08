package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
)

type Storage struct {
	db *sql.DB
}

//	Create DB connection.
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

//	Get original URL from DB.
func (D *Storage) GetURL(shortURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var baseURL string
	var isDeleted bool
	query := `SELECT base_url, is_deleted from urls right join users_url uu on urls.id = uu.url_id where short_url=$1`
	rows, err := D.db.QueryContext(ctx, query, shortURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", interfaces.ErrNotFound
		}
		return "", err
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		return "", err
	}
	for rows.Next() {
		if err = rows.Scan(&baseURL, &isDeleted); err != nil {
			return "", err
		}
		if !isDeleted {
			break
		}
	}

	if isDeleted {
		return "", interfaces.ErrWasDeleted
	}

	return baseURL, nil
}

//	Get all user URLs from DB.
func (D *Storage) GetAllURLsByUserID(userID string) ([]interfaces.ModelURL, error) {
	modelURL := make([]interfaces.ModelURL, 1000)
	var model interfaces.ModelURL
	rows, err := D.db.Query("SELECT short_url, base_url FROM users_url RIGHT JOIN urls u on users_url.url_id=u.id WHERE user_id=$1 and is_deleted=$2;", userID, false)
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

//	Delete user URLs from DB.
func (D *Storage) DelBatchShortURLs(tasks []interfaces.Task) error {

	query := `UPDATE users_url SET is_deleted = true from (SELECT id FROM urls RIGHT JOIN users_url uu on urls.id = uu.url_id WHERE user_id=$1 and short_url=$2) as urls`
	for _, task := range tasks {
		_, err := D.db.Exec(query, task.ID, task.ShortURL)
		if err != nil {
			return err
		}
	}
	return nil
}

//	Add new URL in DB.
func (D *Storage) SetShortURL(userID, shortURL, baseURL string) error {
	var urlID int
	query := `INSERT INTO urls (base_url, short_url) VALUES ($1, $2) RETURNING id `
	err := D.db.QueryRow(query, baseURL, shortURL).Scan(&urlID)
	if err != nil {
		querySelect := `SELECT id FROM urls WHERE base_url = $1;`
		err = D.db.QueryRow(querySelect, baseURL).Scan(&urlID)
		if err != nil {
			return err
		}

		query = `INSERT INTO users_url (user_id, url_id) VALUES ($1, $2) ;`
		_, err = D.db.Exec(query, userID, urlID)
		if err != nil {
			errCode := err.(*pq.Error).Code
			if pgerrcode.IsIntegrityConstraintViolation(string(errCode)) {
				var isDel bool
				query = `SELECT is_deleted from users_url where user_id=$1 and url_id=$2`
				err = D.db.QueryRow(query, userID, urlID).Scan(&isDel)
				if err != nil {
					return err
				}
				if !isDel {
					return interfaces.ErrAlreadyExists
				}
				updateQuery := `UPDATE users_url SET is_deleted = false WHERE user_id = $1 AND url_id = $2`
				_, err = D.db.Exec(updateQuery, userID, urlID)
				if err != nil {
					return err
				}
				return nil
			}
			return err
		}
		return nil
	}

	query = `INSERT INTO users_url (user_id, url_id) VALUES ($1, $2);`
	_, err = D.db.Exec(query, userID, urlID)
	if err != nil {
		errCode := err.(*pq.Error).Code
		if pgerrcode.IsIntegrityConstraintViolation(string(errCode)) {
			return interfaces.ErrAlreadyExists
		}
		return err
	}
	return nil
}

//	Ping DB connection.
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
	  is_deleted boolean default false,
	  CONSTRAINT unique_url UNIQUE (user_id, url_id)
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (D *Storage) Close() error {
	err := D.db.Close()
	return err
}
