package interfaces

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ItIsNotURL       = errors.New("it isn't URL")
)

type Storage interface {
	GetURL(string) (string, error)
	SetShortURL(string, string) error
	Close()
}

type Config interface {
	SrvAddr() string
	HostName() string
}
