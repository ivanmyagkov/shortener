//	Package interfaces for storing interfaces
package interfaces

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrDBConn        = errors.New("DB connection error")
	ErrCreateTable   = errors.New("create tables error")
	ErrPingDB        = errors.New("ping Db error")
	ErrWasDeleted    = errors.New("was deleted")
)

type Storage interface {
	GetURL(shortURL string) (string, error)
	GetAllURLsByUserID(userID string) ([]ModelURL, error)
	SetShortURL(userID, shortURL, baseURL string) error
	DelBatchShortURLs(tasks []Task) error
	Ping() error
	Close() error
}

type Config interface {
	SrvAddr() string
	HostName() string
}

type Users interface {
	CreateSissionID(string2 string) (string, error)
	ReadSessionID(id string) (string, error)
}

type InWorker interface {
	Do(t Task)
	Loop() error
}
type Task struct {
	ID       string
	ShortURL string
}

type ModelURL struct {
	ShortURL string `json:"short_url"`
	BaseURL  string `json:"original_url"`
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
