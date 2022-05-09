package interfaces

type Storage interface {
	GetURL(string) string
	SetShortURL(string, string)
}
