package storage

type Db struct {
	ShortUrl map[string]string
}

func NewDbConn() *Db {
	return &Db{
		ShortUrl: make(map[string]string),
	}

}
