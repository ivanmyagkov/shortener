package interfaces

type Storage interface {
	GetURL(string) string
	SetShortURL(string, string)
}

type Config interface {
	SrvAddr() string
	HostName() string
}
