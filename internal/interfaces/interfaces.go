package interfaces

type Storage interface {
	GetUrl(string) string
	SetShortUrl(string, string)
}
