package interfaces

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrNoContent     = errors.New("no contents for this users")
)

type Storage interface {
	GetURL(shortURL string) (string, error)
	GetAllURLsByUserID(userID string) ([]ModelURL, error)
	SetShortURL(userID, shortURL, baseURL string) error
	Close()
}

type Config interface {
	SrvAddr() string
	HostName() string
}

type Users interface {
	CreateSissionID(string2 string) (string, error)
	ReadSessionID(id string) (string, error)
}

type ModelURL struct {
	ShortURL string `json:"short_url"`
	BaseURL  string `json:"original_url"`
}
